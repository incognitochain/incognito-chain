package lvdb

import "github.com/incognitochain/incognito-chain/common"

func (db *db) StoreBurningConfirm(txID []byte, height uint64) error {
	key := append(burnConfirmPrefix, txID...)
	value := common.Uint64ToBytes(height)
	return db.Put(key, value)
}

func (db *db) GetBurningConfirm(txID []byte) (uint64, error) {
	key := append(burnConfirmPrefix, txID...)
	value, err := db.Get(key)
	if err != nil {
		return 0, err
	}
	height := common.BytesToUint64(value)
	return height, nil
}
