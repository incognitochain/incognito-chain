package finishsync

import (
	"sort"
	"sync"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

type FinishSyncManager struct {
	validators map[byte][]incognitokey.CommitteePublicKey
	mu         *sync.RWMutex
}

func NewManager() *FinishSyncManager {
	return &FinishSyncManager{
		mu: &sync.RWMutex{},
	}
}

func NewManagerWithValue(validators map[byte][]incognitokey.CommitteePublicKey) *FinishSyncManager {
	return &FinishSyncManager{
		validators: validators,
		mu:         &sync.RWMutex{},
	}
}

func (manager *FinishSyncManager) Clone() *FinishSyncManager {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	res := NewManager()
	res.validators = make(map[byte][]incognitokey.CommitteePublicKey)
	for i, v := range manager.validators {
		res.validators[i] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(res.validators[i], v)
	}
	return res
}

//AddFinishedSyncValidators only add validators in valid syncing list and not duplicate in FinishSyncManager.validator list
func (manager *FinishSyncManager) AddFinishedSyncValidators(
	validators []string,
	syncingValidators []incognitokey.CommitteePublicKey,
	shardID byte,
) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	finishedSyncValidators := make(map[string]bool)
	for _, v := range manager.validators[shardID] {
		key, _ := v.ToBase58()
		finishedSyncValidators[key] = true
	}

	validKeys := []string{}
	for _, v := range validators {
		if finishedSyncValidators[v] {
			continue
		}
		validKeys = append(validKeys, v)
	}

	finishedSyncValidators = make(map[string]bool)
	for _, v := range validKeys {
		finishedSyncValidators[v] = true
	}
	count := 0
	lenValidKeys := len(validKeys)
	validKeys = []string{}
	for _, v := range syncingValidators {
		if count == lenValidKeys {
			break
		}
		key, _ := v.ToBase58()
		if finishedSyncValidators[key] {
			validKeys = append(validKeys, key)
			count++
		}
	}
	committeePublicKeys, _ := incognitokey.CommitteeBase58KeyListToStruct(validKeys)
	manager.validators[shardID] = append(manager.validators[shardID], committeePublicKeys...)
}

func (manager *FinishSyncManager) Validators(shardID byte) []incognitokey.CommitteePublicKey {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	res := make([]incognitokey.CommitteePublicKey, len(manager.validators[shardID]))
	copy(res, manager.validators[shardID])
	return res
}

//RemoveValidators only remove FinishSyncManager.validator list ONCE
// ignore validators not in FinishSyncManager.validators list
func (manager *FinishSyncManager) RemoveValidators(validators []incognitokey.CommitteePublicKey, shardID byte) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	finishedSyncValidators := make(map[string]bool)
	for _, v := range validators {
		key, _ := v.ToBase58()
		finishedSyncValidators[key] = true
	}
	count := 0
	for i := 0; i < len(manager.validators[shardID]); i++ {
		v := manager.validators[shardID][i]
		if count == len(validators) {
			break
		}
		key, _ := v.ToBase58()
		if finishedSyncValidators[key] {
			manager.validators[shardID] = append(
				manager.validators[shardID][:i], manager.validators[shardID][i+1:]...)
			i--
			count++
		}
	}
}

//Instructions ....
func (manager *FinishSyncManager) Instructions() []*instruction.FinishSyncInstruction {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	res := []*instruction.FinishSyncInstruction{}
	keys := []int{}
	for i := 0; i < len(manager.validators); i++ {
		keys = append(keys, i)
	}
	sort.Ints(keys)
	for _, v := range keys {
		committeePublicKeys, _ := incognitokey.CommitteeKeyListToString(manager.validators[byte(v)])
		finishSyncInstruction := instruction.NewFinishSyncInstructionWithValue(v, committeePublicKeys)
		res = append(res, finishSyncInstruction)
	}
	return res
}
