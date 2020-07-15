package committeestate

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
)

//ShardCommitteeStateHash
type ShardCommitteeStateHash struct {
	ShardCommitteeHash  common.Hash
	ShardSubstituteHash common.Hash
}

//ShardCommitteeStateV1
type ShardCommitteeStateV1 struct {
	shardCommittee        []incognitokey.CommitteePublicKey
	shardPendingValidator []incognitokey.CommitteePublicKey

	mu *sync.RWMutex
}

//ShardCommitteeEngine
type ShardCommitteeEngine struct {
	shardHeight                      uint64
	shardHash                        common.Hash
	shardID                          byte
	shardCommitteeStateV1            *ShardCommitteeStateV1
	uncommittedShardCommitteeStateV1 *ShardCommitteeStateV1
}

//NewShardCommitteeStateV1 is default constructor for ShardCommitteeStateV1 ...
//Output: pointer of ShardCommitteeStateV1 struct
func NewShardCommitteeStateV1() *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{
		mu: new(sync.RWMutex),
	}
}

//NewShardCommitteeStateV1WithValue is constructor for ShardCommitteeStateV1 with value
//Output: pointer of ShardCommitteeStateV1 struct with value
func NewShardCommitteeStateV1WithValue(shardCommittee, shardPendingValidator []incognitokey.CommitteePublicKey) *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{
		shardCommittee:        shardCommittee,
		shardPendingValidator: shardPendingValidator,
		mu:                    new(sync.RWMutex),
	}
}

//NewShardCommitteeEngine is default constructor for ShardCommitteeEngine
//Output: pointer of ShardCommitteeEngine
func NewShardCommitteeEngine(shardHeight uint64,
	shardHash common.Hash, shardID byte, shardCommitteeStateV1 *ShardCommitteeStateV1) *ShardCommitteeEngine {
	return &ShardCommitteeEngine{
		shardHeight:                      shardHeight,
		shardHash:                        shardHash,
		shardID:                          shardID,
		shardCommitteeStateV1:            shardCommitteeStateV1,
		uncommittedShardCommitteeStateV1: NewShardCommitteeStateV1(),
	}
}

//clone ShardCommitteeStateV1 to new instance
func (committeeState ShardCommitteeStateV1) clone(newCommitteeState *ShardCommitteeStateV1) {
	newCommitteeState.reset()

	newCommitteeState.shardCommittee = make([]incognitokey.CommitteePublicKey, len(committeeState.shardCommittee))
	for i, v := range committeeState.shardCommittee {
		newCommitteeState.shardCommittee[i] = v
	}

	newCommitteeState.shardPendingValidator = make([]incognitokey.CommitteePublicKey, len(committeeState.shardPendingValidator))
	for i, v := range committeeState.shardPendingValidator {
		newCommitteeState.shardPendingValidator[i] = v
	}
}

//reset : reset ShardCommitteeStateV1 to default value
func (committeeState *ShardCommitteeStateV1) reset() {
	committeeState.shardCommittee = make([]incognitokey.CommitteePublicKey, 0)
	committeeState.shardPendingValidator = make([]incognitokey.CommitteePublicKey, 0)
}

//ValidateCommitteeRootHashes validate committee root hashes for checking if it's valid
//	Input: list rootHashes need checking
//	Output: result(boolean) and error
//	Pre-conditions: NULL
//	Post-conditions: NULL
func (engine *ShardCommitteeEngine) ValidateCommitteeRootHashes(rootHashes []common.Hash) (bool, error) {
	panic("Not implemented yet")
}

