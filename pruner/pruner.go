package pruner

import (
	"context"
	"encoding/json"
	"fmt"
	"golang.org/x/sync/semaphore"
	"os"
	"runtime"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/trie"
)

type Pruner struct {
	bestView                sync.Map
	db                      map[int]incdb.Database
	statuses                map[byte]byte
	updateStatusCh          chan UpdateStatus
	TriggerCh               chan ExtendedConfig
	shardInsertLock         map[int]*sync.Mutex
	wg                      *sync.WaitGroup
	PubSubManager           *pubsub.PubSubManager
	currentValidatorShardID int
	stateBloom              map[int]*trie.StateBloom
	addedViewsCache         map[common.Hash]struct{}
}

func NewPrunerWithValue(db map[int]incdb.Database, statuses map[byte]byte) *Pruner {
	return &Pruner{
		bestView:                sync.Map{},
		db:                      db,
		statuses:                statuses,
		updateStatusCh:          make(chan UpdateStatus),
		TriggerCh:               make(chan ExtendedConfig, 1),
		shardInsertLock:         make(map[int]*sync.Mutex),
		currentValidatorShardID: -2,
		wg:                      new(sync.WaitGroup),
		addedViewsCache:         make(map[common.Hash]struct{}),
		stateBloom:              make(map[int]*trie.StateBloom),
	}
}

func (p *Pruner) SetShardInsertLock(sid int, mutex *sync.Mutex) {
	p.shardInsertLock[sid] = mutex
}

func (p *Pruner) LockInsertShardBlock(sid int) {
	if p.shardInsertLock[sid] != nil {
		p.shardInsertLock[sid].Lock()
	}
}

func (p *Pruner) UnlockInsertShardBlock(sid int) {
	if p.shardInsertLock[sid] != nil {
		p.shardInsertLock[sid].Unlock()
	}
}

func (p *Pruner) ReadStatus() {
	for i := 0; i < common.MaxShardNumber; i++ {
		s, _ := rawdbv2.GetPruneStatus(p.db[i]) //ignore error for case not store status yet
		status := s
		if s == rawdbv2.ProcessingPruneByHashStatus {
			status = rawdbv2.WaitingPruneByHashStatus
		}
		if s == rawdbv2.ProcessingPruneByHeightStatus {
			status = rawdbv2.WaitingPruneByHeightStatus
		}
		if s != status {
			rawdbv2.StorePruneStatus(p.db[i], status)
		}
		p.statuses[byte(i)] = status
	}
}

func (p *Pruner) Prune() error {
	cpus := runtime.NumCPU()
	sem := semaphore.NewWeighted(int64(cpus / 2))
	ch := make(chan int)
	stateBloomSize := config.Config().StateBloomSize / uint64(common.MaxShardNumber)
	if cpus/2 < common.MaxShardNumber {
		stateBloomSize = config.Config().StateBloomSize / uint64(cpus/2)
	}

	stopCh := make(chan struct{})
	var count int
	var wg sync.WaitGroup
	go func() {
		for {
			select {
			case shardID := <-ch:
				wg.Add(1)
				go func() {
					if err := p.prune(shardID, false, stateBloomSize); err != nil {
						Logger.log.Error(err)
						return
					}
					wg.Done()
					sem.Release(1)
				}()
			case <-stopCh:
				count++
				if count == common.MaxShardNumber {
					return
				}
			}
		}
	}()
	for i := 0; i < 1; i++ {
		sem.Acquire(context.Background(), 1)
		ch <- i
	}
	wg.Wait()
	return nil
}

