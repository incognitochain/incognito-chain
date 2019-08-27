package blockchain

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
)

/*
	Create New block Shard
	1. Identify Beacon State for this Shard Block: Beacon Hash & Beacon Height & Epoch
		+ Get Beacon Block (B) from Beacon Best State (from Beacon Chain of Shard Node)
		+ Beacon Block (B) must have the same epoch With New Shard Block (S):
		+ If Beacon Block (B) have different height previous shard block PS (previous of S)
		Then Beacon Block (B) epoch greater than Shard Block (S) epoch exact 1 value
		BUT This only works if Shard Best State have the Beacon Height divisible by epoch
		+ Ex: 1 epoch has 50 block
		Example 1:
			shard block with
				height 10,
				epoch: 1,
				beacon block height: 49
			then shard block with
				height 11 must have
				epoch: 1,
				beacon block height: must be 49 or 50
		Example 2:
			shard block with
				height 10,
				epoch: 1,
				beacon block height: 50
			then shard block with
				height is 11 can have 2 option:
				a. epoch: 1, if beacon block height remain 50
				b. epoch: 2, and beacon block must in range from 51-100
				Can have beacon block with height > 100
	2. Build Shard Block Body:
		a. Get Cross Transaction from other shard && Build Cross Shard Tx Custom Token Transaction (if exist)
		b. Get Transactions for New Block
		c. Process Assign Instructions from Beacon Blocks
		c. Generate Instructions
	3. Build Shard Block Header
*/
func (blockGenerator *BlockGenerator) NewBlockShard(shardID byte, round int, crossShards map[byte]uint64, beaconHeight uint64, start time.Time) (*ShardBlock, error) {
	var (
		transactionsForNewBlock = make([]metadata.Transaction, 0)
		totalTxsFee             = make(map[common.Hash]uint64)
		block                   = NewShardBlock()
		instructions            = [][]string{}
		shardPendingValidator   = incognitokey.CommitteeKeyListToString(blockGenerator.chain.BestState.Shard[shardID].ShardPendingValidator)
		currentCommitteePubKeys = incognitokey.CommitteeKeyListToString(blockGenerator.chain.BestState.Shard[shardID].ShardCommittee)
		tempPrivateKey          = blockGenerator.createTempKeyset()
	)
	Logger.log.Criticalf("⛏ Creating Shard Block %+v", blockGenerator.chain.BestState.Shard[shardID].ShardHeight+1)
	//============Build body===============
	// Fetch Beacon information
	BLogger.log.Infof("Producing block: %d", blockGenerator.chain.BestState.Shard[shardID].ShardHeight+1)
	beaconHash, err := blockGenerator.chain.config.DataBase.GetBeaconBlockHashByIndex(beaconHeight)
	if err != nil {
		return nil, err
	}
	beaconBlockBytes, err := blockGenerator.chain.config.DataBase.FetchBeaconBlock(beaconHash)
	if err != nil {
		return nil, err
	}
	beaconBlock := BeaconBlock{}
	err = json.Unmarshal(beaconBlockBytes, &beaconBlock)
	if err != nil {
		return nil, err
	}
	epoch := beaconBlock.Header.Epoch
	if epoch-blockGenerator.chain.BestState.Shard[shardID].Epoch > 1 {
		beaconHeight = blockGenerator.chain.BestState.Shard[shardID].Epoch * blockGenerator.chain.config.ChainParams.Epoch
		newBeaconHash, err := blockGenerator.chain.config.DataBase.GetBeaconBlockHashByIndex(beaconHeight)
		if err != nil {
			return nil, err
		}
		copy(beaconHash[:], newBeaconHash.GetBytes())
		epoch = blockGenerator.chain.BestState.Shard[shardID].Epoch + 1
	}
	//Fetch beacon block from height
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockGenerator.chain.config.DataBase, blockGenerator.chain.BestState.Shard[shardID].BeaconHeight+1, beaconHeight)
	if err != nil {
		return nil, err
	}
	//======Get Transaction For new Block================
	// Get Cross output coin from other shard && produce cross shard transaction
	crossTransactions, crossTxTokenData := blockGenerator.getCrossShardData(shardID, blockGenerator.chain.BestState.Shard[shardID].BeaconHeight, beaconHeight, crossShards)
	crossTxTokenTransactions, _, err := blockGenerator.chain.createNormalTokenTxForCrossShard(&tempPrivateKey, crossTxTokenData, shardID)
	if err != nil {
		return nil, err
	}
	transactionsForNewBlock = append(transactionsForNewBlock, crossTxTokenTransactions...)
	// Get Transaction for new block
	blockCreationLeftOver := blockGenerator.chain.BestState.Shard[shardID].BlockMaxCreateTime.Nanoseconds() - time.Since(start).Nanoseconds()
	txsToAddFromBlock, err := blockGenerator.getTransactionForNewBlock(&tempPrivateKey, shardID, blockGenerator.chain.config.DataBase, beaconBlocks, blockCreationLeftOver)
	if err != nil {
		return nil, err
	}
	transactionsForNewBlock = append(transactionsForNewBlock, txsToAddFromBlock...)
	// build txs with metadata
	txsWithMetadata, err := blockGenerator.chain.BuildResponseTransactionFromTxsWithMetadata(transactionsForNewBlock, &tempPrivateKey)
	if err != nil {
		return nil, err
	}
	transactionsForNewBlock = append(transactionsForNewBlock, txsWithMetadata...)
	// process instruction from beacon
	shardPendingValidator = blockGenerator.chain.processInstructionFromBeacon(beaconBlocks, shardID)
	// Create Instruction
	instructions, shardPendingValidator, shardCommitteePubKeys, err := blockGenerator.chain.generateInstruction(shardID, beaconHeight, beaconBlocks, shardPendingValidator, currentCommitteePubKeys)
	if err != nil {
		return nil, NewBlockChainError(GenerateInstructionError, err)
	}
	if len(instructions) != 0 {
		Logger.log.Info("Shard Producer: Instruction", instructions)
	}
	block.BuildShardBlockBody(instructions, crossTransactions, transactionsForNewBlock)
	//============End Build Body===========
	//============Build Header=============
	previousBlock := blockGenerator.chain.BestState.Shard[shardID].BestBlock
	//TODO calculate fee for another tx type
	for _, tx := range block.Body.Transactions {
		totalTxsFee[*tx.GetTokenID()] += tx.GetTxFee()
		txType := tx.GetType()
		if txType == common.TxCustomTokenPrivacyType {
			txCustomPrivacy := tx.(*transaction.TxCustomTokenPrivacy)
			totalTxsFee[*txCustomPrivacy.GetTokenID()] = txCustomPrivacy.GetTxFeeToken()
		}
	}
	merkleRoots := Merkle{}.BuildMerkleTreeStore(block.Body.Transactions)
	merkleRoot := &common.Hash{}
	if len(merkleRoots) > 0 {
		merkleRoot = merkleRoots[len(merkleRoots)-1]
	}
	previousBlockHash := previousBlock.Hash()
	crossTransactionRoot, err := CreateMerkleCrossTransaction(block.Body.CrossTransactions)
	if err != nil {
		return nil, err
	}
	txInstructions, err := CreateShardInstructionsFromTransactionAndInstruction(block.Body.Transactions, blockGenerator.chain, shardID)
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
	instructionsHash, err := generateHashFromStringArray(totalInstructions)
	if err != nil {
		return nil, NewBlockChainError(InstructionsHashError, err)
	}

	committeeRoot, err := generateHashFromStringArray(shardCommitteePubKeys)
	if err != nil {
		return nil, NewBlockChainError(HashError, err)
	}
	pendingValidatorRoot, err := generateHashFromStringArray(shardPendingValidator)
	if err != nil {
		return nil, NewBlockChainError(HashError, err)
	}
	// Instruction merkle root
	flattenTxInsts, err := FlattenAndConvertStringInst(txInstructions)
	if err != nil {
		return nil, NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from Tx: %+v", err))
	}
	flattenInsts, err := FlattenAndConvertStringInst(instructions)
	if err != nil {
		return nil, NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from block body: %+v", err))
	}
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be preserved in shardprocess
	instMerkleRoot := GetKeccak256MerkleRoot(insts)
	_, shardTxMerkleData := CreateShardTxRoot2(block.Body.Transactions)

	producerPosition := (blockGenerator.chain.BestState.Shard[shardID].ShardProposerIdx + round) % len(currentCommitteePubKeys)

	committeeMiningKeys, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(blockGenerator.chain.BestState.Shard[shardID].ShardCommittee, blockGenerator.chain.BestState.Shard[shardID].ConsensusAlgorithm)

	block.Header = ShardHeader{
		Producer:             committeeMiningKeys[producerPosition],
		ConsensusType:        blockGenerator.chain.BestState.Shard[shardID].ConsensusAlgorithm,
		ShardID:              shardID,
		Version:              SHARD_BLOCK_VERSION,
		PreviousBlockHash:    *previousBlockHash,
		Height:               previousBlock.Header.Height + 1,
		TxRoot:               *merkleRoot,
		ShardTxRoot:          shardTxMerkleData[len(shardTxMerkleData)-1],
		CrossTransactionRoot: *crossTransactionRoot,
		InstructionsRoot:     instructionsHash,
		CrossShardBitMap:     CreateCrossShardByteArray(block.Body.Transactions, shardID),
		CommitteeRoot:        committeeRoot,
		PendingValidatorRoot: pendingValidatorRoot,
		BeaconHeight:         beaconHeight,
		BeaconHash:           beaconHash,
		TotalTxsFee:          totalTxsFee,
		Epoch:                epoch,
		Round:                round,
		Timestamp:            time.Now().Unix(),
	}
	copy(block.Header.InstructionMerkleRoot[:], instMerkleRoot)
	return block, nil
}

