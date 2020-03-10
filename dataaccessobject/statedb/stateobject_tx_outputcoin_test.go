package statedb

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func storeOutputCoin(initRoot common.Hash, db DatabaseAccessWarper, limit int, shardID byte) (common.Hash, map[common.Hash]*OutputCoinState, map[string][][]byte) {
	tokenIDs := testGenerateTokenIDs(limit)
	publicKeys := testGenerateOutputCoinList(10)
	wantM := make(map[common.Hash]*OutputCoinState)
	wantMByPublicKey := make(map[string][][]byte)
	for _, tokenID := range tokenIDs {
		for _, publicKey := range publicKeys {
			outputCoinList := testGenerateOutputCoinList(5)
			key := GenerateOutputCoinObjectKey(tokenID, shardID, publicKey)
			outputCoinState := NewOutputCoinStateWithValue(tokenID, shardID, publicKey, outputCoinList)
			wantM[key] = outputCoinState
			wantMByPublicKey[string(publicKey)] = outputCoinList
		}
	}

	sDB, err := NewWithPrefixTrie(initRoot, db)
	if err != nil {
		panic(err)
	}
	for k, v := range wantM {
		err := sDB.SetStateObject(OutputCoinObjectType, k, v)
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
	tokenID := testGenerateTokenIDs(1)[0]
	shardID := byte(0)
	publicKey := testGeneratePublicKeyList(1)[0]
	outputCoins := testGenerateOutputCoinList(5)

	key := GenerateOutputCoinObjectKey(tokenID, shardID, publicKey)
	outputCoinState := NewOutputCoinStateWithValue(tokenID, shardID, publicKey, outputCoins)

	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(OutputCoinObjectType, key, outputCoinState)
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

	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
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
	wantMs := []map[common.Hash]*OutputCoinState{}
	wantMPublicKeys := []map[string][][]byte{}
	rootHashes := []common.Hash{emptyRoot}
	for index, shardID := range shardIDs {
		tempRootHash, wantM, wantMPublicKey := storeOutputCoin(rootHashes[index], wrarperDB, 50, shardID)
		rootHashes = append(rootHashes, tempRootHash)
		wantMs = append(wantMs, wantM)
		wantMPublicKeys = append(wantMPublicKeys, wantMPublicKey)
	}
	rootHash := rootHashes[len(rootHashes)-1]
	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
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
