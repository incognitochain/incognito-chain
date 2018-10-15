package blockchain

import (
	"errors"
	"fmt"
	"time"

	"github.com/ninjadotorg/cash-prototype/common"
	"github.com/ninjadotorg/cash-prototype/privacy/client"
	"github.com/ninjadotorg/cash-prototype/transaction"
)

func (blockgen *BlkTmplGenerator) NewBlockTemplate(payToAddress client.PaymentAddress, chainID byte) (*BlockTemplate, error) {

	prevBlock := blockgen.chain.BestState[chainID]
	prevBlockHash := blockgen.chain.BestState[chainID].BestBlock.Hash()
	sourceTxns := blockgen.txPool.MiningDescs()

	var txsToAdd []transaction.Transaction
	var txToRemove []transaction.Transaction
	var feeMap map[string]uint64
	var txs []transaction.Transaction

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
	coinbaseTx, err := createSalaryTx(
		salary,
		&payToAddress,
		rt,
		chainID,
	)
	if err != nil {
		return nil, err
	}
	// the 1st tx will be coinbaseTx
	txsToAdd = append([]transaction.Transaction{coinbaseTx}, txsToAdd...)

	merkleRoots := Merkle{}.BuildMerkleTreeStore(txsToAdd)
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
		} else if blockTx.GetType() == common.TxVotingType {
			tx, ok := blockTx.(*transaction.TxVoting)
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
	blockgen.chain.StoreCommitmentsFromListCommitment(commitments, descType, chainID)
	blockgen.chain.StoreNullifiersFromListNullifier(nullifiers, descType, chainID)

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
		SalaryFund:            currentSalaryFund - salary + feeMap[common.TxOutCoinType],
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
		} else if tempBlockTx.GetType() == common.TxVotingType {
			tx, ok := tempBlockTx.(*transaction.TxVoting)
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

/*func createCoinbaseTx(
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
	var reward uint64 = common.DefaultCoinBaseTxReward + feeMap[common.TxOutCoinType] // TODO: probably will need compute reward based on block height

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
	tx := transaction.CreateEmptyTx()
	tx.AddressLastByte = dummyAddress.Apk[len(dummyAddress.Apk)-1]
	rtMap := map[byte][]byte{chainID: rt}
	inputMap := map[byte][]*client.JSInput{chainID: inputs}
	err := tx.BuildNewJSDesc(inputMap, outputs, rtMap, salary, 0)
	if err != nil {
		return nil, err
	}
	tx, err = transaction.SignTx(tx)
	if err != nil {
		return nil, err
	}
	return tx, err
}
