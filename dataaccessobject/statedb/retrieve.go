package statedb

import (
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/trie"
)

func (stateDB *StateDB) Retrieve(
	shouldAddToStateBloom bool, shouldDelete bool,
	stateBloom *trie.StateBloom, keyShouldBeRemoved map[common.Hash]struct{},
	mu *sync.RWMutex,
) (map[common.Hash]struct{}, error) {
	temp := stateDB.trie.NodeIterator(nil)
	it := trie.NewIterator(temp)

	descend := true
	for it.Next(false, descend, false) {
		descend = true
		if len(it.Key) == 0 {
			continue
		}
		key := it.Key
		h := common.Hash{}
		err := h.SetBytes(key)
		if err != nil {
			return keyShouldBeRemoved, err
		}
		if shouldAddToStateBloom {
			if err := stateBloom.Put(key, nil); err != nil {
				return keyShouldBeRemoved, err
			}
		}
		if shouldDelete {
			if ok, err := stateBloom.Contain(key); err != nil {
				return keyShouldBeRemoved, err
			} else if ok {
				descend = false
				continue
			}
			mu.Lock()
			keyShouldBeRemoved[h] = struct{}{}
			mu.Unlock()
		}
	}

	return keyShouldBeRemoved, nil
}
