package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/metrics"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

/*
	// This function should receives block in consensus round
	// It verify validity of this function before sign it
	// This should be verify in the first round of consensus

	Step:
	1. Verify Pre proccessing data
	2. Retrieve beststate for new block, store in local variable
	3. Update: process local beststate with new block
	4. Verify Post processing: updated local beststate and newblock

	Return:
	- No error: valid and can be sign
	- Error: invalid new block
*/
func (blockchain *BlockChain) VerifyPreSignBeaconBlock(beaconBlock *BeaconBlock, isPreSign bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	//========Verify block only
	Logger.log.Infof("BEACON | Verify block for signing process %d, with hash %+v", beaconBlock.Header.Height, *beaconBlock.Hash())
	if err := blockchain.verifyPreProcessingBeaconBlock(beaconBlock, isPreSign); err != nil {
		return err
	}
	//========Verify block with previous best state
	// Get Beststate of previous block == previous best state
	// Clone best state value into new variable
	beaconBestState := NewBeaconBestState()
	if err := beaconBestState.cloneBeaconBestState(blockchain.BestState.Beacon); err != nil {
		return err
	}
	// Verify block with previous best state
	// not verify agg signature in this function
	if err := beaconBestState.verifyBestStateWithBeaconBlock(beaconBlock, false); err != nil {
		return err
	}
	//========Update best state with new block
	if err := beaconBestState.updateBeaconBestState(beaconBlock); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding block
	if err := beaconBestState.verifyPostProcessingBeaconBlock(beaconBlock); err != nil {
		return err
	}
	Logger.log.Infof("BEACON | Block %d, with hash %+v is VALID to be 🖊 signed", beaconBlock.Header.Height, *beaconBlock.Hash())
	return nil
}

