package committeestate

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

type beaconCommitteeStateBase struct {
	beaconCommittee []incognitokey.CommitteePublicKey
	shardCommittee  map[byte][]incognitokey.CommitteePublicKey
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey

	autoStake      map[string]bool                   // committee public key => true or false
	rewardReceiver map[string]privacy.PaymentAddress // incognito public key => reward receiver payment address
	stakingTx      map[string]common.Hash            // committee public key => reward receiver payment address

	mu *sync.RWMutex // beware of this, any class extend this class need to use this mutex carefully
}

func NewBeaconCommitteeStateBase() *beaconCommitteeStateBase {
	return &beaconCommitteeStateBase{
		shardCommittee:  make(map[byte][]incognitokey.CommitteePublicKey),
		shardSubstitute: make(map[byte][]incognitokey.CommitteePublicKey),
		autoStake:       make(map[string]bool),
		rewardReceiver:  make(map[string]privacy.PaymentAddress),
		stakingTx:       make(map[string]common.Hash),
		mu:              new(sync.RWMutex),
	}
}

func NewBeaconCommitteeStateBaseWithValue(
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
) *beaconCommitteeStateBase {
	return &beaconCommitteeStateBase{
		beaconCommittee: beaconCommittee,
		shardCommittee:  shardCommittee,
		shardSubstitute: shardSubstitute,
		autoStake:       autoStake,
		rewardReceiver:  rewardReceiver,
		stakingTx:       stakingTx,
		mu:              new(sync.RWMutex),
	}
}

func (b *beaconCommitteeStateBase) Reset() {
	b.reset()
}

func (b *beaconCommitteeStateBase) reset() {
	b.beaconCommittee = []incognitokey.CommitteePublicKey{}
	b.shardCommittee = make(map[byte][]incognitokey.CommitteePublicKey)
	b.shardSubstitute = make(map[byte][]incognitokey.CommitteePublicKey)
	b.autoStake = make(map[string]bool)
	b.rewardReceiver = make(map[string]privacy.PaymentAddress)
	b.stakingTx = make(map[string]common.Hash)
}

func (b beaconCommitteeStateBase) clone() *beaconCommitteeStateBase {
	newB := NewBeaconCommitteeStateBase()
	newB.beaconCommittee = make([]incognitokey.CommitteePublicKey, len(b.beaconCommittee))
	copy(newB.beaconCommittee, b.beaconCommittee)

	for i, v := range b.shardCommittee {
		newB.shardCommittee[i] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(newB.shardCommittee[i], v)
	}

	for i, v := range b.shardSubstitute {
		newB.shardSubstitute[i] = make([]incognitokey.CommitteePublicKey, len(v))
		copy(newB.shardSubstitute[i], v)
	}

	for k, v := range b.autoStake {
		newB.autoStake[k] = v
	}

	for k, v := range b.rewardReceiver {
		newB.rewardReceiver[k] = v
	}

	for k, v := range b.stakingTx {
		newB.stakingTx[k] = v
	}

	return newB
}

func (b beaconCommitteeStateBase) Version() int {
	panic("Implement version for committee state")
}

func (b beaconCommitteeStateBase) BeaconCommittee() []incognitokey.CommitteePublicKey {
	return b.beaconCommittee
}

func (b beaconCommitteeStateBase) NumberOfAssignedCandidates() int {
	panic("Implement this function")
}

func (b beaconCommitteeStateBase) SetNumberOfAssignedCandidates(numberOfAssignedCandidates int) {
	panic("Implement this function")
}

func (b beaconCommitteeStateBase) ShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	return b.shardCommittee
}

func (b beaconCommitteeStateBase) ShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	return b.shardSubstitute
}

func (b beaconCommitteeStateBase) CandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	panic("Implement this function")
}

func (b beaconCommitteeStateBase) CandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	panic("Implement this function")
}

func (b beaconCommitteeStateBase) ShardCommonPool() []incognitokey.CommitteePublicKey {
	panic("Implement this function")
}

func (b beaconCommitteeStateBase) Terms() map[string]uint64 {
	panic("Implement this function")
}

func (b beaconCommitteeStateBase) AutoStake() map[string]bool {
	return b.autoStake
}

func (b beaconCommitteeStateBase) RewardReceiver() map[string]privacy.PaymentAddress {
	return b.rewardReceiver
}

func (b beaconCommitteeStateBase) StakingTx() map[string]common.Hash {
	return b.stakingTx
}

func (b beaconCommitteeStateBase) Mu() *sync.RWMutex {
	return b.mu
}

func (b beaconCommitteeStateBase) AllCandidateSubstituteCommittees() []string {
	return b.getAllCandidateSubstituteCommittee()
}

func (b beaconCommitteeStateBase) IsEmpty() bool {
	return reflect.DeepEqual(b, NewBeaconCommitteeStateBase())
}

func (b beaconCommitteeStateBase) Hash() (*BeaconCommitteeStateHash, error) {
	if b.IsEmpty() {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}
	// beacon committee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(b.beaconCommittee)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	validatorArr := append([]string{}, beaconCommitteeStr...)

	tempBeaconCommitteeAndValidatorHash, err := common.GenerateHashFromStringArray(validatorArr)
	// Shard candidate root: shard current candidate + shard next candidate

	// Shard Validator root
	shardPendingValidator := make(map[byte][]string)
	for shardID, keys := range b.shardSubstitute {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
		shardPendingValidator[shardID] = keysStr
	}
	shardCommittee := make(map[byte][]string)
	for shardID, keys := range b.shardCommittee {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
		shardCommittee[shardID] = keysStr
	}
	tempShardCommitteeAndValidatorHash, err := common.GenerateHashFromMapByteString(shardPendingValidator, shardCommittee)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	tempAutoStakingHash, err := common.GenerateHashFromMapStringBool(b.autoStake)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	hashes := &BeaconCommitteeStateHash{
		BeaconCommitteeAndValidatorHash: tempBeaconCommitteeAndValidatorHash,
		ShardCommitteeAndValidatorHash:  tempShardCommitteeAndValidatorHash,
		AutoStakeHash:                   tempAutoStakingHash,
	}
	return hashes, nil
}

