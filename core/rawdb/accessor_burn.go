package rawdb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
)

func StoreBurningConfirm(db incdb.Database, txID common.Hash, height uint64, bd *[]incdb.BatchData) error {
	key := append(burnConfirmPrefix, txID[:]...)
	value := common.Uint64ToBytes(height)

	if bd != nil {
		*bd = append(*bd, incdb.BatchData{key, value})
		return nil
	}
	err := db.Put(key, value)
	if err != nil {
		return NewRawdbError(LvdbPutError, err)
	}
	return nil
}

func GetBurningConfirm(db incdb.Database, txID common.Hash) (uint64, error) {
	key := append(burnConfirmPrefix, txID[:]...)
	value, err := db.Get(key)
	if err != nil {
		return 0, NewRawdbError(LvdbGetError, err)
	}
	height, _ := common.BytesToUint64(value)
	return height, nil
}
