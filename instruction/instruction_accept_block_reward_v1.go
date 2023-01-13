package instruction

import (
	"encoding/json"
	"strconv"

	"github.com/incognitochain/incognito-chain/metadata"

	// "errors"

	"github.com/incognitochain/incognito-chain/common"
)

// ShardBlockRewardInfo IS LEGACY CODE, DO NOT MAKE ANY MODIFICATION
type ShardBlockRewardInfo struct {
	ShardReward map[common.Hash]uint64
	Epoch       uint64
}

type AcceptBlockRewardV1 struct {
	ShardID          byte
	TxsFee           map[common.Hash]uint64
	ShardBlockHeight uint64
}

func NewShardReceiveRewardV1WithValue(reward map[common.Hash]uint64, epoch uint64, shardID byte) ([][]string, error) {
	resIns := [][]string{}
	shardBlockRewardInfo := ShardBlockRewardInfo{
		Epoch:       epoch,
		ShardReward: reward,
	}

	contentStr, err := json.Marshal(shardBlockRewardInfo)
	if err != nil {
		return nil, err
	}

	returnedInst := []string{
		strconv.Itoa(SHARD_RECEIVE_REWARD_V1_ACTION),
		strconv.Itoa(int(shardID)),
		SHARD_REWARD_INST,
		string(contentStr),
	}
	resIns = append(resIns, returnedInst)
	return resIns, nil
}

func NewShardReceiveRewardV1FromString(inst string) (*ShardBlockRewardInfo, error) {
	Ins := &ShardBlockRewardInfo{}
	err := json.Unmarshal([]byte(inst), Ins)
	if err != nil {
		return nil, err
	}
	return Ins, nil
}

func NewAcceptBlockRewardV1WithValue(
	shardID byte,
	txsFee map[common.Hash]uint64,
	shardBlockHeight uint64,
) *AcceptBlockRewardV1 {
	return &AcceptBlockRewardV1{
		ShardID:          shardID,
		TxsFee:           txsFee,
		ShardBlockHeight: shardBlockHeight,
	}
}

func NewAcceptedBlockRewardV1FromString(
	inst string,
) (*AcceptBlockRewardV1, error) {
	Ins := &AcceptBlockRewardV1{}
	err := json.Unmarshal([]byte(inst), Ins)
	if err != nil {
		return nil, err
	}
	return Ins, nil
}

func (blockRewardInfo *AcceptBlockRewardV1) String() ([]string, error) {
	content, err := json.Marshal(blockRewardInfo)
	if err != nil {
		return nil, err
	}
	return []string{
		strconv.Itoa(ACCEPT_BLOCK_REWARD_V1_ACTION),
		strconv.Itoa(metadata.BeaconOnly),
		string(content),
	}, nil
}
