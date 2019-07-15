package lvdb

import (
	"github.com/incognitochain/incognito-chain/common"
)

/**
 * NewKeyAddShardRewardRequest create a key for store reward of a shard X at epoch T in db.
 * @param epoch: epoch T
 * @param shardID: shard X
 * @param tokenID: currency unit
 * @return ([]byte, error): Key, error of this process
 */
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

/**
 * NewKeyAddCommitteeReward create a key for store reward of a person P in committee in db.
 * @param committeeAddress: Public key of person P
 * @param tokenID: currency unit
 * @return ([]byte, error): Key, error of this process
 */
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
