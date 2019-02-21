package lvdb

import (
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/pkg/errors"
)

func (db *db) SendInitVoteToken(boardType database.BoardTypeDB, boardIndex uint32, paymentAddress privacy.PaymentAddress, amount uint32) error {
	oldAmount, err := db.GetVoteTokenAmount(boardType, boardIndex, paymentAddress)
	if err != nil {
		oldAmount = 0
	}
	newAmount := oldAmount + amount
	err = db.SetVoteTokenAmount(boardType, boardIndex, paymentAddress, newAmount)

	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}

	return nil
}
