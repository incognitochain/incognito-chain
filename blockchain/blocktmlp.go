package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
)

type BlkTmplGenerator struct {
	txPool      TxPool
	chain       *BlockChain
	rewardAgent RewardAgent
}

type ConstitutionHelper interface {
	GetStartedBlockHeight(generator *BlkTmplGenerator, chainID byte) int32
	CheckSubmitProposalType(tx transaction.Transaction) bool
	CheckVotingProposalType(tx transaction.Transaction) bool
	GetAmountVoteToken(tx transaction.Transaction) uint32
	TxAcceptProposal(originTx transaction.Transaction) transaction.Transaction
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

	//CheckTransactionFee
	CheckTransactionFee(tx transaction.Transaction) (uint64, error)

	// Check tx validate by it self
	ValidateTxByItSelf(tx transaction.Transaction) bool
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

func (blockgen *BlkTmplGenerator) NewBlockTemplate(payToAddress privacy.PaymentAddress, chainID byte) (*Block, error) {

	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	prevBlockHash := blockgen.chain.BestState[chainID].BestBlock.Hash()
	prevCmTree := blockgen.chain.BestState[chainID].CmTree.MakeCopy()
	sourceTxns := blockgen.txPool.MiningDescs()

	var txsToAdd []transaction.Transaction
	var txToRemove []transaction.Transaction
	var buySellReqTxs []transaction.Transaction
	var txTokenVouts map[*common.Hash]*transaction.TxTokenVout
	bondsSold := uint64(0)
	incomeFromBonds := uint64(0)
	totalFee := uint64(0)
	buyBackCoins := uint64(0)

	// Get salary per tx
	salaryPerTx := blockgen.rewardAgent.GetSalaryPerTx(chainID)
	// Get basic salary on block
	basicSalary := blockgen.rewardAgent.GetBasicSalary(chainID)

	// Check if it is the case we need to apply a new proposal
	// 1. newNW < lastNW * 0.9
	// 2. current block height == last Constitution start time + last Constitution execute duration
	if blockgen.neededNewDCBConstitution(chainID) {
		tx, err := blockgen.createRequestConstitutionTxDecs(chainID, DCBConstitutionHelper{})
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		sourceTxns = append(sourceTxns, tx)
	}
	if blockgen.neededNewGovConstitution(chainID) {
		tx, err := blockgen.createRequestConstitutionTxDecs(chainID, GOVConstitutionHelper{})
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		sourceTxns = append(sourceTxns, tx)
	}

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
		// ValidateTransaction vote and propose transaction

		if !blockgen.txPool.ValidateTxByItSelf(tx) {
			txToRemove = append(txToRemove, transaction.Transaction(tx))
			continue
		}

		if tx.GetType() == common.TxBuyFromGOVRequest {
			income, soldAmt, addable := blockgen.checkBuyFromGOVReqTx(chainID, tx, bondsSold)
			if !addable {
				txToRemove = append(txToRemove, tx)
				continue
			}
			bondsSold += soldAmt
			incomeFromBonds += income
			buySellReqTxs = append(buySellReqTxs, tx)
		}

		if tx.GetType() == common.TxBuyBackRequest {
			txTokenVout, buyBackReqTxID, addable := blockgen.checkBuyBackReqTx(chainID, tx, buyBackCoins)
			if !addable {
				txToRemove = append(txToRemove, tx)
				continue
			}
			buyBackCoins += txTokenVout.Value * txTokenVout.BuySellResponse.BuyBackInfo.BuyBackPrice
			txTokenVouts[buyBackReqTxID] = txTokenVout
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
	rt := prevBlock.Header.MerkleRootCommitments.CloneBytes()
	blockHeight := prevBlock.Header.Height + 1

	// Process dividend payout for DCB if needed
	bankDivTxs, bankPayoutAmount, err := blockgen.processBankDividend(rt, chainID, blockHeight)
	if err != nil {
		return nil, err
	}
	for _, tx := range bankDivTxs {
		txsToAdd = append(txsToAdd, tx)
	}

	// Process dividend payout for GOV if needed
	govDivTxs, govPayoutAmount, err := blockgen.processGovDividend(rt, chainID, blockHeight)
	if err != nil {
		return nil, err
	}
	for _, tx := range govDivTxs {
		txsToAdd = append(txsToAdd, tx)
	}

	// Get blocksalary fund from txs
	salaryFundAdd := uint64(0)
	salaryMULTP := uint64(0) //salary multiplier
	for _, blockTx := range txsToAdd {
		if blockTx.GetTxFee() > 0 {
			salaryMULTP++
		}
	}

	// ------------------------ HOW to GET salary on a block-------------------
	// total salary = tx * (salary per tx) + (basic salary on block)
	// ------------------------------------------------------------------------
	totalSalary := salaryMULTP*salaryPerTx + basicSalary
	// create salary tx to pay constant for block producer
	salaryTx, err := transaction.CreateTxSalary(totalSalary, &payToAddress, rt, chainID)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	// create buy/sell response tx to distribute bonds/govs to requesters
	buySellResTx := buildBuySellResponsesTx(
		common.TxBuyFromGOVResponse,
		buySellReqTxs,
		blockgen.chain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.SellingBonds,
	)
	// create buy-back response tx to distribute constants to buy-back requesters
	buyBackResTx := buildBuyBackResponsesTx(
		common.TxBuyBackResponse,
		txTokenVouts,
	)
	// create refund txs
	currentSalaryFund := prevBlock.Header.SalaryFund
	remainingFund := currentSalaryFund + totalFee + salaryFundAdd + incomeFromBonds - (totalSalary + buyBackCoins)
	refundTxs, totalRefundAmt := blockgen.buildRefundTxs(chainID, remainingFund)

	coinbases := []transaction.Transaction{salaryTx}
	if buySellResTx != nil {
		coinbases = append(coinbases, buySellResTx)
	}
	if buyBackResTx != nil {
		coinbases = append(coinbases, buyBackResTx)
	}
	if len(refundTxs) > 0 {
		// coinbases = append(coinbases, refundTxs...)
		for _, refundTx := range refundTxs {
			coinbases = append(coinbases, refundTx)
		}
	}
	txsToAdd = append(coinbases, txsToAdd...)

	// Check for final balance of DCB and GOV
	if currentSalaryFund+totalFee+salaryFundAdd+incomeFromBonds < totalSalary+govPayoutAmount+buyBackCoins+totalRefundAmt {
		return nil, fmt.Errorf("Gov fund is not enough for salary and dividend payout")
	}

	currentBankFund := prevBlock.Header.BankFund
	if currentBankFund < bankPayoutAmount {
		return nil, fmt.Errorf("Bank fund is not enough for dividend payout")
	}

	merkleRoots := Merkle{}.BuildMerkleTreeStore(txsToAdd)
	merkleRoot := merkleRoots[len(merkleRoots)-1]

	block := Block{
		Transactions: make([]transaction.Transaction, 0),
	}
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
		SalaryFund:            currentSalaryFund + incomeFromBonds + totalFee + salaryFundAdd - totalSalary - govPayoutAmount - buyBackCoins - totalRefundAmt,
		BankFund:              prevBlock.Header.BankFund - bankPayoutAmount,
		GOVConstitution:       prevBlock.Header.GOVConstitution, // TODO: need get from gov-params tx
		DCBConstitution:       prevBlock.Header.DCBConstitution, // TODO: need get from dcb-params tx
		LoanParams:            prevBlock.Header.LoanParams,
	}
	if block.Header.GOVConstitution.GOVParams.SellingBonds != nil {
		block.Header.GOVConstitution.GOVParams.SellingBonds.BondsToSell -= bondsSold
	}
	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
		if tx.GetType() == common.TxAcceptDCBProposal {
			block.updateDCBConstitution(tx, blockgen)
		}
		if tx.GetType() == common.TxAcceptGOVProposal {
			block.updateGOVConstitution(tx, blockgen)
		}
	}

	// Add new commitments to merkle tree and save the root
	newTree := prevCmTree
	err = UpdateMerkleTreeForBlock(newTree, &block)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	rt = newTree.GetRoot(common.IncMerkleTreeHeight)
	copy(block.Header.MerkleRootCommitments[:], rt)

	//update the latest AgentDataPoints to block
	// block.AgentDataPoints = agentDataPoints
	return &block, nil
}

