package syncker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"os"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/types"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/wire"
)

type BeaconPeerState struct {
	Timestamp      int64
	BestViewHash   string
	BestViewHeight uint64
	processed      bool
}

type BeaconSyncProcess struct {
	status              string //stop, running
	isCommittee         bool
	isCatchUp           bool
	beaconPeerStates    map[string]BeaconPeerState //sender -> state
	beaconPeerStateCh   chan *wire.MessagePeerState
	blockchain          *blockchain.BlockChain
	network             Network
	chain               Chain
	beaconPool          *BlkPool
	actionCh            chan func()
	lastCrossShardState map[byte]map[byte]uint64
	lastInsert          string
}

func NewBeaconSyncProcess(network Network, bc *blockchain.BlockChain, chain BeaconChainInterface) *BeaconSyncProcess {

	var isOutdatedBlock = func(blk interface{}) bool {
		if blk.(*types.BeaconBlock).GetHeight() < chain.GetFinalViewHeight() {
			return true
		}
		return false
	}

	s := &BeaconSyncProcess{
		status:              STOP_SYNC,
		blockchain:          bc,
		network:             network,
		chain:               chain,
		beaconPool:          NewBlkPool("BeaconPool", isOutdatedBlock),
		beaconPeerStates:    make(map[string]BeaconPeerState),
		beaconPeerStateCh:   make(chan *wire.MessagePeerState),
		actionCh:            make(chan func()),
		lastCrossShardState: make(map[byte]map[byte]uint64),
	}
	go s.syncBeacon()
	go s.insertBeaconBlockFromPool()
	go s.updateConfirmCrossShard()

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		lastHeight := s.chain.GetBestViewHeight()
		for {
			select {
			case f := <-s.actionCh:
				f()
			case beaconPeerState := <-s.beaconPeerStateCh:
				Logger.Debugf("Got new peerstate, last height %v", beaconPeerState.Beacon.Height)
				s.beaconPeerStates[beaconPeerState.SenderID] = BeaconPeerState{
					Timestamp:      beaconPeerState.Timestamp,
					BestViewHash:   beaconPeerState.Beacon.BlockHash.String(),
					BestViewHeight: beaconPeerState.Beacon.Height,
				}
				s.chain.SetReady(true)
			case <-ticker.C:
				for sender, ps := range s.beaconPeerStates {
					if ps.Timestamp < time.Now().Unix()-10 {
						delete(s.beaconPeerStates, sender)
					}
				}
				if lastHeight != s.chain.GetBestViewHeight() {
					s.lastInsert = time.Now().Format("2006-01-02T15:04:05-0700")
					lastHeight = s.chain.GetBestViewHeight()
				}

			}
			if s.status != RUNNING_SYNC {
				time.Sleep(time.Second)
				continue
			}
		}
	}()

	return s
}

func (s *BeaconSyncProcess) start() {
	if s.status == RUNNING_SYNC {
		return
	}
	s.status = RUNNING_SYNC

}

func (s *BeaconSyncProcess) stop() {
	s.status = STOP_SYNC
}

//helper function to access map in atomic way
func (s *BeaconSyncProcess) getBeaconPeerStates() map[string]BeaconPeerState {
	res := make(chan map[string]BeaconPeerState)
	s.actionCh <- func() {
		ps := make(map[string]BeaconPeerState)
		for k, v := range s.beaconPeerStates {
			ps[k] = v
		}
		res <- ps
	}
	return <-res
}

type LastCrossShardBeaconProcess struct {
	BeaconHeight        uint64
	LastCrossShardState map[byte]map[byte]uint64
}

//watching confirm beacon block and update cross shard info (which beacon confirm crossshard block N of shard X)
func (s *BeaconSyncProcess) updateConfirmCrossShard() {
	state := rawdbv2.GetLastBeaconStateConfirmCrossShard(s.chain.GetDatabase())
	lastBeaconStateConfirmCrossX := new(LastCrossShardBeaconProcess)
	_ = json.Unmarshal(state, &lastBeaconStateConfirmCrossX)
	lastBeaconHeightConfirmCrossX := uint64(1)
	if lastBeaconStateConfirmCrossX.BeaconHeight != 0 {
		s.lastCrossShardState = lastBeaconStateConfirmCrossX.LastCrossShardState
		lastBeaconHeightConfirmCrossX = lastBeaconStateConfirmCrossX.BeaconHeight
	}
	Logger.Info("lastBeaconHeightConfirmCrossX", lastBeaconHeightConfirmCrossX)
	for {
		if s.status != RUNNING_SYNC {
			time.Sleep(time.Second)
			continue
		}
		if lastBeaconHeightConfirmCrossX > s.chain.GetFinalViewHeight() {
			//fmt.Println("DEBUG:larger than final view", s.chain.GetFinalViewHeight())
			time.Sleep(time.Second * 5)
			continue
		}
		beaconBlock, err := s.blockchain.FetchConfirmBeaconBlockByHeight(lastBeaconHeightConfirmCrossX)
		if err != nil || beaconBlock == nil {
			//fmt.Println("DEBUG: cannot find beacon block", lastBeaconHeightConfirmCrossX)
			time.Sleep(time.Second * 5)
			continue
		}
		err = processBeaconForConfirmmingCrossShard(s.chain.GetDatabase(), beaconBlock, s.lastCrossShardState)
		if err == nil {
			lastBeaconHeightConfirmCrossX++
			if lastBeaconHeightConfirmCrossX%1000 == 0 {
				Logger.Info("store lastBeaconHeightConfirmCrossX", lastBeaconHeightConfirmCrossX)
				rawdbv2.StoreLastBeaconStateConfirmCrossShard(s.chain.GetDatabase(), LastCrossShardBeaconProcess{lastBeaconHeightConfirmCrossX, s.lastCrossShardState})
			}
		} else {
			Logger.Error(err)
			time.Sleep(time.Second * 5)
		}

	}
}

