package committeestate

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/pkg/errors"
)

type BeaconCommitteeStateV1 struct {
	beaconCommitteeStateBase
	currentEpochShardCandidate []string
	nextEpochShardCandidate    []string
}

func NewBeaconCommitteeStateV1() *BeaconCommitteeStateV1 {
	return &BeaconCommitteeStateV1{
		beaconCommitteeStateBase: *newBeaconCommitteeStateBase(),
	}
}

func NewBeaconCommitteeStateV1WithValue(
	beaconCommittee []string,
	nextEpochShardCandidate []string,
	currentEpochShardCandidate []string,
	shardCurrentValidator map[byte][]string,
	shardSubstituteValidator map[byte][]string,
	autoStaking map[string]bool,
	rewardReceivers map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
) *BeaconCommitteeStateV1 {
	return &BeaconCommitteeStateV1{
		beaconCommitteeStateBase: *newBeaconCommitteeStateBaseWithValue(
			beaconCommittee, shardCurrentValidator, shardSubstituteValidator,
			autoStaking, rewardReceivers, stakingTx,
		),
		nextEpochShardCandidate:    nextEpochShardCandidate,
		currentEpochShardCandidate: currentEpochShardCandidate,
	}
}

func (b *BeaconCommitteeStateV1) Version() int {
	return SELF_SWAP_SHARD_VERSION
}

func (b *BeaconCommitteeStateV1) AssignRuleVersion() int {
	return ASSIGN_RULE_V1
}

func (b BeaconCommitteeStateV1) shallowCopy(newB *BeaconCommitteeStateV1) {
	newB.beaconCommitteeStateBase = b.beaconCommitteeStateBase
	newB.currentEpochShardCandidate = b.currentEpochShardCandidate
	newB.nextEpochShardCandidate = b.nextEpochShardCandidate
}

func (b *BeaconCommitteeStateV1) Clone(db *statedb.StateDB) BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.clone()
}

func (b *BeaconCommitteeStateV1) clone() *BeaconCommitteeStateV1 {
	newB := NewBeaconCommitteeStateV1()
	newB.beaconCommitteeStateBase = *b.beaconCommitteeStateBase.clone()
	newB.currentEpochShardCandidate = make([]string, len(b.currentEpochShardCandidate))
	copy(newB.currentEpochShardCandidate, b.currentEpochShardCandidate)
	newB.nextEpochShardCandidate = make([]string, len(b.nextEpochShardCandidate))
	copy(newB.nextEpochShardCandidate, b.nextEpochShardCandidate)
	return newB
}

func (b *BeaconCommitteeStateV1) reset() {
	b.beaconCommitteeStateBase.reset()
	b.currentEpochShardCandidate = []string{}
	b.nextEpochShardCandidate = []string{}
}

func (b BeaconCommitteeStateV1) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(b.nextEpochShardCandidate)
	return res
}

func (b BeaconCommitteeStateV1) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(b.currentEpochShardCandidate)
	return res
}

func (b BeaconCommitteeStateV1) GetShardCommonPool() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(b.currentEpochShardCandidate)
	res2, _ := incognitokey.CommitteeBase58KeyListToStruct(b.nextEpochShardCandidate)
	return append(res, res2...)
}

func (b BeaconCommitteeStateV1) GetAllCandidateSubstituteCommittee() []string {
	return b.getAllCandidateSubstituteCommittee()
}

func (b *BeaconCommitteeStateV1) getAllCandidateSubstituteCommittee() []string {
	res := []string{}
	for _, committee := range b.shardCommittee {
		res = append(res, committee...)
	}
	for _, substitute := range b.shardSubstitute {
		res = append(res, substitute...)
	}
	res = append(res, b.beaconCommittee...)
	res = append(res, b.currentEpochShardCandidate...)
	res = append(res, b.nextEpochShardCandidate...)
	return res
}

