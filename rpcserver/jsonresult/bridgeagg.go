package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAggState struct {
	BeaconTimeStamp   int64                                     `json:"BeaconTimeStamp"`
	UnifiedTokenInfos map[common.Hash]map[uint]*bridgeagg.Vault `json:"UnifiedTokenInfos"`
	BaseDecimal       uint                                      `json:"BaseDecimal"`
	MaxLenOfPath      int                                       `json:"MaxLenOfPath"`
}

type BridgeAggEstimateReceivedAmount struct {
	ReceivedAmount uint64 `json:"ReceivedAmount"`
	Fee            uint64 `json:"Fee"`
}
