package committeestate

import (
	"fmt"
	"math/big"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

type beaconCommitteeStateBase struct {
	beaconCommittee []string
	shardCommittee  map[byte][]string
	shardSubstitute map[byte][]string

	autoStake      map[string]bool                   // committee public key => true or false
	rewardReceiver map[string]privacy.PaymentAddress // incognito public key => reward receiver payment address
	stakingTx      map[string]common.Hash            // committee public key => reward receiver payment address

	mu *sync.RWMutex // beware of this, any class extend this class need to use this mutex carefully
}

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
) BeaconCommitteeState {
	var committeeState BeaconCommitteeState
	tempBeaconCommittee, _ := incognitokey.CommitteeKeyListToString(beaconCommittee)
	tempNextEpochShardCandidate, _ := incognitokey.CommitteeKeyListToString(nextEpochShardCandidate)
	tempCurrentEpochShardCandidate, _ := incognitokey.CommitteeKeyListToString(currentEpochShardCandidate)
	tempShardCommonPool, _ := incognitokey.CommitteeKeyListToString(shardCommonPool)
	tempShardCommittee := make(map[byte][]string)
	tempShardSubstitute := make(map[byte][]string)
	tempSyncPool := make(map[byte][]string)
	for shardID, v := range shardCommittee {
		tempShardCommittee[shardID], _ = incognitokey.CommitteeKeyListToString(v)
	}
	for shardID, v := range shardSubstitute {
		tempShardSubstitute[shardID], _ = incognitokey.CommitteeKeyListToString(v)
	}
	for shardID, v := range syncPool {
		tempSyncPool[shardID], _ = incognitokey.CommitteeKeyListToString(v)
	}
	switch version {
	case SELF_SWAP_SHARD_VERSION:
		committeeState = NewBeaconCommitteeStateV1WithValue(
			tempBeaconCommittee,
			tempNextEpochShardCandidate,
			tempCurrentEpochShardCandidate,
			tempShardCommittee,
			tempShardSubstitute,
			autoStake,
			rewardReceivers,
			stakingTx,
		)
	case SLASHING_VERSION:
		committeeState = NewBeaconCommitteeStateV2WithValue(
			tempBeaconCommittee,
			tempShardCommittee,
			tempShardSubstitute,
			tempShardCommonPool,
			numberOfAssignedCandidates,
			autoStake,
			rewardReceivers,
			stakingTx,
			swapRule,
		)
	case DCS_VERSION:
		committeeState = NewBeaconCommitteeStateV3WithValue(
			tempBeaconCommittee,
			tempShardCommittee,
			tempShardSubstitute,
			tempShardCommonPool,
			numberOfAssignedCandidates,
			autoStake,
			rewardReceivers,
			stakingTx,
			tempSyncPool,
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

func newBeaconCommitteeStateBase() *beaconCommitteeStateBase {
	return &beaconCommitteeStateBase{
		shardCommittee:  make(map[byte][]string),
		shardSubstitute: make(map[byte][]string),
		autoStake:       make(map[string]bool),
		rewardReceiver:  make(map[string]privacy.PaymentAddress),
		stakingTx:       make(map[string]common.Hash),
		mu:              new(sync.RWMutex),
	}
}

func newBeaconCommitteeStateBaseWithValue(
	beaconCommittee []string,
	shardCommittee map[byte][]string,
	shardSubstitute map[byte][]string,
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
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.clone()
}

func (b beaconCommitteeStateBase) clone() *beaconCommitteeStateBase {
	newB := newBeaconCommitteeStateBase()
	newB.beaconCommittee = common.DeepCopyString(b.beaconCommittee)

	for i, v := range b.shardCommittee {
		newB.shardCommittee[i] = common.DeepCopyString(v)
	}

	for i, v := range b.shardSubstitute {
		newB.shardSubstitute[i] = common.DeepCopyString(v)
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

func (b *beaconCommitteeStateBase) reset() {
	b.beaconCommittee = []string{}
	b.shardCommittee = make(map[byte][]string)
	b.shardSubstitute = make(map[byte][]string)
	b.autoStake = make(map[string]bool)
	b.rewardReceiver = make(map[string]privacy.PaymentAddress)
	b.stakingTx = make(map[string]common.Hash)
}

func (b beaconCommitteeStateBase) Version() int {
	panic("do not use function of beaconCommitteeStateBase struct")
}

func (b beaconCommitteeStateBase) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	return b.getBeaconCommittee()
}

func (b beaconCommitteeStateBase) getBeaconCommittee() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(b.beaconCommittee)
	return res
}

func (b beaconCommitteeStateBase) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

func (b beaconCommitteeStateBase) GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	b.mu.RLock()
	defer b.mu.RUnlock()
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(b.shardCommittee[shardID])
	return res
}

func (b beaconCommitteeStateBase) GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.getShardCommittee()
}

func (b beaconCommitteeStateBase) getShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
	for shardID, committees := range b.shardCommittee {
		shardCommittee[shardID], _ = incognitokey.CommitteeBase58KeyListToStruct(committees)
	}
	return shardCommittee
}

