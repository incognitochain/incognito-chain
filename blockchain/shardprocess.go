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

	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/transaction"
)

/*
	Verify Shard Block Before Signing
	Used for PBFT consensus
	@Notice: this block doesn't have full information (incomplete block)
*/
func (blockchain *BlockChain) VerifyPreSignShardBlock(block *ShardBlock, shardID byte) error {
	//Logger.log.Errorf("\n%v\n%v\n", block.Header, len(block.Body.Transactions))
	if block.Header.ShardID != shardID {
		return errors.New("wrong shard")
	}
	blockchain.BestState.Shard[shardID].lock.Lock()
	defer blockchain.BestState.Shard[shardID].lock.Unlock()
	//========Verify block only
	Logger.log.Infof("SHARD %+v | Verify block for signing process %d, with hash %+v", shardID, block.Header.Height, *block.Hash())
	if err := blockchain.VerifyPreProcessingShardBlock(block, shardID, true); err != nil {
		return err
	}
	//========Verify block with previous best state
	// Get Beststate of previous block == previous best state
	// Clone best state value into new variable
	shardBestState := BestStateShard{}
	// check with current final best state
	// New block must be compatible with current best state
	bestBlockHash := &blockchain.BestState.Shard[shardID].BestBlockHash
	if bestBlockHash.IsEqual(&block.Header.PrevBlockHash) {
		tempMarshal, err := json.Marshal(blockchain.BestState.Shard[shardID])
		if err != nil {
			return NewBlockChainError(UnmashallJsonBlockError, err)
		}
		json.Unmarshal(tempMarshal, &shardBestState)
	}
	// if no match best state found then block is unknown
	if reflect.DeepEqual(shardBestState, BestStateShard{}) {
		return NewBlockChainError(ShardError, errors.New("shard Block does not match with any Shard State in cache or in Database"))
	}
	// Verify block with previous best state
	// not verify agg signature in this function
	prevBeaconHeight := shardBestState.BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.config.DataBase, prevBeaconHeight+1, block.Header.BeaconHeight)
	if err != nil {
		return err
	}
	if err := shardBestState.VerifyBestStateWithShardBlock(block, false, shardID); err != nil {
		return err
	}
	//========Update best state with new block
	if err := shardBestState.Update(block, beaconBlocks); err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding block
	if err := shardBestState.VerifyPostProcessingShardBlock(block, shardID); err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Block %d, with hash %+v is VALID for signing", shardID, block.Header.Height, *block.Hash())
	return nil
}

