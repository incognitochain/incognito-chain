package syncker

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/consensus/consensustypes"
	"github.com/incognitochain/incognito-chain/peerv2"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/wire"
)

type ShardPeerState struct {
	Timestamp      int64
	BestViewHash   string
	BestViewHeight uint64
	processed      bool
}

type ShardSyncProcess struct {
	isCommittee            bool
	isCatchUp              bool
	finalBeaconBlockHeight uint64
	shardID                int
	status                 string                    //stop, running
	shardPeerState         map[string]ShardPeerState //peerid -> state
	shardPeerStateCh       chan *wire.MessagePeerState
	crossShardSyncProcess  *CrossShardSyncProcess
	blockchain             *blockchain.BlockChain
	Network                Network
	Chain                  ShardChainInterface
	beaconChain            Chain
	shardPool              *BlkPool
	actionCh               chan func()
	consensus              peerv2.ConsensusData
	lock                   *sync.RWMutex
}

func NewShardSyncProcess(
	shardID int,
	network Network,
	bc *blockchain.BlockChain,
	beaconChain BeaconChainInterface,
	chain ShardChainInterface,
	consensus peerv2.ConsensusData,
) *ShardSyncProcess {
	var isOutdatedBlock = func(blk interface{}) bool {
		if blk.(*types.ShardBlock).GetHeight() < chain.GetFinalViewHeight() {
			return true
		}
		return false
	}

	s := &ShardSyncProcess{
		shardID:          shardID,
		status:           STOP_SYNC,
		blockchain:       bc,
		Network:          network,
		Chain:            chain,
		beaconChain:      beaconChain,
		shardPool:        NewBlkPool("ShardPool-"+string(shardID), isOutdatedBlock),
		shardPeerState:   make(map[string]ShardPeerState),
		shardPeerStateCh: make(chan *wire.MessagePeerState),
		consensus:        consensus,

		actionCh: make(chan func()),
	}
	s.crossShardSyncProcess = NewCrossShardSyncProcess(network, bc, s, beaconChain)

	go s.syncShardProcess()
	go s.insertShardBlockFromPool()

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		for {
			if s.isCommittee {
				s.crossShardSyncProcess.start()
			} else {
				s.crossShardSyncProcess.stop()
			}

			select {
			case f := <-s.actionCh:
				f()
			case shardPeerState := <-s.shardPeerStateCh:
				for sid, peerShardState := range shardPeerState.Shards {
					if int(sid) == s.shardID {
						s.shardPeerState[shardPeerState.SenderID] = ShardPeerState{
							Timestamp:      shardPeerState.Timestamp,
							BestViewHash:   peerShardState.BlockHash.String(),
							BestViewHeight: peerShardState.Height,
						}
						s.Chain.SetReady(true)
					}
				}
			case <-ticker.C:
				for sender, ps := range s.shardPeerState {
					if ps.Timestamp < time.Now().Unix()-10 {
						delete(s.shardPeerState, sender)
					}
				}
			}
		}
	}()

	return s
}

func (s *ShardSyncProcess) start() {
	if s.status == RUNNING_SYNC {
		return
	}
	s.status = RUNNING_SYNC
}

func (s *ShardSyncProcess) stop() {
	s.status = STOP_SYNC
	s.crossShardSyncProcess.stop()
}

//helper function to access map atomically
func (s *ShardSyncProcess) getShardPeerStates() map[string]ShardPeerState {
	res := make(chan map[string]ShardPeerState)
	s.actionCh <- func() {
		ps := make(map[string]ShardPeerState)
		for k, v := range s.shardPeerState {
			ps[k] = v
		}
		res <- ps
	}
	return <-res
}

//periodically check pool and insert shard block to chain
var insertShardTimeCache, _ = lru.New(10000)

func (s *ShardSyncProcess) insertShardBlockFromPool() {

	insertCnt := 0
	defer func() {
		if insertCnt > 0 {
			s.insertShardBlockFromPool()
		} else {
			time.AfterFunc(time.Second*2, s.insertShardBlockFromPool)
		}
	}()

	//loop all current views, if there is any block connect to the view
	for _, viewHash := range s.Chain.GetAllViewHash() {
		blocks := s.shardPool.GetBlockByPrevHash(viewHash)
		for _, block := range blocks {
			if block == nil {
				continue
			}
			//if already insert and error, last time insert is < 10s then we skip
			insertTime, ok := insertShardTimeCache.Get(viewHash.String())
			if ok && time.Since(insertTime.(time.Time)).Seconds() < 10 {
				continue
			}

			//fullnode delay 1 block (make sure insert final block)
			if os.Getenv("FULLNODE") != "" {
				preBlk := s.shardPool.GetBlockByPrevHash(*block.Hash())
				if len(preBlk) == 0 {
					continue
				}
			}

			insertShardTimeCache.Add(viewHash.String(), time.Now())
			insertCnt++
			//must validate this block when insert
			if err := s.Chain.InsertBlock(block.(types.BlockInterface), true); err != nil {
				Logger.Error("Insert shard block from pool fail", block.GetHeight(), block.Hash(), err)
				continue
			} else {
				previousValidationData := s.shardPool.GetPreviousValidationData(block.GetPrevHash())
				if previousValidationData == common.EmptyString {
					continue
				}
				_, err := consensustypes.DecodeValidationData(previousValidationData)
				if err != nil {
					continue
				}
				err1 := s.Chain.ReplacePreviousValidationData(block.GetPrevHash(), previousValidationData)
				if err1 != nil {
					Logger.Error("Replace Previous Validation Data Fail", block.GetPrevHash(), previousValidationData, err)
				}
			}
			s.shardPool.RemoveBlock(block)
		}
	}
}

