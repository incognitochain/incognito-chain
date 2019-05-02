package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

func (blockgen *BlkTmplGenerator) NewBlockShard(producerKeySet *cashec.KeySet, shardID byte, proposerOffset int, crossShards map[byte]uint64) (*ShardBlock, error) {
	//============Build body=============
	// Fetch Beacon information
	fmt.Printf("[ndh] ========================== Creating shard block[%+v] ==============================", blockgen.chain.BestState.Shard[shardID].ShardHeight+1)
	beaconHeight := blockgen.chain.BestState.Beacon.BeaconHeight
	beaconHash := blockgen.chain.BestState.Beacon.BestBlockHash
	epoch := blockgen.chain.BestState.Beacon.Epoch
	if epoch-blockgen.chain.BestState.Shard[shardID].Epoch > 1 {
		beaconHeight = blockgen.chain.BestState.Shard[shardID].Epoch * common.EPOCH
		newBeaconHash, err := blockgen.chain.config.DataBase.GetBeaconBlockHashByIndex(beaconHeight)
		if err != nil {
			return nil, err
		}
		copy(beaconHash[:], newBeaconHash.GetBytes())
		epoch = blockgen.chain.BestState.Shard[shardID].Epoch + 1
	}
	//Fetch beacon block from height
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockgen.chain.config.DataBase, blockgen.chain.BestState.Shard[shardID].BeaconHeight+1, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	//======Get Transaction For new Block================
	txsToAdd, err1 := blockgen.getTransactionForNewBlock(&producerKeySet.PrivateKey, shardID, blockgen.chain.config.DataBase, beaconBlocks)
	if err1 != nil {
		Logger.log.Error(err1, reflect.TypeOf(err1), reflect.ValueOf(err1))
		return nil, err1
	}
	//======Get Cross output coin from other shard=======
	crossTransactions, crossTxTokenData := blockgen.getCrossShardData(shardID, blockgen.chain.BestState.Shard[shardID].BeaconHeight, beaconHeight, crossShards)
	crossTxTokenTransactions, _ := blockgen.chain.createCustomTokenTxForCrossShard(&producerKeySet.PrivateKey, crossTxTokenData, shardID)
	txsToAdd = append(txsToAdd, crossTxTokenTransactions...)
	//======Create Instruction===========================
	//Assign Instruction
	instructions := [][]string{}
	swapInstruction := []string{}
	assignInstructions := GetAssignInstructionFromBeaconBlock(beaconBlocks, shardID)
	if len(assignInstructions) != 0 {
		Logger.log.Critical("Shard Block Producer AssignInstructions ", assignInstructions)
	}
	shardPendingValidator := blockgen.chain.BestState.Shard[shardID].ShardPendingValidator
	shardCommittee := blockgen.chain.BestState.Shard[shardID].ShardCommittee
	for _, assignInstruction := range assignInstructions {
		shardPendingValidator = append(shardPendingValidator, strings.Split(assignInstruction[1], ",")...)
	}
	//Swap instruction
	// Swap instruction only appear when reach the last block in an epoch
	//@NOTICE: In this block, only pending validator change, shard committees will change in the next block
	if beaconHeight%common.EPOCH == 0 {
		if len(shardPendingValidator) > 0 {
			Logger.log.Critical("shardPendingValidator", shardPendingValidator)
			Logger.log.Critical("shardCommittee", shardCommittee)
			Logger.log.Critical("blockgen.chain.BestState.Shard[shardID].ShardCommitteeSize", blockgen.chain.BestState.Shard[shardID].ShardCommitteeSize)
			Logger.log.Critical("shardID", shardID)
			swapInstruction, shardPendingValidator, shardCommittee, err = CreateSwapAction(shardPendingValidator, shardCommittee, blockgen.chain.BestState.Shard[shardID].ShardCommitteeSize, shardID)
			if err != nil {
				Logger.log.Error(err)
				return nil, err
			}
		}
	}
	if !reflect.DeepEqual(swapInstruction, []string{}) {
		instructions = append(instructions, swapInstruction)
	}

	// Build stand-alone stability instructions
	tradeBondRespInsts, err := blockgen.buildTradeBondConfirmInsts(beaconBlocks, shardID)
	if err != nil {
		return nil, err
	}
	if tradeBondRespInsts != nil && len(tradeBondRespInsts) > 0 {
		instructions = append(instructions, tradeBondRespInsts...)
	}

	block := &ShardBlock{
		Body: ShardBody{
			CrossTransactions: crossTransactions,
			Instructions:      instructions,
			Transactions:      make([]metadata.Transaction, 0),
		},
	}
	for i, tx1 := range txsToAdd {
		Logger.log.Warn(i, tx1.GetType(), tx1.GetMetadata(), "\n")
	}
	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
	}
	if len(instructions) != 0 {
		Logger.log.Critical("Shard Producer: Instruction", instructions)
	}
	//============End Build Body===========

	//============Build Header=============
	merkleRoots := Merkle{}.BuildMerkleTreeStore(block.Body.Transactions)
	merkleRoot := &common.Hash{}
	if len(merkleRoots) > 0 {
		merkleRoot = merkleRoots[len(merkleRoots)-1]
	}
	prevBlock := blockgen.chain.BestState.Shard[shardID].BestBlock
	prevBlockHash := prevBlock.Hash()
	crossTransactionRoot, err := CreateMerkleCrossTransaction(block.Body.CrossTransactions)
	if err != nil {
		return nil, err
	}
	txInstructions, err := CreateShardInstructionsFromTransactionAndIns(block.Body.Transactions, blockgen.chain, shardID, &producerKeySet.PaymentAddress, prevBlock.Header.Height+1, beaconBlocks, beaconHeight)
	if err != nil {
		return nil, err
	}
	totalInstructions := []string{}
	for _, value := range txInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	for _, value := range instructions {
		totalInstructions = append(totalInstructions, value...)
	}
	instructionsHash, err := GenerateHashFromStringArray(totalInstructions)
	if err != nil {
		return nil, NewBlockChainError(HashError, err)
	}
	committeeRoot, err := GenerateHashFromStringArray(shardCommittee)
	if err != nil {
		return nil, NewBlockChainError(HashError, err)
	}
	pendingValidatorRoot, err := GenerateHashFromStringArray(shardPendingValidator)
	if err != nil {
		return nil, NewBlockChainError(HashError, err)
	}
	_, shardTxMerkleData := CreateShardTxRoot2(block.Body.Transactions)
	block.Header = ShardHeader{
		ProducerAddress:      producerKeySet.PaymentAddress,
		ShardID:              shardID,
		Version:              BlockVersion,
		PrevBlockHash:        *prevBlockHash,
		Height:               prevBlock.Header.Height + 1,
		TxRoot:               *merkleRoot,
		ShardTxRoot:          shardTxMerkleData[len(shardTxMerkleData)-1],
		CrossTransactionRoot: *crossTransactionRoot,
		InstructionsRoot:     instructionsHash,
		CrossShards:          CreateCrossShardByteArray(txsToAdd, shardID),
		CommitteeRoot:        committeeRoot,
		PendingValidatorRoot: pendingValidatorRoot,
		BeaconHeight:         beaconHeight,
		BeaconHash:           beaconHash,
		Epoch:                epoch,
		Round:                proposerOffset + 1,
	}
	return block, nil
}

