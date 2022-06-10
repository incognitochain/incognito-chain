package statedb

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

func (stateDB *StateDB) Retrieve(
	db incdb.Database, shouldAddToStateBloom bool, shouldDelete bool, stateBloom *trie.StateBloom,
) (int, int, error) {
	temp := stateDB.trie.NodeIterator(nil)
	it := trie.NewIterator(temp)

	keyShouldBeDeleted := make(map[common.Hash]struct{})
	var totalSize int
	keys := make(map[common.Hash]struct{})
	for it.Next(false) {
		if len(it.Key) == 0 {
			continue
		}
		key := it.Key
		h := common.Hash{}
		err := h.SetBytes(key)
		if err != nil {
			return 0, 0, err
		}
		if shouldAddToStateBloom {
			keys[h] = struct{}{}
			temp, _ := db.Get(key)
			totalSize += len(temp) + len(key)
			if err := stateBloom.Put(key, nil); err != nil {
				return 0, 0, err
			}
		}
		if shouldDelete {
			if ok, err := stateBloom.Contain(key); err != nil {
				return 0, 0, err
			} else if ok {
				continue
			}
			keyShouldBeDeleted[h] = struct{}{}
		}
	}
	if shouldDelete && len(keyShouldBeDeleted) != 0 {
		batch := db.NewBatch()
		for key := range keyShouldBeDeleted {
			temp, _ := db.Get(key.Bytes())
			totalSize += len(temp) + len(key.Bytes())
			if err := batch.Delete(key.Bytes()); err != nil {
				return 0, 0, err
			}
			if batch.ValueSize() >= incdb.IdealBatchSize {
				if err := batch.Write(); err != nil {
					return 0, 0, err
				}
				batch.Reset()
			}
		}
		if batch.ValueSize() > 0 {
			batch.Write()
			batch.Reset()
		}
	}
	if shouldAddToStateBloom {
		fmt.Printf("[state-prune] totalNodes %v data size %v\n", len(keys), totalSize)
	}

	return len(keyShouldBeDeleted), totalSize, nil
}
