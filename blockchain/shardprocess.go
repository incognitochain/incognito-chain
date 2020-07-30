package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/pkg/errors"
)

// VerifyPreSignShardBlock Verify Shard Block Before Signing
// Used for PBFT consensus
// this block doesn't have full information (incomplete block)
func (blockchain *BlockChain) VerifyPreSignShardBlock(shardBlock *ShardBlock, shardID byte) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	blockHeight := shardBlock.Header.Height
	blockHash := shardBlock.Header.Hash()
	Logger.log.Infof("SHARD %+v | Verify ShardBlock for signing process %d, with hash %+v", shardID, blockHeight, blockHash)
	// fetch beacon blocks
	previousBeaconHeight := blockchain.BestState.Shard[shardID].BeaconHeight
	if shardBlock.Header.BeaconHeight > blockchain.BestState.Beacon.BeaconHeight {
		return errors.New(fmt.Sprintf("Beacon %d not ready, latest is %d", shardBlock.Header.BeaconHeight, blockchain.BestState.Beacon.BeaconHeight))
	}
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.GetDatabase(), previousBeaconHeight+1, shardBlock.Header.BeaconHeight)
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
	if err := shardBestState.updateShardBestState(blockchain, shardBlock, beaconBlocks, newCommitteeChange()); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding shardBlock
	if err := shardBestState.verifyPostProcessingShardBlock(shardBlock, shardID); err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Block %d, with hash %+v is VALID for 🖋 signing", shardID, blockHeight, blockHash)
	return nil
}

