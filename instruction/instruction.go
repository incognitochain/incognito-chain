package instruction

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"
)

type Instruction interface {
	GetType() string
	ToString() []string
}

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

type CommitteeStateInstruction struct {
	swapInstructions          []*SwapInstruction
	stakeInstructions         []*StakeInstruction
	assignInstructions        []*AssignInstruction
	stopAutoStakeInstructions []*StopAutoStakeInstruction
}

// ImportCommitteeStateInstruction skip all invalid instructions
func ImportCommitteeStateInstruction(instructions [][]string) *CommitteeStateInstruction {
	instructionManager := new(CommitteeStateInstruction)
	for _, instruction := range instructions {
		if len(instruction) < 1 {
			continue
		}
		switch instruction[0] {
		case SWAP_ACTION:
			swapInstruction, err := ValidateAndImportSwapInstructionFromString(instruction)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, "Skip Swap Instruction"))
				continue
			}
			instructionManager.swapInstructions = append(instructionManager.swapInstructions, swapInstruction)
		case ASSIGN_ACTION:
			assignInstruction, err := ValidateAndImportAssignInstructionFromString(instruction)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, "Skip Assign Instruction"))
				continue
			}
			instructionManager.assignInstructions = append(instructionManager.assignInstructions, assignInstruction)
		case STAKE_ACTION:
			stakeInstruction, err := ValidateAndImportStakeInstructionFromString(instruction)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, "Skip Stake Instruction"))
				continue
			}
			instructionManager.stakeInstructions = append(instructionManager.stakeInstructions, stakeInstruction)
		case STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := ValidateAndImportStopAutoStakeInstructionFromString(instruction)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, "Skip Stop Auto Stake Instruction"))
				continue
			}
			instructionManager.stopAutoStakeInstructions = append(instructionManager.stopAutoStakeInstructions, stopAutoStakeInstruction)
		}
	}
	return instructionManager
}

// the order of instruction must always be maintain
func (i *CommitteeStateInstruction) ToString(action string) [][]string {
	instructions := [][]string{}
	switch action {
	case ASSIGN_ACTION:
		for _, assignInstruction := range i.assignInstructions {
			instructions = append(instructions, assignInstruction.ToString())
		}
	case SWAP_ACTION:
		for _, swapInstruction := range i.swapInstructions {
			instructions = append(instructions, swapInstruction.ToString())
		}
	case STAKE_ACTION:
		for _, stakeInstruction := range i.stakeInstructions {
			instructions = append(instructions, stakeInstruction.ToString())
		}
	case STOP_AUTO_STAKE_ACTION:
		for _, stopAutoStakeInstruction := range i.stopAutoStakeInstructions {
			instructions = append(instructions, stopAutoStakeInstruction.ToString())
		}
	}
	return [][]string{}
}

// FilterInstructions filter duplicate instruction
// duplicate instruction is result from delay of shard and beacon
func (i *CommitteeStateInstruction) ValidateAndFilterStakeInstructionsV1(v *ViewEnvironment) {
	panic("implement me")
}
