package statedb_test

import (
	"math/rand"
	"reflect"
	"strconv"
	"testing"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	_ "github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func storeCommitteeObjectOneShard(role int, initRoot common.Hash, shardID, from, to int) (common.Hash, map[common.Hash]*statedb.CommitteeState) {
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
	sDB, err := statedb.NewWithPrefixTrie(initRoot, warperDBCommitteeTest)
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

func TestStateDB_SetStateObjectCommitteeState(t *testing.T) {
	var shardID = 0
	var err error = nil
	tempCommitteePublicKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
	if err != nil {
		panic(err)
	}
	sampleCommittee := tempCommitteePublicKey[0]
	sampleCommittee2 := tempCommitteePublicKey[1]
	key, _ := statedb.GenerateCommitteeObjectKeyWithRole(statedb.CurrentValidator, shardID, sampleCommittee)
	key2, _ := statedb.GenerateCommitteeObjectKeyWithRole(statedb.CurrentValidator, shardID, sampleCommittee2)
	committeeState := statedb.NewCommitteeStateWithValue(shardID, statedb.CurrentValidator, sampleCommittee, receiverPaymentAddress[0], true)
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBCommitteeTest)
	if err != nil {
		panic(err)
	}
	err = sDB.SetStateObject(statedb.CommitteeObjectType, key, committeeState)
	if err != nil {
		t.Fatal(err)
	}
	err = sDB.SetStateObject(statedb.CommitteeObjectType, key, "committeeState")
	if err == nil {
		t.Fatal("expect error")
	}
	err = sDB.SetStateObject(statedb.CommitteeObjectType, key, []byte("committee state"))
	if err == nil {
		t.Fatal("expect error")
	}
	err = sDB.SetStateObject(statedb.CommitteeObjectType, key2, []byte("committee state"))
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
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err)
	}
	got, has, err := tempStateDB.GetCommitteeState(key)
	if err != nil {
		t.Fatal(err)
	}
	if !has {
		t.Fatal(has)
	}
	if !reflect.DeepEqual(got, committeeState) {
		t.Fatalf("want value %+v but got %+v", committeeState, got)
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

func TestStateDB_GetCurrentValidatorCommitteeState(t *testing.T) {
	rootHash, m := storeCommitteeObjectOneShard(statedb.CurrentValidator, emptyRoot, 0, 0, len(committeePublicKeys))
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

func TestStateDB_GetAllCurrentValidatorCommitteeKey512OneShard(t *testing.T) {
	from, to := 0, 512
	rootHash, m := storeCommitteeObjectOneShard(statedb.CurrentValidator, emptyRoot, 0, from, to)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	gotM := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, []int{0})
	gotMShard0, ok := gotM[0]
	if !ok {
		t.Fatalf("GetAllCommitteeState want shard %+v", 0)
	}
	if len(gotMShard0) != to-from {
		t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", to-from, len(gotM))
	}
	for _, want := range m {
		flag := false
		for _, got := range gotMShard0 {
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

func TestStateDB_GetAllCurrentValidatorCommitteeKey512EightShard(t *testing.T) {
	from, to := 0, 64
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantM := make(map[int][]incognitokey.CommitteePublicKey)
	rootHashes := []common.Hash{emptyRoot}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.CurrentValidator, rootHashes[index], id, from, to)
		from += 64
		to += 64
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantM[id] = append(wantM[id], v.CommitteePublicKey())
		}
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHashes[8], warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	gotM := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, ids)
	for _, id := range ids {
		temp, ok := gotM[id]
		if !ok {
			t.Fatalf("GetAllCommitteeState want shard %+v", id)
		}
		if len(temp) != 64 {
			t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 64, len(temp))
		}
	}
	for id, wants := range wantM {
		flag := false
		for _, want := range wants {
			for _, got := range gotM[id] {
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
}

func TestStateDB_GetAllCurrentValidatorCommitteePublicKey512EightShardMultipleRootHash(t *testing.T) {
	from, to := 0, 32
	maxHeight := 8
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantM := make(map[int][]incognitokey.CommitteePublicKey)
	wantMs := []map[int][]incognitokey.CommitteePublicKey{}
	rootHashesByHeight := []common.Hash{}
	rootHashes := []common.Hash{emptyRoot}
	committeePublicKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
	if err != nil {
		t.Fatal(err)
	}
	committeePublicKey = committeePublicKey[0:512]
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.CurrentValidator, rootHashes[index], id, from, to)
		from += 32
		to += 32
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantM[id] = append(wantM[id], v.CommitteePublicKey())
		}
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHashes[8], warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	gotM := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, ids)
	for _, id := range ids {
		temp, ok := gotM[id]
		if !ok {
			t.Fatalf("GetAllCommitteeState want shard %+v", id)
		}
		if len(temp) != 32 {
			t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 32, len(temp))
		}
	}
	for id, wants := range wantM {
		flag := false
		for _, want := range wants {
			for _, got := range gotM[id] {
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
	wantMs = append(wantMs, wantM)
	rootHashesByHeight = append(rootHashesByHeight, rootHashes[8])
	from = 256
	to = 260
	for i := 1; i < maxHeight; i++ {
		sDB, err := statedb.NewWithPrefixTrie(rootHashesByHeight[i-1], warperDBCommitteeTest)
		if err != nil {
			t.Fatalf("height %+v, err %+v", i, err)
		}
		prevWantM := wantMs[i-1]
		newWantMState := make(map[common.Hash]*statedb.CommitteeState)
		for shardID, publicKeys := range prevWantM {
			for index, value := range publicKeys {
				key, _ := statedb.GenerateCommitteeObjectKeyWithRole(statedb.CurrentValidator, shardID, value)
				committeeState := statedb.NewCommitteeStateWithValue(shardID, statedb.CurrentValidator, value, receiverPaymentAddress[index], true)
				newWantMState[key] = committeeState
			}
		}
		newWantM := make(map[int][]incognitokey.CommitteePublicKey)
		newAddedMState := make(map[common.Hash]*statedb.CommitteeState)
		for _, shardID := range ids {
			tempCommitteePublicKey := committeePublicKey[from:to]
			for index, value := range tempCommitteePublicKey {
				key, _ := statedb.GenerateCommitteeObjectKeyWithRole(statedb.CurrentValidator, shardID, value)
				committeeState := statedb.NewCommitteeStateWithValue(shardID, statedb.CurrentValidator, value, receiverPaymentAddress[index], true)
				newAddedMState[key] = committeeState
			}
			from += 4
			to += 4
			prevWantMStateByShardID := make(map[common.Hash]*statedb.CommitteeState)
			for index, value := range prevWantM[shardID] {
				key, _ := statedb.GenerateCommitteeObjectKeyWithRole(statedb.CurrentValidator, shardID, value)
				committeeState := statedb.NewCommitteeStateWithValue(shardID, statedb.CurrentValidator, value, receiverPaymentAddress[index], true)
				prevWantMStateByShardID[key] = committeeState
			}
			count := 0
			for key, _ := range prevWantMStateByShardID {
				ok := sDB.MarkDeleteStateObject(statedb.CommitteeObjectType, key)
				if sDB.Error() != nil {
					t.Fatal(sDB.Error())
				}
				if !ok {
					t.Fatal("can't mark delete state object " + strconv.Itoa(i) + " " + strconv.Itoa(count))
				}
				delete(newWantMState, key)
				count++
				if count == 4 {
					break
				}
			}
			for k, v := range newAddedMState {
				err := sDB.SetStateObject(statedb.CommitteeObjectType, k, v)
				if err != nil {
					t.Fatal(err)
				}
				newWantMState[k] = v
			}
		}
		for _, v := range newWantMState {
			newWantM[v.ShardID()] = append(newWantM[v.ShardID()], v.CommitteePublicKey())
		}
		rootHash, err := sDB.Commit(true)
		if err != nil {
			t.Fatal(err)
		}
		err = sDB.Database().TrieDB().Commit(rootHash, false)
		if err != nil {
			t.Fatal(err)
		}
		rootHashesByHeight = append(rootHashesByHeight, rootHash)
		wantMs = append(wantMs, newWantM)
	}
	for index, rootHash := range rootHashesByHeight {
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
		if err != nil {
			t.Fatal(err)
		}
		gotM := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, ids)
		for _, id := range ids {
			temp, ok := gotM[id]
			if !ok {
				t.Fatalf("GetAllCommitteeState want shard %+v", id)
			}
			if len(temp) != 32 {
				t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 32, len(temp))
			}
		}
		for id, wants := range wantMs[index] {
			flag := false
			for _, want := range wants {
				for _, got := range gotM[id] {
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
	}
}

func TestStateDB_GetCurrentValidatorCommitteePublicKeyByShardIDState512EightShard(t *testing.T) {
	from, to := 0, 64
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantM := make(map[int][]incognitokey.CommitteePublicKey)
	rootHashes := []common.Hash{emptyRoot}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.CurrentValidator, rootHashes[index], id, from, to)
		from += 64
		to += 64
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantM[id] = append(wantM[id], v.CommitteePublicKey())
		}
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHashes[8], warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	for _, id := range ids {
		gotM := tempStateDB.GetByShardIDCurrentValidatorState(id)
		if len(gotM) != 64 {
			t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 64, len(gotM))
		}
		for _, want := range wantM[id] {
			flag := false
			for _, got := range gotM {
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
}

// SUBSTITUTE VALIDATOR===============================================================
func TestStateDB_GetSubstituteValidatorCommitteeState(t *testing.T) {
	rootHash, m := storeCommitteeObjectOneShard(statedb.SubstituteValidator, emptyRoot, 0, 0, len(committeePublicKeys))
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

func TestStateDB_GetSubstituteValidatorCommitteePublicKeyByShardIDState512EightShard(t *testing.T) {
	from, to := 0, 64
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantM := make(map[int][]incognitokey.CommitteePublicKey)
	rootHashes := []common.Hash{emptyRoot}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.SubstituteValidator, rootHashes[index], id, from, to)
		from += 64
		to += 64
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantM[id] = append(wantM[id], v.CommitteePublicKey())
		}
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHashes[8], warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	for _, id := range ids {
		gotM := tempStateDB.GetByShardIDSubstituteValidatorState(id)
		if len(gotM) != 64 {
			t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 64, len(gotM))
		}
		for _, want := range wantM[id] {
			flag := false
			for _, got := range gotM {
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
}

func TestStateDB_GetAllSubstituteValidatorCommitteeKey512OneShard(t *testing.T) {
	from, to := 0, 512
	rootHash, m := storeCommitteeObjectOneShard(statedb.SubstituteValidator, emptyRoot, 0, from, to)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	gotM := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.SubstituteValidator, []int{0})
	gotMShard0, ok := gotM[0]
	if !ok {
		t.Fatalf("GetAllCommitteeState want shard %+v", 0)
	}
	if len(gotMShard0) != to-from {
		t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", to-from, len(gotM))
	}
	for _, want := range m {
		flag := false
		for _, got := range gotMShard0 {
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

func TestStateDB_GetAllSubstituteValidatorCommitteeKey512EightShard(t *testing.T) {
	from, to := 0, 64
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantM := make(map[int][]incognitokey.CommitteePublicKey)
	rootHashes := []common.Hash{emptyRoot}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.SubstituteValidator, rootHashes[index], id, from, to)
		from += 64
		to += 64
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantM[id] = append(wantM[id], v.CommitteePublicKey())
		}
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHashes[8], warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	gotM := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.SubstituteValidator, ids)
	for _, id := range ids {
		temp, ok := gotM[id]
		if !ok {
			t.Fatalf("GetAllCommitteeState want shard %+v", id)
		}
		if len(temp) != 64 {
			t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 64, len(temp))
		}
	}
	for id, wants := range wantM {
		flag := false
		for _, want := range wants {
			for _, got := range gotM[id] {
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
}

func TestStateDB_AllSubstituteValidatorCommitteePublicKey512EightShardMultipleRootHash(t *testing.T) {
	from, to := 0, 32
	maxHeight := 8
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantM := make(map[int][]incognitokey.CommitteePublicKey)
	wantMs := []map[int][]incognitokey.CommitteePublicKey{}
	rootHashesByHeight := []common.Hash{}
	rootHashes := []common.Hash{emptyRoot}
	committeePublicKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
	if err != nil {
		t.Fatal(err)
	}
	committeePublicKey = committeePublicKey[0:512]
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.SubstituteValidator, rootHashes[index], id, from, to)
		from += 32
		to += 32
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantM[id] = append(wantM[id], v.CommitteePublicKey())
		}
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHashes[8], warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		t.Fatal(err, tempStateDB)
	}
	gotM := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.SubstituteValidator, ids)
	for _, id := range ids {
		temp, ok := gotM[id]
		if !ok {
			t.Fatalf("GetAllCommitteeState want shard %+v", id)
		}
		if len(temp) != 32 {
			t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 32, len(temp))
		}
	}
	for id, wants := range wantM {
		flag := false
		for _, want := range wants {
			for _, got := range gotM[id] {
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
	wantMs = append(wantMs, wantM)
	rootHashesByHeight = append(rootHashesByHeight, rootHashes[8])
	from = 256
	to = 260
	for i := 1; i < maxHeight; i++ {
		sDB, err := statedb.NewWithPrefixTrie(rootHashesByHeight[i-1], warperDBCommitteeTest)
		if err != nil {
			t.Fatalf("height %+v, err %+v", i, err)
		}
		prevWantM := wantMs[i-1]
		newWantMState := make(map[common.Hash]*statedb.CommitteeState)
		for shardID, publicKeys := range prevWantM {
			for index, value := range publicKeys {
				key, _ := statedb.GenerateCommitteeObjectKeyWithRole(statedb.SubstituteValidator, shardID, value)
				committeeState := statedb.NewCommitteeStateWithValue(shardID, statedb.SubstituteValidator, value, receiverPaymentAddress[index], true)
				newWantMState[key] = committeeState
			}
		}
		newWantM := make(map[int][]incognitokey.CommitteePublicKey)
		newAddedMState := make(map[common.Hash]*statedb.CommitteeState)
		for _, shardID := range ids {
			tempCommitteePublicKey := committeePublicKey[from:to]
			for index, value := range tempCommitteePublicKey {
				key, _ := statedb.GenerateCommitteeObjectKeyWithRole(statedb.SubstituteValidator, shardID, value)
				committeeState := statedb.NewCommitteeStateWithValue(shardID, statedb.SubstituteValidator, value, receiverPaymentAddress[index], true)
				newAddedMState[key] = committeeState
			}
			from += 4
			to += 4
			prevWantMStateByShardID := make(map[common.Hash]*statedb.CommitteeState)
			for index, value := range prevWantM[shardID] {
				key, _ := statedb.GenerateCommitteeObjectKeyWithRole(statedb.SubstituteValidator, shardID, value)
				committeeState := statedb.NewCommitteeStateWithValue(shardID, statedb.SubstituteValidator, value, receiverPaymentAddress[index], true)
				prevWantMStateByShardID[key] = committeeState
			}
			count := 0
			for key, _ := range prevWantMStateByShardID {
				ok := sDB.MarkDeleteStateObject(statedb.CommitteeObjectType, key)
				if sDB.Error() != nil {
					t.Fatal(sDB.Error())
				}
				if !ok {
					t.Fatal("can't mark delete state object " + strconv.Itoa(i) + " " + strconv.Itoa(count))
				}
				delete(newWantMState, key)
				count++
				if count == 4 {
					break
				}
			}
			for k, v := range newAddedMState {
				err := sDB.SetStateObject(statedb.CommitteeObjectType, k, v)
				if err != nil {
					t.Fatal(err)
				}
				newWantMState[k] = v
			}
		}
		for _, v := range newWantMState {
			newWantM[v.ShardID()] = append(newWantM[v.ShardID()], v.CommitteePublicKey())
		}
		rootHash, err := sDB.Commit(true)
		if err != nil {
			t.Fatal(err)
		}
		err = sDB.Database().TrieDB().Commit(rootHash, false)
		if err != nil {
			t.Fatal(err)
		}
		rootHashesByHeight = append(rootHashesByHeight, rootHash)
		wantMs = append(wantMs, newWantM)
	}
	for index, rootHash := range rootHashesByHeight {
		tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
		if err != nil {
			t.Fatal(err)
		}
		gotM := tempStateDB.GetAllValidatorCommitteePublicKey(statedb.SubstituteValidator, ids)
		for _, id := range ids {
			temp, ok := gotM[id]
			if !ok {
				t.Fatalf("GetAllCommitteeState want shard %+v", id)
			}
			if len(temp) != 32 {
				t.Fatalf("GetAllCommitteeState want key length %+v but got %+v", 32, len(temp))
			}
		}
		for id, wants := range wantMs[index] {
			flag := false
			for _, want := range wants {
				for _, got := range gotM[id] {
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
	}
}

// SUBSTITUTE VALIDATOR===============================================================

func BenchmarkStateDB_GetCurrentValidatorCommitteePublicKey512EightShard(b *testing.B) {
	from, to := 0, 64
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantM := make(map[int][]incognitokey.CommitteePublicKey)
	rootHashes := []common.Hash{emptyRoot}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.CurrentValidator, rootHashes[index], id, from, to)
		from += 64
		to += 64
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantM[id] = append(wantM[id], v.CommitteePublicKey())
		}
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHashes[8], warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		for _, id := range ids {
			tempStateDB.GetByShardIDCurrentValidatorState(id)
		}
	}
}

func BenchmarkStateDB_GetAllCurrentCandidateCommitteePublicKey512EightShard(b *testing.B) {
	from, to := 0, 64
	ids := []int{0, 1, 2, 3, 4, 5, 6, 7}
	wantM := make(map[int][]incognitokey.CommitteePublicKey)
	rootHashes := []common.Hash{emptyRoot}
	for index, id := range ids {
		tempRootHash, tempM := storeCommitteeObjectOneShard(statedb.CurrentValidator, rootHashes[index], id, from, to)
		from += 64
		to += 64
		rootHashes = append(rootHashes, tempRootHash)
		for _, v := range tempM {
			wantM[id] = append(wantM[id], v.CommitteePublicKey())
		}
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHashes[8], warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, ids)
	}
}
func BenchmarkStateDB_GetAllCurrentCandidateCommitteePublicKey512OneShard(b *testing.B) {
	from, to := 0, 512
	rootHash, _ := storeCommitteeObjectOneShard(statedb.CurrentValidator, emptyRoot, 0, from, to)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		tempStateDB.GetAllValidatorCommitteePublicKey(statedb.CurrentValidator, []int{0})
	}
}

func BenchmarkStateDB_GetCommitteeState512OneShard(b *testing.B) {
	rootHash, m := storeCommitteeObjectOneShard(statedb.CurrentValidator, emptyRoot, 0, 0, 512)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		for key, want := range m {
			got, has, err := tempStateDB.GetCommitteeState(key)
			if err != nil {
				panic(err)
			}
			if !has {
				panic(key)
			}
			if !reflect.DeepEqual(got, want) {
				panic(key)
			}
		}
	}
}

