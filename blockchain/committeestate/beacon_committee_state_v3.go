package committeestate

import (
	"fmt"
	"reflect"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeStateV3 struct {
	beaconCommitteeStateSlashingBase
	syncPool map[byte][]string
}

func NewBeaconCommitteeStateV3() *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
		beaconCommitteeStateSlashingBase: *newBeaconCommitteeStateSlashingBase(),
		syncPool:                         make(map[byte][]string),
	}
}

func NewBeaconCommitteeStateV3WithValue(
	beaconCommittee []string,
	shardCommittee map[byte][]string,
	shardSubstitute map[byte][]string,
	shardCommonPool []string,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	syncPool map[byte][]string,
	swapRule SwapRuleProcessor,
) *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
		beaconCommitteeStateSlashingBase: *newBeaconCommitteeStateSlashingBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute, autoStake, rewardReceiver, stakingTx,
			shardCommonPool, numberOfAssignedCandidates, swapRule,
		),
		syncPool: syncPool,
	}
}

func (b *BeaconCommitteeStateV3) Version() int {
	return DCS_VERSION
}

func (b *BeaconCommitteeStateV3) Clone() BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.clone()
}

func (b *BeaconCommitteeStateV3) clone() *BeaconCommitteeStateV3 {
	newB := NewBeaconCommitteeStateV3()
	newB.beaconCommitteeStateSlashingBase = *b.beaconCommitteeStateSlashingBase.clone()

	for i, v := range b.syncPool {
		newB.syncPool[i] = common.DeepCopyString(v)
	}

	return newB
}

func (b BeaconCommitteeStateV3) GetSyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	b.mu.RLock()
	defer b.mu.RUnlock()
	res := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range b.syncPool {
		res[k], _ = incognitokey.CommitteeBase58KeyListToStruct(v)
	}
	return res
}

func (b BeaconCommitteeStateV3) Hash() (*BeaconCommitteeStateHash, error) {
	if b.isEmpty() {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}

	hashes, err := b.beaconCommitteeStateSlashingBase.Hash()
	if err != nil {
		return nil, err
	}

	tempShardCandidateHash, err := common.GenerateHashFromStringArray(b.shardCommonPool)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	hashes.ShardCandidateHash = tempShardCandidateHash

	syncPool := make(map[byte][]string)
	for shardID, keys := range b.syncPool {
		syncPool[shardID] = keys
	}

	tempSyncPoolHash, err := common.GenerateHashFromMapByteString(syncPool)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	hashes.ShardSyncValidatorsHash = tempSyncPoolHash

	return hashes, nil
}
func (b BeaconCommitteeStateV3) isEmpty() bool {
	return reflect.DeepEqual(b, NewBeaconCommitteeStateV3())
}

func initGenesisBeaconCommitteeStateV3(env *BeaconCommitteeStateEnvironment) *BeaconCommitteeStateV3 {
	beaconCommitteeStateV3 := NewBeaconCommitteeStateV3()
	beaconCommitteeStateV3.initCommitteeState(env)
	return beaconCommitteeStateV3
}

