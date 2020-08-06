package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type CommitteeChange struct {
	NextEpochBeaconCandidateAdded      []incognitokey.CommitteePublicKey
	NextEpochBeaconCandidateRemoved    []incognitokey.CommitteePublicKey
	CurrentEpochBeaconCandidateAdded   []incognitokey.CommitteePublicKey
	CurrentEpochBeaconCandidateRemoved []incognitokey.CommitteePublicKey
	NextEpochShardCandidateAdded       []incognitokey.CommitteePublicKey
	NextEpochShardCandidateRemoved     []incognitokey.CommitteePublicKey
	CurrentEpochShardCandidateAdded    []incognitokey.CommitteePublicKey
	CurrentEpochShardCandidateRemoved  []incognitokey.CommitteePublicKey
	ShardSubstituteAdded               map[byte][]incognitokey.CommitteePublicKey
	ShardSubstituteRemoved             map[byte][]incognitokey.CommitteePublicKey
	ShardCommitteeAdded                map[byte][]incognitokey.CommitteePublicKey
	ShardCommitteeRemoved              map[byte][]incognitokey.CommitteePublicKey
	BeaconSubstituteAdded              []incognitokey.CommitteePublicKey
	BeaconSubstituteRemoved            []incognitokey.CommitteePublicKey
	BeaconCommitteeAdded               []incognitokey.CommitteePublicKey
	BeaconCommitteeRemoved             []incognitokey.CommitteePublicKey
	BeaconCommitteeReplaced            [2][]incognitokey.CommitteePublicKey
	ShardCommitteeReplaced             map[byte][2][]incognitokey.CommitteePublicKey
	StopAutoStake                      []string
}

func NewCommitteeChange() *CommitteeChange {
	committeeChange := &CommitteeChange{
		ShardSubstituteAdded:    make(map[byte][]incognitokey.CommitteePublicKey),
		ShardSubstituteRemoved:  make(map[byte][]incognitokey.CommitteePublicKey),
		ShardCommitteeAdded:     make(map[byte][]incognitokey.CommitteePublicKey),
		ShardCommitteeRemoved:   make(map[byte][]incognitokey.CommitteePublicKey),
		ShardCommitteeReplaced:  make(map[byte][2][]incognitokey.CommitteePublicKey),
		BeaconCommitteeReplaced: [2][]incognitokey.CommitteePublicKey{},
	}
	for i := 0; i < common.MaxShardNumber; i++ {
		shardID := byte(i)
		committeeChange.ShardSubstituteAdded[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardSubstituteRemoved[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardCommitteeAdded[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardCommitteeRemoved[shardID] = []incognitokey.CommitteePublicKey{}
	}
	return committeeChange
}
