package committeestate

import (
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/config"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type beaconCommitteeStateSlashingBase struct {
	beaconCommitteeStateBase

	shardCommonPool            []string
	numberOfAssignedCandidates int
	swapRule                   SwapRuleProcessor
	assignRule                 AssignRuleProcessor
}

func newBeaconCommitteeStateSlashingBase() *beaconCommitteeStateSlashingBase {
	return &beaconCommitteeStateSlashingBase{
		beaconCommitteeStateBase: *newBeaconCommitteeStateBase(),
	}
}

func newBeaconCommitteeStateSlashingBaseWithValue(
	beaconCommittee []string,
	shardCommittee map[byte][]string,
	shardSubstitute map[byte][]string,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	shardCommonPool []string,
	numberOfAssignedCandidates int,
	swapRule SwapRuleProcessor,
	assignRule AssignRuleProcessor,
) *beaconCommitteeStateSlashingBase {
	return &beaconCommitteeStateSlashingBase{
		beaconCommitteeStateBase: *newBeaconCommitteeStateBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute,
			autoStake, rewardReceiver, stakingTx,
		),
		shardCommonPool:            shardCommonPool,
		numberOfAssignedCandidates: numberOfAssignedCandidates,
		swapRule:                   swapRule,
		assignRule:                 assignRule,
	}
}

func (b *beaconCommitteeStateSlashingBase) Version() int {
	panic("implement me")
}

func (b beaconCommitteeStateSlashingBase) AssignRuleVersion() int {
	return b.assignRule.Version()
}

func (b beaconCommitteeStateSlashingBase) shallowCopy(newB *beaconCommitteeStateSlashingBase) {
	newB.beaconCommitteeStateBase = b.beaconCommitteeStateBase
	newB.shardCommonPool = b.shardCommonPool
	newB.numberOfAssignedCandidates = b.numberOfAssignedCandidates
	newB.swapRule = b.swapRule
}

func (b *beaconCommitteeStateSlashingBase) Clone() BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.clone()
}

func (b *beaconCommitteeStateSlashingBase) clone() *beaconCommitteeStateSlashingBase {
	res := newBeaconCommitteeStateSlashingBase()
	res.beaconCommitteeStateBase = *b.beaconCommitteeStateBase.clone()

	res.numberOfAssignedCandidates = b.numberOfAssignedCandidates
	res.shardCommonPool = common.DeepCopyString(b.shardCommonPool)
	res.swapRule = b.swapRule
	res.assignRule = b.assignRule

	return res
}

func (b *beaconCommitteeStateSlashingBase) reset() {
	b.beaconCommitteeStateBase.reset()
	b.shardCommonPool = []string{}
}

func (b beaconCommitteeStateSlashingBase) Hash() (*BeaconCommitteeStateHash, error) {
	if b.isEmpty() {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}

	hashes, err := b.beaconCommitteeStateBase.Hash()
	if err != nil {
		return nil, err
	}
	committeeChange := b.committeeChange
	var tempShardCandidateHash common.Hash
	if !isNilOrShardCandidateHash(b.hashes) &&
		len(committeeChange.NextEpochShardCandidateRemoved) == 0 && len(committeeChange.NextEpochShardCandidateAdded) == 0 &&
		len(committeeChange.CurrentEpochShardCandidateRemoved) == 0 && len(committeeChange.CurrentEpochShardCandidateAdded) == 0 {
		tempShardCandidateHash = b.hashes.ShardCandidateHash
	} else {
		tempShardCandidateHash, err = common.GenerateHashFromStringArray(b.shardCommonPool)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
	}

	hashes.ShardCandidateHash = tempShardCandidateHash

	return hashes, nil
}

func (b beaconCommitteeStateSlashingBase) isEmpty() bool {
	return reflect.DeepEqual(b, newBeaconCommitteeStateSlashingBase())
}

func (b beaconCommitteeStateSlashingBase) GetShardCommonPool() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(b.shardCommonPool)
	return res
}

func (b beaconCommitteeStateSlashingBase) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(b.shardCommonPool[b.numberOfAssignedCandidates:])
	return res
}

func (b beaconCommitteeStateSlashingBase) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(b.shardCommonPool[:b.numberOfAssignedCandidates])
	return res
}

func (b beaconCommitteeStateSlashingBase) GetAllCandidateSubstituteCommittee() []string {
	return b.getAllCandidateSubstituteCommittee()
}

