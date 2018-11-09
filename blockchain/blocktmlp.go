package blockchain

import (
	"time"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
	"github.com/ninjadotorg/constant/transaction"
)

type BlkTmplGenerator struct {
	txPool      TxPool
	chain       *BlockChain
	rewardAgent RewardAgent
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
	GetBasicSalary(chainID byte) uint64
	GetSalaryPerTx(chainID byte) uint64
}

func (self BlkTmplGenerator) Init(txPool TxPool, chain *BlockChain, rewardAgent RewardAgent) (*BlkTmplGenerator, error) {
	return &BlkTmplGenerator{
		txPool:      txPool,
		chain:       chain,
		rewardAgent: rewardAgent,
	}, nil
}

func (blockgen *BlkTmplGenerator) NewBlockTemplate(payToAddress client.PaymentAddress, chainID byte) (*Block, error) {

	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	prevBlockHash := blockgen.chain.BestState[chainID].BestBlock.Hash()
	prevCmTree := blockgen.chain.BestState[chainID].CmTree.MakeCopy()
	sourceTxns := blockgen.txPool.MiningDescs()

	var txsToAdd []transaction.Transaction
	var txToRemove []transaction.Transaction
	// var actionParamTxs []*transaction.ActionParamTx
	totalFee := uint64(0)

	// Get salary per tx
	salaryPerTx := blockgen.rewardAgent.GetSalaryPerTx(chainID)
	// Get basic salary on block
	basicSalary := blockgen.rewardAgent.GetBasicSalary(chainID)

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
		totalFee += tx.GetTxFee()
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
		// return nil, errors.New("no transaction available for this chain")
		Logger.log.Info("Creating empty block...")
	}

concludeBlock:
// Get blocksalary fund from txs
	salaryFundAdd := uint64(0)
	salaryMULTP := uint64(0) //salary multiplier
	for _, blockTx := range txsToAdd {
		if blockTx.GetType() == common.TxVotingType {
			tx, ok := blockTx.(*transaction.TxRegisterCandidate)
			if !ok {
				Logger.log.Error("Transaction not recognized to store in database")
				continue
			}
			salaryFundAdd += tx.GetValue()
		}
		if blockTx.GetTxFee() > 0 {
			salaryMULTP++
		}
	}

	rt := prevBlock.Header.MerkleRootCommitments.CloneBytes()

	// ------------------------ HOW to GET salary on a block-------------------
	// total salary = tx * (salary per tx) + (basic salary on block)
	// ------------------------------------------------------------------------
	totalSalary := salaryMULTP*salaryPerTx + basicSalary
	// create salary tx to pay constant for block producer
	salaryTx, err := createSalaryTx(totalSalary, &payToAddress, rt, chainID)

	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	// the 1st tx will be salaryTx
	txsToAdd = append([]transaction.Transaction{salaryTx}, txsToAdd...)

	merkleRoots := Merkle{}.BuildMerkleTreeStore(txsToAdd)
	merkleRoot := merkleRoots[len(merkleRoots)-1]

	block := Block{
		Transactions: make([]transaction.Transaction, 0),
	}
	currentSalaryFund := prevBlock.Header.SalaryFund
	block.Header = BlockHeader{
		Height:                prevBlock.Header.Height + 1,
		Version:               BlockVersion,
		PrevBlockHash:         *prevBlockHash,
		MerkleRoot:            *merkleRoot,
		MerkleRootCommitments: common.Hash{},
		Timestamp:             time.Now().Unix(),
		BlockCommitteeSigs:    make([]string, common.TotalValidators),
		Committee:             make([]string, common.TotalValidators),
		ChainID:               chainID,
		SalaryFund:            currentSalaryFund - totalSalary + totalFee + salaryFundAdd,
		GovernanceParams:      prevBlock.Header.GovernanceParams, // TODO: need get from gov-params tx
	}
	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
	}

	// Add new commitments to merkle tree and save the root
	newTree := prevCmTree
	err = UpdateMerkleTreeForBlock(newTree, &block)
	if err != nil {
		// TODO check error to process
		return nil, err
	}
	rt = newTree.GetRoot(common.IncMerkleTreeHeight)
	copy(block.Header.MerkleRootCommitments[:], rt)

	//update the latest AgentDataPoints to block
	// block.AgentDataPoints = agentDataPoints

	return &block, nil
}

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

	// Create new notes: first one is salary UTXO, second one has 0 value
	outNote := &client.Note{Value: salary, Apk: receiverAddr.Apk}
	placeHolderOutputNote := &client.Note{Value: 0, Apk: receiverAddr.Apk}

	outputs := []*client.JSOutput{&client.JSOutput{}, &client.JSOutput{}}
	outputs[0].EncKey = receiverAddr.Pkenc
	outputs[0].OutputNote = outNote
	outputs[1].EncKey = receiverAddr.Pkenc
	outputs[1].OutputNote = placeHolderOutputNote

	// Generate proof and sign tx
	tx, err := transaction.CreateEmptyTx(common.TxSalaryType)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	tx.AddressLastByte = dummyAddress.Apk[len(dummyAddress.Apk)-1]
	rtMap := map[byte][]byte{chainID: rt}
	inputMap := map[byte][]*client.JSInput{chainID: inputs}

	// NOTE: always pay salary with constant coin
	err = tx.BuildNewJSDesc(inputMap, outputs, rtMap, salary, 0, false)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	err = tx.SignTx()
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	return tx, nil
}
