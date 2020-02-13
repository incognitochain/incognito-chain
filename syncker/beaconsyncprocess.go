package syncker

import (
	"log"
	"time"
)

type BeaconPeerState struct {
	Timestamp      int64
	BestViewHash   string
	BestViewHeight uint64
}
type S2BPeerState struct {
	Timestamp int64
	Height    map[byte]uint64 //shardid -> height
}

type BeaconSyncProcess struct {
	Status           string //stop, running
	IsCommittee      bool
	BeaconPeerStates map[string]BeaconPeerState //sender -> state
	S2BPeerState     map[string]S2BPeerState    //sender -> state
	Server           Server
	Chain            Chain
}

func NewBeaconSyncProcess(server Server, chain Chain) *BeaconSyncProcess {
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

	for peerID, pState := range s.BeaconPeerStates {

		if pState.BestViewHeight < s.Chain.GetBestViewHeight() {
			continue
		}
		if pState.BestViewHeight == s.Chain.GetBestViewHeight() && pState.BestViewHash == s.Chain.GetBestViewHash() {
			continue
		}
		log.Printf("SYNCKER Request Block from %s height %d hash %s", peerID, s.Chain.GetFinalViewHeight(), s.Chain.GetBestViewHash())
		//ch, stop := s.Server.RequestBlock(peerID, -1, s.Chain.GetFinalViewHash(), s.Chain.GetBestViewHash(), pState.BestViewHash)
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

func (s *BeaconSyncProcess) syncS2BPoolProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncS2BPoolProcess)
	if s.Status != RUNNING_SYNC || !s.IsCommittee {
		return
	}
	for peerID, pState := range s.S2BPeerState {
		for fromSID, height := range pState.Height {
			if height <= s.Server.GetS2BPool(fromSID).GetLatestFinalHeight() {
				continue
			}
			ch, stop := s.Server.RequestS2BBlockPool(peerID, int(fromSID), s.Server.GetS2BPool(fromSID).GetLatestFinalHeight())
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
