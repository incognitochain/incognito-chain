package lvdb

import (
	"github.com/constant-money/constant-chain/common"
	"github.com/pkg/errors"
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

func ParseKeyAddShardRewardRequest(key []byte) (
	epoch uint64,
	shardID byte,
	err error,
) {
	bytesArray, err := ParseKeyToSlice(key, []int{len(ShardRequestRewardPrefix), 4, 1})
	epoch = common.BytesToUint64(bytesArray[1])
	shardID = bytesArray[2][0]
	return
}

func ParseKeyAddCommitteeReward(key []byte) (
	committeeAddress []byte,
	err error,
) {
	bytesArray, err := ParseKeyToSlice(key, []int{len(ShardRequestRewardPrefix), 66})
	if err != nil {
		return nil, err
	}
	committeeAddress = bytesArray[1]
	return committeeAddress, err
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
