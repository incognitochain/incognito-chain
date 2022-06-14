package pruner

import (
	"encoding/json"
	"fmt"
	"runtime"

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
	numWorkers := runtime.NumCPU() - 1
	if numWorkers == 0 {
	}

	Logger.log.Infof("[state-prune] Start state pruning for shardID %v", sID)
	finalHeight, err := p.collectStateBloomData(shardID, db, stateBloom)
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
	var totalNodes, totalSize int
	Logger.log.Infof("[state-prune] begin pruning from height %v", lastPrunedHeight)
	// retrieve all state tree from lastPrunedHeight to finalHeight - 1
	// delete all nodes which are not in state bloom
	for height := lastPrunedHeight; height < finalHeight; height++ {
		count, size, err := p.pruneByHeight(height, db, shardID, stateBloom, rootHashCache)
		if err != nil {
			return err
		}
		totalNodes += count
		totalSize += size
		if height%10000 == 0 {
			Logger.log.Infof("[state-prune] Finish prune for height %v delete totalNodes %v with size %v", height, totalNodes, totalSize)
		}
	}
	Logger.log.Infof("[state-prune] Delete totalNodes %v with size %v", totalNodes, totalSize)
	if err = p.compact(db, totalNodes); err != nil {
		return err
	}
	err = rawdbv2.StorePruneStatus(db, shardID, rawdbv2.FinishPruneStatus)
	if err != nil {
		return err
	}
	Logger.log.Infof("[state-prune] Finish state pruning for shard %v", sID)
	return nil
}

func (p *Pruner) collectStateBloomData(shardID byte, db incdb.Database, stateBloom *trie.StateBloom) (uint64, error) {
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
		Logger.log.Infof("[state-prune] view height %v", v.ShardHeight)
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
		_, _, err = stateDB.Retrieve(db, true, false, stateBloom)
		if err != nil {
			return 0, err
		}
		Logger.log.Infof("[state-prune] Finish retrieve view %s", v.BestBlockHash.String())
	}
	return finalHeight, nil
}

func (p *Pruner) pruneByHeight(
	height uint64, db incdb.Database, shardID byte, stateBloom *trie.StateBloom, rootHashCache *lru.Cache,
) (int, int, error) {
	h, err := rawdbv2.GetFinalizedShardBlockHashByIndex(db, shardID, height)
	if err != nil {
		return 0, 0, err
	}
	data, err := rawdbv2.GetShardRootsHash(db, shardID, *h)
	if err != nil {
		return 0, 0, err
	}
	sRH := &blockchain.ShardRootHash{}
	if err = json.Unmarshal(data, sRH); err != nil {
		return 0, 0, err
	}
	count, size, err := p.pruneTxStateDB(sRH, db, stateBloom, rootHashCache)
	if err != nil {
		return 0, 0, err
	}
	if err = rawdbv2.StoreLastPrunedHeight(db, shardID, height); err != nil {
		return 0, 0, err
	}
	return count, size, nil
}

func (p *Pruner) compact(db incdb.Database, count int) error {
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

func (p *Pruner) pruneTxStateDB(sRH *blockchain.ShardRootHash, db incdb.Database, stateBloom *trie.StateBloom, rootHashCache *lru.Cache) (int, int, error) {
	if _, ok := rootHashCache.Get(sRH.TransactionStateDBRootHash.String()); ok {
		return 0, 0, nil
	}
	var count, size int
	sDB, err := statedb.NewWithPrefixTrie(sRH.TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(db))
	if err != nil {
		return 0, 0, nil
	}
	if count, size, err = sDB.Retrieve(db, false, true, stateBloom); err != nil {
		return 0, 0, err
	}
	if ok := rootHashCache.Add(sRH.TransactionStateDBRootHash.String(), struct{}{}); !ok {
		return 0, 0, fmt.Errorf("Cannot add rootHash to lru cache")
	}
	return count, size, nil
}
