package statedb_test

import (
	"io/ioutil"
	"math/rand"
	"os"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

var (
	warperDBrrTest statedb.DatabaseAccessWarper
)

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_reward")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBrrTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func storeRewardRequest(initRoot common.Hash, warperDB statedb.DatabaseAccessWarper, epoch uint64, shardIDs []byte) (common.Hash, map[common.Hash]*statedb.RewardRequestState) {
	mState := make(map[common.Hash]*statedb.RewardRequestState)
	tokenIDs := generateTokenIDs(maxTokenID)
	for i := uint64(1); i < epoch; i++ {
		for _, shardID := range shardIDs {
			for _, tokenID := range tokenIDs {
				key := statedb.GenerateRewardRequestObjectKey(i, shardID, tokenID)
				amount := uint64(rand.Int() % 100000000000)
				rewardRequestState := statedb.NewRewardRequestStateWithValue(i, shardID, tokenID, amount)
				mState[key] = rewardRequestState
			}
		}
	}
	sDB, err := statedb.NewWithPrefixTrie(initRoot, warperDB)
	if err != nil {
		panic(err)
	}
	for key, value := range mState {
		sDB.SetStateObject(statedb.RewardRequestObjectType, key, value)
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		panic(err)
	}
	return rootHash, mState
}

func TestStateDB_GetAllCommitteeRewardStateByKey(t *testing.T) {
	rootHash, wantM := storeRewardRequest(emptyRoot, warperDBrrTest, defaultMaxEpoch, shardIDs)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBrrTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	for k, v := range wantM {
		gotM, has, err := tempStateDB.GetRewardRequestState(k)
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
	rootHash1, wantM := storeRewardRequest(emptyRoot, warperDBrrTest, defaultMaxEpoch, shardIDs)
	sDB, err := statedb.NewWithPrefixTrie(rootHash1, warperDBrrTest)
	if err != nil || sDB == nil {
		t.Fatal(err)
	}
	for k, v := range wantM {
		gotM, has, err := sDB.GetRewardRequestState(k)
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
	newWantM := make(map[common.Hash]*statedb.RewardRequestState)
	for k, v := range wantM {
		temp := v.Amount() / 2
		newWantM[k] = statedb.NewRewardRequestStateWithValue(v.Epoch(), v.ShardID(), v.TokenID(), temp)
	}
	for k, v := range newWantM {
		sDB.SetStateObject(statedb.RewardRequestObjectType, k, v)
	}
	rootHash2, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash2, false)
	if err != nil {
		panic(err)
	}

	tempStateDB1, err := statedb.NewWithPrefixTrie(rootHash1, warperDBrrTest)
	if err != nil || tempStateDB1 == nil {
		t.Fatal(err)
	}
	for k, v := range wantM {
		gotM, has, err := tempStateDB1.GetRewardRequestState(k)
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

	tempStateDB2, err := statedb.NewWithPrefixTrie(rootHash2, warperDBrrTest)
	if err != nil || tempStateDB2 == nil {
		t.Fatal(err)
	}
	for k, v := range newWantM {
		gotM, has, err := tempStateDB2.GetRewardRequestState(k)
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
