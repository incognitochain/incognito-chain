package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"reflect"
)

type BeaconCommitteeStateV3 struct {
	beaconCommitteeStateSlashingBase
	syncPool map[byte][]incognitokey.CommitteePublicKey
}

func NewBeaconCommitteeStateV3() *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
		beaconCommitteeStateSlashingBase: *newBeaconCommitteeStateSlashingBase(),
		syncPool:                         make(map[byte][]incognitokey.CommitteePublicKey),
	}
}

func NewBeaconCommitteeStateV3WithValue(
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	shardCommonPool []incognitokey.CommitteePublicKey,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	syncPool map[byte][]incognitokey.CommitteePublicKey,
	swapRule SwapRuleProcessor,
) *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
		beaconCommitteeStateSlashingBase: *newBeaconCommitteeStateSlashingBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute, autoStake, rewardReceiver, stakingTx,
			shardCommonPool, numberOfAssignedCandidates, swapRule,
		),
		syncPool: syncPool,
	}
}

func (b *BeaconCommitteeStateV3) Version() uint {
	return DCS_VERSION
}

func (b *BeaconCommitteeStateV3) Clone() BeaconCommitteeState {
	return b.clone()
}

func (b *BeaconCommitteeStateV3) clone() *BeaconCommitteeStateV3 {
	newB := NewBeaconCommitteeStateV3()
	newB.beaconCommitteeStateSlashingBase = *b.beaconCommitteeStateSlashingBase.clone()

	for i, v := range b.syncPool {
		newB.syncPool[i] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(newB.syncPool[i], v)
	}

	return newB
}

func (b BeaconCommitteeStateV3) Hash() (*BeaconCommitteeStateHash, error) {
	if b.isEmpty() {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}

	hashes, err := b.beaconCommitteeStateSlashingBase.Hash()
	if err != nil {
		return nil, err
	}

	shardNextEpochCandidateStr, err := incognitokey.CommitteeKeyListToString(b.shardCommonPool)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	tempShardCandidateHash, err := common.GenerateHashFromStringArray(shardNextEpochCandidateStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	hashes.ShardCandidateHash = tempShardCandidateHash

	syncPool := make(map[byte][]string)
	for shardID, keys := range b.syncPool {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
		syncPool[shardID] = keysStr
	}

	tempSyncPoolHash, err := common.GenerateHashFromMapByteString(syncPool)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	hashes.ShardSyncValidatorsHash = tempSyncPoolHash

	return hashes, nil
}
func (b BeaconCommitteeStateV3) isEmpty() bool {
	return reflect.DeepEqual(b, NewBeaconCommitteeStateV3())
}

func (b *BeaconCommitteeStateV3) GetSyncPool() map[byte][]incognitokey.CommitteePublicKey {
	return b.syncPool
}

//assignToSync assign validatrors to syncPool
// update beacon committee state and committeechange
// UPDATE SYNC POOL ONLY
func (b *BeaconCommitteeStateV3) assignToSync(
	shardID byte,
	candidates []string,
	committeeChange *CommitteeChange,
	beaconHeight uint64,
) *CommitteeChange {
	tempCandidateStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(candidates)
	committeeChange.SyncingPoolAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], tempCandidateStructs...)
	b.syncPool[shardID] = append(b.syncPool[shardID], tempCandidateStructs...)
	return committeeChange
}

//assignToPending assign candidates to pending list
// update beacon state and committeeChange
// UPDATE PENDING LIST ONLY
func (b *BeaconCommitteeStateV3) assignToPending(candidates []string, rand int64, shardID byte, committeeChange *CommitteeChange) *CommitteeChange {
	newCommitteeChange := committeeChange
	for _, candidate := range candidates {
		key := incognitokey.CommitteePublicKey{}
		key.FromString(candidate)
		newCommitteeChange.ShardSubstituteAdded[shardID] = append(newCommitteeChange.ShardSubstituteAdded[shardID], key)
		randomOffset := 0
		if len(b.shardSubstitute[shardID]) != 0 {
			randomOffset = calculateCandidatePosition(candidate, rand, len(b.shardSubstitute[shardID]))
		}
		b.shardSubstitute[shardID] = incognitokey.InsertCommitteePublicKeyToSlice(b.shardSubstitute[shardID], key, randomOffset)
	}
	return newCommitteeChange
}

