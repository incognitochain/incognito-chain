package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

//NewBeaconCommitteeState constructor for BeaconCommitteeState by version
func NewBeaconCommitteeState(
	version int,
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	shardCommonPool []incognitokey.CommitteePublicKey,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceivers map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	syncPool map[byte][]incognitokey.CommitteePublicKey,
	swapRule SwapRuleProcessor,
	nextEpochShardCandidate []incognitokey.CommitteePublicKey,
	currentEpochShardCandidate []incognitokey.CommitteePublicKey,
	beaconHeight uint64,
	beaconBlockHash common.Hash,
) BeaconCommitteeState {
	var committeeState BeaconCommitteeState
	switch version {
	case SELF_SWAP_SHARD_VERSION:
		committeeState = NewBeaconCommitteeStateV1WithValue(
			beaconCommittee,
			[]incognitokey.CommitteePublicKey{},
			nextEpochShardCandidate,
			currentEpochShardCandidate,
			[]incognitokey.CommitteePublicKey{},
			[]incognitokey.CommitteePublicKey{},
			shardCommittee,
			shardSubstitute,
			autoStake,
			rewardReceivers,
			stakingTx,
		)
	case SLASHING_VERSION:
		committeeState = NewBeaconCommitteeStateV2WithValue(
			beaconCommittee,
			shardCommittee,
			shardSubstitute,
			shardCommonPool,
			numberOfAssignedCandidates,
			autoStake,
			rewardReceivers,
			stakingTx,
			swapRule,
		)
	case DCS_VERSION:
		committeeState = NewBeaconCommitteeStateV3WithValue(
			beaconCommittee,
			shardCommittee,
			shardSubstitute,
			shardCommonPool,
			numberOfAssignedCandidates,
			autoStake,
			rewardReceivers,
			stakingTx,
			syncPool,
			swapRule,
		)
	}
	return committeeState
}

//VersionByBeaconHeight get version of committee engine by beaconHeight and config of blockchain
func VersionByBeaconHeight(beaconHeight, consensusV3Height, stakingV3Height uint64) int {
	if beaconHeight >= stakingV3Height {
		return DCS_VERSION
	}
	if beaconHeight >= consensusV3Height {
		return SLASHING_VERSION
	}
	return SELF_SWAP_SHARD_VERSION
}

type beaconCommitteeStateBase struct {
	beaconCommittee []incognitokey.CommitteePublicKey
	shardCommittee  map[byte][]incognitokey.CommitteePublicKey
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey

	autoStake      map[string]bool                   // committee public key => true or false
	rewardReceiver map[string]privacy.PaymentAddress // incognito public key => reward receiver payment address
	stakingTx      map[string]common.Hash            // committee public key => reward receiver payment address

	mu *sync.RWMutex // beware of this, any class extend this class need to use this mutex carefully
}

func newBeaconCommitteeStateBase() *beaconCommitteeStateBase {
	return &beaconCommitteeStateBase{
		shardCommittee:  make(map[byte][]incognitokey.CommitteePublicKey),
		shardSubstitute: make(map[byte][]incognitokey.CommitteePublicKey),
		autoStake:       make(map[string]bool),
		rewardReceiver:  make(map[string]privacy.PaymentAddress),
		stakingTx:       make(map[string]common.Hash),
		mu:              new(sync.RWMutex),
	}
}

func newBeaconCommitteeStateBaseWithValue(
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

func (b beaconCommitteeStateBase) isEmpty() bool {
	return reflect.DeepEqual(b, newBeaconCommitteeStateBase())
}

//Clone:
func (b beaconCommitteeStateBase) Clone() BeaconCommitteeState {
	return b.clone()
}

func (b beaconCommitteeStateBase) clone() *beaconCommitteeStateBase {
	newB := newBeaconCommitteeStateBase()
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
	panic("do not use function of beaconCommitteeStateBase struct")
}

func (b beaconCommitteeStateBase) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	return b.beaconCommittee
}

func (b beaconCommitteeStateBase) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

func (b beaconCommitteeStateBase) GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return b.shardCommittee[shardID]
}

func (b beaconCommitteeStateBase) GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	return b.shardCommittee
}

func (b beaconCommitteeStateBase) GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	return b.shardSubstitute[shardID]
}

func (b beaconCommitteeStateBase) GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	return b.shardSubstitute
}

func (b beaconCommitteeStateBase) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	panic("do not use function of beaconCommitteeStateBase struct")
}

func (b beaconCommitteeStateBase) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	panic("do not use function of beaconCommitteeStateBase struct")
}

func (b beaconCommitteeStateBase) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}
func (b beaconCommitteeStateBase) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

func (b beaconCommitteeStateBase) GetShardCommonPool() []incognitokey.CommitteePublicKey {
	panic("do not use function of beaconCommitteeStateBase struct")
}

func (b beaconCommitteeStateBase) GetAutoStaking() map[string]bool {
	return b.autoStake
}

func (b beaconCommitteeStateBase) GetRewardReceiver() map[string]privacy.PaymentAddress {
	return b.rewardReceiver
}

func (b beaconCommitteeStateBase) GetStakingTx() map[string]common.Hash {
	return b.stakingTx
}

