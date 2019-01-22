package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

func (blockgen *BlkTmplGenerator) NewBlockShard(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, shardID byte) (*ShardBlock, error) {
	//============Build body=============
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
	crossOutputCoin := blockgen.getCrossOutputCoin(shardID)
	instructions := CreateShardActionFromOthers()
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
	//============Build Header=============
	//Get user key set
	userKeySet := cashec.KeySet{}
	userKeySet.ImportFromPrivateKey(privatekey)
	merkleRoots := Merkle{}.BuildMerkleTreeStore(block.Body.Transactions)
	merkleRoot := merkleRoots[len(merkleRoots)-1]
	fmt.Println()
	test, _ := json.Marshal(block.Body.Transactions[0])
	fmt.Println(len(block.Body.Transactions), string(test))
	fmt.Println()
	prevBlock := blockgen.chain.BestState.Shard[shardID].BestShardBlock
	prevBlockHash := blockgen.chain.BestState.Shard[shardID].BestShardBlock.Hash()
	crossOutputCoinRoot := &common.Hash{}
	if len(block.Body.CrossOutputCoin) != 0 {
		crossOutputCoinRoot, err = CreateMerkleCrossOutputCoin(block.Body.CrossOutputCoin)
	}
	if err != nil {
		return nil, err
	}
	actions := CreateShardActionFromTransaction(block.Body.Transactions)
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
	pendingValidatorRoot, err := GenerateHashFromStringArray(blockgen.chain.BestState.Shard[shardID].ShardPendingValidator)
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
		ShardTxRoot:          CreateMerkleRootShard(block.Body.Transactions),
		CrossOutputCoinRoot:  *crossOutputCoinRoot,
		ActionsRoot:          actionsHash,
		CrossShards:          CreateCrossShardByteArray(txsToAdd),
		CommitteeRoot:        committeeRoot,
		PendingValidatorRoot: pendingValidatorRoot,
		BeaconHeight:         blockgen.chain.BestState.Beacon.BeaconHeight,
		BeaconHash:           blockgen.chain.BestState.Beacon.BestBlockHash,
		Epoch:                blockgen.chain.BestState.Beacon.BeaconEpoch,
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

func (blockgen *BlkTmplGenerator) getCrossOutputCoin(shardID byte) []CrossOutputCoin {
	res := []CrossOutputCoin{}
	// get cross shard block
	bestShardHeight := blockgen.chain.BestState.Beacon.BestShardHeight
	crossBlock := blockgen.crossShardPool.GetBlock(bestShardHeight)
	// build CrossOutputCoin
	// TODO: Verify with cross map
	// TODO: Make sure cross shard pool begin with most recent proccessed block
	/*
		1. Get Previous most recent proccess cross shard block
		2. Get beacon height of previous shard block
		3. Search from preBeaconHeight to currentBeaconHeight for cross shard via cross shard byte
		4. Detect in pool
		5. if miss then stop or sync block
		6. Update new most recent proccess cross shard block
	*/
	shardCrossBlock := crossBlock[shardID]
	for _, blk := range shardCrossBlock {
		temp, err := blockgen.chain.config.DataBase.FetchBeaconCommitteeByHeight(blk.Header.BeaconHeight)
		if err != nil {
			break
		}
		shardCommittee := make(map[byte][]string)
		json.Unmarshal(temp, &shardCommittee)
		err = blk.VerifyCrossShardBlock(shardCommittee[shardID])
		if err != nil {
			break
		}
		outputCoin := CrossOutputCoin{
			OutputCoin: blk.CrossOutputCoin,
			ShardID:    shardID,
			BlockHash:  *blk.Hash(),
		}
		res = append(res, outputCoin)
	}
	return res
}

func CreateCrossShardByteArray(txList []metadata.Transaction) (crossIDs []byte) {
	byteMap := make([]byte, TestNetParams.ShardsNum)
	for _, tx := range txList {
		for _, outCoin := range tx.GetProof().OutputCoins {
			lastByte := outCoin.CoinDetails.GetPubKeyLastByte()
			byteMap[lastByte] = 1
		}
	}

	for _, v := range byteMap {
		if byteMap[v] == 1 {
			crossIDs = append(crossIDs, v)
		}
	}

	return crossIDs
}

/*
	Action From Other Source:
	- bpft protocol: swap
	....
*/
func CreateShardActionFromOthers(v ...interface{}) (actions [][]string) {
	//TODO: swap action
	return
}

/*
	Action Generate From Transaction:
	- Stake
	- Stable param: set, del,...
*/
func CreateShardActionFromTransaction(transactions []metadata.Transaction) (actions [][]string) {
	// Generate stake action
	stakeShardPubKey := []string{}
	stakeBeaconPubKey := []string{}
	for _, tx := range transactions {
		tempTx, ok := tx.(*transaction.Tx)
		if !ok {
			panic("Can't create block")
		}
		_ = tempTx
		switch tx.GetMetadataType() {
		// case metadata.BuyFromGOVRequestMeta:
		}

		// shardStaker, beaconStaker, isStake := tempTx.GetStakerFromTransaction()
		// if isStake {
		// 	if strings.Compare(shardStaker, common.EmptyString) != 0 {
		// 		stakeShardPubKey = append(stakeShardPubKey, shardStaker)
		// 	}
		// 	if strings.Compare(beaconStaker, common.EmptyString) != 0 {
		// 		stakeBeaconPubKey = append(stakeBeaconPubKey, beaconStaker)
		// 	}
		// }
	}
	if !reflect.DeepEqual(stakeShardPubKey, []string{}) {
		action := []string{"stake", strings.Join(stakeShardPubKey, ","), "shard"}
		actions = append(actions, action)
	}
	if !reflect.DeepEqual(stakeBeaconPubKey, []string{}) {
		action := []string{"stake", strings.Join(stakeBeaconPubKey, ","), "beacon"}
		actions = append(actions, action)
	}
	//TODO: stable param
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

	// validate tx and calculate total fee
	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx
		txShardID, _ := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
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

func (blockgen *ShardBlock) CreateShardToBeaconBlock() ShardToBeaconBlock {
	block := ShardToBeaconBlock{}
	block.AggregatedSig = blockgen.AggregatedSig
	copy(block.ValidatorsIdx, blockgen.ValidatorsIdx)
	block.ProducerSig = blockgen.ProducerSig
	block.Header = blockgen.Header
	block.Instructions = blockgen.Body.Instructions
	actions := CreateShardActionFromTransaction(blockgen.Body.Transactions)
	block.Instructions = append(block.Instructions, actions...)
	return block
}

func (blk *ShardBlock) CreateAllCrossShardBlock() map[byte]*CrossShardBlock {
	allCrossShard := make(map[byte]*CrossShardBlock)
	for i := 0; i < TestNetParams.ShardsNum; i++ {
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