func (b *BeaconCommitteeStateV3) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	var err error
	incurredInstructions := [][]string{}
	returnStakingInstruction := instruction.NewReturnStakeIns()
	committeeChange := NewCommitteeChange()
	b.mu.Lock()
	defer b.mu.Unlock()

	// snapshot shard common pool in beacon random time
	if env.IsBeaconRandomTime {
		b.numberOfAssignedCandidates = SnapshotShardCommonPoolV2(
			b.shardCommonPool,
			b.shardCommittee,
			b.shardSubstitute,
			env.NumberOfFixedShardBlockValidator,
			env.MinShardCommitteeSize,
			b.swapRule,
		)

		Logger.log.Infof("Block %+v, Number of Snapshot to Assign Candidate %+v", env.BeaconHeight, b.numberOfAssignedCandidates)
	}

	b.addData(env)

	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		switch inst[0] {
		case instruction.STAKE_ACTION:
			stakeInstruction, err := instruction.ValidateAndImportStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, err = b.processStakeInstruction(stakeInstruction, committeeChange)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

		case instruction.RANDOM_ACTION:
			randomInstruction, err := instruction.ValidateAndImportRandomInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = b.processAssignWithRandomInstruction(
				randomInstruction.RandomNumber(), env.numberOfValidator, committeeChange)

		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = b.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange)

		case instruction.SWAP_SHARD_ACTION:
			swapShardInstruction, err := instruction.ValidateAndImportSwapShardInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = b.processSwapShardInstruction(
				swapShardInstruction, env, committeeChange, returnStakingInstruction)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

		case instruction.FINISH_SYNC_ACTION:
			finishSyncInstruction, err := instruction.ValidateAndImportFinishSyncInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange = b.processFinishSyncInstruction(
				finishSyncInstruction, env, committeeChange)
		case instruction.UNSTAKE_ACTION:
			unstakeInstruction, err := instruction.ValidateAndImportUnstakeInstructionFromString(inst)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange, returnStakingInstruction, err = b.processUnstakeInstruction(
				unstakeInstruction, env, committeeChange, returnStakingInstruction)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		}
	}

	hashes, err := b.Hash()
	if err != nil {
		return hashes, committeeChange, incurredInstructions, err
	}
	if !returnStakingInstruction.IsEmpty() {
		incurredInstructions = append(incurredInstructions, returnStakingInstruction.ToString())
	}
	return hashes, committeeChange, incurredInstructions, nil
}

//assignToSync assign validatrors to syncPool
// update beacon committee state and committeechange
// UPDATE SYNC POOL ONLY
func (b *BeaconCommitteeStateV3) assignToSync(
	shardID byte,
	candidates []string,
	committeeChange *CommitteeChange,
) *CommitteeChange {
	committeeChange.AddSyncingPoolAdded(shardID, candidates)
	b.syncPool[shardID] = append(b.syncPool[shardID], candidates...)
	return committeeChange
}

//assignRandomlyToSubstituteList assign candidates to pending list
// update beacon state and committeeChange
// UPDATE PENDING LIST ONLY
func (b *BeaconCommitteeStateV3) assignRandomlyToSubstituteList(candidates []string, rand int64, shardID byte, committeeChange *CommitteeChange) *CommitteeChange {
	for _, candidate := range candidates {
		committeeChange.AddShardSubstituteAdded(shardID, []string{candidate})
		randomOffset := 0
		if len(b.shardSubstitute[shardID]) != 0 {
			randomOffset = calculateNewSubstitutePosition(candidate, rand, len(b.shardSubstitute[shardID]))
		}
		b.shardSubstitute[shardID] = insertValueToSliceByIndex(b.shardSubstitute[shardID], candidate, randomOffset)
	}
	return committeeChange
}

//assignToPending assign candidates to pending list
// update beacon state and committeeChange
// UPDATE PENDING LIST ONLY
func (b *BeaconCommitteeStateV3) assignBackToSubstituteList(candidates []string, shardID byte, committeeChange *CommitteeChange) *CommitteeChange {
	for _, candidate := range candidates {
		committeeChange.AddShardSubstituteAdded(shardID, []string{candidate})
		b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], candidate)
	}
	return committeeChange
}

func (b *BeaconCommitteeStateV3) processAfterNormalSwap(
	env *BeaconCommitteeStateEnvironment,
	outPublicKeys []string,
	committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
) (*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	newCommitteeChange := committeeChange
	candidates, newCommitteeChange, returnStakingInstruction, err := b.getValidatorsByAutoStake(env, outPublicKeys, newCommitteeChange, returnStakingInstruction)
	if err != nil {
		return newCommitteeChange, returnStakingInstruction, err
	}
	newReturnStakingInstruction := returnStakingInstruction
	newCommitteeChange = b.assignBackToSubstituteList(candidates, env.ShardID, newCommitteeChange)
	return newCommitteeChange, newReturnStakingInstruction, nil
}

