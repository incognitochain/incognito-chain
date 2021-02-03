package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/instruction"
)

type BeaconCommitteeEngineV2 struct {
	beaconCommitteeEngineSlashingBase
}

func NewBeaconCommitteeEngineV2(
	beaconHeight uint64,
	beaconHash common.Hash,
	finalBeaconCommitteeStateV2 *BeaconCommitteeStateV2) *BeaconCommitteeEngineV2 {
	Logger.log.Infof("Init Beacon Committee Engine V2, %+v", beaconHeight)
	return &BeaconCommitteeEngineV2{
		beaconCommitteeEngineSlashingBase: *NewBeaconCommitteeEngineSlashingBaseWithValue(
			beaconHeight, beaconHash, &finalBeaconCommitteeStateV2.beaconCommitteeStateBase,
		),
	}
}

//Version :
func (engine BeaconCommitteeEngineV2) Version() uint {
	return SLASHING_VERSION
}

//Clone :
func (engine *BeaconCommitteeEngineV2) Clone() BeaconCommitteeEngine {
	res := &BeaconCommitteeEngineV2{
		beaconCommitteeEngineSlashingBase: *engine.beaconCommitteeEngineSlashingBase.Clone().(*beaconCommitteeEngineSlashingBase),
	}
	return res
}

// UpdateCommitteeState New flow
// Store information from instructions into temp stateDB in env
// When all thing done and no problems, in commit function, we read data in statedb and update
func (engine *BeaconCommitteeEngineV2) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	var err error
	incurredInstructions := [][]string{}
	returnStakingInstruction := instruction.NewReturnStakeIns()
	committeeChange := NewCommitteeChange()
	oldState := engine.finalState.(*BeaconCommitteeStateV2)

	oldState.Mu().RLock()
	defer oldState.Mu().RUnlock()
	if engine.uncommittedState == nil {
		engine.uncommittedState = NewBeaconCommitteeStateV2()
	}
	cloneBeaconCommitteeStateFromTo(oldState, engine.uncommittedState)
	newState := engine.uncommittedState.(*BeaconCommitteeStateV2)

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
	env.newValidators = append(env.newUnassignedCommonPool, env.newAllSubstituteCommittees...)

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
			committeeChange, err = newState.processStakeInstruction(stakeInstruction, committeeChange)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

		case instruction.RANDOM_ACTION:
			randomInstruction, err := instruction.ValidateAndImportRandomInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = newState.processAssignWithRandomInstruction(
				randomInstruction.BtcNonce, env.ActiveShards, committeeChange, oldState)

		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = newState.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange, oldState)

		case instruction.UNSTAKE_ACTION:
			unstakeInstruction, err := instruction.ValidateAndImportUnstakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = newState.processUnstakeInstruction(
				unstakeInstruction, env, committeeChange, returnStakingInstruction, oldState)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

		case instruction.SWAP_SHARD_ACTION:
			swapShardInstruction, err := instruction.ValidateAndImportSwapShardInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = newState.processSwapShardInstruction(
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
