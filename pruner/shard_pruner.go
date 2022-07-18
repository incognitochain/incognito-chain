package pruner

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type ShardPruner struct {
	//state
	shardID     int
	db          incdb.Database
	stateBloom  *trie.StateBloom
	bloomSize   uint64
	role        string
	finalHeight uint64
	bestView    *blockchain.ShardBestState

	//lock
	lock            sync.Mutex
	wg              sync.WaitGroup
	shardInsertLock *sync.Mutex

	//report
	lastTriggerTime      time.Time
	lastProcessingMode   string
	status               int
	lastError            string
	lastProcessingHeight uint64
	storage              uint64
	nodes                uint64
}

func NewShardPruner(sid int, db incdb.Database) *ShardPruner {
	//init object
	sp := &ShardPruner{
		shardID: sid,
		db:      db,
	}
	sp.restoreStatus()
	if sp.lastProcessingHeight == 0 {
		sp.lastProcessingHeight = 1
	}
	return sp
}

func (s *ShardPruner) SetBloomSize(size uint64) {
	s.bloomSize = size
}

func (s *ShardPruner) Stop() {
	s.status = IDLE
}

func (s *ShardPruner) InitBloomState() error {
	//restore best views and final view from database
	allViews := []*blockchain.ShardBestState{}
	views, err := rawdbv2.GetShardBestState(s.db, byte(s.shardID))
	if err != nil {
		Logger.log.Errorf("debug Cannot see shard best state %v", err)
		return err
	}
	err = json.Unmarshal(views, &allViews)
	if err != nil {
		Logger.log.Errorf("debug Cannot unmarshall shard best state", string(views))
		return err
	}
	//collect tree nodes want to keep, add them to state bloom
	if len(allViews) > 0 {
		s.finalHeight = allViews[0].ShardHeight
		s.bestView = allViews[len(allViews)-1]
	} else {
		return errors.New("Cannot retrieve all shard views")
	}
	for _, v := range allViews {
		err = s.addViewToBloom(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ShardPruner) Prune(byHash bool) error {
	s.lock.Lock()
	if s.status != IDLE {
		s.lock.Unlock()
		return fmt.Errorf("Shard %v is not ready! State: %v", s.shardID, s.status)
	}
	s.status = PRUNING
	s.lock.Unlock()

	s.lastTriggerTime = time.Now()

	err := s.InitBloomState()
	if err != nil {
		s.lastError = errors.Wrap(err, "init bloom state fail").Error()
		return err
	}

	if byHash {
		s.lastProcessingMode = "hash"
		s.pruneByHash()
	} else {
		s.lastProcessingMode = "height"
		s.PruneByHeight()
	}
	s.saveStatus()
	s.status = CHECKING
	s.stateBloom = nil
	s.CheckDataIntegrity()
	s.status = IDLE
	return nil
}

func (s *ShardPruner) PruneByHeight() error {
	//prune height
	if s.finalHeight <= 1 {
		s.lastError = ""
		return nil
	}
	for height := s.lastProcessingHeight; height < s.finalHeight; height++ {
		if s.status != PRUNING {
			return nil
		}
		err := func() error {
			s.LockInsertShardBlock() //lock insert shard block
			defer s.UnlockInsertShardBlock()
			s.wg.Wait() //wait for all insert bloom task (in case we have new view)

			//recheck if there is error when handle new view
			if s.status != PRUNING {
				return nil
			}

			storage, node, err := pruneByHeight(s.db, s.shardID, s.stateBloom, height)
			s.storage += storage
			s.nodes += node
			if err != nil {
				return err
			}

			if height%1000 == 0 {
				Logger.log.Infof("[state-prune %v] Finish prune for height %v delete totalNodes %v with storage %v", s.shardID, height, s.nodes, s.storage)
				s.saveStatus()
			}
			s.lastProcessingHeight = height
			return nil
		}()

		if err != nil {
			s.lastError = errors.Wrap(err, "prune by height fail").Error()
		} else {
			s.lastError = ""
		}
	}
	return nil
}

func (s *ShardPruner) pruneByHash() error {
	iter := s.db.NewIteratorWithPrefixStart(rawdbv2.GetShardRootsHashPrefix(byte(s.shardID)), nil)
	defer func() {
		iter.Release()
	}()
	count := 0

	// retrieve all state tree by shard rooth hash prefix
	// delete all nodes which are not in state bloom
	for iter.Next() {
		if s.status != PRUNING {
			return nil
		}

		err := func() error {
			s.LockInsertShardBlock() //lock insert shard block
			defer s.UnlockInsertShardBlock()
			s.wg.Wait() //wait for all handle new view task (in case we have new view)

			//recheck if there is error when handle new view
			if s.status != PRUNING {
				return nil
			}

			key := iter.Key()
			rootHash := &blockchain.ShardRootHash{}
			err := json.Unmarshal(iter.Value(), &rootHash)
			if err != nil {
				return err
			}
			storage, node, err := pruneTxStateDB(s.db, s.stateBloom, rootHash)
			s.storage += storage
			s.nodes += node
			if err != nil {
				return err
			}

			if count%1000 == 0 {
				Logger.log.Infof("[state-prune] Finish prune for key %v totalKeys %v delete totalNodes %v with storage %v", key, count, s.nodes, s.storage)
				s.saveStatus()
			}
			return nil
		}()

		if err != nil {
			s.lastError = err.Error()
			return err
		} else {
			s.lastError = ""
		}

	}
	return nil
}

func (s *ShardPruner) addViewToBloom(v *blockchain.ShardBestState) error {
	Logger.log.Infof("[state-prune %v] Start retrieve view %s at height %v hash %v ",
		v.ShardID, v.BestBlockHash.String(), v.ShardHeight, v.TransactionStateDBRootHash.String())
	var dbAccessWarper = statedb.NewDatabaseAccessWarper(s.db)
	stateDB, err := statedb.NewWithPrefixTrie(v.TransactionStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}
	//Retrieve all state tree for this state
	if s.stateBloom == nil {
		s.stateBloom, _ = trie.NewStateBloomWithSize(s.bloomSize)
		_, err = stateDB.Retrieve(true, false, s.stateBloom, true)
	} else {
		_, err = stateDB.Retrieve(true, false, s.stateBloom, false)
	}

	if err != nil {
		panic(err)
		return err
	}
	Logger.log.Infof("[state-prune %v] Finish retrieve view %s at height %v",
		v.ShardID, v.BestBlockHash.String(), v.ShardHeight)
	return nil
}

func (s *ShardPruner) CheckDataIntegrity() {
	txDB, err := statedb.NewWithPrefixTrie(s.bestView.TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(s.db))
	if err != nil {
		panic(fmt.Sprintf("Something wrong when init txDB %v", s.shardID))
	}
	if err := txDB.Recheck(); err != nil {
		fmt.Println("Recheck TransactionStateDBRootHash", s.bestView.ShardHeight, s.bestView.TransactionStateDBRootHash.String())
		Logger.log.Infof("[state-prune %v] Shard %v Prune data error! %v", s.shardID, s.shardID, err)
		panic(fmt.Sprintf("Prune data error! Shard %v Database corrupt!", s.shardID))
	}
}

func (s *ShardPruner) LockInsertShardBlock() {
	if s.shardInsertLock != nil {
		s.shardInsertLock.Lock()
	}
}

func (s *ShardPruner) UnlockInsertShardBlock() {
	if s.shardInsertLock != nil {
		s.shardInsertLock.Unlock()
	}
}

func (s *ShardPruner) handleNewView(shardBestState *blockchain.ShardBestState) {
	s.wg.Add(1)
	s.bestView = shardBestState
	go func() {
		defer s.wg.Done()
		if s.status == PRUNING {
			err := s.addViewToBloom(shardBestState)
			if err != nil {
				s.lastError = errors.Wrap(err, "handle new view fail").Error()
				s.status = IDLE
				return
			}
		}
	}()
}
