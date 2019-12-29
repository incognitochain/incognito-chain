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
func StoreShardBlockIndex(db incdb.Database, hash common.Hash, index uint64, shardID byte) error {
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

func HasBlock(db incdb.Database, shardID byte, hash common.Hash) (bool, error) {
	keyHash := GetShardHashToBlockKey(hash)
	if ok, err := db.Has(keyHash); err != nil {
		return false, NewRawdbError(HasShardBlockError, fmt.Errorf("has key %+v failed", keyHash))
	} else if ok {
		return true, nil
	}
	return false, nil
}

func FetchBlock(db incdb.Database, hash common.Hash) ([]byte, error) {
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

func GetBlockByIndex(db incdb.Database, shardID byte, index uint64) (map[common.Hash][]byte, error) {
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

func DeleteBlock(db incdb.Database, shardID byte, index uint64, hash common.Hash) error {
	keyHash := GetShardHashToBlockKey(hash)
	keyIndex := GetShardIndexToBlockHashKey(shardID, index, hash)
	if err := db.Delete(keyHash); err != nil {
		return NewRawdbError(DeleteShardBlockError, err)
	}
	if err := db.Delete(keyIndex); err != nil {
		return NewRawdbError(DeleteShardBlockError, err)
	}
	return nil
}
