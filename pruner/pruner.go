package pruner

import (
	"encoding/json"
	"runtime"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

type Pruner struct {
	db                 map[int]incdb.Database
	quitC              chan error
	workerWG           sync.WaitGroup
	cleanWG            sync.WaitGroup
	finalHeight        uint64
	totalRemoveNodes   uint64
	totalRemoveStorage uint64
	keyShouldBeRemoved map[common.Hash]struct{}
	removeMu           *sync.RWMutex
}

func (p *Pruner) reset() {
	p.finalHeight = 0
	p.totalRemoveNodes = 0
	p.totalRemoveStorage = 0
	p.keyShouldBeRemoved = make(map[common.Hash]struct{})
}

func NewPrunerWithValue(db map[int]incdb.Database) *Pruner {
	return &Pruner{
		db:                 db,
		quitC:              make(chan error),
		keyShouldBeRemoved: make(map[common.Hash]struct{}),
		removeMu:           new(sync.RWMutex),
	}
}

func (p *Pruner) PruneImmediately() error {
	for i := 0; i < common.MaxShardNumber; i++ {
		if err := p.Prune(i); err != nil {
			panic(err)
		}
	}
	return nil
}

func (p *Pruner) Prune(sID int) error {
	p.reset()
	stateBloom, err := trie.NewStateBloomWithSize(2048)
	if err != nil {
		panic(err)
	}
	shardID := byte(sID)
	db := p.db[int(shardID)]
	rootHashCache, err := lru.New(100)
	if err != nil {
		panic(err)
	}
	keptViewC := make(chan *blockchain.ShardBestState)
	removeC := make(chan uint64)
	cleanC := make(chan uint64, 1)
	go p.start(db, stateBloom, keptViewC, removeC, cleanC, shardID, rootHashCache)

	Logger.log.Infof("[state-prune] Start state pruning for shardID %v", sID)
	err = p.addDataToStateBloom(shardID, db, stateBloom, keptViewC)
	if err != nil {
		return err
	}
	if p.finalHeight == 0 {
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
	for height := lastPrunedHeight; height < p.finalHeight; height++ {
		p.workerWG.Add(1)
		removeC <- height
	}
	p.workerWG.Wait()
	p.removeMu.RLock()
	if len(p.keyShouldBeRemoved) != 0 {
		err := p.removeNodes(db, shardID, p.finalHeight-1)
		if err != nil {
			return err
		}
	}
	p.removeMu.RUnlock()
	Logger.log.Infof("[state-prune] Delete totalNodes %v with size %v", p.totalRemoveNodes, p.totalRemoveStorage)
	if err = p.compact(db, p.totalRemoveNodes); err != nil {
		return err
	}
	err = rawdbv2.StorePruneStatus(db, shardID, rawdbv2.FinishPruneStatus)
	if err != nil {
		return err
	}
	Logger.log.Infof("[state-prune] Finish state pruning for shard %v", sID)
	p.quitC <- nil
	return nil
}

func (p *Pruner) addDataToStateBloom(
	shardID byte, db incdb.Database, stateBloom *trie.StateBloom, keptViewC chan *blockchain.ShardBestState,
) error {
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
	//collect tree nodes want to keep, add them to state bloom
	for _, v := range allViews {
		p.workerWG.Add(1)
		keptViewC <- v
	}
	p.workerWG.Wait()

	return nil
}

func (p *Pruner) pruneByHeight(
	height uint64, db incdb.Database, shardID byte, stateBloom *trie.StateBloom,
	rootHashCache *lru.Cache, workers chan interface{}, cleanC chan uint64,
) error {
	var err error
	defer func() {
		<-workers
		if err != nil {
			p.quitC <- err
		}
		p.workerWG.Done()
	}()
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
	if height%10 == 0 {
		p.cleanWG.Add(1)
		cleanC <- height
	}
	p.cleanWG.Wait()

	err = p.pruneTxStateDB(sRH, db, stateBloom, rootHashCache)
	if err != nil {
		return err
	}
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

func (p *Pruner) pruneTxStateDB(sRH *blockchain.ShardRootHash, db incdb.Database, stateBloom *trie.StateBloom, rootHashCache *lru.Cache) error {
	if _, ok := rootHashCache.Get(sRH.TransactionStateDBRootHash.String()); ok {
		return nil
	}
	sDB, err := statedb.NewWithPrefixTrie(sRH.TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(db))
	if err != nil {
		return nil
	}
	if p.keyShouldBeRemoved, err = sDB.Retrieve(false, true, stateBloom, p.keyShouldBeRemoved, p.removeMu); err != nil {
		return err
	}
	rootHashCache.Add(sRH.TransactionStateDBRootHash.String(), struct{}{})
	return nil
}

func (p *Pruner) addDataToStateBloomBySingleView(
	v *blockchain.ShardBestState, db incdb.Database, stateBloom *trie.StateBloom, workers chan interface{},
) error {
	var err error
	defer func() {
		<-workers
		if err != nil {
			p.quitC <- err
		}
		p.workerWG.Done()
	}()
	if v.ShardHeight == 1 {
		return nil
	}
	if p.finalHeight == 0 || p.finalHeight > v.ShardHeight {
		p.finalHeight = v.ShardHeight
	}
	Logger.log.Infof("[state-prune] Start retrieve view %s", v.BestBlockHash.String())
	var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
	stateDB, err := statedb.NewWithPrefixTrie(v.TransactionStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}
	//Retrieve all state tree for this state
	_, err = stateDB.Retrieve(true, false, stateBloom, nil, nil)
	if err != nil {
		return err
	}
	Logger.log.Infof("[state-prune] Finish retrieve view %s", v.BestBlockHash.String())
	return nil
}

func (p *Pruner) start(
	db incdb.Database, stateBloom *trie.StateBloom,
	keptViewC chan *blockchain.ShardBestState, removeC chan uint64, cleanC chan uint64,
	shardID byte, rootHashCache *lru.Cache,
) {
	workers := make(chan interface{}, runtime.NumCPU()-1)
	for {
		select {
		case v := <-keptViewC:
			Logger.log.Infof("[state-prune] worker %v collect data at view height %v\n", len(workers), v.ShardHeight)
			workers <- struct{}{}
			go p.addDataToStateBloomBySingleView(v, db, stateBloom, workers)
		case height := <-removeC:
			//TODO: remove this log for production release
			Logger.log.Debugf("[state-prune] worker %v prepare remove data at view height %v\n", len(workers), height)
			workers <- struct{}{}
			go p.pruneByHeight(height, db, shardID, stateBloom, rootHashCache, workers, cleanC)
		case height := <-cleanC:
			Logger.log.Infof("[state-prune] worker %v remove data at view height %v\n", len(workers), height)
			workers <- struct{}{}
			err := p.removeNodes(db, shardID, height)
			if err != nil {
				p.quitC <- err
			}
			<-workers
		case err := <-p.quitC:
			if err != nil {
				panic(err)
			}
			return
		default:
			time.Sleep(time.Millisecond * 100)
		}
	}
}

// removeNodes after removeNodes keys map will be reset to empty value
func (p *Pruner) removeNodes(db incdb.Database, shardID byte, height uint64) error {
	var size uint64
	if len(p.keyShouldBeRemoved) != 0 {
		batch := db.NewBatch()
		p.removeMu.RLock()
		for key := range p.keyShouldBeRemoved {
			temp, _ := db.Get(key.Bytes())
			size += uint64(len(temp) + len(key.Bytes()))
			if err := batch.Delete(key.Bytes()); err != nil {
				return err
			}
			if batch.ValueSize() >= incdb.IdealBatchSize {
				if err := batch.Write(); err != nil {
					return err
				}
				batch.Reset()
			}
		}
		p.removeMu.RUnlock()
		if batch.ValueSize() > 0 {
			batch.Write()
			batch.Reset()
		}
	}
	p.totalRemoveStorage += uint64(size)
	p.removeMu.RLock()
	p.totalRemoveNodes += uint64(len(p.keyShouldBeRemoved))
	p.removeMu.RUnlock()
	if err := rawdbv2.StoreLastPrunedHeight(db, shardID, height); err != nil {
		return err
	}
	if height%10000 == 0 {
		Logger.log.Infof("[state-prune] Finish prune for height %v delete totalNodes %v with size %v", height, p.totalRemoveNodes, p.totalRemoveStorage)
	}
	p.removeMu.Lock()
	p.keyShouldBeRemoved = make(map[common.Hash]struct{})
	p.removeMu.Unlock()
	p.cleanWG.Done()
	return nil
}
