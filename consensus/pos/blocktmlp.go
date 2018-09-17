package pos

import (
	"errors"
	"fmt"
	"time"

	"github.com/ninjadotorg/cash-prototype/blockchain"
	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

// TODO: create block template (move block template from mining to here)

func (g *BlkTmplGenerator) NewBlockTemplate(payToAddress string, chain *blockchain.BlockChain, chainID byte) (*BlockTemplate, error) {

	prevBlock := chain.BestState[chainID].BestBlock
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

	txs, _, _ := extractTxsAndComputeInitialFees(sourceTxns)
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
	blockTxns := txs
	merkleRoots := blockchain.Merkle{}.BuildMerkleTreeStore(blockTxns)
	merkleRoot := merkleRoots[len(merkleRoots)-1]

	var txToRemove []*transaction.Tx
mempoolLoop:
	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx
		//@todo need apply validate tx, logic check all referenced here
		// call function spendTransaction to mark utxo

		txChainID, _ := g.GetTxSenderChain(tx.(*transaction.Tx).AddressHash)
		if txChainID != chainID {
			continue
		}
		utxos, err := g.chain.FetchUtxoView(*tx.(*transaction.Tx))
		_ = utxos
		_ = err
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
			txToRemove = append(txToRemove, tx.(*transaction.Tx))
			continue mempoolLoop
		}
		// g.txSource.Clear()
	}

	for _, tx := range txToRemove {
		g.txSource.RemoveTx(*tx)
	}
	// TODO PoW
	//time.Sleep(time.Second * 15)
	if len(blockTxns) == 0 {
		return nil, errors.New("no transaction available for this chain")
	}
	block := blockchain.Block{}
	block.Header = blockchain.BlockHeader{
		Version:       1,
		PrevBlockHash: *prevBlockHash,
		MerkleRoot:    *merkleRoot,
		Timestamp:     time.Now(),
		Difficulty:    0, //@todo should be create Difficulty logic
		Nonce:         0, //@todo should be create Nonce logic
	}
	for _, tx := range blockTxns {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
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
	//LastUpdated() time.Time

	// MiningDescs returns a slice of mining descriptors for all the
	// transactions in the source pool.
	MiningDescs() []*transaction.TxDesc

	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	//HaveTransaction(hash *common.Hash) bool

	RemoveTx(tx transaction.Tx)

	// TODO using when demo
	Clear()
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
	map[string]float64,
) {
	var txs []transaction.Transaction
	var actionParamTxs []*transaction.ActionParamTx
	var feeMap = map[string]float64{
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
		if len(normalTx.TxOut) > 0 {
			txOutType := normalTx.TxOut[0].TxOutType
			if txOutType == "" {
				txOutType = common.TxOutCoinType
			}
			feeMap[txOutType] += txDesc.Fee
		}
	}
	return txs, actionParamTxs, feeMap
}
