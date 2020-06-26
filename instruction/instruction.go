package instruction

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"
)

type ViewEnvironment struct {
	beaconCommittee                        []incognitokey.CommitteePublicKey
	beaconSubstitute                       []incognitokey.CommitteePublicKey
	candidateShardWaitingForCurrentRandom  []incognitokey.CommitteePublicKey
	candidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey
	candidateShardWaitingForNextRandom     []incognitokey.CommitteePublicKey
	candidateBeaconWaitingForNextRandom    []incognitokey.CommitteePublicKey
	shardCommittee                         map[byte][]incognitokey.CommitteePublicKey
	shardSubstitute                        map[byte][]incognitokey.CommitteePublicKey
}

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

// the order of instruction must always be maintain
func (i *InstructionManager) ToString(action string) [][]string {
	instructions := [][]string{}
	switch action {
	case ASSIGN_ACTION:
		for _, assignInstruction := range i.assignInstructions {
			instructions = append(instructions, assignInstruction.toString())
		}
	case SWAP_ACTION:
		for _, swapInstruction := range i.swapInstructions {
			instructions = append(instructions, swapInstruction.toString())
		}
	case STAKE_ACTION:
		for _, stakeInstruction := range i.stakeInstructions {
			instructions = append(instructions, stakeInstruction.toString())
		}
	case STOP_AUTO_STAKE_ACTION:
		for _, stopAutoStakeInstruction := range i.stopAutoStakeInstructions {
			instructions = append(instructions, stopAutoStakeInstruction.toString())
		}
	}
	return [][]string{}
}

// FilterInstructions filter duplicate instruction
// duplicate instruction is result from delay of shard and beacon
func (i *InstructionManager) ValidateAndFilterStakeInstructionsV1(v *ViewEnvironment) {
	panic("implement me")
}
