package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/pkg/errors"
)

func (blockgen *BlkTmplGenerator) NewBlockShard(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, shardID byte, round int) (*ShardBlock, error) {
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
	crossOutputCoin := blockgen.getCrossOutputCoin(shardID, blockgen.chain.BestState.Shard[shardID].BeaconHeight, beaconHeight)
	//======Create Instruction===========================
	//Assign Instruction
	instructions := [][]string{}
	swapInstruction := []string{}
	assignInstructions := GetAssingInstructionFromBeaconBlock(beaconBlocks, shardID)
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
	actions := CreateShardActionFromTransaction(block.Body.Transactions, blockgen.chain, shardID)
	action := []string{}
	for _, value := range actions {
		action = append(action, value...)
	}
	for _, value := range instructions {
		action = append(action, value...)
	}
	actionsHash, err := GenerateHashFromStringArray(action)
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
	block.Header = ShardHeader{
		Producer:      userKeySet.GetPublicKeyB58(),
		ShardID:       shardID,
		Version:       BlockVersion,
		PrevBlockHash: *prevBlockHash,
		Height:        prevBlock.Header.Height + 1,
		Timestamp:     time.Now().Unix(),
		//TODO: add salary fund
		SalaryFund:           remainingFund,
		TxRoot:               *merkleRoot,
		ShardTxRoot:          *block.Body.CalcMerkleRootShard(blockgen.chain.BestState.Shard[shardID].ActiveShards),
		CrossOutputCoinRoot:  *crossOutputCoinRoot,
		ActionsRoot:          actionsHash,
		CrossShards:          CreateCrossShardByteArray(txsToAdd),
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
		1. Get Previous most recent proccess cross shard block
		2. Get beacon height of previous shard block
		3. Search from preBeaconHeight to currentBeaconHeight for cross shard via cross shard byte
		4. Detect in pool
		5. if miss then stop or sync block
		6. Update new most recent proccess cross shard block
*/
func (blockgen *BlkTmplGenerator) getCrossOutputCoin(shardID byte, lastBeaconHeight uint64, currentBeaconHeight uint64) map[byte][]CrossOutputCoin {
	res := make(map[byte][]CrossOutputCoin)
	crossShardMap := make(map[byte][]CrossShardBlock)
	// get cross shard block
	bestShardHeight := blockgen.chain.BestState.Beacon.BestShardHeight
	allCrossShardBlock := blockgen.crossShardPool.GetBlock(bestShardHeight)
	crossShardBlocks := allCrossShardBlock[shardID]
	currentBestCrossShard := blockgen.chain.BestState.Shard[shardID].BestCrossShard
	// Sort by height
	for _, blk := range crossShardBlocks {
		crossShardMap[blk.Header.ShardID] = append(crossShardMap[blk.Header.ShardID], blk)
	}
	// Get Cross Shard Block
	for crossShardID, crossShardBlock := range crossShardMap {
		sort.SliceStable(crossShardBlock[:], func(i, j int) bool {
			return crossShardBlock[i].Header.Height < crossShardBlock[j].Header.Height
		})
		currentBestCrossShardForThisBlock := currentBestCrossShard[crossShardID]
		for _, blk := range crossShardBlock {
			temp, err := blockgen.chain.config.DataBase.FetchBeaconCommitteeByHeight(blk.Header.BeaconHeight)
			if err != nil {
				break
			}
			shardCommittee := make(map[byte][]string)
			json.Unmarshal(temp, &shardCommittee)
			err = blk.VerifyCrossShardBlock(shardCommittee[crossShardID])
			if err != nil {
				break
			}
			lastBeaconHeight := blockgen.chain.BestState.Shard[shardID].BeaconHeight
			// Get shard state from beacon best state
			/*
				When a shard block is created (ex: shard 1 create block A), it will
				- Send ShardToBeacon Block (A1) to beacon,
					=> ShardToBeacon Block then will be executed and store as ShardState in beacon
				- Send CrossShard Block (A2) to other shard if existed
					=> CrossShard Will be process into CrossOutputCoin
				=> A1 and A2 must have the same header
				- Check if A1 indicates that if A2 is exist or not via CrossShardByteMap

				AND ALSO, check A2 is the only cross shard block after the most recent processed cross shard block
			*/
			passed := false
			for i := lastBeaconHeight + 1; i <= currentBeaconHeight; i++ {
				for shardToBeaconID, shardStates := range blockgen.chain.BestState.Beacon.AllShardState {
					if crossShardID == shardToBeaconID {
						// if the first crossShardblock is not current block then discard current block
						for i := int(currentBestCrossShardForThisBlock); i < len(shardStates); i++ {
							if bytes.Contains(shardStates[i].CrossShard, []byte{shardID}) {
								if shardStates[i].Height == blk.Header.Height {
									passed = true
								}
								break
							}
						}
					}
					if passed {
						break
					}
				}
			}
			if !passed {
				break
			}

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
	return res
}

func GetAssingInstructionFromBeaconBlock(beaconBlocks []*BeaconBlock, shardID byte) [][]string {
	assignInstruction := [][]string{}
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == "assign" && l[2] == "shard" {
				if strings.Compare(l[3], strconv.Itoa(int(shardID))) == 0 {
					assignInstruction = append(assignInstruction, l)
				}
			}
		}
	}
	return assignInstruction
}

func FetchBeaconBlockFromHeight(db database.DatabaseInterface, from uint64, to uint64) ([]*BeaconBlock, error) {
	beaconBlocks := []*BeaconBlock{}
	for i := from; i <= to; i++ {
		hash, err := db.GetBeaconBlockHashByIndex(i)
		if err != nil {
			return beaconBlocks, err
		}
		beaconBlockByte, err := db.FetchBeaconBlock(hash)
		if err != nil {
			return beaconBlocks, err
		}
		beaconBlock := BeaconBlock{}
		err = json.Unmarshal(beaconBlockByte, &beaconBlock)
		if err != nil {
			return beaconBlocks, NewBlockChainError(UnmashallJsonBlockError, err)
		}
		beaconBlocks = append(beaconBlocks, &beaconBlock)
	}
	return beaconBlocks, nil
}

func CreateCrossShardByteArray(txList []metadata.Transaction) (crossIDs []byte) {
	byteMap := make([]byte, common.MAX_SHARD_NUMBER)
	for _, tx := range txList {
		switch tx.GetType() {
		case common.TxNormalType, common.TxSalaryType:
			{
				for _, outCoin := range tx.GetProof().OutputCoins {
					lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
					shardID := common.GetShardIDFromLastByte(lastByte)
					byteMap[common.GetShardIDFromLastByte(shardID)] = 1
				}
			}
		case common.TxCustomTokenType:
			{
				customTokenTx := tx.(*transaction.TxCustomToken)
				for _, out := range customTokenTx.TxTokenData.Vouts {
					lastByte := out.PaymentAddress.Pk[len(out.PaymentAddress.Pk)-1]
					shardID := common.GetShardIDFromLastByte(lastByte)
					byteMap[common.GetShardIDFromLastByte(shardID)] = 1
				}
			}
		case common.TxCustomTokenPrivacyType:
			{
				customTokenTx := tx.(*transaction.TxCustomTokenPrivacy)
				for _, outCoin := range customTokenTx.TxTokenPrivacyData.TxNormal.GetProof().OutputCoins {
					lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
					shardID := common.GetShardIDFromLastByte(lastByte)
					byteMap[common.GetShardIDFromLastByte(shardID)] = 1
				}
			}
		}
	}

	for k := range byteMap {
		if byteMap[k] == 1 {
			crossIDs = append(crossIDs, byte(k))
		}
	}

	return crossIDs
}

/*
	Create Swap Action
	Return param:
	#1: swap instruction
	#2: new pending validator list after swapped
	#3: new committees after swapped
	#4: error
*/
func CreateSwapAction(pendingValidator []string, commitees []string, committeeSize int, shardID byte) ([]string, []string, []string, error) {
	fmt.Println("Shard Producer/Create Swap Action: pendingValidator", pendingValidator)
	fmt.Println("Shard Producer/Create Swap Action: commitees", commitees)
	newPendingValidator, newShardCommittees, shardSwapedCommittees, shardNewCommittees, err := SwapValidator(pendingValidator, commitees, committeeSize, common.OFFSET)
	if err != nil {
		return nil, nil, nil, err
	}
	swapInstruction := []string{"swap", strings.Join(shardNewCommittees, ","), strings.Join(shardSwapedCommittees, ","), "shard", strconv.Itoa(int(shardID))}
	return swapInstruction, newPendingValidator, newShardCommittees, nil
}

/*
	Action Generate From Transaction:
	- Stake
	- Stable param: set, del,...
*/
func CreateShardActionFromTransaction(transactions []metadata.Transaction, bcr metadata.BlockchainRetriever, shardID byte) (actions [][]string) {
	// Generate stake action
	stakeShardPubKey := []string{}
	stakeBeaconPubKey := []string{}
	actions = buildStabilityActions(transactions, bcr, shardID)

	for _, tx := range transactions {
		switch tx.GetMetadataType() {
		case metadata.ShardStakingMeta:
			pk := tx.GetProof().InputCoins[0].CoinDetails.PublicKey.Compress()
			pkb58 := base58.Base58Check{}.Encode(pk, common.ZeroByte)
			stakeShardPubKey = append(stakeShardPubKey, pkb58)
		case metadata.BeaconStakingMeta:
			pk := tx.GetProof().InputCoins[0].CoinDetails.PublicKey.Compress()
			pkb58 := base58.Base58Check{}.Encode(pk, common.ZeroByte)
			stakeBeaconPubKey = append(stakeBeaconPubKey, pkb58)
			//TODO: stable param 0xsancurasolus
			// case metadata.BuyFromGOVRequestMeta:
		}
	}

	if !reflect.DeepEqual(stakeShardPubKey, []string{}) {
		action := []string{"stake", strings.Join(stakeShardPubKey, ","), "shard"}
		actions = append(actions, action)
	}
	if !reflect.DeepEqual(stakeBeaconPubKey, []string{}) {
		action := []string{"stake", strings.Join(stakeBeaconPubKey, ","), "beacon"}
		actions = append(actions, action)
	}

	return actions
}

// get valid tx for specific shard and their fee, also return unvalid tx
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
		totalFee += tx.GetTxFee()
		txsToAdd = append(txsToAdd, tx)
		if len(txsToAdd) == common.MaxTxsInBlock {
			break
		}
	}
	return txsToAdd, txToRemove, totalFee
}

