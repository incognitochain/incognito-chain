package pruner

import (
	"encoding/json"
	"math/big"
	"path/filepath"
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
	db                      map[int]incdb.Database
	statuses                map[byte]byte
	updateStatusCh          chan UpdateStatus
	TriggerCh               chan ExtendedConfig
	shardInsertLock         map[int]*sync.Mutex
	wg                      *sync.WaitGroup
	PubSubManager           *pubsub.PubSubManager
	currentValidatorShardID int
	stateBloom              *trie.StateBloom
	addedViewsCache         map[common.Hash]struct{}
}

func NewPrunerWithValue(db map[int]incdb.Database, statuses map[byte]byte) *Pruner {
	return &Pruner{
		db:                      db,
		statuses:                statuses,
		updateStatusCh:          make(chan UpdateStatus),
		TriggerCh:               make(chan ExtendedConfig, 1),
		shardInsertLock:         make(map[int]*sync.Mutex),
		currentValidatorShardID: -2,
		wg:                      new(sync.WaitGroup),
		addedViewsCache:         make(map[common.Hash]struct{}),
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
	ch := make(chan int, cpus-1)
	stateBloomSize := config.Config().StateBloomSize / uint64(cpus)
	stopCh := make(chan struct{})
	var count int
	var wg sync.WaitGroup
	go func() {
		for {
			select {
			case shardID := <-ch:
				if err := p.prune(shardID, false, stateBloomSize); err != nil {
					panic(err)
				}
				wg.Done()
			case <-stopCh:
				count++
				if count == common.MaxShardNumber {
					return
				}
			}
		}
	}()
	for i := 0; i < common.MaxShardNumber; i++ {
		wg.Add(1)
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
	p.stateBloom, err = trie.NewStateBloomWithSize(stateBloomSize)
	if err != nil {
		panic(err)
	}
	finalHeight, err := p.addDataToStateBloom(shardID, db)
	if err != nil {
		return err
	}
	if finalHeight == 0 {
		return nil
	}
	stopCh := make(chan interface{})
	heightCh := make(chan uint64)
	rootHashCh := make(chan blockchain.ShardRootHash)

	listKeysShouldBeRemoved := &[]map[common.Hash]struct{}{}
	wg := new(sync.WaitGroup)
	for i := 0; i < 1; i++ {
		worker := NewWorker(stopCh, heightCh, rootHashCache, p.stateBloom, listKeysShouldBeRemoved, db, shardID, wg)
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
	newSize, err := common.DirSize(filepath.Join(config.Config().DataDir, config.Config().DatabaseDir))
	if err != nil {
		panic(err)
	}
	if err := rawdbv2.StoreDataSize(db, uint64(newSize)); err != nil {
		panic(err)
	}
	p.reset()
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
	for _, v := range allViews {
		if finalHeight == 0 || finalHeight > v.ShardHeight {
			finalHeight = v.ShardHeight
		}
		err = p.addNewViewToStateBloom(v, db)
		if err != nil {
			return 0, err
		}
	}

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
	Logger.log.Infof("[state-prune] Start retrieve view %s at height %v", v.BestBlockHash.String(), v.ShardHeight)
	var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
	stateDB, err := statedb.NewWithPrefixTrie(v.TransactionStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}
	//Retrieve all state tree for this state
	_, p.stateBloom, err = stateDB.Retrieve(true, false, p.stateBloom)
	if err != nil {
		return err
	}
	p.addedViewsCache[v.BestBlockHash] = struct{}{}
	Logger.log.Infof("[state-prune] Finish retrieve view %s at height %v", v.BestBlockHash.String(), v.ShardHeight)
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
	var nodes, storage uint64
	var err error
	if shouldPruneByHash {
		nodes, storage, err = p.traverseAndDeleteByHash(helper, listKeysShouldBeRemoved)
		if err != nil {
			return err
		}
		Logger.log.Infof("[state-prune] Start compact totalNodes %v with size %v", nodes, storage)
		if err = p.compact(helper.db, nodes); err != nil {
			return err
		}
		Logger.log.Infof("[state-prune] Finish compact totalNodes %v with size %v", nodes, storage)
	} else {
		nodes, storage, err = p.traverseAndDeleteByHeight(helper, listKeysShouldBeRemoved)
		if err != nil {
			return err
		}
	}
	Logger.log.Infof("[state-prune] Delete totalNodes %v with size %v", nodes, storage)

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
		if p.statuses[helper.shardID] == rawdbv2.FinishPruneStatus {
			return nodes, storage, nil
		}
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
	for height := lastPrunedHeight; height < helper.finalHeight; height++ {
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
		if height%10000 == 0 {
			Logger.log.Infof("[state-prune] Finish prune for height %v delete totalNodes %v with storage %v", height, nodes, storage)
		}

		if p.statuses[helper.shardID] == rawdbv2.FinishPruneStatus {
			return nodes, storage, nil
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
}

func (p *Pruner) Start() {
	_, nodeRoleCh, err := p.PubSubManager.RegisterNewSubscriber(pubsub.NodeRoleDetailTopic)
	if err != nil {
		panic(err)
	}

	go p.watchStorageChange()

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

func (p *Pruner) watchStorageChange() {
	if config.Config().EnableAutoPrune {
		for {
			for i := 0; i < common.MaxShardNumber; i++ {
				oldSize, _ := rawdbv2.GetDataSize(p.db[i]) //ignore error if active prune for the first time
				newSize, err := common.DirSize(filepath.Join(config.Config().DataDir, config.Config().DatabaseDir))
				if err != nil {
					panic(err)
				}
				t1 := big.NewInt(0).Mul(big.NewInt(newSize), big.NewInt(10000))
				t2 := big.NewInt(0).Mul(big.NewInt(0).SetUint64(oldSize), big.NewInt(12500))
				if t1.Cmp(t2) >= 0 || config.Config().ForcePrune {
					if p.statuses[byte(i)] != rawdbv2.ProcessingPruneByHashStatus && p.statuses[byte(i)] != rawdbv2.ProcessingPruneByHeightStatus {
						ec := ExtendedConfig{
							Config:  Config{ShouldPruneByHash: false},
							ShardID: byte(i),
						}
						p.TriggerCh <- ec
					}

				}
			}
			time.Sleep(time.Minute * 5)
		}
	}
}

func (p *Pruner) reset() {
	p.wg.Wait()
	p.stateBloom = nil
	p.addedViewsCache = make(map[common.Hash]struct{})
}