func BenchmarkStateDB_GetCommitteeState256OneShard(b *testing.B) {
	rootHash, m := storeCommitteeObjectOneShard(statedb.CurrentValidator, emptyRoot, 0, 0, 256)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for n := 0; n < b.N; n++ {
		for key, want := range m {
			got, has, err := tempStateDB.GetCommitteeState(key)
			if err != nil {
				panic(err)
			}
			if !has {
				panic(key)
			}
			if !reflect.DeepEqual(got, want) {
				panic(key)
			}
		}
	}
}

func BenchmarkStateDB_GetCommitteeState1In1(b *testing.B) {
	key := common.Hash{}
	shardID := 0
	tempCommitteePublicKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
	if err != nil {
		panic(err)
	}
	sampleCommitteePublicKey := tempCommitteePublicKey[0]
	key, _ = statedb.GenerateCommitteeObjectKeyWithRole(statedb.CurrentValidator, shardID, sampleCommitteePublicKey)
	committeeState := statedb.NewCommitteeStateWithValue(shardID, statedb.CurrentValidator, sampleCommitteePublicKey, receiverPaymentAddress[0], true)
	sDB, err := statedb.NewWithPrefixTrie(emptyRoot, warperDBCommitteeTest)
	if err != nil {
		panic(err)
	}
	sDB.SetStateObject(statedb.CommitteeObjectType, key, committeeState)
	rootHash, err := sDB.Commit(true)
	if err != nil {
		panic(err)
	}
	err = sDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		panic(err)
	}
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		tempStateDB.GetCommitteeState(key)
	}
}
func BenchmarkStateDB_GetCommitteeState1In64(b *testing.B) {
	key := common.Hash{}
	shardID := 0
	tempCommitteePublicKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
	if err != nil {
		panic(err)
	}
	sampleCommitteePublicKey := tempCommitteePublicKey[0]
	key, _ = statedb.GenerateCommitteeObjectKeyWithRole(statedb.CurrentValidator, shardID, sampleCommitteePublicKey)
	rootHash, _ := storeCommitteeObjectOneShard(statedb.CurrentValidator, emptyRoot, 0, 0, 64)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		tempStateDB.GetCommitteeState(key)
	}
}
func BenchmarkStateDB_GetCommitteeState1In256(b *testing.B) {
	key := common.Hash{}
	shardID := 0
	tempCommitteePublicKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeePublicKeys)
	if err != nil {
		panic(err)
	}
	sampleCommitteePublicKey := tempCommitteePublicKey[0]
	key, _ = statedb.GenerateCommitteeObjectKeyWithRole(statedb.CurrentValidator, shardID, sampleCommitteePublicKey)
	rootHash, _ := storeCommitteeObjectOneShard(statedb.CurrentValidator, emptyRoot, 0, 0, 256)
	tempStateDB, err := statedb.NewWithPrefixTrie(rootHash, warperDBCommitteeTest)
	if err != nil || tempStateDB == nil {
		panic(err)
	}
	for i := 0; i < b.N; i++ {
		tempStateDB.GetCommitteeState(key)
	}
}
