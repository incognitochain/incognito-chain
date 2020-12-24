package committeestate

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type beaconCommitteeStateBase struct {
	beaconCommittee            []incognitokey.CommitteePublicKey
	shardCommittee             map[byte][]incognitokey.CommitteePublicKey
	shardSubstitute            map[byte][]incognitokey.CommitteePublicKey
	shardCommonPool            []incognitokey.CommitteePublicKey
	probationPool              map[string]signaturecounter.Penalty
	numberOfAssignedCandidates int

	autoStake      map[string]bool                   // committee public key => true or false
	rewardReceiver map[string]privacy.PaymentAddress // incognito public key => reward receiver payment address
	stakingTx      map[string]common.Hash            // committee public key => reward receiver payment address

	swapRule    SwapRule
	unstakeRule UnstakeRule
	//assignRule        AssignRule
	stopAutoStakeRule StopAutoStakeRule
	//randomRule        RandomRule

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
	shardCommonPool []incognitokey.CommitteePublicKey,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	swapRule SwapRule,
) *beaconCommitteeStateBase {
	return &beaconCommitteeStateBase{
		beaconCommittee:            beaconCommittee,
		shardCommittee:             shardCommittee,
		shardSubstitute:            shardSubstitute,
		shardCommonPool:            shardCommonPool,
		numberOfAssignedCandidates: numberOfAssignedCandidates,
		autoStake:                  autoStake,
		rewardReceiver:             rewardReceiver,
		stakingTx:                  stakingTx,
		mu:                         new(sync.RWMutex),
	}
}

func (b *beaconCommitteeStateBase) Reset() {
	b.reset()
}

func (b *beaconCommitteeStateBase) reset() {
	b.beaconCommittee = []incognitokey.CommitteePublicKey{}
	b.shardCommonPool = []incognitokey.CommitteePublicKey{}
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
	newB.numberOfAssignedCandidates = b.numberOfAssignedCandidates
	newB.shardCommonPool = make([]incognitokey.CommitteePublicKey, len(b.shardCommonPool))
	copy(newB.shardCommonPool, b.shardCommonPool)

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

	newB.swapRule = cloneSwapRuleByVersion(b.swapRule)

	return newB
}

func (b beaconCommitteeStateBase) Version() int {
	panic("Implement version for committee state")
}

func (b beaconCommitteeStateBase) BeaconCommittee() []incognitokey.CommitteePublicKey {
	return b.beaconCommittee
}

func (b beaconCommitteeStateBase) ShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	return b.shardCommittee
}

func (b beaconCommitteeStateBase) ShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	return b.shardSubstitute
}

func (b beaconCommitteeStateBase) ShardCommonPool() []incognitokey.CommitteePublicKey {
	return b.shardCommonPool
}

func (b beaconCommitteeStateBase) PropationPool() map[string]signaturecounter.Penalty {
	return b.probationPool
}

func (b beaconCommitteeStateBase) NumberOfAssignedCandidates() int {
	return b.numberOfAssignedCandidates
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
	shardNextEpochCandidateStr, err := incognitokey.CommitteeKeyListToString(b.shardCommonPool)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	tempShardCandidateHash, err := common.GenerateHashFromStringArray(shardNextEpochCandidateStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
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
		ShardCandidateHash:              tempShardCandidateHash,
		ShardCommitteeAndValidatorHash:  tempShardCommitteeAndValidatorHash,
		AutoStakeHash:                   tempAutoStakingHash,
	}
	return hashes, nil
}

func (b *beaconCommitteeStateBase) SetBeaconCommittees(committees []incognitokey.CommitteePublicKey) {
	b.beaconCommittee = []incognitokey.CommitteePublicKey{}
	b.beaconCommittee = append(b.beaconCommittee, committees...)
}

func (b *beaconCommitteeStateBase) SetNumberOfAssignedCandidates(numberOfAssignedCandidates int) {
	b.numberOfAssignedCandidates = numberOfAssignedCandidates
}

