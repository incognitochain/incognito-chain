package pruner

import (
	"encoding/json"
	"runtime"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

type Pruner struct {
	db map[int]incdb.Database
}

func NewPrunerWithValue(db map[int]incdb.Database) *Pruner {
	return &Pruner{
		db: db,
	}
}

func (p *Pruner) PruneImmediately() error {
	for i := 0; i < common.MaxShardNumber; i++ {
		if err := p.Prune(i); err != nil {
			panic(err)
		}
	}
	return nil
}

func (p *Pruner) Prune(sID int) error {
	stateBloom, err := trie.NewStateBloomWithSize(2048)
	if err != nil {
		panic(err)
	}
	shardID := byte(sID)
	db := p.db[int(shardID)]
	rootHashCache, err := lru.New(100)
	if err != nil {
		panic(err)
	}
	Logger.log.Infof("[state-prune] Start state pruning for shardID %v", sID)
	finalHeight, err := p.addDataToStateBloom(shardID, db, stateBloom)
	if err != nil {
		return err
	}
	if finalHeight == 0 {
		return nil
	}

	// get last pruned height before
	var lastPrunedHeight uint64
	temp, err := rawdbv2.GetLastPrunedHeight(db, shardID)
	if err == nil {
		err := json.Unmarshal(temp, &lastPrunedHeight)
		if err != nil {
			return err
		}
	}
	if lastPrunedHeight == 0 {
		lastPrunedHeight = 1
	} else {
		lastPrunedHeight++
	}
	stopCh := make(chan interface{})
	heightCh := make(chan uint64)
	listKeysShouldBeRemoved := &[]map[common.Hash]struct{}{}
	wg := new(sync.WaitGroup)
	for i := 0; i < runtime.NumCPU()-1; i++ {
		worker := NewWorker(stopCh, heightCh, rootHashCache, stateBloom, listKeysShouldBeRemoved, db, shardID, wg)
		go worker.start()
		defer worker.stop()
	}
	Logger.log.Infof("[state-prune] begin pruning from height %v", lastPrunedHeight)
	// retrieve all state tree from lastPrunedHeight to finalHeight - 1
	// delete all nodes which are not in state bloom
	var count, storage uint64
	for height := lastPrunedHeight; height < finalHeight; height++ {
		wg.Add(1)
		heightCh <- height
		if height%100 == 0 || height == finalHeight-1 {
			wg.Wait()
			count, storage, err = p.removeNodes(db, shardID, height, listKeysShouldBeRemoved, count, storage)
			if err != nil {
				return err
			}

		}
	}
	Logger.log.Infof("[state-prune] Delete totalNodes %v with size %v", count, storage)
	if err = p.compact(db, count); err != nil {
		return err
	}
	err = rawdbv2.StorePruneStatus(db, shardID, rawdbv2.FinishPruneStatus)
	if err != nil {
		return err
	}
	Logger.log.Infof("[state-prune] Finish state pruning for shard %v", sID)
	return nil
}

func (p *Pruner) addDataToStateBloom(shardID byte, db incdb.Database, stateBloom *trie.StateBloom) (uint64, error) {
	var finalHeight uint64
	//restore best views and final view from database
	allViews := []*blockchain.ShardBestState{}
	views, err := rawdbv2.GetShardBestState(db, shardID)
	if err != nil {
		Logger.log.Info("debug Cannot see shard best state")
		return 0, err
	}
	err = json.Unmarshal(views, &allViews)
	if err != nil {
		Logger.log.Info("debug Cannot unmarshall shard best state", string(views))
		return 0, err
	}
	//collect tree nodes want to keep, add them to state bloom
	for _, v := range allViews {
		if v.ShardHeight == 1 {
			return 0, nil
		}
		if finalHeight == 0 || finalHeight > v.ShardHeight {
			finalHeight = v.ShardHeight
		}
		Logger.log.Infof("[state-prune] Start retrieve view %s", v.BestBlockHash.String())
		var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
		stateDB, err := statedb.NewWithPrefixTrie(v.TransactionStateDBRootHash, dbAccessWarper)
		if err != nil {
			return 0, err
		}
		//Retrieve all state tree for this state
		_, err = stateDB.Retrieve(true, false, stateBloom)
		if err != nil {
			return 0, err
		}
		Logger.log.Infof("[state-prune] Finish retrieve view %s", v.BestBlockHash.String())
	}

	return finalHeight, nil
}

func (p *Pruner) compact(db incdb.Database, count uint64) error {
	if count >= rangeCompactionThreshold {
		for b := 0x00; b <= 0xf0; b += 0x10 {
			var (
				start = []byte{byte(b)}
				end   = []byte{byte(b + 0x10)}
			)
			if b == 0xf0 {
				end = nil
			}
			if err := db.Compact(start, end); err != nil {
				Logger.log.Error("Database compaction failed", "error", err)
				return err
			}
		}
	}
	return nil
}

// removeNodes after removeNodes keys map will be reset to empty value
func (p *Pruner) removeNodes(
	db incdb.Database, shardID byte, height uint64, listKeysShouldBeRemoved *[]map[common.Hash]struct{},
	totalStorage, totalNodes uint64,
) (uint64, uint64, error) {
	var storage, count uint64

	if len(*listKeysShouldBeRemoved) != 0 {
		batch := db.NewBatch()
		for _, keys := range *listKeysShouldBeRemoved {
			for key := range keys {
				temp, _ := db.Get(key.Bytes())
				storage += uint64(len(temp) + len(key.Bytes()))
				if err := batch.Delete(key.Bytes()); err != nil {
					return 0, 0, err
				}
				if batch.ValueSize() >= incdb.IdealBatchSize {
					if err := batch.Write(); err != nil {
						return 0, 0, err
					}
					batch.Reset()
				}
				count++
			}
		}
		if batch.ValueSize() > 0 {
			batch.Write()
			batch.Reset()
		}
	}
	totalStorage += uint64(storage)
	totalNodes += count
	if err := rawdbv2.StoreLastPrunedHeight(db, shardID, height); err != nil {
		return 0, 0, err
	}
	if height%10000 == 0 {
		Logger.log.Infof("[state-prune] Finish prune for height %v delete totalNodes %v with storage %v", height, totalNodes, totalStorage)
	}
	*listKeysShouldBeRemoved = make([]map[common.Hash]struct{}, 0)
	return totalNodes, totalStorage, nil
}
