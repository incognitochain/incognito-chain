package lvdb

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
)

func (db *db) SendInitDCBVoteToken(boardIndex uint32, pubKey []byte, amount uint32) error {
	key := GetKeyDCBVoteTokenAmount(boardIndex, pubKey)
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if !ok {
		zeroInBytes := common.Uint32ToBytes(uint32(0))
		db.Put(key, zeroInBytes)
	}

	currentAmountInBytes, err := db.lvdb.Get(key, nil)
	currentAmount := common.BytesToUint32(currentAmountInBytes)
	newAmount := currentAmount + amount

	newAmountInBytes := common.Uint32ToBytes(newAmount)
	err = db.Put(key, newAmountInBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	return nil
}

func (db *db) SendInitGOVVoteToken(boardIndex uint32, pubKey []byte, amount uint32) error {
	key := GetGOVVoteTokenAmountKey(boardIndex, pubKey)
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if !ok {
		zeroInBytes := common.Uint32ToBytes(uint32(0))
		db.Put(key, zeroInBytes)
	}

	currentAmountInBytes, err := db.lvdb.Get(key, nil)
	currentAmount := common.BytesToUint32(currentAmountInBytes)
	newAmount := currentAmount + amount

	newAmountInBytes := common.Uint32ToBytes(newAmount)
	err = db.Put(key, newAmountInBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	return nil
}
