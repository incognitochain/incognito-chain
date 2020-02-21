package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/database"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/pubsub"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
)

/*
	Verify Shard Block Before Signing
	Used for PBFT consensus
	@Notice: this block doesn't have full information (incomplete block)
*/
func (blockchain *BlockChain) VerifyPreSignShardBlock(shardBlock *ShardBlock, shardID byte) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	Logger.log.Infof("SHARD %+v | Verify ShardBlock for signing process %d, with hash %+v", shardID, shardBlock.Header.Height, *shardBlock.Hash())
	// fetch beacon blocks

	previousBeaconHeight := blockchain.BestState.Shard[shardID].BeaconHeight
	if shardBlock.Header.BeaconHeight > blockchain.BestState.Beacon.BeaconHeight {
		return errors.New(fmt.Sprintf("Beacon %d not ready, latest is %d", shardBlock.Header.BeaconHeight, blockchain.BestState.Beacon.BeaconHeight))
	}

	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.config.DataBase, previousBeaconHeight+1, shardBlock.Header.BeaconHeight)
	if err != nil {
		return err
	}
	//========Verify shardBlock only
	if err := blockchain.verifyPreProcessingShardBlock(shardBlock, beaconBlocks, shardID, true); err != nil {
		return err
	}
	//========Verify shardBlock with previous best state
	// Get Beststate of previous shardBlock == previous best state
	// Clone best state value into new variable
	shardBestState := NewShardBestState()
	if err := shardBestState.cloneShardBestStateFrom(blockchain.BestState.Shard[shardID]); err != nil {
		return err
	}
	// Verify shardBlock with previous best state
	// DO NOT verify agg signature in this function
	if err := shardBestState.verifyBestStateWithShardBlock(shardBlock, false, shardID); err != nil {
		return err
	}
	//========updateShardBestState best state with new shardBlock
	if err := shardBestState.updateShardBestState(blockchain, shardBlock, beaconBlocks); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding shardBlock
	if err := shardBestState.verifyPostProcessingShardBlock(shardBlock, shardID); err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Block %d, with hash %+v is VALID for 🖋 signing", shardID, shardBlock.Header.Height, *shardBlock.Hash())
	return nil
}

