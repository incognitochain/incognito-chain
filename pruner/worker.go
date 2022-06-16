package pruner

import (
	"fmt"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

type Worker struct {
	stopCh              chan interface{}
	rootHashCh          chan blockchain.ShardRootHash
	keysShouldBeRemoved map[common.Hash]struct{}
	db                  incdb.Database
	shardID             byte
	stateBloom          *trie.StateBloom
	wg                  *sync.WaitGroup
}

func NewWorker(
	stopCh chan interface{}, rootHashCh chan blockchain.ShardRootHash, stateBloom *trie.StateBloom, db incdb.Database, shardID byte, wg *sync.WaitGroup,
) *Worker {
	return &Worker{
		stopCh:              stopCh,
		rootHashCh:          rootHashCh,
		stateBloom:          stateBloom,
		keysShouldBeRemoved: make(map[common.Hash]struct{}),
		db:                  db,
		shardID:             shardID,
		wg:                  wg,
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
			w.wg.Done()
			if err != nil {
				panic(err)
			}
		case <-w.stopCh:
			return
		}
	}
}

func (w *Worker) removeDB() (uint64, uint64, error) {
	var storage = uint64(0)
	var count = uint64(0)

	batch := w.db.NewBatch()
	fmt.Println("remove", len(w.keysShouldBeRemoved))
	for key := range w.keysShouldBeRemoved {
		temp, _ := w.db.Get(key.Bytes())
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

	w.keysShouldBeRemoved = make(map[common.Hash]struct{})
	return count, storage, nil
}

func (w *Worker) pruneByRootHash(rootHash blockchain.ShardRootHash) error {
	err := w.pruneTxStateDB(&rootHash)
	if err != nil {
		return err
	}
	return nil
}

func (w *Worker) pruneTxStateDB(sRH *blockchain.ShardRootHash) error {
	sDB, err := statedb.NewWithPrefixTrie(sRH.TransactionStateDBRootHash, statedb.NewDatabaseAccessWarper(w.db))
	if err != nil {
		return nil
	}
	err = sDB.Retrieve(false, true, w.stateBloom, w.keysShouldBeRemoved)
	if err != nil {
		return err
	}
	return nil
}
