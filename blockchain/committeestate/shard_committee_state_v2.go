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
type ShardCommitteeStateHashV2 struct {
	ShardCommitteeHash  common.Hash
	ShardSubstituteHash common.Hash
}

//ShardCommitteeStateV2
type ShardCommitteeStateV2 struct {
	shardCommittee  []incognitokey.CommitteePublicKey
	shardSubstitute []incognitokey.CommitteePublicKey

	mu *sync.RWMutex
}

//ShardCommitteeEngineV2
type ShardCommitteeEngineV2 struct {
	shardHeight                      uint64
	shardHash                        common.Hash
	shardID                          byte
	shardCommitteeStateV2            *ShardCommitteeStateV2
	uncommittedShardCommitteeStateV2 *ShardCommitteeStateV2
}

//NewShardCommitteeStateV2 is default constructor for ShardCommitteeStateV2 ...
//Output: pointer of ShardCommitteeStateV2 struct
func NewShardCommitteeStateV2() *ShardCommitteeStateV2 {
	return &ShardCommitteeStateV2{
		mu: new(sync.RWMutex),
	}
}

//NewShardCommitteeStateV2WithValue is constructor for ShardCommitteeStateV2 with value
//Output: pointer of ShardCommitteeStateV2 struct with value
func NewShardCommitteeStateV2WithValue(shardCommittee, shardPendingValidator []incognitokey.CommitteePublicKey) *ShardCommitteeStateV2 {
	return &ShardCommitteeStateV2{
		shardCommittee:  shardCommittee,
		shardSubstitute: shardPendingValidator,
		mu:              new(sync.RWMutex),
	}
}

//NewShardCommitteeEngine is default constructor for ShardCommitteeEngineV2
//Output: pointer of ShardCommitteeEngineV2
func NewShardCommitteeEngineV2(shardHeight uint64,
	shardHash common.Hash, shardID byte, shardCommitteeStateV2 *ShardCommitteeStateV2) *ShardCommitteeEngineV2 {
	return &ShardCommitteeEngineV2{
		shardHeight:                      shardHeight,
		shardHash:                        shardHash,
		shardID:                          shardID,
		shardCommitteeStateV2:            shardCommitteeStateV2,
		uncommittedShardCommitteeStateV2: NewShardCommitteeStateV2(),
	}
}

//clone ShardCommitteeStateV2 to new instance
func (committeeState ShardCommitteeStateV2) clone(newCommitteeState *ShardCommitteeStateV2) {
	newCommitteeState.reset()

	newCommitteeState.shardCommittee = make([]incognitokey.CommitteePublicKey, len(committeeState.shardCommittee))
	for i, v := range committeeState.shardCommittee {
		newCommitteeState.shardCommittee[i] = v
	}

	newCommitteeState.shardSubstitute = make([]incognitokey.CommitteePublicKey, len(committeeState.shardSubstitute))
	for i, v := range committeeState.shardSubstitute {
		newCommitteeState.shardSubstitute[i] = v
	}
}

//reset : reset ShardCommitteeStateV2 to default value
func (committeeState *ShardCommitteeStateV2) reset() {
	committeeState.shardCommittee = make([]incognitokey.CommitteePublicKey, 0)
	committeeState.shardSubstitute = make([]incognitokey.CommitteePublicKey, 0)
}

//ValidateCommitteeRootHashes validate committee root hashes for checking if it's valid
func (engine *ShardCommitteeEngineV2) ValidateCommitteeRootHashes(rootHashes []common.Hash) (bool, error) {
	panic("Not implemented yet")
}

//Commit commit committee state change in uncommittedShardCommitteeStateV2 struct
//	- Generate hash from uncommiteed
//	- Check validations of input hash
//	- clone uncommitted to commit
//	- reset uncommitted
func (engine *ShardCommitteeEngineV2) Commit(hashes *ShardCommitteeStateHash) error {
	if reflect.DeepEqual(engine.uncommittedShardCommitteeStateV2, NewShardCommitteeStateV2()) {
		return NewCommitteeStateError(ErrCommitShardCommitteeState, fmt.Errorf("%+v", engine.uncommittedShardCommitteeStateV2))
	}
	engine.uncommittedShardCommitteeStateV2.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV2.mu.Unlock()
	engine.shardCommitteeStateV2.mu.Lock()
	defer engine.shardCommitteeStateV2.mu.Unlock()
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

	engine.uncommittedShardCommitteeStateV2.clone(engine.shardCommitteeStateV2)
	engine.uncommittedShardCommitteeStateV2.reset()
	return nil
}

//AbortUncommittedShardState reset data in uncommittedShardCommitteeStateV2 struct
func (engine *ShardCommitteeEngineV2) AbortUncommittedShardState() {
	engine.uncommittedShardCommitteeStateV2.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV2.mu.Unlock()
	engine.uncommittedShardCommitteeStateV2.reset()
}

