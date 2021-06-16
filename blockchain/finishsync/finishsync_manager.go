package finishsync

import (
	"github.com/incognitochain/incognito-chain/common"
	"sort"
	"sync"

	"github.com/incognitochain/incognito-chain/instruction"
)

// FinishSyncManager manages FinishedSyncValidators in sync pool of attached view
// FinishedSyncValidators must only compatible with sync pool. It must only contain keys that also in sync pool
// FinishSyncManager could maintain different data in different beacon nodes
type FinishSyncManager struct {
	FinishedSyncValidators map[byte]map[string]bool
	mu                     *sync.RWMutex
}

func NewFinishManager() *FinishSyncManager {
	finishedSyncValidators := make(map[byte]map[string]bool)
	for i := 0; i < common.MaxShardNumber; i++ {
		finishedSyncValidators[byte(i)] = make(map[string]bool)
	}
	return &FinishSyncManager{
		FinishedSyncValidators: finishedSyncValidators,
		mu:                     &sync.RWMutex{},
	}
}

func NewManagerWithValue(validators map[byte]map[string]bool) *FinishSyncManager {
	return &FinishSyncManager{
		FinishedSyncValidators: validators,
		mu:                     &sync.RWMutex{},
	}
}

func (manager *FinishSyncManager) Clone() FinishSyncManager {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	res := NewFinishManager()
	for i, validatorList := range manager.FinishedSyncValidators {
		for k, _ := range validatorList {
			res.FinishedSyncValidators[i][k] = true
		}
	}
	return *res
}

// AddFinishedSyncValidators only add FinishedSyncValidators in sync pool of attached view
// and NOT duplicate in FinishSyncManager.validator list
func (manager *FinishSyncManager) AddFinishedSyncValidators(
	newFinishedSyncValidators []string,
	syncPool []string,
	shardID byte,
) {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	finishedValidatorsToAdd := make(map[string]bool)
	// do not allow duplicate key
	for _, v := range newFinishedSyncValidators {
		if manager.FinishedSyncValidators[shardID][v] {
			continue
		}
		finishedValidatorsToAdd[v] = true
	}

	// finished sync FinishedSyncValidators must in sync pool
	count := 0
	for _, v := range syncPool {
		if count == len(finishedValidatorsToAdd) {
			break
		}
		if finishedValidatorsToAdd[v] {
			manager.FinishedSyncValidators[shardID][v] = true
			count++
		}
	}
}

func (manager *FinishSyncManager) Validators(shardID byte) []string {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	res := []string{}
	for k, _ := range manager.FinishedSyncValidators[shardID] {
		res = append(res, k)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}

//RemoveValidators only remove FinishSyncManager.validator list ONCE
// ignore FinishedSyncValidators not in FinishSyncManager.FinishedSyncValidators list
func (manager *FinishSyncManager) RemoveValidators(validators []string, shardID byte) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	for _, v := range validators {
		delete(manager.FinishedSyncValidators[shardID], v)
	}
}

//Instructions ....
func (manager *FinishSyncManager) Instructions() []*instruction.FinishSyncInstruction {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	res := []*instruction.FinishSyncInstruction{}
	keys := []int{}
	for i := 0; i < len(manager.FinishedSyncValidators); i++ {
		keys = append(keys, i)
	}
	sort.Ints(keys)
	for _, v := range keys {
		committeePublicKeys := manager.Validators(byte(v))
		if len(committeePublicKeys) == 0 {
			continue
		}
		finishSyncInstruction := instruction.NewFinishSyncInstructionWithValue(v, committeePublicKeys)
		res = append(res, finishSyncInstruction)
	}
	return res
}