/*
	Insert Shard Block into blockchain
	@Notice: this block must have full information (complete block)
*/
func (blockchain *BlockChain) InsertShardBlock(block *ShardBlock, isValidated bool) error {
	// force non-committee member not to validate blk
	if blockchain.config.UserKeySet != nil && (blockchain.config.NodeMode == common.NODEMODE_AUTO || blockchain.config.NodeMode == common.NODEMODE_SHARD) {
		userRole := blockchain.BestState.Shard[block.Header.ShardID].GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), 0)
		fmt.Println("Shard block received 1", userRole)

		if userRole != common.PROPOSER_ROLE && userRole != common.VALIDATOR_ROLE && userRole != common.PENDING_ROLE {
			isValidated = true
		}
	} else {
		isValidated = true
	}
	shardID := block.Header.ShardID
	blockchain.BestState.Shard[shardID].lock.Lock()
	defer blockchain.BestState.Shard[shardID].lock.Unlock()
	blockHash := block.Header.Hash()
	Logger.log.Infof("SHARD %+v | Check block existence for insert height %+v at hash %+v", block.Header.ShardID, block.Header.Height, blockHash)
	isExist, _ := blockchain.config.DataBase.HasBlock(blockHash)
	if isExist {
		//return nil
		return NewBlockChainError(DuplicateBlockErr, errors.New("This block has been stored already"))
	}
	Logger.log.Infof("SHARD %+v | Begin Insert new block height %+v at hash %+v", block.Header.ShardID, block.Header.Height, blockHash)
	// Verify block with previous best state
	Logger.log.Infof("SHARD %+v | Verify BestState with Block %+v \n", block.Header.ShardID, blockHash)
	if err := blockchain.BestState.Shard[shardID].VerifyBestStateWithShardBlock(block, true, shardID); err != nil {
		return err
	}
	if !isValidated {
		Logger.log.Infof("SHARD %+v | Verify Pre Processing  Block %+v \n", block.Header.ShardID, blockHash)
		if err := blockchain.VerifyPreProcessingShardBlock(block, shardID, false); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Pre Processing Block %+v \n", block.Header.ShardID, blockHash)
	}
	//========Verify block with previous best state
	// check with current final best state
	// block can only be insert if it match the current best state
	bestBlockHash := &blockchain.BestState.Shard[shardID].BestBlockHash
	if !bestBlockHash.IsEqual(&block.Header.PrevBlockHash) {
		return NewBlockChainError(BeaconError, errors.New("beacon Block does not match with any Beacon State in cache or in Database"))
	}

	Logger.log.Infof("SHARD %+v | Update BestState with Block %+v \n", block.Header.ShardID, blockHash)
	//========Update best state with new block
	prevBeaconHeight := blockchain.BestState.Shard[shardID].BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.config.DataBase, prevBeaconHeight+1, block.Header.BeaconHeight)
	if err != nil {
		return err
	}
	// Backup beststate
	if blockchain.config.UserKeySet != nil {
		userRole := blockchain.BestState.Shard[shardID].GetPubkeyRole(blockchain.config.UserKeySet.GetPublicKeyB58(), 0)
		if userRole == common.PROPOSER_ROLE || userRole == common.VALIDATOR_ROLE {
			blockchain.config.DataBase.CleanBackup(true, block.Header.ShardID)
			err = blockchain.BackupCurrentShardState(block, beaconBlocks)
			if err != nil {
				return err
			}
		}
	}

	if err := blockchain.BestState.Shard[shardID].Update(block, beaconBlocks); err != nil {
		return err
	}

	//========Post verififcation: verify new beaconstate with corresponding block
	if !isValidated {
		Logger.log.Infof("SHARD %+v | Verify Post Processing Block %+v \n", block.Header.ShardID, blockHash)
		if err := blockchain.BestState.Shard[shardID].VerifyPostProcessingShardBlock(block, shardID); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Post Processing Block %+v \n", block.Header.ShardID, blockHash)
	}

	//remove staking txid in beststate shard
	go func() {
		for _, l := range block.Body.Instructions {
			if l[0] == SwapAction {
				swapedCommittees := strings.Split(l[2], ",")
				for _, v := range swapedCommittees {
					delete(GetBestStateShard(shardID).StakingTx, v)
				}
			}
		}
	}()

	//=========Remove invalid shard block in pool
	blockchain.config.ShardPool[shardID].SetShardState(blockchain.BestState.Shard[shardID].ShardHeight)

	//Update Cross shard pool: remove invalid block
	go func() {
		blockchain.config.CrossShardPool[shardID].RemoveBlockByHeight(blockchain.BestState.Shard[shardID].BestCrossShard)
		//expectedHeight, _ := blockchain.config.CrossShardPool[shardID].UpdatePool()
		//for fromShardID, height := range expectedHeight {
		//	if height != 0 {
		//		blockchain.SyncBlkCrossShard(false, false, []common.Hash{}, []uint64{height}, fromShardID, shardID, "")
		//	}
		//}
	}()

	go func() {
		//Remove Candidate In pool
		candidates := []string{}
		tokenIDs := []string{}
		for _, tx := range block.Body.Transactions {
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
		blockchain.config.TxPool.RemoveCandidateList(candidates)
		blockchain.config.TxPool.RemoveTokenIDList(tokenIDs)

		//Remove tx out of pool
		go blockchain.config.TxPool.RemoveTx(block.Body.Transactions, true)
		for _, tx := range block.Body.Transactions {
			go func(tx metadata.Transaction) {
				if blockchain.config.IsBlockGenStarted {
					blockchain.config.CRemovedTxs <- tx
				}
			}(tx)
		}
	}()

	//========Store new  Shard block and new shard bestState
	err = blockchain.ProcessStoreShardBlock(block)
	if err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Finish Insert new block %d, with hash %+v", block.Header.ShardID, block.Header.Height, blockHash)
	shardIDForMetric := strconv.Itoa(int(block.Header.ShardID))
	go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
		metrics.Measurement:      metrics.NumOfBlockInsertToChain,
		metrics.MeasurementValue: float64(1),
		metrics.Tag:              metrics.ShardIDTag,
		metrics.TagValue:         metrics.Shard + shardIDForMetric,
	})
	// call FeeEstimator for processing
	if feeEstimator, ok := blockchain.config.FeeEstimator[block.Header.ShardID]; ok {
		go feeEstimator.RegisterBlock(block)
	}
	err = blockchain.updateDatabaseFromBeaconInstructions(beaconBlocks, shardID)
	if err != nil {
		fmt.Printf("[ndh]  - - - [error]1: %+v\n", err)
		return err
	}
	err = blockchain.updateDatabaseFromShardBlock(block)
	if err != nil {
		fmt.Printf("[ndh]  - - - [error]2: %+v\n", err)
		return err
	}
	fmt.Printf("[ndh]  - - - nonerror \n")

	// Save result of BurningConfirm instruction to get proof later
	err = blockchain.storeBurningConfirm(block)
	if err != nil {
		return err
	}

	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewShardblockTopic, block))
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardBeststateTopic, blockchain.BestState.Shard[shardID]))
	return nil
}

