package blockchain

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/metrics"
	"github.com/incognitochain/incognito-chain/pubsub"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
)

/*
	Verify Shard Block Before Signing
	Used for PBFT consensus
	@Notice: this block doesn't have full information (incomplete block)
*/
func (blockchain *BlockChain) VerifyPreSignShardBlock(shardBlock *ShardBlock, shardID byte) error {
	if shardBlock.Header.ShardID != shardID {
		return NewBlockChainError(WrongShardIDError, fmt.Errorf("expect shardID %+v, but get shardID %+v", shardID, shardBlock.Header.ShardID))
	}
	blockchain.BestState.Shard[shardID].lock.Lock()
	defer blockchain.BestState.Shard[shardID].lock.Unlock()
	// fetch beacon blocks
	previousBeaconHeight := blockchain.BestState.Shard[shardID].BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.config.DataBase, previousBeaconHeight+1, shardBlock.Header.BeaconHeight)
	if err != nil {
		return err
	}
	//========Verify shardBlock only
	Logger.log.Infof("SHARD %+v | Verify ShardBlock for signing process %d, with hash %+v", shardID, shardBlock.Header.Height, *shardBlock.Hash())
	if err := blockchain.verifyPreProcessingShardBlock(shardBlock, beaconBlocks, shardID, true); err != nil {
		return err
	}
	//========Verify shardBlock with previous best state
	// Get Beststate of previous shardBlock == previous best state
	// Clone best state value into new variable
	shardBestState := NewShardBestState()
	if err := shardBestState.cloneShardBestState(blockchain.BestState.Shard[shardID]); err != nil {
		return err
	}
	//// check with current final best state
	//// New shardBlock must be compatible with current best state
	//bestBlockHash := &blockchain.BestState.Shard[shardID].BestBlockHash
	//if bestBlockHash.IsEqual(&shardBlock.Header.PreviousBlockHash) {
	//if err := shardBestState.cloneShardBestState(blockchain.BestState.Shard[shardID]); err != nil {
	//	return err
	//}
	//} else {
	//	// if no match best state found then shardBlock is unknown
	//	return NewBlockChainError(ShardBestStateNotCompatibleError, fmt.Errorf("Current Best Block Hash %+v, New Shard Block %+v, Previous Block Hash of New Block %+v", *bestBlockHash, shardBlock.Header.Height, shardBlock.Header.PreviousBlockHash))
	//}
	// Verify shardBlock with previous best state
	// DO NOT verify agg signature in this function
	if err := shardBestState.verifyBestStateWithShardBlock(shardBlock, false, shardID); err != nil {
		return err
	}
	//========updateShardBestState best state with new shardBlock
	if err := shardBestState.updateShardBestState(shardBlock, beaconBlocks); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding shardBlock
	if err := shardBestState.verifyPostProcessingShardBlock(shardBlock, shardID); err != nil {
		return err
	}
	Logger.log.Criticalf("SHARD %+v | Block %d, with hash %+v is VALID for 🖋 signing", shardID, shardBlock.Header.Height, *shardBlock.Hash())
	return nil
}