func (blockgen *BlkTmplGenerator) FinalizeShardBlock(blk *ShardBlock, producerKeyset *cashec.KeySet) error {
	// Signature of producer, sign on hash of header
	blk.Header.Timestamp = time.Now().Unix()
	blockHash := blk.Header.Hash()
	producerSig, err := producerKeyset.SignDataB58(blockHash.GetBytes())
	if err != nil {
		Logger.log.Error(err)
		return err
	}
	blk.ProducerSig = producerSig
	//================End Generate Signature
	return nil
}

/*
	Get Transaction For new Block
*/
func (blockgen *BlkTmplGenerator) getTransactionForNewBlock(privatekey *privacy.PrivateKey, shardID byte, db database.DatabaseInterface, beaconBlocks []*BeaconBlock) ([]metadata.Transaction, error) {
	txsToAdd, txToRemove, _ := blockgen.getPendingTransaction(shardID, beaconBlocks)
	if len(txsToAdd) == 0 {
		Logger.log.Info("Creating empty block...")
	}
	go func() {
		for _, tx := range txToRemove {
			blockgen.txPool.RemoveTx(tx)
		}
	}()

	// Process stability tx, create response txs if needed
	stabilityResponseTxs, err := blockgen.buildStabilityResponseTxsAtShardOnly(txsToAdd, privatekey)
	// Logger.log.Error(stabilityResponseTxs, "-----------------------------\n")
	if err != nil {
		return nil, err
	}
	txsToAdd = append(txsToAdd, stabilityResponseTxs...)
	stabilityResponseTxs, err = blockgen.buildStabilityResponseTxsFromInstructions(beaconBlocks, privatekey, shardID)
	if err != nil {
		return nil, err
	}
	txsToAdd = append(txsToAdd, stabilityResponseTxs...)
	return txsToAdd, nil
}

