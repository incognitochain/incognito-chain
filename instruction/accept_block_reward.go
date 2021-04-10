package instruction

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
)

//AcceptBlockReward store block reward for one subset of committees in shard
type AcceptBlockReward struct {
	subsetID         byte
	shardID          byte
	txsFee           map[common.Hash]uint64
	shardBlockHeight uint64
}

func NewAcceptBlockReward() *AcceptBlockReward {
	return &AcceptBlockReward{}
}

func NewAcceptBlockRewardWithValue(
	subsetID, shardID byte,
	txsFee map[common.Hash]uint64,
	shardBlockHeight uint64,
) *AcceptBlockReward {
	return &AcceptBlockReward{
		subsetID:         subsetID,
		shardID:          shardID,
		txsFee:           txsFee,
		shardBlockHeight: shardBlockHeight,
	}
}

func (a *AcceptBlockReward) SubsetID() byte {
	return a.subsetID
}

func (a *AcceptBlockReward) ShardID() byte {
	return a.shardID
}

//read only function
func (a *AcceptBlockReward) TxsFee() map[common.Hash]uint64 {
	return a.txsFee
}

func (a *AcceptBlockReward) ShardBlockHeight() uint64 {
	return a.shardBlockHeight
}

func (a *AcceptBlockReward) IsEmpty() bool {
	return reflect.DeepEqual(a, NewAcceptBlockReward())
}

func (a *AcceptBlockReward) GetType() string {
	return ACCEPT_BLOCK_REWARD_ACTION
}

func (a *AcceptBlockReward) StringArr() []string {
	acceptBlockRewardStr := []string{ACCEPT_BLOCK_REWARD_ACTION}
	acceptBlockRewardStr = append(acceptBlockRewardStr, strconv.Itoa(int(a.shardID)))
	acceptBlockRewardStr = append(acceptBlockRewardStr, strconv.Itoa(int(a.subsetID)))
	content, _ := json.Marshal(a.txsFee)
	acceptBlockRewardStr = append(acceptBlockRewardStr, string(content))
	acceptBlockRewardStr = append(acceptBlockRewardStr, strconv.FormatUint(a.shardBlockHeight, 10))
	return acceptBlockRewardStr
}

func ValidateAndImportAcceptBlockRewardInstructionFromString(instruction []string) (*AcceptBlockReward, error) {
	if err := ValidateAcceptBlockRewardInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportAcceptBlockRewardInstructionFromString(instruction)
}

func ImportAcceptBlockRewardInstructionFromString(instruction []string) (*AcceptBlockReward, error) {
	acceptBlockRewardIns := NewAcceptBlockReward()
	shardID, err := strconv.Atoi(instruction[1])
	if err != nil {
		return acceptBlockRewardIns, err
	}
	subsetID, err := strconv.Atoi(instruction[2])
	if err != nil {
		return acceptBlockRewardIns, err
	}
	acceptBlockRewardIns.shardID = byte(shardID)
	acceptBlockRewardIns.subsetID = byte(subsetID)

	txsFee := make(map[common.Hash]uint64)
	err = json.Unmarshal([]byte(instruction[3]), &txsFee)
	if err != nil {
		return acceptBlockRewardIns, err
	}
	acceptBlockRewardIns.txsFee = txsFee

	shardBlockHeight, err := strconv.ParseUint(instruction[4], 10, 64)
	if err != nil {
		return acceptBlockRewardIns, err
	}
	acceptBlockRewardIns.shardBlockHeight = shardBlockHeight

	return acceptBlockRewardIns, err
}

func ValidateAcceptBlockRewardInstructionSanity(instruction []string) error {
	if len(instruction) != 5 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != ACCEPT_BLOCK_REWARD_ACTION {
		return fmt.Errorf("invalid accept block reward action, %+v", instruction)
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

	txsFee := make(map[common.Hash]uint64)
	err = json.Unmarshal([]byte(instruction[3]), &txsFee)
	if err != nil {
		return err
	}

	_, err = strconv.ParseUint(instruction[4], 10, 64)
	if err != nil {
		return err
	}
	return nil
}
