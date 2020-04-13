package rawdbv2

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"

	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreLastBeaconHeightConfirmCrossShard(db incdb.Database, height uint64) error {
	if err := db.Put(lastBeaconHeightConfirmCrossShard, common.Uint64ToBytes(height)); err != nil {
		return NewRawdbError(StoreCrossShardNextHeightError, err)
	}
	return nil
}

func GetLastBeaconHeightConfirmCrossShard(db incdb.Database) uint64 {
	lastProcessHeight, _ := db.Get(lastBeaconHeightConfirmCrossShard)
	height, _ := common.BytesToUint64(lastProcessHeight)
	return height
}

func StoreCrossShardNextHeight(db incdb.Database, fromShard byte, toShard byte, curHeight uint64, val []byte) error {
	key := GetCrossShardNextHeightKey(fromShard, toShard, curHeight)
	if err := db.Put(key, val); err != nil {
		return NewRawdbError(StoreCrossShardNextHeightError, err)
	}
	return nil
}

func hasCrossShardNextHeight(db incdb.Database, key []byte) (bool, error) {
	exist, err := db.Has(key)
	if err != nil {
		return false, err
	} else {
		return exist, nil
	}
}

func GetCrossShardNextHeight(db incdb.Database, fromShard byte, toShard byte, curHeight uint64) ([]byte, error) {
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

func RestoreCrossShardNextHeights(db incdb.Database, fromShard byte, toShard byte, curHeight uint64) error {
	key := GetCrossShardNextHeightKey(fromShard, toShard, curHeight)
	curHeightBytes := common.Uint64ToBytes(curHeight)
	heightKey := append(key, curHeightBytes...)
	for {
		nextHeightBytes, err := db.Get(heightKey)
		if err != nil {
			if isOk, err1 := db.Has(heightKey); err1 != nil {
				return NewRawdbError(RestoreCrossShardNextHeightsError, err1)
			} else {
				if !isOk {
					return NewRawdbError(RestoreCrossShardNextHeightsError, err)
				}
			}
		}
		//Delete will not returns error if key doesn't exist.
		err = db.Delete(heightKey)
		if err != nil {
			return NewRawdbError(RestoreCrossShardNextHeightsError, err)
		}
		var nextHeight uint64
		err = binary.Read(bytes.NewReader(nextHeightBytes[:8]), binary.LittleEndian, &nextHeight)
		if err != nil {
			fmt.Println(NewRawdbError(RestoreCrossShardNextHeightsError, err))
		}
		if nextHeight == 0 {
			break
		}
		heightKey = append(key, nextHeightBytes...)
	}
	nextHeightBytes := make([]byte, 8)
	heightKey = append(key, curHeightBytes...)
	if err := db.Put(heightKey, nextHeightBytes); err != nil {
		return NewRawdbError(RestoreCrossShardNextHeightsError, err)
	}
	return nil
}
