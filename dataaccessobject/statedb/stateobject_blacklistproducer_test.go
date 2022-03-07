package statedb

import (
	"reflect"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
)

func storeBlackListProducer(initRoot common.Hash, warperDB DatabaseAccessWarper, beaconHeight uint64, from, to int) (common.Hash, map[common.Hash]*BlackListProducerState, map[string]uint8) {
	mState := make(map[common.Hash]*BlackListProducerState)
	wantM := make(map[string]uint8)
	for _, value := range committeePublicKeys[from:to] {
		key := GenerateBlackListProducerObjectKey(value)
		duration := testGeneratePunishedDuration()
		blackListProducerState := NewBlackListProducerStateWithValue(value, duration, beaconHeight)
		mState[key] = blackListProducerState
		wantM[value] = duration
	}
	sDB, err := NewWithPrefixTrie(initRoot, warperDB)
	if err != nil {
		panic(err)
	}
	for key, value := range mState {
		sDB.SetStateObject(BlackListProducerObjectType, key, value)
	}
	rootHash, _, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false, nil)
	if err != nil {
		panic(err)
	}
	return rootHash, mState, wantM
}
func TestStateDB_GetAllBlackListProducerStateByKey(t *testing.T) {
	rootHash, wantMState, _ := storeBlackListProducer(emptyRoot, wrarperDB, 1, 0, 100)
	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	for k, v := range wantMState {
		gotM, has, err := tempStateDB.getBlackListProducerState(k)
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
	rootHash, _, wantM := storeBlackListProducer(emptyRoot, wrarperDB, 1, 0, 100)
	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	for k, v := range wantM {
		key := GenerateBlackListProducerObjectKey(k)
		gotM, has, err := tempStateDB.getBlackListProducerPunishedEpoch(key)
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
	rootHash, _, wantM := storeBlackListProducer(emptyRoot, wrarperDB, 1, 0, 100)
	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	gotM := tempStateDB.getAllProducerBlackList()
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
	rootHash, wantMState, wantM := storeBlackListProducer(emptyRoot, wrarperDB, 1, 0, 100)
	sDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
	if err != nil || sDB == nil {
		t.Fatal(err)
	}

	newWantMState := make(map[common.Hash]*BlackListProducerState)
	newWantM := make(map[string]uint8)
	for k, v := range wantMState {
		newDuration := v.PunishedEpoches() - 1
		newWantMState[k] = NewBlackListProducerStateWithValue(v.ProducerCommitteePublicKey(), newDuration, v.BeaconHeight()+1)
		newWantM[v.ProducerCommitteePublicKey()] = newDuration
	}

	for key, value := range newWantMState {
		sDB.SetStateObject(BlackListProducerObjectType, key, value)
	}
	rootHash1, _, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash1, false, nil)
	if err != nil {
		panic(err)
	}

	tempStateDB, err := NewWithPrefixTrie(rootHash, wrarperDB)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	gotM := tempStateDB.getAllProducerBlackList()
	for k, v := range wantM {
		if v2, ok := gotM[k]; !ok {
			t.Fatalf("want %+v but got nothing", k)
		} else {
			if v != v2 {
				t.Fatalf("want %+v but got %+v", v, gotM)
			}
		}
	}

	tempStateDB1, err := NewWithPrefixTrie(rootHash1, wrarperDB)
	if err != nil || tempStateDB1 == nil {
		t.Fatal(err)
	}
	newGotM := tempStateDB1.getAllProducerBlackList()
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
