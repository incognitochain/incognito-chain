package lvdb

import (
	"github.com/constant-money/constant-chain/database"
	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

func (db *db) StoreMultiSigsRegistration(
	pubKey []byte,
	multiSigsRegistrationBytes []byte,
) error {
	key := append(multisigsPrefix, pubKey...)
	err := db.Put(key, multiSigsRegistrationBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetMultiSigsRegistration(
	pubKey []byte,
) ([]byte, error) {
	key := append(multisigsPrefix, pubKey...)
	multisigsRegBytes, err := db.lvdb.Get(key, nil)
	if err != nil {
		if err != lvdberr.ErrNotFound {
			return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
		return []byte{}, nil
	}
	return multisigsRegBytes, nil
}
