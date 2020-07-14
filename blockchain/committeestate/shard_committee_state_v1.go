package committeestate

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
)

//ShardCommitteeStateHash :
type ShardCommitteeStateHash struct {
	ShardCommitteeHash  common.Hash
	ShardSubstituteHash common.Hash
}

//ShardCommitteeStateEnvironment :
type ShardCommitteeStateEnvironment struct {
	AddedCommitteesStr                         []string
	ShardHeight                                uint64
	ShardBlockHash                             uint64
	Txs                                        []metadata.Transaction
	Instructions                               [][]string
	BeaconHeight                               uint64
	ChainParamEpoch                            uint64
	EpochBreakPointSwapNewKey                  []uint64
	ShardID                                    byte
	MaxShardCommitteeSize                      int
	MinShardCommitteeSize                      int
	Offset                                     int
	SwapOffset                                 int
	ProducersBlackList                         map[string]uint8
	StakingTx                                  map[string]string
	IsNullView                                 bool
	IsProcessShardBlockInstruction             bool
	IsProcessShardBlockInstructionForKeyListV2 bool
}

//NewShardCommitteeStateEnvironment : Default constructor of ShardCommitteeStateEnvironment
//Output: pointer of ShardCommitteeStateEnvironment
func NewShardCommitteeStateEnvironment(
	addedCommitteesStr []string,
	txs []metadata.Transaction,
	beaconHeight uint64,
	instructions [][]string,
	chainParamEpoch uint64,
	epochBreakPointSwapNewKey []uint64,
	shardID byte,
	maxShardCommitteeSize, minShardCommitteeSize, offset, swapOffset int,
	producersBlackList map[string]uint8,
	stakingTx map[string]string,
	isNullView bool,
	isProcessShardBlockInstruction bool,
	isProcessShardBlockInstructionForKeyListV2 bool) *ShardCommitteeStateEnvironment {
	return &ShardCommitteeStateEnvironment{
		AddedCommitteesStr:             addedCommitteesStr,
		Txs:                            txs,
		BeaconHeight:                   beaconHeight,
		Instructions:                   instructions,
		ChainParamEpoch:                chainParamEpoch,
		EpochBreakPointSwapNewKey:      epochBreakPointSwapNewKey,
		ProducersBlackList:             make(map[string]uint8),
		StakingTx:                      make(map[string]string),
		MaxShardCommitteeSize:          maxShardCommitteeSize,
		MinShardCommitteeSize:          minShardCommitteeSize,
		Offset:                         offset,
		SwapOffset:                     swapOffset,
		ShardID:                        shardID,
		IsNullView:                     isNullView,
		IsProcessShardBlockInstruction: isProcessShardBlockInstruction,
		IsProcessShardBlockInstructionForKeyListV2: isProcessShardBlockInstructionForKeyListV2,
	}
}

//ShardCommitteeStateV1 :
type ShardCommitteeStateV1 struct {
	shardCommittee        []incognitokey.CommitteePublicKey
	shardPendingValidator []incognitokey.CommitteePublicKey

	mu *sync.RWMutex
}

//ShardCommitteeEngine :
type ShardCommitteeEngine struct {
	shardHeight                      uint64
	shardHash                        common.Hash
	shardID                          byte
	shardCommitteeStateV1            *ShardCommitteeStateV1
	uncommittedShardCommitteeStateV1 *ShardCommitteeStateV1
}

//NewShardCommitteeStateV1 : Default constructor for ShardCommitteeStateV1 ...
//Output: pointer of ShardCommitteeStateV1 struct
func NewShardCommitteeStateV1() *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{
		mu: new(sync.RWMutex),
	}
}

//NewShardCommitteeStateV1WithValue : Constructor for ShardCommitteeStateV1 with value
//Output: pointer of ShardCommitteeStateV1 struct with value
func NewShardCommitteeStateV1WithValue(shardCommittee, shardPendingValidator []incognitokey.CommitteePublicKey) *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{
		shardCommittee:        shardCommittee,
		shardPendingValidator: shardPendingValidator,
		mu:                    new(sync.RWMutex),
	}
}

//NewShardCommitteeEngine : Default constructor for ShardCommitteeEngine
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

//clone: clone ShardCommitteeStateV1 to new instance
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