// func (blockGenerator *BlockGenerator) FinalizeShardBlock(blk *ShardBlock, producerKeyset *incognitokey.KeySet) error {
// 	// Signature of producer, sign on hash of header
// 	blk.Header.Timestamp = time.Now().Unix()
// 	blockHash := blk.Header.Hash()
// 	producerSig, err := producerKeyset.SignDataInBase58CheckEncode(blockHash.GetBytes())
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return err
// 	}
// 	blk.ProducerSig = producerSig
// 	//End Generate Signature
// 	return nil
// }

/*
	Get Transaction For new Block
	1. Get pending transaction from blockgen
	2. Keep valid tx & Removed error tx
	3. Build response Transaction For Shard
	4. Build response Transaction For Beacon
	5. Return valid transaction from pending, response transactions from shard and beacon
*/
func (blockGenerator *BlockGenerator) getTransactionForNewBlock(privatekey *privacy.PrivateKey, shardID byte, db database.DatabaseInterface, beaconBlocks []*BeaconBlock, blockCreation int64) ([]metadata.Transaction, error) {
	txsToAdd, txToRemove, _ := blockGenerator.getPendingTransaction(shardID, beaconBlocks, blockCreation)
	if len(txsToAdd) == 0 {
		Logger.log.Info("Creating empty block...")
	}
	go blockGenerator.txPool.RemoveTx(txToRemove, false)
	go func() {
		for _, tx := range txToRemove {
			blockGenerator.chain.config.CRemovedTxs <- tx
		}
	}()
	var responsedTxsBeacon []metadata.Transaction
	var cError chan error
	cError = make(chan error)
	go func() {
		var err error
		responsedTxsBeacon, err = blockGenerator.buildResponseTxsFromBeaconInstructions(beaconBlocks, privatekey, shardID)
		cError <- err
	}()
	nilCount := 0
	for {
		err := <-cError
		if err != nil {
			return nil, err
		}
		nilCount++
		if nilCount == 1 {
			break
		}
	}
	txsToAdd = append(txsToAdd, responsedTxsBeacon...)
	return txsToAdd, nil
}

