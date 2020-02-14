package rawdbv2

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

// StoreShardBlock store block hash => block value and block index => block hash
// record1: prefix-shardid-index-hash => empty
// record2: prefix-hash => block value
func StoreShardBlock(db incdb.Database, shardID byte, index uint64, hash common.Hash, v interface{}) error {
	keyHash := GetShardHashToBlockKey(hash)
	if ok, _ := db.Has(keyHash); ok {
		return NewRawdbError(StoreShardBlockError, fmt.Errorf("key %+v already exists", keyHash))
	}
	keyIndex := GetShardIndexToBlockHashKey(shardID, index, hash)
	if ok, _ := db.Has(keyIndex); ok {
		return NewRawdbError(StoreShardBlockError, fmt.Errorf("key %+v already exists", keyIndex))
	}
	val, err := json.Marshal(v)
	if err != nil {
		return NewRawdbError(StoreShardBlockError, err)
	}
	if err := db.Put(keyIndex, []byte{}); err != nil {
		return NewRawdbError(StoreShardBlockError, err)
	}
	if err := db.Put(keyHash, val); err != nil {
		return NewRawdbError(StoreShardBlockError, err)
	}
	return nil
}

// StoreShardBlockIndex store block hash => block index
// key: i-{hash}
// value: {index-shardID}
func StoreShardBlockIndex(db incdb.Database, shardID byte, index uint64, hash common.Hash) error {
	key := GetShardBlockHashToIndexKey(hash)
	buf := make([]byte, 9)
	tempBuf := common.Uint64ToBytes(index)
	copy(buf, tempBuf)
	buf[8] = shardID
	//{i-[hash]}:index-shardID
	if err := db.Put(key, buf); err != nil {
		return NewRawdbError(StoreShardBlockIndexError, err)
	}

	return nil
}

func HasShardBlock(db incdb.Database, hash common.Hash) (bool, error) {
	keyHash := GetShardHashToBlockKey(hash)
	if ok, err := db.Has(keyHash); err != nil {
		return false, NewRawdbError(HasShardBlockError, fmt.Errorf("has key %+v failed", keyHash))
	} else if ok {
		return true, nil
	}
	return false, nil
}

func DeleteShardBlock(db incdb.Database, shardID byte, index uint64, hash common.Hash) error {
	keyHash := GetShardHashToBlockKey(hash)
	keyIndexToHash := GetShardIndexToBlockHashKey(shardID, index, hash)
	keyIndex := GetShardBlockHashToIndexKey(hash)
	if err := db.Delete(keyHash); err != nil {
		return NewRawdbError(DeleteShardBlockError, err)
	}
	if err := db.Delete(keyIndexToHash); err != nil {
		return NewRawdbError(DeleteShardBlockError, err)
	}
	if err := db.Delete(keyIndex); err != nil {
		return NewRawdbError(DeleteShardBlockError, err)
	}
	return nil
}

func GetShardBlockByHash(db incdb.Database, hash common.Hash) ([]byte, error) {
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

func GetShardBlockByIndex(db incdb.Database, shardID byte, index uint64) (map[common.Hash][]byte, error) {
	m := make(map[common.Hash][]byte)
	indexPrefix := GetShardIndexToBlockHashPrefix(shardID, index)
	iterator := db.NewIteratorWithPrefix(indexPrefix)
	for iterator.Next() {
		key := iterator.Key()
		strs := strings.Split(string(key), string(splitter))
		tempHash := []byte(strs[len(strs)-1])
		hash := common.BytesToHash(tempHash)
		keyHash := GetShardHashToBlockKey(hash)
		if ok, err := db.Has(keyHash); err != nil {
			return nil, NewRawdbError(GetShardBlockByIndexError, fmt.Errorf("has key %+v failed", keyHash))
		} else if !ok {
			return nil, NewRawdbError(GetShardBlockByIndexError, fmt.Errorf("block %+v not exist", hash))
		}
		block, err := db.Get(keyHash)
		if err != nil {
			return nil, NewRawdbError(GetShardBlockByIndexError, err)
		}
		ret := make([]byte, len(block))
		copy(ret, block)
		m[hash] = ret
	}
	return m, nil
}

func GetIndexOfBlock(db incdb.Database, hash common.Hash) (uint64, byte, error) {
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

func StoreShardBestState(db incdb.Database, shardID byte, v interface{}) error {
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

func GetShardBestState(db incdb.Database, shardID byte) ([]byte, error) {
	key := GetShardBestStateKey(shardID)
	shardBestStateBytes, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(StoreShardBestStateError, err)
	}
	return shardBestStateBytes, nil
}

// StoreFeeEstimator - Store data for FeeEstimator object
func StoreFeeEstimator(db incdb.Database, val []byte, shardID byte) error {
	key := GetFeeEstimatorPrefix(shardID)
	err := db.Put(key, val)
	if err != nil {
		return NewRawdbError(StoreFeeEstimatorError, err)
	}
	return nil
}

// GetFeeEstimator - Get data for FeeEstimator object as a json in byte format
func GetFeeEstimator(db incdb.Database, shardID byte) ([]byte, error) {
	key := GetFeeEstimatorPrefix(shardID)
	res, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(GetFeeEstimatorError, err)
	}
	return res, err
}

func StoreShardCommitteeRewardRootHash(db incdb.Database, shardID byte, height uint64, rootHash common.Hash) error {
	key := GetShardCommitteeRewardRootHashKey(shardID, height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreShardCommitteeRewardRootHashError, err)
	}
	return nil
}

