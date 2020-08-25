package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/pkg/errors"
	"reflect"
	"sort"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain/instructionsprocessor"

	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateV2 struct {
	blockHeight uint64
	blockHash   common.Hash

	beaconCommittee []incognitokey.CommitteePublicKey

	shardCommittee              map[byte][]incognitokey.CommitteePublicKey
	shardSubstitute             map[byte][]incognitokey.CommitteePublicKey
	shardCommonPoolNextEpoch    []incognitokey.CommitteePublicKey
	shardCommonPoolCurrentEpoch []incognitokey.CommitteePublicKey

	autoStake      map[string]bool                   // committee public key => reward receiver payment address
	rewardReceiver map[string]privacy.PaymentAddress // incognito public key => reward receiver payment address
	stakingTx      map[string]common.Hash            // committee public key => reward receiver payment address
	//TODO: @hung store number of round
	numberOfRound map[string]int // committee public key => number of round in epoch

	mu *sync.RWMutex
}

type BeaconCommitteeEngineV2 struct {
	beaconHeight                      uint64
	beaconHash                        common.Hash
	finalBeaconCommitteeStateV2       *BeaconCommitteeStateV2
	uncommittedBeaconCommitteeStateV2 *BeaconCommitteeStateV2
	insProcessor                      *instructionsprocessor.BInsProcessor
}

func (b BeaconCommitteeStateV2) clone(newB *BeaconCommitteeStateV2) {
	newB.reset()
	newB.beaconCommittee = b.beaconCommittee
	newB.shardCommonPoolCurrentEpoch = b.shardCommonPoolCurrentEpoch
	newB.shardCommonPoolNextEpoch = b.shardCommonPoolNextEpoch
	for k, v := range b.shardCommittee {
		newB.shardCommittee[k] = v
	}
	for k, v := range b.shardSubstitute {
		newB.shardSubstitute[k] = v
	}
	for k, v := range b.autoStake {
		newB.autoStake[k] = v
	}
	for k, v := range b.numberOfRound {
		newB.numberOfRound[k] = v
	}
	for k, v := range b.rewardReceiver {
		newB.rewardReceiver[k] = v
	}
	for k, v := range b.stakingTx {
		newB.stakingTx[k] = v
	}
}

func (b *BeaconCommitteeStateV2) reset() {
	b.beaconCommittee = []incognitokey.CommitteePublicKey{}
	b.shardCommonPoolCurrentEpoch = []incognitokey.CommitteePublicKey{}
	b.shardCommonPoolNextEpoch = []incognitokey.CommitteePublicKey{}
	b.shardCommittee = make(map[byte][]incognitokey.CommitteePublicKey)
	b.shardSubstitute = make(map[byte][]incognitokey.CommitteePublicKey)
	b.autoStake = make(map[string]bool)
	b.numberOfRound = make(map[string]int)
	b.rewardReceiver = make(map[string]privacy.PaymentAddress)
	b.stakingTx = make(map[string]common.Hash)
}

//GetBeaconHeight :
func (engine BeaconCommitteeEngineV2) GetBeaconHeight() uint64 {
	return engine.beaconHeight
}

//GetBeaconHash :
func (engine BeaconCommitteeEngineV2) GetBeaconHash() common.Hash {
	return engine.beaconHash
}

//GetBeaconCommittee :
func (engine BeaconCommitteeEngineV2) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.beaconCommittee
}

//GetBeaconSubstitute :
func (engine BeaconCommitteeEngineV2) GetBeaconSubstitute() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

//GetCandidateShardWaitingForCurrentRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardCommonPoolCurrentEpoch
}

//GetCandidateBeaconWaitingForCurrentRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

//GetCandidateShardWaitingForNextRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardCommonPoolNextEpoch
}

//GetCandidateBeaconWaitingForNextRandom :
func (engine BeaconCommitteeEngineV2) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

//GetOneShardCommittee :
func (engine BeaconCommitteeEngineV2) GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardCommittee[shardID]
}

//GetShardCommittee :
func (engine BeaconCommitteeEngineV2) GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.finalBeaconCommitteeStateV2.shardCommittee {
		shardCommittee[k] = v
	}
	return shardCommittee
}

