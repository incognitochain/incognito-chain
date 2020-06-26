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
	Action         string
	ChainID        int
	ShardCandidate []string
}

func NewAssignInstruction() *AssignInstruction {
	return &AssignInstruction{Action: ASSIGN_ACTION}
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
		assignIntruction.ShardCandidate = strings.Split(instruction[3], SPLITTER)
	}
	return assignIntruction, nil
}

func (s *AssignInstruction) ToString() []string {
	assignInstructionStr := []string{ASSIGN_ACTION}
	assignInstructionStr = append(assignInstructionStr, strings.Join(s.ShardCandidate, SPLITTER))
	assignInstructionStr = append(assignInstructionStr, "shard")
	assignInstructionStr = append(assignInstructionStr, fmt.Sprintf("%v", s.ChainID))
	return assignInstructionStr
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
