package instruction

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
)

//AcceptBlockRewardV4 store block reward for one subset of committees in shard
type AcceptBlockRewardV4 struct {
	subsetID         byte
	shardID          byte
	txsFee           map[common.Hash]uint64
	shardBlockHeight uint64
	blockVersion     int
}

func NewAcceptBlockRewardV4() *AcceptBlockRewardV4 {
	return &AcceptBlockRewardV4{}
}

func NewAcceptBlockRewardV4WithValue(
	subsetID, shardID byte,
	txsFee map[common.Hash]uint64,
	shardBlockHeight uint64,
	blockVersion int,
) *AcceptBlockRewardV4 {
	return &AcceptBlockRewardV4{
		subsetID:         subsetID,
		shardID:          shardID,
		txsFee:           txsFee,
		shardBlockHeight: shardBlockHeight,
		blockVersion:     blockVersion,
	}
}

func (a *AcceptBlockRewardV4) SubsetID() byte {
	return a.subsetID
}

func (a *AcceptBlockRewardV4) ShardID() byte {
	return a.shardID
}

//read only function
func (a *AcceptBlockRewardV4) TxsFee() map[common.Hash]uint64 {
	return a.txsFee
}

func (a *AcceptBlockRewardV4) ShardBlockHeight() uint64 {
	return a.shardBlockHeight
}

func (a *AcceptBlockRewardV4) IsEmpty() bool {
	return reflect.DeepEqual(a, NewAcceptBlockRewardV4())
}

func (a *AcceptBlockRewardV4) GetType() string {
	return ACCEPT_BLOCK_REWARD_V4_ACTION
}

func (a *AcceptBlockRewardV4) String() []string {
	acceptBlockRewardStr := []string{ACCEPT_BLOCK_REWARD_V3_ACTION}
	acceptBlockRewardStr = append(acceptBlockRewardStr, strconv.Itoa(int(a.shardID)))
	acceptBlockRewardStr = append(acceptBlockRewardStr, strconv.Itoa(int(a.subsetID)))
	content, _ := json.Marshal(a.txsFee)
	acceptBlockRewardStr = append(acceptBlockRewardStr, string(content))
	acceptBlockRewardStr = append(acceptBlockRewardStr, strconv.FormatUint(a.shardBlockHeight, 10))
	return acceptBlockRewardStr
}

func ValidateAndImportAcceptBlockRewardV4InstructionFromString(instruction []string) (*AcceptBlockRewardV4, error) {
	if err := ValidateAcceptBlockRewardInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportAcceptBlockRewardV4InstructionFromString(instruction)
}

func ImportAcceptBlockRewardV4InstructionFromString(instruction []string) (*AcceptBlockRewardV4, error) {
	acceptBlockRewardIns := NewAcceptBlockRewardV4()
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

func ValidateAcceptBlockRewardV4InstructionSanity(instruction []string) error {
	if len(instruction) != 5 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != ACCEPT_BLOCK_REWARD_V3_ACTION {
		return fmt.Errorf("invalid accept block reward action, %+v", instruction)
	}
	shardID, err := strconv.Atoi(instruction[1])
	if err != nil {
		return err
	}

	if shardID < 0 || shardID > common.MaxShardNumber {
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
