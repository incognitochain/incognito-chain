package committeestate

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
)

//ShardCommitteeStateV2
type ShardCommitteeStateV2 struct {
	shardCommittee            []string
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
	res, _ := incognitokey.CommitteeKeyListToString(shardCommittee)
	return &ShardCommitteeStateV2{
		shardCommittee:            res,
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
	newCommitteeState.shardCommittee = common.DeepCopyString(s.shardCommittee)
	newCommitteeState.committeeFromBlock = s.committeeFromBlock
	newCommitteeState.committeesSubsetFromBlock = s.committeesSubsetFromBlock
}

//Version ...
func (s *ShardCommitteeStateV2) Version() int {
	return SLASHING_VERSION
}

//GetShardCommittee get shard committees
func (s *ShardCommitteeStateV2) GetShardCommittee() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(s.shardCommittee)
	return res
}

//GetShardSubstitute get shard pending validators
func (s *ShardCommitteeStateV2) GetShardSubstitute() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

func (s *ShardCommitteeStateV2) GetCommitteeFromBlock() common.Hash {
	return s.committeeFromBlock
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
			stakeInstruction, err := instruction.ValidateAndImportStakeInstructionFromString(beaconInstruction)
			if err != nil {
				panic(err)
			}
			candidates = append(candidates, stakeInstruction.PublicKeys...)
		}
	}

	s.shardCommittee = append(s.shardCommittee, candidates[int(env.ShardID())*
		env.MinShardCommitteeSize():(int(env.ShardID())*env.MinShardCommitteeSize())+env.MinShardCommitteeSize()]...)

	addedCommittees, _ := incognitokey.CommitteeBase58KeyListToStruct(s.shardCommittee)
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
			newShardCommitteeStruct, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{newShardCommittee})
			committeeChange.ShardCommitteeAdded[env.ShardID()] = append(committeeChange.ShardCommitteeAdded[env.ShardID()], newShardCommitteeStruct[0])
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
			oldShardCommitteeStruct, _ := incognitokey.CommitteeBase58KeyListToStruct([]string{oldShardCommittee})
			committeeChange.ShardCommitteeRemoved[env.ShardID()] = append(committeeChange.ShardCommitteeRemoved[env.ShardID()], oldShardCommitteeStruct[0])
		}
	}

	s.shardCommittee = common.DeepCopyString(env.CommitteesFromBeaconView())
	s.committeeFromBlock = env.CommitteesFromBlock()
	return committeeChange, nil
}

//hash generate hashes relate to uncommitted committees of struct ShardCommitteeStateV2
//	append committees and subtitutes to struct and hash it
func (s ShardCommitteeStateV2) hash() (*ShardCommitteeStateHash, error) {

	committeeHash, err := common.GenerateHashFromStringArray(s.shardCommittee)
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
