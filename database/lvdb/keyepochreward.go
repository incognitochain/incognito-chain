package lvdb

import (
	"github.com/incognitochain/incognito-chain/common"
)

func NewKeyAddShardRewardRequest(
	epoch uint64,
	shardID byte,
	tokenID common.Hash,
) ([]byte, error) {
	res := []byte{}
	res = append(res, ShardRequestRewardPrefix...)
	res = append(res, common.Uint64ToBytes(epoch)...)
	res = append(res, shardID)
	res = append(res, tokenID.GetBytes()...)
	return res, nil
}

func NewKeyAddCommitteeReward(
	committeeAddress []byte,
	tokenID common.Hash,
) ([]byte, error) {
	res := []byte{}
	res = append(res, CommitteeRewardPrefix...)
	res = append(res, committeeAddress...)
	res = append(res, tokenID.GetBytes()...)
	return res, nil
}
