package syncker

import "time"

const RUNNING_SYNC = "running_sync"
const STOP_SYNC = "stop_sync"

type ShardPeerState struct {
	PeerID          string
	FinalViewHeight uint64
	BestViewHash    string
	BestViewHeight  uint64
}
type CrossShardPeerState struct {
	PeerID string
	Height uint64
}

type ShardSyncProcess struct {
	ShardID             byte
	Status              string //stop, running
	ShardPeerState      []ShardPeerState
	CrossShardPeerState []CrossShardPeerState
	Server              Server
	Chain               Chain
}

func NewShardSyncProcess(shardID byte, server Server, chain Chain) *ShardSyncProcess {
	s := &ShardSyncProcess{
		ShardID: shardID,
		Status:  STOP_SYNC,
		Server:  server,
		Chain:   chain,
	}

	go s.broadcastPeerStateProcess()
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

func (s *ShardSyncProcess) broadcastPeerStateProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.broadcastPeerStateProcess)
	//TODO: create peerstate info and broadcast
}

func (s *ShardSyncProcess) syncShardProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncShardProcess)
	if s.Status != RUNNING_SYNC {
		return
	}
	for _, pState := range s.ShardPeerState {
		if pState.BestViewHeight < s.Chain.GetBestView().GetHeight() {
			continue
		}
		if pState.BestViewHeight == s.Chain.GetBestView().GetHeight() && pState.BestViewHash == s.Chain.GetBestView().GetHash() {
			continue
		}

		ch, stop := s.Server.RequestBlock(pState.PeerID, int(s.ShardID), s.Chain.GetFinalView().GetHeight(), s.Chain.GetBestView().GetHash())
		for {
			shouldBreak := false
			select {
			case block := <-ch:
				if err := s.Chain.InsertBlock(block); err != nil {
					shouldBreak = true
				}
			}
			if shouldBreak {
				stop <- 1
				break
			}
		}

	}
}

func (s *ShardSyncProcess) syncCrossShardPoolProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncCrossShardPoolProcess)
	if s.Status != RUNNING_SYNC {
		return
	}
	for _, pState := range s.CrossShardPeerState {
		if pState.Height <= s.Server.GetCrossShardPool(s.ShardID).GetLatestFinalHeight() {
			continue
		}

		ch, stop := s.Server.RequestCrossShardBlockPool(pState.PeerID, int(s.ShardID), s.Server.GetCrossShardPool(s.ShardID).GetLatestFinalHeight())
		for {
			shouldBreak := false
			select {
			case block := <-ch:
				if err := s.Server.GetCrossShardPool(s.ShardID).AddBlock(block); err != nil {
					shouldBreak = true
				}
			}
			if shouldBreak {
				stop <- 1
				break
			}
		}
	}

}