func (b beaconCommitteeStateBase) SwapRule() SwapRule {
	return b.swapRule
}

func (b beaconCommitteeStateBase) UnassignedCommonPool() []string {
	commonPoolValidators := []string{}
	candidateShardWaitingForNextRandomStr, _ := incognitokey.CommitteeKeyListToString(b.shardCommonPool[b.numberOfAssignedCandidates:])
	commonPoolValidators = append(commonPoolValidators, candidateShardWaitingForNextRandomStr...)
	return commonPoolValidators
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
	shardCandidates := b.shardCommonPool
	shardCandidatesStr, err := incognitokey.CommitteeKeyListToString(shardCandidates)
	if err != nil {
		panic(err)
	}
	res = append(res, shardCandidatesStr...)
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
	candidateShardWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(b.shardCommonPool[:b.numberOfAssignedCandidates])
	if err != nil {
		return nil, err
	}
	validators = append(validators, candidateShardWaitingForCurrentRandomStr...)

	return validators, nil
}

func (b *beaconCommitteeStateBase) buildReturnStakingInstructionAndDeleteStakerInfo(
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	committeePublicKeyStruct incognitokey.CommitteePublicKey,
	publicKey string,
	stakerInfo *statedb.StakerInfo,
	committeeChange *CommitteeChange,
) (*instruction.ReturnStakeInstruction, *CommitteeChange, error) {
	returnStakingInstruction = buildReturnStakingInstruction(
		returnStakingInstruction,
		publicKey,
		stakerInfo.TxStakingID().String(),
	)
	committeeChange, err := b.deleteStakerInfo(committeePublicKeyStruct, publicKey, committeeChange)
	if err != nil {
		return returnStakingInstruction, committeeChange, err
	}
	return returnStakingInstruction, committeeChange, nil
}

func buildReturnStakingInstruction(
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	publicKey string,
	txStake string,
) *instruction.ReturnStakeInstruction {
	returnStakingInstruction.AddNewRequest(publicKey, txStake)
	return returnStakingInstruction
}

func (b *beaconCommitteeStateBase) deleteStakerInfo(
	committeePublicKeyStruct incognitokey.CommitteePublicKey,
	committeePublicKey string,
	committeeChange *CommitteeChange,
) (*CommitteeChange, error) {
	committeeChange.RemovedStaker = append(committeeChange.RemovedStaker, committeePublicKey)
	delete(b.rewardReceiver, committeePublicKeyStruct.GetIncKeyBase58())
	delete(b.autoStake, committeePublicKey)
	delete(b.stakingTx, committeePublicKey)
	return committeeChange, nil
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
	// var key string
	for index, candidate := range stakeInstruction.PublicKeyStructs {
		committeePublicKey := stakeInstruction.PublicKeys[index]
		b.rewardReceiver[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
		b.autoStake[committeePublicKey] = stakeInstruction.AutoStakingFlag[index]
		b.stakingTx[committeePublicKey] = stakeInstruction.TxStakeHashes[index]
	}
	committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, stakeInstruction.PublicKeyStructs...)
	b.shardCommonPool = append(b.shardCommonPool, stakeInstruction.PublicKeyStructs...)

	return committeeChange, err
}

