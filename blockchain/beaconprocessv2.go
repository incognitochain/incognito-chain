package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/pkg/errors"
	"reflect"
	"sort"
)

func (blockchain *BlockChain) VerifyPreSignBeaconBlockV2(beaconBlock *BeaconBlock, isPreSign bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	// Verify block only
	Logger.log.Infof("BEACON | Verify block for signing process %d, with hash %+v", beaconBlock.Header.Height, *beaconBlock.Hash())
	if err := blockchain.verifyPreProcessingBeaconBlockV2(beaconBlock, isPreSign); err != nil {
		return err
	}
	// Verify block with previous best state
	// Get Beststate of previous block == previous best state
	// Clone best state value into new variable
	beaconBestState := NewBeaconBestState()
	if err := beaconBestState.cloneBeaconBestStateFrom(blockchain.BestState.Beacon); err != nil {
		return err
	}
	// Verify block with previous best state
	// not verify agg signature in this function
	if err := beaconBestState.verifyBestStateWithBeaconBlock(beaconBlock, false, blockchain.config.ChainParams.Epoch); err != nil {
		return err
	}
	// Update best state with new block
	if err := beaconBestState.updateBeaconBestState(beaconBlock, blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.AssignOffset, blockchain.config.ChainParams.RandomTime); err != nil {
		return err
	}
	// Post verififcation: verify new beaconstate with corresponding block
	if err := beaconBestState.verifyPostProcessingBeaconBlock(beaconBlock, blockchain.config.RandomClient); err != nil {
		return err
	}
	Logger.log.Infof("BEACON | Block %d, with hash %+v is VALID to be 🖊 signed", beaconBlock.Header.Height, *beaconBlock.Hash())
	return nil
}