func (p *Pruner) prune(sID int, shouldPruneByHash bool, stateBloomSize uint64) error {
	shardID := byte(sID)
	db := p.db[int(shardID)]
	Logger.log.Infof("[state-prune] Start state pruning for shard %v", sID)
	defer func() {
		Logger.log.Infof("[state-prune] Finish state pruning for shard %v", sID)
	}()
	rootHashCache, err := lru.New(100)
	if err != nil {
		panic(err)
	}

	_, err = os.Stat("/data/shard0")
	if err != nil {
		p.stateBloom[sID], err = trie.NewStateBloomWithSize(stateBloomSize)
		if err != nil {
			panic(err)
		}
	} else {
		p.stateBloom[sID], err = trie.NewStateBloomFromDisk("/data/shard0")
		if err != nil {
			panic(err)
		}
	}

	finalHeight, err := p.addDataToStateBloom(shardID, db)
	if err != nil {
		return err
	}
	if finalHeight == 0 {
		return nil
	}

	err = p.stateBloom[sID].Commit("/data/shard0", "/data/shard0_tmp")
	if err != nil {
		return err
	}

	stopCh := make(chan interface{})
	heightCh := make(chan uint64)
	rootHashCh := make(chan blockchain.ShardRootHash)

	listKeysShouldBeRemoved := &[]map[common.Hash]struct{}{}
	wg := new(sync.WaitGroup)
	for i := 0; i < 1; i++ {
		worker := NewWorker(stopCh, heightCh, rootHashCache, p.stateBloom[sID], listKeysShouldBeRemoved, db, shardID, wg)
		go worker.start()
		defer worker.stop()
	}
	helper := TraverseHelper{
		db:          db,
		shardID:     shardID,
		finalHeight: finalHeight,
		heightCh:    heightCh,
		rootHashCh:  rootHashCh,
		wg:          wg,
	}
	err = p.traverseAndDelete(helper, listKeysShouldBeRemoved, shouldPruneByHash)
	if err != nil {
		return err
	}
	err = rawdbv2.StorePendingPrunedNodes(db, 0)
	if err != nil {
		return err
	}
	p.stateBloom[sID] = nil
	return nil
}

func (p *Pruner) addDataToStateBloom(shardID byte, db incdb.Database) (uint64, error) {
	var finalHeight uint64
	//restore best views and final view from database
	allViews := []*blockchain.ShardBestState{}
	views, err := rawdbv2.GetShardBestState(db, shardID)
	if err != nil {
		Logger.log.Infof("debug Cannot see shard best state %v", err)
		return 0, err
	}
	err = json.Unmarshal(views, &allViews)
	if err != nil {
		Logger.log.Info("debug Cannot unmarshall shard best state", string(views))
		return 0, err
	}
	//collect tree nodes want to keep, add them to state bloom
	//for _, v := range allViews {
	//	if finalHeight == 0 || finalHeight > v.ShardHeight {
	//		finalHeight = v.ShardHeight
	//	}
	v := allViews[len(allViews)-1]
	finalHeight = v.ShardHeight
	p.bestView.Store(shardID, v)
	err = p.addNewViewToStateBloom(v, db)
	if err != nil {
		return 0, err
	}
	//}

	return finalHeight, nil
}

