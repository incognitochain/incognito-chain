package committeestate

import (
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

type BeaconCommitteeEngineV1 struct {
	beaconCommitteeEngineBase
}

func NewBeaconCommitteeEngineV1(
	beaconHeight uint64,
	beaconHash common.Hash,
	beaconCommitteeStateV1 *BeaconCommitteeStateV1) *BeaconCommitteeEngineV1 {
	Logger.log.Infof("Init Beacon Committee Engine V1, %+v", beaconHeight)
	return &BeaconCommitteeEngineV1{
		beaconCommitteeEngineBase: *NewBeaconCommitteeEngineBaseWithValue(
			beaconHeight, beaconHash, beaconCommitteeStateV1,
		),
	}
}

//Clone :
func (engine *BeaconCommitteeEngineV1) Clone() BeaconCommitteeEngine {
	res := &BeaconCommitteeEngineV1{
		beaconCommitteeEngineBase: *engine.beaconCommitteeEngineBase.Clone().(*beaconCommitteeEngineBase),
	}
	return res
}

//Version :
func (engine BeaconCommitteeEngineV1) Version() uint {
	return SELF_SWAP_SHARD_VERSION
}

//GetCandidateShardWaitingForCurrentRandom :
func (engine BeaconCommitteeEngineV1) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return engine.finalState.CandidateShardWaitingForCurrentRandom()
}

//GetCandidateShardWaitingForNextRandom :
func (engine BeaconCommitteeEngineV1) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return engine.finalState.CandidateShardWaitingForNextRandom()
}

//UpdateCommitteeState :
func (engine *BeaconCommitteeEngineV1) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
	*BeaconCommitteeStateHash, *CommitteeChange, [][]string, error) {
	engine.finalState.Mu().RLock()
	if engine.uncommittedState == nil {
		engine.uncommittedState = NewBeaconCommitteeStateV1()
	}
	cloneBeaconCommitteeStateFromTo(engine.finalState, engine.uncommittedState)
	engine.finalState.Mu().RUnlock()
	var err error
	incurredInstructions := [][]string{}
	engine.finalState.Mu().Lock()
	defer engine.finalState.Mu().Unlock()

	newB := engine.uncommittedState.(*BeaconCommitteeStateV1)
	committeeChange := NewCommitteeChange()
	newBeaconCandidates := []incognitokey.CommitteePublicKey{}
	newShardCandidates := []incognitokey.CommitteePublicKey{}

	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		tempNewBeaconCandidates, tempNewShardCandidates := []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		switch inst[0] {
		case instruction.STAKE_ACTION:
			stakeInstruction, err := instruction.ValidateAndImportStakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stake instruction %+v, error %+v", inst, err)
				continue
			}
			tempNewBeaconCandidates, tempNewShardCandidates, err = newB.processStakeInstruction(stakeInstruction, env)
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
			tempNewBeaconCandidates, tempNewShardCandidates, err = newB.processSwapInstruction(swapInstruction, env, committeeChange)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
			}
			newB.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange)
		}
		if len(tempNewBeaconCandidates) > 0 {
			newBeaconCandidates = append(newBeaconCandidates, tempNewBeaconCandidates...)
		}
		if len(tempNewShardCandidates) > 0 {
			newShardCandidates = append(newShardCandidates, tempNewShardCandidates...)
		}
	}

	committeeChange.NextEpochBeaconCandidateAdded = append(committeeChange.NextEpochBeaconCandidateAdded, newBeaconCandidates...)
	newB.nextEpochShardCandidate = append(newB.nextEpochShardCandidate, newShardCandidates...)
	committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, newShardCandidates...)
	if env.IsBeaconRandomTime {
		committeeChange.CurrentEpochShardCandidateAdded = newB.nextEpochShardCandidate
		newB.currentEpochShardCandidate = newB.nextEpochShardCandidate
		Logger.log.Debug("Beacon Process: CandidateShardWaitingForCurrentRandom: ", newB.currentEpochShardCandidate)
		// reset candidate list
		committeeChange.NextEpochShardCandidateRemoved = newB.nextEpochShardCandidate
		newB.nextEpochShardCandidate = []incognitokey.CommitteePublicKey{}
	}
	if env.IsFoundRandomNumber {
		numberOfShardSubstitutes := make(map[byte]int)
		for shardID, shardSubstitute := range newB.shardSubstitute {
			numberOfShardSubstitutes[shardID] = len(shardSubstitute)
		}
		shardCandidatesStr, err := incognitokey.CommitteeKeyListToString(newB.currentEpochShardCandidate)
		if err != nil {
			return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
		}
		remainShardCandidatesStr, assignedCandidates := assignShardCandidate(shardCandidatesStr, numberOfShardSubstitutes, env.RandomNumber, env.AssignOffset, env.ActiveShards)
		remainShardCandidates, err := incognitokey.CommitteeBase58KeyListToStruct(remainShardCandidatesStr)
		if err != nil {
			return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
		}
		committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, remainShardCandidates...)
		// append remain candidate into shard waiting for next random list
		newB.nextEpochShardCandidate = append(newB.nextEpochShardCandidate, remainShardCandidates...)
		// assign candidate into shard pending validator list
		for shardID, candidateListStr := range assignedCandidates {
			candidateList, err := incognitokey.CommitteeBase58KeyListToStruct(candidateListStr)
			if err != nil {
				return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			committeeChange.ShardSubstituteAdded[shardID] = candidateList
			newB.shardSubstitute[shardID] = append(newB.shardSubstitute[shardID], candidateList...)
		}
		committeeChange.CurrentEpochShardCandidateRemoved = newB.currentEpochShardCandidate
		// delete CandidateShardWaitingForCurrentRandom list
		newB.currentEpochShardCandidate = []incognitokey.CommitteePublicKey{}
		// shuffle CandidateBeaconWaitingForCurrentRandom with current random number
	}

	err = newB.processAutoStakingChange(committeeChange, env)
	if err != nil {
		return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}
	hashes, err := engine.uncommittedState.Hash()
	if err != nil {
		return nil, nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}
	return hashes, committeeChange, incurredInstructions, nil
}

