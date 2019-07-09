package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
)

func (blockgen *BlkTmplGenerator) NewBlockShard(producerKeySet *cashec.KeySet, shardID byte, round int, crossShards map[byte]uint64, beaconHeight uint64, start time.Time) (*ShardBlock, error) {
	var txsToAdd = make([]metadata.Transaction, 0)
	totalTxsFee := map[common.Hash]uint64{}
	//============Build body=============
	// Fetch Beacon information
	Logger.log.Infof("Creating shard block%+v", blockgen.chain.BestState.Shard[shardID].ShardHeight+1)
	fmt.Printf("[ndh] Creating shard block%+v", blockgen.chain.BestState.Shard[shardID].ShardHeight+1)
	fmt.Printf("\n[db] producing block: %d\n", blockgen.chain.BestState.Shard[shardID].ShardHeight+1)
	beaconHash, err := blockgen.chain.config.DataBase.GetBeaconBlockHashByIndex(beaconHeight)
	if err != nil {
		return nil, err
	}
	beaconBlockBytes, err := blockgen.chain.config.DataBase.FetchBeaconBlock(beaconHash)
	if err != nil {
		return nil, err
	}
	beaconBlock := BeaconBlock{}
	err = json.Unmarshal(beaconBlockBytes, &beaconBlock)
	if err != nil {
		return nil, err
	}
	epoch := beaconBlock.Header.Epoch
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
	// start, block creation
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
	bridgePubkeyInst := []string{}
	if beaconHeight%common.EPOCH == 0 {
		fmt.Printf("[db] shardPendingValidator: %s\n", shardPendingValidator)
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
		// TODO(@0xbunyip): move inside previous if: only generate instruction when there's a new committee
		// Generate instruction storing merkle root of validators pubkey and send to beacon
		if shardID == byte(1) { // TODO(@0xbunyip): replace with bridge's shardID
			startHeight := blockgen.chain.BestState.Shard[shardID].ShardHeight + 2
			bridgePubkeyInst = buildBridgeSwapConfirmInstruction(shardCommittee, startHeight)
			prevBlock := blockgen.chain.BestState.Shard[shardID].BestBlock
			fmt.Printf("[db] added bridgeCommRoot in shard block %d\n", prevBlock.Header.Height+1)
		}
	}
	if !reflect.DeepEqual(swapInstruction, []string{}) {
		instructions = append(instructions, swapInstruction)
	}

	if len(bridgePubkeyInst) > 0 {
		instructions = append(instructions, bridgePubkeyInst)
		fmt.Printf("[db] build bridge pubkey root inst: %s\n", bridgePubkeyInst)
	}

	// Pick instruction with merkle root of beacon committee's pubkeys and save to bridge block
	// Also, pick BurningConfirm inst and save to bridge block
	if shardID == byte(1) { // TODO(@0xbunyip): replace with bridge's shardID
		// TODO(0xbunyip): validate these instructions in shardprocess
		commPubkeyInst := pickBeaconPubkeyRootInstruction(beaconBlocks)
		if len(commPubkeyInst) > 0 {
			instructions = append(instructions, commPubkeyInst...)
		}

		height := blockgen.chain.BestState.Shard[shardID].ShardHeight + 1
		confirmInsts := pickBurningConfirmInstruction(beaconBlocks, height)
		if len(confirmInsts) > 0 {
			bid := []uint64{}
			for _, b := range beaconBlocks {
				bid = append(bid, b.Header.Height)
			}
			prevBlock := blockgen.chain.BestState.Shard[shardID].BestBlock
			fmt.Printf("[db] picked burning confirm inst: %s %d %v\n", confirmInsts, prevBlock.Header.Height+1, bid)
			instructions = append(instructions, confirmInsts...)
		}
	}

	block := &ShardBlock{
		Body: ShardBody{
			CrossTransactions: crossTransactions,
			Instructions:      instructions,
			Transactions:      make([]metadata.Transaction, 0),
		},
	}
	prevBlock := blockgen.chain.BestState.Shard[shardID].BestBlock
	if len(instructions) != 0 {
		Logger.log.Critical("Shard Producer: Instruction", instructions)
	}

	// rewardInfoInstructions, err := block.getBlockRewardInst(prevBlock.Header.Height + 1)
	// if err != nil {
	// 	Logger.log.Error(err)
	// 	return nil, err
	// }
	// fmt.Printf("[ndh]-[INSTRUCTION AT SHARD] - - %+v\n", rewardInfoInstructions)
	//============End Build Body===========
	blockCreationLeftOver := common.MinShardBlkCreation.Nanoseconds() - time.Since(start).Nanoseconds()
	txsToAddFromBlock, err := blockgen.getTransactionForNewBlock(&producerKeySet.PrivateKey, shardID, blockgen.chain.config.DataBase, beaconBlocks, blockCreationLeftOver)
	block.Body.Transactions = append(block.Body.Transactions, txsToAdd...)
	block.Body.Transactions = append(block.Body.Transactions, txsToAddFromBlock...)
	if err != nil {
		Logger.log.Error(err, reflect.TypeOf(err), reflect.ValueOf(err))
		return nil, err
	}
	err = blockgen.chain.BuildResponseTransactionFromTxsWithMetadata(&block.Body, &producerKeySet.PrivateKey)
	if err != nil {
		fmt.Printf("[ndh] BuildResponseTransactionFromTxsWithMetadata err %+v \n", err)
		return nil, err
	}
	//TODO calculate fee for another tx type
	for _, tx := range block.Body.Transactions {
		totalTxsFee[*tx.GetTokenID()] += tx.GetTxFee()
		txType := tx.GetType()
		// fmt.Printf("[ndh] - - - - TxType %+v\n", txType)
		if txType == common.TxCustomTokenPrivacyType {
			txCustomPrivacy := tx.(*transaction.TxCustomTokenPrivacy)
			totalTxsFee[*txCustomPrivacy.GetTokenID()] = txCustomPrivacy.GetTxFeeToken()
			// fmt.Printf("[ndh]####################### %+v %+v\n", *txCustomPrivacy.GetTokenID(), totalTxsFee[*txCustomPrivacy.GetTokenID()])
		}
	}
	// for key, value := range totalTxsFee {
	// 	fmt.Printf("[ndh] - key %+v Value: %+v\n", key, value)
	// }
	//============Build Header=============
	merkleRoots := Merkle{}.BuildMerkleTreeStore(block.Body.Transactions)
	merkleRoot := &common.Hash{}
	if len(merkleRoots) > 0 {
		merkleRoot = merkleRoots[len(merkleRoots)-1]
	}
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

	// Instruction merkle root
	flattenTxInsts := FlattenAndConvertStringInst(txInstructions)
	flattenInsts := FlattenAndConvertStringInst(instructions)
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be preserved in shardprocess
	instMerkleRoot := GetKeccak256MerkleRoot(insts)
	if len(insts) >= 2 {
		fmt.Printf("[db] block %d has %d insts\n", prevBlock.Header.Height, len(insts))
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
		CrossShards:          CreateCrossShardByteArray(block.Body.Transactions, shardID),
		CommitteeRoot:        committeeRoot,
		PendingValidatorRoot: pendingValidatorRoot,
		BeaconHeight:         beaconHeight,
		BeaconHash:           beaconHash,
		TotalTxsFee:          totalTxsFee,
		Epoch:                epoch,
		Round:                round,
	}
	copy(block.Header.InstructionMerkleRoot[:], instMerkleRoot)
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
func (blockgen *BlkTmplGenerator) getTransactionForNewBlock(privatekey *privacy.PrivateKey, shardID byte, db database.DatabaseInterface, beaconBlocks []*BeaconBlock, blockCreation int64) ([]metadata.Transaction, error) {
	txsToAdd, txToRemove, _ := blockgen.getPendingTransactionV2(shardID, beaconBlocks, blockCreation)
	if len(txsToAdd) == 0 {
		Logger.log.Info("Creating empty block...")
	}
	go blockgen.txPool.RemoveTx(txToRemove, false)
	go func() {
		for _, tx := range txToRemove {
			blockgen.chain.config.CRemovedTxs <- tx
		}
	}()

	var respTxsShard, respTxsBeacon []metadata.Transaction
	var errCh chan error
	errCh = make(chan error)
	go func() {
		var err error
		respTxsShard, err = blockgen.buildStabilityResponseTxsAtShardOnly(txsToAdd, privatekey, shardID)
		errCh <- err
	}()

	go func() {
		var err error
		respTxsBeacon, err = blockgen.buildResponseTxsFromBeaconInstructions(beaconBlocks, privatekey, shardID)
		errCh <- err
	}()

	nilCount := 0
	for {
		err := <-errCh
		if err != nil {
			return nil, err
		}
		nilCount++
		if nilCount == 2 {
			break
		}
	}
	txsToAdd = append(txsToAdd, respTxsShard...)
	txsToAdd = append(txsToAdd, respTxsBeacon...)
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
func (blockgen *BlkTmplGenerator) getPendingTransactionV2(
	shardID byte,
	beaconBlocks []*BeaconBlock,
	blockCreationTime int64,
) (txsToAdd []metadata.Transaction, txToRemove []metadata.Transaction, totalFee uint64) {
	startTime := time.Now()
	sourceTxns := blockgen.GetPendingTxsV2()
	txsProcessTimeInBlockCreation := int64(common.MinShardBlkInterval.Nanoseconds())
	//txsProcessTimeInBlockCreation := int64(float64(common.MinShardBlkInterval.Nanoseconds()) * MaxTxsProcessTimeInBlockCreation)
	var elasped int64
	Logger.log.Critical("Number of transaction get from pool: ", len(sourceTxns))
	isEmpty := blockgen.chain.config.TempTxPool.EmptyPool()
	if !isEmpty {
		panic("TempTxPool Is not Empty")
	}
	currentSize := uint64(0)
	for _, tx := range sourceTxns {
		if tx.IsPrivacy() {
			txsProcessTimeInBlockCreation = blockCreationTime - time.Duration(500*time.Millisecond).Nanoseconds()
		} else {
			txsProcessTimeInBlockCreation = blockCreationTime - time.Duration(50*time.Millisecond).Nanoseconds()
		}
		elasped = time.Since(startTime).Nanoseconds()
		// @txsProcessTimeInBlockCreation is a constant for this current version
		if elasped >= txsProcessTimeInBlockCreation {
			Logger.log.Critical("Shard Producer/Elapsed, Break: ", elasped)
			break
		}
		//Logger.log.Criticalf("Tx index %+v value %+v", i, txDesc)
		txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		if txShardID != shardID {
			continue
		}
		tempTxDesc, err := blockgen.chain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx)
		if err != nil {
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
	}
	Logger.log.Info("Max Transaction In Block ⚡︎ ", MaxTxsInBlock)
	Logger.log.Criticalf("☭ %+v transactions for New Block from pool \n", len(txsToAdd))
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
	for k := range crossTxTokenDataMap {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)

	//	0xmerman optimize using waitgroup
	// var wg sync.WaitGroup
	for _, fromShardID := range keys {
		crossTxTokenDataList := crossTxTokenDataMap[byte(fromShardID)]
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