func (blockchain *BlockChain) InsertBeaconBlockV2(beaconBlock *BeaconBlock, isValidated bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()

	currentBeaconBestState := GetBeaconBestState()
	if currentBeaconBestState.BeaconHeight == beaconBlock.Header.Height && currentBeaconBestState.BestBlock.Header.Timestamp < beaconBlock.Header.Timestamp && currentBeaconBestState.BestBlock.Header.Round < beaconBlock.Header.Round {
		Logger.log.Infof("FORK BEACON, Current Beacon Block Height %+v, Hash %+v | Try to Insert New Beacon Block Height %+v, Hash %+v", currentBeaconBestState.BeaconHeight, currentBeaconBestState.BestBlockHash, beaconBlock.Header.Height, beaconBlock.Header.Hash())
	}

	if beaconBlock.Header.Height != GetBeaconBestState().BeaconHeight+1 {
		return errors.New("Not expected height")
	}

	blockHash := beaconBlock.Header.Hash()
	Logger.log.Infof("BEACON | Begin insert new Beacon Block height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	Logger.log.Infof("BEACON | Check Beacon Block existence before insert block height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	isExist, _ := rawdbv2.HasBeaconBlock(blockchain.GetDatabase(), beaconBlock.Header.Hash())
	if isExist {
		return NewBlockChainError(DuplicateShardBlockError, errors.New("This beaconBlock has been stored already"))
	}
	Logger.log.Infof("BEACON | Begin Insert new Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	if !isValidated {
		Logger.log.Infof("BEACON | Verify Pre Processing, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
		if err := blockchain.verifyPreProcessingBeaconBlock(beaconBlock, false); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("BEACON | SKIP Verify Pre Processing, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	}
	// Verify beaconBlock with previous best state
	if !isValidated {
		Logger.log.Infof("BEACON | Verify Best State With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
		// Verify beaconBlock with previous best state
		if err := blockchain.BestState.Beacon.verifyBestStateWithBeaconBlock(beaconBlock, true, blockchain.config.ChainParams.Epoch); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("BEACON | SKIP Verify Best State With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	}
	// process for slashing, make sure this one is called before update best state
	// since we'd like to process with old committee, not updated committee
	slashErr := blockchain.processForSlashing(beaconBlock)
	if slashErr != nil {
		Logger.log.Errorf("Failed to process slashing with error: %+v", NewBlockChainError(ProcessSlashingError, slashErr))
	}

	// snapshot current beacon committee and shard committee
	snapshotBeaconCommittee, snapshotAllShardCommittee, err := snapshotCommittee(blockchain.BestState.Beacon.BeaconCommittee, blockchain.BestState.Beacon.ShardCommittee)
	if err != nil {
		return NewBlockChainError(SnapshotCommitteeError, err)
	}
	_, snapshotAllShardPending, err := snapshotCommittee([]incognitokey.CommitteePublicKey{}, blockchain.BestState.Beacon.ShardPendingValidator)
	if err != nil {
		return NewBlockChainError(SnapshotCommitteeError, err)
	}

	snapshotShardWaiting := append([]incognitokey.CommitteePublicKey{}, blockchain.BestState.Beacon.CandidateShardWaitingForNextRandom...)
	snapshotShardWaiting = append(snapshotShardWaiting, blockchain.BestState.Beacon.CandidateBeaconWaitingForCurrentRandom...)

	snapshotRewardReceiver, err := snapshotRewardReceiver(blockchain.BestState.Beacon.RewardReceiver)
	if err != nil {
		return NewBlockChainError(SnapshotRewardReceiverError, err)
	}
	Logger.log.Infof("BEACON | Update BestState With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// Update best state with new beaconBlock

	if err := blockchain.BestState.Beacon.updateBeaconBestState(beaconBlock, blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.AssignOffset, blockchain.config.ChainParams.RandomTime); err != nil {
		return err
	}
	// updateNumOfBlocksByProducers updates number of blocks produced by producers
	blockchain.BestState.Beacon.updateNumOfBlocksByProducers(beaconBlock, blockchain.config.ChainParams.Epoch)

	newBeaconCommittee, newAllShardCommittee, err := snapshotCommittee(blockchain.BestState.Beacon.BeaconCommittee, blockchain.BestState.Beacon.ShardCommittee)
	if err != nil {
		return NewBlockChainError(SnapshotCommitteeError, err)
	}
	_, newAllShardPending, err := snapshotCommittee([]incognitokey.CommitteePublicKey{}, blockchain.BestState.Beacon.ShardPendingValidator)
	if err != nil {
		return NewBlockChainError(SnapshotCommitteeError, err)
	}

	notifyHighway := false
	newShardWaiting := append([]incognitokey.CommitteePublicKey{}, blockchain.BestState.Beacon.CandidateShardWaitingForNextRandom...)
	newShardWaiting = append(newShardWaiting, blockchain.BestState.Beacon.CandidateBeaconWaitingForCurrentRandom...)

	isChanged := !reflect.DeepEqual(snapshotBeaconCommittee, newBeaconCommittee)
	if isChanged {
		go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
		notifyHighway = true
	}

	isChanged = !reflect.DeepEqual(snapshotShardWaiting, newShardWaiting)
	if isChanged {
		go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
	}

	//Check shard-pending
	for shardID, committee := range newAllShardPending {
		if _, ok := snapshotAllShardPending[shardID]; ok {
			isChanged := !reflect.DeepEqual(snapshotAllShardPending[shardID], committee)
			if isChanged {
				go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
				notifyHighway = true
			}
		} else {
			go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
			notifyHighway = true
		}
	}
	//Check shard-committee
	for shardID, committee := range newAllShardCommittee {
		if _, ok := snapshotAllShardCommittee[shardID]; ok {
			isChanged := !reflect.DeepEqual(snapshotAllShardCommittee[shardID], committee)
			if isChanged {
				go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
				notifyHighway = true
			}
		} else {
			go blockchain.config.ConsensusEngine.CommitteeChange(common.BeaconChainKey)
			notifyHighway = true
		}
	}

	if !isValidated {
		Logger.log.Infof("BEACON | Verify Post Processing Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
		// Post verification: verify new beacon best state with corresponding beacon block
		if err := blockchain.BestState.Beacon.verifyPostProcessingBeaconBlock(beaconBlock, blockchain.config.RandomClient); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("BEACON | SKIP Verify Post Processing Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	}
	Logger.log.Infof("BEACON | Process Store Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	if err := blockchain.processStoreBeaconBlockV2(beaconBlock, snapshotBeaconCommittee, snapshotAllShardCommittee, snapshotRewardReceiver); err != nil {
		return err
	}
	blockchain.removeOldDataAfterProcessingBeaconBlock()
	Logger.log.Infof("Finish Insert new Beacon Block %+v, with hash %+v \n", beaconBlock.Header.Height, *beaconBlock.Hash())
	if beaconBlock.Header.Height%50 == 0 {
		BLogger.log.Debugf("Inserted beacon height: %d", beaconBlock.Header.Height)
	}
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewBeaconBlockTopic, beaconBlock))
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.BeaconBeststateTopic, blockchain.BestState.Beacon))

	// For masternode: broadcast new committee to highways
	if notifyHighway {
		go blockchain.config.Highway.BroadcastCommittee(
			blockchain.config.ChainParams.Epoch,
			newBeaconCommittee,
			newAllShardCommittee,
			newAllShardPending,
		)
	}
	return nil
}

func (blockchain *BlockChain) verifyPreProcessingBeaconBlockV2(beaconBlock *BeaconBlock, isPreSign bool) error {
	beaconLock := &blockchain.BestState.Beacon.lock
	beaconLock.RLock()
	defer beaconLock.RUnlock()

	//verify version
	if beaconBlock.Header.Version != BEACON_BLOCK_VERSION {
		return NewBlockChainError(WrongVersionError, fmt.Errorf("Expect block version to be equal to %+v but get %+v", BEACON_BLOCK_VERSION, beaconBlock.Header.Version))
	}
	// Verify parent hash exist or not
	previousBlockHash := beaconBlock.Header.PreviousBlockHash
	parentBlockBytes, err := rawdbv2.GetBeaconBlockByHash(blockchain.GetDatabase(), previousBlockHash)
	if err != nil {
		Logger.log.Criticalf("FORK BEACON DETECTED, New Beacon Block Height %+v, Hash %+v, Expected Previous Hash %+v, BUT Current Best State Height %+v and Hash %+v", beaconBlock.Header.Height, beaconBlock.Header.Hash(), beaconBlock.Header.PreviousBlockHash, blockchain.BestState.Beacon.BeaconHeight, blockchain.BestState.Beacon.BestBlockHash)
		blockchain.Synker.SyncBlkBeacon(true, false, false, []common.Hash{previousBlockHash}, nil, 0, 0, "")
		return NewBlockChainError(FetchBeaconBlockError, err)
	}
	previousBeaconBlock := NewBeaconBlock()
	err = json.Unmarshal(parentBlockBytes, previousBeaconBlock)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBeaconBlockError, fmt.Errorf("Failed to unmarshall parent block of block height %+v", beaconBlock.Header.Height))
	}
	// Verify block height with parent block
	if previousBeaconBlock.Header.Height+1 != beaconBlock.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Expect receive beacon block height %+v but get %+v", previousBeaconBlock.Header.Height+1, beaconBlock.Header.Height))
	}
	// Verify epoch with parent block
	if (beaconBlock.Header.Height != 1) && (beaconBlock.Header.Height%blockchain.config.ChainParams.Epoch == 1) && (previousBeaconBlock.Header.Epoch != beaconBlock.Header.Epoch-1) {
		return NewBlockChainError(WrongEpochError, fmt.Errorf("Expect receive beacon block epoch %+v greater than previous block epoch %+v, 1 value", beaconBlock.Header.Epoch, previousBeaconBlock.Header.Epoch))
	}
	// Verify timestamp with parent block
	if beaconBlock.Header.Timestamp <= previousBeaconBlock.Header.Timestamp {
		return NewBlockChainError(WrongTimestampError, fmt.Errorf("Expect receive beacon block with timestamp %+v greater than previous block timestamp %+v", beaconBlock.Header.Timestamp, previousBeaconBlock.Header.Timestamp))
	}
	if !verifyHashFromShardState(beaconBlock.Body.ShardState, beaconBlock.Header.ShardStateHash) {
		return NewBlockChainError(ShardStateHashError, fmt.Errorf("Expect shard state hash to be %+v", beaconBlock.Header.ShardStateHash))
	}
	tempInstructionArr := []string{}
	for _, strs := range beaconBlock.Body.Instructions {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	if hash, ok := verifyHashFromStringArray(tempInstructionArr, beaconBlock.Header.InstructionHash); !ok {
		return NewBlockChainError(InstructionHashError, fmt.Errorf("Expect instruction hash to be %+v but get %+v", beaconBlock.Header.InstructionHash, hash))
	}
	// Shard state must in right format
	// state[i].Height must less than state[i+1].Height and state[i+1].Height - state[i].Height = 1
	for _, shardStates := range beaconBlock.Body.ShardState {
		for i := 0; i < len(shardStates)-2; i++ {
			if shardStates[i+1].Height-shardStates[i].Height != 1 {
				return NewBlockChainError(ShardStateError, fmt.Errorf("Expect Shard State Height to be in the right format, %+v, %+v", shardStates[i+1].Height, shardStates[i].Height))
			}
		}
	}
	// Check if InstructionMerkleRoot is the root of merkle tree containing all instructions in this block
	flattenInsts, err := FlattenAndConvertStringInst(beaconBlock.Body.Instructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, err)
	}
	root := GetKeccak256MerkleRoot(flattenInsts)
	if !bytes.Equal(root, beaconBlock.Header.InstructionMerkleRoot[:]) {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Expect Instruction Merkle Root in Beacon Block Header to be %+v but get %+v", string(beaconBlock.Header.InstructionMerkleRoot[:]), string(root)))
	}
	// if pool does not have one of needed block, fail to verify
	if isPreSign {
		if err := blockchain.verifyPreProcessingBeaconBlockForSigningV2(beaconBlock); err != nil {
			return err
		}
	}
	return nil
}

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
