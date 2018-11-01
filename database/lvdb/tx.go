package lvdb

import (
	"bytes"
	"encoding/json"

	"github.com/ninjadotorg/constant/database"

	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (db *db) StoreNullifiers(nullifier []byte, coinType string, chainId byte) error {
	key := db.getKey("commitment", coinType)
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	txs = append(txs, nullifier)
	b, err := json.Marshal(txs)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) FetchNullifiers(coinType string, chainID byte) ([][]byte, error) {
	key := db.getKey("nullifier", coinType)
	key = append(key, chainID)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return make([][]byte, 0), errors.Wrap(err, "json.Unmarshal")
		}
	}
	return txs, nil
}

func (db *db) HasNullifier(nullifier []byte, coinType string, chainID byte) (bool, error) {
	listNullifiers, err := db.FetchNullifiers(coinType, chainID)
	if err != nil {
		return false, database.NewDatabaseError(database.UnexpectedError, err)
	}
	for _, item := range listNullifiers {
		if bytes.Equal(item, nullifier) {
			return true, nil
		}
	}
	return false, nil
}

func (db *db) CleanNullifiers() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(nullifiers), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

func (db *db) StoreCommitments(commitments []byte, coinType string, chainId byte) error {
	key := db.getKey("commitment", coinType)
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	txs = append(txs, commitments)
	b, err := json.Marshal(txs)
	if err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Marshal"))
	}
	if err := db.lvdb.Put(key, b, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) FetchCommitments(coinType string, chainId byte) ([][]byte, error) {
	key := db.getKey("commitment", coinType)
	key = append(key, chainId)
	res, err := db.lvdb.Get(key, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return make([][]byte, 0), database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "json.Unmarshal"))
		}
	}
	return txs, nil
}
func (db *db) HasCommitment(commitment []byte, coinType string, chainId byte) (bool, error) {
	listCommitments, err := db.FetchCommitments(coinType, chainId)
	if err != nil {
		return false, database.NewDatabaseError(database.UnexpectedError, err)
	}
	for _, item := range listCommitments {
		if bytes.Equal(item, commitment) {
			return true, nil
		}
	}
	return false, nil
}

func (db *db) CleanCommitments() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(commitments), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}

func (db *db) StoreFeeEstimator(val []byte, chainId byte) error {
	if err := db.put(append(feeEstimator, chainId), val); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.put"))
	}
	return nil
}

func (db *db) GetFeeEstimator(chainId byte) ([]byte, error) {
	b, err := db.lvdb.Get(append(feeEstimator, chainId), nil)
	if err != nil {
		return nil, database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
	}
	return b, err
}

func (db *db) CleanFeeEstimator() error {
	iter := db.lvdb.NewIterator(util.BytesPrefix(feeEstimator), nil)
	for iter.Next() {
		err := db.lvdb.Delete(iter.Key(), nil)
		if err != nil {
			return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "db.lvdb.Get"))
		}
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return database.NewDatabaseError(database.UnexpectedError, errors.Wrap(err, "iter.Error"))
	}
	return nil
}
