package committeestate

import (
	"reflect"
	"sync"

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
func NewBeaconCommitteeStateV2WithMu(mu *sync.RWMutex) *BeaconCommitteeStateV2 {

	return &BeaconCommitteeStateV2{
		beaconCommitteeStateSlashingBase: beaconCommitteeStateSlashingBase{
			beaconCommitteeStateBase: beaconCommitteeStateBase{
				shardCommittee:  make(map[byte][]string),
				shardSubstitute: make(map[byte][]string),
				autoStake:       make(map[string]bool),
				rewardReceiver:  make(map[string]privacy.PaymentAddress),
				stakingTx:       make(map[string]common.Hash),
				mu:              mu,
			},
		},
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
	delegateList map[string]string,
	swapRule SwapRuleProcessor,
	assignRule AssignRuleProcessor,
) *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		beaconCommitteeStateSlashingBase: *newBeaconCommitteeStateSlashingBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute,
			autoStake, rewardReceiver, stakingTx, delegateList,
			shardCommonPool,
			numberOfAssignedCandidates, swapRule, assignRule,
		),
	}
}

// shallowCopy maintain dst mutex value
func (b *BeaconCommitteeStateV2) shallowCopy(newB *BeaconCommitteeStateV2) {
	newB.beaconCommitteeStateSlashingBase = b.beaconCommitteeStateSlashingBase
}

// Version :
func (b *BeaconCommitteeStateV2) Version() int {
	return STAKING_FLOW_V2
}

func initGenesisBeaconCommitteeStateV2(env *BeaconCommitteeStateEnvironment) *BeaconCommitteeStateV2 {
	beaconCommitteeStateV2 := NewBeaconCommitteeStateV2()
	beaconCommitteeStateV2.initCommitteeState(env)
	return beaconCommitteeStateV2
}

func (b *BeaconCommitteeStateV2) UpgradeAssignRuleV3() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.assignRule = NewAssignRuleV3()
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
	b.setBeaconCommitteeStateHashes(env.PreviousBlockHashes)

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
		case instruction.RE_DELEGATE:
			redelegateInstruction, err := instruction.ValidateAndImportReDelegateInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
				continue
			}
			b.processReDelegateInstruction(redelegateInstruction, env, committeeChange)
		}
	}

	hashes, err := b.Hash(committeeChange)
	if err != nil {
		return hashes, committeeChange, incurredInstructions, err
	}
	if !returnStakingInstruction.IsEmpty() {
		incurredInstructions = append(incurredInstructions, returnStakingInstruction.ToString())
	}

	return hashes, committeeChange, incurredInstructions, nil
}

// Upgrade check interface method for des
func (b *BeaconCommitteeStateV2) Upgrade(env *BeaconCommitteeStateEnvironment) BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	beaconCommittee, shardCommittee, shardSubstitute,
		shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, delegates := b.getDataForUpgrading(env)

	committeeStateV3 := NewBeaconCommitteeStateV3WithValue(
		beaconCommittee,
		shardCommittee,
		shardSubstitute,
		shardCommonPool,
		numberOfAssignedCandidates,
		autoStake,
		rewardReceiver,
		stakingTx,
		delegates,
		map[byte][]string{},
		NewSwapRuleV3(),
		NewAssignRuleV3(),
	)

	Logger.log.Infof("Upgrade Committee State V2 to V3, swap rule %+v, assign rule %+v",
		reflect.TypeOf(*NewSwapRuleV3()), reflect.TypeOf(*NewAssignRuleV3()))
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
	map[string]string,
) {
	shardCommittee := make(map[byte][]string)
	shardSubstitute := make(map[byte][]string)
	numberOfAssignedCandidates := b.numberOfAssignedCandidates
	autoStake := make(map[string]bool)
	rewardReceiver := make(map[string]privacy.PaymentAddress)
	stakingTx := make(map[string]common.Hash)
	delegates := make(map[string]string)

	beaconCommittee := common.DeepCopyString(b.beaconCommittee)

	for shardID, oneShardCommittee := range b.shardCommittee {
		shardCommittee[shardID] = common.DeepCopyString(oneShardCommittee)
	}
	for shardID, oneShardSubsitute := range b.shardSubstitute {
		shardSubstitute[shardID] = common.DeepCopyString(oneShardSubsitute)
	}
	nextEpochShardCandidate := b.shardCommonPool[numberOfAssignedCandidates:]
	currentEpochShardCandidate := b.shardCommonPool[:numberOfAssignedCandidates]
	shardCandidates := append(currentEpochShardCandidate, nextEpochShardCandidate...)

	shardCommonPool := common.DeepCopyString(shardCandidates)
	for k, v := range b.autoStake {
		autoStake[k] = v
	}
	for k, v := range b.rewardReceiver {
		rewardReceiver[k] = v
	}
	for k, v := range b.stakingTx {
		stakingTx[k] = v
	}
	for k, v := range b.delegate {
		delegates[k] = v
	}

	return beaconCommittee, shardCommittee, shardSubstitute, shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, delegates
}

func (b *beaconCommitteeStateSlashingBase) addData(env *BeaconCommitteeStateEnvironment) {
	env.newUnassignedCommonPool = common.DeepCopyString(b.shardCommonPool[b.numberOfAssignedCandidates:])
	env.newAllSubstituteCommittees, _ = b.getAllSubstituteCommittees()
	env.newAllRoles = append(env.newUnassignedCommonPool, env.newAllSubstituteCommittees...)
	env.shardCommittee = make(map[byte][]string)
	for shardID, committees := range b.shardCommittee {
		env.shardCommittee[shardID] = common.DeepCopyString(committees)
	}
	env.shardSubstitute = make(map[byte][]string)
	for shardID, substitutes := range b.shardSubstitute {
		env.shardSubstitute[shardID] = common.DeepCopyString(substitutes)
	}
	env.numberOfValidator = make([]int, env.ActiveShards)
	for i := 0; i < env.ActiveShards; i++ {
		env.numberOfValidator[i] += len(b.shardCommittee[byte(i)])
		env.numberOfValidator[i] += len(b.shardSubstitute[byte(i)])
	}
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
