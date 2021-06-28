package statedb

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
)

func storeCommitment(initRoot common.Hash, db DatabaseAccessWarper, limit int, shardID byte) (common.Hash, map[common.Hash]*CommitmentState, map[common.Hash][][]byte, map[common.Hash][]uint64, map[common.Hash]uint64) {
	commitmentPerToken := 20
	commitmentList := testGenerateCommitmentList(commitmentPerToken * limit)
	tokenIDs := testGenerateTokenIDs(limit)
	wantM := make(map[common.Hash]*CommitmentState)
	wantIndexM := make(map[common.Hash]common.Hash)
	wantLengthM := make(map[common.Hash]uint64)
	wantMByToken := make(map[common.Hash][][]byte)
	wantIndexMByToken := make(map[common.Hash][]uint64)
	wantLengthMByToken := make(map[common.Hash]uint64)
	for i, tokenID := range tokenIDs {
		commitmentIndex := new(big.Int).SetUint64(1)
		for j := i; j < i+commitmentPerToken; j++ {
			commitment := commitmentList[j]
			key := GenerateCommitmentObjectKey(tokenID, shardID, commitment)
			commitmentState := NewCommitmentStateWithValue(tokenID, shardID, commitment, commitmentIndex)
			wantM[key] = commitmentState
			wantMByToken[tokenID] = append(wantMByToken[tokenID], commitment)

			keyIndex := GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentIndex)
			commitmentIndexState := key
			wantIndexM[keyIndex] = commitmentIndexState
			wantIndexMByToken[tokenID] = append(wantIndexMByToken[tokenID], commitmentIndex.Uint64())

			keyLength := GenerateCommitmentLengthObjectKey(tokenID, shardID)
			commitmentLengthState := commitmentIndex.Uint64()
			wantLengthM[keyLength] = commitmentLengthState
			wantLengthMByToken[tokenID] = commitmentIndex.Uint64()

			temp := commitmentIndex.Uint64()
			commitmentIndex = new(big.Int).SetUint64(temp + 1)
		}
	}

	sDB, err := NewWithPrefixTrie(initRoot, db)
	if err != nil {
		panic(err)
	}
	for k, v := range wantM {
		err := sDB.SetStateObject(CommitmentObjectType, k, v)
		if err != nil {
			panic(err)
		}
	}
	for k, v := range wantIndexM {
		err := sDB.SetStateObject(CommitmentIndexObjectType, k, v)
		if err != nil {
			panic(err)
		}
	}
	for k, v := range wantLengthM {
		err := sDB.SetStateObject(CommitmentLengthObjectType, k, new(big.Int).SetUint64(v))
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
	return rootHash, wantM, wantMByToken, wantIndexMByToken, wantLengthMByToken
}

func TestStateDB_StoreAndGetCommitmentState(t *testing.T) {
	tokenID := testGenerateTokenIDs(1)[0]
	shardID := byte(0)
	commitmentIndex := new(big.Int).SetUint64(0)
	commitment := testGenerateCommitmentList(1)[0]

	key := GenerateCommitmentObjectKey(tokenID, shardID, commitment)
	commitmentState := NewCommitmentStateWithValue(tokenID, shardID, commitment, commitmentIndex)

	keyIndex := GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentIndex)
	commitmentIndexState := key

	keyLength := GenerateCommitmentLengthObjectKey(tokenID, shardID)
	commitmentLengthState := new(big.Int).SetUint64(commitmentIndex.Uint64())

	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(CommitmentObjectType, key, commitmentState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(CommitmentIndexObjectType, keyIndex, commitmentIndexState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(CommitmentLengthObjectType, keyLength, commitmentLengthState)
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
	gotC, has, err := tempStateDB.getCommitmentState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotC, commitmentState) {
		t.Fatalf("getCommitmentState want %+v but got %+v", commitmentState, gotC)
	}

	gotC2, has, err := tempStateDB.getCommitmentIndexState(keyIndex)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotC2, commitmentState) {
		t.Fatalf("getCommitmentState want %+v but got %+v", commitmentState, gotC2)
	}

	gotCLength, has, err := tempStateDB.getCommitmentLengthState(keyLength)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if gotCLength.Uint64() != commitmentLengthState.Uint64() {
		t.Fatalf("getCommitmentState want %+v but got %+v", commitmentLengthState.Uint64(), gotCLength.Uint64())
	}
}