func (b *BeaconCommitteeStateV1) Hash(committeeChange *CommitteeChange) (*BeaconCommitteeStateHash, error) {
	if b.isEmpty() {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}

	res, err := b.beaconCommitteeStateBase.Hash(committeeChange)
	if err != nil {
		return res, err
	}

	var tempShardCandidateHash common.Hash
	if !isNilOrShardCandidateHash(b.hashes) &&
		len(committeeChange.NextEpochShardCandidateRemoved) == 0 && len(committeeChange.NextEpochShardCandidateAdded) == 0 &&
		len(committeeChange.CurrentEpochShardCandidateAdded) == 0 && len(committeeChange.CurrentEpochShardCandidateRemoved) == 0 {
		tempShardCandidateHash = b.hashes.ShardCandidateHash
	} else {
		shardCandidateArr := append([]string{}, b.currentEpochShardCandidate...)
		shardCandidateArr = append(shardCandidateArr, b.nextEpochShardCandidate...)
		// Shard candidate root: shard current candidate + shard next candidate
		tempShardCandidateHash, err = common.GenerateHashFromStringArray(shardCandidateArr)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
	}

	res.ShardCandidateHash = tempShardCandidateHash

	return res, nil
}

func initGenesisBeaconCommitteeStateV1(env *BeaconCommitteeStateEnvironment) *BeaconCommitteeStateV1 {
	beaconCommitteeStateV1 := NewBeaconCommitteeStateV1()
	beaconCommitteeStateV1.initCommitteeState(env)
	return beaconCommitteeStateV1
}

//UpdateCommitteeState :
func (b *BeaconCommitteeStateV1) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.setBeaconCommitteeStateHashes(env.PreviousBlockHashes)
	incurredInstructions := [][]string{}
	committeeChange := NewCommitteeChange()

	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		tempNewShardCandidates := []string{}
		switch inst[0] {
		case instruction.STAKE_ACTION:
			stakeInstruction, err := instruction.ValidateAndImportStakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stake instruction %+v, error %+v", inst, err)
				continue
			}
			_, tempNewShardCandidates, err = b.processStakeInstruction(stakeInstruction, env)
			if err != nil {
				Logger.log.Errorf("SKIP stake instruction %+v, error %+v", inst, err)
				continue
			}
		case instruction.SWAP_ACTION:
			swapInstruction, err := instruction.ValidateAndImportSwapInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP swap instruction %+v, error %+v", inst, err)
				continue
			}
			_, tempNewShardCandidates, err = b.processSwapInstruction(swapInstruction, env, committeeChange)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
				continue
			}
			b.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange)
		}
		if len(tempNewShardCandidates) > 0 {
			b.nextEpochShardCandidate = append(b.nextEpochShardCandidate, tempNewShardCandidates...)
			newShardCandidates, err := incognitokey.CommitteeBase58KeyListToStruct(tempNewShardCandidates)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, newShardCandidates...)
		}

	}
	if env.IsBeaconRandomTime {
		committeeChange.CurrentEpochShardCandidateAdded, _ = incognitokey.CommitteeBase58KeyListToStruct(b.nextEpochShardCandidate)
		b.currentEpochShardCandidate = b.nextEpochShardCandidate
		Logger.log.Debug("Beacon Process: CandidateShardWaitingForCurrentRandom: ", b.currentEpochShardCandidate)
		// reset candidate list
		committeeChange.NextEpochShardCandidateRemoved, _ = incognitokey.CommitteeBase58KeyListToStruct(b.nextEpochShardCandidate)
		b.nextEpochShardCandidate = []string{}
	}
	if env.IsFoundRandomNumber {
		numberOfShardSubstitutes := make(map[byte]int)
		for shardID, shardSubstitute := range b.shardSubstitute {
			numberOfShardSubstitutes[shardID] = len(shardSubstitute)
		}
		shardCandidatesStr := make([]string, len(b.currentEpochShardCandidate))
		copy(shardCandidatesStr, b.currentEpochShardCandidate)
		remainShardCandidates, assignedCandidates := assignShardCandidate(shardCandidatesStr, numberOfShardSubstitutes, env.RandomNumber, env.AssignOffset, env.ActiveShards)
		remainShardCandidatesStr, err := incognitokey.CommitteeBase58KeyListToStruct(remainShardCandidates)
		if err != nil {
			return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
		}
		committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, remainShardCandidatesStr...)
		// append remain candidate into shard waiting for next random list
		b.nextEpochShardCandidate = append(b.nextEpochShardCandidate, remainShardCandidates...)
		// assign candidate into shard pending validator list
		for shardID, candidateList := range assignedCandidates {
			candidateListStr, err := incognitokey.CommitteeBase58KeyListToStruct(candidateList)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange.ShardSubstituteAdded[shardID] = candidateListStr
			b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], candidateList...)
		}
		committeeChange.CurrentEpochShardCandidateRemoved, _ = incognitokey.CommitteeBase58KeyListToStruct(b.currentEpochShardCandidate)
		// delete CandidateShardWaitingForCurrentRandom list
		b.currentEpochShardCandidate = []string{}
		// shuffle CandidateBeaconWaitingForCurrentRandom with current random number
	}

	err := b.processAutoStakingChange(committeeChange, env)
	if err != nil {
		return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}
	hashes, err := b.Hash(committeeChange)
	if err != nil {
		return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}
	return hashes, committeeChange, incurredInstructions, nil
}

