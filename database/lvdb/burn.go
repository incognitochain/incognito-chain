package lvdb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
)

func (db *db) StoreBurningConfirm(txID common.Hash, height uint64, bd *[]database.BatchData) error {
	key := append(burnConfirmPrefix, txID[:]...)
	value := common.Uint64ToBytes(height)

	if bd != nil {
		*bd = append(*bd, database.BatchData{key, value})
		return nil
	}
	return db.Put(key, value)
}

func (db *db) GetBurningConfirm(txID common.Hash) (uint64, error) {
	key := append(burnConfirmPrefix, txID[:]...)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	height, _ := common.BytesToUint64(value)
	return height, nil
}
