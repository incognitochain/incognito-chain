package pruner

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

type Pruner struct {
	db            map[int]incdb.Database
	stateBloomDir string
	stateBlooms   map[int]*trie.StateBloom
}

func NewPrunerWithValue(db map[int]incdb.Database, stateBloomDir string) *Pruner {
	stateBlooms := make(map[int]*trie.StateBloom)
	for i := 0; i < config.Param().ActiveShards; i++ {
		stateBloom, err := trie.NewStateBloomWithSize(2048)
		if err != nil {
			panic(err)
		}
		stateBlooms[i] = stateBloom
	}

	return &Pruner{
		db:            db,
		stateBloomDir: stateBloomDir,
		stateBlooms:   stateBlooms,
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
	//stateBloomDir := filepath.Join(p.stateBloomDir, "shard_"+strconv.Itoa(sID))

	//restore best views and final view from database
	allViews := []*blockchain.ShardBestState{}
	views, err := rawdbv2.GetShardBestState(db, shardID)
	if err != nil {
		fmt.Println("debug Cannot see shard best state")
		return err
	}
	err = json.Unmarshal(views, &allViews)
	if err != nil {
		fmt.Println("debug Cannot unmarshall shard best state", string(views))
		return err
	}
	var finalHeight uint64
	//collect tree nodes want to keep, add them to state bloom
	for _, v := range allViews {
		if v.ShardHeight == 1 {
			return nil
		}
		if finalHeight == 0 || finalHeight > v.ShardHeight {
			finalHeight = v.ShardHeight
		}
		Logger.log.Infof("[state-prune] Start retrieve view %s", v.Hash().String())
		err = v.InitStateRootHash(db)
		if err != nil {
			panic(err)
		}
		//Retrieve all state tree for this state
		err = v.GetCopiedTransactionStateDB().Retrieve(true, false, p.stateBlooms[sID])
		if err != nil {
			panic(err)
		}
		Logger.log.Infof("[state-prune] Finish retrieve view %s", v.Hash().String())
	}

	/*// If the state bloom filter is already committed previously,*/
	/*// reuse it for pruning instead of generating a new one. It's*/
	/*// mandatory because a part of state may already be deleted,*/
	/*// the recovery procedure is necessary.*/
	/*_, err = findBloomFilter(stateBloomDir)*/
	/*if err != nil {*/
	/*return err*/
	/*}*/

	// get last pruned height before
	var lastPrunedHeight uint64
	temp, err := rawdbv2.GetLastPrunedHeight(db, shardID)
	if err == nil {
		height, err := common.BytesToUint64(temp)
		if err != nil {
			panic(err)
		}
		lastPrunedHeight = height
	}
	if lastPrunedHeight == 0 {
		lastPrunedHeight = 1
	}
	// retrieve all state tree from lastPrunedHeight to finalHeight - 1
	// delete all nodes which are not in state bloom
	for height := lastPrunedHeight; height < finalHeight; height++ {
		h, err := rawdbv2.GetFinalizedShardBlockHashByIndex(db, shardID, height)
		if err != nil {
			panic(err)
		}
		data, err := rawdbv2.GetShardRootsHash(db, shardID, *h)
		if err != nil {
			panic(err)
		}
		sRH := &blockchain.ShardRootHash{}
		err = json.Unmarshal(data, sRH)
		if err != nil {
			panic(err)
		}
		sDB, err := statedb.NewWithPrefixTrie(sRH.TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(db))
		if err != nil {
			panic(err)
		}
		err = sDB.Retrieve(false, true, p.stateBlooms[sID])
		if err != nil {
			panic(err)
		}
	}
	return nil
}

func (p *Pruner) recoverPruning(stateBloomPath string) error {
	return nil
}
