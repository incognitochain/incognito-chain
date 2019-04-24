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

	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/transaction"
)

/*
	Verify Shard Block Before Signing
	Used for PBFT consensus
	@Notice: this block doesn't have full information (incomplete block)
*/
func (blockchain *BlockChain) VerifyPreSignShardBlock(block *ShardBlock, shardID byte) error {
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
	if strings.Compare(blockchain.BestState.Shard[shardID].BestBlockHash.String(), block.Header.PrevBlockHash.String()) == 0 {
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
func (blockchain *BlockChain) InsertShardBlock(block *ShardBlock, isProducer bool) error {
	shardID := block.Header.ShardID
	blockchain.BestState.Shard[shardID].lock.Lock()
	defer blockchain.BestState.Shard[shardID].lock.Unlock()
	Logger.log.Infof("SHARD %+v | Check block existence for insert height %+v at hash %+v", block.Header.ShardID, block.Header.Height, *block.Hash())
	isExist, _ := blockchain.config.DataBase.HasBlock(block.Hash())
	if isExist {
		//return nil
		return NewBlockChainError(DuplicateBlockErr, errors.New("This block has been stored already"))
	}
	Logger.log.Infof("SHARD %+v | Begin Insert new block height %+v at hash %+v", block.Header.ShardID, block.Header.Height, *block.Hash())
	if !isProducer {
		Logger.log.Infof("SHARD %+v | Verify Pre Processing  Block %+v \n", block.Header.ShardID, *block.Hash())
		if err := blockchain.VerifyPreProcessingShardBlock(block, shardID, false); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Pre Processing Block %+v \n", block.Header.ShardID, *block.Hash())
	}
	//========Verify block with previous best state
	// check with current final best state
	// block can only be insert if it match the current best state
	if strings.Compare(blockchain.BestState.Shard[shardID].BestBlockHash.String(), block.Header.PrevBlockHash.String()) != 0 {
		return NewBlockChainError(BeaconError, errors.New("beacon Block does not match with any Beacon State in cache or in Database"))
	}

	// Verify block with previous best state
	if !isProducer {
		Logger.log.Infof("SHARD %+v | Verify BestState with Block %+v \n", block.Header.ShardID, *block.Hash())
		if err := blockchain.BestState.Shard[shardID].VerifyBestStateWithShardBlock(block, true, shardID); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify BestState with Block %+v \n", block.Header.ShardID, *block.Hash())
	}
	Logger.log.Infof("SHARD %+v | Update BestState with Block %+v \n", block.Header.ShardID, *block.Hash())
	//========Update best state with new block
	prevBeaconHeight := blockchain.BestState.Shard[shardID].BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain.config.DataBase, prevBeaconHeight+1, block.Header.BeaconHeight)
	if err != nil {
		return err
	}
	if err := blockchain.BestState.Shard[shardID].Update(block, beaconBlocks); err != nil {
		return err
	}

	//========Post verififcation: verify new beaconstate with corresponding block
	if !isProducer {
		Logger.log.Infof("SHARD %+v | Verify Post Processing Block %+v \n", block.Header.ShardID, *block.Hash())
		if err := blockchain.BestState.Shard[shardID].VerifyPostProcessingShardBlock(block, shardID); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Post Processing Block %+v \n", block.Header.ShardID, *block.Hash())
	}

	//=========Remove invalid shard block in pool
	blockchain.config.ShardPool[shardID].SetShardState(blockchain.BestState.Shard[shardID].ShardHeight)

	//Update Cross shard pool: remove invalid block
	go func() {
		blockchain.config.CrossShardPool[shardID].RemoveBlockByHeight(blockchain.BestState.Shard[shardID].BestCrossShard)
		expectedHeight, _ := blockchain.config.CrossShardPool[shardID].UpdatePool()
		for fromShardID, height := range expectedHeight {
			if height != 0 {
				blockchain.SyncBlkCrossShard(false, false, []common.Hash{}, []uint64{height}, fromShardID, shardID, "")
			}
		}
	}()

	// Process stability tx
	err = blockchain.ProcessLoanForBlock(block)
	if err != nil {
		return err
	}

	err = blockchain.processTradeBondTx(block)
	if err != nil {
		return err
	}

	for _, tx := range block.Body.Transactions {
		meta := tx.GetMetadata()
		if meta == nil {
			continue
		}
		err := meta.ProcessWhenInsertBlockShard(tx, blockchain)
		if err != nil {
			return err
		}
	}

	// Process stability stand-alone instructions
	err = blockchain.ProcessStandAloneInstructions(block)
	if err != nil {
		return err
	}

	// Store metadata instruction to local state
	for _, beaconBlock := range beaconBlocks {
		instructions := beaconBlock.Body.Instructions
		for _, inst := range instructions {
			err := blockchain.StoreMetadataInstructions(inst, shardID)
			if err != nil {
				return err
			}
		}
	}

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
	}
	blockchain.config.TxPool.RemoveCandidateList(candidates)
	blockchain.config.TxPool.RemoveTokenIDList(tokenIDs)
	//Remove tx out of pool
	go func() {
		for _, tx := range block.Body.Transactions {
			blockchain.config.TxPool.RemoveTx(tx)
		}
	}()

	//========Store new  Shard block and new shard bestState
	err = blockchain.ProcessStoreShardBlock(block)
	if err != nil {
		return err
	}
	Logger.log.Infof("SHARD %+v | Finish Insert new block %d, with hash %+v", block.Header.ShardID, block.Header.Height, *block.Hash())
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
	if len(block.Body.Transactions) < 1 {
		Logger.log.Infof("No transaction in this block")
	} else {
		Logger.log.Infof("Number of transaction in this block %d", len(block.Body.Transactions))
	}

	if len(block.Body.CrossTransactions) != 0 {
		Logger.log.Critical("ProcessStoreShardBlock/CrossTransactions	", block.Body.CrossTransactions)
	}

	if err := blockchain.CreateAndSaveTxViewPointFromBlock(block); err != nil {
		return err
	}

	for index, tx := range block.Body.Transactions {
		if err := blockchain.StoreTransactionIndex(tx.Hash(), block.Hash(), index); err != nil {
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
	blkHash := block.Header.Hash()
	producerPk := base58.Base58Check{}.Encode(block.Header.ProducerAddress.Pk, common.ZeroByte)
	err := cashec.ValidateDataB58(producerPk, block.ProducerSig, blkHash.GetBytes())
	if err != nil {
		return NewBlockChainError(ProducerError, errors.New("Producer's sig not match"))
	}
	//verify producer
	producerPosition := (blockchain.BestState.Shard[shardID].ShardProposerIdx + block.Header.Round) % len(blockchain.BestState.Shard[shardID].ShardCommittee)
	tempProducer := blockchain.BestState.Shard[shardID].ShardCommittee[producerPosition]
	if strings.Compare(tempProducer, producerPk) != 0 {
		return NewBlockChainError(ProducerError, errors.New("Producer should be should be :"+tempProducer))
	}
	Logger.log.Debugf("SHARD %+v | Begin VerifyPreProcessingShardBlock Block with height %+v at hash %+v", block.Header.ShardID, block.Header.Height, block.Hash())
	if block.Header.ShardID != shardID {
		return NewBlockChainError(ShardIDError, errors.New("Shard should be :"+strconv.Itoa(int(shardID))))
	}
	if block.Header.Version != VERSION {
		return NewBlockChainError(VersionError, errors.New("Version should be :"+strconv.Itoa(VERSION)))
	}
	// Verify parent hash exist or not
	prevBlockHash := block.Header.PrevBlockHash
	parentBlockData, err := blockchain.config.DataBase.FetchBlock(&prevBlockHash)
	if err != nil {
		return NewBlockChainError(DBError, err)
	}
	parentBlock := ShardBlock{}
	json.Unmarshal(parentBlockData, &parentBlock)
	// Verify block height with parent block
	if parentBlock.Header.Height+1 != block.Header.Height {
		return NewBlockChainError(BlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
	}
	// Verify epoch with parent block
	// if block.Header.Height%EPOCH == 0 && parentBlock.Header.Epoch != block.Header.Epoch-1 {
	// 	return NewBlockChainError(EpochError, errors.New("Block height and Epoch is not compatiable"))
	// }
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
		return nil
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
	if !newHash.IsEqual(beaconHash) {
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

	// TODO(@0xbunyip): move to inside isPresig when running validator's node
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
		return NewBlockChainError(TransactionError, errors.New(fmt.Sprintf("There are %d invalid txs...", len(invalidTxs))))
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
				blockchain.SyncBlkCrossShard(false, false, []common.Hash{}, heights, fromShard, shardID, "")
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
	if err := cashec.ValidateDataB58(producerPubkey, block.ProducerSig, blockHash.GetBytes()); err != nil {
		return NewBlockChainError(SignatureError, err)
	}
	//=============End Verify producer signature
	//=============Verify aggegrate signature
	if isVerifySig {
		if len(bestStateShard.ShardCommittee) > 3 && len(block.ValidatorsIdx[1]) < (len(bestStateShard.ShardCommittee)>>1) {
			return NewBlockChainError(SignatureError, errors.New("block validators and Shard committee is not compatible"))
		}
		ValidateAggSignature(block.ValidatorsIdx, bestStateShard.ShardCommittee, block.AggregatedSig, block.R, block.Hash())
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

	// Add pending validator
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
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
	isEmpty := blockChain.config.TempTxPool.EmptyPool()
	if !isEmpty {
		panic("TempTxPool Is not Empty")
	}
	index := 0
	salaryCount := 0

	for _, tx := range txs {
		if !tx.IsSalaryTx() {
			if tx.GetType() == common.TxCustomTokenType {
				customTokenTx := tx.(*transaction.TxCustomToken)
				if customTokenTx.TxTokenData.Type == transaction.CustomTokenCrossShard {
					salaryCount++
					continue
				}
			}
			_, err := blockChain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx)
			if err != nil {
				return err
			}
			index++
		} else {
			salaryCount++
		}
	}
	blockChain.config.TempTxPool.EmptyPool()
	if index == len(txs)-salaryCount {
		return nil
	} else {
		return NewBlockChainError(TransactionError, errors.New("Some Transactions in new Block maybe invalid"))
	}
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
	if strings.Compare(hash.String(), hashFromTxs.String()) != 0 {
		return errors.New("Cross Token Data from Cross Shard Block Not Compatible with Cross Token Data in New Block")
	}
	return nil
}

//=====================Util for shard====================