func processBeaconForConfirmmingCrossShard(database incdb.Database, beaconBlock *types.BeaconBlock, lastCrossShardState map[byte]map[byte]uint64) error {
	if beaconBlock != nil && beaconBlock.Body.ShardState != nil {
		for fromShard, shardBlocks := range beaconBlock.Body.ShardState {
			for _, shardBlock := range shardBlocks {
				for _, toShard := range shardBlock.CrossShard {

					if fromShard == toShard {
						continue
					}
					if lastCrossShardState[fromShard] == nil {
						lastCrossShardState[fromShard] = make(map[byte]uint64)
					}
					lastHeight := lastCrossShardState[fromShard][toShard] // get last cross shard height from shardID  to crossShardShardID
					waitHeight := shardBlock.Height

					info := blockchain.NextCrossShardInfo{
						waitHeight,
						shardBlock.Hash.String(),
						beaconBlock.GetHeight(),
						beaconBlock.Hash().String(),
					}
					//Logger.Info("DEBUG: processBeaconForConfirmmingCrossShard ", fromShard, toShard, info)
					b, _ := json.Marshal(info)
					Logger.Info("debug StoreCrossShardNextHeight", fromShard, toShard, lastHeight, string(b))
					err := rawdbv2.StoreCrossShardNextHeight(database, fromShard, toShard, lastHeight, b)
					if err != nil {
						return err
					}
					if lastCrossShardState[fromShard] == nil {
						lastCrossShardState[fromShard] = make(map[byte]uint64)
					}
					lastCrossShardState[fromShard][toShard] = waitHeight //update lastHeight to waitHeight
				}
			}
		}
	}
	return nil
}

//periodically check pool and insert into pool (in case some fork block or block come early)
var insertBeaconTimeCache, _ = lru.New(1000)

func (s *BeaconSyncProcess) insertBeaconBlockFromPool() {

	insertCnt := 0
	defer func() {
		if insertCnt > 0 {
			s.insertBeaconBlockFromPool()
		} else {
			time.AfterFunc(time.Second*2, s.insertBeaconBlockFromPool)
		}
	}()
	//Logger.Debugf("insertBeaconBlockFromPool Start")
	//loop all current views, if there is any block connect to the view
	for _, viewHash := range s.chain.GetAllViewHash() {
		blks := s.beaconPool.GetBlockByPrevHash(viewHash)
		for _, blk := range blks {
			if blk == nil {
				continue
			}
			//Logger.Debugf("insertBeaconBlockFromPool blk %v %v", blk.GetHeight(), blk.Hash().String())
			//if already insert and error, last time insert is < 10s then we skip
			insertTime, ok := insertBeaconTimeCache.Get(viewHash.String())
			if ok && time.Since(insertTime.(time.Time)).Seconds() < 10 {
				continue
			}

			//fullnode delay 1 block (make sure insert final block)
			if os.Getenv("FULLNODE") != "" {
				preBlk := s.beaconPool.GetBlockByPrevHash(*blk.Hash())
				if len(preBlk) == 0 {
					continue
				}
			}

			insertBeaconTimeCache.Add(viewHash.String(), time.Now())
			insertCnt++
			//must validate this block when insert
			if err := s.chain.InsertBlock(blk.(types.BlockInterface), common.BASIC_VALIDATION); err != nil {
				Logger.Error("Insert beacon block from pool fail", blk.GetHeight(), blk.Hash(), err)
				continue
			}
			s.beaconPool.RemoveBlock(blk)
		}
	}

}

func (s *BeaconSyncProcess) streamBlockFromHighway() chan *types.BeaconBlock {
	fromHeight := s.chain.GetBestViewHeight() + 1
	beaconCh := make(chan *types.BeaconBlock, 500)
	time.Sleep(time.Second * 5)
	go func() {
		for {
		REPEAT:
			ctx, _ := context.WithTimeout(context.Background(), time.Minute)
			ch, err := s.network.RequestBeaconBlocksViaStream(ctx, "", fromHeight, fromHeight+100)
			if err != nil || ch == nil {
				time.Sleep(time.Second * 30)
				continue
			}
			tmpHeight := fromHeight
			for {
				select {
				case blk := <-ch:
					if !isNil(blk) {
						beaconCh <- blk.(*types.BeaconBlock)
						fromHeight = blk.GetHeight() + 1
					} else {
						if tmpHeight == fromHeight {
							time.Sleep(time.Second * 20)
						}
						goto REPEAT
					}
				}
			}
		}

	}()
	return beaconCh
}

