package pruner

import (
	"context"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/peerv2"
	"github.com/incognitochain/incognito-chain/pubsub"
	"golang.org/x/sync/semaphore"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

type PrunerManager struct {
	ShardPruner   map[int]*ShardPruner
	PubSubManager *pubsub.PubSubManager
	consensus     peerv2.ConsensusData
}

func NewPrunerManager() *PrunerManager {
	prunerManager := &PrunerManager{
		ShardPruner: make(map[int]*ShardPruner),
	}

	return prunerManager
}

const NUM_BLOCK_TRIGGER_PRUNE = 10

func (s *PrunerManager) Init() error {
	cfg := config.LoadConfig()
	db, err := incdb.OpenMultipleDB("leveldb", filepath.Join(cfg.DataDir, cfg.DatabaseDir))
	if err != nil {
		return err
	}
	for sid := 0; sid < common.MaxShardNumber; sid++ {
		s.ShardPruner[sid] = NewShardPruner(sid, db[sid])
	}

	go func() {
		for {
			for _, shardPruner := range s.ShardPruner {
				//if shard pruner not run -> then check condition to trigger prune
				if shardPruner.status != RUNNING {
					latest := false
					if shardPruner.bestView != nil && common.CalculateTimeSlot(shardPruner.bestView.BestBlock.GetProposeTime()) == common.CalculateTimeSlot(time.Now().Unix()) {
						latest = true
					}
					//if auto prune, latest block, and bestview block > last processing block
					if config.Config().EnableAutoPrune && latest && shardPruner.bestView.ShardHeight > shardPruner.lastProcessingHeight+NUM_BLOCK_TRIGGER_PRUNE {
						//if not committee or forceprunce => trigger prune
						if shardPruner.role != common.CommitteeRole || config.Config().ForcePrune {
							shardPruner.PruneByHeight()
						}
					}
				}
			}
			time.Sleep(time.Second)
			for sid, val := range s.consensus.GetOneValidatorForEachConsensusProcess() {
				s.ShardPruner[sid].SetCommitteeRole(val.State.Role)
			}
		}
	}()
	return nil
}

func (s *PrunerManager) OfflinePrune() {
	cpus := runtime.NumCPU()
	sem := semaphore.NewWeighted(int64(cpus / 2))
	ch := make(chan int)
	stateBloomSize := config.Config().StateBloomSize / uint64(common.MaxShardNumber)
	if cpus/2 < common.MaxShardNumber {
		stateBloomSize = config.Config().StateBloomSize / uint64(cpus/2)
	}
	stopCh := make(chan struct{})
	var count int
	var wg sync.WaitGroup
	go func() {
		for {
			select {
			case shardID := <-ch:
				wg.Add(1)
				go func() {
					s.ShardPruner[shardID].SetBloomSize(stateBloomSize)
					if err := s.ShardPruner[shardID].PruneByHeight(); err != nil {
						Logger.log.Error(err)
						return
					}
					wg.Done()
					sem.Release(1)
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

func (s *PrunerManager) Report() error {
	for sid := 0; sid < common.MaxShardNumber; sid++ {
		s.ShardPruner[sid].Report()
	}
	return nil
}

func (p *PrunerManager) SetShardInsertLock(sid int, mutex *sync.Mutex) {
	p.ShardPruner[sid].shardInsertLock = mutex
}

func (p *PrunerManager) InsertNewView(shardBestState *blockchain.ShardBestState) {
	sid := shardBestState.ShardID
	p.ShardPruner[int(sid)].handleNewView(shardBestState)
}