func updateDCBConstitution(block *Block, tx transaction.Transaction, blockgen *BlkTmplGenerator) error {
	txAcceptDCBProposal := tx.(transaction.TxAcceptDCBProposal)
	_, _, _, getTx, err := blockgen.chain.GetTransactionByHash(txAcceptDCBProposal.DCBProposalTXID)
	DCBProposal := getTx.(*transaction.TxSubmitDCBProposal)
	if err != nil {
		return err
	}
	block.Header.DCBConstitution.StartedBlockHeight = block.Header.Height
	block.Header.DCBConstitution.ExecuteDuration = DCBProposal.DCBProposalData.ExecuteDuration
	block.Header.DCBConstitution.ProposalTXID = txAcceptDCBProposal.DCBProposalTXID
	block.Header.DCBConstitution.CurrentDCBNationalWelfare = GetOracleDCBNationalWelfare()

	//	proposalParams := DCBProposal.DCBProposalData.DCBParams // not use yet
	block.Header.DCBConstitution.DCBParams = DCBParams{}
	return nil
}

func updateGOVConstitution(block *Block, tx transaction.Transaction, blockgen *BlkTmplGenerator) error {
	txAcceptGOVProposal := tx.(transaction.TxAcceptGOVProposal)
	_, _, _, getTx, err := blockgen.chain.GetTransactionByHash(txAcceptGOVProposal.GOVProposalTXID)
	GOVProposal := getTx.(*transaction.TxSubmitGOVProposal)
	if err != nil {
		return err
	}
	block.Header.GOVConstitution.StartedBlockHeight = block.Header.Height
	block.Header.GOVConstitution.ExecuteDuration = GOVProposal.GOVProposalData.ExecuteDuration
	block.Header.GOVConstitution.ProposalTXID = txAcceptGOVProposal.GOVProposalTXID
	block.Header.GOVConstitution.CurrentGOVNationalWelfare = GetOracleGOVNationalWelfare()

	proposalParams := GOVProposal.GOVProposalData.GOVParams
	block.Header.GOVConstitution.GOVParams = GOVParams{
		proposalParams.SalaryPerTx,
		proposalParams.BasicSalary,
		&SellingBonds{
			proposalParams.SellingBonds.BondsToSell,
			proposalParams.SellingBonds.BondPrice,
			proposalParams.SellingBonds.Maturity,
			proposalParams.SellingBonds.BuyBackPrice,
			proposalParams.SellingBonds.StartSellingAt,
			proposalParams.SellingBonds.SellingWithin,
		},
		&RefundInfo{},
	}
	return nil
}

func GetOracleDCBNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!")
	return 1234
}
func GetOracleGOVNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!")
	return 1234
}

//1. Current National welfare (NW)  < lastNW * 0.9 (Emergency case)
//2. Block height == last constitution start time + last constitution window
func (blockgen *BlkTmplGenerator) neededNewDCBConstitution(chainID byte) bool {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastDCBConstitution := BestBlock.Header.DCBConstitution
	if GetOracleDCBNationalWelfare() < lastDCBConstitution.CurrentDCBNationalWelfare*ThresholdRatioOfDCBCrisis/100 ||
		BestBlock.Header.Height+1 == lastDCBConstitution.StartedBlockHeight+lastDCBConstitution.ExecuteDuration {
		return true
	}
	return false
}
func (blockgen *BlkTmplGenerator) neededNewGovConstitution(chainID byte) bool {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastGovConstitution := BestBlock.Header.GOVConstitution
	if GetOracleGOVNationalWelfare() < lastGovConstitution.CurrentGOVNationalWelfare*ThresholdRatioOfGovCrisis/100 ||
		BestBlock.Header.Height+1 == lastGovConstitution.StartedBlockHeight+lastGovConstitution.ExecuteDuration {
		return true
	}
	return false
}

func (blockgen *BlkTmplGenerator) createRequestConstitutionTxDecs(
	chainID byte,
	ConstitutionHelper ConstitutionHelper,
) (*transaction.TxDesc, error) {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock

	// count vote from lastConstitution.StartedBlockHeight to Bestblock height
	CountVote := make(map[common.Hash]int64)
	Transaction := make(map[common.Hash]*transaction.Transaction)
	for blockHeight := ConstitutionHelper.GetStartedBlockHeight(blockgen, chainID); blockHeight < BestBlock.Header.Height; blockHeight += 1 {
		//retrieve block from block's height
		hashBlock, err := blockgen.chain.config.DataBase.GetBlockByIndex(blockHeight, chainID)
		if err != nil {
			return nil, err
		}
		blockBytes, err := blockgen.chain.config.DataBase.FetchBlock(hashBlock)
		if err != nil {
			return nil, err
		}
		block := Block{}
		err = json.Unmarshal(blockBytes, &block)
		if err != nil {
			return nil, err
		}
		//count vote of this block
		for _, tx := range block.Transactions {
			_, exist := CountVote[*tx.Hash()]
			if ConstitutionHelper.CheckSubmitProposalType(tx) {
				if exist {
					return nil, err
				}
				CountVote[*tx.Hash()] = 0
				Transaction[*tx.Hash()] = &tx
			} else {
				if ConstitutionHelper.CheckVotingProposalType(tx) {
					if !exist {
						return nil, err
					}
					CountVote[*tx.Hash()] += int64(ConstitutionHelper.GetAmountVoteToken(tx))
				}
			}
		}
	}

	// get transaction and create transaction desc
	var maxVote int64
	var res common.Hash
	for key, value := range CountVote {
		if value > maxVote {
			maxVote = value
			res = key
		}
	}

	acceptedSubmitProposalTransaction := ConstitutionHelper.TxAcceptProposal(*Transaction[res])

	AcceptedTransactionDesc := transaction.TxDesc{
		Tx:     acceptedSubmitProposalTransaction,
		Added:  time.Now(),
		Height: BestBlock.Header.Height,
		Fee:    0,
	}
	return &AcceptedTransactionDesc, nil
}

