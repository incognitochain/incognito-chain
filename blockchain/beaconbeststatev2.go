package blockchain

import (
	"github.com/incognitochain/incognito-chain/common"
)

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
