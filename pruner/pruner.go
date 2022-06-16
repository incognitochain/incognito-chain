package pruner

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain"
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
	if err := p.Prune(2); err != nil {
		panic(err)
	}
	panic(1)
	//for i := 0; i < common.MaxShardNumber; i++ {
	//	if err := p.Prune(i); err != nil {
	//		panic(err)
	//	}
	//}
	return nil
}

func (p *Pruner) Prune(sID int) error {
	stateBloom, err := trie.NewStateBloomWithSize(2048)
	if err != nil {
		panic(err)
	}
	shardID := byte(sID)
	db := p.db[int(shardID)]
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

	// get last pruned height before
	lastPrunedKey, err := rawdbv2.GetLastPrunedKeyTrie(db, shardID)
	if err == nil {
		return err
	}
	stopCh := make(chan interface{})
	rootHashCh := make(chan blockchain.ShardRootHash)

	wg := new(sync.WaitGroup)
	workerGroup := []*Worker{}
	NUMPROCESS := 40
	for i := 0; i < NUMPROCESS; i++ {
		worker := NewWorker(stopCh, rootHashCh, stateBloom, db, shardID, wg)
		workerGroup = append(workerGroup, worker)
		go worker.start()
		defer worker.stop()
	}
	// retrieve all state tree by shard rooth hash prefix
	// delete all nodes which are not in state bloom
	var start []byte
	if len(lastPrunedKey) != 0 {
		start = lastPrunedKey
	}
	var nodes, storage, count uint64
	iter := db.NewIteratorWithPrefixStart(rawdbv2.GetShardRootsHashPrefix(shardID), start)
	defer func() {
		iter.Release()
	}()

	for iter.Next() {
		_ = iter.Key()
		rootHash := blockchain.ShardRootHash{}
		err := json.Unmarshal(iter.Value(), &rootHash)
		if err != nil {
			return err
		}
		wg.Add(1)
		rootHashCh <- rootHash

		Logger.log.Infof("[state-prune] Prune counter ... %v %v", count, rootHash.TransactionStateDBRootHash.String())
		if count%uint64(NUMPROCESS) == 0 {
			wg.Wait()
			for _, w := range workerGroup {
				wg.Add(1)
				go func() {
					c, t, e := w.removeDB()
					if e != nil {
						fmt.Println(e)
					}
					storage += t
					nodes += c
					wg.Done()
				}()
			}
			wg.Wait()

			Logger.log.Infof("[state-prune] Finish prune for %v roothash: delete totalNodes %v with storage %v", count, nodes, storage)
		}
		count++
	}

	wg.Wait()
	for _, w := range workerGroup {
		wg.Add(1)
		go func() {
			c, t, e := w.removeDB()
			if e != nil {
				fmt.Println(e)
			}
			storage += t
			nodes += c
			wg.Done()
		}()
	}
	wg.Wait()
	if err = p.compact(db, nodes); err != nil {
		return err
	}

	Logger.log.Infof("[state-prune] Delete totalNodes %v with size %v", nodes, storage)
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
		err = stateDB.Retrieve(true, false, stateBloom, nil)
		if err != nil {
			return 0, err
		}
		Logger.log.Infof("[state-prune] Finish retrieve view %s", v.BestBlockHash.String())
		break
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
