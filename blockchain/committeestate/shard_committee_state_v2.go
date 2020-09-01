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
	shardCommittee []incognitokey.CommitteePublicKey
	//TODO: @hung remove shard substitute
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
func NewShardCommitteeStateV2WithValue(shardCommittee, shardSubstitute []incognitokey.CommitteePublicKey) *ShardCommitteeStateV2 {
	return &ShardCommitteeStateV2{
		shardCommittee:  shardCommittee,
		shardSubstitute: shardSubstitute,
		mu:              new(sync.RWMutex),
	}
}

//NewShardCommitteeEngineV1 is default constructor for ShardCommitteeEngineV2
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
func (s ShardCommitteeStateV2) clone(newCommitteeState *ShardCommitteeStateV2) {
	newCommitteeState.reset()

	newCommitteeState.shardCommittee = make([]incognitokey.CommitteePublicKey, len(s.shardCommittee))
	for i, v := range s.shardCommittee {
		newCommitteeState.shardCommittee[i] = v
	}

	newCommitteeState.shardSubstitute = make([]incognitokey.CommitteePublicKey, len(s.shardSubstitute))
	for i, v := range s.shardSubstitute {
		newCommitteeState.shardSubstitute[i] = v
	}
}

//reset : reset ShardCommitteeStateV2 to default value
func (s *ShardCommitteeStateV2) reset() {
	s.shardCommittee = make([]incognitokey.CommitteePublicKey, 0)
	s.shardSubstitute = make([]incognitokey.CommitteePublicKey, 0)
}

//GetShardCommittee get shard committees
func (engine *ShardCommitteeEngineV2) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.shardCommitteeStateV2.shardCommittee
}

//GetShardSubstitute get shard pending validators
func (engine *ShardCommitteeEngineV2) GetShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey {
	return engine.shardCommitteeStateV2.shardSubstitute
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

//InitCommitteeState init committee state at genesis block or anytime restore program
//	- call function processInstructionFromBeacon for process instructions received from beacon
//	- call function processShardBlockInstruction for process shard block instructions
func (engine *ShardCommitteeEngineV2) InitCommitteeState(env ShardCommitteeStateEnvironment) {
	engine.shardCommitteeStateV2.mu.Lock()
	defer engine.shardCommitteeStateV2.mu.Unlock()

	committeeChange := NewCommitteeChange()
	candidates := []string{}

	for _, beaconInstruction := range env.BeaconInstructions() {
		if beaconInstruction[0] == instruction.STAKE_ACTION {
			candidates = strings.Split(beaconInstruction[1], ",")
		}
	}

	newShardCandidateStructs := []incognitokey.CommitteePublicKey{}
	for _, candidate := range candidates {
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

	engine.shardCommitteeStateV2.shardCommittee = append(engine.shardCommitteeStateV2.shardCommittee,
		addedCommittees...)
	committeeChange.ShardCommitteeAdded[env.ShardID()] = addedCommittees

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
	var err error
	newCommitteeState := engine.uncommittedShardCommitteeStateV2
	committeeChange := NewCommitteeChange()

	committeeChange, err = newCommitteeState.processShardBlockInstruction(env, committeeChange)
	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}

	hashes, err := engine.generateUncommittedCommitteeHashes()
	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}

	return hashes, committeeChange, nil
}

func (engine *ShardCommitteeEngineV2) GenerateConfirmShardSwapInstruction(env ShardCommitteeStateEnvironment) (*instruction.ConfirmShardSwapInstruction, []string, error) {
	confirmShardSwapInstruction := instruction.NewConfirmShardSwapInstruction()
	for _, beaconInstruction := range env.BeaconInstructions() {
		if len(beaconInstruction) == 0 {
			continue
		}
		if beaconInstruction[0] == instruction.REQUEST_SHARD_SWAP_ACTION {
			Logger.log.Infof("GenerateConfirmShardSwapInstruction, shard height %+v, beacon height %+v", env.ShardHeight(), env.BeaconHeight())
			requestShardSwapInstruction, err := instruction.ValidateAndImportRequestShardSwapInstructionFromString(beaconInstruction)
			if err != nil {
				// Return Error for debug purpose
				return &instruction.ConfirmShardSwapInstruction{}, []string{}, err
			}
			if byte(requestShardSwapInstruction.ChainID) == env.ShardID() {
				confirmShardSwapInstruction = instruction.ConvertRequestToConfirmShardSwapInstruction(requestShardSwapInstruction)
				shardCommittees, _ := incognitokey.CommitteeKeyListToString(engine.shardCommitteeStateV2.shardCommittee)
				fixedProducerShardValidators := shardCommittees[:env.NumberOfFixedBlockValidators()]
				shardCommittees = shardCommittees[env.NumberOfFixedBlockValidators():]
				newShardCommittees, err := getNewShardCommittees(confirmShardSwapInstruction, shardCommittees)
				if err != nil {
					return &instruction.ConfirmShardSwapInstruction{}, []string{}, err
				}
				newShardCommittees = append(fixedProducerShardValidators, newShardCommittees...)
				Logger.log.Infof("GenerateConfirmShardSwapInstruction, confirmShardSwapInstruction %+v ", confirmShardSwapInstruction)
				return confirmShardSwapInstruction, newShardCommittees, nil
			}
		}
	}
	return &instruction.ConfirmShardSwapInstruction{}, []string{}, nil
}

