package pruner

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
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
	status               int
	lastError            string
	lastProcessingHeight uint64
	storage              uint64
	nodes                uint64
}

func NewShardPruner(sid int, db incdb.Database) *ShardPruner {
	//get last prune height
	lastPrunedHeight, _ := rawdbv2.GetLastPrunedHeight(db)
	if lastPrunedHeight == 0 {
		lastPrunedHeight = 1
	} else {
		lastPrunedHeight++
	}
	//init object
	sp := &ShardPruner{
		shardID:              sid,
		db:                   db,
		lastProcessingHeight: lastPrunedHeight,
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
	//init bloom state
	var err error
	s.stateBloom, err = trie.NewStateBloomWithSize(s.bloomSize)
	if err != nil {
		return err
	}

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
	}
	for _, v := range allViews {
		err = s.addViewToBloom(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ShardPruner) PruneByHeight() error {
	s.lock.Lock()
	if s.status != RUNNING {
		s.lock.Unlock()
		return fmt.Errorf("Shard %v is pruning at height %v", s.shardID, s.lastProcessingHeight)
	}
	s.status = RUNNING
	s.lock.Unlock()

	s.lastTriggerTime = time.Now()

	err := s.InitBloomState()
	if err != nil {
		s.lastError = err.Error()
		s.status = ERROR
		return err
	}

	//prune height
	for height := s.lastProcessingHeight; height < s.finalHeight; height++ {
		if s.status != RUNNING {
			return nil
		}
		s.LockInsertShardBlock() //lock insert shard block

		s.wg.Wait() //wait for all insert bloom task (in case we have new view)

		storage, node, err := s.pruneByHeight(height)
		s.storage += storage
		s.nodes += node
		if err != nil {
			s.lastError = err.Error()
			s.status = ERROR
			s.UnlockInsertShardBlock()
			break
		}

		s.UnlockInsertShardBlock()
		if height%10000 == 0 {
			Logger.log.Infof("[state-prune %v] Finish prune for height %v delete totalNodes %v with storage %v", s.shardID, height, s.nodes, s.storage)
		}
	}
	s.stateBloom = nil
	s.CheckDataIntegrity()
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
	_, _, err = stateDB.Retrieve(true, false, s.stateBloom)
	if err != nil {
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
	go func() {
		defer s.wg.Done()
		if s.status == RUNNING {
			err := s.addViewToBloom(shardBestState)
			if err != nil {
				s.lastError = err.Error()
				s.status = ERROR
				return
			}
		}
	}()
}

func (s *ShardPruner) SetCommitteeRole(role string) {
	s.role = role
}