//UpdateCommitteeState update committeState from valid data before
//	- call process instructions from beacon
//	- check conditions for epoch timestamp
//		+ process shard block instructions for key
//			+ process shard block instructions normally
//	- hash for checking commit later
//	- Only call once in new or insert block process
func (engine *ShardCommitteeEngineV2) UpdateCommitteeState(
	env ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash, *CommitteeChange, error) {
	engine.uncommittedShardCommitteeStateV2.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV2.mu.Unlock()
	engine.shardCommitteeStateV2.mu.RLock()
	engine.shardCommitteeStateV2.clone(engine.uncommittedShardCommitteeStateV2)
	engine.shardCommitteeStateV2.mu.RUnlock()
	newCommitteeState := engine.uncommittedShardCommitteeStateV2

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
func (engine *ShardCommitteeEngineV2) InitCommitteeState(env ShardCommitteeStateEnvironment) {
	engine.shardCommitteeStateV2.mu.Lock()
	defer engine.shardCommitteeStateV2.mu.Unlock()

	committeeState := engine.shardCommitteeStateV2
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
				committeeState.shardSubstitute = append(committeeState.shardSubstitute, assignInstruction.ShardCandidatesStruct...)
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

	engine.shardCommitteeStateV2.shardCommittee = append(engine.shardCommitteeStateV2.shardCommittee,
		addedCommittees...)
	engine.shardCommitteeStateV2.shardSubstitute = append(engine.shardCommitteeStateV2.shardSubstitute,
		newShardPendingValidator...)
}

//GetShardCommittee get shard committees
func (engine *ShardCommitteeEngineV2) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.shardCommitteeStateV2.shardCommittee
}

//GetShardPendingValidator get shard pending validators
func (engine *ShardCommitteeEngineV2) GetShardPendingValidator(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.shardCommitteeStateV2.shardSubstitute
}

//generateUncommittedCommitteeHashes generate hashes relate to uncommitted committees of struct ShardCommitteeEngineV2
//	append committees and subtitutes to struct and hash it
func (engine ShardCommitteeEngineV2) generateUncommittedCommitteeHashes() (*ShardCommitteeStateHash, error) {
	if reflect.DeepEqual(engine.uncommittedShardCommitteeStateV2, NewBeaconCommitteeStateV1()) {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}

	newCommitteeState := engine.uncommittedShardCommitteeStateV2

	committeesStr, err := incognitokey.CommitteeKeyListToString(newCommitteeState.shardCommittee)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	committeeHash, err := common.GenerateHashFromStringArray(committeesStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	pendingValidatorsStr, err := incognitokey.CommitteeKeyListToString(newCommitteeState.shardSubstitute)
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
func (committeeState *ShardCommitteeStateV2) processInstructionFromBeacon(
	listInstructions [][]string,
	shardID byte,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {

	addedSubstituteValidator := []incognitokey.CommitteePublicKey{}

	for _, inst := range listInstructions {
		assignInstruction, err := instruction.ValidateAndImportAssignInstructionFromString(inst)
		if err == nil && assignInstruction.ChainID == int(shardID) {
			addedSubstituteValidator = append(addedSubstituteValidator, assignInstruction.ShardCandidatesStruct...)
			committeeState.shardSubstitute = append(committeeState.shardSubstitute, assignInstruction.ShardCandidatesStruct...)
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
func (committeeState *ShardCommitteeStateV2) processShardBlockInstruction(
	env ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {
	var err error
	newCommitteeChange := committeeChange
	shardID := env.ShardID()
	shardPendingValidator, err := incognitokey.CommitteeKeyListToString(committeeState.shardSubstitute)
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
					env.SwapOffset())

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

	committeeState.shardSubstitute, err = incognitokey.CommitteeBase58KeyListToStruct(shardPendingValidator)
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
func (committeeState *ShardCommitteeStateV2) processShardBlockInstructionForKeyListV2(
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
func (engine *ShardCommitteeEngineV2) ProcessInstructionFromBeacon(
	env ShardCommitteeStateEnvironment) (*CommitteeChange, error) {
	newCommitteeState := &ShardCommitteeStateV2{}
	engine.shardCommitteeStateV2.mu.RLock()
	engine.shardCommitteeStateV2.clone(newCommitteeState)
	engine.shardCommitteeStateV2.mu.RUnlock()

	committeeChange, err := newCommitteeState.processInstructionFromBeacon(
		env.BeaconInstructions(),
		env.ShardID(), NewCommitteeChange())

	if err != nil {
		return nil, err
	}

	return committeeChange, nil
}

//ProcessInstructionFromShard :
func (engine *ShardCommitteeEngineV2) ProcessInstructionFromShard(env ShardCommitteeStateEnvironment) (*CommitteeChange, error) {
	return nil, nil
}