func (b *BeaconCommitteeStateV1) processStakeInstruction(
	stakeInstruction *instruction.StakeInstruction,
	env *BeaconCommitteeStateEnvironment,
) ([]string, []string, error) {
	committeeChange := NewCommitteeChange()
	committeeChange, err := b.beaconCommitteeStateBase.processStakeInstruction(stakeInstruction, committeeChange)
	if err != nil {
		return []string{}, []string{}, err
	}
	newShardCandidates, _ := incognitokey.CommitteeKeyListToString(committeeChange.NextEpochShardCandidateAdded)

	err = statedb.StoreStakerInfo(
		env.ConsensusStateDB,
		stakeInstruction.PublicKeyStructs,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
		env.BeaconHeight,
	)

	if err != nil {
		return []string{}, newShardCandidates, err
	}

	return []string{}, newShardCandidates, nil
}

func (b *BeaconCommitteeStateV1) processStopAutoStakeInstruction(
	stopAutoStakeInstruction *instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) {
	for _, committeePublicKey := range stopAutoStakeInstruction.CommitteePublicKeys {
		if common.IndexOfStr(committeePublicKey, b.getAllCandidateSubstituteCommittee()) == -1 {
			// if not found then delete auto staking data for this public key if present
			if _, ok := b.autoStake[committeePublicKey]; ok {
				delete(b.autoStake, committeePublicKey)
			}
		} else {
			// if found in committee list then turn off auto staking
			if _, ok := b.autoStake[committeePublicKey]; ok {
				b.autoStake[committeePublicKey] = false
				committeeChange.StopAutoStake = append(committeeChange.StopAutoStake, committeePublicKey)
			}
		}
	}
}