/*
	Insert Shard Block into blockchain
	@Notice: this block must have full information (complete block)
*/
func (blockchain *BlockChain) InsertShardBlock(shardBlock *ShardBlock, isValidated bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	shardID := shardBlock.Header.ShardID
	blockHash := shardBlock.Header.Hash()

	shardLock := &blockchain.BestState.Shard[shardID].lock
	shardLock.Lock()
	defer shardLock.Unlock()

	if shardBlock.Header.Height != GetBestStateShard(shardID).ShardHeight+1 {
		return errors.New("Not expected height")
	}

	Logger.log.Criticalf("SHARD %+v | Begin insert new block height %+v with hash %+v", shardID, shardBlock.Header.Height, blockHash)
	Logger.log.Infof("SHARD %+v | Check block existence for insert height %+v with hash %+v", shardID, shardBlock.Header.Height, blockHash)
	currentShardBestState := blockchain.BestState.Shard[shardBlock.Header.ShardID]

	if currentShardBestState.ShardHeight == shardBlock.Header.Height && currentShardBestState.BestBlock.Header.Timestamp < shardBlock.Header.Timestamp && currentShardBestState.BestBlock.Header.Round < shardBlock.Header.Round {
		currentShardHeight := currentShardBestState.ShardHeight
		currentShardHash := currentShardBestState.BestBlockHash
		Logger.log.Infof("FORK SHARDID %+v, Current Block Height %+v, Block Hash %+v | Try To Insert New Shard Block Height %+v, Hash %+v", shardBlock.Header.ShardID, currentShardBestState.ShardHeight, currentShardBestState.BestBlockHash, shardBlock.Header.Height, shardBlock.Header.Hash())
		if err := blockchain.ValidateBlockWithPrevShardBestState(shardBlock); err != nil {
			Logger.log.Error(err)
			return err
		}
		if err := blockchain.RevertShardState(shardBlock.Header.ShardID); err != nil {
			panic(err)
		}
		Logger.log.Infof("REVERTED SHARDID %+v, Revert Current Block Height %+v, Block Hash %+v", shardBlock.Header.ShardID, currentShardHeight, currentShardHash)
	}

	if shardBlock.Header.Height != GetBestStateShard(shardID).ShardHeight+1 {
		return errors.New("Not expected height")
	}
	// force non-committee member not to validate blk
	// if blockchain.config.UserKeySet != nil && (blockchain.config.NodeMode == common.NODEMODE_AUTO || blockchain.config.NodeMode == common.NODEMODE_SHARD) {
	// 	userRole := blockchain.BestState.Shard[block.Header.ShardID].GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyInBase58CheckEncode(), 0)
	// 	fmt.Println("Shard block received 1", userRole)

	// 	if userRole != common.PROPOSER_ROLE && userRole != common.VALIDATOR_ROLE && userRole != common.PENDING_ROLE {
	// 		isValidated = true
	// 	}
	// } else {
	// 	isValidated = true
	// }
	isExist, _ := blockchain.config.DataBase.HasBlock(blockHash)
	if isExist {
		return NewBlockChainError(DuplicateShardBlockError, fmt.Errorf("SHARD %+v, block height %+v wit hash %+v has been stored already", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash))
	}
	// fetch beacon blocks
	previousBeaconHeight := blockchain.BestState.Shard[shardID].BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.config.DataBase, previousBeaconHeight+1, shardBlock.Header.BeaconHeight)
	if err != nil {
		return NewBlockChainError(FetchBeaconBlocksError, err)
	}
	if !isValidated {
		Logger.log.Infof("SHARD %+v | Verify Pre Processing, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
		if err := blockchain.verifyPreProcessingShardBlock(shardBlock, beaconBlocks, shardID, false); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Pre Processing, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	}
	// Verify block with previous best state
	Logger.log.Infof("SHARD %+v | Verify BestState With Shard Block, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	if err := blockchain.BestState.Shard[shardID].verifyBestStateWithShardBlock(shardBlock, true, shardID); err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | BackupCurrentShardState, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	// Backup beststate
	err = blockchain.config.DataBase.CleanBackup(false, shardBlock.Header.ShardID)
	if err != nil {
		return NewBlockChainError(CleanBackUpError, err)
	}
	err = blockchain.BackupCurrentShardState(shardBlock, beaconBlocks)
	if err != nil {
		return NewBlockChainError(BackUpBestStateError, err)
	}

	oldCommittee, err := incognitokey.CommitteeKeyListToString(blockchain.BestState.Shard[shardID].ShardCommittee)
	if err != nil {
		return err
	}

	Logger.log.Infof("SHARD %+v | Update ShardBestState, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	if err := blockchain.BestState.Shard[shardID].updateShardBestState(blockchain, shardBlock, beaconBlocks); err != nil {
		errRevert := blockchain.revertShardBestState(shardID)
		if errRevert != nil {
			return errors.WithStack(errRevert)
		}
		return err
	}

	Logger.log.Infof("SHARD %+v | Update NumOfBlocksByProducers, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	// update number of blocks produced by producers to shard best state
	blockchain.BestState.Shard[shardID].updateNumOfBlocksByProducers(shardBlock)

	newCommittee, err := incognitokey.CommitteeKeyListToString(blockchain.BestState.Shard[shardID].ShardCommittee)
	if err != nil {
		return err
	}
	if !common.CompareStringArray(oldCommittee, newCommittee) {
		go blockchain.config.ConsensusEngine.CommitteeChange(common.GetShardChainKey(shardID))
	}

	//========Post verififcation: verify new beaconstate with corresponding block
	if !isValidated {
		Logger.log.Infof("SHARD %+v | Verify Post Processing, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
		if err := blockchain.BestState.Shard[shardID].verifyPostProcessingShardBlock(shardBlock, shardID); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Post Processing, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	}
	Logger.log.Infof("SHARD %+v | Remove Data After Processed, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	blockchain.removeOldDataAfterProcessingShardBlock(shardBlock, shardID)
	Logger.log.Infof("SHARD %+v | Update Beacon Instruction, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	err = blockchain.updateDatabaseFromBeaconInstructions(beaconBlocks, shardID)
	if err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Store New Shard Block And Update Data, block height %+v with hash %+v \n", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	//========Store new  Shard block and new shard bestState
	err = blockchain.processStoreShardBlockAndUpdateDatabase(shardBlock)
	if err != nil {
		revertErr := blockchain.revertShardState(shardID)
		if revertErr != nil {
			return errors.WithStack(revertErr)
		}
		return err
	}
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewShardblockTopic, shardBlock))
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardBeststateTopic, blockchain.BestState.Shard[shardID]))
	//shardIDForMetric := strconv.Itoa(int(shardBlock.Header.ShardID))
	//go metrics.AnalyzeTimeSeriesMetricDataWithTime(map[string]interface{}{
	//	metrics.Measurement:      metrics.NumOfBlockInsertToChain,
	//	metrics.MeasurementValue: float64(1),
	//	metrics.Tag:              metrics.ShardIDTag,
	//	metrics.TagValue:         metrics.Shard + shardIDForMetric,
	//	metrics.Time:             shardBlock.Header.Timestamp,
	//})
	//if shardBlock.Header.Height > 2 {
	//	go metrics.AnalyzeTimeSeriesMetricDataWithTime(map[string]interface{}{
	//		metrics.Measurement:      metrics.NumOfRoundPerBlock,
	//		metrics.MeasurementValue: float64(shardBlock.Header.Round),
	//		metrics.Tag:              metrics.ShardIDTag,
	//		metrics.TagValue:         metrics.Shard + shardIDForMetric,
	//		metrics.Time:             shardBlock.Header.Timestamp,
	//	})
	//}
	//if shardBlock.Header.Height != 1 {
	//	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
	//		metrics.Measurement:      metrics.TxInOneBlock,
	//		metrics.MeasurementValue: float64(len(shardBlock.Body.Transactions)),
	//		metrics.Tag:              metrics.BlockHeightTag,
	//		metrics.TagValue:         fmt.Sprintf("%d-%d", shardBlock.Header.ShardID, shardBlock.Header.Height),
	//	})
	//}
	Logger.log.Infof("SHARD %+v | 🔗 Finish Insert new block %d, with hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	return nil
}

// updateNumOfBlocksByProducers updates number of blocks produced by producers to shard best state
func (shardBestState *ShardBestState) updateNumOfBlocksByProducers(shardBlock *ShardBlock) {
	isSwapInstContained := false
	for _, inst := range shardBlock.Body.Instructions {
		if len(inst) > 0 && inst[0] == SwapAction {
			isSwapInstContained = true
			break
		}
	}
	producer := shardBlock.GetProducerPubKeyStr()
	if isSwapInstContained {
		// reset number of blocks produced by producers
		shardBestState.NumOfBlocksByProducers = map[string]uint64{
			producer: 1,
		}
	} else {
		// Update number of blocks produced by producers in epoch
		numOfBlks, found := shardBestState.NumOfBlocksByProducers[producer]
		if !found {
			shardBestState.NumOfBlocksByProducers[producer] = 1
		} else {
			shardBestState.NumOfBlocksByProducers[producer] = numOfBlks + 1
		}
	}
}

/* Verify Pre-prosessing data
This function DOES NOT verify new block with best state
DO NOT USE THIS with GENESIS BLOCK
	Condition:
	- Producer Address is not empty
	- ShardID: of received block same shardID
	- Version: shard block version is one of pre-defined versions
	- Parent (previous) block must be found in database ( current block point to an exist block in database )
	- Height: parent block height + 1
	- Epoch: blockHeight % Epoch ? Parent Epoch + 1 : Current Epoch
	- Timestamp: block timestamp must be greater than previous block timestamp
	- TransactionRoot: rebuild transaction root from txs in block and compare with transaction root in header
	- ShardTxRoot: rebuild shard transaction root from txs in block and compare with shard transaction root in header
	- CrossOutputCoinRoot: rebuild cross shard output root from cross output coin in block and compare with cross shard output coin
		+ cross output coin must be re-created (from cross shard block) if verify block for signing
	- InstructionRoot: rebuild instruction root from instructions and txs in block and compare with instruction root in header
		+ instructions must be re-created (from beacon block and instruction) if verify block for signing
	- InstructionMerkleRoot: rebuild instruction root from instructions and txs in block and compare with instruction root in header
	- TotalTxFee: calculate tx token fee from all transaction in block then compare with header
	- CrossShars: Verify Cross Shard Bitmap
	- BeaconHeight: fetch beacon hash using beacon height in current shard block
	- BeaconHash: compare beacon hash in database with beacon hash in shard block
	- Verify swap instruction
	- Validate transaction created from miner via instruction
	- Validate Response Transaction From Transaction with Metadata
	- ALL Transaction in block: see in verifyTransactionFromNewBlock
*/
func (blockchain *BlockChain) verifyPreProcessingShardBlock(shardBlock *ShardBlock, beaconBlocks []*BeaconBlock, shardID byte, isPreSign bool) error {
	//verify producer sig
	Logger.log.Debugf("SHARD %+v | Begin verifyPreProcessingShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	if shardBlock.Header.ShardID != shardID {
		return NewBlockChainError(WrongShardIDError, fmt.Errorf("Expect receive shardBlock from Shard ID %+v but get %+v", shardID, shardBlock.Header.ShardID))
	}
	// if len(shardBlock.Header.ProducerAddress.Bytes()) != 66 {
	// 	return NewBlockChainError(ProducerError, fmt.Errorf("Expect has length 66 but get %+v", len(shardBlock.Header.ProducerAddress.Bytes())))
	// }
	if shardBlock.Header.Version != SHARD_BLOCK_VERSION {
		return NewBlockChainError(WrongVersionError, fmt.Errorf("Expect shardBlock version %+v but get %+v", SHARD_BLOCK_VERSION, shardBlock.Header.Version))
	}

	if shardBlock.Header.Height > blockchain.BestState.Shard[shardID].ShardHeight+1 {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Expect shardBlock height %+v but get %+v", blockchain.BestState.Shard[shardID].ShardHeight+1, shardBlock.Header.Height))
	}
	// Verify parent hash exist or not
	previousBlockHash := shardBlock.Header.PreviousBlockHash
	previousShardBlockData, err := blockchain.config.DataBase.FetchBlock(previousBlockHash)
	if err != nil {
		Logger.log.Criticalf("FORK SHARD DETECTED shardID=%+v at BlockHeight=%+v hash=%+v pre-hash=%+v",
			shardID,
			shardBlock.Header.Height,
			shardBlock.Hash().String(),
			previousBlockHash.String())
		blockchain.Synker.SyncBlkShard(shardID, true, false, false, []common.Hash{previousBlockHash}, nil, 0, 0, "")
		Logger.log.Critical("SEND REQUEST FOR BLOCK HASH", previousBlockHash.String(), shardBlock.Header.Height, shardBlock.Header.ShardID)
		if !isPreSign {
			revertErr := blockchain.revertShardState(shardID)
			if revertErr != nil {
				Logger.log.Error("blockchain.revertShardState error", revertErr)
				return errors.WithStack(revertErr)
			}
		}
		return NewBlockChainError(FetchPreviousBlockError, err)
	}
	previousShardBlock := ShardBlock{}
	err = json.Unmarshal(previousShardBlockData, &previousShardBlock)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	// Verify shardBlock height with parent shardBlock
	if previousShardBlock.Header.Height+1 != shardBlock.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Expect receive shardBlock height %+v but get %+v", previousShardBlock.Header.Height+1, shardBlock.Header.Height))
	}
	// Verify timestamp with parent shardBlock
	if shardBlock.Header.Timestamp <= previousShardBlock.Header.Timestamp {
		return NewBlockChainError(WrongTimestampError, fmt.Errorf("Expect receive shardBlock has timestamp must be greater than %+v but get %+v", previousShardBlock.Header.Timestamp, shardBlock.Header.Timestamp))
	}
	// Verify transaction root
	txMerkleTree := Merkle{}.BuildMerkleTreeStore(shardBlock.Body.Transactions)
	txRoot := &common.Hash{}
	if len(txMerkleTree) > 0 {
		txRoot = txMerkleTree[len(txMerkleTree)-1]
	}
	if !bytes.Equal(shardBlock.Header.TxRoot.GetBytes(), txRoot.GetBytes()) {
		return NewBlockChainError(TransactionRootHashError, fmt.Errorf("Expect transaction root hash %+v but get %+v", shardBlock.Header.TxRoot, txRoot))
	}
	// Verify ShardTx Root
	_, shardTxMerkleData := CreateShardTxRoot2(shardBlock.Body.Transactions)
	shardTxRoot := shardTxMerkleData[len(shardTxMerkleData)-1]
	if !bytes.Equal(shardBlock.Header.ShardTxRoot.GetBytes(), shardTxRoot.GetBytes()) {
		return NewBlockChainError(ShardTransactionRootHashError, fmt.Errorf("Expect shard transaction root hash %+v but get %+v", shardBlock.Header.ShardTxRoot, shardTxRoot))
	}
	// Verify crossTransaction coin
	if !VerifyMerkleCrossTransaction(shardBlock.Body.CrossTransactions, shardBlock.Header.CrossTransactionRoot) {
		return NewBlockChainError(CrossShardTransactionRootHashError, fmt.Errorf("Expect cross shard transaction root hash %+v", shardBlock.Header.CrossTransactionRoot))
	}
	// Verify Action
	txInstructions, err := CreateShardInstructionsFromTransactionAndInstruction(shardBlock.Body.Transactions, blockchain, shardID)
	if err != nil {
		Logger.log.Error(err)
		return NewBlockChainError(ShardIntructionFromTransactionAndInstructionError, err)
	}
	if !isPreSign {
		totalInstructions := []string{}
		for _, value := range txInstructions {
			totalInstructions = append(totalInstructions, value...)
		}
		for _, value := range shardBlock.Body.Instructions {
			totalInstructions = append(totalInstructions, value...)
		}
		if hash, ok := verifyHashFromStringArray(totalInstructions, shardBlock.Header.InstructionsRoot); !ok {
			return NewBlockChainError(InstructionsHashError, fmt.Errorf("Expect instruction hash to be %+v but get %+v at block %+v hash %+v", shardBlock.Header.InstructionsRoot, hash, shardBlock.Header.Height, shardBlock.Hash().String()))
		}
	}
	totalTxsFee := make(map[common.Hash]uint64)
	for _, tx := range shardBlock.Body.Transactions {
		totalTxsFee[*tx.GetTokenID()] += tx.GetTxFee()
		txType := tx.GetType()
		if txType == common.TxCustomTokenPrivacyType {
			txCustomPrivacy := tx.(*transaction.TxCustomTokenPrivacy)
			totalTxsFee[*txCustomPrivacy.GetTokenID()] = txCustomPrivacy.GetTxFeeToken()
		}
	}
	tokenIDsfromTxs := make([]common.Hash, 0)
	for tokenID := range totalTxsFee {
		tokenIDsfromTxs = append(tokenIDsfromTxs, tokenID)
	}
	sort.Slice(tokenIDsfromTxs, func(i int, j int) bool {
		res, _ := tokenIDsfromTxs[i].Cmp(&tokenIDsfromTxs[j])
		return res == -1
	})
	tokenIDsfromBlock := make([]common.Hash, 0)
	for tokenID := range shardBlock.Header.TotalTxsFee {
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
		if totalTxsFee[tokenID] != shardBlock.Header.TotalTxsFee[tokenID] {
			return NewBlockChainError(WrongBlockTotalFeeError, fmt.Errorf("Expect Total Fee to be equal, From Txs %+v, From Block Header %+v", totalTxsFee[tokenID], shardBlock.Header.TotalTxsFee[tokenID]))
		}
	}
	// Verify Cross Shards
	crossShards := CreateCrossShardByteArray(shardBlock.Body.Transactions, shardID)
	if len(crossShards) != len(shardBlock.Header.CrossShardBitMap) {
		return NewBlockChainError(CrossShardBitMapError, fmt.Errorf("Expect number of cross shardID is %+v but get %+v", len(shardBlock.Header.CrossShardBitMap), len(crossShards)))
	}
	for index := range crossShards {
		if crossShards[index] != shardBlock.Header.CrossShardBitMap[index] {
			return NewBlockChainError(CrossShardBitMapError, fmt.Errorf("Expect Cross Shard Bitmap of shardID %+v is %+v but get %+v", index, shardBlock.Header.CrossShardBitMap[index], crossShards[index]))
		}
	}
	// Check if InstructionMerkleRoot is the root of merkle tree containing all instructions in this shardBlock
	flattenTxInsts, err := FlattenAndConvertStringInst(txInstructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from Tx: %+v", err))
	}
	flattenInsts, err := FlattenAndConvertStringInst(shardBlock.Body.Instructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from shardBlock body: %+v", err))
	}
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be the same as when creating new shard shardBlock
	root := GetKeccak256MerkleRoot(insts)
	if !bytes.Equal(root, shardBlock.Header.InstructionMerkleRoot[:]) {
		return NewBlockChainError(InstructionMerkleRootError, fmt.Errorf("Expect transaction merkle root to be %+v but get %+v", shardBlock.Header.InstructionMerkleRoot, string(root)))
	}
	//Get beacon hash by height in db
	//If hash not found then fail to verify
	beaconHash, err := blockchain.config.DataBase.GetBeaconBlockHashByIndex(shardBlock.Header.BeaconHeight)
	if err != nil {
		return NewBlockChainError(FetchBeaconBlockHashError, err)
	}
	//Hash in db must be equal to hash in shard shardBlock
	newHash, err := common.Hash{}.NewHash(shardBlock.Header.BeaconHash.GetBytes())
	if err != nil {
		return NewBlockChainError(HashError, err)
	}
	if !newHash.IsEqual(&beaconHash) {
		return NewBlockChainError(BeaconBlockNotCompatibleError, fmt.Errorf("Expect beacon shardBlock hash to be %+v but get %+v", beaconHash, newHash))
	}
	// Swap instruction
	for _, l := range shardBlock.Body.Instructions {
		if l[0] == "swap" {
			if l[3] != "shard" || l[4] != strconv.Itoa(int(shardID)) {
				return NewBlockChainError(SwapInstructionError, fmt.Errorf("invalid swap instruction %+v", l))
			}
		}
	}
	// Verify response transactions
	instsForValidations := [][]string{}
	instsForValidations = append(instsForValidations, shardBlock.Body.Instructions...)
	for _, beaconBlock := range beaconBlocks {
		instsForValidations = append(instsForValidations, beaconBlock.Body.Instructions...)
	}
	invalidTxs, err := blockchain.verifyMinerCreatedTxBeforeGettingInBlock(instsForValidations, shardBlock.Body.Transactions, shardID)
	if err != nil {
		return NewBlockChainError(TransactionCreatedByMinerError, err)
	}
	if len(invalidTxs) > 0 {
		return NewBlockChainError(TransactionCreatedByMinerError, fmt.Errorf("There are %d invalid txs", len(invalidTxs)))
	}
	err = blockchain.ValidateResponseTransactionFromTxsWithMetadata(shardBlock)
	if err != nil {
		return NewBlockChainError(ResponsedTransactionWithMetadataError, err)
	}
	// Get cross shard shardBlock from pool
	// @NOTICE: COMMENT to bypass verify cross shard shardBlock
	if isPreSign {
		err := blockchain.verifyPreProcessingShardBlockForSigning(shardBlock, beaconBlocks, txInstructions, shardID)
		if err != nil {
			return err
		}
	}
	Logger.log.Debugf("SHARD %+v | Finish verifyPreProcessingShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	return err
}

/*
	VerifyPreProcessingShardBlockForSigning verify shard block before a validator signs new shard block
	- Verify Transactions In New Block
	- Generate Instruction (from beacon), create instruction root and compare instruction root with instruction root in header
	- Get Cross Output Data from cross shard block (shard pool) and verify cross transaction hash
	- Get Cross Tx Custom Token from cross shard block (shard pool) then verify
*/
func (blockchain *BlockChain) verifyPreProcessingShardBlockForSigning(shardBlock *ShardBlock, beaconBlocks []*BeaconBlock, txInstructions [][]string, shardID byte) error {
	var err error
	var isOldBeaconHeight = false
	// Verify Transaction
	//get beacon height from shard block
	beaconHeight := shardBlock.Header.BeaconHeight
	if err := blockchain.verifyTransactionFromNewBlock(shardBlock.Body.Transactions, int64(beaconHeight)); err != nil {
		return NewBlockChainError(TransactionFromNewBlockError, err)
	}
	// Verify Instruction
	instructions := [][]string{}
	shardCommittee, err := incognitokey.CommitteeKeyListToString(blockchain.BestState.Shard[shardID].ShardCommittee)
	if err != nil {
		return err
	}
	shardPendingValidator, _ := blockchain.processInstructionFromBeacon(beaconBlocks, shardID)
	if blockchain.BestState.Shard[shardID].BeaconHeight == shardBlock.Header.BeaconHeight {
		isOldBeaconHeight = true
	}
	instructions, shardPendingValidator, shardCommittee, err = blockchain.generateInstruction(shardID, shardBlock.Header.BeaconHeight, isOldBeaconHeight, beaconBlocks, shardPendingValidator, shardCommittee)
	if err != nil {
		return NewBlockChainError(GenerateInstructionError, err)
	}
	totalInstructions := []string{}
	for _, value := range txInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	for _, value := range instructions {
		totalInstructions = append(totalInstructions, value...)
	}
	if hash, ok := verifyHashFromStringArray(totalInstructions, shardBlock.Header.InstructionsRoot); !ok {
		return NewBlockChainError(InstructionsHashError, fmt.Errorf("Expect instruction hash to be %+v but %+v", shardBlock.Header.InstructionsRoot, hash))
	}
	// Verify Cross Shard Output Coin and Custom Token Transaction
	toShard := shardID
	crossShardLimit := blockchain.config.CrossShardPool[toShard].GetLatestValidBlockHeight()
	toShardAllCrossShardBlock := blockchain.config.CrossShardPool[toShard].GetValidBlock(crossShardLimit)
	for fromShard, crossTransactions := range shardBlock.Body.CrossTransactions {
		toShardCrossShardBlocks, ok := toShardAllCrossShardBlock[fromShard]
		if !ok {
			heights := []uint64{}
			for _, crossTransaction := range crossTransactions {
				heights = append(heights, crossTransaction.BlockHeight)
			}
			blockchain.Synker.SyncBlkCrossShard(false, false, []common.Hash{}, heights, fromShard, shardID, "")
			return NewBlockChainError(CrossShardBlockError, fmt.Errorf("Cross Shard Block From Shard %+v Not Found in Pool", fromShard))
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
						return NewBlockChainError(NextCrossShardBlockError, err)
					}
					if nextHeight != crossTransaction.BlockHeight {
						return NewBlockChainError(NextCrossShardBlockError, fmt.Errorf("Next Cross Shard Block Height %+v is Not Expected, Expect Next block Height %+v from shard %+v ", toShardCrossShardBlock.Header.Height, nextHeight, fromShard))
					}
					startHeight = nextHeight
					beaconHeight, err := blockchain.FindBeaconHeightForCrossShardBlock(toShardCrossShardBlock.Header.BeaconHeight, toShardCrossShardBlock.Header.ShardID, toShardCrossShardBlock.Header.Height)
					if err != nil {
						break
					}
					temp, err := blockchain.config.DataBase.FetchShardCommitteeByHeight(beaconHeight)
					if err != nil {
						return NewBlockChainError(FetchShardCommitteeError, err)
					}
					shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
					err = json.Unmarshal(temp, &shardCommittee)
					if err != nil {
						return NewBlockChainError(UnmashallJsonShardCommitteesError, err)
					}
					err = toShardCrossShardBlock.VerifyCrossShardBlock(blockchain, shardCommittee[toShardCrossShardBlock.Header.ShardID])
					if err != nil {
						return NewBlockChainError(VerifyCrossShardBlockError, err)
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
						return NewBlockChainError(CrossTransactionHashError, fmt.Errorf("Cross Output Coin From New Block %+v not compatible with cross shard block in pool %+v", targetHash, hash))
					}
					if true {
						toShardCrossShardBlocks = toShardCrossShardBlocks[index:]
						isValids++
						break
					}
				}
			}
		}
		if len(crossTransactions) != isValids {
			return NewBlockChainError(CrossShardBlockError, fmt.Errorf("Can't not verify all cross shard block from shard %+v", fromShard))
		}
	}
	return nil
}

