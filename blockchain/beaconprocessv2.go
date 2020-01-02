package blockchain

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"reflect"
	"sort"
)

func (blockchain *BlockChain) verifyPreProcessingBeaconBlockForSigningV2(beaconBlock *BeaconBlock) error {
	var err error
	rewardByEpochInstruction := [][]string{}
	tempShardStates := make(map[byte][]ShardState)
	stakeInstructions := [][]string{}
	validStakePublicKeys := []string{}
	swapInstructions := make(map[byte][][]string)
	bridgeInstructions := [][]string{}
	acceptedBlockRewardInstructions := [][]string{}
	stopAutoStakingInstructions := [][]string{}
	statefulActionsByShardID := map[byte][][]string{}
	// Get Reward Instruction By Epoch
	if beaconBlock.Header.Height%blockchain.config.ChainParams.Epoch == 1 {
		rewardByEpochInstruction, err = blockchain.BuildRewardInstructionByEpoch(beaconBlock.Header.Height, beaconBlock.Header.Epoch-1)
		if err != nil {
			return NewBlockChainError(BuildRewardInstructionError, err)
		}
	}
	// get shard to beacon blocks from pool
	allShardBlocks := blockchain.config.ShardToBeaconPool.GetValidBlock(nil)

	var keys []int
	for k := range beaconBlock.Body.ShardState {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	for _, value := range keys {
		shardID := byte(value)
		shardBlocks, ok := allShardBlocks[shardID]
		shardStates := beaconBlock.Body.ShardState[shardID]
		if !ok && len(shardStates) > 0 {
			return NewBlockChainError(GetShardToBeaconBlocksError, fmt.Errorf("Expect to get from pool ShardToBeacon Block from Shard %+v but failed", shardID))
		}
		// repeatly compare each shard to beacon block and shard state in new beacon block body
		if len(shardBlocks) >= len(shardStates) {
			shardBlocks = shardBlocks[:len(beaconBlock.Body.ShardState[shardID])]
			for index, shardState := range shardStates {
				if shardBlocks[index].Header.Height != shardState.Height {
					return NewBlockChainError(ShardStateHeightError, fmt.Errorf("Expect shard state height to be %+v but get %+v from pool", shardState.Height, shardBlocks[index].Header.Height))
					return NewBlockChainError(ShardStateHeightError, fmt.Errorf("Expect shard state height to be %+v but get %+v from pool(shard %v)", shardState.Height, shardBlocks[index].Header.Height, shardID))
				}
				blockHash := shardBlocks[index].Header.Hash()
				if !blockHash.IsEqual(&shardState.Hash) {
					return NewBlockChainError(ShardStateHashError, fmt.Errorf("Expect shard state height %+v has hash %+v but get %+v from pool", shardState.Height, shardState.Hash, shardBlocks[index].Header.Hash()))
				}
				if !reflect.DeepEqual(shardBlocks[index].Header.CrossShardBitMap, shardState.CrossShard) {
					return NewBlockChainError(ShardStateCrossShardBitMapError, fmt.Errorf("Expect shard state height %+v has bitmap %+v but get %+v from pool", shardState.Height, shardState.CrossShard, shardBlocks[index].Header.CrossShardBitMap))
				}
			}
			// Only accept block in one epoch
			for _, shardBlock := range shardBlocks {
				currentCommittee := blockchain.BestState.Beacon.GetAShardCommittee(shardID)
				errValidation := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(shardBlock, currentCommittee, beaconBestState.ShardConsensusAlgorithm[shardID])
				if errValidation != nil {
					return NewBlockChainError(ShardStateError, fmt.Errorf("Fail to verify with Shard To Beacon Block %+v, error %+v", shardBlock.Header.Height, err))
				}
			}
			for _, shardBlock := range shardBlocks {
				tempShardState, stakeInstruction, tempValidStakePublicKeys, swapInstruction, bridgeInstruction, acceptedBlockRewardInstruction, stopAutoStakingInstruction, statefulActions := blockchain.GetShardStateFromBlock(beaconBlock.Header.Height, shardBlock, shardID, false, validStakePublicKeys)
				tempShardStates[shardID] = append(tempShardStates[shardID], tempShardState[shardID])
				stakeInstructions = append(stakeInstructions, stakeInstruction...)
				swapInstructions[shardID] = append(swapInstructions[shardID], swapInstruction[shardID]...)
				bridgeInstructions = append(bridgeInstructions, bridgeInstruction...)
				acceptedBlockRewardInstructions = append(acceptedBlockRewardInstructions, acceptedBlockRewardInstruction)
				stopAutoStakingInstructions = append(stopAutoStakingInstructions, stopAutoStakingInstruction...)
				validStakePublicKeys = append(validStakePublicKeys, tempValidStakePublicKeys...)

				// group stateful actions by shardID
				_, found := statefulActionsByShardID[shardID]
				if !found {
					statefulActionsByShardID[shardID] = statefulActions
				} else {
					statefulActionsByShardID[shardID] = append(statefulActionsByShardID[shardID], statefulActions...)
				}
			}
		} else {
			return NewBlockChainError(GetShardToBeaconBlocksError, fmt.Errorf("Expect to get more than %+v ShardToBeaconBlock but only get %+v (shard %v)", len(beaconBlock.Body.ShardState[shardID]), len(shardBlocks), shardID))
		}
	}
	// build stateful instructions
	statefulInsts := blockchain.buildStatefulInstructionsV2(blockchain.BestState.Beacon.featureStateDB, statefulActionsByShardID, beaconBlock.Header.Height)
	bridgeInstructions = append(bridgeInstructions, statefulInsts...)
	tempInstruction, err := blockchain.BestState.Beacon.GenerateInstruction(beaconBlock.Header.Height,
		stakeInstructions, swapInstructions, stopAutoStakingInstructions,
		blockchain.BestState.Beacon.CandidateShardWaitingForCurrentRandom,
		bridgeInstructions, acceptedBlockRewardInstructions,
		blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.RandomTime, blockchain)
	if err != nil {
		return err
	}
	if len(rewardByEpochInstruction) != 0 {
		tempInstruction = append(tempInstruction, rewardByEpochInstruction...)
	}
	tempInstructionArr := []string{}
	for _, strs := range tempInstruction {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	tempInstructionHash, err := generateHashFromStringArray(tempInstructionArr)
	if err != nil {
		return NewBlockChainError(GenerateInstructionHashError, fmt.Errorf("Fail to generate hash for instruction %+v", tempInstructionArr))
	}
	if !tempInstructionHash.IsEqual(&beaconBlock.Header.InstructionHash) {
		return NewBlockChainError(InstructionHashError, fmt.Errorf("Expect Instruction Hash in Beacon Header to be %+v, but get %+v, validator instructions: %+v", beaconBlock.Header.InstructionHash, tempInstructionHash, tempInstruction))
	}
	return nil
}

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
	err = blockchain.processBridgeInstructionsV2(beaconBestState.featureStateDB, beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessBridgeInstructionError, err)
	}
	// execute, store
	err = blockchain.processPDEInstructionsV2(beaconBestState.featureStateDB, beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessPDEInstructionError, err)
	}
	consensusRootHash, err := beaconBestState.consensusStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = beaconBestState.consensusStateDB.Database().TrieDB().Commit(consensusRootHash, false)
	if err != nil {
		return err
	}
	err = beaconBestState.consensusStateDB.Reset(consensusRootHash)
	if err != nil {
		return err
	}
	featureRootHash, err := beaconBestState.featureStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = beaconBestState.featureStateDB.Database().TrieDB().Commit(featureRootHash, false)
	if err != nil {
		return err
	}
	err = beaconBestState.featureStateDB.Reset(featureRootHash)
	if err != nil {
		return err
	}
	rewardRootHash, err := beaconBestState.rewardStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = beaconBestState.rewardStateDB.Database().TrieDB().Commit(rewardRootHash, false)
	if err != nil {
		return err
	}
	err = beaconBestState.rewardStateDB.Reset(rewardRootHash)
	if err != nil {
		return err
	}
	beaconBestState.ConsensusStateRootHash[beaconBlock.Header.Height] = consensusRootHash
	beaconBestState.FeatureStateRootHash[beaconBlock.Header.Height] = featureRootHash
	beaconBestState.RewardStateRootHash[beaconBlock.Header.Height] = rewardRootHash
	//statedb===========================START
	return nil
}
