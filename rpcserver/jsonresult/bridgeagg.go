package jsonresult

import (
	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/common"
)

type BridgeAggState struct {
	BeaconTimeStamp   int64                                            `json:"BeaconTimeStamp"`
	UnifiedTokenInfos map[common.Hash]map[common.Hash]*bridgeagg.Vault `json:"UnifiedTokenInfos"`
}
