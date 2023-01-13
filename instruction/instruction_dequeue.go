package instruction

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/log/proto"
)

var (
	ErrDequeueInstruction = errors.New("dequeue instruction error")
)

type DequeueInstruction struct {
	Reason      string
	DequeueList map[int][]int //shardID -> pending index
	instructionBase
}

func NewDequeueInstructionWithValue(reason string, dequeueList map[int][]int) *DequeueInstruction {
	dequeueInstruction := &DequeueInstruction{
		reason,
		dequeueList,
		instructionBase{
			featureID: proto.FID_CONSENSUS_SHARD,
			logOnly:   false,
		},
	}
	return dequeueInstruction
}

func NewDequeueInstruction() *DequeueInstruction {
	return &DequeueInstruction{
		instructionBase: instructionBase{
			featureID: proto.FID_CONSENSUS_SHARD,
			logOnly:   false,
		},
	}
}

func (f *DequeueInstruction) GetType() string {
	return DEQUEUE
}

func (f *DequeueInstruction) IsEmpty() bool {
	return len(f.DequeueList) == 0
}

func (f *DequeueInstruction) ToString() []string {
	dequeueInstruction := []string{DEQUEUE}
	dequeueInstruction = append(dequeueInstruction, f.Reason)
	b, _ := json.Marshal(f.DequeueList)
	dequeueInstruction = append(dequeueInstruction, string(b))
	return dequeueInstruction
}

func ValidateAndImportDequeueInstructionFromString(instruction []string) (*DequeueInstruction, error) {
	if err := ValidateDequeueInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportDequeueInstructionFromString(instruction)
}

// ImportFeatureEnableInstructionFromString is unsafe method
func ImportDequeueInstructionFromString(instruction []string) (*DequeueInstruction, error) {
	dequeueInstruction := NewDequeueInstruction()
	dequeueInstruction.Reason = instruction[1]
	if err := json.Unmarshal([]byte(instruction[2]), &dequeueInstruction.DequeueList); err != nil {
		return nil, err
	}
	return dequeueInstruction, nil
}

// ValidateFeatureEnableInstructionSanity ...
func ValidateDequeueInstructionSanity(instruction []string) error {
	if len(instruction) != 3 {
		return fmt.Errorf("%+v: invalid length, %+v", ErrDequeueInstruction, instruction)
	}
	if instruction[0] != DEQUEUE {
		return fmt.Errorf("%+v: invalid dequeue action, %+v", ErrDequeueInstruction, instruction)
	}

	if len(instruction[2]) == 0 {
		return fmt.Errorf("%+v: zero dequeue list, %+v", ErrDequeueInstruction, instruction)
	}
	return nil
}
