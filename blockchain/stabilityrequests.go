package blockchain

import (
	"bytes"
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type buyBackFromInfo struct {
	paymentAddress privacy.PaymentAddress
	buyBackPrice   uint64
	value          uint64
	requestedTxID  *common.Hash
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
		return common.FalseValue, dcbTokensSold
	}
	if !bytes.Equal(issuingReq.AssetType[:], common.DCBTokenID[:]) {
		return common.TrueValue, dcbTokensSold
	}
	header := blockgen.chain.BestState[chainID].BestBlock.Header
	saleDBCTOkensByUSDData := header.DCBConstitution.DCBParams.SaleDCBTokensByUSDData
	oracleParams := header.Oracle
	dcbTokenPrice := uint64(1)
	if oracleParams.DCBToken != 0 {
		dcbTokenPrice = oracleParams.DCBToken
	}
	dcbTokensReq := issuingReq.DepositedAmount / dcbTokenPrice
	if dcbTokensSold+dcbTokensReq > saleDBCTOkensByUSDData.Amount {
		return common.FalseValue, dcbTokensSold
	}
	return common.TrueValue, dcbTokensSold + dcbTokensReq
}

func (blockgen *BlkTmplGenerator) checkBuyBackReqTx(
	chainID byte,
	tx metadata.Transaction,
	buyBackConsts uint64,
) (*buyBackFromInfo, bool) {
	buyBackReqTx, ok := tx.(*transaction.TxCustomToken)
	if !ok {
		Logger.log.Error(errors.New("Could not parse BuyBackRequest tx (custom token tx)."))
		return nil, common.FalseValue
	}
	vins := buyBackReqTx.TxTokenData.Vins
	if len(vins) == 0 {
		Logger.log.Error(errors.New("No existed Vins from BuyBackRequest tx"))
		return nil, common.FalseValue
	}
	priorTxID := vins[0].TxCustomTokenID
	_, _, _, priorTx, err := blockgen.chain.GetTransactionByHash(&priorTxID)
	if err != nil {
		Logger.log.Error(err)
		return nil, common.FalseValue
	}
	priorCustomTokenTx, ok := priorTx.(*transaction.TxCustomToken)
	if !ok {
		Logger.log.Error(errors.New("Could not parse prior TxCustomToken."))
		return nil, common.FalseValue
	}

	priorMeta := priorCustomTokenTx.GetMetadata()
	if priorMeta == nil {
		Logger.log.Error(errors.New("No existed metadata in priorCustomTokenTx"))
		return nil, common.FalseValue
	}
	buySellResMeta, ok := priorMeta.(*metadata.BuySellResponse)
	if !ok {
		Logger.log.Error(errors.New("Could not parse BuySellResponse metadata."))
		return nil, common.FalseValue
	}
	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	if buySellResMeta.StartSellingAt+buySellResMeta.Maturity > uint64(prevBlock.Header.Height)+1 {
		Logger.log.Error("The token is not overdued yet.")
		return nil, common.FalseValue
	}
	// check remaining constants in GOV fund is enough or not
	buyBackReqMeta := buyBackReqTx.GetMetadata()
	buyBackReq, ok := buyBackReqMeta.(*metadata.BuyBackRequest)
	if !ok {
		Logger.log.Error(errors.New("Could not parse BuyBackRequest metadata."))
		return nil, common.FalseValue
	}
	buyBackValue := buyBackReq.Amount * buySellResMeta.BuyBackPrice
	if buyBackConsts+buyBackValue > prevBlock.Header.SalaryFund {
		return nil, common.FalseValue
	}
	buyBackFromInfo := &buyBackFromInfo{
		paymentAddress: buyBackReq.PaymentAddress,
		buyBackPrice:   buySellResMeta.BuyBackPrice,
		value:          buyBackReq.Amount,
		requestedTxID:  tx.Hash(),
	}
	return buyBackFromInfo, common.TrueValue
}

func (blockgen *BlkTmplGenerator) checkBuyFromGOVReqTx(
	chainID byte,
	tx metadata.Transaction,
	bondsSold uint64,
) (uint64, uint64, bool) {
	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	sellingBondsParams := prevBlock.Header.GOVConstitution.GOVParams.SellingBonds
	if uint64(prevBlock.Header.Height)+1 > sellingBondsParams.StartSellingAt+sellingBondsParams.SellingWithin {
		return 0, 0, common.FalseValue
	}

	buySellReqMeta := tx.GetMetadata()
	req, ok := buySellReqMeta.(*metadata.BuySellRequest)
	if !ok {
		return 0, 0, common.FalseValue
	}

	if bondsSold+req.Amount > sellingBondsParams.BondsToSell { // run out of bonds for selling
		return 0, 0, common.FalseValue
	}
	return req.Amount * req.BuyPrice, req.Amount, common.TrueValue
}

