package statedb_test

import (
	"bytes"
	"github.com/incognitochain/incognito-chain/trie"
	"io/ioutil"
	"os"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb/lvdb"
)

var (
	warperDBSerialNumberTest statedb.DatabaseAccessWarper

	serialNumber1   = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	serialNumber2   = []byte{10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	serialNumber3   = []byte{11, 21, 31, 41, 51, 61, 71, 81, 91, 101}
	serialNumber1It = []byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29}
	serialNumber2It = []byte{30, 31, 32, 33, 34, 35, 36, 37, 38, 39}
	serialNumber3It = []byte{40, 41, 42, 43, 44, 45, 46, 47, 48, 49}

	serialNumber1Hash = common.HashH(serialNumber1)
	serialNumber2Hash = common.HashH(serialNumber2)
	serialNumber3Hash = common.HashH(serialNumber3)
)
var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBSerialNumberTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestStoreAndGetSerialNumberObject(t *testing.T) {
	sDB, err := statedb.New(emptyRoot, warperDBSerialNumberTest)
	if err != nil {
		t.Fatal(err)
	}
	if sDB == nil {
		t.Fatal("statedb is nil")
	}

	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber1Hash, serialNumber1)
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber2Hash, serialNumber2)
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber3Hash, serialNumber3)

	rootHash1, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(rootHash1.Bytes(), emptyRoot.Bytes()) == 0 {
		t.Fatal("root hash is empty")
	}
	err = warperDBSerialNumberTest.TrieDB().Commit(rootHash1, false)
	if err != nil {
		t.Fatal(err)
	}

	tempStateDB, err := statedb.New(rootHash1, warperDBSerialNumberTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	sn1, _ := tempStateDB.GetSerialNumber(serialNumber1Hash)
	if bytes.Compare(sn1, serialNumber1) != 0 {
		t.Fatalf("Serial number 1 expect %+v but get %+v", serialNumber1, sn1)
	}
	sn2, _ := tempStateDB.GetSerialNumber(serialNumber2Hash)
	if bytes.Compare(sn2, serialNumber2) != 0 {
		t.Fatalf("Serial number 2 expect %+v but get %+v", serialNumber2, sn2)
	}
	sn3, _ := tempStateDB.GetSerialNumber(serialNumber3Hash)
	if bytes.Compare(sn3, serialNumber3) != 0 {
		t.Fatalf("Serial number 3 expect %+v but get %+v", serialNumber3, sn3)
	}
}

