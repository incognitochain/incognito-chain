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
	terms    map[string]uint64
}

func NewBeaconCommitteeStateV3() *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
		beaconCommitteeStateSlashingBase: *NewBeaconCommitteeStateSlashingBase(),
		syncPool:                         make(map[byte][]incognitokey.CommitteePublicKey),
		terms:                            make(map[string]uint64),
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
	terms map[string]uint64,
	swapRule SwapRule,
) *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
		beaconCommitteeStateSlashingBase: *NewBeaconCommitteeStateSlashingBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute, autoStake, rewardReceiver, stakingTx,
			shardCommonPool, numberOfAssignedCandidates, swapRule,
		),
		syncPool: syncPool,
		terms:    terms,
	}
}

func (b *BeaconCommitteeStateV3) cloneFrom(fromB BeaconCommitteeStateV3) {
	b.reset()
	b.beaconCommitteeStateSlashingBase.cloneFrom(fromB.beaconCommitteeStateSlashingBase)

	for i, v := range fromB.syncPool {
		b.syncPool[i] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(b.syncPool[i], v)
	}

	for i, v := range fromB.terms {
		b.terms[i] = v
	}
}

func (b *BeaconCommitteeStateV3) reset() {
	b.beaconCommitteeStateSlashingBase.reset()
	b.syncPool = map[byte][]incognitokey.CommitteePublicKey{}
	b.terms = map[string]uint64{}
}

func (b *BeaconCommitteeStateV3) clone() *BeaconCommitteeStateV3 {
	newB := NewBeaconCommitteeStateV3()
	newB.beaconCommitteeStateSlashingBase = *b.beaconCommitteeStateSlashingBase.clone()

	for i, v := range b.syncPool {
		newB.syncPool[i] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(newB.syncPool[i], v)
	}

	for i, v := range b.terms {
		newB.terms[i] = v
	}

	return newB
}

func (b *BeaconCommitteeStateV3) Version() int {
	return DCS_VERSION
}

func (b *BeaconCommitteeStateV3) SyncPool() map[byte][]incognitokey.CommitteePublicKey {
	return b.syncPool
}

func (b *BeaconCommitteeStateV3) assignToSync(
	shardID byte,
	candidates []string,
	committeeChange *CommitteeChange) *CommitteeChange {
	tempCandidateStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(candidates)
	committeeChange.ShardSyncingAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], tempCandidateStructs...)
	b.syncPool[shardID] = append(b.syncPool[shardID], tempCandidateStructs...)

	return committeeChange
}

func (b *BeaconCommitteeStateV3) assignAfterNormalSwapOut(
	candidates []string, rand int64, activeShards int, committeeChange *CommitteeChange,
	oldState BeaconCommitteeState, oldShardID byte,
) *CommitteeChange {
	newCommitteeChange := committeeChange
	assignedCandidates := b.getAssignCandidates(candidates, rand, activeShards, oldState)

	for shardID, tempCandidates := range assignedCandidates {
		if shardID == oldShardID {
			newCommitteeChange = b.assignToPending(tempCandidates, rand, shardID, newCommitteeChange)
		} else {
			newCommitteeChange = b.assignToSync(shardID, tempCandidates, newCommitteeChange)
		}
	}
	return newCommitteeChange
}

func (b *BeaconCommitteeStateV3) assignToPending(candidates []string, rand int64, shardID byte, committeeChange *CommitteeChange) *CommitteeChange {
	newCommitteeChange := committeeChange

	for _, candidate := range candidates {
		key := incognitokey.CommitteePublicKey{}
		key.FromString(candidate)
		newCommitteeChange.ShardSubstituteAdded[shardID] = append(newCommitteeChange.ShardSubstituteAdded[shardID], key)
		randomOffset := calculateCandidatePosition(candidate, rand, len(b.shardSubstitute[shardID]))
		b.shardSubstitute[shardID] = incognitokey.InsertCommitteePublicKeyToSlice(b.shardSubstitute[shardID], key, randomOffset)
	}

	return newCommitteeChange
}

func (b *BeaconCommitteeStateV3) processAssignInstruction(
	assignInstruction *instruction.AssignInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (
	*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	newCommitteeChange := committeeChange
	newCommitteeChange.ShardSyncingRemoved[byte(assignInstruction.ChainID)] =
		append(newCommitteeChange.ShardSyncingRemoved[byte(assignInstruction.ChainID)], assignInstruction.ShardCandidatesStruct...)
	b.syncPool[byte(assignInstruction.ChainID)] = b.syncPool[byte(assignInstruction.ChainID)][len(assignInstruction.ShardCandidates):]

	candidates, newCommitteeChange, returnStakingInstruction, err := b.getValidatorsByAutoStake(env, assignInstruction.ShardCandidates, newCommitteeChange, returnStakingInstruction)
	if err != nil {
		return newCommitteeChange, returnStakingInstruction, err
	}

	committeeChange = b.assignToPending(
		candidates,
		env.RandomNumber,
		byte(assignInstruction.ChainID),
		newCommitteeChange)

	return newCommitteeChange, returnStakingInstruction, nil
}

func (b *BeaconCommitteeStateV3) processAfterNormalSwap(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	newCommitteeChange := committeeChange

	candidates, newCommitteeChange, returnStakingInstruction, err := b.getValidatorsByAutoStake(env, outPublicKeys, newCommitteeChange, returnStakingInstruction)
	if err != nil {
		return newCommitteeChange, returnStakingInstruction, err
	}
	newReturnStakingInstruction := returnStakingInstruction

	for i := 0; i < len(candidates); i++ {
		candidate := candidates[i]
		key := incognitokey.CommitteePublicKey{}
		err := key.FromBase58(candidate)
		if err != nil {
			return newCommitteeChange, returnStakingInstruction, err
		}
		if env.BeaconHeight-b.Terms()[candidate]-committeeTerm < 0 {
			newCommitteeChange.ShardSubstituteAdded[env.ShardID] = append(newCommitteeChange.ShardSubstituteAdded[env.ShardID], key)
			newCommitteeChange.ShardCommitteeRemoved[env.ShardID] = append(newCommitteeChange.ShardCommitteeRemoved[env.ShardID], key)
			candidates = append(candidates[:i], candidates[i+1:]...)
		}
	}
	newCommitteeChange = b.assignAfterNormalSwapOut(candidates, env.RandomNumber, env.ActiveShards, newCommitteeChange, oldState, env.ShardID)

	return newCommitteeChange, newReturnStakingInstruction, nil
}

func (b *BeaconCommitteeStateV3) processAssignWithRandomInstruction(
	rand int64,
	activeShards int,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	newCommitteeChange, candidates := b.updateCandidatesByRandom(committeeChange, oldState)
	assignedCandidates := b.getAssignCandidates(candidates, rand, activeShards, oldState)
	for shardID, candidates := range assignedCandidates {
		newCommitteeChange = b.assignToSync(shardID, candidates, newCommitteeChange)
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
	)
	if err != nil {
		return nil, returnStakingInstruction, err
	}
	newCommitteeChange.SlashingCommittee[shardID] = append(committeeChange.SlashingCommittee[shardID], slashingCommittees...)

	return newCommitteeChange, returnStakingInstruction, nil
}
