package committeestate

import (
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
	panic("implement me")
}

//Commit : Commit commitee state change in uncommittedShardCommitteeStateV1 struct
//Pre-conditions: uncommittedShardCommitteeStateV1 has been inited
//Input: ShardCommitteeStateEnvironment for enviroment config
//Output: error
func (engine *ShardCommitteeEngine) Commit(env *ShardCommitteeStateEnvironment) error {
	panic("implement me")
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
func (engine *ShardCommitteeEngine) UpdateCommitteeState(env *ShardCommitteeStateEnvironment) (*ShardCommitteeStateEnvironment, *CommitteeChange, error) {
	panic("implement me")
}

//InitCommitteeState : Init committee state at genesis block or anytime restore program
//Pre-conditions: right config from files or env variables
//Input: env variables ShardCommitteeStateEnvironment
//Output: NULL
func (engine *ShardCommitteeEngine) InitCommitteeState(env *ShardCommitteeStateEnvironment) {
	engine.uncommittedShardCommitteeStateV1.mu.Lock()
	defer engine.uncommittedShardCommitteeStateV1.mu.Unlock()
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
