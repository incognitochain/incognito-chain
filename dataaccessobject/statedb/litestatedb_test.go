package statedb

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"math/rand"
	"os"
	"sort"
	"testing"
)

func genRandomKV() (common.Hash, []byte) {
	r := [32]byte{}
	for i := 0; i < 32; i++ {
		r[i] = byte(rand.Intn(256))
	}
	h := common.HashH(r[:])
	return h, r[:]
}
func TestLiteStateDB(t *testing.T) {
	//init DB and txDB
	os.RemoveAll("./tmp")
	db, err := incdb.Open("leveldb", "./tmp")
	if err != nil {
		t.Error(err)
	}
	txDB, err := NewLiteStateDB("./tmp/state", common.Hash{}, common.Hash{}, db)
	if err != nil {
		t.Error(err)
	}

	//generate data
	var randKey []common.Hash
	var randValue [][]byte
	rand.Seed(1)
	for i := 0; i < 100; i++ {
		k, v := genRandomKV()
		randKey = append(randKey, k)
		randValue = append(randValue, v)
	}
	//flow test
	emptyIter := txDB.liteStateDB.NewIteratorwithPrefix([]byte{})
	if emptyIter.Next() {
		t.Error("Empty iterator must return false immediately")
		t.FailNow()
	}
	/*
		Test Set&Get&Copy LiteStateDB
	*/

	// check set & get object
	txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[0], randValue[0])
	getData, _ := txDB.getTestObject(randKey[0])
	if !bytes.Equal(getData, randValue[0]) { // must return equal
		t.Error(errors.New("Cannot store live object to newTxDB"))
	}
	txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[1], randValue[1])
	getData, _ = txDB.getTestObject(randKey[1])
	if !bytes.Equal(getData, randValue[1]) { // must return equal
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

	for i := 0; i < 10; i++ {
		newTxDB.getOrNewStateObjectWithValue(TestObjectType, randKey[i+30], randValue[i+30])
		_, err := newTxDB.Commit(true)
		if err != nil {
			t.Error(err)
		}
	}
	newAgg, _ := newTxDB.Commit(true)

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
	aggHash1, err := txDB.Commit(true)
	if err != nil {
		t.Error(err)
	}
	if !aggHash.IsEqual(&aggHash1) { // must equal
		t.Error(errors.New("Commit without set new key must giving same aggHash"))
	}
	fmt.Println("aggHash", aggHash)
	//this will panic, to prevent flow error
	//txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[2], randValue[2])

	stateObj, _ := txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[2], randValue[2])
	stateObj.MarkDelete()
	aggHash3, err := txDB.Commit(true)
	if err != nil {
		t.Error(err)
	}
	if aggHash.IsEqual(&aggHash3) { // must return different
		t.Error(errors.New("Different key set must return different aggregated hash"))
	}

	for i := 0; i < 10; i++ {
		txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[i+3], randValue[i+3])
		_, err := txDB.Commit(true)
		if err != nil {
			t.Error(err)
		}
	}

	txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[13], randValue[13])
	aggHash4, err := txDB.Commit(true)
	if err != nil {
		t.Error(err)
	}
	stateObj, _ = txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[14], randValue[14])
	stateObj.MarkDelete()
	aggHash5, err := txDB.Commit(true)
	if err != nil {
		t.Error(err)
	}

	txDB.Finalized(db, aggHash4)

	//check finalized database
	iterator := txDB.liteStateDB.db.NewIteratorWithPrefix([]byte{})
	iteratorKeyArray := []string{}
	dataSizeArray := []string{}
	for iterator.Next() {
		k := iterator.Key()
		h, e := common.Hash{}.NewHash(k[len(PREFIX_LITESTATEDB):])
		if e != nil {
			panic(e)
		}
		iteratorKeyArray = append(iteratorKeyArray, h.String())
	}

	for _, key := range randKey[:14] {
		dataSizeArray = append(dataSizeArray, key.String())
	}
	sort.Strings(iteratorKeyArray)
	sort.Strings(dataSizeArray)

	fmt.Println(iteratorKeyArray)
	fmt.Println(dataSizeArray)

	if !common.CompareStringArray(iteratorKeyArray, dataSizeArray) {
		t.Error("Finalized database error!")
	}

	//iterator on lite statedb
	iterator = txDB.liteStateDB.NewIteratorwithPrefix([]byte{})
	iteratorKeyArray = []string{}
	dataSizeArray = []string{}
	for iterator.Next() {
		k := iterator.Key()
		h, e := common.Hash{}.NewHash(k[:])
		if e != nil {
			panic(e)
		}
		iteratorKeyArray = append(iteratorKeyArray, h.String())
	}

	for _, key := range randKey[:15] {
		dataSizeArray = append(dataSizeArray, key.String())
	}
	sort.Strings(iteratorKeyArray)
	sort.Strings(dataSizeArray)
	fmt.Println(iteratorKeyArray)
	fmt.Println(dataSizeArray)
	if !common.CompareStringArray(iteratorKeyArray, dataSizeArray) {
		t.Error("Iterator on litestatedb error")
	}

	//compare restore liteStateDB node link list with current txDB
	restoreTxDB, err := NewLiteStateDB("./tmp/state", aggHash5, aggHash4, db)
	if err != nil {
		t.Error(err)
	}
	compareStateNodeList(restoreTxDB.liteStateDB.headStateNode.previousLink, txDB.liteStateDB.headStateNode.previousLink, t)
	//check mark delete object
	delObj, err := restoreTxDB.getDeletedStateObject(TestObjectType, randKey[2])
	if !delObj.IsDeleted() {
		fmt.Println(delObj.IsDeleted(), delObj.GetValueBytes())
		t.Error(err)
	}
	delObj, err = restoreTxDB.getDeletedStateObject(TestObjectType, randKey[14])
	if !delObj.IsDeleted() {
		t.Error(err)
	}
	normalObj, err := restoreTxDB.getStateObject(TestObjectType, randKey[13])
	if normalObj.IsDeleted() {
		t.Error(err)
	}
	////compare restore liteStateDB node link list with current newTxDB
	restoreTxDB, err = NewLiteStateDB("./tmp/state", newAgg, emptyRoot, db)
	if err != nil {
		t.Error(err)
	}
	compareStateNodeList(restoreTxDB.liteStateDB.headStateNode.previousLink, newTxDB.liteStateDB.headStateNode.previousLink, t)
}

func compareStateNodeList(restoreStateNode *StateNode, originStateNode *StateNode, t *testing.T) {
	for {
		if restoreStateNode.aggregateHash.String() != originStateNode.aggregateHash.String() {
			t.Error(errors.New("Restore txDB is not correct! Different aggregateHash"))
			break
		}
		prevLink := restoreStateNode.previousLink
		if prevLink != nil {
			restoreStateNode = prevLink
			originStateNode = originStateNode.previousLink
			if originStateNode == nil {
				fmt.Println("1", prevLink.aggregateHash.String())
				t.Error(errors.New("Restore txDB is not correct! Different state number, origin nil first"))
				break
			}
		} else {
			if originStateNode.previousLink != nil {
				t.Error(errors.New("Restore txDB is not correct! Different state number, restore nil first"))
				break
			} else {
				break
			}
		}
	}
}
