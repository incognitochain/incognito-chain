package statedb

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"os"
	"testing"
)

func genRandomKV() (common.Hash, []byte) {
	r := common.RandBytes(32)
	h := common.HashH(r)
	return h, r
}
func TestLiteStateDB(t *testing.T) {
	//init DB and txDB
	os.Remove("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Error(err)
	}
	txDB, err := NewLiteStateDB("./tmp", 0, db)
	if err != nil {
		t.Error(err)
	}

	//generate data
	var randKey []common.Hash
	var randValue [][]byte
	for i := 0; i < 100; i++ {
		k, v := genRandomKV()
		randKey = append(randKey, k)
		randValue = append(randValue, v)
	}
	//flow test

	/*
		Test Set&Get&Copy LiteStateDB
	*/

	// check set & get object
	txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[0], randValue[0])
	getData, _ := txDB.getTestObject(randKey[0])
	if !bytes.Equal(getData, randValue[0]) { // must return equal
		t.Error(errors.New("Cannot store live object to newTxDB"))
	}

	//clone new txDB: must remove old live state
	newTxDB := txDB.Copy()
	getData, _ = newTxDB.getTestObject(randKey[0])
	if len(getData) != 0 { // must return empty
		t.Error(errors.New("Copy stateDB but data of other live state still exist"))
	}

	newTxDB.getOrNewStateObjectWithValue(TestObjectType, randKey[1], randValue[1])
	getData, _ = newTxDB.getTestObject(randKey[1])
	if !bytes.Equal(getData, randValue[1]) { // must return equal
		t.Error(errors.New("Cannot store live object to newTxDB"))
	}

	//clone new txDB: other txDB pointer should still work
	getData, _ = txDB.getTestObject(randKey[0])
	if !bytes.Equal(getData, randValue[0]) { // must return equal
		t.Error(errors.New("Cannot store live object to newTxDB"))
	}

	/*
		Test Commit LiteStateDB
	*/
	aggHash, err := txDB.Commit(true)
	if err != nil {
		t.Error(err)
	}
	_, err = txDB.Commit(true)
	if err == nil {
		t.Error(errors.New("Must trigger commit twice error"))
	}
	//this will panic, to prevent flow error
	//txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[2], randValue[2])

	txDB.Finalized(emptyRoot)
	txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[2], randValue[2])
	aggHash3, err := txDB.Commit(true)
	if err != nil {
		t.Error(err)
	}
	if aggHash.IsEqual(&aggHash3) { // must return different
		t.Error(errors.New("Different key set must return different aggregated hash"))
	}
	fmt.Println(aggHash3)
}