func GetShardCommitteeRewardRootHash(db incdb.Database, shardID byte, height uint64) (common.Hash, error) {
	key := GetShardCommitteeRewardRootHashKey(shardID, height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetShardCommitteeRewardRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func DeleteShardCommitteeRewardRootHash(db incdb.Database, shardID byte, height uint64) error {
	key := GetShardCommitteeRewardRootHashKey(shardID, height)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(StoreShardCommitteeRewardRootHashError, err)
	}
	return nil
}

func StoreShardConsensusRootHash(db incdb.Database, shardID byte, height uint64, rootHash common.Hash) error {
	key := GetShardConsensusRootHashKey(shardID, height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreShardConsensusRootHashError, err)
	}
	return nil
}

func GetShardConsensusRootHash(db incdb.Database, shardID byte, height uint64) (common.Hash, error) {
	key := GetShardConsensusRootHashKey(shardID, height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetShardConsensusRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func DeleteShardConsensusRootHash(db incdb.Database, shardID byte, height uint64) error {
	key := GetShardConsensusRootHashKey(shardID, height)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(DeleteShardConsensusRootHashError, err)
	}
	return nil
}

func StoreShardFeatureRootHash(db incdb.Database, shardID byte, height uint64, rootHash common.Hash) error {
	key := GetShardFeatureRootHashKey(shardID, height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreShardFeatureRootHashError, err)
	}
	return nil
}

func GetShardFeatureRootHash(db incdb.Database, shardID byte, height uint64) (common.Hash, error) {
	key := GetShardFeatureRootHashKey(shardID, height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetShardFeatureRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func DeleteShardFeatureRootHash(db incdb.Database, shardID byte, height uint64) error {
	key := GetShardFeatureRootHashKey(shardID, height)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(DeleteShardFeatureRootHashError, err)
	}
	return nil
}

func StoreShardTransactionRootHash(db incdb.Database, shardID byte, height uint64, rootHash common.Hash) error {
	key := GetShardTransactionRootHashKey(shardID, height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreShardTransactionRootHashError, err)
	}
	return nil
}

func GetShardTransactionRootHash(db incdb.Database, shardID byte, height uint64) (common.Hash, error) {
	key := GetShardTransactionRootHashKey(shardID, height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetShardTransactionRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func DeleteShardTransactionRootHash(db incdb.Database, shardID byte, height uint64) error {
	key := GetShardTransactionRootHashKey(shardID, height)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(DeleteShardTransactionRootHashError, err)
	}
	return nil
}

func StoreShardSlashRootHash(db incdb.Database, shardID byte, height uint64, rootHash common.Hash) error {
	key := GetShardSlashRootHashKey(shardID, height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreShardSlashRootHashError, err)
	}
	return nil
}

func GetShardSlashRootHash(db incdb.Database, shardID byte, height uint64) (common.Hash, error) {
	key := GetShardSlashRootHashKey(shardID, height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetShardSlashRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func DeleteShardSlashRootHash(db incdb.Database, shardID byte, height uint64) error {
	key := GetShardSlashRootHashKey(shardID, height)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(StoreShardSlashRootHashError, err)
	}
	return nil
}