func (p *Pruner) addNewViewToStateBloom(
	v *blockchain.ShardBestState, db incdb.Database,
) error {
	if _, found := p.addedViewsCache[v.BestBlockHash]; found {
		return nil
	}
	if v.ShardHeight == 1 {
		return nil
	}
	Logger.log.Infof("[state-prune %v] Start retrieve view %s at height %v", v.ShardID, v.BestBlockHash.String(), v.ShardHeight)
	var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
	stateDB, err := statedb.NewWithPrefixTrie(v.TransactionStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}
	//Retrieve all state tree for this state
	_, _, err = stateDB.Retrieve(true, false, p.stateBloom[int(v.ShardID)])
	if err != nil {
		return err
	}
	p.addedViewsCache[v.BestBlockHash] = struct{}{}
	Logger.log.Infof("[state-prune %v] Finish retrieve view %s at height %v", v.ShardID, v.BestBlockHash.String(), v.ShardHeight)
	return nil
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

func (p *Pruner) traverseAndDelete(
	helper TraverseHelper, listKeysShouldBeRemoved *[]map[common.Hash]struct{}, shouldPruneByHash bool,
) error {
	if helper.finalHeight <= 1 {
		return nil
	}
	var nodes, storage uint64
	var err error
	if shouldPruneByHash {
		nodes, storage, err = p.traverseAndDeleteByHash(helper, listKeysShouldBeRemoved)
		if err != nil {
			return err
		}
		Logger.log.Infof("[state-prune %v] Start compact totalNodes %v with size %v", helper.shardID, nodes, storage)
		if err = p.compact(helper.db, nodes); err != nil {
			return err
		}
		Logger.log.Infof("[state-prune %v] Finish compact totalNodes %v with size %v", helper.shardID, nodes, storage)
		sBestView, ok := p.bestView.Load(helper.shardID)
		if !ok {
			panic("Something wrong when get shard bestview")
		}
		txDB, err := statedb.NewWithPrefixTrie(sBestView.(*blockchain.ShardBestState).TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(helper.db))
		if err != nil {
			panic(fmt.Sprintf("Something wrong when init txDB %v", helper.shardID))
		}
		if err := txDB.Recheck(); err != nil {
			Logger.log.Infof("[state-prune %v] Shard %v Prune data error! %v", helper.shardID, helper.shardID, err)
			panic(fmt.Sprintf("Prune data error! Shard %v Database corrupt!", helper.shardID))
		}
	} else {
		nodes, storage, err = p.traverseAndDeleteByHeight(helper, listKeysShouldBeRemoved)
		if err != nil {
			return err
		}
		sBestView, ok := p.bestView.Load(helper.shardID)
		if !ok {
			panic(fmt.Sprintf("Something wrong when get shard bestview %v", helper.shardID))
		}

		txDB, err := statedb.NewWithPrefixTrie(sBestView.(*blockchain.ShardBestState).TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(helper.db))
		if err != nil {
			panic(fmt.Sprintf("Something wrong when init txDB %v", helper.shardID))
		}
		if err := txDB.Recheck(); err != nil {
			Logger.log.Infof("[state-prune %v] Shard %v Prune data error! %v", helper.shardID, helper.shardID, err)
			panic(fmt.Sprintf("Prune data error! Shard %v Database corrupt!", helper.shardID))
		}
	}
	p.statuses[helper.shardID] = rawdbv2.FinishPruneStatus
	Logger.log.Infof("[state-prune %v] Delete totalNodes %v with size %v", helper.shardID, nodes, storage)

	return nil
}

func (p *Pruner) traverseAndDeleteByHash(
	helper TraverseHelper, listKeysShouldBeRemoved *[]map[common.Hash]struct{},
) (uint64, uint64, error) {
	var nodes, storage, count uint64
	lastPrunedKey, err := rawdbv2.GetLastPrunedKeyTrie(helper.db)
	var start []byte
	if len(lastPrunedKey) != 0 {
		start = lastPrunedKey
	}
	Logger.log.Infof("[state-prune] begin pruning from key %v", lastPrunedKey)
	nodes, _ = rawdbv2.GetPendingPrunedNodes(helper.db) // not checking error avoid case not store pruned node yet
	iter := helper.db.NewIteratorWithPrefixStart(rawdbv2.GetShardRootsHashPrefix(helper.shardID), start)
	defer func() {
		iter.Release()
	}()
	var finalPrunedKey []byte

	// retrieve all state tree by shard rooth hash prefix
	// delete all nodes which are not in state bloom
	for iter.Next() {
		if p.statuses[helper.shardID] == rawdbv2.FinishPruneStatus {
			return nodes, storage, nil
		}
		p.LockInsertShardBlock(int(helper.shardID))
		p.wg.Wait()
		key := iter.Key()
		rootHash := blockchain.ShardRootHash{}
		err := json.Unmarshal(iter.Value(), &rootHash)
		if err != nil {
			return 0, 0, err
		}
		helper.wg.Add(1)
		helper.rootHashCh <- rootHash
		finalPrunedKey = key
		helper.wg.Wait()
		nodes, storage, err = p.removeNodes(helper.db, helper.shardID, key, 0, listKeysShouldBeRemoved, nodes, storage, true)
		p.UnlockInsertShardBlock(int(helper.shardID))
		if err != nil {
			return 0, 0, err
		}
		if count%10000 == 0 {
			Logger.log.Infof("[state-prune] Finish prune for key %v totalKeys %v delete totalNodes %v with storage %v", key, count, nodes, storage)
		}
		finalPrunedKey = []byte{}
		count++

	}
	if len(finalPrunedKey) != 0 {
		nodes, storage, err = p.removeNodes(helper.db, helper.shardID, finalPrunedKey, 0, listKeysShouldBeRemoved, nodes, storage, true)
		if err != nil {
			return 0, 0, err
		}
	}
	iter.Release()
	return nodes, storage, nil
}

func (p *Pruner) traverseAndDeleteByHeight(
	helper TraverseHelper, listKeysShouldBeRemoved *[]map[common.Hash]struct{},
) (uint64, uint64, error) {
	var nodes, storage uint64
	var err error
	// get last pruned height before
	lastPrunedHeight, _ := rawdbv2.GetLastPrunedHeight(helper.db)
	if lastPrunedHeight == 0 {
		lastPrunedHeight = 1
	} else {
		lastPrunedHeight++
	}

	fmt.Println(lastPrunedHeight, helper.finalHeight)
	for height := lastPrunedHeight; height < helper.finalHeight; height++ {
		if p.statuses[helper.shardID] == rawdbv2.FinishPruneStatus {
			return nodes, storage, nil
		}
		p.LockInsertShardBlock(int(helper.shardID)) //lock insert shard block

		p.wg.Wait() //wait for all insert bloom task (in case we have new view)

		helper.wg.Add(1)
		helper.heightCh <- height
		helper.wg.Wait()

		nodes, storage, err = p.removeNodes(helper.db, helper.shardID, nil, height, listKeysShouldBeRemoved, nodes, storage, false)
		p.UnlockInsertShardBlock(int(helper.shardID))
		if err != nil {
			return 0, 0, err
		}
		cnt := 0
		if height%10000 == 0 {
			Logger.log.Infof("[state-prune %v] Finish prune for height %v delete totalNodes %v with storage %v", helper.shardID, height, nodes, storage)
			cnt++
			if cnt == 10 {
				break
			}
		}
	}
	return nodes, storage, nil
}

// removeNodes after removeNodes keys map will be reset to empty value
func (p *Pruner) removeNodes(
	db incdb.Database, shardID byte, key []byte, height uint64,
	listKeysShouldBeRemoved *[]map[common.Hash]struct{}, totalNodes, totalStorage uint64, shouldPruneByHash bool,
) (uint64, uint64, error) {
	var storage, count uint64

	if len(*listKeysShouldBeRemoved) != 0 {
		keysShouldBeRemoved := make(map[common.Hash]struct{})
		if len(*listKeysShouldBeRemoved) == 1 {
			keysShouldBeRemoved = (*listKeysShouldBeRemoved)[0]
		} else {
			for _, keys := range *listKeysShouldBeRemoved {
				for key := range keys {
					keysShouldBeRemoved[key] = struct{}{}
				}
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

	if shouldPruneByHash {
		if err := rawdbv2.StoreLastPrunedKeyTrie(db, key); err != nil {
			return 0, 0, err
		}
		if err := rawdbv2.StorePendingPrunedNodes(db, totalNodes); err != nil {
			return 0, 0, err
		}
	} else {
		if err := rawdbv2.StoreLastPrunedHeight(db, height); err != nil {
			return 0, 0, err
		}
	}

	*listKeysShouldBeRemoved = make([]map[common.Hash]struct{}, 0)
	return totalNodes, totalStorage, nil
}

func (p *Pruner) InsertNewView(shardBestState *blockchain.ShardBestState) {
	p.wg.Add(1)
	go func() {
		if err := p.handleNewView(shardBestState); err != nil {
			panic(err)
		}
		p.wg.Done()
	}()
	//TODO: if sync latest && (lastheight < shardbeststate - 1000) && auto prune || force prune => trigger prunning
}

func (p *Pruner) Start() {
	_, nodeRoleCh, err := p.PubSubManager.RegisterNewSubscriber(pubsub.NodeRoleDetailTopic)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case updateStatus := <-p.updateStatusCh:
			if err := p.updateStatus(updateStatus); err != nil {
				panic(err)
			}
		case nodeRole := <-nodeRoleCh:
			newRole, ok := nodeRole.Value.(*pubsub.NodeRole)
			if ok {
				Logger.log.Infof("Receive new role %v at shard %v", newRole.Role, newRole.CID)
				if err := p.handleNewRole(newRole); err != nil {
					panic(err)
				}
			} else {
				Logger.log.Errorf("Cannot parse node role %v", *nodeRole)
			}
		case ec := <-p.TriggerCh:
			status := rawdbv2.ProcessingPruneByHeightStatus
			if ec.ShouldPruneByHash {
				status = rawdbv2.ProcessingPruneByHashStatus
			}
			if p.currentValidatorShardID <= common.BeaconChainID {
				p.triggerUpdateStatus(UpdateStatus{ShardID: ec.ShardID, Status: status})
			} else {
				if byte(p.currentValidatorShardID) != ec.ShardID {
					p.triggerUpdateStatus(UpdateStatus{ShardID: ec.ShardID, Status: status})
				} else {
					status := rawdbv2.WaitingPruneByHeightStatus
					if ec.ShouldPruneByHash {
						status = rawdbv2.WaitingPruneByHashStatus
					}
					p.triggerUpdateStatus(UpdateStatus{ShardID: ec.ShardID, Status: status})
				}
			}
		}
	}
}

func (p *Pruner) updateStatus(status UpdateStatus) error {
	p.statuses[status.ShardID] = status.Status
	err := rawdbv2.StorePruneStatus(p.db[int(status.ShardID)], status.Status)
	if err != nil {
		return err
	}

	if status.Status == rawdbv2.ProcessingPruneByHashStatus || status.Status == rawdbv2.ProcessingPruneByHeightStatus {
		shouldPruneByHash := status.Status == rawdbv2.ProcessingPruneByHashStatus
		if err := p.prune(int(status.ShardID), shouldPruneByHash, config.Config().StateBloomSize); err != nil {
			return err
		}
		p.statuses[status.ShardID] = rawdbv2.FinishPruneStatus
		err := rawdbv2.StorePruneStatus(p.db[int(status.ShardID)], rawdbv2.FinishPruneStatus)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pruner) handleNewRole(newRole *pubsub.NodeRole) error {
	if newRole.CID > common.BeaconChainID {
		if newRole.Role == common.CommitteeRole && !config.Config().ForcePrune {
			switch p.statuses[byte(newRole.CID)] {
			case rawdbv2.ProcessingPruneByHeightStatus, rawdbv2.WaitingPruneByHeightStatus, rawdbv2.ProcessingPruneByHashStatus, rawdbv2.WaitingPruneByHashStatus:
				p.triggerUpdateStatus(UpdateStatus{ShardID: byte(newRole.CID), Status: rawdbv2.FinishPruneStatus})
			}
		}
	}
	p.currentValidatorShardID = newRole.CID
	return nil
}

func (p *Pruner) handleNewView(shardBestState *blockchain.ShardBestState) error {
	status := p.statuses[shardBestState.ShardID]
	if common.CalculateTimeSlot(time.Now().Unix()) == common.CalculateTimeSlot(shardBestState.BestBlock.GetProduceTime()) || config.Config().ForcePrune {
		if status == rawdbv2.ProcessingPruneByHashStatus || status == rawdbv2.ProcessingPruneByHeightStatus {
			Logger.log.Infof("Process new view %s at shard %v", shardBestState.BestBlockHash.String(), shardBestState.ShardID)
			p.bestView.Store(shardBestState.ShardID, shardBestState)
			err := p.addNewViewToStateBloom(shardBestState, p.db[int(shardBestState.ShardID)])
			if err != nil {
				panic(err)
			}
		}
		if status == rawdbv2.WaitingPruneByHashStatus || status == rawdbv2.WaitingPruneByHeightStatus {
			s := rawdbv2.ProcessingPruneByHeightStatus
			if status == rawdbv2.WaitingPruneByHashStatus {
				s = rawdbv2.ProcessingPruneByHashStatus
			}
			p.triggerUpdateStatus(UpdateStatus{ShardID: shardBestState.ShardID, Status: s})
		}
	}
	return nil
}

func (p *Pruner) triggerUpdateStatus(status UpdateStatus) {
	go func() {
		p.updateStatusCh <- status
	}()
}