// buildResponseTxsFromBeaconInstructions builds response txs from beacon instructions
func (blockGenerator *BlockGenerator) buildResponseTxsFromBeaconInstructions(beaconBlocks []*BeaconBlock, producerPrivateKey *privacy.PrivateKey, shardID byte) ([]metadata.Transaction, error) {
	responsedTxs := []metadata.Transaction{}
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == SwapAction {
				for _, v := range strings.Split(l[2], ",") {
					tx, err := blockGenerator.buildReturnStakingAmountTx(v, producerPrivateKey)
					if err != nil {
						Logger.log.Error(err)
						continue
					}
					responsedTxs = append(responsedTxs, tx)
				}

			}
			if l[0] == StakeAction || l[0] == RandomAction || l[0] == AssignAction || l[0] == SwapAction {
				continue
			}
			if len(l) <= 2 {
				continue
			}
			metaType, err := strconv.Atoi(l[0])
			if err != nil {
				return nil, err
			}
			var newTx metadata.Transaction
			switch metaType {
			case metadata.IssuingETHRequestMeta:
				if len(l) >= 4 && l[2] == "accepted" {
					newTx, err = blockGenerator.buildETHIssuanceTx(l[3], producerPrivateKey, shardID)
				}
			case metadata.IssuingRequestMeta:
				if len(l) >= 4 && l[2] == "accepted" {
					newTx, err = blockGenerator.buildIssuanceTx(l[3], producerPrivateKey, shardID)
				}

			default:
				continue
			}
			if err != nil {
				return nil, err
			}
			if newTx != nil {
				responsedTxs = append(responsedTxs, newTx)
			}
		}
	}
	return responsedTxs, nil
}

