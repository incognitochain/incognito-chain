package lvdb

import (
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
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
	multisigsRegBytes, err := db.Get(key)
	return multisigsRegBytes, err
}
