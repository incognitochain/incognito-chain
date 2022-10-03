package pruner

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incdb"
	"golang.org/x/sync/semaphore"
)

type Config struct {
	ShouldPruneByHash bool `json:"ShouldPruneByHash"`
}

type PrunerManager struct {
	ShardPruner map[int]*ShardPruner
	JobRquest   map[int]*Config
}

func NewPrunerManager(db map[int]incdb.Database) *PrunerManager {
	prunerManager := &PrunerManager{
		ShardPruner: make(map[int]*ShardPruner),
		JobRquest:   make(map[int]*Config),
	}
	for sid := 0; sid < common.MaxShardNumber; sid++ {
		prunerManager.ShardPruner[sid] = NewShardPruner(sid, db[sid])
	}

	return prunerManager
}

func (s *PrunerManager) Start() error {
	for {
		for sid, shardPruner := range s.ShardPruner {
			//if shard pruner not run -> then check condition to trigger prune
			if shardPruner.status == IDLE {
				latest := false
				if shardPruner.bestView != nil && shardPruner.bestView.CalculateTimeSlot(shardPruner.bestView.BestBlock.GetProposeTime()) == shardPruner.bestView.CalculateTimeSlot(time.Now().Unix()) {
					latest = true
				}
				//if auto prune, sync latest block, and bestview block > last processing block
				if config.Config().EnableAutoPrune && latest && shardPruner.bestView.ShardHeight > shardPruner.lastProcessingHeight+config.Config().NumBlockTriggerPrune {
					shardPruner.SetBloomSize(config.Config().StateBloomSize)
					shardPruner.Prune(false)
				} else if req, ok := s.JobRquest[sid]; ok { //request for this shard from RPC
					shardPruner.SetBloomSize(config.Config().StateBloomSize)
					if req.ShouldPruneByHash {
						shardPruner.Prune(true)
					} else {
						shardPruner.Prune(false)
					}
					delete(s.JobRquest, sid)
					//s.JobRquest[sid] = nil //unmark request for this shard
				}
			}
		}
		time.Sleep(time.Second)
	}
}

// run parallel based on available CPU
func (s *PrunerManager) OfflinePrune() {
	cpus := runtime.NumCPU()
	semNum := cpus / 2
	if semNum > 2 {
		semNum = 2
	}
	sem := semaphore.NewWeighted(int64(semNum))
	ch := make(chan int)

	stateBloomSize := config.Config().StateBloomSize / uint64(semNum*2)

	stopCh := make(chan struct{})
	var count int
	var wg sync.WaitGroup
	go func() {
		for {
			select {
			case shardID := <-ch:
				wg.Add(1)
				go func() {
					if _, ok := s.ShardPruner[shardID]; !ok {
						fmt.Println("ShardPrunter is not ready")
					}
					s.ShardPruner[shardID].SetBloomSize(stateBloomSize)
					s.ShardPruner[shardID].Prune(false)
					wg.Done()
					sem.Release(1)
					Logger.log.Infof("Shard %v finish prune", shardID)
					b, _ := json.MarshalIndent(s.ShardPruner[shardID].Report(), "", "\t")
					fmt.Println(string(b))
				}()
			case <-stopCh:
				count++
				if count == common.MaxShardNumber {
					return
				}
			}
		}
	}()
	for i := 0; i < common.MaxShardNumber; i++ {
		sem.Acquire(context.Background(), 1)
		ch <- i
	}
	wg.Wait()
}

func (p *PrunerManager) SetShardInsertLock(sid int, mutex *sync.Mutex) {
	p.ShardPruner[sid].shardInsertLock = mutex
}

func (p *PrunerManager) InsertNewView(shardBestState *blockchain.ShardBestState) {
	sid := shardBestState.ShardID
	p.ShardPruner[int(sid)].handleNewView(shardBestState)
}

func (s *PrunerManager) Report() map[int]ShardPrunerReport {
	res := map[int]ShardPrunerReport{}
	for sid := 0; sid < common.MaxShardNumber; sid++ {
		res[sid] = s.ShardPruner[sid].Report()
	}
	return res
}
