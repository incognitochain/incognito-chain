package statedb_test

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

func storeCommitment(initRoot common.Hash, db statedb.DatabaseAccessWarper, limit int, shardID byte) (common.Hash, map[common.Hash]*statedb.CommitmentState, map[common.Hash][][]byte) {
	commitmentIndex := new(big.Int).SetUint64(0)
	commitmentPerToken := 5
	commitmentList := generateCommitmentList(commitmentPerToken * limit)
	tokenIDs := generateTokenIDs(limit)
	wantM := make(map[common.Hash]*statedb.CommitmentState)
	wantIndexM := make(map[common.Hash]common.Hash)
	wantLengthM := make(map[common.Hash]uint64)
	wantMByToken := make(map[common.Hash][][]byte)
	wantIndexMByToken := make(map[common.Hash][]uint64)
	wantLengthMByToken := make(map[common.Hash]uint64)
	for i, tokenID := range tokenIDs {
		for j := i; j < i+commitmentPerToken; j++ {
			commitment := commitmentList[j]
			key := statedb.GenerateCommitmentObjectKey(tokenID, shardID, commitment)
			commitmentState := statedb.NewCommitmentStateWithValue(tokenID, shardID, commitment, commitmentIndex)
			wantM[key] = commitmentState
			wantMByToken[tokenID] = append(wantMByToken[tokenID], commitment)

			keyIndex := statedb.GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentIndex)
			commitmentIndexState := key
			wantIndexM[keyIndex] = commitmentIndexState
			wantIndexMByToken[tokenID] = append(wantIndexMByToken[tokenID], commitmentIndex.Uint64())

			keyLength := statedb.GenerateCommitmentLengthObjectKey(tokenID, shardID)
			commitmentLengthState := commitmentIndex.Uint64()
			wantLengthM[keyLength] = commitmentLengthState
			wantLengthMByToken[tokenID] = commitmentIndex.Uint64()

			temp := commitmentIndex.Uint64()
			commitmentIndex.SetUint64(temp + 1)
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

func TestStateDB_StoreAndGetCommitmentState(t *testing.T) {
	tokenID := generateTokenIDs(1)[0]
	shardID := byte(0)
	commitmentIndex := new(big.Int).SetUint64(1)
	commitment := generateCommitmentList(1)[0]

	key := statedb.GenerateCommitmentObjectKey(tokenID, shardID, commitment)
	commitmentState := statedb.NewCommitmentStateWithValue(tokenID, shardID, commitment, commitmentIndex)

	keyIndex := statedb.GenerateCommitmentIndexObjectKey(tokenID, shardID, commitmentIndex)
	commitmentIndexState := key

	keyLength := statedb.GenerateCommitmentLengthObjectKey(tokenID, shardID)
	commitmentLengthState := new(big.Int).SetUint64(commitmentIndex.Uint64())

	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBTxTest)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.CommitmentObjectType, key, commitmentState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.CommitmentIndexObjectType, keyIndex, commitmentIndexState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.CommitmentLengthObjectType, keyLength, commitmentLengthState)
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
	gotC, has, err := tempStateDB.GetCommitmentState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotC, commitmentState) {
		t.Fatalf("GetCommitmentState want %+v but got %+v", commitmentState, gotC)
	}

	gotC2, has, err := tempStateDB.GetCommitmentIndexState(keyIndex)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(gotC2, commitmentState) {
		t.Fatalf("GetCommitmentState want %+v but got %+v", commitmentState, gotC2)
	}

	gotCLength, has, err := tempStateDB.GetCommitmentLengthState(keyLength)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if gotCLength.Uint64() != commitmentLengthState.Uint64() {
		t.Fatalf("GetCommitmentState want %+v but got %+v", commitmentLengthState.Uint64(), gotCLength.Uint64())
	}

}

//
//func TestStateDB_GetAllSerialNumberByPrefix(t *testing.T) {
//	wantMs := []map[common.Hash]*statedb.CommitmentState{}
//	wantMByTokens := []map[common.Hash][][]byte{}
//	rootHashes := []common.Hash{emptyRoot}
//	for index, shardID := range shardIDs {
//		tempRootHash, wantM, wantMByToken := storeSerialNumber(rootHashes[index], warperDBTxTest, 50, shardID)
//		rootHashes = append(rootHashes, tempRootHash)
//		wantMs = append(wantMs, wantM)
//		wantMByTokens = append(wantMByTokens, wantMByToken)
//	}
//	rootHash := rootHashes[len(rootHashes)-1]
//	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBTxTest)
//	if err != nil {
//		t.Fatal(err)
//	}
//	for index, shardID := range shardIDs {
//		tempWantMByToken := wantMByTokens[index]
//		for tokenID, wantSerialNumberList := range tempWantMByToken {
//			gotCerialNumberList := tempStateDB.GetAllSerialNumberByPrefix(tokenID, shardID)
//			for _, wantSerialNumber := range wantSerialNumberList {
//				flag := false
//				for _, gotCerialNumber := range gotCerialNumberList {
//					if bytes.Compare(wantSerialNumber, gotCerialNumber) == 0 {
//						flag = true
//						break
//					}
//				}
//				if !flag {
//					t.Fatalf("GetAllSerialNumberByPrefix shard %+v didn't got %+v", shardID, wantSerialNumber)
//				}
//			}
//		}
//	}
//}
