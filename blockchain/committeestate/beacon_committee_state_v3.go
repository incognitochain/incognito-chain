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

func (b *BeaconCommitteeStateV3) assign(
	candidates []string, rand int64, activeShards int, committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	assignedCandidates := b.getAssignCandidates(candidates, rand, activeShards, oldState)
	for shardID, tempCandidates := range assignedCandidates {
		tempCandidateStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(tempCandidates)
		committeeChange.ShardSyncingAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], tempCandidateStructs...)
		b.syncPool[shardID] = append(b.syncPool[shardID], tempCandidateStructs...)
	}
	return committeeChange
}

func (b *BeaconCommitteeStateV3) assignAfterNormalSwapOut(
	candidates []string, rand int64, activeShards int, committeeChange *CommitteeChange,
	oldState BeaconCommitteeState, oldShardID byte,
) *CommitteeChange {
	newCommitteeChange := committeeChange
	assignedCandidates := b.getAssignCandidates(candidates, rand, activeShards, oldState)

	for shardID, tempCandidates := range assignedCandidates {
		tempCandidateStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(tempCandidates)
		if shardID == oldShardID {
			committeeChange.ShardSubstituteAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], tempCandidateStructs...)
			b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], tempCandidateStructs...)
		} else {
			committeeChange.ShardSyncingAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], tempCandidateStructs...)
			b.syncPool[shardID] = append(b.syncPool[shardID], tempCandidateStructs...)
		}
	}
	return newCommitteeChange
}

//TODO: @tin reimplement here
func (b *BeaconCommitteeStateV3) assignToShard(candidates []string, rand int64, shardID byte, committeeChange *CommitteeChange) *CommitteeChange {
	newCommitteeChange := committeeChange

	for _, candidate := range candidates {
		key := incognitokey.CommitteePublicKey{}
		key.FromString(candidate)
		newCommitteeChange.ShardSubstituteAdded[shardID] = append(newCommitteeChange.ShardSubstituteAdded[shardID], key)
		randomPosition := 0
		for randomPosition >= len(b.shardCommittee[shardID]) {
			randomPosition = calculateCandidatePosition(candidate, rand, len(b.shardCommittee[shardID])+len(b.shardSubstitute[shardID]))
		}
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
	candidates, newCommitteeChange, returnStakingInstruction, err := b.getValidatorsByAutoStake(env, assignInstruction.ShardCandidates, newCommitteeChange, returnStakingInstruction)
	if err != nil {
		return committeeChange, returnStakingInstruction, err
	}
	newReturnStakingInstruction := returnStakingInstruction
	committeeChange = b.assignToShard(
		candidates,
		env.RandomNumber,
		byte(assignInstruction.ChainID),
		newCommitteeChange)

	b.syncPool[byte(assignInstruction.ChainID)] = b.syncPool[byte(assignInstruction.ChainID)][len(assignInstruction.ShardCandidates):]
	return newCommitteeChange, newReturnStakingInstruction, nil
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