func (b *BeaconCommitteeStateV3) processAfterNormalSwap(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	newCommitteeChange := committeeChange
	candidates, newCommitteeChange, returnStakingInstruction, err := b.getValidatorsByAutoStake(env, outPublicKeys, newCommitteeChange, returnStakingInstruction, oldState)
	if err != nil {
		return newCommitteeChange, returnStakingInstruction, err
	}
	newReturnStakingInstruction := returnStakingInstruction
	newCommitteeChange = b.assignToPending(candidates, env.RandomNumber, env.ShardID, newCommitteeChange)
	return newCommitteeChange, newReturnStakingInstruction, nil
}

//processAssignWithRandomInstruction assign candidates to syncPool
// update beacon state and committeechange
func (b *BeaconCommitteeStateV3) processAssignWithRandomInstruction(
	rand int64,
	activeShards int,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
	beaconHeight uint64,
) *CommitteeChange {
	newCommitteeChange, candidates := b.updateCandidatesByRandom(committeeChange, oldState)
	assignedCandidates := b.getAssignCandidates(candidates, rand, activeShards, oldState)
	for shardID, candidates := range assignedCandidates {
		newCommitteeChange = b.assignToSync(shardID, candidates, newCommitteeChange, beaconHeight)
	}
	return newCommitteeChange
}

func (b *BeaconCommitteeStateV3) processSwapShardInstruction(
	swapShardInstruction *instruction.SwapShardInstruction,
	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (
	*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	shardID := byte(swapShardInstruction.ChainID)
	env.ShardID = shardID

	// process normal swap out
	newCommitteeChange, _, normalSwapOutCommittees, slashingCommittees, err := b.processNormalSwap(swapShardInstruction, env, committeeChange, oldState)

	// process after swap for assign old committees to current shard pool
	newCommitteeChange, returnStakingInstruction, err = b.processAfterNormalSwap(env,
		normalSwapOutCommittees,
		newCommitteeChange,
		returnStakingInstruction,
		oldState,
	)
	if err != nil {
		return nil, returnStakingInstruction, err
	}

	// process slashing after normal swap out
	returnStakingInstruction, newCommitteeChange, err = b.processSlashing(
		env,
		slashingCommittees,
		returnStakingInstruction,
		newCommitteeChange,
		oldState,
	)
	if err != nil {
		return nil, returnStakingInstruction, err
	}
	newCommitteeChange.SlashingCommittee[shardID] = append(committeeChange.SlashingCommittee[shardID], slashingCommittees...)

	return newCommitteeChange, returnStakingInstruction, nil
}

func (b *BeaconCommitteeStateV3) removeValidatorsFromSyncPool(validators []string, shardID byte) {
	finishedSyncValidators := make(map[string]bool)
	for _, v := range validators {
		finishedSyncValidators[v] = true
	}
	count := 0
	for i := 0; i < len(b.syncPool[shardID]); i++ {
		if count == len(validators) {
			break
		}
		v := b.syncPool[shardID][i]
		key, _ := v.ToBase58()
		if finishedSyncValidators[key] {
			b.syncPool[shardID] = append(b.syncPool[shardID][:i], b.syncPool[shardID][i+1:]...)
			i--
			count++
		}
	}
}

func (b *BeaconCommitteeStateV3) processFinishSyncInstruction(
	finishSyncInstruction *instruction.FinishSyncInstruction,
	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) (
	*CommitteeChange, error) {
	newCommitteeChange := committeeChange
	newCommitteeChange.SyncingPoolRemoved[byte(finishSyncInstruction.ChainID)] =
		append(newCommitteeChange.SyncingPoolRemoved[byte(finishSyncInstruction.ChainID)], finishSyncInstruction.PublicKeysStruct...)
	newCommitteeChange.FinishedSyncValidators[byte(finishSyncInstruction.ChainID)] = append(
		newCommitteeChange.FinishedSyncValidators[byte(finishSyncInstruction.ChainID)],
		finishSyncInstruction.PublicKeys...,
	)
	b.removeValidatorsFromSyncPool(finishSyncInstruction.PublicKeys, byte(finishSyncInstruction.ChainID))

	committeeChange = b.assignToPending(
		finishSyncInstruction.PublicKeys,
		env.RandomNumber,
		byte(finishSyncInstruction.ChainID),
		newCommitteeChange)

	return newCommitteeChange, nil
}

func (b *BeaconCommitteeStateV3) processUnstakeInstruction(
	unstakeInstruction *instruction.UnstakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,

) *CommitteeChange {
	return b.turnOffAutoStake(env.newValidators, unstakeInstruction.CommitteePublicKeys, committeeChange, oldState)
}

func (b *BeaconCommitteeStateV3) AllSyncingValidators() []string {
	res := []string{}
	for _, syncingValidators := range b.syncPool {
		str, err := incognitokey.CommitteeKeyListToString(syncingValidators)
		if err != nil {
			return []string{}
		}
		res = append(res, str...)
	}
	return res
}

func (b BeaconCommitteeStateV3) GetSyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	b.mu.RLock()
	defer b.mu.RUnlock()
	res := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range b.syncPool {
		res[k] = v
	}
	return res
}

