package pruner

import (
	"sync"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

type Worker struct {
	stopCh                  chan interface{}
	rootHashCh              chan blockchain.ShardRootHash
	rootHashCache           *lru.Cache
	listKeysShouldBeRemoved *[]map[common.Hash]struct{}
	db                      incdb.Database
	shardID                 byte
	stateBloom              *trie.StateBloom
	wg                      *sync.WaitGroup
}

func NewWorker(
	stopCh chan interface{}, rootHashCh chan blockchain.ShardRootHash, rootHashCache *lru.Cache, stateBloom *trie.StateBloom,
	listKeysShouldBeRemoved *[]map[common.Hash]struct{}, db incdb.Database, shardID byte, wg *sync.WaitGroup,
) *Worker {
	return &Worker{
		stopCh:                  stopCh,
		rootHashCh:              rootHashCh,
		rootHashCache:           rootHashCache,
		stateBloom:              stateBloom,
		listKeysShouldBeRemoved: listKeysShouldBeRemoved,
		db:                      db,
		shardID:                 shardID,
		wg:                      wg,
	}
}

func (w *Worker) stop() {
	w.stopCh <- struct{}{}
}

func (w *Worker) start() {
	for {
		select {
		case rootHash := <-w.rootHashCh:
			err := w.pruneByRootHash(rootHash)
			if err != nil {
				panic(err)
			}
		case <-w.stopCh:
			return
		}
	}
}

func (w *Worker) pruneByRootHash(rootHash blockchain.ShardRootHash) error {
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

func (w *Worker) pruneTxStateDB(sRH *blockchain.ShardRootHash) error {
	sDB, err := statedb.NewWithPrefixTrie(sRH.TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(w.db))
	if err != nil {
		return nil
	}
	keysShouldBeRemoved, err := sDB.Retrieve(false, true, w.stateBloom)
	if err != nil {
		return err
	}
	if len(keysShouldBeRemoved) != 0 {
		*w.listKeysShouldBeRemoved = append(*w.listKeysShouldBeRemoved, keysShouldBeRemoved)
	}
	return nil
}
