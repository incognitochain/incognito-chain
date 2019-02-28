package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

func (blockgen *BlkTmplGenerator) NewBlockShard(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, shardID byte, round int, crossShards map[byte]uint64) (*ShardBlock, error) {
	//============Build body=============
	// Fetch Beacon information
	beaconHeight := blockgen.chain.BestState.Beacon.BeaconHeight
	beaconHash := blockgen.chain.BestState.Beacon.BestBlockHash
	fmt.Println("Shard Producer/NewBlockShard, Beacon Height / Before", beaconHeight)
	fmt.Println("Shard Producer/NewBlockShard, Beacon Hash / Before", beaconHash)
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
	fmt.Println("Shard Producer/NewBlockShard, Beacon Height / After", beaconHeight)
	fmt.Println("Shard Producer/NewBlockShard, Beacon Hash / After", beaconHash)
	fmt.Println("Shard Producer/NewBlockShard, Beacon Epoch", epoch)
	//Fetch beacon block from height
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockgen.chain.config.DataBase, blockgen.chain.BestState.Shard[shardID].BeaconHeight+1, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	//======Get Transaction For new Block================
	txsToAdd, remainingFund := blockgen.getTransactionForNewBlock(payToAddress, privatekey, shardID, blockgen.chain.config.DataBase, beaconBlocks)
	//======Get Cross output coin from other shard=======
	crossOutputCoin := blockgen.getCrossOutputCoin(shardID, blockgen.chain.BestState.Shard[shardID].BeaconHeight, beaconHeight, crossShards)
	fmt.Println("crossOutputCoin", crossOutputCoin)
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
	block := &ShardBlock{
		Body: ShardBody{
			CrossOutputCoin: crossOutputCoin,
			Instructions:    instructions,
			Transactions:    make([]metadata.Transaction, 0),
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
	merkleRoot := merkleRoots[len(merkleRoots)-1]
	prevBlock := blockgen.chain.BestState.Shard[shardID].BestBlock
	prevBlockHash := prevBlock.Hash()

	crossOutputCoinRoot := &common.Hash{}
	if len(block.Body.CrossOutputCoin) != 0 {
		crossOutputCoinRoot, err = CreateMerkleCrossOutputCoin(block.Body.CrossOutputCoin)
	}
	if err != nil {
		return nil, err
	}
	txInstructions := CreateShardInstructionsFromTransaction(block.Body.Transactions, blockgen.chain, shardID)
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
		Producer:             userKeySet.GetPublicKeyB58(),
		ShardID:              shardID,
		Version:              BlockVersion,
		PrevBlockHash:        *prevBlockHash,
		Height:               prevBlock.Header.Height + 1,
		Timestamp:            time.Now().Unix(),
		SalaryFund:           remainingFund,
		TxRoot:               *merkleRoot,
		ShardTxRoot:          shardTxMerkleData[len(shardTxMerkleData)-1],
		CrossOutputCoinRoot:  *crossOutputCoinRoot,
		InstructionsRoot:     instructionsHash,
		CrossShards:          CreateCrossShardByteArray(txsToAdd, shardID),
		CommitteeRoot:        committeeRoot,
		PendingValidatorRoot: pendingValidatorRoot,
		BeaconHeight:         beaconHeight,
		BeaconHash:           beaconHash,
		Epoch:                epoch,
		Round:                round,
	}
	// Create producer signature
	blkHeaderHash := block.Header.Hash()
	sig, err := userKeySet.SignDataB58(blkHeaderHash.GetBytes())
	if err != nil {
		return nil, err
	}
	block.ProducerSig = sig
	_ = remainingFund
	return block, nil
}

/*
	Get Transaction For new Block
*/
func (blockgen *BlkTmplGenerator) getTransactionForNewBlock(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, shardID byte, db database.DatabaseInterface, beaconBlocks []*BeaconBlock) ([]metadata.Transaction, uint64) {
	txsToAdd, txToRemove, totalFee := blockgen.getPendingTransaction(shardID)
	if len(txsToAdd) == 0 {
		Logger.log.Info("Creating empty block...")
	}
	// Remove unrelated shard tx
	// TODO: Check again Txpool should be remove after create block is successful
	for _, tx := range txToRemove {
		blockgen.txPool.RemoveTx(tx)
	}
	// Calculate coinbases
	salaryPerTx := blockgen.rewardAgent.GetSalaryPerTx(shardID)
	basicSalary := blockgen.rewardAgent.GetBasicSalary(shardID)
	salaryFundAdd := uint64(0)
	salaryMULTP := uint64(0) //salary multiplier
	for _, blockTx := range txsToAdd {
		if blockTx.GetTxFee() > 0 {
			salaryMULTP++
		}
	}
	totalSalary := salaryMULTP*salaryPerTx + basicSalary
	salaryTx := new(transaction.Tx)
	err := salaryTx.InitTxSalary(totalSalary, payToAddress, privatekey, blockgen.chain.config.DataBase, nil)
	if err != nil {
		panic(err)
	}
	currentSalaryFund := uint64(0)
	remainingFund := currentSalaryFund + totalFee + salaryFundAdd - totalSalary
	coinbases := []metadata.Transaction{salaryTx}
	txsToAdd = append(coinbases, txsToAdd...)

	// Process stability tx, create response txs if needed
	stabilityResponseTxs, err := blockgen.buildStabilityResponseTxsAtShardOnly(txsToAdd, privatekey)
	if err != nil {
		panic(err)
	}
	txsToAdd = append(txsToAdd, stabilityResponseTxs...)

	// Process stability instructions, create response txs if needed
	stabilityResponseTxs, err = blockgen.buildStabilityResponseTxsFromInstructions(beaconBlocks, privatekey, shardID)
	if err != nil {
		panic(err)
	}
	txsToAdd = append(txsToAdd, stabilityResponseTxs...)
	return txsToAdd, remainingFund
}

/*
	build CrossOutputCoin
		1. Get information for CrossShardBlock Validation
			- Get Valid Shard Block from Pool
			- Get Current Cross Shard State: BestCrossShard.ShardHeight
			- Get Current Cross Shard Bytemap height: BestCrossShard.BeaconHeight
			- Get Shard Committee for Cross Shard Block via Beacon Height
		2. Validate
			- Greater than current cross shard state
			- Cross Shard Block Signature
			- Next Cross Shard Block via Beacon Bytemap:
				// 	When a shard block is created (ex: shard 1 create block A), it will
				// 	- Send ShardToBeacon Block (A1) to beacon,
				// 		=> ShardToBeacon Block then will be executed and store as ShardState in beacon
				// 	- Send CrossShard Block (A2) to other shard if existed
				// 		=> CrossShard Will be process into CrossOutputCoin
				// 	=> A1 and A2 must have the same header
				// 	- Check if A1 indicates that if A2 is exist or not via CrossShardByteMap
				// 	AND ALSO, check A2 is the only cross shard block after the most recent processed cross shard block
				// =====> Store Current and Next cross shard block in DB
		3. if miss Cross Shard Block according to beacon bytemap then stop discard the rest
		4. After validation: process valid block, extract cross output coin
*/
func (blockgen *BlkTmplGenerator) getCrossOutputCoin(shardID byte, lastBeaconHeight uint64, currentBeaconHeight uint64, crossShards map[byte]uint64) map[byte][]CrossOutputCoin {
	res := make(map[byte][]CrossOutputCoin)
	// crossShardMap := make(map[byte][]CrossShardBlock)
	// get cross shard block

	allCrossShardBlock := blockgen.crossShardPool[shardID].GetValidBlock(crossShards)
	fmt.Println("ShardProducer/AllCrosshardblock", allCrossShardBlock)
	// Get Cross Shard Block
	for _, crossShardBlock := range allCrossShardBlock {
		sort.SliceStable(crossShardBlock[:], func(i, j int) bool {
			return crossShardBlock[i].Header.Height < crossShardBlock[j].Header.Height
		})
		//TODO: @COMMENT for testing get crossoutput coin
		// currentBestCrossShardForThisBlock := currentBestCrossShard.ShardHeight[crossShardID]
		index := 0
		for _, blk := range crossShardBlock {
			if blk.Header.Height <= blockgen.chain.BestState.Shard[shardID].BestCrossShard.ShardHeight[blk.Header.ShardID] {
				break
			}
			temp, err := blockgen.chain.config.DataBase.FetchCommitteeByEpoch(blk.Header.Epoch)
			if err != nil {
				break
			}
			shardCommittee := make(map[byte][]string)
			json.Unmarshal(temp, &shardCommittee)
			err = blk.VerifyCrossShardBlock(shardCommittee[blk.Header.ShardID])
			fmt.Println("ShardProducer/VerifyCrossShardBlock", err == nil)
			if err != nil {
				break
			}
			index++
			//TODO: Verify block with beacon cross sahrd byte map (via function in DB)

			// lastBeaconHeight := blockgen.chain.BestState.Shard[shardID].BeaconHeight
			// // Get shard state from beacon best state
			// passed := false
			// for i := lastBeaconHeight + 1; i <= currentBeaconHeight; i++ {
			// 	shardStates, ok := blockgen.chain.BestState.Beacon.AllShardState[crossShardID]
			// 	if ok {
			// 		// if the first crossShardblock is not current block then discard current block
			// 		for i := int(currentBestCrossShardForThisBlock); i < len(shardStates); i++ {
			// 			if bytes.Contains(shardStates[i].CrossShard, []byte{shardID}) {
			// 				if shardStates[i].Height == blk.Header.Height {
			// 					passed = true
			// 				}
			// 				break
			// 			}
			// 		}
			// 	}
			// }
			// if !passed {
			// 	break
			// }
		}
		for _, blk := range crossShardBlock[:index] {
			outputCoin := CrossOutputCoin{
				OutputCoin:  blk.CrossOutputCoin,
				BlockHash:   *blk.Hash(),
				BlockHeight: blk.Header.Height,
			}
			res[blk.Header.ShardID] = append(res[blk.Header.ShardID], outputCoin)
		}
	}
	for _, crossOutputcoin := range res {
		sort.SliceStable(crossOutputcoin[:], func(i, j int) bool {
			return crossOutputcoin[i].BlockHeight < crossOutputcoin[j].BlockHeight
		})
	}
	fmt.Println("ShardProducer/CrossOutputcoin Number of cross output coin", len(res[byte(0)]))
	fmt.Println("ShardProducer/CrossOutputcoin", res)
	return res
}

/*
	1. Get valid tx for specific shard and their fee, also return unvalid tx
		a. Validate Tx By it self
		b. Validate Tx with Blockchain
	2. Remove unvalid Tx out of pool
	3. Keep valid tx for new block
	4. Return total fee of tx
*/
func (blockgen *BlkTmplGenerator) getPendingTransaction(shardID byte) (txsToAdd []metadata.Transaction, txToRemove []metadata.Transaction, totalFee uint64) {
	sourceTxns := blockgen.txPool.MiningDescs()

	// get tx and wait for more if not enough
	if len(sourceTxns) < common.MinTxsInBlock {
		<-time.Tick(common.MinBlockWaitTime * time.Second)
		sourceTxns = blockgen.txPool.MiningDescs()
		if len(sourceTxns) == 0 {
			<-time.Tick(common.MaxBlockWaitTime * time.Second)
			sourceTxns = blockgen.txPool.MiningDescs()
		}
	}

	//TODO: sort transaction base on fee and check limit block size
	// StartingPriority, fee, size, time

	// validate tx and calculate total fee
	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx
		txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		if txShardID != shardID {
			continue
		}
		// TODO: need to determine a tx is in privacy format or not
		if !tx.ValidateTxByItself(tx.IsPrivacy(), blockgen.chain.config.DataBase, blockgen.chain, shardID) {
			txToRemove = append(txToRemove, metadata.Transaction(tx))
			continue
		}
		if err := tx.ValidateTxWithBlockChain(blockgen.chain, shardID, blockgen.chain.config.DataBase); err != nil {
			txToRemove = append(txToRemove, metadata.Transaction(tx))
			continue
		}
		totalFee += tx.GetTxFee()
		txsToAdd = append(txsToAdd, tx)
		if len(txsToAdd) == common.MaxTxsInBlock {
			break
		}
	}
	return txsToAdd, txToRemove, totalFee
}
