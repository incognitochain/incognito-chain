package blockchain

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

//restoreBeaconCommittee ...
func (beaconBestState *BeaconBestState) restoreBeaconCommittee() error {

	beaconBestState.BeaconCommittee = []incognitokey.CommitteePublicKey{}

	committeePublicKey := statedb.GetBeaconCommittee(beaconBestState.consensusStateDB)
	for _, v := range committeePublicKey {
		beaconBestState.BeaconCommittee = append(beaconBestState.BeaconCommittee, v)
	}

	return nil
}

//restoreShardCommittee ...
func (beaconBestState *BeaconBestState) restoreShardCommittee() error {

	beaconBestState.BeaconCommittee = []incognitokey.CommitteePublicKey{}

	committeePublicKey := statedb.GetBeaconCommittee(beaconBestState.consensusStateDB)
	for _, v := range committeePublicKey {
		beaconBestState.BeaconCommittee = append(beaconBestState.BeaconCommittee, v)
	}

	return nil
}

//RecoverCommittee ...
func (shardBestState *ShardBestState) restoreCommittee(shardID byte) error {

	shardBestState.ShardCommittee = []incognitokey.CommitteePublicKey{}

	committeePublicKey := statedb.GetOneShardCommittee(shardBestState.consensusStateDB, shardID)

	for _, v := range committeePublicKey {
		shardBestState.ShardCommittee = append(shardBestState.ShardCommittee, v)
	}

	// fmt.Println("[optimize-bestsate] {ShardBestState.restoreCommittee()} len(shardBestState.ShardCommittee):", len(shardBestState.ShardCommittee))
	// fmt.Println("[optimize-bestsate] {ShardBestState.restoreCommittee()} len(committeePublicKey):", len(committeePublicKey))

	return nil
}
