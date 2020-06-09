package rawdbv2

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

//func FinalizedBeaconBlock(db incdb.Database, hash common.Hash) error {
//	key := GetLastBeaconBlockKey()
//	if err := db.Put(key, hash[:]); err != nil {
//		return NewRawdbError(FinalizedBeaconBlockError, err)
//	}
//	iter := db.NewIteratorWithPrefix(GetViewPrefixWithValue(hash))
//	for iter.Next() {
//		key := make([]byte, len(iter.Key()))
//		copy(key, iter.Key())
//		if err := db.Delete(key); err != nil {
//			return NewRawdbError(FinalizedBeaconBlockError, err)
//		}
//	}
//	return nil
//}
//
//func GetFinalizedBeaconBlock(db incdb.Database) (common.Hash, error) {
//	key := GetLastBeaconBlockKey()
//	res, err := db.Get(key)
//	if err != nil {
//		return common.Hash{}, NewRawdbError(GetFinalizedBeaconBlockError, err)
//	}
//	h, err := common.Hash{}.NewHash(res)
//	if err != nil {
//		return common.Hash{}, NewRawdbError(GetFinalizedBeaconBlockError, err)
//	}
//	return *h, nil
//}

// StoreBeaconBlock store block hash => block value and block index => block hash
// record1: prefix-index-hash => empty
// record2: prefix-hash => block value
func StoreBeaconBlock(db incdb.KeyValueWriter, index uint64, hash common.Hash, v interface{}) error {
	keyHash := GetBeaconHashToBlockKey(hash)
	//keyIndex := GetBeaconIndexToBlockHashKey(index, hash)
	val, err := json.Marshal(v)
	if err != nil {
		return NewRawdbError(StoreBeaconBlockError, err)
	}
	//if err := db.Put(keyIndex, []byte{}); err != nil {
	//	return NewRawdbError(StoreBeaconBlockError, err)
	//}
	if err := db.Put(keyHash, val); err != nil {
		return NewRawdbError(StoreBeaconBlockError, err)
	}
	return nil
}

//func StoreBeaconBlockWithView(db incdb.Database, view common.Hash, height uint64, blockHash common.Hash) error {
//	key := GetViewBeaconKey(view, height)
//	err := db.Put(key, blockHash[:])
//	if err != nil {
//		return NewRawdbError(StoreBeaconBlockWithViewError, err)
//	}
//	return nil
//}

//func GetBeaconBlockByView(db incdb.Database, view common.Hash) (map[uint64]common.Hash, error) {
//	iter := db.NewIteratorWithPrefix(GetViewPrefixWithValue(view))
//	m := make(map[uint64]common.Hash)
//	for iter.Next() {
//		key := make([]byte, len(iter.Key()))
//		copy(key, iter.Key())
//		value := make([]byte, len(iter.Value()))
//		copy(value, iter.Value())
//		strs := strings.Split(string(key), string(splitter))
//		tempHeight := strs[2]
//		height, err := common.BytesToUint64([]byte(tempHeight))
//		if err != nil {
//			return nil, NewRawdbError(GetBeaconBlockByViewError, err)
//		}
//		h, err := common.Hash{}.NewHash(value)
//		if err != nil {
//			return nil, NewRawdbError(GetBeaconBlockByViewError, err)
//		}
//		m[height] = *h
//	}
//	return m, nil
//}

//func UpdateBeaconBlockView(db incdb.Database, oldView common.Hash, newView common.Hash) error {
//	iter := db.NewIteratorWithPrefix(GetViewPrefixWithValue(oldView))
//	for iter.Next() {
//		oldKey := make([]byte, len(iter.Key()))
//		copy(oldKey, iter.Key())
//		value := make([]byte, len(iter.Value()))
//		copy(value, iter.Value())
//		strs := strings.Split(string(oldKey), string(splitter))
//		tempHeight := strs[2]
//		height, err := common.BytesToUint64([]byte(tempHeight))
//		if err != nil {
//			return NewRawdbError(UpdateBeaconBlockViewError, err)
//		}
//		newKey := GetViewBeaconKey(newView, height)
//		if err := db.Put(newKey, value); err != nil {
//			return NewRawdbError(UpdateBeaconBlockViewError, err)
//		}
//		if err := db.Delete(oldKey); err != nil {
//			return NewRawdbError(UpdateBeaconBlockViewError, err)
//		}
//	}
//	return nil
//}

//func DeleteBeaconBlockByView(db incdb.Database, view common.Hash) error {
//	iter := db.NewIteratorWithPrefix(GetViewPrefixWithValue(view))
//	for iter.Next() {
//		key := make([]byte, len(iter.Key()))
//		copy(key, iter.Key())
//		value := make([]byte, len(iter.Value()))
//		copy(value, iter.Value())
//		strs := strings.Split(string(key), string(splitter))
//		tempHeight := strs[2]
//		height, err := common.BytesToUint64([]byte(tempHeight))
//		if err != nil {
//			return NewRawdbError(DeleteBeaconBlockByViewError, err)
//		}
//		h, err := common.Hash{}.NewHash(value)
//		if err != nil {
//			return NewRawdbError(DeleteBeaconBlockByViewError, err)
//		}
//		if err := DeleteBeaconBlock(db, height, *h); err != nil {
//			return NewRawdbError(DeleteBeaconBlockByViewError, err)
//		}
//		if err := db.Delete(key); err != nil {
//			return NewRawdbError(DeleteBeaconBlockByViewError, err)
//		}
//	}
//	return nil
//}

