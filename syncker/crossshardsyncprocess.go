package syncker

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"time"
)

type CrossShardSyncProcess struct {
	Status                string //stop, running
	Server                Server
	ShardID               int
	ShardSyncProcess      *ShardSyncProcess
	BeaconChain           BeaconChainInterface
	CrossShardPool        *CrossShardBlkPool
	lastRequestCrossShard map[byte]uint64
	requestPool           map[byte]map[common.Hash]*CrossXReq
	actionCh              chan func()
}

type CrossXReq struct {
	height uint64
	time   time.Time
}

func NewCrossShardSyncProcess(server Server, shardSyncProcess *ShardSyncProcess, beaconChain BeaconChainInterface) *CrossShardSyncProcess {
	s := &CrossShardSyncProcess{
		Status:           STOP_SYNC,
		Server:           server,
		BeaconChain:      beaconChain,
		ShardSyncProcess: shardSyncProcess,
		actionCh:         make(chan func()),
	}

	go s.syncCrossShard()
	return s
}

func (s *CrossShardSyncProcess) Start() {
	if s.Status == RUNNING_SYNC {
		return
	}
	s.Status = RUNNING_SYNC
	s.lastRequestCrossShard = s.BeaconChain.GetCurrentCrossShardHeightToShard(byte(s.ShardID))
	s.requestPool = make(map[byte]map[common.Hash]*CrossXReq)

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		for {
			if s.Status != RUNNING_SYNC || !s.ShardSyncProcess.FewBlockBehind {
				time.Sleep(time.Second)
				continue
			}

			select {
			case <-ticker.C:
				s.pullCrossShardBlock() //we need batching 500ms per request
			case f := <-s.actionCh:
				f()
			}
		}
	}()
}

func (s *CrossShardSyncProcess) Stop() {
	s.Status = STOP_SYNC
}

func (s *CrossShardSyncProcess) syncCrossShard() {
	for {
		reqCnt := 0
		if s.Status != RUNNING_SYNC || !s.ShardSyncProcess.FewBlockBehind {
			time.Sleep(time.Second * 5)
			continue
		}

		//get last confirm crossshard -> process request until retrieve info
		for i := 0; i < s.Server.GetChainParam().ActiveShards; i++ {
			for {
				requestHeight := s.lastRequestCrossShard[byte(i)]
				nextHeight := s.Server.FetchNextCrossShard(i, s.ShardID, requestHeight)
				if nextHeight == 0 {
					break
				}
				beaconBlock, err := s.Server.FetchBeaconBlockConfirmCrossShardHeight(i, s.ShardID, nextHeight)
				if err != nil {
					break
				}

				for _, shardState := range beaconBlock.Body.ShardState[byte(i)] {
					if shardState.Height == nextHeight {
						s.actionCh <- func() {
							reqCnt++
							if s.requestPool[byte(i)] == nil {
								s.requestPool[byte(i)] = make(map[common.Hash]*CrossXReq)
							}
							s.requestPool[byte(i)][shardState.Hash] = nil
						}
						s.lastRequestCrossShard[byte(i)] = nextHeight
						break
					}
				}
			}
		}

		if reqCnt == 0 {
			time.Sleep(time.Second * 5)
		}
	}
}

func (s *CrossShardSyncProcess) pullCrossShardBlock() {
	currentCrossShardStatus := s.BeaconChain.GetCurrentCrossShardHeightToShard(byte(s.ShardID))
	for fromSID, reqs := range s.requestPool {
		reqHash := []common.Hash{}
		for hash, reqTime := range reqs {
			//if not request or (time out and cross shard not confirm and in pool yet)
			if reqTime == nil || (reqTime.time.Add(time.Second*10).Before(time.Now()) && s.CrossShardPool.BlkPoolByHash[hash.String()] == nil && reqTime.height > currentCrossShardStatus[fromSID]) {
				reqHash = append(reqHash, hash)
			}
		}
		if err := s.Server.PushMessageGetBlockCrossShardByHash(fromSID, byte(s.ShardID), reqHash, false, ""); err != nil {
			fmt.Println("syncker: Cannot PushMessageGetBlockCrossShardByHash")
		}

	}
}
