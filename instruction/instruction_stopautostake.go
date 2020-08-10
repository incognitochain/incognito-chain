package instruction

import (
	"fmt"
	"strings"
)

type StopAutoStakeInstruction struct {
	PublicKeys []string
}

func NewStopAutoStakeInstructionWithValue(publicKeys []string) *StopAutoStakeInstruction {
	return &StopAutoStakeInstruction{PublicKeys: publicKeys}
}

func NewStopAutoStakeInstruction() *StopAutoStakeInstruction {
	return &StopAutoStakeInstruction{}
}

func (s *StopAutoStakeInstruction) GetType() string {
	return STOP_AUTO_STAKE_ACTION
}

func (s *StopAutoStakeInstruction) ToString() []string {
	stopAutoStakeInstructionStr := []string{STOP_AUTO_STAKE_ACTION}
	stopAutoStakeInstructionStr = append(stopAutoStakeInstructionStr, strings.Join(s.PublicKeys, SPLITTER))
	return stopAutoStakeInstructionStr
}

func ValidateAndImportStopAutoStakeInstructionFromString(instruction []string) (*StopAutoStakeInstruction, error) {
	if err := ValidateStopAutoStakeInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportStopAutoStakeInstructionFromString(instruction), nil
}

func ImportStopAutoStakeInstructionFromString(instruction []string) *StopAutoStakeInstruction {
	stopAutoStakeInstruction := NewStopAutoStakeInstruction()
	if len(instruction[1]) > 0 {
		publicKeys := strings.Split(instruction[1], SPLITTER)
		stopAutoStakeInstruction.PublicKeys = publicKeys
	}
	return stopAutoStakeInstruction
}

func ValidateStopAutoStakeInstructionSanity(instruction []string) error {
	if len(instruction) != 2 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != STOP_AUTO_STAKE_ACTION {
		return fmt.Errorf("invalid stop auto stake action, %+v", instruction)
	}
	return nil
}
