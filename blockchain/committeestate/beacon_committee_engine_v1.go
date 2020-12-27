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
		beaconCommitteeEngineBase: beaconCommitteeEngineBase{
			beaconHeight:     beaconHeight,
			beaconHash:       beaconHash,
			finalState:       beaconCommitteeStateV1,
			uncommittedState: NewBeaconCommitteeStateV1(),
		},
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
	engine.uncommittedState = cloneBeaconCommitteeStateFrom(engine.finalState)
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
		Logger.log.Info("[dcs] shardCandidatesStr:", shardCandidatesStr)
		remainShardCandidatesStr, assignedCandidates := assignShardCandidate(shardCandidatesStr, numberOfShardSubstitutes, env.RandomNumber, env.AssignOffset, env.ActiveShards)
		Logger.log.Info("[dcs] remainShardCandidatesStr:", remainShardCandidatesStr)
		Logger.log.Info("[dcs] assignedCandidates:", assignedCandidates)
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

func (engine *BeaconCommitteeEngineV1) GenerateAssignInstruction(rand int64, assignOffset int, activeShards int, beaconHeight uint64) []*instruction.AssignInstruction {
	candidates, _ := incognitokey.CommitteeKeyListToString(engine.finalState.CandidateShardWaitingForCurrentRandom())
	numberOfPendingValidator := make(map[byte]int)
	shardPendingValidator := engine.finalState.ShardSubstitute()
	for i := 0; i < activeShards; i++ {
		if pendingValidators, ok := shardPendingValidator[byte(i)]; ok {
			numberOfPendingValidator[byte(i)] = len(pendingValidators)
		} else {
			numberOfPendingValidator[byte(i)] = 0
		}
	}
	assignedCandidates := make(map[byte][]string)
	shuffledCandidate := shuffleShardCandidate(candidates, rand)
	for _, candidate := range shuffledCandidate {
		shardID := calculateCandidateShardID(candidate, rand, activeShards)
		if numberOfPendingValidator[shardID]+1 <= assignOffset {
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
	for _, v := range instructions {
		Logger.log.Info("[dcs] v.String():", v.ToString())
	}
	return instructions
}

// GenerateAllSwapShardInstructions do nothing
func (b *BeaconCommitteeEngineV1) GenerateAllSwapShardInstructions(env *BeaconCommitteeStateEnvironment) (
	[]*instruction.SwapShardInstruction, error) {
	return []*instruction.SwapShardInstruction{}, nil
}
