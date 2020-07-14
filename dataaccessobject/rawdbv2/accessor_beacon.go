package rawdbv2

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreBeaconRootsHash(db incdb.KeyValueWriter, hash common.Hash, rootsHash interface{}) error {
	key := GetBeaconRootsHashKey(hash)
	b, _ := json.Marshal(rootsHash)
	err := db.Put(key, b)
	if err != nil {
		return NewRawdbError(StoreShardConsensusRootHashError, err)
	}
	return nil
}

func GetBeaconRootsHash(db incdb.KeyValueReader, hash common.Hash) ([]byte, error) {
	key := GetBeaconRootsHashKey(hash)
	data, err := db.Get(key)
	return data, err
}

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
