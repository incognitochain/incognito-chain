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
	ShardSyncingAdded                  map[byte][]incognitokey.CommitteePublicKey
	ShardSyncingRemoved                map[byte][]incognitokey.CommitteePublicKey
	ShardCommitteeAdded                map[byte][]incognitokey.CommitteePublicKey
	ShardCommitteeRemoved              map[byte][]incognitokey.CommitteePublicKey
	BeaconSubstituteAdded              []incognitokey.CommitteePublicKey
	BeaconSubstituteRemoved            []incognitokey.CommitteePublicKey
	BeaconCommitteeAdded               []incognitokey.CommitteePublicKey
	BeaconCommitteeRemoved             []incognitokey.CommitteePublicKey
	BeaconCommitteeReplaced            [2][]incognitokey.CommitteePublicKey
	ShardCommitteeReplaced             map[byte][2][]incognitokey.CommitteePublicKey
	StopAutoStake                      []string
	RemovedStaker                      []string
	SlashingCommittee                  map[byte][]string
}

//GetStakerKeys ...
func (committeeChange *CommitteeChange) StakerKeys() []incognitokey.CommitteePublicKey {
	return committeeChange.NextEpochShardCandidateAdded
}

func (committeeChange *CommitteeChange) RemovedStakers() []incognitokey.CommitteePublicKey {
	res := []incognitokey.CommitteePublicKey{}
	res, _ = incognitokey.CommitteeBase58KeyListToStruct(committeeChange.RemovedStaker)
	return res
}

func (committeeChange *CommitteeChange) StopAutoStakeKeys() []incognitokey.CommitteePublicKey {
	res := []incognitokey.CommitteePublicKey{}
	res, _ = incognitokey.CommitteeBase58KeyListToStruct(committeeChange.StopAutoStake)
	return res
}

func (committeeChange *CommitteeChange) IsShardCommitteeChange() bool {
	for _, res := range committeeChange.ShardSubstituteAdded {
		if len(res) > 0 {
			return true
		}
	}
	for _, res := range committeeChange.ShardSubstituteRemoved {
		if len(res) > 0 {
			return true
		}
	}
	return false
}

func NewCommitteeChange() *CommitteeChange {
	committeeChange := &CommitteeChange{
		ShardSubstituteAdded:    make(map[byte][]incognitokey.CommitteePublicKey),
		ShardSubstituteRemoved:  make(map[byte][]incognitokey.CommitteePublicKey),
		ShardCommitteeAdded:     make(map[byte][]incognitokey.CommitteePublicKey),
		ShardCommitteeRemoved:   make(map[byte][]incognitokey.CommitteePublicKey),
		ShardCommitteeReplaced:  make(map[byte][2][]incognitokey.CommitteePublicKey),
		BeaconCommitteeReplaced: [2][]incognitokey.CommitteePublicKey{},
		SlashingCommittee:       make(map[byte][]string),
		ShardSyncingAdded:       make(map[byte][]incognitokey.CommitteePublicKey),
		ShardSyncingRemoved:     make(map[byte][]incognitokey.CommitteePublicKey),
	}
	for i := 0; i < common.MaxShardNumber; i++ {
		shardID := byte(i)
		committeeChange.ShardSubstituteAdded[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardSubstituteRemoved[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardCommitteeAdded[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardCommitteeRemoved[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardSyncingAdded[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardSyncingRemoved[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.SlashingCommittee[shardID] = []string{}
	}
	return committeeChange
}
