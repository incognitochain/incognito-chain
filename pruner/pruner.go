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
	Logger.log.Infof("[state-prune] begin pruning from height %v", lastPrunedHeight)
	// retrieve all state tree from lastPrunedHeight to finalHeight - 1
	// delete all nodes which are not in state bloom
	for height := lastPrunedHeight; height < finalHeight; height++ {
		err := p.pruneByHeight(height, db, shardID)
		if err != nil {
			return err
		}
		if height%50 == 0 {
			Logger.log.Infof("[state-prune] Finish prune for height %v", height)
		}
	}
	if err = p.compact(db); err != nil {
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
		err = v.GetCopiedTransactionStateDB().Retrieve(db, true, false, p.stateBlooms[int(shardID)])
		if err != nil {
			return 0, err
		}
		Logger.log.Infof("[state-prune] Finish retrieve view %s", v.BestBlockHash.String())
	}
	return finalHeight, nil
}

func (p *Pruner) pruneByHeight(height uint64, db incdb.Database, shardID byte) error {
	h, err := rawdbv2.GetFinalizedShardBlockHashByIndex(db, shardID, height)
	if err != nil {
		return err
	}
	data, err := rawdbv2.GetShardRootsHash(db, shardID, *h)
	if err != nil {
		return err
	}
	sRH := &blockchain.ShardRootHash{}
	if err = json.Unmarshal(data, sRH); err != nil {
		return err
	}
	sDB, err := statedb.NewWithPrefixTrie(sRH.TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(db))
	if err != nil {
		return nil
	}
	if err = sDB.Retrieve(db, false, true, p.stateBlooms[int(shardID)]); err != nil {
		return err
	}
	if err = rawdbv2.StoreLastPrunedHeight(db, shardID, height); err != nil {
		return err
	}
	return nil
}

func (p *Pruner) compact(db incdb.Database) error {
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
	return nil
}