func getNewShardCommittees(
	confirmShardSwapInstruction *instruction.ConfirmShardSwapInstruction,
	shardCommittees []string,
) ([]string, error) {
	newShardCommittees, err := removeValidatorV2(shardCommittees, confirmShardSwapInstruction.OutPublicKeys)
	if err != nil {
		return []string{}, err
	}
	newShardCommittees = append(newShardCommittees, confirmShardSwapInstruction.InPublicKeys...)
	return newShardCommittees, nil
}
func (engine *ShardCommitteeEngineV2) GenerateSwapInstruction(env ShardCommitteeStateEnvironment) (*instruction.SwapInstruction, []string, []string, error) {
	shardSubsitutes, _ := incognitokey.CommitteeKeyListToString(engine.shardCommitteeStateV2.shardSubstitute)
	shardCommittees, _ := incognitokey.CommitteeKeyListToString(engine.shardCommitteeStateV2.shardCommittee)
	return instruction.NewSwapInstruction(), shardSubsitutes, shardCommittees, nil
}

// processInstructionFromBeacon process instruction from beacon blocks
//	- Get all subtitutes in shard
//  - Loop over the list instructions:
//		+ Create Assign instruction struct from assign instruction string
//	- Update shard subtitute added in committee change struct
//	- Only call once in new or insert block process
func (s *ShardCommitteeStateV2) processInstructionFromBeacon(
	listInstructions [][]string,
	shardID byte,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {
	return committeeChange, nil
}

//processShardBlockInstruction process shard block instruction for sending to beacon
//	- get list instructions from input environment
//	- loop over the list instructions
//		+ Check type of instructions and process itp
//		+ At this moment, there will be only swap action for this function
//	- After process all instructions, we will updatew commitee change variable
//	- Only call once in new or insert block process
func (s *ShardCommitteeStateV2) processShardBlockInstruction(
	env ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {
	var err error
	shardID := env.ShardID()
	shardCommittees, err := incognitokey.CommitteeKeyListToString(s.shardCommittee)
	if err != nil {
		return nil, err
	}
	fixedProducerShardValidators := s.shardCommittee[:env.NumberOfFixedBlockValidators()]
	shardCommittees = shardCommittees[env.NumberOfFixedBlockValidators():]
	// Swap committee
	for _, inst := range env.ShardInstructions() {
		if len(inst) == 0 {
			continue
		}
		if inst[0] == instruction.CONFIRM_SHARD_SWAP_ACTION {
			confirmShardSwapInstruction, err := instruction.ValidateAndImportConfirmShardSwapInstructionFromString(inst)
			if err != nil {
				return committeeChange, err
			}
			tempNewShardCommittees, err := getNewShardCommittees(confirmShardSwapInstruction, shardCommittees)
			if err != nil {
				return committeeChange, err
			}
			newShardCommittees, _ := incognitokey.CommitteeBase58KeyListToStruct(tempNewShardCommittees)
			s.shardCommittee = append(fixedProducerShardValidators, newShardCommittees...)
			committeeChange.ShardCommitteeAdded[shardID] = append(committeeChange.ShardCommitteeAdded[shardID], confirmShardSwapInstruction.InPublicKeyStructs...)
			committeeChange.ShardCommitteeRemoved[shardID] = append(committeeChange.ShardCommitteeAdded[shardID], confirmShardSwapInstruction.OutPublicKeyStructs...)
		}
	}
	return committeeChange, nil
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

	substitutesStr, err := incognitokey.CommitteeKeyListToString(newCommitteeState.shardSubstitute)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	substituteHash, err := common.GenerateHashFromStringArray(substitutesStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	return &ShardCommitteeStateHash{
		ShardCommitteeHash:  committeeHash,
		ShardSubstituteHash: substituteHash,
	}, nil
}