func (b beaconCommitteeStateBase) GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	b.mu.RLock()
	defer b.mu.RUnlock()
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(b.shardSubstitute[shardID])
	return res
}

func (b beaconCommitteeStateBase) GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.getShardSubstitute()
}

func (b beaconCommitteeStateBase) getShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	shardSubstitute := make(map[byte][]incognitokey.CommitteePublicKey)
	for shardID, substitute := range b.shardSubstitute {
		shardSubstitute[shardID], _ = incognitokey.CommitteeBase58KeyListToStruct(substitute)
	}
	return shardSubstitute
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
	b.mu.RLock()
	defer b.mu.RUnlock()
	res := make(map[string]bool)
	for k, v := range b.autoStake {
		res[k] = v
	}
	return res
}

func (b beaconCommitteeStateBase) GetRewardReceiver() map[string]privacy.PaymentAddress {
	b.mu.RLock()
	defer b.mu.RUnlock()
	res := make(map[string]privacy.PaymentAddress)
	for k, v := range b.rewardReceiver {
		res[k] = v
	}
	return res
}

func (b beaconCommitteeStateBase) GetStakingTx() map[string]common.Hash {
	b.mu.RLock()
	defer b.mu.RUnlock()
	res := make(map[string]common.Hash)
	for k, v := range b.stakingTx {
		res[k] = v
	}
	return res
}

func (b beaconCommitteeStateBase) GetAllCandidateSubstituteCommittee() []string {
	return b.getAllCandidateSubstituteCommittee()
}

func (b beaconCommitteeStateBase) GetSyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	return make(map[byte][]incognitokey.CommitteePublicKey)
}

func (b *beaconCommitteeStateBase) GetNumberOfActiveShards() int {
	return len(b.shardCommittee)
}