//processAssignWithRandomInstruction assign candidates to syncPool
// update beacon state and committeechange
func (b *BeaconCommitteeStateV3) processAssignWithRandomInstruction(
	rand int64,
	numberOfValidator []int,
	committeeChange *CommitteeChange,
) *CommitteeChange {
	newCommitteeChange, candidates := b.getCandidatesForRandomAssignment(committeeChange)
	assignedCandidates := b.assignCandidates(candidates, rand, numberOfValidator)
	for shardID, candidates := range assignedCandidates {
		newCommitteeChange = b.assignToSync(shardID, candidates, newCommitteeChange)
	}
	return newCommitteeChange
}

func (b *BeaconCommitteeStateV3) processSwapShardInstruction(
	swapShardInstruction *instruction.SwapShardInstruction,
	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange,
	returnStakingInstruction *instruction.ReturnStakeInstruction,
) (
	*CommitteeChange, *instruction.ReturnStakeInstruction, error) {
	shardID := byte(swapShardInstruction.ChainID)
	env.ShardID = shardID

	// process normal swap out
	newCommitteeChange, _, normalSwapOutCommittees, slashingCommittees, err := b.processSwap(swapShardInstruction, env, committeeChange)

	// process after swap for assign old committees to current shard pool
	newCommitteeChange, returnStakingInstruction, err = b.processAfterNormalSwap(env,
		normalSwapOutCommittees,
		newCommitteeChange,
		returnStakingInstruction,
	)
	if err != nil {
		return nil, returnStakingInstruction, err
	}

	// process slashing after normal swap out
	returnStakingInstruction, newCommitteeChange, err = b.processSlashing(
		shardID,
		env,
		slashingCommittees,
		returnStakingInstruction,
		newCommitteeChange,
	)
	if err != nil {
		return nil, returnStakingInstruction, err
	}

	return newCommitteeChange, returnStakingInstruction, nil
}

// removeValidatorsFromSyncPool remove validator in sync pool regardless input ordered or sync pool ordered
func (b *BeaconCommitteeStateV3) removeValidatorsFromSyncPool(validators []string, shardID byte) {
	finishedSyncValidators := make(map[string]bool)
	for _, v := range validators {
		finishedSyncValidators[v] = true
	}
	originSyncPoolLength := 0
	for i := 0; i < len(b.syncPool[shardID]); {
		if originSyncPoolLength == len(validators) {
			break
		}
		key := b.syncPool[shardID][i]
		if finishedSyncValidators[key] {
			b.syncPool[shardID] = append(b.syncPool[shardID][:i], b.syncPool[shardID][i+1:]...)
			i--
			originSyncPoolLength++
		}
		i++
	}
}

//processFinishSyncInstruction move validators from pending to sync pool
// validators MUST in sync pool
func (b *BeaconCommitteeStateV3) processFinishSyncInstruction(
	finishSyncInstruction *instruction.FinishSyncInstruction,
	env *BeaconCommitteeStateEnvironment, committeeChange *CommitteeChange,
) *CommitteeChange {
	committeeChange.AddSyncingPoolRemoved(byte(finishSyncInstruction.ChainID), finishSyncInstruction.PublicKeys)
	committeeChange.AddFinishedSyncValidators(byte(finishSyncInstruction.ChainID), finishSyncInstruction.PublicKeys)
	b.removeValidatorsFromSyncPool(finishSyncInstruction.PublicKeys, byte(finishSyncInstruction.ChainID))

	committeeChange = b.
		assignRandomlyToSubstituteList(
			finishSyncInstruction.PublicKeys,
			env.RandomNumber,
			byte(finishSyncInstruction.ChainID),
			committeeChange)

	return committeeChange
}

func (b *BeaconCommitteeStateV3) addData(env *BeaconCommitteeStateEnvironment) {
	env.newUnassignedCommonPool = common.DeepCopyString(b.shardCommonPool[b.numberOfAssignedCandidates:])
	env.newAllSubstituteCommittees, _ = b.getAllSubstituteCommittees()
	env.newAllRoles = append([]string{}, env.newUnassignedCommonPool...)
	env.newAllRoles = append(env.newAllRoles, env.newAllSubstituteCommittees...)
	for _, syncPoolValidators := range b.syncPool {
		env.newAllRoles = append(env.newAllRoles, syncPoolValidators...)
	}
	env.shardCommittee = make(map[byte][]string)
	for shardID, committees := range b.shardCommittee {
		env.shardCommittee[shardID] = common.DeepCopyString(committees)
	}
	env.shardSubstitute = make(map[byte][]string)
	for shardID, substitutes := range b.shardSubstitute {
		env.shardSubstitute[shardID] = common.DeepCopyString(substitutes)
	}
	env.numberOfValidator = make([]int, env.ActiveShards)
	for i := 0; i < env.ActiveShards; i++ {
		env.numberOfValidator[i] += len(b.shardCommittee[byte(i)])
		env.numberOfValidator[i] += len(b.shardSubstitute[byte(i)])
	}

}

