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
	warperDBblTest statedb.DatabaseAccessWarper
)

var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_blacklist")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBblTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func storeBlackListProducer(initRoot common.Hash, warperDB statedb.DatabaseAccessWarper, beaconHeight uint64, from, to int) (common.Hash, map[common.Hash]*statedb.BlackListProducerState, map[string]uint8) {
	mState := make(map[common.Hash]*statedb.BlackListProducerState)
	wantM := make(map[string]uint8)
	for _, value := range committeePublicKeys[from:to] {
		key := statedb.GenerateBlackListProducerObjectKey(value)
		duration := generatePunishedDuration()
		blackListProducerState := statedb.NewBlackListProducerStateWithValue(value, duration, beaconHeight)
		mState[key] = blackListProducerState
		wantM[value] = duration
	}
	sDB, err := statedb.NewWithPrefixTrie(initRoot, warperDB)
	if err != nil {
		panic(err)
	}
	for key, value := range mState {
		sDB.SetStateObject(statedb.BlackListProducerObjectType, key, value)
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
func TestStateDB_GetAllBlackListProducerStateByKey(t *testing.T) {
	rootHash, wantMState, _ := storeBlackListProducer(emptyRoot, warperDBblTest, 1, 0, 100)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBblTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	for k, v := range wantMState {
		gotM, has, err := tempStateDB.GetBlackListProducerState(k)
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
func TestStateDB_GetBlackListProducerPunishedEpoch(t *testing.T) {
	rootHash, _, wantM := storeBlackListProducer(emptyRoot, warperDBblTest, 1, 0, 100)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBblTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	for k, v := range wantM {
		key := statedb.GenerateBlackListProducerObjectKey(k)
		gotM, has, err := tempStateDB.GetBlackListProducerPunishedEpoch(key)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatal(has)
		}
		if v != gotM {
			t.Fatalf("want %+v but got %+v", v, gotM)
		}
	}
}

func TestStateDB_GetAllBlackListProducerState(t *testing.T) {
	rootHash, _, wantM := storeBlackListProducer(emptyRoot, warperDBblTest, 1, 0, 100)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBblTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	gotM := tempStateDB.GetAllProducerBlackList()
	for k, v := range wantM {
		if v2, ok := gotM[k]; !ok {
			t.Fatalf("want %+v but got nothing", k)
		} else {
			if v != v2 {
				t.Fatalf("want %+v but got %+v", v, gotM)
			}
		}
	}
}

func TestStateDB_GetAllBlackListProducerStateMultipleRootHash(t *testing.T) {
	rootHash, wantMState, wantM := storeBlackListProducer(emptyRoot, warperDBblTest, 1, 0, 100)
	sDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBblTest)
	if err != nil || sDB == nil {
		t.Fatal(err)
	}

	newWantMState := make(map[common.Hash]*statedb.BlackListProducerState)
	newWantM := make(map[string]uint8)
	for k, v := range wantMState {
		newDuration := v.PunishedEpoches() - 1
		newWantMState[k] = statedb.NewBlackListProducerStateWithValue(v.ProducerCommitteePublicKey(), newDuration, v.BeaconHeight()+1)
		newWantM[v.ProducerCommitteePublicKey()] = newDuration
	}

	for key, value := range newWantMState {
		sDB.SetStateObject(statedb.BlackListProducerObjectType, key, value)
	}
	rootHash1, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash1, false)
	if err != nil {
		panic(err)
	}

	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBblTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	gotM := tempStateDB.GetAllProducerBlackList()
	for k, v := range wantM {
		if v2, ok := gotM[k]; !ok {
			t.Fatalf("want %+v but got nothing", k)
		} else {
			if v != v2 {
				t.Fatalf("want %+v but got %+v", v, gotM)
			}
		}
	}

	tempStateDB1, err := statedb.NewWithPrefixTrie(rootHash1, warperDBblTest)
	if err != nil || tempStateDB1 == nil {
		t.Fatal(err)
	}
	newGotM := tempStateDB1.GetAllProducerBlackList()
	for k, v := range newWantM {
		if v2, ok := newGotM[k]; !ok {
			t.Fatalf("want %+v but got nothing", k)
		} else {
			if v != v2 {
				t.Fatalf("want %+v but got %+v", v, gotM)
			}
		}
	}
}
