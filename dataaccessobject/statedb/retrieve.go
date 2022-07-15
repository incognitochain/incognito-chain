package statedb

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/trie"
)

func (stateDB *StateDB) Recheck() error {
	temp := stateDB.trie.NodeIterator(nil)
	it := trie.NewIterator(temp)
	cnt := 0
	descend := true
	for it.Next(true, descend, true) {
		cnt++
		if cnt%100000 == 0 {
			fmt.Println(cnt)
		}
		if len(it.Key) == 0 {
			continue
		}
	}
	fmt.Println("Total node check:", cnt)
	return it.Err
}

func (stateDB *StateDB) Retrieve(
	shouldAddToStateBloom bool, shouldDelete bool, stateBloom *trie.StateBloom,
) (map[common.Hash]struct{}, *trie.StateBloom, error) {
	temp := stateDB.trie.NodeIterator(nil)
	it := trie.NewIterator(temp)
	keysShouldBeRemoved := make(map[common.Hash]struct{})
	cnt := 0
	descend := true
	returnErr := false
	if shouldAddToStateBloom {
		returnErr = true
	}
	for it.Next(false, descend, returnErr) {
		cnt++
		if cnt%100000 == 0 {
			fmt.Println(cnt)
		}
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
	if shouldAddToStateBloom && it.Err != nil {
		fmt.Println(it.Err)
		panic(1)
	}
	return keysShouldBeRemoved, stateBloom, nil
}
