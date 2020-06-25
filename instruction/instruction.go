package instruction

import "github.com/pkg/errors"

type InstructionManager struct {
	swapInstructions          []*SwapInstruction
	stakeInstructions         []*StakeInstruction
	assignInstructions        []*AssignInstruction
	stopAutoStakeInstructions []*StopAutoStakeInstruction
}

func ImportInstructionFromStringArray(instructions [][]string, chainID int) (*InstructionManager, error) {
	instructionManager := new(InstructionManager)
	for _, instruction := range instructions {
		if len(instruction) < 1 {
			continue
		}
		switch instruction[0] {
		case SWAP_ACTION:
			swapInstruction, err := importSwapInstructionFromString(instruction, chainID)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, ""))
			}
			instructionManager.swapInstructions = append(instructionManager.swapInstructions, swapInstruction)
		case ASSIGN_ACTION:
			assignInstruction, err := importAssignInstructionFromString(instruction)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, ""))
			}
			instructionManager.assignInstructions = append(instructionManager.assignInstructions, assignInstruction)
		case STAKE_ACTION:
			stakeInstruction, err := importStakeInstructionFromString(instruction)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, ""))
			}
			instructionManager.stakeInstructions = append(instructionManager.stakeInstructions, stakeInstruction)
		case STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := importStopAutoStakeInstructionFromString(instruction)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, ""))
			}
			instructionManager.stopAutoStakeInstructions = append(instructionManager.stopAutoStakeInstructions, stopAutoStakeInstruction)
		}
	}
	return instructionManager, nil
}

// FilterInstructions filter duplicate instruction
// duplicate instruction is result from delay of shard and beacon
func (i *InstructionManager) FilterInstructions() {

}

func (i *InstructionManager) ToString() [][]string {
	return [][]string{}
}