func (b *BeaconCommitteeStateV3) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
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
		b.numberOfAssignedCandidates = SnapshotShardCommonPoolV2(
			oldB.shardCommonPool,
			oldB.shardCommittee,
			oldB.shardSubstitute,
			env.NumberOfFixedShardBlockValidator,
			env.MinShardCommitteeSize,
			oldB.swapRule,
		)

		Logger.log.Infof("Block %+v, Number of Snapshot to Assign Candidate %+v", env.BeaconHeight, b.numberOfAssignedCandidates)
	}

	env.newUnassignedCommonPool, _ = incognitokey.CommitteeKeyListToString(b.shardCommonPool[b.numberOfAssignedCandidates:])
	env.newAllSubstituteCommittees, _ = b.getAllSubstituteCommittees()
	env.newValidators = append(env.newUnassignedCommonPool, env.newAllSubstituteCommittees...)
	env.newValidators = append(env.newValidators, b.AllSyncingValidators()...)

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
				randomInstruction.BtcNonce, env.ActiveShards, committeeChange, oldB, env.BeaconHeight)

		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = b.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange, oldB)

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

		case instruction.FINISH_SYNC_ACTION:
			finishSyncInstruction, err := instruction.ValidateAndImportFinishSyncInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, err = b.processFinishSyncInstruction(
				finishSyncInstruction, env, committeeChange, oldB)
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

//SplitReward ...
func (b *BeaconCommitteeStateV3) Process(
	env *BeaconCommitteeStateEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64,
	map[common.Hash]uint64, map[common.Hash]uint64, error,
) {
	devPercent := uint64(env.DAOPercent)
	allCoinTotalReward := env.TotalReward
	rewardForBeacon := map[common.Hash]uint64{}
	rewardForShard := map[common.Hash]uint64{}
	rewardForIncDAO := map[common.Hash]uint64{}
	rewardForCustodian := map[common.Hash]uint64{}
	lenBeaconCommittees := uint64(len(b.GetBeaconCommittee()))
	lenShardCommittees := uint64(len(b.GetShardCommittee()[env.ShardID]))

	if len(allCoinTotalReward) == 0 {
		Logger.log.Info("Beacon Height %+v, 😭 found NO reward", env.BeaconHeight)
		return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
	}

	for key, totalReward := range allCoinTotalReward {
		totalRewardForDAOAndCustodians := devPercent * totalReward / 100
		totalRewardForShardAndBeaconValidators := totalReward - totalRewardForDAOAndCustodians
		shardWeight := float64(lenShardCommittees)
		beaconWeight := 2 * float64(lenBeaconCommittees) / float64(len(b.GetShardCommittee()))
		totalValidatorWeight := shardWeight + beaconWeight

		rewardForShard[key] = uint64(shardWeight * float64(totalRewardForShardAndBeaconValidators) / totalValidatorWeight)
		Logger.log.Infof("totalRewardForDAOAndCustodians tokenID %v - %v\n", key.String(), totalRewardForDAOAndCustodians)

		if env.IsSplitRewardForCustodian {
			rewardForCustodian[key] += env.PercentCustodianReward * totalRewardForDAOAndCustodians / 100
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians - rewardForCustodian[key]
		} else {
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians
		}
		rewardForBeacon[key] += totalReward - (rewardForShard[key] + totalRewardForDAOAndCustodians)
	}

	return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
}
