package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

type Pdexv3State struct {
	BeaconTimeStamp      int64                                   `json:"BeaconTimeStamp"`
	Params               *pdex.Params                            `json:"Params,omitempty"`
	PoolPairs            *map[string]*pdex.PoolPairState         `json:"PoolPairs,omitempty"`
	WaitingContributions *map[string]*rawdbv2.Pdexv3Contribution `json:"WaitingContributions,omitempty"`
	NftIDs               *map[string]uint64                      `json:"NftIDs,omitempty"`
	StakingPools         *map[string]*pdex.StakingPoolState      `json:"StakingPools,omitempty"`
}

type Pdexv3LPValue struct {
	PoolValue   map[string]uint64 `json:"PoolValue"`
	LPReward    map[string]uint64 `json:"LPReward"`
	OrderReward map[string]uint64 `json:"OrderReward"`
	PoolReward  map[string]uint64 `json:"PoolReward"`
}
