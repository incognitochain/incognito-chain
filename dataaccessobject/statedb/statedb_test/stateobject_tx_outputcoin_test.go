package statedb_test

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func storeOutputCoin(initRoot common.Hash, db statedb.DatabaseAccessWarper, limit int, shardID byte) (common.Hash, map[common.Hash]*statedb.OutputCoinState, map[string][][]byte) {
	tokenIDs := generateTokenIDs(limit)
	publicKeys := generateOutputCoinList(10)
	wantM := make(map[common.Hash]*statedb.OutputCoinState)
	wantMByPublicKey := make(map[string][][]byte)
	for _, tokenID := range tokenIDs {
		for _, publicKey := range publicKeys {
			outputCoinList := generateOutputCoinList(5)
			key := statedb.GenerateOutputCoinObjectKey(tokenID, shardID, publicKey)
			outputCoinState := statedb.NewOutputCoinStateWithValue(tokenID, shardID, publicKey, outputCoinList)
			wantM[key] = outputCoinState
			wantMByPublicKey[string(publicKey)] = outputCoinList
		}
	}

	sDB, err := statedb.NewWithPrefixTrie(initRoot, db)
	if err != nil {
		panic(err)
	}
	for k, v := range wantM {
		err := sDB.SetStateObject(statedb.OutputCoinObjectType, k, v)
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
	return rootHash, wantM, wantMByPublicKey
}

func TestStateDB_StoreAndGetOutputCoinState(t *testing.T) {
	tokenID := generateTokenIDs(1)[0]
	shardID := byte(0)
	publicKey := generatePublicKeyList(1)[0]
	outputCoins := generateOutputCoinList(5)

	key := statedb.GenerateOutputCoinObjectKey(tokenID, shardID, publicKey)
	outputCoinState := statedb.NewOutputCoinStateWithValue(tokenID, shardID, publicKey, outputCoins)

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBTxTest)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.OutputCoinObjectType, key, outputCoinState)
	if err != nil {
		t.Fatal(err)
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
	gotO, has, err := tempStateDB.GetOutputCoinState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotO, outputCoinState) {
		t.Fatalf("GetOutputCoinState want %+v but got %+v", outputCoinState, gotO)
	}
}

func TestStateDB_GetMultipleOutputCoinState(t *testing.T) {
	wantMs := []map[common.Hash]*statedb.OutputCoinState{}
	wantMPublicKeys := []map[string][][]byte{}
	rootHashes := []common.Hash{emptyRoot}
	for index, shardID := range shardIDs {
		tempRootHash, wantM, wantMPublicKey := storeOutputCoin(rootHashes[index], warperDBTxTest, 50, shardID)
		rootHashes = append(rootHashes, tempRootHash)
		wantMs = append(wantMs, wantM)
		wantMPublicKeys = append(wantMPublicKeys, wantMPublicKey)
	}
	rootHash := rootHashes[len(rootHashes)-1]
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBTxTest)
	if err != nil {
		t.Fatal(err)
	}
	for index, _ := range shardIDs {
		tempWantM := wantMs[index]
		for key, wantO := range tempWantM {
			gotO, has, err := tempStateDB.GetOutputCoinState(key)
			if err != nil {
				t.Fatal(err)
			}
			if !has {
				t.Fatal(has)
			}
			if !reflect.DeepEqual(wantO, gotO) {
				t.Fatalf("GetOutputCoinState want %+v got %+v ", wantO, gotO)
			}
		}
	}
}
