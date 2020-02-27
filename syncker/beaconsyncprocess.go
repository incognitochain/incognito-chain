package syncker

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"reflect"
	"sync"
	"time"
)

type BeaconPeerState struct {
	Timestamp      int64
	BestViewHash   string
	BestViewHeight uint64
	processed      bool
}

type S2BPeerState struct {
	Timestamp int64
	Height    map[byte]uint64 //shardid -> height
	processed bool
}

type BeaconSyncProcess struct {
	Status           string //stop, running
	IsCommittee      bool
	FewBlockBehind   bool
	BeaconPeerStates map[string]BeaconPeerState //sender -> state
	S2BPeerState     map[string]S2BPeerState    //sender -> state
	Server           Server
	Chain            Chain
	ChainID          int

	BeaconPool *BlkPool
	lock       *sync.RWMutex
}

func NewBeaconSyncProcess(server Server, chain Chain) *BeaconSyncProcess {
	s := &BeaconSyncProcess{
		Status:     STOP_SYNC,
		Server:     server,
		Chain:      chain,
		BeaconPool: NewBlkPool("BeaconPool"),
		lock:       new(sync.RWMutex),
	}

	go s.syncBeaconProcess()
	go s.insertBeaconBlockFromPool()
	go s.syncS2BPoolProcess()
	return s
}

func (s *BeaconSyncProcess) Start(chainID int) {
	s.ChainID = chainID
	s.Status = RUNNING_SYNC
}

func (s *BeaconSyncProcess) Stop() {
	s.Status = STOP_SYNC
}

func isNil(v interface{}) bool {
	return v == nil || (reflect.ValueOf(v).Kind() == reflect.Ptr && reflect.ValueOf(v).IsNil())
}

func (s *BeaconSyncProcess) insertBeaconBlockFromPool() {
	defer func() {
		if s.FewBlockBehind {
			time.AfterFunc(time.Millisecond*100, s.insertBeaconBlockFromPool)
		} else {
			time.AfterFunc(time.Second*1, s.insertBeaconBlockFromPool)
		}
	}()

	if !s.FewBlockBehind {
		return
	}
	var blk common.BlockPoolInterface
	if s.ChainID == -1 {
		blk = s.BeaconPool.GetNextBlock(s.Chain.GetBestViewHash(), true)
	} else {
		blk = s.BeaconPool.GetNextBlock(s.Chain.GetBestViewHash(), false)
	}

	if isNil(blk) {
		return
	}
	//fmt.Println("Syncker: Insert beacon from pool", blk.(common.BlockInterface).GetHeight())
	s.BeaconPool.RemoveBlock(blk.GetHash())
	if err := s.Chain.ValidateBlockSignatures(blk.(common.BlockInterface), s.Chain.GetCommittee()); err != nil {
		return
	}

	if err := s.Chain.InsertBlk(blk.(common.BlockInterface)); err != nil {
	}
}

func (s *BeaconSyncProcess) syncBeaconProcess() {
	s.Chain.SetReady(s.FewBlockBehind)
	requestCnt := 0

	defer func() {
		if requestCnt > 0 {
			s.FewBlockBehind = false
			time.AfterFunc(0, s.syncBeaconProcess)
		} else {
			if len(s.BeaconPeerStates) > 0 {
				s.FewBlockBehind = true
			}
			time.AfterFunc(time.Second*1, s.syncBeaconProcess)
		}
	}()

	if s.Status != RUNNING_SYNC {
		return
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	for peerID, pState := range s.BeaconPeerStates {

		toHeight := pState.BestViewHeight
		if s.ChainID != -1 { //if not beacon committee, not insert the newest block (incase we need to revert beacon block)
			toHeight -= 1
		}

		if toHeight <= s.Chain.GetBestViewHeight() {
			continue
		}

		blockBuffer := []common.BlockInterface{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		ch, err := s.Server.RequestBlocksViaStream(ctx, peerID, -1, proto.BlkType_BlkBc, s.Chain.GetBestViewHeight()+1, s.Chain.GetFinalViewHeight(), toHeight, pState.BestViewHash)
		if err != nil {
			fmt.Println("Syncker: create channel fail")
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
					insertBlkCnt := 0
					for {
						//time1 := time.Now()
						if successBlk, err := InsertBatchBlock(s.Chain, blockBuffer); err != nil {
							//fmt.Printf("Syncker Insert %d beacon (from %d to %d) elaspse %f \n", successBlk, blockBuffer[0].GetHeight(), blockBuffer[successBlk-1].GetHeight(), time.Since(time1).Seconds())
							goto CANCEL_REQUEST
						} else {
							insertBlkCnt += successBlk
							//fmt.Printf("Syncker Insert %d beacon (from %d to %d) elaspse %f \n", successBlk, blockBuffer[0].GetHeight(), blockBuffer[len(blockBuffer)-1].GetHeight(), time.Since(time1).Seconds())
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
					//fmt.Println("Syncker: blk nil")
					goto CANCEL_REQUEST
				}

			}
		}

	CANCEL_REQUEST:
		//fmt.Println("Syncker: request cancel")
		return
	}
}

func (s *BeaconSyncProcess) syncS2BPoolProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncS2BPoolProcess)
	if s.Status != RUNNING_SYNC || !s.IsCommittee || !s.FewBlockBehind {
		return
	}

	s.lock.RLock()
	defer s.lock.RUnlock()
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
