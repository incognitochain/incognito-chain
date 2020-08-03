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
func (engine *ShardCommitteeEngine) ValidateCommitteeRootHashes(rootHashes []common.Hash) (bool, error) {
	panic("Not implemented yet")
}

//Commit commit committee state change in uncommittedShardCommitteeStateV1 struct
//	- Generate hash from uncommiteed
//	- Check validations of input hash
//	- clone uncommitted to commit
//	- reset uncommitted
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
func (engine *ShardCommitteeEngine) AbortUncommittedShardState() {
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()
	engine.uncommittedShardCommitteeStateV1.reset()
}

//UpdateCommitteeState update committeState from valid data before
//	- call process instructions from beacon
//	- check conditions for epoch timestamp
//		+ process shard block instructions for key
//			+ process shard block instructions normally
//	- hash for checking commit later
//	- Only call once in new or insert block process
func (engine *ShardCommitteeEngine) UpdateCommitteeState(
	env ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash, *CommitteeChange, error) {
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()
	engine.shardCommitteeStateV1.mu.RLock()
	engine.shardCommitteeStateV1.clone(engine.uncommittedShardCommitteeStateV1)
	engine.shardCommitteeStateV1.mu.RUnlock()
	newCommitteeState := engine.uncommittedShardCommitteeStateV1

	committeeChange, err := newCommitteeState.processInstructionFromBeacon(env.BeaconInstructions(), env.ShardID(), NewCommitteeChange())

	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}

	if common.IndexOfUint64(env.BeaconHeight()/env.ChainParamEpoch(), env.EpochBreakPointSwapNewKey()) > -1 {
		committeeChange, err = newCommitteeState.processShardBlockInstructionForKeyListV2(env, committeeChange)
	} else {
		committeeChange, err = newCommitteeState.processShardBlockInstruction(env, committeeChange)
	}

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
//	- call function processInstructionFromBeacon for process instructions received from beacon
//	- call function processShardBlockInstruction for process shard block instructions
func (engine *ShardCommitteeEngine) InitCommitteeState(env ShardCommitteeStateEnvironment) {
	engine.shardCommitteeStateV1.mu.Lock()
	defer engine.shardCommitteeStateV1.mu.Unlock()

	committeeState := engine.shardCommitteeStateV1
	committeeChange := NewCommitteeChange()

	shardPendingValidator := []string{}
	newShardPendingValidator := []incognitokey.CommitteePublicKey{}

	shardsCommittees := []string{}

	for _, beaconInstruction := range env.BeaconInstructions() {
		if beaconInstruction[0] == instruction.STAKE_ACTION {
			shardsCommittees = strings.Split(beaconInstruction[1], ",")
		}
		if beaconInstruction[0] == instruction.ASSIGN_ACTION {
			assignInstruction, err := instruction.ValidateAndImportAssignInstructionFromString(beaconInstruction)
			if err == nil && assignInstruction.ChainID == int(env.ShardID()) {
				shardPendingValidator = append(shardPendingValidator, assignInstruction.ShardCandidates...)
				newShardPendingValidator = append(newShardPendingValidator, assignInstruction.ShardCandidatesStruct...)
				committeeState.shardPendingValidator = append(committeeState.shardPendingValidator, assignInstruction.ShardCandidatesStruct...)
			}
		}
	}

	newShardCandidateStructs := []incognitokey.CommitteePublicKey{}
	for _, candidate := range shardsCommittees {
		key := incognitokey.CommitteePublicKey{}
		err := key.FromBase58(candidate)
		if err != nil {
			panic(err)
		}
		newShardCandidateStructs = append(newShardCandidateStructs, key)
	}

	addedCommittees := []incognitokey.CommitteePublicKey{}
	addedCommittees = append(addedCommittees, newShardCandidateStructs[int(env.ShardID())*
		env.MinShardCommitteeSize():(int(env.ShardID())*env.MinShardCommitteeSize())+env.MinShardCommitteeSize()]...)

	_, err := committeeState.processShardBlockInstruction(env, committeeChange)
	if err != nil {
		panic(err)
	}

	engine.shardCommitteeStateV1.shardCommittee = append(engine.shardCommitteeStateV1.shardCommittee,
		addedCommittees...)
	engine.shardCommitteeStateV1.shardPendingValidator = append(engine.shardCommitteeStateV1.shardPendingValidator,
		newShardPendingValidator...)
}

