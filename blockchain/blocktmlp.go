package blockchain

import (
	"bytes"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
)

type BlkTmplGenerator struct {
	txPool      TxPool
	chain       *BlockChain
	rewardAgent RewardAgent
}

type ConstitutionHelper interface {
	GetStartedNormalVote(generator *BlkTmplGenerator, chainID byte) uint32
	CheckSubmitProposalType(tx metadata.Transaction) bool
	CheckVotingProposalType(tx metadata.Transaction) bool
	GetAmountVoteTokenOfTx(tx metadata.Transaction) uint64
	TxAcceptProposal(txId *common.Hash, voter metadata.Voter) metadata.Transaction
	GetLowerCaseBoardType() string
	GetConstitutionEndedBlockHeight(generator *BlkTmplGenerator, chainID byte) uint32
	CreatePunishDecryptTx([]byte) metadata.Metadata
	GetSealerPubKey(metadata.Transaction) [][]byte
	NewTxRewardProposalSubmitter(blockgen *BlkTmplGenerator, receiverAddress *privacy.PaymentAddress, minerPrivateKey *privacy.SpendingKey) (metadata.Transaction, error)
	GetPaymentAddressFromSubmitProposalMetadata(tx metadata.Transaction) *privacy.PaymentAddress
	GetPubKeyVoter(blockgen *BlkTmplGenerator, chainID byte) ([]byte, error)
	GetPrizeProposal() uint32
	GetTopMostVoteGovernor(blockgen *BlkTmplGenerator) (database.CandidateList, error)
	GetCurrentBoardPubKeys(blockgen *BlkTmplGenerator) [][]byte
	GetAmountVoteTokenOfBoard(blockgen *BlkTmplGenerator, pubKey []byte, boardIndex uint32) uint64
	GetBoardSumToken(blockgen *BlkTmplGenerator) uint64
	GetBoardFund(blockgen *BlkTmplGenerator) uint64
	GetTokenID() *common.Hash
	GetAmountOfVoteToBoard(blockgen *BlkTmplGenerator, candidatePubKey []byte, voterPubKey []byte, boardIndex uint32) uint64
	GetBoard(chain *BlockChain) Governor
	GetConstitutionInfo(chain *BlockChain) ConstitutionInfo
	GetCurrentNationalWelfare(chain *BlockChain) int32
	GetThresholdRatioOfCrisis() int32
	GetOldNationalWelfare(chain *BlockChain) int32
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
	MiningDescs() []*metadata.TxDesc

	// HaveTransaction returns whether or not the passed transaction hash
	// exists in the source pool.
	HaveTransaction(hash *common.Hash) bool

	// RemoveTx remove tx from tx resource
	RemoveTx(tx metadata.Transaction) error

	//CheckTransactionFee
	// CheckTransactionFee(tx metadata.Transaction) (uint64, error)

	// Check tx validate by it self
	// ValidateTxByItSelf(tx metadata.Transaction) bool
}

type RewardAgent interface {
	GetBasicSalary(chainID byte) uint64
	GetSalaryPerTx(chainID byte) uint64
}

type buyBackFromInfo struct {
	paymentAddress privacy.PaymentAddress
	buyBackPrice   uint64
	value          uint64
	requestedTxID  *common.Hash
}

func (self BlkTmplGenerator) Init(txPool TxPool, chain *BlockChain, rewardAgent RewardAgent) (*BlkTmplGenerator, error) {
	return &BlkTmplGenerator{
		txPool:      txPool,
		chain:       chain,
		rewardAgent: rewardAgent,
	}, nil
}

