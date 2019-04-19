package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/databasemp"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Key: tx-{type}-{txHash}
// Value: transaction(byte value)-Splitter-otherDescValue(byte Value)
func (db *db) AddTransaction(txHash *common.Hash, txType string, valueTx []byte, valueDesc []byte) error {
	key := db.GetKey(txHash)
	value := append([]byte(txType),Splitter...)
	value = append(value, valueTx...)
	value = append(value, Splitter...)
	value = append(value, valueDesc...)
	if err := db.lvdb.Put(key,value, nil); err != nil {
		return databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "db.lvdb.Put"))
	}
	return nil
}
func (db *db) RemoveTransaction(txHash *common.Hash) error {
	key := db.GetKey(txHash)
	if err := db.lvdb.Delete(key, nil); err != nil {
		return databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}
	return nil
}

func (db *db) GetTransaction(txHash *common.Hash) ([]byte, error){
	key := db.GetKey(txHash)
	value, err := db.lvdb.Get(key,nil)
	if err != nil {
		return []byte{}, databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return value, nil
}

func (db *db) HasTransaction(txHash *common.Hash) (bool,error) {
	key := db.GetKey(txHash)
	ret, err := db.lvdb.Has(key, nil)
	if err != nil {
		return false, databasemp.NewDatabaseMempoolError(databasemp.NotExistValue, err)
	}
	return ret, nil
}

func (db *db) Reset() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(txKeyPrefix), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

func (db *db) Load() ([][]byte, error) {
	txs := [][]byte{}
	iter := db.lvdb.NewIterator(util.BytesPrefix(txKeyPrefix), nil)
	for iter.Next() {
		value := iter.Value()
		newValue := make([]byte, len(value))
		copy(newValue, value)
		txs = append(txs, newValue)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return txs, databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return txs, nil
}
