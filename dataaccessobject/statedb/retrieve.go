package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/trie"
)

func (stateDB *StateDB) Retrieve(
	shouldAddToStateBloom bool, shouldDelete bool, stateBloom *trie.StateBloom, keysShouldBeRemoved map[common.Hash]struct{},
) error {
	temp := stateDB.trie.NodeIterator(nil)
	it := trie.NewIterator(temp)

	descend := true
	cnt := 0

	for it.Next(false, descend, false) {
		cnt++
		descend = true
		if len(it.Key) == 0 {
			continue
		}
		key := it.Key
		h := common.Hash{}
		err := h.SetBytes(key)
		if err != nil {
			return err
		}
		if shouldAddToStateBloom || shouldDelete {
			if _, ok := keysShouldBeRemoved[h]; ok {
				descend = false
			}

			if ok, err := stateBloom.Contain(key); err != nil {
				return err
			} else if ok {
				descend = false
				continue
			}
		}
		if shouldAddToStateBloom {
			if cnt%100000 == 0 {
				fmt.Println(cnt)

			}
			if err := stateBloom.Put(key, nil); err != nil {
				return err
			}
		}

		if shouldDelete {
			keysShouldBeRemoved[h] = struct{}{}
		}
	}
	fmt.Println("keysShouldBeRemoved", len(keysShouldBeRemoved))
	return nil
}