func (s *ShardSyncProcess) syncShardProcess() {
	for {
		requestCnt := 0
		if s.status != RUNNING_SYNC {
			s.isCatchUp = false
			time.Sleep(time.Second * 5)
			continue
		}

		for peerID, pState := range s.getShardPeerStates() {
			requestCnt += s.streamFromPeer(peerID, pState)
		}

		if requestCnt > 0 {
			s.isCatchUp = false
			// s.syncShardProcess()
		} else {
			if len(s.shardPeerState) > 0 {
				s.isCatchUp = true
				committeeView := s.blockchain.BeaconChain.GetBestView().(*blockchain.BeaconBestState)
				if s.finalBeaconBlockHeight < committeeView.BeaconHeight {
					s.finalBeaconBlockHeight = committeeView.BeaconHeight
					if committeeView.CommitteeStateVersion() == committeestate.DCS_VERSION {
						syncingValidators := s.consensus.SyncingValidatorsByShardID(s.shardID)
						if committeeView.ShouldSendFinishSyncMessage(syncingValidators, byte(s.shardID)) {
							msg := &wire.MessageFinishSync{
								CommitteePublicKey: syncingValidators,
								ShardID:            byte(s.shardID),
							}
							s.Network.PublishMessageToShard(msg, common.BeaconChainID)
						}
					}
				}
			}
			time.Sleep(time.Second * 5)
		}
	}
}

func (s *ShardSyncProcess) streamFromPeer(peerID string, pState ShardPeerState) (requestCnt int) {
	if pState.processed {
		return
	}

	blockBuffer := []types.BlockInterface{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer func() {
		if requestCnt == 0 {
			pState.processed = true
		}
		cancel()
	}()

	if pState.processed {
		return
	}
	toHeight := pState.BestViewHeight

	//fullnode delay 1 block (make sure insert final block)
	if os.Getenv("FULLNODE") != "" {
		toHeight = pState.BestViewHeight - 1
		if toHeight <= s.Chain.GetBestViewHeight() {
			return
		}
	}

	//if is behind, and
	//if peerstate show fork, sync that peerID
	if pState.BestViewHeight < s.Chain.GetBestViewHeight() || (pState.BestViewHeight == s.Chain.GetBestViewHeight() && s.Chain.GetBestViewHash() == pState.BestViewHash) {
		return
	}

	if pState.BestViewHeight == s.Chain.GetBestViewHeight() && s.Chain.GetBestViewHash() != pState.BestViewHash {
		for _, h := range s.Chain.GetAllViewHash() { //check if block exist in multiview, then return
			if h.String() == pState.BestViewHash {
				return
			}
		}
	}

	if pState.BestViewHeight > s.Chain.GetBestViewHeight() {
		requestCnt++
		peerID = ""
	}

	//incase, we have long multiview chain, just sync last 100 block (very low probability that we have fork more than 100 blocks
	fromHeight := s.Chain.GetFinalViewHeight() + 1
	if s.Chain.GetBestViewHeight()-100 > fromHeight {
		fromHeight = s.Chain.GetBestViewHeight()
	}

	//stream
	ch, err := s.Network.RequestShardBlocksViaStream(ctx, peerID, s.shardID, fromHeight, toHeight)
	if err != nil || ch == nil {
		fmt.Println("Syncker: create channel fail")
		requestCnt = 0
		return
	}

	insertTime := time.Now()
	for {
		select {
		case blk := <-ch:
			if !isNil(blk) {
				blockBuffer = append(blockBuffer, blk)

				if blk.(*types.ShardBlock).Header.BeaconHeight > s.beaconChain.GetBestViewHeight() {
					time.Sleep(30 * time.Second)
				}
				// if blk.(*blockchain.ShardBlock).Header.BeaconHeight > s.beaconChain.GetBestViewHeight() {
				// 	Logger.Infof("Cannot find beacon for inserting shard block")
				// 	return
				// }
			}

			if uint64(len(blockBuffer)) >= 500 || (len(blockBuffer) > 0 && (isNil(blk) || time.Since(insertTime) > time.Millisecond*2000)) {
				insertBlkCnt := 0
				for {
					time1 := time.Now()
					if successBlk, err := InsertBatchBlock(s.Chain, blockBuffer); err != nil {
						Logger.Errorf("Fail to Insert Batch Block, %+v", err)
						return
					} else {
						insertBlkCnt += successBlk
						fmt.Printf("Syncker Insert %d shard %d block(from %d to %d) elaspse %f \n", successBlk, s.shardID, blockBuffer[0].GetHeight(), blockBuffer[len(blockBuffer)-1].GetHeight(), time.Since(time1).Seconds())
						if successBlk >= len(blockBuffer) || successBlk == 0 {
							break
						}
						blockBuffer = blockBuffer[successBlk:]
					}
				}

				insertTime = time.Now()
				blockBuffer = []types.BlockInterface{}
			}

			if isNil(blk) && len(blockBuffer) == 0 {
				return
			}
		}
	}

}
