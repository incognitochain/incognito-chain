package instruction

import "fmt"

type StopAutoStakingInstruction struct {
	Action    int
	PublicKey string
}

func importStopAutoStakingInstructionFromString([]string) []*StopAutoStakingInstruction {
}

func (s *StopAutoStakingInstruction) toString() []string {
	return []string{}
}

func validateStopAutoStakeInstructionSanity(instruction []string) error {
	if len(instruction) != 2 {
		return fmt.Errorf("invalid length, %+v", instruction)
	}
	if instruction[0] != stopAutoStake {
		return fmt.Errorf("invalid stop auto stake action, %+v", instruction)
	}
	return nil
}
