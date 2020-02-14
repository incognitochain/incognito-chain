package rawdbv2

import (
	"github.com/incognitochain/incognito-chain/incdb"
)

func StorePreviousBeaconBestState(db incdb.Database, data []byte) error {
	key := GetPreviousBestStateKey(-1)
	if err := db.Put(key, data); err != nil {
		return NewRawdbError(StorePreviousBeaconBestStateError, err)
	}
	return nil
}

func GetPreviousBeaconBestState(db incdb.Database) ([]byte, error) {
	key := GetPreviousBestStateKey(-1)
	res, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(GetPreviousBeaconBestStateError, err)
	}
	return res, nil
}

func CleanUpPreviousBeaconBestState(db incdb.Database) error {
	key := GetPreviousBestStateKey(-1)
	if err := db.Delete(key); err != nil {
		return NewRawdbError(CleanUpPreviousBeaconBestStateError, err)
	}
	return nil
}

func StorePreviousShardBestState(db incdb.Database, shardID byte, data []byte) error {
	key := GetPreviousBestStateKey(int(shardID))
	if err := db.Put(key, data); err != nil {
		return NewRawdbError(StorePreviousShardBestStateError, err)
	}
	return nil
}

func GetPreviousShardBestState(db incdb.Database, shardID byte) ([]byte, error) {
	key := GetPreviousBestStateKey(int(shardID))
	res, err := db.Get(key)
	if err != nil {
		return nil, NewRawdbError(GetPreviousShardBestStateError, err)
	}
	return res, nil
}

func CleanUpPreviousShardBestState(db incdb.Database, shardID byte) error {
	key := GetPreviousBestStateKey(int(shardID))
	if err := db.Delete(key); err != nil {
		return NewRawdbError(CleanUpPreviousShardBestStateError, err)
	}
	return nil
}
