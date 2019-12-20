package statedb_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	_ "github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

var (
	warperDBTxTest statedb.DatabaseAccessWarper
)
var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_tx")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBTxTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func storeSerialNumber(initRoot common.Hash, db statedb.DatabaseAccessWarper, limit int, shardID byte) (common.Hash, map[common.Hash]*statedb.SerialNumberState, map[common.Hash][][]byte) {
	serialNumberPerToken := 5
	serialNumberList := generateSerialNumberList(serialNumberPerToken * limit)
	tokenIDs := generateTokenIDs(limit)
	wantM := make(map[common.Hash]*statedb.SerialNumberState)
	wantMByToken := make(map[common.Hash][][]byte)
	for i, tokenID := range tokenIDs {
		for j := i; j < i+serialNumberPerToken; j++ {
			serialNumber := serialNumberList[j]
			key := statedb.GenerateSerialNumberObjectKey(tokenID, shardID, serialNumber)
			serialNumberState := statedb.NewSerialNumberStateWithValue(tokenID, shardID, serialNumber)
			wantM[key] = serialNumberState
			wantMByToken[tokenID] = append(wantMByToken[tokenID], serialNumber)
		}
	}

	sDB, err := statedb.NewWithPrefixTrie(initRoot, db)
	if err != nil {
		panic(err)
	}
	for k, v := range wantM {
		err := sDB.SetStateObject(statedb.SerialNumberObjectType, k, v)
		if err != nil {
			panic(err)
		}
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		panic(err)
	}
	return rootHash, wantM, wantMByToken
}

func TestStateDB_StoreAndGetSerialNumberState(t *testing.T) {
	tokenID := generateTokenIDs(1)[0]
	shardID := byte(0)
	serialNumber := generateSerialNumberList(1)[0]
	serialNumber2 := generateSerialNumberList(1)[0]

	key := statedb.GenerateSerialNumberObjectKey(tokenID, shardID, serialNumber)
	serialNumberState := statedb.NewSerialNumberStateWithValue(tokenID, shardID, serialNumber)
	key2 := statedb.GenerateSerialNumberObjectKey(tokenID, shardID, serialNumber2)

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBTxTest)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.SerialNumberObjectType, key, serialNumberState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.SerialNumberObjectType, key2, serialNumber2)
	if err == nil {
		t.Fatal("expect error")
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		t.Fatal(err)
	}

	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBTxTest)
	if err != nil {
		t.Fatal(err)
	}
	gotS, has, err := tempStateDB.GetSerialNumberState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotS, serialNumberState) {
		t.Fatalf("GetSerialNumberState want %+v but got %+v", serialNumberState, gotS)
	}
	gotS2, has, err := tempStateDB.GetSerialNumberState(key2)
	if has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotS2, statedb.NewSerialNumberState()) {
		t.Fatalf("GetSerialNumberState want %+v but got %+v", statedb.NewSerialNumberState(), gotS2)
	}
}

func TestStateDB_GetAllSerialNumberByPrefix(t *testing.T) {
	wantMs := []map[common.Hash]*statedb.SerialNumberState{}
	wantMByTokens := []map[common.Hash][][]byte{}
	rootHashes := []common.Hash{emptyRoot}
	for index, shardID := range shardIDs {
		tempRootHash, wantM, wantMByToken := storeSerialNumber(rootHashes[index], warperDBTxTest, 50, shardID)
		rootHashes = append(rootHashes, tempRootHash)
		wantMs = append(wantMs, wantM)
		wantMByTokens = append(wantMByTokens, wantMByToken)
	}
	rootHash := rootHashes[len(rootHashes)-1]
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBTxTest)
	if err != nil {
		t.Fatal(err)
	}
	for index, shardID := range shardIDs {
		tempWantMByToken := wantMByTokens[index]
		for tokenID, wantSerialNumberList := range tempWantMByToken {
			gotSerialNumberList := tempStateDB.GetAllSerialNumberByPrefix(tokenID, shardID)
			for _, wantSerialNumber := range wantSerialNumberList {
				flag := false
				for _, gotSerialNumber := range gotSerialNumberList {
					if bytes.Compare(wantSerialNumber, gotSerialNumber) == 0 {
						flag = true
						break
					}
				}
				if !flag {
					t.Fatalf("GetAllSerialNumberByPrefix shard %+v didn't got %+v", shardID, wantSerialNumber)
				}
			}
		}
	}
}