func (blockgen *BlkTmplGenerator) processDividend(
	rt []byte,
	chainID byte,
	proposal *transaction.PayoutProposal,
	blockHeight int32,
) ([]*transaction.TxDividendPayout, uint64, error) {
	payoutAmount := uint64(0)

	// TODO(@0xbunyip): how to execute payout dividend proposal
	dividendTxs := []*transaction.TxDividendPayout{}
	if false && chainID == 0 && blockHeight%transaction.PayoutFrequency == 0 { // only chain 0 process dividend proposals
		totalTokenSupply, tokenHolders, amounts, err := blockgen.chain.GetAmountPerAccount(proposal)
		if err != nil {
			return nil, 0, err
		}

		infos := []transaction.DividendInfo{}
		// Build tx to pay dividend to each holder
		for i, holder := range tokenHolders {
			holderAddress := (&privacy.PaymentAddress{}).FromBytes(holder)
			info := transaction.DividendInfo{
				TokenHolder: *holderAddress,
				Amount:      amounts[i] / totalTokenSupply,
			}
			payoutAmount += info.Amount
			infos = append(infos, info)

			if len(infos) > transaction.MaxDivTxsPerBlock {
				break // Pay dividend to only some token holders in this block
			}
		}

		dividendTxs, err = transaction.BuildDividendTxs(infos, rt, chainID, proposal)
		if err != nil {
			return nil, 0, err
		}
	}
	return dividendTxs, payoutAmount, nil
}

func (blockgen *BlkTmplGenerator) processBankDividend(rt []byte, chainID byte, blockHeight int32) ([]*transaction.TxDividendPayout, uint64, error) {
	tokenID := &common.Hash{} // TODO(@0xbunyip): hard-code tokenID of BANK token and get proposal
	proposal := &transaction.PayoutProposal{
		TokenID: tokenID,
	}
	return blockgen.processDividend(rt, chainID, proposal, blockHeight)
}

func (blockgen *BlkTmplGenerator) processGovDividend(rt []byte, chainID byte, blockHeight int32) ([]*transaction.TxDividendPayout, uint64, error) {
	tokenID := &common.Hash{} // TODO(@0xbunyip): hard-code tokenID of GOV token and get proposal
	proposal := &transaction.PayoutProposal{
		TokenID: tokenID,
	}
	return blockgen.processDividend(rt, chainID, proposal, blockHeight)
}

