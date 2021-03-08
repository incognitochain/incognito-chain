package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

type BeaconCommitteeEngineV3 struct {
	beaconCommitteeEngineSlashingBase
}

func NewBeaconCommitteeEngineV3(
	beaconHeight uint64,
	beaconHash common.Hash,
	finalBeaconCommitteeStateV3 *BeaconCommitteeStateV3) *BeaconCommitteeEngineV3 {
	Logger.log.Infof("Init Beacon Committee Engine V2, %+v", beaconHeight)
	return &BeaconCommitteeEngineV3{
		beaconCommitteeEngineSlashingBase: *NewBeaconCommitteeEngineSlashingBaseWithValue(
			beaconHeight, beaconHash, finalBeaconCommitteeStateV3, NewBeaconCommitteeStateV3(),
		),
	}
}

//Version :
func (engine BeaconCommitteeEngineV3) Version() uint {
	return DCS_VERSION
}

//Clone :
func (engine *BeaconCommitteeEngineV3) Clone() BeaconCommitteeEngine {
	res := &BeaconCommitteeEngineV3{
		beaconCommitteeEngineSlashingBase: *engine.beaconCommitteeEngineSlashingBase.Clone().(*beaconCommitteeEngineSlashingBase),
	}
	return res
}

func (engine *BeaconCommitteeEngineV3) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	var err error
	incurredInstructions := [][]string{}
	returnStakingInstruction := instruction.NewReturnStakeIns()
	committeeChange := NewCommitteeChange()
	oldState := engine.finalState.(*BeaconCommitteeStateV3)

	oldState.Mu().RLock()
	defer oldState.Mu().RUnlock()
	if engine.uncommittedState == nil {
		engine.uncommittedState = NewBeaconCommitteeStateV3()
	}
	cloneBeaconCommitteeStateFromTo(oldState, engine.uncommittedState)
	newState := engine.uncommittedState.(*BeaconCommitteeStateV3)

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
	env.newValidators = append(env.newValidators, newState.AllSyncingValidators()...)

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
				randomInstruction.BtcNonce, env.ActiveShards, committeeChange, oldState, env.BeaconHeight)

		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = newState.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange, oldState)

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

		case instruction.FINISH_SYNC_ACTION:
			finishSyncInstruction, err := instruction.ValidateAndImportFinishSyncInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, err = newState.processFinishSyncInstruction(
				finishSyncInstruction, env, committeeChange, oldState)
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

func (engine BeaconCommitteeEngineV3) SyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	engine.finalState.Mu().RLock()
	defer engine.finalState.Mu().RUnlock()
	res := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.finalState.SyncPool() {
		res[k] = v
	}
	return res
}
