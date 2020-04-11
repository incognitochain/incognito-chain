package metadata

import (
	"github.com/incognitochain/incognito-chain/common"
)

type ResetFeatureRewardRequestContent struct {
	TokenID common.Hash
	TxReqID common.Hash
	ShardID byte
}

func NewResetFeatureRewardRequestContent(tokenID common.Hash, TxReqID common.Hash, shardID byte) (*ResetFeatureRewardRequestContent, error) {
	return &ResetFeatureRewardRequestContent{
		TokenID: tokenID,
		TxReqID: TxReqID,
		ShardID: shardID,
	}, nil
}