func (blockgen *BlkTmplGenerator) processDividend(
	proposal *metadata.DividendProposal,
	blockHeight int32,
	producerPrivateKey *privacy.SpendingKey,
) ([]*transaction.Tx, uint64, error) {
	payoutAmount := uint64(0)
	// TODO(@0xbunyip): how to execute payout dividend proposal
	dividendTxs := []*transaction.Tx{}
	if common.FalseValue && blockHeight%metadata.PayoutFrequency == 0 { // only chain 0 process dividend proposals
		totalTokenSupply, tokenHolders, amounts, err := blockgen.chain.GetAmountPerAccount(proposal)
		if err != nil || totalTokenSupply == 0 {
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

		dividendTxs, err = transaction.BuildDividendTxs(infos, proposal, producerPrivateKey, blockgen.chain.GetDatabase())
		if err != nil {
			return nil, 0, err
		}
	}
	return dividendTxs, payoutAmount, nil
}

func (blockgen *BlkTmplGenerator) processBankDividend(blockHeight int32, producerPrivateKey *privacy.SpendingKey) ([]*transaction.Tx, uint64, error) {
	tokenID, err := (&common.Hash{}).NewHash(common.DCBTokenID[:])
	if err != nil {
		return nil, 0, err
	}
	proposal := &metadata.DividendProposal{
		TokenID: tokenID,
	}
	return blockgen.processDividend(proposal, blockHeight, producerPrivateKey)
}

func (blockgen *BlkTmplGenerator) processGovDividend(blockHeight int32, producerPrivateKey *privacy.SpendingKey) ([]*transaction.Tx, uint64, error) {
	tokenID, err := (&common.Hash{}).NewHash(common.GOVTokenID[:])
	if err != nil {
		return nil, 0, err
	}
	proposal := &metadata.DividendProposal{
		TokenID: tokenID,
	}
	return blockgen.processDividend(proposal, blockHeight, producerPrivateKey)
}

func (blockgen *BlkTmplGenerator) checkAndGroupTxs(
	sourceTxns []*metadata.TxDesc,
	chainID byte,
	privatekey *privacy.SpendingKey,
) (map[string][]metadata.Transaction, map[string]uint64, []*buyBackFromInfo, error) {
	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	blockHeight := prevBlock.Header.Height + 1
	rt := []byte{}

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
	bankPayoutAmount := uint64(0)
	govPayoutAmount := uint64(0)

	for _, txDesc := range sourceTxns {
		tx := txDesc.Tx
		txChainID, err := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
		if txChainID != chainID || err != nil {
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

	// TODO(@0xbunyip): cap #tx to common.MaxTxsInBlock
	// Process dividend payout for DCB if needed
	bankDivTxs, bankPayoutAmount, err := blockgen.processBankDividend(blockHeight, privatekey)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, tx := range bankDivTxs {
		txsToAdd = append(txsToAdd, tx)
	}

	// Process dividend payout for GOV if needed
	govDivTxs, govPayoutAmount, err := blockgen.processGovDividend(blockHeight, privatekey)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, tx := range govDivTxs {
		txsToAdd = append(txsToAdd, tx)
	}

	// Process crowdsale for DCB
	dcbSaleTxs, removableTxs, err := blockgen.processCrowdsale(sourceTxns, rt, chainID, privatekey)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, tx := range dcbSaleTxs {
		txsToAdd = append(txsToAdd, tx)
	}
	for _, tx := range removableTxs {
		txsToAdd = append(txsToAdd, tx)
	}

	// Build CMB responses
	cmbInitRefundTxs, err := blockgen.buildCMBRefund(sourceTxns, chainID, privatekey)
	if err != nil {
		return nil, nil, nil, err
	}
	for _, tx := range cmbInitRefundTxs {
		txsToAdd = append(txsToAdd, tx)
	}

	txGroups := map[string][]metadata.Transaction{
		"txsToAdd":                 txsToAdd,
		"txToRemove":               txToRemove,
		"buySellReqTxs":            buySellReqTxs,
		"issuingReqTxs":            issuingReqTxs,
		"updatingOracleBoardTxs":   updatingOracleBoardTxs,
		"multiSigsRegistrationTxs": multiSigsRegistrationTxs,
	}
	accumulativeValues := map[string]uint64{
		"bondsSold":        bondsSold,
		"dcbTokensSold":    dcbTokensSold,
		"incomeFromBonds":  incomeFromBonds,
		"totalFee":         totalFee,
		"buyBackCoins":     buyBackCoins,
		"govPayoutAmount":  govPayoutAmount,
		"bankPayoutAmount": bankPayoutAmount,
	}
	return txGroups, accumulativeValues, buyBackFromInfos, nil
}