/*
	This function will verify the validation of a block with some best state in cache or current best state
	Get beacon state of this block
	For example, new blockHeight is 91 then beacon state of this block must have height 90
	OR new block has previous has is beacon best block hash
	- Producer
	- committee length and validatorIndex length
	- Producer + sig
	- New Shard Block has parent (previous) hash is current shard state best block hash (compatible with current beststate)
	- New Shard Block Height must be compatible with best shard state
	- New Shard Block has beacon must higher or equal to beacon height of shard best state
*/
func (shardBestState *ShardBestState) verifyBestStateWithShardBlock(shardBlock *ShardBlock, isVerifySig bool, shardID byte) error {
	Logger.log.Debugf("SHARD %+v | Begin VerifyBestStateWithShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	//verify producer via index
	producerPublicKey := shardBlock.Header.Producer
	producerPosition := (shardBestState.ShardProposerIdx + shardBlock.Header.Round) % len(shardBestState.ShardCommittee)

	//verify producer
	tempProducer, err := shardBestState.ShardCommittee[producerPosition].ToBase58() //.GetMiningKeyBase58(common.BridgeConsensus)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	if strings.Compare(tempProducer, producerPublicKey) != 0 {
		return NewBlockChainError(ProducerError, fmt.Errorf("Producer should be should be %+v", tempProducer))
	}
	//=============End Verify producer signature
	//=============Verify aggegrate signature
	// if isVerifySig {
	// TODO: validator index condition
	// if len(shardBestState.ShardCommittee) > 3 && len(shardBlock.ValidatorsIdx[1]) < (len(shardBestState.ShardCommittee)>>1) {
	// 	return NewBlockChainError(ShardCommitteeLengthAndCommitteeIndexError, fmt.Errorf("Expect Number of Committee Size greater than 3 but get %+v", len(shardBestState.ShardCommittee)))
	// }
	// err := ValidateAggSignature(shardBlock.ValidatorsIdx, shardBestState.ShardCommittee, shardBlock.AggregatedSig, shardBlock.R, shardBlock.Hash())
	// if err != nil {
	// 	return err
	// }
	// }
	//=============End Verify Aggegrate signature
	// check with current final best state
	// shardBlock can only be insert if it match the current best state
	if !shardBestState.BestBlockHash.IsEqual(&shardBlock.Header.PreviousBlockHash) {
		return NewBlockChainError(ShardBestStateNotCompatibleError, fmt.Errorf("Current Best Block Hash %+v, New Shard Block %+v, Previous Block Hash of New Block %+v", shardBestState.BestBlockHash, shardBlock.Header.Height, shardBlock.Header.PreviousBlockHash))
	}
	if shardBestState.ShardHeight+1 != shardBlock.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Shard Block height of new Shard Block should be %+v, but get %+v", shardBestState.ShardHeight+1, shardBlock.Header.Height))
	}
	if shardBlock.Header.BeaconHeight < shardBestState.BeaconHeight {
		return NewBlockChainError(ShardBestStateBeaconHeightNotCompatibleError, fmt.Errorf("Shard Block contain invalid beacon height, current beacon height %+v but get %+v ", shardBestState.BeaconHeight, shardBlock.Header.BeaconHeight))
	}
	Logger.log.Debugf("SHARD %+v | Finish VerifyBestStateWithShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	return nil
}