func (b *BeaconCommitteeStateV1) processSwapInstruction(
	swapInstruction *instruction.SwapInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) ([]string, []string, error) {
	newBeaconCandidates := []string{}
	newShardCandidates := []string{}
	if common.IndexOfUint64(env.BeaconHeight/env.EpochLengthV1, env.EpochBreakPointSwapNewKey) > -1 || swapInstruction.IsReplace {
		err := b.processReplaceInstruction(swapInstruction, committeeChange, env)
		if err != nil {
			return newBeaconCandidates, newShardCandidates, err
		}
	} else {
		Logger.log.Debug("Swap Instruction In Public Keys", swapInstruction.InPublicKeys)
		Logger.log.Debug("Swap Instruction Out Public Keys", swapInstruction.OutPublicKeys)
		if swapInstruction.ChainID != instruction.BEACON_CHAIN_ID {
			shardID := byte(swapInstruction.ChainID)
			// delete in public key out of sharding pending validator list
			if len(swapInstruction.InPublicKeys) > 0 {
				shardSubstituteStr := make([]string, len(b.shardSubstitute[shardID]))
				copy(shardSubstituteStr, b.shardSubstitute[shardID])
				tempShardSubstitute, err := removeValidatorV1(shardSubstituteStr, swapInstruction.InPublicKeys)
				if err != nil {
					return newBeaconCandidates, newShardCandidates, err
				}
				// update shard pending validator
				committeeChange.ShardSubstituteRemoved[shardID] = append(committeeChange.ShardSubstituteRemoved[shardID], swapInstruction.InPublicKeyStructs...)
				b.shardSubstitute[shardID] = make([]string, len(tempShardSubstitute))
				copy(b.shardSubstitute[shardID], tempShardSubstitute)
				// add new public key to committees
				committeeChange.ShardCommitteeAdded[shardID] = append(committeeChange.ShardCommitteeAdded[shardID], swapInstruction.InPublicKeyStructs...)
				b.shardCommittee[shardID] = append(b.shardCommittee[shardID], swapInstruction.InPublicKeys...)
			}
			// delete out public key out of current committees
			if len(swapInstruction.OutPublicKeys) > 0 {
				//for _, value := range outPublickeyStructs {
				//	delete(b,cue.GetIncKeyBase58(
				//}
				tempShardCommittees, err := removeValidatorV1(b.shardCommittee[shardID], swapInstruction.OutPublicKeys)
				if err != nil {
					return newBeaconCandidates, newShardCandidates, err
				}
				b.shardCommittee[shardID] = make([]string, len(tempShardCommittees))
				copy(b.shardCommittee[shardID], tempShardCommittees)
				// remove old public key in shard committee update shard committee
				committeeChange.ShardCommitteeRemoved[shardID] = append(committeeChange.ShardCommitteeRemoved[shardID], swapInstruction.OutPublicKeyStructs...)
				// Check auto stake in out public keys list
				// if auto staking not found or flag auto stake is false then do not re-stake for this out public key
				// if auto staking flag is true then system will automatically add this out public key to current candidate list
				for index, outPublicKey := range swapInstruction.OutPublicKeys {
					stakerInfo, has, err := statedb.GetShardStakerInfo(env.ConsensusStateDB, outPublicKey)
					if err != nil {
						panic(err)
					}
					if !has {
						panic(errors.Errorf("Can not found info of this public key %v", outPublicKey))
					}
					if stakerInfo.AutoStaking() {
						newShardCandidates = append(newShardCandidates, swapInstruction.OutPublicKeys[index])
					} else {
						delete(b.rewardReceiver, swapInstruction.OutPublicKeyStructs[index].GetIncKeyBase58())
						delete(b.autoStake, outPublicKey)
						delete(b.stakingTx, outPublicKey)
					}
				}
			}
		}
	}
	return newBeaconCandidates, newShardCandidates, nil
}

