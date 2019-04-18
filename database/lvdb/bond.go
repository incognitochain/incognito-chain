package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) StoreSoldBondTypes(
	bondID *common.Hash,
	sellingBondsParamsBytes []byte,
) error {
	key := append(bondTypePrefix, bondID[:]...)
	err := db.Put(key, sellingBondsParamsBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetSoldBondTypes() ([][]byte, error) {
	iter := db.NewIterator(util.BytesPrefix(bondTypePrefix), nil)
	results := [][]byte{}
	for iter.Next() {
		value := iter.Value()
		results = append(results, value)
	}
	iter.Release()
	err := iter.Error()
	if err != nil {
		if err != lvdberr.ErrNotFound {
			return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
		return [][]byte{}, nil
	}
	return results, nil
}
