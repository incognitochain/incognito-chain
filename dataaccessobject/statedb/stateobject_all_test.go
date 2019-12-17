package statedb_test

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/trie"
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

var (
	warperDBAllTest statedb.DatabaseAccessWarper
)

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_all")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBAllTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func storeAllStateObjectForTesting(initRoot common.Hash) (common.Hash, map[int][]incognitokey.CommitteePublicKey, map[string]map[common.Hash]int) {
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	sDB, err := statedb.NewWithPrefixTrie(initRoot, warperDBAllTest)
	if err != nil {
		panic(err)
	}
	wantMCommittee := make(map[int][]incognitokey.CommitteePublicKey)
	wantMReward := make(map[string]map[common.Hash]int)
	// Committee
	from, to := 0, 64
	for _, shardID := range ids {
		mCommittee := make(map[common.Hash]*statedb.CommitteeState)
		tempCommitteePublicKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
		if err != nil {
			panic(err)
		}
		tempCommitteePublicKey = tempCommitteePublicKey[from:to]
		for index, value := range tempCommitteePublicKey {
			key, _ := statedb.GenerateCommitteeObjectKeyWithRole(statedb.CurrentValidator, shardID, value)
			committeeState := statedb.NewCommitteeStateWithValue(shardID, statedb.CurrentValidator, value, receiverPaymentAddress[index], true)
			mCommittee[key] = committeeState
			wantMCommittee[shardID] = append(wantMCommittee[shardID], value)
		}
		for key, value := range mCommittee {
			sDB.SetStateObject(statedb.CommitteeObjectType, key, value)
		}
		from += 64
		to += 64
	}

	// Committee Reward
	mRewardReceiverStates := make(map[common.Hash]*statedb.CommitteeRewardState)
	for _, value := range incognitoPublicKeys {
		key, _ := statedb.GenerateCommitteeRewardObjectKey(value)
		reward := generateTokenMapWithAmount()
		rewardReceiverState := statedb.NewCommitteeRewardStateWithValue(reward, value)
		mRewardReceiverStates[key] = rewardReceiverState
		wantMReward[value] = reward
	}

	for key, value := range mRewardReceiverStates {
		sDB.SetStateObject(statedb.CommitteeRewardObjectType, key, value)
	}

	// COMMIT
	rootHash, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		panic(err)
	}
	return rootHash, wantMCommittee, wantMReward
}

func TestStateDB_GetAllStateObject(t *testing.T) {
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	rootHash, wantMCommittee, wantMReward := storeAllStateObjectForTesting(emptyRoot)
	// GOT to verify
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBAllTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	gotMCommittee := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, ids)
	for _, id := range ids {
		temp, ok := gotMCommittee[id]
		if !ok {
			t.Fatalf("GetAllCommitteeState want shard %+v", id)
		}
		if len(temp) != 64 {
			t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 64, len(temp))
		}
	}
	for id, wants := range wantMCommittee {
		flag := false
		for _, want := range wants {
			for _, got := range gotMCommittee[id] {
				if reflect.DeepEqual(got, want) {
					flag = true
					break
				}
			}
			if !flag {
				t.Fatalf("GetAllCommitteeState shard %+v want %+v but didn't get anything", id, want)
			}
		}
	}
	gotMRewardReceiver := tempStateDB.GetAllCommitteeReward()
	for k, v1 := range gotMRewardReceiver {
		if v2, ok := wantMReward[k]; !ok {
			t.Fatalf("want %+v but get nothing", k)
		} else {
			if !reflect.DeepEqual(v2, v1) {
				t.Fatalf("want %+v but got %+v", v2, v1)
			}
		}
	}
}

func BenchmarkStateDB_GetAllRewardReceiverInFullData(b *testing.B) {
	rootHash, _, _ := storeAllStateObjectForTesting(emptyRoot)
	// GOT to verify
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBAllTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetAllCommitteeReward()
	}
}

func BenchmarkStateDB_GetAllCommitteeInFullData(b *testing.B) {
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	rootHash, _, _ := storeAllStateObjectForTesting(emptyRoot)
	// GOT to verify
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBAllTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, ids)
	}
}