func (b *beaconCommitteeStateBase) processStopAutoStakeInstruction(
	stopAutoStakeInstruction *instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	// b == newstate -> only write
	// oldstate -> only read

	//careful with this variable
	// validators := oldState.getAllCandidateSubstituteCommittee()
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

func (b *beaconCommitteeStateBase) processAssignWithRandomInstruction(
	rand int64,
	activeShards int,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	// b == newstate -> only write
	// oldstate -> only read
	newCommitteeChange := committeeChange
	candidateStructs := oldState.ShardCommonPool()[:b.numberOfAssignedCandidates]
	candidates, _ := incognitokey.CommitteeKeyListToString(candidateStructs)
	newCommitteeChange = b.assign(candidates, rand, activeShards, newCommitteeChange, oldState)
	newCommitteeChange.NextEpochShardCandidateRemoved = append(newCommitteeChange.NextEpochShardCandidateRemoved, candidateStructs...)
	b.shardCommonPool = b.shardCommonPool[b.numberOfAssignedCandidates:]
	b.numberOfAssignedCandidates = 0

	return newCommitteeChange
}

func (b *beaconCommitteeStateBase) assign(
	candidates []string, rand int64, activeShards int, committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	numberOfValidator := make([]int, activeShards)
	for i := 0; i < activeShards; i++ {
		numberOfValidator[byte(i)] += len(oldState.ShardSubstitute()[byte(i)])
		numberOfValidator[byte(i)] += len(oldState.ShardCommittee()[byte(i)])
	}

	assignedCandidates := assignShardCandidateV2(candidates, numberOfValidator, rand)
	for shardID, tempCandidates := range assignedCandidates {
		tempCandidateStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(tempCandidates)
		committeeChange.ShardSubstituteAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], tempCandidateStructs...)
		b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], tempCandidateStructs...)
	}
	return committeeChange
}

//processSwapShardInstruction update committees state by swap shard instruction
// Process single swap shard instruction for and update committee state
func (b *beaconCommitteeStateBase) processSwapShardInstruction(
	swapShardInstruction *instruction.SwapShardInstruction,
	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (
	*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	var err error
	shardID := byte(swapShardInstruction.ChainID)
	committees := oldState.ShardCommittee()[shardID]
	substitutes := oldState.ShardSubstitute()[shardID]
	tempCommittees, _ := incognitokey.CommitteeKeyListToString(committees)
	tempSubstitutes, _ := incognitokey.CommitteeKeyListToString(substitutes)

	comparedShardSwapInstruction, newCommittees, _,
		slashingCommittees, normalSwapOutCommittees := b.swapRule.GenInstructions(
		shardID,
		tempCommittees,
		tempSubstitutes,
		env.MinShardCommitteeSize,
		env.MaxShardCommitteeSize,
		instruction.SWAP_BY_END_EPOCH,
		env.NumberOfFixedShardBlockValidator,
		env.DcsMaxShardCommitteeSize,
		env.DcsMinShardCommitteeSize,
		env.MissingSignaturePenalty,
	)

	if len(slashingCommittees) > 0 {
		Logger.log.Infof("SHARD %+v, Epoch %+v, Slashing Committees %+v", shardID, env.Epoch, slashingCommittees)
	} else {
		Logger.log.Infof("SHARD %+v, Epoch %+v, NO Slashing Committees", shardID, env.Epoch)
	}

	if !reflect.DeepEqual(comparedShardSwapInstruction.InPublicKeys, swapShardInstruction.InPublicKeys) {
		return nil, returnStakingInstruction, fmt.Errorf("expect swap in keys %+v, got %+v",
			comparedShardSwapInstruction.InPublicKeys, swapShardInstruction.InPublicKeys)
	}

	if !reflect.DeepEqual(comparedShardSwapInstruction.OutPublicKeys, swapShardInstruction.OutPublicKeys) {
		return nil, returnStakingInstruction, fmt.Errorf("expect swap out keys %+v, got %+v",
			comparedShardSwapInstruction.OutPublicKeys, swapShardInstruction.OutPublicKeys)
	}

	b.shardCommittee[shardID], _ = incognitokey.CommitteeBase58KeyListToStruct(newCommittees)
	b.shardSubstitute[shardID] = b.shardSubstitute[shardID][len(swapShardInstruction.InPublicKeys):]

	committeeChange.ShardCommitteeRemoved[shardID] = append(committeeChange.ShardCommitteeRemoved[shardID],
		incognitokey.DeepCopy(swapShardInstruction.OutPublicKeyStructs)...)
	committeeChange.ShardSubstituteRemoved[shardID] = append(committeeChange.ShardSubstituteRemoved[shardID],
		incognitokey.DeepCopy(swapShardInstruction.InPublicKeyStructs)...)
	committeeChange.ShardCommitteeAdded[shardID] = append(committeeChange.ShardCommitteeAdded[shardID],
		incognitokey.DeepCopy(swapShardInstruction.InPublicKeyStructs)...)

	// process after swap for assign old committees to current shard pool
	committeeChange, returnStakingInstruction, err = b.processAfterNormalSwap(env,
		normalSwapOutCommittees,
		committeeChange,
		returnStakingInstruction,
		oldState,
	)
	if err != nil {
		return nil, returnStakingInstruction, err
	}

	returnStakingInstruction, committeeChange, err = b.processSlashing(
		env,
		slashingCommittees,
		returnStakingInstruction,
		committeeChange,
	)
	if err != nil {
		return nil, returnStakingInstruction, err
	}

	committeeChange.SlashingCommittee[shardID] = append(committeeChange.SlashingCommittee[shardID], slashingCommittees...)

	return committeeChange, returnStakingInstruction, nil
}

// processAfterNormalSwap process swapped out committee public key
// - auto stake is false then remove completely out of any committee, candidate, substitute list
// - auto stake is true then using assignment rule v2 to assign this committee public key
func (b *beaconCommitteeStateBase) processAfterNormalSwap(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	candidates := []string{}

	outPublicKeyStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
	for index, outPublicKey := range outPublicKeys {
		stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, outPublicKey)
		if err != nil {
			return committeeChange, returnStakingInstruction, err
		}
		if !has {
			return committeeChange, returnStakingInstruction, errors.Errorf("Can not found info of this public key %v", outPublicKey)
		}
		if stakerInfo.AutoStaking() {
			candidates = append(candidates, outPublicKey)
		} else {
			returnStakingInstruction, committeeChange, err = b.buildReturnStakingInstructionAndDeleteStakerInfo(
				returnStakingInstruction,
				outPublicKeyStructs[index],
				outPublicKey,
				stakerInfo,
				committeeChange,
			)

			if err != nil {
				return committeeChange, returnStakingInstruction, err
			}
		}
	}

	committeeChange = b.assign(candidates, env.RandomNumber, env.ActiveShards, committeeChange, oldState)
	return committeeChange, returnStakingInstruction, nil
}

