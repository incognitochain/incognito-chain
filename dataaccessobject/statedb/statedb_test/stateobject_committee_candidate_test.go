package statedb_test

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
	"testing"
)

func TestStateDB_GetNextEpochCandidateCommitteeState(t *testing.T) {
	rootHash, m := storeCommitteeObjectOneShard(statedb.NextEpochCandidate, emptyRoot, 0, 0, len(committeePublicKeys))
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	for key, want := range m {
		got, has, err := tempStateDB.GetCommitteeState(key)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatalf("GetCommitteeState want key %+v, value %+v but didn't get anything", key, want)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("GetCommitteeState want key %+v, value %+v but got value %+v ", key, want, got)
		}
	}
}

func TestStateDB_GetCurrentEpochCandidateCommitteeState(t *testing.T) {
	rootHash, m := storeCommitteeObjectOneShard(statedb.CurrentEpochCandidate, emptyRoot, 0, 0, len(committeePublicKeys))
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	for key, want := range m {
		got, has, err := tempStateDB.GetCommitteeState(key)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatalf("GetCommitteeState want key %+v, value %+v but didn't get anything", key, want)
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("GetCommitteeState want key %+v, value %+v but got value %+v ", key, want, got)
		}
	}
}

func TestStateDB_GetAllCurrentEpochCandidateCommitteeKey512EightShard(t *testing.T) {
	from, to := 0, 512
	wantM := []incognitokey.CommitteePublicKey{}
	tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.CurrentEpochCandidate, emptyRoot, statedb.CandidateShardID, from, to)
	for _, v := range tempM {
		wantM = append(wantM, v.CommitteePublicKey())
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(tempRootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	gotM := tempStateDB.GetAllCandidateCommitteePublicKey(statedb.CurrentEpochCandidate)

	if len(gotM) != to-from {
		t.Fatalf("GetAllCandidateCommitteePublicKey want key length %+v but got %+v", to-from, len(gotM))
	}
	for _, want := range wantM {
		flag := false
		for _, got := range gotM {
			if reflect.DeepEqual(got, want) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("GetAllCandidateCommitteePublicKey want %+v but didn't get anything", want)
		}
	}
}
func TestStateDB_GetAllNextEpochCandidateCommitteeKey(t *testing.T) {
	from, to := 0, 512
	rootHash, m := storeCommitteeObjectOneShard(statedb.NextEpochCandidate, emptyRoot, 0, from, to)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	gotM := tempStateDB.GetAllCandidateCommitteePublicKey(statedb.NextEpochCandidate)
	if len(gotM) != to-from {
		t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", to-from, len(gotM))
	}
	for _, want := range m {
		flag := false
		for _, got := range gotM {
			if reflect.DeepEqual(got, want.CommitteePublicKey()) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("GetAllCommitteeState want %+v but didn't get anything", want.CommitteePublicKey())
		}
	}
}

func TestStateDB_GetAllCurrentEpochCandidateCommitteeKey(t *testing.T) {
	from, to := 0, 512
	rootHash, m := storeCommitteeObjectOneShard(statedb.CurrentEpochCandidate, emptyRoot, 0, from, to)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	gotM := tempStateDB.GetAllCandidateCommitteePublicKey(statedb.CurrentEpochCandidate)
	if len(gotM) != to-from {
		t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", to-from, len(gotM))
	}
	for _, want := range m {
		flag := false
		for _, got := range gotM {
			if reflect.DeepEqual(got, want.CommitteePublicKey()) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("GetAllCommitteeState want %+v but didn't get anything", want.CommitteePublicKey())
		}
	}
}

func TestStateDB_GetAllNextEpochCandidateCommitteeKey512EightShard(t *testing.T) {
	from, to := 0, 512
	wantM := []incognitokey.CommitteePublicKey{}
	tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.NextEpochCandidate, emptyRoot, statedb.CandidateShardID, from, to)
	for _, v := range tempM {
		wantM = append(wantM, v.CommitteePublicKey())
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(tempRootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	gotM := tempStateDB.GetAllCandidateCommitteePublicKey(statedb.NextEpochCandidate)

	if len(gotM) != to-from {
		t.Fatalf("GetAllCandidateCommitteePublicKey want key length %+v but got %+v", to-from, len(gotM))
	}
	for _, want := range wantM {
		flag := false
		for _, got := range gotM {
			if reflect.DeepEqual(got, want) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("GetAllCandidateCommitteePublicKey want %+v but didn't get anything", want)
		}
	}
}
