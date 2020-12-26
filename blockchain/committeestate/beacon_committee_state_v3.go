package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeStateV3 struct {
	beaconCommitteeStateBase
	syncPool map[byte][]incognitokey.CommitteePublicKey
	terms    map[string]uint64
}

func NewBeaconCommitteeStateV3() *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
		beaconCommitteeStateBase: *NewBeaconCommitteeStateBase(),
		syncPool:                 make(map[byte][]incognitokey.CommitteePublicKey),
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
		beaconCommitteeStateBase: *NewBeaconCommitteeStateBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute, shardCommonPool,
			numberOfAssignedCandidates, autoStake, rewardReceiver, stakingTx, swapRule,
		),
		syncPool: syncPool,
	}
}

func (b *BeaconCommitteeStateV3) clone() *BeaconCommitteeStateV3 {
	newB := NewBeaconCommitteeStateV3()

	newB.beaconCommitteeStateBase = *b.beaconCommitteeStateBase.clone()

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

func (b *BeaconCommitteeStateV3) assignShardWithRandomNumber(candidates []string, rand int64, lenSubstitute, lenCommittees int, committeeChange *CommitteeChange) *CommitteeChange {
	newCommitteeChange := committeeChange

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
	newReturnStakingInstruction := returnStakingInstruction

	newCommitteeChange.ShardSyncingRemoved[byte(assignInstruction.ChainID)] =
		append(newCommitteeChange.ShardSyncingRemoved[byte(assignInstruction.ChainID)], assignInstruction.ShardCandidatesStruct...)
	candidates, newCommitteeChange, returnStakingInstruction, err := b.getValidatorsByAutoStake(env, assignInstruction.ShardCandidates, newCommitteeChange, newReturnStakingInstruction)
	if err != nil {
		return committeeChange, returnStakingInstruction, err
	}
	newCommitteeChange.RemovedStaker = append(newCommitteeChange.RemovedStaker, newReturnStakingInstruction.PublicKeys...)
	committeeChange = b.assignShardWithRandomNumber(
		candidates,
		env.RandomNumber,
		len(b.shardSubstitute[byte(assignInstruction.ChainID)]),
		len(b.shardCommittee[byte(assignInstruction.ChainID)]),
		newCommitteeChange)

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
	newReturnStakingInstruction := returnStakingInstruction

	candidates, newCommitteeChange, returnStakingInstruction, err := b.getValidatorsByAutoStake(env, outPublicKeys, newCommitteeChange, returnStakingInstruction)
	if err != nil {
		return newCommitteeChange, returnStakingInstruction, err
	}
	newCommitteeChange.RemovedStaker = append(newCommitteeChange.RemovedStaker, newReturnStakingInstruction.PublicKeys...)

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

	return newCommitteeChange, returnStakingInstruction, nil
}