/*
	build CrossTransaction
		1. Get information for CrossShardBlock Validation
			- Get Valid Shard Block from Pool
			- Get Current Cross Shard State: BestCrossShard.ShardHeight
			- Get Current Cross Shard Bytemap height: BestCrossShard.BeaconHeight
			- Get Shard Committee for Cross Shard Block via Beacon Height
			   + Using FetchCrossShardNextHeight function in Database to determine next block height
		2. Validate
			- Greater than current cross shard state
			- Cross Shard Block Signature
			- Next Cross Shard Block via Beacon Bytemap:
				// 	When a shard block is created (ex: shard 1 create block A), it will
				// 	- Send ShardToBeacon Block (A1) to beacon,
				// 		=> ShardToBeacon Block then will be executed and store as ShardState in beacon
				// 	- Send CrossShard Block (A2) to other shard if existed
				// 		=> CrossShard Will be process into CrossTransaction
				// 	=> A1 and A2 must have the same header
				// 	- Check if A1 indicates that if A2 is exist or not via CrossShardByteMap
				// 	AND ALSO, check A2 is the only cross shard block after the most recent processed cross shard block
				// =====> Store Current and Next cross shard block in DB
		3. if miss Cross Shard Block according to beacon bytemap then stop discard the rest
		4. After validation: process valid block, extract cross output coin
*/
func (blockgen *BlkTmplGenerator) getCrossShardData(shardID byte, lastBeaconHeight uint64, currentBeaconHeight uint64, crossShards map[byte]uint64) (map[byte][]CrossTransaction, map[byte][]CrossTxTokenData) {
	crossTransactions := make(map[byte][]CrossTransaction)
	crossTxTokenData := make(map[byte][]CrossTxTokenData)
	// get cross shard block

	allCrossShardBlock := blockgen.crossShardPool[shardID].GetValidBlock(crossShards)
	// Get Cross Shard Block
	for fromShard, crossShardBlock := range allCrossShardBlock {
		sort.SliceStable(crossShardBlock[:], func(i, j int) bool {
			return crossShardBlock[i].Header.Height < crossShardBlock[j].Header.Height
		})
		indexs := []int{}
		toShard := shardID
		startHeight := blockgen.chain.BestState.Shard[toShard].BestCrossShard[fromShard]
		for index, blk := range crossShardBlock {
			if blk.Header.Height <= startHeight {
				break
			}
			nextHeight, err := blockgen.chain.config.DataBase.FetchCrossShardNextHeight(fromShard, toShard, startHeight)
			if err != nil {
				break
			}
			if nextHeight != blk.Header.Height {
				continue
			}
			startHeight = nextHeight
			temp, err := blockgen.chain.config.DataBase.FetchCommitteeByEpoch(blk.Header.BeaconHeight)
			if err != nil {
				break
			}
			shardCommittee := make(map[byte][]string)
			json.Unmarshal(temp, &shardCommittee)
			err = blk.VerifyCrossShardBlock(shardCommittee[blk.Header.ShardID])
			if err != nil {
				break
			}
			indexs = append(indexs, index)
		}
		for _, index := range indexs {
			blk := crossShardBlock[index]
			crossTransaction := CrossTransaction{
				OutputCoin:       blk.CrossOutputCoin,
				TokenPrivacyData: blk.CrossTxTokenPrivacyData,
				BlockHash:        *blk.Hash(),
				BlockHeight:      blk.Header.Height,
			}
			crossTransactions[blk.Header.ShardID] = append(crossTransactions[blk.Header.ShardID], crossTransaction)
			txTokenData := CrossTxTokenData{
				TxTokenData: blk.CrossTxTokenData,
				BlockHash:   *blk.Hash(),
				BlockHeight: blk.Header.Height,
			}
			crossTxTokenData[blk.Header.ShardID] = append(crossTxTokenData[blk.Header.ShardID], txTokenData)
		}
	}
	for _, crossTxTokenData := range crossTxTokenData {
		sort.SliceStable(crossTxTokenData[:], func(i, j int) bool {
			return crossTxTokenData[i].BlockHeight < crossTxTokenData[j].BlockHeight
		})
	}

	for _, crossTransaction := range crossTransactions {
		sort.SliceStable(crossTransaction[:], func(i, j int) bool {
			return crossTransaction[i].BlockHeight < crossTransaction[j].BlockHeight
		})
	}
	return crossTransactions, crossTxTokenData
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
func (blockgen *BlkTmplGenerator) getPendingTransaction(
	shardID byte,
	beaconBlocks []*BeaconBlock,
) (txsToAdd []metadata.Transaction, txToRemove []metadata.Transaction, totalFee uint64) {
	sourceTxns := blockgen.txPool.MiningDescs()
	isEmpty := blockgen.chain.config.TempTxPool.EmptyPool()
	if !isEmpty {
		panic("TempTxPool Is not Empty")
	}
	currentSize := uint64(0)
	startTime := time.Now()

	instsForValidations := [][]string{}
	for _, beaconBlock := range beaconBlocks {
		instsForValidations = append(instsForValidations, beaconBlock.Body.Instructions...)
	}
	instUsed := make([]int, len(instsForValidations))
	accumulatedData := component.UsedInstData{
		TradeActivated: map[string]bool{},
	}

	for i, txDesc := range sourceTxns {
		Logger.log.Criticalf("Tx index %+v value %+v", i, txDesc)
		tx := txDesc.Tx
		txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		if txShardID != shardID {
			continue
		}
		tempTxDesc, err := blockgen.chain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx)
		if err != nil {
			txToRemove = append(txToRemove, tx)
			continue
		}
		ok, err := tx.VerifyMinerCreatedTxBeforeGettingInBlock(instsForValidations, instUsed, shardID, blockgen.chain, &accumulatedData)
		if err != nil || !ok {
			txToRemove = append(txToRemove, tx)
			continue
		}

		tempTx := tempTxDesc.Tx
		totalFee += tx.GetTxFee()

		tempSize := tempTx.GetTxActualSize()
		if currentSize+tempSize >= common.MaxBlockSize {
			break
		}
		currentSize += tempSize
		txsToAdd = append(txsToAdd, tempTx)
		if len(txsToAdd) == common.MaxTxsInBlock {
			break
		}
		// Time bound condition for block creation
		//if time for getting transaction exceed half of MinShardBlkInterval then break
		elasped := time.Since(startTime)
		if elasped >= common.MinShardBlkInterval/2 {
			break
		}
	}
	blockgen.chain.config.TempTxPool.EmptyPool()
	return txsToAdd, txToRemove, totalFee
}

