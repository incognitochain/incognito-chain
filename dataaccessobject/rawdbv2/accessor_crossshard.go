package rawdbv2

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreLastBeaconStateConfirmCrossShard(db incdb.Database, state interface{}) error {
	key := GetLastBeaconHeightConfirmCrossShardKey()
	val, _ := json.Marshal(state)
	if err := db.Put(key, val); err != nil {
		return NewRawdbError(StoreCrossShardNextHeightError, err)
	}
	return nil
}

func GetLastBeaconStateConfirmCrossShard(db incdb.Database) []byte {
	key := GetLastBeaconHeightConfirmCrossShardKey()
	lastState, _ := db.Get(key)
	return lastState
}

func StoreCrossShardNextHeight(db incdb.KeyValueWriter, fromShard byte, toShard byte, curHeight uint64, val []byte) error {
	key := GetCrossShardNextHeightKey(fromShard, toShard, curHeight)
	if err := db.Put(key, val); err != nil {
		return NewRawdbError(StoreCrossShardNextHeightError, err)
	}
	return nil
}

func hasCrossShardNextHeight(db incdb.KeyValueReader, key []byte) (bool, error) {
	exist, err := db.Has(key)
	if err != nil {
		return false, err
	} else {
		return exist, nil
	}
}

func GetCrossShardNextHeight(db incdb.KeyValueReader, fromShard byte, toShard byte, curHeight uint64) ([]byte, error) {
	key := GetCrossShardNextHeightKey(fromShard, toShard, curHeight)
	if _, err := hasCrossShardNextHeight(db, key); err != nil {
		return nil, NewRawdbError(FetchCrossShardNextHeightError, err)
	}
	nextCrossShardInfo, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(FetchCrossShardNextHeightError, err)
	}
	return nextCrossShardInfo, nil
}
