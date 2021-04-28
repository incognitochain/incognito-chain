package finishsync

import (
	"github.com/incognitochain/incognito-chain/common"
	"sort"
	"sync"

	"github.com/incognitochain/incognito-chain/instruction"
)

// FinishSyncManager manages finishedSyncValidators in sync pool of attached view
//  already sent finish sync message to this node
// FinishSyncManager could maintain different data in different beacon nodes
type FinishSyncManager struct {
	finishedSyncValidators map[byte]map[string]bool
	mu                     *sync.RWMutex
}

func NewFinishManager() *FinishSyncManager {
	finishedSyncValidators := make(map[byte]map[string]bool)
	for i := 0; i < common.MaxShardNumber; i++ {
		finishedSyncValidators[byte(i)] = make(map[string]bool)
	}
	return &FinishSyncManager{
		finishedSyncValidators: finishedSyncValidators,
		mu:                     &sync.RWMutex{},
	}
}

func NewManagerWithValue(validators map[byte]map[string]bool) *FinishSyncManager {
	return &FinishSyncManager{
		finishedSyncValidators: validators,
		mu:                     &sync.RWMutex{},
	}
}

func (manager *FinishSyncManager) Clone() *FinishSyncManager {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	res := NewFinishManager()
	for i, validatorList := range manager.finishedSyncValidators {
		for k, _ := range validatorList {
			res.finishedSyncValidators[i][k] = true
		}
	}
	return res
}

// AddFinishedSyncValidators only add finishedSyncValidators in sync pool of attached view
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
		if manager.finishedSyncValidators[shardID][v] {
			continue
		}
		finishedValidatorsToAdd[v] = true
	}

	// finished sync finishedSyncValidators must in sync pool
	count := 0
	for _, v := range syncPool {
		if count == len(finishedValidatorsToAdd) {
			break
		}
		if finishedValidatorsToAdd[v] {
			manager.finishedSyncValidators[shardID][v] = true
			count++
		}
	}
}

func (manager *FinishSyncManager) Validators(shardID byte) []string {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	res := []string{}
	for k, _ := range manager.finishedSyncValidators[shardID] {
		res = append(res, k)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}

//RemoveValidators only remove FinishSyncManager.validator list ONCE
// ignore finishedSyncValidators not in FinishSyncManager.finishedSyncValidators list
func (manager *FinishSyncManager) RemoveValidators(validators []string, shardID byte) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	for _, v := range validators {
		delete(manager.finishedSyncValidators[shardID], v)
	}
}

//Instructions ....
func (manager *FinishSyncManager) Instructions() []*instruction.FinishSyncInstruction {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	res := []*instruction.FinishSyncInstruction{}
	keys := []int{}
	for i := 0; i < len(manager.finishedSyncValidators); i++ {
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