/*
	Process Instruction From Beacon Blocks:
	- Assign Instruction: get more pending validator from beacon and return new list of pending validator
*/
func (blockchain *BlockChain) processInstructionFromBeacon(beaconBlocks []*BeaconBlock, shardID byte) []string {
	shardPendingValidator := incognitokey.CommitteeKeyListToString(blockchain.BestState.Shard[shardID].ShardPendingValidator)
	assignInstructions := GetAssignInstructionFromBeaconBlock(beaconBlocks, shardID)
	if len(assignInstructions) != 0 {
		Logger.log.Info("Shard Block Producer Assign Instructions ", assignInstructions)
	}
	for _, assignInstruction := range assignInstructions {
		shardPendingValidator = append(shardPendingValidator, strings.Split(assignInstruction[1], ",")...)
	}
	return shardPendingValidator
}

/*
	Generate Instruction:
	- Swap: at the end of beacon epoch
	- Brigde: at the end of beacon epoch
	Return params:
	#1: instruction list
	#2: shardpendingvalidator
	#3: shardcommittee
	#4: error
*/
func (blockchain *BlockChain) generateInstruction(shardID byte, beaconHeight uint64, beaconBlocks []*BeaconBlock, shardPendingValidator []string, shardCommittee []string) ([][]string, []string, []string, error) {
	var (
		instructions          = [][]string{}
		bridgeSwapConfirmInst = []string{}
		swapInstruction       = []string{}
		err                   error
	)
	if beaconHeight%blockchain.config.ChainParams.Epoch == 0 {
		if len(shardPendingValidator) > 0 {
			Logger.log.Info("ShardPendingValidator", shardPendingValidator)
			Logger.log.Info("ShardCommittee", shardCommittee)
			Logger.log.Info("MaxShardCommitteeSize", blockchain.BestState.Shard[shardID].MaxShardCommitteeSize)
			Logger.log.Info("ShardID", shardID)
			swapInstruction, shardPendingValidator, shardCommittee, err = CreateSwapAction(shardPendingValidator, shardCommittee, blockchain.BestState.Shard[shardID].MaxShardCommitteeSize, shardID)
			if err != nil {
				Logger.log.Error(err)
				return instructions, shardPendingValidator, shardCommittee, err
			}
			// Generate instruction storing merkle root of validators pubkey and send to beacon
			bridgeID := byte(common.BridgeShardID)
			if shardID == bridgeID {
				startHeight := blockchain.BestState.Shard[shardID].ShardHeight + 2
				bridgeSwapConfirmInst = buildBridgeSwapConfirmInstruction(shardCommittee, startHeight)
				prevBlock := blockchain.BestState.Shard[shardID].BestBlock
				BLogger.log.Infof("Add Bridge swap inst in ShardID %+v block %d", shardID, prevBlock.Header.Height+1)
			}
		}
	}
	if len(swapInstruction) > 0 {
		instructions = append(instructions, swapInstruction)
	}
	if len(bridgeSwapConfirmInst) > 0 {
		instructions = append(instructions, bridgeSwapConfirmInst)
		Logger.log.Infof("Build bridge swap confirm inst: %s \n", bridgeSwapConfirmInst)
	}
	// Pick instruction with merkle root of beacon committee's pubkeys and save to bridge block
	// Also, pick BurningConfirm inst and save to bridge block
	bridgeID := byte(common.BridgeShardID)
	if shardID == bridgeID {
		prevBlock := blockchain.BestState.Shard[shardID].BestBlock
		commPubkeyInst := pickBeaconSwapConfirmInst(beaconBlocks)
		if len(commPubkeyInst) > 0 {
			instructions = append(instructions, commPubkeyInst...)
			BLogger.log.Infof("Found beacon swap confirm inst and add to bridge block %d: %s", prevBlock.Header.Height+1, commPubkeyInst)
		}
		height := blockchain.BestState.Shard[shardID].ShardHeight + 1
		confirmInsts := pickBurningConfirmInstruction(beaconBlocks, height)
		if len(confirmInsts) > 0 {
			bid := []uint64{}
			for _, b := range beaconBlocks {
				bid = append(bid, b.Header.Height)
			}
			Logger.log.Infof("Picked burning confirm inst: %s %d %v\n", confirmInsts, prevBlock.Header.Height+1, bid)
			instructions = append(instructions, confirmInsts...)
		}
	}
	return instructions, shardPendingValidator, shardCommittee, nil
}

