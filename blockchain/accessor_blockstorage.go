package rawdbv2

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

var (
	beaconConfirmShardBlockPrefix   = []byte("b-c-s" + string(splitter))
	blockHashToValidationDataPrefix = []byte("b-h-v-d" + string(splitter))
	blockHashToFFIndexPrefix        = []byte("b-h-ff-i" + string(splitter))
	splitter                        = []byte("-[-]-")
)

func GetBlockHashToFFIndexKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(blockHashToFFIndexPrefix))
	temp = append(temp, blockHashToFFIndexPrefix...)
	return append(temp, hash[:]...)
}

func GetBlockHashToValidationDataKey(hash common.Hash) []byte {
	temp := make([]byte, 0, len(blockHashToValidationDataPrefix))
	temp = append(temp, blockHashToValidationDataPrefix...)
	return append(temp, hash[:]...)
}

func GetBeaconConfirmShardBlockPrefix(shardID byte, index uint64) []byte {
	buf := common.Uint64ToBytes(index)
	temp := make([]byte, 0, len(beaconConfirmShardBlockPrefix))
	temp = append(temp, beaconConfirmShardBlockPrefix...)
	key := append(temp, shardID)
	key = append(key, splitter...)
	key = append(key, buf...)
	return key
}

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

//beacon confirm sid , heigh -> hash
func StoreBeaconConfirmInstantFinalityShardBlock(db incdb.KeyValueWriter, sid byte, index uint64, hash common.Hash) error {
	keyHash := GetBeaconConfirmShardBlockPrefix(sid, index)
	if err := db.Put(keyHash, hash.Bytes()); err != nil {
		return NewRawdbError(StoreShardBlockIndexError, err)
	}
	return nil
}

func HasBeaconConfirmInstantFinalityShardBlock(db incdb.KeyValueReader, sid byte, index uint64) (bool, error) {
	keyHash := GetBeaconConfirmShardBlockPrefix(sid, index)
	return db.Has(keyHash)
}

func GetBeaconConfirmInstantFinalityShardBlock(db incdb.KeyValueReader, sid byte, index uint64) (*common.Hash, error) {
	keyHash := GetBeaconConfirmShardBlockPrefix(sid, index)
	val, err := db.Get(keyHash)
	if err != nil {
		return nil, NewRawdbError(GetShardBlockByIndexError, err)
	}
	h, err := common.Hash{}.NewHash(val)
	if err != nil {
		return nil, NewRawdbError(GetShardBlockByIndexError, err)
	}
	return h, nil
}