func (b *beaconCommitteeStateBase) SetBeaconCommittees(committees []incognitokey.CommitteePublicKey) {
	b.beaconCommittee = []incognitokey.CommitteePublicKey{}
	b.beaconCommittee = append(b.beaconCommittee, committees...)
}

func (b beaconCommitteeStateBase) SwapRule() SwapRule {
	panic("Implement this function")
}

func (b beaconCommitteeStateBase) UnassignedCommonPool() []string {
	return []string{}
}

func (b beaconCommitteeStateBase) AllSubstituteCommittees() []string {
	committees, _ := b.getAllSubstituteCommittees()
	return committees
}

func (b *beaconCommitteeStateBase) getAllCandidateSubstituteCommittee() []string {
	res := []string{}
	for _, committee := range b.shardCommittee {
		shardCommitteeStr, err := incognitokey.CommitteeKeyListToString(committee)
		if err != nil {
			panic(err)
		}
		res = append(res, shardCommitteeStr...)
	}
	for _, substitute := range b.shardSubstitute {
		beaconSubstituteStr, err := incognitokey.CommitteeKeyListToString(substitute)
		if err != nil {
			panic(err)
		}
		res = append(res, beaconSubstituteStr...)
	}
	beaconCommittee := b.beaconCommittee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconCommittee)
	if err != nil {
		panic(err)
	}
	res = append(res, beaconCommitteeStr...)
	return res
}

func (b *beaconCommitteeStateBase) getAllSubstituteCommittees() ([]string, error) {
	validators := []string{}

	for _, committee := range b.shardCommittee {
		committeeStr, err := incognitokey.CommitteeKeyListToString(committee)
		if err != nil {
			return nil, err
		}
		validators = append(validators, committeeStr...)
	}
	for _, substitute := range b.shardSubstitute {
		substituteStr, err := incognitokey.CommitteeKeyListToString(substitute)
		if err != nil {
			return nil, err
		}
		validators = append(validators, substituteStr...)
	}

	beaconCommittee := b.beaconCommittee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconCommittee)
	if err != nil {
		return nil, err
	}
	validators = append(validators, beaconCommitteeStr...)
	return validators, nil
}

func (b *beaconCommitteeStateBase) stopAutoStake(publicKey string, committeeChange *CommitteeChange) *CommitteeChange {
	b.autoStake[publicKey] = false
	committeeChange.StopAutoStake = append(committeeChange.StopAutoStake, publicKey)
	return committeeChange
}

func (b *beaconCommitteeStateBase) processStakeInstruction(
	stakeInstruction *instruction.StakeInstruction,
	committeeChange *CommitteeChange,
) (*CommitteeChange, error) {
	var err error
	for index, candidate := range stakeInstruction.PublicKeyStructs {
		committeePublicKey := stakeInstruction.PublicKeys[index]
		b.rewardReceiver[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
		b.autoStake[committeePublicKey] = stakeInstruction.AutoStakingFlag[index]
		b.stakingTx[committeePublicKey] = stakeInstruction.TxStakeHashes[index]
	}
	committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, stakeInstruction.PublicKeyStructs...)

	return committeeChange, err
}

func (b *beaconCommitteeStateBase) processStopAutoStakeInstruction(
	stopAutoStakeInstruction *instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	//careful with this variable
	validators := env.newAllCandidateSubstituteCommittee
	for _, committeePublicKey := range stopAutoStakeInstruction.CommitteePublicKeys {
		if common.IndexOfStr(committeePublicKey, validators) == -1 {
			// if not found then delete auto staking data for this public key if present
			if _, ok := oldState.AutoStake()[committeePublicKey]; ok {
				delete(b.autoStake, committeePublicKey)
			}
		} else {
			// if found in committee list then turn off auto staking
			if _, ok := oldState.AutoStake()[committeePublicKey]; ok {
				committeeChange = b.stopAutoStake(committeePublicKey, committeeChange)
			}
		}
	}
	return committeeChange
}

func (b *beaconCommitteeStateBase) SyncPool() map[byte][]incognitokey.CommitteePublicKey {
	return map[byte][]incognitokey.CommitteePublicKey{}
}

func SnapshotShardCommonPoolV2(
	shardCommonPool []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	numberOfFixedValidator int,
	minCommitteeSize int,
	swapRule SwapRule,
) (numberOfAssignedCandidates int) {
	for k, v := range shardCommittee {
		assignPerShard := swapRule.AssignOffset(
			len(shardSubstitute[k]),
			len(v),
			numberOfFixedValidator,
			minCommitteeSize,
		)
		numberOfAssignedCandidates += assignPerShard
	}

	if numberOfAssignedCandidates > len(shardCommonPool) {
		numberOfAssignedCandidates = len(shardCommonPool)
	}
	return numberOfAssignedCandidates
}

func (b *beaconCommitteeStateBase) SetSwapRule(swapRule SwapRule) {
	panic("Implement this function")
}
