package lvdb

import (
	"strings"

	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) updateCMB(
	cmbInitKey []byte,
	reserve []byte,
	members [][]byte,
	capital uint64,
	txHash []byte,
	state uint8,
	fine uint64,
) error {
	newValue, err := getCMBInitValue(reserve, members, capital, txHash, state, fine)
	if err != nil {
		return errUnexpected(err, "getCMBValue")
	}

	if err := db.Put(cmbInitKey, newValue); err != nil {
		return errUnexpected(err, "put cmb main account")
	}
	return nil
}

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
	fine := uint64(0)
	return db.updateCMB(cmbInitKey, reserveAccount, members, capital, txHash, state, fine)
}

func (db *db) GetCMB(mainAccount []byte) ([]byte, [][]byte, uint64, []byte, uint8, uint64, error) {
	cmbInitKey := getCMBInitKey(mainAccount)
	cmbInitValue, err := db.Get(cmbInitKey)
	if err != nil {
		return nil, nil, 0, nil, 0, 0, err
	}
	return parseCMBInitValue(cmbInitValue)
}

func (db *db) UpdateCMBState(mainAccount []byte, state uint8) error {
	cmbInitKey := getCMBInitKey(mainAccount)
	cmbInitValue, err := db.Get(cmbInitKey)
	if err != nil {
		return err
	}
	reserve, members, capital, txHash, _, fine, err := parseCMBInitValue(cmbInitValue)
	if err != nil {
		return err
	}
	return db.updateCMB(cmbInitKey, reserve, members, capital, txHash, state, fine)
}

func (db *db) UpdateCMBFine(mainAccount []byte, fine uint64) error {
	cmbInitKey := getCMBInitKey(mainAccount)
	cmbInitValue, err := db.Get(cmbInitKey)
	if err != nil {
		return err
	}
	reserve, members, capital, txHash, state, _, err := parseCMBInitValue(cmbInitValue)
	if err != nil {
		return err
	}
	return db.updateCMB(cmbInitKey, reserve, members, capital, txHash, state, fine)
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

func (db *db) StoreWithdrawRequest(contractID []byte, txHash []byte) error {
	cmbWithdrawRequestKey := getCMBWithdrawRequestKey(contractID)
	state := metadata.WithdrawRequested
	cmbWithdrawRequestValue := getCMBWithdrawRequestValue(txHash, state)
	if err := db.Put(cmbWithdrawRequestKey, cmbWithdrawRequestValue); err != nil {
		return errUnexpected(err, "put cmb withdraw request")
	}
	return nil
}

func (db *db) GetWithdrawRequest(contractID []byte) ([]byte, uint8, error) {
	cmbWithdrawRequestKey := getCMBWithdrawRequestKey(contractID)
	cmbWithdrawRequestValue, err := db.Get(cmbWithdrawRequestKey)
	if err != nil {
		return nil, 0, err
	}
	return parseWithdrawRequestValue(cmbWithdrawRequestValue)
}

func (db *db) UpdateWithdrawRequestState(contractID []byte, state uint8) error {
	cmbWithdrawRequestKey := getCMBWithdrawRequestKey(contractID)
	cmbWithdrawRequestValue, err := db.Get(cmbWithdrawRequestKey)
	if err != nil {
		return err
	}
	txHash, _, err := parseWithdrawRequestValue(cmbWithdrawRequestValue)
	if err != nil {
		return errUnexpected(err, "parseWithdrawRequestValue")
	}
	newValue := getCMBWithdrawRequestValue(txHash, state)
	if err := db.Put(cmbWithdrawRequestKey, newValue); err != nil {
		return errUnexpected(err, "put cmb withdraw request")
	}
	return nil
}

func (db *db) StoreNoticePeriod(blockHeight uint64, txReqHash []byte) error {
	cmbNoticeKey := getCMBNoticeKey(blockHeight, txReqHash)
	cmbNoticeValue := []byte{1}
	if err := db.Put(cmbNoticeKey, cmbNoticeValue); err != nil {
		return errUnexpected(err, "put cmb notice")
	}
	return nil
}

func (db *db) GetNoticePeriod(blockHeight uint64) ([][]byte, error) {
	txReqHash := []byte{} // empty hash to get all
	txHashes := [][]byte{}
	cmbNoticeKey := getCMBNoticeKey(blockHeight, txReqHash)
	iter := db.lvdb.NewIterator(util.BytesPrefix(cmbNoticeKey), nil)
	for iter.Next() {
		key := string(iter.Key())
		keys := strings.Split(key, string(Splitter)) // cmbnotice-blockHeight-[-]-txHash
		txHashes = append(txHashes, []byte(keys[1]))
	}
	iter.Release()
	return txHashes, nil
}