func (b beaconCommitteeStateSlashingBase) getAllCandidateSubstituteCommittee() []string {
	res := []string{}
	res = b.beaconCommitteeStateBase.getAllCandidateSubstituteCommittee()
	res = append(res, b.shardCommonPool...)
	return res
}

func (b beaconCommitteeStateSlashingBase) getAllSubstituteCommittees() ([]string, error) {
	validators, err := b.beaconCommitteeStateBase.getAllSubstituteCommittees()
	if err != nil {
		return []string{}, err
	}

	candidateShardWaitingForCurrentRandomStr := common.DeepCopyString(b.shardCommonPool[:b.numberOfAssignedCandidates])
	validators = append(validators, candidateShardWaitingForCurrentRandomStr...)
	return validators, nil
}

func (b *beaconCommitteeStateSlashingBase) initCommitteeState(env *BeaconCommitteeStateEnvironment) {
	b.beaconCommitteeStateBase.initCommitteeState(env)
	b.swapRule = GetSwapRuleVersion(env.BeaconHeight, config.Param().ConsensusParam.StakingFlowV3Height)
	b.assignRule = GetAssignRuleVersion(env.BeaconHeight, config.Param().ConsensusParam.StakingFlowV2Height, config.Param().ConsensusParam.AssignRuleV3Height)
}

func (b *beaconCommitteeStateSlashingBase) GenerateSwapShardInstructions(
	env *BeaconCommitteeStateEnvironment,
) (
	[]*instruction.SwapShardInstruction,
	error,
) {
	b.addData(env)
	swapShardInstructions := []*instruction.SwapShardInstruction{}
	slashedChange := NewCommitteeChange()
	for i := 0; i < env.ActiveShards; i++ {
		shardID := byte(i)
		tempCommittees := common.DeepCopyString(b.shardCommittee[shardID])
		tempSubstitutes := common.DeepCopyString(b.shardSubstitute[shardID])

		swapShardInstruction, _, _, slashedCommittee, _ := b.swapRule.Process(
			shardID,
			tempCommittees,
			tempSubstitutes,
			env.MinShardCommitteeSize,
			env.MaxShardCommitteeSize,
			instruction.SWAP_BY_END_EPOCH,
			env.NumberOfFixedShardBlockValidator,
			env.MissingSignaturePenalty,
		)
		if len(slashedCommittee) > 0 {
			slashedChange.SlashingCommittee[shardID] = slashedCommittee
		}
		if !swapShardInstruction.IsEmpty() {
			swapShardInstructions = append(swapShardInstructions, swapShardInstruction)
		} else {
			Logger.log.Infof("Generate empty swap shard instructions")
		}
	}
	return swapShardInstructions, nil
}

func (b *beaconCommitteeStateSlashingBase) buildReturnStakingInstructionAndDeleteStakerInfo(returnStakingInstruction *instruction.ReturnStakeInstruction, committeePublicKeyStruct incognitokey.CommitteePublicKey, publicKey string, stakerInfo *statedb.ShardStakerInfo) *instruction.ReturnStakeInstruction {
	//only return staking for non-foundation node
	if !config.Param().GenesisParam.AllFoundationNode[publicKey] {
		returnStakingInstruction = buildReturnStakingInstruction(
			returnStakingInstruction,
			publicKey,
			stakerInfo.TxStakingID().String(),
		)
		Logger.log.Info("Build returnstaking:", publicKey)
	} else {
		Logger.log.Info("Not Build returnstaking:", publicKey)
	}

	b.deleteStakerInfo(committeePublicKeyStruct, publicKey)

	return returnStakingInstruction
}

func buildReturnStakingInstruction(
	returnStakingInstruction *instruction.ReturnStakeInstruction,
	publicKey string,
	txStake string,
) *instruction.ReturnStakeInstruction {
	returnStakingInstruction.AddNewRequest(publicKey, txStake)
	return returnStakingInstruction
}

func (b *beaconCommitteeStateSlashingBase) deleteStakerInfo(committeePublicKeyStruct incognitokey.CommitteePublicKey, committeePublicKey string) {
	delete(b.rewardReceiver, committeePublicKeyStruct.GetIncKeyBase58())
	delete(b.autoStake, committeePublicKey)
	delete(b.stakingTx, committeePublicKey)
	b.committeeChange.AddRemovedStaker(committeePublicKey)
}

func (b *beaconCommitteeStateSlashingBase) processStakeInstruction(
	stakeInstruction *instruction.StakeInstruction,
) error {
	err := b.beaconCommitteeStateBase.processStakeInstruction(stakeInstruction)
	// b.committeeChange = b.beaconCommitteeStateBase.committeeChange
	if stakeInstruction.Chain == instruction.SHARD_INST {
		b.shardCommonPool = append(b.shardCommonPool, stakeInstruction.PublicKeys...)
	}
	return err
}

