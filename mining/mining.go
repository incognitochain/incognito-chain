package mining

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/transaction"
	"github.com/ninjadotorg/cash-prototype/wallet"
)

/*type txPrioItem struct {
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
		if tx.GetType() == common.TxActionParamsType {
			actionParamTxs = append(actionParamTxs, tx.(*transaction.ActionParamTx))
		}
	}
	return actionParamTxs
}*/

func getMedians(agentDataPoints []*blockchain.AgentDataPoint) (
	float64, float64, float64,
) {
	agentDataPointsLen := len(agentDataPoints)
	if agentDataPointsLen == 0 {
		return 0, 0, 0
	}
	var sumOfCoins float64 = 0
	var sumOfBonds float64 = 0
	var sumOfTaxs float64 = 0
	for _, dataPoint := range agentDataPoints {
		sumOfCoins += dataPoint.NumOfCoins
		sumOfBonds += dataPoint.NumOfBonds
		sumOfTaxs += dataPoint.Tax
	}
	return float64(sumOfCoins / float64(agentDataPointsLen)), float64(sumOfBonds / float64(agentDataPointsLen)), float64(sumOfTaxs / float64(agentDataPointsLen))
}

func calculateReward(
	agentDataPoints map[string]*blockchain.AgentDataPoint,
	feeMap map[string]uint64,
) map[string]uint64 {
	if len(agentDataPoints) < NUMBER_OF_MAKING_DECISION_AGENTS {
		return map[string]uint64{
			"coins": DEFAULT_COINS + feeMap[common.TxOutCoinType],
			"bonds": DEFAULT_BONDS + feeMap[common.TxOutBondType],
		}
	}

	// group actions by their purpose (ie. issuing or contracting)
	issuingCoinsActions := []*blockchain.AgentDataPoint{}
	contractingCoinsActions := []*blockchain.AgentDataPoint{}
	for _, dataPoint := range agentDataPoints {
		if (dataPoint.NumOfCoins > 0 && dataPoint.NumOfBonds > 0) || (dataPoint.NumOfCoins > 0 && dataPoint.Tax > 0) {
			continue
		}
		if dataPoint.NumOfCoins > 0 {
			issuingCoinsActions = append(issuingCoinsActions, dataPoint)
			continue
		}
		contractingCoinsActions = append(contractingCoinsActions, dataPoint)
	}
	if math.Max(float64(len(issuingCoinsActions)), float64(len(contractingCoinsActions))) < (math.Floor(float64(len(agentDataPoints)/2)) + 1) {
		return map[string]uint64{
			"coins": DEFAULT_COINS + feeMap[common.TxOutCoinType],
			"bonds": DEFAULT_BONDS + feeMap[common.TxOutBondType],
		}
	}

	if len(issuingCoinsActions) == len(contractingCoinsActions) {
		return map[string]uint64{
			"coins": DEFAULT_COINS + feeMap[common.TxOutCoinType],
			"bonds": DEFAULT_BONDS + feeMap[common.TxOutBondType],
		}
	}

	if len(issuingCoinsActions) < len(contractingCoinsActions) {
		_, medianBond, medianTax := getMedians(contractingCoinsActions)
		coins := uint64(math.Floor((100 - medianTax) * 0.01 * float64(feeMap[common.TxOutCoinType])))
		burnedCoins := feeMap[common.TxOutCoinType] - coins
		bonds := uint64(math.Floor(medianBond)) + feeMap[common.TxOutBondType] + burnedCoins
		return map[string]uint64{
			"coins":       coins,
			"bonds":       bonds,
			"burnedCoins": burnedCoins,
		}
	}
	// issuing coins
	medianCoin, _, _ := getMedians(issuingCoinsActions)

	return map[string]uint64{
		"coins": uint64(math.Floor(medianCoin)) + feeMap[common.TxOutCoinType],
		"bonds": feeMap[common.TxOutBondType],
	}
}