func (blockgen *BlkTmplGenerator) NewBlockTemplate(payToAddress *privacy.PaymentAddress, privatekey *privacy.SpendingKey, chainID byte) (*Block, error) {

	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	prevBlockHash := blockgen.chain.BestState[chainID].BestBlock.Hash()
	//prevCmTree := blockgen.chain.BestState[chainID].CmTree.MakeCopy()
	sourceTxns := blockgen.txPool.MiningDescs()

	var txsToAdd []metadata.Transaction
	var txToRemove []metadata.Transaction
	var buySellReqTxs []metadata.Transaction
	var issuingReqTxs []metadata.Transaction
	var updatingOracleBoardTxs []metadata.Transaction
	var multiSigsRegistrationTxs []metadata.Transaction
	var buyBackFromInfos []*buyBackFromInfo
	bondsSold := uint64(0)
	dcbTokensSold := uint64(0)
	incomeFromBonds := uint64(0)
	totalFee := uint64(0)
	buyBackCoins := uint64(0)

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
				// return nil, errors.Zero("No TxNormal")
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

		// TODO: 0xbunyip need to determine a tx is in privacy format or not
		if !tx.ValidateTxByItself(tx.IsPrivacy(), blockgen.chain.config.DataBase, blockgen.chain, chainID) {
			txToRemove = append(txToRemove, metadata.Transaction(tx))
			continue
		}

		meta := tx.GetMetadata()
		if meta != nil && !meta.ValidateBeforeNewBlock(tx, blockgen.chain, chainID) {
			txToRemove = append(txToRemove, metadata.Transaction(tx))
			continue
		}

		switch tx.GetMetadataType() {
		case metadata.BuyFromGOVRequestMeta:
			{
				income, soldAmt, addable := blockgen.checkBuyFromGOVReqTx(chainID, tx, bondsSold)
				if !addable {
					txToRemove = append(txToRemove, tx)
					continue
				}
				bondsSold += soldAmt
				incomeFromBonds += income
				buySellReqTxs = append(buySellReqTxs, tx)
			}
		case metadata.BuyBackRequestMeta:
			{
				buyBackFromInfo, addable := blockgen.checkBuyBackReqTx(chainID, tx, buyBackCoins)
				if !addable {
					txToRemove = append(txToRemove, tx)
					continue
				}
				buyBackCoins += (buyBackFromInfo.buyBackPrice + buyBackFromInfo.value)
				buyBackFromInfos = append(buyBackFromInfos, buyBackFromInfo)
			}
		case metadata.IssuingRequestMeta:
			{
				addable, newDCBTokensSold := blockgen.checkIssuingReqTx(chainID, tx, dcbTokensSold)
				dcbTokensSold = newDCBTokensSold
				if !addable {
					txToRemove = append(txToRemove, tx)
					continue
				}
				issuingReqTxs = append(issuingReqTxs, tx)
			}
		case metadata.UpdatingOracleBoardMeta:
			{
				updatingOracleBoardTxs = append(updatingOracleBoardTxs, tx)
			}
		case metadata.MultiSigsRegistrationMeta:
			{
				multiSigsRegistrationTxs = append(multiSigsRegistrationTxs, tx)
			}
		}

		totalFee += tx.GetTxFee()
		txsToAdd = append(txsToAdd, tx)
		if len(txsToAdd) == common.MaxTxsInBlock {
			break
		}
	}

	// check len of txs in block
	if len(txsToAdd) == 0 {
		// return nil, errors.Zero("no transaction available for this chain")
		Logger.log.Info("Creating empty block...")
	}

