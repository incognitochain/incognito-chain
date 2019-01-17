package blockchain

import (
	"errors"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

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
) ([]metadata.Transaction, uint64) {
	amt := uint64(0)
	if estimatedRefundAmt <= remainingFund {
		amt = estimatedRefundAmt
	} else {
		amt = remainingFund
	}
	actualRefundAmt := amt / uint64(len(addresses))
	var refundTxs []metadata.Transaction
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
) ([]metadata.Transaction, uint64) {
	if remainingFund <= 0 {
		Logger.log.Info("GOV fund is not enough for refund.")
		return []metadata.Transaction{}, 0
	}
	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	header := prevBlock.Header
	govParams := header.GOVConstitution.GOVParams
	refundInfo := govParams.RefundInfo
	if refundInfo == nil {
		Logger.log.Info("Refund info is not existed.")
		return []metadata.Transaction{}, 0
	}
	lookbackBlockHeight := header.Height - common.RefundPeriod
	if lookbackBlockHeight < 0 {
		return []metadata.Transaction{}, 0
	}
	lookbackBlock, err := blockgen.chain.GetBlockByBlockHeight(lookbackBlockHeight, chainID)
	if err != nil {
		Logger.log.Error(err)
		return []metadata.Transaction{}, 0
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
		addr := lookbackTx.GetSenderAddress()
		if addr == nil {
			continue
		}
		txValue := lookbackTx.CalculateTxValue()
		if txValue > refundInfo.ThresholdToLargeTx {
			continue
		}
		addresses = append(addresses, addr)
		smallTxHashes = append(smallTxHashes, tx.Hash())
		estimatedRefundAmt += refundInfo.RefundAmount
	}
	if len(addresses) == 0 {
		return []metadata.Transaction{}, 0
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

func (blockgen *BlkTmplGenerator) processLoan(sourceTxns []*metadata.TxDesc, producerPrivateKey *privacy.SpendingKey) (uint64, []metadata.Transaction, []metadata.Transaction) {
	amount := uint64(0)
	loanUnlockTxs := []metadata.Transaction{}
	removableTxs := []metadata.Transaction{}
	for _, txDesc := range sourceTxns {
		if txDesc.Tx.GetMetadataType() == metadata.LoanPaymentMeta {
			paymentAmount, err := blockgen.calculateInterestPaid(txDesc.Tx)
			if err != nil {
				removableTxs = append(removableTxs, txDesc.Tx)
				continue
			}
			amount += paymentAmount
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
			amounts := []uint64{meta.LoanAmount}
			txNormals, err := transaction.BuildCoinbaseTxs([]*privacy.PaymentAddress{meta.ReceiveAddress}, amounts, producerPrivateKey, blockgen.chain.GetDatabase(), uplockMetaList)
			if err != nil {
				removableTxs = append(removableTxs, txDesc.Tx)
				continue
			}
			loanUnlockTxs = append(loanUnlockTxs, txNormals[0]) // There's only one tx
		}
	}
	return amount, loanUnlockTxs, removableTxs
}

func (blockgen *BlkTmplGenerator) buildBuyBackResponseTxs(
	buyBackFromInfos []*buyBackFromInfo,
	chainID byte,
	privatekey *privacy.SpendingKey,
) ([]metadata.Transaction, error) {
	if len(buyBackFromInfos) == 0 {
		return []metadata.Transaction{}, nil
	}

	// prevBlock := blockgen.chain.BestState[chainID].BestBlock
	var buyBackResTxs []metadata.Transaction
	for _, buyBackFromInfo := range buyBackFromInfos {
		buyBackAmount := buyBackFromInfo.value * buyBackFromInfo.buyBackPrice
		buyBackRes := metadata.NewBuyBackResponse(*buyBackFromInfo.requestedTxID, metadata.BuyBackResponseMeta)
		buyBackResTx := new(transaction.Tx)
		err := buyBackResTx.InitTxSalary(buyBackAmount, &buyBackFromInfo.paymentAddress, privatekey, blockgen.chain.GetDatabase(), buyBackRes)
		if err != nil {
			return []metadata.Transaction{}, err
		}
		buyBackResTxs = append(buyBackResTxs, buyBackResTx)
	}
	return buyBackResTxs, nil
}

// buildBuySellResponsesTx
// the tx is to distribute tokens (bond, gov, ...) to token requesters
func (blockgen *BlkTmplGenerator) buildBuySellResponsesTx(
	buySellReqTxs []metadata.Transaction,
	sellingBondsParam *params.SellingBonds,
) ([]metadata.Transaction, error) {
	if len(buySellReqTxs) == 0 {
		return []metadata.Transaction{}, nil
	}
	var resTxs []metadata.Transaction
	for _, reqTx := range buySellReqTxs {
		resTx, err := buildSingleBuySellResponseTx(reqTx, sellingBondsParam)
		if err != nil {
			return []metadata.Transaction{}, err
		}
		resTxs = append(resTxs, resTx)
	}
	return resTxs, nil
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

func (blockgen *BlkTmplGenerator) buildResponseTxs(
	chainID byte,
	sourceTxns []*metadata.TxDesc,
	privatekey *privacy.SpendingKey,
	txGroups *txGroups,
	accumulativeValues *accumulativeValues,
	buyBackFromInfos []*buyBackFromInfo,
) (*txGroups, *accumulativeValues, map[string]uint64, error) {
	prevBlock := blockgen.chain.BestState[chainID].BestBlock
	// create buy/sell response txs to distribute bonds/govs to requesters
	buySellResTxs, err := blockgen.buildBuySellResponsesTx(
		txGroups.buySellReqTxs,
		blockgen.chain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams.SellingBonds,
	)
	if err != nil {
		Logger.log.Error(err)
		return nil, nil, nil, err
	}
	// create buy-back response txs to distribute constants to buy-back requesters
	buyBackResTxs, err := blockgen.buildBuyBackResponseTxs(buyBackFromInfos, chainID, privatekey)
	if err != nil {
		Logger.log.Error(err)
		return nil, nil, nil, err
	}

	oracleRewardTxs, totalOracleRewards, updatedOracleValues, err := blockgen.buildOracleRewardTxs(chainID, privatekey)
	if err != nil {
		Logger.log.Error(err)
		return nil, nil, nil, err
	}

	// create refund txs
	currentSalaryFund := prevBlock.Header.SalaryFund
	remainingFund := currentSalaryFund + accumulativeValues.totalFee + accumulativeValues.incomeFromBonds - (accumulativeValues.totalSalary + accumulativeValues.buyBackCoins + totalOracleRewards)
	refundTxs, totalRefundAmt := blockgen.buildRefundTxs(chainID, remainingFund, privatekey)

	issuingResTxs, err := blockgen.buildIssuingResTxs(chainID, txGroups.issuingReqTxs, privatekey)
	if err != nil {
		Logger.log.Error(err)
		return nil, nil, nil, err
	}

	// Get loan payment amount to add to DCB fund
	loanPaymentAmount, unlockTxs, removableTxs := blockgen.processLoan(sourceTxns, privatekey)
	for _, tx := range removableTxs {
		txGroups.txToRemove = append(txGroups.txToRemove, tx)
	}
	txGroups.buySellResTxs = buySellResTxs
	txGroups.buyBackResTxs = buyBackResTxs
	txGroups.oracleRewardTxs = oracleRewardTxs
	txGroups.refundTxs = refundTxs
	txGroups.issuingResTxs = issuingResTxs
	txGroups.unlockTxs = unlockTxs

	accumulativeValues.totalOracleRewards = totalOracleRewards
	accumulativeValues.totalRefundAmt = totalRefundAmt
	accumulativeValues.loanPaymentAmount = loanPaymentAmount
	accumulativeValues.currentSalaryFund = currentSalaryFund
	return txGroups, accumulativeValues, updatedOracleValues, nil
}