func createCoinbaseTx(
	params *blockchain.Params,
	receiverAddr *client.PaymentAddress,
	rewardMap map[string]uint64,
	rt []byte,
) (*transaction.Tx, error) {
	// Create Proof for the joinsplit op
	inputs := make([]*client.JSInput, 2)
	inputs[0] = transaction.CreateRandomJSInput()
	inputs[1] = transaction.CreateRandomJSInput()

	// Get reward
	// TODO(@0xbunyip): implement bonds reward
	fmt.Printf("reward map: %+v\n", rewardMap)
	var reward uint64
	for rewardType, rewardValue := range rewardMap {
		if rewardValue <= 0 {
			continue
		}
		if rewardType == "coins" {
			reward += rewardValue
		}
	}

	// Create new notes: first one is coinbase UTXO, second one has 0 value
	outNote := &client.Note{Value: reward, Apk: receiverAddr.Apk}
	placeHolderOutputNote := &client.Note{Value: 0, Apk: receiverAddr.Apk}

	outputs := []*client.JSOutput{&client.JSOutput{}, &client.JSOutput{}}
	outputs[0].EncKey = receiverAddr.Pkenc
	outputs[0].OutputNote = outNote
	outputs[1].EncKey = receiverAddr.Pkenc
	outputs[1].OutputNote = placeHolderOutputNote

	// Shuffle output notes randomly (if necessary)

	// Generate proof and sign tx
	tx := transaction.NewTxTemplate()
	var coinbaseTxFee uint64 // Zero fee for coinbase tx
	err := tx.BuildNewJSDesc(inputs, outputs, rt, reward, coinbaseTxFee)
	if err != nil {
		return nil, err
	}
	tx, err = transaction.SignTx(tx)
	if err != nil {
		return nil, err
	}
	return tx, err
}

// /**
// // createCoinbaseTx returns a coinbase transaction paying an appropriate subsidy
// // based on the passed block height to the provided address.  When the address
// // is nil, the coinbase transaction will instead be redeemable by anyone.
// */
// func createCoinbaseTx(
// 	params *blockchain.Params,
// 	coinbaseScript []byte,
// 	addr string,
// 	rewardMap map[string]uint64,
// ) (*transaction.Tx, error) {
// 	// Create the script to pay to the provided payment address if one was
// 	// specified.  Otherwise create a script that allows the coinbase to be
// 	// redeemable by anyone.
// 	var pkScript []byte
// 	_ = pkScript

// 	pkScript = []byte(addr) //@todo add public key of the receiver where

// 	//create new tx
// 	tx := &transaction.Tx{
// 		Version: 1,
// 		Type:    common.TxNormalType,
// 		//TxIn:    make([]transaction.TxIn, 0, 2),
// 		//TxOut:   make([]transaction.TxOut, 0, 1),
// 	}
// 	//create outpoint
// 	//outPoint := &transaction.OutPoint{
// 	//	Hash: common.Hash{},
// 	//	Vout: transaction.MaxPrevOutIndex,
// 	//}

// 	//txIn := *transaction.TxIn{}.NewTxIn(outPoint, coinbaseScript)
// 	//tx.AddTxIn(txIn)
// 	//@todo add value of tx out logic
// 	for rewardType, rewardValue := range rewardMap {
// 		if rewardValue <= 0 {
// 			continue
// 		}
// 		txOutTypeMap := map[string]string{
// 			"coins":       common.TxOutCoinType,
// 			"bonds":       common.TxOutBondType,
// 			"burnedCoins": common.TxOutCoinType,
// 		}
// 		if rewardType == "burnedCoins" {
// 			pkScript = []byte(DEFAULT_ADDRESS_FOR_BURNING)
// 		}
// 		_ = txOutTypeMap
// 		//txOut := *transaction.TxOut{}.NewTxOut(rewardValue, pkScript, txOutTypeMap[rewardType])
// 		//tx.AddTxOut(txOut)
// 	}

// 	return tx, nil
// }

// spendTransaction updates the passed view by marking the inputs to the passed
// transaction as spent.  It also adds all outputs in the passed transaction
// which are not provably unspendable as available unspent transaction outputs.
/*func spendTransaction(utxoView *blockchain.UtxoViewpoint, tx *transaction.Tx, height int32) error {
	for _, txIn := range tx.TxIn {
		entry := utxoView.LookupEntry(txIn.PreviousOutPoint)
		if entry != nil {
			entry.Spend()
		}
	}

	utxoView.AddTxOuts(tx, height)
	return nil
}*/