// processAfterNormalSwap process swapped out committee public key
// if number of round is less than MAX_NUMBER_OF_ROUND go back to THAT shard pool, and increase number of round
// if number of round is equal to or greater than MAX_NUMBER_OF_ROUND
// - auto stake is false then remove completely out of any committee, candidate, substitute list
// - auto stake is true then using assignment rule v2 to assign this committee public key
func (b *beaconCommitteeStateBase) processSlashing(
	env *BeaconCommitteeStateEnvironment,
	slashingPublicKeys []string,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	committeeChange *CommitteeChange,
) (*instruction.ReturnStakeInstruction, *CommitteeChange, error) {
	slashingPublicKeyStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(slashingPublicKeys)
	for index, outPublicKey := range slashingPublicKeys {
		stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, outPublicKey)
		if err != nil {
			return returnStakingInstruction, committeeChange, err
		}
		if !has {
			return returnStakingInstruction, committeeChange, fmt.Errorf("Can not found info of this public key %v", outPublicKey)
		}
		returnStakingInstruction, committeeChange, err = b.buildReturnStakingInstructionAndDeleteStakerInfo(
			returnStakingInstruction,
			slashingPublicKeyStructs[index],
			outPublicKey,
			stakerInfo,
			committeeChange,
		)
		if err != nil {
			return returnStakingInstruction, committeeChange, err
		}
	}

	return returnStakingInstruction, committeeChange, nil
}

