package pruner

import (
	"encoding/json"
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

type Worker struct {
	stopCh                  chan interface{}
	heightCh                chan uint64
	rootHashCh              chan blockchain.ShardRootHash
	rootHashCache           *lru.Cache
	listKeysShouldBeRemoved *[]map[common.Hash]struct{}
	db                      incdb.Database
	shardID                 byte
	stateBloom              *trie.StateBloom
	wg                      *sync.WaitGroup
	insertLock              *sync.Mutex
}

func NewWorker(
	stopCh chan interface{}, heightCh chan uint64, rootHashCache *lru.Cache, stateBloom *trie.StateBloom,
	listKeysShouldBeRemoved *[]map[common.Hash]struct{}, db incdb.Database, shardID byte, wg *sync.WaitGroup,
	insertLock *sync.Mutex,
) *Worker {
	return &Worker{
		stopCh:                  stopCh,
		heightCh:                heightCh,
		rootHashCache:           rootHashCache,
		stateBloom:              stateBloom,
		listKeysShouldBeRemoved: listKeysShouldBeRemoved,
		db:                      db,
		shardID:                 shardID,
		wg:                      wg,
		insertLock:              insertLock,
	}
}

func (w *Worker) stop() {
	w.stopCh <- struct{}{}
}

func (w *Worker) start() {
	for {
		select {
		case height := <-w.heightCh:
			err := w.pruneByHeight(height)
			if err != nil {
				panic(err)
			}
		case hash := <-w.rootHashCh:
			err := w.pruneByHash(hash)
			if err != nil {
				panic(err)
			}
		case <-w.stopCh:
			return
		}
	}
}

func (w *Worker) pruneByHash(rootHash blockchain.ShardRootHash) error {
	defer func() {
		w.wg.Done()
	}()
	if _, ok := w.rootHashCache.Get(rootHash.TransactionStateDBRootHash.String()); ok {
		return nil
	}
	err := w.pruneTxStateDB(&rootHash)
	if err != nil {
		return err
	}
	w.rootHashCache.Add(rootHash.TransactionStateDBRootHash.String(), struct{}{})
	return nil
}

func (w *Worker) pruneByHeight(height uint64) error {
	defer func() {
		w.wg.Done()
	}()
	h, err := rawdbv2.GetFinalizedShardBlockHashByIndex(w.db, w.shardID, height)
	if err != nil {
		return err
	}
	data, err := rawdbv2.GetShardRootsHash(w.db, w.shardID, *h)
	sRH := &blockchain.ShardRootHash{}
	if err = json.Unmarshal(data, sRH); err != nil {
		return err
	}
	return w.pruneTxStateDB(sRH)
}

func (w *Worker) pruneTxStateDB(sRH *blockchain.ShardRootHash) error {
	sDB, err := statedb.NewWithPrefixTrie(sRH.TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(w.db))
	if err != nil {
		return nil
	}
	keysShouldBeRemoved, _, err := sDB.Retrieve(false, true, w.stateBloom)
	if err != nil {
		return err
	}
	if len(keysShouldBeRemoved) != 0 {
		*w.listKeysShouldBeRemoved = append(*w.listKeysShouldBeRemoved, keysShouldBeRemoved)
	}
	return nil
}
