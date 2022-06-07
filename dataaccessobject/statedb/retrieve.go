package statedb

import "github.com/incognitochain/incognito-chain/trie"

func (stateDB *StateDB) Retrieve(shouldAddToStateBloom bool, shouldDelete bool, stateBloom *trie.StateBloom) error {
	temp := stateDB.trie.NodeIterator(nil)
	it := trie.NewIterator(temp)
	for it.Next(false) {
		if shouldAddToStateBloom {

		}
		if shouldDelete {

		}
	}
	return nil
}
