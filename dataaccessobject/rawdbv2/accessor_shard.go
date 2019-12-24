package rawdbv2

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreShardBlock(db incdb.Database, shardID byte, index uint64, hash common.Hash, v interface{}) error {
	keyHash := GetShardBlockHashKey(shardID, hash)
	if ok, _ := db.Has(keyHash); ok {
		return NewRawdbError(StoreShardBlockError, fmt.Errorf("key %+v already exists", keyHash))
	}
	keyIndex := GetShardBlockIndexKey(shardID, index, hash)
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

func HasShardBlock(db incdb.Database, shardID byte, hash common.Hash) (bool, error) {
	keyHash := GetShardBlockHashKey(shardID, hash)
	if ok, err := db.Has(keyHash); err != nil {
		return false, NewRawdbError(HasShardBlockError, fmt.Errorf("has key %+v failed", keyHash))
	} else if ok {
		return true, nil
	}
	return false, nil
}

func GetShardBlockByHash(db incdb.Database, shardID byte, hash common.Hash) ([]byte, error) {
	keyHash := GetShardBlockHashKey(shardID, hash)
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
	indexPrefix := GetShardBlockIndexPrefix(shardID, index)
	iterator := db.NewIteratorWithPrefix(indexPrefix)
	for iterator.Next() {
		key := iterator.Key()
		strs := strings.Split(string(key), string(splitter))
		tempHash := []byte(strs[len(strs)-1])
		hash := common.BytesToHash(tempHash)
		keyHash := GetShardBlockHashKey(shardID, hash)
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

func DeleteShardBlock(db incdb.Database, shardID byte, index uint64, hash common.Hash) error {
	keyHash := GetShardBlockHashKey(shardID, hash)
	keyIndex := GetShardBlockIndexKey(shardID, index, hash)
	if err := db.Delete(keyHash); err != nil {
		return NewRawdbError(DeleteShardBlockError, err)
	}
	if err := db.Delete(keyIndex); err != nil {
		return NewRawdbError(DeleteShardBlockError, err)
	}
	return nil
}