func TestStoreAndGetSerialNumberObjectSameKeyDifferentValue(t *testing.T) {
	sDB, err := statedb.New(emptyRoot, warperDBSerialNumberTest)
	if err != nil {
		t.Fatal(err)
	}
	if sDB == nil {
		t.Fatal("statedb is nil")
	}
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber1Hash, serialNumber1)
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber2Hash, serialNumber2)
	rootHash1, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(rootHash1.Bytes(), emptyRoot.Bytes()) == 0 {
		t.Fatal("root hash is empty")
	}
	err = warperDBSerialNumberTest.TrieDB().Commit(rootHash1, false)
	if err != nil {
		t.Fatal(err)
	}

	if err := sDB.Reset(emptyRoot); err != nil {
		t.Fatal(err)
	}
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber1Hash, serialNumber1)
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber2Hash, serialNumber3)
	rootHash2, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(rootHash2.Bytes(), emptyRoot.Bytes()) == 0 {
		t.Fatal("root hash is empty")
	}
	err = warperDBSerialNumberTest.TrieDB().Commit(rootHash2, false)
	if err != nil {
		t.Fatal(err)
	}

	tempStateDB, err := statedb.New(rootHash1, warperDBSerialNumberTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	sn1, _ := tempStateDB.GetSerialNumber(serialNumber1Hash)
	if bytes.Compare(sn1, serialNumber1) != 0 {
		t.Fatalf("Serial number 1 expect %+v but get %+v", serialNumber1, sn1)
	}
	sn2, _ := tempStateDB.GetSerialNumber(serialNumber2Hash)
	if bytes.Compare(sn2, serialNumber2) != 0 {
		t.Fatalf("Serial number 2 expect %+v but get %+v", serialNumber2, sn2)
	}
	sn3, _ := tempStateDB.GetSerialNumber(serialNumber3Hash)
	if bytes.Compare(sn3, []byte{}) != 0 {
		t.Fatalf("Serial number 3 expect %+v but get %+v", serialNumber3, sn3)
	}
	tempStateDB2, err := statedb.New(rootHash2, warperDBSerialNumberTest)
	if err != nil || tempStateDB2 == nil {
		t.Fatal(err, tempStateDB2)
	}
	sn1, _ = tempStateDB2.GetSerialNumber(serialNumber1Hash)
	if bytes.Compare(sn1, serialNumber1) != 0 {
		t.Fatalf("Serial number 1 expect %+v but get %+v", serialNumber1, sn1)
	}
	sn2, _ = tempStateDB2.GetSerialNumber(serialNumber2Hash)
	if bytes.Compare(sn2, serialNumber3) != 0 {
		t.Fatalf("Serial number 2 expect %+v but get %+v", serialNumber2, sn2)
	}
	sn3, _ = tempStateDB2.GetSerialNumber(serialNumber3Hash)
	if bytes.Compare(sn3, []byte{}) != 0 {
		t.Fatalf("Serial number 3 expect %+v but get %+v", serialNumber3, sn3)
	}
}
func TestStoreAndGetDifferentSerialNumberObject(t *testing.T) {
	sDB, err := statedb.New(emptyRoot, warperDBSerialNumberTest)
	if err != nil {
		t.Fatal(err)
	}
	if sDB == nil {
		t.Fatal("statedb is nil")
	}

	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber1Hash, serialNumber1)
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber2Hash, serialNumber2)

	rootHash1, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(rootHash1.Bytes(), emptyRoot.Bytes()) == 0 {
		t.Fatal("root hash is empty")
	}
	err = warperDBSerialNumberTest.TrieDB().Commit(rootHash1, false)
	if err != nil {
		t.Fatal(err)
	}

	if err := sDB.Reset(emptyRoot); err != nil {
		t.Fatal(err)
	}
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber1Hash, serialNumber1)
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber3Hash, serialNumber3)
	rootHash2, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(rootHash2.Bytes(), emptyRoot.Bytes()) == 0 {
		t.Fatal("root hash is empty")
	}
	err = warperDBSerialNumberTest.TrieDB().Commit(rootHash2, false)
	if err != nil {
		t.Fatal(err)
	}

	tempStateDB1, err := statedb.New(rootHash1, warperDBSerialNumberTest)
	if err != nil || tempStateDB1 == nil {
		t.Fatal(err, tempStateDB1)
	}
	sn1, _ := tempStateDB1.GetSerialNumber(serialNumber1Hash)
	if bytes.Compare(sn1, serialNumber1) != 0 {
		t.Fatalf("Serial number 1 expect %+v but get %+v", serialNumber1, sn1)
	}
	sn2, _ := tempStateDB1.GetSerialNumber(serialNumber2Hash)
	if bytes.Compare(sn2, serialNumber2) != 0 {
		t.Fatalf("Serial number 2 expect %+v but get %+v", serialNumber2, sn2)
	}
	sn3, _ := tempStateDB1.GetSerialNumber(serialNumber3Hash)
	if bytes.Compare(sn3, []byte{}) != 0 {
		t.Fatalf("Serial number 3 expect %+v but get %+v", serialNumber3, sn3)
	}
	tempStateDB2, err := statedb.New(rootHash2, warperDBSerialNumberTest)
	if err != nil || tempStateDB2 == nil {
		t.Fatal(err, tempStateDB2)
	}
	sn1, _ = tempStateDB2.GetSerialNumber(serialNumber1Hash)
	if bytes.Compare(sn1, serialNumber1) != 0 {
		t.Fatalf("Serial number 1 expect %+v but get %+v", serialNumber1, sn1)
	}
	sn3, _ = tempStateDB2.GetSerialNumber(serialNumber3Hash)
	if bytes.Compare(sn3, serialNumber3) != 0 {
		t.Fatalf("Serial number 3 expect %+v but get %+v", serialNumber3, sn3)
	}
	sn2, _ = tempStateDB2.GetSerialNumber(serialNumber2Hash)
	if bytes.Compare(sn2, []byte{}) != 0 {
		t.Fatalf("Serial number 2 expect %+v but get %+v", serialNumber2, sn2)
	}
}

