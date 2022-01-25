package rawdbv2

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject"
	"github.com/incognitochain/incognito-chain/incdb"
)

// ReadPreimage retrieves a single preimage of the provided hash.
func ReadPreimage(db incdb.KeyValueReader, hash common.Hash) []byte {
	data, _ := db.Get(preimageKey(hash))
	return data
}

// WritePreimages writes the provided set of preimages to the database.
func WritePreimages(db incdb.KeyValueWriter, preimages map[common.Hash][]byte) {
	for hash, preimage := range preimages {
		if err := db.Put(preimageKey(hash), preimage); err != nil {
			dataaccessobject.Logger.Log.Critical("Failed to store trie preimage", "err", err)
		}
	}
}

// ReadTrieNode retrieves the trie node of the provided hash.
func ReadTrieNode(db incdb.KeyValueReader, hash common.Hash) []byte {
	data, _ := db.Get(hash.Bytes())
	return data
}

// WriteTrieNode writes the provided trie node database.
func WriteTrieNode(db incdb.KeyValueWriter, hash common.Hash, node []byte) {
	if err := db.Put(hash.Bytes(), node); err != nil {
		dataaccessobject.Logger.Log.Critical("Failed to store trie node", "err", err)
	}
}

// DeleteTrieNode deletes the specified trie node from the database.
func DeleteTrieNode(db incdb.KeyValueWriter, hash common.Hash) {
	if err := db.Delete(hash.Bytes()); err != nil {
		dataaccessobject.Logger.Log.Critical("Failed to delete trie node", "err", err)
	}
}

func StoreFlatFileTransactionIndex(db incdb.KeyValueWriter, hash common.Hash, indexes []int) error {

	key := GetFlatFileTransactionIndexKey(hash)

	value, err := json.Marshal(indexes)
	if err != nil {
		return err
	}

	if err := db.Put(key, value); err != nil {
		return err
	}

	return nil
}

func GetFlatFileTransactionIndex(db incdb.KeyValueReader, hash common.Hash) ([]int, error) {

	indexes := []int{}
	key := GetFlatFileTransactionIndexKey(hash)

	value, err := db.Get(key)
	if err != nil {
		return indexes, err
	}

	if err := json.Unmarshal(value, &indexes); err != nil {
		return indexes, err
	}

	return indexes, nil
}

func HasFlatFileTransactionIndex(db incdb.KeyValueReader, hash common.Hash) (bool, error) {
	return db.Has(GetFlatFileTransactionIndexKey(hash))
}

func DeleteFlatFileTransactionIndex(db incdb.KeyValueWriter, hash common.Hash) error {
	return db.Delete(GetFlatFileTransactionIndexKey(hash))
}

func StoreLatestPivotBlock(db incdb.KeyValueWriter, hash common.Hash) error {
	return db.Put(fullSyncPivotBlockKey, hash[:])
}

func GetLatestPivotBlock(db incdb.KeyValueReader) (common.Hash, error) {

	value, err := db.Get(fullSyncPivotBlockKey)
	if err != nil {
		return common.Hash{}, err
	}

	h, err := common.Hash{}.NewHash(value)

	return *h, err
}
