package statedb

import (
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

func (stateDB *StateDB) Retrieve(db incdb.Database, shouldAddToStateBloom bool, shouldDelete bool, stateBloom *trie.StateBloom) error {
	temp := stateDB.trie.NodeIterator(nil)
	it := trie.NewIterator(temp)
	batch := db.NewBatch()
	for it.Next(false) {
		key := it.Key
		if shouldAddToStateBloom {
			if err := stateBloom.Put(key, nil); err != nil {
				return err
			}
		}
		if shouldDelete {
			if ok, err := stateBloom.Contain(key); err != nil {
				return err
			} else if ok {
				continue
			}
			if err := batch.Delete(key); err != nil {
				return err
			}
			if batch.ValueSize() >= incdb.IdealBatchSize {
				if err := batch.Write(); err != nil {
					return err
				}
				batch.Reset()
			}
		}
	}
	if batch.ValueSize() > 0 {
		batch.Write()
		batch.Reset()
	}
	return nil
}
