package lvdb

import (
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/pkg/errors"
)

func (db *db) StoreCMB(
	mainAccount []byte,
	members [][]byte,
	capital uint64,
	txHash []byte,
) error {
	ok, err := db.HasValue(mainAccount)
	if err != nil {
		return errUnexpected(err, "error retrieving cmb")
	}
	if !ok {
		return database.NewDatabaseError(database.KeyExisted, errors.Errorf("CMB main account existed"))
	}
	cmbInitKey := getCMBInitKey(mainAccount)

	state := metadata.CMBRequested
	cmbValue, err := getCMBInitValue(capital, members, txHash, state)
	if err != nil {
		return errUnexpected(err, "getCMBValue")
	}

	if err := db.Put(cmbInitKey, cmbValue); err != nil {
		return errUnexpected(err, "put cmb main account")
	}
	return nil
}

func (db *db) GetCMB(mainAccount []byte) ([][]byte, uint64, []byte, uint8, error) {
	cmbInitValue, err := db.Get(mainAccount)
	if err != nil {
		return nil, 0, nil, 0, err
	}
	return parseCMBInitValue(cmbInitValue)
}
