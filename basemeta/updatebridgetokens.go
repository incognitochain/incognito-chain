package basemeta

import "github.com/incognitochain/incognito-chain/common"

type UpdatingInfo struct {
	CountUpAmt      uint64
	DeductAmt       uint64
	TokenID         common.Hash
	ExternalTokenID []byte
	IsCentralized   bool
}