func buildSingleBuySellResponseTx(
	buySellReqTx *transaction.TxBuySellRequest,
	sellingBondsParam *SellingBonds,
) transaction.TxTokenVout {
	buyBackInfo := &transaction.BuyBackInfo{
		Maturity:     sellingBondsParam.Maturity,
		BuyBackPrice: sellingBondsParam.BuyBackPrice,
	}
	buySellResponse := &transaction.BuySellResponse{
		BuyBackInfo:   buyBackInfo,
		AssetID:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s%s", sellingBondsParam.Maturity, sellingBondsParam.BuyBackPrice, sellingBondsParam.StartSellingAt))),
		RequestedTxID: buySellReqTx.Hash(),
	}
	return transaction.TxTokenVout{
		Value:           buySellReqTx.Amount,
		PaymentAddress:  buySellReqTx.PaymentAddress,
		BuySellResponse: buySellResponse,
	}
}

func (blockgen *BlkTmplGenerator) checkBuyFromGOVReqTx(
	chainID byte,
	tx transaction.Transaction,
	bondsSold uint64,
) (uint64, uint64, bool) {
	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	sellingBondsParams := prevBlock.Header.GOVConstitution.GOVParams.SellingBonds
	if uint32(prevBlock.Header.Height)+1 > sellingBondsParams.StartSellingAt+sellingBondsParams.SellingWithin {
		return 0, 0, false
	}

	reqTx, ok := tx.(*transaction.TxBuySellRequest)
	if !ok {
		return 0, 0, false
	}
	if bondsSold+reqTx.Amount > sellingBondsParams.BondsToSell { // run out of bonds for selling
		return 0, 0, false
	}
	return reqTx.Amount * reqTx.BuyPrice, reqTx.Amount, true
}

// buildBuySellResponsesTx
// the tx is to distribute tokens (bond, gov, ...) to token requesters
func buildBuySellResponsesTx(
	coinbaseTxType string,
	buySellReqTxs []transaction.Transaction,
	sellingBondsParam *SellingBonds,
) *transaction.TxCustomToken {
	if len(buySellReqTxs) == 0 {
		return nil
	}

	txTokenData := transaction.TxTokenData{
		Type:           transaction.CustomTokenInit,
		Amount:         0,
		PropertyName:   "",
		PropertySymbol: coinbaseTxType,
		Vins:           []transaction.TxTokenVin{},
	}
	var txTokenVouts []transaction.TxTokenVout
	for _, reqTx := range buySellReqTxs {
		tx, _ := reqTx.(*transaction.TxBuySellRequest)
		txTokenVout := buildSingleBuySellResponseTx(tx, sellingBondsParam)
		txTokenVouts = append(txTokenVouts, txTokenVout)
	}
	txTokenData.Vouts = txTokenVouts
	return &transaction.TxCustomToken{
		TxTokenData: txTokenData,
	}
}

func (blockgen *BlkTmplGenerator) checkBuyBackReqTx(
	chainID byte,
	tx transaction.Transaction,
	buyBackConsts uint64,
) (*transaction.TxTokenVout, *common.Hash, bool) {
	txBuyBackReq, ok := tx.(*transaction.TxBuyBackRequest)
	if !ok {
		return nil, nil, false
	}
	_, _, _, buyFromGOVReqTx, err := blockgen.chain.GetTransactionByHash(txBuyBackReq.BuyBackFromTxID)
	if err != nil {
		Logger.log.Error(err)
		return nil, nil, false
	}
	customTokenTx, ok := buyFromGOVReqTx.(*transaction.TxCustomToken)
	if !ok {
		return nil, nil, false
	}
	vout := customTokenTx.TxTokenData.Vouts[txBuyBackReq.VoutIndex]
	buyBackInfo := vout.BuySellResponse.BuyBackInfo
	if buyBackInfo == nil {
		Logger.log.Error("Missing buy-back info")
		return nil, nil, false
	}
	prevBlock := blockgen.chain.BestState[chainID].BestBlock

	if buyBackInfo.StartSellingAt+buyBackInfo.Maturity > uint32(prevBlock.Header.Height)+1 {
		Logger.log.Error("The token is not overdued yet.")
		return nil, nil, false
	}
	// check remaining constants in GOV fun is enough or not
	buyBackAmount := vout.Value * vout.BuySellResponse.BuyBackInfo.BuyBackPrice
	if buyBackConsts+buyBackAmount > prevBlock.Header.SalaryFund {
		return nil, nil, false
	}
	return &vout, tx.Hash(), true
}

