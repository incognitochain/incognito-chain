package v2utils

import (
	"math/big"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
)

type StateChange struct {
	PoolPairs    map[string]*PoolPairChange
	StakingPools map[string]*StakingPoolChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		PoolPairs:    make(map[string]*PoolPairChange),
		StakingPools: make(map[string]*StakingPoolChange),
	}
}

type StakingPoolChange struct {
	Stakers         map[string]*StakerChange
	RewardsPerShare map[string]bool
}

type StakerChange struct {
	IsChanged           bool
	Rewards             map[string]bool
	LastRewardsPerShare map[string]bool
}

func NewStakingChange() *StakingPoolChange {
	return &StakingPoolChange{
		Stakers:         make(map[string]*StakerChange),
		RewardsPerShare: make(map[string]bool),
	}
}

func NewStakerChange() *StakerChange {
	return &StakerChange{
		Rewards:             make(map[string]bool),
		LastRewardsPerShare: make(map[string]bool),
	}
}

type PoolPairChange struct {
	IsChanged         bool
	Shares            map[string]*ShareChange
	OrderIDs          map[string]bool
	LpFeesPerShare    map[string]bool
	ProtocolFees      map[string]bool
	StakingPoolFees   map[string]bool
	MakingVolume      map[string]*MakingVolumeChange
	OrderRewards      map[string]*OrderRewardChange
	LmRewardsPerShare map[string]bool
	LmLockedShare     map[string]map[uint64]bool
}

func NewPoolPairChange() *PoolPairChange {
	return &PoolPairChange{
		Shares:            make(map[string]*ShareChange),
		OrderIDs:          make(map[string]bool),
		LpFeesPerShare:    make(map[string]bool),
		LmRewardsPerShare: make(map[string]bool),
		ProtocolFees:      make(map[string]bool),
		StakingPoolFees:   make(map[string]bool),
		MakingVolume:      make(map[string]*MakingVolumeChange),
		OrderRewards:      make(map[string]*OrderRewardChange),
		LmLockedShare:     make(map[string]map[uint64]bool),
	}
}

type ShareChange struct {
	IsChanged             bool
	TradingFees           map[string]bool
	LastLPFeesPerShare    map[string]bool
	LastLmRewardsPerShare map[string]bool
}

func NewShareChange() *ShareChange {
	return &ShareChange{
		TradingFees:           make(map[string]bool),
		LastLPFeesPerShare:    make(map[string]bool),
		LastLmRewardsPerShare: make(map[string]bool),
	}
}

type MakingVolumeChange struct {
	Volume map[string]bool
}

func NewMakingVolumeChange() *MakingVolumeChange {
	return &MakingVolumeChange{
		Volume: make(map[string]bool),
	}
}

type OrderRewardChange struct {
	UncollectedReward map[string]bool
}

func NewOrderRewardChange() *OrderRewardChange {
	return &OrderRewardChange{
		UncollectedReward: make(map[string]bool),
	}
}

func GetChangedElementsFromMapUint64(map0, map1 map[common.Hash]uint64) map[string]bool {
	res := map[string]bool{}
	for k, v := range map0 {
		if m, ok := map1[k]; !ok || !reflect.DeepEqual(m, v) {
			res[k.String()] = true
		}
	}

	for k, v := range map1 {
		if m, ok := map0[k]; !ok || !reflect.DeepEqual(m, v) {
			res[k.String()] = true
		}
	}
	return res
}

func GetChangedElementsFromMapBigInt(map0, map1 map[common.Hash]*big.Int) map[string]bool {
	res := map[string]bool{}
	for k, v := range map0 {
		if m, ok := map1[k]; !ok || !reflect.DeepEqual(m, v) {
			res[k.String()] = true
		}
	}

	for k, v := range map1 {
		if m, ok := map0[k]; !ok || !reflect.DeepEqual(m, v) {
			res[k.String()] = true
		}
	}
	return res
}

func GetChangedElementsFromMapStringBigInt(map0, map1 map[string]*big.Int) map[string]bool {
	res := map[string]bool{}
	for k, v := range map0 {
		if m, ok := map1[k]; !ok || !reflect.DeepEqual(m, v) {
			res[k] = true
		}
	}

	for k, v := range map1 {
		if m, ok := map0[k]; !ok || !reflect.DeepEqual(m, v) {
			res[k] = true
		}
	}
	return res
}

func GetChangedElementsFromMapStringMapUint64(map0, map1 map[string]map[uint64]uint64) map[string]map[uint64]bool {
	res := map[string]map[uint64]bool{}
	for k, v := range map0 {
		res[k] = getChangedElementsFromMapUint64Uint64(v, map1[k])
	}

	for k, v := range map1 {
		res[k] = getChangedElementsFromMapUint64Uint64(v, map0[k])
	}
	return res
}

func getChangedElementsFromMapUint64Uint64(map0, map1 map[uint64]uint64) map[uint64]bool {
	res := map[uint64]bool{}
	if map1 == nil && map0 != nil {
		for k := range map0 {
			res[k] = true
		}
		return res
	}
	if map0 == nil && map1 != nil {
		for k := range map1 {
			res[k] = true
		}
		return res
	}
	for k, v := range map0 {
		if m, ok := map1[k]; !ok || !reflect.DeepEqual(m, v) {
			res[k] = true
		}
	}

	for k, v := range map1 {
		if m, ok := map0[k]; !ok || !reflect.DeepEqual(m, v) {
			res[k] = true
		}
	}
	return res
}