//GetOneShardSubstitute :
func (engine BeaconCommitteeEngineV2) GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.finalBeaconCommitteeStateV2.shardSubstitute[shardID]
}

//GetShardSubstitute :
func (engine BeaconCommitteeEngineV2) GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	shardSubstitute := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range engine.finalBeaconCommitteeStateV2.shardSubstitute {
		shardSubstitute[k] = v
	}
	return shardSubstitute
}

//GetAutoStaking :
func (engine BeaconCommitteeEngineV2) GetAutoStaking() map[string]bool {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	autoStake := make(map[string]bool)
	for k, v := range engine.finalBeaconCommitteeStateV2.autoStake {
		autoStake[k] = v
	}
	return autoStake
}

func (engine BeaconCommitteeEngineV2) GetRewardReceiver() map[string]privacy.PaymentAddress {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	rewardReceiver := make(map[string]privacy.PaymentAddress)
	for k, v := range engine.finalBeaconCommitteeStateV2.rewardReceiver {
		rewardReceiver[k] = v
	}
	return rewardReceiver
}

func (engine BeaconCommitteeEngineV2) GetStakingTx() map[string]common.Hash {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	stakingTx := make(map[string]common.Hash)
	for k, v := range engine.finalBeaconCommitteeStateV2.stakingTx {
		stakingTx[k] = v
	}
	return stakingTx
}

func (engine *BeaconCommitteeEngineV2) GetAllCandidateSubstituteCommittee() []string {
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	defer engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	return engine.finalBeaconCommitteeStateV2.getAllCandidateSubstituteCommittee()
}

func (engine *BeaconCommitteeEngineV2) Commit(hashes *BeaconCommitteeStateHash) error {
	if reflect.DeepEqual(engine.uncommittedBeaconCommitteeStateV2, NewBeaconCommitteeStateV1()) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("%+v", engine.uncommittedBeaconCommitteeStateV2))
	}
	engine.uncommittedBeaconCommitteeStateV2.mu.Lock()
	defer engine.uncommittedBeaconCommitteeStateV2.mu.Unlock()
	engine.finalBeaconCommitteeStateV2.mu.Lock()
	defer engine.finalBeaconCommitteeStateV2.mu.Unlock()
	comparedHashes, err := engine.generateUncommittedCommitteeHashes()
	if err != nil {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, err)
	}
	if !comparedHashes.BeaconCommitteeAndValidatorHash.IsEqual(&hashes.BeaconCommitteeAndValidatorHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("Uncommitted BeaconCommitteeAndValidatorHash want value %+v but have %+v", comparedHashes.BeaconCommitteeAndValidatorHash, hashes.BeaconCommitteeAndValidatorHash))
	}
	if !comparedHashes.BeaconCandidateHash.IsEqual(&hashes.BeaconCandidateHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("Uncommitted BeaconCandidateHash want value %+v but have %+v", comparedHashes.BeaconCandidateHash, hashes.BeaconCandidateHash))
	}
	if !comparedHashes.ShardCandidateHash.IsEqual(&hashes.ShardCandidateHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("Uncommitted ShardCandidateHash want value %+v but have %+v", comparedHashes.ShardCandidateHash, hashes.ShardCandidateHash))
	}
	if !comparedHashes.ShardCommitteeAndValidatorHash.IsEqual(&hashes.ShardCommitteeAndValidatorHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("Uncommitted ShardCommitteeAndValidatorHash want value %+v but have %+v", comparedHashes.ShardCommitteeAndValidatorHash, hashes.ShardCommitteeAndValidatorHash))
	}
	if !comparedHashes.AutoStakeHash.IsEqual(&hashes.AutoStakeHash) {
		return NewCommitteeStateError(ErrCommitBeaconCommitteeState, fmt.Errorf("Uncommitted AutoStakingHash want value %+v but have %+v", comparedHashes.AutoStakeHash, hashes.AutoStakeHash))
	}
	engine.uncommittedBeaconCommitteeStateV2.clone(engine.finalBeaconCommitteeStateV2)
	engine.uncommittedBeaconCommitteeStateV2.reset()
	return nil
}

