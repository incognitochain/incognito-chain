package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
)

func (blockgen *BlkTmplGenerator) NewBlockShard(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, shardID byte, proposerOffset int, crossShards map[byte]uint64) (*ShardBlock, error) {
	//============Build body=============
	// Fetch Beacon information
	beaconHeight := blockgen.chain.BestState.Beacon.BeaconHeight
	beaconHash := blockgen.chain.BestState.Beacon.BestBlockHash
	// fmt.Println("Shard Producer/NewBlockShard, Beacon Height", beaconHeight)
	// fmt.Println("Shard Producer/NewBlockShard, Beacon Hash", beaconHash)
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
	fmt.Println("Shard Producer/NewBlockShard, Beacon Height", beaconHeight)
	fmt.Println("Shard Producer/NewBlockShard, Beacon Hash", beaconHash)
	fmt.Println("Shard Producer/NewBlockShard, Beacon Epoch", epoch)
	//Fetch beacon block from height
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockgen.chain.config.DataBase, blockgen.chain.BestState.Shard[shardID].BeaconHeight+1, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	//======Get Transaction For new Block================
	txsToAdd, err := blockgen.getTransactionForNewBlock(payToAddress, privatekey, shardID, blockgen.chain.config.DataBase, beaconBlocks)
	if err != nil {
		//@todo @0xjackalope remove panic
		panic(err)
		Logger.log.Error(err)
		return nil, err
	}
	//======Get Cross output coin from other shard=======
	crossTransactions, crossTxTokenData := blockgen.getCrossShardData(shardID, blockgen.chain.BestState.Shard[shardID].BeaconHeight, beaconHeight, crossShards)
	crossTxTokenTransactions, _ := blockgen.chain.createCustomTokenTxForCrossShard(privatekey, crossTxTokenData, shardID)
	txsToAdd = append(txsToAdd, crossTxTokenTransactions...)
	fmt.Println("crossOutputCoin", crossTransactions)
	fmt.Println("Shard Producer crossTxTokenTransactions", crossTxTokenTransactions)
	//======Create Instruction===========================
	//Assign Instruction
	instructions := [][]string{}
	swapInstruction := []string{}
	assignInstructions := GetAssignInstructionFromBeaconBlock(beaconBlocks, shardID)
	fmt.Println("Shard Block Producer AssignInstructions ", assignInstructions)
	shardPendingValidator := blockgen.chain.BestState.Shard[shardID].ShardPendingValidator
	shardCommittee := blockgen.chain.BestState.Shard[shardID].ShardCommittee
	for _, assignInstruction := range assignInstructions {
		shardPendingValidator = append(shardPendingValidator, strings.Split(assignInstruction[1], ",")...)
	}
	fmt.Println("Shard Producer: shardPendingValidator", shardPendingValidator)
	fmt.Println("Shard Producer: shardCommitee", shardCommittee)
	//Swap instruction
	// Swap instruction only appear when reach the last block in an epoch
	//@NOTICE: In this block, only pending validator change, shard committees will change in the next block
	if beaconHeight%common.EPOCH == 0 {
		if len(shardPendingValidator) > 0 {
			swapInstruction, shardPendingValidator, shardCommittee, err = CreateSwapAction(shardPendingValidator, shardCommittee, blockgen.chain.BestState.Shard[shardID].ShardCommitteeSize, shardID)
			fmt.Println("Shard Producer: swapInstruction", swapInstruction)
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
	stabilityInsts, err := blockgen.buildDividendSubmitInsts(privatekey, shardID)
	if err != nil {
		return nil, err
	}
	if stabilityInsts != nil && len(stabilityInsts) > 0 {
		instructions = append(instructions, stabilityInsts...)
	}

	block := &ShardBlock{
		Body: ShardBody{
			CrossTransactions: crossTransactions,
			Instructions:      instructions,
			Transactions:      make([]metadata.Transaction, 0),
		},
	}
	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
	}
	fmt.Println("Shard Producer: Instruction", instructions)
	fmt.Printf("Number of Transaction in blocks %+v \n", len(block.Body.Transactions))
	//============End Build Body===========

	//============Build Header=============
	//Get user key set
	userKeySet := cashec.KeySet{}
	userKeySet.ImportFromPrivateKey(privatekey)
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
	fmt.Printf("[db] buildActionReq to get hash for new shard block\n")
	txInstructions := CreateShardInstructionsFromTransactionAndIns(block.Body.Transactions, blockgen.chain, shardID, payToAddress, prevBlock.Header.Height+1, beaconBlocks)
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
	_, shardTxMerkleData := CreateShardTxRoot(block.Body.Transactions)
	fmt.Println("ShardProducer/Shard Tx Root", shardTxMerkleData[len(shardTxMerkleData)-1])
	block.Header = ShardHeader{
		ProducerAddress:      payToAddress,
		Producer:             userKeySet.GetPublicKeyB58(),
		ShardID:              shardID,
		Version:              BlockVersion,
		PrevBlockHash:        *prevBlockHash,
		Height:               prevBlock.Header.Height + 1,
		Timestamp:            time.Now().Unix(),
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
	// Create producer signature
	blkHeaderHash := block.Header.Hash()
	sig, err := userKeySet.SignDataB58(blkHeaderHash.GetBytes())
	if err != nil {
		return nil, err
	}
	block.ProducerSig = sig
	return block, nil
}

/*
	Get Transaction For new Block
*/
func (blockgen *BlkTmplGenerator) getTransactionForNewBlock(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, shardID byte, db database.DatabaseInterface, beaconBlocks []*BeaconBlock) ([]metadata.Transaction, error) {
	txsToAdd, txToRemove, _ := blockgen.getPendingTransaction(shardID)
	if len(txsToAdd) == 0 {
		Logger.log.Info("Creating empty block...")
	}
	// Remove unrelated shard tx
	for _, tx := range txToRemove {
		blockgen.txPool.RemoveTx(tx)
	}

	// Process new dividend proposal and build new dividend payment txs
	divTxs, err := blockgen.buildDividendPaymentTxs(privatekey, shardID)
	if err != nil {
		return nil, err
	}
	for _, tx := range divTxs {
		if tx != nil {
			txsToAdd = append(txsToAdd, tx)
		}
	}

	// Process stability tx, create response txs if needed
	stabilityResponseTxs, err := blockgen.buildStabilityResponseTxsAtShardOnly(txsToAdd, privatekey)
	if err != nil {
		return nil, err
	}
	txsToAdd = append(txsToAdd, stabilityResponseTxs...)

	// Process stability instructions, create response txs if needed
	//fmt.Printf("[db] start build resp from inst with %d beaconBlocks\n", len(beaconBlocks))
	//for _, b := range beaconBlocks {
	//	if len(b.Body.Instructions) > 0 {
	//		fmt.Printf("[db] b height: %d\n", b.Header.Height)
	//		fmt.Printf("[db] b inst: %+v\n", b.Body.Instructions)
	//	}
	//}
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
	fmt.Println("ShardProducer/AllCrosshardblock", allCrossShardBlock)
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
			temp, err := blockgen.chain.config.DataBase.FetchCommitteeByEpoch(blk.Header.Epoch)
			if err != nil {
				break
			}
			shardCommittee := make(map[byte][]string)
			json.Unmarshal(temp, &shardCommittee)
			err = blk.VerifyCrossShardBlock(shardCommittee[blk.Header.ShardID])
			fmt.Println("ShardProducer/VerifyCrossShardBlock", err == nil)
			if err != nil {
				fmt.Println("Shard Producer/FAIL TO Verify Crossshard block", err)
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
	fmt.Println("ShardProducer/Get data from cross shard block to shard ", shardID)
	fmt.Println("ShardProducer/crossTransactions Number of cross transaction", len(crossTransactions[shardID]))
	fmt.Println("ShardProducer/crossTransactions", crossTransactions)
	fmt.Println("ShardProducer/crossTxTokenData Number of cross custom tx token", len(crossTxTokenData[shardID]))
	fmt.Println("ShardProducer/crossTxTokenData", crossTxTokenData)
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
func (blockgen *BlkTmplGenerator) getPendingTransaction(shardID byte) (txsToAdd []metadata.Transaction, txToRemove []metadata.Transaction, totalFee uint64) {
	sourceTxns := blockgen.txPool.MiningDescs()

	//TODO: UNCOMMENT To avoid produce too many empty block
	// get tx and wait for more if not enough
	// if len(sourceTxns) < common.MinTxsInBlock {
	// 	<-time.Tick(common.MinBlockWaitTime * time.Second)
	// 	sourceTxns = blockgen.txPool.MiningDescs()
	// 	if len(sourceTxns) == 0 {
	// 		<-time.Tick(common.MaxBlockWaitTime * time.Second)
	// 		sourceTxns = blockgen.txPool.MiningDescs()
	// 	}
	// }

	//TODO: sort transaction base on fee and check limit block size
	// StartingPriority, fee, size, time
	fmt.Println("TempTxPool", reflect.TypeOf(blockgen.chain.config.TempTxPool))
	isEmpty := blockgen.chain.config.TempTxPool.EmptyPool()
	if !isEmpty {
		panic("TempTxPool Is not Empty")
	}
	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx
		tempTxDesc, err := blockgen.chain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx)
		tempTx := tempTxDesc.Tx
		if err != nil {
			txToRemove = append(txToRemove, metadata.Transaction(tempTx))
			continue
		}
		totalFee += tx.GetTxFee()
		txsToAdd = append(txsToAdd, tempTx)
		if len(txsToAdd) == common.MaxTxsInBlock {
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
func (blockchain *BlockChain) createCustomTokenTxForCrossShard(privatekey *privacy.SpendingKey, crossTxTokenDataMap map[byte][]CrossTxTokenData, shardID byte) ([]metadata.Transaction, []transaction.TxTokenData) {
	var keys []int
	txs := []metadata.Transaction{}
	txTokenDataList := []transaction.TxTokenData{}
	listCustomTokens, err := blockchain.ListCustomToken()
	if err != nil {
		panic("Can't Retrieve List Custom Token in Database")
	}
	for k := range crossTxTokenDataMap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
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
						listCustomTokens,
						blockchain.config.DataBase,
						nil,
						false,
						shardID,
					)
					if err != nil {
						fmt.Printf("Fail to create Transaction for Cross Shard Tx Token, err %+v \n", err)
						panic("")
					}
					fmt.Println("CreateCustomTokenTxForCrossShard/ tx", tx)
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
// func (blockchain *BlockChain) createCustomTokenPrivacyTxForCrossShard(privatekey *privacy.SpendingKey, contentCrossTokenPrivacyDataMap map[byte][]ContentCrossTokenPrivacyData, shardID byte) ([]metadata.Transaction, []transaction.TxTokenData) {
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
