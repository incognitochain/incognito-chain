package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/instruction"
)

type BeaconCommitteeEngineV3 struct {
	beaconCommitteeEngineBase
}

func NewBeaconCommitteeEngineV3(
	beaconHeight uint64,
	beaconHash common.Hash,
	finalBeaconCommitteeStateV3 *BeaconCommitteeStateV3) *BeaconCommitteeEngineV3 {
	Logger.log.Infof("Init Beacon Committee Engine V2, %+v", beaconHeight)
	return &BeaconCommitteeEngineV3{
		beaconCommitteeEngineBase: beaconCommitteeEngineBase{
			beaconHeight:     beaconHeight,
			beaconHash:       beaconHash,
			finalState:       finalBeaconCommitteeStateV3,
			uncommittedState: NewBeaconCommitteeStateV3(),
		},
	}
}

//Version :
func (engine BeaconCommitteeEngineV3) Version() uint {
	return DCS_VERSION
}

func (engine *BeaconCommitteeEngineV3) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	var err error
	incurredInstructions := [][]string{}
	returnStakingInstruction := instruction.NewReturnStakeIns()
	committeeChange := NewCommitteeChange()
	oldState := engine.finalState

	oldState.Mu().RLock()
	defer oldState.Mu().RUnlock()

	engine.uncommittedState = cloneBeaconCommitteeStateFrom(oldState)
	newState := engine.uncommittedState

	newState.Mu().Lock()
	defer newState.Mu().Unlock()

	// snapshot shard common pool in beacon random time
	if env.IsBeaconRandomTime {
		newState.SetNumberOfAssignedCandidates(SnapshotShardCommonPoolV2(
			oldState.ShardCommonPool(),
			oldState.ShardCommittee(),
			oldState.ShardSubstitute(),
			env.NumberOfFixedShardBlockValidator,
			env.MinShardCommitteeSize,
			oldState.SwapRule(),
		))

		Logger.log.Infof("Block %+v, Number of Snapshot to Assign Candidate %+v", env.BeaconHeight, newState.NumberOfAssignedCandidates())
	}

	env.newUnassignedCommonPool = newState.UnassignedCommonPool()
	env.newAllSubstituteCommittees = newState.AllSubstituteCommittees()
	env.newAllCandidateSubstituteCommittee = append(env.newUnassignedCommonPool, env.newAllSubstituteCommittees...)

	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		switch inst[0] {
		case instruction.STAKE_ACTION:
			stakeInstruction, err := instruction.ValidateAndImportStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, err = newState.ProcessStakeInstruction(stakeInstruction, committeeChange)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		case instruction.RANDOM_ACTION:
			randomInstruction, err := instruction.ValidateAndImportRandomInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = newState.ProcessAssignWithRandomInstruction(
				randomInstruction.BtcNonce, env.ActiveShards, committeeChange, oldState)
		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = newState.ProcessStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange, oldState)
		case instruction.UNSTAKE_ACTION:
			unstakeInstruction, err := instruction.ValidateAndImportUnstakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = newState.ProcessUnstakeInstruction(
				unstakeInstruction, env, committeeChange, returnStakingInstruction, oldState)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		case instruction.SWAP_SHARD_ACTION:
			swapShardInstruction, err := instruction.ValidateAndImportSwapShardInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = newState.ProcessSwapShardInstruction(
				swapShardInstruction, env, committeeChange, returnStakingInstruction, oldState)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		}
	}

	hashes, err := newState.Hash()
	if err != nil {
		return hashes, committeeChange, incurredInstructions, err
	}
	if !returnStakingInstruction.IsEmpty() {
		incurredInstructions = append(incurredInstructions, returnStakingInstruction.ToString())
	}

	return hashes, committeeChange, incurredInstructions, nil
}

//GenerateAssignInstruction generate assign instructions for assign from syncing pool to shard pending pool
// TODO: @tin Overridew from parent function and add validators from syncpool to shard pending pool
func (engine *BeaconCommitteeEngineV3) GenerateAssignInstruction(rand int64, assignOffset int, activeShards int) ([]*instruction.AssignInstruction, []string, map[byte][]string) {
	return []*instruction.AssignInstruction{}, []string{}, make(map[byte][]string)
}
