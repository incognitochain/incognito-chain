package types

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
)

type CrossOutputCoin struct {
	BlockHeight uint64
	BlockHash   common.Hash
	OutputCoin  []privacy.Coin
}

func (crossOutputCoin CrossOutputCoin) Hash() common.Hash {
	res := []byte{}
	res = append(res, crossOutputCoin.BlockHash.GetBytes()...)
	for _, coins := range crossOutputCoin.OutputCoin {
		res = append(res, coins.Bytes()...)
	}
	return common.HashH(res)
}
