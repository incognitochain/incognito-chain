package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/databasemp"
	"github.com/pkg/errors"
)

func (db *db) AddTransaction(txHash *common.Hash, value []byte) error {
	key := db.GetKey(txHash)
	if err := db.lvdb.Put(key,value, nil); err != nil {
		return databasemp.NewDatabaseMempoolError(databasemp.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return nil
}

func (db *db) RemoveTransaction(txHash *common.Hash) error {
	return nil
}

func (db *db) GetTransaction(txHash *common.Hash) ([]byte, error){
	return []byte{}, nil
}

func (db *db) HasTransaction(txHash *common.Hash) error {
	return nil
}

func (db *db) Reset() error {
	return nil
}
