package finishsync

import (
	"sort"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/instruction"
)

var (
	DefaultFinishSyncMsgPool *FinishSyncMsgPool
)

// FinishSyncMsgPool manages FinishedSyncValidators in sync pool of attached view
// FinishedSyncValidators must only compatible with sync pool. It must only contain keys that also in sync pool
// FinishSyncMsgPool could maintain different data in different beacon nodes
type FinishSyncMsgPool struct {
	FinishedSyncValidators map[byte]map[string]bool `json:"FinishedSyncValidators"`
	mu                     *sync.RWMutex
}

func NewFinishSyncMsgPool() FinishSyncMsgPool {
	finishedSyncValidators := make(map[byte]map[string]bool)
	for i := 0; i < common.MaxShardNumber; i++ {
		finishedSyncValidators[byte(i)] = make(map[string]bool)
	}
	return FinishSyncMsgPool{
		FinishedSyncValidators: finishedSyncValidators,
		mu:                     &sync.RWMutex{},
	}
}

func NewDefaultFinishSyncMsgPool() {
	finishedSyncValidators := make(map[byte]map[string]bool)
	for i := 0; i < common.MaxShardNumber; i++ {
		finishedSyncValidators[byte(i)] = make(map[string]bool)
	}
	DefaultFinishSyncMsgPool = &FinishSyncMsgPool{
		FinishedSyncValidators: finishedSyncValidators,
		mu:                     &sync.RWMutex{},
	}
}

func NewFinishSyncMsgPoolWithValue(validators map[byte]map[string]bool) FinishSyncMsgPool {
	return FinishSyncMsgPool{
		FinishedSyncValidators: validators,
		mu:                     &sync.RWMutex{},
	}
}

func (f *FinishSyncMsgPool) Clone() FinishSyncMsgPool {

	f.mu.RLock()
	defer f.mu.RUnlock()

	res := NewFinishSyncMsgPool()
	for i, validatorList := range f.FinishedSyncValidators {
		for k, _ := range validatorList {
			res.FinishedSyncValidators[i][k] = true
		}
	}

	return res
}

func (f *FinishSyncMsgPool) GetFinishedSyncValidators() map[byte][]string {

	f.mu.RLock()
	defer f.mu.RUnlock()

	res := make(map[byte][]string)
	for shardID, finishedSyncValidators := range f.FinishedSyncValidators {
		for k, _ := range finishedSyncValidators {
			res[shardID] = append(res[shardID], k)
		}
	}

	return res
}

// AddFinishedSyncValidators only add FinishedSyncValidators in sync pool of attached view
// and NOT duplicate in FinishSyncMsgPool.validator list
func (f *FinishSyncMsgPool) AddFinishedSyncValidators(
	newFinishedSyncValidators []string,
	syncPool []string,
	shardID byte,
) {

	f.mu.Lock()
	defer f.mu.Unlock()

	finishedValidatorsToAdd := make(map[string]bool)
	// do not allow duplicate key
	for _, v := range newFinishedSyncValidators {
		if f.FinishedSyncValidators[shardID][v] {
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
			f.FinishedSyncValidators[shardID][v] = true
			count++
		}
	}
}

func (f *FinishSyncMsgPool) Validators(shardID byte) []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	res := []string{}
	for k, _ := range f.FinishedSyncValidators[shardID] {
		res = append(res, k)
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i] < res[j]
	})
	return res
}

//RemoveValidators only remove FinishSyncMsgPool.validator list ONCE
// ignore FinishedSyncValidators not in FinishSyncMsgPool.FinishedSyncValidators list
func (f *FinishSyncMsgPool) RemoveValidators(validators []string, shardID byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, v := range validators {
		delete(f.FinishedSyncValidators[shardID], v)
	}
}

//Instructions ....
func (f *FinishSyncMsgPool) Instructions(allSyncPool map[byte][]string) []*instruction.FinishSyncInstruction {

	f.mu.RLock()
	defer f.mu.RUnlock()

	res := []*instruction.FinishSyncInstruction{}
	keys := []int{}
	for i := 0; i < len(f.FinishedSyncValidators); i++ {
		keys = append(keys, i)
	}

	sort.Ints(keys)

	for _, v := range keys {
		syncPool, ok := allSyncPool[byte(v)]
		if !ok {
			continue
		}

		committeePublicKeys := []string{}
		finishSyncMsgValidators := f.FinishedSyncValidators[byte(v)]

		for _, validator := range syncPool {
			has := finishSyncMsgValidators[validator]
			if has {
				committeePublicKeys = append(committeePublicKeys, validator)
			}
		}
		if len(committeePublicKeys) == 0 {
			continue
		}
		finishSyncInstruction := instruction.NewFinishSyncInstructionWithValue(v, committeePublicKeys)
		res = append(res, finishSyncInstruction)
	}

	return res
}

func (f *FinishSyncMsgPool) Clean(allSyncPoolValidators map[byte][]string) {

	for {
		f.mu.Lock()
		f.clean(allSyncPoolValidators)
		f.mu.Unlock()
		time.Sleep(5 * time.Minute)
	}

}

func (f *FinishSyncMsgPool) clean(allSyncPoolValidators map[byte][]string) {
	for shardID, finishedSyncValidators := range f.FinishedSyncValidators {
		Logger.Infof("Finish Sync Msg Pool, ShardID %+v, Length %+v", shardID, len(finishedSyncValidators))
		syncPoolValidators, ok := allSyncPoolValidators[shardID]
		if !ok {
			f.FinishedSyncValidators[shardID] = make(map[string]bool)
		}
		for finishSyncMsg, _ := range finishedSyncValidators {
			has := false
			for _, syncPoolValidator := range syncPoolValidators {
				if syncPoolValidator == finishSyncMsg {
					has = true
					break
				}
			}
			if !has {
				delete(f.FinishedSyncValidators[shardID], finishSyncMsg)
			}
		}
	}
}
