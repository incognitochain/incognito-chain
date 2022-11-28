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
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/pkg/errors"
)

func InitGenesisShardCommitteeState(beaconHeight, stakingFlowV2, stakingFlowV3, stakingFlowV4 uint64,
	env *ShardCommitteeStateEnvironment) ShardCommitteeState {
	version := VersionByBeaconHeight(beaconHeight, stakingFlowV2, stakingFlowV3, stakingFlowV4)
	switch version {
	case SELF_SWAP_SHARD_VERSION:
		return initGenesisShardCommitteeStateV1(env)
	case STAKING_FLOW_V2:
		return initGenesisShardCommitteeStateV2(env)
	case STAKING_FLOW_V3, STAKING_FLOW_V4:
		return initGenesisShardCommitteeStateV3(env)
	default:
		panic("not a valid shard committee state version")
	}
}

// ShardCommitteeStateHash
type ShardCommitteeStateHash struct {
	ShardCommitteeHash        common.Hash
	ShardSubstituteHash       common.Hash
	CommitteeFromBlock        common.Hash
	SubsetCommitteesFromBlock common.Hash
}

// ShardCommitteeStateV1
type ShardCommitteeStateV1 struct {
	shardCommittee  []string
	shardSubstitute []string

	mu *sync.RWMutex
}

// NewShardCommitteeStateV1 is default constructor for ShardCommitteeStateV1 ...
// Output: pointer of ShardCommitteeStateV1 struct
func NewShardCommitteeStateV1() *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{
		mu: new(sync.RWMutex),
	}
}

// NewShardCommitteeStateV1WithValue is constructor for ShardCommitteeStateV1 with value
// Output: pointer of ShardCommitteeStateV1 struct with value
func NewShardCommitteeStateV1WithValue(shardCommittee, shardSubstitute []incognitokey.CommitteePublicKey) *ShardCommitteeStateV1 {
	tempShardCommittee, _ := incognitokey.CommitteeKeyListToString(shardCommittee)
	tempShardSubstitute, _ := incognitokey.CommitteeKeyListToString(shardSubstitute)
	return &ShardCommitteeStateV1{
		shardCommittee:  tempShardCommittee,
		shardSubstitute: tempShardSubstitute,
		mu:              new(sync.RWMutex),
	}
}

// clone ShardCommitteeStateV1 to new instance
func (s ShardCommitteeStateV1) clone(newCommitteeState *ShardCommitteeStateV1) {
	newCommitteeState.shardCommittee = common.DeepCopyString(s.shardCommittee)
	newCommitteeState.shardSubstitute = common.DeepCopyString(s.shardSubstitute)
}

// Clone ...
func (s *ShardCommitteeStateV1) Clone() ShardCommitteeState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	newS := NewShardCommitteeStateV1()
	s.clone(newS)
	return newS
}

// Version get version of engine
func (s *ShardCommitteeStateV1) Version() int {
	return SELF_SWAP_SHARD_VERSION
}

// GetShardCommittee get shard committees
func (s *ShardCommitteeStateV1) GetShardCommittee() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(s.shardCommittee)
	return res
}

// GetShardSubstitute get shard pending validators
func (s *ShardCommitteeStateV1) GetShardSubstitute() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(s.shardSubstitute)
	return res
}

func (s *ShardCommitteeStateV1) SubsetCommitteesFromBlock() common.Hash {
	return common.Hash{}
}

