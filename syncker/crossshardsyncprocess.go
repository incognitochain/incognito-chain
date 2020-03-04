package syncker

import (
	"context"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"time"
)

type CrossShardSyncProcess struct {
	status                string //stop, running
	server                Server
	shardID               int
	shardSyncProcess      *ShardSyncProcess
	beaconChain           BeaconChainInterface
	crossShardPool        *CrossShardBlkPool
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
		status:           STOP_SYNC,
		server:           server,
		beaconChain:      beaconChain,
		shardSyncProcess: shardSyncProcess,
		crossShardPool:   NewCrossShardBlkPool("crossshard"),
		shardID:          shardSyncProcess.ShardID,
		actionCh:         make(chan func()),
	}

	go s.syncCrossShard()
	go s.pullCrossShardBlock()
	return s
}

func (s *CrossShardSyncProcess) start() {
	if s.status == RUNNING_SYNC {
		return
	}
	s.status = RUNNING_SYNC
	s.lastRequestCrossShard = s.shardSyncProcess.Chain.GetCrossShardState()
	s.requestPool = make(map[byte]map[common.Hash]*CrossXReq)
	go func() {
		for {
			f := <-s.actionCh
			f()
		}
	}()
}

func (s *CrossShardSyncProcess) stop() {
	s.status = STOP_SYNC
}

func (s *CrossShardSyncProcess) getRequestPool() map[byte]map[common.Hash]*CrossXReq {
	res := make(chan map[byte]map[common.Hash]*CrossXReq)
	s.actionCh <- func() {
		pool := make(map[byte]map[common.Hash]*CrossXReq)
		for k, v := range s.requestPool {
			for i, j := range v {
				if pool[k] == nil {
					pool[k] = make(map[common.Hash]*CrossXReq)
				}
				pool[k][i] = j
			}
		}
		res <- pool
	}
	return <-res
}

func (s *CrossShardSyncProcess) setRequestPool(fromSID int, hash common.Hash, crossReq *CrossXReq) {
	res := make(chan int)
	s.actionCh <- func() {
		if s.requestPool[byte(fromSID)] == nil {
			s.requestPool[byte(fromSID)] = make(map[common.Hash]*CrossXReq)
		}
		s.requestPool[byte(fromSID)][hash] = crossReq
		res <- 1
	}
	<-res
}

func (s *CrossShardSyncProcess) syncCrossShard() {
	for {
		reqCnt := 0
		if s.status != RUNNING_SYNC || !s.shardSyncProcess.FewBlockBehind {
			time.Sleep(time.Second * 5)
			continue
		}
		//get last confirm crossshard -> process request until retrieve info
		for i := 0; i < s.server.GetChainParam().ActiveShards; i++ {
			for {
				if i == s.shardID {
					break
				}
				requestHeight := s.lastRequestCrossShard[byte(i)]
				nextHeight := s.server.FetchNextCrossShard(i, s.shardID, requestHeight)
				//fmt.Println("crossdebug FetchNextCrossShard", i, s.ShardID, requestHeight, nextHeight)
				if nextHeight == 0 {
					break
				}
				beaconBlock, err := s.server.FetchBeaconBlockConfirmCrossShardHeight(i, s.shardID, nextHeight)
				if err != nil {
					break
				}
				//fmt.Println("crossdebug beaconBlock", beaconBlock.Body.ShardState[byte(i)])
				for _, shardState := range beaconBlock.Body.ShardState[byte(i)] {
					//fmt.Println("crossdebug shardState.Height", shardState.Height, nextHeight)
					fromSID := i
					if shardState.Height == nextHeight {
						reqCnt++
						s.setRequestPool(fromSID, shardState.Hash, &CrossXReq{time: nil, height: shardState.Height})
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
	defer time.AfterFunc(time.Second*1, s.pullCrossShardBlock)

	currentCrossShardStatus := s.shardSyncProcess.Chain.GetCrossShardState()
	for fromSID, reqs := range s.getRequestPool() {
		reqHash := []common.Hash{}
		reqHeight := []uint64{}
		for hash, req := range reqs {
			//if not request or (time out and cross shard not confirm and in pool yet)
			if req.height > currentCrossShardStatus[fromSID] && !s.crossShardPool.HasBlock(hash) && (req.time == nil || (req.time.Add(time.Second * 10).Before(time.Now()))) {
				reqHash = append(reqHash, hash)
				reqHeight = append(reqHeight, req.height)
				t := time.Now()
				reqs[hash].time = &t
			}
		}
		if len(reqHash) > 0 {
			//fmt.Println("crossdebug: PushMessageGetBlockCrossShardByHash", fromSID, byte(s.ShardID), reqHeight)
			s.streamCrossBlkFromPeer(int(fromSID), reqHeight)
		}

	}
}

func (s *CrossShardSyncProcess) streamCrossBlkFromPeer(fromSID int, height []uint64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	//stream
	ch, err := s.server.RequestCrossShardBlocksViaStream(ctx, "", fromSID, s.shardID, height)
	if err != nil {
		fmt.Println("Syncker: create channel fail")
		return
	}

	//receive
	blkCnt := int(0)
	for {
		blkCnt++
		select {
		case blk := <-ch:
			if !isNil(blk) {
				fmt.Println("syncker: Insert crossShard block", blk.GetHeight(), blk.Hash().String())
				s.crossShardPool.AddBlock(blk.(common.CrossShardBlkPoolInterface))
			} else {
				break
			}
		}
		if blkCnt > 100 {
			break
		}
	}
}
