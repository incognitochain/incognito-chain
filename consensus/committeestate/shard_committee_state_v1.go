package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ShardCommitteeStateV1 struct {
	ShardCommittee        map[byte][]incognitokey.CommitteePublicKey
	ShardPendingValidator map[byte][]incognitokey.CommitteePublicKey
}

func NewShardCommitteeStateV1() *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{}
}

func NewShardCommitteeStateV1WithValue(shardCommittee map[byte][]incognitokey.CommitteePublicKey, shardPendingValidator map[byte][]incognitokey.CommitteePublicKey) *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{ShardCommittee: shardCommittee, ShardPendingValidator: shardPendingValidator}
}

func (s ShardCommitteeStateV1) GenerateCommitteeRootHashes(shardID byte, instruction []string) ([]common.Hash, error) {
	panic("implement me")
}

func (s ShardCommitteeStateV1) UpdateCommitteeState(shardID byte, instruction []string) (interface{}, error) {
	panic("implement me")
}

func (s ShardCommitteeStateV1) ValidateCommitteeRootHashes(shardID byte, rootHashes []common.Hash) (bool, error) {
	panic("implement me")
}

func (s ShardCommitteeStateV1) GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {
	panic("implement me")
}

func (s ShardCommitteeStateV1) GetShardPendingValidator(shardID byte) []incognitokey.CommitteePublicKey {
	panic("implement me")
}
