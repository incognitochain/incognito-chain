package committeestate

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
)

//ShardCommitteeStateHash :
type ShardCommitteeStateHash struct {
	ShardCommitteeHash  common.Hash
	ShardSubstituteHash common.Hash
}

//ShardCommitteeStateEnvironment :
type ShardCommitteeStateEnvironment struct {
	txs                       []metadata.Transaction
	beaconInstructions        [][]string
	newBeaconHeight           uint64
	chainParamEpoch           uint64
	epochBreakPointSwapNewKey []uint64
}

//NewShardCommitteeStateEnvironment : Default constructor of ShardCommitteeStateEnvironment
//Output: pointer of ShardCommitteeStateEnvironment
func NewShardCommitteeStateEnvironment(txs []metadata.Transaction,
	beaconInstructions [][]string,
	newBeaconHeight uint64,
	chainParamEpoch uint64,
	epochBreakPointSwapNewKey []uint64) *ShardCommitteeStateEnvironment {
	return &ShardCommitteeStateEnvironment{
		txs:                       txs,
		newBeaconHeight:           newBeaconHeight,
		chainParamEpoch:           chainParamEpoch,
		epochBreakPointSwapNewKey: epochBreakPointSwapNewKey,
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

	if comparedHashes.ShardCommitteeHash.IsEqual(&hashes.ShardCommitteeHash) {
		return NewCommitteeStateError(ErrCommitShardCommitteeState, fmt.Errorf("Uncommitted ShardCommitteeHash want value %+v but have %+v",
			comparedHashes.ShardCommitteeHash, hashes.ShardCommitteeHash))
	}

	if comparedHashes.ShardSubstituteHash.IsEqual(&hashes.ShardSubstituteHash) {
		return NewCommitteeStateError(ErrCommitShardCommitteeState, fmt.Errorf("Uncommitted ShardSubstituteHash want value %+v but have %+v",
			comparedHashes.ShardSubstituteHash, hashes.ShardSubstituteHash))
	}

	engine.uncommittedShardCommitteeStateV1.clone(engine.shardCommitteeStateV1)
	engine.uncommittedShardCommitteeStateV1.reset()
	return nil
}

//AbortUncommittedBeaconState : Reset data in uncommittedShardCommitteeStateV1 struct
//Pre-conditions: uncommittedShardCommitteeStateV1 has been inited
//Input: NULL
//Output: error
func (engine *ShardCommitteeEngine) AbortUncommittedBeaconState() {
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()
	engine.uncommittedShardCommitteeStateV1.reset()
}

//UpdateCommitteeState : Update committeState from valid data before
//Pre-conditions: Validate committee state
//Input: env variables ShardCommitteeStateEnvironment
//Output: New ShardCommitteeEngineV1 and committee changes, error
func (engine *ShardCommitteeEngine) UpdateCommitteeState(env *ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash, *CommitteeChange, error) {
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()
	engine.shardCommitteeStateV1.mu.RLock()
	engine.shardCommitteeStateV1.clone(engine.uncommittedShardCommitteeStateV1)
	// env.allCandidateSubstituteCommittee = engine.beaconCommitteeStateV1.getAllCandidateSubstituteCommittee()
	engine.shardCommitteeStateV1.mu.RUnlock()
	// newCommitteeState := engine.uncommittedShardCommitteeStateV1
	committeeChange := NewCommitteeChange()
	// newShardCandidates := []incognitokey.CommitteePublicKey{}
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
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()

	committeeState := engine.shardCommitteeStateV1

	newCommittees := []incognitokey.CommitteePublicKey{}
	newPendingValidators := []incognitokey.CommitteePublicKey{}

	committeeState.shardCommittee = append(committeeState.shardCommittee, newCommittees...)
	committeeState.shardPendingValidator = append(committeeState.shardPendingValidator, newPendingValidators...)
}

//GetShardCommittee : Get shard committees
//Input: NULL
//Output: list array of incognito public keys
func (engine *ShardCommitteeEngine) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	engine.shardCommitteeStateV1.mu.RLock()
	defer engine.shardCommitteeStateV1.mu.Unlock()
	return engine.shardCommitteeStateV1.shardCommittee
}

//GetShardPendingValidator : Get shard pending validators
//Input: NULL
//Output: list array of incognito public keys
func (engine *ShardCommitteeEngine) GetShardPendingValidator(shardID byte) []incognitokey.CommitteePublicKey {
	engine.shardCommitteeStateV1.mu.RLock()
	defer engine.shardCommitteeStateV1.mu.Unlock()
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
