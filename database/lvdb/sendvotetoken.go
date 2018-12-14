package lvdb

import (
	"encoding/binary"

	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
)

func (db *db) SendInitDCBVoteToken(startedBlock uint32, pubKey []byte, amount uint64) error {
	key := db.GetKey(string(DCBVoteTokenAmountPrefix), string(startedBlock)+string(pubKey))
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if !ok {
		zeroInBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(zeroInBytes, uint64(0))
		db.Put(key, zeroInBytes)
	}

	currentAmountInBytes, err := db.lvdb.Get(key, nil)
	currentAmount := binary.LittleEndian.Uint64(currentAmountInBytes)
	newAmount := currentAmount + amount

	newAmountInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newAmountInBytes, newAmount)
	err = db.Put(key, newAmountInBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	return nil
}

func (db *db) SendInitGOVVoteToken(startedBlock uint32, pubKey []byte, amount uint64) error {
	key := db.GetKey(string(GOVVoteTokenAmountPrefix), string(startedBlock)+string(pubKey))
	ok, err := db.HasValue(key)
	if err != nil {
		return err
	}
	if !ok {
		zeroInBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(zeroInBytes, uint64(0))
		db.Put(key, zeroInBytes)
	}

	currentAmountInBytes, err := db.lvdb.Get(key, nil)
	currentAmount := binary.LittleEndian.Uint64(currentAmountInBytes)
	newAmount := currentAmount + amount

	newAmountInBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(newAmountInBytes, newAmount)
	err = db.Put(key, newAmountInBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	return nil
}
