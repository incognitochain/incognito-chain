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

// //initRawDBHash ...
// func (beaconBestState *BeaconBestState) initRawDBHash(bc *BlockChain) error {

// 	// get beaconBestState.BeaconPreCommitteeHash from epoch
// 	beaconBestState.BeaconPreCommitteeHash = common.Hash{}

// 	// get beaconBestState.ShardPreCommitteeHash from epoch
// 	beaconBestState.ShardPreCommitteeHash = common.Hash{}

// 	return nil
// }

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

	// shardPreCommitteeInfo := beststate.ShardPreCommitteeInfo{
	// 	// ShardPendingValidator:                 []incognitokey.CommitteePublicKey{},
	// 	CandidateShardWaitingForCurrentRandom: []incognitokey.CommitteePublicKey{},
	// 	CandidateShardWaitingForNextRandom:    []incognitokey.CommitteePublicKey{},
	// }

	// //Restore beacon pending validators
	// beaconPreCommitteeInfoData, err := rawdbv2.GetShardPreCommitteeInfo(bc.GetShardChainDatabase(), beaconBestState.BeaconPreCommitteeHash)
	// if err != nil {
	// 	return err
	// }

	// err = json.Unmarshal(beaconPreCommitteeInfoData, beaconPreCommitteeInfo)

	// if err != nil {
	// 	return err
	// }

	// beaconBestState.BeaconPendingValidator = make([]incognitokey.CommitteePublicKey, len(beaconPreCommitteeInfo.BeaconPendingValidator))
	// for i, v := range beaconPreCommitteeInfo.BeaconPendingValidator {
	// 	beaconBestState.BeaconPendingValidator[i] = v
	// }

	// beaconBestState.CandidateBeaconWaitingForCurrentRandom = make([]incognitokey.CommitteePublicKey, len(beaconPreCommitteeInfo.CandidateBeaconWaitingForCurrentRandom))
	// for i, v := range beaconPreCommitteeInfo.CandidateBeaconWaitingForCurrentRandom {
	// 	beaconBestState.CandidateBeaconWaitingForCurrentRandom[i] = v
	// }

	// // beaconBestState.CandidateBeaconWaitingForNextRandom = make([]incognitokey.CommitteePublicKey, len(beaconPreCommitteeInfo.CandidateBeaconWaitingForNextRandom))
	// // for i, v := range beaconPreCommitteeInfo.CandidateBeaconWaitingForNextRandom {
	// // 	beaconBestState.CandidateBeaconWaitingForNextRandom[i] = v
	// // }

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

//storeBeaconPreCommitteeHash ...
func (beaconBestState *BeaconBestState) storeShardPreCommitteeHash(db incdb.KeyValueWriter, bc *BlockChain) error {
	return nil
}
