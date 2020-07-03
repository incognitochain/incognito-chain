package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"sync"
)

type ShardCommitteeStateHash struct {
	ShardCommitteeHash  common.Hash
	ShardSubstituteHash common.Hash
}

type ShardCommitteeStateEnvironment struct {
	txs                       []metadata.Transaction
	beaconInstructions        [][]string
	newBeaconHeight           uint64
	chainParamEpoch           uint64
	epochBreakPointSwapNewKey []uint64
}

type ShardCommitteeStateV1 struct {
	ShardCommittee        map[byte][]incognitokey.CommitteePublicKey
	ShardPendingValidator map[byte][]incognitokey.CommitteePublicKey

	mu *sync.RWMutex
}

type ShardCommitteeEngineV1 struct {
	shardHeight                      uint64
	shardHash                        common.Hash
	shardID                          byte
	shardCommitteeStateV1            *ShardCommitteeStateV1
	uncommittedShardCommitteeStateV1 *ShardCommitteeStateV1
}

func NewShardCommitteeStateV1() *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{
		mu: new(sync.RWMutex),
	}
}

func NewShardCommitteeStateV1WithValue(shardCommittee map[byte][]incognitokey.CommitteePublicKey, shardPendingValidator map[byte][]incognitokey.CommitteePublicKey) *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{
		ShardCommittee:        shardCommittee,
		ShardPendingValidator: shardPendingValidator,
		mu:                    new(sync.RWMutex),
	}
}

func (s ShardCommitteeEngineV1) Commit(env *ShardCommitteeStateEnvironment) error {
	panic("implement me")
}

func (s ShardCommitteeEngineV1) AbortUncommittedBeaconState() {
	panic("implement me")
}

func (s ShardCommitteeEngineV1) UpdateCommitteeState(env *ShardCommitteeStateEnvironment) (*ShardCommitteeStateEnvironment, *CommitteeChange, error) {
	panic("implement me")
}

func (s ShardCommitteeEngineV1) InitCommitteeState(env *ShardCommitteeStateEnvironment) {
	panic("implement me")
}

func (s ShardCommitteeEngineV1) ValidateCommitteeRootHashes(rootHashes []common.Hash) (bool, error) {
	panic("implement me")
}

func (s ShardCommitteeEngineV1) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (s ShardCommitteeEngineV1) GetShardPendingValidator(shardID byte) []incognitokey.CommitteePublicKey {
	panic("implement me")
}