//Commit commit committee state change in uncommittedShardCommitteeStateV1 struct
//	Pre-conditions: uncommittedShardCommitteeStateV1 has been inited
//	Input: Shard Committee hash
//	Output: error
//	Post-conditions: NULL
//	Flow:
//		- Generate hash from uncommiteed
//		- Check validations of input hash
//		- clone uncommitted to commit
//		- reset uncommitted
func (engine *ShardCommitteeEngine) Commit(hashes *ShardCommitteeStateHash) error {
	if reflect.DeepEqual(engine.uncommittedShardCommitteeStateV1, NewShardCommitteeStateV1()) {
		return NewCommitteeStateError(ErrCommitShardCommitteeState, fmt.Errorf("%+v", engine.uncommittedShardCommitteeStateV1))
	}
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()
	engine.shardCommitteeStateV1.mu.Lock()
	defer engine.shardCommitteeStateV1.mu.Unlock()
	comparedHashes, err := engine.generateUncommittedCommitteeHashes()
	if err != nil {
		return NewCommitteeStateError(ErrCommitShardCommitteeState, err)
	}

	if !comparedHashes.ShardCommitteeHash.IsEqual(&hashes.ShardCommitteeHash) {
		return NewCommitteeStateError(ErrCommitShardCommitteeState, fmt.Errorf("Uncommitted ShardCommitteeHash want value %+v but have %+v",
			comparedHashes.ShardCommitteeHash, hashes.ShardCommitteeHash))
	}

	if !comparedHashes.ShardSubstituteHash.IsEqual(&hashes.ShardSubstituteHash) {
		return NewCommitteeStateError(ErrCommitShardCommitteeState, fmt.Errorf("Uncommitted ShardSubstituteHash want value %+v but have %+v",
			comparedHashes.ShardSubstituteHash, hashes.ShardSubstituteHash))
	}

	engine.uncommittedShardCommitteeStateV1.clone(engine.shardCommitteeStateV1)
	engine.uncommittedShardCommitteeStateV1.reset()
	return nil
}

//AbortUncommittedShardState reset data in uncommittedShardCommitteeStateV1 struct
//	Pre-conditions: uncommittedShardCommitteeStateV1 has been inited
//	Input: NULL
//	Output: error
//	Post-conditions: NULL
//	Flow: Reset uncommitted shard state
func (engine *ShardCommitteeEngine) AbortUncommittedShardState() {
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()
	engine.uncommittedShardCommitteeStateV1.reset()
}

//UpdateCommitteeState update committeState from valid data before
//	Pre-conditions: Validate committee state
//	Input: env variables ShardCommitteeStateEnvironment
// 	Output: New ShardCommitteeEngineV1 and committee changes, error
// 	Post-conditions: NULL
// 		Flow:
//			- call process instructions from beacon
//			- check conditions for epoch timestamp
//				+ process shard block instructions for key
//				+ process shard block instructions normally
//			- hash for checking commit later
func (engine *ShardCommitteeEngine) UpdateCommitteeState(
	env ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash, *CommitteeChange, error) {
	var err error
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()
	engine.shardCommitteeStateV1.mu.RLock()
	engine.shardCommitteeStateV1.clone(engine.uncommittedShardCommitteeStateV1)
	engine.shardCommitteeStateV1.mu.RUnlock()
	newCommitteeState := engine.uncommittedShardCommitteeStateV1
	committeeChange := NewCommitteeChange()

	// fmt.Println("[committee-state] committeeChange before processInstructionFromBeacon :", committeeChange

	if len(committeeChange.CurrentEpochShardCandidateAdded) != 0 ||
		len(committeeChange.CurrentEpochShardCandidateRemoved) != 0 ||
		len(committeeChange.NextEpochBeaconCandidateAdded) != 0 ||
		len(committeeChange.NextEpochShardCandidateRemoved) != 0 {
		fmt.Println("[committee-state] committeeChange before processInstructionFromBeacon :", committeeChange)
	}

	err = newCommitteeState.processInstructionFromBeacon(env.RecentSubtitutesStr(),
		env.Instructions(), env.ShardID(), committeeChange)
	if err != nil {
		return nil, nil, err
	}

	if len(committeeChange.CurrentEpochShardCandidateAdded) != 0 ||
		len(committeeChange.CurrentEpochShardCandidateRemoved) != 0 ||
		len(committeeChange.NextEpochBeaconCandidateAdded) != 0 ||
		len(committeeChange.NextEpochShardCandidateRemoved) != 0 {
		fmt.Println("[committee-state] committeeChange before processInstructionFromBeacon :", committeeChange)
	}

	// fmt.Println("[committee-state] committeeChange before process shard block instruction :", committeeChange)

	if common.IndexOfUint64(env.BeaconHeight()/env.ChainParamEpoch(), env.EpochBreakPointSwapNewKey()) > -1 &&
		env.IsProcessShardBlockInstructionForKeyListV2() {
		err = newCommitteeState.processShardBlockInstructionForKeyListV2(env, committeeChange)
	} else {
		if common.IndexOfUint64(env.BeaconHeight()/env.ChainParamEpoch(), env.EpochBreakPointSwapNewKey()) > -1 {
			err = newCommitteeState.processShardBlockInstruction(env, committeeChange)
		}
	}

	if len(committeeChange.CurrentEpochShardCandidateAdded) != 0 ||
		len(committeeChange.CurrentEpochShardCandidateRemoved) != 0 ||
		len(committeeChange.NextEpochBeaconCandidateAdded) != 0 ||
		len(committeeChange.NextEpochShardCandidateRemoved) != 0 {
		fmt.Println("[committee-state] committeeChange before processInstructionFromBeacon :", committeeChange)
	}

	// fmt.Println("[committee-state] committeeChange after process shard block instruction :", committeeChange)

	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}

	hashes, err := engine.generateUncommittedCommitteeHashes()
	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}
	return hashes, committeeChange, nil
}

