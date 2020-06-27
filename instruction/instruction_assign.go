package instruction

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	ErrAssignInstruction = errors.New("assign instruction error")
)

type AssignInstruction struct {
	ChainID         int
	ShardCandidates []string
}

func NewAssignInstructionWithValue(chainID int, shardCandidates []string) *AssignInstruction {
	return &AssignInstruction{ChainID: chainID, ShardCandidates: shardCandidates}
}

func NewAssignInstruction() *AssignInstruction {
	return &AssignInstruction{}
}

func (a *AssignInstruction) GetType() string {
	return ASSIGN_ACTION
}

func (a *AssignInstruction) ToString() []string {
	assignInstructionStr := []string{ASSIGN_ACTION}
	assignInstructionStr = append(assignInstructionStr, strings.Join(a.ShardCandidates, SPLITTER))
	assignInstructionStr = append(assignInstructionStr, "shard")
	assignInstructionStr = append(assignInstructionStr, fmt.Sprintf("%v", a.ChainID))
	return assignInstructionStr
}

func (a *AssignInstruction) SetChainID(chainID int) *AssignInstruction {
	a.ChainID = chainID
	return a
}

func (a *AssignInstruction) SetShardCandidates(shardCandidates []string) *AssignInstruction {
	a.ShardCandidates = shardCandidates
	return a
}

func importAssignInstructionFromString(instruction []string) (*AssignInstruction, error) {
	if err := validateAssignInstructionSanity(instruction); err != nil {
		return nil, err
	}
	assignIntruction := NewAssignInstruction()
	tempShardID := instruction[2]
	chainID, err := strconv.Atoi(tempShardID)
	assignIntruction.ChainID = chainID
	if err != nil {
		return nil, err
	}
	if len(instruction[3]) > 0 {
		assignIntruction.ShardCandidates = strings.Split(instruction[3], SPLITTER)
	}
	return assignIntruction, nil
}

func validateAssignInstructionSanity(instruction []string) error {
	if len(instruction) != 4 {
		return fmt.Errorf("%+v: invalid length, %+v", ErrAssignInstruction, instruction)
	}
	if instruction[0] != ASSIGN_ACTION {
		return fmt.Errorf("%+v: invalid assign action, %+v", ErrAssignInstruction, instruction)
	}
	if instruction[2] != SHARD_INST {
		return fmt.Errorf("%+v: invalid assign chain ID, %+v", ErrAssignInstruction, instruction)
	}
	return nil
}