func (b *beaconCommitteeStateSlashingBase) getCandidatesForRandomAssignment() []string {
	candidates := b.shardCommonPool[:b.numberOfAssignedCandidates]
	b.committeeChange.AddNextEpochShardCandidateRemoved(candidates)
	b.shardCommonPool = b.shardCommonPool[b.numberOfAssignedCandidates:]
	b.numberOfAssignedCandidates = 0
	return candidates
}

func (b *beaconCommitteeStateSlashingBase) processAssignWithRandomInstruction(
	rand int64,
	numberOfValidator []int,
) {
	candidates := b.getCandidatesForRandomAssignment()
	b.assign(candidates, rand, numberOfValidator)
	return
}

func (b *beaconCommitteeStateSlashingBase) processRandomAssignment(
	candidates []string,
	rand int64,
	numberOfValidator []int,
) map[byte][]string {
	assignedCandidates := b.assignRule.Process(candidates, numberOfValidator, rand)
	return assignedCandidates
}

func (b *beaconCommitteeStateSlashingBase) assign(
	candidates []string, rand int64, numberOfValidator []int,
) {
	assignedCandidates := b.processRandomAssignment(candidates, rand, numberOfValidator)
	for shardID, tempCandidates := range assignedCandidates {
		tempCandidateStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(tempCandidates)
		b.committeeChange.ShardSubstituteAdded[shardID] = append(b.committeeChange.ShardSubstituteAdded[shardID], tempCandidateStructs...)
		b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], tempCandidates...)
	}
}

func (b *beaconCommitteeStateSlashingBase) processSwap(
	swapShardInstruction *instruction.SwapShardInstruction,
	env *BeaconCommitteeStateEnvironment,
) ([]string, []string, []string, error) {
	shardID := byte(swapShardInstruction.ChainID)
	newCommitteeChange := b.committeeChange
	committees := env.shardCommittee[shardID]
	substitutes := env.shardSubstitute[shardID]
	tempCommittees := common.DeepCopyString(committees)
	tempSubstitutes := common.DeepCopyString(substitutes)

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
		return []string{}, []string{}, []string{},
			fmt.Errorf("expect swap in keys %+v, got %+v",
				comparedShardSwapInstruction.InPublicKeys, swapShardInstruction.InPublicKeys)
	}

	if !reflect.DeepEqual(comparedShardSwapInstruction.OutPublicKeys, swapShardInstruction.OutPublicKeys) {
		return []string{}, []string{}, []string{},
			fmt.Errorf("expect swap out keys %+v, got %+v",
				comparedShardSwapInstruction.OutPublicKeys, swapShardInstruction.OutPublicKeys)
	}
	b.shardCommittee[shardID] = common.DeepCopyString(newCommittees)
	b.shardSubstitute[shardID] = b.shardSubstitute[shardID][len(swapShardInstruction.InPublicKeys):]

	newCommitteeChange.AddShardCommitteeRemoved(shardID, swapShardInstruction.OutPublicKeys)
	newCommitteeChange.AddShardSubstituteRemoved(shardID, swapShardInstruction.InPublicKeys)
	newCommitteeChange.AddShardCommitteeAdded(shardID, swapShardInstruction.InPublicKeys)
	b.committeeChange = newCommitteeChange
	return swapShardInstruction.InPublicKeys, normalSwapOutCommittees, slashingCommittees, nil
}

// processSwapShardInstruction update committees state by swap shard instruction
// Process single swap shard instruction for and update committee state
func (b *beaconCommitteeStateSlashingBase) processSwapShardInstruction(
	swapShardInstruction *instruction.SwapShardInstruction,
	numberOfValidator []int,
	env *BeaconCommitteeStateEnvironment,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
) (
	*instruction.ReturnStakeInstruction,
	error,
) {
	shardID := byte(swapShardInstruction.ChainID)

	_, normalSwapOutCommittees, slashingCommittees, err := b.processSwap(swapShardInstruction, env)
	if err != nil {
		return returnStakingInstruction, err
	}

	// process after swap for assign old committees to current shard pool
	returnStakingInstruction, err = b.processAfterNormalSwap(
		env,
		normalSwapOutCommittees,
		numberOfValidator,
		returnStakingInstruction,
	)
	if err != nil {
		return returnStakingInstruction, err
	}

	//process slashing after normal swap out
	returnStakingInstruction, err = b.processSlashing(
		shardID,
		env,
		slashingCommittees,
		returnStakingInstruction,
	)
	if err != nil {
		return returnStakingInstruction, err
	}

	return returnStakingInstruction, nil
}

