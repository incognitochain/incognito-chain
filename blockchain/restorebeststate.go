package blockchain

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

//RecoverCommittee ...
func (shardBestState *ShardBestState) RestoreCommittee(shardID byte, chain *BlockChain) error {

	committeePublicKey := statedb.GetOneShardCommittee(shardBestState.consensusStateDB, shardID)

	shardBestState.ShardCommittee = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		shardBestState.ShardCommittee[i] = v
	}

	return nil
}

//RestorePendingValidators ...
func (shardBestState *ShardBestState) RestorePendingValidators(shardID byte, bc *BlockChain) error {

	committeePublicKey := statedb.GetOneShardSubstituteValidator(shardBestState.consensusStateDB, shardID)
	shardBestState.ShardPendingValidator = make([]incognitokey.CommitteePublicKey, len(committeePublicKey))
	for i, v := range committeePublicKey {
		shardBestState.ShardPendingValidator[i] = v
	}
	return nil
}

// //
// func (shardBestState *ShardBestState) restoreViewFromHash(blockchain *BlockChain) error {
// 	return nil
// }

//RestoreBeaconViewStateFromHash ...
func (beaconBestState *BeaconBestState) RestoreBeaconViewStateFromHash(blockchain *BlockChain) error {
	err := beaconBestState.InitStateRootHash(blockchain)
	if err != nil {
		return err
	}
	//best block
	block, _, err := blockchain.GetBeaconBlockByHash(beaconBestState.BestBlockHash)
	if err != nil || block == nil {
		return err
	}
	beaconBestState.BestBlock = *block
	beaconBestState.BeaconHeight = block.GetHeight()
	beaconCommitteeEngine := InitBeaconCommitteeEngineV1(beaconBestState.ActiveShards, beaconBestState.consensusStateDB, beaconBestState.BeaconHeight, beaconBestState.BestBlockHash)
	beaconBestState.beaconCommitteeEngine = beaconCommitteeEngine
	return nil
}
