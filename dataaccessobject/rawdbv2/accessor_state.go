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

func StoreLatestPivotBlock(db incdb.KeyValueWriter, shardID byte, hash common.Hash) error {
	return db.Put(GetFullSyncPivotBlockKey(shardID), hash[:])
}

func HasLatestPivotBlock(db incdb.KeyValueReader, shardID byte) (bool, error) {
	return db.Has(GetFullSyncPivotBlockKey(shardID))
}

func GetLatestPivotBlock(db incdb.KeyValueReader, shardID byte) (common.Hash, error) {

	value, err := db.Get(GetFullSyncPivotBlockKey(shardID))
	if err != nil {
		return common.Hash{}, err
	}

	h, err := common.Hash{}.NewHash(value)

	return *h, err
}

func StoreFlatFileStateObjectIndex(db incdb.KeyValueWriter, hash common.Hash, indexes [][]int) error {

	key := GetFlatFileStateObjectIndexKey(hash)

	value, err := json.Marshal(indexes)
	if err != nil {
		return err
	}

	if err := db.Put(key, value); err != nil {
		return err
	}

	return nil
}

func GetFlatFileStateObjectIndex(db incdb.KeyValueReader, hash common.Hash) ([][]int, error) {

	indexes := [][]int{}
	key := GetFlatFileStateObjectIndexKey(hash)

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
	return db.Has(GetFlatFileStateObjectIndexKey(hash))
}

func DeleteFlatFileTransactionIndex(db incdb.KeyValueWriter, hash common.Hash) error {
	return db.Delete(GetFlatFileStateObjectIndexKey(hash))
}