// StoreBeaconBlockIndex store block hash => block index
// key: i-{hash}
// value: {index-shardID}
//func StoreBeaconBlockIndex(db incdb.KeyValueWriter, index uint64, hash common.Hash) error {
//	key := GetBeaconBlockHashToIndexKey(hash)
//	buf := common.Uint64ToBytes(index)
//	err := db.Put(key, buf)
//	if err != nil {
//		return NewRawdbError(StoreBeaconBlockIndexError, err)
//	}
//	return nil
//}

func HasBeaconBlock(db incdb.KeyValueReader, hash common.Hash) (bool, error) {
	keyHash := GetBeaconHashToBlockKey(hash)
	if ok, err := db.Has(keyHash); err != nil {
		return false, NewRawdbError(HasBeaconBlockError, fmt.Errorf("has key %+v failed", keyHash))
	} else if ok {
		return true, nil
	}
	return false, nil
}

func GetBeaconBlockByHash(db incdb.KeyValueReader, hash common.Hash) ([]byte, error) {
	keyHash := GetBeaconHashToBlockKey(hash)
	if ok, err := db.Has(keyHash); err != nil {
		return []byte{}, NewRawdbError(GetBeaconBlockByHashError, fmt.Errorf("has key %+v failed", keyHash))
	} else if !ok {
		return []byte{}, NewRawdbError(GetBeaconBlockByHashError, fmt.Errorf("block %+v not exist", hash))
	}
	block, err := db.Get(keyHash)
	if err != nil {
		return nil, NewRawdbError(GetBeaconBlockByHashError, err)
	}
	ret := make([]byte, len(block))
	copy(ret, block)
	return ret, nil
}

//func GetBeaconBlockByIndex(db incdb.Database, index uint64) (map[common.Hash][]byte, error) {
//	m := make(map[common.Hash][]byte)
//	indexPrefix := GetBeaconIndexToBlockHashPrefix(index)
//	iterator := db.NewIteratorWithPrefix(indexPrefix)
//	for iterator.Next() {
//		key := iterator.Key()
//		strs := strings.Split(string(key), string(splitter))
//		tempHash := []byte(strs[len(strs)-1])
//		hash := common.BytesToHash(tempHash)
//		keyHash := GetBeaconHashToBlockKey(hash)
//		if ok, err := db.Has(keyHash); err != nil {
//			return nil, NewRawdbError(GetBeaconBlockByIndexError, fmt.Errorf("has key %+v failed", keyHash))
//		} else if !ok {
//			return nil, NewRawdbError(GetBeaconBlockByIndexError, fmt.Errorf("block %+v not exist", hash))
//		}
//		block, err := db.Get(keyHash)
//		if err != nil {
//			return nil, NewRawdbError(GetBeaconBlockByIndexError, err)
//		}
//		ret := make([]byte, len(block))
//		copy(ret, block)
//		m[hash] = ret
//	}
//	return m, nil
//}

//func GetIndexOfBeaconBlock(db incdb.KeyValueReader, hash common.Hash) (uint64, error) {
//	key := GetBeaconBlockHashToIndexKey(hash)
//	buf, err := db.Get(key)
//	if err != nil {
//		return 0, NewRawdbError(GetIndexOfBeaconBlockError, err)
//	}
//	index, err := common.BytesToUint64(buf)
//	if err != nil {
//		return 0, NewRawdbError(GetIndexOfBeaconBlockError, err)
//	}
//	return index, nil
//}

//func DeleteBeaconBlock(db incdb.KeyValueWriter, index uint64, hash common.Hash) error {
//	keyHash := GetBeaconHashToBlockKey(hash)
//	keyIndexToHash := GetBeaconIndexToBlockHashKey(index, hash)
//	keyIndex := GetBeaconBlockHashToIndexKey(hash)
//	if err := db.Delete(keyHash); err != nil {
//		return NewRawdbError(DeleteBeaconBlockError, err)
//	}
//	if err := db.Delete(keyIndexToHash); err != nil {
//		return NewRawdbError(DeleteBeaconBlockError, err)
//	}
//	if err := db.Delete(keyIndex); err != nil {
//		return NewRawdbError(DeleteBeaconBlockError, err)
//	}
//	return nil
//}

//func GetBeaconBlockHashByIndex(db incdb.Database, index uint64) ([]common.Hash, error) {
//	beaconBlockHashes := []common.Hash{}
//	indexPrefix := GetBeaconIndexToBlockHashPrefix(index)
//	iterator := db.NewIteratorWithPrefix(indexPrefix)
//	for iterator.Next() {
//		key := iterator.Key()
//		strs := strings.Split(string(key), string(splitter))
//		tempHash := []byte(strs[len(strs)-1])
//		hash := common.BytesToHash(tempHash)
//		beaconBlockHashes = append(beaconBlockHashes, hash)
//	}
//	if len(beaconBlockHashes) == 0 {
//		return beaconBlockHashes, errors.New("beacon block hash not found")
//	}
//	return beaconBlockHashes, nil
//}