// initGenesisShardCommitteeStateV1 init committee state at genesis block or anytime restore program
//   - call function processInstructionFromBeacon for process instructions received from beacon
//   - call function processShardBlockInstruction for process shard block instructions
func initGenesisShardCommitteeStateV1(env *ShardCommitteeStateEnvironment) *ShardCommitteeStateV1 {
	s := NewShardCommitteeStateV1()

	shardPendingValidator := []string{}
	newSubstituteStructs := []incognitokey.CommitteePublicKey{}

	shardsCommittees := []string{}

	for _, beaconInstruction := range env.BeaconInstructions {
		if beaconInstruction[0] == instruction.STAKE_ACTION {
			shardsCommittees = strings.Split(beaconInstruction[1], ",")
		}
		if beaconInstruction[0] == instruction.ASSIGN_ACTION {
			assignInstruction, err := instruction.ValidateAndImportAssignInstructionFromString(beaconInstruction)
			if err == nil && assignInstruction.ChainID == int(env.ShardID) {
				shardPendingValidator = append(shardPendingValidator, assignInstruction.ShardCandidates...)
				newSubstituteStructs = append(newSubstituteStructs, assignInstruction.ShardCandidatesStruct...)
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

	newCommitteeStructs := []incognitokey.CommitteePublicKey{}
	newCommitteeStructs = append(newCommitteeStructs, newShardCandidateStructs[int(env.ShardID)*
		env.MinShardCommitteeSize:(int(env.ShardID)*env.MinShardCommitteeSize)+env.MinShardCommitteeSize]...)

	newCommittees, _ := incognitokey.CommitteeKeyListToString(newCommitteeStructs)
	newSubstitutes, _ := incognitokey.CommitteeKeyListToString(newSubstituteStructs)

	s.shardCommittee = append(s.shardCommittee, newCommittees...)
	s.shardSubstitute = append(s.shardSubstitute, newSubstitutes...)

	return s
}

// UpdateCommitteeState update committeState from valid data before
//   - call process instructions from beacon
//   - check conditions for epoch timestamp
//   - process shard block instructions for key
//   - process shard block instructions normally
//   - hash for checking commit later
//   - Only call once in new or insert block process
func (s *ShardCommitteeStateV1) UpdateCommitteeState(
	env *ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash, *CommitteeChange, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	committeeChange, err := s.processInstructionFromBeacon(env.BeaconInstructions, env.ShardID, NewCommitteeChange())

	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}

	if common.IndexOfUint64(env.Epoch, env.EpochBreakPointSwapNewKey) > -1 {
		committeeChange, err = s.processShardBlockInstructionForKeyListV2(env, committeeChange)
	} else {
		committeeChange, err = s.processShardBlockInstruction(env, committeeChange)
	}

	if err != nil {
		return nil, nil, err
	}

	hashes, err := s.hash()
	if err != nil {
		return nil, nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}

	return hashes, committeeChange, nil
}

func (s *ShardCommitteeStateV1) GenerateSwapInstructions(env *ShardCommitteeStateEnvironment) (*instruction.SwapInstruction, []string, []string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	shardSubstitutes := common.DeepCopyString(s.shardSubstitute)
	shardCommittees := common.DeepCopyString(s.shardCommittee)
	fixedProducerShardValidators := make([]string, env.NumberOfFixedBlockValidators)
	copy(fixedProducerShardValidators, shardCommittees[:env.NumberOfFixedBlockValidators])
	shardCommittees = shardCommittees[env.NumberOfFixedBlockValidators:]
	swapInstruction, shardPendingValidator, shardCommittees, err := createSwapInstruction(
		shardSubstitutes,
		shardCommittees,
		env.MaxShardCommitteeSize,
		env.MinShardCommitteeSize,
		env.ShardID,
		env.Offset,
		env.SwapOffset,
	)
	if err != nil {
		Logger.log.Error(err)
		return swapInstruction, shardPendingValidator, shardCommittees, err
	}
	return swapInstruction, shardPendingValidator, append(fixedProducerShardValidators, shardCommittees...), nil
}

// hash generate hashes relate to uncommitted committees of struct ShardCommitteeStateV1
//
//	append committees and subtitutes to struct and hash it
func (s ShardCommitteeStateV1) hash() (*ShardCommitteeStateHash, error) {
	if reflect.DeepEqual(s, NewShardCommitteeStateV1()) {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, empty uncommitted state")
	}

	committeeHash, err := common.GenerateHashFromStringArray(s.shardCommittee)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	substituteHash, err := common.GenerateHashFromStringArray(s.shardSubstitute)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	return &ShardCommitteeStateHash{
		ShardCommitteeHash:  committeeHash,
		ShardSubstituteHash: substituteHash,
	}, nil
}

// processnstructionFromBeacon process instruction from beacon blocks
//   - Get all subtitutes in shard
//   - Loop over the list instructions:
//   - Create Assign instruction struct from assign instruction string
//   - Update shard subtitute added in committee change struct
//   - Only call once in new or insert block process
func (s *ShardCommitteeStateV1) processInstructionFromBeacon(
	listInstructions [][]string,
	shardID byte,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {
	newSubstitutes := extractAssignInstruction(listInstructions, shardID)
	temp, _ := incognitokey.CommitteeBase58KeyListToStruct(newSubstitutes)
	s.shardSubstitute = append(s.shardSubstitute, newSubstitutes...)
	committeeChange.ShardSubstituteAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], temp...)
	return committeeChange, nil
}

func extractAssignInstruction(listInstructions [][]string, shardID byte) []string {
	newSubstitutes := []string{}
	for _, inst := range listInstructions {
		assignInstruction, err := instruction.ValidateAndImportAssignInstructionFromString(inst)
		if err == nil && assignInstruction.ChainID == int(shardID) {
			newSubstitutes = append(newSubstitutes, assignInstruction.ShardCandidates...)
		}
	}
	return newSubstitutes
}

// processShardBlockInstruction process shard block instruction for sending to beacon
//   - get list instructions from input environment
//   - loop over the list instructions
//   - Check type of instructions and process itp
//   - At this moment, there will be only swap action for this function
//   - After process all instructions, we will updatew commitee change variable
//   - Only call once in new or insert block process
func (committeeState *ShardCommitteeStateV1) processShardBlockInstruction(
	env *ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {
	newCommitteeChange := committeeChange
	shardID := env.ShardID
	if len(env.ShardInstructions) != 0 {
		Logger.log.Debugf("Shard Process/processShardBlockInstruction: Shard Instruction %+v", env.ShardInstructions)
	}
	// Swap committee
	for _, ins := range env.ShardInstructions {
		swapInstruction, err := instruction.ValidateAndImportSwapInstructionFromString(ins)
		if err == nil {
			shardPendingValidator := committeeState.shardSubstitute
			shardCommittee := committeeState.shardCommittee
			fixedProducerShardValidators := shardCommittee[:env.NumberOfFixedBlockValidators]
			shardCommittee = shardCommittee[env.NumberOfFixedBlockValidators:]
			shardSwappedCommittees := []string{}
			shardNewCommittees := []string{}
			// #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator
			maxShardCommitteeSize := env.MaxShardCommitteeSize - env.NumberOfFixedBlockValidators
			var minShardCommitteeSize int
			if env.MinShardCommitteeSize-env.NumberOfFixedBlockValidators < 0 {
				minShardCommitteeSize = 0
			} else {
				minShardCommitteeSize = env.MinShardCommitteeSize - env.NumberOfFixedBlockValidators
			}
			shardPendingValidator, shardCommittee, shardSwappedCommittees, shardNewCommittees, err =
				SwapValidator(shardPendingValidator,
					shardCommittee, maxShardCommitteeSize,
					minShardCommitteeSize, env.Offset,
					env.SwapOffset)

			if err != nil {
				Logger.log.Errorf("SHARD %+v | Blockchain Error %+v", err)
				return nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

			if !reflect.DeepEqual(swapInstruction.OutPublicKeys, shardSwappedCommittees) {
				return nil, NewCommitteeStateError(ErrUpdateCommitteeState,
					errors.Errorf("Expect swapped committees to be %+v but get %+v",
						swapInstruction.OutPublicKeys, shardSwappedCommittees))
			}

			if !reflect.DeepEqual(swapInstruction.InPublicKeys, shardNewCommittees) {
				return nil, NewCommitteeStateError(ErrUpdateCommitteeState,
					errors.Errorf("Expect new committees to be %+v but get %+v",
						swapInstruction.InPublicKeys, shardNewCommittees))
			}

			shardNewCommitteesStruct, err := incognitokey.CommitteeBase58KeyListToStruct(shardNewCommittees)
			if err != nil {
				return nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

			shardSwappedCommitteesStruct, err := incognitokey.CommitteeBase58KeyListToStruct(shardSwappedCommittees)
			if err != nil {
				return nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
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
			committeeState.shardSubstitute = shardPendingValidator
			if err != nil {
				return nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}

			committeeState.shardCommittee = append(fixedProducerShardValidators, shardCommittee...)
			if err != nil {
				return nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			newCommitteeChange.ShardCommitteeAdded[shardID] = shardNewCommitteesStruct
			newCommitteeChange.ShardCommitteeRemoved[shardID] = shardSwappedCommitteesStruct
			newCommitteeChange.ShardSubstituteRemoved[shardID] = filteredShardSubstituteRemoved
			newCommitteeChange.ShardSubstituteAdded[shardID] = filteredShardSubstituteAdded
			Logger.log.Infof("SHARD %+v | Swap: Out committee %+v", shardID, shardSwappedCommittees)
			Logger.log.Infof("SHARD %+v | Swap: In committee %+v", shardID, shardNewCommittees)
			break
		}
	}

	return newCommitteeChange, nil
}

// processShardBlockInstructionForKeyListV2 process shard block instructions for key list v2
//   - get list instructions from input environment
//   - loop over the list instructions
//   - Check type of instructions and process it
//   - At this moment, there will be only swap action for this function
func (s *ShardCommitteeStateV1) processShardBlockInstructionForKeyListV2(
	env *ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {
	shardID := env.ShardID
	newCommitteeChange := committeeChange
	for _, inst := range env.ShardInstructions {
		if inst[0] == instruction.SWAP_ACTION {
			swapInstruction, err := instruction.ValidateAndImportSwapInstructionFromString(inst)
			if err != nil {
				return nil, NewCommitteeStateError(ErrUpdateCommitteeState, err)
			}
			removedCommitteeSize := len(swapInstruction.InPublicKeys)
			remainedShardCommittees := common.DeepCopyString(s.shardCommittee[removedCommitteeSize:])
			tempShardSwappedCommittees := common.DeepCopyString(s.shardCommittee[:env.MinShardCommitteeSize])
			if !reflect.DeepEqual(swapInstruction.OutPublicKeys, tempShardSwappedCommittees) {
				return nil, NewCommitteeStateError(ErrUpdateCommitteeState,
					fmt.Errorf("expect swapped committe %+v but got %+v", tempShardSwappedCommittees, swapInstruction.OutPublicKeys))
			}
			s.shardCommittee = common.DeepCopyString(append(swapInstruction.InPublicKeys, remainedShardCommittees...))
			committeeReplace := [2][]incognitokey.CommitteePublicKey{}
			committeeReplace[common.REPLACE_IN] = incognitokey.DeepCopy(swapInstruction.InPublicKeyStructs)
			committeeReplace[common.REPLACE_OUT] = incognitokey.DeepCopy(swapInstruction.OutPublicKeyStructs)
			committeeChange.ShardCommitteeReplaced[shardID] = committeeReplace
		}
	}
	return newCommitteeChange, nil
}

// ProcessInstructionFromBeacon : process instrucction from beacon
func (s ShardCommitteeStateV1) ProcessAssignInstructions(env *ShardCommitteeStateEnvironment) []incognitokey.CommitteePublicKey {
	newSubstitutes := extractAssignInstruction(env.BeaconInstructions, env.ShardID)
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(newSubstitutes)
	return res
}

func (s ShardCommitteeStateV1) BuildTotalTxsFeeFromTxs(txs []metadata.Transaction) map[common.Hash]uint64 {
	totalTxsFee := make(map[common.Hash]uint64)
	for _, tx := range txs {
		totalTxsFee[*tx.GetTokenID()] += tx.GetTxFee()
		if tx.GetType() == common.TxCustomTokenPrivacyType {
			txTokenData := transaction.GetTxTokenDataFromTransaction(tx)
			totalTxsFee[txTokenData.PropertyID] = txTokenData.TxNormal.GetTxFee()
		}
	}
	return totalTxsFee
}
