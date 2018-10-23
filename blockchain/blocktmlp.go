package blockchain

import (
	"errors"
	"fmt"
	"time"

	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/privacy/client"
	"github.com/ninjadotorg/cash/transaction"
)

func (blockgen *BlkTmplGenerator) NewBlockTemplate(payToAddress client.PaymentAddress, chainID byte) (*BlockTemplate, error) {

	prevBlock := blockgen.chain.BestState[chainID]
	prevBlockHash := blockgen.chain.BestState[chainID].BestBlock.Hash()
	sourceTxns := blockgen.txPool.MiningDescs()

	var txsToAdd []transaction.Transaction
	var txToRemove []transaction.Transaction
	// var actionParamTxs []*transaction.ActionParamTx
	var feeMap = map[string]uint64{
		fmt.Sprintf(common.AssetTypeCoin):     0,
		fmt.Sprintf(common.AssetTypeBond):     0,
		fmt.Sprintf(common.AssetTypeGovToken): 0,
		fmt.Sprintf(common.AssetTypeDcbToken): 0,
	}

	// Get reward
	salary := blockgen.rewardAgent.GetSalary()

	if len(sourceTxns) < common.MinTxsInBlock {
		// if len of sourceTxns < MinTxsInBlock -> wait for more transactions
		Logger.log.Info("not enough transactions. Wait for more...")
		<-time.Tick(common.MinBlockWaitTime * time.Second)
		sourceTxns = blockgen.txPool.MiningDescs()
		if len(sourceTxns) == 0 {
			<-time.Tick(common.MaxBlockWaitTime * time.Second)
			sourceTxns = blockgen.txPool.MiningDescs()
			if len(sourceTxns) == 0 {
				// return nil, errors.New("No Tx")
				Logger.log.Info("Creating empty block...")
				goto concludeBlock
			}
		}
	}

	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx
		txChainID, _ := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
		if txChainID != chainID {
			continue
		}
		if !tx.ValidateTransaction() {
			txToRemove = append(txToRemove, transaction.Transaction(tx))
			continue
		}
		txType := tx.GetType()
		txFee := uint64(0)
		switch txType {
		case common.TxActionParamsType:
			// actionParamTxs = append(actionParamTxs, tx.(*transaction.ActionParamTx))
			continue
		case common.TxVotingType:
			txFee = tx.(*transaction.TxVoting).Fee
		case common.TxNormalType:
			txFee = tx.(*transaction.Tx).Fee
		}
		feeMap[txType] += txFee
		txsToAdd = append(txsToAdd, tx)
		if len(txsToAdd) == common.MaxTxsInBlock {
			break
		}
	}

	for _, tx := range txToRemove {
		blockgen.txPool.RemoveTx(tx)
	}

	// check len of txs in block
	if len(txsToAdd) == 0 {
		return nil, errors.New("no transaction available for this chain")
	}

concludeBlock:
	rt := blockgen.chain.BestState[chainID].BestBlock.Header.MerkleRootCommitments.CloneBytes()
	salaryTx, err := createSalaryTx(salary, &payToAddress, rt, chainID)
	if err != nil {
		return nil, err
	}
	// the 1st tx will be coinbaseTx
	txsToAdd = append([]transaction.Transaction{salaryTx}, txsToAdd...)

	merkleRoots := Merkle{}.BuildMerkleTreeStore(txsToAdd)
	merkleRoot := merkleRoots[len(merkleRoots)-1]

	// Get salary fund from txs
	salaryFund := uint64(0)
	for _, blockTx := range txsToAdd {
		if blockTx.GetType() == common.TxVotingType {
			tx, ok := blockTx.(*transaction.TxVoting)
			if !ok {
				Logger.log.Error("Transaction not recognized to store in database")
				continue
			}
			salaryFund += tx.GetValue()
		}
	}

	block := Block{}
	currentSalaryFund := blockgen.chain.BestState[chainID].BestBlock.Header.SalaryFund
	block.Header = BlockHeader{
		Version:               1,
		PrevBlockHash:         *prevBlockHash,
		MerkleRoot:            *merkleRoot,
		MerkleRootCommitments: common.Hash{},
		Timestamp:             time.Now().Unix(),
		BlockCommitteeSigs:    make([]string, common.TotalValidators),
		Committee:             make([]string, common.TotalValidators),
		ChainID:               chainID,
		SalaryFund:            currentSalaryFund - salary + feeMap[common.AssetTypeCoin] + salaryFund,
	}
	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
	}

	// Add new commitments to merkle tree and save the root
	newTree := blockgen.chain.BestState[chainID].CmTree.MakeCopy()
	UpdateMerkleTreeForBlock(newTree, &block)
	rt = newTree.GetRoot(common.IncMerkleTreeHeight)
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

