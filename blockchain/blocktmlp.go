package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/privacy-protocol/client"
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
	// var actionParamTxs []*transaction.ActionParamTx
	totalFee := uint64(0)

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
	var buySellReqTxs []*transaction.BuySellRequestTx
	for _, blockTx := range txsToAdd {
		if blockTx.GetType() == common.TxBuyFromGOVRequest {
			buySellReqTx, ok := blockTx.(*transaction.BuySellRequestTx)
			if !ok {
				Logger.log.Error("Transaction not recognized to store in database")
				continue
			}
			buySellReqTxs = append(buySellReqTxs, buySellReqTx)
		}
		// if blockTx.GetType() == common.TxRegisterCandidateType {
		// 	tx, ok := blockTx.(*transaction.TxRegisterCandidate)
		// 	if !ok {
		// 		Logger.log.Error("Transaction not recognized to store in database")
		// 		continue
		// 	}
		// 	salaryFundAdd += tx.GetValue()
		// }
		if blockTx.GetTxFee() > 0 {
			salaryMULTP++
		}
	}

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
	buySellResTx := buildBuySellResponsesTx(
		common.TxBuyFromGOVResponse,
		buySellReqTxs,
		blockgen.chain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.SellingBonds,
	)

	// the 1st tx will be salaryTx
	txsToAdd = append([]transaction.Transaction{salaryTx, buySellResTx}, txsToAdd...)

	// Check for final balance of DCB and GOV
	currentSalaryFund := prevBlock.Header.SalaryFund
	if currentSalaryFund < totalSalary+totalFee+salaryFundAdd-govPayoutAmount {
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
		SalaryFund:            currentSalaryFund - totalSalary + totalFee + salaryFundAdd,
		BankFund:              prevBlock.Header.BankFund - bankPayoutAmount,
		GOVConstitution:       prevBlock.Header.GOVConstitution, // TODO: need get from gov-params tx
		DCBConstitution:       prevBlock.Header.DCBConstitution, // TODO: need get from dcb-params tx
		LoanParams:            prevBlock.Header.LoanParams,
	}
	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
		if tx.GetType() == common.TxAcceptDCBProposal {
			updateDCBConstitution(&block, tx, blockgen)
		}
		if tx.GetType() == common.TxAcceptGOVProposal {
			updateGOVConstitution(&block, tx, blockgen)
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

// createSalaryTx
// Blockchain use this tx to pay a reward(salary) to miner of chain
// #1 - salary:
// #2 - receiverAddr:
// #3 - rt
// #4 - chainID
func createSalaryTx(
	salary uint64,
	receiverAddr *privacy.PaymentAddress,
	rt []byte,
	chainID byte,
) (*transaction.Tx, error) {
	// Create Proof for the joinsplit op
	inputs := make([]*client.JSInput, 2)
	inputs[0] = transaction.CreateRandomJSInput(nil)
	inputs[1] = transaction.CreateRandomJSInput(inputs[0].Key)
	dummyAddress := client.GenPaymentAddress(*inputs[0].Key)

	// Create new notes: first one is salary UTXO, second one has 0 value
	outNote := &client.Note{Value: salary, Apk: receiverAddr.Pk}
	placeHolderOutputNote := &client.Note{Value: 0, Apk: receiverAddr.Pk}

	outputs := []*client.JSOutput{&client.JSOutput{}, &client.JSOutput{}}
	outputs[0].EncKey = receiverAddr.Tk
	outputs[0].OutputNote = outNote
	outputs[1].EncKey = receiverAddr.Tk
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
	err = tx.BuildNewJSDesc(inputMap, outputs, rtMap, salary, 0, true)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	err = tx.SignTx()
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	return tx, nil
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
	buySellReqTx *transaction.BuySellRequestTx,
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

// buildBuySellResponsesTx
// the tx is to distribute tokens (bond, gov, ...) to token requesters
func buildBuySellResponsesTx(
	coinbaseTxType string,
	buySellReqTxs []*transaction.BuySellRequestTx,
	sellingBondsParam *SellingBonds,
) *transaction.TxCustomToken {
	txTokenData := transaction.TxTokenData{
		Type:           transaction.CustomTokenInit,
		Amount:         0,
		PropertyName:   "",
		PropertySymbol: coinbaseTxType,
		Vins:           []transaction.TxTokenVin{},
	}
	var txTokenVouts []transaction.TxTokenVout
	for _, reqTx := range buySellReqTxs {
		txTokenVout := buildSingleBuySellResponseTx(reqTx, sellingBondsParam)
		txTokenVouts = append(txTokenVouts, txTokenVout)
	}
	txTokenData.Vouts = txTokenVouts
	return &transaction.TxCustomToken{
		TxTokenData: txTokenData,
	}
}
