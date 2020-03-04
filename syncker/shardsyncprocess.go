package syncker

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/wire"
	"sync"
	"time"
)

type ShardPeerState struct {
	Timestamp      int64
	BestViewHash   string
	BestViewHeight uint64
	processed      bool
}

type ShardSyncProcess struct {
	IsCommittee           bool
	FewBlockBehind        bool
	ShardID               int
	Status                string                    //stop, running
	ShardPeerState        map[string]ShardPeerState //peerid -> state
	ShardPeerStateCh      chan *wire.MessagePeerState
	CrossShardSyncProcess *CrossShardSyncProcess
	Server                Server
	Chain                 ShardChainInterface
	BeaconChain           Chain
	ShardPool             *BlkPool
	actionCh              chan func()
	lock                  *sync.RWMutex
}

func NewShardSyncProcess(shardID int, server Server, beaconChain BeaconChainInterface, chain ShardChainInterface) *ShardSyncProcess {
	s := &ShardSyncProcess{
		ShardID:          shardID,
		Status:           STOP_SYNC,
		Server:           server,
		Chain:            chain,
		BeaconChain:      beaconChain,
		ShardPool:        NewBlkPool("ShardPool-" + string(shardID)),
		ShardPeerState:   make(map[string]ShardPeerState),
		ShardPeerStateCh: make(chan *wire.MessagePeerState),

		actionCh: make(chan func()),
	}
	s.CrossShardSyncProcess = NewCrossShardSyncProcess(server, s, beaconChain)

	go s.syncShardProcess()
	go s.insertShardBlockFromPool()
	return s
}

func (s *ShardSyncProcess) start() {
	if s.Status == RUNNING_SYNC {
		return
	}
	s.Status = RUNNING_SYNC

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		for {
			if s.IsCommittee {
				s.CrossShardSyncProcess.start()
			} else {
				s.CrossShardSyncProcess.stop()
			}

			select {
			case f := <-s.actionCh:
				f()
			case shardPeerState := <-s.ShardPeerStateCh:
				for sid, peerShardState := range shardPeerState.Shards {
					if int(sid) == s.ShardID {
						s.ShardPeerState[shardPeerState.SenderID] = ShardPeerState{
							Timestamp:      shardPeerState.Timestamp,
							BestViewHash:   peerShardState.BlockHash.String(),
							BestViewHeight: peerShardState.Height,
						}
					}
				}
			case <-ticker.C:
				s.Chain.SetReady(s.FewBlockBehind)
			}
		}
	}()

}

func (s *ShardSyncProcess) stop() {
	s.Status = STOP_SYNC
	s.CrossShardSyncProcess.stop()
}

func (s *ShardSyncProcess) getShardPeerStates() map[string]ShardPeerState {
	res := make(chan map[string]ShardPeerState)
	s.actionCh <- func() {
		ps := make(map[string]ShardPeerState)
		for k, v := range s.ShardPeerState {
			ps[k] = v
		}
		res <- ps
	}
	return <-res
}

func (s *ShardSyncProcess) insertShardBlockFromPool() {
	defer func() {
		if s.FewBlockBehind {
			time.AfterFunc(time.Millisecond*100, s.insertShardBlockFromPool)
		} else {
			time.AfterFunc(time.Second*1, s.insertShardBlockFromPool)
		}
	}()

	if !s.FewBlockBehind {
		return
	}
	var blk common.BlockPoolInterface
	blk = s.ShardPool.GetNextBlock(s.Chain.GetBestViewHash(), true)

	if isNil(blk) {
		return
	}

	fmt.Println("Syncker: Insert shard from pool", blk.(common.BlockInterface).GetHeight())
	s.ShardPool.RemoveBlock(blk.Hash().String())
	if err := s.Chain.ValidateBlockSignatures(blk.(common.BlockInterface), s.Chain.GetCommittee()); err != nil {
		return
	}

	if err := s.Chain.InsertBlk(blk.(common.BlockInterface)); err != nil {
	}
}

func (s *ShardSyncProcess) syncShardProcess() {
	for {
		requestCnt := 0
		if s.Status != RUNNING_SYNC {
			s.FewBlockBehind = false
			time.Sleep(time.Second * 5)
			continue
		}

		for peerID, pState := range s.getShardPeerStates() {
			requestCnt += s.streamFromPeer(peerID, pState)
		}

		if requestCnt > 0 {
			s.FewBlockBehind = false
			s.syncShardProcess()
		} else {
			if len(s.ShardPeerState) > 0 {
				s.FewBlockBehind = true
			}
			time.Sleep(time.Second * 5)
		}
	}

}

func (s *ShardSyncProcess) streamFromPeer(peerID string, pState ShardPeerState) (requestCnt int) {
	blockBuffer := []common.BlockInterface{}
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

	if pState.BestViewHeight <= s.Chain.GetBestViewHeight() {
		return
	}

	//fmt.Println("SYNCKER Request Shard Block", peerID, s.ShardID, s.Chain.GetBestViewHeight()+1, pState.BestViewHeight)
	ch, err := s.Server.RequestShardBlocksViaStream(ctx, peerID, s.ShardID, s.Chain.GetBestViewHeight()+1, pState.BestViewHeight)
	if err != nil {
		fmt.Println("Syncker: create channel fail")
		return
	}

	requestCnt++
	insertTime := time.Now()
	for {
		select {
		case blk := <-ch:
			if !isNil(blk) {
				blockBuffer = append(blockBuffer, blk)
				if blk.(*blockchain.ShardBlock).Header.BeaconHeight > s.BeaconChain.GetBestViewHeight() {
					time.Sleep(5 * time.Second)
				}
				if blk.(*blockchain.ShardBlock).Header.BeaconHeight > s.BeaconChain.GetBestViewHeight() {
					return
				}
			}

			if len(blockBuffer) >= 350 || (len(blockBuffer) > 0 && (isNil(blk) || time.Since(insertTime) > time.Millisecond*1000)) {
				insertBlkCnt := 0
				for {
					time1 := time.Now()
					if successBlk, err := InsertBatchBlock(s.Chain, blockBuffer); err != nil {
						return
					} else {
						insertBlkCnt += successBlk
						fmt.Printf("Syncker Insert %d shard (from %d to %d) elaspse %f \n", successBlk, blockBuffer[0].GetHeight(), blockBuffer[len(blockBuffer)-1].GetHeight(), time.Since(time1).Seconds())
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
