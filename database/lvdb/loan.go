package lvdb

import (
	"fmt"

	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
)

func (db *db) GetLoanWithdrawed(loanID []byte) (bool, error) {
	keyLoanWithdrawed := string(loanWithdrawedPrefix) + string(loanID)
	value, err := db.Get([]byte(keyLoanWithdrawed))
	if err != nil {
		return false, err
	}
	return value[0] > 0, err
}

func (db *db) StoreLoanWithdrawed(loanID []byte) error {
	keyLoanWithdrawed := string(loanWithdrawedPrefix) + string(loanID)
	valueLoanWithdrawed := []byte{1}
	if err := db.Put([]byte(keyLoanWithdrawed), valueLoanWithdrawed); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}

func (db *db) GetLoanRequestTx(loanID []byte) ([]byte, error) {
	keyLoanID := string(loanIDKeyPrefix) + string(loanID) + string(loanRequestPostfix)
	loanReqTx, err := db.Get([]byte(keyLoanID))
	return loanReqTx, err
}

func (db *db) StoreLoanPayment(loanID []byte, principle, interest uint64, deadline uint64) error {
	loanPaymentKey := string(loanPaymentKeyPrefix) + string(loanID)
	loanPaymentValue := getLoanPaymentValue(principle, interest, deadline)

	fmt.Printf("[db] Putting key %x, value %x\n", loanPaymentKey, loanPaymentValue)
	if err := db.Put([]byte(loanPaymentKey), []byte(loanPaymentValue)); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}

func (db *db) GetLoanPayment(loanID []byte) (uint64, uint64, uint64, error) {
	loanPaymentKey := string(loanPaymentKeyPrefix) + string(loanID)
	loanPaymentValue, err := db.Get([]byte(loanPaymentKey))
	if err != nil {
		return 0, 0, 0, err
	}

	fmt.Printf("Found payment %x: %x\n", loanPaymentKey, loanPaymentValue)
	return parseLoanPaymentValue(loanPaymentValue)
}
