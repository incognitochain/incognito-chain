package lvdb

import (
	"github.com/ninjadotorg/constant/database"
	"github.com/pkg/errors"
)

func (db *db) StoreCMB(mainAccount []byte, members [][]byte, capital uint64) error {
	ok, err := db.HasValue(mainAccount)
	if err != nil {
		return errUnexpected(err, "error retrieving cmb")
	}
	if !ok {
		return database.NewDatabaseError(database.KeyExisted, errors.Errorf("CMB main account existed"))
	}

	cmbValue, err := getCMBValue(capital, members)
	if err != nil {
		return errUnexpected(err, "getCMBValue")
	}
	if err := db.Put(mainAccount, cmbValue); err != nil {
		return errUnexpected(err, "put cmb main account")
	}
	return nil
}

func (db *db) GetCMB(saleID []byte) (int32, []byte, uint64, []byte, uint64, error) {
	return 0, nil, 0, nil, 0, nil
}
