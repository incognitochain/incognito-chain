package database

import (
	"github.com/internet-cash/prototype/blockchain"
	"github.com/syndtr/goleveldb/leveldb"
	"encoding/json"
	"path/filepath"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"github.com/btcsuite/btcd/database"
	"fmt"
)

func init() {
	dbCreator := func(name string, dir string) (DB, error) {
		return NewGoLevelDB(name, dir)
	}
	registerDBCreator(LevelDBBackend, dbCreator, false)
	registerDBCreator(GoLevelDBBackend, dbCreator, false)
}

var _ DB = (*LevelDB)(nil)

type LevelDB struct {
	db *leveldb.DB
}

func NewGoLevelDB(name string, dir string) (*LevelDB, error) {
	dbPath := filepath.Join(dir, name+".db")
	db, err := leveldb.OpenFile(dbPath, nil)
	if err != nil {
		return nil, err
	}
	database := &LevelDB{
		db: db,
	}
	return database, nil
}

// Implements DB.
func (db *LevelDB) Get(key []byte) []byte {
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


func (db *LevelDB) Set(key []byte, value []byte) {
	key = nonNilBytes(key)
	value = nonNilBytes(value)
	err := db.db.Put(key, value, nil)
	if err != nil {
		//cmn.PanicCrisis(err)
	}
}

func nonNilBytes(bz []byte) []byte {
	if bz == nil {
		return []byte{}
	}
	return bz
}

func (db LevelDB) PutChain([]*blockchain.Block) {

}

func (db LevelDB) GetChain() []*blockchain.Block {
	var key = []byte("")
	data := db.Get(key)
	var chain []*blockchain.Block

	if len(data) > 1 {
		err := json.Unmarshal(data,&chain)
		if err != nil {
			fmt.Print("got error")
		}
	}

	return chain
}