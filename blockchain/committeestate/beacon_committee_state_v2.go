package committeestate

import (
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateV2 struct {
	beaconCommitteeStateSlashingBase
}

func NewBeaconCommitteeStateV2() *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		beaconCommitteeStateSlashingBase: *newBeaconCommitteeStateSlashingBase(),
	}
}

func NewBeaconCommitteeStateV2WithValue(
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	shardCommonPool []incognitokey.CommitteePublicKey,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	swapRule SwapRule,
) *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		beaconCommitteeStateSlashingBase: *newBeaconCommitteeStateSlashingBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute,
			autoStake, rewardReceiver, stakingTx,
			shardCommonPool,
			numberOfAssignedCandidates, swapRule,
		),
	}
}

func (b BeaconCommitteeStateV2) Version() uint {
	return SLASHING_VERSION
}

func (b *BeaconCommitteeStateV2) cloneFrom(fromB BeaconCommitteeStateV2) {
	b.reset()
	b.beaconCommitteeStateSlashingBase.cloneFrom(fromB.beaconCommitteeStateSlashingBase)
}

func (b *BeaconCommitteeStateV2) clone() *BeaconCommitteeStateV2 {
	res := NewBeaconCommitteeStateV2()
	res.beaconCommitteeStateSlashingBase = *b.beaconCommitteeStateSlashingBase.clone()
	return res
}

func (b *BeaconCommitteeStateV2) reset() {
	b.beaconCommitteeStateBase.reset()
}

//Upgrade check interface method for des
func (engine *BeaconCommitteeStateV2) Upgrade(env *BeaconCommitteeStateEnvironment) BeaconCommitteeState {
	beaconCommittee, shardCommittee, shardSubstitute,
		shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule := engine.getDataForUpgrading(env)

	committeeStateV3 := NewBeaconCommitteeStateV3WithValue(
		beaconCommittee,
		shardCommittee,
		shardSubstitute,
		shardCommonPool,
		numberOfAssignedCandidates,
		autoStake,
		rewardReceiver,
		stakingTx,
		map[byte][]incognitokey.CommitteePublicKey{},
		swapRule,
	)
	return committeeStateV3
}

// UpdateCommitteeState New flow
// Store information from instructions into temp stateDB in env
// When all thing done and no problems, in commit function, we read data in statedb and update
func (b *BeaconCommitteeStateV2) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	var err error
	incurredInstructions := [][]string{}
	returnStakingInstruction := instruction.NewReturnStakeIns()
	committeeChange := NewCommitteeChange()
	oldB := b.clone()

	oldB.mu.RLock()
	defer oldB.mu.RUnlock()
	b.mu.Lock()
	defer b.mu.Unlock()

	// snapshot shard common pool in beacon random time
	if env.IsBeaconRandomTime {
		b.SetNumberOfAssignedCandidates(SnapshotShardCommonPoolV2(
			oldB.GetShardCommonPool(),
			oldB.GetShardCommittee(),
			oldB.GetShardSubstitute(),
			env.NumberOfFixedShardBlockValidator,
			env.MinShardCommitteeSize,
			oldB.SwapRule(),
		))

		Logger.log.Infof("Block %+v, Number of Snapshot to Assign Candidate %+v", env.BeaconHeight, b.NumberOfAssignedCandidates())
	}

	env.newUnassignedCommonPool = b.UnassignedCommonPool()
	env.newAllSubstituteCommittees = b.AllSubstituteCommittees()
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
			committeeChange, err = b.processStakeInstruction(stakeInstruction, committeeChange)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

		case instruction.RANDOM_ACTION:
			randomInstruction, err := instruction.ValidateAndImportRandomInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = b.processAssignWithRandomInstruction(
				randomInstruction.BtcNonce, env.ActiveShards, committeeChange, oldB)

		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = b.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange, oldB)

		case instruction.UNSTAKE_ACTION:
			unstakeInstruction, err := instruction.ValidateAndImportUnstakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = b.processUnstakeInstruction(
				unstakeInstruction, env, committeeChange, returnStakingInstruction, oldB)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

		case instruction.SWAP_SHARD_ACTION:
			swapShardInstruction, err := instruction.ValidateAndImportSwapShardInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = b.processSwapShardInstruction(
				swapShardInstruction, env, committeeChange, returnStakingInstruction, oldB)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		}
	}

	hashes, err := b.Hash()
	if err != nil {
		return hashes, committeeChange, incurredInstructions, err
	}
	if !returnStakingInstruction.IsEmpty() {
		incurredInstructions = append(incurredInstructions, returnStakingInstruction.ToString())
	}

	return hashes, committeeChange, incurredInstructions, nil
}

func (b *BeaconCommitteeStateV2) getDataForUpgrading(env *BeaconCommitteeStateEnvironment) (
	[]incognitokey.CommitteePublicKey,
	map[byte][]incognitokey.CommitteePublicKey,
	map[byte][]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	int,
	map[string]bool,
	map[string]privacy.PaymentAddress,
	map[string]common.Hash,
	SwapRule,
) {
	beaconCommittee, shardCommittee, shardSubstitute,
		shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule := b.getDataForUpgrading(env)

	numberOfAssignedCandidates = b.NumberOfAssignedCandidates()
	shardCommonPool = make([]incognitokey.CommitteePublicKey, numberOfAssignedCandidates)
	copy(shardCommonPool, b.shardCommonPool)
	return beaconCommittee, shardCommittee, shardSubstitute, shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule
}
