package lvdb

import (
	"bytes"
	"encoding/json"

	"github.com/ninjadotorg/cash-prototype/database"

	"github.com/pkg/errors"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
)

func (db *db) StoreNullifiers(nullifier []byte, typeJoinSplitDesc string, chainId byte) error {
	key := db.getKey("commitment", typeJoinSplitDesc)
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

func (db *db) FetchNullifiers(typeJoinSplitDesc string, chainId byte) ([][]byte, error) {
	key := db.getKey("nullifier", typeJoinSplitDesc)
	key = append(key, chainId)
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

func (db *db) HasNullifier(nullifier []byte, typeJoinSplitDesc string, chainId byte) (bool, error) {
	listNullifiers, err := db.FetchNullifiers(typeJoinSplitDesc, chainId)
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
	return nil
}

func (db *db) StoreCommitments(commitments []byte, typeJoinSplitDesc string, chainId byte) error {
	key := db.getKey("commitment", typeJoinSplitDesc)
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

func (db *db) FetchCommitments(typeJoinSplitDesc string, chainId byte) ([][]byte, error) {
	key := db.getKey("commitment", typeJoinSplitDesc)
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
func (db *db) HasCommitment(commitment []byte, typeJoinSplitDesc string, chainId byte) (bool, error) {
	listCommitments, err := db.FetchCommitments(typeJoinSplitDesc, chainId)
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
	return nil
}