/*
	updateShardBestState beststate with new shard block:
	- New Previous Shard BlockHash
	- New BestShardBlockHash
	- New BestBeaconHash
	- New Best Shard Block
	- New Best Shard Height
	- New Beacon Height
	- ShardProposerIdx of new shard block
	- Execute stake instruction, store staking transaction (if exist)
	- Execute assign instruction, add new pending validator (if exist)
	- Execute swap instruction, swap pending validator and committee (if exist)
*/
func (shardBestState *ShardBestState) updateShardBestState(blockchain *BlockChain, shardBlock *ShardBlock, beaconBlocks []*BeaconBlock) error {
	Logger.log.Debugf("SHARD %+v | Begin update Beststate with new Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	var (
		err error
	)
	shardBestState.BestBlockHash = *shardBlock.Hash()
	shardBestState.BestBeaconHash = shardBlock.Header.BeaconHash
	shardBestState.BestBlock = shardBlock
	shardBestState.BestBlockHash = *shardBlock.Hash()
	shardBestState.ShardHeight = shardBlock.Header.Height
	shardBestState.Epoch = shardBlock.Header.Epoch
	shardBestState.BeaconHeight = shardBlock.Header.BeaconHeight
	shardBestState.TotalTxns += uint64(len(shardBlock.Body.Transactions))
	shardBestState.NumTxns = uint64(len(shardBlock.Body.Transactions))
	if shardBlock.Header.Height == 1 {
		shardBestState.ShardProposerIdx = 0
	} else {
		shardBestState.ShardProposerIdx = (shardBestState.ShardProposerIdx + shardBlock.Header.Round) % len(shardBestState.ShardCommittee)
	}
	//shardBestState.processBeaconBlocks(shardBlock, beaconBlocks)
	shardPendingValidator, stakingTx := blockchain.processInstructionFromBeacon(beaconBlocks, shardBlock.Header.ShardID)
	shardBestState.ShardPendingValidator, err = incognitokey.CommitteeBase58KeyListToStruct(shardPendingValidator)
	if err != nil {
		return err
	}
	for stakePublicKey, txHash := range stakingTx {
		shardBestState.StakingTx[stakePublicKey] = txHash
	}
	err = shardBestState.processShardBlockInstruction(blockchain, shardBlock)
	if err != nil {
		return err
	}
	//updateShardBestState best cross shard
	for shardID, crossShardBlock := range shardBlock.Body.CrossTransactions {
		shardBestState.BestCrossShard[shardID] = crossShardBlock[len(crossShardBlock)-1].BlockHeight
	}
	//======BEGIN For testing and benchmark
	temp := 0
	for _, tx := range shardBlock.Body.Transactions {
		//detect transaction that's not salary
		if !tx.IsSalaryTx() {
			temp++
		}
	}
	shardBestState.TotalTxnsExcludeSalary += uint64(temp)
	//======END
	Logger.log.Debugf("SHARD %+v | Finish update Beststate with new Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	return nil
}
func (shardBestState *ShardBestState) initShardBestState(blockchain *BlockChain, genesisShardBlock *ShardBlock, genesisBeaconBlock *BeaconBlock) error {
	shardBestState.BestBeaconHash = *ChainTestParam.GenesisBeaconBlock.Hash()
	shardBestState.BestBlock = genesisShardBlock
	shardBestState.BestBlockHash = *genesisShardBlock.Hash()
	shardBestState.ShardHeight = genesisShardBlock.Header.Height
	shardBestState.Epoch = genesisShardBlock.Header.Epoch
	shardBestState.BeaconHeight = genesisShardBlock.Header.BeaconHeight
	shardBestState.TotalTxns += uint64(len(genesisShardBlock.Body.Transactions))
	shardBestState.NumTxns = uint64(len(genesisShardBlock.Body.Transactions))
	shardBestState.ShardProposerIdx = 0
	//shardBestState.processBeaconBlocks(genesisShardBlock, []*BeaconBlock{genesisBeaconBlock})
	shardPendingValidator, stakingTx := blockchain.processInstructionFromBeacon([]*BeaconBlock{genesisBeaconBlock}, genesisShardBlock.Header.ShardID)

	shardPendingValidatorStr, err := incognitokey.CommitteeBase58KeyListToStruct(shardPendingValidator)
	if err != nil {
		return err
	}
	shardBestState.ShardPendingValidator = append(shardBestState.ShardPendingValidator, shardPendingValidatorStr...)
	for stakePublicKey, txHash := range stakingTx {
		shardBestState.StakingTx[stakePublicKey] = txHash
	}
	err = shardBestState.processShardBlockInstruction(blockchain, genesisShardBlock)
	if err != nil {
		return err
	}
	shardBestState.ConsensusAlgorithm = common.BlsConsensus
	shardBestState.NumOfBlocksByProducers = make(map[string]uint64)
	return nil
}

func (shardBestState *ShardBestState) processShardBlockInstruction(blockchain *BlockChain, shardBlock *ShardBlock) error {
	var err error
	shardPendingValidator, err := incognitokey.CommitteeKeyListToString(shardBestState.ShardPendingValidator)
	if err != nil {
		return err
	}
	shardID := shardBestState.ShardID
	shardCommittee, err := incognitokey.CommitteeKeyListToString(shardBestState.ShardCommittee)
	if err != nil {
		return err
	}
	shardSwappedCommittees := []string{}
	shardNewCommittees := []string{}
	if len(shardBlock.Body.Instructions) != 0 {
		Logger.log.Info("Shard Process/updateShardBestState: Shard Instruction", shardBlock.Body.Instructions)
	}

	producersBlackList, err := blockchain.getUpdatedProducersBlackList(false, int(shardID), shardCommittee, shardBlock.Header.BeaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	// Swap committee
	for _, l := range shardBlock.Body.Instructions {
		if l[0] == SwapAction {
			// #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator
			shardPendingValidator, shardCommittee, shardSwappedCommittees, shardNewCommittees, err = SwapValidator(shardPendingValidator, shardCommittee, shardBestState.MaxShardCommitteeSize, shardBestState.MinShardCommitteeSize, blockchain.config.ChainParams.Offset, producersBlackList, blockchain.config.ChainParams.SwapOffset)
			if err != nil {
				Logger.log.Errorf("SHARD %+v | Blockchain Error %+v", err)
				return NewBlockChainError(SwapValidatorError, err)
			}
			swapedCommittees := []string{}
			if len(l[2]) != 0 && l[2] != "" {
				swapedCommittees = strings.Split(l[2], ",")
			}

			for _, v := range swapedCommittees {
				if txId, ok := shardBestState.StakingTx[v]; ok {
					if checkReturnStakingTxExistence(txId, shardBlock) {
						delete(GetBestStateShard(shardBestState.ShardID).StakingTx, v)
					}
				}
			}
			if !reflect.DeepEqual(swapedCommittees, shardSwappedCommittees) {
				return NewBlockChainError(SwapValidatorError, fmt.Errorf("Expect swapped committees to be %+v but get %+v", swapedCommittees, shardSwappedCommittees))
			}

			newCommittees := []string{}
			if len(l[1]) > 0 {
				newCommittees = strings.Split(l[1], ",")
			}
			if !reflect.DeepEqual(newCommittees, shardNewCommittees) {
				return NewBlockChainError(SwapValidatorError, fmt.Errorf("Expect new committees to be %+v but get %+v", newCommittees, shardNewCommittees))
			}
			Logger.log.Infof("SHARD %+v | Swap: Out committee %+v", shardBlock.Header.ShardID, shardSwappedCommittees)
			Logger.log.Infof("SHARD %+v | Swap: In committee %+v", shardBlock.Header.ShardID, shardNewCommittees)
		}
	}
	shardBestState.ShardPendingValidator, err = incognitokey.CommitteeBase58KeyListToStruct(shardPendingValidator)
	if err != nil {
		return err
	}
	shardBestState.ShardCommittee, err = incognitokey.CommitteeBase58KeyListToStruct(shardCommittee)
	if err != nil {
		return err
	}
	return nil
}

/*
	verifyPostProcessingShardBlock
	- commitee root
	- pending validator root
*/
func (shardBestState *ShardBestState) verifyPostProcessingShardBlock(shardBlock *ShardBlock, shardID byte) error {
	Logger.log.Debugf("SHARD %+v | Begin VerifyPostProcessing Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())

	shardCommitteeStr, err := incognitokey.CommitteeKeyListToString(shardBestState.ShardCommittee)
	if err != nil {
		return err
	}
	if hash, ok := verifyHashFromStringArray(shardCommitteeStr, shardBlock.Header.CommitteeRoot); !ok {
		return NewBlockChainError(ShardCommitteeRootHashError, fmt.Errorf("Expect shard committee root hash to be %+v but get %+v", shardBlock.Header.CommitteeRoot, hash))
	}

	shardPendingValidatorStr, err := incognitokey.CommitteeKeyListToString(shardBestState.ShardPendingValidator)
	if err != nil {
		return err
	}
	if hash, ok := verifyHashFromStringArray(shardPendingValidatorStr, shardBlock.Header.PendingValidatorRoot); !ok {
		return NewBlockChainError(ShardPendingValidatorRootHashError, fmt.Errorf("Expect shard pending validator root hash to be %+v but get %+v", shardBlock.Header.PendingValidatorRoot, hash))
	}
	tempHash, isOk := verifyHashFromMapStringString(shardBestState.StakingTx, shardBlock.Header.StakingTxRoot)
	if !isOk {
		return NewBlockChainError(ShardPendingValidatorRootHashError, fmt.Errorf("Expect shard pending validator root hash to be %+v but get %+v", shardBlock.Header.PendingValidatorRoot, tempHash))
	}
	Logger.log.Debugf("SHARD %+v | Finish VerifyPostProcessing Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
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
func (blockchain *BlockChain) verifyTransactionFromNewBlock(txs []metadata.Transaction, beaconHeight int64) error {
	if len(txs) == 0 {
		return nil
	}
	isEmpty := blockchain.config.TempTxPool.EmptyPool()
	if !isEmpty {
		panic("TempTxPool Is not Empty")
	}
	defer blockchain.config.TempTxPool.EmptyPool()
	listTxs := []metadata.Transaction{}
	for _, tx := range txs {
		if !tx.IsSalaryTx() {
			listTxs = append(listTxs, tx)
		}
	}
	_, err := blockchain.config.TempTxPool.MaybeAcceptBatchTransactionForBlockProducing(listTxs, beaconHeight)
	if err != nil {
		for index, tx := range listTxs {
			if blockchain.config.TempTxPool.HaveTransaction(tx.Hash()) {
				continue
			}
			_, err1 := blockchain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx, beaconHeight)
			if err1 != nil {
				return NewBlockChainError(TransactionFromNewBlockError, fmt.Errorf("Transaction %+v, index %+v get %+v ", *tx.Hash(), index, err1))
			}
		}
	}
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
func (blockchain *BlockChain) processStoreShardBlockAndUpdateDatabase(shardBlock *ShardBlock) error {
	blockHash := shardBlock.Hash().String()
	batchPutData := []database.BatchData{}

	Logger.log.Infof("SHARD %+v | Process store block height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, *shardBlock.Hash())
	if err := blockchain.StoreShardBlock(shardBlock, &batchPutData); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := blockchain.StoreShardBlockIndex(shardBlock, &batchPutData); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := blockchain.StoreShardBestState(shardBlock.Header.ShardID, &batchPutData); err != nil {
		return NewBlockChainError(StoreBestStateError, err)
	}

	if len(shardBlock.Body.CrossTransactions) != 0 {
		Logger.log.Critical("processStoreShardBlockAndUpdateDatabase/CrossTransactions	", shardBlock.Body.CrossTransactions)
	}
	if err := blockchain.CreateAndSaveTxViewPointFromBlock(shardBlock, &batchPutData); err != nil {
		return NewBlockChainError(FetchAndStoreTransactionError, err)
	}

	for index, tx := range shardBlock.Body.Transactions {
		if err := blockchain.StoreTransactionIndex(tx.Hash(), shardBlock.Header.Hash(), index, &batchPutData); err != nil {
			Logger.log.Errorf("Transaction in block with hash %+v and index %+v: %+v, err %+v", blockHash, index, tx, err)
			return NewBlockChainError(FetchAndStoreTransactionError, err)
		}
		// Process Transaction Metadata
		metaType := tx.GetMetadataType()
		if metaType == metadata.WithDrawRewardResponseMeta {
			_, requesterRes, amountRes, coinID := tx.GetTransferData()
			err := blockchain.config.DataBase.RemoveCommitteeReward(requesterRes, amountRes, *coinID, &batchPutData)
			if err != nil {
				return NewBlockChainError(RemoveCommitteeRewardError, err)
			}
		}
		Logger.log.Debugf("Transaction in block with hash", blockHash, "and index", index)
	}
	// Store Incomming Cross Shard
	if err := blockchain.CreateAndSaveCrossTransactionCoinViewPointFromBlock(shardBlock, &batchPutData); err != nil {
		return NewBlockChainError(FetchAndStoreCrossTransactionError, err)
	}
	err := blockchain.StoreIncomingCrossShard(shardBlock, &batchPutData)
	if err != nil {
		return NewBlockChainError(StoreIncomingCrossShardError, err)
	}
	// Save result of BurningConfirm instruction to get proof later
	err = blockchain.storeBurningConfirm(shardBlock, &batchPutData)
	if err != nil {
		return NewBlockChainError(StoreBurningConfirmError, err)
	}

	// Update bridge issuance request status
	err = blockchain.updateBridgeIssuanceStatus(shardBlock, &batchPutData)
	if err != nil {
		return NewBlockChainError(UpdateBridgeIssuanceStatusError, err)
	}

	// call FeeEstimator for processing
	if feeEstimator, ok := blockchain.config.FeeEstimator[shardBlock.Header.ShardID]; ok {
		err := feeEstimator.RegisterBlock(shardBlock)
		if err != nil {
			Logger.log.Warn(NewBlockChainError(RegisterEstimatorFeeError, err))
		}
	}
	Logger.log.Infof("SHARD %+v | 🔎 %d transactions in block height %+v \n", shardBlock.Header.ShardID, len(shardBlock.Body.Transactions), shardBlock.Header.Height)
	return blockchain.config.DataBase.PutBatch(batchPutData)
	//return nil
}

//func (blockchain *BlockChain) updateDatabaseWithTransactionMetadata(shardBlock *ShardBlock) error {
//	db := blockchain.config.DataBase
//	for _, tx := range shardBlock.Body.Transactions {
//		metaType := tx.GetMetadataType()
//		var err error
//		if metaType == metadata.WithDrawRewardResponseMeta {
//			_, requesterRes, amountRes, coinID := tx.GetTransferData()
//			err = db.RemoveCommitteeReward(requesterRes, amountRes, *coinID)
//			if err != nil {
//				return NewBlockChainError(RemoveCommitteeRewardError, err)
//			}
//		}
//	}
//	return nil
//}

/*
	- Remove Staking TX in Shard BestState from instruction
	- Set Shard State for removing old Shard Block in Pool
	- Remove Old Cross Shard Block
	- Remove Init Tokens ID in Mempool
	- Remove Candiates in Mempool
	- Remove Transaction in Mempool and Block Generator
*/
func (blockchain *BlockChain) removeOldDataAfterProcessingShardBlock(shardBlock *ShardBlock, shardID byte) {
	//remove staking txid in beststate shard
	//go func() {
	//	for _, l := range shardBlock.Body.Instructions {
	//		if l[0] == SwapAction {
	//			swapedCommittees := strings.Split(l[2], ",")
	//			for _, v := range swapedCommittees {
	//				delete(GetBestStateShard(shardID).StakingTx, v)
	//			}
	//		}
	//	}
	//}()
	//=========Remove invalid shard block in pool
	go blockchain.config.ShardPool[shardID].SetShardState(blockchain.BestState.Shard[shardID].ShardHeight)
	//updateShardBestState Cross shard pool: remove invalid block
	go func() {
		blockchain.config.CrossShardPool[shardID].RemoveBlockByHeight(blockchain.BestState.Shard[shardID].BestCrossShard)
	}()
	go func() {
		//Remove Candidate In pool
		candidates := []string{}
		for _, tx := range shardBlock.Body.Transactions {
			if blockchain.config.IsBlockGenStarted {
				blockchain.config.CRemovedTxs <- tx
			}
			if tx.GetMetadata() != nil {
				if tx.GetMetadata().GetType() == metadata.ShardStakingMeta || tx.GetMetadata().GetType() == metadata.BeaconStakingMeta {
					stakingMetadata, ok := tx.GetMetadata().(*metadata.StakingMetadata)
					if !ok {
						continue
					}
					candidates = append(candidates, stakingMetadata.CommitteePublicKey)
				}
			}
		}
		go blockchain.config.TxPool.RemoveCandidateList(candidates)

		//Remove tx out of pool
		go blockchain.config.TxPool.RemoveTx(shardBlock.Body.Transactions, true)
	}()
}

//=====================Util for shard====================