func buildSingleTxTokenResVout(
	buyBackTxID *common.Hash,
	txTokenReqVout *transaction.TxTokenVout,
) transaction.TxTokenVout {
	buySellResponse := &transaction.BuySellResponse{
		BuyBackInfo:   nil,
		AssetID:       "",
		RequestedTxID: buyBackTxID,
	}
	buyBackAmount := txTokenReqVout.Value * txTokenReqVout.BuySellResponse.BuyBackInfo.BuyBackPrice
	return transaction.TxTokenVout{
		Value:           buyBackAmount,
		PaymentAddress:  txTokenReqVout.PaymentAddress,
		BuySellResponse: buySellResponse,
	}
}

// buildBuyBackResponsesTx
// the tx is to pay constants back to token buy back requesters
func buildBuyBackResponsesTx(
	coinbaseTxType string,
	txTokenReqVouts map[*common.Hash]*transaction.TxTokenVout,
) *transaction.TxCustomToken {
	if len(txTokenReqVouts) == 0 {
		return nil
	}

	txTokenData := transaction.TxTokenData{
		Type:           transaction.CustomTokenInit,
		Amount:         0,
		PropertyName:   "",
		PropertySymbol: coinbaseTxType,
		Vins:           []transaction.TxTokenVin{},
	}
	var txTokenResVouts []transaction.TxTokenVout
	for buyBackTxID, txTokenReqVout := range txTokenReqVouts {
		txTokenResVout := buildSingleTxTokenResVout(buyBackTxID, txTokenReqVout)
		txTokenResVouts = append(txTokenResVouts, txTokenResVout)
	}
	txTokenData.Vouts = txTokenResVouts
	return &transaction.TxCustomToken{
		TxTokenData: txTokenData,
	}
}

func calculateAmountOfRefundTxs(
	addresses []*privacy.PaymentAddress,
	estimatedRefundAmt uint64,
	remainingFund uint64,
	rt []byte,
	chainID byte,
) ([]*transaction.Tx, uint64) {
	amt := uint64(0)
	if estimatedRefundAmt <= remainingFund {
		amt = estimatedRefundAmt
	} else {
		amt = remainingFund
	}
	actualRefundAmt := amt / uint64(len(addresses))
	var refundTxs []*transaction.Tx
	for i := 0; i < len(addresses); i++ {
		addr := addresses[i]
		refundTx, err := transaction.CreateTxSalary(actualRefundAmt, addr, rt, chainID)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		refundTxs = append(refundTxs, refundTx)
	}
	return refundTxs, amt
}

func (blockgen *BlkTmplGenerator) buildRefundTxs(
	chainID byte,
	remainingFund uint64,
) ([]*transaction.Tx, uint64) {
	if remainingFund <= 0 {
		Logger.log.Info("GOV fund is not enough for refund.")
		return []*transaction.Tx{}, 0
	}
	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	header := prevBlock.Header
	govParams := header.GOVConstitution.GOVParams
	refundInfo := govParams.RefundInfo
	if refundInfo == nil {
		Logger.log.Info("Refund info is not existed.")
		return []*transaction.Tx{}, 0
	}
	lookbackBlockHeight := header.Height - common.RefundPeriod
	if lookbackBlockHeight < 0 {
		return []*transaction.Tx{}, 0
	}
	lookbackBlock, err := blockgen.chain.GetBlockByBlockHeight(lookbackBlockHeight, chainID)
	if err != nil {
		Logger.log.Error(err)
		return []*transaction.Tx{}, 0
	}
	var addresses []*privacy.PaymentAddress
	estimatedRefundAmt := uint64(0)
	for _, tx := range lookbackBlock.Transactions {
		if tx.GetType() != common.TxNormalType {
			continue
		}
		lookbackTx, ok := tx.(*transaction.Tx)
		if !ok {
			continue
		}
		addr, txValue := lookbackTx.CalculateTxValue()
		if addr == nil || txValue > refundInfo.ThresholdToLargeTx {
			continue
		}
		addresses = append(addresses, addr)
		estimatedRefundAmt += refundInfo.RefundAmount
	}
	refundTxs, totalRefundAmt := calculateAmountOfRefundTxs(
		addresses,
		estimatedRefundAmt,
		remainingFund,
		prevBlock.Header.MerkleRootCommitments.CloneBytes(),
		chainID,
	)
	return refundTxs, totalRefundAmt
}