func extractTxsAndComputeInitialFees(txDescs []*TxDesc) (
	[]transaction.Transaction,
	[]*transaction.ActionParamTx,
	map[string]uint64,
) {
	var txs []transaction.Transaction
	var actionParamTxs []*transaction.ActionParamTx
	var feeMap = map[string]uint64{
		fmt.Sprintf(common.TxOutCoinType): 0,
		fmt.Sprintf(common.TxOutBondType): 0,
	}
	for _, txDesc := range txDescs {
		tx := txDesc.Tx
		txs = append(txs, tx)
		txType := tx.GetType()
		if txType == common.TxActionParamsType {
			actionParamTxs = append(actionParamTxs, tx.(*transaction.ActionParamTx))
			continue
		}
		normalTx, _ := tx.(*transaction.Tx)
		feeMap[normalTx.Descs[0].Type] += txDesc.Fee
	}
	return txs, actionParamTxs, feeMap
}

func getLatestAgentDataPoints(
	chain *blockchain.BlockChain,
	actionParamTxs []*transaction.ActionParamTx,
) map[string]*blockchain.AgentDataPoint {
	agentDataPoints := map[string]*blockchain.AgentDataPoint{}
	bestBlock := chain.BestState.BestBlock

	if bestBlock != nil && bestBlock.AgentDataPoints != nil {
		agentDataPoints = bestBlock.AgentDataPoints
	}

	for _, actionParamTx := range actionParamTxs {
		inputAgentID := actionParamTx.Param.AgentID

		_, ok := agentDataPoints[inputAgentID]
		if !ok || actionParamTx.LockTime > agentDataPoints[inputAgentID].LockTime {
			agentDataPoints[inputAgentID] = &blockchain.AgentDataPoint{
				AgentID:          actionParamTx.Param.AgentID,
				AgentSig:         actionParamTx.Param.AgentSig,
				NumOfCoins:       actionParamTx.Param.NumOfCoins,
				NumOfBonds:       actionParamTx.Param.NumOfBonds,
				Tax:              actionParamTx.Param.Tax,
				EligibleAgentIDs: actionParamTx.Param.EligibleAgentIDs,
				LockTime:         actionParamTx.LockTime,
			}
		}
	}

	// in case of not being enough number of agents
	dataPointsLen := len(agentDataPoints)
	if dataPointsLen < NUMBER_OF_MAKING_DECISION_AGENTS {
		return agentDataPoints
	}

	// check add/remove agents by number of votes
	votesForAgents := map[string]int{}
	for _, dataPoint := range agentDataPoints {
		for _, eligibleAgentID := range dataPoint.EligibleAgentIDs {
			if _, ok := votesForAgents[eligibleAgentID]; !ok {
				votesForAgents[eligibleAgentID] = 1
				continue
			}
			votesForAgents[eligibleAgentID] += 1
		}
	}

	for agentID, votes := range votesForAgents {
		if votes < int(math.Floor(float64(dataPointsLen/2))+1) {
			delete(agentDataPoints, agentID)
		}
	}

	return agentDataPoints
}

