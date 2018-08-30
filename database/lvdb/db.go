package lvdb

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/syndtr/goleveldb/leveldb"
	lvdberr "github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/syndtr/goleveldb/leveldb/util"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/database"
)

var (
	blockKeyPrefix = []byte("b-")
	usedTxKey      = []byte("usedTx")
	bestBlockKey   = []byte("bestBlock")
)

func open(dbPath string) (database.DB, error) {
	ldb, err := leveldb.OpenFile(filepath.Join(dbPath, "db"), nil)
	if err != nil {
		return nil, errors.Wrapf(err, "leveldb.OpenFile %s", dbPath)
	}
	return &db{ldb: ldb}, nil
}

type db struct {
	ldb *leveldb.DB
}

func (db *db) hasBlock(key []byte) bool {
	ret, err := db.ldb.Has(key, nil)
	if err != nil {
		return false
	}
	return ret
}

type hasher interface {
	Hash() *common.Hash
}

func (db *db) StoreBlock(v interface{}) error {
	h, ok := v.(hasher)
	if !ok {
		return errors.New("v must implement Hash() method")
	}
	var (
		hash = h.Hash()
		key  = append(blockKeyPrefix, hash[:]...)
	)
	if db.hasBlock(key) {
		return errors.Errorf("block %s already exists", hash.String())
	}
	val, err := json.Marshal(v)
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
	if exists := db.hasBlock(db.getKey(hash)); exists {
		return true, nil
	}
	return false, nil
}

func (db *db) FetchBlock(hash *common.Hash) ([]byte, error) {
	block, err := db.ldb.Get(db.getKey(hash), nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.ldb.Get")
	}
	return block, nil
}

func (db *db) StoreTx(tx []byte) error {
	res, err := db.ldb.Get(usedTxKey, nil)
	if err != nil && err != lvdberr.ErrNotFound {
		return errors.Wrap(err, "db.ldb.Get")
	}

	var txs [][]byte
	if len(res) > 0 {
		if err := json.Unmarshal(res, &txs); err != nil {
			return errors.Wrap(err, "json.Unmarshal")
		}
	}
	txs = append(txs, tx)
	b, err := json.Marshal(txs)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}
	if err := db.ldb.Put(usedTxKey, b, nil); err != nil {
		return err
	}
	return nil
}

func (db *db) StoreBestBlock(v interface{}) error {
	val, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}
	if err := db.put(bestBlockKey, val); err != nil {
		return errors.Wrap(err, "db.Put")
	}
	return nil
}

func (db *db) FetchBestBlock() ([]byte, error) {
	block, err := db.ldb.Get(bestBlockKey, nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.ldb.Get")
	}
	return block, nil
}

func (db *db) StoreBlockIndex(h *common.Hash, idx int32) error {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, idx); err != nil {
		return errors.Wrapf(err, "binary.Write %d", idx)
	}

	if err := db.ldb.Put(db.getKey(h), buf.Bytes(), nil); err != nil {
		return errors.Wrap(err, "db.ldb.Put")
	}
	if err := db.ldb.Put(buf.Bytes(), h[:], nil); err != nil {
		return errors.Wrap(err, "db.ldb.Put")
	}
	return nil
}

func (db *db) GetIndexOfBlock(h *common.Hash) (int32, error) {
	b, err := db.ldb.Get(db.getKey(h), nil)
	if err != nil {
		return 0, errors.Wrap(err, "db.ldb.Get")
	}

	var idx int32
	if err := binary.Read(bytes.NewReader(b), binary.LittleEndian, &idx); err != nil {
		return 0, errors.Wrap(err, "binary.Read")
	}
	return idx, nil
}

func (db *db) GetBlockByIndex(idx int32) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, idx); err != nil {
		return nil, errors.Wrapf(err, "binary.Write %d", idx)
	}
	b, err := db.ldb.Get(buf.Bytes(), nil)
	if err != nil {
		return nil, errors.Wrap(err, "db.ldb.Get")
	}
	return b, nil
}

func (db *db) FetchAllBlocks() ([]*common.Hash, error) {
	var keys []*common.Hash
	iter := db.ldb.NewIterator(util.BytesPrefix(blockKeyPrefix), nil)
	for iter.Next() {
		var h common.Hash
		key := iter.Key()[len(blockKeyPrefix):]
		for i, b := range key {
			h[i] = b
		}
		keys = append(keys, &h)
	}
	iter.Release()
	if err := iter.Error(); err != nil {
		return nil, errors.Wrap(err, "iter.Error")
	}
	return keys, nil
}

func (db *db) getKey(h *common.Hash) []byte {
	var key []byte
	key = append(blockKeyPrefix, h[:]...)
	return key
}