/*
	Insert Shard Block into blockchain
	@Notice: this block must have full information (complete block)
*/
func (blockchain *BlockChain) InsertShardBlock(block *ShardBlock, isValidated bool) error {
	shardID := block.Header.ShardID
	blockHash := block.Header.Hash()
	blockchain.BestState.Shard[shardID].lock.Lock()
	defer blockchain.BestState.Shard[shardID].lock.Unlock()
	Logger.log.Criticalf("SHARD %+v | Begin insert new block height %+v with hash %+v", block.Header.ShardID, block.Header.Height, blockHash)
	// Force non-committee (validator) member not to validate blk
	if blockchain.config.UserKeySet != nil && (blockchain.config.NodeMode == common.NODEMODE_AUTO || blockchain.config.NodeMode == common.NODEMODE_SHARD) {
		userRole := blockchain.BestState.Shard[block.Header.ShardID].GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyInBase58CheckEncode(), 0)
		Logger.log.Infof("SHARD %+v | Shard block height %+v with hash %+v, User Role %+v ", block.Header.ShardID, block.Header.Height, blockHash, userRole)
		if userRole != common.PROPOSER_ROLE && userRole != common.VALIDATOR_ROLE && userRole != common.PENDING_ROLE {
			isValidated = true
		}
	} else {
		isValidated = true
	}
	Logger.log.Infof("SHARD %+v | Check block existence for insert height %+v with hash %+v", block.Header.ShardID, block.Header.Height, blockHash)
	isExist, _ := blockchain.config.DataBase.HasBlock(blockHash)
	if isExist {
		return NewBlockChainError(DuplicateShardBlockError, fmt.Errorf("SHARD %+v, block height %+v wit hash %+v has been stored already", block.Header.ShardID, block.Header.Height, blockHash))
	}
	// fetch beacon blocks
	previousBeaconHeight := blockchain.BestState.Shard[shardID].BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.config.DataBase, previousBeaconHeight+1, block.Header.BeaconHeight)
	if err != nil {
		return NewBlockChainError(FetchBeaconBlocksError, err)
	}
	if !isValidated {
		Logger.log.Infof("SHARD %+v | Verify Pre Processing, block height %+v with hash %+v \n", block.Header.ShardID, block.Header.Height, blockHash)
		if err := blockchain.verifyPreProcessingShardBlock(block, beaconBlocks, shardID, false); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Pre Processing, block height %+v with hash %+v \n", block.Header.ShardID, block.Header.Height, blockHash)
	}
	// Verify block with previous best state
	Logger.log.Infof("SHARD %+v | Verify BestState With Shard Block, block height %+v with hash %+v \n", block.Header.ShardID, block.Header.Height, blockHash)
	if err := blockchain.BestState.Shard[shardID].verifyBestStateWithShardBlock(block, true, shardID); err != nil {
		return err
	}
	//// check with current final best state
	//// block can only be insert if it match the current best state
	//bestBlockHash := &blockchain.BestState.Shard[shardID].BestBlockHash
	//if !bestBlockHash.IsEqual(&block.Header.PreviousBlockHash) {
	//	return NewBlockChainError(BeaconError, errors.New("beacon Block does not match with any Beacon State in cache or in Database"))
	//}

	Logger.log.Infof("SHARD %+v | Update ShardBestState, block height %+v with hash %+v \n", block.Header.ShardID, block.Header.Height, blockHash)
	//========updateShardBestState best state with new block
	//previousBeaconHeight := blockchain.BestState.Shard[shardID].BeaconHeight
	//beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.config.DataBase, previousBeaconHeight+1, block.Header.BeaconHeight)
	//if err != nil {
	//	return err
	//}
	// Backup beststate
	if blockchain.config.UserKeySet != nil {
		userRole := blockchain.BestState.Shard[shardID].GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyInBase58CheckEncode(), 0)
		if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
			err = blockchain.config.DataBase.CleanBackup(true, block.Header.ShardID)
			if err != nil {
				return err
			}
			err = blockchain.BackupCurrentShardState(block, beaconBlocks)
			if err != nil {
				return err
			}
		}
	}
	if err := blockchain.BestState.Shard[shardID].updateShardBestState(block, beaconBlocks); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding block
	if !isValidated {
		Logger.log.Infof("SHARD %+v | Verify Post Processing, block height %+v with hash %+v \n", block.Header.ShardID, block.Header.Height, blockHash)
		if err := blockchain.BestState.Shard[shardID].verifyPostProcessingShardBlock(block, shardID); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Post Processing, block height %+v with hash %+v \n", block.Header.ShardID, block.Header.Height, blockHash)
	}
	Logger.log.Infof("SHARD %+v | Remove Data After Processed, block height %+v with hash %+v \n", block.Header.ShardID, block.Header.Height, blockHash)
	go blockchain.removeOldDataAfterProcessing(block, shardID)
	Logger.log.Infof("SHARD %+v | Update Beacon Instruction, block height %+v with hash %+v \n", block.Header.ShardID, block.Header.Height, blockHash)
	err = blockchain.updateDatabaseFromBeaconInstructions(beaconBlocks, shardID)
	if err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Store New Shard Block And Update Data, block height %+v with hash %+v \n", block.Header.ShardID, block.Header.Height, blockHash)
	//========Store new  Shard block and new shard bestState
	err = blockchain.processStoreShardBlockAndUpdateDatabase(block)
	if err != nil {
		return err
	}
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewShardblockTopic, block))
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardBeststateTopic, blockchain.BestState.Shard[shardID]))
	shardIDForMetric := strconv.Itoa(int(block.Header.ShardID))
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.NumOfBlockInsertToChain,
		metrics.MeasurementValue: float64(1),
		metrics.Tag:              metrics.ShardIDTag,
		metrics.TagValue:         metrics.Shard + shardIDForMetric,
	})
	Logger.log.Criticalf("SHARD %+v | 🔗 Finish Insert new block %d, with hash %+v", block.Header.ShardID, block.Header.Height, blockHash)
	return nil
}

