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

func (blockgen *BlkTmplGenerator) NewBlockShard(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, shardID byte) (*ShardBlock, error) {
	//============Build body=============
	beaconHeight := blockgen.chain.BestState.Beacon.BeaconHeight
	beaconHash := blockgen.chain.BestState.Beacon.BestBlockHash
	epoch := blockgen.chain.BestState.Beacon.BeaconEpoch
	if epoch-blockgen.chain.BestState.Shard[shardID].Epoch > 1 {
		beaconHeight = blockgen.chain.BestState.Shard[shardID].Epoch * common.EPOCH
		epoch = blockgen.chain.BestState.Shard[shardID].Epoch + 1
	}

	// Get valid transaction (add tx, remove tx, fee of add tx)
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
		Logger.log.Error(err)
		return nil, err
	}
	currentSalaryFund := uint64(0)
	remainingFund := currentSalaryFund + totalFee + salaryFundAdd - totalSalary
	coinbases := []metadata.Transaction{salaryTx}
	txsToAdd = append(coinbases, txsToAdd...)

	//Crossoutputcoint
	crossOutputCoin := blockgen.getCrossOutputCoin(shardID, blockgen.chain.BestState.Shard[shardID].BeaconHeight, beaconHeight)
	//Assign Instruction
	//Fetch beacon block from height
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockgen.chain.config.DataBase, blockgen.chain.BestState.Shard[shardID].BeaconHeight+1, beaconHeight)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	assignInstructions := GetAssingInstructionFromBeaconBlock(beaconBlocks, shardID)
	fmt.Println("Shard Block Producer AssignInstructions ", assignInstructions)
	shardPendingValidator := blockgen.chain.BestState.Shard[shardID].ShardPendingValidator
	shardCommittees := blockgen.chain.BestState.Shard[shardID].ShardCommittee
	for _, assignInstruction := range assignInstructions {
		shardPendingValidator = append(shardPendingValidator, strings.Split(assignInstruction[1], ",")...)
	}
	//Swap instruction
	instructions := [][]string{}
	swapInstruction := []string{}
	// Swap instruction only appear when reach the last block in an epoch
	//@NOTICE: In this block, only pending validator change, shard committees will change in the next block
	if beaconHeight%common.EPOCH == 0 {
		swapInstruction, err = CreateSwapAction(shardPendingValidator, shardCommittees, shardID)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
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

	// Process stability tx, create response txs if needed
	stabilityResponseTxs, err := blockgen.buildStabilityResponseTxsAtShardOnly(txsToAdd, privatekey)
	if err != nil {
		return nil, err
	}
	for _, tx := range stabilityResponseTxs {
		txsToAdd = append(txsToAdd, tx)
	}

	// Process stability instructions, create response txs if needed
	stabilityResponseTxs, err = blockgen.buildStabilityResponseTxsFromInstructions(beaconBlocks, privatekey, shardID)
	if err != nil {
		return nil, err
	}
	for _, tx := range stabilityResponseTxs {
		txsToAdd = append(txsToAdd, tx)
	}

	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
	}
	//============Build Header=============
	fmt.Printf("Number of Transaction in blocks %+v \n", len(block.Body.Transactions))
	//Get user key set
	userKeySet := cashec.KeySet{}
	userKeySet.ImportFromPrivateKey(privatekey)
	merkleRoots := Merkle{}.BuildMerkleTreeStore(block.Body.Transactions)
	merkleRoot := merkleRoots[len(merkleRoots)-1]
	prevBlock := blockgen.chain.BestState.Shard[shardID].BestShardBlock
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
	committeeRoot, err := GenerateHashFromStringArray(blockgen.chain.BestState.Shard[shardID].ShardCommittee)
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
		ShardTxRoot:          *block.Body.CalcMerkleRootShard(),
		CrossOutputCoinRoot:  *crossOutputCoinRoot,
		ActionsRoot:          actionsHash,
		CrossShards:          CreateCrossShardByteArray(txsToAdd),
		CommitteeRoot:        committeeRoot,
		PendingValidatorRoot: pendingValidatorRoot,
		BeaconHeight:         beaconHeight,
		BeaconHash:           beaconHash,
		Epoch:                epoch,
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
	byteMap := make([]byte, common.SHARD_NUMBER)
	for _, tx := range txList {
		for _, outCoin := range tx.GetProof().OutputCoins {
			lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
			shardID := common.GetShardIDFromLastByte(lastByte)
			byteMap[common.GetShardIDFromLastByte(shardID)] = 1
		}
	}

	for k, _ := range byteMap {
		if byteMap[k] == 1 {
			crossIDs = append(crossIDs, byte(k))
		}
	}

	return crossIDs
}

/*
	Action From Other Source:
	- bpft protocol: swap
	....
*/
func CreateSwapAction(commitees []string, pendingValidator []string, shardID byte) ([]string, error) {
	_, _, shardSwapedCommittees, shardNewCommittees, err := SwapValidator(pendingValidator, commitees, common.COMMITEES, common.OFFSET)
	if err != nil {
		return nil, err
	}
	swapInstruction := []string{"swap", strings.Join(shardNewCommittees, ","), strings.Join(shardSwapedCommittees, ","), "shard", strconv.Itoa(int(shardID))}
	return swapInstruction, nil
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
	copy(block.ValidatorsIdx, blk.ValidatorsIdx)
	block.ProducerSig = blk.ProducerSig
	block.Header = blk.Header
	block.Instructions = blk.Body.Instructions
	actions := CreateShardActionFromTransaction(blk.Body.Transactions, bcr, blk.Header.ShardID)
	block.Instructions = append(block.Instructions, actions...)
	return &block
}

func (blk *ShardBlock) CreateAllCrossShardBlock() map[byte]*CrossShardBlock {
	allCrossShard := make(map[byte]*CrossShardBlock)
	if common.SHARD_NUMBER == 1 {
		return allCrossShard
	}
	for i := 0; i < common.SHARD_NUMBER; i++ {
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
	copy(crossShard.ValidatorsIdx, block.ValidatorsIdx)
	crossShard.ProducerSig = block.ProducerSig
	crossShard.Header = block.Header
	crossShard.MerklePathShard = merklePathShard
	crossShard.CrossOutputCoin = utxoList
	return crossShard, nil
}