concludeBlock:
	// rt := prevBlock.Header.MerkleRootCommitments.CloneBytes()
	rt := []byte{}
	blockHeight := prevBlock.Header.Height + 1

	// TODO(@0xbunyip): cap #tx to common.MaxTxsInBlock
	// Process dividend payout for DCB if needed
	bankDivTxs, bankPayoutAmount, err := blockgen.processBankDividend(blockHeight, privatekey)
	if err != nil {
		return nil, err
	}
	for _, tx := range bankDivTxs {
		txsToAdd = append(txsToAdd, tx)
	}

	// Process dividend payout for GOV if needed
	govDivTxs, govPayoutAmount, err := blockgen.processGovDividend(blockHeight, privatekey)
	if err != nil {
		return nil, err
	}
	for _, tx := range govDivTxs {
		txsToAdd = append(txsToAdd, tx)
	}

	// Process crowdsale for DCB
	dcbSaleTxs, removableTxs, err := blockgen.processCrowdsale(sourceTxns, rt, chainID, privatekey)
	if err != nil {
		return nil, err
	}
	for _, tx := range dcbSaleTxs {
		txsToAdd = append(txsToAdd, tx)
	}
	for _, tx := range removableTxs {
		txToRemove = append(txToRemove, tx)
	}

	// Build CMB responses
	cmbInitRefundTxs, err := blockgen.buildCMBRefund(sourceTxns, chainID, privatekey)
	if err != nil {
		return nil, err
	}
	for _, tx := range cmbInitRefundTxs {
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
	salaryTx := new(transaction.Tx)
	err = salaryTx.InitTxSalary(totalSalary, payToAddress, privatekey, blockgen.chain.config.DataBase, nil)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	// create buy/sell response txs to distribute bonds/govs to requesters
	buySellResTxs, err := blockgen.buildBuySellResponsesTx(
		buySellReqTxs,
		blockgen.chain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.SellingBonds,
	)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	// create buy-back response txs to distribute constants to buy-back requesters
	buyBackResTxs, err := blockgen.buildBuyBackResponseTxs(buyBackFromInfos, chainID, privatekey)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	oracleRewardTxs, totalOracleRewards, updatedOracleValues, err := blockgen.buildOracleRewardTxs(chainID, privatekey)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	// create refund txs
	currentSalaryFund := prevBlock.Header.SalaryFund
	remainingFund := currentSalaryFund + totalFee + salaryFundAdd + incomeFromBonds - (totalSalary + buyBackCoins + totalOracleRewards)
	refundTxs, totalRefundAmt := blockgen.buildRefundTxs(chainID, remainingFund, privatekey)

	issuingResTxs, err := blockgen.buildIssuingResTxs(chainID, issuingReqTxs, privatekey)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	// Get loan payment amount to add to DCB fund
	loanPaymentAmount, unlockTxs, removableTxs := blockgen.processLoan(sourceTxns, privatekey)
	for _, tx := range removableTxs {
		txToRemove = append(txToRemove, tx)
	}

	coinbases := []metadata.Transaction{salaryTx}
	// Voting transaction
	// Check if it is the case we need to apply a new proposal
	// 1. newNW < lastNW * 0.9
	// 2. current block height == last Constitution start time + last Constitution execute duration
	if blockgen.chain.readyNewConstitution(DCBConstitutionHelper{}) {
		blockgen.chain.config.DataBase.SetEncryptionLastBlockHeight("dcb", uint32(prevBlock.Header.Height+1))
		blockgen.chain.config.DataBase.SetEncryptFlag("dcb", uint32(common.Lv3EncryptionFlag))
		tx, err := blockgen.createAcceptConstitutionAndPunishTxAndRewardSubmitter(chainID, DCBConstitutionHelper{}, privatekey)
		coinbases = append(coinbases, tx...)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		rewardTx, err := blockgen.createRewardProposalWinnerTx(chainID, DCBConstitutionHelper{})
		coinbases = append(coinbases, rewardTx)
	}
	if blockgen.chain.readyNewConstitution(GOVConstitutionHelper{}) {
		blockgen.chain.config.DataBase.SetEncryptionLastBlockHeight("gov", uint32(prevBlock.Header.Height+1))
		blockgen.chain.config.DataBase.SetEncryptFlag("gov", uint32(common.Lv3EncryptionFlag))
		tx, err := blockgen.createAcceptConstitutionAndPunishTxAndRewardSubmitter(chainID, GOVConstitutionHelper{}, privatekey)
		coinbases = append(coinbases, tx...)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
		rewardTx, err := blockgen.createRewardProposalWinnerTx(chainID, GOVConstitutionHelper{})
		coinbases = append(coinbases, rewardTx)
	}

	if blockgen.neededNewDCBGovernor(chainID) {
		coinbases = append(coinbases, blockgen.UpdateNewGovernor(DCBConstitutionHelper{}, chainID, privatekey)...)
	}
	if blockgen.neededNewGOVGovernor(chainID) {
		coinbases = append(coinbases, blockgen.UpdateNewGovernor(GOVConstitutionHelper{}, chainID, privatekey)...)
	}

	for _, tx := range unlockTxs {
		coinbases = append(coinbases, tx)
	}
	for _, resTx := range buySellResTxs {
		coinbases = append(coinbases, resTx)
	}
	for _, resTx := range buyBackResTxs {
		coinbases = append(coinbases, resTx)
	}
	for _, resTx := range issuingResTxs {
		coinbases = append(coinbases, resTx)
	}
	for _, refundTx := range refundTxs {
		coinbases = append(coinbases, refundTx)
	}
	for _, oracleRewardTx := range oracleRewardTxs {
		coinbases = append(coinbases, oracleRewardTx)
	}

	txsToAdd = append(coinbases, txsToAdd...)

	for _, tx := range txToRemove {
		blockgen.txPool.RemoveTx(tx)
	}

	// Check for final balance of DCB and GOV
	if currentSalaryFund+totalFee+salaryFundAdd+incomeFromBonds < totalSalary+govPayoutAmount+buyBackCoins+totalRefundAmt+totalOracleRewards {
		return nil, fmt.Errorf("Gov fund is not enough for salary and dividend payout")
	}

	currentBankFund := prevBlock.Header.BankFund
	if currentBankFund < bankPayoutAmount { // Can't spend loan payment just received in this block
		return nil, fmt.Errorf("Bank fund is not enough for dividend payout")
	}

	merkleRoots := Merkle{}.BuildMerkleTreeStore(txsToAdd)
	merkleRoot := merkleRoots[len(merkleRoots)-1]

	block := Block{
		Transactions: make([]metadata.Transaction, 0),
	}

	block.Header = BlockHeader{
		Height:        prevBlock.Header.Height + 1,
		Version:       BlockVersion,
		PrevBlockHash: *prevBlockHash,
		MerkleRoot:    *merkleRoot,
		// MerkleRootCommitments: common.Hash{},
		Timestamp:          time.Now().Unix(),
		BlockCommitteeSigs: make([]string, common.TotalValidators),
		Committee:          make([]string, common.TotalValidators),
		ChainID:            chainID,
		SalaryFund:         currentSalaryFund + incomeFromBonds + totalFee + salaryFundAdd - totalSalary - govPayoutAmount - buyBackCoins - totalRefundAmt - totalOracleRewards,
		BankFund:           prevBlock.Header.BankFund + loanPaymentAmount - bankPayoutAmount,
		GOVConstitution:    prevBlock.Header.GOVConstitution, // TODO: 0xbunyip need get from gov-params tx
		DCBConstitution:    prevBlock.Header.DCBConstitution, // TODO: 0xbunyip need get from dcb-params tx
		Oracle:             prevBlock.Header.Oracle,
	}
	if block.Header.GOVConstitution.GOVParams.SellingBonds != nil {
		block.Header.GOVConstitution.GOVParams.SellingBonds.BondsToSell -= bondsSold
	}
	if block.Header.DCBConstitution.DCBParams.SaleDBCTOkensByUSDData != nil {
		block.Header.DCBConstitution.DCBParams.SaleDBCTOkensByUSDData.Amount -= dcbTokensSold
	}

	blockgen.updateOracleValues(&block, updatedOracleValues)
	err = blockgen.updateOracleBoard(&block, updatingOracleBoardTxs)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
		// Handle if this transaction change something in block header or database
		if tx.GetMetadataType() == metadata.AcceptDCBProposalMeta {
			block.updateDCBConstitution(tx, blockgen)
		}
		if tx.GetMetadataType() == metadata.AcceptGOVProposalMeta {
			block.updateGOVConstitution(tx, blockgen)
		}
		if tx.GetMetadataType() == metadata.AcceptDCBBoardMeta {
			block.UpdateDCBBoard(tx)
		}
		if tx.GetMetadataType() == metadata.AcceptGOVBoardMeta {
			block.UpdateGOVBoard(tx)
		}
		if tx.GetMetadataType() == metadata.RewardDCBProposalSubmitterMeta {
			block.UpdateDCBFund(tx)
		}
		if tx.GetMetadataType() == metadata.RewardGOVProposalSubmitterMeta {
			block.UpdateGOVFund(tx)
		}
	}

	// register multisigs addresses
	err = blockgen.registerMultiSigsAddresses(multiSigsRegistrationTxs)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	// Add new commitments to merkle tree and save the root
	/*newTree := prevCmTree
	err = UpdateMerkleTreeForBlock(newTree, &block)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	rt = newTree.GetRoot(common.IncMerkleTreeHeight)
	copy(block.Header.MerkleRootCommitments[:], rt)*/

	return &block, nil
}

func GetOracleDCBNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!")
	return 1234
}
func GetOracleGOVNationalWelfare() int32 {
	fmt.Print("Get national welfare. It is constant now. Need to change !!!")
	return 1234
}

func (blockgen *BlkTmplGenerator) neededNewDCBGovernor(chainID byte) bool {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	return int32(BestBlock.Header.DCBGovernor.EndBlock) == BestBlock.Header.Height+2
}
func (blockgen *BlkTmplGenerator) neededNewGOVGovernor(chainID byte) bool {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	return int32(BestBlock.Header.GOVGovernor.EndBlock) == BestBlock.Header.Height+2
}

func (blockgen *BlkTmplGenerator) processDividend(
	proposal *metadata.DividendProposal,
	blockHeight int32,
	producerPrivateKey *privacy.SpendingKey,
) ([]*transaction.Tx, uint64, error) {
	payoutAmount := uint64(0)
	// TODO(@0xbunyip): how to execute payout dividend proposal
	dividendTxs := []*transaction.Tx{}
	if false && blockHeight%metadata.PayoutFrequency == 0 { // only chain 0 process dividend proposals
		totalTokenSupply, tokenHolders, amounts, err := blockgen.chain.GetAmountPerAccount(proposal)
		if err != nil {
			return nil, 0, err
		}

		infos := []metadata.DividendInfo{}
		// Build tx to pay dividend to each holder
		for i, holder := range tokenHolders {
			holderAddrInBytes, _, err := base58.Base58Check{}.Decode(holder)
			if err != nil {
				return nil, 0, err
			}
			holderAddress := (&privacy.PaymentAddress{}).SetBytes(holderAddrInBytes)
			info := metadata.DividendInfo{
				TokenHolder: *holderAddress,
				Amount:      amounts[i] / totalTokenSupply,
			}
			payoutAmount += info.Amount
			infos = append(infos, info)

			if len(infos) > metadata.MaxDivTxsPerBlock {
				break // Pay dividend to only some token holders in this block
			}
		}

		dividendTxs, err = buildDividendTxs(infos, proposal, producerPrivateKey, blockgen.chain.GetDatabase())
		if err != nil {
			return nil, 0, err
		}
	}
	return dividendTxs, payoutAmount, nil
}

func (blockgen *BlkTmplGenerator) processBankDividend(blockHeight int32, producerPrivateKey *privacy.SpendingKey) ([]*transaction.Tx, uint64, error) {
	tokenID, _ := (&common.Hash{}).NewHash(common.DCBTokenID[:])
	proposal := &metadata.DividendProposal{
		TokenID: tokenID,
	}
	return blockgen.processDividend(proposal, blockHeight, producerPrivateKey)
}

func (blockgen *BlkTmplGenerator) processGovDividend(blockHeight int32, producerPrivateKey *privacy.SpendingKey) ([]*transaction.Tx, uint64, error) {
	tokenID, _ := (&common.Hash{}).NewHash(common.GOVTokenID[:])
	proposal := &metadata.DividendProposal{
		TokenID: tokenID,
	}
	return blockgen.processDividend(proposal, blockHeight, producerPrivateKey)
}