/* Verify Pre-prosessing data
This function DOES NOT verify new block with best state
DO NOT USE THIS with GENESIS BLOCK
	- Producer
	- ShardID: of received block same shardID
	- Version
	- Parent hash
	- Height = parent hash + 1
	- Epoch = blockHeight % Epoch ? Parent Epoch + 1
	- Timestamp can not excess some limit
	- TxRoot
	- ShardTxRoot
	- CrossOutputCoinRoot
	- ActionsRoot
	- BeaconHeight
	- BeaconHash
	- Swap instruction
	- ALL Transaction in block: see in verifyTransactionFromNewBlock
*/
func (blockchain *BlockChain) verifyPreProcessingShardBlock(block *ShardBlock, beaconBlocks []*BeaconBlock, shardID byte, isPresig bool) error {
	//verify producer sig
	Logger.log.Debugf("SHARD %+v | Begin verifyPreProcessingShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	if block.Header.ShardID != shardID {
		return NewBlockChainError(WrongShardIDError, fmt.Errorf("Expect receive block from Shard ID %+v but get %+v", shardID, block.Header.ShardID))
	}
	if block.Header.Version != VERSION {
		return NewBlockChainError(WrongVersionError, fmt.Errorf("Expect block version %+v but get %+v", VERSION, block.Header.Version))
	}
	// Verify parent hash exist or not
	previousBlockHash := block.Header.PreviousBlockHash
	previousShardBlockData, err := blockchain.config.DataBase.FetchBlock(previousBlockHash)
	if err != nil {
		return NewBlockChainError(FetchPreviousBlockError, err)
	}
	previousShardBlock := ShardBlock{}
	err = json.Unmarshal(previousShardBlockData, &previousShardBlock)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	// Verify block height with parent block
	if previousShardBlock.Header.Height+1 != block.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Expect receive block height %+v but get %+v", previousShardBlock.Header.Height+1, block.Header.Height))
	}
	// Verify timestamp with parent block
	if block.Header.Timestamp <= previousShardBlock.Header.Timestamp {
		return NewBlockChainError(WrongTimestampError, fmt.Errorf("Expect receive block has timestamp must be greater than %+v but get %+v", previousShardBlock.Header.Timestamp, block.Header.Timestamp))
	}
	// Verify transaction root
	txMerkle := Merkle{}.BuildMerkleTreeStore(block.Body.Transactions)
	txRoot := &common.Hash{}
	if len(txMerkle) > 0 {
		txRoot = txMerkle[len(txMerkle)-1]
	}
	if !bytes.Equal(block.Header.TxRoot.GetBytes(), txRoot.GetBytes()) {
		return NewBlockChainError(TransactionRootHashError, fmt.Errorf("Expect transaction root hash %+v but get %+v", block.Header.TxRoot, txRoot))
	}
	// Verify ShardTx Root
	_, shardTxMerkleData := CreateShardTxRoot2(block.Body.Transactions)
	shardTxRoot := shardTxMerkleData[len(shardTxMerkleData)-1]
	if !bytes.Equal(block.Header.ShardTxRoot.GetBytes(), shardTxRoot.GetBytes()) {
		return NewBlockChainError(ShardTransactionRootHashError, fmt.Errorf("Expect shard transaction root hash %+v but get %+v", block.Header.ShardTxRoot, shardTxRoot))
	}
	// Verify crossTransaction coin
	if !VerifyMerkleCrossTransaction(block.Body.CrossTransactions, block.Header.CrossTransactionRoot) {
		return NewBlockChainError(CrossShardTransactionRootHashError, fmt.Errorf("Expect cross shard transaction root hash %+v", block.Header.CrossTransactionRoot))
	}
	// Verify Action
	//beaconBlocks, err := FetchBeaconBlockFromHeight(
	//	blockchain.config.DataBase,
	//	blockchain.BestState.Shard[block.Header.ShardID].BeaconHeight+1,
	//	block.Header.BeaconHeight,
	//)
	//if err != nil {
	//	Logger.log.Error(err)
	//	return err
	//}

	//wrongTxsFee := errors.New("Wrong blockheader totalTxs fee")

	totalTxsFee := make(map[common.Hash]uint64)
	for _, tx := range block.Body.Transactions {
		totalTxsFee[*tx.GetTokenID()] += tx.GetTxFee()
		txType := tx.GetType()
		if txType == common.TxCustomTokenPrivacyType {
			txCustomPrivacy := tx.(*transaction.TxCustomTokenPrivacy)
			totalTxsFee[*txCustomPrivacy.GetTokenID()] = txCustomPrivacy.GetTxFeeToken()
		}
	}
	tokenIDsfromTxs := make([]common.Hash, 0)
	for tokenID, _ := range totalTxsFee {
		tokenIDsfromTxs = append(tokenIDsfromTxs, tokenID)
	}
	sort.Slice(tokenIDsfromTxs, func(i int, j int) bool {
		res, _ := tokenIDsfromTxs[i].Cmp(&tokenIDsfromTxs[j])
		return res == -1
	})
	tokenIDsfromBlock := make([]common.Hash, 0)
	for tokenID, _ := range block.Header.TotalTxsFee {
		tokenIDsfromBlock = append(tokenIDsfromBlock, tokenID)
	}
	sort.Slice(tokenIDsfromBlock, func(i int, j int) bool {
		res, _ := tokenIDsfromBlock[i].Cmp(&tokenIDsfromBlock[j])
		return res == -1
	})
	if len(tokenIDsfromTxs) != len(tokenIDsfromBlock) {
		return NewBlockChainError(WrongBlockTotalFeeError, fmt.Errorf("Expect Total Fee to be equal, From Txs %+v, From Block %+v", len(tokenIDsfromTxs), len(tokenIDsfromBlock)))
	}
	for i, tokenID := range tokenIDsfromTxs {
		if !tokenIDsfromTxs[i].IsEqual(&tokenIDsfromBlock[i]) {
			return NewBlockChainError(WrongBlockTotalFeeError, fmt.Errorf("Expect Total Fee to be equal, From Txs %+v, From Block %+v", tokenIDsfromTxs[i], tokenIDsfromBlock[i]))
		}
		if totalTxsFee[tokenID] != block.Header.TotalTxsFee[tokenID] {
			return NewBlockChainError(WrongBlockTotalFeeError, fmt.Errorf("Expect Total Fee to be equal, From Txs %+v, From Block Header %+v", totalTxsFee[tokenID], block.Header.TotalTxsFee[tokenID]))
		}
	}
	txInstructions, err := CreateShardInstructionsFromTransactionAndIns(
		block.Body.Transactions,
		blockchain,
		shardID,
		&block.Header.ProducerAddress,
		block.Header.Height,
		beaconBlocks,
		block.Header.BeaconHeight,
	)
	if err != nil {
		Logger.log.Error(err)
		return NewBlockChainError(ShardIntructionFromTransactionAndInsError, err)
	}
	if !isPresig {
		totalInstructions := []string{}
		for _, value := range txInstructions {
			totalInstructions = append(totalInstructions, value...)
		}
		for _, value := range block.Body.Instructions {
			totalInstructions = append(totalInstructions, value...)
		}
		isOk := VerifyHashFromStringArray(totalInstructions, block.Header.InstructionsRoot)
		if !isOk {
			return NewBlockChainError(InstructionsHashError, fmt.Errorf("Expect instruction hash to be %+v", block.Header.InstructionsRoot))
		}
	}
	// Check if InstructionMerkleRoot is the root of merkle tree containing all instructions in this block
	flattenTxInsts, err := FlattenAndConvertStringInst(txInstructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from Tx: %+v", err))
	}
	flattenInsts, err := FlattenAndConvertStringInst(block.Body.Instructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from block body: %+v", err))
	}
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be the same as when creating new shard block
	root := GetKeccak256MerkleRoot(insts)
	if !bytes.Equal(root, block.Header.InstructionMerkleRoot[:]) {
		return NewBlockChainError(InstructionMerkleRootError, fmt.Errorf("Expect transaction merkle root to be %+v but get %+v", block.Header.InstructionMerkleRoot, string(root)))
	}
	//Get beacon hash by height in db
	//If hash not found then fail to verify
	beaconHash, err := blockchain.config.DataBase.GetBeaconBlockHashByIndex(block.Header.BeaconHeight)
	if err != nil {
		return err
	}
	//Hash in db must be equal to hash in shard block
	newHash, err := common.Hash{}.NewHash(block.Header.BeaconHash.GetBytes())
	if err != nil {
		return NewBlockChainError(HashError, err)
	}
	if !newHash.IsEqual(&beaconHash) {
		return NewBlockChainError(BeaconError, errors.New("beacon block height and beacon block hash are not compatible in Database"))
	}
	// Swap instruction
	for _, l := range block.Body.Instructions {
		if l[0] == "swap" {
			if l[3] != "shard" || l[4] != strconv.Itoa(int(shardID)) {
				return NewBlockChainError(InstructionError, errors.New("swap instruction is invalid"))
			}
		}
	}
	// Verify response transactions
	instsForValidations := [][]string{}
	instsForValidations = append(instsForValidations, block.Body.Instructions...)
	for _, beaconBlock := range beaconBlocks {
		instsForValidations = append(instsForValidations, beaconBlock.Body.Instructions...)
	}
	invalidTxs, err := blockchain.verifyMinerCreatedTxBeforeGettingInBlock(instsForValidations, block.Body.Transactions, shardID)
	if err != nil {
		return NewBlockChainError(TransactionError, err)
	}
	if len(invalidTxs) > 0 {
		return NewBlockChainError(TransactionError, fmt.Errorf("There are %d invalid txs", len(invalidTxs)))
	}
	err = blockchain.ValidateResponseTransactionFromTxsWithMetadata(&block.Body)
	if err != nil {
		return err
	}
	// Get cross shard block from pool
	// @NOTICE: COMMENT to bypass verify cross shard block
	if isPresig {
		// Verify Transaction
		if err := blockchain.verifyTransactionFromNewBlock(block.Body.Transactions); err != nil {
			return NewBlockChainError(TransactionError, err)
		}
		// Verify Instruction
		instructions := [][]string{}
		shardCommittee := blockchain.BestState.Shard[shardID].ShardCommittee
		shardPendingValidator := blockchain.processInstructionFromBeacon(beaconBlocks, shardID)
		instructions, shardPendingValidator, shardCommittee, err = blockchain.generateInstruction(shardID, block.Header.BeaconHeight, beaconBlocks, shardPendingValidator, shardCommittee)
		if err != nil {
			return NewBlockChainError(InstructionError, err)
		}
		totalInstructions := []string{}
		for _, value := range txInstructions {
			totalInstructions = append(totalInstructions, value...)
		}
		for _, value := range instructions {
			totalInstructions = append(totalInstructions, value...)
		}
		isOk := VerifyHashFromStringArray(totalInstructions, block.Header.InstructionsRoot)
		if !isOk {
			return NewBlockChainError(HashError, errors.New("Error verify action root"))
		}
		// Verify Cross Shard Output Coin and Custom Token Transaction
		crossTxTokenData := make(map[byte][]CrossTxTokenData)
		toShard := shardID
		crossShardLimit := blockchain.config.CrossShardPool[toShard].GetLatestValidBlockHeight()
		toShardAllCrossShardBlock := blockchain.config.CrossShardPool[toShard].GetValidBlock(crossShardLimit)
		for fromShard, crossTransactions := range block.Body.CrossTransactions {
			toShardCrossShardBlocks, ok := toShardAllCrossShardBlock[fromShard]
			if !ok {
				heights := []uint64{}
				for _, crossTransaction := range crossTransactions {
					heights = append(heights, crossTransaction.BlockHeight)
				}
				blockchain.Synker.SyncBlkCrossShard(false, false, []common.Hash{}, heights, fromShard, shardID, "")
				return NewBlockChainError(CrossShardBlockError, errors.New("Cross Shard Block From Shard "+strconv.Itoa(int(fromShard))+" Not Found in Pool"))
			}
			sort.SliceStable(toShardCrossShardBlocks[:], func(i, j int) bool {
				return toShardCrossShardBlocks[i].Header.Height < toShardCrossShardBlocks[j].Header.Height
			})
			startHeight := blockchain.BestState.Shard[toShard].BestCrossShard[fromShard]
			isValids := 0
			for _, crossTransaction := range crossTransactions {
				for index, toShardCrossShardBlock := range toShardCrossShardBlocks {
					//Compare block height and block hash
					if crossTransaction.BlockHeight == toShardCrossShardBlock.Header.Height {
						nextHeight, err := blockchain.config.DataBase.FetchCrossShardNextHeight(fromShard, toShard, startHeight)
						if err != nil {
							return NewBlockChainError(CrossShardBlockError, err)
						}
						if nextHeight != crossTransaction.BlockHeight {
							return NewBlockChainError(CrossShardBlockError, errors.New("Next Cross Shard Block "+strconv.Itoa(int(toShardCrossShardBlock.Header.Height))+"is Not Expected block Height "+strconv.Itoa(int(nextHeight))+" from shard "+strconv.Itoa(int(fromShard))))
						}
						startHeight = nextHeight
						temp, err := blockchain.config.DataBase.FetchCommitteeByHeight(toShardCrossShardBlock.Header.BeaconHeight)
						if err != nil {
							return NewBlockChainError(CrossShardBlockError, err)
						}
						shardCommittee := make(map[byte][]string)
						json.Unmarshal(temp, &shardCommittee)
						err = toShardCrossShardBlock.VerifyCrossShardBlock(shardCommittee[toShardCrossShardBlock.Header.ShardID])
						if err != nil {
							return NewBlockChainError(CrossShardBlockError, err)
						}
						compareCrossTransaction := CrossTransaction{
							TokenPrivacyData: toShardCrossShardBlock.CrossTxTokenPrivacyData,
							OutputCoin:       toShardCrossShardBlock.CrossOutputCoin,
							BlockHash:        *toShardCrossShardBlock.Hash(),
							BlockHeight:      toShardCrossShardBlock.Header.Height,
						}
						targetHash := crossTransaction.Hash()
						hash := compareCrossTransaction.Hash()
						if !hash.IsEqual(&targetHash) {
							return NewBlockChainError(CrossShardBlockError, errors.New("Cross Output Coin From New Block not compatible with cross shard block in pool"))
						}
						txTokenData := CrossTxTokenData{
							TxTokenData: toShardCrossShardBlock.CrossTxTokenData,
							BlockHash:   *toShardCrossShardBlock.Hash(),
							BlockHeight: toShardCrossShardBlock.Header.Height,
						}
						crossTxTokenData[toShardCrossShardBlock.Header.ShardID] = append(crossTxTokenData[toShardCrossShardBlock.Header.ShardID], txTokenData)
						if true {
							toShardCrossShardBlocks = toShardCrossShardBlocks[index:]
							isValids++
							break
						}
					}
				}
			}
			if len(crossTransactions) != isValids {
				return NewBlockChainError(CrossShardBlockError, errors.New("Can't not verify all cross shard block from shard "+strconv.Itoa(int(fromShard))))
			}
		}
		if err := blockchain.verifyCrossShardCustomToken(crossTxTokenData, shardID, block.Body.Transactions); err != nil {
			return NewBlockChainError(CrossShardBlockError, err)
		}
	}
	Logger.log.Debugf("SHARD %+v | Finish verifyPreProcessingShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return err
}