//ValidateCommitteeRootHashes : Validate committee root hashes for checking if it's valid
//Input: list rootHashes need checking
//Output: result(boolean) and error
func (engine *ShardCommitteeEngine) ValidateCommitteeRootHashes(rootHashes []common.Hash) (bool, error) {
	panic("Not implemented yet")
}

//Commit : Commit commitee state change in uncommittedShardCommitteeStateV1 struct
//Pre-conditions: uncommittedShardCommitteeStateV1 has been inited
//Input: Shard Committee hash
//Output: error
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

	fmt.Println(comparedHashes)

	// if hashes != nil || comparedHashes != nil {

	// 	if comparedHashes.ShardCommitteeHash.IsEqual(&hashes.ShardCommitteeHash) {
	// 		return NewCommitteeStateError(ErrCommitShardCommitteeState, fmt.Errorf("Uncommitted ShardCommitteeHash want value %+v but have %+v",
	// 			comparedHashes.ShardCommitteeHash, hashes.ShardCommitteeHash))
	// 	}

	// 	if comparedHashes.ShardSubstituteHash.IsEqual(&hashes.ShardSubstituteHash) {
	// 		return NewCommitteeStateError(ErrCommitShardCommitteeState, fmt.Errorf("Uncommitted ShardSubstituteHash want value %+v but have %+v",
	// 			comparedHashes.ShardSubstituteHash, hashes.ShardSubstituteHash))
	// 	}

	// }

	engine.uncommittedShardCommitteeStateV1.clone(engine.shardCommitteeStateV1)
	engine.uncommittedShardCommitteeStateV1.reset()
	return nil
}

//AbortUncommittedShardState : Reset data in uncommittedShardCommitteeStateV1 struct
//Pre-conditions: uncommittedShardCommitteeStateV1 has been inited
//Input: NULL
//Output: error
func (engine *ShardCommitteeEngine) AbortUncommittedShardState() {
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()
	engine.uncommittedShardCommitteeStateV1.reset()
}

