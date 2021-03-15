package committeestate

import (
	"fmt"
	"reflect"

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
	numberOfAssignedCandidates int

	swapRule SwapRuleProcessor
}

func newBeaconCommitteeStateSlashingBase() *beaconCommitteeStateSlashingBase {
	return &beaconCommitteeStateSlashingBase{
		beaconCommitteeStateBase: *newBeaconCommitteeStateBase(),
	}
}

func newBeaconCommitteeStateSlashingBaseWithValue(
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	shardCommonPool []incognitokey.CommitteePublicKey,
	numberOfAssignedCandidates int,
	swapRule SwapRuleProcessor,
) *beaconCommitteeStateSlashingBase {
	return &beaconCommitteeStateSlashingBase{
		beaconCommitteeStateBase: *newBeaconCommitteeStateBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute,
			autoStake, rewardReceiver, stakingTx,
		),
		shardCommonPool:            shardCommonPool,
		numberOfAssignedCandidates: numberOfAssignedCandidates,
		swapRule:                   swapRule,
	}
}

func (b beaconCommitteeStateSlashingBase) Version() int {
	panic("implement me")
}

func (b *beaconCommitteeStateSlashingBase) Clone() BeaconCommitteeState {
	return b.clone()
}

func (b beaconCommitteeStateSlashingBase) clone() *beaconCommitteeStateSlashingBase {
	res := newBeaconCommitteeStateSlashingBase()
	res.beaconCommitteeStateBase = *b.beaconCommitteeStateBase.clone()

	res.numberOfAssignedCandidates = b.numberOfAssignedCandidates
	res.shardCommonPool = make([]incognitokey.CommitteePublicKey, len(b.shardCommonPool))
	copy(res.shardCommonPool, b.shardCommonPool)
	res.swapRule = cloneSwapRuleByVersion(b.swapRule)

	return res
}

func (b beaconCommitteeStateSlashingBase) Hash() (*BeaconCommitteeStateHash, error) {
	if b.isEmpty() {
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

func (b beaconCommitteeStateSlashingBase) isEmpty() bool {
	return reflect.DeepEqual(b, newBeaconCommitteeStateSlashingBase())
}

func (b beaconCommitteeStateSlashingBase) GetShardCommonPool() []incognitokey.CommitteePublicKey {
	return b.shardCommonPool
}

func (b beaconCommitteeStateSlashingBase) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return b.shardCommonPool[b.numberOfAssignedCandidates:]
}

func (b beaconCommitteeStateSlashingBase) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return b.shardCommonPool[:b.numberOfAssignedCandidates]
}

