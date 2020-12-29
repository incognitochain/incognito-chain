package committeestate

import (
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type beaconCommitteeStateSlashingBase struct {
	beaconCommitteeStateBase

	shardCommonPool            []incognitokey.CommitteePublicKey
	probationPool              map[string]signaturecounter.Penalty
	numberOfAssignedCandidates int

	swapRule SwapRule
}

func NewBeaconCommitteeStateSlashingBase() *beaconCommitteeStateSlashingBase {
	return &beaconCommitteeStateSlashingBase{
		beaconCommitteeStateBase: *NewBeaconCommitteeStateBase(),
		probationPool:            map[string]signaturecounter.Penalty{},
	}
}

func NewBeaconCommitteeStateSlashingBaseWithValue(
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	shardCommonPool []incognitokey.CommitteePublicKey,
	probationPool map[string]signaturecounter.Penalty,
	numberOfAssignedCandidates int,
	swapRule SwapRule,
) *beaconCommitteeStateSlashingBase {
	return &beaconCommitteeStateSlashingBase{
		beaconCommitteeStateBase: *NewBeaconCommitteeStateBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute,
			autoStake, rewardReceiver, stakingTx,
		),
		shardCommonPool:            shardCommonPool,
		probationPool:              probationPool,
		numberOfAssignedCandidates: numberOfAssignedCandidates,
		swapRule:                   swapRule,
	}
}

func (b beaconCommitteeStateSlashingBase) ShardCommonPool() []incognitokey.CommitteePublicKey {
	return b.shardCommonPool
}

func (b beaconCommitteeStateSlashingBase) ProbationPool() map[string]signaturecounter.Penalty {
	return b.probationPool
}

func (b beaconCommitteeStateSlashingBase) NumberOfAssignedCandidates() int {
	return b.numberOfAssignedCandidates
}

func (b beaconCommitteeStateSlashingBase) SwapRule() SwapRule {
	return b.swapRule
}

func (b beaconCommitteeStateSlashingBase) clone() *beaconCommitteeStateSlashingBase {
	res := NewBeaconCommitteeStateSlashingBase()
	res.beaconCommitteeStateBase = *b.beaconCommitteeStateBase.clone()

	res.numberOfAssignedCandidates = b.numberOfAssignedCandidates
	res.shardCommonPool = make([]incognitokey.CommitteePublicKey, len(b.shardCommonPool))
	copy(res.shardCommonPool, b.shardCommonPool)
	for i, v := range b.probationPool {
		res.probationPool[i] = v
	}
	res.swapRule = cloneSwapRuleByVersion(b.swapRule)

	return res
}

func (b *beaconCommitteeStateSlashingBase) reset() {
	b.beaconCommitteeStateBase.reset()
	b.numberOfAssignedCandidates = 0
	b.probationPool = map[string]signaturecounter.Penalty{}
	b.shardCommonPool = []incognitokey.CommitteePublicKey{}
	b.swapRule = nil // be careful here
}

func (b beaconCommitteeStateSlashingBase) Version() int {
	panic("Implement version for committee state slashing")
}

func (b beaconCommitteeStateSlashingBase) getAllCandidateSubstituteCommittee() []string {
	res := []string{}
	res = b.beaconCommitteeStateBase.getAllCandidateSubstituteCommittee()
	shardCandidates := b.shardCommonPool
	shardCandidatesStr, err := incognitokey.CommitteeKeyListToString(shardCandidates)
	if err != nil {
		panic(err)
	}
	res = append(res, shardCandidatesStr...)
	return res
}

func (b beaconCommitteeStateSlashingBase) IsEmpty() bool {
	return reflect.DeepEqual(b, NewBeaconCommitteeStateSlashingBase())
}

func (b beaconCommitteeStateSlashingBase) Hash() (*BeaconCommitteeStateHash, error) {
	if b.IsEmpty() {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}
	hashes, err := b.beaconCommitteeStateBase.Hash()
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
	return hashes, nil
}

func (b beaconCommitteeStateSlashingBase) UnassignedCommonPool() []string {
	commonPoolValidators := []string{}
	candidateShardWaitingForNextRandomStr, _ := incognitokey.CommitteeKeyListToString(b.shardCommonPool[b.numberOfAssignedCandidates:])
	commonPoolValidators = append(commonPoolValidators, candidateShardWaitingForNextRandomStr...)
	return commonPoolValidators
}

func (b beaconCommitteeStateSlashingBase) getAllSubstituteCommittees() ([]string, error) {
	validators, err := b.beaconCommitteeStateBase.getAllSubstituteCommittees()
	if err != nil {
		return []string{}, err
	}

	candidateShardWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(b.shardCommonPool[:b.numberOfAssignedCandidates])
	if err != nil {
		return nil, err
	}
	validators = append(validators, candidateShardWaitingForCurrentRandomStr...)
	return validators, nil
}

