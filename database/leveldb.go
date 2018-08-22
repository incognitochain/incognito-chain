package database

import (
	"path/filepath"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

func init() {
	dbCreator := func(name string, dir string) (DB, error) {
		return NewLevelDB(name, dir)
	}
	registerDBCreator(LevelDBBackend, dbCreator, false)
}

var _ DB = (*LevelDB)(nil)

type LevelDB struct {
	db *leveldb.DB
}

func NewLevelDB(name string, dir string) (*LevelDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	_data, err := leveldb.OpenFile(dbPath, nil)
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

func (db *LevelDB) View(fn func(Tx) error) error {
	return nil
}

//save block
func (db LevelDB) SaveBlock(key []byte, value []byte) (bool, error){
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	err := db.db.Put(key, value, nil)
	if err != nil {
		return false, err
	}
	return true, nil
}

//get block
func (db LevelDB) GetBlock(key []byte) ([]byte) {
	key = nonNilBytes(key)
	res, err := db.db.Get(key, nil)
	if err != nil {
		if err == errors.ErrNotFound {
			return nil
		}
		panic(err)
	}

	return res
}