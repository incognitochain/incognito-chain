package lvdb

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/databasemp"
	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
)

type db struct {
	lvdb *leveldb.DB
}

var (
	txKeyPrefix = []byte("tx-")
	Splitter    = []byte("-[-]-")
)

func (db *db) GetKey(key interface{}) []byte {
	var dbkey []byte
	dbkey = append(txKeyPrefix, key.(*common.Hash)[:]...)
	return dbkey
}

func open(dbPath string) (databasemp.DatabaseInterface, error) {
	lvdb, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, databasemp.NewDatabaseMempoolError(databasemp.OpenDbErr, errors.Wrapf(err, "levelvdb.OpenFile %s", dbPath))
	}
	return &db{lvdb: lvdb}, nil
}

func (db *db) Close() error {
	return errors.Wrap(db.lvdb.Close(), "db.lvdb.Close")
}

func (db *db) HasValue(key []byte) (bool, error) {
	ret, err := db.lvdb.Has(key, nil)
	if err != nil {
		return false, databasemp.NewDatabaseMempoolError(databasemp.NotExistValue, err)
	}
	return ret, nil
}

func (db *db) Put(key, value []byte) error {
	if err := db.Put(key, value); err != nil {
		return databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "db.lvdb.Put"))
	}
	return nil
}

func (db *db) Delete(key []byte) error {
	err := db.lvdb.Delete(key, nil)
	if err != nil {
		return databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "db.lvdb.Delete"))
	}
	return nil
}

func (db *db) Get(key []byte) ([]byte, error) {
	value, err := db.Get(key)
	if err != nil {
		return nil, databasemp.NewDatabaseMempoolError(databasemp.LvDbNotFound, errors.Wrap(err, "db.lvdb.Get"))
	}
	return value, nil
}
