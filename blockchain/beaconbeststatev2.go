package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type committeeChange struct {
	nextEpochBeaconCandidateAdded      []incognitokey.CommitteePublicKey
	nextEpochBeaconCandidateRemoved    []incognitokey.CommitteePublicKey
	currentEpochBeaconCandidateAdded   []incognitokey.CommitteePublicKey
	currentEpochBeaconCandidateRemoved []incognitokey.CommitteePublicKey
	nextEpochShardCandidateAdded       []incognitokey.CommitteePublicKey
	nextEpochShardCandidateRemoved     []incognitokey.CommitteePublicKey
	currentEpochShardCandidateAdded    []incognitokey.CommitteePublicKey
	currentEpochShardCandidateRemoved  []incognitokey.CommitteePublicKey
	shardSubstituteAdded               map[byte][]incognitokey.CommitteePublicKey
	shardSubstituteRemoved             map[byte][]incognitokey.CommitteePublicKey
	shardCommitteeAdded                map[byte][]incognitokey.CommitteePublicKey
	shardCommitteeRemoved              map[byte][]incognitokey.CommitteePublicKey
	beaconSubstituteAdded              []incognitokey.CommitteePublicKey
	beaconSubstituteRemoved            []incognitokey.CommitteePublicKey
	beaconCommitteeAdded               []incognitokey.CommitteePublicKey
	beaconCommitteeRemoved             []incognitokey.CommitteePublicKey
}

func newCommitteeChange() *committeeChange {
	return &committeeChange{
		shardSubstituteAdded:   make(map[byte][]incognitokey.CommitteePublicKey),
		shardSubstituteRemoved: make(map[byte][]incognitokey.CommitteePublicKey),
		shardCommitteeAdded:    make(map[byte][]incognitokey.CommitteePublicKey),
		shardCommitteeRemoved:  make(map[byte][]incognitokey.CommitteePublicKey),
	}
}

func (beaconBestState *BeaconBestState) GetConsensusStateRootHash(height uint64) common.Hash {
	beaconBestState.lock.RLock()
	defer beaconBestState.lock.RUnlock()
	return beaconBestState.ConsensusStateRootHash[height]
}

func (beaconBestState *BeaconBestState) GetFeatureStateRootHash(height uint64) common.Hash {
	beaconBestState.lock.RLock()
	defer beaconBestState.lock.RUnlock()
	return beaconBestState.FeatureStateRootHash[height]
}
