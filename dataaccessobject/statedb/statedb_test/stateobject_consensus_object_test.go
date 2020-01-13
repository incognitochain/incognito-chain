package statedb_test

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/trie"
	"io/ioutil"
	"math/rand"
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

func storeCommitteeObjectOneShardForTestConsensus(role int, initRoot common.Hash, shardID, from, to int) (common.Hash, map[common.Hash]*statedb.CommitteeState) {
	m := make(map[common.Hash]*statedb.CommitteeState)
	tempCommitteePublicKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
	if err != nil {
		panic(err)
	}
	tempCommitteePublicKey = tempCommitteePublicKey[from:to]
	for index, value := range tempCommitteePublicKey {
		key, _ := statedb.GenerateCommitteeObjectKeyWithRole(role, shardID, value)
		autoStaking := false
		if rand.Int()%2 == 0 {
			autoStaking = true
		}
		committeeState := statedb.NewCommitteeStateWithValue(shardID, role, value, receiverPaymentAddress[index], autoStaking)
		m[key] = committeeState
	}
	sDB, err := statedb.NewWithPrefixTrie(initRoot, warperDBAllTest)
	if err != nil {
		panic(err)
	}
	for key, value := range m {
		sDB.SetStateObject(statedb.CommitteeObjectType, key, value)
	}
	rootHash, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		panic(err)
	}
	return rootHash, m
}

func storeAllConsensusStateObjectForTesting(initRoot common.Hash) (
	common.Hash,
	map[int][]incognitokey.CommitteePublicKey,
	map[int][]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	map[string]string,
	map[string]bool,
	map[string]map[common.Hash]uint64,
	map[common.Hash]*statedb.RewardRequestState,
	map[string]uint8,
) {
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	shardIDs := []byte{0, 1, 2, 3, 4, 5, 6, 7}
	wantMCommittee := make(map[int][]incognitokey.CommitteePublicKey)
	wantMSubstituteValidator := make(map[int][]incognitokey.CommitteePublicKey)
	wantNextEpochCandidate := []incognitokey.CommitteePublicKey{}
	wantCurrentEpochCandidate := []incognitokey.CommitteePublicKey{}
	wantMRewardReceiver := make(map[string]string)
	wantMAutoStaking := make(map[string]bool)
	wantMCommitteeReward := make(map[string]map[common.Hash]uint64)
	wantMRewardRequest := make(map[common.Hash]*statedb.RewardRequestState)
	wantMBlackListProducer := make(map[string]uint8)
	// Committee
	from, to := 0, 32
	wantCurrentValidatorM := make(map[int][]incognitokey.CommitteePublicKey)
	rootHashes := []common.Hash{emptyRoot}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShardForTestConsensus(statedb.CurrentValidator, rootHashes[index], id, from, to)
		from += 32
		to += 32
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantCurrentValidatorM[id] = append(wantCurrentValidatorM[id], v.CommitteePublicKey())
			tempString, _ := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.CommitteePublicKey()})
			wantMRewardReceiver[tempString[0]] = v.RewardReceiver()
			wantMAutoStaking[tempString[0]] = v.AutoStaking()
		}
	}

	tempRootHash := rootHashes[8]
	to = from + 8
	rootHashes = []common.Hash{tempRootHash}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShardForTestConsensus(statedb.SubstituteValidator, rootHashes[index], id, from, to)
		from += 8
		to += 8
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantMSubstituteValidator[id] = append(wantMSubstituteValidator[id], v.CommitteePublicKey())
			tempString, _ := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.CommitteePublicKey()})
			wantMRewardReceiver[tempString[0]] = v.RewardReceiver()
			wantMAutoStaking[tempString[0]] = v.AutoStaking()
		}
	}
	to = from + 80
	tempRootHash = rootHashes[8]
	tempRootHash, tempM := storeCommitteeObjectOneShardForTestConsensus(statedb.NextEpochShardCandidate, tempRootHash, statedb.CandidateShardID, from, to)
	for _, v := range tempM {
		wantNextEpochCandidate = append(wantNextEpochCandidate, v.CommitteePublicKey())
		tempString, _ := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.CommitteePublicKey()})
		wantMRewardReceiver[tempString[0]] = v.RewardReceiver()
		wantMAutoStaking[tempString[0]] = v.AutoStaking()
	}

	from += 80
	to += 80
	tempRootHash, tempM = storeCommitteeObjectOneShardForTestConsensus(statedb.CurrentEpochShardCandidate, tempRootHash, statedb.CandidateShardID, from, to)
	for _, v := range tempM {
		wantCurrentEpochCandidate = append(wantCurrentEpochCandidate, v.CommitteePublicKey())
		tempString, _ := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{v.CommitteePublicKey()})
		wantMRewardReceiver[tempString[0]] = v.RewardReceiver()
		wantMAutoStaking[tempString[0]] = v.AutoStaking()
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(tempRootHash, warperDBAllTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	tempRootHash, _, wantMCommitteeReward = storeCommitteeReward(tempRootHash, warperDBAllTest)
	tempRootHash, wantMRewardRequest = storeRewardRequest(tempRootHash, warperDBAllTest, defaultMaxEpoch, shardIDs)
	tempRootHash, _, wantMBlackListProducer = storeBlackListProducer(tempRootHash, warperDBAllTest, 1, 460, 500)
	return tempRootHash, wantMCommittee, wantMSubstituteValidator, wantNextEpochCandidate, wantCurrentEpochCandidate, wantMRewardReceiver, wantMAutoStaking, wantMCommitteeReward, wantMRewardRequest, wantMBlackListProducer
}