func (engine *BeaconCommitteeEngineV2) AbortUncommittedBeaconState() {
	engine.uncommittedBeaconCommitteeStateV2.mu.Lock()
	defer engine.uncommittedBeaconCommitteeStateV2.mu.Unlock()
	engine.uncommittedBeaconCommitteeStateV2.reset()
}

func (engine *BeaconCommitteeEngineV2) InitCommitteeState(env *BeaconCommitteeStateEnvironment) {
	engine.finalBeaconCommitteeStateV2.mu.Lock()
	defer engine.finalBeaconCommitteeStateV2.mu.Unlock()
	b := engine.finalBeaconCommitteeStateV2
	newBeaconCandidates := []incognitokey.CommitteePublicKey{}
	newShardCandidates := []incognitokey.CommitteePublicKey{}
	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		if inst[0] == instruction.STAKE_ACTION {
			stakeInstruction := instruction.ImportInitStakeInstructionFromString(inst)
			newBeaconCandidates := []incognitokey.CommitteePublicKey{}
			newShardCandidates := []incognitokey.CommitteePublicKey{}
			for index, candidate := range stakeInstruction.PublicKeyStructs {
				b.rewardReceiver[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
				b.autoStake[stakeInstruction.PublicKeys[index]] = stakeInstruction.AutoStakingFlag[index]
				b.numberOfRound[stakeInstruction.PublicKeys[index]] = 1
				b.stakingTx[stakeInstruction.PublicKeys[index]] = stakeInstruction.TxStakeHashes[index]
			}
			if stakeInstruction.Chain == instruction.BEACON_INST {
				newBeaconCandidates = append(newBeaconCandidates, stakeInstruction.PublicKeyStructs...)
			} else {
				newShardCandidates = append(newShardCandidates, stakeInstruction.PublicKeyStructs...)
			}
			_ = statedb.StoreStakerInfo(
				env.ConsensusStateDB,
				stakeInstruction.PublicKeyStructs,
				b.rewardReceiver,
				b.autoStake,
				b.stakingTx,
			)
		}
	}
	b.beaconCommittee = append(b.beaconCommittee, newBeaconCandidates...)
	for shardID := 0; shardID < env.ActiveShards; shardID++ {
		b.shardCommittee[byte(shardID)] = append(b.shardCommittee[byte(shardID)], newShardCandidates[shardID*env.MinShardCommitteeSize:(shardID+1)*env.MinShardCommitteeSize]...)
	}
}

// New flow
// Store information from instructions into temp stateDB in env
// When all thing done and no problems, in commit function, we read data in statedb and update
// BeaconCommitteeStateV2
func (engine *BeaconCommitteeEngineV2) UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (*BeaconCommitteeStateHash, *CommitteeChange, error) {
	engine.uncommittedBeaconCommitteeStateV2.mu.Lock()
	defer engine.uncommittedBeaconCommitteeStateV2.mu.Unlock()
	engine.finalBeaconCommitteeStateV2.mu.RLock()
	engine.finalBeaconCommitteeStateV2.clone(engine.uncommittedBeaconCommitteeStateV2)
	env.allCandidateSubstituteCommittee = engine.finalBeaconCommitteeStateV2.getAllCandidateSubstituteCommittee()
	engine.finalBeaconCommitteeStateV2.mu.RUnlock()
	newB := engine.uncommittedBeaconCommitteeStateV2
	committeeChange := NewCommitteeChange()
	isAssignInstruction := false
	for _, inst := range env.BeaconInstructions {
		if len(inst) == 0 {
			continue
		}
		switch inst[0] {
		case instruction.STAKE_ACTION:
			stakeInstruction, err := instruction.ValidateAndImportStakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stake instruction %+v, error %+v", inst, err)
				continue
			}
			committeeChange, err = newB.processStakeInstruction(stakeInstruction, committeeChange, env)
			if err != nil {
				Logger.log.Errorf("SKIP stake instruction %+v, error %+v", inst, err)
				continue
			}
		case instruction.ASSIGN_ACTION:
			assignInstruction, err := instruction.ValidateAndImportAssignInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
			}
			committeeChange, err = newB.processAssignInstruction(assignInstruction, env, committeeChange)
			if err != nil {
				return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			isAssignInstruction = true
		case instruction.SWAP_ACTION:
			swapInstruction, err := instruction.ValidateAndImportSwapInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP swap instruction %+v, error %+v", inst, err)
				continue
			}
			committeeChange, err = newB.processSwapInstruction(swapInstruction, env, committeeChange)
			if err != nil {
				return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
		case instruction.STOP_AUTO_STAKE_ACTION:
			stopAutoStakeInstruction, err := instruction.ValidateAndImportStopAutoStakeInstructionFromString(inst)
			if err != nil {
				Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
			}
			committeeChange = newB.processStopAutoStakeInstruction(stopAutoStakeInstruction, env, committeeChange)
		}
	}
	// snapshot shard common pool in beacon random time
	if env.IsBeaconRandomTime {
		committeeChange = newB.snapshotShardCommonPool(committeeChange)
	}
	if isAssignInstruction {
		committeeChange.CurrentEpochShardCandidateRemoved = newB.shardCommonPoolCurrentEpoch
		newB.shardCommonPoolCurrentEpoch = []incognitokey.CommitteePublicKey{}
	}
	err := newB.processAutoStakingChange(committeeChange, env)
	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}
	hashes, err := engine.generateUncommittedCommitteeHashes()
	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}
	return hashes, committeeChange, nil
	// panic("implement me")
}