//sync beacon
func (s *BeaconSyncProcess) syncBeacon() {
	regression := os.Getenv("REGRESSION")

	//if regression, we sync from highway, not care about fork and peerstate
	if regression == "1" {
		beaconCh := s.streamBlockFromHighway()
		for {
			nextHeight := s.chain.GetBestViewHeight() + 1
			beaconBlock := <-beaconCh
			if nextHeight != beaconBlock.GetHeight() {
				Logger.Error("Something wrong", nextHeight, beaconBlock.GetHeight())
				panic(1)
			}

		BEACON_WAIT:
			shouldWait := false
			for sid, shardStates := range beaconBlock.Body.ShardState {
				if len(shardStates) > 0 && shardStates[len(shardStates)-1].Height > s.blockchain.GetChain(int(sid)).GetFinalView().GetHeight() {
					shouldWait = true
				}
			}
			if !shouldWait {
				err := s.chain.InsertBlock(beaconBlock, common.REGRESSION_TEST)
				if err != nil {
					Logger.Error(err)
					goto BEACON_WAIT
				}
			} else {
				time.Sleep(time.Millisecond * 5)
				goto BEACON_WAIT
			}
		}
		return
	}

	//if not regression, we sync from peer
	for {
		requestCnt := 0
		if s.status != RUNNING_SYNC {
			s.isCatchUp = false
			time.Sleep(time.Second)
			continue
		}

		for peerID, pState := range s.getBeaconPeerStates() {
			requestCnt += s.streamFromPeer(peerID, pState)
		}

		//last check, if we still need to sync more
		if requestCnt > 0 {
			s.isCatchUp = false
		} else {
			if len(s.beaconPeerStates) > 0 {
				s.isCatchUp = true
			}
			time.Sleep(time.Second * 5)
		}

	}
}

func (s *BeaconSyncProcess) streamFromPeer(peerID string, pState BeaconPeerState) (requestCnt int) {
	if pState.processed {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	blockBuffer := []types.BlockInterface{}
	defer func() {
		if requestCnt == 0 {
			pState.processed = true
		}
		cancel()
	}()

	toHeight := pState.BestViewHeight
	//process param

	//fullnode delay 1 block (make sure insert final block)
	if os.Getenv("FULLNODE") != "" {
		toHeight = toHeight - 1
		if toHeight <= s.chain.GetBestViewHeight() {
			return
		}
	}

	//if is behind, and
	//if peerstate show fork, sync that block
	if pState.BestViewHeight < s.chain.GetBestViewHeight() || (pState.BestViewHeight == s.chain.GetBestViewHeight() && s.chain.GetBestViewHash() == pState.BestViewHash) {
		return
	}

	if pState.BestViewHeight == s.chain.GetBestViewHeight() && s.chain.GetBestViewHash() != pState.BestViewHash {
		for _, h := range s.chain.GetAllViewHash() { //check if block exist in multiview, then return
			if h.String() == pState.BestViewHash {
				return
			}
		}
	}

	if pState.BestViewHeight > s.chain.GetBestViewHeight() {
		requestCnt++
		peerID = ""
	}

	//incase, we have long multiview chain, just sync last 100 block (very low probability that we have fork more than 100 blocks)
	fromHeight := s.chain.GetFinalViewHeight() + 1
	if s.chain.GetBestViewHeight()-100 > fromHeight {
		fromHeight = s.chain.GetBestViewHeight()
	}

	//stream
	ch, err := s.network.RequestBeaconBlocksViaStream(ctx, peerID, fromHeight, toHeight)
	if err != nil || ch == nil {
		fmt.Println("Syncker: create channel fail")
		requestCnt = 0
		return
	}

	//receive
	insertTime := time.Now()
	for {
		select {
		case blk := <-ch:
			if !isNil(blk) {
				blockBuffer = append(blockBuffer, blk)
			}

			if uint64(len(blockBuffer)) >= blockchain.DefaultMaxBlkReqPerPeer || (len(blockBuffer) > 0 && (isNil(blk) || time.Since(insertTime) > time.Millisecond*2000)) {
				insertBlkCnt := 0
				for {
					time1 := time.Now()

					/*for _, v := range blockBuffer {*/
					//Logger.Infof("[config] v height %v proposetime %v", v.GetHeight(), v.GetProposeTime())
					/*}*/

					if successBlk, err := InsertBatchBlock(s.chain, blockBuffer); err != nil {
						if successBlk == 0 {
							fmt.Println(err)
						}
						return
					} else {
						insertBlkCnt += successBlk
						Logger.Infof("Syncker Insert %d beacon block (from %d to %d) elaspse %f \n", successBlk, blockBuffer[0].GetHeight(), blockBuffer[len(blockBuffer)-1].GetHeight(), time.Since(time1).Seconds())
						if successBlk >= len(blockBuffer) || successBlk == 0 {
							return
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
