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
	db       map[int]incdb.Database
	Statuses map[int]byte
	StatusMu sync.RWMutex
	mu       sync.Mutex
}

func NewPrunerWithValue(db map[int]incdb.Database, statuses map[int]byte) *Pruner {
	return &Pruner{
		db:       db,
		Statuses: statuses,
	}
}

func (p *Pruner) ReadStatus() {
	for i := 0; i < common.MaxShardNumber; i++ {
		status, _ := rawdbv2.GetPruneStatus(p.db[i], byte(i)) //ignore error for case not store status yet
		p.Statuses[i] = status
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
	p.mu.Lock()
	defer func() {
		p.mu.Unlock()
	}()
	shardID := byte(sID)
	db := p.db[int(shardID)]
	err := rawdbv2.StorePruneStatus(db, shardID, rawdbv2.ProcessingPruneStatus)
	if err != nil {
		return err
	}
	p.StatusMu.Lock()
	p.Statuses[sID] = rawdbv2.ProcessingPruneStatus
	p.StatusMu.Unlock()
	stateBloom, err := trie.NewStateBloomWithSize(2048)
	if err != nil {
		panic(err)
	}
	rootHashCache, err := lru.New(100)
	if err != nil {
		panic(err)
	}
	Logger.log.Infof("[state-prune] Start state pruning for shard %v", sID)
	defer func() {
		Logger.log.Infof("[state-prune] Finish state pruning for shard %v", sID)
	}()
	finalHeight, err := p.addDataToStateBloom(shardID, db, stateBloom)
	if err != nil {
		return err
	}
	if finalHeight == 0 {
		return nil
	}

	stopCh := make(chan interface{})
	rootHashCh := make(chan blockchain.ShardRootHash)
	listKeysShouldBeRemoved := &[]map[common.Hash]struct{}{}
	wg := new(sync.WaitGroup)
	for i := 0; i < runtime.NumCPU()-1; i++ {
		worker := NewWorker(stopCh, rootHashCh, rootHashCache, stateBloom, listKeysShouldBeRemoved, db, shardID, wg)
		go worker.start()
		defer worker.stop()
	}
	nodes, storage, err := p.traverseAndDelete(db, shardID, listKeysShouldBeRemoved, rootHashCh, wg)
	if err != nil {
		return err
	}
	if err = p.compact(db, nodes); err != nil {
		return err
	}
	Logger.log.Infof("[state-prune] Delete totalNodes %v with size %v", nodes, storage)
	err = rawdbv2.StorePruneStatus(db, shardID, rawdbv2.FinishPruneStatus)
	if err != nil {
		return err
	}
	err = rawdbv2.StorePendingPrunedNodes(db, shardID, 0)
	if err != nil {
		return err
	}
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
		Logger.log.Infof("[state-prune] Start retrieve view %s at height %v", v.BestBlockHash.String(), v.ShardHeight)
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
		Logger.log.Infof("[state-prune] Finish retrieve view %s at height %v", v.BestBlockHash.String(), v.ShardHeight)
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
	db incdb.Database, shardID byte, key []byte,
	listKeysShouldBeRemoved *[]map[common.Hash]struct{}, totalNodes, totalStorage uint64,
) (uint64, uint64, error) {
	var storage, count uint64

	if len(*listKeysShouldBeRemoved) != 0 {
		keysShouldBeRemoved := make(map[common.Hash]struct{})
		for _, keys := range *listKeysShouldBeRemoved {
			for key := range keys {
				keysShouldBeRemoved[key] = struct{}{}
			}
		}
		batch := db.NewBatch()
		for key := range keysShouldBeRemoved {
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
		if batch.ValueSize() > 0 {
			batch.Write()
			batch.Reset()
		}
	}
	totalStorage += uint64(storage)
	totalNodes += count
	if err := rawdbv2.StoreLastPrunedKeyTrie(db, shardID, key); err != nil {
		return 0, 0, err
	}
	if err := rawdbv2.StorePendingPrunedNodes(db, shardID, totalNodes); err != nil {
		return 0, 0, err
	}
	*listKeysShouldBeRemoved = make([]map[common.Hash]struct{}, 0)
	return totalNodes, totalStorage, nil
}

func (p *Pruner) traverseAndDelete(
	db incdb.Database, shardID byte, listKeysShouldBeRemoved *[]map[common.Hash]struct{},
	rootHashCh chan blockchain.ShardRootHash, wg *sync.WaitGroup,
) (uint64, uint64, error) {
	var nodes, storage, count uint64
	// get last pruned height before
	lastPrunedKey, err := rawdbv2.GetLastPrunedKeyTrie(db, shardID)
	if err == nil {
		return 0, 0, err
	}
	var start []byte
	if len(lastPrunedKey) != 0 {
		start = lastPrunedKey
	}
	Logger.log.Infof("[state-prune] begin pruning from key %v", lastPrunedKey)
	nodes, _ = rawdbv2.GetPendingPrunedNodes(db, shardID) // not checking error avoid case not store pruned node yet
	iter := db.NewIteratorWithPrefixStart(rawdbv2.GetShardRootsHashPrefix(shardID), start)
	var finalPrunedKey []byte

	// retrieve all state tree by shard rooth hash prefix
	// delete all nodes which are not in state bloom
	for iter.Next() {
		key := iter.Key()
		rootHash := blockchain.ShardRootHash{}
		err := json.Unmarshal(iter.Value(), &rootHash)
		if err != nil {
			return 0, 0, err
		}
		wg.Add(1)
		rootHashCh <- rootHash
		finalPrunedKey = key
		if count%uint64(runtime.NumCPU()) == 0 {
			wg.Wait()
			nodes, storage, err = p.removeNodes(db, shardID, key, listKeysShouldBeRemoved, nodes, storage)
			if err != nil {
				return 0, 0, err
			}
			if count%10000 == 0 {
				Logger.log.Infof("[state-prune] Finish prune for key %v totalKeys %v delete totalNodes %v with storage %v", key, count, nodes, storage)
			}
			finalPrunedKey = []byte{}
		}
		count++
	}
	if len(finalPrunedKey) != 0 {
		nodes, storage, err = p.removeNodes(db, shardID, finalPrunedKey, listKeysShouldBeRemoved, nodes, storage)
		if err != nil {
			return 0, 0, err
		}
	}
	iter.Release()
	return nodes, storage, nil
}