//GetShardCommittee get shard committees
func (engine *ShardCommitteeEngine) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.shardCommitteeStateV1.shardCommittee
}

//GetShardPendingValidator get shard pending validators
func (engine *ShardCommitteeEngine) GetShardPendingValidator(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.shardCommitteeStateV1.shardPendingValidator
}

//generateUncommittedCommitteeHashes generate hashes relate to uncommitted committees of struct ShardCommitteeEngine
//	append committees and subtitutes to struct and hash it
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
//	- Get all subtitutes in shard
//  - Loop over the list instructions:
//		+ Create Assign instruction struct from assign instruction string
//	- Update shard subtitute added in committee change struct
//	- Only call once in new or insert block process
func (committeeState *ShardCommitteeStateV1) processInstructionFromBeacon(
	listInstructions [][]string,
	shardID byte,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {

	addedSubstituteValidator := []incognitokey.CommitteePublicKey{}

	for _, inst := range listInstructions {
		assignInstruction, err := instruction.ValidateAndImportAssignInstructionFromString(inst)
		if err == nil && assignInstruction.ChainID == int(shardID) {
			addedSubstituteValidator = append(addedSubstituteValidator, assignInstruction.ShardCandidatesStruct...)
			committeeState.shardPendingValidator = append(committeeState.shardPendingValidator, assignInstruction.ShardCandidatesStruct...)
		}
	}

	committeeChange.ShardSubstituteAdded[shardID] = addedSubstituteValidator

	return committeeChange, nil
}

//processShardBlockInstruction process shard block instruction for sending to beacon
//	- get list instructions from input environment
//	- loop over the list instructions
//		+ Check type of instructions and process itp
//		+ At this moment, there will be only swap action for this function
//	- After process all instructions, we will updatew commitee change variable
//	- Only call once in new or insert block process
func (committeeState *ShardCommitteeStateV1) processShardBlockInstruction(
	env ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {
	var err error
	newCommitteeChange := committeeChange
	shardID := env.ShardID()
	shardPendingValidator, err := incognitokey.CommitteeKeyListToString(committeeState.shardPendingValidator)
	if err != nil {
		return nil, err
	}
	shardCommittee, err := incognitokey.CommitteeKeyListToString(committeeState.shardCommittee)
	if err != nil {
		return nil, err
	}
	fixedProducerShardValidators := shardCommittee[:env.NumberOfFixedBlockValidators()]
	shardCommittee = shardCommittee[env.NumberOfFixedBlockValidators():]
	shardSwappedCommittees := []string{}
	shardNewCommittees := []string{}
	if len(env.ShardInstructions()) != 0 {
		Logger.log.Debugf("Shard Process/processShardBlockInstruction: Shard Instruction %+v", env.ShardInstructions())
	}
	// Swap committee
	for _, ins := range env.ShardInstructions() {
		swapInstruction, err := instruction.ValidateAndImportSwapInstructionFromString(ins)
		if err == nil {
			// #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator
			maxShardCommitteeSize := env.MaxShardCommitteeSize() - env.NumberOfFixedBlockValidators()
			var minShardCommitteeSize int
			if env.MinShardCommitteeSize()-env.NumberOfFixedBlockValidators() < 0 {
				minShardCommitteeSize = 0
			} else {
				minShardCommitteeSize = env.MinShardCommitteeSize() - env.NumberOfFixedBlockValidators()
			}
			shardPendingValidator, shardCommittee, shardSwappedCommittees, shardNewCommittees, err =
				SwapValidator(shardPendingValidator,
					shardCommittee, maxShardCommitteeSize,
					minShardCommitteeSize, env.Offset(),
					env.ProducersBlackList(), env.SwapOffset())

			if err != nil {
				Logger.log.Errorf("SHARD %+v | Blockchain Error %+v", err)
				return nil, err
			}

			if !reflect.DeepEqual(swapInstruction.OutPublicKeys, shardSwappedCommittees) {
				return nil, fmt.Errorf("Expect swapped committees to be %+v but get %+v",
					swapInstruction.OutPublicKeys, shardSwappedCommittees)
			}

			if !reflect.DeepEqual(swapInstruction.InPublicKeys, shardNewCommittees) {
				return nil, fmt.Errorf("Expect new committees to be %+v but get %+v",
					swapInstruction.InPublicKeys, shardNewCommittees)
			}

			shardNewCommitteesStruct, err := incognitokey.CommitteeBase58KeyListToStruct(shardNewCommittees)
			if err != nil {
				return nil, err
			}

			shardSwappedCommitteesStruct, err := incognitokey.CommitteeBase58KeyListToStruct(shardSwappedCommittees)
			if err != nil {
				return nil, err
			}

			beforeFilterShardSubstituteAdded := newCommitteeChange.ShardSubstituteAdded[shardID]
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

			newCommitteeChange.ShardCommitteeAdded[shardID] = shardNewCommitteesStruct
			newCommitteeChange.ShardCommitteeRemoved[shardID] = shardSwappedCommitteesStruct
			newCommitteeChange.ShardSubstituteRemoved[shardID] = filteredShardSubstituteRemoved
			newCommitteeChange.ShardSubstituteAdded[shardID] = filteredShardSubstituteAdded
			Logger.log.Infof("SHARD %+v | Swap: Out committee %+v", shardID, shardSwappedCommittees)
			Logger.log.Infof("SHARD %+v | Swap: In committee %+v", shardID, shardNewCommittees)
		}
	}

	committeeState.shardPendingValidator, err = incognitokey.CommitteeBase58KeyListToStruct(shardPendingValidator)
	if err != nil {
		return nil, err
	}

	committeeState.shardCommittee, err = incognitokey.CommitteeBase58KeyListToStruct(append(fixedProducerShardValidators, shardCommittee...))
	if err != nil {
		return nil, err
	}

	return newCommitteeChange, nil
}

//processShardBlockInstructionForKeyListV2 process shard block instructions for key list v2
//	- get list instructions from input environment
//	- loop over the list instructions
//		+ Check type of instructions and process it
//		+ At this moment, there will be only swap action for this function
func (committeeState *ShardCommitteeStateV1) processShardBlockInstructionForKeyListV2(
	env ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {
	shardID := env.ShardID()
	newCommitteeChange := committeeChange
	for _, inst := range env.ShardInstructions() {
		if inst[0] == instruction.SWAP_ACTION {
			swapInstruction, err := instruction.ValidateAndImportSwapInstructionFromString(inst)
			if err != nil {
				return nil, err
			}
			removedCommitteeSize := len(swapInstruction.InPublicKeys)
			remainedShardCommittees := committeeState.shardCommittee[removedCommitteeSize:]
			tempShardSwappedCommittees := committeeState.shardCommittee[:env.MinShardCommitteeSize()]
			if !reflect.DeepEqual(swapInstruction.OutPublicKeyStructs, tempShardSwappedCommittees) {
				return nil, fmt.Errorf("expect swapped committe %+v but got %+v", tempShardSwappedCommittees, swapInstruction.OutPublicKeyStructs)
			}
			shardCommitteesStruct := append(swapInstruction.InPublicKeyStructs, remainedShardCommittees...)
			committeeState.shardCommittee = shardCommitteesStruct
			newCommitteeChange.ShardCommitteeAdded[shardID] = swapInstruction.InPublicKeyStructs
			newCommitteeChange.ShardCommitteeRemoved[shardID] = swapInstruction.OutPublicKeyStructs
		}
	}
	return newCommitteeChange, nil
}

//ProcessInstructionFromBeacon : process instrucction from beacon
func (engine *ShardCommitteeEngine) ProcessInstructionFromBeacon(
	env ShardCommitteeStateEnvironment) (*CommitteeChange, error) {
	newCommitteeState := &ShardCommitteeStateV1{}
	engine.shardCommitteeStateV1.mu.RLock()
	engine.shardCommitteeStateV1.clone(newCommitteeState)
	engine.shardCommitteeStateV1.mu.RUnlock()

	committeeChange, err := newCommitteeState.processInstructionFromBeacon(
		env.BeaconInstructions(),
		env.ShardID(), NewCommitteeChange())

	if err != nil {
		return nil, err
	}

	return committeeChange, nil
}

//ProcessInstructionFromShard :
func (engine *ShardCommitteeEngine) ProcessInstructionFromShard(env ShardCommitteeStateEnvironment) (*CommitteeChange, error) {
	return nil, nil
}
