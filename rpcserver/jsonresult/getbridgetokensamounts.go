package jsonresult

import "github.com/constant-money/constant-chain/common"

type GetBridgeTokensAmounts struct {
	BridgeTokensAmounts map[string]GetBridgeTokensAmount // key is currency type
}

type GetBridgeTokensAmount struct {
	TokenID *common.Hash `json:"TokenId"`
	Amount  uint64       `json:"Amount"`
}