func TestStateDB_GetGetAllCommitmentStateByPrefix(t *testing.T) {
	wantMs := []map[common.Hash]*CommitmentState{}
	wantMByTokens := []map[common.Hash][][]byte{}
	wantIndexMByTokens := []map[common.Hash][]uint64{}
	wantLengthMByTokens := []map[common.Hash]uint64{}
	rootHashes := []common.Hash{emptyRoot}
	for index, shardID := range shardIDs {
		tempRootHash, wantM, wantMByToken, wantIndexMByToken, wantLengthMByToken := storeCommitment(rootHashes[index], wrarperDB, 50, shardID)
		rootHashes = append(rootHashes, tempRootHash)
		wantMs = append(wantMs, wantM)
		wantMByTokens = append(wantMByTokens, wantMByToken)
		wantIndexMByTokens = append(wantIndexMByTokens, wantIndexMByToken)
		wantLengthMByTokens = append(wantLengthMByTokens, wantLengthMByToken)
	}
	rootHash := rootHashes[len(rootHashes)-1]
	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	for index, shardID := range shardIDs {
		tempWantIndexMByToken := wantIndexMByTokens[index]
		tempWantMByToken := wantMByTokens[index]
		for tokenID, wantIndexList := range tempWantIndexMByToken {
			gotCIndexList := tempStateDB.getAllCommitmentStateByPrefix(tokenID, shardID)
			for gotC, gotCIndex := range gotCIndexList {
				flag := false
				for _, wantCIndex := range wantIndexList {
					if gotCIndex == wantCIndex {
						flag = true
						break
					}
				}
				if !flag {
					t.Fatalf("getAllCommitmentStateByPrefix shard %+v didn't want %+v", shardID, gotCIndex)
				}
				flag2 := false
				for _, wantCBytes := range tempWantMByToken[tokenID] {
					wantC := base58.Base58Check{}.Encode(wantCBytes, common.Base58Version)
					if gotC == wantC {
						flag2 = true
						break
					}
				}
				if !flag2 {
					t.Fatalf("getAllCommitmentStateByPrefix shard %+v didn't want %+v", shardID, gotC)
				}
			}
			keyLength := GenerateCommitmentLengthObjectKey(tokenID, shardID)
			gotCLength, has, err := tempStateDB.getCommitmentLengthState(keyLength)
			if err != nil {
				t.Fatal(err)
			}
			if !has {
				t.Fatal(has)
			}
			if gotCLength.Uint64() != wantLengthMByTokens[index][tokenID] {
				t.Fatalf("getAllSerialNumberByPrefix shard %+v want %+v but got %+v", shardID, wantLengthMByTokens[shardID][tokenID], gotCLength.Uint64())
			}
		}
	}
}

//func TestStateDB_StoreCommitments(t *testing.T) {
//	tokenID := common.PRVCoinID
//	shardID := byte(0)
//	commitments := testGenerateCommitmentList(20)
//	sDB, err := NewWithPrefixTrie(emptyRoot, wrarperDB)
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = StoreCommitments(sDB, tokenID, []byte{}, commitments, shardID)
//	if err != nil {
//		t.Fatal(err)
//	}
//	rootHash, err := sDB.Commit(true)
//	if err != nil {
//		t.Fatal(err)
//	}
//	err = sDB.Database().TrieDB().Commit(rootHash, false)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
//	if err != nil {
//		t.Fatal(err)
//	}
//	res, err := GetCommitmentLength(tempStateDB, tokenID, shardID)
//	if err != nil {
//		t.Fatal(err)
//	}
//	if res.Uint64() != 20 {
//		t.Fatalf("want 20 but got %+v", res.Uint64())
//	}
//}
