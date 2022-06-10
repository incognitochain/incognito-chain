package pruner

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

type Pruner struct {
	db          map[int]incdb.Database
	stateBlooms map[int]*trie.StateBloom
}

func NewPrunerWithValue(db map[int]incdb.Database) *Pruner {
	stateBlooms := make(map[int]*trie.StateBloom)
	for i := 0; i < common.MaxShardNumber; i++ {
		stateBloom, err := trie.NewStateBloomWithSize(2048)
		if err != nil {
			panic(err)
		}
		stateBlooms[i] = stateBloom
	}

	return &Pruner{
		db:          db,
		stateBlooms: stateBlooms,
	}
}

func (p *Pruner) Prune() error {
	for i := 0; i < common.MaxShardNumber; i++ {
		if err := p.prune(i); err != nil {
			panic(err)
		}
	}
	return nil
}

func (p *Pruner) prune(sID int) error {
	shardID := byte(sID)
	db := p.db[int(shardID)]

	Logger.log.Infof("[state-prune] Start state pruning for shardID %v", sID)
	finalHeight, err := p.collectStateBloomData(shardID, db)
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
		count, size, err := p.pruneByHeight(height, db, shardID)
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
	Logger.log.Infof("[state-prune] Finish state pruning for shard %v", sID)
	return nil
}

func (p *Pruner) collectStateBloomData(shardID byte, db incdb.Database) (uint64, error) {
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
		err = v.InitStateRootHash(db)
		if err != nil {
			return 0, err
		}
		//Retrieve all state tree for this state
		_, _, err = v.GetCopiedTransactionStateDB().Retrieve(db, true, false, p.stateBlooms[int(shardID)])
		if err != nil {
			return 0, err
		}
		Logger.log.Infof("[state-prune] Finish retrieve view %s", v.BestBlockHash.String())
	}
	return finalHeight, nil
}

func (p *Pruner) pruneByHeight(height uint64, db incdb.Database, shardID byte) (int, int, error) {
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
	var count, size int
	sDB, err := statedb.NewWithPrefixTrie(sRH.TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(db))
	if err != nil {
		return 0, 0, nil
	}
	if count, size, err = sDB.Retrieve(db, false, true, p.stateBlooms[int(shardID)]); err != nil {
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
