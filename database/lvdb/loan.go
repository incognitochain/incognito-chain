package lvdb

import (
	"fmt"

	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
)

func (db *db) GetLoanRequestTx(loanID []byte) ([]byte, error) {
	keyLoanID := string(loanIDKeyPrefix) + string(loanID) + string(loanRequestPostfix)
	//	if ok, _ := db.HasValue([]byte(keyLoanID)); !ok {
	//		return nil, database.NewDatabaseError(database.KeyExisted, errors.Errorf("loan ID not existed %+v", keyLoanID))
	//	}
	loanReqTx, err := db.Get([]byte(keyLoanID))
	return loanReqTx, err
}

func (db *db) StoreLoanPayment(loanID []byte, principle, interest uint64, deadline uint64) error {
	loanPaymentKey := string(loanPaymentKeyPrefix) + string(loanID)
	loanPaymentValue := getLoanPaymentValue(principle, interest, deadline)

	fmt.Printf("Putting key %x, value %x\n", loanPaymentKey, loanPaymentValue)
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
