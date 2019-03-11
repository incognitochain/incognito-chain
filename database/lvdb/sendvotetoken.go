package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/pkg/errors"
)

func (db *db) SendInitVoteToken(boardType common.BoardType, boardIndex uint32, paymentAddress privacy.PaymentAddress, amount uint32) error {
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
