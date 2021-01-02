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
			beaconHeight, beaconHash, &finalBeaconCommitteeStateV3.beaconCommitteeStateBase,
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

		case instruction.ASSIGN_ACTION:
			assignInstruction, err := instruction.ValidateAndImportAssignInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = newState.processAssignInstruction(
				assignInstruction, env, committeeChange, returnStakingInstruction, oldState)
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
func (engine *BeaconCommitteeEngineV3) GenerateAssignInstruction(rand int64, assignOffset int, activeShards int, beaconHeight uint64) []*instruction.AssignInstruction {
	assignInstructions := []*instruction.AssignInstruction{}

	for i := 0; i < activeShards; i++ {
		shardID := byte(i)
		syncingValidators, _ := incognitokey.CommitteeKeyListToString(engine.finalState.SyncPool()[shardID])

		validKeys := []string{}
		for _, v := range syncingValidators {
			if beaconHeight-syncTerm-engine.finalState.Terms()[v] < 0 {
				break
			}
			validKeys = append(validKeys, v)
		}

		validKeysStruct, _ := incognitokey.CommitteeBase58KeyListToStruct(validKeys)
		assignInstruction := &instruction.AssignInstruction{
			ChainID:               int(shardID),
			ShardCandidates:       validKeys,
			ShardCandidatesStruct: validKeysStruct,
		}

		if !assignInstruction.IsEmpty() {
			assignInstructions = append(assignInstructions, assignInstruction)
		} else {
			Logger.log.Infof("Generate empty assign instruction beacon hash: %s & height: %v \n", engine.beaconHash, engine.beaconHeight)
		}
	}

	return assignInstructions
}