func TestStateDB_GetAllConsensusStateObject(t *testing.T) {
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	rootHash, wantMCommittee, wantMSubstituteValidator, wantNextEpochCandidate, wantCurrentEpochCandidate, wantMRewardReceiver, wantMAutoStaking, wantMCommitteeReward, wantMRewardRequest, wantMBlackListProducer := storeAllConsensusStateObjectForTesting(emptyRoot)
	// GOT to verify
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBAllTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	gotMCommittee := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, ids)
	for _, id := range ids {
		temp, ok := gotMCommittee[id]
		if !ok {
			t.Fatalf("GetAllValidatorCommitteePublicKey want shard %+v", id)
		}
		if len(temp) != 32 {
			t.Fatalf("GetAllValidatorCommitteePublicKey want key length %+v but got %+v", 32, len(temp))
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
				t.Fatalf("GetAllValidatorCommitteePublicKey shard %+v want %+v but didn't get anything", id, want)
			}
		}
	}

	gotMSubstituteValidator := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.SubstituteValidator, ids)
	for _, id := range ids {
		temp, ok := gotMSubstituteValidator[id]
		if !ok {
			t.Fatalf("GetAllValidatorCommitteePublicKey want shard %+v", id)
		}
		if len(temp) != 8 {
			t.Fatalf("GetAllValidatorCommitteePublicKey want key length %+v but got %+v", 8, len(temp))
		}
	}
	for id, wants := range wantMSubstituteValidator {
		flag := false
		for _, want := range wants {
			for _, got := range gotMSubstituteValidator[id] {
				if reflect.DeepEqual(got, want) {
					flag = true
					break
				}
			}
			if !flag {
				t.Fatalf("GetAllValidatorCommitteePublicKey shard %+v want %+v but didn't get anything", id, want)
			}
		}
	}

	gotNextEpochCandidate := tempStateDB.GetAllCandidateCommitteePublicKey(statedb.NextEpochShardCandidate)
	if len(gotNextEpochCandidate) != 80 {
		t.Fatalf("GetAllCandidateCommitteePublicKey want key length %+v but got %+v", 80, len(gotNextEpochCandidate))
	}
	for id, want := range wantNextEpochCandidate {
		flag := false
		for _, got := range gotNextEpochCandidate {
			if reflect.DeepEqual(got, want) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("GetAllCandidateCommitteePublicKey shard %+v want %+v but didn't get anything", id, want)
		}
	}

	gotCurrentEpochCandidate := tempStateDB.GetAllCandidateCommitteePublicKey(statedb.CurrentEpochShardCandidate)
	if len(gotCurrentEpochCandidate) != 80 {
		t.Fatalf("GetAllCandidateCommitteePublicKey want key length %+v but got %+v", 80, len(gotCurrentEpochCandidate))
	}
	for id, want := range wantCurrentEpochCandidate {
		flag := false
		for _, got := range gotCurrentEpochCandidate {
			if reflect.DeepEqual(got, want) {
				flag = true
				break
			}
		}
		if !flag {
			t.Fatalf("GetAllCandidateCommitteePublicKey %+v want %+v but didn't get anything", id, want)
		}
	}

	_, _, _, _, _, _, gotMRewardReceiver, gotMAutoStaking := tempStateDB.GetAllCommitteeState(ids)
	for k, v1 := range gotMRewardReceiver {
		if v2, ok := wantMRewardReceiver[k]; !ok {
			t.Fatalf("want %+v but get nothing", k)
		} else {
			if !reflect.DeepEqual(v2, v1) {
				t.Fatalf("want %+v but got %+v", v2, v1)
			}
		}
	}
	for k, v1 := range gotMAutoStaking {
		if v2, ok := wantMAutoStaking[k]; !ok {
			t.Fatalf("want %+v but get nothing", k)
		} else {
			if !reflect.DeepEqual(v2, v1) {
				t.Fatalf("want %+v but got %+v", v2, v1)
			}
		}
	}
	gotMCommitteeReward := tempStateDB.GetAllCommitteeReward()
	for k1, v1 := range wantMCommitteeReward {
		if v2, ok := gotMCommitteeReward[k1]; !ok {
			t.Fatalf("want %+v but got nothing", k1)
		} else {
			for k11, v11 := range v1 {
				if v22, ok := v2[k11]; !ok {
					t.Fatalf("want %+v but got nothing", k11)
				} else {
					if v11 != v22 {
						t.Fatalf("want %+v but got %+v", v11, v22)
					}
				}
			}
		}
	}
	for k, v := range wantMRewardRequest {
		gotRewardRequest, has, err := tempStateDB.GetRewardRequestAmount(k)
		if err != nil {
			t.Fatal(err)
		}
		if !has {
			t.Fatal(has)
		}
		if v.Amount() != gotRewardRequest {
			t.Fatalf("want %+v but got %+v", v.Amount(), gotRewardRequest)
		}
	}
	gotMBlackListProducer := tempStateDB.GetAllProducerBlackList()
	for k, v := range wantMBlackListProducer {
		if v2, ok := gotMBlackListProducer[k]; !ok {
			t.Fatalf("want %+v but got nothing", k)
		} else if v != v2 {
			t.Fatalf("want %+v but got %+v", v, v2)
		}
	}
}

func BenchmarkStateDB_GetAllCommitteeRewardInFullData(b *testing.B) {
	rootHash, _, _, _, _, _, _, _, _, _ := storeAllConsensusStateObjectForTesting(emptyRoot)
	// GOT to verify
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBAllTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetAllCommitteeReward()
	}
}

func BenchmarkStateDB_GetAllAutoStakingInFullData(b *testing.B) {
	rootHash, _, _, _, _, _, _, _, _, _ := storeAllConsensusStateObjectForTesting(emptyRoot)
	// GOT to verify
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBAllTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetAllCommitteeState([]int{0, 1, 2, 3, 4, 5, 6, 7, 8})
	}
}
func BenchmarkStateDB_GetAllCommitteeInFullData(b *testing.B) {
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	rootHash, _, _, _, _, _, _, _, _, _ := storeAllConsensusStateObjectForTesting(emptyRoot)
	// GOT to verify
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBAllTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, ids)
	}
}
