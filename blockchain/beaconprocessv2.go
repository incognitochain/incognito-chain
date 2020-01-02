package blockchain

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (blockchain *BlockChain) processStoreBeaconBlockV2(beaconBlock *BeaconBlock, snapshotBeaconCommittees []incognitokey.CommitteePublicKey, snapshotAllShardCommittees map[byte][]incognitokey.CommitteePublicKey, snapshotRewardReceivers map[string]string,
) error {
	Logger.log.Infof("BEACON | Process Store Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, beaconBlock.Header.Hash())
	blockHash := beaconBlock.Header.Hash()
	//statedb===========================START
	var err error
	err = statedb.StoreCurrentEpochCandidate(beaconBestState.consensusStateDB, beaconBestState.CandidateShardWaitingForCurrentRandom, beaconBestState.RewardReceiver, beaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreNextEpochCandidate(beaconBestState.consensusStateDB, beaconBestState.CandidateShardWaitingForNextRandom, beaconBestState.RewardReceiver, beaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreAllShardSubstitutesValidator(beaconBestState.consensusStateDB, beaconBestState.ShardPendingValidator, beaconBestState.RewardReceiver, beaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreAllShardCommittee(beaconBestState.consensusStateDB, beaconBestState.ShardPendingValidator, beaconBestState.RewardReceiver, beaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreBeaconSubstituteValidator(beaconBestState.consensusStateDB, beaconBestState.BeaconPendingValidator, beaconBestState.RewardReceiver, beaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreBeaconCommittee(beaconBestState.consensusStateDB, beaconBestState.BeaconCommittee, beaconBestState.RewardReceiver, beaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	//statedb===========================END
	//================================Store cross shard state ==================================
	if beaconBlock.Body.ShardState != nil {
		GetBeaconBestState().lock.Lock()
		lastCrossShardState := GetBeaconBestState().LastCrossShardState
		for fromShard, shardBlocks := range beaconBlock.Body.ShardState {
			for _, shardBlock := range shardBlocks {
				for _, toShard := range shardBlock.CrossShard {
					if fromShard == toShard {
						continue
					}
					if lastCrossShardState[fromShard] == nil {
						lastCrossShardState[fromShard] = make(map[byte]uint64)
					}
					lastHeight := lastCrossShardState[fromShard][toShard] // get last cross shard height from shardID  to crossShardShardID
					waitHeight := shardBlock.Height
					err := rawdbv2.StoreCrossShardNextHeight(blockchain.GetDatabase(), fromShard, toShard, lastHeight, waitHeight)
					if err != nil {
						GetBeaconBestState().lock.Unlock()
						return NewBlockChainError(StoreCrossShardNextHeightError, err)
					}
					//beacon process shard_to_beacon in order so cross shard next height also will be saved in order
					//dont care overwrite this value
					err = rawdbv2.StoreCrossShardNextHeight(blockchain.GetDatabase(), fromShard, toShard, waitHeight, 0)
					if err != nil {
						GetBeaconBestState().lock.Unlock()
						return NewBlockChainError(StoreCrossShardNextHeightError, err)
					}
					if lastCrossShardState[fromShard] == nil {
						lastCrossShardState[fromShard] = make(map[byte]uint64)
					}
					lastCrossShardState[fromShard][toShard] = waitHeight //update lastHeight to waitHeight
				}
			}
			blockchain.config.CrossShardPool[fromShard].UpdatePool()
		}
		GetBeaconBestState().lock.Unlock()
	}
	//=============================END Store cross shard state ==================================
	if err := rawdbv2.StoreBeaconBlockIndex(blockchain.GetDatabase(), blockHash, beaconBlock.Header.Height); err != nil {
		return NewBlockChainError(StoreBeaconBlockIndexError, err)
	}
	Logger.log.Debugf("Store Beacon BestState Height %+v", beaconBlock.Header.Height)
	if err := rawdbv2.StoreBeaconBestState(blockchain.GetDatabase(), beaconBestState); err != nil {
		return NewBlockChainError(StoreBeaconBestStateError, err)
	}
	Logger.log.Debugf("Store Beacon Block Height %+v with Hash %+v ", beaconBlock.Header.Height, blockHash)
	if err := rawdbv2.StoreBeaconBlock(blockchain.GetDatabase(), beaconBlock.Header.Height, blockHash, beaconBlock); err != nil {
		return NewBlockChainError(StoreBeaconBlockError, err)
	}
	//statedb===========================START
	err = blockchain.updateDatabaseWithBlockRewardInfoV2(beaconBlock)
	if err != nil {
		return NewBlockChainError(UpdateDatabaseWithBlockRewardInfoError, err)
	}
	// execute, store
	err = blockchain.processBridgeInstructionsV2(beaconBestState.consensusStateDB, beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessBridgeInstructionError, err)
	}
	// execute, store
	err = blockchain.processPDEInstructionsV2(beaconBestState.consensusStateDB, beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessPDEInstructionError, err)
	}
	rootHash, err := beaconBestState.consensusStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = beaconBestState.consensusStateDB.Database().TrieDB().Commit(rootHash, false)
	if err != nil {
		return err
	}
	err = beaconBestState.consensusStateDB.Reset(rootHash)
	if err != nil {
		return err
	}
	//statedb===========================START
	return nil
}