func buildSingleBuySellResponseTx(
	buySellReqTx metadata.Transaction,
	sellingBondsParam *params.SellingBonds,
) (*transaction.TxCustomToken, error) {
	bondID := sellingBondsParam.GetID()
	buySellRes := metadata.NewBuySellResponse(
		*buySellReqTx.Hash(),
		sellingBondsParam.StartSellingAt,
		sellingBondsParam.Maturity,
		sellingBondsParam.BuyBackPrice,
		bondID[:],
		metadata.BuyFromGOVResponseMeta,
	)

	buySellReqMeta := buySellReqTx.GetMetadata()
	buySellReq, ok := buySellReqMeta.(*metadata.BuySellRequest)
	if !ok {
		return nil, errors.New("Could not assert BuySellRequest metadata.")
	}
	txTokenVout := transaction.TxTokenVout{
		Value:          buySellReq.Amount,
		PaymentAddress: buySellReq.PaymentAddress,
	}

	var propertyID [common.HashSize]byte
	copy(propertyID[:], bondID[:])
	txTokenData := transaction.TxTokenData{
		Type:       transaction.CustomTokenInit,
		Mintable:   true,
		Amount:     buySellReq.Amount,
		PropertyID: common.Hash(propertyID),
		Vins:       []transaction.TxTokenVin{},
		Vouts:      []transaction.TxTokenVout{txTokenVout},
	}
	txTokenData.PropertyName = txTokenData.PropertyID.String()
	txTokenData.PropertySymbol = txTokenData.PropertyID.String()
	resTx := &transaction.TxCustomToken{
		TxTokenData: txTokenData,
	}
	resTx.Type = common.TxCustomTokenType
	resTx.SetMetadata(buySellRes)
	return resTx, nil
}

func (blockgen *BlkTmplGenerator) checkBuyFromGOVReqTx(
	chainID byte,
	tx metadata.Transaction,
	bondsSold uint64,
) (uint64, uint64, bool) {
	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	sellingBondsParams := prevBlock.Header.GOVConstitution.GOVParams.SellingBonds
	if uint32(prevBlock.Header.Height)+1 > sellingBondsParams.StartSellingAt+sellingBondsParams.SellingWithin {
		return 0, 0, false
	}

	buySellReqMeta := tx.GetMetadata()
	req, ok := buySellReqMeta.(*metadata.BuySellRequest)
	if !ok {
		return 0, 0, false
	}

	if bondsSold+req.Amount > sellingBondsParams.BondsToSell { // run out of bonds for selling
		return 0, 0, false
	}
	return req.Amount * req.BuyPrice, req.Amount, true
}

// buildBuySellResponsesTx
// the tx is to distribute tokens (bond, gov, ...) to token requesters
func (blockgen *BlkTmplGenerator) buildBuySellResponsesTx(
	buySellReqTxs []metadata.Transaction,
	sellingBondsParam *params.SellingBonds,
) ([]*transaction.TxCustomToken, error) {
	if len(buySellReqTxs) == 0 {
		return []*transaction.TxCustomToken{}, nil
	}
	var resTxs []*transaction.TxCustomToken
	for _, reqTx := range buySellReqTxs {
		resTx, err := buildSingleBuySellResponseTx(reqTx, sellingBondsParam)
		if err != nil {
			return []*transaction.TxCustomToken{}, err
		}
		resTxs = append(resTxs, resTx)
	}
	return resTxs, nil
}

func (blockgen *BlkTmplGenerator) checkBuyBackReqTx(
	chainID byte,
	tx metadata.Transaction,
	buyBackConsts uint64,
) (*buyBackFromInfo, bool) {
	buyBackReqTx, ok := tx.(*transaction.TxCustomToken)
	if !ok {
		Logger.log.Error(errors.New("Could not parse BuyBackRequest tx (custom token tx)."))
		return nil, false
	}
	vins := buyBackReqTx.TxTokenData.Vins
	if len(vins) == 0 {
		Logger.log.Error(errors.New("No existed Vins from BuyBackRequest tx"))
		return nil, false
	}
	priorTxID := vins[0].TxCustomTokenID
	_, _, _, priorTx, err := blockgen.chain.GetTransactionByHash(&priorTxID)
	if err != nil {
		Logger.log.Error(err)
		return nil, false
	}
	priorCustomTokenTx, ok := priorTx.(*transaction.TxCustomToken)
	if !ok {
		Logger.log.Error(errors.New("Could not parse prior TxCustomToken."))
		return nil, false
	}

	priorMeta := priorCustomTokenTx.GetMetadata()
	if priorMeta == nil {
		Logger.log.Error(errors.New("No existed metadata in priorCustomTokenTx"))
		return nil, false
	}
	buySellResMeta, ok := priorMeta.(*metadata.BuySellResponse)
	if !ok {
		Logger.log.Error(errors.New("Could not parse BuySellResponse metadata."))
		return nil, false
	}
	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	if buySellResMeta.StartSellingAt+buySellResMeta.Maturity > uint32(prevBlock.Header.Height)+1 {
		Logger.log.Error("The token is not overdued yet.")
		return nil, false
	}
	// check remaining constants in GOV fund is enough or not
	buyBackReqMeta := buyBackReqTx.GetMetadata()
	buyBackReq, ok := buyBackReqMeta.(*metadata.BuyBackRequest)
	if !ok {
		Logger.log.Error(errors.New("Could not parse BuyBackRequest metadata."))
		return nil, false
	}
	buyBackValue := buyBackReq.Amount * buySellResMeta.BuyBackPrice
	if buyBackConsts+buyBackValue > prevBlock.Header.SalaryFund {
		return nil, false
	}
	buyBackFromInfo := &buyBackFromInfo{
		paymentAddress: buyBackReq.PaymentAddress,
		buyBackPrice:   buySellResMeta.BuyBackPrice,
		value:          buyBackReq.Amount,
		requestedTxID:  tx.Hash(),
	}
	return buyBackFromInfo, true
}

