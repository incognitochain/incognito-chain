package rawdbv2

import (
	"bytes"
	"encoding/binary"

	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreCrossShardNextHeight(db incdb.Database, fromShard byte, toShard byte, curHeight uint64, nextHeight uint64) error {
	key := GetCrossShardNextHeightKey(fromShard, toShard, curHeight)
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, nextHeight)
	if err := db.Put(key, buf); err != nil {
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

func FetchCrossShardNextHeight(db incdb.Database, fromShard byte, toShard byte, curHeight uint64) (uint64, error) {
	key := GetCrossShardNextHeightKey(fromShard, toShard, curHeight)
	if _, err := hasCrossShardNextHeight(db, key); err != nil {
		return 0, NewRawdbError(FetchCrossShardNextHeightError, err)
	}
	tempNextHeight, err := db.Get(key)
	if err != nil {
		return 0, NewRawdbError(FetchCrossShardNextHeightError, err)
	}
	var nextHeight uint64
	err = binary.Read(bytes.NewReader(tempNextHeight[:8]), binary.LittleEndian, &nextHeight)
	if err != nil {
		return 0, NewRawdbError(FetchCrossShardNextHeightError, err)
	}
	return nextHeight, nil
}