func (blk *ShardBlock) CreateShardToBeaconBlock(bcr metadata.BlockchainRetriever) *ShardToBeaconBlock {
	block := ShardToBeaconBlock{}
	block.AggregatedSig = blk.AggregatedSig
	block.ValidatorsIdx = make([][]int, 2)
	block.ValidatorsIdx[0] = append(block.ValidatorsIdx[0], blk.ValidatorsIdx[0]...)
	block.ValidatorsIdx[1] = append(block.ValidatorsIdx[1], blk.ValidatorsIdx[1]...)
	block.R = blk.R
	block.ProducerSig = blk.ProducerSig
	block.Header = blk.Header
	block.Instructions = blk.Body.Instructions
	actions := CreateShardActionFromTransaction(blk.Body.Transactions, bcr, blk.Header.ShardID)
	block.Instructions = append(block.Instructions, actions...)
	return &block
}

func (blk *ShardBlock) CreateAllCrossShardBlock(activeShards int) map[byte]*CrossShardBlock {
	allCrossShard := make(map[byte]*CrossShardBlock)
	if activeShards == 1 {
		return allCrossShard
	}
	for i := 0; i < activeShards; i++ {
		crossShard, err := blk.CreateCrossShardBlock(byte(i))
		if crossShard != nil && err == nil {
			allCrossShard[byte(i)] = crossShard
		}
	}
	return allCrossShard
}

func (block *ShardBlock) CreateCrossShardBlock(shardID byte) (*CrossShardBlock, error) {
	crossShard := &CrossShardBlock{}
	utxoList := getOutCoinCrossShard(block.Body.Transactions, shardID)
	if len(utxoList) == 0 {
		return nil, nil
	}
	merklePathShard, merkleShardRoot := GetMerklePathCrossShard(block.Body.Transactions, shardID)
	if merkleShardRoot != block.Header.TxRoot {
		return crossShard, NewBlockChainError(CrossShardBlockError, errors.New("MerkleRootShard mismatch"))
	}

	//Copy signature and header
	crossShard.AggregatedSig = block.AggregatedSig

	crossShard.ValidatorsIdx = make([][]int, 2)
	crossShard.ValidatorsIdx[0] = append(crossShard.ValidatorsIdx[0], block.ValidatorsIdx[0]...)
	crossShard.ValidatorsIdx[1] = append(crossShard.ValidatorsIdx[1], block.ValidatorsIdx[1]...)

	crossShard.R = block.R
	crossShard.ProducerSig = block.ProducerSig
	crossShard.Header = block.Header
	crossShard.MerklePathShard = merklePathShard
	crossShard.CrossOutputCoin = utxoList
	return crossShard, nil
}