func (g *BlkTmplGenerator) NewBlockTemplate(payToAddress string, chain *blockchain.BlockChain) (*BlockTemplate, error) {
	// get prevblock
	prevBlock := chain.BestState.BestBlock
	prevBlockHash := chain.BestState.BestBlockHash

	sourceTxns := g.txSource.MiningDescs()

	if len(sourceTxns) == 0 {
		return nil, errors.New("No Tx")
	}

	txs, actionParamTxs, feeMap := extractTxsAndComputeInitialFees(sourceTxns)
	//@todo we need apply sort rules for sourceTxns here

	agentDataPoints := getLatestAgentDataPoints(chain, actionParamTxs)
	rewardMap := calculateReward(agentDataPoints, feeMap)

	_ = []byte("1234567890123456789012") //@todo should be create function create basescript

	receiverKeyset, _ := wallet.Base58CheckDeserialize(payToAddress)

	// dependers := make(map[common.Hash]map[common.Hash]*txPrioItem)

	// blockTxns := make([]transaction.Transaction, 0, len(sourceTxns))
	// blockTxns = append(blockTxns, coinbaseTx)
	blockTxns := txs

	//merkleRoots := blockchain.Merkle{}.BuildMerkleTreeStore(blockTxns)
	//merkleRoot := merkleRoots[len(merkleRoots)-1]

mempoolLoop:
	for _, blockTx := range blockTxns {
		tx, ok := blockTx.(*transaction.Tx)
		if !ok {
			return nil, fmt.Errorf("Transaction in block not recognized")
		}

		//@todo need apply validate tx, logic check all referenced here
		for _, desc := range tx.Descs {
			view, err := g.chain.FetchTxViewPoint(desc.Type)
			_ = view
			_ = err
		}
		//_ = utxos
		//_ = err
		/*
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
		if !tx.ValidateTransaction() {
			continue mempoolLoop
		}

		err := g.txSource.RemoveTx(&blockTx)
		if err != nil {
			Logger.log.Error(err)
		}
	}

	coinbaseTx, err := createCoinbaseTx(&blockchain.Params{}, &receiverKeyset.KeyPair.PublicKey, rewardMap, g.chain.BestState.BestBlock.Header.MerkleRootCommitments.CloneBytes())
	if err != nil {
		return nil, err
	}
	blockTxns = append([]transaction.Transaction{coinbaseTx}, blockTxns...)

	// TODO PoW
	//time.Sleep(time.Second * 15)

	// Store commitments and nullifiers in database
	var descType string
	commitments := [][]byte{}
	nullifiers := [][]byte{}
	for _, blockTx := range blockTxns {
		if blockTx.GetType() == common.TxNormalType {
			tx, ok := blockTx.(*transaction.Tx)
			if ok == false {
				return nil, fmt.Errorf("Transaction not recognized to store in database")
			}

			for _, desc := range tx.Descs {
				for _, cm := range desc.Commitments {
					commitments = append(commitments, cm)
				}

				for _, nf := range desc.Nullifiers {
					nullifiers = append(nullifiers, nf)
				}
				descType = desc.Type
			}
		}
	}
	// TODO(@0xsirrush): check if cm and nf should be saved here (when generate block template)
	// or when UpdateBestState
	g.chain.StoreCommitmentsFromListCommitment(commitments, descType)
	g.chain.StoreNullifiersFromListNullifier(nullifiers, descType)

	block := blockchain.Block{}
	block.Header = blockchain.BlockHeader{
		Version:       1,
		PrevBlockHash: *prevBlockHash,
		//MerkleRoot:            *merkleRoot,
		MerkleRootCommitments: common.Hash{},
		Timestamp:             time.Now().Unix(),
		Difficulty:            0, //@todo should be create Difficulty logic
		Nonce:                 0, //@todo should be create Nonce logic
	}
	for _, tx := range blockTxns {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
	}

	// Add new commitments to merkle tree and save the root
	newTree := g.chain.BestState.CmTree.MakeCopy()
	fmt.Printf("[newBlockTemplate] old tree rt: %x\n", newTree.GetRoot(common.IncMerkleTreeHeight))
	blockchain.UpdateMerkleTreeForBlock(newTree, &block)
	rt := newTree.GetRoot(common.IncMerkleTreeHeight)
	fmt.Printf("[newBlockTemplate] updated tree rt: %x\n", rt)
	copy(block.Header.MerkleRootCommitments[:], rt)

	//update the latest AgentDataPoints to block
	block.AgentDataPoints = agentDataPoints
	// Set height
	block.Height = prevBlock.Height + 1

	blockTemp := &BlockTemplate{
		Block: &block,
	}
	return blockTemp, nil
}

type BlkTmplGenerator struct {
	txSource    TxSource
	chain       *blockchain.BlockChain
	chainParams *blockchain.Params
	policy      *Policy
}

type BlockTemplate struct {
	Block *blockchain.Block

	// Fees []float64
}

/**
// TxSource represents a source of transactions to consider for inclusion in
// new blocks.
//
// The interface contract requires that all of these methods are safe for
// concurrent access with respect to the source.
Note mempool(tx pool) is a TxSource
*/
type TxSource interface {
	// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
	LastUpdated() time.Time

	// MiningDescs returns a slice of mining descriptors for all the
	// transactions in the source pool.
	MiningDescs() []*TxDesc

	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	HaveTransaction(hash *common.Hash) bool

	// RemoveTx remove tx from tx resource
	RemoveTx(tx *transaction.Transaction) error
}

/**
NewBlkTmplGenerator - create a block template generator from mempool package
*/
func NewBlkTmplGeneratorByMempool(mempool TxSource, chain *blockchain.BlockChain) *BlkTmplGenerator {
	return &BlkTmplGenerator{
		txSource: mempool,
		chain:    chain,
	}
}
