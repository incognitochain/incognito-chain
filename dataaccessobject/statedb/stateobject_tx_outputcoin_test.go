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
			for _, outputCoin := range outputCoinList {
				key := GenerateOutputCoinObjectKey(tokenID, shardID, publicKey, outputCoin)
				outputCoinState := NewOutputCoinStateWithValue(tokenID, shardID, publicKey, outputCoin)
				wantM[key] = outputCoinState
			}
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

	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false, nil)
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
	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	wantM := make(map[common.Hash]*OutputCoinState)
	for _, outputCoin := range outputCoins {
		key := GenerateOutputCoinObjectKey(tokenID, shardID, publicKey, outputCoin)
		outputCoinState := NewOutputCoinStateWithValue(tokenID, shardID, publicKey, outputCoin)
		wantM[key] = outputCoinState
		err = sDB.SetStateObject(OutputCoinObjectType, key, outputCoinState)
		if err != nil {
			t.Fatal(err)
		}
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
	for key, outputCoinState := range wantM {
		gotO, has, err := tempStateDB.getOutputCoinState(key)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatal(has)
		}
		if !reflect.DeepEqual(gotO, outputCoinState) {
			t.Fatalf("getOutputCoinState want %+v but got %+v", outputCoinState, gotO)
		}
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
			gotO, has, err := tempStateDB.getOutputCoinState(key)
			if err != nil {
				t.Fatal(err)
			}
			if !has {
				t.Fatal(has)
			}
			if !reflect.DeepEqual(wantO, gotO) {
				t.Fatalf("getOutputCoinState want %+v got %+v ", wantO, gotO)
			}
		}
	}
}
