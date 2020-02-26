package syncker

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/peerv2/proto"
	"sync"
	"time"
)

const RUNNING_SYNC = "running_sync"
const STOP_SYNC = "stop_sync"

type ShardPeerState struct {
	Timestamp      int64
	BestViewHash   string
	BestViewHeight uint64
	processed      bool
}
type CrossShardPeerState struct {
	Timestamp int64
	Height    map[byte]uint64 //fromShardID -> hieght
	processed bool
}

type ShardSyncProcess struct {
	IsCommittee         bool
	FewBlockBehind      bool
	ShardID             int
	Status              string                         //stop, running
	ShardPeerState      map[string]ShardPeerState      //peerid -> state
	CrossShardPeerState map[string]CrossShardPeerState //peerID -> state
	Server              Server
	Chain               Chain
	lock                *sync.RWMutex
}

func NewShardSyncProcess(shardID int, server Server, chain Chain) *ShardSyncProcess {
	s := &ShardSyncProcess{
		ShardID: shardID,
		Status:  STOP_SYNC,
		Server:  server,
		Chain:   chain,
		lock:    new(sync.RWMutex),
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

	s.Chain.SetReady(s.FewBlockBehind)

	requestCnt := 0
	defer func() {
		if requestCnt > 0 {
			s.FewBlockBehind = false
			time.AfterFunc(0, s.syncShardProcess)
		} else {
			if len(s.ShardPeerState) > 0 {
				s.FewBlockBehind = true
			}
			time.AfterFunc(time.Second*10, s.syncShardProcess)
		}
	}()

	if s.Status != RUNNING_SYNC {
		return
	}

	s.lock.RLock()
	defer s.lock.RUnlock()

	for peerID, pState := range s.ShardPeerState {
		if pState.processed {
			continue
		}

		if pState.BestViewHeight <= s.Chain.GetBestViewHeight() {
			continue
		}

		blockBuffer := []common.BlockInterface{}
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		fmt.Println("SYNCKER Request Shard Block", peerID, s.ShardID, s.Chain.GetBestViewHeight()+1, pState.BestViewHeight, pState.BestViewHash, s.Chain.GetBestViewHash())

		ch, err := s.Server.RequestBlocksViaStream(ctx, peerID, s.ShardID, proto.BlkType_BlkShard, s.Chain.GetBestViewHeight()+1, s.Chain.GetFinalViewHeight(), pState.BestViewHeight, pState.BestViewHash)
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
						time1 := time.Now()
						if successBlk, err := InsertBatchBlock(s.Chain, blockBuffer); err != nil {
							fmt.Println("Syncker:", err)
							goto CANCEL_REQUEST
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
					fmt.Println("Syncker: blk nil")
					goto CANCEL_REQUEST
				}

			}
		}

	CANCEL_REQUEST:
		fmt.Println("Syncker: request cancel")
		pState.processed = true
		return
	}
}

func (s *ShardSyncProcess) syncCrossShardPoolProcess() {
	defer time.AfterFunc(time.Millisecond*500, s.syncCrossShardPoolProcess)
	if s.Status != RUNNING_SYNC || !s.IsCommittee || !s.FewBlockBehind {
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