func (blockgen *BlkTmplGenerator) buildBuyBackResponseTxs(
	buyBackFromInfos []*buyBackFromInfo,
	chainID byte,
	privatekey *privacy.SpendingKey,
) ([]*transaction.Tx, error) {
	if len(buyBackFromInfos) == 0 {
		return []*transaction.Tx{}, nil
	}

	// prevBlock := blockgen.chain.BestState[chainID].BestBlock
	var buyBackResTxs []*transaction.Tx
	for _, buyBackFromInfo := range buyBackFromInfos {
		buyBackAmount := buyBackFromInfo.value * buyBackFromInfo.buyBackPrice
		buyBackRes := metadata.NewBuyBackResponse(*buyBackFromInfo.requestedTxID, metadata.BuyBackResponseMeta)
		buyBackResTx := new(transaction.Tx)
		err := buyBackResTx.InitTxSalary(buyBackAmount, &buyBackFromInfo.paymentAddress, privatekey, blockgen.chain.GetDatabase(), buyBackRes)
		if err != nil {
			return []*transaction.Tx{}, err
		}
		buyBackResTxs = append(buyBackResTxs, buyBackResTx)
	}
	return buyBackResTxs, nil
}

func (blockgen *BlkTmplGenerator) checkIssuingReqTx(
	chainID byte,
	tx metadata.Transaction,
	dcbTokensSold uint64,
) (bool, uint64) {
	issuingReqMeta := tx.GetMetadata()
	issuingReq, ok := issuingReqMeta.(*metadata.IssuingRequest)
	if !ok {
		Logger.log.Error(errors.New("Could not parse IssuingRequest metadata"))
		return false, dcbTokensSold
	}
	if !bytes.Equal(issuingReq.AssetType[:], common.DCBTokenID[:]) {
		return true, dcbTokensSold
	}
	header := blockgen.chain.BestState[chainID].BestBlock.Header
	saleDBCTOkensByUSDData := header.DCBConstitution.DCBParams.SaleDBCTOkensByUSDData
	oracleParams := header.Oracle
	dcbTokenPrice := uint64(1)
	if oracleParams.DCBToken != 0 {
		dcbTokenPrice = oracleParams.DCBToken
	}
	dcbTokensReq := issuingReq.DepositedAmount / dcbTokenPrice
	if dcbTokensSold+dcbTokensReq > saleDBCTOkensByUSDData.Amount {
		return false, dcbTokensSold
	}
	return true, dcbTokensSold + dcbTokensReq
}

func (blockgen *BlkTmplGenerator) buildIssuingResTxs(
	chainID byte,
	issuingReqTxs []metadata.Transaction,
	privatekey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	oracleParams := prevBlock.Header.Oracle

	issuingResTxs := []metadata.Transaction{}
	for _, issuingReqTx := range issuingReqTxs {
		meta := issuingReqTx.GetMetadata()
		issuingReq, ok := meta.(*metadata.IssuingRequest)
		if !ok {
			return []metadata.Transaction{}, errors.New("Could not parse IssuingRequest metadata.")
		}
		if issuingReq.AssetType == common.DCBTokenID {
			issuingRes := metadata.NewIssuingResponse(*issuingReqTx.Hash(), metadata.IssuingResponseMeta)
			dcbTokenPrice := uint64(1)
			if oracleParams.DCBToken != 0 {
				dcbTokenPrice = oracleParams.DCBToken
			}
			issuingAmt := issuingReq.DepositedAmount / dcbTokenPrice
			txTokenVout := transaction.TxTokenVout{
				Value:          issuingAmt,
				PaymentAddress: issuingReq.ReceiverAddress,
			}
			txTokenData := transaction.TxTokenData{
				Type:       transaction.CustomTokenInit,
				Amount:     issuingAmt,
				PropertyID: common.Hash(common.DCBTokenID),
				Vins:       []transaction.TxTokenVin{},
				Vouts:      []transaction.TxTokenVout{txTokenVout},
				// PropertyName:   "",
				// PropertySymbol: coinbaseTxType,
			}
			resTx := &transaction.TxCustomToken{
				TxTokenData: txTokenData,
			}
			resTx.Type = common.TxCustomTokenType
			resTx.SetMetadata(issuingRes)
			issuingResTxs = append(issuingResTxs, resTx)
			continue
		}
		if issuingReq.AssetType == common.ConstantID {
			constantPrice := uint64(1)
			if oracleParams.Constant != 0 {
				constantPrice = oracleParams.Constant
			}
			issuingAmt := issuingReq.DepositedAmount / constantPrice
			issuingRes := metadata.NewIssuingResponse(*issuingReqTx.Hash(), metadata.IssuingResponseMeta)
			resTx := new(transaction.Tx)
			err := resTx.InitTxSalary(issuingAmt, &issuingReq.ReceiverAddress, privatekey, blockgen.chain.GetDatabase(), issuingRes)
			if err != nil {
				return []metadata.Transaction{}, err
			}
			issuingResTxs = append(issuingResTxs, resTx)
		}
	}
	return issuingResTxs, nil
}

