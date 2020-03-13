package syncker

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/wire"
	"time"
)

type BeaconPeerState struct {
	Timestamp      int64
	BestViewHash   string
	BestViewHeight uint64
	processed      bool
}

type BeaconSyncProcess struct {
	status                         string //stop, running
	isCommittee                    bool
	isCatchUp                      bool
	beaconPeerStates               map[string]BeaconPeerState //sender -> state
	beaconPeerStateCh              chan *wire.MessagePeerState
	server                         Server
	chain                          Chain
	beaconPool                     *BlkPool
	s2bSyncProcess                 *S2BSyncProcess
	actionCh                       chan func()
	lastProcessConfirmBeaconHeight uint64
	lastCrossShardState            map[byte]map[byte]uint64
}

func NewBeaconSyncProcess(server Server, chain BeaconChainInterface) *BeaconSyncProcess {
	s := &BeaconSyncProcess{
		status:                         STOP_SYNC,
		server:                         server,
		chain:                          chain,
		beaconPool:                     NewBlkPool("BeaconPool"),
		beaconPeerStates:               make(map[string]BeaconPeerState),
		beaconPeerStateCh:              make(chan *wire.MessagePeerState),
		actionCh:                       make(chan func()),
		lastProcessConfirmBeaconHeight: 1,
		lastCrossShardState:            make(map[byte]map[byte]uint64),
	}
	s.s2bSyncProcess = NewS2BSyncProcess(server, s, chain)
	go s.syncBeacon()
	go s.insertBeaconBlockFromPool()
	go s.updateConfirmCrossShard()
	return s
}

func (s *BeaconSyncProcess) start(isCommittee bool) {
	if s.status == RUNNING_SYNC {
		return
	}
	s.isCommittee = isCommittee
	s.status = RUNNING_SYNC

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		for {
			if s.isCommittee {
				s.s2bSyncProcess.start()
			} else {
				s.s2bSyncProcess.stop()
			}

			select {
			case f := <-s.actionCh:
				f()
			case beaconPeerState := <-s.beaconPeerStateCh:
				s.beaconPeerStates[beaconPeerState.SenderID] = BeaconPeerState{
					Timestamp:      beaconPeerState.Timestamp,
					BestViewHash:   beaconPeerState.Beacon.BlockHash.String(),
					BestViewHeight: beaconPeerState.Beacon.Height,
				}
			case <-ticker.C:
				s.chain.SetReady(s.isCatchUp)
				for sender, ps := range s.beaconPeerStates {
					if ps.Timestamp < time.Now().Unix()-10 {
						delete(s.beaconPeerStates, sender)
					}
				}
			}
			if s.status != RUNNING_SYNC {
				time.Sleep(time.Second)
				continue
			}
		}
	}()

}

func (s *BeaconSyncProcess) stop() {
	s.status = STOP_SYNC
	s.s2bSyncProcess.stop()
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

type NextCrossShardInfo struct {
	nextCrossShardHeight uint64
	nextCrossShardHash   string
	confirmBeaconHeight  uint64
	confirmBeaconHash    string
}

//watching confirm beacon block and update cross shard info (which beacon confirm crossshard block N of shard X)
func (s *BeaconSyncProcess) updateConfirmCrossShard() {
	//TODO: update lastUpdateConfirmCrossShard using DB
	for {
		if s.lastProcessConfirmBeaconHeight > s.chain.GetFinalViewHeight() {
			time.Sleep(time.Second * 5)
			continue
		}
		beaconBlock, err := s.server.FetchConfirmBeaconBlockByHeight(s.lastProcessConfirmBeaconHeight)
		if err != nil || beaconBlock == nil {
			time.Sleep(time.Second * 5)
			continue
		}

		err = processBeaconForConfirmmingCrossShard(s.server.GetIncDatabase(), beaconBlock, s.lastCrossShardState)
		if err == nil {
			s.lastProcessConfirmBeaconHeight++
		} else {
			fmt.Println(err)
			time.Sleep(time.Second * 5)
		}

	}
}

func processBeaconForConfirmmingCrossShard(database incdb.Database, beaconBlock *blockchain.BeaconBlock, lastCrossShardState map[byte]map[byte]uint64) error {
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

					info := NextCrossShardInfo{
						waitHeight,
						shardBlock.Hash.String(),
						beaconBlock.GetHeight(),
						beaconBlock.Hash().String(),
					}
					b, _ := json.Marshal(info)
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

//periodically check pool and insert into pool
func (s *BeaconSyncProcess) insertBeaconBlockFromPool() {
	defer func() {
		if s.isCatchUp {
			time.AfterFunc(time.Millisecond*100, s.insertBeaconBlockFromPool)
		} else {
			time.AfterFunc(time.Second*1, s.insertBeaconBlockFromPool)
		}
	}()

	if !s.isCatchUp {
		return
	}
	var blk common.BlockPoolInterface
	blk = s.beaconPool.GetNextBlock(s.chain.GetBestViewHash())

	if isNil(blk) {
		return
	}

	fmt.Println("Syncker: Insert beacon from pool", blk.(common.BlockInterface).GetHeight())
	s.beaconPool.RemoveBlock(blk.Hash().String())
	if err := s.chain.ValidateBlockSignatures(blk.(common.BlockInterface), s.chain.GetCommittee()); err != nil {
		return
	}

	if err := s.chain.InsertBlk(blk.(common.BlockInterface)); err != nil {
	}
}

//sync beacon
func (s *BeaconSyncProcess) syncBeacon() {
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
	blockBuffer := []common.BlockInterface{}
	defer func() {
		if requestCnt == 0 {
			pState.processed = true
		}
		cancel()
	}()

	toHeight := pState.BestViewHeight
	//process param

	if toHeight <= s.chain.GetBestViewHeight() {
		return
	}

	//stream
	ch, err := s.server.RequestBeaconBlocksViaStream(ctx, peerID, s.chain.GetBestViewHeight()+1, toHeight)
	if err != nil {
		fmt.Println("Syncker: create channel fail")
		return
	}

	//receive
	requestCnt++
	insertTime := time.Now()
	for {
		select {
		case blk := <-ch:
			if !isNil(blk) {
				blockBuffer = append(blockBuffer, blk)
			}

			if len(blockBuffer) >= 350 || (len(blockBuffer) > 0 && (isNil(blk) || time.Since(insertTime) > time.Millisecond*1000)) {
				insertBlkCnt := 0
				for {
					if successBlk, err := InsertBatchBlock(s.chain, blockBuffer); err != nil {
						return
					} else {
						insertBlkCnt += successBlk
						if successBlk >= len(blockBuffer) {
							break
						}
						blockBuffer = blockBuffer[successBlk:]
					}
				}
				insertTime = time.Now()
				blockBuffer = []common.BlockInterface{}
			}
			if isNil(blk) && len(blockBuffer) == 0 {
				return
			}
		}
	}
}
