package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

//restoreBeaconCommittee ...
func (beaconBestState *BeaconBestState) restoreBeaconCommittee() error {

	committeePublicKey := statedb.GetBeaconCommittee(beaconBestState.consensusStateDB)
	beaconBestState.BeaconCommittee = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		beaconBestState.BeaconCommittee[i] = v
	}

	return nil
}

//restoreShardCommittee ...
func (beaconBestState *BeaconBestState) restoreShardCommittee() error {

	beaconBestState.ShardCommittee = make(map[byte][]incognitokey.CommitteePublicKey)
	for i := 0; i < beaconBestState.ActiveShards; i++ {
		committeePublicKey := statedb.GetOneShardCommittee(beaconBestState.consensusStateDB, byte(i))
		beaconBestState.ShardCommittee[byte(i)] = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
		for index, value := range committeePublicKey {
			beaconBestState.ShardCommittee[byte(i)][index] = value
		}
	}

	return nil
}

//restoreBeaconPendingValidator ...
func (beaconBestState *BeaconBestState) restoreBeaconPendingValidator() error {

	committeePublicKey := statedb.GetBeaconSubstituteValidator(beaconBestState.consensusStateDB)
	beaconBestState.BeaconPendingValidator = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		beaconBestState.BeaconPendingValidator[i] = v
	}
	return nil
}

//restoreShardPendingValidator ...
func (beaconBestState *BeaconBestState) restoreShardPendingValidator() error {

	beaconBestState.ShardPendingValidator = make(map[byte][]incognitokey.CommitteePublicKey)
	for i := 0; i < beaconBestState.ActiveShards; i++{
		committeePublicKey := statedb.GetOneShardSubstituteValidator(beaconBestState.consensusStateDB, byte(i))
		beaconBestState.ShardPendingValidator[byte(i)] = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
		for index, value := range committeePublicKey{
			beaconBestState.ShardPendingValidator[byte(i)][index] = value
		}
	}
	return nil
}

//restoreCandidateShardWaitingForCurrentRandom ...
func (beaconBestState *BeaconBestState) restoreCandidateShardWaitingForCurrentRandom() error {

	//GetCurrentEpochCandidate
	committeePublicKey := statedb.GetCurrentEpochCandidate(beaconBestState.consensusStateDB)
	beaconBestState.CandidateShardWaitingForCurrentRandom = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		beaconBestState.CandidateShardWaitingForCurrentRandom[i] = v
	}
	return nil
}

//restoreCandidateBeaconWaitingForCurrentRandom ...
func (beaconBestState *BeaconBestState) restoreCandidateBeaconWaitingForCurrentRandom() error {

	//TODO: @tin
	// For further development, when beacon is round robin for community -> change here

	beaconBestState.CandidateBeaconWaitingForCurrentRandom = make([]incognitokey.CommitteePublicKey, 0)
	return nil
}

//restoreCandidateShardWaitingForNextRandom ...
func (beaconBestState *BeaconBestState) restoreCandidateShardWaitingForNextRandom() error {
	//GetNextEpochCandidate
	committeePublicKey := statedb.GetNextEpochCandidate(beaconBestState.consensusStateDB)
	beaconBestState.CandidateShardWaitingForNextRandom = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey{
		beaconBestState.CandidateShardWaitingForNextRandom[i] = v
	}
	return nil
}

//restoreCandidateBeaconWaitingForNextRandom ...
func (beaconBestState *BeaconBestState) restoreCandidateBeaconWaitingForNextRandom() error {

	//TODO: @tin
	// For further development, when beacon is round robin for community -> change here

	beaconBestState.CandidateBeaconWaitingForNextRandom = make([]incognitokey.CommitteePublicKey, 0)
	return nil
}

//RecoverCommittee ...
func (shardBestState *ShardBestState) restoreCommittee(shardID byte) error {

	committeePublicKey := statedb.GetOneShardCommittee(shardBestState.consensusStateDB, shardID)

	fmt.Println("[optimize-beststate] len(committeePublicKey):", len(committeePublicKey))
	for _, v := range committeePublicKey {
		key, _ := v.ToBase58()
		fmt.Println("[optimize-beststate] key:", key)
	}

	shardBestState.ShardCommittee = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		shardBestState.ShardCommittee[i] = v
	}

	return nil
}

//restorePendingValidators ...
func (shardBestState *ShardBestState) restorePendingValidators(shardID byte, bc *BlockChain) error {

	committeePublicKey := statedb.GetOneShardSubstituteValidator(shardBestState.consensusStateDB, shardID)
	shardBestState.ShardPendingValidator = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		shardBestState.ShardPendingValidator[i] = v
	}
	return nil
}