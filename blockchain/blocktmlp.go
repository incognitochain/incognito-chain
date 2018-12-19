package blockchain

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/voting"
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
	GetAmountVoteToken(tx metadata.Transaction) uint64
	TxAcceptProposal(txId *common.Hash) metadata.Transaction
	GetLowerCaseBoardType() string
	GetEndedBlockHeight(generator *BlkTmplGenerator, chainID byte) uint32
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
	currentBlockHeight := uint32(prevBlock.Header.Height + 1)

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

		// TODO: need to determine a tx is in privacy format or not
		if !tx.ValidateTxByItself(tx.IsPrivacy(), blockgen.chain.config.DataBase, blockgen.chain, chainID) {
			txToRemove = append(txToRemove, metadata.Transaction(tx))
			continue
		}

		startedDCBPivot := prevBlock.Header.DCBConstitution.StartedBlockHeight
		endedDCBPivot := prevBlock.Header.DCBConstitution.GetEndedBlockHeight()
		lv3DCBPivot := endedDCBPivot - common.EncryptionPhaseDuration
		lv2DCBPivot := lv3DCBPivot - common.EncryptionPhaseDuration
		lv1DCBPivot := lv2DCBPivot - common.EncryptionPhaseDuration
		startedGOVPivot := prevBlock.Header.GOVConstitution.StartedBlockHeight
		endedGOVPivot := prevBlock.Header.GOVConstitution.GetEndedBlockHeight()
		lv3GOVPivot := endedGOVPivot - common.EncryptionPhaseDuration
		lv2GOVPivot := lv3GOVPivot - common.EncryptionPhaseDuration
		lv1GOVPivot := lv2GOVPivot - common.EncryptionPhaseDuration
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
		case metadata.NormalDCBBallotMetaFromSealer:
			if !(currentBlockHeight < endedDCBPivot && currentBlockHeight >= lv1DCBPivot) {
				continue
			}
		case metadata.NormalDCBBallotMetaFromOwner:
			if !(currentBlockHeight < endedDCBPivot && currentBlockHeight >= lv1DCBPivot) {
				continue
			}
		case metadata.SealedLv1DCBBallotMeta:
			if !(currentBlockHeight < lv1DCBPivot && currentBlockHeight >= lv2DCBPivot) {
				continue
			}
		case metadata.SealedLv2DCBBallotMeta:
			if !(currentBlockHeight < lv2DCBPivot && currentBlockHeight >= lv3DCBPivot) {
				continue
			}
		case metadata.SealedLv3DCBBallotMeta:
			if !(currentBlockHeight < lv3DCBPivot && currentBlockHeight >= startedDCBPivot) {
				continue
			}
		case metadata.NormalGOVBallotMetaFromSealer:
			if !(currentBlockHeight < endedGOVPivot && currentBlockHeight >= lv1GOVPivot) {
				continue
			}
		case metadata.NormalGOVBallotMetaFromOwner:
			if !(currentBlockHeight < endedGOVPivot && currentBlockHeight >= lv1GOVPivot) {
				continue
			}
		case metadata.SealedLv1GOVBallotMeta:
			if !(currentBlockHeight < lv1GOVPivot && currentBlockHeight >= lv2GOVPivot) {
				continue
			}
		case metadata.SealedLv2GOVBallotMeta:
			if !(currentBlockHeight < lv2GOVPivot && currentBlockHeight >= lv3GOVPivot) {
				continue
			}
		case metadata.SealedLv3GOVBallotMeta:
			if !(currentBlockHeight < lv3GOVPivot && currentBlockHeight >= startedGOVPivot) {
				continue
			}
		case metadata.IssuingRequestMeta:
			addable, newDCBTokensSold := blockgen.checkIssuingReqTx(chainID, tx, dcbTokensSold)
			dcbTokensSold = newDCBTokensSold
			if !addable {
				txToRemove = append(txToRemove, tx)
				continue
			}
			issuingReqTxs = append(issuingReqTxs, tx)
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
	salaryTx, err := transaction.CreateTxSalary(totalSalary, payToAddress, privatekey, blockgen.chain.config.DataBase)
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
	buyBackResTxs, err := blockgen.buildBuyBackResponsesTx(buyBackFromInfos, chainID, privatekey)
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}

	// create refund txs
	currentSalaryFund := prevBlock.Header.SalaryFund
	remainingFund := currentSalaryFund + totalFee + salaryFundAdd + incomeFromBonds - (totalSalary + buyBackCoins)
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
	if blockgen.neededNewDCBConstitution(chainID) {
		tx, err := blockgen.createAcceptConstitutionAndPunishTx(chainID, DCBConstitutionHelper{})
		coinbases = append(coinbases, tx...)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
	}
	if blockgen.neededNewGOVConstitution(chainID) {
		tx, err := blockgen.createAcceptConstitutionAndPunishTx(chainID, GOVConstitutionHelper{})
		coinbases = append(coinbases, tx...)
		if err != nil {
			Logger.log.Error(err)
			return nil, err
		}
	}

	if int32(prevBlock.Header.DCBGovernor.EndBlock) == prevBlock.Header.Height+1 {
		newBoardList, _ := blockgen.chain.config.DataBase.GetTopMostVoteDCBGovernor(common.NumberOfDCBGovernors)
		sort.Sort(newBoardList)
		sumOfVote := uint64(0)
		var newDCBBoardPubKey [][]byte
		for _, i := range newBoardList {
			newDCBBoardPubKey = append(newDCBBoardPubKey, i.PubKey)
			sumOfVote += i.VoteAmount
		}

		coinbases = append(coinbases, blockgen.createAcceptDCBBoardTx(newDCBBoardPubKey, sumOfVote))
		coinbases = append(coinbases, blockgen.CreateSendDCBVoteTokenToGovernorTx(chainID, newBoardList, sumOfVote)...)

		coinbases = append(coinbases, blockgen.CreateSendBackDCBTokenAfterVoteFail(chainID, newDCBBoardPubKey)...)
		// Todo @0xjackalope: send reward to old board and delete them from database before send back token to new board
		//xxx add to pool
	}

	if int32(prevBlock.Header.GOVGovernor.EndBlock) == prevBlock.Header.Height+1 {
		newBoardList, _ := blockgen.chain.config.DataBase.GetTopMostVoteGOVGovernor(common.NumberOfGOVGovernors)
		sort.Sort(newBoardList)
		sumOfVote := uint64(0)
		var newGOVBoardPubKey [][]byte
		for _, i := range newBoardList {
			newGOVBoardPubKey = append(newGOVBoardPubKey, i.PubKey)
			sumOfVote += i.VoteAmount
		}

		coinbases = append(coinbases, blockgen.createAcceptGOVBoardTx(newGOVBoardPubKey, sumOfVote))
		coinbases = append(coinbases, blockgen.CreateSendGOVVoteTokenToGovernorTx(chainID, newBoardList, sumOfVote)...)

		coinbases = append(coinbases, blockgen.CreateSendBackGOVTokenAfterVoteFail(chainID, newGOVBoardPubKey)...)
		// Todo @0xjackalope: send reward to old board and delete them from database before send back token to new board
		//xxx add to pool
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

	txsToAdd = append(coinbases, txsToAdd...)

	for _, tx := range txToRemove {
		blockgen.txPool.RemoveTx(tx)
	}

	// Check for final balance of DCB and GOV
	if currentSalaryFund+totalFee+salaryFundAdd+incomeFromBonds < totalSalary+govPayoutAmount+buyBackCoins+totalRefundAmt {
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
		SalaryFund:         currentSalaryFund + incomeFromBonds + totalFee + salaryFundAdd - totalSalary - govPayoutAmount - buyBackCoins - totalRefundAmt,
		BankFund:           prevBlock.Header.BankFund + loanPaymentAmount - bankPayoutAmount,
		GOVConstitution:    prevBlock.Header.GOVConstitution, // TODO: need get from gov-params tx
		DCBConstitution:    prevBlock.Header.DCBConstitution, // TODO: need get from dcb-params tx
	}
	if block.Header.GOVConstitution.GOVParams.SellingBonds != nil {
		block.Header.GOVConstitution.GOVParams.SellingBonds.BondsToSell -= bondsSold
	}
	if block.Header.DCBConstitution.DCBParams.SaleDBCTOkensByUSDData != nil {
		block.Header.DCBConstitution.DCBParams.SaleDBCTOkensByUSDData.Amount -= dcbTokensSold
	}

	for _, tx := range txsToAdd {
		if err := block.AddTransaction(tx); err != nil {
			return nil, err
		}
		// Handle if this transaction change something in block header
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

//1. Current National welfare (NW)  < lastNW * 0.9 (Emergency case)
//2. Block height == last constitution start time + last constitution window
func (blockgen *BlkTmplGenerator) neededNewDCBConstitution(chainID byte) bool {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastDCBConstitution := BestBlock.Header.DCBConstitution
	if GetOracleDCBNationalWelfare() < lastDCBConstitution.CurrentDCBNationalWelfare*ThresholdRatioOfDCBCrisis/100 ||
		uint32(BestBlock.Header.Height+1) == lastDCBConstitution.StartedBlockHeight+lastDCBConstitution.ExecuteDuration {
		return true
	}
	return false
}
func (blockgen *BlkTmplGenerator) neededNewGOVConstitution(chainID byte) bool {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock
	lastGovConstitution := BestBlock.Header.GOVConstitution
	if GetOracleGOVNationalWelfare() < lastGovConstitution.CurrentGOVNationalWelfare*ThresholdRatioOfGovCrisis/100 ||
		uint32(BestBlock.Header.Height+1) == lastGovConstitution.StartedBlockHeight+lastGovConstitution.ExecuteDuration {
		return true
	}
	return false
}

func (blockgen *BlkTmplGenerator) createAcceptConstitutionTxDecs(
	chainID byte,
	ConstitutionHelper ConstitutionHelper,
) (*metadata.TxDesc, error) {
	BestBlock := blockgen.chain.BestState[chainID].BestBlock

	// count vote from lastConstitution.StartedBlockHeight to Bestblock height
	CountVote := make(map[common.Hash]int64)
	Transaction := make(map[common.Hash]*metadata.Transaction)
	for blockHeight := ConstitutionHelper.GetStartedNormalVote(blockgen, chainID); blockHeight < uint32(BestBlock.Header.Height); blockHeight += 1 {
		//retrieve block from block's height
		hashBlock, err := blockgen.chain.config.DataBase.GetBlockByIndex(int32(blockHeight), chainID)
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

	acceptedSubmitProposalTransaction := ConstitutionHelper.TxAcceptProposal(&res)

	AcceptedTransactionDesc := metadata.TxDesc{
		Tx:     acceptedSubmitProposalTransaction,
		Added:  time.Now(),
		Height: BestBlock.Header.Height,
		Fee:    0,
	}
	return &AcceptedTransactionDesc, nil
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
			// TODO(@0xbunyip): holder here is Pk only, change to use both Pk and Pkenc
			holderAddr, err := hex.DecodeString(holder)
			if err != nil {
				return nil, 0, err
			}
			holderAddress := (&privacy.PaymentAddress{}).FromBytes(holderAddr)
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
	sellingBondsParam *voting.SellingBonds,
) (*transaction.TxCustomToken, error) {
	bondID := fmt.Sprintf("%s%s%s", sellingBondsParam.Maturity, sellingBondsParam.BuyBackPrice, sellingBondsParam.StartSellingAt)
	additionalSuffix := make([]byte, 24-len(bondID))
	bondIDBytes := append([]byte(bondID), additionalSuffix...)

	buySellRes := metadata.BuySellResponse{
		RequestedTxID:  buySellReqTx.Hash(),
		StartSellingAt: sellingBondsParam.StartSellingAt,
		Maturity:       sellingBondsParam.Maturity,
		BuyBackPrice:   sellingBondsParam.BuyBackPrice,
		BondID:         bondIDBytes,
	}
	buySellRes.Type = metadata.BuyFromGOVResponseMeta

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
	copy(propertyID[:], append(common.BondTokenID[0:8], bondIDBytes...))
	txTokenData := transaction.TxTokenData{
		Type:       transaction.CustomTokenInit,
		Amount:     buySellReq.Amount,
		PropertyID: common.Hash(propertyID),
		Vins:       []transaction.TxTokenVin{},
		Vouts:      []transaction.TxTokenVout{txTokenVout},
		// PropertyName:   "",
		// PropertySymbol: coinbaseTxType,
	}
	resTx := &transaction.TxCustomToken{
		TxTokenData: txTokenData,
	}
	resTx.Type = common.TxCustomTokenType
	resTx.Metadata = &buySellRes
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
	sellingBondsParam *voting.SellingBonds,
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
	buyBackReqMeta := tx.GetMetadata()
	buyBackReq, ok := buyBackReqMeta.(*metadata.BuyBackRequest)
	if !ok {
		Logger.log.Error(errors.New("Could not parse BuyBackRequest metadata."))
		return nil, false
	}
	_, _, _, fromTx, err := blockgen.chain.GetTransactionByHash(&buyBackReq.BuyBackFromTxID)
	if err != nil {
		Logger.log.Error(err)
		return nil, false
	}
	customTokenTx, ok := fromTx.(*transaction.TxCustomToken)
	if !ok {
		Logger.log.Error(errors.New("Could not parse TxCustomToken."))
		return nil, false
	}
	fromTxMeta := fromTx.GetMetadata()
	buySellRes, ok := fromTxMeta.(*metadata.BuySellResponse)
	if !ok {
		Logger.log.Error(errors.New("Could not parse BuySellResponse metadata."))
		return nil, false
	}

	vout := customTokenTx.TxTokenData.Vouts[buyBackReq.VoutIndex]
	prevBlock := blockgen.chain.BestState[chainID].BestBlock

	if buySellRes.StartSellingAt+buySellRes.Maturity > uint32(prevBlock.Header.Height)+1 {
		Logger.log.Error("The token is not overdued yet.")
		return nil, false
	}
	// check remaining constants in GOV fund is enough or not
	buyBackAmount := vout.Value * buySellRes.BuyBackPrice
	if buyBackConsts+buyBackAmount > prevBlock.Header.SalaryFund {
		return nil, false
	}
	buyBackFromInfo := &buyBackFromInfo{
		paymentAddress: vout.PaymentAddress,
		buyBackPrice:   buySellRes.BuyBackPrice,
		value:          vout.Value,
		requestedTxID:  tx.Hash(),
	}
	return buyBackFromInfo, true
}

func (blockgen *BlkTmplGenerator) buildBuyBackResponsesTx(
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
		buyBackResTx, err := transaction.CreateTxSalary(buyBackAmount, &buyBackFromInfo.paymentAddress, privatekey, blockgen.chain.GetDatabase())
		if err != nil {
			return []*transaction.Tx{}, err
		}
		buyBackRes := &metadata.BuyBackResponse{
			RequestedTxID: buyBackFromInfo.requestedTxID,
		}
		buyBackRes.Type = metadata.BuyBackResponseMeta
		buyBackResTx.SetMetadata(buyBackRes)
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
			issuingRes := metadata.IssuingResponse{
				RequestedTxID: issuingReqTx.Hash(),
			}
			issuingRes.Type = metadata.IssuingResponseMeta
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
			resTx.Metadata = &issuingRes
			issuingResTxs = append(issuingResTxs, resTx)
			continue
		}
		if issuingReq.AssetType == common.ConstantID {
			constantPrice := uint64(1)
			if oracleParams.Constant != 0 {
				constantPrice = oracleParams.Constant
			}
			issuingAmt := issuingReq.DepositedAmount / constantPrice
			resTx, err := transaction.CreateTxSalary(issuingAmt, &issuingReq.ReceiverAddress, privatekey, blockgen.chain.GetDatabase())
			if err != nil {
				return []metadata.Transaction{}, err
			}
			issuingRes := &metadata.IssuingResponse{
				RequestedTxID: issuingReqTx.Hash(),
			}
			issuingRes.Type = metadata.IssuingResponseMeta
			resTx.SetMetadata(issuingRes)
			issuingResTxs = append(issuingResTxs, resTx)
		}
	}
	return issuingResTxs, nil
}

func calculateAmountOfRefundTxs(
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
		refundTx, err := transaction.CreateTxSalary(actualRefundAmt, addr, privatekey, db)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		refundTx.Type = common.TxRefundType
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
	if len(addresses) == 0 {
		return []*transaction.Tx{}, 0
	}
	refundTxs, totalRefundAmt := calculateAmountOfRefundTxs(
		addresses,
		estimatedRefundAmt,
		remainingFund,
		blockgen.chain.GetDatabase(),
		privatekey,
	)
	return refundTxs, totalRefundAmt
}

func (blockgen *BlkTmplGenerator) buildResponseForCrowdsale(
	tx *transaction.TxCustomToken,
	saleDataMap map[string]*voting.SaleData,
	unspentTokenMap map[string]([]transaction.TxTokenVout),
	rt []byte,
	chainID byte,
	saleID []byte,
	producerPrivateKey *privacy.SpendingKey,
) (*transaction.TxCustomToken, error) {
	accountDCB, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbPk := accountDCB.KeySet.PaymentAddress.Pk
	saleData := saleDataMap[string(saleID)]

	// Get price for asset
	prices := blockgen.chain.BestState[chainID].BestBlock.Header.Oracle.Bonds
	// TODO(@0xbunyip): validate sale data in proposal to admit only valid pair of assets
	txResponse := &transaction.TxCustomToken{}
	err := errors.New("Incorrect assets for crowdsale")
	sellingAsset := &common.Hash{}
	copy(sellingAsset[:], saleData.SellingAsset)
	if bytes.Equal(sellingAsset[:], common.ConstantID[:]) {
		if bytes.Equal(saleData.BuyingAsset, common.OffchainAssetID[:]) {
			return nil, nil // no response for offchain asset
		}

		tokenAmount, valuesInConstant := getTxTokenValue(tx.TxTokenData, saleData.BuyingAsset, dcbPk, prices)
		if tokenAmount > saleData.BuyingAmount || valuesInConstant > saleData.SellingAmount {
			// User sent too many token, reject request
			return nil, fmt.Errorf("Crowdsale reached limit")
		}
		// Update amount of buying/selling asset of the crowdsale
		saleData.BuyingAmount -= tokenAmount
		saleData.SellingAmount -= valuesInConstant
		txResponse, err = buildResponseForCoin(tx, valuesInConstant, saleData.SaleID, producerPrivateKey, blockgen.chain.GetDatabase())
	} else if bytes.Equal(sellingAsset[:8], common.BondTokenID[:8]) || bytes.Equal(sellingAsset[:], common.DCBTokenID[:]) {
		// Get unspent token UTXO to send to user
		if _, ok := unspentTokenMap[string(sellingAsset[:])]; !ok {
			unspentTxTokenOuts, err := blockgen.chain.GetUnspentTxCustomTokenVout(accountDCB.KeySet, sellingAsset)
			if err == nil {
				unspentTokenMap[string(sellingAsset[:])] = unspentTxTokenOuts
			} else {
				unspentTokenMap[string(sellingAsset[:])] = []transaction.TxTokenVout{}
			}
		}
		mint := bytes.Equal(sellingAsset[:], common.DCBTokenID[:]) // Mint DCB token, transfer bonds
		constantAmount, valuesInToken := getTxValue(&tx.Tx, sellingAsset[:], dcbPk, prices)
		if constantAmount > saleData.BuyingAmount || valuesInToken > saleData.SellingAmount {
			return nil, fmt.Errorf("Crowdsale reached limit")
		}
		saleData.BuyingAmount -= constantAmount
		saleData.SellingAmount -= valuesInToken
		txResponse, err = buildResponseForToken(tx, valuesInToken, sellingAsset[:], rt, chainID, unspentTokenMap, saleData.SaleID, mint)
	}
	return txResponse, err
}

func (blockgen *BlkTmplGenerator) processCrowdsale(sourceTxns []*metadata.TxDesc, rt []byte, chainID byte, producerPrivateKey *privacy.SpendingKey) ([]*transaction.TxCustomToken, []metadata.Transaction, error) {
	txsToRemove := []metadata.Transaction{}
	txsResponse := []*transaction.TxCustomToken{}

	// Get unspent bond tx to spend if needed
	unspentTokenMap := map[string]([]transaction.TxTokenVout){}
	saleDataMap := map[string]*voting.SaleData{}
	for _, txDesc := range sourceTxns {
		if txDesc.Tx.GetMetadataType() != metadata.CrowdsaleRequestMeta {
			continue
		}

		tx, ok := (txDesc.Tx).(*transaction.TxCustomToken)
		if !ok {
			txsToRemove = append(txsToRemove, tx)
		}

		// Create corresponding response to send selling asset
		// Get buying and selling asset from current sale
		meta := txDesc.Tx.GetMetadata()
		if meta == nil {
			txsToRemove = append(txsToRemove, tx)
			continue
		}
		metaRequest, ok := meta.(*metadata.CrowdsaleRequest)
		if !ok {
			txsToRemove = append(txsToRemove, tx)
			continue
		}
		if _, ok := saleDataMap[string(metaRequest.SaleID)]; !ok {
			saleData, err := blockgen.chain.GetCrowdsaleData(metaRequest.SaleID)
			if err != nil {
				txsToRemove = append(txsToRemove, tx)
				continue
			}

			saleDataMap[string(metaRequest.SaleID)] = saleData
		}
		txResponse, err := blockgen.buildResponseForCrowdsale(tx, saleDataMap, unspentTokenMap, rt, chainID, metaRequest.SaleID, producerPrivateKey)
		if err != nil {
			txsToRemove = append(txsToRemove, tx)
		} else if txResponse != nil {
			txsResponse = append(txsResponse, txResponse)
		}
	}
	return txsResponse, txsToRemove, nil
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
			pks := [][]byte{meta.ReceiveAddress.Pk[:]}
			tks := [][]byte{meta.ReceiveAddress.Tk[:]}
			amounts := []uint64{meta.LoanAmount}
			txNormals, err := buildCoinbaseTxs(pks, tks, amounts, producerPrivateKey, blockgen.chain.GetDatabase())
			if err != nil {
				removableTxs = append(removableTxs, txDesc.Tx)
				continue
			}
			unlockMeta := &metadata.LoanUnlock{
				LoanID: make([]byte, len(withdrawMeta.LoanID)),
			}
			copy(unlockMeta.LoanID, withdrawMeta.LoanID)
			txNormals[0].Metadata = unlockMeta
			loanUnlockTxs = append(loanUnlockTxs, txNormals[0]) // There's only one tx
		}
	}
	return amount, loanUnlockTxs, removableTxs
}