//UpdateCommitteeState : Update committeState from valid data before
//Pre-conditions: Validate committee state
//Input: env variables ShardCommitteeStateEnvironment
//Output: New ShardCommitteeEngineV1 and committee changes, error
func (engine *ShardCommitteeEngine) UpdateCommitteeState(
	env *ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash, *CommitteeChange, error) {
	var err error
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()
	engine.shardCommitteeStateV1.mu.RLock()
	engine.shardCommitteeStateV1.clone(engine.uncommittedShardCommitteeStateV1)
	engine.shardCommitteeStateV1.mu.RUnlock()
	newCommitteeState := engine.uncommittedShardCommitteeStateV1
	committeeChange := NewCommitteeChange()

	// TODO: separate function, each function should hold only one responsible
	// Avoid using boolean condition to branch
	// @tin

	err = newCommitteeState.processInstructionFromBeacon(env.IsNullView,
		env.Instructions, env.ShardID, committeeChange)
	if err != nil {
		return nil, nil, err
	}

	if common.IndexOfUint64(env.BeaconHeight/env.ChainParamEpoch, env.EpochBreakPointSwapNewKey) > -1 {
		if env.IsProcessShardBlockInstructionForKeyListV2 {
			err = newCommitteeState.processShardBlockInstructionForKeyListV2(env, committeeChange)
		}
	} else {
		if env.IsProcessShardBlockInstruction {
			err = newCommitteeState.processShardBlockInstruction(env, committeeChange)
		}
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

//InitCommitteeState : Init committee state at genesis block or anytime restore program
//Pre-conditions: right config from files or env variables
//Input: env variables ShardCommitteeStateEnvironment
//Output: NULL
func (engine *ShardCommitteeEngine) InitCommitteeState(env *ShardCommitteeStateEnvironment) {
	engine.shardCommitteeStateV1.mu.Lock()
	defer engine.shardCommitteeStateV1.mu.Unlock()

	committeeState := engine.shardCommitteeStateV1

	committeeChange := NewCommitteeChange()

	err := committeeState.processInstructionFromBeacon(env.IsNullView,
		env.Instructions, env.ShardID, committeeChange)
	if err != nil {
		panic(err)
	}

	err = committeeState.processShardBlockInstruction(env, committeeChange)
	if err != nil {
		panic(err)
	}

	committess, err := incognitokey.CommitteeBase58KeyListToStruct(env.AddedCommitteesStr)
	if err != nil {
		panic(err)
	}

	engine.shardCommitteeStateV1.shardCommittee = append(engine.shardCommitteeStateV1.shardCommittee, committess...)

}

//GetShardCommittee : Get shard committees
//Input: NULL
//Output: list array of incognito public keys
func (engine *ShardCommitteeEngine) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	// engine.shardCommitteeStateV1.mu.RLock()
	// defer engine.shardCommitteeStateV1.mu.RUnlock()
	return engine.shardCommitteeStateV1.shardCommittee
}

//GetShardPendingValidator : Get shard pending validators
//Input: NULL
//Output: list array of incognito public keys
func (engine *ShardCommitteeEngine) GetShardPendingValidator(shardID byte) []incognitokey.CommitteePublicKey {
	// engine.shardCommitteeStateV1.mu.RLock()
	// defer engine.shardCommitteeStateV1.mu.RUnlock()
	return engine.shardCommitteeStateV1.shardPendingValidator
}

//generateUncommittedCommitteeHashes : generate hashes relate to uncommitted committees of struct ShardCommitteeEngine
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

//processInstructionFromBeacon :
// Pre-conditions:
// Input:
// Output:
// Post-conditions:
// Process Instruction From Beacon Blocks:
func (committeeState *ShardCommitteeStateV1) processInstructionFromBeacon(
	isNullView bool,
	listInstructions [][]string,
	shardID byte,
	committeeChange *CommitteeChange) error {

	shardPendingValidator := []string{}

	if !isNullView {
		var err error
		shardPendingValidator, err = incognitokey.
			CommitteeKeyListToString(committeeState.shardPendingValidator)
		if err != nil {
			return err
		}
	}

	newShardPendingValidator := []string{}

	for _, ins := range listInstructions {
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

	var err error
	committeeChange.ShardSubstituteAdded[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(newShardPendingValidator)
	if err != nil {
		return err
	}

	return nil
}

//processShardBlockInstruction :
// Pre-conditions:
// Input:
// Output:
// Post-conditions:
func (committeeState *ShardCommitteeStateV1) processShardBlockInstruction(
	env *ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) error {

	var err error
	shardID := env.ShardID
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
	if len(env.Instructions) != 0 {
		Logger.log.Debugf("Shard Process/processShardBlockInstruction: Shard Instruction %+v", env.Instructions)
	}

	// Swap committee
	for _, l := range env.Instructions {
		if l[0] == instruction.SWAP_ACTION {
			// #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator
			shardPendingValidator, shardCommittee, shardSwappedCommittees, shardNewCommittees, err = SwapValidator(shardPendingValidator,
				shardCommittee, env.MaxShardCommitteeSize, env.MinShardCommitteeSize, env.Offset,
				env.ProducersBlackList, env.SwapOffset)
			if err != nil {
				Logger.log.Errorf("SHARD %+v | Blockchain Error %+v", err)
				return err
			}
			swapedCommittees := []string{}
			if len(l[2]) != 0 && l[2] != "" {
				swapedCommittees = strings.Split(l[2], ",")
			}
			for _, v := range swapedCommittees {
				if txID, ok := env.StakingTx[v]; ok {
					if checkReturnStakingTxExistence(txID, env.Txs) {
						delete(env.StakingTx, v)
					}
				}
			}
			if !reflect.DeepEqual(swapedCommittees, shardSwappedCommittees) {
				return fmt.Errorf("Expect swapped committees to be %+v but get %+v", swapedCommittees, shardSwappedCommittees)
			}
			newCommittees := []string{}
			if len(l[1]) > 0 {
				newCommittees = strings.Split(l[1], ",")
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

//processShardBlockInstructionForKeyListV2 :
// Pre-conditions:
// Input:
// Output:
// Post-conditions:
func (committeeState *ShardCommitteeStateV1) processShardBlockInstructionForKeyListV2(
	env *ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) error {
	shardID := env.ShardID
	for _, ins := range env.Instructions {
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
			tempShardSwappedCommittees := committeeState.shardCommittee[:env.MinShardCommitteeSize]
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
