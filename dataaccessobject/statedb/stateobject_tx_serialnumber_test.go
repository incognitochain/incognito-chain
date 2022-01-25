package statedb

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	_ "github.com/incognitochain/incognito-chain/incdb"
)

func storeSerialNumber(initRoot common.Hash, db DatabaseAccessWarper, limit int, shardID byte) (common.Hash, map[common.Hash]*SerialNumberState, map[common.Hash][][]byte) {
	serialNumberPerToken := 5
	serialNumberList := testGenerateSerialNumberList(serialNumberPerToken * limit)
	tokenIDs := testGenerateTokenIDs(limit)
	wantM := make(map[common.Hash]*SerialNumberState)
	wantMByToken := make(map[common.Hash][][]byte)
	for i, tokenID := range tokenIDs {
		for j := i; j < i+serialNumberPerToken; j++ {
			serialNumber := serialNumberList[j]
			key := GenerateSerialNumberObjectKey(tokenID, shardID, serialNumber)
			serialNumberState := NewSerialNumberStateWithValue(tokenID, shardID, serialNumber)
			wantM[key] = serialNumberState
			wantMByToken[tokenID] = append(wantMByToken[tokenID], serialNumber)
		}
	}

	sDB, err := NewWithPrefixTrie(initRoot, db)
	if err != nil {
		panic(err)
	}
	for k, v := range wantM {
		err := sDB.SetStateObject(SerialNumberObjectType, k, v)
		if err != nil {
			panic(err)
		}
	}
	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		panic(err)
	}
	return rootHash, wantM, wantMByToken
}

func TestStateDB_StoreAndGetSerialNumberState(t *testing.T) {
	tokenID := testGenerateTokenIDs(1)[0]
	shardID := byte(0)
	serialNumber := testGenerateSerialNumberList(1)[0]
	serialNumber2 := testGenerateSerialNumberList(1)[0]

	key := GenerateSerialNumberObjectKey(tokenID, shardID, serialNumber)
	serialNumberState := NewSerialNumberStateWithValue(tokenID, shardID, serialNumber)
	key2 := GenerateSerialNumberObjectKey(tokenID, shardID, serialNumber2)

	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(SerialNumberObjectType, key, serialNumberState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(SerialNumberObjectType, key2, serialNumber2)
	if err == nil {
		t.Fatal("expect error")
	}
	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		t.Fatal(err)
	}

	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	gotS, has, err := tempStateDB.getSerialNumberState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotS, serialNumberState) {
		t.Fatalf("getSerialNumberState want %+v but got %+v", serialNumberState, gotS)
	}
	gotS2, has, err := tempStateDB.getSerialNumberState(key2)
	if has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotS2, NewSerialNumberState()) {
		t.Fatalf("getSerialNumberState want %+v but got %+v", NewSerialNumberState(), gotS2)
	}
}

func TestStateDB_GetAllSerialNumberByPrefix(t *testing.T) {
	wantMs := []map[common.Hash]*SerialNumberState{}
	wantMByTokens := []map[common.Hash][][]byte{}
	rootHashes := []common.Hash{emptyRoot}
	for index, shardID := range shardIDs {
		tempRootHash, wantM, wantMByToken := storeSerialNumber(rootHashes[index], wrarperDB, 50, shardID)
		rootHashes = append(rootHashes, tempRootHash)
		wantMs = append(wantMs, wantM)
		wantMByTokens = append(wantMByTokens, wantMByToken)
	}
	rootHash := rootHashes[len(rootHashes)-1]
	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	for index, shardID := range shardIDs {
		tempWantMByToken := wantMByTokens[index]
		for tokenID, wantSerialNumberList := range tempWantMByToken {
			gotSerialNumberList := tempStateDB.getAllSerialNumberByPrefix(tokenID, shardID)
			for _, wantSerialNumber := range wantSerialNumberList {
				flag := false
				for _, gotSerialNumber := range gotSerialNumberList {
					if bytes.Compare(wantSerialNumber, gotSerialNumber) == 0 {
						flag = true
						break
					}
				}
				if !flag {
					t.Fatalf("getAllSerialNumberByPrefix shard %+v didn't got %+v", shardID, wantSerialNumber)
				}
			}
		}
	}
}