func (b beaconCommitteeStateBase) Hash() (*BeaconCommitteeStateHash, error) {
	if b.isEmpty() {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}
	tempBeaconCommitteeAndValidatorHash, err := common.GenerateHashFromStringArray(b.beaconCommittee)
	// Shard candidate root: shard current candidate + shard next candidate

	// Shard Validator root
	shardPendingValidator := make(map[byte][]string)
	for shardID, keys := range b.shardSubstitute {
		shardPendingValidator[shardID] = keys
	}
	tempShardCommitteeAndValidatorHash, err := common.GenerateHashFromTwoMapByteString(shardPendingValidator, b.shardCommittee)
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

func (b *beaconCommitteeStateBase) initCommitteeState(env *BeaconCommitteeStateEnvironment) {
	b.mu.Lock()
	defer b.mu.Unlock()
	newBeaconCandidates := []string{}
	newShardCandidates := []string{}
	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		if inst[0] == instruction.STAKE_ACTION {
			stakeInstruction := instruction.ImportInitStakeInstructionFromString(inst)
			for index, candidate := range stakeInstruction.PublicKeyStructs {
				b.rewardReceiver[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
				b.autoStake[stakeInstruction.PublicKeys[index]] = stakeInstruction.AutoStakingFlag[index]
				b.stakingTx[stakeInstruction.PublicKeys[index]] = stakeInstruction.TxStakeHashes[index]
			}
			if stakeInstruction.Chain == instruction.BEACON_INST {
				newBeaconCandidates = append(newBeaconCandidates, stakeInstruction.PublicKeys...)
			} else {
				newShardCandidates = append(newShardCandidates, stakeInstruction.PublicKeys...)
			}
			err := statedb.StoreStakerInfo(
				env.ConsensusStateDB,
				stakeInstruction.PublicKeyStructs,
				b.rewardReceiver,
				b.autoStake,
				b.stakingTx,
			)
			if err != nil {
				panic(err)
			}
		}
	}
	b.beaconCommittee = []string{}
	b.beaconCommittee = append(b.beaconCommittee, newBeaconCandidates...)
	Logger.log.Info("[dcs] newShardCandidates:", newShardCandidates)
	for shardID := 0; shardID < env.ActiveShards; shardID++ {
		Logger.log.Info("[dcs] env.MinShardCommitteeSize:", env.MinShardCommitteeSize)
		b.shardCommittee[byte(shardID)] = append(
			b.shardCommittee[byte(shardID)],
			newShardCandidates[shardID*env.MinShardCommitteeSize:(shardID+1)*env.MinShardCommitteeSize]...,
		)
	}
}

func (b *beaconCommitteeStateBase) getAllCandidateSubstituteCommittee() []string {
	res := []string{}
	for _, committee := range b.shardCommittee {
		res = append(res, committee...)
	}
	for _, substitute := range b.shardSubstitute {
		res = append(res, substitute...)
	}
	res = append(res, b.beaconCommittee...)
	return res
}

func (b *beaconCommitteeStateBase) getAllSubstituteCommittees() ([]string, error) {
	res := []string{}

	for _, committee := range b.shardCommittee {
		res = append(res, committee...)
	}

	for _, substitute := range b.shardSubstitute {
		res = append(res, substitute...)
	}
	res = append(res, b.beaconCommittee...)
	return res, nil
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
	committeeChange.AddStopAutoStake(publicKey)
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
) *CommitteeChange {
	for _, committeePublicKey := range stopAutoStakeKeys {
		if common.IndexOfStr(committeePublicKey, validators) == -1 {
			// if not found then delete auto staking data for this public key if present
			if _, ok := b.autoStake[committeePublicKey]; ok {
				delete(b.autoStake, committeePublicKey)
			}
		} else {
			// if found in committee list then turn off auto staking
			if autoStake, ok := b.autoStake[committeePublicKey]; ok {
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
) *CommitteeChange {
	return b.turnOffAutoStake(env.newAllRoles, stopAutoStakeInstruction.CommitteePublicKeys, committeeChange)
}

//SplitReward ...
func (b *beaconCommitteeStateBase) SplitReward(
	env *SplitRewardEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64,
	map[common.Hash]uint64, map[common.Hash]uint64, error,
) {
	panic("do not use function of beaconCommitteeStateBase struct")
}

func SnapshotShardCommonPoolV2(
	shardCommonPool []string,
	shardCommittee map[byte][]string,
	shardSubstitute map[byte][]string,
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

func (b beaconCommitteeStateBase) GenerateRandomInstructions(env *BeaconCommitteeStateEnvironment) (*instruction.RandomInstruction, int64) {
	res := []byte{}
	bestBeaconBlockHash := env.BeaconHash
	res = append(res, bestBeaconBlockHash.Bytes()...)
	for i := 0; i < env.ActiveShards; i++ {
		shardID := byte(i)
		bestShardBlockHash := env.BestShardHash[shardID]
		res = append(res, bestShardBlockHash.Bytes()...)
	}

	bigInt := new(big.Int)
	bigInt = bigInt.SetBytes(res)
	randomNumber := int64(bigInt.Uint64())
	randomInstruction := instruction.NewRandomInstructionWithValue(
		randomNumber,
	)
	return randomInstruction, randomNumber
}
