package jsonresult

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type BridgeAggState struct {
	BeaconTimeStamp     int64                                                        `json:"BeaconTimeStamp"`
	UnifiedTokenVaults  map[common.Hash]map[common.Hash]*statedb.BridgeAggVaultState `json:"UnifiedTokenVaults"`
	WaitingUnshieldReqs map[common.Hash][]*statedb.BridgeAggWaitingUnshieldReq       `json:"WaitingUnshieldReqs"`
	Param               *statedb.BridgeAggParamState                                 `json:"Param"`
	BaseDecimal         uint8                                                        `json:"BaseDecimal"`
	MaxLenOfPath        uint8                                                        `json:"MaxLenOfPath"`
}

type BridgeAggEstimateFeeByBurntAmount struct {
	BurntAmount    uint64 `json:"BurntAmount"`
	Fee            uint64 `json:"Fee"`
	ReceivedAmount uint64 `json:"ReceivedAmount"`

	MaxFee            uint64 `json:"MaxFee"`
	MinReceivedAmount uint64 `json:"MinReceivedAmount"`
}

type BridgeAggEstimateFeeByReceivedAmount struct {
	ReceivedAmount uint64 `json:"ReceivedAmount"`
	Fee            uint64 `json:"Fee"`
	BurntAmount    uint64 `json:"BurntAmount"`

	MaxFee         uint64 `json:"MaxFee"`
	MaxBurntAmount uint64 `json:"MaxBurntAmount"`
}

type BridgeAggEstimateReward struct {
	ReceivedAmount uint64 `json:"ReceivedAmount"`
	Reward         uint64 `json:"Reward"`
}
