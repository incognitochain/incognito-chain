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

func IsConsensusInstruction(action string) bool {
	return action == RANDOM_ACTION ||
		action == SWAP_ACTION ||
		action == STAKE_ACTION ||
		action == BEACON_STAKE_ACTION ||
		action == ASSIGN_ACTION ||
		action == STOP_AUTO_STAKE_ACTION ||
		action == SET_ACTION ||
		action == SWAP_SHARD_ACTION ||
		action == UNSTAKE_ACTION ||
		action == ACCEPT_BLOCK_REWARD_V3_ACTION ||
		action == SHARD_RECEIVE_REWARD_V3_ACTION ||
		action == FINISH_SYNC_ACTION ||
		action == SHARD_INST ||
		action == BEACON_INST ||
		action == RETURN_ACTION ||
		action == RETURN_BEACON_ACTION ||
		action == ADD_STAKING_ACTION ||
		action == RE_DELEGATE
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

func ValidateAndImportInstructionFromString(inst []string) (
	instStr Instruction,
	err error,
) {
	action := inst[0]
	var buildInstructionFromString func(x []string) (Instruction, error)
	if !IsConsensusInstruction(action) {
		return nil, errors.Errorf("this inst %v is not Consensus instruction", inst)
	}
	switch action {
	case STAKE_ACTION:
		buildInstructionFromString = BuildStakeInstructionFromString
	case RANDOM_ACTION:
		buildInstructionFromString = BuildRandomInstructionFromString
	case STOP_AUTO_STAKE_ACTION:
		buildInstructionFromString = BuildStopAutoStakeInstructionFromString
	case SWAP_ACTION:
		buildInstructionFromString = BuildSwapInstructionFromString
	case SWAP_SHARD_ACTION:
		buildInstructionFromString = BuildSwapShardInstructionFromString
	case FINISH_SYNC_ACTION:
		buildInstructionFromString = BuildFinishSyncInstructionFromString
	case UNSTAKE_ACTION:
		buildInstructionFromString = BuildUnstakeInstructionFromString
	case ADD_STAKING_ACTION:
		buildInstructionFromString = BuildAddStakingInstructionFromString
	case RETURN_ACTION:
		buildInstructionFromString = BuildReturnStakingInstructionFromString
	case RETURN_BEACON_ACTION:
		buildInstructionFromString = BuildReturnBeaconStakingInstructionFromString
	case RE_DELEGATE:
		buildInstructionFromString = BuildReDelegateInstructionFromString
	default:
		panic(action)
	}
	instStr, err = buildInstructionFromString(inst)
	if err != nil {
		return nil, errors.Errorf("SKIP %v instruction %+v, error %+v", action, inst, err)
	}
	return instStr, nil
}

func ValidateAndImportConsensusInstructionFromListString(insts [][]string) []Instruction {
	iInsts := []Instruction{}
	for _, instString := range insts {
		iInst, err := ValidateAndImportInstructionFromString(instString)
		if err != nil {
			Logger.Log.Error(err)
			continue
		}
		iInsts = append(iInsts, iInst)
	}
	return iInsts
}

func BuildStakeInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateStakeInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportStakeInstructionFromString(instruction), nil
}
func BuildRandomInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateRandomInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportRandomInstructionFromString(instruction), nil
}
func BuildStopAutoStakeInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateStopAutoStakeInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportStopAutoStakeInstructionFromString(instruction), nil
}

func BuildSwapInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateSwapInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportSwapInstructionFromString(instruction), nil
}

func BuildSwapShardInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateSwapShardInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportSwapShardInstructionFromString(instruction), nil
}

func BuildFinishSyncInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateFinishSyncInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportFinishSyncInstructionFromString(instruction)
}

func BuildUnstakeInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateUnstakeInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportUnstakeInstructionFromString(instruction), nil
}

func BuildAddStakingInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateAddStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportAddStakingInstructionFromString(instruction), nil
}

func BuildReturnStakingInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateReturnStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReturnStakingInstructionFromString(instruction)
}

func BuildReturnBeaconStakingInstructionFromString(instruction []string) (Instruction, error) {
	if err := ValidateReturnBeaconStakingInstructionSanity(instruction); err != nil {
		return nil, err
	}
	return ImportReturnBeaconStakingInstructionFromString(instruction)
}
