package instruction

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/log/proto"
)

type ShardReceiveRewardV3 struct {
	reward   map[common.Hash]uint64
	epoch    uint64
	shardID  byte
	subsetID byte
	instructionBase
}

func NewShardReceiveRewardV3() *ShardReceiveRewardV3 {
	return &ShardReceiveRewardV3{
		instructionBase: instructionBase{
			featureID: proto.FID_CONSENSUS_SHARD,
			logOnly:   false,
		},
	}
}

func NewShardReceiveRewardV3WithValue(
	reward map[common.Hash]uint64,
	epoch uint64, shardID, subsetID byte,
) *ShardReceiveRewardV3 {
	return &ShardReceiveRewardV3{
		reward:   reward,
		epoch:    epoch,
		shardID:  shardID,
		subsetID: subsetID,
		instructionBase: instructionBase{
			featureID: proto.FID_CONSENSUS_SHARD,
			logOnly:   false,
		},
	}
}

func (shardSubsetReward *ShardReceiveRewardV3) SubsetID() byte {
	return shardSubsetReward.subsetID
}

func (shardSubsetReward *ShardReceiveRewardV3) SetSubsetID(subsetID byte) *ShardReceiveRewardV3 {
	shardSubsetReward.subsetID = subsetID
	return shardSubsetReward
}

func (shardSubsetReward *ShardReceiveRewardV3) ShardID() byte {
	return shardSubsetReward.shardID
}

// read only function
func (shardSubsetReward *ShardReceiveRewardV3) Reward() map[common.Hash]uint64 {
	return shardSubsetReward.reward
}

func (shardSubsetReward *ShardReceiveRewardV3) Epoch() uint64 {
	return shardSubsetReward.epoch
}

func (shardSubsetReward *ShardReceiveRewardV3) IsEmpty() bool {
	return reflect.DeepEqual(shardSubsetReward, NewShardReceiveRewardV3())
}

func (shardSubsetReward *ShardReceiveRewardV3) GetType() string {
	return SHARD_RECEIVE_REWARD_V3_ACTION
}

func (shardSubsetReward *ShardReceiveRewardV3) String() []string {
	shardSubsetRewardStr := []string{SHARD_RECEIVE_REWARD_V3_ACTION}
	shardSubsetRewardStr = append(shardSubsetRewardStr, strconv.Itoa(int(shardSubsetReward.shardID)))
	shardSubsetRewardStr = append(shardSubsetRewardStr, strconv.Itoa(int(shardSubsetReward.subsetID)))
	content, _ := json.Marshal(shardSubsetReward.reward)
	shardSubsetRewardStr = append(shardSubsetRewardStr, string(content))
	shardSubsetRewardStr = append(shardSubsetRewardStr, strconv.FormatUint(shardSubsetReward.epoch, 10))
	return shardSubsetRewardStr
}

func ValidateAndImportShardReceiveRewardV3InstructionFromString(instruction []string) (*ShardReceiveRewardV3, error) {
	if err := ValidateShardReceiveRewardV3InstructionFromString(instruction); err != nil {
		return nil, err
	}
	return ImportShardReceiveRewardV3InstructionFromString(instruction)
}

func ImportShardReceiveRewardV3InstructionFromString(instruction []string) (*ShardReceiveRewardV3, error) {
	shardSubsetReward := NewShardReceiveRewardV3()
	shardID, err := strconv.Atoi(instruction[1])
	if err != nil {
		return shardSubsetReward, err
	}
	subsetID, err := strconv.Atoi(instruction[2])
	if err != nil {
		return shardSubsetReward, err
	}
	shardSubsetReward.shardID = byte(shardID)
	shardSubsetReward.subsetID = byte(subsetID)

	reward := make(map[common.Hash]uint64)
	err = json.Unmarshal([]byte(instruction[3]), &reward)
	if err != nil {
		return shardSubsetReward, err
	}
	shardSubsetReward.reward = reward

	epoch, err := strconv.ParseUint(instruction[4], 10, 64)
	if err != nil {
		return shardSubsetReward, err
	}
	shardSubsetReward.epoch = epoch

	return shardSubsetReward, err
}

func ValidateShardReceiveRewardV3InstructionFromString(instruction []string) error {
	if len(instruction) != 5 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != SHARD_RECEIVE_REWARD_V3_ACTION {
		return fmt.Errorf("invalid shard subset reward action, %+v", instruction)
	}
	shardID, err := strconv.Atoi(instruction[1])
	if err != nil {
		return err
	}

	if shardID < 0 || shardID > 9 {
		return errors.New("shardID is out of range for byte")
	}

	subsetID, err := strconv.Atoi(instruction[2])
	if err != nil {
		return err
	}

	if subsetID < 0 || subsetID > 2 {
		return errors.New("subsetID is out of range for byte")
	}

	reward := make(map[common.Hash]uint64)
	err = json.Unmarshal([]byte(instruction[3]), &reward)
	if err != nil {
		return err
	}

	_, err = strconv.ParseUint(instruction[4], 10, 64)
	if err != nil {
		return err
	}
	return nil
}