// InsertShardBlock Insert Shard Block into blockchain
// this block must have full information (complete block)
func (blockchain *BlockChain) InsertShardBlock(shardBlock *ShardBlock, isValidated bool) error {
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	shardID := shardBlock.Header.ShardID
	blockHash := shardBlock.Header.Hash()
	blockHeight := shardBlock.Header.Height
	committeeChange := newCommitteeChange()
	shardLock := &blockchain.BestState.Shard[shardID].lock
	shardLock.Lock()
	defer shardLock.Unlock()
	if blockHeight != GetBestStateShard(shardID).ShardHeight+1 {
		return errors.New("Not expected height")
	}
	Logger.log.Criticalf("SHARD %+v | Begin insert new block height %+v with hash %+v", shardID, blockHeight, blockHash)
	Logger.log.Infof("SHARD %+v | Check block existence for insert height %+v with hash %+v", shardID, blockHeight, blockHash)
	tempShardBestState := blockchain.BestState.Shard[shardID]
	if tempShardBestState.ShardHeight == blockHeight && tempShardBestState.BestBlock.Header.Timestamp < shardBlock.Header.Timestamp && tempShardBestState.BestBlock.Header.Round < shardBlock.Header.Round {
		Logger.log.Infof("FORK SHARDID %+v, Current Block Height %+v, Block Hash %+v | Try To Insert New Shard Block Height %+v, Hash %+v", shardID, tempShardBestState.ShardHeight, tempShardBestState.BestBlockHash, blockHeight, blockHash)
		if err := blockchain.ValidateBlockWithPreviousShardBestState(shardBlock); err != nil {
			Logger.log.Error(err)
			return err
		}
		if err := blockchain.revertShardState(shardBlock.Header.ShardID); err != nil {
			Logger.log.Error(err)
			return err
		}
		Logger.log.Debugf("REVERTED SHARDID %+v, Revert Current Block Height %+v, Block Hash %+v", shardBlock.Header.ShardID, tempShardBestState.ShardHeight, tempShardBestState.BestBlockHash)
	}
	if blockHeight != GetBestStateShard(shardID).ShardHeight+1 {
		return errors.New("Not expected height")
	}
	// fetch beacon blocks
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.GetDatabase(), tempShardBestState.BeaconHeight+1, shardBlock.Header.BeaconHeight)
	if err != nil {
		return NewBlockChainError(FetchBeaconBlocksError, err)
	}
	if !isValidated {
		Logger.log.Infof("SHARD %+v | Verify Pre Processing, block height %+v with hash %+vt \n", shardID, blockHeight, blockHash)
		if err := blockchain.verifyPreProcessingShardBlock(shardBlock, beaconBlocks, shardID, false); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Pre Processing, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	}
	// Verify block with previous best state
	Logger.log.Infof("SHARD %+v | Verify BestState With Shard Block, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	if err := tempShardBestState.verifyBestStateWithShardBlock(shardBlock, true, shardID); err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | BackupCurrentShardState, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	// Backup beststate
	err = rawdbv2.CleanUpPreviousShardBestState(blockchain.GetDatabase(), shardID)
	if err != nil {
		return NewBlockChainError(CleanBackUpError, err)
	}
	err = blockchain.BackupCurrentShardState(shardBlock)
	if err != nil {
		return NewBlockChainError(BackUpBestStateError, err)
	}
	oldCommittee, err := incognitokey.CommitteeKeyListToString(tempShardBestState.ShardCommittee)
	if err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Update ShardBestState, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	if err := tempShardBestState.updateShardBestState(blockchain, shardBlock, beaconBlocks, committeeChange); err != nil {
		if errRevert := blockchain.revertShardBestState(shardID); errRevert != nil {
			return errRevert
		}
		return err
	}

	Logger.log.Infof("SHARD %+v | Update NumOfBlocksByProducers, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	// update number of blocks produced by producers to shard best state
	tempShardBestState.updateNumOfBlocksByProducers(shardBlock)
	newCommittee, err := incognitokey.CommitteeKeyListToString(tempShardBestState.ShardCommittee)
	if err != nil {
		if errRevert := blockchain.revertShardBestState(shardID); errRevert != nil {
			return errRevert
		}
		return err
	}
	if !common.CompareStringArray(oldCommittee, newCommittee) {
		go blockchain.config.ConsensusEngine.CommitteeChange(common.GetShardChainKey(shardID))
	}
	//========Post verification: verify new beaconstate with corresponding block
	if !isValidated {
		Logger.log.Infof("SHARD %+v | Verify Post Processing, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
		if err := tempShardBestState.verifyPostProcessingShardBlock(shardBlock, shardID); err != nil {
			if errRevert := blockchain.revertShardBestState(shardID); errRevert != nil {
				return errRevert
			}
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Post Processing, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	}
	Logger.log.Infof("SHARD %+v | Remove Data After Processed, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	Logger.log.Infof("SHARD %+v | Update Beacon Instruction, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	err = blockchain.processSalaryInstructions(tempShardBestState.rewardStateDB, beaconBlocks, shardID)
	if err != nil {
		if errRevert := blockchain.revertShardBestState(shardID); errRevert != nil {
			return errRevert
		}
		return err
	}
	Logger.log.Infof("SHARD %+v | Store New Shard Block And Update Data, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	//========Store new  Shard block and new shard bestState
	confirmBeaconBlock := NewBeaconBlock()
	if len(beaconBlocks) > 0 {
		confirmBeaconBlock = beaconBlocks[len(beaconBlocks)-1]
	} else {
		confirmBeaconBlock, _, err = blockchain.GetBeaconBlockByHash(shardBlock.Header.BeaconHash)
		if err != nil {
			if errRevert := blockchain.revertShardBestState(shardID); errRevert != nil {
				return errRevert
			}
			return err
		}
	}
	err = blockchain.processStoreShardBlock(shardBlock, committeeChange, confirmBeaconBlock)
	if err != nil {
		revertErr := blockchain.revertShardState(shardID)
		if revertErr != nil {
			return revertErr
		}
		return err
	}
	blockchain.removeOldDataAfterProcessingShardBlock(shardBlock, shardID)
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewShardblockTopic, shardBlock))
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardBeststateTopic, tempShardBestState))
	Logger.log.Infof("SHARD %+v | 🔗 Finish Insert new block %d, with hash %+v", shardID, blockHeight, blockHash)
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

// verifyPreProcessingShardBlock DOES NOT verify new block with best state
// DO NOT USE THIS with GENESIS BLOCK
// Verification condition:
//	- Producer Address is not empty
//	- ShardID: of received block same shardID
//	- Version: shard block version is one of pre-defined versions
//	- Parent (previous) block must be found in database ( current block point to an exist block in database )
//	- Height: parent block height + 1
//	- Epoch: blockHeight % Epoch ? Parent Epoch + 1 : Current Epoch
//	- Timestamp: block timestamp must be greater than previous block timestamp
//	- TransactionRoot: rebuild transaction root from txs in block and compare with transaction root in header
//	- ShardTxRoot: rebuild shard transaction root from txs in block and compare with shard transaction root in header
//	- CrossOutputCoinRoot: rebuild cross shard output root from cross output coin in block and compare with cross shard output coin
//		+ cross output coin must be re-created (from cross shard block) if verify block for signing
//	- InstructionRoot: rebuild instruction root from instructions and txs in block and compare with instruction root in header
//		+ instructions must be re-created (from beacon block and instruction) if verify block for signing
//	- InstructionMerkleRoot: rebuild instruction root from instructions and txs in block and compare with instruction root in header
//	- TotalTxFee: calculate tx token fee from all transaction in block then compare with header
//	- CrossShars: Verify Cross Shard Bitmap
//	- BeaconHeight: fetch beacon hash using beacon height in current shard block
//	- BeaconHash: compare beacon hash in database with beacon hash in shard block
//	- Verify swap instruction
//	- Validate transaction created from miner via instruction
//	- Validate Response Transaction From Transaction with Metadata
//	- ALL Transaction in block: see in verifyTransactionFromNewBlock
func (blockchain *BlockChain) verifyPreProcessingShardBlock(shardBlock *ShardBlock, beaconBlocks []*BeaconBlock, shardID byte, isPreSign bool) error {
	//verify producer sig
	tempShardBestState := blockchain.BestState.Shard[shardID]
	Logger.log.Debugf("SHARD %+v | Begin verifyPreProcessingShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	if shardBlock.Header.ShardID != shardID {
		return NewBlockChainError(WrongShardIDError, fmt.Errorf("Expect receive shardBlock from Shard ID %+v but get %+v", shardID, shardBlock.Header.ShardID))
	}
	if shardBlock.Header.Version != SHARD_BLOCK_VERSION {
		return NewBlockChainError(WrongVersionError, fmt.Errorf("Expect shardBlock version %+v but get %+v", SHARD_BLOCK_VERSION, shardBlock.Header.Version))
	}
	if shardBlock.Header.Height > tempShardBestState.ShardHeight+1 {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Expect shardBlock height %+v but get %+v", tempShardBestState.ShardHeight+1, shardBlock.Header.Height))
	}
	// Verify parent hash exist or not
	previousBlockHash := shardBlock.Header.PreviousBlockHash
	previousShardBlockData, err := rawdbv2.GetShardBlockByHash(blockchain.GetDatabase(), previousBlockHash)
	if err != nil {
		Logger.log.Criticalf("FORK SHARD DETECTED shardID=%+v at BlockHeight=%+v hash=%+v pre-hash=%+v",
			shardID,
			shardBlock.Header.Height,
			shardBlock.Hash().String(),
			previousBlockHash.String())
		blockchain.Synker.SyncBlkShard(shardID, true, false, false, []common.Hash{previousBlockHash}, nil, 0, 0, "")
		Logger.log.Critical("SEND REQUEST FOR BLOCK HASH", previousBlockHash.String(), shardBlock.Header.Height, shardBlock.Header.ShardID)
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
	_, shardTxMerkleData := CreateShardTxRoot(shardBlock.Body.Transactions)
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
	beaconHashs, err := rawdbv2.GetBeaconBlockHashByIndex(blockchain.GetDatabase(), shardBlock.Header.BeaconHeight)
	if err != nil {
		return NewBlockChainError(FetchBeaconBlockHashError, err)
	}
	beaconHash := beaconHashs[0]
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
	if isPreSign {
		err := blockchain.verifyPreProcessingShardBlockForSigning(shardBlock, beaconBlocks, txInstructions, shardID)
		if err != nil {
			return err
		}
	}
	Logger.log.Debugf("SHARD %+v | Finish verifyPreProcessingShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	return err
}

// VerifyPreProcessingShardBlockForSigning verify shard block before a validator signs new shard block
//	- Verify Transactions In New Block
//	- Generate Instruction (from beacon), create instruction root and compare instruction root with instruction root in header
//	- Get Cross Output Data from cross shard block (shard pool) and verify cross transaction hash
//	- Get Cross Tx Custom Token from cross shard block (shard pool) then verify
//
func (blockchain *BlockChain) verifyPreProcessingShardBlockForSigning(shardBlock *ShardBlock, beaconBlocks []*BeaconBlock, txInstructions [][]string, shardID byte) error {
	var err error
	var isOldBeaconHeight = false
	// Verify Transaction
	//get beacon height from shard block
	beaconHeight := shardBlock.Header.BeaconHeight
	if err := blockchain.verifyTransactionFromNewBlock(shardID, shardBlock.Body.Transactions, int64(beaconHeight)); err != nil {
		return NewBlockChainError(TransactionFromNewBlockError, err)
	}
	// Verify Instruction
	instructions := [][]string{}
	shardCommittee, err := incognitokey.CommitteeKeyListToString(blockchain.BestState.Shard[shardID].ShardCommittee)
	if err != nil {
		return err
	}
	shardPendingValidator, _, _ := blockchain.processInstructionFromBeacon(beaconBlocks, shardID, newCommitteeChange())
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
					nextHeight, err := rawdbv2.GetCrossShardNextHeight(blockchain.GetDatabase(), fromShard, toShard, startHeight)
					if err != nil {
						return NewBlockChainError(NextCrossShardBlockError, err)
					}
					if nextHeight != crossTransaction.BlockHeight {
						return NewBlockChainError(NextCrossShardBlockError, fmt.Errorf("Next Cross Shard Block Height %+v is Not Expected, Expect Next block Height %+v from shard %+v ", toShardCrossShardBlock.Header.Height, nextHeight, fromShard))
					}
					startHeight = nextHeight
					beaconHeight, err := blockchain.FindBeaconHeightForCrossShardBlock(toShardCrossShardBlock.Header.BeaconHeight, toShardCrossShardBlock.Header.ShardID, toShardCrossShardBlock.Header.Height)
					if err != nil {
						Logger.log.Errorf("%+v", err)
						break
					}
					beaconHeight = beaconHeight - 1
					consensusRootHash, err := blockchain.GetBeaconConsensusRootHash(blockchain.GetDatabase(), beaconHeight)
					if err != nil {
						Logger.log.Errorf("Can't found ConsensusStateRootHash of beacon height %+v, error %+v", beaconHeight, err)
						break
					}
					stateDB, err := statedb.NewWithPrefixTrie(consensusRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
					if err != nil {
						Logger.log.Errorf("Init trie err", err)
						break
					}
					shardCommittee := statedb.GetOneShardCommittee(stateDB, toShardCrossShardBlock.Header.ShardID)
					Logger.log.Criticalf("Shard %+v, committee %+v", toShardCrossShardBlock.Header.ShardID, shardCommittee)
					err = toShardCrossShardBlock.VerifyCrossShardBlock(blockchain, shardCommittee)
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

// verifyBestStateWithShardBlock will verify the validation of a block with some best state in cache or current best state
// Get beacon state of this block
// For example, new blockHeight is 91 then beacon state of this block must have height 90
// OR new block has previous has is beacon best block hash
//	- Producer
//	- committee length and validatorIndex length
//	- Producer + sig
//	- New Shard Block has parent (previous) hash is current shard state best block hash (compatible with current beststate)
//	- New Shard Block Height must be compatible with best shard state
//	- New Shard Block has beacon must higher or equal to beacon height of shard best state
func (shardBestState *ShardBestState) verifyBestStateWithShardBlock(shardBlock *ShardBlock, isVerifySig bool, shardID byte) error {
	Logger.log.Debugf("SHARD %+v | Begin VerifyBestStateWithShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	//verify producer via index
	producerPublicKey := shardBlock.Header.Producer
	producerPosition := shardBestState.GetProducerIndexFromBlock(shardBlock)

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

// updateShardBestState beststate with new shard block:
//	- New Previous Shard BlockHash
//	- New BestShardBlockHash
//	- New BestBeaconHash
//	- New Best Shard Block
//	- New Best Shard Height
//	- New Beacon Height
//	- ShardProposerIdx of new shard block
//	- Execute stake instruction, store staking transaction (if exist)
//	- Execute assign instruction, add new pending validator (if exist)
//	- Execute swap instruction, swap pending validator and committee (if exist)
func (shardBestState *ShardBestState) updateShardBestState(blockchain *BlockChain, shardBlock *ShardBlock, beaconBlocks []*BeaconBlock, committeeChange *committeeChange) error {
	Logger.log.Debugf("SHARD %+v | Begin update Beststate with new Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	var (
		err     error
		shardID = shardBlock.Header.ShardID
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
	} else { //TODO: 0xsirush revert this code
		shardBestState.ShardProposerIdx = 0 // (shardBestState.ShardProposerIdx + shardBlock.Header.Round) % len(shardBestState.ShardCommittee)
	}
	//shardBestState.processBeaconBlocks(shardBlock, beaconBlocks)
	shardPendingValidator, newShardPendingValidator, stakingTx := blockchain.processInstructionFromBeacon(beaconBlocks, shardBlock.Header.ShardID, committeeChange)
	shardBestState.ShardPendingValidator, err = incognitokey.CommitteeBase58KeyListToStruct(shardPendingValidator)
	if err != nil {
		return err
	}
	committeeChange.shardSubstituteAdded[shardID], err = incognitokey.CommitteeBase58KeyListToStruct(newShardPendingValidator)
	if err != nil {
		return err
	}
	for stakePublicKey, txHash := range stakingTx {
		shardBestState.StakingTx[stakePublicKey] = txHash
	}

	//hot fix, update shard staking tx
	if len(beaconBlocks) > 0 {
		replaceInThisBlock := false

		//incase, replace in this shard block (confirm beacon block with height NEWSTAKINGTX_HEIGHT_SWITCH)
		NEWSTAKINGTX_HEIGHT_SWITCH := blockchain.config.ChainParams.NewStakingTxBeaconHeighSwitch
		for _, beaconBlock := range beaconBlocks {
			if beaconBlock.GetHeight() == NEWSTAKINGTX_HEIGHT_SWITCH {
				replaceInThisBlock = true

				//retore staking for this shardID by reading json file
				type StakingInfo struct {
					Height        uint64
					StakingTxRoot map[int]string
					StakingTx     map[int]map[string]string
				}

				var CloneStakingTx = func(s *StakingInfo, sid int) map[string]string {
					res := make(map[string]string)
					for k, v := range s.StakingTx[sid] {
						res[k] = v
					}
					return res
				}

				stakingInfoResult := &StakingInfo{}

				fd, err := os.OpenFile("stakingtx.json", os.O_RDONLY, 0600)
				if err != nil {
					Logger.log.Error(err)
					return errors.New("Cannot open stakingtx.json")
				}

				b, err := ioutil.ReadAll(fd)
				if err != nil {
					Logger.log.Error(err)
					return errors.New("Cannot read stakingtx.json")
				}

				err = json.Unmarshal(b, stakingInfoResult)
				if err != nil {
					Logger.log.Error(err)
					return errors.New("Unmarshal stakingtx.json")
				}

				if stakingInfoResult.Height < 500000 {
					Logger.log.Error(err)
					return errors.New("Unepected height of stakingtx")
				}

				for sid, stakingMap := range stakingInfoResult.StakingTx {
					hash, err := GenerateHashFromMapStringString(stakingMap)
					if err != nil {
						panic(err)
					}
					if stakingInfoResult.StakingTxRoot[sid] != hash.String() {
						Logger.log.Error(err)
						return errors.New("Unepected stakingtx root hash")
					}
				}

				Logger.log.Infof("NEWTX: Replace stakingtx, shardID %v, root hash %v", int(shardBestState.ShardID), stakingInfoResult.StakingTxRoot[int(shardBestState.ShardID)])
				shardBestState.StakingTx = CloneStakingTx(stakingInfoResult, int(shardBestState.ShardID))
				break
			}
		}

		//incase, after the replacement
		if !replaceInThisBlock && beaconBlocks[0].GetHeight() > NEWSTAKINGTX_HEIGHT_SWITCH {
			for _, beaconBlock := range beaconBlocks {
				for _, l := range beaconBlock.Body.Instructions {
					// Process Swap Instruction, and delete stakingtx if we get return
					if l[0] == SwapAction {
						for _, outPublicKey := range strings.Split(l[2], ",") {
							if txId, ok := shardBestState.StakingTx[outPublicKey]; ok {
								if checkReturnStakingTxExistence(txId, shardBlock) {
									delete(shardBestState.StakingTx, outPublicKey)
								}
							}
						}
					}
				}
			}
		}
	}

	if common.IndexOfUint64(shardBlock.Header.BeaconHeight/blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.EpochBreakPointSwapNewKey) > -1 {
		err = shardBestState.processShardBlockInstructionForKeyListV2(blockchain, shardBlock, committeeChange)
	} else {
		err = shardBestState.processShardBlockInstruction(blockchain, shardBlock, committeeChange)
	}
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

func (shardBestState *ShardBestState) initShardBestState(blockchain *BlockChain, db incdb.Database, genesisShardBlock *ShardBlock, genesisBeaconBlock *BeaconBlock) error {
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
	shardPendingValidator, _, stakingTx := blockchain.processInstructionFromBeacon([]*BeaconBlock{genesisBeaconBlock}, genesisShardBlock.Header.ShardID, newCommitteeChange())

	shardPendingValidatorStr, err := incognitokey.CommitteeBase58KeyListToStruct(shardPendingValidator)
	if err != nil {
		return err
	}
	shardBestState.ShardPendingValidator = append(shardBestState.ShardPendingValidator, shardPendingValidatorStr...)
	for stakePublicKey, txHash := range stakingTx {
		shardBestState.StakingTx[stakePublicKey] = txHash
	}
	err = shardBestState.processShardBlockInstruction(blockchain, genesisShardBlock, newCommitteeChange())
	if err != nil {
		return err
	}
	shardBestState.ConsensusAlgorithm = common.BlsConsensus
	shardBestState.NumOfBlocksByProducers = make(map[string]uint64)
	//statedb===========================START
	dbAccessWarper := statedb.NewDatabaseAccessWarper(db)
	shardBestState.consensusStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	shardBestState.transactionStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	shardBestState.featureStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	shardBestState.rewardStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	shardBestState.slashStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	//statedb===========================END
	return nil
}

func (shardBestState *ShardBestState) processShardBlockInstruction(blockchain *BlockChain, shardBlock *ShardBlock, committeeChange *committeeChange) error {
	var err error
	shardID := shardBestState.ShardID
	shardPendingValidator, err := incognitokey.CommitteeKeyListToString(shardBestState.ShardPendingValidator)
	if err != nil {
		return err
	}
	shardCommittee, err := incognitokey.CommitteeKeyListToString(shardBestState.ShardCommittee)
	if err != nil {
		return err
	}
	// TODO: 0xmerman
	fixedProducerShardValidators := shardCommittee[:NumberOfFixedBlockValidators]
	shardCommittee = shardCommittee[NumberOfFixedBlockValidators:]

	shardSwappedCommittees := []string{}
	shardNewCommittees := []string{}
	if len(shardBlock.Body.Instructions) != 0 {
		Logger.log.Debugf("Shard Process/updateShardBestState: Shard Instruction %+v", shardBlock.Body.Instructions)
	}
	producersBlackList, err := blockchain.getUpdatedProducersBlackList(blockchain.BestState.Beacon.slashStateDB, false, int(shardID), shardCommittee, shardBlock.Header.BeaconHeight)
	if err != nil {
		return err
	}
	// Swap committee
	for _, l := range shardBlock.Body.Instructions {
		if l[0] == SwapAction {
			// #1 remaining pendingValidators, #2 new currentValidators #3 swapped out validator, #4 incoming validator
			maxShardCommitteeSize := shardBestState.MaxShardCommitteeSize - NumberOfFixedBlockValidators
			var minShardCommitteeSize int
			if blockchain.BestState.Shard[shardID].MinShardCommitteeSize-NumberOfFixedBlockValidators < 0 {
				minShardCommitteeSize = 0
			} else {
				minShardCommitteeSize = blockchain.BestState.Shard[shardID].MinShardCommitteeSize - NumberOfFixedBlockValidators
			}
			shardPendingValidator, shardCommittee, shardSwappedCommittees, shardNewCommittees, err = SwapValidator(shardPendingValidator, shardCommittee, maxShardCommitteeSize, minShardCommitteeSize, blockchain.config.ChainParams.Offset, producersBlackList, blockchain.config.ChainParams.SwapOffset)
			if err != nil {
				Logger.log.Errorf("SHARD %+v | Blockchain Error %+v", err)
				return NewBlockChainError(SwapValidatorError, err)
			}
			swapedCommittees := []string{}
			if len(l[2]) != 0 && l[2] != "" {
				swapedCommittees = strings.Split(l[2], ",")
			}
			//this code never run( checkReturnStakingTxExistence always false)
			//for _, v := range swapedCommittees {
			//	if txId, ok := shardBestState.StakingTx[v]; ok {
			//		if checkReturnStakingTxExistence(txId, shardBlock) {
			//			delete(GetBestStateShard(shardBestState.ShardID).StakingTx, v)
			//		}
			//	}
			//}
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
			shardNewCommitteesStruct, err := incognitokey.CommitteeBase58KeyListToStruct(shardNewCommittees)
			if err != nil {
				return err
			}
			shardSwappedCommitteesStruct, err := incognitokey.CommitteeBase58KeyListToStruct(shardSwappedCommittees)
			if err != nil {
				return err
			}
			beforeFilterShardSubstituteAdded := committeeChange.shardSubstituteAdded[shardID]
			filteredShardSubstituteAdded := []incognitokey.CommitteePublicKey{}
			filteredShardSubstituteRemoved := []incognitokey.CommitteePublicKey{}
			for _, currentShardSubstituteAdded := range beforeFilterShardSubstituteAdded {
				flag := false
				for _, newShardCommitteeAdded := range shardNewCommitteesStruct {
					if currentShardSubstituteAdded.IsEqual(newShardCommitteeAdded) {
						flag = true
						break
					}
				}
				if !flag {
					filteredShardSubstituteAdded = append(filteredShardSubstituteAdded, currentShardSubstituteAdded)
				}
			}
			for _, newShardCommitteeAdded := range shardNewCommitteesStruct {
				flag := false
				for _, currentShardSubstituteAdded := range beforeFilterShardSubstituteAdded {
					if currentShardSubstituteAdded.IsEqual(newShardCommitteeAdded) {
						flag = true
						break
					}
				}
				if !flag {
					filteredShardSubstituteRemoved = append(filteredShardSubstituteRemoved, newShardCommitteeAdded)
				}
			}
			committeeChange.shardCommitteeAdded[shardID] = shardNewCommitteesStruct
			committeeChange.shardCommitteeRemoved[shardID] = shardSwappedCommitteesStruct
			committeeChange.shardSubstituteRemoved[shardID] = filteredShardSubstituteRemoved
			committeeChange.shardSubstituteAdded[shardID] = filteredShardSubstituteAdded
			Logger.log.Infof("SHARD %+v | Swap: Out committee %+v", shardID, shardSwappedCommittees)
			Logger.log.Infof("SHARD %+v | Swap: In committee %+v", shardID, shardNewCommittees)
		}
	}
	shardBestState.ShardPendingValidator, err = incognitokey.CommitteeBase58KeyListToStruct(shardPendingValidator)
	if err != nil {
		return err
	}
	//TODO: merman
	shardBestState.ShardCommittee, err = incognitokey.CommitteeBase58KeyListToStruct(append(fixedProducerShardValidators, shardCommittee...))
	if err != nil {
		return err
	}
	return nil
}

func (shardBestState *ShardBestState) processShardBlockInstructionForKeyListV2(blockchain *BlockChain, shardBlock *ShardBlock, committeeChange *committeeChange) error {
	shardID := shardBlock.Header.ShardID
	for _, instruction := range shardBlock.Body.Instructions {
		if instruction[0] == SwapAction {
			shardPendingValidatorStruct := shardBestState.ShardPendingValidator
			inPublicKeys := strings.Split(instruction[1], ",")
			inPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(inPublicKeys)
			if err != nil {
				return err
			}
			outPublicKeys := strings.Split(instruction[2], ",")
			outPublicKeyStructs, err := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
			if err != nil {
				return err
			}
			inRewardReceiver := strings.Split(instruction[6], ",")
			if len(inPublicKeys) != len(outPublicKeys) {
				return NewBlockChainError(ProcessSwapInstructionError, fmt.Errorf("length new committee %+v, length out committee %+v", len(inPublicKeys), len(outPublicKeys)))
			}
			if len(inPublicKeys) != len(inRewardReceiver) {
				return NewBlockChainError(ProcessSwapInstructionError, fmt.Errorf("length new committee %+v, new reward receiver %+v", len(inPublicKeys), len(inRewardReceiver)))
			}
			removedCommitteeSize := len(inPublicKeys)
			remainedShardCommittees := shardBestState.ShardCommittee[removedCommitteeSize:]
			tempShardSwappedCommittees := shardBestState.ShardCommittee[:shardBestState.MinShardCommitteeSize]
			if !reflect.DeepEqual(outPublicKeyStructs, tempShardSwappedCommittees) {
				return NewBlockChainError(SwapValidatorError, fmt.Errorf("expect swapped committe %+v but got %+v", tempShardSwappedCommittees, outPublicKeyStructs))
			}
			shardCommitteesStruct := append(inPublicKeyStructs, remainedShardCommittees...)
			shardBestState.ShardPendingValidator = shardPendingValidatorStruct
			shardBestState.ShardCommittee = shardCommitteesStruct
			// committeeChange.shardCommitteeAdded[shardID] = inPublicKeyStructs
			// committeeChange.shardCommitteeRemoved[shardID] = outPublicKeyStructs
			committeeReplace := [2][]incognitokey.CommitteePublicKey{}
			committeeReplace[common.REPLACE_IN] = append(committeeChange.shardCommitteeAdded[shardID], inPublicKeyStructs...)
			committeeReplace[common.REPLACE_OUT] = append(committeeChange.shardCommitteeAdded[shardID], outPublicKeyStructs...)
			committeeChange.shardCommitteeReplaced[shardID] = committeeReplace
		}
	}
	return nil
}

// verifyPostProcessingShardBlock
//	- commitee root
//	- pending validator root
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
	if hash, isOk := verifyHashFromStringArray(shardPendingValidatorStr, shardBlock.Header.PendingValidatorRoot); !isOk {
		return NewBlockChainError(ShardPendingValidatorRootHashError, fmt.Errorf("Expect shard pending validator root hash to be %+v but get %+v", shardBlock.Header.PendingValidatorRoot, hash))
	}

	if hash, isOk := verifyHashFromMapStringString(shardBestState.StakingTx, shardBlock.Header.StakingTxRoot); !isOk {
		return NewBlockChainError(ShardPendingValidatorRootHashError, fmt.Errorf("Expect shard %v staking root hash to be %+v but get %+v", shardID, shardBlock.Header.StakingTxRoot, hash))
	}

	Logger.log.Debugf("SHARD %+v | Finish VerifyPostProcessing Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	return nil
}

// Verify Transaction with these condition:
//	1. Validate tx version
//	2. Validate fee with tx size
//	3. Validate type of tx
//	4. Validate with other txs in block:
// 	- Normal Transaction:
// 	- Custom Tx:
//	4.1 Validate Init Custom Token
//	5. Validate sanity data of tx
//	6. Validate data in tx: privacy proof, metadata,...
//	7. Validate tx with blockchain: douple spend, ...
//	8. Check tx existed in block
//	9. Not accept a salary tx
//	10. Check duplicate staker public key in block
//	11. Check duplicate Init Custom Token in block
func (blockchain *BlockChain) verifyTransactionFromNewBlock(shardID byte, txs []metadata.Transaction, beaconHeight int64) error {
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
	_, err := blockchain.config.TempTxPool.MaybeAcceptBatchTransactionForBlockProducing(shardID, listTxs, beaconHeight)
	if err != nil {
		Logger.log.Errorf("Batching verify transactions from new block err: %+v\n Trying verify one by one", err)
		for index, tx := range listTxs {
			if blockchain.config.TempTxPool.HaveTransaction(tx.Hash()) {
				continue
			}
			_, err1 := blockchain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx, beaconHeight)
			if err1 != nil {
				Logger.log.Errorf("One by one verify txs at index %d error: %+v", index, err1)
				return NewBlockChainError(TransactionFromNewBlockError, fmt.Errorf("Transaction %+v, index %+v get %+v ", *tx.Hash(), index, err1))
			}
		}
	}
	return nil
}

// processStoreShardBlock Store All information after Insert
//	- Shard Block
//	- Shard Best State
//	- Transaction => UTXO, serial number, snd, commitment
//	- Cross Output Coin => UTXO, snd, commmitment
//	- Store transaction metadata:
//		+ Withdraw Metadata
//	- Store incoming cross shard block
//	- Store Burning Confirmation
//	- Update Mempool fee estimator
func (blockchain *BlockChain) processStoreShardBlock(shardBlock *ShardBlock, committeeChange *committeeChange, beaconBlock *BeaconBlock) error {
	shardID := shardBlock.Header.ShardID
	blockHeight := shardBlock.Header.Height
	blockHash := shardBlock.Header.Hash()
	rewardReceiver := make(map[string]string)
	autoStaking := make(map[string]bool)
	tempShardBestState := blockchain.BestState.Shard[shardID]
	Logger.log.Infof("SHARD %+v | Process store block height %+v at hash %+v", shardBlock.Header.ShardID, blockHeight, *shardBlock.Hash())
	if len(shardBlock.Body.CrossTransactions) != 0 {
		Logger.log.Critical("processStoreShardBlock/CrossTransactions	", shardBlock.Body.CrossTransactions)
	}
	featureStateRootHash, err := blockchain.GetBeaconFeatureRootHash(blockchain.GetDatabase(), beaconBlock.Header.Height)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, fmt.Errorf("can't get ConsensusStateRootHash of height %+v ,error %+v", beaconBlock.Header.Height, err))
	}
	featureStateDB, err := statedb.NewWithPrefixTrie(featureStateRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := blockchain.CreateAndSaveTxViewPointFromBlock(shardBlock, tempShardBestState.transactionStateDB, featureStateDB); err != nil {
		return NewBlockChainError(FetchAndStoreTransactionError, err)
	}

	for index, tx := range shardBlock.Body.Transactions {
		if err := rawdbv2.StoreTransactionIndex(blockchain.GetDatabase(), *tx.Hash(), shardBlock.Header.Hash(), index); err != nil {
			return NewBlockChainError(FetchAndStoreTransactionError, err)
		}
		// Process Transaction Metadata
		metaType := tx.GetMetadataType()
		if metaType == metadata.WithDrawRewardResponseMeta {
			_, publicKey, amountRes, coinID := tx.GetTransferData()
			err = statedb.RemoveCommitteeReward(tempShardBestState.rewardStateDB, publicKey, amountRes, *coinID)
			if err != nil {
				return NewBlockChainError(RemoveCommitteeRewardError, err)
			}
		}
		Logger.log.Debugf("Transaction in block with hash", blockHash, "and index", index)
	}
	// Store Incomming Cross Shard
	if err := blockchain.CreateAndSaveCrossTransactionViewPointFromBlock(shardBlock, tempShardBestState.transactionStateDB); err != nil {
		return NewBlockChainError(FetchAndStoreCrossTransactionError, err)
	}
	// Save result of BurningConfirm instruction to get proof later
	err = blockchain.storeBurningConfirm(tempShardBestState.featureStateDB, shardBlock)
	if err != nil {
		return NewBlockChainError(StoreBurningConfirmError, err)
	}
	// Update bridge issuance request status
	err = blockchain.updateBridgeIssuanceStatus(tempShardBestState.featureStateDB, shardBlock)
	if err != nil {
		return NewBlockChainError(UpdateBridgeIssuanceStatusError, err)
	}
	// call FeeEstimator for processing
	if feeEstimator, ok := blockchain.config.FeeEstimator[shardBlock.Header.ShardID]; ok {
		err := feeEstimator.RegisterBlock(shardBlock)
		if err != nil {
			Logger.log.Debug(NewBlockChainError(RegisterEstimatorFeeError, err))
		}
	}
	if len(committeeChange.shardCommitteeAdded[shardID]) > 0 || len(committeeChange.shardSubstituteAdded[shardID]) > 0 {
		consensusStateRootHash, err := blockchain.GetBeaconConsensusRootHash(blockchain.GetDatabase(), beaconBlock.Header.Height)
		if err != nil {
			return NewBlockChainError(StoreShardBlockError, fmt.Errorf("can't get ConsensusStateRootHash of height %+v ,error %+v", beaconBlock.Header.Height, err))
		}
		consensusStateDB, err := statedb.NewWithPrefixTrie(consensusStateRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetDatabase()))
		if err != nil {
			return NewBlockChainError(StoreShardBlockError, err)
		}
		rewardReceiver, autoStaking = statedb.GetRewardReceiverAndAutoStaking(consensusStateDB, blockchain.GetShardIDs())
		//statedb===========================START
		err = statedb.StoreOneShardCommittee(tempShardBestState.consensusStateDB, shardID, committeeChange.shardCommitteeAdded[shardID], rewardReceiver, autoStaking)
		if err != nil {
			return NewBlockChainError(StoreShardBlockError, err)
		}
		err = statedb.StoreOneShardSubstitutesValidator(tempShardBestState.consensusStateDB, shardID, committeeChange.shardSubstituteAdded[shardID], rewardReceiver, autoStaking)
		if err != nil {
			return NewBlockChainError(StoreShardBlockError, err)
		}
	}
	if common.IndexOfUint64(shardBlock.Header.BeaconHeight/blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.EpochBreakPointSwapNewKey) > -1 {
		for _, instruction := range shardBlock.Body.Instructions {
			if instruction[0] == SwapAction {
				inRewardReceiver := strings.Split(instruction[6], ",")
				outPublicKeys := strings.Split(instruction[2], ",")
				outPublicKeyStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(outPublicKeys)
				inPublicKeys := strings.Split(instruction[1], ",")
				inPublicKeyStructs, _ := incognitokey.CommitteeBase58KeyListToStruct(inPublicKeys)
				removedCommittee := len(inPublicKeys)
				for i := 0; i < removedCommittee; i++ {
					delete(autoStaking, outPublicKeys[i])
					delete(rewardReceiver, outPublicKeyStructs[i].GetIncKeyBase58())
					autoStaking[inPublicKeys[i]] = false
					rewardReceiver[inPublicKeyStructs[i].GetIncKeyBase58()] = inRewardReceiver[i]
				}
				break
			}
		}
		err = statedb.ReplaceOneShardCommittee(tempShardBestState.consensusStateDB, shardID, committeeChange.shardCommitteeReplaced[shardID], rewardReceiver, autoStaking)
		if err != nil {
			return NewBlockChainError(StoreShardBlockError, err)
		}
	}
	err = statedb.DeleteOneShardCommittee(tempShardBestState.consensusStateDB, shardID, committeeChange.shardCommitteeAdded[shardID])
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = statedb.DeleteOneShardSubstitutesValidator(tempShardBestState.consensusStateDB, shardID, committeeChange.shardSubstituteAdded[shardID])
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	// consensus root hash
	consensusRootHash, err := tempShardBestState.consensusStateDB.Commit(true)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = tempShardBestState.consensusStateDB.Database().TrieDB().Commit(consensusRootHash, false)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	// transaction root hash
	transactionRootHash, err := tempShardBestState.transactionStateDB.Commit(true)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = tempShardBestState.transactionStateDB.Database().TrieDB().Commit(transactionRootHash, false)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	// feature root hash
	featureRootHash, err := tempShardBestState.featureStateDB.Commit(true)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = tempShardBestState.featureStateDB.Database().TrieDB().Commit(featureRootHash, false)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	// reward root hash
	rewardRootHash, err := tempShardBestState.rewardStateDB.Commit(true)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = tempShardBestState.rewardStateDB.Database().TrieDB().Commit(rewardRootHash, false)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	// slash root hash
	slashRootHash, err := tempShardBestState.slashStateDB.Commit(true)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = tempShardBestState.slashStateDB.Database().TrieDB().Commit(slashRootHash, false)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	tempShardBestState.consensusStateDB.ClearObjects()
	tempShardBestState.transactionStateDB.ClearObjects()
	tempShardBestState.featureStateDB.ClearObjects()
	tempShardBestState.rewardStateDB.ClearObjects()
	tempShardBestState.slashStateDB.ClearObjects()
	batchData := blockchain.GetDatabase().NewBatch()
	if err := rawdbv2.StoreShardConsensusRootHash(batchData, shardID, blockHeight, consensusRootHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardTransactionRootHash(batchData, shardID, blockHeight, transactionRootHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardFeatureRootHash(batchData, shardID, blockHeight, featureRootHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardCommitteeRewardRootHash(batchData, shardID, blockHeight, rewardRootHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardSlashRootHash(batchData, shardID, blockHeight, slashRootHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	//statedb===========================END
	if err := rawdbv2.StoreShardBlock(batchData, shardID, blockHeight, blockHash, shardBlock); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardBlockIndex(batchData, shardID, blockHeight, blockHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardBestState(batchData, shardID, tempShardBestState); err != nil {
		return NewBlockChainError(StoreBestStateError, err)
	}
	if err := batchData.Write(); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	Logger.log.Infof("SHARD %+v | 🔎 %d transactions in block height %+v \n", shardBlock.Header.ShardID, len(shardBlock.Body.Transactions), blockHeight)
	return nil
}

// removeOldDataAfterProcessingShardBlock remove outdate data from pool and beststate
//	- Remove Staking TX in Shard BestState from instruction
//	- Set Shard State for removing old Shard Block in Pool
//	- Remove Old Cross Shard Block
//	- Remove Init Tokens ID in Mempool
//	- Remove Candiates in Mempool
//	- Remove Transaction in Mempool and Block Generator
func (blockchain *BlockChain) removeOldDataAfterProcessingShardBlock(shardBlock *ShardBlock, shardID byte) {
	// Remove invalid shard block in pool
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