/*
	1. Get valid tx for specific shard and their fee, also return unvalid tx
		a. Validate Tx By it self
		b. Validate Tx with Blockchain
	2. Remove unvalid Tx out of pool
	3. Keep valid tx for new block
	4. Return total fee of tx
*/
// get valid tx for specific shard and their fee, also return unvalid tx
func (blockchain *BlockChain) createCustomTokenTxForCrossShard(privatekey *privacy.PrivateKey, crossTxTokenDataMap map[byte][]CrossTxTokenData, shardID byte) ([]metadata.Transaction, []transaction.TxTokenData) {
	var keys []int
	txs := []metadata.Transaction{}
	txTokenDataList := []transaction.TxTokenData{}
	// 0xsirrush updated: check existed tokenID
	//listCustomTokens, err := blockchain.ListCustomToken()
	//if err != nil {
	//	panic("Can't Retrieve List Custom Token in Database")
	//}
	for k := range crossTxTokenDataMap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	//	0xBahamoot optimize using waitgroup
	// var wg sync.WaitGroup

	for _, fromShardID := range keys {
		crossTxTokenDataList, _ := crossTxTokenDataMap[byte(fromShardID)]
		//crossTxTokenData is already sorted by block height
		for _, crossTxTokenData := range crossTxTokenDataList {
			for _, txTokenData := range crossTxTokenData.TxTokenData {

				if privatekey != nil {
					tx := &transaction.TxCustomToken{}
					tokenParam := &transaction.CustomTokenParamTx{
						PropertyID:     txTokenData.PropertyID.String(),
						PropertyName:   txTokenData.PropertyName,
						PropertySymbol: txTokenData.PropertySymbol,
						Amount:         txTokenData.Amount,
						TokenTxType:    transaction.CustomTokenCrossShard,
						Receiver:       txTokenData.Vouts,
					}
					err := tx.Init(
						privatekey,
						nil,
						nil,
						0,
						tokenParam,
						//listCustomTokens,
						blockchain.config.DataBase,
						nil,
						false,
						shardID,
					)
					if err != nil {
						panic("")
					}
					txs = append(txs, tx)
				} else {
					tempTxTokenData := cloneTxTokenDataForCrossShard(txTokenData)
					tempTxTokenData.Vouts = txTokenData.Vouts
					txTokenDataList = append(txTokenDataList, tempTxTokenData)
				}

			}
		}
	}
	return txs, txTokenDataList
}

