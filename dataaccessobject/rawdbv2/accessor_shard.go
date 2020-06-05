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

func FinalizedShardBlock(db incdb.Database, shardID byte, hash common.Hash) error {
	key := GetLastShardBlockKey(shardID)
	if err := db.Put(key, hash[:]); err != nil {
		return NewRawdbError(FinalizedShardBlockError, err)
	}
	iter := db.NewIteratorWithPrefix(GetViewPrefixWithValue(hash))
	for iter.Next() {
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		if err := db.Delete(key); err != nil {
			return NewRawdbError(FinalizedShardBlockError, err)
		}
	}
	return nil
}

func GetFinalizedShardBlock(db incdb.Database, shardID byte) (common.Hash, error) {
	key := GetLastShardBlockKey(shardID)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetFinalizedShardBlockError, err)
	}
	h, err := common.Hash{}.NewHash(res)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetFinalizedShardBlockError, err)
	}
	return *h, nil
}

// StoreShardBlock store block hash => block value and block index => block hash
// record1: prefix-shardid-index-hash => empty
// record2: prefix-hash => block value
func StoreShardBlock(db incdb.KeyValueWriter, shardID byte, index uint64, hash common.Hash, v interface{}) error {
	keyHash := GetShardHashToBlockKey(hash)
	keyIndex := GetShardIndexToBlockHashKey(shardID, index, hash)
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

func StoreShardBlockWithView(db incdb.Database, view common.Hash, shardID byte, height uint64, blockHash common.Hash) error {
	key := GetViewShardKey(view, shardID, height)
	err := db.Put(key, blockHash[:])
	if err != nil {
		return NewRawdbError(StoreShardBlockWithViewError, err)
	}
	return nil
}

func GetShardBlockByView(db incdb.Database, view common.Hash) (map[uint64]common.Hash, error) {
	iter := db.NewIteratorWithPrefix(GetViewPrefixWithValue(view))
	m := make(map[uint64]common.Hash)
	for iter.Next() {
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		strs := strings.Split(string(key), string(splitter))
		tempHeight := strs[3]
		height, err := common.BytesToUint64([]byte(tempHeight))
		if err != nil {
			return nil, NewRawdbError(GetShardBlockByViewError, err)
		}
		h, err := common.Hash{}.NewHash(value)
		if err != nil {
			return nil, NewRawdbError(GetShardBlockByViewError, err)
		}
		m[height] = *h
	}
	return m, nil
}

func UpdateShardBlockView(db incdb.Database, oldView common.Hash, newView common.Hash) error {
	iter := db.NewIteratorWithPrefix(GetViewPrefixWithValue(oldView))
	for iter.Next() {
		oldKey := make([]byte, len(iter.Key()))
		copy(oldKey, iter.Key())
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		strs := strings.Split(string(oldKey), string(splitter))
		shardID := []byte(strs[2])[0]
		tempHeight := strs[3]
		height, err := common.BytesToUint64([]byte(tempHeight))
		if err != nil {
			return NewRawdbError(UpdateShardBlockViewError, err)
		}
		newKey := GetViewShardKey(newView, shardID, height)
		if err := db.Put(newKey, value); err != nil {
			return NewRawdbError(UpdateShardBlockViewError, err)
		}
		if err := db.Delete(oldKey); err != nil {
			return NewRawdbError(UpdateShardBlockViewError, err)
		}
	}
	return nil
}

func DeleteShardBlockByView(db incdb.Database, view common.Hash) error {
	iter := db.NewIteratorWithPrefix(GetViewPrefixWithValue(view))
	for iter.Next() {
		key := make([]byte, len(iter.Key()))
		copy(key, iter.Key())
		strs := strings.Split(string(key), string(splitter))
		shardID := []byte(strs[2])[0]
		tempHeight := strs[3]
		height, err := common.BytesToUint64([]byte(tempHeight))
		if err != nil {
			return NewRawdbError(DeleteShardBlockByViewError, err)
		}
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		h, err := common.Hash{}.NewHash(value)
		if err != nil {
			return NewRawdbError(DeleteShardBlockByViewError, err)
		}
		if err := DeleteShardBlock(db, shardID, height, *h); err != nil {
			return NewRawdbError(DeleteShardBlockByViewError, err)
		}
		if err := db.Delete(key); err != nil {
			return NewRawdbError(DeleteShardBlockByViewError, err)
		}
	}
	return nil
}

// StoreShardBlockIndex store block hash => block index
// key: i-{hash}
// value: {index-shardID}
func StoreShardBlockIndex(db incdb.KeyValueWriter, shardID byte, index uint64, hash common.Hash) error {
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

func HasShardBlock(db incdb.KeyValueReader, hash common.Hash) (bool, error) {
	keyHash := GetShardHashToBlockKey(hash)
	if ok, err := db.Has(keyHash); err != nil {
		return false, NewRawdbError(HasShardBlockError, fmt.Errorf("has key %+v failed", keyHash))
	} else if ok {
		return true, nil
	}
	return false, nil
}

func DeleteShardBlock(db incdb.KeyValueWriter, shardID byte, index uint64, hash common.Hash) error {
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

func StoreShardCommitteeRewardRootHash(db incdb.KeyValueWriter, shardID byte, height uint64, rootHash common.Hash) error {
	key := GetShardCommitteeRewardRootHashKey(shardID, height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreShardCommitteeRewardRootHashError, err)
	}
	return nil
}

func GetShardCommitteeRewardRootHash(db incdb.KeyValueReader, shardID byte, height uint64) (common.Hash, error) {
	key := GetShardCommitteeRewardRootHashKey(shardID, height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetShardCommitteeRewardRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func DeleteShardCommitteeRewardRootHash(db incdb.KeyValueWriter, shardID byte, height uint64) error {
	key := GetShardCommitteeRewardRootHashKey(shardID, height)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(StoreShardCommitteeRewardRootHashError, err)
	}
	return nil
}

func StoreShardConsensusRootHash(db incdb.KeyValueWriter, shardID byte, height uint64, rootHash common.Hash) error {
	key := GetShardConsensusRootHashKey(shardID, height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreShardConsensusRootHashError, err)
	}
	return nil
}

func GetShardConsensusRootHash(db incdb.KeyValueReader, shardID byte, height uint64) (common.Hash, error) {
	key := GetShardConsensusRootHashKey(shardID, height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetShardConsensusRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func DeleteShardConsensusRootHash(db incdb.KeyValueWriter, shardID byte, height uint64) error {
	key := GetShardConsensusRootHashKey(shardID, height)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(DeleteShardConsensusRootHashError, err)
	}
	return nil
}

func StoreShardFeatureRootHash(db incdb.KeyValueWriter, shardID byte, height uint64, rootHash common.Hash) error {
	key := GetShardFeatureRootHashKey(shardID, height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreShardFeatureRootHashError, err)
	}
	return nil
}

func GetShardFeatureRootHash(db incdb.KeyValueReader, shardID byte, height uint64) (common.Hash, error) {
	key := GetShardFeatureRootHashKey(shardID, height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetShardFeatureRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func DeleteShardFeatureRootHash(db incdb.KeyValueWriter, shardID byte, height uint64) error {
	key := GetShardFeatureRootHashKey(shardID, height)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(DeleteShardFeatureRootHashError, err)
	}
	return nil
}

func StoreShardTransactionRootHash(db incdb.KeyValueWriter, shardID byte, height uint64, rootHash common.Hash) error {
	key := GetShardTransactionRootHashKey(shardID, height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreShardTransactionRootHashError, err)
	}
	return nil
}

func GetShardTransactionRootHash(db incdb.KeyValueReader, shardID byte, height uint64) (common.Hash, error) {
	key := GetShardTransactionRootHashKey(shardID, height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetShardTransactionRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func DeleteShardTransactionRootHash(db incdb.KeyValueWriter, shardID byte, height uint64) error {
	key := GetShardTransactionRootHashKey(shardID, height)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(DeleteShardTransactionRootHashError, err)
	}
	return nil
}

func StoreShardSlashRootHash(db incdb.KeyValueWriter, shardID byte, height uint64, rootHash common.Hash) error {
	key := GetShardSlashRootHashKey(shardID, height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreShardSlashRootHashError, err)
	}
	return nil
}

func GetShardSlashRootHash(db incdb.KeyValueReader, shardID byte, height uint64) (common.Hash, error) {
	key := GetShardSlashRootHashKey(shardID, height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetShardSlashRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func DeleteShardSlashRootHash(db incdb.KeyValueWriter, shardID byte, height uint64) error {
	key := GetShardSlashRootHashKey(shardID, height)
	err := db.Delete(key)
	if err != nil {
		return NewRawdbError(StoreShardSlashRootHashError, err)
	}
	return nil
}
