package lvdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
	"github.com/ninjadotorg/constant/transaction"
	"strconv"
	"strings"

	"github.com/ninjadotorg/constant/database"

	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) StoreNullifiers(nullifier []byte, chainId byte) error {
	key := db.getKey(string(nullifiersPrefix), "")
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	txs = append(txs, nullifier)
	b, err := json.Marshal(txs)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) FetchNullifiers(chainID byte) ([][]byte, error) {
	key := db.getKey(string(nullifiersPrefix), "")
	key = append(key, chainID)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return make([][]byte, 0), errors.Wrap(err, "json.Unmarshal")
		}
	}
	return txs, nil
}

func (db *db) HasNullifier(nullifier []byte, chainID byte) (bool, error) {
	listNullifiers, err := db.FetchNullifiers(chainID)
	if err != nil {
		return false, database.NewDatabaseError(database.UnexpectedError, err)
	}
	for _, item := range listNullifiers {
		if bytes.Equal(item, nullifier) {
			return true, nil
		}
	}
	return false, nil
}

func (db *db) CleanNullifiers() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(nullifiersPrefix), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

func (db *db) StoreCommitments(commitments []byte, chainId byte) error {
	key := db.getKey(string(commitmentsPrefix), "")
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	txs = append(txs, commitments)
	b, err := json.Marshal(txs)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) FetchCommitments(chainId byte) ([][]byte, error) {
	key := db.getKey(string(commitmentsPrefix), "")
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	return txs, nil
}
func (db *db) HasCommitment(commitment []byte, chainId byte) (bool, error) {
	listCommitments, err := db.FetchCommitments(chainId)
	if err != nil {
		return false, database.NewDatabaseError(database.UnexpectedError, err)
	}
	for _, item := range listCommitments {
		if bytes.Equal(item, commitment) {
			return true, nil
		}
	}
	return false, nil
}

func (db *db) CleanCommitments() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(commitmentsPrefix), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

func (db *db) StoreFeeEstimator(val []byte, chainId byte) error {
	if err := db.put(append(feeEstimator, chainId), val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.put"))
	}
	return nil
}

func (db *db) GetFeeEstimator(chainId byte) ([]byte, error) {
	b, err := db.lvdb.Get(append(feeEstimator, chainId), nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return b, err
}

func (db *db) CleanFeeEstimator() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(feeEstimator), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

/*
	StoreTransactionIndex
	Store tx detail location
  Key: prefixTx-txHash
	Value: blockHash-blockIndex
*/
func (db *db) StoreTransactionIndex(txId *common.Hash, blockHash *common.Hash, index int) error {
	key := string(transactionKeyPrefix) + txId.String()
	value := blockHash.String() + string(spliter) + strconv.Itoa(index)
	fmt.Println("Key in StoreTransactionIndex", key)
	fmt.Println("Value in StoreTransactionIndex", value)
	if err := db.lvdb.Put([]byte(key), []byte(value), nil); err != nil {
		return err
	}

	return nil
}

/*
  Get Transaction by ID
*/

func (db *db) GetTransactionIndexById(txId *common.Hash)  (*common.Hash, int, error) {
	fmt.Println("TxID in GetTransactionById", txId.String())
	key := string(transactionKeyPrefix) + txId.String()
	_, err := db.hasValue([]byte(key))
	if err != nil {
		fmt.Println("ERROR in finding transaction id",txId.String(), err)
		return nil, -1, err
	}
	res, err := db.lvdb.Get([]byte(key), nil)
	if err != nil {
		return nil, -1, err;
	}
	reses := strings.Split(string(res),(string(spliter)))
	hash, err :=  common.Hash{}.NewHashFromStr(reses[0])
	if err != nil {
		return nil, -1, err;
	}
	index, err := strconv.Atoi(reses[1])
	if err != nil {
		return nil, -1, err;
	}
	fmt.Println("BlockHash", hash, "Transaction index", index)
	return hash, index, nil
}
/*
	Store Transaction in Light mode
	1. Key -> value : prefix(privateky)privateKey-chainId-(999999999 - blockHeight)-(999999999 - txIndex) 		-> 		tx
	2. Key -> value :							prefix(transaction)txHash 												->  	privateKey-chainId-blockHeight-txIndex

*/
func (db *db)  StoreTransactionLightMode(privateKey *client.SpendingKey, chainId byte, blockHeight int32, txIndex int, tx *transaction.Tx) error {
	const(
		bigNumber = 999999999
	)
	reverseBlockHeight := bigNumber - blockHeight
	reverseTxIndex := bigNumber - txIndex
	key1 := string(privateKeyPrefix) + privateKey.String() + string(spliter) + string(chainId) + string(reverseBlockHeight) + string(reverseTxIndex)
	key2 := string(privateKeyPrefix) + tx.Hash().String()

	if ok, _ := db.hasValue([]byte(key1)); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("tx %s already exists", key1))
	}
	if ok, _ := db.hasValue([]byte(key2)); ok {
		return database.NewDatabaseError(database.BlockExisted, errors.Errorf("tx %s already exists", key2))
	}

	value,err := json.Marshal(tx)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.put([]byte(key1), value); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	if err := db.put([]byte(key2), []byte(key1)); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.Put"))
	}
	return nil
}
