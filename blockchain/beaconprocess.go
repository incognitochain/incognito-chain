package blockchain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/instruction"
	"reflect"
	"sort"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/btc"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
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
	//get view that block link to
	preHash := beaconBlock.Header.PreviousBlockHash
	view := blockchain.BeaconChain.GetViewByHash(preHash)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if view == nil {
		blockchain.config.Syncker.SyncMissingBeaconBlock(ctx, "", preHash)
	}

	for {
		select {
		case <-ctx.Done():
			return errors.New(fmt.Sprintf("BeaconBlock %v link to wrong view (%s)", beaconBlock.GetHeight(), preHash.String()))
		default:
			view = blockchain.BeaconChain.GetViewByHash(preHash)
			if view != nil {
				goto CONTINUE_VERIFY
			}
			time.Sleep(time.Second)
		}
	}
CONTINUE_VERIFY:

	curView := view.(*BeaconBestState)
	// Verify block only
	Logger.log.Infof("BEACON | Verify block for signing process %d, with hash %+v", beaconBlock.Header.Height, *beaconBlock.Hash())
	if err := blockchain.verifyPreProcessingBeaconBlock(curView, beaconBlock, isPreSign); err != nil {
		return err
	}
	// Verify block with previous best state
	// not verify agg signature in this function
	if err := curView.verifyBestStateWithBeaconBlock(blockchain, beaconBlock, false, blockchain.config.ChainParams.Epoch); err != nil {
		return err
	}
	// Update best state with new block
	newBestState, hashes, _, err := curView.updateBeaconBestState(beaconBlock, blockchain)
	if err != nil {
		return err
	}
	// Post verififcation: verify new beaconstate with corresponding block
	if err := newBestState.verifyPostProcessingBeaconBlock(beaconBlock, blockchain.config.RandomClient, hashes); err != nil {
		return err
	}
	Logger.log.Infof("BEACON | Block %d, with hash %+v is VALID to be 🖊 signed", beaconBlock.Header.Height, *beaconBlock.Hash())
	return nil
}

// var bcTmp time.Duration
// var bcStart time.Time
// var bcAllTime time.Duration