func (b *BeaconCommitteeEngineV2) GenerateShardSwapInstruction(env *BeaconCommitteeStateEnvironment) ([][]string, error) {
	return [][]string{}, nil
}

func (engine *BeaconCommitteeEngineV2) GenerateAssignInstruction(rand int64, assignOffset int, activeShards int) ([][]string, []string, map[byte][]string) {
	remainCandidates, assignedCandidates := engine.generateAssignInstruction(rand, activeShards)
	var keys []int
	instructions := [][]string{}
	for k := range assignedCandidates {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, key := range keys {
		shardID := byte(key)
		candidates := assignedCandidates[shardID]
		Logger.log.Infof("Assign Candidate at Shard %+v: %+v", shardID, candidates)
		shardAssignInstruction := instruction.NewAssignInstructionWithValue(int(shardID), candidates)
		instructions = append(instructions, shardAssignInstruction.ToString())
	}
	return instructions, remainCandidates, assignedCandidates
}

func (engine *BeaconCommitteeEngineV2) generateAssignInstruction(rand int64, activeShards int) ([]string, map[byte][]string) {
	candidates, _ := incognitokey.CommitteeKeyListToString(engine.finalBeaconCommitteeStateV2.shardCommonPoolCurrentEpoch)
	numberOfValidator := make([]int, activeShards)
	for i := 0; i < activeShards; i++ {
		numberOfValidator[byte(i)] += len(engine.finalBeaconCommitteeStateV2.shardSubstitute[byte(i)])
		numberOfValidator[byte(i)] += len(engine.finalBeaconCommitteeStateV2.shardCommittee[byte(i)])
	}
	assignedCandidates := assignShardCandidateV2(candidates, numberOfValidator, rand)
	return []string{}, assignedCandidates
}

func (b *BeaconCommitteeStateV2) snapshotShardCommonPool(committeeChange *CommitteeChange) *CommitteeChange {
	b.shardCommonPoolCurrentEpoch = b.shardCommonPoolNextEpoch
	committeeChange.CurrentEpochShardCandidateAdded = b.shardCommonPoolNextEpoch
	committeeChange.NextEpochShardCandidateRemoved = b.shardCommonPoolNextEpoch
	b.shardCommonPoolNextEpoch = []incognitokey.CommitteePublicKey{}
	return committeeChange
}

func (b *BeaconCommitteeStateV2) processStakeInstruction(
	stakeInstruction *instruction.StakeInstruction,
	committeeChange *CommitteeChange,
	env *BeaconCommitteeStateEnvironment,
) (*CommitteeChange, error) {
	for index, candidate := range stakeInstruction.PublicKeyStructs {
		b.rewardReceiver[candidate.GetIncKeyBase58()] = stakeInstruction.RewardReceiverStructs[index]
		b.autoStake[stakeInstruction.PublicKeys[index]] = stakeInstruction.AutoStakingFlag[index]
		b.numberOfRound[stakeInstruction.PublicKeys[index]] = 0
		b.stakingTx[stakeInstruction.PublicKeys[index]] = stakeInstruction.TxStakeHashes[index]
	}
	committeeChange.NextEpochShardCandidateAdded = append(committeeChange.NextEpochShardCandidateAdded, stakeInstruction.PublicKeyStructs...)
	b.shardCommonPoolNextEpoch = append(b.shardCommonPoolNextEpoch, stakeInstruction.PublicKeyStructs...)
	err := statedb.StoreStakerInfo(
		env.ConsensusStateDB,
		stakeInstruction.PublicKeyStructs,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
	)
	if err != nil {
		return committeeChange, err
	}
	return committeeChange, nil
}

func (b *BeaconCommitteeStateV2) processStopAutoStakeInstruction(
	stopAutoStakeInstruction *instruction.StopAutoStakeInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) *CommitteeChange {
	for _, committeePublicKey := range stopAutoStakeInstruction.PublicKeys {
		if common.IndexOfStr(committeePublicKey, env.allCandidateSubstituteCommittee) == -1 {
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
	return committeeChange
}

func (b *BeaconCommitteeStateV2) processAssignInstruction(
	assignInstruction *instruction.AssignInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) (*CommitteeChange, error) {
	assignedCandidates := []incognitokey.CommitteePublicKey{}
	if env.IsFoundRandomNumber == false {
		return committeeChange, fmt.Errorf("Found Assign Instruction %+v but Found random number is %+v", assignInstruction, false)
	}
	shardID := byte(assignInstruction.ChainID)
	assignedCandidates = assignInstruction.ShardCandidatesStruct
	b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], assignedCandidates...)
	committeeChange.ShardSubstituteAdded[shardID] = assignedCandidates
	return committeeChange, nil
}

func (b *BeaconCommitteeStateV2) processSwapInstruction(
	swapInstruction *instruction.SwapInstruction,
	env *BeaconCommitteeStateEnvironment,
	committeeChange *CommitteeChange,
) (*CommitteeChange, error) {
	shardID := byte(swapInstruction.ChainID)
	// delete in public key out of sharding pending validator list
	if len(swapInstruction.InPublicKeys) > 0 {
		shardSubstituteStr, err := incognitokey.CommitteeKeyListToString(b.shardSubstitute[shardID])
		if err != nil {
			return committeeChange, err
		}
		tempShardSubstitute, err := removeValidatorV2(shardSubstituteStr, swapInstruction.InPublicKeys)
		if err != nil {
			return committeeChange, err
		}
		// update shard pending validator
		committeeChange.ShardSubstituteRemoved[shardID] = append(committeeChange.ShardSubstituteRemoved[shardID], swapInstruction.InPublicKeyStructs...)
		b.shardSubstitute[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(tempShardSubstitute)
		if err != nil {
			return committeeChange, err
		}
		// add new public key to committees
		committeeChange.ShardCommitteeAdded[shardID] = append(committeeChange.ShardCommitteeAdded[shardID], swapInstruction.InPublicKeyStructs...)
		b.shardCommittee[shardID] = append(b.shardCommittee[shardID], swapInstruction.InPublicKeyStructs...)
		for _, inPublicKey := range swapInstruction.InPublicKeys {
			b.numberOfRound[inPublicKey] = b.numberOfRound[inPublicKey] + 1
		}
	}
	// delete out public key out of current committees
	if len(swapInstruction.OutPublicKeys) > 0 {
		shardCommitteeStr, err := incognitokey.CommitteeKeyListToString(b.shardCommittee[shardID])
		if err != nil {
			return committeeChange, err
		}
		tempShardCommittees, err := removeValidatorV2(shardCommitteeStr, swapInstruction.OutPublicKeys)
		if err != nil {
			return committeeChange, err
		}
		// remove old public key in shard committee update shard committee
		committeeChange.ShardCommitteeRemoved[shardID] = append(committeeChange.ShardCommitteeRemoved[shardID], swapInstruction.OutPublicKeyStructs...)
		b.shardCommittee[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(tempShardCommittees)
		if err != nil {
			return committeeChange, err
		}
		// check number of round staying in shard pool
		backToSubstitutesIndex := []int{}
		swappedOutSubstitutesIndex := []int{}
		for index, outPublicKey := range swapInstruction.OutPublicKeys {
			if b.numberOfRound[outPublicKey] >= MAX_NUMBER_OF_ROUND {
				swappedOutSubstitutesIndex = append(swappedOutSubstitutesIndex, index)
			} else {
				backToSubstitutesIndex = append(backToSubstitutesIndex, index)
				b.numberOfRound[outPublicKey] += 1
			}
		}
		// Check auto stake in swappedOutSubstitutes list
		// if auto staking not found or flag auto stake is false then do not re-stake for this out public key
		// if auto staking flag is true then system will automatically add this out public key to current candidate list
		for _, index := range swappedOutSubstitutesIndex {
			// @NOTICE: these lines of code is for debug purpose
			stakerInfo, has, err := statedb.GetStakerInfo(env.ConsensusStateDB, swapInstruction.OutPublicKeys[index])
			if err != nil {
				panic(err)
			}
			if !has {
				panic(errors.Errorf("Can not found info of this public key %v", swapInstruction.OutPublicKeys[index]))
			}
			if stakerInfo.AutoStaking() {
				b.shardCommonPoolNextEpoch = append(b.shardCommonPoolNextEpoch, swapInstruction.OutPublicKeyStructs[index])
			} else {
				delete(b.rewardReceiver, swapInstruction.OutPublicKeyStructs[index].GetIncKeyBase58())
				delete(b.autoStake, swapInstruction.OutPublicKeys[index])
			}
		}
		for _, index := range backToSubstitutesIndex {
			b.shardSubstitute[shardID] = append(b.shardSubstitute[shardID], swapInstruction.OutPublicKeyStructs[index])
			committeeChange.ShardSubstituteAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], swapInstruction.OutPublicKeyStructs[index])
		}
	}
	return committeeChange, nil
}

func (engine *BeaconCommitteeEngineV2) generateUncommittedCommitteeHashes() (*BeaconCommitteeStateHash, error) {
	if reflect.DeepEqual(engine.uncommittedBeaconCommitteeStateV2, NewBeaconCommitteeStateV1()) {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}
	newB := engine.uncommittedBeaconCommitteeStateV2
	// beacon committee
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(newB.beaconCommittee)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	validatorArr := append([]string{}, beaconCommitteeStr...)

	tempBeaconCommitteeAndValidatorHash, err := common.GenerateHashFromStringArray(validatorArr)
	// Shard candidate root: shard current candidate + shard next candidate
	shardCurrentEpochCandidateStr, err := incognitokey.CommitteeKeyListToString(newB.shardCommonPoolCurrentEpoch)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	shardNextEpochCandidateStr, err := incognitokey.CommitteeKeyListToString(newB.shardCommonPoolNextEpoch)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	tempShardCandidateHash, err := common.GenerateHashFromStringArray(append(shardCurrentEpochCandidateStr, shardNextEpochCandidateStr...))
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}
	// Shard Validator root
	shardPendingValidator := make(map[byte][]string)
	for shardID, keys := range newB.shardSubstitute {
		keysStr, err := incognitokey.CommitteeKeyListToString(keys)
		if err != nil {
			return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
		}
		shardPendingValidator[shardID] = keysStr
	}
	shardCommittee := make(map[byte][]string)
	for shardID, keys := range newB.shardCommittee {
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
	tempAutoStakingHash, err := common.GenerateHashFromMapStringBool(newB.autoStake)
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

func (b *BeaconCommitteeStateV2) getAllCandidateSubstituteCommittee() []string {
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
	shardCandidates := append(b.shardCommonPoolNextEpoch, b.shardCommonPoolCurrentEpoch...)
	shardCandidatesStr, err := incognitokey.CommitteeKeyListToString(shardCandidates)
	if err != nil {
		panic(err)
	}
	res = append(res, shardCandidatesStr...)
	return res
}

func (b *BeaconCommitteeStateV2) processAutoStakingChange(committeeChange *CommitteeChange, env *BeaconCommitteeStateEnvironment) error {
	stopAutoStakingIncognitoKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeeChange.StopAutoStake)
	if err != nil {
		return err
	}
	err = statedb.StoreStakerInfo(
		env.ConsensusStateDB,
		stopAutoStakingIncognitoKey,
		b.rewardReceiver,
		b.autoStake,
		b.stakingTx,
	)
	return nil
}
