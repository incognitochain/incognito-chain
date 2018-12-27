package lvdb

import (
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
)

func (db *db) StoreMultiSigsRegistration(
	paymentAddressBytes []byte,
	multiSigsRegistrationBytes []byte,
) error {
	key := append(multisigsPrefix, paymentAddressBytes...)
	err := db.Put(key, multiSigsRegistrationBytes)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.put"))
	}
	return nil
}

func (db *db) GetMultiSigsRegistration(
	paymentAddressBytes []byte,
) ([]byte, error) {
	key := append(multisigsPrefix, paymentAddressBytes...)
	multisigsRegBytes, err := db.Get(key)
	return multisigsRegBytes, err
}
