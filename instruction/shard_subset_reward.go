package instruction

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
)

type ShardSubsetReward struct {
	reward   map[common.Hash]uint64
	epoch    uint64
	shardID  byte
	subsetID byte
}

func NewShardSubsetReward() *ShardSubsetReward {
	return &ShardSubsetReward{}
}

func NewShardSubsetRewardWithValue(
	reward map[common.Hash]uint64,
	epoch uint64, shardID, subsetID byte,
) *ShardSubsetReward {
	return &ShardSubsetReward{
		reward:   reward,
		epoch:    epoch,
		shardID:  shardID,
		subsetID: subsetID,
	}
}

func (shardSubsetReward *ShardSubsetReward) SubsetID() byte {
	return shardSubsetReward.subsetID
}

func (shardSubsetReward *ShardSubsetReward) ShardID() byte {
	return shardSubsetReward.shardID
}

//read only function
func (shardSubsetReward *ShardSubsetReward) Reward() map[common.Hash]uint64 {
	return shardSubsetReward.reward
}

func (shardSubsetReward *ShardSubsetReward) Epoch() uint64 {
	return shardSubsetReward.epoch
}

func (shardSubsetReward *ShardSubsetReward) IsEmpty() bool {
	return reflect.DeepEqual(shardSubsetReward, NewShardSubsetReward())
}

func (shardSubsetReward *ShardSubsetReward) GetType() string {
	return SHARD_SUBSET_REWARD_ACTION
}

func (shardSubsetReward *ShardSubsetReward) StringArr() []string {
	shardSubsetRewardStr := []string{SHARD_SUBSET_REWARD_ACTION}
	shardSubsetRewardStr = append(shardSubsetRewardStr, strconv.Itoa(int(shardSubsetReward.shardID)))
	shardSubsetRewardStr = append(shardSubsetRewardStr, strconv.Itoa(int(shardSubsetReward.subsetID)))
	content, _ := json.Marshal(shardSubsetReward.reward)
	shardSubsetRewardStr = append(shardSubsetRewardStr, string(content))
	shardSubsetRewardStr = append(shardSubsetRewardStr, strconv.FormatUint(shardSubsetReward.epoch, 10))
	return shardSubsetRewardStr
}

func ValidateAndImportShardSubsetRewardInstructionFromString(instruction []string) (*ShardSubsetReward, error) {
	if err := ValidateShardSubsetRewardInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportShardSubsetRewardInstructionFromString(instruction)
}

func ImportShardSubsetRewardInstructionFromString(instruction []string) (*ShardSubsetReward, error) {
	shardSubsetReward := NewShardSubsetReward()
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

func ValidateShardSubsetRewardInstructionSanity(instruction []string) error {
	if len(instruction) != 5 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != SHARD_SUBSET_REWARD_ACTION {
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
