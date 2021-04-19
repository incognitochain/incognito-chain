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
	SyncingPoolAdded                   map[byte][]incognitokey.CommitteePublicKey
	SyncingPoolRemoved                 map[byte][]incognitokey.CommitteePublicKey
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
	FinishedSyncValidators             map[byte][]string
	SlashingCommittee                  map[byte][]string
}

func (committeeChange *CommitteeChange) AddNextEpochShardCandidateRemoved(nextEpochShardCandidateRemoved []string) *CommitteeChange {
	temp, _ := incognitokey.CommitteeBase58KeyListToStruct(nextEpochShardCandidateRemoved)
	committeeChange.NextEpochShardCandidateRemoved = append(committeeChange.NextEpochShardCandidateRemoved, temp...)
	return committeeChange
}

func (committeeChange *CommitteeChange) AddShardSubstituteAdded(shardID byte, shardSubstituteAdded []string) *CommitteeChange {
	temp, _ := incognitokey.CommitteeBase58KeyListToStruct(shardSubstituteAdded)
	committeeChange.ShardSubstituteAdded[shardID] = append(committeeChange.ShardSubstituteAdded[shardID], temp...)
	return committeeChange
}

func (committeeChange *CommitteeChange) AddSyncingPoolAdded(shardID byte, syncingPoolAdded []string) *CommitteeChange {
	temp, _ := incognitokey.CommitteeBase58KeyListToStruct(syncingPoolAdded)
	committeeChange.SyncingPoolAdded[shardID] = append(committeeChange.SyncingPoolAdded[shardID], temp...)
	return committeeChange
}

func (committeeChange *CommitteeChange) AddSyncingPoolRemoved(shardID byte, syncingPoolRemoved []string) *CommitteeChange {
	temp, _ := incognitokey.CommitteeBase58KeyListToStruct(syncingPoolRemoved)
	committeeChange.SyncingPoolRemoved[shardID] = append(committeeChange.SyncingPoolRemoved[shardID], temp...)
	return committeeChange
}

func (committeeChange *CommitteeChange) AddFinishedSyncValidators(shardID byte, finishedSyncValidators []string) *CommitteeChange {
	committeeChange.FinishedSyncValidators[shardID] = append(committeeChange.FinishedSyncValidators[shardID], finishedSyncValidators...)
	return committeeChange
}

func (committeeChange *CommitteeChange) AddRemovedStaker(removedStaker string) *CommitteeChange {
	committeeChange.RemovedStaker = append(committeeChange.RemovedStaker, removedStaker)
	return committeeChange
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
		FinishedSyncValidators:  make(map[byte][]string),
		SyncingPoolAdded:        make(map[byte][]incognitokey.CommitteePublicKey),
		SyncingPoolRemoved:      make(map[byte][]incognitokey.CommitteePublicKey),
	}
	for i := 0; i < common.MaxShardNumber; i++ {
		shardID := byte(i)
		committeeChange.ShardSubstituteAdded[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardSubstituteRemoved[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardCommitteeAdded[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.ShardCommitteeRemoved[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.SyncingPoolAdded[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.SyncingPoolRemoved[shardID] = []incognitokey.CommitteePublicKey{}
		committeeChange.SlashingCommittee[shardID] = []string{}
		committeeChange.FinishedSyncValidators[shardID] = []string{}
	}
	return committeeChange
}
