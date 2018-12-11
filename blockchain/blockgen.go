package blockchain

import (
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

type BlkTmplGeneratorNew struct {
	txPool      TxPool
	chain       *BlockChain
	rewardAgent RewardAgent
}

func (self BlkTmplGeneratorNew) Init(txPool TxPool, chain *BlockChain, rewardAgent RewardAgent) (*BlkTmplGenerator, error) {
	return &BlkTmplGenerator{
		txPool:      txPool,
		chain:       chain,
		rewardAgent: rewardAgent,
	}, nil
}

func (blockgen *BlkTmplGeneratorNew) NewBlockTemplate(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, chainID byte, blockType string) (*BlockV2, error) {

	// 	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	// 	prevBlockHash := blockgen.chain.BestState[chainID].BestBlock.Hash()
	// 	sourceTxns := blockgen.txPool.MiningDescs()

	// 	var txsToAdd []transaction.Transaction
	// 	var txToRemove []transaction.Transaction
	// 	var buySellReqTxs []transaction.Transaction
	// 	//var txTokenVouts map[*common.Hash]*transaction.TxTokenVout
	// 	bondsSold := uint64(0)
	// 	incomeFromBonds := uint64(0)
	// 	totalFee := uint64(0)
	// 	buyBackCoins := uint64(0)

	// 	// Get salary per tx
	// 	salaryPerTx := blockgen.rewardAgent.GetSalaryPerTx(chainID)
	// 	// Get basic salary on block
	// 	basicSalary := blockgen.rewardAgent.GetBasicSalary(chainID)

	// 	// Check if it is the case we need to apply a new proposal
	// 	// 1. newNW < lastNW * 0.9
	// 	// 2. current block height == last Constitution start time + last Constitution execute duration
	// 	/*if blockgen.neededNewDCBConstitution(chainID) {
	// 		tx, err := blockgen.createRequestConstitutionTxDecs(chainID, DCBConstitutionHelper{})
	// 		if err != nil {
	// 			Logger.log.Error(err)
	// 			return nil, err
	// 		}
	// 		sourceTxns = append(sourceTxns, tx)
	// 	}
	// 	if blockgen.neededNewGovConstitution(chainID) {
	// 		tx, err := blockgen.createRequestConstitutionTxDecs(chainID, GOVConstitutionHelper{})
	// 		if err != nil {
	// 			Logger.log.Error(err)
	// 			return nil, err
	// 		}
	// 		sourceTxns = append(sourceTxns, tx)
	// 	}*/

	// 	if len(sourceTxns) < common.MinTxsInBlock {
	// 		// if len of sourceTxns < MinTxsInBlock -> wait for more transactions
	// 		Logger.log.Info("not enough transactions. Wait for more...")
	// 		<-time.Tick(common.MinBlockWaitTime * time.Second)
	// 		sourceTxns = blockgen.txPool.MiningDescs()
	// 		if len(sourceTxns) == 0 {
	// 			<-time.Tick(common.MaxBlockWaitTime * time.Second)
	// 			sourceTxns = blockgen.txPool.MiningDescs()
	// 			if len(sourceTxns) == 0 {
	// 				// return nil, errors.New("No TxNormal")
	// 				Logger.log.Info("Creating empty block...")
	// 				goto concludeBlock
	// 			}
	// 		}
	// 	}

	// 	for _, txDesc := range sourceTxns {
	// 		tx := txDesc.Tx
	// 		txChainID, _ := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
	// 		if txChainID != chainID {
	// 			continue
	// 		}
	// 		// ValidateTransaction vote and propose transaction

	// 		if !blockgen.txPool.ValidateTxByItSelf(tx) {
	// 			txToRemove = append(txToRemove, transaction.Transaction(tx))
	// 			continue
	// 		}

	// 		/*if tx.GetType() == common.TxBuyFromGOVRequest {
	// 			income, soldAmt, addable := blockgen.checkBuyFromGOVReqTx(chainID, tx, bondsSold)
	// 			if !addable {
	// 				txToRemove = append(txToRemove, tx)
	// 				continue
	// 			}
	// 			bondsSold += soldAmt
	// 			incomeFromBonds += income
	// 			buySellReqTxs = append(buySellReqTxs, tx)
	// 		}

	// 		if tx.GetType() == common.TxBuyBackRequest {
	// 			txTokenVout, buyBackReqTxID, addable := blockgen.checkBuyBackReqTx(chainID, tx, buyBackCoins)
	// 			if !addable {
	// 				txToRemove = append(txToRemove, tx)
	// 				continue
	// 			}
	// 			buyBackCoins += txTokenVout.Value * txTokenVout.BuySellResponse.BuyBackInfo.BuyBackPrice
	// 			txTokenVouts[buyBackReqTxID] = txTokenVout
	// 		}*/

	// 		totalFee += tx.GetTxFee()
	// 		txsToAdd = append(txsToAdd, tx)
	// 		if len(txsToAdd) == common.MaxTxsInBlock {
	// 			break
	// 		}
	// 	}

	// 	for _, tx := range txToRemove {
	// 		blockgen.txPool.RemoveTx(tx)
	// 	}

	// 	// check len of txs in block
	// 	if len(txsToAdd) == 0 {
	// 		// return nil, errors.New("no transaction available for this chain")
	// 		Logger.log.Info("Creating empty block...")
	// 	}

	// concludeBlock:
	// 	//rt := prevBlock.Header.MerkleRootCommitments.CloneBytes()
	// 	_ = prevBlock.Header.Height + 1

	// 	// TODO
	// 	bankPayoutAmount := uint64(0)
	// 	// Process dividend payout for DCB if needed
	// 	/*bankDivTxs, bankPayoutAmount, err := blockgen.processBankDividend(rt, chainID, blockHeight)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	for _, tx := range bankDivTxs {
	// 		txsToAdd = append(txsToAdd, tx)
	// 	}*/

	// 	// TODO
	// 	// Process dividend payout for GOV if needed
	// 	/*govDivTxs, govPayoutAmount, err := blockgen.processGovDividend(rt, chainID, blockHeight)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	for _, tx := range govDivTxs {
	// 		txsToAdd = append(txsToAdd, tx)
	// 	}*/

	// 	// Process crowdsale for DCB
	// 	/*dcbSaleTxs, removableTxs, err := blockgen.processCrowdsale(sourceTxns, rt, chainID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	for _, tx := range dcbSaleTxs {
	// 		txsToAdd = append(txsToAdd, tx)
	// 	}
	// 	for _, tx := range removableTxs {
	// 		txToRemove = append(txToRemove, tx)
	// 	}*/

	// 	// Get blocksalary fund from txs
	// 	salaryFundAdd := uint64(0)
	// 	salaryMULTP := uint64(0) //salary multiplier
	// 	for _, blockTx := range txsToAdd {
	// 		if blockTx.GetTxFee() > 0 {
	// 			salaryMULTP++
	// 		}
	// 	}

	// 	// ------------------------ HOW to GET salary on a block-------------------
	// 	// total salary = tx * (salary per tx) + (basic salary on block)
	// 	// ------------------------------------------------------------------------
	// 	totalSalary := salaryMULTP*salaryPerTx + basicSalary
	// 	// create salary tx to pay constant for block producer
	// 	salaryTx, err := transaction.CreateTxSalary(totalSalary, payToAddress, privatekey, blockgen.chain.config.DataBase)
	// 	if err != nil {
	// 		Logger.log.Error(err)
	// 		return nil, err
	// 	}
	// 	// create buy/sell response txs to distribute bonds/govs to requesters
	// 	// buySellResTxs := blockgen.buildBuySellResponsesTx(
	// 	// 	common.TxBuyFromGOVResponse,
	// 	// 	buySellReqTxs,
	// 	// 	blockgen.chain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.SellingBonds,
	// 	// )

	// 	// create buy-back response txs to distribute constants to buy-back requesters
	// 	//buyBackResTxs, err := blockgen.buildBuyBackResponsesTx(common.TxBuyBackResponse, txTokenVouts, chainID)
	// 	currentSalaryFund := prevBlock.Header.SalaryFund
	// 	// create refund txs
	// 	//remainingFund := currentSalaryFund + totalFee + salaryFundAdd + incomeFromBonds - (totalSalary + buyBackCoins)
	// 	//refundTxs, totalRefundAmt := blockgen.buildRefundTxs(chainID, remainingFund)

	// 	coinbases := []transaction.Transaction{salaryTx}
	// 	for _, resTx := range buySellResTxs {
	// 		coinbases = append(coinbases, resTx)
	// 	}
	// 	/*for _, resTx := range buyBackResTxs {
	// 		coinbases = append(coinbases, resTx)
	// 	}
	// 	for _, refundTx := range refundTxs {
	// 		coinbases = append(coinbases, refundTx)
	// 	}*/
	// 	txsToAdd = append(coinbases, txsToAdd...)
	// 	govPayoutAmount := uint64(0)
	// 	totalRefundAmt := uint64(0)

	// 	// Check for final balance of DCB and GOV
	// 	if currentSalaryFund+totalFee+salaryFundAdd+incomeFromBonds < totalSalary+govPayoutAmount+buyBackCoins+totalRefundAmt {
	// 		return nil, fmt.Errorf("Gov fund is not enough for salary and dividend payout")
	// 	}

	// 	/*currentBankFund := prevBlock.Header.BankFund
	// 	if currentBankFund < bankPayoutAmount {
	// 		return nil, fmt.Errorf("Bank fund is not enough for dividend payout")
	// 	}*/

	// 	merkleRoots := Merkle{}.BuildMerkleTreeStore(txsToAdd)
	// 	merkleRoot := merkleRoots[len(merkleRoots)-1]

	// 	block := Block{
	// 		Transactions: make([]transaction.Transaction, 0),
	// 	}
	// 	block.Header = BlockHeader{
	// 		Height:        prevBlock.Header.Height + 1,
	// 		Version:       BlockVersion,
	// 		PrevBlockHash: *prevBlockHash,
	// 		MerkleRoot:    *merkleRoot,
	// 		//MerkleRootCommitments: common.Hash{},
	// 		Timestamp:          time.Now().Unix(),
	// 		BlockCommitteeSigs: make([]string, common.TotalValidators),
	// 		Committee:          make([]string, common.TotalValidators),
	// 		ChainID:            chainID,
	// 		SalaryFund:         currentSalaryFund + incomeFromBonds + totalFee + salaryFundAdd - totalSalary - govPayoutAmount - buyBackCoins - totalRefundAmt,
	// 		BankFund:           prevBlock.Header.BankFund - bankPayoutAmount,
	// 		GOVConstitution:    prevBlock.Header.GOVConstitution, // TODO: need get from gov-params tx
	// 		DCBConstitution:    prevBlock.Header.DCBConstitution, // TODO: need get from dcb-params tx
	// 		LoanParams:         prevBlock.Header.LoanParams,
	// 	}
	// 	if block.Header.GOVConstitution.GOVParams.SellingBonds != nil {
	// 		block.Header.GOVConstitution.GOVParams.SellingBonds.BondsToSell -= bondsSold
	// 	}
	// 	for _, tx := range txsToAdd {
	// 		if err := block.AddTransaction(tx); err != nil {
	// 			return nil, err
	// 		}
	// 		/* TODO
	// 		if tx.GetType() == common.TxAcceptDCBProposal {
	// 			block.updateDCBConstitution(tx, blockgen)
	// 		}
	// 		if tx.GetType() == common.TxAcceptGOVProposal {
	// 			block.updateGOVConstitution(tx, blockgen)
	// 		}*/
	// 	}

	// 	// Add new commitments to merkle tree and save the root
	// 	/*newTree := prevCmTree
	// 	err = UpdateMerkleTreeForBlock(newTree, &block)
	// 	if err != nil {
	// 		Logger.log.Error(err)
	// 		return nil, err
	// 	}
	// 	rt = newTree.GetRoot(common.IncMerkleTreeHeight)
	// 	copy(block.Header.MerkleRootCommitments[:], rt)*/

	// 	//update the latest AgentDataPoints to block
	// 	// block.AgentDataPoints = agentDataPoints
	// 	return &block, nil
	return &BlockV2{}, nil
}

// tmpBlk := &blockchain.BlockV2{
// 	AggregatedSig: []byte{0, 0, 0, 0},
// 	ValidatorsIdx: []int{1, 2, 3},
// 	ProducerSig:   []byte{0, 0, 0, 0},
// 	Type:          "beacon",
// 	Header: blockchain.BlockHeaderBeacon{
// 		BlockHeaderGeneric: blockchain.BlockHeaderGeneric{
// 			Height: 1,
// 		},
// 		TestParam: "lskdfjglsfj;fgjs;",
// 	},
// 	Body: blockchain.BlockBodyBeacon{},
// }
// fmt.Println(tmpBlk)
// test, err := json.Marshal(tmpBlk)
// if err != nil {
// 	fmt.Println(err)
// }
// fmt.Println(string(test))

// decodeBlk := &blockchain.BlockV2{}
// err = decodeBlk.UnmarshalJSON(test)
// if err != nil {
// 	fmt.Println(err)
// }
// fmt.Println(decodeBlk)
// test2, err := json.Marshal(decodeBlk)
// if err != nil {
// 	fmt.Println(err)
// }
// fmt.Println(string(test2))
// return