func (b *beaconCommitteeStateSlashingBase) classifyValidatorsByAutoStake(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
) ([]string, *instruction.ReturnStakeInstruction, error) {
	candidates := []string{}
	outPublicKeyStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
	for index, outPublicKey := range outPublicKeys {
		stakerInfo, has, err := statedb.GetShardStakerInfo(env.ConsensusStateDB, outPublicKey)
		if err != nil {
			return candidates, returnStakingInstruction, err
		}
		if !has {
			return candidates, returnStakingInstruction, errors.Errorf("Can not found info of this public key %v", outPublicKey)
		}
		if stakerInfo.AutoStaking() {
			candidates = append(candidates, outPublicKey)
		} else {
			returnStakingInstruction = b.buildReturnStakingInstructionAndDeleteStakerInfo(
				returnStakingInstruction,
				outPublicKeyStructs[index],
				outPublicKey,
				stakerInfo,
			)
		}
	}

	return candidates, returnStakingInstruction, nil
}

// processAfterNormalSwap process swapped out committee public key
// - auto stake is false then remove completely out of any committee, candidate, substitute list
// - auto stake is true then using assignment rule v2 toassign this committee public key
func (b *beaconCommitteeStateSlashingBase) processAfterNormalSwap(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	numberOfValidator []int,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
) (*instruction.ReturnStakeInstruction, error) {
	candidates, returnStakingInstruction, err := b.classifyValidatorsByAutoStake(env, outPublicKeys, returnStakingInstruction)
	if err != nil {
		return returnStakingInstruction, err
	}
	b.assign(candidates, env.RandomNumber, numberOfValidator)

	return returnStakingInstruction, nil
}

// processSlashing process slashing committee public key
// force unstake and return staking amount for slashed committee
func (b *beaconCommitteeStateSlashingBase) processSlashing(
	shardID byte,
	env *BeaconCommitteeStateEnvironment,
	slashingPublicKeys []string,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
) (*instruction.ReturnStakeInstruction, error) {
	slashingPublicKeyStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(slashingPublicKeys)
	for index, outPublicKey := range slashingPublicKeys {
		stakerInfo, has, err := statedb.GetShardStakerInfo(env.ConsensusStateDB, outPublicKey)
		if err != nil {
			return returnStakingInstruction, err
		}
		if !has {
			return returnStakingInstruction, fmt.Errorf("Can not found info of this public key %v", outPublicKey)
		}
		returnStakingInstruction = b.buildReturnStakingInstructionAndDeleteStakerInfo(
			returnStakingInstruction,
			slashingPublicKeyStructs[index],
			outPublicKey,
			stakerInfo,
		)
	}
	b.committeeChange.AddSlashingCommittees(shardID, slashingPublicKeys)

	return returnStakingInstruction, nil
}

// processUnstakeInstruction : process unstake instruction from beacon block
func (b *beaconCommitteeStateSlashingBase) processUnstakeInstruction(
	unstakeInstruction *instruction.UnstakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
) (*instruction.ReturnStakeInstruction, error) {
	for index, publicKey := range unstakeInstruction.CommitteePublicKeys {
		if common.IndexOfStr(publicKey, env.newUnassignedCommonPool) == -1 {
			// if found in committee list then turn off auto staking
			if _, ok := b.autoStake[publicKey]; ok {
				b.turnOffStopAutoStake(publicKey)
			}
		} else {
			indexCandidate := common.IndexOfStr(publicKey, b.shardCommonPool)
			b.shardCommonPool = append(b.shardCommonPool[:indexCandidate], b.shardCommonPool[indexCandidate+1:]...)
			stakerInfo, has, err := statedb.GetShardStakerInfo(env.ConsensusStateDB, publicKey)
			if err != nil {
				return returnStakingInstruction, err
			}
			if !has {
				return returnStakingInstruction, errors.New("Can't find staker info")
			}
			b.committeeChange.AddNextEpochShardCandidateRemoved([]string{unstakeInstruction.CommitteePublicKeys[index]})

			returnStakingInstruction = b.buildReturnStakingInstructionAndDeleteStakerInfo(
				returnStakingInstruction,
				unstakeInstruction.CommitteePublicKeysStruct[index],
				publicKey,
				stakerInfo,
			)
		}
	}

	return returnStakingInstruction, nil
}
