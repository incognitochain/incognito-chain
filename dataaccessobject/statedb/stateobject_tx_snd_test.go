package statedb

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	_ "github.com/incognitochain/incognito-chain/incdb"
)

func storeSNDerivator(initRoot common.Hash, db DatabaseAccessWarper, limit int) (common.Hash, map[common.Hash]*SNDerivatorState, map[common.Hash][][]byte) {
	sndPerToken := 5
	sndList := testGenerateSNDList(sndPerToken * limit)
	tokenIDs := testGenerateTokenIDs(limit)
	wantM := make(map[common.Hash]*SNDerivatorState)
	wantMByToken := make(map[common.Hash][][]byte)
	for i, tokenID := range tokenIDs {
		for j := i; j < i+sndPerToken; j++ {
			snd := sndList[j]
			key := GenerateSNDerivatorObjectKey(tokenID, snd)
			sndState := NewSNDerivatorStateWithValue(tokenID, snd)
			wantM[key] = sndState
			wantMByToken[tokenID] = append(wantMByToken[tokenID], snd)
		}
	}
	sDB, err := NewWithPrefixTrie(initRoot, db)
	if err != nil {
		panic(err)
	}
	for k, v := range wantM {
		err := sDB.SetStateObject(SNDerivatorObjectType, k, v)
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

func TestStateDB_StoreAndGetSNDerivatorrState(t *testing.T) {
	tokenID := testGenerateTokenIDs(1)[0]
	snd := testGenerateSNDList(1)[0]
	snd2 := testGenerateSNDList(1)[0]

	key := GenerateSNDerivatorObjectKey(tokenID, snd)
	sndState := NewSNDerivatorStateWithValue(tokenID, snd)
	key2 := GenerateSNDerivatorObjectKey(tokenID, snd2)

	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(SNDerivatorObjectType, key, sndState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(SNDerivatorObjectType, key2, snd2)
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
	gotSND, has, err := tempStateDB.getSNDerivatorState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotSND, sndState) {
		t.Fatalf("getSNDerivatorState want %+v but got %+v", sndState, gotSND)
	}
	gotSND2, has, err := tempStateDB.getSNDerivatorState(key2)
	if has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotSND2, NewSNDerivatorState()) {
		t.Fatalf("getSNDerivatorState want %+v but got %+v", NewSNDerivatorState(), gotSND2)
	}
}

func TestStateDB_GetAllSNDerivatorByPrefix(t *testing.T) {
	rootHash, _, wantMByToken := storeSNDerivator(emptyRoot, wrarperDB, 50)
	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	for tokenID, wantSNDList := range wantMByToken {
		gotSerialNumberList := tempStateDB.getAllSNDerivatorStateByPrefix(tokenID)
		for _, gotSND := range gotSerialNumberList {
			flag := false
			for _, wantSND := range wantSNDList {
				if bytes.Compare(wantSND, gotSND) == 0 {
					flag = true
					break
				}
			}
			if !flag {
				t.Fatalf("getAllSerialNumberByPrefix didn't got %+v", gotSND)
			}
		}
	}
}
