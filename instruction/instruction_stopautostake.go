package instruction

import (
	"fmt"
	"strings"
)

type StopAutoStakeInstruction struct {
	PublicKeys []string
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

func importStopAutoStakeInstructionFromString(instruction []string) (*StopAutoStakeInstruction, error) {
	if err := validateStopAutoStakeInstructionSanity(instruction); err != nil {
		return nil, err
	}
	stopAutoStakeInstruction := NewStopAutoStakeInstruction()
	if len(instruction[1]) > 0 {
		publicKeys := strings.Split(instruction[1], SPLITTER)
		stopAutoStakeInstruction.PublicKeys = publicKeys
	}
	return stopAutoStakeInstruction, nil
}

func validateStopAutoStakeInstructionSanity(instruction []string) error {
	if len(instruction) != 2 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != STOP_AUTO_STAKE_ACTION {
		return fmt.Errorf("invalid stop auto stake action, %+v", instruction)
	}
	return nil
}
