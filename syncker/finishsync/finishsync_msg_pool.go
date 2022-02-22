package finishsync

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/instruction"
	"sort"
	"sync"
)

var (
	DefaultFinishSyncMsgPool *FinishSyncMsgPool
)

// FinishSyncMsgPool manages FinishedSyncValidators in sync pool of attached view
// FinishedSyncValidators must only compatible with sync pool. It must only contain keys that also in sync pool
// FinishSyncMsgPool could maintain different data in different beacon nodes
type FinishSyncMsgPool struct {
	FinishedSyncValidators map[byte]map[string]bool `json:"FinishedSyncValidators"`
	ReceiveTime            map[string]uint64        //beacon block height that receive the msg
	mu                     *sync.RWMutex
}

func NewFinishSyncMsgPool() FinishSyncMsgPool {
	finishedSyncValidators := make(map[byte]map[string]bool)
	for i := 0; i < common.MaxShardNumber; i++ {
		finishedSyncValidators[byte(i)] = make(map[string]bool)
	}
	return FinishSyncMsgPool{
		FinishedSyncValidators: finishedSyncValidators,
		ReceiveTime:            make(map[string]uint64),
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
		ReceiveTime:            make(map[string]uint64),
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
		for k := range validatorList {
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
		for k := range finishedSyncValidators {
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
	beaconHeight uint64,
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
			f.ReceiveTime[v] = beaconHeight
			count++
		}
	}
}

func (f *FinishSyncMsgPool) Validators(shardID byte) []string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	res := []string{}
	for k := range f.FinishedSyncValidators[shardID] {
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
		delete(f.ReceiveTime, v)
	}
}

//Instructions ....
func (f *FinishSyncMsgPool) Instructions(allSyncPool map[byte][]string, currentBeaconHeight uint64) []*instruction.FinishSyncInstruction {

	f.mu.RLock()
	defer f.mu.RUnlock()
	//fmt.Println("debug 1", allSyncPool, currentBeaconHeight, f.FinishedSyncValidators, f.ReceiveTime)
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
			receiveTime := f.ReceiveTime[validator]
			//receive msg should be from 1 epoch ago (prevent random number control)
			//fmt.Println("debug 2", validator, has, receiveTime, currentBeaconHeight-config.Param().EpochParam.NumberOfBlockInEpoch)
			if has && receiveTime < currentBeaconHeight-config.Param().EpochParam.NumberOfBlockInEpoch {
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
	f.clean(allSyncPoolValidators)
}

func (f *FinishSyncMsgPool) clean(allSyncPoolValidators map[byte][]string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for shardID, finishedSyncValidators := range f.FinishedSyncValidators {
		Logger.Infof("Finish Sync Msg Pool, ShardID %+v, Length %+v", shardID, len(finishedSyncValidators))
		syncPoolValidators, _ := allSyncPoolValidators[shardID]
		for validator := range finishedSyncValidators {
			has := false
			for _, syncPoolValidator := range syncPoolValidators {
				if syncPoolValidator == validator {
					has = true
					break
				}
			}
			if !has {
				//fmt.Println("debug detete", validator)
				//Logger.Info(errors.New("debug detete"))
				delete(f.FinishedSyncValidators[shardID], validator)
				delete(f.ReceiveTime, validator)
			}
		}
	}
}
