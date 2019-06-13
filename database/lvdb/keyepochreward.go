package lvdb

import (
	"github.com/incognitochain/incognito-chain/common"
)

func NewKeyAddShardRewardRequest(
	epoch uint64,
	shardID byte,
) ([]byte, error) {
	res := []byte{}
	res = append(res, ShardRequestRewardPrefix...)
	res = append(res, common.Uint64ToBytes(epoch)...)
	res = append(res, shardID)
	return res, nil
}

func NewKeyAddCommitteeReward(
	committeeAddress []byte,
) ([]byte, error) {
	res := []byte{}
	res = append(res, CommitteeRewardPrefix...)
	res = append(res, committeeAddress...)
	return res, nil
}
