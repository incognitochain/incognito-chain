package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/pkg/errors"
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

func ParseKeyAddShardRewardRequest(key []byte) (
	epoch uint64,
	shardID byte,
	tokenID common.Hash,
	err error,
) {
	bytesArray, err := ParseKeyToSlice(key, []int{len(ShardRequestRewardPrefix), 4, 1, common.HashSize})
	if err != nil {
		return epoch, shardID, tokenID, err
	}
	epoch = common.BytesToUint64(bytesArray[1])
	shardID = bytesArray[2][0]
	tmpTokenID, err := common.NewHash(bytesArray[3])
	tokenID = *tmpTokenID
	return epoch, shardID, tokenID, err
}

func ParseKeyAddCommitteeReward(key []byte) (
	committeeAddress []byte,
	tokenID common.Hash,
	err error,
) {
	bytesArray, err := ParseKeyToSlice(key, []int{len(ShardRequestRewardPrefix), 66, common.HashSize})
	if err != nil {
		return nil, tokenID, err
	}
	committeeAddress = bytesArray[1]
	tmpTokenID, err := common.NewHash(bytesArray[2])
	tokenID = *tmpTokenID
	return committeeAddress, tokenID, err
}

func ParseKeyToSlice(key []byte, length []int) ([][]byte, error) {
	pos := GetPosFromLength(length)
	if pos[len(pos)-1] != len(key) {
		return nil, errors.New("key and length of args not match")
	}
	res := make([][]byte, 0)
	for i := 0; i < len(pos)-1; i++ {
		res = append(res, key[pos[i]:pos[i+1]])
	}
	return res, nil
}

func GetPosFromLength(length []int) []int {
	pos := []int{0}
	for i := 0; i < len(length); i++ {
		pos = append(pos, pos[i]+length[i])
	}
	return pos
}
