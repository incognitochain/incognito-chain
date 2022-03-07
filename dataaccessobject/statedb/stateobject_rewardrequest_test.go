package statedb

import (
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func storeRewardRequest(initRoot common.Hash, warperDB DatabaseAccessWarper, epoch uint64, shardIDs []byte) (common.Hash, map[common.Hash]*RewardRequestState) {
	mState := make(map[common.Hash]*RewardRequestState)
	tokenIDs := testGenerateTokenIDs(maxTokenID)
	for i := uint64(1); i < epoch; i++ {
		for _, shardID := range shardIDs {
			for _, tokenID := range tokenIDs {
				key := GenerateRewardRequestObjectKey(i, shardID, tokenID)
				amount := uint64(rand.Int() % 100000000000)
				rewardRequestState := NewRewardRequestStateWithValue(i, shardID, tokenID, amount)
				mState[key] = rewardRequestState
			}
		}
	}
	sDB, err := NewWithPrefixTrie(initRoot, warperDB)
	if err != nil {
		panic(err)
	}
	for key, value := range mState {
		sDB.SetStateObject(RewardRequestObjectType, key, value)
	}
	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		panic(err)
	}
	return rootHash, mState
}

func TestStateDB_GetAllCommitteeRewardStateByKey(t *testing.T) {
	rootHash, wantM := storeRewardRequest(emptyRoot, wrarperDB, defaultMaxEpoch, shardIDs)
	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	for k, v := range wantM {
		gotM, has, err := tempStateDB.getRewardRequestState(k)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatal(has)
		}
		if !reflect.DeepEqual(v, gotM) {
			t.Fatalf("want %+v but got %+v", v, gotM)
		}
	}
}

func TestStateDB_UpdateAndGetAllCommitteeRewardStateByKey(t *testing.T) {
	rootHash1, wantM := storeRewardRequest(emptyRoot, wrarperDB, defaultMaxEpoch, shardIDs)
	sDB, err := NewWithPrefixTrie(rootHash1, wrarperDB)
	if err != nil || sDB == nil {
		t.Fatal(err)
	}
	for k, v := range wantM {
		gotM, has, err := sDB.getRewardRequestState(k)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatal(has)
		}
		if !reflect.DeepEqual(v, gotM) {
			t.Fatalf("want %+v but got %+v", v, gotM)
		}
	}
	newWantM := make(map[common.Hash]*RewardRequestState)
	for k, v := range wantM {
		temp := v.Amount() / 2
		newWantM[k] = NewRewardRequestStateWithValue(v.Epoch(), v.ShardID(), v.TokenID(), temp)
	}
	for k, v := range newWantM {
		sDB.SetStateObject(RewardRequestObjectType, k, v)
	}
	rootHash2, _, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash2, false, nil)
	if err != nil {
		panic(err)
	}

	tempStateDB1, err := NewWithPrefixTrie(rootHash1, wrarperDB)
	if err != nil || tempStateDB1 == nil {
		t.Fatal(err)
	}
	for k, v := range wantM {
		gotM, has, err := tempStateDB1.getRewardRequestState(k)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatal(has)
		}
		if !reflect.DeepEqual(v, gotM) {
			t.Fatalf("want %+v but got %+v", v, gotM)
		}
	}

	tempStateDB2, err := NewWithPrefixTrie(rootHash2, wrarperDB)
	if err != nil || tempStateDB2 == nil {
		t.Fatal(err)
	}
	for k, v := range newWantM {
		gotM, has, err := tempStateDB2.getRewardRequestState(k)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatal(has)
		}
		if !reflect.DeepEqual(v, gotM) {
			t.Fatalf("want %+v but got %+v", v, gotM)
		}
	}
	for k, v := range wantM {
		if v2, ok := newWantM[k]; !ok {
			t.Fatalf("want %+v but got nothing", k)
		} else {
			if v2.Amount() != v.Amount()/2 {
				t.Fatalf("expect %+v but got %+v", v.Amount()/2, v2.Amount())
			}
		}
	}
}

