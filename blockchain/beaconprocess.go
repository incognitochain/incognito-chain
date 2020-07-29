package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

	"github.com/incognitochain/incognito-chain/blockchain/btc"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/pkg/errors"
)

// VerifyPreSignBeaconBlock should receives block in consensus round
// It verify validity of this function before sign it
// This should be verify in the first round of consensus
//	Step:
//	1. Verify Pre proccessing data
//	2. Retrieve beststate for new block, store in local variable
//	3. Update: process local beststate with new block
//	4. Verify Post processing: updated local beststate and newblock
//	Return:
//	- No error: valid and can be sign
//	- Error: invalid new block
func (blockchain *BlockChain) VerifyPreSignBeaconBlock(beaconBlock *BeaconBlock, isPreSign bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	// Verify block only
	Logger.log.Infof("BEACON | Verify block for signing process %d, with hash %+v", beaconBlock.Header.Height, *beaconBlock.Hash())
	committeeChange := newCommitteeChange()
	if err := blockchain.verifyPreProcessingBeaconBlock(beaconBlock, isPreSign); err != nil {
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
	if err := beaconBestState.updateBeaconBestState(beaconBlock, blockchain, committeeChange); err != nil {
		return err
	}
	// Post verififcation: verify new beaconstate with corresponding block
	if err := beaconBestState.verifyPostProcessingBeaconBlock(beaconBlock, blockchain.config.RandomClient); err != nil {
		return err
	}
	Logger.log.Infof("BEACON | Block %d, with hash %+v is VALID to be 🖊 signed", beaconBlock.Header.Height, *beaconBlock.Hash())
	return nil
}

func (blockchain *BlockChain) InsertBeaconBlock(beaconBlock *BeaconBlock, isValidated bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	currentBeaconBestState := GetBeaconBestState()
	blockHash := beaconBlock.Header.Hash()
	committeeChange := newCommitteeChange()
	if currentBeaconBestState.BeaconHeight == beaconBlock.Header.Height && currentBeaconBestState.BestBlock.Header.Timestamp < beaconBlock.Header.Timestamp && currentBeaconBestState.BestBlock.Header.Round < beaconBlock.Header.Round {
		currentBeaconHeight := currentBeaconBestState.BeaconHeight
		currentBeaconHash := currentBeaconBestState.BestBlockHash
		Logger.log.Infof("FORK BEACON, Current Beacon Block Height %+v, Hash %+v | Try to Insert New Beacon Block Height %+v, Hash %+v", currentBeaconHeight, currentBeaconHash, beaconBlock.Header.Height, beaconBlock.Header.Hash())
		if err := blockchain.ValidateBlockWithPreviousBeaconBestState(beaconBlock); err != nil {
			Logger.log.Error(err)
			return err
		}
		if err := blockchain.revertBeaconState(); err != nil {
			Logger.log.Error(err)
			return err
		}
		Logger.log.Infof("REVERTED BEACON, Revert Beacon Block Height %+v, Hash %+v", currentBeaconHeight, currentBeaconHash)
	}

	if beaconBlock.Header.Height != GetBeaconBestState().BeaconHeight+1 {
		return errors.New("Not expected height")
	}
	Logger.log.Infof("BEACON | Begin insert new Beacon Block height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	Logger.log.Debugf("BEACON | Begin Insert new Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	if !isValidated {
		Logger.log.Debugf("BEACON | Verify Pre Processing, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
		if err := blockchain.verifyPreProcessingBeaconBlock(beaconBlock, false); err != nil {
			return err
		}
	} else {
		Logger.log.Debugf("BEACON | SKIP Verify Pre Processing, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	}
	// Verify beaconBlock with previous best state
	if !isValidated {
		Logger.log.Debugf("BEACON | Verify Best State With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
		// Verify beaconBlock with previous best state
		if err := blockchain.BestState.Beacon.verifyBestStateWithBeaconBlock(beaconBlock, true, blockchain.config.ChainParams.Epoch); err != nil {
			return err
		}
	} else {
		Logger.log.Debugf("BEACON | SKIP Verify Best State With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	}
	// Backup beststate
	err := rawdbv2.CleanUpPreviousBeaconBestState(blockchain.GetDatabase())
	if err != nil {
		return NewBlockChainError(CleanBackUpError, err)
	}
	err = blockchain.BackupCurrentBeaconState(beaconBlock)
	if err != nil {
		return NewBlockChainError(BackUpBestStateError, err)
	}
	// process for slashing, make sure this one is called before update best state
	// since we'd like to process with old committee, not updated committee
	slashErr := blockchain.processForSlashing(blockchain.BestState.Beacon.slashStateDB, beaconBlock)
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
	Logger.log.Debugf("BEACON | Update BestState With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// Update best state with new beaconBlock

	if err := blockchain.BestState.Beacon.updateBeaconBestState(beaconBlock, blockchain, committeeChange); err != nil {
		errRevert := blockchain.revertBeaconBestState()
		if errRevert != nil {
			return errRevert
		}
		return err
	}
	// updateNumOfBlocksByProducers updates number of blocks produced by producers
	blockchain.BestState.Beacon.updateNumOfBlocksByProducers(beaconBlock, blockchain.config.ChainParams.Epoch)

	newBeaconCommittee, newAllShardCommittee, err := snapshotCommittee(blockchain.BestState.Beacon.BeaconCommittee, blockchain.BestState.Beacon.ShardCommittee)
	if err != nil {
		errRevert := blockchain.revertBeaconBestState()
		if errRevert != nil {
			return errRevert
		}
		return NewBlockChainError(SnapshotCommitteeError, err)
	}
	_, newAllShardPending, err := snapshotCommittee([]incognitokey.CommitteePublicKey{}, blockchain.BestState.Beacon.ShardPendingValidator)
	if err != nil {
		errRevert := blockchain.revertBeaconBestState()
		if errRevert != nil {
			return errRevert
		}
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
		Logger.log.Debugf("BEACON | Verify Post Processing Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
		// Post verification: verify new beacon best state with corresponding beacon block
		if err := blockchain.BestState.Beacon.verifyPostProcessingBeaconBlock(beaconBlock, blockchain.config.RandomClient); err != nil {
			errRevert := blockchain.revertBeaconBestState()
			if errRevert != nil {
				return errRevert
			}
			return err
		}
	} else {
		Logger.log.Debugf("BEACON | SKIP Verify Post Processing Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	}
	Logger.log.Infof("BEACON | Process Store Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	if err := blockchain.processStoreBeaconBlock(beaconBlock, snapshotBeaconCommittee, snapshotAllShardCommittee, snapshotRewardReceiver, committeeChange); err != nil {
		errRevert := blockchain.revertBeaconState()
		if errRevert != nil {
			return errRevert
		}
		return err
	}
	blockchain.removeOldDataAfterProcessingBeaconBlock()
	Logger.log.Infof("🔗 Finish Insert new Beacon Block %+v, with hash %+v \n", beaconBlock.Header.Height, *beaconBlock.Hash())
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

// updateNumOfBlocksByProducers updates number of blocks produced by producers
func (beaconBestState *BeaconBestState) updateNumOfBlocksByProducers(beaconBlock *BeaconBlock, chainParamEpoch uint64) {
	producer := beaconBlock.GetProducerPubKeyStr()
	if beaconBlock.GetHeight()%chainParamEpoch == 1 {
		beaconBestState.NumOfBlocksByProducers = map[string]uint64{
			producer: 1,
		}
	}
	// Update number of blocks produced by producers in epoch
	numOfBlks, found := beaconBestState.NumOfBlocksByProducers[producer]
	if !found {
		beaconBestState.NumOfBlocksByProducers[producer] = 1
	} else {
		beaconBestState.NumOfBlocksByProducers[producer] = numOfBlks + 1
	}
}

func (blockchain *BlockChain) removeOldDataAfterProcessingBeaconBlock() {
	//=========Remove beacon beaconBlock in pool
	go blockchain.config.BeaconPool.SetBeaconState(blockchain.BestState.Beacon.BeaconHeight)
	go blockchain.config.BeaconPool.RemoveBlock(blockchain.BestState.Beacon.BeaconHeight)
	//=========Remove shard to beacon beaconBlock in pool

	go func() {
		shardHeightMap := blockchain.BestState.Beacon.GetBestShardHeight()
		//force release readLock first, before execute the params in below function (which use same readLock).
		//if writeLock occur before release, readLock will be block
		blockchain.config.ShardToBeaconPool.SetShardState(shardHeightMap)
	}()
}

// VerifyPreProcessingBeaconBlock DOES NOT verify new block with best state
// DO NOT USE THIS with GENESIS BLOCK
// - Producer sanity data
// - Version: compatible with predefined version
// - Previous Block exist in database, fetch previous block by previous hash of new beacon block
// - Check new beacon block height is equal to previous block height + 1
// - Epoch = blockHeight % Epoch == 1 ? Previous Block Epoch + 1 : Previous Block Epoch
// - Timestamp of new beacon block is greater than previous beacon block timestamp
// - ShardStateHash: rebuild shard state hash from shard state body and compare with shard state hash in block header
// - InstructionHash: rebuild instruction hash from instruction body and compare with instruction hash in block header
// - InstructionMerkleRoot: rebuild instruction merkle root from instruction body and compare with instruction merkle root in block header
// - If verify block for signing then verifyPreProcessingBeaconBlockForSigning
func (blockchain *BlockChain) verifyPreProcessingBeaconBlock(beaconBlock *BeaconBlock, isPreSign bool) error {
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
		if err := blockchain.verifyPreProcessingBeaconBlockForSigning(beaconBlock); err != nil {
			return err
		}
	}
	return nil
}

// verifyPreProcessingBeaconBlockForSigning verify block before signing
// Must pass these following condition:
//  - Rebuild Reward By Epoch Instruction
//  - Get All Shard To Beacon Block in Shard To Beacon Pool
//  - For all Shard To Beacon Blocks in each Shard
//	  + Compare all shard height of shard states in body and these Shard To Beacon Blocks (got from pool)
//		* Must be have the same range of height
//		* Compare CrossShardBitMap of each Shard To Beacon Block and Shard State in New Beacon Block Body
//	  + After finish comparing these shard to beacon blocks with shard states in new beacon block body
//		* Verifying Shard To Beacon Block Agg Signature
//		* Only accept block in one epoch
//	  + Get Instruction from these Shard To Beacon Blocks:
//		* Stake Instruction
//		* Swap Instruction
//		* Bridge Instruction
//		* Block Reward Instruction
//	+ Generate Instruction Hash from all recently got instructions
//	+ Compare just created Instruction Hash with Instruction Hash In Beacon Header
func (blockchain *BlockChain) verifyPreProcessingBeaconBlockForSigning(beaconBlock *BeaconBlock) error {
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
	rewardForCustodianByEpoch := map[common.Hash]uint64{}

	portalParams := blockchain.GetPortalParams(beaconBlock.GetHeight())

	// Get Reward Instruction By Epoch
	if beaconBlock.Header.Height%blockchain.config.ChainParams.Epoch == 1 {
		featureStateDB := beaconBestState.GetCopiedFeatureStateDB()
		totalLockedCollateral, err := getTotalLockedCollateralInEpoch(featureStateDB)
		if err != nil {
			return NewBlockChainError(GetTotalLockedCollateralError, err)
		}
		isSplitRewardForCustodian := totalLockedCollateral > 0
		percentCustodianRewards := portalParams.MaxPercentCustodianRewards
		if totalLockedCollateral < portalParams.MinLockCollateralAmountInEpoch {
			percentCustodianRewards = portalParams.MinPercentCustodianRewards
		}

		rewardByEpochInstruction, rewardForCustodianByEpoch, err = blockchain.buildRewardInstructionByEpoch(beaconBlock.Header.Height, beaconBlock.Header.Epoch-1, blockchain.BestState.Beacon.GetCopiedRewardStateDB(), isSplitRewardForCustodian, percentCustodianRewards)
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
	statefulInsts := blockchain.buildStatefulInstructions(blockchain.BestState.Beacon.featureStateDB, statefulActionsByShardID, beaconBlock.Header.Height, rewardForCustodianByEpoch, portalParams)
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

//  verifyBestStateWithBeaconBlock verify the validation of a block with some best state in cache or current best state
//  Get beacon state of this block
//  For example, new blockHeight is 91 then beacon state of this block must have height 90
//  OR new block has previous has is beacon best block hash
//  - Get producer via index and compare with producer address in beacon block headerproce
//  - Validate public key and signature sanity
//  - Validate Agg Signature
//  - Beacon Best State has best block is previous block of new beacon block
//  - Beacon Best State has height compatible with new beacon block
//  - Beacon Best State has epoch compatible with new beacon block
//  - Beacon Best State has best shard height compatible with shard state of new beacon block
//  - New Stake public key must not found in beacon best state (candidate, pending validator, committee)
func (beaconBestState *BeaconBestState) verifyBestStateWithBeaconBlock(beaconBlock *BeaconBlock, isVerifySig bool, chainParamEpoch uint64) error {
	beaconBestState.lock.RLock()
	defer beaconBestState.lock.RUnlock()
	//verify producer via index
	producerPublicKey := beaconBlock.Header.Producer
	producerPosition := beaconBestState.GetProducerIndexFromBlock(beaconBlock)
	tempProducer, err := beaconBestState.BeaconCommittee[producerPosition].ToBase58() //.GetMiningKeyBase58(common.BridgeConsensus)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	if strings.Compare(string(tempProducer), producerPublicKey) != 0 {
		return NewBlockChainError(BeaconBlockProducerError, fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, producerPublicKey))
	}

	//=============End Verify Aggegrate signature
	if !beaconBestState.BestBlockHash.IsEqual(&beaconBlock.Header.PreviousBlockHash) {
		return NewBlockChainError(BeaconBestStateBestBlockNotCompatibleError, errors.New("previous us block should be :"+beaconBestState.BestBlockHash.String()))
	}
	if beaconBestState.BeaconHeight+1 != beaconBlock.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(beaconBlock.Header.Height+1))))
	}
	if beaconBlock.Header.Height%chainParamEpoch == 1 && beaconBestState.Epoch+1 != beaconBlock.Header.Epoch {
		return NewBlockChainError(WrongEpochError, fmt.Errorf("Expect beacon block height %+v has epoch %+v but get %+v", beaconBlock.Header.Height, beaconBestState.Epoch+1, beaconBlock.Header.Epoch))
	}
	if beaconBlock.Header.Height%chainParamEpoch != 1 && beaconBestState.Epoch != beaconBlock.Header.Epoch {
		return NewBlockChainError(WrongEpochError, fmt.Errorf("Expect beacon block height %+v has epoch %+v but get %+v", beaconBlock.Header.Height, beaconBestState.Epoch, beaconBlock.Header.Epoch))
	}
	// check shard states of new beacon block and beacon best state
	// shard state of new beacon block must be greater or equal to current best shard height
	for shardID, shardStates := range beaconBlock.Body.ShardState {
		if bestShardHeight, ok := beaconBestState.BestShardHeight[shardID]; !ok {
			if shardStates[0].Height != 2 {
				return NewBlockChainError(BeaconBestStateBestShardHeightNotCompatibleError, fmt.Errorf("Shard %+v best height not found in beacon best state", shardID))
			}
		} else {
			if len(shardStates) > 0 {
				if bestShardHeight > shardStates[0].Height {
					return NewBlockChainError(BeaconBestStateBestShardHeightNotCompatibleError, fmt.Errorf("Expect Shard %+v has state greater than or equal to %+v but get %+v", shardID, bestShardHeight, shardStates[0].Height))
				}
				if bestShardHeight < shardStates[0].Height && bestShardHeight+1 != shardStates[0].Height {
					return NewBlockChainError(BeaconBestStateBestShardHeightNotCompatibleError, fmt.Errorf("Expect Shard %+v has state %+v but get %+v", shardID, bestShardHeight+1, shardStates[0].Height))
				}
			}
		}
	}
	//=============Verify Stake Public Key
	newBeaconCandidate, newShardCandidate := GetStakingCandidate(*beaconBlock)
	if !reflect.DeepEqual(newBeaconCandidate, []string{}) {
		validBeaconCandidate := beaconBestState.GetValidStakers(newBeaconCandidate)
		if !reflect.DeepEqual(validBeaconCandidate, newBeaconCandidate) {
			return NewBlockChainError(CandidateError, errors.New("beacon candidate list is INVALID"))
		}
	}
	if !reflect.DeepEqual(newShardCandidate, []string{}) {
		validShardCandidate := beaconBestState.GetValidStakers(newShardCandidate)
		if !reflect.DeepEqual(validShardCandidate, newShardCandidate) {
			return NewBlockChainError(CandidateError, errors.New("shard candidate list is INVALID"))
		}
	}
	//=============End Verify Stakers
	return nil
}

//  verifyPostProcessingBeaconBlock verify block after update beacon best state
//  - Validator root: BeaconCommittee + BeaconPendingValidator
//  - Beacon Candidate root: CandidateBeaconWaitingForCurrentRandom + CandidateBeaconWaitingForNextRandom
//  - Shard Candidate root: CandidateShardWaitingForCurrentRandom + CandidateShardWaitingForNextRandom
//  - Shard Validator root: ShardCommittee + ShardPendingValidator
//  - Random number if have in instruction
func (beaconBestState *BeaconBestState) verifyPostProcessingBeaconBlock(beaconBlock *BeaconBlock, randomClient btc.RandomClient) error {
	beaconBestState.lock.RLock()
	defer beaconBestState.lock.RUnlock()
	var (
		strs []string
	)

	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee)
	if err != nil {
		panic(err)
	}
	strs = append(strs, beaconCommitteeStr...)

	beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconPendingValidator)
	if err != nil {
		panic(err)
	}
	strs = append(strs, beaconPendingValidatorStr...)
	if hash, ok := verifyHashFromStringArray(strs, beaconBlock.Header.BeaconCommitteeAndValidatorRoot); !ok {
		return NewBlockChainError(BeaconCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v but get %+v", beaconBlock.Header.BeaconCommitteeAndValidatorRoot, hash))
	}
	strs = []string{}

	candidateBeaconWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateBeaconWaitingForCurrentRandom)
	if err != nil {
		panic(err)
	}
	strs = append(strs, candidateBeaconWaitingForCurrentRandomStr...)

	candidateBeaconWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateBeaconWaitingForNextRandom)
	if err != nil {
		panic(err)
	}
	strs = append(strs, candidateBeaconWaitingForNextRandomStr...)
	if hash, ok := verifyHashFromStringArray(strs, beaconBlock.Header.BeaconCandidateRoot); !ok {
		return NewBlockChainError(BeaconCandidateRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v but get %+v", beaconBlock.Header.BeaconCandidateRoot, hash))
	}
	strs = []string{}

	candidateShardWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForCurrentRandom)
	if err != nil {
		panic(err)
	}
	strs = append(strs, candidateShardWaitingForCurrentRandomStr...)

	candidateShardWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForNextRandom)
	if err != nil {
		panic(err)
	}
	strs = append(strs, candidateShardWaitingForNextRandomStr...)
	if hash, ok := verifyHashFromStringArray(strs, beaconBlock.Header.ShardCandidateRoot); !ok {
		return NewBlockChainError(ShardCandidateRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v but get %+v", beaconBlock.Header.ShardCandidateRoot, hash))
	}

	shardPendingValidator := make(map[byte][]string)
	for shardID, keyList := range beaconBestState.ShardPendingValidator {
		keyListStr, err := incognitokey.CommitteeKeyListToString(keyList)
		if err != nil {
			return err
		}
		shardPendingValidator[shardID] = keyListStr
	}

	shardCommittee := make(map[byte][]string)
	for shardID, keyList := range beaconBestState.ShardCommittee {
		keyListStr, err := incognitokey.CommitteeKeyListToString(keyList)
		if err != nil {
			return err
		}
		shardCommittee[shardID] = keyListStr
	}
	ok := verifyHashFromMapByteString(shardPendingValidator, shardCommittee, beaconBlock.Header.ShardCommitteeAndValidatorRoot)
	if !ok {
		return NewBlockChainError(ShardCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v", beaconBlock.Header.ShardCommitteeAndValidatorRoot))
	}
	if hash, ok := verifyHashFromMapStringBool(beaconBestState.AutoStaking, beaconBlock.Header.AutoStakingRoot); !ok {
		return NewBlockChainError(ShardCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v but get %+v", beaconBlock.Header.AutoStakingRoot, hash))
	}
	if !TestRandom {
		//COMMENT FOR TESTING
		instructions := beaconBlock.Body.Instructions
		for _, l := range instructions {
			if l[0] == "random" {
				startTime := time.Now()
				// ["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
				nonce, err := strconv.Atoi(l[1])
				if err != nil {
					Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
					return NewBlockChainError(UnExpectedError, err)
				}
				ok, err = randomClient.VerifyNonceWithTimestamp(startTime, beaconBestState.BlockMaxCreateTime, beaconBestState.CurrentRandomTimeStamp, int64(nonce))
				Logger.log.Infof("Verify Random number %+v", ok)
				if err != nil {
					Logger.log.Error("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
					return NewBlockChainError(UnExpectedError, err)
				}
				if !ok {
					return NewBlockChainError(RandomError, errors.New("Error verify random number"))
				}
			}
		}
	}
	return nil
}

// updateBeaconBestState update beststate with new beacon block
func (beaconBestState *BeaconBestState) updateBeaconBestState(beaconBlock *BeaconBlock, blockchain *BlockChain, committeeChange *committeeChange) error {
	beaconBestState.lock.Lock()
	defer beaconBestState.lock.Unlock()
	var chainParamEpoch = blockchain.config.ChainParams.Epoch
	var chainParamAssignOffset = blockchain.config.ChainParams.AssignOffset
	var randomTime = blockchain.config.ChainParams.RandomTime
	Logger.log.Debugf("Start processing new block at height %d, with hash %+v", beaconBlock.Header.Height, *beaconBlock.Hash())
	newBeaconCandidate := []incognitokey.CommitteePublicKey{}
	newShardCandidate := []incognitokey.CommitteePublicKey{}
	// Logger.log.Infof("Start processing new block at height %d, with hash %+v", newBlock.Header.Height, *newBlock.Hash())
	if beaconBlock == nil {
		return errors.New("null pointer")
	}
	// signal of random parameter from beacon block
	randomFlag := false
	// update BestShardHash, BestBlock, BestBlockHash
	beaconBestState.PreviousBestBlockHash = beaconBestState.BestBlockHash
	beaconBestState.BestBlockHash = *beaconBlock.Hash()
	beaconBestState.BestBlock = *beaconBlock
	beaconBestState.Epoch = beaconBlock.Header.Epoch
	beaconBestState.BeaconHeight = beaconBlock.Header.Height
	if beaconBlock.Header.Height == 1 {
		beaconBestState.BeaconProposerIndex = 0
	} else {
		//TODO: 0xsirush revert this code
		beaconBestState.BeaconProposerIndex = 0 //(beaconBestState.BeaconProposerIndex + beaconBlock.Header.Round) % len(beaconBestState.BeaconCommittee)
	}
	if beaconBestState.BestShardHash == nil {
		beaconBestState.BestShardHash = make(map[byte]common.Hash)
	}
	if beaconBestState.BestShardHeight == nil {
		beaconBestState.BestShardHeight = make(map[byte]uint64)
	}
	// Update new best new block hash
	for shardID, shardStates := range beaconBlock.Body.ShardState {
		beaconBestState.BestShardHash[shardID] = shardStates[len(shardStates)-1].Hash
		beaconBestState.BestShardHeight[shardID] = shardStates[len(shardStates)-1].Height
	}
	// processing instruction
	snapshotAutoStaking := make(map[string]bool)
	for k, v := range beaconBestState.AutoStaking {
		snapshotAutoStaking[k] = v
	}
	for _, instruction := range beaconBlock.Body.Instructions {
		err, tempRandomFlag, tempNewBeaconCandidate, tempNewShardCandidate := beaconBestState.processInstruction(instruction, blockchain, committeeChange, snapshotAutoStaking)
		if err != nil {
			return err
		}
		if tempRandomFlag {
			randomFlag = tempRandomFlag
		}
		if len(tempNewBeaconCandidate) > 0 {
			newBeaconCandidate = append(newBeaconCandidate, tempNewBeaconCandidate...)
		}
		if len(tempNewShardCandidate) > 0 {
			newShardCandidate = append(newShardCandidate, tempNewShardCandidate...)
		}
	}
	// update candidate list after processing instructions
	beaconBestState.CandidateBeaconWaitingForNextRandom = append(beaconBestState.CandidateBeaconWaitingForNextRandom, newBeaconCandidate...)
	committeeChange.nextEpochBeaconCandidateAdded = append(committeeChange.nextEpochBeaconCandidateAdded, newBeaconCandidate...)
	beaconBestState.CandidateShardWaitingForNextRandom = append(beaconBestState.CandidateShardWaitingForNextRandom, newShardCandidate...)
	committeeChange.nextEpochShardCandidateAdded = append(committeeChange.nextEpochShardCandidateAdded, newShardCandidate...)
	if beaconBestState.BeaconHeight%chainParamEpoch == 1 && beaconBestState.BeaconHeight != 1 {
		// Begin of each epoch
		beaconBestState.IsGetRandomNumber = false
		// Before get random from bitcoin
	} else if beaconBestState.BeaconHeight%chainParamEpoch >= randomTime {
		// snap shot candidate list, prepare to get random number (beaconHeight == random time)
		if beaconBestState.BeaconHeight%chainParamEpoch == randomTime {
			// snapshot candidate list
			committeeChange.currentEpochShardCandidateAdded = beaconBestState.CandidateShardWaitingForNextRandom
			beaconBestState.CandidateShardWaitingForCurrentRandom = beaconBestState.CandidateShardWaitingForNextRandom
			committeeChange.currentEpochBeaconCandidateAdded = beaconBestState.CandidateBeaconWaitingForNextRandom
			beaconBestState.CandidateBeaconWaitingForCurrentRandom = beaconBestState.CandidateBeaconWaitingForNextRandom
			Logger.log.Debug("Beacon Process: CandidateShardWaitingForCurrentRandom: ", beaconBestState.CandidateShardWaitingForCurrentRandom)
			Logger.log.Debug("Beacon Process: CandidateBeaconWaitingForCurrentRandom: ", beaconBestState.CandidateBeaconWaitingForCurrentRandom)
			// reset candidate list
			committeeChange.nextEpochShardCandidateRemoved = beaconBestState.CandidateShardWaitingForNextRandom
			beaconBestState.CandidateShardWaitingForNextRandom = []incognitokey.CommitteePublicKey{}
			committeeChange.nextEpochBeaconCandidateRemoved = beaconBestState.CandidateBeaconWaitingForNextRandom
			beaconBestState.CandidateBeaconWaitingForNextRandom = []incognitokey.CommitteePublicKey{}
			// assign random timestamp
			beaconBestState.CurrentRandomTimeStamp = beaconBlock.Header.Timestamp
		}
		// if get new random number (beaconHeight > random time)
		// Assign candidate to shard
		// assign CandidateShardWaitingForCurrentRandom to ShardPendingValidator with CurrentRandom
		if randomFlag {
			beaconBestState.IsGetRandomNumber = true
			numberOfPendingValidator := make(map[byte]int)
			for shardID, pendingValidators := range beaconBestState.ShardPendingValidator {
				numberOfPendingValidator[shardID] = len(pendingValidators)
			}
			shardCandidatesStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.CandidateShardWaitingForCurrentRandom)
			if err != nil {
				panic(err)
			}
			remainShardCandidatesStr, assignedCandidates := assignShardCandidate(shardCandidatesStr, numberOfPendingValidator, beaconBestState.CurrentRandomNumber, chainParamAssignOffset, beaconBestState.ActiveShards)
			remainShardCandidates, err := incognitokey.CommitteeBase58KeyListToStruct(remainShardCandidatesStr)
			if err != nil {
				panic(err)
			}
			committeeChange.nextEpochShardCandidateAdded = append(committeeChange.nextEpochShardCandidateAdded, remainShardCandidates...)
			// append remain candidate into shard waiting for next random list
			beaconBestState.CandidateShardWaitingForNextRandom = append(beaconBestState.CandidateShardWaitingForNextRandom, remainShardCandidates...)
			// assign candidate into shard pending validator list
			for shardID, candidateListStr := range assignedCandidates {
				candidateList, err := incognitokey.CommitteeBase58KeyListToStruct(candidateListStr)
				if err != nil {
					panic(err)
				}
				committeeChange.shardSubstituteAdded[shardID] = candidateList
				beaconBestState.ShardPendingValidator[shardID] = append(beaconBestState.ShardPendingValidator[shardID], candidateList...)
			}
			committeeChange.currentEpochShardCandidateRemoved = beaconBestState.CandidateShardWaitingForCurrentRandom
			// delete CandidateShardWaitingForCurrentRandom list
			beaconBestState.CandidateShardWaitingForCurrentRandom = []incognitokey.CommitteePublicKey{}
			// shuffle CandidateBeaconWaitingForCurrentRandom with current random number
			newBeaconPendingValidator, err := ShuffleCandidate(beaconBestState.CandidateBeaconWaitingForCurrentRandom, beaconBestState.CurrentRandomNumber)
			if err != nil {
				return NewBlockChainError(ShuffleBeaconCandidateError, err)
			}
			committeeChange.currentEpochBeaconCandidateRemoved = beaconBestState.CandidateBeaconWaitingForCurrentRandom
			beaconBestState.CandidateBeaconWaitingForCurrentRandom = []incognitokey.CommitteePublicKey{}
			committeeChange.beaconSubstituteAdded = newBeaconPendingValidator
			beaconBestState.BeaconPendingValidator = append(beaconBestState.BeaconPendingValidator, newBeaconPendingValidator...)
		}
	}
	if err := beaconBestState.processAutoStakingChange(committeeChange); err != nil {
		return NewBlockChainError(ProcessAutoStakingError, err)
	}
	return nil
}

func (beaconBestState *BeaconBestState) initBeaconBestState(genesisBeaconBlock *BeaconBlock, blockchain *BlockChain, db incdb.Database) error {
	var (
		newBeaconCandidate = []incognitokey.CommitteePublicKey{}
		newShardCandidate  = []incognitokey.CommitteePublicKey{}
	)
	Logger.log.Info("Process Update Beacon Best State With Beacon Genesis Block")
	beaconBestState.lock.Lock()
	defer beaconBestState.lock.Unlock()
	beaconBestState.PreviousBestBlockHash = beaconBestState.BestBlockHash
	beaconBestState.BestBlockHash = *genesisBeaconBlock.Hash()
	beaconBestState.BestBlock = *genesisBeaconBlock
	beaconBestState.Epoch = genesisBeaconBlock.Header.Epoch
	beaconBestState.BeaconHeight = genesisBeaconBlock.Header.Height
	beaconBestState.BeaconProposerIndex = 0
	beaconBestState.BestShardHash = make(map[byte]common.Hash)
	beaconBestState.BestShardHeight = make(map[byte]uint64)
	// Update new best new block hash
	for shardID, shardStates := range genesisBeaconBlock.Body.ShardState {
		beaconBestState.BestShardHash[shardID] = shardStates[len(shardStates)-1].Hash
		beaconBestState.BestShardHeight[shardID] = shardStates[len(shardStates)-1].Height
	}
	// update param
	snapshotAutoStaking := make(map[string]bool)
	for k, v := range beaconBestState.AutoStaking {
		snapshotAutoStaking[k] = v
	}
	for _, instruction := range genesisBeaconBlock.Body.Instructions {
		err, _, tempNewBeaconCandidate, tempNewShardCandidate := beaconBestState.processInstruction(instruction, blockchain, newCommitteeChange(), snapshotAutoStaking)
		if err != nil {
			return err
		}
		newBeaconCandidate = append(newBeaconCandidate, tempNewBeaconCandidate...)
		newShardCandidate = append(newShardCandidate, tempNewShardCandidate...)
	}
	beaconBestState.BeaconCommittee = append(beaconBestState.BeaconCommittee, newBeaconCandidate...)
	beaconBestState.ConsensusAlgorithm = common.BlsConsensus
	beaconBestState.ShardConsensusAlgorithm = make(map[byte]string)
	for shardID := 0; shardID < beaconBestState.ActiveShards; shardID++ {
		beaconBestState.ShardCommittee[byte(shardID)] = append(beaconBestState.ShardCommittee[byte(shardID)], newShardCandidate[shardID*beaconBestState.MinShardCommitteeSize:(shardID+1)*beaconBestState.MinShardCommitteeSize]...)
		beaconBestState.ShardConsensusAlgorithm[byte(shardID)] = common.BlsConsensus
	}
	beaconBestState.Epoch = 1
	beaconBestState.NumOfBlocksByProducers = make(map[string]uint64)
	//statedb===========================START
	var err error
	dbAccessWarper := statedb.NewDatabaseAccessWarper(db)
	beaconBestState.featureStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	beaconBestState.consensusStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	beaconBestState.rewardStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	beaconBestState.slashStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	//statedb===========================END
	return nil
}

//  processInstruction, process these instruction:
//  - Random Instruction format
//		["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
//	- store random number into beststate
//  - Swap Instruction format
//		["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
//		["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
//    + Update shard/beacon pending validator and shard/beacon committee in beststate
//  - Stake Instruction
//	  + format
//		["stake", "pubkey1,pubkey2,..." "shard" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,..." flag]
//		["stake", "pubkey1,pubkey2,..." "beacon" "txStake1,txStake2,..." "rewardReceiver1,rewardReceiver2,..." flag]
//	  + Get Stake public key and for later storage
//  Return param
//  #1 error
//  #2 random flag
//  #3 new beacon candidate
//  #4 new shard candidate

func (beaconBestState *BeaconBestState) processInstruction(instruction []string, blockchain *BlockChain, committeeChange *committeeChange, autoStaking map[string]bool) (error, bool, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey) {
	newBeaconCandidates := []incognitokey.CommitteePublicKey{}
	newShardCandidates := []incognitokey.CommitteePublicKey{}
	if len(instruction) < 1 {
		return nil, false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
	}
	// ["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
	if instruction[0] == RandomAction {
		temp, err := strconv.Atoi(instruction[1])
		if err != nil {
			return NewBlockChainError(ProcessRandomInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		beaconBestState.CurrentRandomNumber = int64(temp)
		Logger.log.Infof("Random number found %d", beaconBestState.CurrentRandomNumber)
		return nil, true, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
	}
	if instruction[0] == StopAutoStake {
		committeePublicKeys := strings.Split(instruction[1], ",")
		for _, committeePublicKey := range committeePublicKeys {
			allCommitteeValidatorCandidate := beaconBestState.getAllCommitteeValidatorCandidateFlattenList()
			// check existence in all committee list
			if common.IndexOfStr(committeePublicKey, allCommitteeValidatorCandidate) == -1 {
				// if not found then delete auto staking data for this public key if present
				if _, ok := beaconBestState.AutoStaking[committeePublicKey]; ok {
					delete(beaconBestState.AutoStaking, committeePublicKey)
				}
			} else {
				// if found in committee list then turn off auto staking
				if _, ok := beaconBestState.AutoStaking[committeePublicKey]; ok {
					beaconBestState.AutoStaking[committeePublicKey] = false
					committeeChange.stopAutoStaking = append(committeeChange.stopAutoStaking, committeePublicKey)
				}
			}
		}
	}
	if instruction[0] == SwapAction {
		if common.IndexOfUint64(beaconBestState.BeaconHeight/blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.EpochBreakPointSwapNewKey) > -1 || len(instruction) == 7 {
			err := beaconBestState.processSwapInstructionForKeyListV2(instruction, committeeChange)
			if err != nil {
				return err, false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
			}
			return nil, false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		} else {
			Logger.log.Debug("Swap Instruction", instruction)
			inPublickeys := strings.Split(instruction[1], ",")
			Logger.log.Debug("Swap Instruction In Public Keys", inPublickeys)
			inPublickeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(inPublickeys)
			if err != nil {
				return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
			}
			outPublickeys := strings.Split(instruction[2], ",")
			Logger.log.Debug("Swap Instruction Out Public Keys", outPublickeys)
			outPublickeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(outPublickeys)
			if err != nil {
				if len(outPublickeys) != 0 {
					return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
			}

			if instruction[3] == "shard" {
				temp, err := strconv.Atoi(instruction[4])
				if err != nil {
					return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
				}
				shardID := byte(temp)
				// delete in public key out of sharding pending validator list
				if len(instruction[1]) > 0 {
					shardPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.ShardPendingValidator[shardID])
					if err != nil {
						return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					tempShardPendingValidator, err := RemoveValidator(shardPendingValidatorStr, inPublickeys)
					if err != nil {
						return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					// update shard pending validator
					committeeChange.shardSubstituteRemoved[shardID] = append(committeeChange.shardSubstituteRemoved[shardID], inPublickeyStructs...)
					beaconBestState.ShardPendingValidator[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(tempShardPendingValidator)
					if err != nil {
						return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					// add new public key to committees
					committeeChange.shardCommitteeAdded[shardID] = append(committeeChange.shardCommitteeAdded[shardID], inPublickeyStructs...)
					beaconBestState.ShardCommittee[shardID] = append(beaconBestState.ShardCommittee[shardID], inPublickeyStructs...)
				}
				// delete out public key out of current committees
				if len(instruction[2]) > 0 {
					//for _, value := range outPublickeyStructs {
					//	delete(beaconBestState.RewardReceiver, value.GetIncKeyBase58())
					//}
					shardCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.ShardCommittee[shardID])
					if err != nil {
						return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					tempShardCommittees, err := RemoveValidator(shardCommitteeStr, outPublickeys)
					if err != nil {
						return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					// remove old public key in shard committee update shard committee
					committeeChange.shardCommitteeRemoved[shardID] = append(committeeChange.shardCommitteeRemoved[shardID], outPublickeyStructs...)
					beaconBestState.ShardCommittee[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(tempShardCommittees)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					// Check auto stake in out public keys list
					// if auto staking not found or flag auto stake is false then do not re-stake for this out public key
					// if auto staking flag is true then system will automatically add this out public key to current candidate list
					for index, outPublicKey := range outPublickeys {
						if isAutoStaking, ok := autoStaking[outPublicKey]; !ok {
							if _, ok := beaconBestState.RewardReceiver[outPublicKey]; ok {
								delete(beaconBestState.RewardReceiver, outPublickeyStructs[index].GetIncKeyBase58())
							}
							continue
						} else {
							if !isAutoStaking {
								// delete this flag for next time staking
								delete(beaconBestState.RewardReceiver, outPublickeyStructs[index].GetIncKeyBase58())
								delete(beaconBestState.AutoStaking, outPublicKey)
							} else {
								shardCandidate, err := incognitokey.CommitteeBase58KeyListToStruct([]string{outPublicKey})
								if err != nil {
									return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
								}
								newShardCandidates = append(newShardCandidates, shardCandidate...)
							}
						}
					}
				}
			} else if instruction[3] == "beacon" {
				if len(instruction[1]) > 0 {
					beaconPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconPendingValidator)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					tempBeaconPendingValidator, err := RemoveValidator(beaconPendingValidatorStr, inPublickeys)
					if err != nil {
						return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					// update beacon pending validator
					committeeChange.beaconSubstituteRemoved = append(committeeChange.beaconSubstituteRemoved, inPublickeyStructs...)
					beaconBestState.BeaconPendingValidator, err = incognitokey.CommitteeBase58KeyListToStruct(tempBeaconPendingValidator)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					// add new public key to beacon committee
					committeeChange.beaconCommitteeAdded = append(committeeChange.beaconCommitteeAdded, inPublickeyStructs...)
					beaconBestState.BeaconCommittee = append(beaconBestState.BeaconCommittee, inPublickeyStructs...)
				}
				if len(instruction[2]) > 0 {
					// delete reward receiver
					//for _, value := range outPublickeyStructs {
					//	delete(beaconBestState.RewardReceiver, value.GetIncKeyBase58())
					//}
					beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconBestState.BeaconCommittee)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					tempBeaconCommittes, err := RemoveValidator(beaconCommitteeStr, outPublickeys)
					if err != nil {
						return NewBlockChainError(ProcessSwapInstructionError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					// remove old public key in beacon committee and update beacon best state
					committeeChange.beaconCommitteeRemoved = append(committeeChange.beaconCommitteeRemoved, outPublickeyStructs...)
					beaconBestState.BeaconCommittee, err = incognitokey.CommitteeBase58KeyListToStruct(tempBeaconCommittes)
					if err != nil {
						return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
					}
					for index, outPublicKey := range outPublickeys {
						if isAutoStaking, ok := autoStaking[outPublicKey]; !ok {
							if _, ok := beaconBestState.RewardReceiver[outPublicKey]; ok {
								delete(beaconBestState.RewardReceiver, outPublickeyStructs[index].GetIncKeyBase58())
							}
							continue
						} else {
							if !isAutoStaking {
								delete(beaconBestState.RewardReceiver, outPublickeyStructs[index].GetIncKeyBase58())
								delete(beaconBestState.AutoStaking, outPublicKey)
							} else {
								beaconCandidate, err := incognitokey.CommitteeBase58KeyListToStruct([]string{outPublicKey})
								if err != nil {
									return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
								}
								newBeaconCandidates = append(newBeaconCandidates, beaconCandidate...)
							}
						}
					}
				}
			}
		}
		return nil, false, newBeaconCandidates, newShardCandidates
	}
	// Update candidate
	// get staking candidate list and store
	// store new staking candidate
	if instruction[0] == StakeAction && instruction[2] == "beacon" {
		beaconCandidates := strings.Split(instruction[1], ",")
		beaconCandidatesStructs, err := incognitokey.CommitteeBase58KeyListToStruct(beaconCandidates)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		beaconRewardReceivers := strings.Split(instruction[4], ",")
		beaconAutoReStaking := strings.Split(instruction[5], ",")
		if len(beaconCandidatesStructs) != len(beaconRewardReceivers) && len(beaconRewardReceivers) != len(beaconAutoReStaking) {
			return NewBlockChainError(StakeInstructionError, fmt.Errorf("Expect Beacon Candidate (length %+v) and Beacon Reward Receiver (length %+v) and Beacon Auto ReStaking (lenght %+v) have equal length", len(beaconCandidates), len(beaconRewardReceivers), len(beaconAutoReStaking))), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		for index, candidate := range beaconCandidatesStructs {
			beaconBestState.RewardReceiver[candidate.GetIncKeyBase58()] = beaconRewardReceivers[index]
			if beaconAutoReStaking[index] == "true" {
				beaconBestState.AutoStaking[beaconCandidates[index]] = true
			} else {
				beaconBestState.AutoStaking[beaconCandidates[index]] = false
			}
		}

		newBeaconCandidates = append(newBeaconCandidates, beaconCandidatesStructs...)
		return nil, false, newBeaconCandidates, newShardCandidates
	}
	if instruction[0] == StakeAction && instruction[2] == "shard" {
		shardCandidates := strings.Split(instruction[1], ",")
		shardCandidatesStructs, err := incognitokey.CommitteeBase58KeyListToStruct(shardCandidates)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		shardRewardReceivers := strings.Split(instruction[4], ",")
		shardAutoReStaking := strings.Split(instruction[5], ",")
		if len(shardCandidates) != len(shardRewardReceivers) && len(shardRewardReceivers) != len(shardAutoReStaking) {
			return NewBlockChainError(StakeInstructionError, fmt.Errorf("Expect Beacon Candidate (length %+v) and Beacon Reward Receiver (length %+v) and Shard Auto ReStaking (length %+v) have equal length", len(shardCandidates), len(shardRewardReceivers), len(shardAutoReStaking))), false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
		}
		for index, candidate := range shardCandidatesStructs {
			beaconBestState.RewardReceiver[candidate.GetIncKeyBase58()] = shardRewardReceivers[index]
			if shardAutoReStaking[index] == "true" {
				beaconBestState.AutoStaking[shardCandidates[index]] = true
			} else {
				beaconBestState.AutoStaking[shardCandidates[index]] = false
			}
		}
		newShardCandidates = append(newShardCandidates, shardCandidatesStructs...)
		return nil, false, newBeaconCandidates, newShardCandidates
	}
	return nil, false, []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}
}

func (beaconBestState *BeaconBestState) processSwapInstructionForKeyListV2(instruction []string, committeeChange *committeeChange) error {
	if instruction[0] == SwapAction {
		if instruction[1] == "" && instruction[2] == "" {
			return nil
		}
		inPublicKeys := strings.Split(instruction[1], ",")
		inPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(inPublicKeys)
		if err != nil {
			return NewBlockChainError(UnExpectedError, err)
		}
		outPublicKeys := strings.Split(instruction[2], ",")
		outPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
		if err != nil {
			if len(outPublicKeys) != 0 {
				return NewBlockChainError(UnExpectedError, err)
			}
		}
		inRewardReceiver := strings.Split(instruction[6], ",")
		if len(inPublicKeys) != len(outPublicKeys) {
			return NewBlockChainError(ProcessSwapInstructionError, fmt.Errorf("length new committee %+v, length out committee %+v", len(inPublicKeys), len(outPublicKeys)))
		}
		if len(inPublicKeys) != len(inRewardReceiver) {
			return NewBlockChainError(ProcessSwapInstructionError, fmt.Errorf("length new committee %+v, new reward receiver %+v", len(inPublicKeys), len(inRewardReceiver)))
		}
		removedCommittee := len(inPublicKeys)
		if instruction[3] == "shard" {
			temp, err := strconv.Atoi(instruction[4])
			if err != nil {
				return NewBlockChainError(ProcessSwapInstructionError, err)
			}
			shardID := byte(temp)
			// update shard pending validator
			// committeeChange.shardCommitteeRemoved[shardID] = append(committeeChange.shardCommitteeRemoved[shardID], outPublicKeyStructs...)
			// add new public key to committees
			// committeeChange.shardCommitteeAdded[shardID] = append(committeeChange.shardCommitteeAdded[shardID], inPublicKeyStructs...)
			committeeReplace := [2][]incognitokey.CommitteePublicKey{}
			committeeReplace[common.REPLACE_OUT] = append(committeeReplace[common.REPLACE_OUT], outPublicKeyStructs...)
			committeeReplace[common.REPLACE_IN] = append(committeeReplace[common.REPLACE_IN], inPublicKeyStructs...)
			committeeChange.shardCommitteeReplaced[shardID] = committeeReplace
			remainedShardCommittees := beaconBestState.ShardCommittee[shardID][removedCommittee:]
			beaconBestState.ShardCommittee[shardID] = append(inPublicKeyStructs, remainedShardCommittees...)
		} else if instruction[3] == "beacon" {
			// committeeChange.beaconCommitteeRemoved = append(committeeChange.beaconCommitteeRemoved, outPublicKeyStructs...)
			// add new public key to committees
			// committeeChange.beaconCommitteeAdded = append(committeeChange.beaconCommitteeAdded, inPublicKeyStructs...)
			committeeChange.beaconCommitteeReplaced[common.REPLACE_OUT] = append(committeeChange.beaconCommitteeReplaced[common.REPLACE_OUT], outPublicKeyStructs...)
			committeeChange.beaconCommitteeReplaced[common.REPLACE_IN] = append(committeeChange.beaconCommitteeReplaced[common.REPLACE_IN], inPublicKeyStructs...)
			remainedBeaconCommittees := beaconBestState.BeaconCommittee[removedCommittee:]
			beaconBestState.BeaconCommittee = append(inPublicKeyStructs, remainedBeaconCommittees...)
		}
		for i := 0; i < removedCommittee; i++ {
			delete(beaconBestState.AutoStaking, outPublicKeys[i])
			delete(beaconBestState.RewardReceiver, outPublicKeyStructs[i].GetIncKeyBase58())
			beaconBestState.AutoStaking[inPublicKeys[i]] = false
			beaconBestState.RewardReceiver[inPublicKeyStructs[i].GetIncKeyBase58()] = inRewardReceiver[i]
		}
	}
	return nil
}

func (beaconBestState *BeaconBestState) processAutoStakingChange(committeeChange *committeeChange) error {
	stopAutoStakingIncognitoKey, err := incognitokey.CommitteeBase58KeyListToStruct(committeeChange.stopAutoStaking)
	if err != nil {
		return err
	}
	for _, committeePublicKey := range stopAutoStakingIncognitoKey {
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, committeeChange.nextEpochBeaconCandidateAdded) > -1 {
			continue
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, committeeChange.currentEpochBeaconCandidateAdded) > -1 {
			continue
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, committeeChange.nextEpochShardCandidateAdded) > -1 {
			continue
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, committeeChange.currentEpochShardCandidateAdded) > -1 {
			continue
		}
		flag := false
		for _, v := range committeeChange.shardSubstituteAdded {
			if incognitokey.IndexOfCommitteeKey(committeePublicKey, v) > -1 {
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		for _, v := range committeeChange.shardCommitteeAdded {
			if incognitokey.IndexOfCommitteeKey(committeePublicKey, v) > -1 {
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, committeeChange.beaconSubstituteAdded) > -1 {
			continue
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, committeeChange.beaconCommitteeAdded) > -1 {
			continue
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, beaconBestState.CandidateBeaconWaitingForNextRandom) > -1 {
			committeeChange.nextEpochBeaconCandidateAdded = append(committeeChange.nextEpochBeaconCandidateAdded, committeePublicKey)
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, beaconBestState.CandidateBeaconWaitingForCurrentRandom) > -1 {
			committeeChange.currentEpochBeaconCandidateAdded = append(committeeChange.currentEpochBeaconCandidateAdded, committeePublicKey)
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, beaconBestState.CandidateShardWaitingForNextRandom) > -1 {
			committeeChange.nextEpochShardCandidateAdded = append(committeeChange.nextEpochShardCandidateAdded, committeePublicKey)
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, beaconBestState.CandidateShardWaitingForCurrentRandom) > -1 {
			committeeChange.currentEpochShardCandidateAdded = append(committeeChange.currentEpochShardCandidateAdded, committeePublicKey)
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, beaconBestState.BeaconPendingValidator) > -1 {
			committeeChange.beaconSubstituteAdded = append(committeeChange.beaconSubstituteAdded, committeePublicKey)
		}
		if incognitokey.IndexOfCommitteeKey(committeePublicKey, beaconBestState.BeaconCommittee) > -1 {
			committeeChange.beaconCommitteeAdded = append(committeeChange.beaconCommitteeAdded, committeePublicKey)
		}
		for k, v := range beaconBestState.ShardCommittee {
			if incognitokey.IndexOfCommitteeKey(committeePublicKey, v) > -1 {
				committeeChange.shardCommitteeAdded[k] = append(committeeChange.shardCommitteeAdded[k], committeePublicKey)
				flag = true
				break
			}
		}
		if flag {
			continue
		}
		for k, v := range beaconBestState.ShardPendingValidator {
			if incognitokey.IndexOfCommitteeKey(committeePublicKey, v) > -1 {
				committeeChange.shardSubstituteAdded[k] = append(committeeChange.shardSubstituteAdded[k], committeePublicKey)
				flag = true
				break
			}
		}
		if flag {
			continue
		}
	}
	return nil
}

func (blockchain *BlockChain) processStoreBeaconBlock(beaconBlock *BeaconBlock, snapshotBeaconCommittees []incognitokey.CommitteePublicKey, snapshotAllShardCommittees map[byte][]incognitokey.CommitteePublicKey, snapshotRewardReceivers map[string]string, committeeChange *committeeChange) error {
	Logger.log.Infof("BEACON | Process Store Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, beaconBlock.Header.Hash())
	var err error
	blockHash := beaconBlock.Header.Hash()
	blockHeight := beaconBlock.Header.Height
	tempBeaconBestState := blockchain.BestState.Beacon
	//statedb===========================START
	// Added
	err = statedb.StoreCurrentEpochShardCandidate(tempBeaconBestState.consensusStateDB, committeeChange.currentEpochShardCandidateAdded, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreNextEpochShardCandidate(tempBeaconBestState.consensusStateDB, committeeChange.nextEpochShardCandidateAdded, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreCurrentEpochBeaconCandidate(tempBeaconBestState.consensusStateDB, committeeChange.currentEpochBeaconCandidateAdded, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreNextEpochBeaconCandidate(tempBeaconBestState.consensusStateDB, committeeChange.nextEpochBeaconCandidateAdded, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreAllShardSubstitutesValidator(tempBeaconBestState.consensusStateDB, committeeChange.shardSubstituteAdded, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreAllShardCommittee(tempBeaconBestState.consensusStateDB, committeeChange.shardCommitteeAdded, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.ReplaceAllShardCommittee(tempBeaconBestState.consensusStateDB, committeeChange.shardCommitteeReplaced, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreBeaconSubstituteValidator(tempBeaconBestState.consensusStateDB, committeeChange.beaconSubstituteAdded, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.StoreBeaconCommittee(tempBeaconBestState.consensusStateDB, committeeChange.beaconCommitteeAdded, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	err = statedb.ReplaceBeaconCommittee(tempBeaconBestState.consensusStateDB, committeeChange.beaconCommitteeReplaced, tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return err
	}
	// Deleted
	err = statedb.DeleteCurrentEpochShardCandidate(tempBeaconBestState.consensusStateDB, committeeChange.currentEpochShardCandidateRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteNextEpochShardCandidate(tempBeaconBestState.consensusStateDB, committeeChange.nextEpochShardCandidateRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteCurrentEpochBeaconCandidate(tempBeaconBestState.consensusStateDB, committeeChange.currentEpochBeaconCandidateRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteNextEpochBeaconCandidate(tempBeaconBestState.consensusStateDB, committeeChange.nextEpochBeaconCandidateRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteAllShardSubstitutesValidator(tempBeaconBestState.consensusStateDB, committeeChange.shardSubstituteRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteAllShardCommittee(tempBeaconBestState.consensusStateDB, committeeChange.shardCommitteeRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteBeaconSubstituteValidator(tempBeaconBestState.consensusStateDB, committeeChange.beaconSubstituteRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteBeaconCommittee(tempBeaconBestState.consensusStateDB, committeeChange.beaconCommitteeRemoved)
	if err != nil {
		return err
	}
	// Remove shard reward request of old epoch
	// this value is no longer needed because, old epoch reward has been split and send to shard
	if beaconBlock.Header.Height%blockchain.config.ChainParams.Epoch == 2 {
		statedb.RemoveRewardOfShardByEpoch(tempBeaconBestState.rewardStateDB, beaconBlock.Header.Epoch-1)
	}
	err = blockchain.addShardRewardRequestToBeacon(beaconBlock, tempBeaconBestState.rewardStateDB)
	if err != nil {
		return NewBlockChainError(UpdateDatabaseWithBlockRewardInfoError, err)
	}
	// execute, store
	err = blockchain.processBridgeInstructions(tempBeaconBestState.featureStateDB, beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessBridgeInstructionError, err)
	}
	// execute, store
	err = blockchain.processPDEInstructions(tempBeaconBestState.featureStateDB, beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessPDEInstructionError, err)
	}

	// execute, store
	err = blockchain.processPortalInstructions(tempBeaconBestState.featureStateDB, beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessPortalInstructionError, err)
	}

	// execute, store
	err = blockchain.processRelayingInstructions(beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessPortalRelayingError, err)
	}

	consensusRootHash, err := tempBeaconBestState.consensusStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = tempBeaconBestState.consensusStateDB.Database().TrieDB().Commit(consensusRootHash, false)
	if err != nil {
		return err
	}
	featureRootHash, err := tempBeaconBestState.featureStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = tempBeaconBestState.featureStateDB.Database().TrieDB().Commit(featureRootHash, false)
	if err != nil {
		return err
	}
	rewardRootHash, err := tempBeaconBestState.rewardStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = tempBeaconBestState.rewardStateDB.Database().TrieDB().Commit(rewardRootHash, false)
	if err != nil {
		return err
	}
	slashRootHash, err := tempBeaconBestState.slashStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = tempBeaconBestState.slashStateDB.Database().TrieDB().Commit(slashRootHash, false)
	if err != nil {
		return err
	}
	tempBeaconBestState.consensusStateDB.ClearObjects()
	tempBeaconBestState.rewardStateDB.ClearObjects()
	tempBeaconBestState.featureStateDB.ClearObjects()
	tempBeaconBestState.slashStateDB.ClearObjects()
	//statedb===========================END
	//================================Store cross shard state ==================================
	if beaconBlock.Body.ShardState != nil {
		//GetBeaconBestState().lock.Lock()
		lastCrossShardState := tempBeaconBestState.LastCrossShardState
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
						//GetBeaconBestState().lock.Unlock()
						return NewBlockChainError(StoreCrossShardNextHeightError, err)
					}
					//beacon process shard_to_beacon in order so cross shard next height also will be saved in order
					//dont care overwrite this value
					err = rawdbv2.StoreCrossShardNextHeight(blockchain.GetDatabase(), fromShard, toShard, waitHeight, 0)
					if err != nil {
						//GetBeaconBestState().lock.Unlock()
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
		//GetBeaconBestState().lock.Unlock()
	}
	//=============================END Store cross shard state ==================================
	batch := blockchain.GetDatabase().NewBatch()
	//State Root Hash
	if err := rawdbv2.StoreConsensusStateRootHash(batch, blockHeight, consensusRootHash); err != nil {
		return NewBlockChainError(StoreBeaconBlockError, err)
	}
	if err := rawdbv2.StoreRewardStateRootHash(batch, blockHeight, rewardRootHash); err != nil {
		return NewBlockChainError(StoreBeaconBlockError, err)
	}
	if err := rawdbv2.StoreFeatureStateRootHash(batch, blockHeight, featureRootHash); err != nil {
		return NewBlockChainError(StoreBeaconBlockError, err)
	}
	if err := rawdbv2.StoreSlashStateRootHash(batch, blockHeight, slashRootHash); err != nil {
		return NewBlockChainError(StoreBeaconBlockError, err)
	}
	if err := rawdbv2.StoreBeaconBlockIndex(batch, blockHeight, blockHash); err != nil {
		return NewBlockChainError(StoreBeaconBlockIndexError, err)
	}
	Logger.log.Debugf("Store Beacon BestState Height %+v", blockHeight)
	beaconBestStateBytes, err := json.Marshal(tempBeaconBestState)
	if err != nil {
		return NewBlockChainError(StoreBeaconBestStateError, err)
	}
	if err := rawdbv2.StoreBeaconBestState(batch, beaconBestStateBytes); err != nil {
		return NewBlockChainError(StoreBeaconBestStateError, err)
	}
	Logger.log.Debugf("Store Beacon Block Height %+v with Hash %+v ", blockHeight, blockHash)
	if err := rawdbv2.StoreBeaconBlock(batch, blockHeight, blockHash, beaconBlock); err != nil {
		return NewBlockChainError(StoreBeaconBlockError, err)
	}
	if err := batch.Write(); err != nil {
		return NewBlockChainError(StoreBeaconBlockError, err)
	}
	return nil
}
