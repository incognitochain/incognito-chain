package lvdb

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
)

func (db *db) SendInitVoteToken(boardType string, boardIndex uint32, paymentAddress privacy.PaymentAddress, amount uint32) error {
	key := GetKeyVoteTokenAmount(boardType, boardIndex, paymentAddress)
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