//processUnstakeInstruction : process unstake instruction from beacon block
func (b *beaconCommitteeStateBase) processUnstakeInstruction(
	unstakeInstruction *instruction.UnstakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	// b == newstate -> only write
	// oldstate -> only read
	shardCommonPoolStr, _ := incognitokey.CommitteeKeyListToString(b.shardCommonPool)
	for index, publicKey := range unstakeInstruction.CommitteePublicKeys {
		if common.IndexOfStr(publicKey, env.newUnassignedCommonPool) == -1 {
			if common.IndexOfStr(publicKey, env.newAllSubstituteCommittees) != -1 {
				// if found in committee list then turn off auto staking
				if _, ok := oldState.AutoStake()[publicKey]; ok {
					committeeChange = b.stopAutoStake(publicKey, committeeChange)
				}
			}
		} else {
			indexCandidate := common.IndexOfStr(publicKey, shardCommonPoolStr)
			if indexCandidate == -1 {
				return committeeChange, returnStakingInstruction, errors.Errorf("Committee public key: %s is not valid for any committee sets", publicKey)
			}
			shardCommonPoolStr = append(shardCommonPoolStr[:indexCandidate], shardCommonPoolStr[indexCandidate+1:]...)
			b.shardCommonPool = append(b.shardCommonPool[:indexCandidate], b.shardCommonPool[indexCandidate+1:]...)
			stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, publicKey)
			if err != nil {
				return committeeChange, returnStakingInstruction, err
			}
			if !has {
				return committeeChange, returnStakingInstruction, errors.New("Can't find staker info")
			}

			committeeChange.NextEpochShardCandidateRemoved =
				append(committeeChange.NextEpochShardCandidateRemoved, unstakeInstruction.CommitteePublicKeysStruct[index])

			returnStakingInstruction, committeeChange, err = b.buildReturnStakingInstructionAndDeleteStakerInfo(
				returnStakingInstruction,
				unstakeInstruction.CommitteePublicKeysStruct[index],
				publicKey,
				stakerInfo,
				committeeChange,
			)

			if err != nil {
				return committeeChange, returnStakingInstruction, errors.New("Can't find staker info")
			}
		}
	}

	return committeeChange, returnStakingInstruction, nil
}

func (b *beaconCommitteeStateBase) ProcessStakeInstruction(
	stakeInstruction *instruction.StakeInstruction,
	committeeChange *CommitteeChange,
) (*CommitteeChange, error) {
	return b.processStakeInstruction(stakeInstruction, committeeChange)
}

func (b *beaconCommitteeStateBase) ProcessStopAutoStakeInstruction(
	stopAutoStakeInstruction *instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	return b.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange, oldState)
}

func (b *beaconCommitteeStateBase) ProcessAssignWithRandomInstruction(
	rand int64,
	activeShards int,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	return b.processAssignWithRandomInstruction(rand, activeShards, committeeChange, oldState)
}

func (b *beaconCommitteeStateBase) ProcessSwapShardInstruction(
	swapShardInstruction *instruction.SwapShardInstruction,
	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (
	*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	return b.processSwapShardInstruction(swapShardInstruction, env, committeeChange, returnStakingInstruction, oldState)
}

//processUnstakeInstruction : process unstake instruction from beacon block
func (b *beaconCommitteeStateBase) ProcessUnstakeInstruction(
	unstakeInstruction *instruction.UnstakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	return b.processUnstakeInstruction(unstakeInstruction, env, committeeChange, returnStakingInstruction, oldState)
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
	b.swapRule = swapRule
}

func (b *beaconCommitteeStateBase) ProcessAssignInstruction(
	assignInstruction *instruction.AssignInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) (
	*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	return b.processAssignInstruction(assignInstruction, env, committeeChange)
}

func (b *beaconCommitteeStateBase) processAssignInstruction(
	assignInstruction *instruction.AssignInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) (
	*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	return committeeChange, &instruction.ReturnStakeInstruction{}, nil
}
