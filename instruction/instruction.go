package instruction

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
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

type InstructionManager struct {
	swapInstructions          []*SwapInstruction
	stakeInstructions         []*StakeInstruction
	assignInstructions        []*AssignInstruction
	stopAutoStakeInstructions []*StopAutoStakeInstruction
}

// the order of instruction must always be maintain
func (i *InstructionManager) ToString(action string) [][]string {
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
func (i *InstructionManager) ValidateAndFilterStakeInstructionsV1(v *ViewEnvironment) {
	panic("implement me")
}
