package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain/bridgehub"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type BridgeHubState struct {
	BeaconTimeStamp int64                                      `json:"BeaconTimeStamp"`
	StakingInfos    map[string]*statedb.BridgeStakingInfoState `json:"StakingInfos"`
	BridgeInfos     map[string]*bridgehub.BridgeInfo           `json:"BridgeInfos"`
	TokenPrices     map[string]uint64                          `json:"TokenPrices"`
	Params          *statedb.BridgeHubParamState               `json:"Params"`
}

// type BridgeInfo struct {
// 	Info          *statedb.BridgeInfoState                 `json:"Info"`
// 	PTokenAmounts map[string]*statedb.BridgeHubPTokenState `json:"PTokenAmounts"`
// }
