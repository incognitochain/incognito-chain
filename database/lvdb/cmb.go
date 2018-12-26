package lvdb

import (
	"strings"

	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) StoreCMB(
	mainAccount []byte,
	reserveAccount []byte,
	members [][]byte,
	capital uint64,
	txHash []byte,
) error {
	ok, err := db.HasValue(mainAccount)
	if err != nil {
		return errUnexpected(err, "error retrieving cmb")
	}
	if !ok {
		return database.NewDatabaseError(database.KeyExisted, errors.Errorf("CMB main account existed"))
	}
	cmbInitKey := getCMBInitKey(mainAccount)

	state := metadata.CMBRequested
	cmbValue, err := getCMBInitValue(reserveAccount, members, capital, txHash, state)
	if err != nil {
		return errUnexpected(err, "getCMBValue")
	}

	if err := db.Put(cmbInitKey, cmbValue); err != nil {
		return errUnexpected(err, "put cmb main account")
	}
	return nil
}

func (db *db) GetCMB(mainAccount []byte) ([]byte, [][]byte, uint64, []byte, uint8, error) {
	cmbInitKey := getCMBInitKey(mainAccount)
	cmbInitValue, err := db.Get(cmbInitKey)
	if err != nil {
		return nil, nil, 0, nil, 0, err
	}
	return parseCMBInitValue(cmbInitValue)
}

func (db *db) UpdateCMBState(mainAccount []byte, state uint8) error {
	cmbInitKey := getCMBInitKey(mainAccount)
	cmbInitValue, err := db.Get(cmbInitKey)
	if err != nil {
		return err
	}
	reserve, members, capital, txHash, _, err := parseCMBInitValue(cmbInitValue)
	newValue, err := getCMBInitValue(reserve, members, capital, txHash, state)
	if err != nil {
		return errUnexpected(err, "getCMBValue")
	}

	if err := db.Put(cmbInitKey, newValue); err != nil {
		return errUnexpected(err, "put cmb main account")
	}
	return nil
}

func (db *db) StoreCMBResponse(mainAccount, approver []byte) error {
	cmbResponseKey := getCMBResponseKey(mainAccount, approver)
	cmbResponseValue := []byte{1}
	if err := db.Put(cmbResponseKey, cmbResponseValue); err != nil {
		return errUnexpected(err, "put cmb response")
	}
	return nil
}

func (db *db) GetCMBResponse(mainAccount []byte) ([][]byte, error) {
	approver := []byte{} // empty approver to get all
	approvers := [][]byte{}
	cmbResponseKey := getCMBResponseKey(mainAccount, approver)
	iter := db.lvdb.NewIterator(util.BytesPrefix(cmbResponseKey), nil)
	for iter.Next() {
		key := string(iter.Key())
		keys := strings.Split(key, string(Splitter)) // cmbres-mainAccount-[-]-approver
		approvers = append(approvers, []byte(keys[1]))
	}
	iter.Release()
	return approvers, nil
}

func (db *db) StoreDepositSend(contractID []byte, txHash []byte) error {
	cmbDepositSendKey := getCMBDepositSendKey(contractID)
	cmbDepositSendValue := getCMBDepositSendValue(txHash)
	if err := db.Put(cmbDepositSendKey, cmbDepositSendValue); err != nil {
		return errUnexpected(err, "put cmb deposit send")
	}
	return nil
}

func (db *db) GetDepositSend(contractID []byte) ([]byte, error) {
	cmbDepositSendKey := getCMBDepositSendKey(contractID)
	cmbDepositSendValue, err := db.Get(cmbDepositSendKey)
	if err != nil {
		return nil, err
	}
	return cmbDepositSendValue, nil
}
