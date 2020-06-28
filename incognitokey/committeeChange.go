package incognitokey

import "github.com/incognitochain/incognito-chain/common"

type CommitteeChange struct {
	NextEpochBeaconCandidateAdded      []CommitteePublicKey
	NextEpochBeaconCandidateRemoved    []CommitteePublicKey
	CurrentEpochBeaconCandidateAdded   []CommitteePublicKey
	CurrentEpochBeaconCandidateRemoved []CommitteePublicKey
	NextEpochShardCandidateAdded       []CommitteePublicKey
	NextEpochShardCandidateRemoved     []CommitteePublicKey
	CurrentEpochShardCandidateAdded    []CommitteePublicKey
	CurrentEpochShardCandidateRemoved  []CommitteePublicKey
	ShardSubstituteAdded               map[byte][]CommitteePublicKey
	ShardSubstituteRemoved             map[byte][]CommitteePublicKey
	ShardCommitteeAdded                map[byte][]CommitteePublicKey
	ShardCommitteeRemoved              map[byte][]CommitteePublicKey
	BeaconSubstituteAdded              []CommitteePublicKey
	BeaconSubstituteRemoved            []CommitteePublicKey
	BeaconCommitteeAdded               []CommitteePublicKey
	BeaconCommitteeRemoved             []CommitteePublicKey
	StopAutoStake                      []string
}

func NewCommitteeChange() *CommitteeChange {
	committeeChange := &CommitteeChange{
		ShardSubstituteAdded:   make(map[byte][]CommitteePublicKey),
		ShardSubstituteRemoved: make(map[byte][]CommitteePublicKey),
		ShardCommitteeAdded:    make(map[byte][]CommitteePublicKey),
		ShardCommitteeRemoved:  make(map[byte][]CommitteePublicKey),
	}
	for i := 0; i < common.MaxShardNumber; i++ {
		shardID := byte(i)
		committeeChange.ShardSubstituteAdded[shardID] = []CommitteePublicKey{}
		committeeChange.ShardSubstituteRemoved[shardID] = []CommitteePublicKey{}
		committeeChange.ShardCommitteeAdded[shardID] = []CommitteePublicKey{}
		committeeChange.ShardCommitteeRemoved[shardID] = []CommitteePublicKey{}
	}
	return committeeChange
}
