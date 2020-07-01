package rawdbv2

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

// StoreShardBlock store block hash => block value and block index => block hash
// record1: prefix-shardid-index-hash => empty
// record2: prefix-hash => block value
func StoreShardBlock(db incdb.KeyValueWriter, hash common.Hash, v interface{}) error {
	keyHash := GetShardHashToBlockKey(hash)
	val, err := json.Marshal(v)
	if err != nil {
		return NewRawdbError(StoreShardBlockError, err)
	}
	if err := db.Put(keyHash, val); err != nil {
		return NewRawdbError(StoreShardBlockError, err)
	}
	return nil
}

func HasShardBlock(db incdb.KeyValueReader, hash common.Hash) (bool, error) {
	keyHash := GetShardHashToBlockKey(hash)
	if ok, err := db.Has(keyHash); err != nil {
		return false, NewRawdbError(HasShardBlockError, fmt.Errorf("has key %+v failed", keyHash))
	} else if ok {
		return true, nil
	}
	return false, nil
}

func GetShardBlockByHash(db incdb.KeyValueReader, hash common.Hash) ([]byte, error) {
	keyHash := GetShardHashToBlockKey(hash)
	if ok, err := db.Has(keyHash); err != nil {
		return []byte{}, NewRawdbError(GetShardBlockByHashError, fmt.Errorf("has key %+v failed", keyHash))
	} else if !ok {
		return []byte{}, NewRawdbError(GetShardBlockByHashError, fmt.Errorf("block %+v not exist", hash))
	}
	block, err := db.Get(keyHash)
	if err != nil {
		return nil, NewRawdbError(GetShardBlockByHashError, err)
	}
	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}

func GetIndexOfBlock(db incdb.KeyValueReader, hash common.Hash) (uint64, byte, error) {
	var index uint64
	var shardID byte
	key := GetShardBlockHashToIndexKey(hash)
	value, err := db.Get(key)
	if err != nil {
		return index, shardID, NewRawdbError(GetIndexOfBlockError, err)
	}
	if err := binary.Read(bytes.NewReader(value[:8]), binary.LittleEndian, &index); err != nil {
		return 0, 0, NewRawdbError(GetIndexOfBlockError, err)
	}
	if err = binary.Read(bytes.NewReader(value[8:]), binary.LittleEndian, &shardID); err != nil {
		return 0, 0, NewRawdbError(GetIndexOfBlockError, err)
	}
	return index, shardID, nil
}

func StoreShardBestState(db incdb.KeyValueWriter, shardID byte, v interface{}) error {
	key := GetShardBestStateKey(shardID)
	val, err := json.Marshal(v)
	if err != nil {
		return NewRawdbError(StoreShardBestStateError, err)
	}
	if err := db.Put(key, val); err != nil {
		return NewRawdbError(StoreShardBestStateError, err)
	}
	return nil
}

func GetShardBestState(db incdb.KeyValueReader, shardID byte) ([]byte, error) {
	key := GetShardBestStateKey(shardID)
	shardBestStateBytes, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(StoreShardBestStateError, err)
	}
	return shardBestStateBytes, nil
}

// StoreFeeEstimator - Store data for FeeEstimator object
func StoreFeeEstimator(db incdb.KeyValueWriter, val []byte, shardID byte) error {
	key := GetFeeEstimatorPrefix(shardID)
	err := db.Put(key, val)
	if err != nil {
		return NewRawdbError(StoreFeeEstimatorError, err)
	}
	return nil
}

// GetFeeEstimator - Get data for FeeEstimator object as a json in byte format
func GetFeeEstimator(db incdb.KeyValueReader, shardID byte) ([]byte, error) {
	key := GetFeeEstimatorPrefix(shardID)
	res, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(GetFeeEstimatorError, err)
	}
	return res, err
}
func StoreShardRootsHash(db incdb.KeyValueWriter, shardID byte, hash common.Hash, rootsHash interface{}) error {
	key := GetShardRootsHashKey(shardID, hash)
	b, _ := json.Marshal(rootsHash)
	err := db.Put(key, b)
	if err != nil {
		return NewRawdbError(StoreShardConsensusRootHashError, err)
	}
	return nil
}

func GetShardRootsHash(db incdb.KeyValueReader, shardID byte, hash common.Hash) ([]byte, error) {
	key := GetShardRootsHashKey(shardID, hash)
	return db.Get(key)
}
