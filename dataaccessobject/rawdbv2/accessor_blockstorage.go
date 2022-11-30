package rawdbv2

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

// store block hash => block value
func StoreFlatFileIndexByBlockHash(db incdb.KeyValueWriter, hash common.Hash, index uint64) error {
	keyHash := GetBlockHashToFFIndexKey(hash)
	buf := common.Uint64ToBytes(index)
	if err := db.Put(keyHash, buf); err != nil {
		return NewRawdbError(StoreFFIndexError, err)
	}
	return nil
}

func GetFlatFileIndexByBlockHash(db incdb.KeyValueReader, hash common.Hash) (uint64, error) {
	keyHash := GetBlockHashToFFIndexKey(hash)
	if data, err := db.Get(keyHash); err != nil {
		return 0, NewRawdbError(StoreFFIndexError, err)
	} else {
		return common.BytesToUint64(data)
	}

}

// store block hash => validation data
func StoreValidationDataByBlockHash(db incdb.KeyValueWriter, hash common.Hash, val []byte) error {
	keyHash := GetBlockHashToValidationDataKey(hash)
	if err := db.Put(keyHash, val); err != nil {
		return NewRawdbError(StoreFFIndexError, err)
	}
	return nil
}

func GetValidationDataByBlockHash(db incdb.KeyValueReader, hash common.Hash) ([]byte, error) {
	keyHash := GetBlockHashToValidationDataKey(hash)
	if data, err := db.Get(keyHash); err != nil {
		return nil, NewRawdbError(StoreFFIndexError, err)
	} else {
		return data, nil
	}

}
