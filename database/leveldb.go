package database

import (
	"path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
	"sync"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"fmt"
)

func init() {
	dbCreator := func(name string, dir string, cache int, handles int) (DB, error) {
		return NewLevelDB(name, dir, cache, handles)
	}
	registerDBCreator(LevelDBBackend, dbCreator, false)
}

var _ DB = (*LevelDB)(nil)

type LevelDB struct {
	db *leveldb.DB

	quitLock sync.Mutex      // Mutex protecting the quit channel access
	quitChan chan chan error // Quit channel to stop the metrics collection before closing the database

}

func NewLevelDB(name string, dir string, cache int, handles int) (*LevelDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	_data, err := leveldb.OpenFile(dbPath, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	})
	if err != nil {
		return nil, err
	}
	database := &LevelDB{
		db: _data,
	}
	return database, nil
}


func nonNilBytes(bz []byte) []byte {
	if bz == nil {
		return []byte{}
	}
	return bz
}



//get block
func (db LevelDB) Has(key []byte) (bool, error) {
	key = nonNilBytes(key)
	res, err := db.db.Get(key, nil)
	if err != nil {
		return false, err
	}

	return (res != nil) , nil
}

func (db *LevelDB) Get(key []byte) ([]byte, error) {
	key = nonNilBytes(key)
	res, err := db.db.Get(key, nil)
	if err != nil {
		return nil, err
	}

	return res , nil
}

func (db *LevelDB) Delete(key []byte)  (bool, error) {
	key = nonNilBytes(key)
	err := db.db.Delete(key, nil)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (db *LevelDB) Put(key []byte, value []byte)  (bool, error) {
	value = nonNilBytes(value)
	err := db.db.Put(key, value, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (db *LevelDB) NewIterator() iterator.Iterator {
	return db.db.NewIterator(nil, nil)
}

// NewIteratorWithPrefix returns a iterator to iterate over subset of database content with a particular prefix.
func (db *LevelDB) NewIteratorWithPrefix(prefix []byte) iterator.Iterator {
	return db.db.NewIterator(util.BytesPrefix(prefix), nil)
}

func (db *LevelDB) Close() {
	db.quitLock.Lock()
	defer db.quitLock.Unlock()

	if db.quitChan != nil {
		errc := make(chan error)
		db.quitChan <- errc
		if err := <-errc; err != nil {
			fmt.Print("Metrics collection failed", "err", err)
		}
		db.quitChan = nil
	}
	err := db.db.Close()
	if err == nil {
		fmt.Print("Database closed")
	} else {
		fmt.Print("Failed to close database", "err", err)
	}
}

type lBDBatch struct {
	db   *leveldb.DB
	b    *leveldb.Batch
	size int
}

func (lbdb *lBDBatch) Put(key []byte, value []byte) (bool, error) {
	lbdb.b.Put(key, value)
	lbdb.size += len(value)
	return true, nil
}

func (lbdb *lBDBatch) Delete(key []byte) (bool, error) {
	lbdb.b.Delete(key)
	lbdb.size += 1
	return true, nil
}

func (lbdb *lBDBatch) ValueSize() int {
	return lbdb.size
}

func (lbdb *lBDBatch) Write() error {
	return lbdb.db.Write(lbdb.b, nil)
}

func (lbdb *lBDBatch) Reset() {
	lbdb.b.Reset()
	lbdb.size = 0
}

func (db *LevelDB) NewBatch() Batch {
	batch :=  new(leveldb.Batch)
	return &lBDBatch{db: db.db, b: batch}
}

type table struct {
	db     LevelDB
	prefix string
}

func NewTable(db LevelDB, prefix string) DB {
	return &table{
		db:     db,
		prefix: prefix,
	}
}

func (dt *table) Put(key []byte, value []byte) (bool, error) {
	return dt.db.Put(append([]byte(dt.prefix), key...), value)
}

func (dt *table) Has(key []byte) (bool, error) {
	return dt.db.Has(append([]byte(dt.prefix), key...))
}

func (dt *table) Get(key []byte) ([]byte, error) {
	return dt.db.Get(append([]byte(dt.prefix), key...))
}

func (dt *table) Delete(key []byte) (bool, error) {
	return dt.db.Delete(append([]byte(dt.prefix), key...))
}

func (dt *table) Close() {
	//@todo nothing todo :)
}

type tableBatch struct {
	batch  Batch
	prefix string
}

// NewTableBatch returns a Batch object which prefixes all keys with a given string.
func NewTableBatch(db LevelDB, prefix string) Batch {
	return &tableBatch{db.NewBatch(), prefix}
}

func (dt *table) NewBatch() Batch {
	return &tableBatch{dt.db.NewBatch(), dt.prefix}
}

func (tb *tableBatch) Put(key, value []byte) (bool, error) {
	return tb.batch.Put(append([]byte(tb.prefix), key...), value)
}

func (tb *tableBatch) Delete(key []byte) (bool, error) {
	return tb.batch.Delete(append([]byte(tb.prefix), key...))
}

func (tb *tableBatch) Write() error {
	return tb.batch.Write()
}

func (tb *tableBatch) ValueSize() int {
	return tb.batch.ValueSize()
}

func (tb *tableBatch) Reset() {
	tb.batch.Reset()
}


