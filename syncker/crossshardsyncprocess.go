package syncker

import (
	"github.com/incognitochain/incognito-chain/wire"
	"sync"
	"time"
)

type CrossShardPeerState struct {
	Timestamp int64
	Height    map[byte]uint64 //fromShardID -> hieght
	processed bool
}

type CrossShardSyncProcess struct {
	ShardID             int
	Status              string                    //stop, running
	ShardPeerState      map[string]ShardPeerState //peerid -> state
	ShardPeerStateCh    chan *wire.MessagePeerState
	CrossShardPeerState map[string]CrossShardPeerState //peerID -> state
	Server              Server
	Chain               Chain
	BeaconChain         Chain
	ShardPool           *BlkPool
	ShardSyncProcess    *ShardSyncProcess
	lock                *sync.RWMutex
}

func NewCrossShardSyncProcess(shardID int, shardSyncProcess *ShardSyncProcess, server Server, beaconChain, chain Chain) *CrossShardSyncProcess {
	s := &CrossShardSyncProcess{
		ShardID:          shardID,
		Status:           STOP_SYNC,
		Server:           server,
		Chain:            chain,
		BeaconChain:      beaconChain,
		ShardPool:        NewBlkPool("ShardPool-" + string(shardID)),
		ShardSyncProcess: shardSyncProcess,
	}
	return s
}

func (s *CrossShardSyncProcess) Start() {
	s.Status = RUNNING_SYNC
	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		for {
			if s.Status != RUNNING_SYNC {
				break
			}
			select {
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
				s.syncCrossShardPoolProcess()
			}
		}
	}()

}

func (s *CrossShardSyncProcess) Stop() {
	s.Status = STOP_SYNC
}

func (s *CrossShardSyncProcess) syncCrossShardPoolProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncCrossShardPoolProcess)
	if !s.ShardSyncProcess.FewBlockBehind {
		return
	}
	//TODO : sync CrossShard direct from shard node

	//TODO optional later: sync CrossShard  from other validator pool
	//for peerID, pState := range s.CrossShardPeerState {
	//	for fromSID, height := range pState.Height {
	//		if height <= s.Server.GetCrossShardPool(s.ShardID).GetLatestCrossShardFinalHeight(fromSID) {
	//			continue
	//		}
	//		ch, stop := s.Server.RequestCrossShardBlock(peerID, int(s.ShardID), s.Server.GetCrossShardPool(s.ShardID).GetLatestCrossShardFinalHeight(fromSID))
	//		for {
	//			shouldBreak := false
	//			select {
	//			case block := <-ch:
	//				if err := s.Server.GetCrossShardPool(s.ShardID).AddBlock(block); err != nil {
	//					shouldBreak = true
	//				}
	//			}
	//			if shouldBreak {
	//				stop <- 1
	//				break
	//			}
	//		}
	//	}
	//}

}