func TestStateDB_AddShardRewardRequest(t *testing.T) {
	stateDB, err := NewWithPrefixTrie(common.EmptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	amount := uint64(10000)
	epoch1 := uint64(1)
	shardID0 := byte(0)
	//shardID1 := byte(1)
	err = AddShardRewardRequest(stateDB, epoch1, shardID0, common.PRVCoinID, amount)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, _, err := stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = stateDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	gotAmount0, err := GetRewardOfShardByEpoch(stateDB, epoch1, shardID0, common.PRVCoinID)
	if err != nil {
		t.Fatal(err)
	}
	if gotAmount0 != amount {
		t.Fatalf("want %+v but got %+v", amount, gotAmount0)
	}
	err = AddShardRewardRequest(stateDB, epoch1, shardID0, common.PRVCoinID, amount*3)
	if err != nil {
		t.Fatal(err)
	}
	rootHash, _, err = stateDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = stateDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	gotAmount1, err := GetRewardOfShardByEpoch(stateDB, epoch1, shardID0, common.PRVCoinID)
	if err != nil {
		t.Fatal(err)
	}
	if gotAmount1 != amount*4 {
		t.Fatalf("want %+v but got %+v", amount*4, gotAmount1)
	}
}

func TestStateDB_AddShardRewardRequest5000(t *testing.T) {
	stateDB, err := NewWithPrefixTrie(common.EmptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	maxEpoch := 5000
	wantReward := make(map[uint64]uint64)
	shardID0 := byte(0)
	for i := 0; i < maxEpoch; i++ {
		epoch := uint64(i)
		amount := rand.Uint64()
		err = AddShardRewardRequest(stateDB, epoch, shardID0, common.PRVCoinID, amount)
		if err != nil {
			t.Fatal(err)
		}
		rootHash, _, err := stateDB.Commit(true)
		if err != nil {
			t.Fatal(err)
		}
		err = stateDB.Database().TrieDB().Commit(rootHash, false, nil)
		if err != nil {
			t.Fatal(err)
		}
		wantReward[epoch] = amount
	}
	for i := 0; i < maxEpoch; i++ {
		epoch := uint64(i)
		gotAmount0, err := GetRewardOfShardByEpoch(stateDB, epoch, shardID0, common.PRVCoinID)
		if err != nil {
			t.Fatal(err)
		}
		if gotAmount0 != wantReward[epoch] {
			t.Fatalf("Epoch %+v, want reward %+v, got %+v", epoch, wantReward[epoch], gotAmount0)
		}
	}
}

func TestStateDB_GetAllTokenIDForReward(t *testing.T) {
	wantMTokenIDs := make(map[uint64][]common.Hash)
	maxEpoch := 100
	amount := uint64(1000)
	shardID := byte(0)
	stateDB, err := NewWithPrefixTrie(common.EmptyRoot, wrarperDB)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < maxEpoch; i++ {
		epoch := uint64(i)
		tokenIDs := testGenerateTokenIDs(10)
		for _, tokenID := range tokenIDs {
			err := AddShardRewardRequest(stateDB, epoch, shardID, tokenID, amount)
			if err != nil {
				t.Fatal(err)
			}
		}
		rootHash, _, err := stateDB.Commit(true)
		if err != nil {
			t.Fatal(err)
		}
		err = stateDB.Database().TrieDB().Commit(rootHash, false, nil)
		if err != nil {
			t.Fatal(err)
		}
		wantMTokenIDs[epoch] = tokenIDs
	}
	tempStateDB := stateDB.Copy()
	for i := 0; i < maxEpoch; i++ {
		epoch := uint64(i)
		gotTokenIDs := GetAllTokenIDForReward(tempStateDB, epoch)
		wantTokenIDs := wantMTokenIDs[epoch]
		for _, wantTokenID := range wantTokenIDs {
			flag := false
			for _, gotTokenID := range gotTokenIDs {
				if wantTokenID.IsEqual(&gotTokenID) {
					flag = true
					break
				}
			}
			if !flag {
				t.Fatalf("epoch %+v, want %+v tokenID, got nothing", epoch, wantTokenID)
			}
		}
		sort.Slice(wantTokenIDs, func(i, j int) bool {
			return wantTokenIDs[i].String() < wantTokenIDs[j].String()
		})

		sort.Slice(gotTokenIDs, func(i, j int) bool {
			return gotTokenIDs[i].String() < gotTokenIDs[j].String()
		})

		for index, _ := range wantTokenIDs {
			if !wantTokenIDs[index].IsEqual(&gotTokenIDs[index]) {
				t.Fatalf("want %+v, but got %+v", wantTokenIDs[index], gotTokenIDs[index])
			}
		}
	}
}