func (engine *BeaconCommitteeEngineV1) AssignInstructions(env *BeaconCommitteeStateEnvironment) []*instruction.AssignInstruction {
	candidates, _ := incognitokey.CommitteeKeyListToString(engine.finalState.CandidateShardWaitingForCurrentRandom())
	numberOfPendingValidator := make(map[byte]int)
	shardPendingValidator := engine.finalState.ShardSubstitute()
	for i := 0; i < len(engine.finalState.ShardCommittee()); i++ {
		if pendingValidators, ok := shardPendingValidator[byte(i)]; ok {
			numberOfPendingValidator[byte(i)] = len(pendingValidators)
		} else {
			numberOfPendingValidator[byte(i)] = 0
		}
	}
	assignedCandidates := make(map[byte][]string)
	shuffledCandidate := shuffleShardCandidate(candidates, env.RandomNumber)
	for _, candidate := range shuffledCandidate {
		shardID := calculateCandidateShardID(candidate, env.RandomNumber, len(engine.finalState.ShardCommittee()))
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

//SplitReward ...
func (b *BeaconCommitteeEngineV1) SplitReward(
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

	if len(allCoinTotalReward) == 0 {
		Logger.log.Info("Beacon Height %+v, 😭 found NO reward", env.BeaconHeight)
		return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
	}

	for key, totalReward := range allCoinTotalReward {
		rewardForBeacon[key] += 2 * ((100 - devPercent) * totalReward) / ((uint64(env.ActiveShards) + 2) * 100)
		totalRewardForDAOAndCustodians := uint64(devPercent) * totalReward / uint64(100)

		Logger.log.Infof("[test-salary] totalRewardForDAOAndCustodians tokenID %v - %v\n",
			key.String(), totalRewardForDAOAndCustodians)

		if env.IsSplitRewardForCustodian {
			rewardForCustodian[key] += uint64(env.PercentCustodianReward) * totalRewardForDAOAndCustodians / uint64(100)
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians - rewardForCustodian[key]
		} else {
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians
		}

		rewardForShard[key] = totalReward - (rewardForBeacon[key] + totalRewardForDAOAndCustodians)
	}

	return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
}

//IsSwapTime read from interface des
func (engine BeaconCommitteeEngineV1) IsSwapTime(beaconHeight, numberOfBlockEachEpoch uint64) bool {
	if beaconHeight%numberOfBlockEachEpoch == 0 {
		return true
	} else {
		return false
	}
}

//Upgrade check interface method for des
func (engine BeaconCommitteeEngineV1) Upgrade(env *BeaconCommitteeStateEnvironment) BeaconCommitteeEngine {
	beaconCommittee, shardCommittee, shardSubstitute,
		shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule := engine.getDataForUpgrading(env)

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
	)
	committeeEngine := NewBeaconCommitteeEngineV2(
		env.BeaconHeight,
		env.BeaconHash,
		committeeStateV2,
	)
	return committeeEngine
}
