package lvdb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/databasemp"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// Key: tx-{type}-{txHash}
// Value: transaction(byte value)-Splitter-otherDescValue(byte Value)
func (db *db) AddTransaction(txHash *common.Hash, txType string, valueTx []byte, valueDesc []byte) error {
	key := db.GetKey(txHash)
	value := append([]byte(txType), Splitter...)
	value = append(value, valueTx...)
	value = append(value, Splitter...)
	value = append(value, valueDesc...)
	if err := db.Put(key, value); err != nil {
		return databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "db.lvdb.Put"))
	}
	return nil
}
func (db *db) RemoveTransaction(txHash *common.Hash) error {
	key := db.GetKey(txHash)
	if err := db.Delete(key); err != nil {
		return databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}
	return nil
}

func (db *db) GetTransaction(txHash *common.Hash) ([]byte, error) {
	key := db.GetKey(txHash)
	value, err := db.Get(key)
	if err != nil {
		return []byte{}, databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return value, nil
}

func (db *db) HasTransaction(txHash *common.Hash) (bool, error) {
	key := db.GetKey(txHash)
	ret, err := db.HasValue(key)
	if err != nil {
		return false, databasemp.NewDatabaseMempoolError(databasemp.NotExistValue, err)
	}
	return ret, nil
}

func (db *db) Reset() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(txKeyPrefix), nil)
	for iter.Next() {
		err := db.Delete(iter.Key())
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

func (db *db) Load() ([][]byte, [][]byte, error) {
	txHashes := [][]byte{}
	txs := [][]byte{}
	iter := db.lvdb.NewIterator(util.BytesPrefix(txKeyPrefix), nil)
	for iter.Next() {
		key := iter.Key()
		newKey := make([]byte, len(key))
		copy(newKey, key)
		txHashes = append(txHashes, newKey)
		value := iter.Value()
		newValue := make([]byte, len(value))
		copy(newValue, value)
		txs = append(txs, newValue)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return txHashes, txs, databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return txHashes, txs, nil
}
