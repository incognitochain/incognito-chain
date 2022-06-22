package statedb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/trie"
)

func (stateDB *StateDB) Retrieve(
	shouldAddToStateBloom bool, shouldDelete bool, stateBloom *trie.StateBloom,
) (map[common.Hash]struct{}, *trie.StateBloom, error) {
	temp := stateDB.trie.NodeIterator(nil)
	it := trie.NewIterator(temp)
	keysShouldBeRemoved := make(map[common.Hash]struct{})

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
			return nil, stateBloom, err
		}
		if shouldAddToStateBloom || shouldDelete {
			if ok, err := stateBloom.Contain(key); err != nil {
				return nil, stateBloom, err
			} else if ok {
				descend = false
				continue
			}
		}
		if shouldAddToStateBloom {
			if err := stateBloom.Put(key, nil); err != nil {
				return nil, stateBloom, err
			}
		}
		if shouldDelete {
			keysShouldBeRemoved[h] = struct{}{}
		}
	}

	return keysShouldBeRemoved, stateBloom, nil
}