func StoreBeaconViews(db incdb.KeyValueWriter, val []byte) error {
	key := GetBeaconViewsKey()
	if err := db.Put(key, val); err != nil {
		return NewRawdbError(StoreBeaconBestStateError, err)
	}
	return nil
}

func GetBeaconViews(db incdb.KeyValueReader) ([]byte, error) {
	key := GetBeaconViewsKey()
	block, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(GetBeaconBestStateError, err)
	}
	return block, nil
}

func StoreBeaconConsensusStateRootHash(db incdb.KeyValueWriter, height uint64, rootHash common.Hash) error {
	key := GetBeaconConsensusRootHashKey(height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreBeaconConsensusRootHashError, err)
	}
	return nil
}

func GetBeaconConsensusStateRootHash(db incdb.KeyValueReader, height uint64) (common.Hash, error) {
	key := GetBeaconConsensusRootHashKey(height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetBeaconConsensusRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func StoreBeaconRewardStateRootHash(db incdb.KeyValueWriter, height uint64, rootHash common.Hash) error {
	key := GetBeaconRewardRootHashKey(height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreBeaconRewardRootHashError, err)
	}
	return nil
}

func GetBeaconRewardStateRootHash(db incdb.KeyValueReader, height uint64) (common.Hash, error) {
	key := GetBeaconRewardRootHashKey(height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetBeaconRewardRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func StoreBeaconFeatureStateRootHash(db incdb.KeyValueWriter, height uint64, rootHash common.Hash) error {
	key := GetBeaconFeatureRootHashKey(height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreBeaconFeatureRootHashError, err)
	}
	return nil
}

func GetBeaconFeatureStateRootHash(db incdb.KeyValueReader, height uint64) (common.Hash, error) {
	key := GetBeaconFeatureRootHashKey(height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetBeaconFeatureRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

func StoreBeaconSlashStateRootHash(db incdb.KeyValueWriter, height uint64, rootHash common.Hash) error {
	key := GetBeaconSlashRootHashKey(height)
	err := db.Put(key, rootHash[:])
	if err != nil {
		return NewRawdbError(StoreBeaconSlashRootHashError, err)
	}
	return nil
}

func GetBeaconSlashStateRootHash(db incdb.KeyValueReader, height uint64) (common.Hash, error) {
	key := GetBeaconSlashRootHashKey(height)
	res, err := db.Get(key)
	if err != nil {
		return common.Hash{}, NewRawdbError(GetBeaconSlashRootHashError, err)
	}
	return common.BytesToHash(res), nil
}

//StoreBeaconPreCommitteeInfo ...
func StoreBeaconPreCommitteeInfo(db incdb.KeyValueWriter, hash common.Hash, value []byte) error {
	err := db.Put(getBeaconPreCommitteeInfoKey(hash), value)
	if err != nil {
		return NewRawdbError(StoreBeaconPreCommitteeInfoError, err)
	}
	return nil
}

//StoreShardPreCommitteeInfo ...
func StoreShardPreCommitteeInfo(db incdb.KeyValueWriter, hash common.Hash, value []byte) error {
	err := db.Put(getShardPreCommitteeInfoKey(hash), value)
	if err != nil {
		return NewRawdbError(StoreBeaconPreCommitteeInfoError, err)
	}
	return nil
}

//GetBeaconPreCommitteeInfo ...
func GetBeaconPreCommitteeInfo(db incdb.KeyValueReader, hash common.Hash) ([]byte, error) {
	res, err := db.Get(getBeaconPreCommitteeInfoKey(hash))
	if err != nil {
		return nil, NewRawdbError(GetBeaconPreCommitteeInfoError, err)
	}
	return res, nil
}

//GetShardPreCommitteeInfo ...
func GetShardPreCommitteeInfo(db incdb.KeyValueReader, hash common.Hash) ([]byte, error) {
	res, err := db.Get(getShardPreCommitteeInfoKey(hash))
	if err != nil {
		return nil, NewRawdbError(GetBeaconPreCommitteeInfoError, err)
	}
	return res, nil
}

//GetShardPendingValidators ...
func GetShardPendingValidators(db incdb.KeyValueReader, hash common.Hash) ([]byte, error) {
	res, err := db.Get(getShardPendingValidatorsKey(hash))
	if err != nil {
		return nil, NewRawdbError(GetShardPendingValidatorsError, err)
	}
	return res, nil
}

//StoreShardPendingValidators ...
func StoreShardPendingValidators(db incdb.KeyValueWriter, hash common.Hash, value []byte) error {
	err := db.Put(getShardPendingValidatorsKey(hash), value)
	if err != nil {
		return NewRawdbError(StoreBeaconPreCommitteeInfoError, err)
	}
	return nil
}