func (blockchain *BlockChain) insertETHHeaders(
	shardBlock *ShardBlock,
	tx metadata.Transaction,
) error {
	relayingRewardMeta := tx.GetMetadata().(*metadata.ETHHeaderRelayingReward)
	reqID := relayingRewardMeta.RequestedTxID
	txs := shardBlock.Body.Transactions
	var relayingTx metadata.Transaction
	for _, item := range txs {
		metaType := item.GetMetadataType()
		if metaType == metadata.ETHHeaderRelayingMeta &&
			bytes.Equal(reqID[:], item.Hash()[:]) {
			relayingTx = item
			break
		}
	}
	relayingMeta := relayingTx.GetMetadata().(*metadata.ETHHeaderRelaying)
	lc := blockchain.LightEthereum.GetLightChain()
	_, err := lc.InsertHeaderChain(relayingMeta.ETHHeaders, 0)
	if err != nil {
		fmt.Println("haha insert error: ", err)
		return err
	}
	fmt.Println("haha insert header chain ok")
	return nil
}

func (blockchain *BlockChain) insertStuffForIssuingETHRes(
	tx metadata.Transaction,
) error {
	db := blockchain.GetDatabase()
	issuingETHResdMeta := tx.GetMetadata().(*metadata.IssuingETHResponse)
	err := db.InsertETHTxHashIssued(issuingETHResdMeta.UniqETHTx)
	if err != nil {
		return err
	}
	err = db.UpdateBridgeTokenPairInfo(
		*tx.GetTokenID(),
		issuingETHResdMeta.ExternalTokenID,
		false,
	)
	fmt.Println("haha finally")
	return err
}

