package committeestate

import (
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
)

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
}

func NewShardCommitteeStateV1() *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{}
}

func NewShardCommitteeStateV1WithValue(shardCommittee map[byte][]incognitokey.CommitteePublicKey, shardPendingValidator map[byte][]incognitokey.CommitteePublicKey) *ShardCommitteeStateV1 {
	return &ShardCommitteeStateV1{ShardCommittee: shardCommittee, ShardPendingValidator: shardPendingValidator}
}
