package mining

import (
	"time"

	"github.com/internet-cash/prototype/blockchain"
	"github.com/internet-cash/prototype/common"
	"github.com/internet-cash/prototype/transaction"
)

// createCoinbaseTx returns a coinbase transaction paying an appropriate subsidy
// based on the passed block height to the provided address.  When the address
// is nil, the coinbase transaction will instead be redeemable by anyone.

func createCoinbaseTx(params *blockchain.Params, coinbaseScript []byte, addr string) (*transaction.Tx, error) {
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
	txOut := *transaction.TxOut{}.NewTxOut(float64(1.5), pkScript)
	tx.AddTxOut(txOut)

	return tx, nil
}

func (g *BlkTmplGenerator) NewBlockTemplate(payToAddress string, chain *blockchain.BlockChain) (*BlockTemplate, error) {

	prevBlockHash := chain.BestBlock.Hash()
	sourceTxns := g.txSource
	//@todo we need apply sort rules for sourceTxns here

	coinbaseScript := []byte("1234567890123456789012") //@todo should be create function create basescript

	coinbaseTx, err := createCoinbaseTx(&blockchain.Params{}, coinbaseScript, payToAddress)
	if err != nil {
		return nil, err
	}

	blockTxns := make([]*transaction.Tx, 0, len(sourceTxns))
	blockTxns = append(blockTxns, coinbaseTx)

	merkleRoots := blockchain.Merkle{}.BuildMerkleTreeStore(blockTxns)
	merkleRoot := merkleRoots[len(merkleRoots)-1]
mempoolLoop:
	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx
		//@todo need apply validate tx, logic check all referenced here
		if tx.ValidateTransaction() {
			continue mempoolLoop
		}
	}
	txFees := make([]int64, 0, 1)
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
		if err := msgBlock.AddTransaction(*tx); err != nil {
			return nil, err
		}
	}

	msgBlock.BlockHash = prevBlockHash

	return &BlockTemplate{
		Block: &msgBlock,
		Fees:  txFees,
	}, nil

}

func NewBlkTmplGenerator(txSource []*TxDesc, chain *blockchain.BlockChain) *BlkTmplGenerator {
	return &BlkTmplGenerator{
		txSource: txSource,
		chain:    chain,
	}
}