//InitCommitteeState init committee state at genesis block or anytime restore program
// Pre-conditions: right config from files or env variables
// Input: env variables ShardCommitteeStateEnvironment
// Output: NULL
// Post-conditions: NULL
// Flow:
//	- call function processInstructionFromBeacon for process instructions received from beacon
//	- call function processShardBlockInstruction for process shard block instructions
func (engine *ShardCommitteeEngine) InitCommitteeState(env ShardCommitteeStateEnvironment) {
	engine.shardCommitteeStateV1.mu.Lock()
	defer engine.shardCommitteeStateV1.mu.Unlock()

	committeeState := engine.shardCommitteeStateV1
	committeeChange := NewCommitteeChange()
	err := committeeState.processInstructionFromBeacon(env.RecentSubtitutesStr(),
		env.Instructions(), env.ShardID(), committeeChange)
	if err != nil {
		panic(err)
	}
	err = committeeState.processShardBlockInstruction(env, committeeChange)
	if err != nil {
		panic(err)
	}
	committess, err := incognitokey.CommitteeBase58KeyListToStruct(env.RecentCommitteesStr())
	if err != nil {
		panic(err)
	}
	engine.shardCommitteeStateV1.shardCommittee = append(engine.shardCommitteeStateV1.shardCommittee, committess...)
}

//GetShardCommittee get shard committees
//Input: ShardID
//Output: list array of incognito public keys
func (engine *ShardCommitteeEngine) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.shardCommitteeStateV1.shardCommittee
}

//GetShardPendingValidator get shard pending validators
//Input: ShardID
//Output: list array of incognito public keys
func (engine *ShardCommitteeEngine) GetShardPendingValidator(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.shardCommitteeStateV1.shardPendingValidator
}

