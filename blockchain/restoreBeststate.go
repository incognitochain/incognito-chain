package blockchain

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/beststate"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

//restoreBeaconCommittee ...
func (beaconBestState *BeaconBestState) restoreBeaconCommittee() error {
	committeePublicKey := statedb.GetBeaconCommittee(beaconBestState.consensusStateDB)
	beaconBestState.BeaconCommittee = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		beaconBestState.BeaconCommittee[i] = v
	}

	// fmt.Println("[optimize-beststate] {BeaconBestState.restoreBeaconCommittee()} len(beaconBestState.BeaconCommittee):", len(beaconBestState.BeaconCommittee))

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

	// fmt.Println("[optimize-beststate] {BeaconBestState.restoreShardCommittee()} len(beaconBestState.ShardCommittee):", len(beaconBestState.ShardCommittee))

	return nil
}

//restoreBeaconPreCommitteeInfo ...
func (beaconBestState *BeaconBestState) restoreBeaconPreCommitteeInfo(bc *BlockChain) error {

	beaconPreCommitteeInfo := beststate.BeaconPreCommitteeInfo{
		BeaconPendingValidator:                 []incognitokey.CommitteePublicKey{},
		CandidateBeaconWaitingForCurrentRandom: []incognitokey.CommitteePublicKey{},
		CandidateBeaconWaitingForNextRandom:    []incognitokey.CommitteePublicKey{},
	}

	if beaconBestState.BeaconHeight <= 2 {
		return nil
	}

	//Restore beacon pending validators
	beaconPreCommitteeInfoData, err := rawdbv2.GetBeaconPreCommitteeInfo(bc.GetBeaconChainDatabase(), beaconBestState.BeaconPreCommitteeHash)
	if err != nil {
		return err
	}

	err = json.Unmarshal(beaconPreCommitteeInfoData, beaconPreCommitteeInfo)
	if err != nil {
		return err
	}

	beaconBestState.BeaconPendingValidator = make([]incognitokey.CommitteePublicKey, len(beaconPreCommitteeInfo.BeaconPendingValidator))
	for i, v := range beaconPreCommitteeInfo.BeaconPendingValidator {
		beaconBestState.BeaconPendingValidator[i] = v
	}

	beaconBestState.CandidateBeaconWaitingForCurrentRandom = make([]incognitokey.CommitteePublicKey, len(beaconPreCommitteeInfo.CandidateBeaconWaitingForCurrentRandom))
	for i, v := range beaconPreCommitteeInfo.CandidateBeaconWaitingForCurrentRandom {
		beaconBestState.CandidateBeaconWaitingForCurrentRandom[i] = v
	}

	beaconBestState.CandidateBeaconWaitingForNextRandom = make([]incognitokey.CommitteePublicKey, len(beaconPreCommitteeInfo.CandidateBeaconWaitingForNextRandom))
	for i, v := range beaconPreCommitteeInfo.CandidateBeaconWaitingForNextRandom {
		beaconBestState.CandidateBeaconWaitingForNextRandom[i] = v
	}

	return nil
}

