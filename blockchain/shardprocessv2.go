package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/incdb"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/pkg/errors"
)

//VerifyPreSignShardBlockV2 Verify Shard Block Before Signing
//Used for PBFT consensus
//@Notice: this block doesn't have full information (incomplete block)
func (blockchain *BlockChain) VerifyPreSignShardBlockV2(shardBlock *ShardBlock, shardID byte) error {
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
	beaconBlocks, err := FetchBeaconBlockFromHeightV2(blockchain.GetDatabase(), previousBeaconHeight+1, shardBlock.Header.BeaconHeight)
	if err != nil {
		return err
	}
	//========Verify shardBlock only
	if err := blockchain.verifyPreProcessingShardBlockV2(shardBlock, beaconBlocks, shardID, true); err != nil {
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
	if err := shardBestState.updateShardBestStateV2(blockchain, shardBlock, beaconBlocks, newCommitteeChange()); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding shardBlock
	if err := shardBestState.verifyPostProcessingShardBlock(shardBlock, shardID); err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Block %d, with hash %+v is VALID for 🖋 signing", shardID, blockHeight, blockHash)
	return nil
}

//InsertShardBlockV2 Insert Shard Block into blockchain
//@Notice: this block must have full information (complete block)
func (blockchain *BlockChain) InsertShardBlockV2(shardBlock *ShardBlock, isValidated bool) error {
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
		if err := blockchain.ValidateBlockWithPreviousShardBestStateV2(shardBlock); err != nil {
			Logger.log.Error(err)
			return err
		}
		if err := blockchain.RevertShardStateV2(shardBlock.Header.ShardID); err != nil {
			Logger.log.Error(err)
			return err
		}
	}
	if blockHeight != GetBestStateShard(shardID).ShardHeight+1 {
		return errors.New("Not expected height")
	}
	isExist, _ := rawdbv2.HasShardBlock(blockchain.GetDatabase(), blockHash)
	if isExist {
		return NewBlockChainError(DuplicateShardBlockError, fmt.Errorf("SHARD %+v, block height %+v wit hash %+v has been stored already", shardID, blockHeight, blockHash))
	}
	// fetch beacon blocks
	beaconBlocks, err := FetchBeaconBlockFromHeightV2(blockchain.GetDatabase(), tempShardBestState.BeaconHeight+1, shardBlock.Header.BeaconHeight)
	if err != nil {
		return NewBlockChainError(FetchBeaconBlocksError, err)
	}
	if !isValidated {
		Logger.log.Infof("SHARD %+v | Verify Pre Processing, block height %+v with hash %+vt \n", shardID, blockHeight, blockHash)
		if err := blockchain.verifyPreProcessingShardBlockV2(shardBlock, beaconBlocks, shardID, false); err != nil {
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
	err = blockchain.BackupCurrentShardStateV2(shardBlock)
	if err != nil {
		return NewBlockChainError(BackUpBestStateError, err)
	}
	oldCommittee, err := incognitokey.CommitteeKeyListToString(tempShardBestState.ShardCommittee)
	if err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Update ShardBestState, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	if err := tempShardBestState.updateShardBestStateV2(blockchain, shardBlock, beaconBlocks, committeeChange); err != nil {
		return err
	}

	Logger.log.Infof("SHARD %+v | Update NumOfBlocksByProducers, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	// update number of blocks produced by producers to shard best state
	tempShardBestState.updateNumOfBlocksByProducers(shardBlock)
	newCommittee, err := incognitokey.CommitteeKeyListToString(tempShardBestState.ShardCommittee)
	if err != nil {
		return err
	}
	if !common.CompareStringArray(oldCommittee, newCommittee) {
		go blockchain.config.ConsensusEngine.CommitteeChange(common.GetShardChainKey(shardID))
	}
	//========Post verification: verify new beaconstate with corresponding block
	if !isValidated {
		Logger.log.Infof("SHARD %+v | Verify Post Processing, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
		if err := tempShardBestState.verifyPostProcessingShardBlock(shardBlock, shardID); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Post Processing, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	}
	Logger.log.Infof("SHARD %+v | Remove Data After Processed, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	Logger.log.Infof("SHARD %+v | Update Beacon Instruction, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	err = blockchain.processSalaryInstructionsV2(tempShardBestState.rewardStateDB, beaconBlocks, shardID)
	if err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Store New Shard Block And Update Data, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	//========Store new  Shard block and new shard bestState
	err = blockchain.processStoreShardBlockV2(shardBlock, committeeChange)
	if err != nil {
		revertErr := blockchain.revertShardStateV2(shardID)
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

func (blockchain *BlockChain) verifyPreProcessingShardBlockV2(shardBlock *ShardBlock, beaconBlocks []*BeaconBlock, shardID byte, isPreSign bool) error {
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
		return NewBlockChainError(TransactionRootHashError, fmt.Errorf("Expect transaction root hash %+v but get %+v", shardBlock.Header.TxRoot, *txRoot))
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
	Logger.log.Infof("Shard Proccess Instruction from Transaction %+v", txInstructions)
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
	err = blockchain.ValidateResponseTransactionFromTxsWithMetadataV2(shardBlock)
	if err != nil {
		return NewBlockChainError(ResponsedTransactionWithMetadataError, err)
	}
	// Get cross shard shardBlock from pool
	if isPreSign {
		err := blockchain.verifyPreProcessingShardBlockForSigningV2(shardBlock, beaconBlocks, txInstructions, shardID)
		if err != nil {
			return err
		}
	}
	Logger.log.Debugf("SHARD %+v | Finish verifyPreProcessingShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	return err
}

func (blockchain *BlockChain) verifyPreProcessingShardBlockForSigningV2(shardBlock *ShardBlock, beaconBlocks []*BeaconBlock, txInstructions [][]string, shardID byte) error {
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
	shardPendingValidator, _, _ := blockchain.processInstructionFromBeaconV2(beaconBlocks, shardID, newCommitteeChange())
	if blockchain.BestState.Shard[shardID].BeaconHeight == shardBlock.Header.BeaconHeight {
		isOldBeaconHeight = true
	}
	instructions, shardPendingValidator, shardCommittee, err = blockchain.generateInstructionV2(shardID, shardBlock.Header.BeaconHeight, isOldBeaconHeight, beaconBlocks, shardPendingValidator, shardCommittee)
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
					beaconHeight, err := blockchain.FindBeaconHeightForCrossShardBlockV2(toShardCrossShardBlock.Header.BeaconHeight, toShardCrossShardBlock.Header.ShardID, toShardCrossShardBlock.Header.Height)
					if err != nil {
						Logger.log.Errorf("%+v", err)
						break
					}
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
func (shardBestState *ShardBestState) initShardBestStateV2(blockchain *BlockChain, db incdb.Database, genesisShardBlock *ShardBlock, genesisBeaconBlock *BeaconBlock) error {
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
	shardPendingValidator, _, stakingTx := blockchain.processInstructionFromBeacon([]*BeaconBlock{genesisBeaconBlock}, genesisShardBlock.Header.ShardID)

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

func (shardBestState *ShardBestState) updateShardBestStateV2(blockchain *BlockChain, shardBlock *ShardBlock, beaconBlocks []*BeaconBlock, committeeChange *committeeChange) error {
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
	} else {
		shardBestState.ShardProposerIdx = (shardBestState.ShardProposerIdx + shardBlock.Header.Round) % len(shardBestState.ShardCommittee)
	}
	//shardBestState.processBeaconBlocks(shardBlock, beaconBlocks)
	shardPendingValidator, newShardPendingValidator, stakingTx := blockchain.processInstructionFromBeaconV2(beaconBlocks, shardBlock.Header.ShardID, committeeChange)
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
	err = shardBestState.processShardBlockInstructionV2(blockchain, shardBlock, committeeChange)
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

func (shardBestState *ShardBestState) processShardBlockInstructionV2(blockchain *BlockChain, shardBlock *ShardBlock, committeeChange *committeeChange) error {
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
	shardSwappedCommittees := []string{}
	shardNewCommittees := []string{}
	if len(shardBlock.Body.Instructions) != 0 {
		Logger.log.Info("Shard Process/updateShardBestState: Shard Instruction", shardBlock.Body.Instructions)
	}
	producersBlackList, err := blockchain.getUpdatedProducersBlackListV2(blockchain.BestState.Beacon.slashStateDB, false, int(shardID), shardCommittee, shardBlock.Header.BeaconHeight)
	if err != nil {
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
	shardBestState.ShardCommittee, err = incognitokey.CommitteeBase58KeyListToStruct(shardCommittee)
	if err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) processStoreShardBlockV2(shardBlock *ShardBlock, committeeChange *committeeChange) error {
	shardID := shardBlock.Header.ShardID
	blockHeight := shardBlock.Header.Height
	blockHash := shardBlock.Header.Hash()
	tempShardBestState := blockchain.BestState.Shard[shardID]
	tempBeaconBestState := blockchain.BestState.Beacon
	Logger.log.Infof("SHARD %+v | Process store block height %+v at hash %+v", shardBlock.Header.ShardID, blockHeight, *shardBlock.Hash())
	if len(shardBlock.Body.CrossTransactions) != 0 {
		Logger.log.Critical("processStoreShardBlockV2/CrossTransactions	", shardBlock.Body.CrossTransactions)
	}
	if err := blockchain.CreateAndSaveTxViewPointFromBlockV2(shardBlock, tempShardBestState.transactionStateDB, tempBeaconBestState.featureStateDB); err != nil {
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
			err := statedb.RemoveCommitteeReward(tempShardBestState.rewardStateDB, publicKey, amountRes, *coinID)
			if err != nil {
				return NewBlockChainError(RemoveCommitteeRewardError, err)
			}
		}
		Logger.log.Debugf("Transaction in block with hash", blockHash, "and index", index)
	}
	// Store Incomming Cross Shard
	if err := blockchain.CreateAndSaveCrossTransactionViewPointFromBlockV2(shardBlock, tempShardBestState.transactionStateDB); err != nil {
		return NewBlockChainError(FetchAndStoreCrossTransactionError, err)
	}
	// Save result of BurningConfirm instruction to get proof later
	err := blockchain.storeBurningConfirmV2(tempShardBestState.featureStateDB, shardBlock)
	if err != nil {
		return NewBlockChainError(StoreBurningConfirmError, err)
	}
	// Update bridge issuance request status
	err = blockchain.updateBridgeIssuanceStatusV2(tempShardBestState.featureStateDB, shardBlock)
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
	//statedb===========================START
	err = statedb.StoreOneShardCommittee(tempShardBestState.consensusStateDB, shardID, committeeChange.shardCommitteeAdded[shardID], tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = statedb.StoreOneShardSubstitutesValidator(tempShardBestState.consensusStateDB, shardID, committeeChange.shardSubstituteAdded[shardID], tempBeaconBestState.RewardReceiver, tempBeaconBestState.AutoStaking)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
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
	if err := rawdbv2.StoreShardConsensusRootHash(blockchain.GetDatabase(), shardID, blockHeight, consensusRootHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardTransactionRootHash(blockchain.GetDatabase(), shardID, blockHeight, transactionRootHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardFeatureRootHash(blockchain.GetDatabase(), shardID, blockHeight, featureRootHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardCommitteeRewardRootHash(blockchain.GetDatabase(), shardID, blockHeight, rewardRootHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardSlashRootHash(blockchain.GetDatabase(), shardID, blockHeight, slashRootHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	//statedb===========================END
	if err := rawdbv2.StoreShardBlock(blockchain.GetDatabase(), shardID, blockHeight, blockHash, shardBlock); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardBlockIndex(blockchain.GetDatabase(), shardID, blockHeight, blockHash); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	if err := rawdbv2.StoreShardBestState(blockchain.GetDatabase(), shardID, tempShardBestState); err != nil {
		return NewBlockChainError(StoreBestStateError, err)
	}
	Logger.log.Infof("SHARD %+v | 🔎 %d transactions in block height %+v \n", shardBlock.Header.ShardID, len(shardBlock.Body.Transactions), blockHeight)
	return nil
}