func (b *BeaconCommitteeStateV1) processReplaceInstruction(
	swapInstruction *instruction.SwapInstruction,
	committeeChange *CommitteeChange,
	env *BeaconCommitteeStateEnvironment,
) error {
	removedCommittee := len(swapInstruction.InPublicKeys)
	if swapInstruction.ChainID == instruction.BEACON_CHAIN_ID {
		committeeChange.BeaconCommitteeReplaced[common.REPLACE_OUT] = append(committeeChange.BeaconCommitteeReplaced[common.REPLACE_OUT], swapInstruction.OutPublicKeyStructs...)
		// add new public key to committees
		committeeChange.BeaconCommitteeReplaced[common.REPLACE_IN] = append(committeeChange.BeaconCommitteeReplaced[common.REPLACE_IN], swapInstruction.InPublicKeyStructs...)
		remainedBeaconCommittees := make([]string, len(b.beaconCommittee[removedCommittee:]))
		copy(remainedBeaconCommittees, b.beaconCommittee[removedCommittee:])
		newCommittees := make([]string, len(swapInstruction.InPublicKeys))
		copy(newCommittees, swapInstruction.InPublicKeys)
		b.beaconCommittee = append(newCommittees, remainedBeaconCommittees...)
	} else {
		shardID := byte(swapInstruction.ChainID)
		committeeReplace := [2][]incognitokey.CommitteePublicKey{}
		// update shard COMMITTEE
		committeeReplace[common.REPLACE_OUT] = append(committeeReplace[common.REPLACE_OUT], swapInstruction.OutPublicKeyStructs...)
		// add new public key to committees
		committeeReplace[common.REPLACE_IN] = append(committeeReplace[common.REPLACE_IN], swapInstruction.InPublicKeyStructs...)
		committeeChange.ShardCommitteeReplaced[shardID] = committeeReplace
		remainedShardCommittees := b.shardCommittee[shardID][removedCommittee:]
		newCommittees := make([]string, len(swapInstruction.InPublicKeys))
		copy(newCommittees, swapInstruction.InPublicKeys)
		b.shardCommittee[shardID] = append(newCommittees, remainedShardCommittees...)
	}
	for index := 0; index < removedCommittee; index++ {
		delete(b.autoStake, swapInstruction.OutPublicKeys[index])
		delete(b.stakingTx, swapInstruction.OutPublicKeys[index])
		delete(b.rewardReceiver, swapInstruction.OutPublicKeyStructs[index].GetIncKeyBase58())
		b.autoStake[swapInstruction.InPublicKeys[index]] = false
		b.rewardReceiver[swapInstruction.InPublicKeyStructs[index].GetIncKeyBase58()] = swapInstruction.NewRewardReceiverStructs[index]
		b.stakingTx[swapInstruction.InPublicKeys[index]] = common.HashH([]byte{0})
	}
	err := statedb.StoreStakerInfo(
		env.ConsensusStateDB,
		swapInstruction.InPublicKeyStructs,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
		env.BeaconHeight,
	)
	return err
}