//restoreShardPreCommitteeInfo ...
func (beaconBestState *BeaconBestState) restoreShardPreCommitteeInfo(bc *BlockChain) error {

	shardPreCommitteeInfo := beststate.ShardPreCommitteeInfo{
		ShardPendingValidatorHashs:            make(map[byte]common.Hash),
		CandidateShardWaitingForCurrentRandom: []incognitokey.CommitteePublicKey{},
		CandidateShardWaitingForNextRandom:    []incognitokey.CommitteePublicKey{},
	}

	//Restore beacon pending validators
	shardPreCommitteeInfoData, err := rawdbv2.GetShardPreCommitteeInfo(bc.GetBeaconChainDatabase(), beaconBestState.ShardPreCommitteeHash)
	if err != nil {
		return err
	}

	err = json.Unmarshal(shardPreCommitteeInfoData, shardPreCommitteeInfo)

	if err != nil {
		return err
	}

	beaconBestState.ShardPendingValidator = make(map[byte][]incognitokey.CommitteePublicKey)
	for shardID, v := range shardPreCommitteeInfo.ShardPendingValidatorHashs {

		pendingValidatorsData, err := rawdbv2.GetShardPendingValidators(bc.GetShardChainDatabase(shardID), v)
		if err != nil {
			return err
		}

		pendingValidators := []incognitokey.CommitteePublicKey{}
		err = json.Unmarshal(pendingValidatorsData, pendingValidators)
		if err != nil {
			return err
		}

		beaconBestState.ShardPendingValidator[shardID] = make([]incognitokey.CommitteePublicKey, len(pendingValidators))

		for index, value := range pendingValidators {
			beaconBestState.ShardPendingValidator[shardID][index] = value
		}
	}

	beaconBestState.CandidateShardWaitingForCurrentRandom = make([]incognitokey.CommitteePublicKey, len(shardPreCommitteeInfo.CandidateShardWaitingForCurrentRandom))
	for i, v := range shardPreCommitteeInfo.CandidateShardWaitingForCurrentRandom {
		beaconBestState.CandidateShardWaitingForCurrentRandom[i] = v
	}

	beaconBestState.CandidateShardWaitingForNextRandom = make([]incognitokey.CommitteePublicKey, len(shardPreCommitteeInfo.CandidateShardWaitingForNextRandom))
	for i, v := range shardPreCommitteeInfo.CandidateShardWaitingForNextRandom {
		beaconBestState.CandidateShardWaitingForNextRandom[i] = v
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

	// fmt.Println("[optimize-beststate] {ShardBestState.restoreCommittee()} len(shardBestState.ShardCommittee):", len(shardBestState.ShardCommittee))

	return nil
}

//restorePendingValidators ...
func (shardBestState *ShardBestState) restorePendingValidators(shardID byte, bc *BlockChain) error {

	pendingValidatorsData, err := rawdbv2.GetShardPendingValidators(bc.GetShardChainDatabase(shardID), shardBestState.ShardPendingValidatorHash)
	if err != nil {
		return err
	}

	pendingValidators := []incognitokey.CommitteePublicKey{}

	err = json.Unmarshal(pendingValidatorsData, pendingValidators)
	if err != nil {
		return err
	}

	shardBestState.ShardPendingValidator = make([]incognitokey.CommitteePublicKey, len(pendingValidators))
	for i, v := range pendingValidators {
		shardBestState.ShardPendingValidator[i] = v
	}

	return nil
}

//storeBeaconPreCommitteeHash ...
func (beaconBestState *BeaconBestState) storeBeaconPreCommitteeHash(db incdb.KeyValueWriter, bc *BlockChain) error {
	/// Use temp struct for storing BeaconPreCommitteeInfo
	beaconPreCommitteeInfo := beststate.BeaconPreCommitteeInfo{}

	beaconPreCommitteeInfo.BeaconPendingValidator = make([]incognitokey.CommitteePublicKey, len(beaconBestState.BeaconPendingValidator))
	for i, v := range beaconBestState.BeaconPendingValidator {
		beaconPreCommitteeInfo.BeaconPendingValidator[i] = v
	}

	beaconPreCommitteeInfo.CandidateBeaconWaitingForCurrentRandom = make([]incognitokey.CommitteePublicKey, len(beaconBestState.CandidateBeaconWaitingForCurrentRandom))
	for i, v := range beaconBestState.CandidateShardWaitingForCurrentRandom {
		beaconPreCommitteeInfo.CandidateBeaconWaitingForCurrentRandom[i] = v
	}

	beaconPreCommitteeInfo.CandidateBeaconWaitingForNextRandom = make([]incognitokey.CommitteePublicKey, len(beaconBestState.CandidateBeaconWaitingForNextRandom))
	for i, v := range beaconBestState.CandidateBeaconWaitingForNextRandom {
		beaconPreCommitteeInfo.CandidateBeaconWaitingForNextRandom[i] = v
	}

	// Get BeaconPreCommitteeInfo Bytes
	bytes, err := beaconPreCommitteeInfo.MarshalJSON()
	if err != nil {
		return err
	}

	// Hash BeaconPreCommitteeInfo
	hash := common.BytesToHash(bytes)
	// Add to BeaconPreCommitteeHash to BeaconBestState
	beaconBestState.BeaconPreCommitteeHash = hash

	// Save and check cache value here
	if _, ok := bc.BeaconChain.hashHistory.Get(hash.String()); !ok {
		bc.BeaconChain.hashHistory.Add(hash.String(), true)
		err := rawdbv2.StoreBeaconPreCommitteeInfo(db, hash, bytes)
		if err != nil {
			return err
		}
	}
	/// End of save and check cache value

	return nil
}

//storewShardPreCommitteeHash ...
func (beaconBestState *BeaconBestState) storeShardPreCommitteeHash(db incdb.KeyValueWriter, bc *BlockChain) error {
	/// Use temp struct for storing ShardPreCommitteeInfo
	shardPreCommitteeInfo := beststate.ShardPreCommitteeInfo{
		ShardPendingValidatorHashs: make(map[byte]common.Hash),
	}

	for shardID, pendingValidators := range beaconBestState.ShardPendingValidator {
		pendingValidatorsData, err := json.Marshal(pendingValidators)
		if err != nil {
			return err
		}
		hash := common.BytesToHash(pendingValidatorsData)
		shardPreCommitteeInfo.ShardPendingValidatorHashs[shardID] = hash
	}

	shardPreCommitteeInfo.CandidateShardWaitingForCurrentRandom = make([]incognitokey.CommitteePublicKey, len(beaconBestState.CandidateShardWaitingForCurrentRandom))
	for i, v := range beaconBestState.CandidateShardWaitingForCurrentRandom {
		shardPreCommitteeInfo.CandidateShardWaitingForCurrentRandom[i] = v
	}

	shardPreCommitteeInfo.CandidateShardWaitingForNextRandom = make([]incognitokey.CommitteePublicKey, len(beaconBestState.CandidateShardWaitingForNextRandom))
	for i, v := range beaconBestState.CandidateBeaconWaitingForNextRandom {
		shardPreCommitteeInfo.CandidateShardWaitingForNextRandom[i] = v
	}

	// Get ShardPreCommitteeInfo Bytes
	bytes, err := shardPreCommitteeInfo.MarshalJSON()
	if err != nil {
		return err
	}

	// Hash BeaconPreCommitteeInfo
	hash := common.BytesToHash(bytes)
	// Add to BeaconPreCommitteeHash to BeaconBestState
	beaconBestState.BeaconPreCommitteeHash = hash

	// Save and check cache value here
	if _, ok := bc.BeaconChain.hashHistory.Get(hash.String()); !ok {
		bc.BeaconChain.hashHistory.Add(hash.String(), true)
		err := rawdbv2.StoreShardPreCommitteeInfo(db, hash, bytes)
		if err != nil {
			return err
		}
	}
	/// End of save and check cache value

	return nil
}

//storePendingValidators ...
func (shardBestState *ShardBestState) storePendingValidators(db incdb.KeyValueWriter, bc *BlockChain) error {

	pendingValidatorsData, err := json.Marshal(shardBestState.ShardPendingValidator)
	if err != nil {
		return err
	}
	hash := common.BytesToHash(pendingValidatorsData)
	shardBestState.ShardPendingValidatorHash = hash

	// Save and check cache value here
	if _, ok := bc.ShardChain[shardBestState.ShardID].hashHistory.Get(hash.String()); !ok {
		bc.BeaconChain.hashHistory.Add(hash.String(), true)
		err := rawdbv2.StoreShardPendingValidators(db, hash, hash.Bytes())
		if err != nil {
			return err
		}
	}
	/// End of save and check cache value

	return nil
}
