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
	SwapInstructions          []*SwapInstruction
	StakeInstructions         []*StakeInstruction
	AssignInstructions        []*AssignInstruction
	StopAutoStakeInstructions []*StopAutoStakeInstruction
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
			instructionManager.SwapInstructions = append(instructionManager.SwapInstructions, swapInstruction)
		case ASSIGN_ACTION:
			assignInstruction, err := ValidateAndImportAssignInstructionFromString(instruction)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, "Skip Assign Instruction"))
				continue
			}
			instructionManager.AssignInstructions = append(instructionManager.AssignInstructions, assignInstruction)
		case STAKE_ACTION:
			stakeInstruction, err := ValidateAndImportStakeInstructionFromString(instruction)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, "Skip Stake Instruction"))
				continue
			}
			instructionManager.StakeInstructions = append(instructionManager.StakeInstructions, stakeInstruction)
		case STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := ValidateAndImportStopAutoStakeInstructionFromString(instruction)
			if err != nil {
				Logger.Log.Error(errors.Wrap(err, "Skip Stop Auto Stake Instruction"))
				continue
			}
			instructionManager.StopAutoStakeInstructions = append(instructionManager.StopAutoStakeInstructions, stopAutoStakeInstruction)
		}
	}
	return instructionManager
}

// the order of instruction must always be maintain
func (i *CommitteeStateInstruction) ToString(action string) [][]string {
	instructions := [][]string{}
	switch action {
	case ASSIGN_ACTION:
		for _, assignInstruction := range i.AssignInstructions {
			instructions = append(instructions, assignInstruction.ToString())
		}
	case SWAP_ACTION:
		for _, swapInstruction := range i.SwapInstructions {
			instructions = append(instructions, swapInstruction.ToString())
		}
	case STAKE_ACTION:
		for _, stakeInstruction := range i.StakeInstructions {
			instructions = append(instructions, stakeInstruction.ToString())
		}
	case STOP_AUTO_STAKE_ACTION:
		for _, stopAutoStakeInstruction := range i.StopAutoStakeInstructions {
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