func (b beaconCommitteeStateSlashingBase) GetAllCandidateSubstituteCommittee() []string {
	return b.getAllCandidateSubstituteCommittee()
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

//IsSwapTime read from interface des
func (beaconCommitteeStateSlashingBase) IsSwapTime(beaconHeight, numberOfBlockEachEpoch uint64) bool {
	if beaconHeight%numberOfBlockEachEpoch == 1 {
		return true
	} else {
		return false
	}
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

func (b *beaconCommitteeStateSlashingBase) InitCommitteeState(env *BeaconCommitteeStateEnvironment) {
	b.beaconCommitteeStateBase.InitCommitteeState(env)
	b.swapRule = SwapRuleByEnv(env)
}

func (b *beaconCommitteeStateSlashingBase) GenerateAllSwapShardInstructions(
	env *BeaconCommitteeStateEnvironment) (
	[]*instruction.SwapShardInstruction, error) {
	swapShardInstructions := []*instruction.SwapShardInstruction{}
	for i := 0; i < len(b.shardCommittee); i++ {
		shardID := byte(i)
		committees := b.shardCommittee[shardID]
		substitutes := b.shardSubstitute[shardID]
		tempCommittees, _ := incognitokey.CommitteeKeyListToString(committees)
		tempSubstitutes, _ := incognitokey.CommitteeKeyListToString(substitutes)

		swapShardInstruction, _, _, _, _ := b.swapRule.Process(
			shardID,
			tempCommittees,
			tempSubstitutes,
			env.MinShardCommitteeSize,
			env.MaxShardCommitteeSize,
			instruction.SWAP_BY_END_EPOCH,
			env.NumberOfFixedShardBlockValidator,
			env.MissingSignaturePenalty,
		)

		if !swapShardInstruction.IsEmpty() {
			swapShardInstructions = append(swapShardInstructions, swapShardInstruction)
		} else {
			Logger.log.Infof("Generate empty swap shard instructions")
		}
	}
	return swapShardInstructions, nil
}

func (b *beaconCommitteeStateSlashingBase) buildReturnStakingInstructionAndDeleteStakerInfo(
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	committeePublicKeyStruct incognitokey.CommitteePublicKey,
	publicKey string,
	stakerInfo *statedb.StakerInfo,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) (*instruction.ReturnStakeInstruction, *CommitteeChange, error) {
	returnStakingInstruction = buildReturnStakingInstruction(
		returnStakingInstruction,
		publicKey,
		stakerInfo.TxStakingID().String(),
	)
	committeeChange, err := b.removeFromState(committeePublicKeyStruct, publicKey, committeeChange, oldState)
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

func (b *beaconCommitteeStateSlashingBase) removeFromState(
	committeePublicKeyStruct incognitokey.CommitteePublicKey,
	committeePublicKey string,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) (*CommitteeChange, error) {
	delete(b.rewardReceiver, committeePublicKeyStruct.GetIncKeyBase58())
	delete(b.autoStake, committeePublicKey)
	delete(b.stakingTx, committeePublicKey)
	committeeChange.RemovedStaker = append(committeeChange.RemovedStaker, committeePublicKey)

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

func (b *beaconCommitteeStateSlashingBase) updateCandidatesByRandom(
	committeeChange *CommitteeChange, oldState BeaconCommitteeState,
) (*CommitteeChange, []string) {
	newCommitteeChange := committeeChange
	candidateStructs := oldState.GetShardCommonPool()[:b.numberOfAssignedCandidates]
	candidates, _ := incognitokey.CommitteeKeyListToString(candidateStructs)
	newCommitteeChange.NextEpochShardCandidateRemoved = append(newCommitteeChange.NextEpochShardCandidateRemoved, candidateStructs...)
	b.shardCommonPool = b.shardCommonPool[b.numberOfAssignedCandidates:]
	b.numberOfAssignedCandidates = 0
	return newCommitteeChange, candidates
}

func (b *beaconCommitteeStateSlashingBase) processAssignWithRandomInstruction(
	rand int64,
	activeShards int,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	newCommitteeChange, candidates := b.updateCandidatesByRandom(committeeChange, oldState)
	newCommitteeChange = b.assign(candidates, rand, activeShards, newCommitteeChange, oldState)
	return newCommitteeChange
}

func (b *beaconCommitteeStateSlashingBase) getAssignCandidates(candidates []string, rand int64, activeShards int, oldState BeaconCommitteeState) map[byte][]string {
	numberOfValidator := make([]int, activeShards)
	for i := 0; i < activeShards; i++ {
		numberOfValidator[byte(i)] += len(oldState.GetShardSubstitute()[byte(i)])
		numberOfValidator[byte(i)] += len(oldState.GetShardCommittee()[byte(i)])
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

func (b *beaconCommitteeStateSlashingBase) processNormalSwap(
	swapShardInstruction *instruction.SwapShardInstruction,
	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange,
	oldState BeaconCommitteeState) (
	*CommitteeChange, []string, []string, []string, error,
) {
	shardID := byte(swapShardInstruction.ChainID)
	newCommitteeChange := committeeChange
	committees := oldState.GetShardCommittee()[shardID]
	substitutes := oldState.GetShardSubstitute()[shardID]
	tempCommittees, _ := incognitokey.CommitteeKeyListToString(committees)
	tempSubstitutes, _ := incognitokey.CommitteeKeyListToString(substitutes)

	comparedShardSwapInstruction, newCommittees, _,
		slashingCommittees, normalSwapOutCommittees := b.swapRule.Process(
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
		return newCommitteeChange, []string{}, []string{}, []string{},
			fmt.Errorf("expect swap in keys %+v, got %+v",
				comparedShardSwapInstruction.InPublicKeys, swapShardInstruction.InPublicKeys)
	}

	if !reflect.DeepEqual(comparedShardSwapInstruction.OutPublicKeys, swapShardInstruction.OutPublicKeys) {
		return newCommitteeChange, []string{}, []string{}, []string{},
			fmt.Errorf("expect swap out keys %+v, got %+v",
				comparedShardSwapInstruction.OutPublicKeys, swapShardInstruction.OutPublicKeys)
	}
	b.shardCommittee[shardID], _ = incognitokey.CommitteeBase58KeyListToStruct(newCommittees)
	b.shardSubstitute[shardID] = b.shardSubstitute[shardID][len(swapShardInstruction.InPublicKeys):]

	newCommitteeChange.ShardCommitteeRemoved[shardID] = append(newCommitteeChange.ShardCommitteeRemoved[shardID],
		incognitokey.DeepCopy(swapShardInstruction.OutPublicKeyStructs)...)
	newCommitteeChange.ShardSubstituteRemoved[shardID] = append(newCommitteeChange.ShardSubstituteRemoved[shardID],
		incognitokey.DeepCopy(swapShardInstruction.InPublicKeyStructs)...)
	newCommitteeChange.ShardCommitteeAdded[shardID] = append(newCommitteeChange.ShardCommitteeAdded[shardID],
		incognitokey.DeepCopy(swapShardInstruction.InPublicKeyStructs)...)

	return newCommitteeChange, swapShardInstruction.InPublicKeys, normalSwapOutCommittees, slashingCommittees, nil
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
	shardID := byte(swapShardInstruction.ChainID)

	newCommitteeChange, _, normalSwapOutCommittees, slashingCommittees, err := b.processNormalSwap(swapShardInstruction, env, committeeChange, oldState)
	if err != nil {
		return nil, returnStakingInstruction, err
	}

	// process after swap for assign old committees to current shard pool
	newCommitteeChange, returnStakingInstruction, err = b.processAfterNormalSwap(
		env,
		normalSwapOutCommittees,
		newCommitteeChange,
		returnStakingInstruction,
		oldState,
	)
	if err != nil {
		return nil, returnStakingInstruction, err
	}

	//process slashing after normal swap out
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

func (b *beaconCommitteeStateSlashingBase) getValidatorsByAutoStake(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	oldState BeaconCommitteeState,
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
				oldState,
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
	candidates, committeeChange, returnStakingInstruction, err := b.getValidatorsByAutoStake(env, outPublicKeys, committeeChange, returnStakingInstruction, oldState)
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
	oldState BeaconCommitteeState,
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
			oldState,
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
				if _, ok := oldState.GetAutoStaking()[publicKey]; ok {
					committeeChange = b.turnOffStopAutoStake(publicKey, committeeChange)
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
				oldState,
			)

			if err != nil {
				return committeeChange, returnStakingInstruction, errors.New("Can't find staker info")
			}
		}
	}

	return committeeChange, returnStakingInstruction, nil
}

//SplitReward ...
func (engine *beaconCommitteeStateSlashingBase) Process(
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
	lenBeaconCommittees := uint64(len(engine.GetBeaconCommittee()))
	lenShardCommittees := uint64(len(engine.GetShardCommittee()[env.ShardID]))

	if len(allCoinTotalReward) == 0 {
		Logger.log.Info("Beacon Height %+v, 😭 found NO reward", env.BeaconHeight)
		return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
	}

	for key, totalReward := range allCoinTotalReward {
		totalRewardForDAOAndCustodians := devPercent * totalReward / 100
		totalRewardForShardAndBeaconValidators := totalReward - totalRewardForDAOAndCustodians
		shardWeight := float64(lenShardCommittees)
		beaconWeight := 2 * float64(lenBeaconCommittees) / float64(env.ActiveShards)
		totalValidatorWeight := shardWeight + beaconWeight

		rewardForShard[key] = uint64(shardWeight * float64(totalRewardForShardAndBeaconValidators) / totalValidatorWeight)
		Logger.log.Infof("[test-salary] totalRewardForDAOAndCustodians tokenID %v - %v\n",
			key.String(), totalRewardForDAOAndCustodians)

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
