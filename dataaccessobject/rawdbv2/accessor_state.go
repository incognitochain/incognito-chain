package rawdbv2

import (
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

func StoreCommitteeChangeCheckpoint(db incdb.KeyValueWriter, data []byte) error {
	key := GetCommitteeCheckpointKey()
	if err := db.Put(key, data); err != nil {
		return NewRawdbError(StoreShardBestStateError, err)
	}
	return nil
}

func GetCommitteeChangeCheckpoint(db incdb.KeyValueReader) ([]byte, error) {
	key := GetCommitteeCheckpointKey()
	chkPntBytes, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	return chkPntBytes, nil
}

func GetDatabaseConfig(db incdb.KeyValueReader) ([]byte, error) {
	key := GetDatabaseConfigFromDBKey()
	dbConfigBytes, err := db.Get(key)
	if err != nil {
		return nil, err
	}
	return dbConfigBytes, nil
}

func StoreDatabaseConfig(db incdb.KeyValueWriter, data []byte) error {
	key := GetDatabaseConfigFromDBKey()
	return db.Put(key, data)
}
