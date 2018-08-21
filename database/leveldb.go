package database

import (
	"encoding/json"
	"path/filepath"
	"fmt"

	"github.com/internet-cash/prototype/blockchain"
	"github.com/internet-cash/prototype/common"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

func init() {
	dbCreator := func(name string, dir string) (DB, error) {
		return NewLevelDB(name, dir)
	}
	registerDBCreator(LevelDBBackend, dbCreator, false)
}

var _ DB = (LevelDB)(nil)

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

// Implements DB.
func (db *LevelDB) get(key []byte) []byte {
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


func (db *LevelDB) set(key []byte, value []byte) {
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

//save block
func (db LevelDB) SaveBlock(block *blockchain.Block) (bool, error){
	data, err := json.Marshal(block)
	if err != nil {
		return false, err
	}
	var key = []byte(block.Hash().String())
	db.set(key, data)
	return true, nil
}

//get block
func (db LevelDB) GetBlock(hash common.Hash) *blockchain.Block {
	var key = []byte(hash.String())
	data := db.get(key)
	var chain *blockchain.Block

	if len(data) > 1 {
		err := json.Unmarshal(data,&chain)
		if err != nil {
			fmt.Print("got error")
		}
	}

	return chain
}