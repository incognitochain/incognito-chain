package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

type Pdexv3State struct {
	BeaconTimeStamp      int64                                  `json:"BeaconTimeStamp"`
	Params               pdex.Params                            `json:"Params"`
	PoolPairs            map[string]*pdex.PoolPairState         `json:"PoolPairs"`
	WaitingContributions map[string]*rawdbv2.Pdexv3Contribution `json:"WaitingContributions"`
	NftIDs               map[string]uint64                      `json:"NftIDs"`
	StakingPools         map[string]*pdex.StakingPoolState      `json:"StakingPools"`
}