func (blockchain *BlockChain) InsertBeaconBlock(beaconBlock *BeaconBlock, shouldValidate bool) error {
	blockHash := beaconBlock.Hash()
	Logger.log.Infof("BEACON | InsertBeaconBlock  %+v with hash %+v", beaconBlock.Header.Height, blockHash.String())
	// if beaconBlock.GetHeight() == 2 {
	// 	bcTmp = 0
	// 	bcStart = time.Now()
	// 	bcAllTime = time.Since(bcStart)
	// }
	// defer func(h uint64) {
	// 	bcAllTime = time.Since(bcStart)
	// 	if h%1000 == 0 {
	// 		bcTmp = bcAllTime - bcTmp
	// 		Logger.log.Infof("[BenchmarkBeacon] Time for insert from 2->%v: %v, Avg: %v", h, bcAllTime, bcAllTime.Seconds()/float64(h-2+1))
	// 		Logger.log.Infof("[BenchmarkBeacon] Time for insert 1000 blks [%v-%v]: %v; Avg: %v", h-1000+1, h, bcTmp, bcTmp.Seconds()/1000)
	// 	}
	// }(beaconBlock.GetHeight())
	blockchain.BeaconChain.insertLock.Lock()
	defer blockchain.BeaconChain.insertLock.Unlock()
	startTimeStoreBeaconBlock := time.Now()
	//check view if exited
	checkView := blockchain.BeaconChain.GetViewByHash(*beaconBlock.Hash())
	if checkView != nil {
		return nil
	}

	//get view that block link to
	preHash := beaconBlock.Header.PreviousBlockHash
	preView := blockchain.BeaconChain.GetViewByHash(preHash)
	if preView == nil {
		return errors.New(fmt.Sprintf("BeaconBlock %v link to wrong view (%s)", beaconBlock.GetHeight(), preHash.String()))
	}
	curView := preView.(*BeaconBestState)

	if beaconBlock.Header.Height != curView.BeaconHeight+1 {
		return errors.New("Not expected height")
	}

	Logger.log.Debugf("BEACON | Begin Insert new Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	if shouldValidate {
		Logger.log.Debugf("BEACON | Verify Pre Processing, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
		if err := blockchain.verifyPreProcessingBeaconBlock(curView, beaconBlock, false); err != nil {
			return err
		}
	} else {
		Logger.log.Debugf("BEACON | SKIP Verify Pre Processing, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	}
	// Verify beaconBlock with previous best state
	if shouldValidate {
		Logger.log.Debugf("BEACON | Verify Best State With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
		// Verify beaconBlock with previous best state
		if err := curView.verifyBestStateWithBeaconBlock(blockchain, beaconBlock, true, blockchain.config.ChainParams.Epoch); err != nil {
			return err
		}
		beaconCommittee := curView.GetBeaconCommittee()
		if err := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(beaconBlock, beaconCommittee); err != nil {
			return err
		}
	} else {
		Logger.log.Debugf("BEACON | SKIP Verify Best State With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	}

	// Backup beststate
	err := rawdbv2.CleanUpPreviousBeaconBestState(blockchain.GetBeaconChainDatabase())
	if err != nil {
		return NewBlockChainError(CleanBackUpError, err)
	}

	// process for slashing, make sure this one is called before update best state
	// since we'd like to process with old committee, not updated committee
	slashErr := blockchain.processForSlashing(curView.slashStateDB, beaconBlock)
	if slashErr != nil {
		Logger.log.Errorf("Failed to process slashing with error: %+v", NewBlockChainError(ProcessSlashingError, slashErr))
	}
	Logger.log.Debugf("BEACON | Update BestState With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// Update best state with new beaconBlock
	newBestState, hashes, committeeChange, err := curView.updateBeaconBestState(beaconBlock, blockchain)
	if err != nil {
		curView.beaconCommitteeEngine.AbortUncommittedBeaconState()
		return err
	}
	var err2 error
	defer func() {
		if err2 != nil {
			newBestState.beaconCommitteeEngine.AbortUncommittedBeaconState()
		}
	}()
	// updateNumOfBlocksByProducers updates number of blocks produced by producers
	newBestState.updateNumOfBlocksByProducers(beaconBlock, blockchain.config.ChainParams.Epoch)
	if shouldValidate {
		Logger.log.Debugf("BEACON | Verify Post Processing Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
		// Post verification: verify new beacon best state with corresponding beacon block
		if err2 = newBestState.verifyPostProcessingBeaconBlock(beaconBlock, blockchain.config.RandomClient, hashes); err2 != nil {
			return err2
		}
	} else {
		Logger.log.Debugf("BEACON | SKIP Verify Post Processing Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	}
	Logger.log.Infof("BEACON | Update Committee State Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	if err2 := newBestState.beaconCommitteeEngine.Commit(hashes); err2 != nil {
		return err2
	}
	Logger.log.Infof("BEACON | Process Store Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	if err2 := blockchain.processStoreBeaconBlock(newBestState, beaconBlock, committeeChange); err2 != nil {
		return err2
	}
	Logger.log.Infof("BEACON | Finish Insert new Beacon Block %+v, with hash %+v", beaconBlock.Header.Height, *beaconBlock.Hash())
	if beaconBlock.Header.Height%50 == 0 {
		BLogger.log.Debugf("Inserted beacon height: %d", beaconBlock.Header.Height)
	}
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewBeaconBlockTopic, beaconBlock))
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.BeaconBeststateTopic, newBestState))
	// For masternode: broadcast new committee to highways
	beaconInsertBlockTimer.UpdateSince(startTimeStoreBeaconBlock)
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

/*
	VerifyPreProcessingBeaconBlock
	This function DOES NOT verify new block with best state
	DO NOT USE THIS with GENESIS BLOCK
	- Producer sanity data
	- Version: compatible with predefined version
	- Previous Block exist in database, fetch previous block by previous hash of new beacon block
	- Check new beacon block height is equal to previous block height + 1
	- Epoch = blockHeight % Epoch == 1 ? Previous Block Epoch + 1 : Previous Block Epoch
	- Timestamp of new beacon block is greater than previous beacon block timestamp
	- ShardStateHash: rebuild shard state hash from shard state body and compare with shard state hash in block header
	- InstructionHash: rebuild instruction hash from instruction body and compare with instruction hash in block header
	- InstructionMerkleRoot: rebuild instruction merkle root from instruction body and compare with instruction merkle root in block header
	- If verify block for signing then verifyPreProcessingBeaconBlockForSigning
*/
func (blockchain *BlockChain) verifyPreProcessingBeaconBlock(curView *BeaconBestState, beaconBlock *BeaconBlock, isPreSign bool) error {
	// if len(beaconBlock.Header.Producer) == 0 {
	// 	return NewBlockChainError(ProducerError, fmt.Errorf("Expect has length 66 but get %+v", len(beaconBlock.Header.Producer)))
	// }
	startTimeVerifyPreProcessingBeaconBlock := time.Now()
	// Verify parent hash exist or not
	previousBlockHash := beaconBlock.Header.PreviousBlockHash
	parentBlockBytes, err := rawdbv2.GetBeaconBlockByHash(blockchain.GetBeaconChainDatabase(), previousBlockHash)
	if err != nil {
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
	beaconVerifyPreprocesingTimer.UpdateSince(startTimeVerifyPreProcessingBeaconBlock)
	if isPreSign {
		if err := blockchain.verifyPreProcessingBeaconBlockForSigning(curView, beaconBlock); err != nil {
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
func (blockchain *BlockChain) verifyPreProcessingBeaconBlockForSigning(curView *BeaconBestState, beaconBlock *BeaconBlock) error {
	var err error
	startTimeVerifyPreProcessingBeaconBlockForSigning := time.Now()
	rewardByEpochInstruction := [][]string{}
	tempShardStates := make(map[byte][]ShardState)
	stakeInstructions := []*instruction.StakeInstruction{}
	validStakePublicKeys := []string{}
	swapInstructions := make(map[byte][]*instruction.SwapInstruction)
	bridgeInstructions := [][]string{}
	acceptedBlockRewardInstructions := [][]string{}
	stopAutoStakeInstructions := []*instruction.StopAutoStakeInstruction{}
	statefulActionsByShardID := map[byte][][]string{}
	rewardForCustodianByEpoch := map[common.Hash]uint64{}

	portalParams := blockchain.GetPortalParams(beaconBlock.GetHeight())

	// Get Reward Instruction By Epoch
	if beaconBlock.Header.Height%blockchain.config.ChainParams.Epoch == 1 {
		featureStateDB := curView.GetBeaconFeatureStateDB()
		totalLockedCollateral, err := getTotalLockedCollateralInEpoch(featureStateDB)
		if err != nil {
			return NewBlockChainError(GetTotalLockedCollateralError, err)
		}
		isSplitRewardForCustodian := totalLockedCollateral > 0
		percentCustodianRewards := portalParams.MaxPercentCustodianRewards
		if totalLockedCollateral < portalParams.MinLockCollateralAmountInEpoch {
			percentCustodianRewards = portalParams.MinPercentCustodianRewards
		}

		rewardByEpochInstruction, rewardForCustodianByEpoch, err = blockchain.buildRewardInstructionByEpoch(curView, beaconBlock.Header.Height, beaconBlock.Header.Epoch-1, curView.GetBeaconRewardStateDB(), isSplitRewardForCustodian, percentCustodianRewards)
		if err != nil {
			return NewBlockChainError(BuildRewardInstructionError, err)
		}
	}
	// get shard to beacon blocks from pool
	allRequiredShardBlockHeight := make(map[byte][]uint64)
	for shardID, shardstates := range beaconBlock.Body.ShardState {
		heights := []uint64{}
		for _, state := range shardstates {
			heights = append(heights, state.Height)
		}
		sort.Slice(heights, func(i, j int) bool {
			return heights[i] < heights[j]
		})
		allRequiredShardBlockHeight[shardID] = heights
	}
	allShardBlocks, err := blockchain.GetShardBlocksForBeaconValidator(allRequiredShardBlockHeight)
	if err != nil {
		return NewBlockChainError(GetShardBlocksForBeaconProcessError, fmt.Errorf("Unable to get required shard block for beacon process"))
	}
	keys := []int{}
	for shardID, shardBlocks := range allShardBlocks {
		strs := fmt.Sprintf("GetShardState shardID: %+v, Height", shardID)
		for _, shardBlock := range shardBlocks {
			strs += fmt.Sprintf(" %d", shardBlock.Header.Height)
		}
		Logger.log.Info(strs)
		keys = append(keys, int(shardID))
	}
	sort.Ints(keys)
	for _, v := range keys {
		shardID := byte(v)
		shardBlocks := allShardBlocks[shardID]
		shardStates := beaconBlock.Body.ShardState[shardID]
		// repeatly compare each shard to beacon block and shard state in new beacon block body
		if len(shardBlocks) >= len(shardStates) {
			shardBlocks = shardBlocks[:len(beaconBlock.Body.ShardState[shardID])]
			for _, shardBlock := range shardBlocks {
				tempShardState, stakeInstruction, tempValidStakePublicKeys, swapInstruction, bridgeInstruction, acceptedBlockRewardInstruction, stopAutoStakingInstruction, statefulActions := blockchain.GetShardStateFromBlock(curView, beaconBlock.Header.Height, shardBlock, shardID, false, validStakePublicKeys)
				tempShardStates[shardID] = append(tempShardStates[shardID], tempShardState[shardID])
				stakeInstructions = append(stakeInstructions, stakeInstruction...)
				swapInstructions[shardID] = append(swapInstructions[shardID], swapInstruction[shardID]...)
				bridgeInstructions = append(bridgeInstructions, bridgeInstruction...)
				acceptedBlockRewardInstructions = append(acceptedBlockRewardInstructions, acceptedBlockRewardInstruction)
				stopAutoStakeInstructions = append(stopAutoStakeInstructions, stopAutoStakingInstruction...)
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
			return NewBlockChainError(GetShardBlocksForBeaconProcessError, fmt.Errorf("Expect to get more than %+v Shard Block but only get %+v (shard %v)", len(beaconBlock.Body.ShardState[shardID]), len(shardBlocks), shardID))
		}
	}
	// build stateful instructions
	statefulInsts := blockchain.buildStatefulInstructions(curView.featureStateDB, statefulActionsByShardID, beaconBlock.Header.Height, rewardForCustodianByEpoch, portalParams)
	bridgeInstructions = append(bridgeInstructions, statefulInsts...)

	tempInstruction, err := curView.GenerateInstruction(beaconBlock.Header.Height,
		stakeInstructions, swapInstructions, stopAutoStakeInstructions,
		curView.GetCandidateShardWaitingForCurrentRandom(),
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
	beaconVerifyPreprocesingForPreSignTimer.UpdateSince(startTimeVerifyPreProcessingBeaconBlockForSigning)
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
func (beaconBestState *BeaconBestState) verifyBestStateWithBeaconBlock(blockchain *BlockChain, beaconBlock *BeaconBlock, isVerifySig bool, chainParamEpoch uint64) error {
	//verify producer via index
	startTimeVerifyWithBestState := time.Now()
	if err := blockchain.config.ConsensusEngine.ValidateProducerPosition(beaconBlock, beaconBestState.BeaconProposerIndex, beaconBestState.GetBeaconCommittee(), beaconBestState.MinBeaconCommitteeSize); err != nil {
		return err
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
					return NewBlockChainError(BeaconBestStateBestShardHeightNotCompatibleError, fmt.Errorf("Expect Shard %+v has state greater than to %+v but get %+v", shardID, bestShardHeight, shardStates[0].Height))
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
	beaconVerifyWithBestStateTimer.UpdateSince(startTimeVerifyWithBestState)
	return nil
}

//  verifyPostProcessingBeaconBlock verify block after update beacon best state
//  - Validator root: BeaconCommittee + BeaconPendingValidator
//  - Beacon Candidate root: CandidateBeaconWaitingForCurrentRandom + CandidateBeaconWaitingForNextRandom
//  - Shard Candidate root: CandidateShardWaitingForCurrentRandom + CandidateShardWaitingForNextRandom
//  - Shard Validator root: ShardCommittee + ShardPendingValidator
//  - Random number if have in instruction
func (beaconBestState *BeaconBestState) verifyPostProcessingBeaconBlock(beaconBlock *BeaconBlock, randomClient btc.RandomClient, hashes *committeestate.BeaconCommitteeStateHash) error {
	startTimeVerifyPostProcessingBeaconBlock := time.Now()
	if !hashes.BeaconCommitteeAndValidatorHash.IsEqual(&beaconBlock.Header.BeaconCommitteeAndValidatorRoot) {
		return NewBlockChainError(BeaconCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect %+v but get %+v", beaconBlock.Header.BeaconCommitteeAndValidatorRoot, hashes.BeaconCommitteeAndValidatorHash))
	}
	if !hashes.BeaconCandidateHash.IsEqual(&beaconBlock.Header.BeaconCandidateRoot) {
		return NewBlockChainError(BeaconCandidateRootError, fmt.Errorf("Expect %+v but get %+v", beaconBlock.Header.BeaconCandidateRoot, hashes.BeaconCandidateHash))
	}
	if !hashes.ShardCandidateHash.IsEqual(&beaconBlock.Header.ShardCandidateRoot) {
		return NewBlockChainError(ShardCandidateRootError, fmt.Errorf("Expect %+v but get %+v", beaconBlock.Header.ShardCandidateRoot, hashes.ShardCandidateHash))
	}
	if !hashes.ShardCommitteeAndValidatorHash.IsEqual(&beaconBlock.Header.ShardCommitteeAndValidatorRoot) {
		return NewBlockChainError(ShardCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect %+v but get %+v", beaconBlock.Header.ShardCommitteeAndValidatorRoot, hashes.ShardCommitteeAndValidatorHash))
	}
	if !hashes.AutoStakeHash.IsEqual(&beaconBlock.Header.AutoStakingRoot) {
		return NewBlockChainError(ShardCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect %+v but get %+v", beaconBlock.Header.AutoStakingRoot, hashes.AutoStakeHash))
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
				ok, err := randomClient.VerifyNonceWithTimestamp(startTime, beaconBestState.BlockMaxCreateTime, beaconBestState.CurrentRandomTimeStamp, int64(nonce))
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
	beaconVerifyPostProcessingTimer.UpdateSince(startTimeVerifyPostProcessingBeaconBlock)
	return nil
}

/*
	Update Beststate with new Block
*/
func (oldBestState *BeaconBestState) updateBeaconBestState(beaconBlock *BeaconBlock, blockchain *BlockChain) (
	*BeaconBestState, *committeestate.BeaconCommitteeStateHash, *committeestate.CommitteeChange, error) {
	startTimeUpdateBeaconBestState := time.Now()
	beaconBestState := NewBeaconBestState()
	if err := beaconBestState.cloneBeaconBestStateFrom(oldBestState); err != nil {
		return nil, nil, nil, err
	}
	var chainParamEpoch = blockchain.config.ChainParams.Epoch
	var randomTime = blockchain.config.ChainParams.RandomTime
	var isBeginRandom = false
	var isFoundRandomInstruction = false
	Logger.log.Debugf("Start processing new block at height %d, with hash %+v", beaconBlock.Header.Height, *beaconBlock.Hash())
	// signal of random parameter from beacon block
	// update BestShardHash, BestBlock, BestBlockHash
	beaconBestState.PreviousBestBlockHash = beaconBestState.BestBlockHash
	beaconBestState.BestBlockHash = *beaconBlock.Hash()
	beaconBestState.BestBlock = *beaconBlock
	beaconBestState.Epoch = beaconBlock.Header.Epoch
	beaconBestState.BeaconHeight = beaconBlock.Header.Height
	if beaconBlock.Header.Height == 1 {
		beaconBestState.BeaconProposerIndex = 0
	} else {
		for i, v := range oldBestState.GetBeaconCommittee() {
			b58Str, _ := v.ToBase58()
			if b58Str == beaconBlock.Header.Producer {
				beaconBestState.BeaconProposerIndex = i
				break
			}
		}
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
	for _, inst := range beaconBlock.Body.Instructions {
		if inst[0] == instruction.RANDOM_ACTION {
			if err := instruction.ValidateRandomInstructionSanity(inst); err != nil {
				return nil, nil, nil, NewBlockChainError(ProcessRandomInstructionError, err)
			}
			randomInstruction := instruction.ImportRandomInstructionFromString(inst)
			beaconBestState.CurrentRandomNumber = randomInstruction.BtcNonce
			beaconBestState.IsGetRandomNumber = true
			isFoundRandomInstruction = true
			Logger.log.Infof("Random number found %d", beaconBestState.CurrentRandomNumber)
		}
	}
	if beaconBestState.BeaconHeight%chainParamEpoch == 1 && beaconBestState.BeaconHeight != 1 {
		// Begin of each epoch
		beaconBestState.IsGetRandomNumber = false
		// Before get random from bitcoin
	} else if beaconBestState.BeaconHeight%chainParamEpoch >= randomTime {
		if beaconBestState.BeaconHeight%chainParamEpoch == randomTime {
			beaconBestState.CurrentRandomTimeStamp = beaconBlock.Header.Timestamp
			isBeginRandom = true
		}
	}
	env := beaconBestState.NewBeaconCommitteeStateEnvironment(blockchain.config.ChainParams, beaconBlock.Body.Instructions, isFoundRandomInstruction, isBeginRandom)
	hashes, committeeChange, err := beaconBestState.beaconCommitteeEngine.UpdateCommitteeState(env)
	if err != nil {
		return nil, nil, nil, NewBlockChainError(UpdateBeaconCommitteeStateError, err)
	}
	beaconBestState.updateNumOfBlocksByProducers(beaconBlock, chainParamEpoch)
	beaconUpdateBestStateTimer.UpdateSince(startTimeUpdateBeaconBestState)
	return beaconBestState, hashes, committeeChange, nil
}

func (beaconBestState *BeaconBestState) initBeaconBestState(genesisBeaconBlock *BeaconBlock, blockchain *BlockChain, db incdb.Database) error {
	Logger.log.Info("Process Update Beacon Best State With Beacon Genesis Block")
	beaconBestState.PreviousBestBlockHash = beaconBestState.BestBlockHash
	beaconBestState.BestBlockHash = *genesisBeaconBlock.Hash()
	beaconBestState.BestBlock = *genesisBeaconBlock
	beaconBestState.Epoch = genesisBeaconBlock.Header.Epoch
	beaconBestState.BeaconHeight = genesisBeaconBlock.Header.Height
	beaconBestState.BeaconProposerIndex = 0
	beaconBestState.BestShardHash = make(map[byte]common.Hash)
	beaconBestState.BestShardHeight = make(map[byte]uint64)
	for i := 0; i < beaconBestState.ActiveShards; i++ {
		shardID := byte(i)
		beaconBestState.BestShardHeight[shardID] = 0
	}
	// Update new best new block hash
	for shardID, shardStates := range genesisBeaconBlock.Body.ShardState {
		beaconBestState.BestShardHash[shardID] = shardStates[len(shardStates)-1].Hash
		beaconBestState.BestShardHeight[shardID] = shardStates[len(shardStates)-1].Height
	}
	// update param
	beaconBestState.ConsensusAlgorithm = common.BlsConsensus
	beaconBestState.ShardConsensusAlgorithm = make(map[byte]string)
	for shardID := 0; shardID < beaconBestState.ActiveShards; shardID++ {
		beaconBestState.ShardConsensusAlgorithm[byte(shardID)] = common.BlsConsensus
	}
	beaconBestState.beaconCommitteeEngine.InitCommitteeState(beaconBestState.NewBeaconCommitteeStateEnvironment(blockchain.config.ChainParams, genesisBeaconBlock.Body.Instructions, false, false))
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
	beaconBestState.ConsensusStateDBRootHash = common.EmptyRoot
	beaconBestState.SlashStateDBRootHash = common.EmptyRoot
	beaconBestState.RewardStateDBRootHash = common.EmptyRoot
	beaconBestState.FeatureStateDBRootHash = common.EmptyRoot
	if err != nil {
		return err
	}
	//statedb===========================END
	return nil
}

func (blockchain *BlockChain) processStoreBeaconBlock(
	newBestState *BeaconBestState,
	beaconBlock *BeaconBlock,
	committeeChange *committeestate.CommitteeChange,
) error {
	startTimeProcessStoreBeaconBlock := time.Now()
	Logger.log.Debugf("BEACON | Process Store Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, beaconBlock.Header.Hash())
	blockHash := beaconBlock.Header.Hash()
	blockHeight := beaconBlock.Header.Height

	var err error
	//statedb===========================START
	// Added
	err = statedb.StoreCurrentEpochShardCandidate(newBestState.consensusStateDB, committeeChange.CurrentEpochShardCandidateAdded)
	if err != nil {
		return err
	}
	err = statedb.StoreNextEpochShardCandidate(newBestState.consensusStateDB, committeeChange.NextEpochShardCandidateAdded, newBestState.GetRewardReceiver(), newBestState.GetAutoStaking(), newBestState.GetStakingTx())
	if err != nil {
		return err
	}
	err = statedb.StoreCurrentEpochBeaconCandidate(newBestState.consensusStateDB, committeeChange.CurrentEpochBeaconCandidateAdded)
	if err != nil {
		return err
	}
	err = statedb.StoreNextEpochBeaconCandidate(newBestState.consensusStateDB, committeeChange.NextEpochBeaconCandidateAdded, newBestState.GetRewardReceiver(), newBestState.GetAutoStaking(), newBestState.GetStakingTx())
	if err != nil {
		return err
	}
	err = statedb.StoreAllShardSubstitutesValidator(newBestState.consensusStateDB, committeeChange.ShardSubstituteAdded)
	if err != nil {
		return err
	}
	err = statedb.StoreAllShardCommittee(newBestState.consensusStateDB, committeeChange.ShardCommitteeAdded)
	if err != nil {
		return err
	}
	err = statedb.ReplaceAllShardCommittee(newBestState.consensusStateDB, committeeChange.ShardCommitteeReplaced)
	if err != nil {
		return err
	}
	err = statedb.StoreBeaconSubstituteValidator(newBestState.consensusStateDB, committeeChange.BeaconSubstituteAdded)
	if err != nil {
		return err
	}
	err = statedb.StoreBeaconCommittee(newBestState.consensusStateDB, committeeChange.BeaconCommitteeAdded)
	if err != nil {
		return err
	}
	err = statedb.ReplaceBeaconCommittee(newBestState.consensusStateDB, committeeChange.BeaconCommitteeReplaced)
	if err != nil {
		return err
	}
	// Deleted
	err = statedb.DeleteCurrentEpochShardCandidate(newBestState.consensusStateDB, committeeChange.CurrentEpochShardCandidateRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteNextEpochShardCandidate(newBestState.consensusStateDB, committeeChange.NextEpochShardCandidateRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteCurrentEpochBeaconCandidate(newBestState.consensusStateDB, committeeChange.CurrentEpochBeaconCandidateRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteNextEpochBeaconCandidate(newBestState.consensusStateDB, committeeChange.NextEpochBeaconCandidateRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteAllShardSubstitutesValidator(newBestState.consensusStateDB, committeeChange.ShardSubstituteRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteAllShardCommittee(newBestState.consensusStateDB, committeeChange.ShardCommitteeRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteBeaconSubstituteValidator(newBestState.consensusStateDB, committeeChange.BeaconSubstituteRemoved)
	if err != nil {
		return err
	}
	err = statedb.DeleteBeaconCommittee(newBestState.consensusStateDB, committeeChange.BeaconCommitteeRemoved)
	if err != nil {
		return err
	}

	blockchain.processForSlashing(newBestState.slashStateDB, beaconBlock)

	// Remove shard reward request of old epoch
	// this value is no longer needed because, old epoch reward has been split and send to shard
	if beaconBlock.Header.Height%blockchain.config.ChainParams.Epoch == 2 {
		statedb.RemoveRewardOfShardByEpoch(newBestState.rewardStateDB, beaconBlock.Header.Epoch-1)
	}
	err = blockchain.addShardRewardRequestToBeacon(beaconBlock, newBestState.rewardStateDB)
	if err != nil {
		return NewBlockChainError(UpdateDatabaseWithBlockRewardInfoError, err)
	}
	// execute, store
	err = blockchain.processBridgeInstructions(newBestState.featureStateDB, beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessBridgeInstructionError, err)
	}
	// execute, store PDE instruction
	err = blockchain.processPDEInstructions(newBestState.featureStateDB, beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessPDEInstructionError, err)
	}

	// execute, store Portal Instruction
	//if (blockchain.config.ChainParams.Net == Mainnet) || (blockchain.config.ChainParams.Net == Testnet && beaconBlock.Header.Height > 1500000) {
	err = blockchain.processPortalInstructions(newBestState.featureStateDB, beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessPortalInstructionError, err)
	}
	//}

	// execute, store Ralaying Instruction
	err = blockchain.processRelayingInstructions(beaconBlock)
	if err != nil {
		return NewBlockChainError(ProcessPortalRelayingError, err)
	}

	//store beacon block hash by index to consensus state db => mark this block hash is for this view at this height
	if err := statedb.StoreBeaconBlockHashByIndex(newBestState.consensusStateDB, blockHeight, blockHash); err != nil {
		return err
	}

	consensusRootHash, err := newBestState.consensusStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = newBestState.consensusStateDB.Database().TrieDB().Commit(consensusRootHash, false)
	if err != nil {
		return err
	}

	newBestState.ConsensusStateDBRootHash = consensusRootHash
	featureRootHash, err := newBestState.featureStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = newBestState.featureStateDB.Database().TrieDB().Commit(featureRootHash, false)
	if err != nil {
		return err
	}
	newBestState.FeatureStateDBRootHash = featureRootHash
	rewardRootHash, err := newBestState.rewardStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = newBestState.rewardStateDB.Database().TrieDB().Commit(rewardRootHash, false)
	if err != nil {
		return err
	}
	newBestState.RewardStateDBRootHash = rewardRootHash
	slashRootHash, err := newBestState.slashStateDB.Commit(true)
	if err != nil {
		return err
	}
	err = newBestState.slashStateDB.Database().TrieDB().Commit(slashRootHash, false)
	if err != nil {
		return err
	}
	newBestState.SlashStateDBRootHash = slashRootHash
	newBestState.consensusStateDB.ClearObjects()
	newBestState.rewardStateDB.ClearObjects()
	newBestState.featureStateDB.ClearObjects()
	newBestState.slashStateDB.ClearObjects()
	//statedb===========================END

	batch := blockchain.GetBeaconChainDatabase().NewBatch()
	//State Root Hash
	bRH := BeaconRootHash{
		ConsensusStateDBRootHash: consensusRootHash,
		FeatureStateDBRootHash:   featureRootHash,
		RewardStateDBRootHash:    rewardRootHash,
		SlashStateDBRootHash:     slashRootHash,
	}

	if err := rawdbv2.StoreBeaconRootsHash(batch, blockHash, bRH); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}

	if err := rawdbv2.StoreBeaconBlock(batch, blockHeight, blockHash, beaconBlock); err != nil {
		return NewBlockChainError(StoreBeaconBlockError, err)
	}
	blockchain.BeaconChain.multiView.AddView(newBestState)
	err = blockchain.BackupBeaconViews(batch)
	if err != nil {
		panic("Backup shard view error")
	}
	if err := batch.Write(); err != nil {
		return NewBlockChainError(StoreBeaconBlockError, err)
	}
	beaconStoreBlockTimer.UpdateSince(startTimeProcessStoreBeaconBlock)

	//backup
	//if (newBestState.BeaconHeight+1)%blockchain.config.ChainParams.Epoch == 0 {
	//	blockchain.GetBeaconChainDatabase().Close()
	//	blockchain.GetBeaconChainDatabase().Backup(fmt.Sprintf("../../../backup/beacon/%d", newBestState.Epoch))
	//	blockchain.GetBeaconChainDatabase().ReOpen()
	//}

	return nil
}
