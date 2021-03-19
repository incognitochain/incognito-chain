package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
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
	beaconCommittee []string,
	shardCommittee map[byte][]string,
	shardSubstitute map[byte][]string,
	shardCommonPool []string,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	swapRule SwapRuleProcessor,
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

func (b BeaconCommitteeStateV2) Version() int {
	return SLASHING_VERSION
}

func (b *BeaconCommitteeStateV2) Clone() BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.clone()
}

func (b *BeaconCommitteeStateV2) clone() *BeaconCommitteeStateV2 {
	res := NewBeaconCommitteeStateV2()
	res.beaconCommitteeStateSlashingBase = *b.beaconCommitteeStateSlashingBase.clone()
	return res
}

func InitCommitteeStateV2(env *BeaconCommitteeStateEnvironment) *BeaconCommitteeStateV2 {
	beaconCommitteeStateV2 := NewBeaconCommitteeStateV2()
	beaconCommitteeStateV2.initCommitteeState(env)
	return beaconCommitteeStateV2
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
	b.mu.Lock()
	defer b.mu.Unlock()

	// snapshot shard common pool in beacon random time
	if env.IsBeaconRandomTime {
		b.numberOfAssignedCandidates = SnapshotShardCommonPoolV2(
			b.shardCommonPool,
			b.shardCommittee,
			b.shardSubstitute,
			env.NumberOfFixedShardBlockValidator,
			env.MinShardCommitteeSize,
			b.swapRule,
		)

		Logger.log.Infof("Block %+v, Number of Snapshot to Assign Candidate %+v", env.BeaconHeight, b.numberOfAssignedCandidates)
	}
	b.addData(env)
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
				randomInstruction.RandomNumber(), env.numberOfValidator, committeeChange)

		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = b.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange)

		case instruction.UNSTAKE_ACTION:
			unstakeInstruction, err := instruction.ValidateAndImportUnstakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = b.processUnstakeInstruction(
				unstakeInstruction, env, committeeChange, returnStakingInstruction)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

		case instruction.SWAP_SHARD_ACTION:
			swapShardInstruction, err := instruction.ValidateAndImportSwapShardInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = b.processSwapShardInstruction(
				swapShardInstruction, env.numberOfValidator, env, committeeChange, returnStakingInstruction)
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

//Upgrade check interface method for des
func (b *BeaconCommitteeStateV2) Upgrade(env *BeaconCommitteeStateEnvironment) BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	beaconCommittee, shardCommittee, shardSubstitute,
		shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule := b.getDataForUpgrading(env)

	committeeStateV3 := NewBeaconCommitteeStateV3WithValue(
		beaconCommittee,
		shardCommittee,
		shardSubstitute,
		shardCommonPool,
		numberOfAssignedCandidates,
		autoStake,
		rewardReceiver,
		stakingTx,
		map[byte][]string{},
		swapRule,
	)
	return committeeStateV3
}

func (b *BeaconCommitteeStateV2) getDataForUpgrading(env *BeaconCommitteeStateEnvironment) (
	[]string,
	map[byte][]string,
	map[byte][]string,
	[]string,
	int,
	map[string]bool,
	map[string]privacy.PaymentAddress,
	map[string]common.Hash,
	SwapRuleProcessor,
) {
	beaconCommittee := make([]string, len(b.beaconCommittee))
	shardCommittee := make(map[byte][]string)
	shardSubstitute := make(map[byte][]string)
	shardCommonPool := make([]string, len(b.shardCommittee))
	numberOfAssignedCandidates := b.numberOfAssignedCandidates
	autoStake := make(map[string]bool)
	rewardReceiver := make(map[string]privacy.PaymentAddress)
	stakingTx := make(map[string]common.Hash)
	swapRule := b.swapRule

	copy(beaconCommittee, b.beaconCommittee)
	for shardID, oneShardCommittee := range b.shardCommittee {
		shardCommittee[shardID] = make([]string, len(oneShardCommittee))
		copy(shardCommittee[shardID], oneShardCommittee)
	}
	for shardID, oneShardSubsitute := range b.shardSubstitute {
		shardSubstitute[shardID] = make([]string, len(oneShardSubsitute))
		copy(shardSubstitute[shardID], oneShardSubsitute)
	}
	nextEpochShardCandidate := b.shardCommonPool[numberOfAssignedCandidates:]
	currentEpochShardCandidate := b.shardCommonPool[:numberOfAssignedCandidates]
	shardCandidates := append(currentEpochShardCandidate, nextEpochShardCandidate...)

	copy(shardCommonPool, shardCandidates)
	for k, v := range b.autoStake {
		autoStake[k] = v
	}
	for k, v := range b.rewardReceiver {
		rewardReceiver[k] = v
	}
	for k, v := range b.stakingTx {
		stakingTx[k] = v
	}

	return beaconCommittee, shardCommittee, shardSubstitute, shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule
}
