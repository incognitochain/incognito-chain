package finishsync

import (
	"sync"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

type Manager struct {
	validators map[byte][]incognitokey.CommitteePublicKey
	mu         *sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		mu: &sync.RWMutex{},
	}
}

func NewManagerWithValue(validators map[byte][]incognitokey.CommitteePublicKey) *Manager {
	return &Manager{
		validators: validators,
		mu:         &sync.RWMutex{},
	}
}

func (manager *Manager) Clone() *Manager {
	res := NewManager()
	res.validators = make(map[byte][]incognitokey.CommitteePublicKey)
	for i, v := range manager.validators {
		res.validators[i] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(res.validators[i], v)
	}
	return res
}

func (manager *Manager) AddFinishedSyncValidators(
	validators []string,
	syncingValidators []incognitokey.CommitteePublicKey,
	shardID byte,
) {
	manager.mu.Lock()
	defer func() {
		manager.mu.Unlock()
	}()
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

func (manager *Manager) Validators(shardID byte) []incognitokey.CommitteePublicKey {
	manager.mu.RLock()
	defer func() {
		manager.mu.RUnlock()
	}()
	return manager.validators[shardID]
}

func (manager *Manager) RemoveValidators(validators []incognitokey.CommitteePublicKey, shardID byte) {
	manager.mu.Lock()
	defer func() {
		manager.mu.Unlock()
	}()
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
