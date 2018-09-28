package ppos

import (
	"errors"
	"fmt"
	"time"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

// TODO: create block template (move block template from mining to here)

// func getLatestAgentDataPoints(
// 	chain *blockchain.BlockChain,
// 	actionParamTxs []*transaction.ActionParamTx,
// ) map[string]*blockchain.AgentDataPoint {
// 	agentDataPoints := map[string]*blockchain.AgentDataPoint{}
// 	bestBlock := chain.BestState.BestBlock

// 	if bestBlock != nil && bestBlock.AgentDataPoints != nil {
// 		agentDataPoints = bestBlock.AgentDataPoints
// 	}

// 	for _, actionParamTx := range actionParamTxs {
// 		inputAgentID := actionParamTx.Param.AgentID

// 		_, ok := agentDataPoints[inputAgentID]
// 		if !ok || actionParamTx.LockTime > agentDataPoints[inputAgentID].LockTime {
// 			agentDataPoints[inputAgentID] = &blockchain.AgentDataPoint{
// 				AgentID:          actionParamTx.Param.AgentID,
// 				AgentSig:         actionParamTx.Param.AgentSig,
// 				NumOfCoins:       actionParamTx.Param.NumOfCoins,
// 				NumOfBonds:       actionParamTx.Param.NumOfBonds,
// 				Tax:              actionParamTx.Param.Tax,
// 				EligibleAgentIDs: actionParamTx.Param.EligibleAgentIDs,
// 				LockTime:         actionParamTx.LockTime,
// 			}
// 		}
// 	}

// 	// in case of not being enough number of agents
// 	dataPointsLen := len(agentDataPoints)
// 	if dataPointsLen < NUMBER_OF_MAKING_DECISION_AGENTS {
// 		return agentDataPoints
// 	}

// 	// check add/remove agents by number of votes
// 	votesForAgents := map[string]int{}
// 	for _, dataPoint := range agentDataPoints {
// 		for _, eligibleAgentID := range dataPoint.EligibleAgentIDs {
// 			if _, ok := votesForAgents[eligibleAgentID]; !ok {
// 				votesForAgents[eligibleAgentID] = 1
// 				continue
// 			}
// 			votesForAgents[eligibleAgentID] += 1
// 		}
// 	}

// 	for agentID, votes := range votesForAgents {
// 		if votes < int(math.Floor(float64(dataPointsLen/2))+1) {
// 			delete(agentDataPoints, agentID)
// 		}
// 	}

// 	return agentDataPoints
// }

func (g *BlkTmplGenerator) NewBlockTemplate(payToAddress client.PaymentAddress, chain *blockchain.BlockChain, chainID byte) (*BlockTemplate, error) {

	prevBlock := chain.BestState[chainID]
	prevBlockHash := chain.BestState[chainID].BestBlock.Hash()
	sourceTxns := g.txSource.MiningDescs()
	if len(sourceTxns) < MIN_TXs {
		// if len of sourceTxns < MIN_TXs -> wait for more transactions
		Logger.log.Info("not enough transactions. Wait for more...")
		fmt.Println(sourceTxns)
		<-time.Tick(MAX_BLOCK_WAIT_TIME * time.Second)
		sourceTxns = g.txSource.MiningDescs()
		if len(sourceTxns) == 0 {
			return nil, errors.New("No Tx")
		}
	}

	txs, _, feeMap := extractTxsAndComputeInitialFees(sourceTxns)
	//@todo we need apply sort rules for sourceTxns here

	// agentDataPoints := getLatestAgentDataPoints(chain, actionParamTxs)
	// rewardMap := calculateReward(agentDataPoints, feeMap)

	// coinbaseScript := []byte("1234567890123456789012") //@todo should be create function create basescript

	// coinbaseTx, err := createCoinbaseTx(&blockchain.Params{}, coinbaseScript, payToAddress, rewardMap)
	// if err != nil {
	// 	return nil, err
	// }

	// dependers := make(map[common.Hash]map[common.Hash]*txPrioItem)

	// blockTxns := make([]transaction.Transaction, 0, len(sourceTxns))
	// blockTxns = append(blockTxns, coinbaseTx)

	// blockTxns := append([]transaction.Transaction{coinbaseTx}, txs...)
	//receiverKeyset, _ := wallet.Base58CheckDeserialize(payToAddress)
	//_ = receiverKeyset

	var txsToAdd []transaction.Transaction

	var txToRemove []transaction.Transaction
	// mempoolLoop:
	for _, tx := range txs {
		// tx, ok := txDesc.(*transaction.Tx)
		// if !ok {
		// 	return nil, fmt.Errorf("Transaction in block not recognized")
		// }

		//@todo need apply validate tx, logic check all referenced here
		// call function spendTransaction to mark utxo

		txChainID, _ := g.GetTxSenderChain(tx.GetSenderAddrLastByte())
		if txChainID != chainID {
			continue
		}
		/*for _, desc := range tx.Descs {
			view, err := g.chain.FetchTxViewPoint(desc.Type)
			_ = view
			_ = err
		}*/
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
			txToRemove = append(txToRemove, transaction.Transaction(tx))
		}
		txsToAdd = append(txsToAdd, tx)
		if len(txsToAdd) == MAX_TXs_IN_BLOCK {
			break
		}
		// g.txSource.Clear()
	}

	for _, tx := range txToRemove {
		g.txSource.RemoveTx(tx)
	}

	// check len of txs in block
	if len(txsToAdd) == 0 {
		return nil, errors.New("no transaction available for this chain")
	}

	rt := g.chain.BestState[chainID].BestBlock.Header.MerkleRootCommitments.CloneBytes()
	coinbaseTx, err := createCoinbaseTx(
		&blockchain.Params{},
		&payToAddress,
		rt,
		chainID,
		feeMap,
	)
	if err != nil {
		return nil, err
	}
	// the 1st tx will be coinbaseTx
	txsToAdd = append([]transaction.Transaction{coinbaseTx}, txsToAdd...)

	merkleRoots := blockchain.Merkle{}.BuildMerkleTreeStore(txsToAdd)
	merkleRoot := merkleRoots[len(merkleRoots)-1]
	// TODO PoW
	//time.Sleep(time.Second * 15)

	// Store commitments and nullifiers in database
	var descType string
	commitments := [][]byte{}
	nullifiers := [][]byte{}
	for _, blockTx := range txsToAdd {
		if blockTx.GetType() == common.TxNormalType {
			tx, ok := blockTx.(*transaction.Tx)
			if !ok {
				Logger.log.Error("Transaction not recognized to store in database")
				continue
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
	g.chain.StoreCommitmentsFromListCommitment(commitments, descType, chainID)
	g.chain.StoreNullifiersFromListNullifier(nullifiers, descType, chainID)

	block := blockchain.Block{}
	block.Header = blockchain.BlockHeader{
		Version:               1,
		PrevBlockHash:         *prevBlockHash,
		MerkleRoot:            *merkleRoot,
		MerkleRootCommitments: common.Hash{},
		Timestamp:             time.Now().Unix(),
		Difficulty:            0, //@todo should be create Difficulty logic
		Nonce:                 0, //@todo should be create Nonce logic
		BlockCommitteeSigs:    []string{},
		Committee:             []string{},
		ChainID:               chainID,
	}
	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
	}

	// Add new commitments to merkle tree and save the root
	newTree := g.chain.BestState[chainID].CmTree.MakeCopy()
	fmt.Printf("[newBlockTemplate] old tree rt: %x\n", newTree.GetRoot(common.IncMerkleTreeHeight))
	blockchain.UpdateMerkleTreeForBlock(newTree, &block)
	rt = newTree.GetRoot(common.IncMerkleTreeHeight)
	fmt.Printf("[newBlockTemplate] updated tree rt: %x\n", rt)
	copy(block.Header.MerkleRootCommitments[:], rt)

	//update the latest AgentDataPoints to block
	// block.AgentDataPoints = agentDataPoints
	// Set height
	block.Height = prevBlock.Height + 1

	blockTemp := &BlockTemplate{
		Block: &block,
	}
	return blockTemp, nil
}

func (g *BlkTmplGenerator) GetTxSenderChain(senderLastByte byte) (byte, error) {
	modResult := senderLastByte % 100
	for index := byte(0); index < 5; index++ {
		if (modResult-index)%5 == 0 {
			// result := byte((modResult - index) / 5)
			return byte((modResult - index) / 5), nil
		}
	}
	return 0, errors.New("can't get sender's chainID")
}

type BlkTmplGenerator struct {
	txSource    TxSource
	chain       *blockchain.BlockChain
	chainParams *blockchain.Params
	// policy      *Policy
}

type BlockTemplate struct {
	Block *blockchain.Block

	// Fees []float64
}

// TxSource represents a source of transactions to consider for inclusion in
// new blocks.
//
// The interface contract requires that all of these methods are safe for
// concurrent access with respect to the source.
type TxSource interface {
	// LastUpdated returns the last time a transaction was added to or
	// removed from the source pool.
	LastUpdated() time.Time

	// MiningDescs returns a slice of mining descriptors for all the
	// transactions in the source pool.
	MiningDescs() []*transaction.TxDesc

	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	HaveTransaction(hash *common.Hash) bool

	// RemoveTx remove tx from tx resource
	RemoveTx(tx transaction.Transaction) error
}

func NewBlkTmplGenerator(txSource TxSource, chain *blockchain.BlockChain) *BlkTmplGenerator {
	return &BlkTmplGenerator{
		txSource: txSource,
		chain:    chain,
	}
}

func extractTxsAndComputeInitialFees(txDescs []*transaction.TxDesc) (
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

func createCoinbaseTx(
	params *blockchain.Params,
	receiverAddr *client.PaymentAddress,
	rt []byte,
	chainID byte,
	feeMap map[string]uint64,
) (*transaction.Tx, error) {
	// Create Proof for the joinsplit op
	inputs := make([]*client.JSInput, 2)
	inputs[0] = transaction.CreateRandomJSInput()
	inputs[1] = transaction.CreateRandomJSInput()

	// Get reward
	// TODO(@0xbunyip): implement bonds reward
	var reward uint64 = DEFAULT_MINING_REWARD + feeMap[common.TxOutCoinType] // TODO: probably will need compute reward based on block height

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
	rtMap := map[byte][]byte{chainID: rt}
	inputMap := map[byte][]*client.JSInput{chainID: inputs}
	err := tx.BuildNewJSDesc(inputMap, outputs, rtMap, reward, coinbaseTxFee)
	if err != nil {
		return nil, err
	}
	tx.AddressLastByte = receiverAddr.Apk[len(receiverAddr.Apk)-1]
	tx, err = transaction.SignTx(tx)
	if err != nil {
		return nil, err
	}
	return tx, err
}
