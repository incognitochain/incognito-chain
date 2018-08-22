package mining

import (
	"time"
	"math"
	// "fmt"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

type txPrioItem struct {
	tx       transaction.Transaction
	fee      int64
	priority float64
	feePerKB int64

	dependsOn map[common.Hash]struct{}
}


func filterActionParamsTxs(block *blockchain.Block) []*transaction.ActionParamTx {
	allTxs := block.Transactions
	var actionParamTxs []*transaction.ActionParamTx
	for _, tx := range allTxs {
		if tx.GetType() == ACTION_PARAMS_TRANSACTION_TYPE {
			actionParamTxs = append(actionParamTxs, tx.(*transaction.ActionParamTx))
		}
	}
	return actionParamTxs
}


func getRecentActionParamsTxs(numOfBlocks int, chain *blockchain.BlockChain) []*transaction.ActionParamTx {
	if chain == nil || chain.BestBlock == nil {
		return []*transaction.ActionParamTx{}
	}
	actionParamTxs := []*transaction.ActionParamTx{}
	bestBlock := chain.BestBlock
	actionParamTxsInBestBlock := filterActionParamsTxs(bestBlock)
	actionParamTxs = append(actionParamTxs, actionParamTxsInBestBlock...)
	prevBlockHash := bestBlock.Header.PrevBlockHash

	for i := 0; i < numOfBlocks - 1; i++ {
		block, ok := chain.Blocks[&prevBlockHash]
		if !ok {
			return actionParamTxs
		}
		actionParamTxsInBlock := filterActionParamsTxs(block)
		actionParamTxs = append(actionParamTxs, actionParamTxsInBlock...)
		prevBlockHash = block.Header.PrevBlockHash
	}
	return actionParamTxs
}


func getMedians(actionParamTxs []*transaction.ActionParamTx) (float64, float64, float64) {
	sumOfCoins := 0
	sumOfBonds := 0
	var sumOfTaxs float64 = 0
	for _, tx := range actionParamTxs {
		sumOfCoins += tx.Param.NumOfIssuingCoins
		sumOfBonds += tx.Param.NumOfIssuingBonds
		sumOfTaxs += tx.Param.Tax
	}
	return float64(sumOfCoins / len(actionParamTxs)), float64(sumOfBonds / len(actionParamTxs)), float64(sumOfTaxs / float64(len(actionParamTxs)))
}


func calculateReward(actionParamTxs []*transaction.ActionParamTx, txFees []float64) (map[string]float64) {
	latestTxsByAgentId := map[string]*transaction.ActionParamTx{}
	for _, tx := range actionParamTxs {

		agentId := tx.Param.AgentID
		existingTx, ok := latestTxsByAgentId[agentId]
		if !ok {
			latestTxsByAgentId[agentId] = tx
			continue
		}
		if existingTx.LockTime < tx.LockTime {
			latestTxsByAgentId[agentId] = tx
		}
	}
	if len(latestTxsByAgentId) < NUMBER_OF_MAKING_DECISION_AGENTS {
		return map[string]float64{
			"coins": DEFAULT_COINS,
			"bonds": DEFAULT_BONDS,
		}
	}

	// get group of action params tx that issuing coins
	issuingCoinsActions := []*transaction.ActionParamTx{}
	contractingCoinsActions := []*transaction.ActionParamTx{}
	for _, tx := range latestTxsByAgentId {
		if (tx.Param.NumOfIssuingCoins > 0 && tx.Param.NumOfIssuingBonds > 0) || (tx.Param.NumOfIssuingCoins > 0 && tx.Param.Tax > 0) {
			continue
		}
		if tx.Param.NumOfIssuingCoins > 0 {
			issuingCoinsActions = append(issuingCoinsActions, tx)
		} else {
			contractingCoinsActions = append(contractingCoinsActions, tx)
		}
	}
	if math.Max(float64(len(issuingCoinsActions)), float64(len(contractingCoinsActions))) < (math.Floor(float64(len(latestTxsByAgentId) / 2)) + 1) {
		return map[string]float64{
			"coins": DEFAULT_COINS,
			"bonds": DEFAULT_BONDS,
		}
	}

	if len(issuingCoinsActions) == len(contractingCoinsActions) {
		return map[string]float64{
			"coins": DEFAULT_COINS,
			"bonds": DEFAULT_BONDS,
		}
	}
	if len(issuingCoinsActions) < len(contractingCoinsActions) {
		_, medianBond, medianTax := getMedians(contractingCoinsActions)
		var coins float64
		coins = 0
		for _, fee := range txFees {
			coins += (100 - medianTax) * 0.01 * fee
		}
		// TODO: remember that there are 2 type of tx out: coin and bond -> recalculate by type
		return map[string]float64{
			"coins": coins,
			"bonds": medianBond,
		}
	}
	// issuing coins
	medianCoin, _, _ := getMedians(contractingCoinsActions)
	return map[string]float64{
		"coins": medianCoin,
		"bonds": 0,
	}
}


// createCoinbaseTx returns a coinbase transaction paying an appropriate subsidy
// based on the passed block height to the provided address.  When the address
// is nil, the coinbase transaction will instead be redeemable by anyone.

func createCoinbaseTx(
	params *blockchain.Params,
	coinbaseScript []byte,
	addr string,
	rewardMap map[string]float64,
) (*transaction.Tx, error) {
	// Create the script to pay to the provided payment address if one was
	// specified.  Otherwise create a script that allows the coinbase to be
	// redeemable by anyone.
	var pkScript []byte

	pkScript = []byte(addr) //@todo add public key of the receiver where

	//create new tx
	tx := &transaction.Tx{
		Version: 1,
		TxIn:    make([]transaction.TxIn, 0, 2),
		TxOut:   make([]transaction.TxOut, 0, 1),
	}
	//create outpoint
	outPoint := &transaction.OutPoint{
		Hash: common.Hash{},
		Vout: 1,
	}

	txIn := *transaction.TxIn{}.NewTxIn(outPoint, coinbaseScript)
	tx.AddTxIn(txIn)
	//@todo add value of tx out logic
	for _, rewardValue := range rewardMap {
		if rewardValue > 0 {
			// TODO: add reward type to txOut
			txOut := *transaction.TxOut{}.NewTxOut(rewardValue, pkScript)
			tx.AddTxOut(txOut)
		}
	}

	return tx, nil
}

func (g *BlkTmplGenerator) NewBlockTemplate(payToAddress string, chain *blockchain.BlockChain) (*BlockTemplate, error) {

	prevBlockHash := chain.BestBlock.Hash()
	sourceTxns := g.txSource
	//@todo we need apply sort rules for sourceTxns here


	// TODO: need to compute real txFees from transactions
	actionParamTxs := getRecentActionParamsTxs(NUMBER_OF_LAST_BLOCKS, chain)
	txFees := make([]float64, 0, 1)
	rewardMap := calculateReward(actionParamTxs, txFees)

	coinbaseScript := []byte("1234567890123456789012") //@todo should be create function create basescript

	coinbaseTx, err := createCoinbaseTx(&blockchain.Params{}, coinbaseScript, payToAddress, rewardMap)
	if err != nil {
		return nil, err
	}

	// dependers := make(map[common.Hash]map[common.Hash]*txPrioItem)

	blockTxns := make([]transaction.Transaction, 0, len(sourceTxns))
	blockTxns = append(blockTxns, coinbaseTx)

	merkleRoots := blockchain.Merkle{}.BuildMerkleTreeStore(blockTxns)
	merkleRoot := merkleRoots[len(merkleRoots)-1]

	// txFees := make([]int64, 0, 0)

mempoolLoop:
	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx
		//@todo need apply validate tx, logic check all referenced here

		/*utxos, err := g.chain.FetchUtxoView(&tx)
		if err != nil {
			fmt.Print("Unable to fetch utxo view for tx %s: %v",
				tx.Hash(), err)
			continue
		}
		prioItem := &txPrioItem{tx: tx}
		for _, txIn := range tx.TxIn {
			originHash := &txIn.PreviousOutPoint.Hash
			entry := utxos.LookupEntry(txIn.PreviousOutPoint)
			if entry == nil || entry.IsSpent() {
				if !TxPool.HaveTx(originHash) {
					fmt.Print("Skipping tx %s because it "+
						"references unspent output %s "+
						"which is not available",
						tx.Hash(), txIn.PreviousOutPoint)
					continue mempoolLoop
				}

				// The transaction is referencing another
				// transaction in the source pool, so setup an
				// ordering dependency.
				deps, exists := dependers[*originHash]
				if !exists {
					deps = make(map[common.Hash]*txPrioItem)
					dependers[*originHash] = deps
				}
				deps[*prioItem.tx.Hash()] = prioItem
				if prioItem.dependsOn == nil {
					prioItem.dependsOn = make(
						map[common.Hash]struct{})
				}
				prioItem.dependsOn[*originHash] = struct{}{}

				// Skip the check below. We already know the
				// referenced transaction is available.
				continue
			}
		}*/
		if tx.ValidateTransaction() {
			continue mempoolLoop
		}
	}

	// txFees := make([]int64, 0, 1)

	var msgBlock blockchain.Block
	msgBlock.Header = blockchain.BlockHeader{
		Version:       1,
		PrevBlockHash: *prevBlockHash,
		MerkleRoot:    *merkleRoot,
		Timestamp:     time.Now(),
		Difficulty:    0, //@todo should be create Difficulty logic
		Nonce:         0, //@todo should be create Nonce logic
	}
	for _, tx := range blockTxns {
		if err := msgBlock.AddTransaction(tx); err != nil {
			return nil, err
		}
	}

	msgBlock.BlockHash = prevBlockHash

	return &BlockTemplate{
		Block: &msgBlock,
		Fees:  txFees, // TODO: need Fees here?????
	}, nil

}

func NewBlkTmplGenerator(txSource []*TxDesc, chain *blockchain.BlockChain) *BlkTmplGenerator {
	return &BlkTmplGenerator{
		txSource: txSource,
		chain:    chain,
	}
}
