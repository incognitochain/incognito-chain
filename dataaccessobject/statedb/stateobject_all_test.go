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
	"strings"
	"testing"
)

var (
	warperDBAllTest statedb.DatabaseAccessWarper
)

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_committee")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBAllTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestStateDB_GetAllStateObject(t *testing.T) {
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBAllTest)
	if err != nil {
		panic(err)
	}
	wantMCommittee := make(map[int][]incognitokey.CommitteePublicKey)
	wantMRewardReceiver := make(map[string]string)
	// Committee
	from, to := 0, 64
	for _, shardID := range ids {
		mCommittee := make(map[common.Hash]*statedb.CommitteeState)
		tempCommitteePublicKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
		if err != nil {
			panic(err)
		}
		tempCommitteePublicKey = tempCommitteePublicKey[from:to]
		for _, value := range tempCommitteePublicKey {
			key, _ := statedb.GenerateCommitteeObjectKey(shardID, value)
			committeeState := statedb.NewCommitteeStateWithValue(shardID, value)
			mCommittee[key] = committeeState
			wantMCommittee[shardID] = append(wantMCommittee[shardID], value)
		}
		for key, value := range mCommittee {
			sDB.SetStateObject(statedb.CommitteeObjectType, key, value)
		}
		from += 64
		to += 64
	}

	// Reward Receiver
	mRewardReceiverStates := make(map[common.Hash]*statedb.RewardReceiverState)
	for index, value := range incognitoPublicKey {
		key, _ := statedb.GenerateRewardReceiverObjectKey(value)
		rewardReceiverState := statedb.NewRewardReceiverStateWithValue(value, receiverPaymentAddress[index])
		mRewardReceiverStates[key] = rewardReceiverState
		wantMRewardReceiver[value] = receiverPaymentAddress[index]
	}

	for key, value := range mRewardReceiverStates {
		sDB.SetStateObject(statedb.RewardReceiverObjectType, key, value)
	}

	// COMMIT
	rootHash, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		t.Fatal(err)
	}

	// GOT to verify
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBAllTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	gotMCommittee := tempStateDB.GetAllCommitteeState(ids)
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
	gotMRewardReceiver := tempStateDB.GetAllRewardReceiverState()
	for k, v1 := range gotMRewardReceiver {
		if v2, ok := wantMRewardReceiver[k]; !ok {
			t.Fatalf("want %+v but get nothing", k)
		} else {
			if strings.Compare(v2, v1) != 0 {
				t.Fatalf("want %+v but got %+v", v2, v1)
			}
		}
	}
}
