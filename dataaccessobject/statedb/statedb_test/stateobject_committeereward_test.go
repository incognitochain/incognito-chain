package statedb_test

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/trie"
)

var (
	warperDBrewardTest statedb.DatabaseAccessWarper
)

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_reward")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBrewardTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func storeCommitteeReward(initRoot common.Hash, warperDB statedb.DatabaseAccessWarper) (common.Hash, map[common.Hash]*statedb.CommitteeRewardState, map[string]map[common.Hash]int) {
	mState := make(map[common.Hash]*statedb.CommitteeRewardState)
	wantM := make(map[string]map[common.Hash]int)
	for index, value := range incognitoPublicKeys {
		key, _ := statedb.GenerateCommitteeRewardObjectKey(value)
		reward := generateTokenMapWithAmount()
		rewardReceiverState := statedb.NewCommitteeRewardStateWithValue(reward, incognitoPublicKeys[index])
		mState[key] = rewardReceiverState
		wantM[value] = reward
	}
	sDB, err := statedb.NewWithPrefixTrie(initRoot, warperDB)
	if err != nil {
		panic(err)
	}
	for key, value := range mState {
		sDB.SetStateObject(statedb.CommitteeRewardObjectType, key, value)
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		panic(err)
	}
	return rootHash, mState, wantM
}

func TestStateDB_GetAllCommitteeRewardState(t *testing.T) {
	rootHash, wantM, _ := storeCommitteeReward(emptyRoot, warperDBrewardTest)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBrewardTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	for k, v := range wantM {
		gotM, has, err := tempStateDB.GetCommitteeRewardState(k)
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

func TestStateDB_StoreAndGetRewardReceiver(t *testing.T) {
	var err error = nil
	key, _ := statedb.GenerateCommitteeRewardObjectKey(incognitoPublicKeys[0])
	key2, _ := statedb.GenerateCommitteeRewardObjectKey(incognitoPublicKeys[1])
	rewardReceiverState := statedb.NewCommitteeRewardStateWithValue(generateTokenMapWithAmount(), incognitoPublicKeys[0])
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBrewardTest)
	if err != nil {
		panic(err)
	}
	err = sDB.SetStateObject(statedb.CommitteeRewardObjectType, key, rewardReceiverState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.CommitteeRewardObjectType, key, "committee reward")
	if err == nil {
		t.Fatal("expect error")
	}
	err = sDB.SetStateObject(statedb.CommitteeRewardObjectType, key, []byte("committee reward"))
	if err == nil {
		t.Fatal("expect error")
	}
	err = sDB.SetStateObject(statedb.CommitteeRewardObjectType, key2, []byte("committee reward"))
	if err == nil {
		t.Fatal("expect error")
	}
	stateObjects := sDB.GetStateObjectMapForTestOnly()
	if _, ok := stateObjects[key2]; ok {
		t.Fatalf("want nothing but got %+v", key2)
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		t.Fatal(err)
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBrewardTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	got, has, err := tempStateDB.GetCommitteeRewardState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(got, rewardReceiverState) {
		t.Fatalf("want value %+v but got %+v", rewardReceiverState, got)
	}

	got2, has2, err := tempStateDB.GetCommitteeState(key2)
	if err != nil {
		t.Fatal(err)
	}
	if has2 {
		t.Fatal(has2)
	}
	if !reflect.DeepEqual(got2, statedb.NewCommitteeState()) {
		t.Fatalf("want value %+v but got %+v", statedb.NewCommitteeState(), got2)
	}
}

func TestStateDB_GetAllRewardReceiverStateMultipleRootHash(t *testing.T) {
	offset := 9
	maxHeight := int(len(incognitoPublicKeys) / offset)
	rootHashes := []common.Hash{emptyRoot}
	wantMs := []map[string]map[common.Hash]int{}
	for i := 0; i < maxHeight; i++ {
		sDB, err := statedb.NewWithPrefixTrie(rootHashes[i], warperDBrewardTest)
		if err != nil || sDB == nil {
			t.Fatal(err)
		}
		tempKeys := incognitoPublicKeys[i*9 : (i+1)*9]
		tempM := make(map[string]map[common.Hash]int)
		prevWantM := make(map[string]map[common.Hash]int)
		if i != 0 {
			prevWantM = wantMs[i-1]
		}
		for k, v := range prevWantM {
			tempM[k] = v
		}
		for _, value := range tempKeys {
			key, _ := statedb.GenerateCommitteeRewardObjectKey(value)
			reward := generateTokenMapWithAmount()
			rewardReceiverState := statedb.NewCommitteeRewardStateWithValue(reward, value)
			err := sDB.SetStateObject(statedb.CommitteeRewardObjectType, key, rewardReceiverState)
			if err != nil {
				t.Fatal(err)
			}
			tempM[value] = reward
		}
		rootHash, err := sDB.Commit(true)
		if err != nil {
			t.Fatal(err)
		}
		err = sDB.Database().TrieDB().Commit(rootHash, false)
		if err != nil {
			t.Fatal(err)
		}
		wantMs = append(wantMs, tempM)
		rootHashes = append(rootHashes, rootHash)
	}
	for index, rootHash := range rootHashes[1:] {
		wantM := wantMs[index]
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBrewardTest)
		if err != nil || tempStateDB == nil {
			t.Fatal(err)
		}
		gotM := tempStateDB.GetAllCommitteeReward()
		for k, v1 := range gotM {
			if v2, ok := wantM[k]; !ok {
				t.Fatalf("want %+v but get nothing", k)
			} else {
				if !reflect.DeepEqual(v2, v1) {
					t.Fatalf("want %+v but got %+v", v2, v1)
				}
			}
		}
	}
}
