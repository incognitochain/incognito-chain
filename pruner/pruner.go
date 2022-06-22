package pruner

import (
	"encoding/json"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
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
	ForwardCh               chan ExtendedConfig
	insertLock              *sync.Mutex
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
		ForwardCh:               make(chan ExtendedConfig, 1),
		currentValidatorShardID: -2,
		insertLock:              new(sync.Mutex),
		addedViewsCache:         make(map[common.Hash]struct{}),
	}
}

func (p *Pruner) ReadStatus() {
	for i := 0; i < common.MaxShardNumber; i++ {
		status, _ := rawdbv2.GetPruneStatus(p.db[i], byte(i)) //ignore error for case not store status yet
		p.statuses[byte(i)] = status
	}
}

func (p *Pruner) PruneImmediately() error {
	for i := 0; i < common.MaxShardNumber; i++ {
		if err := p.Prune(i, false); err != nil {
			panic(err)
		}
	}
	return nil
}

func (p *Pruner) Prune(sID int, shouldPruneByHash bool) error {
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
	p.stateBloom, err = trie.NewStateBloomWithSize(2048)
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
		worker := NewWorker(stopCh, heightCh, rootHashCache, p.stateBloom, listKeysShouldBeRemoved, db, shardID, wg, p.insertLock)
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
	err = rawdbv2.StorePendingPrunedNodes(db, shardID, 0)
	if err != nil {
		return err
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
	lastPrunedKey, err := rawdbv2.GetLastPrunedKeyTrie(helper.db, helper.shardID)
	var start []byte
	if len(lastPrunedKey) != 0 {
		start = lastPrunedKey
	}
	Logger.log.Infof("[state-prune] begin pruning from key %v", lastPrunedKey)
	nodes, _ = rawdbv2.GetPendingPrunedNodes(helper.db, helper.shardID) // not checking error avoid case not store pruned node yet
	iter := helper.db.NewIteratorWithPrefixStart(rawdbv2.GetShardRootsHashPrefix(helper.shardID), start)
	defer func() {
		iter.Release()
	}()
	var finalPrunedKey []byte

	// retrieve all state tree by shard rooth hash prefix
	// delete all nodes which are not in state bloom
	for iter.Next() {
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
	var nodes, storage, count uint64
	// get last pruned height before
	lastPrunedHeight, err := rawdbv2.GetLastPrunedHeight(helper.db, helper.shardID)
	if err == nil {
		return 0, 0, err
	}
	if lastPrunedHeight == 0 {
		lastPrunedHeight = 1
	} else {
		lastPrunedHeight++
	}
	for height := lastPrunedHeight; height < helper.finalHeight; height++ {
		p.insertLock.Lock()
		helper.wg.Add(1)
		helper.heightCh <- height
		helper.wg.Wait()
		nodes, storage, err = p.removeNodes(helper.db, helper.shardID, nil, height, listKeysShouldBeRemoved, nodes, storage, false)
		if err != nil {
			return 0, 0, err
		}
		if count%10000 == 0 {
			Logger.log.Infof("[state-prune] Finish prune for height %v delete totalNodes %v with storage %v", height, nodes, storage)
		}
		count++
		p.insertLock.Unlock()
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
		if err := rawdbv2.StoreLastPrunedKeyTrie(db, shardID, key); err != nil {
			return 0, 0, err
		}
		if err := rawdbv2.StorePendingPrunedNodes(db, shardID, totalNodes); err != nil {
			return 0, 0, err
		}
	} else {
		if err := rawdbv2.StoreLastPrunedHeight(db, shardID, height); err != nil {
			return 0, 0, err
		}
	}

	*listKeysShouldBeRemoved = make([]map[common.Hash]struct{}, 0)
	return totalNodes, totalStorage, nil
}

func (p *Pruner) Start() {
	_, nodeRoleCh, err := p.PubSubManager.RegisterNewSubscriber(pubsub.NodeRoleDetailTopic)
	if err != nil {
		panic(err)
	}
	_, newShardBestStateCh, err := p.PubSubManager.RegisterNewSubscriber(pubsub.ShardBeststateTopic)
	if err != nil {
		panic(err)
	}

	for {
		select {
		case updateStatus := <-p.updateStatusCh:
			p.statuses[updateStatus.ShardID] = updateStatus.Status
			err := rawdbv2.StorePruneStatus(p.db[int(updateStatus.ShardID)], byte(updateStatus.ShardID), updateStatus.Status)
			if err != nil {
				panic(err)
			}
			if updateStatus.Status == rawdbv2.ProcessingPruneStatus {
				if err := p.Prune(int(updateStatus.ShardID), updateStatus.ShouldPruneByHash); err != nil {
					panic(err)
				}
				p.statuses[updateStatus.ShardID] = rawdbv2.FinishPruneStatus
				err := rawdbv2.StorePruneStatus(p.db[int(updateStatus.ShardID)], byte(updateStatus.ShardID), rawdbv2.FinishPruneStatus)
				if err != nil {
					panic(err)
				}
			}
		case nodeRole := <-nodeRoleCh:
			newRole, ok := nodeRole.Value.(*pubsub.NodeRole)
			if ok {
				Logger.log.Infof("Receive new role %v at shard %v", newRole.Role, newRole.CID)
				if newRole.CID > common.BeaconChainID {
					if newRole.Role == common.CommitteeRole {
						switch p.statuses[byte(newRole.CID)] {
						case rawdbv2.ProcessingPruneStatus, rawdbv2.WaitingPruneByHashStatus, rawdbv2.WaitingPruneByHeightStatus:
							p.updateStatusCh <- UpdateStatus{ShardID: byte(newRole.CID), Status: rawdbv2.PausePruneStatus}
						}
					}
					p.currentValidatorShardID = newRole.CID
				}
			} else {
				Logger.log.Errorf("Cannot parse node role %v", *nodeRole)
			}
		case newShardBestState := <-newShardBestStateCh:
			// if new insert block reach to current
			shardBestState, ok := newShardBestState.Value.(*blockchain.ShardBestState)
			if ok {
				status := p.statuses[shardBestState.ShardID]
				if common.CalculateTimeSlot(time.Now().Unix()) == common.CalculateTimeSlot(shardBestState.BestBlock.GetProduceTime()) {
					if status == rawdbv2.ProcessingPruneStatus {
						p.insertLock.Lock()
						err = p.addNewViewToStateBloom(shardBestState, p.db[int(shardBestState.ShardID)])
						if err != nil {
							panic(err)
						}
						p.insertLock.Unlock()
					}
					if status == rawdbv2.WaitingPruneByHashStatus || status == rawdbv2.WaitingPruneByHeightStatus {
						shouldPruneByHash := status == rawdbv2.WaitingPruneByHashStatus
						p.updateStatusCh <- UpdateStatus{ShardID: shardBestState.ShardID, Status: rawdbv2.ProcessingPruneStatus, ShouldPruneByHash: shouldPruneByHash}
					}
				}
			} else {
				Logger.log.Errorf("Cannot parse newShardBestState %v", newShardBestState)
			}
		case ec := <-p.ForwardCh:
			if p.currentValidatorShardID <= common.BeaconChainID {
				p.updateStatusCh <- UpdateStatus{ShardID: ec.ShardID, Status: rawdbv2.ProcessingPruneStatus}
			} else {
				if byte(p.currentValidatorShardID) != ec.ShardID {
					p.updateStatusCh <- UpdateStatus{ShardID: ec.ShardID, Status: rawdbv2.ProcessingPruneStatus}
				} else {
					status := rawdbv2.WaitingPruneByHeightStatus
					if ec.ShouldPruneByHash {
						status = rawdbv2.WaitingPruneByHashStatus
					}
					p.updateStatusCh <- UpdateStatus{ShardID: ec.ShardID, Status: status}
				}
			}
		}
	}
}

func (p *Pruner) reset() {
	p.stateBloom = nil
	p.insertLock = new(sync.Mutex)
	p.addedViewsCache = make(map[common.Hash]struct{})
}
