package types

import "github.com/incognitochain/incognito-chain/common"

type CrossShardTokenPrivacyMetaData struct {
	TokenID        common.Hash
	PropertyName   string
	PropertySymbol string
	Type           int    // action type
	Mintable       bool   // default false
	Amount         uint64 // init amount
}
