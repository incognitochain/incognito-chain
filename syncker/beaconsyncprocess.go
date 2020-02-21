package syncker

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"reflect"
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
	RemainOneBlock   bool
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

func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

func (s *BeaconSyncProcess) syncBeaconProcess() {
	requestCnt := 0
	defer func() {
		if requestCnt > 0 {
			time.AfterFunc(0, s.syncBeaconProcess)
		} else {
			time.AfterFunc(time.Millisecond*500, s.syncBeaconProcess)
		}
	}()

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

		blockBuffer := []common.BlockInterface{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		ch, err := s.Server.RequestBlocksViaChannel(ctx, peerID, -1, s.Chain.GetBestViewHeight()+1, s.Chain.GetFinalViewHeight(), pState.BestViewHeight, pState.BestViewHash)
		if err != nil {
			continue
		}

		requestCnt++
		insertTime := time.Now()
		for {
			select {
			case blk := <-ch:
				if !isNil(blk) {
					blockBuffer = append(blockBuffer, blk)
				}

				if len(blockBuffer) >= 350 || (len(blockBuffer) > 0 && (isNil(blk) || time.Since(insertTime) > time.Millisecond*1000)) {
					//if err := s.Chain.InsertBatchBlock(blockBuffer); err != nil {
					//	goto CANCEL_REQUEST
					//}
					fmt.Println("SYNCKER insert", len(blockBuffer))
					time.Sleep(2000 * time.Millisecond)
					insertTime = time.Now()
					blockBuffer = []common.BlockInterface{}
				}

				if isNil(blk) && len(blockBuffer) == 0 {
					goto CANCEL_REQUEST
				}

			}
		}

	CANCEL_REQUEST:
		return
	}
}

func (s *BeaconSyncProcess) syncS2BPoolProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncS2BPoolProcess)
	if s.Status != RUNNING_SYNC || !s.IsCommittee || !s.RemainOneBlock {
		return
	}
	//sync when status is enable and in committee and remain only one syncing beacon block
	//TODO : sync S2B direct from shard node

	//TODO optional later : sync S2B from other validator pool
	//for peerID, pState := range s.S2BPeerState {
	//	for fromSID, height := range pState.Height {
	//		if height <= s.Server.GetS2BPool(fromSID).GetLatestFinalHeight() {
	//			continue
	//		}
	//		ch, stop := s.Server.RequestS2BBlock(peerID, int(fromSID), s.Server.GetS2BPool(fromSID).GetLatestFinalHeight())
	//		for {
	//			shouldBreak := false
	//			select {
	//			case block := <-ch:
	//				if err := s.Server.GetS2BPool(fromSID).AddBlock(block); err != nil {
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