func (blockchain *BlockChain) InsertBeaconBlock(beaconBlock *BeaconBlock, isValidated bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	blockHash := beaconBlock.Header.Hash()
	Logger.log.Infof("BEACON | Begin insert new Beacon Block height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	Logger.log.Infof("BEACON | Check Beacon Block existence before insert block height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	isExist, _ := blockchain.config.DataBase.HasBeaconBlock(beaconBlock.Header.Hash())
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
		if err := blockchain.BestState.Beacon.verifyBestStateWithBeaconBlock(beaconBlock, true); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("BEACON | SKIP Verify Best State With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	}
	// Backup beststate
	if blockchain.config.UserKeySet != nil {
		userRole, _ := blockchain.BestState.Beacon.GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyInBase58CheckEncode(), 0)
		if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
			err := blockchain.config.DataBase.CleanBackup(false, 0)
			if err != nil {
				return NewBlockChainError(CleanBackUpError, err)
			}
			err = blockchain.BackupCurrentBeaconState(beaconBlock)
			if err != nil {
				return NewBlockChainError(BackUpBestStateError, err)
			}
		}
	}
	Logger.log.Infof("BEACON | Update BestState With Beacon Block, Beacon Block Height %+v with hash %+v", beaconBlock.Header.Height, blockHash)
	// Update best state with new beaconBlock
	if err := blockchain.BestState.Beacon.updateBeaconBestState(beaconBlock); err != nil {
		return err
	}
	if !isValidated {
		Logger.log.Infof("BEACON | Verify Post Processing Beacon Block %+v \n", blockHash)
		//========Post verififcation: verify new beaconstate with corresponding beaconBlock
		if err := blockchain.BestState.Beacon.verifyPostProcessingBeaconBlock(beaconBlock); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("BEACON | SKIP Verify Post Processing Block %+v \n", blockHash)
	}
	if err := blockchain.processStoreBeaconBlock(beaconBlock); err != nil {
		Logger.log.Error(err)
		return err
	}
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.NumOfBlockInsertToChain,
		metrics.MeasurementValue: float64(1),
		metrics.Tag:              metrics.ShardIDTag,
		metrics.TagValue:         metrics.Beacon,
	})
	Logger.log.Infof("Finish Insert new Beacon Block %+v, with hash %+v \n", beaconBlock.Header.Height, *beaconBlock.Hash())
	if beaconBlock.Header.Height%50 == 0 {
		BLogger.log.Debugf("Inserted beacon height: %d", beaconBlock.Header.Height)
	}
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewBeaconBlockTopic, beaconBlock))
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.BeaconBeststateTopic, blockchain.BestState.Beacon))
	return nil
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
func (blockchain *BlockChain) verifyPreProcessingBeaconBlock(beaconBlock *BeaconBlock, isPreSign bool) error {
	blockchain.BestState.Beacon.lock.RLock()
	defer blockchain.BestState.Beacon.lock.RUnlock()
	if len(beaconBlock.Header.ProducerAddress.Bytes()) != 66 {
		return NewBlockChainError(ProducerError, fmt.Errorf("Expect %+v has length 66 but get %+v", len(beaconBlock.Header.ProducerAddress.Bytes())))
	}
	//verify version
	if beaconBlock.Header.Version != BEACON_BLOCK_VERSION {
		return NewBlockChainError(WrongVersionError, fmt.Errorf("Expect block version to be equal to %+v but get %+v", BEACON_BLOCK_VERSION, beaconBlock.Header.Version))
	}
	// Verify parent hash exist or not
	previousBlockHash := beaconBlock.Header.PreviousBlockHash
	parentBlockBytes, err := blockchain.config.DataBase.FetchBeaconBlock(previousBlockHash)
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
	if (beaconBlock.Header.Height != 1) && (beaconBlock.Header.Height%common.EPOCH == 1) && (previousBeaconBlock.Header.Epoch != beaconBlock.Header.Epoch-1) {
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
	if !verifyHashFromStringArray(tempInstructionArr, beaconBlock.Header.InstructionHash) {
		return NewBlockChainError(InstructionHashError, fmt.Errorf("Expect instruction hash to be %+v", beaconBlock.Header.InstructionHash))
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

/*
	verifyPreProcessingBeaconBlockForSigning
	Must pass these following condition:
	- Rebuild Reward By Epoch Instruction
	- Get All Shard To Beacon Block in Shard To Beacon Pool
	- For all Shard To Beacon Blocks in each Shard
		+ Compare all shard height of shard states in body and these Shard To Beacon Blocks (got from pool)
			* Must be have the same range of height
			* Compare CrossShardBitMap of each Shard To Beacon Block and Shard State in New Beacon Block Body
		+ After finish comparing these shard to beacon blocks with shard states in new beacon block body
			* Verifying Shard To Beacon Block Agg Signature
			* Only accept block in one epoch
		+ Get Instruction from these Shard To Beacon Blocks:
			* Stake Instruction
			* Swap Instruction
			* Bridge Instruction
			* Block Reward Instruction
		+ Generate Instruction Hash from all recently got instructions
		+ Compare just created Instruction Hash with Instruction Hash In Beacon Header
*/
func (blockchain *BlockChain) verifyPreProcessingBeaconBlockForSigning(beaconBlock *BeaconBlock) error {
	var err error
	rewardByEpochInstruction := [][]string{}
	tempShardStates := make(map[byte][]ShardState)
	stakeInstructions := [][]string{}
	swapInstructions := make(map[byte][][]string)
	bridgeInstructions := [][]string{}
	acceptedBlockRewardInstructions := [][]string{}
	// Get Reward Instruction By Epoch
	if beaconBlock.Header.Height%common.EPOCH == 1 {
		rewardByEpochInstruction, err = blockchain.BuildRewardInstructionByEpoch(beaconBlock.Header.Epoch - 1)
		if err != nil {
			return NewBlockChainError(BuildRewardInstructionError, err)
		}
	}
	// get shard to beacon blocks from pool
	allShardBlocks := blockchain.config.ShardToBeaconPool.GetValidBlock(nil)
	var keys []int
	for k := range allShardBlocks {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, value := range keys {
		shardID := byte(value)
		shardBlocks := allShardBlocks[shardID]
		// repeatly compare each shard to beacon block and shard state in new beacon block body
		if len(shardBlocks) >= len(beaconBlock.Body.ShardState[shardID]) {
			shardBlocks = shardBlocks[:len(beaconBlock.Body.ShardState[shardID])]
			shardStates := beaconBlock.Body.ShardState[shardID]
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
			for index, shardBlock := range shardBlocks {
				currentCommittee := blockchain.BestState.Beacon.GetAShardCommittee(shardID)
				currentPendingValidator := blockchain.BestState.Beacon.GetAShardPendingValidator(shardID)
				hash := shardBlock.Header.Hash()
				err := ValidateAggSignature(shardBlock.ValidatorsIndex, currentCommittee, shardBlock.AggregatedSig, shardBlock.R, &hash)
				if index == 0 && err != nil {
					currentCommittee, _, _, _, err = SwapValidator(currentPendingValidator, currentCommittee, blockchain.BestState.Beacon.MaxShardCommitteeSize, common.OFFSET)
					if err != nil {
						return NewBlockChainError(SwapValidatorError, fmt.Errorf("Failed to swap validator when try to verify shard to beacon block %+v, error %+v", shardBlock.Header.Height, err))
					}
					err = ValidateAggSignature(shardBlock.ValidatorsIndex, currentCommittee, shardBlock.AggregatedSig, shardBlock.R, &hash)
					if err != nil {
						return NewBlockChainError(SignatureError, fmt.Errorf("Failed to verify Signature of Shard To Beacon Block %+v, error %+v", shardBlock.Header.Height, err))
					}
				}
				if index != 0 && err != nil {
					return NewBlockChainError(ShardStateError, fmt.Errorf("Fail to verify with Shard To Beacon Block %+v, error %+v", shardBlock.Header.Height, err))
				}
			}
			for _, shardBlock := range shardBlocks {
				tempShardState, stakeInstruction, swapInstruction, bridgeInstruction, acceptedBlockRewardInstruction := blockchain.GetShardStateFromBlock(beaconBlock.Header.Height, shardBlock, shardID)
				tempShardStates[shardID] = append(tempShardStates[shardID], tempShardState[shardID])
				stakeInstructions = append(stakeInstructions, stakeInstruction...)
				swapInstructions[shardID] = append(swapInstructions[shardID], swapInstruction[shardID]...)
				bridgeInstructions = append(bridgeInstructions, bridgeInstruction...)
				acceptedBlockRewardInstructions = append(acceptedBlockRewardInstructions, acceptedBlockRewardInstruction)
			}
		} else {
			return NewBlockChainError(GetShardBlocksError, fmt.Errorf("Expect to get more than %+v ShardToBeaconBlock but only get %+v", len(beaconBlock.Body.ShardState[shardID]), len(shardBlocks)))
		}
	}
	tempInstruction := blockchain.BestState.Beacon.GenerateInstruction(beaconBlock.Header.Height, stakeInstructions, swapInstructions, blockchain.BestState.Beacon.CandidateShardWaitingForCurrentRandom, bridgeInstructions, acceptedBlockRewardInstructions)
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
		return NewBlockChainError(InstructionHashError, fmt.Errorf("Expect Instruction Hash in Beacon Header to be %+v, but get %+v", beaconBlock.Header.InstructionHash, tempInstructionHash))
	}
	return nil
}

/*
	This function will verify the validation of a block with some best state in cache or current best state
	Get beacon state of this block
	For example, new blockHeight is 91 then beacon state of this block must have height 90
	OR new block has previous has is beacon best block hash
	- Get producer via index and compare with producer address in beacon block header
	- Validate public key and signature sanity
	- Validate Agg Signature
	- Beacon Best State has best block is previous block of new beacon block
	- Beacon Best State has height compatible with new beacon block
	- Beacon Best State has epoch compatible with new beacon block
	- Beacon Best State has best shard height compatible with shard state of new beacon block
	- New Stake public key must not found in beacon best state (candidate, pending validator, committee)
*/
func (beaconBestState *BeaconBestState) verifyBestStateWithBeaconBlock(beaconBlock *BeaconBlock, isVerifySig bool) error {
	beaconBestState.lock.RLock()
	defer beaconBestState.lock.RUnlock()
	hash := beaconBlock.Header.Hash()
	//verify producer via index
	producerPublicKey := base58.Base58Check{}.Encode(beaconBlock.Header.ProducerAddress.Pk, common.ZeroByte)
	producerPosition := (beaconBestState.BeaconProposerIndex + beaconBlock.Header.Round) % len(beaconBestState.BeaconCommittee)
	tempProducer := beaconBestState.BeaconCommittee[producerPosition]
	if strings.Compare(tempProducer, producerPublicKey) != 0 {
		return NewBlockChainError(BeaconBlockProducerError, fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, producerPublicKey))
	}
	err := incognitokey.ValidateDataB58(producerPublicKey, beaconBlock.ProducerSig, hash.GetBytes())
	if err != nil {
		return NewBlockChainError(BeaconBlockSignatureError, fmt.Errorf("Producer Public Key %+v, Producer Signature %+v, Hash %+v", producerPublicKey, beaconBlock.ProducerSig, hash))
	}
	//=============Verify aggegrate signature
	if isVerifySig {
		// ValidatorIdx must > Number of Beacon Committee / 2 AND Number of Beacon Committee > 3
		if len(beaconBestState.BeaconCommittee) > 3 && len(beaconBlock.ValidatorsIndex[1]) < (len(beaconBestState.BeaconCommittee)>>1) {
			return NewBlockChainError(BeaconCommitteeLengthAndCommitteeIndexError, fmt.Errorf("Expect Number of Committee Size greater than 3 but get %+v", len(beaconBestState.BeaconCommittee)))
		}
		err := ValidateAggSignature(beaconBlock.ValidatorsIndex, beaconBestState.BeaconCommittee, beaconBlock.AggregatedSig, beaconBlock.R, beaconBlock.Hash())
		if err != nil {
			return NewBlockChainError(BeaconBlockSignatureError, err)
		}
	}
	//=============End Verify Aggegrate signature
	if !beaconBestState.BestBlockHash.IsEqual(&beaconBlock.Header.PreviousBlockHash) {
		return NewBlockChainError(BeaconBestStateBestBlockNotCompatibleError, errors.New("previous us block should be :"+beaconBestState.BestBlockHash.String()))
	}
	if beaconBestState.BeaconHeight+1 != beaconBlock.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(beaconBlock.Header.Height+1))))
	}
	if beaconBlock.Header.Height%common.EPOCH == 1 && beaconBestState.Epoch+1 != beaconBlock.Header.Epoch {
		return NewBlockChainError(WrongEpochError, fmt.Errorf("Expect beacon block height %+v has epoch %+v but get %+v", beaconBlock.Header.Height, beaconBestState.Epoch+1, beaconBlock.Header.Epoch))
	}
	if beaconBlock.Header.Height%common.EPOCH != 1 && beaconBestState.Epoch != beaconBlock.Header.Epoch {
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

/* Verify Post-processing data
- Validator root: BeaconCommittee + BeaconPendingValidator
- Beacon Candidate root: CandidateBeaconWaitingForCurrentRandom + CandidateBeaconWaitingForNextRandom
- Shard Candidate root: CandidateShardWaitingForCurrentRandom + CandidateShardWaitingForNextRandom
- Shard Validator root: ShardCommittee + ShardPendingValidator
- Random number if have in instruction
*/
func (beaconBestState *BeaconBestState) verifyPostProcessingBeaconBlock(beaconBlock *BeaconBlock) error {
	beaconBestState.lock.RLock()
	defer beaconBestState.lock.RUnlock()
	var (
		strs []string
		ok   bool
	)
	strs = append(strs, beaconBestState.BeaconCommittee...)
	strs = append(strs, beaconBestState.BeaconPendingValidator...)
	ok = verifyHashFromStringArray(strs, beaconBlock.Header.BeaconCommitteeAndValidatorRoot)
	if !ok {
		return NewBlockChainError(BeaconCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v", beaconBlock.Header.BeaconCommitteeAndValidatorRoot))
	}
	strs = []string{}
	strs = append(strs, beaconBestState.CandidateBeaconWaitingForCurrentRandom...)
	strs = append(strs, beaconBestState.CandidateBeaconWaitingForNextRandom...)
	ok = verifyHashFromStringArray(strs, beaconBlock.Header.BeaconCandidateRoot)
	if !ok {
		return NewBlockChainError(BeaconCandidateRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v", beaconBlock.Header.BeaconCandidateRoot))
	}
	strs = []string{}
	strs = append(strs, beaconBestState.CandidateShardWaitingForCurrentRandom...)
	strs = append(strs, beaconBestState.CandidateShardWaitingForNextRandom...)
	ok = verifyHashFromStringArray(strs, beaconBlock.Header.ShardCandidateRoot)
	if !ok {
		return NewBlockChainError(ShardCandidateRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v", beaconBlock.Header.ShardCandidateRoot))
	}
	ok = verifyHashFromMapByteString(beaconBestState.ShardPendingValidator, beaconBestState.ShardCommittee, beaconBlock.Header.ShardCommitteeAndValidatorRoot)
	if !ok {
		return NewBlockChainError(ShardCommitteeAndPendingValidatorRootError, fmt.Errorf("Expect Beacon Committee and Validator Root to be %+v", beaconBlock.Header.ShardCommitteeAndValidatorRoot))
	}
	// COMMENT FOR TESTING
	// instructions := block.Body.Instructions
	// for _, l := range instructions {
	// 	if l[0] == "random" {
	// 		temp, err := strconv.Atoi(l[3])
	// 		if err != nil {
	// 			Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
	// 			return NewBlockChainError(UnExpectedError, err)
	// 		}
	// 		ok, err = btc.VerifyNonceWithTimestamp(beaconBestState.CurrentRandomTimeStamp, int64(temp))
	// 		Logger.log.Infof("Verify Random number %+v", ok)
	// 		if err != nil {
	// 			Logger.log.Error("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
	// 			return NewBlockChainError(UnExpectedError, err)
	// 		}
	// 		if !ok {
	// 			return NewBlockChainError(RandomError, errors.New("Error verify random number"))
	// 		}
	// 	}
	// }
	return nil
}

/*
	Update Beststate with new Block
*/
func (beaconBestState *BeaconBestState) updateBeaconBestState(beaconBlock *BeaconBlock) error {
	beaconBestState.lock.Lock()
	defer beaconBestState.lock.Unlock()
	Logger.log.Debugf("Start processing new block at height %d, with hash %+v", beaconBlock.Header.Height, *beaconBlock.Hash())
	newBeaconCandidate := []string{}
	newShardCandidate := []string{}
	randomFlag := false // signal of random parameter from beacon block
	// update beacon best state
	beaconBestState.PreviousBestBlockHash = beaconBestState.BestBlockHash
	beaconBestState.BestBlockHash = *beaconBlock.Hash()
	beaconBestState.BestBlock = *beaconBlock
	beaconBestState.Epoch = beaconBlock.Header.Epoch
	beaconBestState.BeaconHeight = beaconBlock.Header.Height
	beaconBestState.BeaconProposerIndex = common.IndexOfStr(base58.Base58Check{}.Encode(beaconBlock.Header.ProducerAddress.Pk, common.ZeroByte), beaconBestState.BeaconCommittee)
	// Update new best new block hash
	for shardID, shardStates := range beaconBlock.Body.ShardState {
		beaconBestState.BestShardHash[shardID] = shardStates[len(shardStates)-1].Hash
		beaconBestState.BestShardHeight[shardID] = shardStates[len(shardStates)-1].Height
	}
	// processing instruction
	for _, instruction := range beaconBlock.Body.Instructions {
		err, tempRandomFlag, tempNewBeaconCandidate, tempNewShardCandidate := beaconBestState.processInstruction(instruction)
		if err != nil {
			return err
		}
		if tempRandomFlag {
			randomFlag = tempRandomFlag
		}
		newBeaconCandidate = append(newBeaconCandidate, tempNewBeaconCandidate...)
		newShardCandidate = append(newShardCandidate, tempNewShardCandidate...)
	}
	// update candidate list after processing instructions
	beaconBestState.CandidateBeaconWaitingForNextRandom = append(beaconBestState.CandidateBeaconWaitingForNextRandom, newBeaconCandidate...)
	beaconBestState.CandidateShardWaitingForNextRandom = append(beaconBestState.CandidateShardWaitingForNextRandom, newShardCandidate...)

	if beaconBestState.BeaconHeight%common.EPOCH == 1 && beaconBestState.BeaconHeight != 1 {
		// Begin of each epoch
		beaconBestState.IsGetRandomNumber = false
		// Before get random from bitcoin
	} else if beaconBestState.BeaconHeight%common.EPOCH >= common.RANDOM_TIME {
		// After get random from bitcoin
		if beaconBestState.BeaconHeight%common.EPOCH == common.RANDOM_TIME {
			// snapshot candidate list
			beaconBestState.CandidateShardWaitingForCurrentRandom = beaconBestState.CandidateShardWaitingForNextRandom
			beaconBestState.CandidateBeaconWaitingForCurrentRandom = beaconBestState.CandidateBeaconWaitingForNextRandom
			Logger.log.Info("Beacon Process: CandidateShardWaitingForCurrentRandom: ", beaconBestState.CandidateShardWaitingForCurrentRandom)
			Logger.log.Info("Beacon Process: CandidateBeaconWaitingForCurrentRandom: ", beaconBestState.CandidateBeaconWaitingForCurrentRandom)
			// reset candidate list
			beaconBestState.CandidateShardWaitingForNextRandom = []string{}
			beaconBestState.CandidateBeaconWaitingForNextRandom = []string{}
			// assign random timestamp
			beaconBestState.CurrentRandomTimeStamp = beaconBlock.Header.Timestamp
		}
		// if get new random number
		// Assign candidate to shard
		// assign CandidateShardWaitingForCurrentRandom to ShardPendingValidator with CurrentRandom
		if randomFlag {
			beaconBestState.IsGetRandomNumber = true
			err := AssignValidatorShard(beaconBestState.ShardPendingValidator, beaconBestState.CandidateShardWaitingForCurrentRandom, beaconBestState.CurrentRandomNumber, beaconBestState.ActiveShards)
			if err != nil {
				return NewBlockChainError(AssignValidatorToShardError, err)
			}
			// delete CandidateShardWaitingForCurrentRandom list
			beaconBestState.CandidateShardWaitingForCurrentRandom = []string{}
			// Shuffle candidate
			// shuffle CandidateBeaconWaitingForCurrentRandom with current random number
			newBeaconPendingValidator, err := ShuffleCandidate(beaconBestState.CandidateBeaconWaitingForCurrentRandom, beaconBestState.CurrentRandomNumber)
			if err != nil {
				return NewBlockChainError(ShuffleBeaconCandidateError, err)
			}
			beaconBestState.CandidateBeaconWaitingForCurrentRandom = []string{}
			beaconBestState.BeaconPendingValidator = append(beaconBestState.BeaconPendingValidator, newBeaconPendingValidator...)
		}
	}
	return nil
}

func (beaconBestState *BeaconBestState) initBeaconBestState(genesisBeaconBlock *BeaconBlock) error {
	var (
		newBeaconCandidate = []string{}
		newShardCandidate  = []string{}
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
	for _, instruction := range genesisBeaconBlock.Body.Instructions {
		err, _, tempNewBeaconCandidate, tempNewShardCandidate := beaconBestState.processInstruction(instruction)
		if err != nil {
			return err
		}
		newBeaconCandidate = append(newBeaconCandidate, tempNewBeaconCandidate...)
		newShardCandidate = append(newShardCandidate, tempNewShardCandidate...)
	}
	beaconBestState.BeaconCommittee = make([]string, beaconBestState.MaxBeaconCommitteeSize)
	copy(beaconBestState.BeaconCommittee, newBeaconCandidate[:beaconBestState.MaxBeaconCommitteeSize])
	for shardID := 0; shardID < beaconBestState.ActiveShards; shardID++ {
		beaconBestState.ShardCommittee[byte(shardID)] = append(beaconBestState.ShardCommittee[byte(shardID)], newShardCandidate[shardID*beaconBestState.MinShardCommitteeSize:(shardID+1)*beaconBestState.MinShardCommitteeSize]...)
	}
	beaconBestState.Epoch = 1
	return nil
}

/*
	processInstruction, process these instruction:
	- Random Instruction
		+ format
			["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
		+ store random number into beststate
	- Swap Instruction
		+ format
			["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "shard" "shardID"]
			["swap" "inPubkey1,inPubkey2,..." "outPupkey1, outPubkey2,..." "beacon"]
        + Update shard/beacon pending validator and shard/beacon committee in beststate
	- Stake Instruction
		+ format
			["stake" "pubkey1,pubkey2,..." "beacon"]
			["stake" "pubkey1,pubkey2,..." "shard"]
		+ Get Stake public key and for later storage
	Return param
	#1 error
	#2 random flag
	#3 new beacon candidate
	#4 new shard candidate

*/
func (beaconBestState *BeaconBestState) processInstruction(instruction []string) (error, bool, []string, []string) {
	newBeaconCandidate := []string{}
	newShardCandidate := []string{}
	if len(instruction) < 1 {
		return nil, false, []string{}, []string{}
	}
	// ["random" "{nonce}" "{blockheight}" "{timestamp}" "{bitcoinTimestamp}"]
	if instruction[0] == RandomAction {
		temp, err := strconv.Atoi(instruction[1])
		if err != nil {
			return NewBlockChainError(ProcessRandomInstructionError, err), false, []string{}, []string{}
		}
		beaconBestState.CurrentRandomNumber = int64(temp)
		Logger.log.Infof("Random number found %+v", beaconBestState.CurrentRandomNumber)
		return nil, true, []string{}, []string{}
	}

	if instruction[0] == SwapAction {
		Logger.log.Info("Swap Instruction", instruction)
		inPubkeys := strings.Split(instruction[1], ",")
		outPubkeys := strings.Split(instruction[2], ",")
		Logger.log.Info("Swap Instruction inPubkeys", inPubkeys)
		Logger.log.Info("Swap Instruction outPubkeys", outPubkeys)
		if instruction[3] == "shard" {
			temp, err := strconv.Atoi(instruction[4])
			if err != nil {
				return NewBlockChainError(ProcessSwapInstructionError, err), false, []string{}, []string{}
			}
			shardID := byte(temp)
			// delete in public key out of sharding pending validator list
			if len(instruction[1]) > 0 {
				tempShardPendingValidator, err := RemoveValidator(beaconBestState.ShardPendingValidator[shardID], inPubkeys)
				if err != nil {
					return NewBlockChainError(ProcessSwapInstructionError, err), false, []string{}, []string{}
				}
				// update shard pending validator
				beaconBestState.ShardPendingValidator[shardID] = tempShardPendingValidator
				// add new public key to committees
				beaconBestState.ShardCommittee[shardID] = append(beaconBestState.ShardCommittee[shardID], inPubkeys...)
			}
			// delete out public key out of current committees
			if len(instruction[2]) > 0 {
				tempShardCommittees, err := RemoveValidator(beaconBestState.ShardCommittee[shardID], outPubkeys)
				if err != nil {
					return NewBlockChainError(ProcessSwapInstructionError, err), false, []string{}, []string{}
				}
				// remove old public key in shard committee update shard committee
				beaconBestState.ShardCommittee[shardID] = tempShardCommittees
			}
		} else if instruction[3] == "beacon" {
			if len(instruction[1]) > 0 {
				tempBeaconPendingValidator, err := RemoveValidator(beaconBestState.BeaconPendingValidator, inPubkeys)
				if err != nil {
					return NewBlockChainError(ProcessSwapInstructionError, err), false, []string{}, []string{}
				}
				// update beacon pending validator
				beaconBestState.BeaconPendingValidator = tempBeaconPendingValidator
				// add new public key to beacon committee
				beaconBestState.BeaconCommittee = append(beaconBestState.BeaconCommittee, inPubkeys...)
			}
			if len(instruction[2]) > 0 {
				tempBeaconCommittes, err := RemoveValidator(beaconBestState.BeaconCommittee, outPubkeys)
				if err != nil {
					return NewBlockChainError(ProcessSwapInstructionError, err), false, []string{}, []string{}
				}
				// remove old public key in beacon committee and update beacon best state
				beaconBestState.BeaconCommittee = tempBeaconCommittes
			}
		}
		return nil, false, []string{}, []string{}
	}
	// Update candidate
	// get staking candidate list and store
	// store new staking candidate
	if instruction[0] == StakeAction && instruction[2] == "beacon" {
		beacon := strings.Split(instruction[1], ",")
		newBeaconCandidate = append(newBeaconCandidate, beacon...)
		return nil, false, newBeaconCandidate, newShardCandidate
	}

	if instruction[0] == StakeAction && instruction[2] == "shard" {
		shard := strings.Split(instruction[1], ",")
		newShardCandidate = append(newShardCandidate, shard...)
		return nil, false, newBeaconCandidate, newShardCandidate
	}
	return nil, false, []string{}, []string{}
}
func (blockchain *BlockChain) processStoreBeaconBlock(beaconBlock *BeaconBlock) error {
	blockHash := beaconBlock.Header.Hash()
	for shardID, shardStates := range beaconBlock.Body.ShardState {
		for _, shardState := range shardStates {
			err := blockchain.config.DataBase.StoreAcceptedShardToBeacon(shardID, beaconBlock.Header.Height, shardState.Hash)
			if err != nil {
				return NewBlockChainError(StoreAcceptedShardToBeaconError, err)
			}
		}
	}
	Logger.log.Infof("BEACON | Store Committee in Height %+v \n", beaconBlock.Header.Height)
	if err := blockchain.config.DataBase.StoreShardCommitteeByHeight(beaconBlock.Header.Height, blockchain.BestState.Beacon.GetShardCommittee()); err != nil {
		return NewBlockChainError(StoreShardCommitteeByHeightError, err)
	}
	if err := blockchain.config.DataBase.StoreBeaconCommitteeByHeight(beaconBlock.Header.Height, blockchain.BestState.Beacon.BeaconCommittee); err != nil {
		return NewBlockChainError(StoreBeaconCommitteeByHeightError, err)
	}
	//=========Store cross shard state ==================================
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
					err := blockchain.config.DataBase.StoreCrossShardNextHeight(fromShard, toShard, lastHeight, waitHeight)
					if err != nil {
						return NewBlockChainError(StoreCrossShardNextHeightError, err)
					}
					//beacon process shard_to_beacon in order so cross shard next height also will be saved in order
					//dont care overwrite this value
					err = blockchain.config.DataBase.StoreCrossShardNextHeight(fromShard, toShard, waitHeight, 0)
					if err != nil {
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
	// ************ Store beaconBlock at last
	//========Store new Beaconblock and new Beacon bestState in cache
	Logger.log.Info("Store Beacon BestState")
	if err := blockchain.StoreBeaconBestState(); err != nil {
		return NewBlockChainError(StoreBeaconBestStateError, err)
	}
	Logger.log.Info("Store Beacon Block ", beaconBlock.Header.Height, blockHash)
	if err := blockchain.config.DataBase.StoreBeaconBlock(beaconBlock, blockHash); err != nil {
		return NewBlockChainError(StoreBeaconBlockError, err)
	}
	if err := blockchain.config.DataBase.StoreBeaconBlockIndex(blockHash, beaconBlock.Header.Height); err != nil {
		return NewBlockChainError(StoreBeaconBlockIndexError, err)
	}
	//=========Remove beacon beaconBlock in pool
	go blockchain.config.BeaconPool.SetBeaconState(blockchain.BestState.Beacon.BeaconHeight)
	go blockchain.config.BeaconPool.RemoveBlock(blockchain.BestState.Beacon.BeaconHeight)
	//=========Remove shard to beacon beaconBlock in pool
	//Logger.log.Info("Remove beaconBlock from pool beaconBlock with hash  ", *beaconBlock.Hash(), beaconBlock.Header.Height, blockchain.BestState.Beacon.BestShardHeight)
	go blockchain.config.ShardToBeaconPool.SetShardState(blockchain.BestState.Beacon.GetBestShardHeight())
	err := blockchain.updateDatabaseFromBeaconBlock(beaconBlock)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	// execute, store
	err = blockchain.processBridgeInstructions(beaconBlock)
	if err != nil {
		Logger.log.Errorf("Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
		return NewBlockChainError(UnExpectedError, err)
	}
	return nil
}