type BlkTmplGenerator struct {
	txPool      TxPool
	chain       *BlockChain
	rewardAgent RewardAgent
	// chainParams *blockchain.Params
	// policy      *Policy
}

type BlockTemplate struct {
	Block *Block
}

// txPool represents a source of transactions to consider for inclusion in
// new blocks.
//
// The interface contract requires that all of these methods are safe for
// concurrent access with respect to the source.
type TxPool interface {
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

type RewardAgent interface {
	GetSalary() uint64
}

func (self BlkTmplGenerator) Init(txPool TxPool, chain *BlockChain, rewardAgent RewardAgent) (*BlkTmplGenerator, error) {
	return &BlkTmplGenerator{
		txPool:      txPool,
		chain:       chain,
		rewardAgent: rewardAgent,
	}, nil
}

/* TODO:
func extractTxsAndComputeInitialFees(txDescs []*transaction.TxDesc) (
	[]transaction.Transaction,
	[]*transaction.ActionParamTx,
	map[string]uint64,
) {
	var txs []transaction.Transaction
	var actionParamTxs []*transaction.ActionParamTx
	var feeMap = map[string]uint64{
		fmt.Sprintf(common.AssetTypeCoin):     0,
		fmt.Sprintf(common.AssetTypeBond):     0,
		fmt.Sprintf(common.AssetTypeGovToken): 0,
		fmt.Sprintf(common.AssetTypeDcbToken): 0,
	}
	for _, txDesc := range txDescs {
		tx := txDesc.Tx
		txs = append(txs, tx)
		txType := tx.GetType()
		txFee := uint64(0)
		switch txType {
		case common.TxActionParamsType:
			actionParamTxs = append(actionParamTxs, tx.(*transaction.ActionParamTx))
			continue
		case common.TxVotingType:
			txFee = tx.(*transaction.TxVoting).Fee
		case common.TxNormalType:
			txFee = tx.(*transaction.Tx).Fee
		}
		normalTx, _ := tx.(*transaction.Tx)
		feeMap[normalTx.Descs[0].Type] += txFee
	}
	return txs, actionParamTxs, feeMap
}*/

// createSalaryTx
// Blockchain use this tx to pay a reward(salary) to miner of chain
// #1 - salary:
// #2 - receiverAddr:
// #3 - rt
// #4 - chainID
func createSalaryTx(
	salary uint64,
	receiverAddr *client.PaymentAddress,
	rt []byte,
	chainID byte,
) (*transaction.Tx, error) {
	// Create Proof for the joinsplit op
	inputs := make([]*client.JSInput, 2)
	inputs[0] = transaction.CreateRandomJSInput(nil)
	inputs[1] = transaction.CreateRandomJSInput(inputs[0].Key)
	dummyAddress := client.GenPaymentAddress(*inputs[0].Key)

	// Create new notes: first one is coinbase UTXO, second one has 0 value
	outNote := &client.Note{Value: salary, Apk: receiverAddr.Apk}
	placeHolderOutputNote := &client.Note{Value: 0, Apk: receiverAddr.Apk}

	outputs := []*client.JSOutput{&client.JSOutput{}, &client.JSOutput{}}
	outputs[0].EncKey = receiverAddr.Pkenc
	outputs[0].OutputNote = outNote
	outputs[1].EncKey = receiverAddr.Pkenc
	outputs[1].OutputNote = placeHolderOutputNote

	// Generate proof and sign tx
	tx, err := transaction.CreateEmptyTx()
	if err != nil {
		return nil, err
	}
	tx.AddressLastByte = dummyAddress.Apk[len(dummyAddress.Apk)-1]
	rtMap := map[byte][]byte{chainID: rt}
	inputMap := map[byte][]*client.JSInput{chainID: inputs}
	err = tx.BuildNewJSDesc(inputMap, outputs, rtMap, salary, 0, false)
	if err != nil {
		return nil, err
	}
	tx, err = transaction.SignTx(tx)
	if err != nil {
		return nil, err
	}
	return tx, err
}
