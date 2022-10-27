package rawdbv2

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
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

// StoreBeaconBlock store block hash => block value
func StoreBeaconBlockByHash(db incdb.KeyValueWriter, hash common.Hash, v interface{}) error {
	keyHash := GetBeaconHashToBlockKey(hash)
	val, err := json.Marshal(v)
	if err != nil {
		return NewRawdbError(StoreBeaconBlockError, err)
	}
	if err := db.Put(keyHash, val); err != nil {
		return NewRawdbError(StoreBeaconBlockError, err)
	}
	return nil
}

func StoreFinalizedBeaconBlockHashByIndex(db incdb.KeyValueWriter, index uint64, hash common.Hash) error {
	keyHash := GetBeaconIndexToBlockHashKey(index)
	if err := db.Put(keyHash, hash.Bytes()); err != nil {
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

func GetFinalizedBeaconBlockHashByIndex(db incdb.KeyValueReader, index uint64) (*common.Hash, error) {
	keyHash := GetBeaconIndexToBlockHashKey(index)
	val, err := db.Get(keyHash)
	if err != nil {
		return nil, NewRawdbError(GetBeaconBlockByIndexError, err)
	}
	h, err := common.Hash{}.NewHash(val)
	return h, err
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

func StoreCacheCommitteeFromBlock(db incdb.KeyValueWriter, hash common.Hash, cid int, cpks []incognitokey.CommitteePublicKey) error {
	key := GetCacheCommitteeFromBlockKey(hash, cid)
	b, err := json.Marshal(cpks)
	if err != nil {
		return NewRawdbError(StoreCommmitteeFromBlockCacheError, err)
	}
	if err := db.Put(key, b); err != nil {
		return NewRawdbError(StoreBeaconBestStateError, err)
	}
	return nil
}

func GetAllCacheCommitteeFromBlock(db incdb.Database) (map[int]map[common.Hash][]incognitokey.CommitteePublicKey, error) {
	res := map[int]map[common.Hash][]incognitokey.CommitteePublicKey{}
	it := db.NewIteratorWithPrefix(cacheCommitteeFromBlockPrefix)
	for it.Next() {
		cpks := []incognitokey.CommitteePublicKey{}
		err := json.Unmarshal(it.Value(), &cpks)
		if err != nil {
			return nil, NewRawdbError(GetCommmitteeFromBlockCacheError, err)
		}
		key := it.Key()
		keyData := key[len(cacheCommitteeFromBlockPrefix):]

		h := common.Hash{}
		err = h.SetBytes(keyData[:32])
		if err != nil {
			return nil, NewRawdbError(GetCommmitteeFromBlockCacheError, err)
		}
		cid, err := common.BytesToInt32(keyData[32:])
		if err != nil {
			return nil, NewRawdbError(GetCommmitteeFromBlockCacheError, err)
		}
		if _, ok := res[int(cid)]; !ok {
			res[int(cid)] = make(map[common.Hash][]incognitokey.CommitteePublicKey)
		}
		res[int(cid)][h] = cpks
	}
	return res, nil
}

func GetCacheCommitteeFromBlock(db incdb.Database, hash common.Hash, cid int) ([]incognitokey.CommitteePublicKey, error) {
	key := GetCacheCommitteeFromBlockKey(hash, cid)
	b, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(GetCommmitteeFromBlockCacheError, err)
	}
	cpks := []incognitokey.CommitteePublicKey{}
	err = json.Unmarshal(b, &cpks)
	if err != nil {
		return nil, NewRawdbError(GetCommmitteeFromBlockCacheError, err)
	}
	return cpks, nil
}
