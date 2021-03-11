package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeStateV3 struct {
	beaconCommitteeStateSlashingBase
	syncPool map[byte][]incognitokey.CommitteePublicKey
}

func NewBeaconCommitteeStateV3() *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
		beaconCommitteeStateSlashingBase: *NewBeaconCommitteeStateSlashingBase(),
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
	swapRule SwapRule,
) *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
		beaconCommitteeStateSlashingBase: *NewBeaconCommitteeStateSlashingBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute, autoStake, rewardReceiver, stakingTx,
			shardCommonPool, numberOfAssignedCandidates, swapRule,
		),
		syncPool: syncPool,
	}
}

func (b *BeaconCommitteeStateV3) cloneFrom(fromB BeaconCommitteeStateV3) {
	b.reset()
	b.beaconCommitteeStateSlashingBase.cloneFrom(fromB.beaconCommitteeStateSlashingBase)

	for i, v := range fromB.syncPool {
		b.syncPool[i] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(b.syncPool[i], v)
	}
}

func (b *BeaconCommitteeStateV3) reset() {
	b.beaconCommitteeStateSlashingBase.reset()
	b.syncPool = map[byte][]incognitokey.CommitteePublicKey{}
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

func (b *BeaconCommitteeStateV3) Version() int {
	return DCS_VERSION
}

func (b *BeaconCommitteeStateV3) SyncPool() map[byte][]incognitokey.CommitteePublicKey {
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

//SplitReward ...
func (b *BeaconCommitteeStateV3) SplitReward(
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

func (b BeaconCommitteeStateV3) SyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	b.mu.RLock()
	defer b.mu.RUnlock()
	res := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range b.syncPool {
		res[k] = v
	}
	return res
}