/*
	getCrossShardData get cross shard data from cross shard block
		1. Get Cross Shard Block and Validate
			a. Get Valid Cross Shard Block from Cross Shard Pool
			b. Get Current Cross Shard State: Last Cross Shard Block From other Shard (FS) to this shard (TS) (Ex: last cross shard block from Shard 0 to Shard 1)
			c. Get Next Cross Shard Block Height from other Shard (FS) to this shard (TS)
			   + Using FetchCrossShardNextHeight function in Database to determine next block height
			d. Fetch Other Shard (FS) Committee at Next Cross Shard Block Height for Validation
		2. Validate
			a. Get Next Cross Shard Height from Database
			a. Cross Shard Block Height is Next Cross Shard Height from Database (if miss Cross Shard Block according to beacon bytemap then stop discard the rest)
			b. Verify Cross Shard Block Signature
		4. After validation:
			- Process valid block to extract:
				+ Cross output coin
				+ Cross Normal Token
*/
func (blockGenerator *BlockGenerator) getCrossShardData(toShard byte, lastBeaconHeight uint64, currentBeaconHeight uint64, crossShards map[byte]uint64) (map[byte][]CrossTransaction, map[byte][]CrossTxTokenData) {
	crossTransactions := make(map[byte][]CrossTransaction)
	crossTxTokenData := make(map[byte][]CrossTxTokenData)
	// get cross shard block
	allCrossShardBlock := blockGenerator.crossShardPool[toShard].GetValidBlock(crossShards)
	// Get Cross Shard Block
	for fromShard, crossShardBlock := range allCrossShardBlock {
		sort.SliceStable(crossShardBlock[:], func(i, j int) bool {
			return crossShardBlock[i].Header.Height < crossShardBlock[j].Header.Height
		})
		indexs := []int{}
		startHeight := blockGenerator.chain.BestState.Shard[toShard].BestCrossShard[fromShard]
		for index, crossShardBlock := range crossShardBlock {
			if crossShardBlock.Header.Height <= startHeight {
				break
			}
			nextHeight, err := blockGenerator.chain.config.DataBase.FetchCrossShardNextHeight(fromShard, toShard, startHeight)
			if err != nil {
				break
			}
			if nextHeight != crossShardBlock.Header.Height {
				continue
			}
			startHeight = nextHeight
			beaconHeight, err := blockGenerator.chain.FindBeaconHeightForCrossShardBlock(crossShardBlock.Header.BeaconHeight, crossShardBlock.Header.ShardID, crossShardBlock.Header.Height)
			if err != nil {
				break
			}
			temp, err := blockGenerator.chain.config.DataBase.FetchShardCommitteeByHeight(beaconHeight)
			if err != nil {
				break
			}
			shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
			err = json.Unmarshal(temp, &shardCommittee)
			if err != nil {
				break
			}
			err = crossShardBlock.VerifyCrossShardBlock(blockGenerator.chain, shardCommittee[crossShardBlock.Header.ShardID])
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
	Verify Transaction with these condition: defined in mempool.go
*/
func (blockGenerator *BlockGenerator) getPendingTransaction(
	shardID byte,
	beaconBlocks []*BeaconBlock,
	blockCreationTime int64,
) (txsToAdd []metadata.Transaction, txToRemove []metadata.Transaction, totalFee uint64) {
	startTime := time.Now()
	sourceTxns := blockGenerator.GetPendingTxsV2()
	txsProcessTimeInBlockCreation := int64(blockGenerator.chain.BestState.Shard[shardID].BlockMaxCreateTime.Nanoseconds())
	var elasped int64
	Logger.log.Info("Number of transaction get from Block Generator: ", len(sourceTxns))
	isEmpty := blockGenerator.chain.config.TempTxPool.EmptyPool()
	if !isEmpty {
		return []metadata.Transaction{}, []metadata.Transaction{}, 0
	}
	currentSize := uint64(0)
	for _, tx := range sourceTxns {
		if tx.IsPrivacy() {
			txsProcessTimeInBlockCreation = blockCreationTime - time.Duration(500*time.Millisecond).Nanoseconds()
		} else {
			txsProcessTimeInBlockCreation = blockCreationTime - time.Duration(50*time.Millisecond).Nanoseconds()
		}
		elasped = time.Since(startTime).Nanoseconds()
		if elasped >= txsProcessTimeInBlockCreation {
			Logger.log.Info("Shard Producer/Elapsed, Break: ", elasped)
			break
		}
		txShardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
		if txShardID != shardID {
			continue
		}
		tempTxDesc, err := blockGenerator.chain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx)
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
	Logger.log.Criticalf(" 🔎 %+v transactions for New Block from pool \n", len(txsToAdd))
	blockGenerator.chain.config.TempTxPool.EmptyPool()
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
func (blockchain *BlockChain) createNormalTokenTxForCrossShard(privatekey *privacy.PrivateKey, crossTxTokenDataMap map[byte][]CrossTxTokenData, shardID byte) ([]metadata.Transaction, []transaction.TxNormalTokenData, error) {
	var keys []int
	txs := []metadata.Transaction{}
	txTokenDataList := []transaction.TxNormalTokenData{}
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
					tx := &transaction.TxNormalToken{}
					tokenParam := &transaction.CustomTokenParamTx{
						PropertyID:     txTokenData.PropertyID.String(),
						PropertyName:   txTokenData.PropertyName,
						PropertySymbol: txTokenData.PropertySymbol,
						Amount:         txTokenData.Amount,
						TokenTxType:    transaction.CustomTokenCrossShard,
						Receiver:       txTokenData.Vouts,
					}
					err := tx.Init(
						transaction.NewTxNormalTokenInitParam(privatekey,
							nil,
							nil,
							0,
							tokenParam,
							//listCustomTokens,
							blockchain.config.DataBase,
							nil,
							false,
							shardID))
					if err != nil {
						return []metadata.Transaction{}, []transaction.TxNormalTokenData{}, NewBlockChainError(CreateNormalTokenTxForCrossShardError, err)
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
	return txs, txTokenDataList, nil
}

/*
	Find Beacon Block with compatible shard states of cross shard block
*/
func (blockchain *BlockChain) FindBeaconHeightForCrossShardBlock(beaconHeight uint64, fromShardID byte, crossShardBlockHeight uint64) (uint64, error) {
	for {
		beaconBlock, err := blockchain.GetBeaconBlockByHeight(beaconHeight)
		if err != nil {
			return 0, NewBlockChainError(FetchBeaconBlockError, err)
		}
		if shardStates, ok := beaconBlock.Body.ShardState[fromShardID]; ok {
			for _, shardState := range shardStates {
				if shardState.Height == crossShardBlockHeight {
					return beaconBlock.Header.Height, nil
				}
			}
		}
		beaconHeight += 1
	}
}

func (blockGenerator *BlockGenerator) createTempKeyset() privacy.PrivateKey {
	rand.Seed(time.Now().UnixNano())
	seed := make([]byte, 16)
	rand.Read(seed)
	return privacy.GeneratePrivateKey(seed)
}