/*
	This function will verify the validation of a block with some best state in cache or current best state
	Get beacon state of this block
	For example, new blockHeight is 91 then beacon state of this block must have height 90
	OR new block has previous has is beacon best block hash
	- Producer
	- committee length and validatorIndex length
	- Producer + sig
	- Has parent hash is current best state best blockshard hash (compatible with current beststate)
	- Block Height
	- Beacon Height
	- Action root
*/
func (shardBestState *ShardBestState) verifyBestStateWithShardBlock(shardBlock *ShardBlock, isVerifySig bool, shardID byte) error {
	Logger.log.Debugf("SHARD %+v | Begin VerifyBestStateWithShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	// Cal next producer
	// Verify next producer
	//=============Verify producer signature
	producerPosition := (shardBestState.ShardProposerIdx + shardBlock.Header.Round) % len(shardBestState.ShardCommittee)
	producerPubkey := shardBestState.ShardCommittee[producerPosition]
	blockHash := shardBlock.Header.Hash()
	fmt.Println("V58", producerPubkey, shardBlock.ProducerSig, blockHash.GetBytes(), base58.Base58Check{}.Encode(shardBlock.Header.ProducerAddress.Pk, common.ZeroByte))
	//verify producer
	tempProducer := shardBestState.ShardCommittee[producerPosition]
	if strings.Compare(tempProducer, producerPubkey) != 0 {
		return NewBlockChainError(ProducerError, errors.New("Producer should be should be :"+tempProducer))
	}
	if err := incognitokey.ValidateDataB58(producerPubkey, shardBlock.ProducerSig, blockHash.GetBytes()); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	//=============End Verify producer signature
	//=============Verify aggegrate signature
	if isVerifySig {
		if len(shardBestState.ShardCommittee) > 3 && len(shardBlock.ValidatorsIdx[1]) < (len(shardBestState.ShardCommittee)>>1) {
			return NewBlockChainError(SignatureError, errors.New("shardBlock validators and Shard committee is not compatible"))
		}
		err := ValidateAggSignature(shardBlock.ValidatorsIdx, shardBestState.ShardCommittee, shardBlock.AggregatedSig, shardBlock.R, shardBlock.Hash())
		if err != nil {
			return err
		}
	}
	//=============End Verify Aggegrate signature
	// check with current final best state
	// shardBlock can only be insert if it match the current best state
	bestBlockHash := shardBestState.BestBlockHash
	if !bestBlockHash.IsEqual(&shardBlock.Header.PreviousBlockHash) {
		return NewBlockChainError(ShardBestStateNotCompatibleError, fmt.Errorf("Current Best Block Hash %+v, New Shard Block %+v, Previous Block Hash of New Block %+v", bestBlockHash, shardBlock.Header.Height, shardBlock.Header.PreviousBlockHash))
	}
	if shardBestState.ShardHeight+1 != shardBlock.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Shard Block height of new Shard Block should be %+v, but get %+v", shardBestState.ShardHeight+1, shardBlock.Header.Height))
	}
	if shardBlock.Header.BeaconHeight < shardBestState.BeaconHeight {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Shard Block contain invalid beacon height, current beacon height %+v but get %+v ", shardBestState.BeaconHeight, shardBlock.Header.BeaconHeight))
	}
	Logger.log.Debugf("SHARD %+v | Finish VerifyBestStateWithShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	return nil
}

/*
	updateShardBestState beststate with new block
		PrevShardBlockHash
		BestShardBlockHash
		BestBeaconHash
		BestShardBlock
		ShardHeight
		BeaconHeight
		ShardProposerIdx

		Add pending validator
		Swap shard committee if detect new epoch of beacon
*/
func (shardBestState *ShardBestState) updateShardBestState(block *ShardBlock, beaconBlocks []*BeaconBlock) error {
	Logger.log.Debugf("SHARD %+v | Begin update Beststate with new Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	var (
		err                   error
		shardSwapedCommittees []string
		shardNewCommittees    []string
	)
	shardBestState.BestBlockHash = *block.Hash()
	if block.Header.BeaconHeight == 1 {
		shardBestState.BestBeaconHash = *ChainTestParam.GenesisBeaconBlock.Hash()
	} else {
		shardBestState.BestBeaconHash = block.Header.BeaconHash
	}
	if block.Header.Height == 1 {
		shardBestState.BestCrossShard = make(map[byte]uint64)
	}
	shardBestState.BestBlock = block
	shardBestState.BestBlockHash = *block.Hash()
	shardBestState.ShardHeight = block.Header.Height
	shardBestState.Epoch = block.Header.Epoch
	shardBestState.BeaconHeight = block.Header.BeaconHeight
	shardBestState.TotalTxns += uint64(len(block.Body.Transactions))
	shardBestState.NumTxns = uint64(len(block.Body.Transactions))
	//======BEGIN For testing and benchmark
	temp := 0
	for _, tx := range block.Body.Transactions {
		//detect transaction that's not salary
		if !tx.IsSalaryTx() {
			temp++
		}
	}
	shardBestState.TotalTxnsExcludeSalary += uint64(temp)
	//======END
	if block.Header.Height == 1 {
		shardBestState.ShardProposerIdx = 0
	} else {
		shardBestState.ShardProposerIdx = common.IndexOfStr(base58.Base58Check{}.Encode(block.Header.ProducerAddress.Pk, common.ZeroByte), shardBestState.ShardCommittee)
	}

	newBeaconCandidate := []string{}
	newShardCandidate := []string{}
	// Add pending validator
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {

			if l[0] == StakeAction && l[2] == "beacon" {
				beacon := strings.Split(l[1], ",")
				newBeaconCandidate = append(newBeaconCandidate, beacon...)
				if len(l) == 4 {
					for i, v := range strings.Split(l[3], ",") {
						GetBestStateShard(shardBestState.ShardID).StakingTx[newBeaconCandidate[i]] = v
					}
				}
			}
			if l[0] == StakeAction && l[2] == "shard" {
				shard := strings.Split(l[1], ",")
				newShardCandidate = append(newShardCandidate, shard...)
				if len(l) == 4 {
					for i, v := range strings.Split(l[3], ",") {
						GetBestStateShard(shardBestState.ShardID).StakingTx[newShardCandidate[i]] = v
					}
				}
			}

			if l[0] == "assign" && l[2] == "shard" {
				if l[3] == strconv.Itoa(int(block.Header.ShardID)) {
					Logger.log.Infof("SHARD %+v | Old ShardPendingValidatorList %+v", block.Header.ShardID, shardBestState.ShardPendingValidator)
					shardBestState.ShardPendingValidator = append(shardBestState.ShardPendingValidator, strings.Split(l[1], ",")...)
					Logger.log.Infof("SHARD %+v | New ShardPendingValidatorList %+v", block.Header.ShardID, shardBestState.ShardPendingValidator)
				}
			}
		}
	}
	if len(block.Body.Instructions) != 0 {
		Logger.log.Critical("Shard Process/updateShardBestState: ALL Instruction", block.Body.Instructions)
	}

	// Swap committee
	for _, l := range block.Body.Instructions {

		if l[0] == "swap" {
			shardBestState.ShardPendingValidator, shardBestState.ShardCommittee, shardSwapedCommittees, shardNewCommittees, err = SwapValidator(shardBestState.ShardPendingValidator, shardBestState.ShardCommittee, shardBestState.MaxShardCommitteeSize, common.OFFSET)
			if err != nil {
				Logger.log.Errorf("SHARD %+v | Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			swapedCommittees := []string{}
			if len(l[2]) != 0 && l[2] != "" {
				swapedCommittees = strings.Split(l[2], ",")
			}
			newCommittees := strings.Split(l[1], ",")

			for _, v := range swapedCommittees {
				delete(GetBestStateShard(shardBestState.ShardID).StakingTx, v)
			}

			if !reflect.DeepEqual(swapedCommittees, shardSwapedCommittees) {
				return NewBlockChainError(SwapError, errors.New("invalid shard swapped committees"))
			}
			if !reflect.DeepEqual(newCommittees, shardNewCommittees) {
				return NewBlockChainError(SwapError, errors.New("invalid shard new committees"))
			}
			Logger.log.Infof("SHARD %+v | Swap: Out committee %+v", block.Header.ShardID, shardSwapedCommittees)
			Logger.log.Infof("SHARD %+v | Swap: In committee %+v", block.Header.ShardID, shardNewCommittees)
		}
	}
	//updateShardBestState best cross shard
	for shardID, crossShardBlock := range block.Body.CrossTransactions {
		shardBestState.BestCrossShard[shardID] = crossShardBlock[len(crossShardBlock)-1].BlockHeight
	}

	Logger.log.Debugf("SHARD %+v | Finish update Beststate with new Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return nil
}

/*
	verifyPostProcessingShardBlock
	- commitee root
	- pending validator root
*/
func (shardBestState *ShardBestState) verifyPostProcessingShardBlock(block *ShardBlock, shardID byte) error {
	var (
		isOk bool
	)
	Logger.log.Debugf("SHARD %+v | Begin VerifyPostProcessing Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	isOk = VerifyHashFromStringArray(shardBestState.ShardCommittee, block.Header.CommitteeRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("Error verify Committee root"))
	}
	isOk = VerifyHashFromStringArray(shardBestState.ShardPendingValidator, block.Header.PendingValidatorRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("Error verify Pending validator root"))
	}
	Logger.log.Debugf("SHARD %+v | Finish VerifyPostProcessing Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return nil
}

/*
	Verify Transaction with these condition:
	1. Validate tx version
	2. Validate fee with tx size
	3. Validate type of tx
	4. Validate with other txs in block:
 	- Normal Transaction:
 	- Custom Tx:
	4.1 Validate Init Custom Token
	5. Validate sanity data of tx
	6. Validate data in tx: privacy proof, metadata,...
	7. Validate tx with blockchain: douple spend, ...
	8. Check tx existed in block
	9. Not accept a salary tx
	10. Check duplicate staker public key in block
	11. Check duplicate Init Custom Token in block
*/
func (blockChain *BlockChain) verifyTransactionFromNewBlock(txs []metadata.Transaction) error {
	if len(txs) == 0 {
		return nil
	}
	isEmpty := blockChain.config.TempTxPool.EmptyPool()
	if !isEmpty {
		panic("TempTxPool Is not Empty")
	}
	defer blockChain.config.TempTxPool.EmptyPool()

	err := blockChain.config.TempTxPool.ValidateTxList(txs)
	if err != nil {
		Logger.log.Errorf("Error validating transaction in block creation: %+v \n", err)
		return NewBlockChainError(TransactionError, errors.New("Some Transactions in New Block IS invalid"))
	}
	// TODO: uncomment to synchronize validate method with shard process and mempool
	//for _, tx := range txs {
	//	if !tx.IsSalaryTx() {
	//		if tx.GetType() == common.TxCustomTokenType {
	//			customTokenTx := tx.(*transaction.TxCustomToken)
	//			if customTokenTx.TxTokenData.Type == transaction.CustomTokenCrossShard {
	//				continue
	//			}
	//		}
	//		_, err := blockChain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx)
	//		if err != nil {
	//			return err
	//		}
	//	}
	//}
	return nil
}

/*
	Store All information after Insert
	- Shard Block
	- Shard Best State
	- Transaction => UTXO, serial number, snd, commitment
	- Cross Output Coin => UTXO, snd, commmitment
	- Store transaction metadata:
		+ Withdraw Metadata
	- Store incoming cross shard block
	- Store Burning Confirmation
	- Update Mempool fee estimator
*/
func (blockchain *BlockChain) processStoreShardBlockAndUpdateDatabase(block *ShardBlock) error {
	blockHash := block.Hash().String()
	Logger.log.Infof("SHARD %+v | Process store block height %+v at hash %+v", block.Header.ShardID, block.Header.Height, *block.Hash())

	if err := blockchain.StoreShardBlock(block); err != nil {
		return err
	}

	if err := blockchain.StoreShardBlockIndex(block); err != nil {
		return err
	}

	if err := blockchain.StoreShardBestState(block.Header.ShardID); err != nil {
		return err
	}
	if len(block.Body.CrossTransactions) != 0 {
		Logger.log.Critical("processStoreShardBlockAndUpdateDatabase/CrossTransactions	", block.Body.CrossTransactions)
	}

	if err := blockchain.CreateAndSaveTxViewPointFromBlock(block); err != nil {
		return err
	}

	for index, tx := range block.Body.Transactions {
		if err := blockchain.StoreTransactionIndex(tx.Hash(), block.Header.Hash(), index); err != nil {
			Logger.log.Error("ERROR", err, "Transaction in block with hash", blockHash, "and index", index, ":", tx)
			return NewBlockChainError(UnExpectedError, err)
		}
		// Process Transaction Metadata
		metaType := tx.GetMetadataType()
		if metaType == metadata.WithDrawRewardResponseMeta {
			_, requesterRes, amountRes, coinID := tx.GetTransferData()
			err := blockchain.config.DataBase.RemoveCommitteeReward(requesterRes, amountRes, *coinID)
			if err != nil {
				return err
			}
		}
		Logger.log.Debugf("Transaction in block with hash", blockHash, "and index", index)
	}
	// Store Incomming Cross Shard
	if err := blockchain.CreateAndSaveCrossTransactionCoinViewPointFromBlock(block); err != nil {
		return err
	}
	err := blockchain.StoreIncomingCrossShard(block)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	// Save result of BurningConfirm instruction to get proof later
	err = blockchain.storeBurningConfirm(block)
	if err != nil {
		return err
	}
	// call FeeEstimator for processing
	if feeEstimator, ok := blockchain.config.FeeEstimator[block.Header.ShardID]; ok {
		err := feeEstimator.RegisterBlock(block)
		if err != nil {
			return NewBlockChainError(RegisterEstimatorFeeError, err)
		}
	}
	Logger.log.Criticalf("SHARD %+v | 🎉︎ %d transactions in block height %+v \n", block.Header.ShardID, len(block.Body.Transactions), block.Header.Height)
	if block.Header.Height != 1 {
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement:      metrics.TxInOneBlock,
			metrics.MeasurementValue: float64(len(block.Body.Transactions)),
			metrics.Tag:              metrics.BlockHeightTag,
			metrics.TagValue:         fmt.Sprintf("%d", block.Header.Height),
		})
	}
	return nil
}

func (blockchain *BlockChain) updateDatabaseWithTransactionMetadata(
	shardBlock *ShardBlock,
) error {
	db := blockchain.config.DataBase
	for _, tx := range shardBlock.Body.Transactions {
		metaType := tx.GetMetadataType()
		var err error
		if metaType == metadata.WithDrawRewardResponseMeta {
			_, requesterRes, amountRes, coinID := tx.GetTransferData()
			err = db.RemoveCommitteeReward(requesterRes, amountRes, *coinID)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

/*
	- Remove Staking TX in Shard BestState from instruction
	- Set Shard State for removing old Shard Block in Pool
	- Remove Old Cross Shard Block
	- Remove Init Tokens ID in Mempool
	- Remove Candiates in Mempool
	- Remove Transaction in Mempool and Block Generator
*/
func (blockchain *BlockChain) removeOldDataAfterProcessing(shardBlock *ShardBlock, shardID byte) {
	//remove staking txid in beststate shard
	go func() {
		for _, l := range shardBlock.Body.Instructions {
			if l[0] == SwapAction {
				swapedCommittees := strings.Split(l[2], ",")
				for _, v := range swapedCommittees {
					delete(GetBestStateShard(shardID).StakingTx, v)
				}
			}
		}
	}()
	//=========Remove invalid shard block in pool
	go blockchain.config.ShardPool[shardID].SetShardState(blockchain.BestState.Shard[shardID].ShardHeight)
	//updateShardBestState Cross shard pool: remove invalid block
	go func() {
		blockchain.config.CrossShardPool[shardID].RemoveBlockByHeight(blockchain.BestState.Shard[shardID].BestCrossShard)
	}()
	go func() {
		//Remove Candidate In pool
		candidates := []string{}
		tokenIDs := []string{}
		for _, tx := range shardBlock.Body.Transactions {
			if tx.GetMetadata() != nil {
				if tx.GetMetadata().GetType() == metadata.ShardStakingMeta || tx.GetMetadata().GetType() == metadata.BeaconStakingMeta {
					pubkey := base58.Base58Check{}.Encode(tx.GetSigPubKey(), common.ZeroByte)
					candidates = append(candidates, pubkey)
				}
			}
			if tx.GetType() == common.TxCustomTokenType {
				customTokenTx := tx.(*transaction.TxCustomToken)
				if customTokenTx.TxTokenData.Type == transaction.CustomTokenInit {
					tokenID := customTokenTx.TxTokenData.PropertyID.String()
					tokenIDs = append(tokenIDs, tokenID)
				}
			}
			if blockchain.config.IsBlockGenStarted {
				blockchain.config.CRemovedTxs <- tx
			}
		}
		go blockchain.config.TxPool.RemoveCandidateList(candidates)
		go blockchain.config.TxPool.RemoveTokenIDList(tokenIDs)

		//Remove tx out of pool
		go blockchain.config.TxPool.RemoveTx(shardBlock.Body.Transactions, true)
		for _, tx := range shardBlock.Body.Transactions {
			go func(tx metadata.Transaction) {
				if blockchain.config.IsBlockGenStarted {
					blockchain.config.CRemovedTxs <- tx
				}
			}(tx)
		}
	}()
}

func (blockchain *BlockChain) verifyCrossShardCustomToken(CrossTxTokenData map[byte][]CrossTxTokenData, shardID byte, txs []metadata.Transaction) error {
	txTokenDataListFromTxs := []transaction.TxTokenData{}
	_, txTokenDataList := blockchain.createCustomTokenTxForCrossShard(nil, CrossTxTokenData, shardID)
	hash, err := calHashFromTxTokenDataList(txTokenDataList)
	if err != nil {
		return err
	}
	for _, tx := range txs {
		if tx.GetType() == common.TxCustomTokenType {
			txCustomToken := tx.(*transaction.TxCustomToken)
			if txCustomToken.TxTokenData.Type == transaction.CustomTokenCrossShard {
				txTokenDataListFromTxs = append(txTokenDataListFromTxs, txCustomToken.TxTokenData)
			}
		}
	}
	hashFromTxs, err := calHashFromTxTokenDataList(txTokenDataListFromTxs)
	if err != nil {
		return err
	}
	if !hash.IsEqual(&hashFromTxs) {
		return errors.New("Cross Token Data from Cross Shard Block Not Compatible with Cross Token Data in New Block")
	}
	return nil
}

//=====================Util for shard====================
