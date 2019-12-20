package statedb_test

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	_ "github.com/incognitochain/incognito-chain/incdb"
)

func storeSNDerivator(initRoot common.Hash, db statedb.DatabaseAccessWarper, limit int) (common.Hash, map[common.Hash]*statedb.SNDerivatorState, map[common.Hash][][]byte) {
	sndPerToken := 5
	sndList := generateSNDList(sndPerToken * limit)
	tokenIDs := generateTokenIDs(limit)
	wantM := make(map[common.Hash]*statedb.SNDerivatorState)
	wantMByToken := make(map[common.Hash][][]byte)
	for i, tokenID := range tokenIDs {
		for j := i; j < i+sndPerToken; j++ {
			snd := sndList[j]
			key := statedb.GenerateSNDerivatorObjectKey(tokenID, snd)
			sndState := statedb.NewSNDerivatorStateWithValue(tokenID, snd)
			wantM[key] = sndState
			wantMByToken[tokenID] = append(wantMByToken[tokenID], snd)
		}
	}
	sDB, err := statedb.NewWithPrefixTrie(initRoot, db)
	if err != nil {
		panic(err)
	}
	for k, v := range wantM {
		err := sDB.SetStateObject(statedb.SNDerivatorObjectType, k, v)
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

func TestStateDB_StoreAndGetSNDerivatorrState(t *testing.T) {
	tokenID := generateTokenIDs(1)[0]
	snd := generateSNDList(1)[0]
	snd2 := generateSNDList(1)[0]

	key := statedb.GenerateSNDerivatorObjectKey(tokenID, snd)
	sndState := statedb.NewSNDerivatorStateWithValue(tokenID, snd)
	key2 := statedb.GenerateSNDerivatorObjectKey(tokenID, snd2)

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBTxTest)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.SNDerivatorObjectType, key, sndState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.SNDerivatorObjectType, key2, snd2)
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
	gotSND, has, err := tempStateDB.GetSNDerivatorState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotSND, sndState) {
		t.Fatalf("GetSNDerivatorState want %+v but got %+v", sndState, gotSND)
	}
	gotSND2, has, err := tempStateDB.GetSNDerivatorState(key2)
	if has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotSND2, statedb.NewSNDerivatorState()) {
		t.Fatalf("GetSNDerivatorState want %+v but got %+v", statedb.NewSNDerivatorState(), gotSND2)
	}
}

func TestStateDB_GetAllSNDerivatorByPrefix(t *testing.T) {
	rootHash, _, wantMByToken := storeSNDerivator(emptyRoot, warperDBTxTest, 50)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBTxTest)
	if err != nil {
		t.Fatal(err)
	}
	for tokenID, wantSNDList := range wantMByToken {
		gotSerialNumberList := tempStateDB.GetAllSNDerivatorStateByPrefix(tokenID)
		for _, gotSND := range gotSerialNumberList {
			flag := false
			for _, wantSND := range wantSNDList {
				if bytes.Compare(wantSND, gotSND) == 0 {
					flag = true
					break
				}
			}
			if !flag {
				t.Fatalf("GetAllSerialNumberByPrefix didn't got %+v", gotSND)
			}
		}
	}
}