func (b beaconCommitteeStateBase) GetAllCandidateSubstituteCommittee() []string {
	return b.getAllCandidateSubstituteCommittee()
}

func (b beaconCommitteeStateBase) GetSyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	panic("do not use function of beaconCommitteeStateBase struct")
}

func (b *beaconCommitteeStateBase) ActiveShards() int {
	return len(b.shardCommittee)
}

//IsSwapTime is this the moment for process a swbap action
func (b beaconCommitteeStateBase) IsSwapTime(beaconHeight, numberBlocksEachEpoch uint64) bool {
	panic("Implement this function")
}

func (b beaconCommitteeStateBase) Hash() (*BeaconCommitteeStateHash, error) {
	if b.isEmpty() {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}
	// beacon committee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(b.beaconCommittee)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	validatorArr := append([]string{}, beaconCommitteeStr...)
	// beacon committee
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
	tempShardCommitteeAndValidatorHash, err := common.GenerateHashFromTwoMapByteString(shardPendingValidator, shardCommittee)
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

func (b *beaconCommitteeStateBase) InitCommitteeState(env *BeaconCommitteeStateEnvironment) {
	b.mu.Lock()
	defer b.mu.Unlock()
	newBeaconCandidates := []incognitokey.CommitteePublicKey{}
	newShardCandidates := []incognitokey.CommitteePublicKey{}
	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		if inst[0] == instruction.STAKE_ACTION {
			stakeInstruction := instruction.ImportInitStakeInstructionFromString(inst)
			for index, candidate := range stakeInstruction.PublicKeyStructs {
				b.GetRewardReceiver()[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
				b.GetAutoStaking()[stakeInstruction.PublicKeys[index]] = stakeInstruction.AutoStakingFlag[index]
				b.GetStakingTx()[stakeInstruction.PublicKeys[index]] = stakeInstruction.TxStakeHashes[index]
			}
			if stakeInstruction.Chain == instruction.BEACON_INST {
				newBeaconCandidates = append(newBeaconCandidates, stakeInstruction.PublicKeyStructs...)
			} else {
				newShardCandidates = append(newShardCandidates, stakeInstruction.PublicKeyStructs...)
			}
			err := statedb.StoreStakerInfo(
				env.ConsensusStateDB,
				stakeInstruction.PublicKeyStructs,
				b.GetRewardReceiver(),
				b.GetAutoStaking(),
				b.GetStakingTx(),
			)
			if err != nil {
				panic(err)
			}
		}
	}
	b.beaconCommittee = []incognitokey.CommitteePublicKey{}
	b.beaconCommittee = append(b.beaconCommittee, newBeaconCandidates...)
	for shardID := 0; shardID < env.ActiveShards; shardID++ {
		b.GetShardCommittee()[byte(shardID)] = append(b.GetShardCommittee()[byte(shardID)], newShardCandidates[shardID*env.MinShardCommitteeSize:(shardID+1)*env.MinShardCommitteeSize]...)
	}
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

func (b *beaconCommitteeStateBase) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash,
	*CommitteeChange,
	[][]string,
	error) {
	return nil, nil, [][]string{}, nil
}

func (b *beaconCommitteeStateBase) turnOffStopAutoStake(publicKey string, committeeChange *CommitteeChange) *CommitteeChange {
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

func (b *beaconCommitteeStateBase) turnOffAutoStake(
	validators, stopAutoStakeKeys []string,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	for _, committeePublicKey := range stopAutoStakeKeys {
		if common.IndexOfStr(committeePublicKey, validators) == -1 {
			// if not found then delete auto staking data for this public key if present
			if _, ok := oldState.GetAutoStaking()[committeePublicKey]; ok {
				delete(b.autoStake, committeePublicKey)
			}
		} else {
			// if found in committee list then turn off auto staking
			if autoStake, ok := oldState.GetAutoStaking()[committeePublicKey]; ok {
				if autoStake {
					committeeChange = b.turnOffStopAutoStake(committeePublicKey, committeeChange)
				}
			}
		}
	}
	return committeeChange
}

func (b *beaconCommitteeStateBase) processStopAutoStakeInstruction(
	stopAutoStakeInstruction *instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	return b.turnOffAutoStake(env.newValidators, stopAutoStakeInstruction.CommitteePublicKeys, committeeChange, oldState)
}

func (b *beaconCommitteeStateBase) GetSyncPool() map[byte][]incognitokey.CommitteePublicKey {
	panic("do not use function of beaconCommitteeStateBase struct")
}

//SplitReward ...
func (b *beaconCommitteeStateBase) Process(
	env *BeaconCommitteeStateEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64,
	map[common.Hash]uint64, map[common.Hash]uint64, error,
) {
	panic("do not use function of beaconCommitteeStateBase struct")
}

func SnapshotShardCommonPoolV2(
	shardCommonPool []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	numberOfFixedValidator int,
	minCommitteeSize int,
	swapRule SwapRuleProcessor,
) (numberOfAssignedCandidates int) {
	for k, v := range shardCommittee {
		assignPerShard := swapRule.CalculateAssignOffset(
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

//Upgrade upgrade committee engine by version
func (b beaconCommitteeStateBase) Upgrade(env *BeaconCommitteeStateEnvironment) BeaconCommitteeState {
	panic("Implement this function")
}