func (b *beaconCommitteeStateSlashingBase) SetSwapRule(swapRule SwapRule) {
	b.swapRule = swapRule
}

func (b *beaconCommitteeStateSlashingBase) buildReturnStakingInstructionAndDeleteStakerInfo(
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

func (b *beaconCommitteeStateSlashingBase) deleteStakerInfo(
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

func (b *beaconCommitteeStateSlashingBase) processStakeInstruction(
	stakeInstruction *instruction.StakeInstruction,
	committeeChange *CommitteeChange,
) (*CommitteeChange, error) {
	newCommitteeChange, err := b.beaconCommitteeStateBase.processStakeInstruction(stakeInstruction, committeeChange)
	b.shardCommonPool = append(b.shardCommonPool, stakeInstruction.PublicKeyStructs...)
	return newCommitteeChange, err
}

func (b *beaconCommitteeStateSlashingBase) processAssignWithRandomInstruction(
	rand int64,
	activeShards int,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	newCommitteeChange := committeeChange
	candidateStructs := oldState.ShardCommonPool()[:b.numberOfAssignedCandidates]
	candidates, _ := incognitokey.CommitteeKeyListToString(candidateStructs)
	newCommitteeChange = b.assign(candidates, rand, activeShards, newCommitteeChange, oldState)
	newCommitteeChange.NextEpochShardCandidateRemoved = append(newCommitteeChange.NextEpochShardCandidateRemoved, candidateStructs...)
	b.shardCommonPool = b.shardCommonPool[b.numberOfAssignedCandidates:]
	b.numberOfAssignedCandidates = 0

	return newCommitteeChange
}

func (b *beaconCommitteeStateSlashingBase) getAssignCandidates(candidates []string, rand int64, activeShards int, oldState BeaconCommitteeState) map[byte][]string {
	numberOfValidator := make([]int, activeShards)
	for i := 0; i < activeShards; i++ {
		numberOfValidator[byte(i)] += len(oldState.ShardSubstitute()[byte(i)])
		numberOfValidator[byte(i)] += len(oldState.ShardCommittee()[byte(i)])
	}

	assignedCandidates := assignShardCandidateV2(candidates, numberOfValidator, rand)
	return assignedCandidates
}

func (b *beaconCommitteeStateSlashingBase) assign(
	candidates []string, rand int64, activeShards int, committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	assignedCandidates := b.getAssignCandidates(candidates, rand, activeShards, oldState)
	for shardID, tempCandidates := range assignedCandidates {
		tempCandidateStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(tempCandidates)
		committeeChange.ShardSubstituteAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], tempCandidateStructs...)
		b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], tempCandidateStructs...)
	}
	return committeeChange
}

//processSwapShardInstruction update committees state by swap shard instruction
// Process single swap shard instruction for and update committee state
func (b *beaconCommitteeStateSlashingBase) processSwapShardInstruction(
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

func (b *beaconCommitteeStateSlashingBase) getValidatorsByAutoStake(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
) ([]string, *CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	candidates := []string{}

	outPublicKeyStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
	for index, outPublicKey := range outPublicKeys {
		stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, outPublicKey)
		if err != nil {
			return candidates, committeeChange, returnStakingInstruction, err
		}
		if !has {
			return candidates, committeeChange, returnStakingInstruction, errors.Errorf("Can not found info of this public key %v", outPublicKey)
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
				return candidates, committeeChange, returnStakingInstruction, err
			}
		}
	}

	return candidates, committeeChange, returnStakingInstruction, nil
}

// processAfterNormalSwap process swapped out committee public key
// - auto stake is false then remove completely out of any committee, candidate, substitute list
// - auto stake is true then using assignment rule v2 to assign this committee public key
func (b *beaconCommitteeStateSlashingBase) processAfterNormalSwap(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	candidates, committeeChange, returnStakingInstruction, err := b.getValidatorsByAutoStake(env, outPublicKeys, committeeChange, returnStakingInstruction)
	if err != nil {
		return committeeChange, returnStakingInstruction, err
	}
	committeeChange = b.assign(candidates, env.RandomNumber, env.ActiveShards, committeeChange, oldState)

	return committeeChange, returnStakingInstruction, nil
}

// processAfterNormalSwap process swapped out committee public key
// if number of round is less than MAX_NUMBER_OF_ROUND go back to THAT shard pool, and increase number of round
// if number of round is equal to or greater than MAX_NUMBER_OF_ROUND
// - auto stake is false then remove completely out of any committee, candidate, substitute list
// - auto stake is true then using assignment rule v2 to assign this committee public key
func (b *beaconCommitteeStateSlashingBase) processSlashing(
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
func (b *beaconCommitteeStateSlashingBase) processUnstakeInstruction(
	unstakeInstruction *instruction.UnstakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
) (*CommitteeChange, *instruction.ReturnStakeInstruction, error) {

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