func TestDeleteSerialNumberObject(t *testing.T) {
	sDB, err := statedb.New(emptyRoot, warperDBSerialNumberTest)
	if err != nil {
		t.Fatal(err)
	}
	if sDB == nil {
		t.Fatal("statedb is nil")
	}

	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber1Hash, serialNumber1)
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber2Hash, serialNumber2)
	sDB.SetStateObject(statedb.SerialNumberObjectType, serialNumber3Hash, serialNumber3)
	rootHash1, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(rootHash1.Bytes(), emptyRoot.Bytes()) == 0 {
		t.Fatal("root hash is empty")
	}
	err = warperDBSerialNumberTest.TrieDB().Commit(rootHash1, false)
	if err != nil {
		t.Fatal(err)
	}

	sDB.MarkDeleteStateObject(statedb.SerialNumberObjectType, serialNumber3Hash)
	rootHash2, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Compare(rootHash2.Bytes(), emptyRoot.Bytes()) == 0 {
		t.Fatal("root hash is empty")
	}
	err = warperDBSerialNumberTest.TrieDB().Commit(rootHash2, false)
	if err != nil {
		t.Fatal(err)
	}

	tempStateDB1, err := statedb.New(rootHash1, warperDBSerialNumberTest)
	if err != nil || tempStateDB1 == nil {
		t.Fatal(err, tempStateDB1)
	}
	sn1, _ := tempStateDB1.GetSerialNumber(serialNumber1Hash)
	if bytes.Compare(sn1, serialNumber1) != 0 {
		t.Fatalf("Serial number 1 expect %+v but get %+v", serialNumber1, sn1)
	}
	sn2, _ := tempStateDB1.GetSerialNumber(serialNumber2Hash)
	if bytes.Compare(sn2, serialNumber2) != 0 {
		t.Fatalf("Serial number 2 expect %+v but get %+v", serialNumber2, sn2)
	}
	sn3, _ := tempStateDB1.GetSerialNumber(serialNumber3Hash)
	if bytes.Compare(sn3, serialNumber3) != 0 {
		t.Fatalf("Serial number 3 expect %+v but get %+v", serialNumber3, sn3)
	}
	tempStateDB2, err := statedb.New(rootHash2, warperDBSerialNumberTest)
	if err != nil || tempStateDB2 == nil {
		t.Fatal(err, tempStateDB2)
	}
	sn1, _ = tempStateDB2.GetSerialNumber(serialNumber1Hash)
	if bytes.Compare(sn1, serialNumber1) != 0 {
		t.Fatalf("Serial number 1 expect %+v but get %+v", serialNumber1, sn1)
	}
	sn2, _ = tempStateDB2.GetSerialNumber(serialNumber2Hash)
	if bytes.Compare(sn2, serialNumber2) != 0 {
		t.Fatalf("Serial number 2 expect %+v but get %+v", serialNumber2, sn2)
	}
	sn3, _ = tempStateDB2.GetSerialNumber(serialNumber3Hash)
	if bytes.Compare(sn3, []byte{}) != 0 {
		t.Fatalf("Serial number 3 expect %+v but get %+v", serialNumber3, sn3)
	}
}
