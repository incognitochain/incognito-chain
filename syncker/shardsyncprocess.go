package syncker

import "time"

const RUNNING_SYNC = "running_sync"
const STOP_SYNC = "stop_sync"

type ShardPeerState struct {
	Timestamp      int64
	BestViewHash   string
	BestViewHeight uint64
}
type CrossShardPeerState struct {
	Timestamp int64
	Height    map[byte]uint64 //fromShardID -> hieght
}

type ShardSyncProcess struct {
	IsCommittee         bool
	RemainOneBlock      bool
	ShardID             int
	Status              string                         //stop, running
	ShardPeerState      map[string]ShardPeerState      //peerid -> state
	CrossShardPeerState map[string]CrossShardPeerState //peerID -> state
	Server              Server
	Chain               Chain
}

func NewShardSyncProcess(shardID int, server Server, chain Chain) *ShardSyncProcess {
	s := &ShardSyncProcess{
		ShardID: shardID,
		Status:  STOP_SYNC,
		Server:  server,
		Chain:   chain,
	}

	go s.syncShardProcess()
	go s.syncCrossShardPoolProcess()
	return s
}

func (s *ShardSyncProcess) Start() {
	s.Status = RUNNING_SYNC
}

func (s *ShardSyncProcess) Stop() {
	s.Status = STOP_SYNC
}

func (s *ShardSyncProcess) syncShardProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncShardProcess)
	if s.Status != RUNNING_SYNC {
		return
	}
	for _, pState := range s.ShardPeerState {
		if pState.BestViewHeight < s.Chain.GetBestViewHeight() {
			continue
		}
		if pState.BestViewHeight == s.Chain.GetBestViewHeight() && pState.BestViewHash == s.Chain.GetBestViewHash() {
			continue
		}

		//ch, stop := s.Server.RequestBlocksViaStream(PeerID, int(s.ShardID), s.Chain.GetFinalViewHash(), s.Chain.GetBestViewHash(), pState.BestViewHash)
		//for {
		//	shouldBreak := false
		//	select {
		//	case _ = <-ch:
		//		//if err := s.Chain.InsertBlock(block); err != nil {
		//		//	shouldBreak = true
		//		//}
		//	}
		//	if shouldBreak {
		//		stop <- 1
		//		break
		//	}
		//}
	}
}

func (s *ShardSyncProcess) syncCrossShardPoolProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncCrossShardPoolProcess)
	if s.Status != RUNNING_SYNC || !s.IsCommittee || !s.RemainOneBlock {
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
