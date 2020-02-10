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
	Height string
}

type ShardSyncProcess struct {
	ShardID             int
	Status              string //stop, running
	ShardPeerState      []ShardPeerState
	CrossShardPeerState []CrossShardPeerState
	Server              Server
	Chain               Chain
}

func NewShardSyncProcess(shardID byte, server Server, chain Chain) *ShardSyncProcess {
	s := &ShardSyncProcess{
		ShardID: int(shardID),
		Status:  STOP_SYNC,
		Server:  server,
		Chain:   chain,
	}

	go s.broadcastPeerStateProcess()
	go s.syncShardProcess()
	go s.syncCrossShardProcess()
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
		if pState.BestViewHeight < s.Chain.GetShardBestView().GetHeight() {
			continue
		}
		if pState.BestViewHeight == s.Chain.GetShardBestView().GetHeight() && pState.BestViewHash == s.Chain.GetShardBestView().GetHash() {
			continue
		}

		ch := s.Server.RequestBlock(pState.PeerID, s.ShardID, s.Chain.GetShardFinalView().GetHeight(), s.Chain.GetShardBestView().GetHash(), -1)
		select {
		case block := <-ch:
			s.Chain.InsertBlock(block)
		}
	}
}

func (s *ShardSyncProcess) syncCrossShardProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncCrossShardProcess)
	if s.Status != RUNNING_SYNC {
		return
	}
	//TODO: Sync Cross Shard Block from PeerState

}