func calculateAmountOfRefundTxs(
	smallTxHashes []*common.Hash,
	addresses []*privacy.PaymentAddress,
	estimatedRefundAmt uint64,
	remainingFund uint64,
	db database.DatabaseInterface,
	privatekey *privacy.SpendingKey,
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
		refundMeta := metadata.NewRefund(*smallTxHashes[i], metadata.RefundMeta)
		refundTx := new(transaction.Tx)
		err := refundTx.InitTxSalary(actualRefundAmt, addr, privatekey, db, refundMeta)
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
	privatekey *privacy.SpendingKey,
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
	addresses := []*privacy.PaymentAddress{}
	smallTxHashes := []*common.Hash{}
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
		smallTxHashes = append(smallTxHashes, tx.Hash())
		estimatedRefundAmt += refundInfo.RefundAmount
	}
	if len(addresses) == 0 {
		return []*transaction.Tx{}, 0
	}
	refundTxs, totalRefundAmt := calculateAmountOfRefundTxs(
		smallTxHashes,
		addresses,
		estimatedRefundAmt,
		remainingFund,
		blockgen.chain.GetDatabase(),
		privatekey,
	)
	return refundTxs, totalRefundAmt
}

func (blockgen *BlkTmplGenerator) processLoan(sourceTxns []*metadata.TxDesc, producerPrivateKey *privacy.SpendingKey) (uint64, []*transaction.Tx, []metadata.Transaction) {
	amount := uint64(0)
	loanUnlockTxs := []*transaction.Tx{}
	removableTxs := []metadata.Transaction{}
	for _, txDesc := range sourceTxns {
		if txDesc.Tx.GetMetadataType() == metadata.LoanPaymentMeta {
			paymentMeta := txDesc.Tx.GetMetadata().(*metadata.LoanPayment)
			_, _, _, err := blockgen.chain.config.DataBase.GetLoanPayment(paymentMeta.LoanID)
			if err != nil {
				removableTxs = append(removableTxs, txDesc.Tx)
				continue
			}
			paymentAmount := uint64(0)
			accountDCB, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
			dcbPk := accountDCB.KeySet.PaymentAddress.Pk
			txNormal := txDesc.Tx.(*transaction.Tx)
			for _, coin := range txNormal.Proof.OutputCoins {
				if bytes.Equal(coin.CoinDetails.PublicKey.Compress(), dcbPk) {
					paymentAmount += coin.CoinDetails.Value
				}
			}
			if !paymentMeta.PayPrinciple { // Only keep interest
				amount += paymentAmount
			}
		} else if txDesc.Tx.GetMetadataType() == metadata.LoanWithdrawMeta {
			withdrawMeta := txDesc.Tx.GetMetadata().(*metadata.LoanWithdraw)
			meta, err := blockgen.chain.GetLoanRequestMeta(withdrawMeta.LoanID)
			if err != nil {
				removableTxs = append(removableTxs, txDesc.Tx)
				continue
			}

			unlockMeta := &metadata.LoanUnlock{
				LoanID:       make([]byte, len(withdrawMeta.LoanID)),
				MetadataBase: metadata.MetadataBase{Type: metadata.LoanUnlockMeta},
			}
			copy(unlockMeta.LoanID, withdrawMeta.LoanID)
			uplockMetaList := []metadata.Metadata{unlockMeta}
			pks := [][]byte{meta.ReceiveAddress.Pk[:]}
			tks := [][]byte{meta.ReceiveAddress.Tk[:]}
			amounts := []uint64{meta.LoanAmount}
			txNormals, err := buildCoinbaseTxs(pks, tks, amounts, producerPrivateKey, blockgen.chain.GetDatabase(), uplockMetaList)
			if err != nil {
				removableTxs = append(removableTxs, txDesc.Tx)
				continue
			}
			loanUnlockTxs = append(loanUnlockTxs, txNormals[0]) // There's only one tx
		}
	}
	return amount, loanUnlockTxs, removableTxs
}