func (blockchain *BlockChain) updateDatabaseFromShardBlock(
	shardBlock *ShardBlock,
) error {
	db := blockchain.config.DataBase
	for _, tx := range shardBlock.Body.Transactions {
		metaType := tx.GetMetadataType()
		var err error
		if metaType == metadata.WithDrawRewardResponseMeta {
			_, requesterRes, amountRes, coinID := tx.GetTransferData()
			err = db.RemoveCommitteeReward(requesterRes, amountRes, *coinID)
		} else if metaType == metadata.ETHHeaderRelayingRewardMeta {
			err = blockchain.insertETHHeaders(shardBlock, tx)
		} else if metaType == metadata.IssuingETHResponseMeta {
			err = blockchain.insertStuffForIssuingETHRes(tx)
		}
		if err != nil {
			return err
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
*/
func (blockchain *BlockChain) ProcessStoreShardBlock(block *ShardBlock) error {
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

	// Process transaction db
	Logger.log.Criticalf("SHARD %+v | ⚒︎ %d transactions in block height %+v \n", block.Header.ShardID, len(block.Body.Transactions), block.Header.Height)
	if block.Header.Height != 1 {
		go metrics.AnalyzeTimeSeriesMetricData(map[string]interface{}{
			metrics.Measurement:      metrics.TxInOneBlock,
			metrics.MeasurementValue: float64(len(block.Body.Transactions)),
			metrics.Tag:              metrics.BlockHeightTag,
			metrics.TagValue:         fmt.Sprintf("%d", block.Header.Height),
		})
	}
	if len(block.Body.CrossTransactions) != 0 {
		Logger.log.Critical("ProcessStoreShardBlock/CrossTransactions	", block.Body.CrossTransactions)
	}

	if err := blockchain.CreateAndSaveTxViewPointFromBlock(block); err != nil {
		return err
	}

	for index, tx := range block.Body.Transactions {
		if err := blockchain.StoreTransactionIndex(tx.Hash(), block.Header.Hash(), index); err != nil {
			Logger.log.Error("ERROR", err, "Transaction in block with hash", blockHash, "and index", index, ":", tx)
			return NewBlockChainError(UnExpectedError, err)
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
- ALL Transaction in block: see in VerifyTransactionFromNewBlock
*/
func (blockchain *BlockChain) VerifyPreProcessingShardBlock(block *ShardBlock, shardID byte, isPresig bool) error {
	//verify producer sig
	Logger.log.Debugf("SHARD %+v | Begin VerifyPreProcessingShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	if block.Header.ShardID != shardID {
		return NewBlockChainError(ShardIDError, errors.New("Shard should be :"+strconv.Itoa(int(shardID))))
	}
	if block.Header.Version != VERSION {
		return NewBlockChainError(VersionError, errors.New("Version should be :"+strconv.Itoa(VERSION)))
	}
	// Verify parent hash exist or not
	prevBlockHash := block.Header.PrevBlockHash
	parentBlockData, err := blockchain.config.DataBase.FetchBlock(prevBlockHash)
	if err != nil {
		return NewBlockChainError(DBError, err)
	}
	parentBlock := ShardBlock{}
	err = json.Unmarshal(parentBlockData, &parentBlock)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}
	// Verify block height with parent block
	if parentBlock.Header.Height+1 != block.Header.Height {
		return NewBlockChainError(BlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
	}
	// Verify timestamp with parent block
	if block.Header.Timestamp <= parentBlock.Header.Timestamp {
		return NewBlockChainError(TimestampError, errors.New("timestamp of new block can't equal to parent block"))
	}
	// Verify transaction root
	txMerkle := Merkle{}.BuildMerkleTreeStore(block.Body.Transactions)
	txRoot := &common.Hash{}
	if len(txMerkle) > 0 {
		txRoot = txMerkle[len(txMerkle)-1]
	}

	if !bytes.Equal(block.Header.TxRoot.GetBytes(), txRoot.GetBytes()) {
		return NewBlockChainError(HashError, errors.New("can't Verify Transaction Root"))
	}
	// Verify ShardTx Root
	_, shardTxMerkleData := CreateShardTxRoot2(block.Body.Transactions)
	shardTxRoot := shardTxMerkleData[len(shardTxMerkleData)-1]
	if !bytes.Equal(block.Header.ShardTxRoot.GetBytes(), shardTxRoot.GetBytes()) {
		return NewBlockChainError(HashError, errors.New("can't Verify CrossShardTransaction Root"))
	}
	// Verify crossTransaction coin
	if !VerifyMerkleCrossTransaction(block.Body.CrossTransactions, block.Header.CrossTransactionRoot) {
		return NewBlockChainError(HashError, errors.New("can't Verify CrossOutputCoin Root"))
	}
	// Verify Action
	beaconBlocks, err := FetchBeaconBlockFromHeight(
		blockchain.config.DataBase,
		blockchain.BestState.Shard[block.Header.ShardID].BeaconHeight+1,
		block.Header.BeaconHeight,
	)
	if err != nil {
		Logger.log.Error(err)
		return err
	}

	wrongTxsFee := errors.New("Wrong blockheader totalTxs fee")

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
		return tokenIDsfromTxs[i].Cmp(&tokenIDsfromTxs[j]) == -1
	})

	tokenIDsfromBlock := make([]common.Hash, 0)
	for tokenID, _ := range block.Header.TotalTxsFee {
		tokenIDsfromBlock = append(tokenIDsfromBlock, tokenID)
	}
	sort.Slice(tokenIDsfromBlock, func(i int, j int) bool {
		return tokenIDsfromBlock[i].Cmp(&tokenIDsfromBlock[j]) == -1
	})

	if len(tokenIDsfromTxs) != len(tokenIDsfromBlock) {
		return wrongTxsFee
	}

	for i, tokenID := range tokenIDsfromTxs {
		if !tokenIDsfromTxs[i].IsEqual(&tokenIDsfromBlock[i]) {
			return wrongTxsFee
		}
		if totalTxsFee[tokenID] != block.Header.TotalTxsFee[tokenID] {
			return wrongTxsFee
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
		return nil
	}
	totalInstructions := []string{}
	for _, value := range txInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	for _, value := range block.Body.Instructions {
		totalInstructions = append(totalInstructions, value...)
	}
	isOk := VerifyHashFromStringArray(totalInstructions, block.Header.InstructionsRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("Error verify action root"))
	}

	// Check if InstructionMerkleRoot is the root of merkle tree containing all instructions in this block
	flattenTxInsts := FlattenAndConvertStringInst(txInstructions)
	flattenInsts := FlattenAndConvertStringInst(block.Body.Instructions)
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be the same as when creating new shard block
	root := GetKeccak256MerkleRoot(insts)
	if !bytes.Equal(root, block.Header.InstructionMerkleRoot[:]) {
		return NewBlockChainError(HashError, errors.New("invalid InstructionMerkleRoot"))
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

	// Verify stability transactions
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
		fmt.Println("haha failed: ", invalidTxs[0].GetMetadata())
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
		if err := blockchain.VerifyTransactionFromNewBlock(block.Body.Transactions); err != nil {
			return NewBlockChainError(TransactionError, err)
		}

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
						temp, err := blockchain.config.DataBase.FetchCommitteeByEpoch(toShardCrossShardBlock.Header.BeaconHeight)
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
		if err := blockchain.VerifyCrossShardCustomToken(crossTxTokenData, shardID, block.Body.Transactions); err != nil {
			return NewBlockChainError(CrossShardBlockError, err)
		}
	}
	Logger.log.Debugf("SHARD %+v | Finish VerifyPreProcessingShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
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
func (bestStateShard *BestStateShard) VerifyBestStateWithShardBlock(block *ShardBlock, isVerifySig bool, shardID byte) error {
	Logger.log.Debugf("SHARD %+v | Begin VerifyBestStateWithShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	// Cal next producer
	// Verify next producer
	//=============Verify producer signature
	producerPosition := (bestStateShard.ShardProposerIdx + block.Header.Round) % len(bestStateShard.ShardCommittee)
	producerPubkey := bestStateShard.ShardCommittee[producerPosition]
	blockHash := block.Header.Hash()
	fmt.Println("V58", producerPubkey, block.ProducerSig, blockHash.GetBytes(), base58.Base58Check{}.Encode(block.Header.ProducerAddress.Pk, common.ZeroByte))
	//verify producer
	tempProducer := bestStateShard.ShardCommittee[producerPosition]
	if strings.Compare(tempProducer, producerPubkey) != 0 {
		return NewBlockChainError(ProducerError, errors.New("Producer should be should be :"+tempProducer))
	}
	if err := cashec.ValidateDataB58(producerPubkey, block.ProducerSig, blockHash.GetBytes()); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	//=============End Verify producer signature
	//=============Verify aggegrate signature
	if isVerifySig {
		if len(bestStateShard.ShardCommittee) > 3 && len(block.ValidatorsIdx[1]) < (len(bestStateShard.ShardCommittee)>>1) {
			return NewBlockChainError(SignatureError, errors.New("block validators and Shard committee is not compatible"))
		}
		err := ValidateAggSignature(block.ValidatorsIdx, bestStateShard.ShardCommittee, block.AggregatedSig, block.R, block.Hash())
		if err != nil {
			return err
		}
	}
	//=============End Verify Aggegrate signature
	if bestStateShard.ShardHeight+1 != block.Header.Height {
		return NewBlockChainError(BlockHeightError, errors.New("block height of new block should be : "+strconv.Itoa(int(bestStateShard.ShardHeight+1))))
	}
	if block.Header.BeaconHeight < bestStateShard.BeaconHeight {
		return NewBlockChainError(BlockHeightError, errors.New("block contain invalid beacon height"))
	}
	Logger.log.Debugf("SHARD %+v | Finish VerifyBestStateWithShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return nil
}

/*
	Update beststate with new block
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
func (bestStateShard *BestStateShard) Update(block *ShardBlock, beaconBlocks []*BeaconBlock) error {
	Logger.log.Debugf("SHARD %+v | Begin update Beststate with new Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	var (
		err                   error
		shardSwapedCommittees []string
		shardNewCommittees    []string
	)
	bestStateShard.BestBlockHash = *block.Hash()
	if block.Header.BeaconHeight == 1 {
		bestStateShard.BestBeaconHash = *ChainTestParam.GenesisBeaconBlock.Hash()
	} else {
		bestStateShard.BestBeaconHash = block.Header.BeaconHash
	}
	if block.Header.Height == 1 {
		bestStateShard.BestCrossShard = make(map[byte]uint64)
	}
	bestStateShard.BestBlock = block
	bestStateShard.BestBlockHash = *block.Hash()
	bestStateShard.ShardHeight = block.Header.Height
	bestStateShard.Epoch = block.Header.Epoch
	bestStateShard.BeaconHeight = block.Header.BeaconHeight
	bestStateShard.TotalTxns += uint64(len(block.Body.Transactions))
	bestStateShard.NumTxns = uint64(len(block.Body.Transactions))
	//======BEGIN For testing and benchmark
	temp := 0
	for _, tx := range block.Body.Transactions {
		//detect transaction that's not salary
		if !tx.IsSalaryTx() {
			temp++
		}
	}
	bestStateShard.TotalTxnsExcludeSalary += uint64(temp)
	//======END
	if block.Header.Height == 1 {
		bestStateShard.ShardProposerIdx = 0
	} else {
		bestStateShard.ShardProposerIdx = common.IndexOfStr(base58.Base58Check{}.Encode(block.Header.ProducerAddress.Pk, common.ZeroByte), bestStateShard.ShardCommittee)
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
						GetBestStateShard(bestStateShard.ShardID).StakingTx[newBeaconCandidate[i]] = v
					}
				}
			}
			if l[0] == StakeAction && l[2] == "shard" {
				shard := strings.Split(l[1], ",")
				newShardCandidate = append(newShardCandidate, shard...)
				if len(l) == 4 {
					for i, v := range strings.Split(l[3], ",") {
						GetBestStateShard(bestStateShard.ShardID).StakingTx[newShardCandidate[i]] = v
					}
				}
			}

			if l[0] == "assign" && l[2] == "shard" {
				if l[3] == strconv.Itoa(int(block.Header.ShardID)) {
					Logger.log.Infof("SHARD %+v | Old ShardPendingValidatorList %+v", block.Header.ShardID, bestStateShard.ShardPendingValidator)
					bestStateShard.ShardPendingValidator = append(bestStateShard.ShardPendingValidator, strings.Split(l[1], ",")...)
					Logger.log.Infof("SHARD %+v | New ShardPendingValidatorList %+v", block.Header.ShardID, bestStateShard.ShardPendingValidator)
				}
			}
		}
	}
	if len(block.Body.Instructions) != 0 {
		Logger.log.Critical("Shard Process/Update: ALL Instruction", block.Body.Instructions)
	}

	// Swap committee
	for _, l := range block.Body.Instructions {

		if l[0] == "swap" {
			bestStateShard.ShardPendingValidator, bestStateShard.ShardCommittee, shardSwapedCommittees, shardNewCommittees, err = SwapValidator(bestStateShard.ShardPendingValidator, bestStateShard.ShardCommittee, bestStateShard.ShardCommitteeSize, common.OFFSET)
			if err != nil {
				Logger.log.Errorf("SHARD %+v | Blockchain Error %+v", NewBlockChainError(UnExpectedError, err))
				return NewBlockChainError(UnExpectedError, err)
			}
			swapedCommittees := strings.Split(l[2], ",")
			newCommittees := strings.Split(l[1], ",")

			for _, v := range swapedCommittees {
				delete(GetBestStateShard(bestStateShard.ShardID).StakingTx, v)
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
	//Update best cross shard
	for shardID, crossShardBlock := range block.Body.CrossTransactions {
		bestStateShard.BestCrossShard[shardID] = crossShardBlock[len(crossShardBlock)-1].BlockHeight
	}

	Logger.log.Debugf("SHARD %+v | Finish update Beststate with new Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	return nil
}

/*
	VerifyPostProcessingShardBlock
	- commitee root
	- pending validator root
*/
func (blockchain *BestStateShard) VerifyPostProcessingShardBlock(block *ShardBlock, shardID byte) error {
	var (
		isOk bool
	)
	Logger.log.Debugf("SHARD %+v | Begin VerifyPostProcessing Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	isOk = VerifyHashFromStringArray(blockchain.ShardCommittee, block.Header.CommitteeRoot)
	if !isOk {
		return NewBlockChainError(HashError, errors.New("Error verify Committee root"))
	}
	isOk = VerifyHashFromStringArray(blockchain.ShardPendingValidator, block.Header.PendingValidatorRoot)
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
func (blockChain *BlockChain) VerifyTransactionFromNewBlock(txs []metadata.Transaction) error {
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
	return nil

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
func (blockchain *BlockChain) VerifyCrossShardCustomToken(CrossTxTokenData map[byte][]CrossTxTokenData, shardID byte, txs []metadata.Transaction) error {
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