func (b *BeaconCommitteeStateV1) processAutoStakingChange(committeeChange *CommitteeChange, env *BeaconCommitteeStateEnvironment) error {
	stopAutoStakingIncognitoKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeeChange.StopAutoStake)
	if err != nil {
		return err
	}

	if len(stopAutoStakingIncognitoKey) != 0 {
		err := statedb.SaveStopAutoStakerInfo(env.ConsensusStateDB, stopAutoStakingIncognitoKey, b.autoStake)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BeaconCommitteeStateV1) GenerateAssignInstructions(env *BeaconCommitteeStateEnvironment) []*instruction.AssignInstruction {
	candidates := make([]string, len(b.currentEpochShardCandidate))
	copy(candidates, b.currentEpochShardCandidate)
	numberOfPendingValidator := make(map[byte]int)
	shardPendingValidator := b.shardSubstitute
	for i := 0; i < len(b.shardCommittee); i++ {
		if pendingValidators, ok := shardPendingValidator[byte(i)]; ok {
			numberOfPendingValidator[byte(i)] = len(pendingValidators)
		} else {
			numberOfPendingValidator[byte(i)] = 0
		}
	}
	assignedCandidates := make(map[byte][]string)
	shuffledCandidate := shuffleShardCandidate(candidates, env.RandomNumber)
	for _, candidate := range shuffledCandidate {
		shardID := calculateCandidateShardID(candidate, env.RandomNumber, len(b.shardCommittee))
		if numberOfPendingValidator[shardID]+1 <= env.AssignOffset {
			assignedCandidates[shardID] = append(assignedCandidates[shardID], candidate)
			numberOfPendingValidator[shardID] += 1

		}
	}
	var keys []int
	for k := range assignedCandidates {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	instructions := []*instruction.AssignInstruction{}
	for _, key := range keys {
		shardID := byte(key)
		candidates := assignedCandidates[shardID]
		Logger.log.Infof("Assign Candidate at Shard %+v: %+v", shardID, candidates)
		shardAssignInstruction := instruction.NewAssignInstructionWithValue(int(shardID), candidates)
		instructions = append(instructions, shardAssignInstruction)
	}
	return instructions
}

func (b BeaconCommitteeStateV1) IsFinishSync(string) bool {
	panic("This should not be callsed")
}

//Upgrade check interface method for des
func (b *BeaconCommitteeStateV1) Upgrade(env *BeaconCommitteeStateEnvironment) BeaconCommitteeState {
	b.mu.RLock()
	defer b.mu.RUnlock()
	beaconCommittee, shardCommittee, shardSubstitute,
		shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule, assignRule := b.getDataForUpgrading(env)

	committeeStateV2 := NewBeaconCommitteeStateV2WithValue(
		beaconCommittee,
		shardCommittee,
		shardSubstitute,
		shardCommonPool,
		numberOfAssignedCandidates,
		autoStake,
		rewardReceiver,
		stakingTx,
		swapRule,
		assignRule,
	)
	Logger.log.Infof("Upgrade Committee State V2 to V3, swap rule %+v, assign rule $+v",
		reflect.TypeOf(swapRule), reflect.TypeOf(assignRule))
	return committeeStateV2
}

func (b *BeaconCommitteeStateV1) getDataForUpgrading(env *BeaconCommitteeStateEnvironment) (
	[]string,
	map[byte][]string,
	map[byte][]string,
	[]string,
	int,
	map[string]bool,
	map[string]privacy.PaymentAddress,
	map[string]common.Hash,
	SwapRuleProcessor,
	AssignRuleProcessor,
) {
	beaconCommittee := make([]string, len(b.beaconCommittee))
	shardCommittee := make(map[byte][]string)
	shardSubstitute := make(map[byte][]string)
	numberOfAssignedCandidates := len(b.currentEpochShardCandidate)
	autoStake := make(map[string]bool)
	rewardReceiver := make(map[string]privacy.PaymentAddress)
	stakingTx := make(map[string]common.Hash)

	copy(beaconCommittee, b.beaconCommittee)
	for shardID, oneShardCommittee := range b.shardCommittee {
		shardCommittee[shardID] = make([]string, len(oneShardCommittee))
		copy(shardCommittee[shardID], oneShardCommittee)
	}
	for shardID, oneShardSubstitute := range b.shardSubstitute {
		shardSubstitute[shardID] = make([]string, len(oneShardSubstitute))
		copy(shardSubstitute[shardID], oneShardSubstitute)
	}
	currentEpochShardCandidate := b.currentEpochShardCandidate
	nextEpochShardCandidate := b.nextEpochShardCandidate
	shardCandidates := append(currentEpochShardCandidate, nextEpochShardCandidate...)
	shardCommonPool := make([]string, len(shardCandidates))
	copy(shardCommonPool, shardCandidates)
	for k, v := range b.autoStake {
		autoStake[k] = v
	}
	for k, v := range b.rewardReceiver {
		rewardReceiver[k] = v
	}
	for k, v := range b.stakingTx {
		stakingTx[k] = v
	}

	swapRule := GetSwapRuleVersion(env.BeaconHeight, env.StakingV3Height)
	assignRule := GetAssignRuleVersion(env.BeaconHeight, env.StakingV2Height, env.AssignRuleV3Height)
	return beaconCommittee, shardCommittee, shardSubstitute, shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule, assignRule
}
