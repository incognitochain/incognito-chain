package statedb

import (
	"bytes"
	"errors"
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
		t.Fatal(err)
	}
	txDB, err := NewWithMode("txDB", common.STATEDB_LITE_MODE, db, *NewEmptyRebuildInfo(common.STATEDB_LITE_MODE), nil)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal("Empty iterator must return false immediately")
	}
	/*
		Test Set&Get&Copy LiteStateDB
	*/

	// check set & get object
	txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[0], randValue[0])
	getData, _ := txDB.getTestObject(randKey[0])
	if !bytes.Equal(getData, randValue[0]) { // must return equal
		t.Fatal(errors.New("Cannot store live object to newTxDB"))
	}
	txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[1], randValue[1])
	getData, _ = txDB.getTestObject(randKey[1])
	if !bytes.Equal(getData, randValue[1]) { // must return equal
		t.Fatal(errors.New("Cannot store live object to newTxDB"))
	}

	//clone new txDB: must remove old live state
	newTxDB := txDB.Copy()
	getData, _ = newTxDB.getTestObject(randKey[0])
	if len(getData) != 0 { // must return empty
		t.Fatal(errors.New("Copy stateDB but data of other live state still exist"))
	}

	newTxDB.getOrNewStateObjectWithValue(TestObjectType, randKey[1], randValue[1])
	getData, _ = newTxDB.getTestObject(randKey[1])
	if !bytes.Equal(getData, randValue[1]) { // must return equal
		t.Fatal(errors.New("Cannot store live object to newTxDB"))
	}

	for i := 0; i < 10; i++ {
		newTxDB.getOrNewStateObjectWithValue(TestObjectType, randKey[i+30], randValue[i+30])
		_, _, err := newTxDB.Commit(true)
		if err != nil {
			t.Fatal(err)
		}
	}
	_, newAggRoot, _ := newTxDB.Commit(true)

	//clone new txDB: other txDB pointer should still work
	getData, _ = txDB.getTestObject(randKey[0])
	if !bytes.Equal(getData, randValue[0]) { // must return equal
		t.Fatal(errors.New("Cannot store live object to newTxDB"))
	}

	/*
		Test Commit LiteStateDB
	*/
	aggHash, _, err := txDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	aggHash1, _, err := txDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println("aggHash", aggHash, aggHash1)
	if !aggHash.IsEqual(&aggHash1) { // must equal
		t.Fatal(errors.New("Commit without set new key must giving same aggHash"))
	}

	//this will panic, to prevent flow error
	//txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[2], randValue[2])

	stateObj, _ := txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[2], randValue[2])
	stateObj.MarkDelete()
	aggHash3, _, err := txDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	if aggHash.IsEqual(&aggHash3) { // must return different
		t.Fatal(errors.New("Different key set must return different aggregated hash"))
	}

	for i := 0; i < 10; i++ {
		txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[i+3], randValue[i+3])
		_, _, err := txDB.Commit(true)
		if err != nil {
			t.Fatal(err)
		}
	}

	txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[13], randValue[13])
	_, aggHash4Rebuild, err := txDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println("aggHash4.String()", aggHash4.String())
	txDB.Finalized(false, *aggHash4Rebuild)

	txDB1 := txDB.Copy()

	stateObj, _ = txDB1.getOrNewStateObjectWithValue(TestObjectType, randKey[14], randValue[14])
	_, _, err = txDB1.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	stateObj, _ = txDB1.getOrNewStateObjectWithValue(TestObjectType, randKey[15], randValue[15])
	stateObj.MarkDelete()
	_, agghash5aRoot, err := txDB1.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	stateObj, _ = txDB.getOrNewStateObjectWithValue(TestObjectType, randKey[14], randValue[14])
	stateObj.MarkDelete()
	_, aggHash5Root, err := txDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}

	//check finalized database
	//fmt.Println("=====================>", txDB.liteStateDB.dbPrefix)
	iterator := txDB.liteStateDB.db.NewIteratorWithPrefix([]byte(txDB.liteStateDB.dbPrefix))
	iteratorKeyArray := []string{}
	dataSizeArray := []string{}
	for iterator.Next() {
		k := iterator.Key()

		h, e := common.Hash{}.NewHash(k[len(txDB.liteStateDB.dbPrefix):])
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

	//fmt.Println(iteratorKeyArray)
	//fmt.Println(dataSizeArray)

	if !common.CompareStringArray(iteratorKeyArray, dataSizeArray) {
		t.Fatal("Finalized database error!")
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
	//fmt.Println(iteratorKeyArray)
	//fmt.Println(dataSizeArray)
	if !common.CompareStringArray(iteratorKeyArray, dataSizeArray) {
		t.Fatal("Iterator on litestatedb error")
	}

	//compare restore liteStateDB node link list with current txDB
	restoreTxDB, err := NewWithMode("txDB", common.STATEDB_LITE_MODE, db, *aggHash5Root, nil)
	if err != nil {
		t.Error(err)
	}
	if compareStateNodeList(restoreTxDB.liteStateDB.headStateNode.previousLink, txDB.liteStateDB.headStateNode.previousLink, t) != nil {
		t.Fatal(err)
	}

	//check mark delete object
	delObj, err := restoreTxDB.getDeletedStateObject(TestObjectType, randKey[2])
	if !delObj.IsDeleted() {
		//fmt.Println(delObj.IsDeleted(), delObj.GetValueBytes())
		t.Fatal(err)
	}
	delObj, err = restoreTxDB.getDeletedStateObject(TestObjectType, randKey[14])
	if !delObj.IsDeleted() {
		t.Fatal(err)
	}
	normalObj, err := restoreTxDB.getStateObject(TestObjectType, randKey[13])
	if normalObj.IsDeleted() {
		t.Fatal(err)
	}
	//fmt.Println("xxxxxxxxxxxxxxxxxxxxxxxxxxxxx", agghash5a.String())
	////restore from finalized hash

	restoreTxDB, err = NewWithMode("txDB", common.STATEDB_LITE_MODE, db, *agghash5aRoot, nil)
	if err != nil {
		t.Fatal(err)
	}
	if compareStateNodeList(restoreTxDB.liteStateDB.headStateNode.previousLink, txDB1.liteStateDB.headStateNode.previousLink, t) != nil {
		t.Fatal(err)
	}

	////restore from finalized hash, should return error
	restoreTxDB, err = NewWithMode("txDB", common.STATEDB_LITE_MODE, db, *newAggRoot, nil)
	if err != nil {
		t.Fatal(err)
	}
	//fmt.Println(err)
}

func compareStateNodeList(restoreStateNode *StateNode, originStateNode *StateNode, t *testing.T) error {
	for {
		if restoreStateNode.aggregateHash.String() != originStateNode.aggregateHash.String() {
			return errors.New("Restore txDB is not correct! Different aggregateHash")
		}
		prevLink := restoreStateNode.previousLink
		//fmt.Println("check prev nil", restoreStateNode.aggregateHash.String())
		if prevLink != nil {
			restoreStateNode = prevLink
			originStateNode = originStateNode.previousLink
			if originStateNode == nil {
				//fmt.Println("1", prevLink.aggregateHash.String())
				return errors.New("Restore txDB is not correct! Different state number, origin nil first")

			}
		} else {
			if originStateNode.previousLink != nil {
				return errors.New("Restore txDB is not correct! Different state number, restore nil first")
			} else {
				break
			}
		}
	}
	return nil
}
