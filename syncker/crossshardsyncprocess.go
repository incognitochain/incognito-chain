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
	time   *time.Time
}

func NewCrossShardSyncProcess(server Server, shardSyncProcess *ShardSyncProcess, beaconChain BeaconChainInterface) *CrossShardSyncProcess {
	s := &CrossShardSyncProcess{
		Status:           STOP_SYNC,
		Server:           server,
		BeaconChain:      beaconChain,
		ShardSyncProcess: shardSyncProcess,
		CrossShardPool:   NewCrossShardBlkPool("crossshard"),
		ShardID:          shardSyncProcess.ShardID,
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
	s.lastRequestCrossShard = s.ShardSyncProcess.Chain.GetCrossShardState()
	s.requestPool = make(map[byte]map[common.Hash]*CrossXReq)

	go func() {
		ticker := time.NewTicker(time.Millisecond * 500)
		for {
			select {
			case f := <-s.actionCh:
				f()
			case <-ticker.C:
				if s.Status != RUNNING_SYNC || !s.ShardSyncProcess.FewBlockBehind {
					time.Sleep(time.Second)
					continue
				}
				s.pullCrossShardBlock() //we need batching 500ms per request
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
				if i == s.ShardID {
					break
				}
				requestHeight := s.lastRequestCrossShard[byte(i)]
				nextHeight := s.Server.FetchNextCrossShard(i, s.ShardID, requestHeight)
				//fmt.Println("crossdebug FetchNextCrossShard", i, s.ShardID, requestHeight, nextHeight)
				if nextHeight == 0 {
					break
				}
				beaconBlock, err := s.Server.FetchBeaconBlockConfirmCrossShardHeight(i, s.ShardID, nextHeight)
				if err != nil {
					break
				}
				//fmt.Println("crossdebug beaconBlock", beaconBlock.Body.ShardState[byte(i)])
				for _, shardState := range beaconBlock.Body.ShardState[byte(i)] {
					//fmt.Println("crossdebug shardState.Height", shardState.Height, nextHeight)
					fromSID := i
					if shardState.Height == nextHeight {
						s.actionCh <- func() {
							reqCnt++
							if s.requestPool[byte(fromSID)] == nil {
								s.requestPool[byte(fromSID)] = make(map[common.Hash]*CrossXReq)
							}
							s.requestPool[byte(fromSID)][shardState.Hash] = &CrossXReq{
								time:   nil,
								height: shardState.Height,
							}
						}
						s.lastRequestCrossShard[byte(fromSID)] = nextHeight
						break
					}
				}
			}
		}

		if reqCnt == 0 {
			time.Sleep(time.Second * 15)
		}
	}
}

func (s *CrossShardSyncProcess) pullCrossShardBlock() {
	//TODO: should limit the number of request block
	currentCrossShardStatus := s.ShardSyncProcess.Chain.GetCrossShardState()
	for fromSID, reqs := range s.requestPool {
		reqHash := []common.Hash{}
		for hash, req := range reqs {
			//if not request or (time out and cross shard not confirm and in pool yet)
			if req.time == nil || (req.time.Add(time.Second*10).Before(time.Now()) && s.CrossShardPool.BlkPoolByHash[hash.String()] == nil && req.height > currentCrossShardStatus[fromSID]) {
				reqHash = append(reqHash, hash)
				t := time.Now()
				reqs[hash].time = &t
			}
		}
		fmt.Println("crossdebug: PushMessageGetBlockCrossShardByHash", fromSID, byte(s.ShardID), reqHash)
		//if err := s.Server.PushMessageGetBlockCrossShardByHash(fromSID, byte(s.ShardID), reqHash, false, ""); err != nil {
		//	fmt.Println("crossdebug: Cannot PushMessageGetBlockCrossShardByHash")
		//}
	}
}
