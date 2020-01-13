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
	warperDBCommitteeTest statedb.DatabaseAccessWarper
)
var _ = func() (_ struct{}) {
	dbPath, err := ioutil.TempDir(os.TempDir(), "test_committee")
	if err != nil {
		panic(err)
	}
	diskBD, _ := incdb.Open("leveldb", dbPath)
	warperDBCommitteeTest = statedb.NewDatabaseAccessWarper(diskBD)
	trie.Logger.Init(common.NewBackend(nil).Logger("test", true))
	return
}()

func TestStateDB_GetMixCommitteePublicKey(t *testing.T) {
	from, to := 0, 32
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantCurrentValidatorM := make(map[int][]incognitokey.CommitteePublicKey)
	rootHashes := []common.Hash{emptyRoot}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.CurrentValidator, rootHashes[index], id, from, to)
		from += 32
		to += 32
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantCurrentValidatorM[id] = append(wantCurrentValidatorM[id], v.CommitteePublicKey())
		}
	}
	tempRootHash := rootHashes[8]
	wantSubstituteValidatorM := make(map[int][]incognitokey.CommitteePublicKey)
	to = from + 8
	rootHashes = []common.Hash{tempRootHash}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.SubstituteValidator, rootHashes[index], id, from, to)
		from += 8
		to += 8
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantSubstituteValidatorM[id] = append(wantSubstituteValidatorM[id], v.CommitteePublicKey())
		}
	}
	to = from + 80
	tempRootHash = rootHashes[8]
	wantNextEpochCandidateM := []incognitokey.CommitteePublicKey{}
	tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.NextEpochShardCandidate, tempRootHash, statedb.CandidateShardID, from, to)
	for _, v := range tempM {
		wantNextEpochCandidateM = append(wantNextEpochCandidateM, v.CommitteePublicKey())
	}

	from += 80
	to += 80
	wantCurrentEpochCandidateM := []incognitokey.CommitteePublicKey{}
	tempRootHash, tempM = storeCommitteeObjectOneShard(statedb.CurrentEpochShardCandidate, tempRootHash, statedb.CandidateShardID, from, to)
	for _, v := range tempM {
		wantCurrentEpochCandidateM = append(wantCurrentEpochCandidateM, v.CommitteePublicKey())
	}

	tempStateDB, err := statedb.NewWithPrefixTrie(tempRootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}

	gotCurrentValidatorM := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, ids)
	for _, id := range ids {
		temp, ok := gotCurrentValidatorM[id]
		if !ok {
			t.Fatalf("GetAllCommitteeState want shard %+v", id)
		}
		if len(temp) != 32 {
			t.Fatalf("GetAllCommitteeState shard %+v want key length %+v but got %+v", id, 32, len(temp))
		}
	}
	for id, wants := range wantCurrentValidatorM {
		flag := false
		for _, want := range wants {
			for _, got := range gotCurrentValidatorM[id] {
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

	gotSubstituteValidatorM := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.SubstituteValidator, ids)
	for _, id := range ids {
		temp, ok := gotSubstituteValidatorM[id]
		if !ok {
			t.Fatalf("GetAllCommitteeState want shard %+v", id)
		}
		if len(temp) != 8 {
			t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 8, len(temp))
		}
	}
	for id, wants := range wantSubstituteValidatorM {
		flag := false
		for _, want := range wants {
			for _, got := range gotSubstituteValidatorM[id] {
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

	gotNextEpochCandidateM := tempStateDB.GetAllCandidateCommitteePublicKey(statedb.NextEpochShardCandidate)
	if len(gotNextEpochCandidateM) != 80 {
		t.Fatalf("GetAllCandidateCommitteePublicKey want key length %+v but got %+v", to-from, len(gotNextEpochCandidateM))
	}
	for _, want := range wantNextEpochCandidateM {
		flag := false
		for _, got := range gotNextEpochCandidateM {
			if reflect.DeepEqual(got, want) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("GetAllCandidateCommitteePublicKey want %+v but didn't get anything", want)
		}
	}

	gotCurrentEpochCandidateM := tempStateDB.GetAllCandidateCommitteePublicKey(statedb.CurrentEpochShardCandidate)
	if len(gotNextEpochCandidateM) != 80 {
		t.Fatalf("GetAllCandidateCommitteePublicKey want key length %+v but got %+v", to-from, len(gotCurrentEpochCandidateM))
	}
	for _, want := range wantCurrentEpochCandidateM {
		flag := false
		for _, got := range gotCurrentEpochCandidateM {
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

func TestStateDB_GetMixCommitteeState(t *testing.T) {
	from, to := 0, 32
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantCurrentValidatorM := make(map[int][]incognitokey.CommitteePublicKey)
	wantSubstituteValidatorM := make(map[int][]incognitokey.CommitteePublicKey)
	wantNextEpochCandidateM := []incognitokey.CommitteePublicKey{}
	wantCurrentEpochCandidateM := []incognitokey.CommitteePublicKey{}
	wantRewardReceiverM := make(map[string]string)
	wantAutoStakingM := make(map[string]bool)
	rootHashes := []common.Hash{emptyRoot}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.CurrentValidator, rootHashes[index], id, from, to)
		from += 32
		to += 32
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantCurrentValidatorM[id] = append(wantCurrentValidatorM[id], v.CommitteePublicKey())
			tempCurrentValidatorString, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.CommitteePublicKey()})
			if err != nil {
				t.Fatal(err)
			}
			wantRewardReceiverM[tempCurrentValidatorString[0]] = v.RewardReceiver()
			wantAutoStakingM[tempCurrentValidatorString[0]] = v.AutoStaking()
		}
	}
	tempRootHash := rootHashes[8]
	to = from + 8
	rootHashes = []common.Hash{tempRootHash}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.SubstituteValidator, rootHashes[index], id, from, to)
		from += 8
		to += 8
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantSubstituteValidatorM[id] = append(wantSubstituteValidatorM[id], v.CommitteePublicKey())
			tempCurrentValidatorString, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.CommitteePublicKey()})
			if err != nil {
				t.Fatal(err)
			}
			wantRewardReceiverM[tempCurrentValidatorString[0]] = v.RewardReceiver()
			wantAutoStakingM[tempCurrentValidatorString[0]] = v.AutoStaking()
		}
	}
	to = from + 80
	tempRootHash = rootHashes[8]
	tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.NextEpochShardCandidate, tempRootHash, statedb.CandidateShardID, from, to)
	for _, v := range tempM {
		wantNextEpochCandidateM = append(wantNextEpochCandidateM, v.CommitteePublicKey())
		tempCurrentValidatorString, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.CommitteePublicKey()})
		if err != nil {
			t.Fatal(err)
		}
		wantRewardReceiverM[tempCurrentValidatorString[0]] = v.RewardReceiver()
		wantAutoStakingM[tempCurrentValidatorString[0]] = v.AutoStaking()
	}

	from += 80
	to += 80
	tempRootHash, tempM = storeCommitteeObjectOneShard(statedb.CurrentEpochShardCandidate, tempRootHash, statedb.CandidateShardID, from, to)
	for _, v := range tempM {
		wantCurrentEpochCandidateM = append(wantCurrentEpochCandidateM, v.CommitteePublicKey())
		tempCurrentValidatorString, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.CommitteePublicKey()})
		if err != nil {
			t.Fatal(err)
		}
		wantRewardReceiverM[tempCurrentValidatorString[0]] = v.RewardReceiver()
		wantAutoStakingM[tempCurrentValidatorString[0]] = v.AutoStaking()
	}

	tempStateDB, err := statedb.NewWithPrefixTrie(tempRootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}

	gotCurrentValidatorM, gotSubstituteValidatorM, gotNextEpochCandidateM, gotCurrentEpochCandidateM, _, _, gotRewardReceiverM, gotAutoStakingM := tempStateDB.GetAllCommitteeState(ids)
	for _, id := range ids {
		temp, ok := gotCurrentValidatorM[id]
		if !ok {
			t.Fatalf("GetAllCommitteeState want shard %+v", id)
		}
		if len(temp) != 32 {
			t.Fatalf("GetAllCommitteeState shard %+v want key length %+v but got %+v", id, 32, len(temp))
		}
	}
	for id, wants := range wantCurrentValidatorM {
		flag := false
		for _, want := range wants {
			for _, got := range gotCurrentValidatorM[id] {
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

	for _, id := range ids {
		temp, ok := gotSubstituteValidatorM[id]
		if !ok {
			t.Fatalf("GetAllCommitteeState want shard %+v", id)
		}
		if len(temp) != 8 {
			t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 8, len(temp))
		}
	}
	for id, wants := range wantSubstituteValidatorM {
		flag := false
		for _, want := range wants {
			for _, got := range gotSubstituteValidatorM[id] {
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

	if len(gotNextEpochCandidateM) != 80 {
		t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", to-from, len(gotNextEpochCandidateM))
	}
	for _, want := range wantNextEpochCandidateM {
		flag := false
		for _, got := range gotNextEpochCandidateM {
			if reflect.DeepEqual(got, want) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("GetAllCommitteeState want %+v but didn't get anything", want)
		}
	}

	if len(gotNextEpochCandidateM) != 80 {
		t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", to-from, len(gotCurrentEpochCandidateM))
	}
	for _, want := range wantCurrentEpochCandidateM {
		flag := false
		for _, got := range gotCurrentEpochCandidateM {
			if reflect.DeepEqual(got, want) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("GetAllCommitteeState want %+v but didn't get anything", want)
		}
	}

	for k1, v1 := range wantRewardReceiverM {
		if v2, ok := gotRewardReceiverM[k1]; !ok {
			t.Fatalf("GetAllCommitteeState want %+v but got nothing", k1)
		} else {
			if strings.Compare(v1, v2) != 0 {
				t.Fatalf("GetAllCommitteeState want %+v but got %+v", v1, v2)
			}
		}
	}

	for k1, v1 := range wantAutoStakingM {
		if v2, ok := gotAutoStakingM[k1]; !ok {
			t.Fatalf("GetAllCommitteeState want %+v but got nothing", k1)
		} else {
			if v1 != v2 {
				t.Fatalf("GetAllCommitteeState want %+v but got %+v", v1, v2)
			}
		}
	}
}