//generateUncommittedCommitteeHashes generate hashes relate to uncommitted committees of struct ShardCommitteeEngine
//	Input: NULL
//	Pre-conditions: NULL
//	Output:
//	Post-conditions: NULL
//	Flow:
//		- append committees and subtitutes to struct and hash it
func (engine ShardCommitteeEngine) generateUncommittedCommitteeHashes() (*ShardCommitteeStateHash, error) {
	if reflect.DeepEqual(engine.uncommittedShardCommitteeStateV1, NewBeaconCommitteeStateV1()) {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}

	newCommitteeState := engine.uncommittedShardCommitteeStateV1

	committeesStr, err := incognitokey.CommitteeKeyListToString(newCommitteeState.shardCommittee)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	committeeHash, err := common.GenerateHashFromStringArray(committeesStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	pendingValidatorsStr, err := incognitokey.CommitteeKeyListToString(newCommitteeState.shardPendingValidator)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	pendingValidatorHash, err := common.GenerateHashFromStringArray(pendingValidatorsStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	return &ShardCommitteeStateHash{
		ShardCommitteeHash:  committeeHash,
		ShardSubstituteHash: pendingValidatorHash,
	}, nil
}

//processnstructionFromBeacon process instruction from beacon blocks
// Pre-conditions: NULL
//	Input: list instructions, recent subtitutes validators in string data structure, shardID, committee change
//	Output: error
// 	Post-conditions: NULL
// 	Flow:
//		- Get all subtitutes in shard
//  	- Loop over the list instructions:
//			+ Create Assign instruction struct from assign instruction string
//		- Update shard subtitute added in committee change struct
func (committeeState *ShardCommitteeStateV1) processInstructionFromBeacon(
	recentSubtitutesStr []string,
	listInstructions [][]string,
	shardID byte,
	committeeChange *CommitteeChange) error {
	var err error
	shardPendingValidator := []string{}
	newShardPendingValidator := []string{}
	shardPendingValidator = append(shardPendingValidator, recentSubtitutesStr...)
	for _, ins := range listInstructions {
		//TODO: convert ins to intruction.AssignInstruction
		if ins[0] == instruction.ASSIGN_ACTION && ins[2] == "shard" {
			assignInstruction, err := instruction.ImportAssignInstructionFromString(ins)
			if err != nil {
				return err
			}
			if assignInstruction.ChainID == int(shardID) {
				tempNewShardPendingValidator := assignInstruction.ShardCandidates
				shardPendingValidator = append(shardPendingValidator, tempNewShardPendingValidator...)
				newShardPendingValidator = append(newShardPendingValidator, tempNewShardPendingValidator...)
				committeeState.shardPendingValidator = append(committeeState.shardPendingValidator, assignInstruction.ShardCandidatesStruct...)
			}
		}
	}

	if len(newShardPendingValidator) != 0 {
		fmt.Println("[committee-state] newShardPendingValidator:", newShardPendingValidator)
	}

	if len(committeeChange.ShardSubstituteAdded[shardID]) != 0 {
		fmt.Println("[committee-state] committeeChange.ShardSubstituteAdded[shardID]:", committeeChange.ShardSubstituteAdded[shardID])
	}

	committeeChange.ShardSubstituteAdded[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(newShardPendingValidator)
	if err != nil {
		return err
	}
	return nil
}

//processShardBlockInstruction process shard block instruction for sending to beacon
//	Pre-conditions: NULL
//	Input: ShardCommitteeStateEnvironment, committee change
//	Output: error
//	Post-conditions: NULL
//	Flow:
//		- get list instructions from input environment
//		- loop over the list instructions
//			+ Check type of instructions and process itp
//			+ At this moment, there will be only swap action for this function
//		- After process all instructions, we will updatew commitee change variable
func (committeeState *ShardCommitteeStateV1) processShardBlockInstruction(
	env ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) error {
	var err error
	shardID := env.ShardID()
	shardPendingValidator, err := incognitokey.CommitteeKeyListToString(committeeState.shardPendingValidator)
	if err != nil {
		return err
	}
	shardCommittee, err := incognitokey.CommitteeKeyListToString(committeeState.shardCommittee)
	if err != nil {
		return err
	}
	shardSwappedCommittees := []string{}
	shardNewCommittees := []string{}
	if len(env.Instructions()) != 0 {
		Logger.log.Debugf("Shard Process/processShardBlockInstruction: Shard Instruction %+v", env.Instructions())
	}
	// Swap committee
	for _, inst := range env.Instructions() {
		//TODO: convert ins to intruction.SwapInstruction
		if inst[0] == instruction.SWAP_ACTION {
			// #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator
			shardPendingValidator, shardCommittee, shardSwappedCommittees, shardNewCommittees, err = SwapValidator(shardPendingValidator,
				shardCommittee, env.MaxShardCommitteeSize(), env.MinShardCommitteeSize(), env.Offset(),
				env.ProducersBlackList(), env.SwapOffset())
			if err != nil {
				Logger.log.Errorf("SHARD %+v | Blockchain Error %+v", err)
				return err
			}
			swapedCommittees := []string{}
			if len(inst[2]) != 0 && inst[2] != "" {
				swapedCommittees = strings.Split(inst[2], ",")
			}
			for _, v := range swapedCommittees {
				if txID, ok := env.StakingTx()[v]; ok {
					if checkReturnStakingTxExistence(txID, env.Txs()) {
						delete(env.StakingTx(), v)
					}
				}
			}
			if !reflect.DeepEqual(swapedCommittees, shardSwappedCommittees) {
				return fmt.Errorf("Expect swapped committees to be %+v but get %+v", swapedCommittees, shardSwappedCommittees)
			}
			newCommittees := []string{}
			if len(inst[1]) > 0 {
				newCommittees = strings.Split(inst[1], ",")
			}
			if !reflect.DeepEqual(newCommittees, shardNewCommittees) {
				return fmt.Errorf("Expect new committees to be %+v but get %+v", newCommittees, shardNewCommittees)
			}
			shardNewCommitteesStruct, err := incognitokey.CommitteeBase58KeyListToStruct(shardNewCommittees)
			if err != nil {
				return err
			}
			shardSwappedCommitteesStruct, err := incognitokey.CommitteeBase58KeyListToStruct(shardSwappedCommittees)
			if err != nil {
				return err
			}
			beforeFilterShardSubstituteAdded := committeeChange.ShardSubstituteAdded[shardID]
			filteredShardSubstituteAdded := []incognitokey.CommitteePublicKey{}
			filteredShardSubstituteRemoved := []incognitokey.CommitteePublicKey{}
			for _, currentShardSubstituteAdded := range beforeFilterShardSubstituteAdded {
				flag := false
				for _, newShardCommitteeAdded := range shardNewCommitteesStruct {
					if currentShardSubstituteAdded.IsEqual(newShardCommitteeAdded) {
						flag = true
						break
					}
				}
				if !flag {
					filteredShardSubstituteAdded = append(filteredShardSubstituteAdded, currentShardSubstituteAdded)
				}
			}
			for _, newShardCommitteeAdded := range shardNewCommitteesStruct {
				flag := false
				for _, currentShardSubstituteAdded := range beforeFilterShardSubstituteAdded {
					if currentShardSubstituteAdded.IsEqual(newShardCommitteeAdded) {
						flag = true
						break
					}
				}
				if !flag {
					filteredShardSubstituteRemoved = append(filteredShardSubstituteRemoved, newShardCommitteeAdded)
				}
			}
			committeeChange.ShardCommitteeAdded[shardID] = shardNewCommitteesStruct
			committeeChange.ShardCommitteeRemoved[shardID] = shardSwappedCommitteesStruct
			committeeChange.ShardSubstituteRemoved[shardID] = filteredShardSubstituteRemoved
			committeeChange.ShardSubstituteAdded[shardID] = filteredShardSubstituteAdded
			Logger.log.Infof("SHARD %+v | Swap: Out committee %+v", shardID, shardSwappedCommittees)
			Logger.log.Infof("SHARD %+v | Swap: In committee %+v", shardID, shardNewCommittees)
		}
	}
	committeeState.shardPendingValidator, err = incognitokey.CommitteeBase58KeyListToStruct(shardPendingValidator)
	if err != nil {
		return err
	}
	committeeState.shardCommittee, err = incognitokey.CommitteeBase58KeyListToStruct(shardCommittee)
	if err != nil {
		return err
	}
	return nil
}

//processShardBlockInstructionForKeyListV2 process shard block instructions for key list v2
//	Pre-conditions: NULL
//	Input: ShardCommitteeStateEnvironment, CommitteeChange
//	Output: error
//	Post-conditions: NULL
//	Flow:
//		- get list instructions from input environment
//		- loop over the list instructions
//			+ Check type of instructions and process it
//			+ At this moment, there will be only swap action for this function
func (committeeState *ShardCommitteeStateV1) processShardBlockInstructionForKeyListV2(
	env ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) error {
	shardID := env.ShardID()
	for _, ins := range env.Instructions() {
		if ins[0] == instruction.SWAP_ACTION {
			shardPendingValidatorStruct := committeeState.shardPendingValidator
			inPublicKeys := strings.Split(ins[1], ",")
			inPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(inPublicKeys)
			if err != nil {
				return err
			}
			outPublicKeys := strings.Split(ins[2], ",")
			outPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
			if err != nil {
				return err
			}
			inRewardReceiver := strings.Split(ins[6], ",")
			if len(inPublicKeys) != len(outPublicKeys) {
				return fmt.Errorf("length new committee %+v, length out committee %+v", len(inPublicKeys), len(outPublicKeys))
			}
			if len(inPublicKeys) != len(inRewardReceiver) {
				return fmt.Errorf("length new committee %+v, new reward receiver %+v", len(inPublicKeys), len(inRewardReceiver))
			}
			removedCommitteeSize := len(inPublicKeys)
			remainedShardCommittees := committeeState.shardCommittee[removedCommitteeSize:]
			tempShardSwappedCommittees := committeeState.shardCommittee[:env.MinShardCommitteeSize()]
			if !reflect.DeepEqual(outPublicKeyStructs, tempShardSwappedCommittees) {
				return fmt.Errorf("expect swapped committe %+v but got %+v", tempShardSwappedCommittees, outPublicKeyStructs)
			}
			shardCommitteesStruct := append(inPublicKeyStructs, remainedShardCommittees...)
			committeeState.shardPendingValidator = shardPendingValidatorStruct
			committeeState.shardCommittee = shardCommitteesStruct
			committeeChange.ShardCommitteeAdded[shardID] = inPublicKeyStructs
			committeeChange.ShardCommitteeRemoved[shardID] = outPublicKeyStructs
		}
	}
	return nil
}