// /*
// 	1. Get valid tx for specific shard and their fee, also return unvalid tx
// 		a. Validate Tx By it self
// 		b. Validate Tx with Blockchain
// 	2. Remove unvalid Tx out of pool
// 	3. Keep valid tx for new block
// 	4. Return total fee of tx
// */
// // get valid tx for specific shard and their fee, also return unvalid tx
// func (blockchain *BlockChain) createCustomTokenPrivacyTxForCrossShard(privatekey *privacy.PrivateKey, contentCrossTokenPrivacyDataMap map[byte][]ContentCrossTokenPrivacyData, shardID byte) ([]metadata.Transaction, []transaction.TxTokenData) {
// 	var keys []int
// 	compressContentCrossTokenPrivacyData := make(map[byte][]ContentCrossTokenPrivacyData)
// 	txs := []metadata.Transaction{}
// 	txTokenDataList := []transaction.TxTokenData{}
// 	listCustomTokens, err := blockchain.ListCustomToken()
// 	if err != nil {
// 		panic("Can't Retrieve List Custom Token in Database")
// 	}
// 	for k := range contentCrossTokenPrivacyDataMap {
// 		keys = append(keys, int(k))
// 	}
// 	sort.Ints(keys)
// 	for _, fromShardID := range keys {
// 		crossTxTokenPrivacyDataList, _ := contentCrossTokenPrivacyDataMap[byte(fromShardID)]
// 		//crossTxTokenData is already sorted by block height
// 		for _, crossTxTokenPrivacyData := range crossTxTokenPrivacyDataList {
// 			for _, txTokenData := range crossTxTokenPrivacyData.TxTokenData {
// 				if privatekey != nil {
// 					tx := &transaction.TxCustomToken{}
// 					tokenParam := &transaction.CustomTokenParamTx{
// 						PropertyID:     txTokenData.PropertyID.String(),
// 						PropertyName:   txTokenData.PropertyName,
// 						PropertySymbol: txTokenData.PropertySymbol,
// 						Amount:         txTokenData.Amount,
// 						TokenTxType:    transaction.CustomTokenCrossShard,
// 						Receiver:       txTokenData.Vouts,
// 					}
// 					err := tx.Init(
// 						privatekey,
// 						nil,
// 						nil,
// 						0,
// 						tokenParam,
// 						listCustomTokens,
// 						blockchain.config.DataBase,
// 						nil,
// 						false,
// 						shardID,
// 					)
// 					if err != nil {
// 						fmt.Printf("Fail to create Transaction for Cross Shard Tx Token, err %+v \n", err)
// 						panic("")
// 					}
// 					fmt.Println("CreateCustomTokenTxForCrossShard/ tx", tx)
// 					txs = append(txs, tx)
// 				} else {
// 					tempTxTokenData := cloneTxTokenDataForCrossShard(txTokenData)
// 					tempTxTokenData.Vouts = txTokenData.Vouts
// 					txTokenDataList = append(txTokenDataList, tempTxTokenData)
// 				}
// 			}
// 		}
// 	}
// 	return txs, txTokenDataList
// }
