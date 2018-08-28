package lvdb

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
)

func open(dbPath string) (database.DB, error) {
	ldb, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, errors.Wrapf(err, "leveldb.OpenFile %s", dbPath)
	}
	return &db{ldb}, nil
}

type db struct {
	ldb *leveldb.DB
}

func (db *db) hasKey(key []byte) (bool, error) {
	if key == nil {
		return false, nil
	}
	res, err := db.ldb.Get(key, nil)
	if err != nil {
		return false, errors.Wrap(err, "db.ldb.Get")
	}
	if res == nil {
		return false, nil
	}
	return true, nil
}

func (db *db) hasBlock(key []byte) bool {
	ret, err := db.hasKey(key)
	if err != nil {
		return false
	}
	return ret
}

func (db *db) StoreBlock(b *blockchain.Block) error {
	var (
		hash = b.Hash()
		key  = hash[:]
	)
	if db.hasBlock(key) {
		return errors.Errorf("block %s already exists", hash.String())
	}
	val, err := json.Marshal(b)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}
	if err := db.put(key, val); err != nil {
		return errors.Wrap(err, "db.Put")
	}
	return nil
}

func (db *db) Close() error {
	return errors.Wrap(db.ldb.Close(), "db.ldb.Close")
}

func (db *db) put(key, value []byte) error {
	if err := db.ldb.Put(key, value, nil); err != nil {
		return errors.Wrap(err, "db.ldb.Put")
	}
	return nil
}

func (db *db) HasBlock(hash *common.Hash) (bool, error) {
	if exists := db.hasBlock(hash[:]); exists {
		return true, nil
	}
	return false, nil
}

func (db *db) FetchBlock(hash *common.Hash) ([]byte, error) {
	block, err := db.ldb.Get(hash[:], nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.ldb.Get")
	}
	return block, nil
}
