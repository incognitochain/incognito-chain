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
)

//ShardCommitteeStateV2
type ShardCommitteeStateV2 struct {
	shardCommittee            []incognitokey.CommitteePublicKey
	committeeFromBlock        common.Hash //Committees From Beacon Block Hash
	committeesSubsetFromBlock common.Hash //Committees From Beacon Block Hash

	mu *sync.RWMutex
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
func NewShardCommitteeStateV2WithValue(
	shardCommittee []incognitokey.CommitteePublicKey,
	committeeFromBlockHash common.Hash,
	committeesSubsetFromBlockHash common.Hash,
) *ShardCommitteeStateV2 {
	return &ShardCommitteeStateV2{
		shardCommittee:            incognitokey.DeepCopy(shardCommittee),
		committeeFromBlock:        committeeFromBlockHash,
		committeesSubsetFromBlock: committeesSubsetFromBlockHash,
		mu:                        new(sync.RWMutex),
	}
}

//Clone ...
func (s *ShardCommitteeStateV2) Clone() ShardCommitteeState {
	newS := NewShardCommitteeStateV2()
	s.clone(newS)
	return newS
}

//clone ShardCommitteeStateV2 to new instance
func (s ShardCommitteeStateV2) clone(newCommitteeState *ShardCommitteeStateV2) {
	newCommitteeState.shardCommittee = incognitokey.DeepCopy(s.shardCommittee)
	newCommitteeState.committeeFromBlock = s.committeeFromBlock
	newCommitteeState.committeesSubsetFromBlock = s.committeesSubsetFromBlock
}

//Version ...
func (s *ShardCommitteeStateV2) Version() int {
	return SLASHING_VERSION
}

//GetShardCommittee get shard committees
func (s *ShardCommitteeStateV2) GetShardCommittee() []incognitokey.CommitteePublicKey {
	return incognitokey.DeepCopy(s.shardCommittee)
}

//GetShardSubstitute get shard pending validators
func (s *ShardCommitteeStateV2) GetShardSubstitute() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

func (s *ShardCommitteeStateV2) GetCommitteeFromBlock() common.Hash {
	return s.committeeFromBlock
}

//InitCommitteeState init committee state at genesis block or anytime restore program
//	- call function processInstructionFromBeacon for process instructions received from beacon
//	- call function processShardBlockInstruction for process shard block instructions
func (s *ShardCommitteeStateV2) InitCommitteeState(env ShardCommitteeStateEnvironment) {
	s.mu.Lock()
	defer s.mu.Unlock()

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

	s.shardCommittee = incognitokey.DeepCopy(addedCommittees)
	committeeChange.ShardCommitteeAdded[env.ShardID()] = addedCommittees

}

//InitGenesisShardCommitteeStateV2 init committee state at genesis block or anytime restore program
//	- call function processInstructionFromBeacon for process instructions received from beacon
//	- call function processShardBlockInstruction for process shard block instructions
func InitGenesisShardCommitteeStateV2(env ShardCommitteeStateEnvironment) *ShardCommitteeStateV2 {
	s := NewShardCommitteeStateV2()

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

	s.shardCommittee = incognitokey.DeepCopy(addedCommittees)
	committeeChange.ShardCommitteeAdded[env.ShardID()] = addedCommittees
	return s
}

//UpdateCommitteeState update committeState from valid data before
//	- call process instructions from beacon
//	- check conditions for epoch timestamp
//		+ process shard block instructions for key
//			+ process shard block instructions normally
//	- hash for checking commit later
//	- Only call once in new or insert block process
func (s *ShardCommitteeStateV2) UpdateCommitteeState(
	env ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash, *CommitteeChange, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	committeeChange, err := s.forceUpdateCommitteesFromBeacon(env, NewCommitteeChange())
	if err != nil {
		return nil, NewCommitteeChange(), NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}
	hashes, err := s.hash()
	if err != nil {
		return nil, NewCommitteeChange(), NewCommitteeStateError(ErrUpdateCommitteeState, err)
	}

	return hashes, committeeChange, nil
}

func getNewShardCommittees(
	shardCommittees []string,
) ([]string, error) {
	return shardCommittees, nil
}

// processSwapShardInstruction: process swap shard instruction
func (s *ShardCommitteeStateV2) processSwapShardInstruction(
	swapShardInstruction *instruction.SwapShardInstruction,
	env ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {

	newCommitteeChange := committeeChange
	chainID := byte(swapShardInstruction.ChainID)
	tempSwapOutPublicKeys := swapShardInstruction.OutPublicKeyStructs
	tempSwapInPublicKeys := swapShardInstruction.InPublicKeyStructs
	numberFixedValidators := env.NumberOfFixedBlockValidators()

	// process list shard committees
	for _, v := range tempSwapOutPublicKeys {
		s.shardCommittee = append(s.shardCommittee[:numberFixedValidators], s.shardCommittee[numberFixedValidators+1:]...)
		newCommitteeChange.ShardCommitteeRemoved[chainID] = append(newCommitteeChange.ShardCommitteeRemoved[chainID], v)
	}
	s.shardCommittee = append(s.shardCommittee, tempSwapInPublicKeys...)
	newCommitteeChange.ShardCommitteeAdded[chainID] = append(newCommitteeChange.ShardCommitteeAdded[chainID], tempSwapInPublicKeys...)

	return newCommitteeChange, nil
}

func (s *ShardCommitteeStateV2) forceUpdateCommitteesFromBeacon(
	env ShardCommitteeStateEnvironment,
	committeeChange *CommitteeChange) (*CommitteeChange, error) {
	for _, newShardCommittee := range env.CommitteesFromBeaconView() {
		flag := false
		for _, oldShardCommittee := range s.shardCommittee {
			if reflect.DeepEqual(newShardCommittee, oldShardCommittee) {
				flag = true
				break
			}
		}
		if !flag {
			committeeChange.ShardCommitteeAdded[env.ShardID()] = append(committeeChange.ShardCommitteeAdded[env.ShardID()], newShardCommittee)
		}
	}

	for _, oldShardCommittee := range s.shardCommittee {
		flag := false
		for _, newShardCommittee := range env.CommitteesFromBeaconView() {
			if reflect.DeepEqual(oldShardCommittee, newShardCommittee) {
				flag = true
				break
			}
		}
		if !flag {
			committeeChange.ShardCommitteeRemoved[env.ShardID()] = append(committeeChange.ShardCommitteeRemoved[env.ShardID()], oldShardCommittee)
		}
	}

	s.shardCommittee = incognitokey.DeepCopy(env.CommitteesFromBeaconView())
	s.committeeFromBlock = env.CommitteesFromBlock()
	return committeeChange, nil
}

//hash generate hashes relate to uncommitted committees of struct ShardCommitteeStateV2
//	append committees and subtitutes to struct and hash it
func (s ShardCommitteeStateV2) hash() (*ShardCommitteeStateHash, error) {

	committeesStr, err := incognitokey.CommitteeKeyListToString(s.shardCommittee)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	committeeHash, err := common.GenerateHashFromStringArray(committeesStr)
	if err != nil {
		return nil, fmt.Errorf("Generate Uncommitted Root Hash, error %+v", err)
	}

	substitutesStr, err := incognitokey.CommitteeKeyListToString([]incognitokey.CommitteePublicKey{})
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
		CommitteeFromBlock:  s.committeeFromBlock,
	}, nil
}

func (s ShardCommitteeStateV2) BuildTotalTxsFeeFromTxs(txs []metadata.Transaction) map[common.Hash]uint64 {
	totalTxsFee := make(map[common.Hash]uint64)
	for _, tx := range txs {
		switch tx.GetType() {
		case common.TxNormalType:
			totalTxsFee[common.PRVCoinID] += tx.GetTxFee()
		case common.TxCustomTokenPrivacyType:
			totalTxsFee[common.PRVCoinID] += tx.GetTxFee()
			txCustomPrivacy := tx.(*transaction.TxCustomTokenPrivacy)
			totalTxsFee[*txCustomPrivacy.GetTokenID()] += txCustomPrivacy.GetTxFeeToken()
			Logger.log.Info("[slashing] totalTxsFee[*txCustomPrivacy.GetTokenID()] :", totalTxsFee[*txCustomPrivacy.GetTokenID()])
		default:
			Logger.log.Infof("[reward] Skip building reward for transaction %s \n", tx.Hash().String())
		}
		Logger.log.Info("[slashing] totalTxsFee[common.PRVCoinID]:", totalTxsFee[common.PRVCoinID])
	}
	return totalTxsFee
}

func (s *ShardCommitteeStateV2) GetCommitteesSubsetFromBlock() common.Hash {
	return s.committeesSubsetFromBlock
}
