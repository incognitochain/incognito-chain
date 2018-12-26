package lvdb

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
)

func (db *db) SendInitDCBVoteToken(startedBlock uint32, pubKey []byte, amount uint32) error {
	key := GetDCBVoteTokenAmountKey(startedBlock, pubKey)
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

func (db *db) SendInitGOVVoteToken(startedBlock uint32, pubKey []byte, amount uint32) error {
	key := GetGOVVoteTokenAmountKey(startedBlock, pubKey)
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
