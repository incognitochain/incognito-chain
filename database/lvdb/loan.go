package lvdb

import (
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

/*
 Store loan id with tx hash:
 - for loan request
	key1: loanID-ID-req
	value1: txHash

 - for load response
	key2: loanID-ID-res
	value2: txHash

 - for tx hash related to a loan ID
	key3: loanTx-txHash
  value3: loanId-req/res
*/
func (db *db) StoreLoanRequest(loanID, txHash []byte) error {
	keyLoanID := string(loanIDKeyPrefix) + string(loanID) + string(loanRequestPostfix)
	valueLoanID := string(txHash)

	if ok, _ := db.HasValue([]byte(keyLoanID)); ok {
		return database.NewDatabaseError(database.KeyExisted, errors.Errorf("loan ID existed %+v", keyLoanID))
	}
	if err := db.Put([]byte(keyLoanID), []byte(valueLoanID)); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}

	keyTxHash := string(loanTxKeyPrefix) + string(txHash)
	valueTxHash := string(loanID) + string(loanRequestPostfix)
	if ok, _ := db.HasValue([]byte(keyTxHash)); ok {
		return database.NewDatabaseError(database.KeyExisted, errors.Errorf("loan transaction hash existed %+v", keyTxHash))
	}
	if err := db.Put([]byte(keyTxHash), []byte(valueTxHash)); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}

func (db *db) StoreLoanResponse(loanID, txHash []byte) error {
	keyLoanID := string(loanIDKeyPrefix) + string(loanID) + string(loanResponsePostfix)
	valueLoanID := string(txHash)
	if ok, _ := db.HasValue([]byte(keyLoanID)); ok {
		return database.NewDatabaseError(database.KeyExisted, errors.Errorf("loan ID existed %+v", keyLoanID))
	}
	if err := db.Put([]byte(keyLoanID), []byte(valueLoanID)); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}

	keyTxHash := string(loanTxKeyPrefix) + string(txHash)
	valueTxHash := string(loanID) + string(loanResponsePostfix)
	if ok, _ := db.HasValue([]byte(keyTxHash)); ok {
		return database.NewDatabaseError(database.KeyExisted, errors.Errorf("loan transaction hash existed %+v", keyTxHash))
	}
	if err := db.Put([]byte(keyTxHash), []byte(valueTxHash)); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}

func (db *db) GetLoanTxs(loanID []byte) ([][]byte, error) {
	loanIdPrefix := string(loanIDKeyPrefix) + string(loanID)
	iter := db.lvdb.NewIterator(util.BytesPrefix([]byte(loanIdPrefix)), nil)
	results := [][]byte{}
	for iter.Next() {
		value := make([]byte, len(iter.Value()))
		copy(value, iter.Value())
		results = append(results, value)
	}
	iter.Release()
	return results, nil
}

func (db *db) StoreLoanPayment(loanID []byte, principle, interest uint64, deadline uint32) error {
	// TODO(@0xbunyip): implement
	return nil
}

func (db *db) GetLoanPayment(loanID []byte) (uint64, uint64, uint32, error) {
	// TODO(@0xbunyip): implement
	return 0, 0, 0, nil
}
