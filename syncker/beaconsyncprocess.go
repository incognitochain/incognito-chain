package syncker

import "time"

type BeaconPeerState struct {
	PeerID          string
	FinalViewHeight uint64
	BestViewHash    string
	BestViewHeight  uint64
}
type S2BPeerState struct {
	PeerID string
	Height map[byte]uint64
}

type BeaconSyncProcess struct {
	Status         string //stop, running
	ShardPeerState []BeaconPeerState
	S2BPeerState   []S2BPeerState
	Server         Server
	Chain          Chain
}

func NewBeaconSyncProcess(shardID byte, server Server, chain Chain) *BeaconSyncProcess {
	s := &BeaconSyncProcess{
		Status: STOP_SYNC,
		Server: server,
		Chain:  chain,
	}

	go s.syncBeaconProcess()
	go s.syncS2BPoolProcess()
	return s
}

func (s *BeaconSyncProcess) Start() {
	s.Status = RUNNING_SYNC
}

func (s *BeaconSyncProcess) Stop() {
	s.Status = STOP_SYNC
}

func (s *BeaconSyncProcess) syncBeaconProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncBeaconProcess)
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

		ch, stop := s.Server.RequestBlock(pState.PeerID, -1, s.Chain.GetFinalView().GetHeight(), s.Chain.GetBestView().GetHash())
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

func (s *BeaconSyncProcess) syncS2BPoolProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncS2BPoolProcess)
	if s.Status != RUNNING_SYNC {
		return
	}
	for _, pState := range s.S2BPeerState {
		for fromSID, height := range pState.Height {
			if height <= s.Server.GetS2BPool(fromSID).GetLatestFinalHeight() {
				continue
			}
			ch, stop := s.Server.RequestS2BBlockPool(pState.PeerID, int(fromSID), s.Server.GetS2BPool(fromSID).GetLatestFinalHeight())
			for {
				shouldBreak := false
				select {
				case block := <-ch:
					if err := s.Server.GetS2BPool(fromSID).AddBlock(block); err != nil {
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

}
