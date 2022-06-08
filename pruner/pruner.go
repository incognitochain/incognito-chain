package pruner

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/config"
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
	for i := 0; i < config.Param().ActiveShards; i++ {
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
	for i := 0; i < config.Param().ActiveShards; i++ {
		if err := p.prune(i); err != nil {
			panic(err)
		}
	}
	return nil
}

func (p *Pruner) prune(sID int) error {
	shardID := byte(sID)
	db := p.db[int(shardID)]

	Logger.log.Infof("[state-prune] Start prune for shardID %v", sID)
	//restore best views and final view from database
	allViews := []*blockchain.ShardBestState{}
	views, err := rawdbv2.GetShardBestState(db, shardID)
	if err != nil {
		Logger.log.Info("debug Cannot see shard best state")
		return err
	}
	err = json.Unmarshal(views, &allViews)
	if err != nil {
		Logger.log.Info("debug Cannot unmarshall shard best state", string(views))
		return err
	}
	var finalHeight uint64
	//collect tree nodes want to keep, add them to state bloom
	for _, v := range allViews {
		Logger.log.Infof("[state-prune] view height %v", v.ShardHeight)
		if v.ShardHeight == 1 {
			return nil
		}
		if finalHeight == 0 || finalHeight > v.ShardHeight {
			finalHeight = v.ShardHeight
		}
		Logger.log.Infof("[state-prune] Start retrieve view %s", v.BestBlockHash.String())
		err = v.InitStateRootHash(db)
		if err != nil {
			return err
		}
		//Retrieve all state tree for this state
		err = v.GetCopiedTransactionStateDB().Retrieve(db, true, false, p.stateBlooms[sID])
		if err != nil {
			return err
		}
		Logger.log.Infof("[state-prune] Finish retrieve view %s", v.BestBlockHash.String())
	}
	// get last pruned height before
	var count, lastPrunedHeight uint64
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
		count++
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
			continue
		}
		if err = sDB.Retrieve(db, false, true, p.stateBlooms[sID]); err != nil {
			return err
		}
		if err = rawdbv2.StoreLastPrunedHeight(db, shardID, height); err != nil {
			return err
		}
		if count == 100 {
			Logger.log.Infof("[state-prune] Finish prune for height %v", height)
			count = 0
		}
	}
	Logger.log.Infof("[state-prune] Finish state pruning for shard %v", sID)
	return nil
}