func (blockgen *BlkTmplGenerator) UpdateNewGovernor(helper ConstitutionHelper, chainID byte, minerPrivateKey *privacy.SpendingKey) []metadata.Transaction {
	txs := make([]metadata.Transaction, 0)
	newBoardList, _ := helper.GetTopMostVoteGovernor(blockgen)
	sort.Sort(newBoardList)
	sumOfVote := uint64(0)
	var newDCBBoardPubKey [][]byte
	for _, i := range newBoardList {
		newDCBBoardPubKey = append(newDCBBoardPubKey, i.PubKey)
		sumOfVote += i.VoteAmount
	}

	txs = append(txs, blockgen.createAcceptDCBBoardTx(newDCBBoardPubKey, sumOfVote))
	txs = append(txs, blockgen.CreateSendDCBVoteTokenToGovernorTx(chainID, newBoardList, sumOfVote)...)

	txs = append(txs, blockgen.CreateSendBackTokenAfterVoteFail(helper.GetLowerCaseBoardType(), chainID, newDCBBoardPubKey)...)

	txs = append(txs, blockgen.CreateSendRewardOldBoard(helper, minerPrivateKey)...)

	return txs
}

func (blockgen *BlkTmplGenerator) CreateSingleShareRewardOldBoard(
	helper ConstitutionHelper,
	chairPubKey []byte,
	voterPubKey []byte,
	amountOfCoin uint64,
	amountOfToken uint64,
	minerPrivateKey *privacy.SpendingKey,
) metadata.Transaction {
	paymentAddressByte := blockgen.chain.config.DataBase.GetPaymentAddressFromPubKey(voterPubKey)
	paymentAddress := privacy.PaymentAddress{}
	paymentAddress.SetBytes(paymentAddressByte)
	tx := transaction.Tx{}
	rewardShareOldBoardMeta := metadata.NewRewardShareOldBoardMetadata(chairPubKey, voterPubKey, helper.GetLowerCaseBoardType())
	tx.InitTxSalary(amountOfCoin, &paymentAddress, minerPrivateKey, blockgen.chain.config.DataBase, rewardShareOldBoardMeta)
	txTokenData := transaction.TxTokenData{
		Type:       transaction.CustomTokenInit,
		Amount:     amountOfToken,
		PropertyID: *helper.GetTokenID(),
		Vins:       []transaction.TxTokenVin{},
		Vouts:      []transaction.TxTokenVout{},
	}

	txCustomToken := transaction.TxCustomToken{
		Tx:          tx,
		TxTokenData: txTokenData,
	}
	return &txCustomToken
}

func (blockgen *BlkTmplGenerator) CreateShareRewardOldBoard(
	helper ConstitutionHelper,
	chairPubKey []byte,
	totalAmountCoinReward uint64,
	totalAmountTokenReward uint64,
	totalVoteAmount uint64,
	minerPrivateKey *privacy.SpendingKey,
) []metadata.Transaction {
	txs := make([]metadata.Transaction, 0)

	voterList := blockgen.chain.config.DataBase.GetBoardVoterList(helper.GetLowerCaseBoardType(), chairPubKey, blockgen.chain.GetCurrentBoardIndex(helper))
	boardIndex := blockgen.chain.GetCurrentBoardIndex(helper)
	for _, pubKey := range voterList {
		amountOfVote := helper.GetAmountOfVoteToBoard(blockgen, chairPubKey, pubKey, boardIndex)
		amountOfCoin := amountOfVote * totalAmountCoinReward / totalVoteAmount
		amountOfToken := amountOfVote * totalAmountTokenReward / totalVoteAmount
		blockgen.CreateSingleShareRewardOldBoard(helper, chairPubKey, pubKey, amountOfCoin, amountOfToken, minerPrivateKey)
	}
	return txs
}

func (blockgen *BlkTmplGenerator) GetCoinTermReward(helper ConstitutionHelper) uint64 {
	return helper.GetBoardFund(blockgen) * common.PercentageBoardSalary / common.BasePercentage
}

func (blockgen *BlkTmplGenerator) CreateSendRewardOldBoard(helper ConstitutionHelper, minerPrivateKey *privacy.SpendingKey) []metadata.Transaction {
	txs := make([]metadata.Transaction, 0)
	voteTokenAmount := make(map[string]uint64)
	sumVoteTokenAmount := uint64(0)
	pubKeys := helper.GetCurrentBoardPubKeys(blockgen)
	totalAmountOfTokenReward := helper.GetBoardSumToken(blockgen) // Total amount of token
	totalAmountOfCoinReward := blockgen.GetCoinTermReward(helper)
	helper.GetBoardFund(blockgen)
	//reward for each by voteDCBList
	for _, i := range pubKeys {
		amount := helper.GetAmountVoteTokenOfBoard(blockgen, i, blockgen.chain.GetCurrentBoardIndex(helper))
		voteTokenAmount[string(i)] = amount
		sumVoteTokenAmount += amount
	}
	for pubKey, voteAmount := range voteTokenAmount {
		percentageReward := voteAmount * common.BasePercentage / sumVoteTokenAmount
		amountTokenReward := totalAmountOfTokenReward * uint64(percentageReward) / common.BasePercentage
		amountCoinReward := totalAmountOfCoinReward * uint64(percentageReward) / common.BasePercentage
		txs = append(txs, blockgen.CreateShareRewardOldBoard(helper, []byte(pubKey), amountCoinReward, amountTokenReward, voteAmount, minerPrivateKey)...)
		//todo @0xjackalope: reward for chair
	}
	return txs
}
