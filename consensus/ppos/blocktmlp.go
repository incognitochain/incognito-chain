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

func (g *BlkTmplGenerator) NewBlockTemplate(payToAddress client.PaymentAddress, chain *blockchain.BlockChain, chainID byte) (*BlockTemplate, error) {

	prevBlock := chain.BestState[chainID]
	prevBlockHash := chain.BestState[chainID].BestBlock.Hash()
	sourceTxns := g.txSource.MiningDescs()

	var txsToAdd []transaction.Transaction
	var txToRemove []transaction.Transaction
	var feeMap map[string]uint64
	var txs []transaction.Transaction

	if len(sourceTxns) < common.MIN_TXs {
		// if len of sourceTxns < MIN_TXs -> wait for more transactions
		Logger.log.Info("not enough transactions. Wait for more...")
		fmt.Println(sourceTxns)
		<-time.Tick(common.MIN_BLOCK_WAIT_TIME * time.Second)
		sourceTxns = g.txSource.MiningDescs()
		if len(sourceTxns) == 0 {
			<-time.Tick(common.MAX_BLOCK_WAIT_TIME * time.Second)
			sourceTxns = g.txSource.MiningDescs()
			if len(sourceTxns) == 0 {
				// return nil, errors.New("No Tx")
				Logger.log.Info("Creating empty block...")
				goto concludeBlock
			}
		}
	}

	txs, _, feeMap = extractTxsAndComputeInitialFees(sourceTxns)

	// mempoolLoop:
	for _, tx := range txs {
		// tx, ok := txDesc.(*transaction.Tx)
		// if !ok {
		// 	return nil, fmt.Errorf("Transaction in block not recognized")
		// }

		//@todo need apply validate tx, logic check all referenced here
		// call function spendTransaction to mark utxo

		txChainID, _ := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
		if txChainID != chainID {
			continue
		}
		if !tx.ValidateTransaction() {
			txToRemove = append(txToRemove, transaction.Transaction(tx))
		}
		txsToAdd = append(txsToAdd, tx)
		if len(txsToAdd) == common.MAX_TXs_IN_BLOCK {
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

concludeBlock:
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
		// BlockCommitteeSigs:    []string{},
		// Committee:             []string{},
		CommitteeSigs: make(map[string]string),
		ChainID:       chainID,
	}
	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
	}

	// Add new commitments to merkle tree and save the root
	newTree := g.chain.BestState[chainID].CmTree.MakeCopy()
	fmt.Printf("[newBlockTemplate] old tree rt: %x\n", newTree.GetRoot(common.IncMerkleTreeHeight))
	g.chain.UpdateMerkleTreeForBlock(newTree, &block)
	rt = newTree.GetRoot(common.IncMerkleTreeHeight)
	fmt.Printf("[newBlockTemplate] updated tree rt: %x\n", rt)
	copy(block.Header.MerkleRootCommitments[:], rt)

	for _, tempBlockTx := range block.Transactions {
		if tempBlockTx.GetType() == common.TxNormalType {
			tx, ok := tempBlockTx.(*transaction.Tx)
			if ok == false {
				Logger.log.Errorf("Transaction in block not valid")
			}

			for _, desc := range tx.Descs {
				for _, cm := range desc.Commitments {
					Logger.log.Infof("%x", cm[:])
				}
			}
		}
	}

	//update the latest AgentDataPoints to block
	// block.AgentDataPoints = agentDataPoints
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
	// policy      *Policy
}

type BlockTemplate struct {
	Block *blockchain.Block
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
	inputs[0] = transaction.CreateRandomJSInput(nil)
	inputs[1] = transaction.CreateRandomJSInput(inputs[0].Key)
	dummyAddress := client.GenPaymentAddress(*inputs[0].Key)

	// Get reward
	// TODO(@0xbunyip): implement bonds reward
	var reward uint64 = common.DEFAULT_MINING_REWARD + feeMap[common.TxOutCoinType] // TODO: probably will need compute reward based on block height

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
	tx := transaction.CreateEmptyTx()
	tx.AddressLastByte = dummyAddress.Apk[len(dummyAddress.Apk)-1]
	var coinbaseTxFee uint64 // Zero fee for coinbase tx
	rtMap := map[byte][]byte{chainID: rt}
	inputMap := map[byte][]*client.JSInput{chainID: inputs}
	err := tx.BuildNewJSDesc(inputMap, outputs, rtMap, reward, coinbaseTxFee)
	if err != nil {
		return nil, err
	}
	tx, err = transaction.SignTx(tx)
	if err != nil {
		return nil, err
	}
	return tx, err
}
