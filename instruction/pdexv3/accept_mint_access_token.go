package pdexv3

import "github.com/incognitochain/incognito-chain/common"

type AcceptMintAccessToken struct {
	burntAmount uint64
	otaReceiver string
	shardID     byte
	txReqID     common.Hash
}

func NewAcceptMintAccessToken() *AcceptMintAccessToken {
	return &AcceptMintAccessToken{}
}

func NewAcceptMintAccessTokenWithValue(
	burntAmount uint64,
	otaReceiver string,
	shardID byte,
	txReqID common.Hash,
) *AcceptMintAccessToken {
	return &AcceptMintAccessToken{
		burntAmount: burntAmount,
		otaReceiver: otaReceiver,
		shardID:     shardID,
		txReqID:     txReqID,
	}
}

func (a *AcceptMintAccessToken) StringSlice() ([]string, error) {
	res := []string{}
	return res, nil
}