func (b *BeaconCommitteeStateV3) AllSyncingValidators() []string {
	res := []string{}
	for _, syncingValidators := range b.syncPool {
		res = append(res, syncingValidators...)
	}
	return res
}

//SplitReward ...
func (b *BeaconCommitteeStateV3) SplitReward(
	env *SplitRewardEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64,
	map[common.Hash]uint64, map[common.Hash]uint64, error,
) {
	devPercent := uint64(env.DAOPercent)
	allCoinTotalReward := env.TotalReward // total reward for shard subset
	rewardForBeacon := map[common.Hash]uint64{}
	rewardForShardSubset := map[common.Hash]uint64{}
	rewardForIncDAO := map[common.Hash]uint64{}
	rewardForCustodian := map[common.Hash]uint64{}
	lenBeaconCommittees := uint64(len(b.getBeaconCommittee()))
	lenShardSubsetCommittees := uint64(len(b.getShardCommittee()[env.ShardID]) / int(env.MaxSubsetCommittees))
	if len(b.getShardCommittee()[env.ShardID])%int(env.MaxSubsetCommittees) != 0 {
		if (env.SubsetID % env.MaxSubsetCommittees) == 0 {
			lenShardSubsetCommittees += uint64(len(b.getShardCommittee()[env.ShardID]) % int(env.MaxSubsetCommittees))
		}
	}

	if len(allCoinTotalReward) == 0 {
		Logger.log.Info("Beacon Height %+v, 😭 found NO reward", env.BeaconHeight)
		return rewardForBeacon, rewardForShardSubset, rewardForIncDAO, rewardForCustodian, nil
	}

	for coinID, totalReward := range allCoinTotalReward {
		totalRewardForDAOAndCustodians := devPercent * totalReward / 100
		totalRewardForShardAndBeaconValidators := totalReward - totalRewardForDAOAndCustodians
		shardSubsetWeight := float64(lenShardSubsetCommittees)
		beaconWeight := float64(lenBeaconCommittees) / float64(len(b.shardCommittee))
		totalValidatorWeight := shardSubsetWeight + beaconWeight

		rewardForShardSubset[coinID] = uint64(shardSubsetWeight * float64(totalRewardForShardAndBeaconValidators) / totalValidatorWeight)
		Logger.log.Infof("totalRewardForDAOAndCustodians tokenID %v - %v\n", coinID.String(), totalRewardForDAOAndCustodians)

		if env.IsSplitRewardForCustodian {
			rewardForCustodian[coinID] += env.PercentCustodianReward * totalRewardForDAOAndCustodians / 100
			rewardForIncDAO[coinID] += totalRewardForDAOAndCustodians - rewardForCustodian[coinID]
		} else {
			rewardForIncDAO[coinID] += totalRewardForDAOAndCustodians
		}
		rewardForBeacon[coinID] += totalReward - (rewardForShardSubset[coinID] + totalRewardForDAOAndCustodians)
	}
	return rewardForBeacon, rewardForShardSubset, rewardForIncDAO, rewardForCustodian, nil
}

func (b *BeaconCommitteeStateV3) GetAllCandidateSubstituteCommittee() []string {
	return b.getAllCandidateSubstituteCommittee()
}

func (b *BeaconCommitteeStateV3) getAllCandidateSubstituteCommittee() []string {
	res := b.beaconCommitteeStateSlashingBase.getAllCandidateSubstituteCommittee()
	for _, validators := range b.syncPool {
		res = append(res, validators...)
	}
	return res
}
