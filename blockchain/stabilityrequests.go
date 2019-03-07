package blockchain

// type buyBackFromInfo struct {
// 	paymentAddress privacy.PaymentAddress
// 	buyBackPrice   uint64
// 	value          uint64
// 	requestedTxID  *common.Hash
// }

// type txGroups struct {
// 	txsToAdd                 []metadata.Transaction
// 	txToRemove               []metadata.Transaction
// 	buySellReqTxs            []metadata.Transaction
// 	buyGOVTokensReqTxs       []metadata.Transaction
// 	issuingReqTxs            []metadata.Transaction
// 	updatingOracleBoardTxs   []metadata.Transaction
// 	multiSigsRegistrationTxs []metadata.Transaction
// 	unlockTxs                []metadata.Transaction
// 	buySellResTxs            []metadata.Transaction
// 	buyGOVTokensResTxs       []metadata.Transaction
// 	buyBackResTxs            []metadata.Transaction
// 	issuingResTxs            []metadata.Transaction
// 	refundTxs                []metadata.Transaction
// 	oracleRewardTxs          []metadata.Transaction
// }

// type accumulativeValues struct {
// 	bondsSold           uint64
// 	govTokensSold       uint64
// 	dcbTokensSold       uint64
// 	incomeFromBonds     uint64
// 	incomeFromGOVTokens uint64
// 	totalFee            uint64
// 	buyBackCoins        uint64
// 	govPayoutAmount     uint64
// 	bankPayoutAmount    uint64
// 	totalSalary         uint64
// 	currentSalaryFund   uint64
// 	totalRefundAmt      uint64
// 	totalOracleRewards  uint64
// 	loanPaymentAmount   uint64
// }

// func (blockgen *BlkTmplGenerator) checkIssuingReqTx(
// 	shardID byte,
// 	tx metadata.Transaction,
// 	dcbTokensSold uint64,
// ) (bool, uint64) {
// 	issuingReqMeta := tx.GetMetadata()
// 	issuingReq, ok := issuingReqMeta.(*metadata.IssuingRequest)
// 	if !ok {
// 		Logger.log.Error(errors.New("Could not parse IssuingRequest metadata"))
// 		return false, dcbTokensSold
// 	}
// 	if !bytes.Equal(issuingReq.AssetType[:], common.DCBTokenID[:]) {
// 		return true, dcbTokensSold
// 	}
// 	header := blockgen.chain.BestState[shardID].BestBlock.Header
// 	saleDBCTOkensByUSDData := header.DCBConstitution.DCBParams.SaleDCBTokensByUSDData
// 	oracleParams := header.Oracle
// 	dcbTokenPrice := uint64(1)
// 	if oracleParams.DCBToken != 0 {
// 		dcbTokenPrice = oracleParams.DCBToken
// 	}
// 	dcbTokensReq := issuingReq.DepositedAmount / dcbTokenPrice
// 	if dcbTokensSold+dcbTokensReq > saleDBCTOkensByUSDData.Amount {
// 		return false, dcbTokensSold
// 	}
// 	return true, dcbTokensSold + dcbTokensReq
// }

// func (blockgen *BlkTmplGenerator) checkBuyBackReqTx(
// 	shardID byte,
// 	tx metadata.Transaction,
// 	buyBackConsts uint64,
// ) (*buyBackFromInfo, bool) {
// 	buyBackReqTx, ok := tx.(*transaction.TxCustomToken)
// 	if !ok {
// 		Logger.log.Error(errors.New("Could not parse BuyBackRequest tx (custom token tx)."))
// 		return nil, false
// 	}
// 	vins := buyBackReqTx.TxTokenData.Vins
// 	if len(vins) == 0 {
// 		Logger.log.Error(errors.New("No existed Vins from BuyBackRequest tx"))
// 		return nil, false
// 	}
// 	priorTxID := vins[0].TxCustomTokenID
// 	_, _, _, priorTx, err := blockgen.chain.GetTransactionByHash(&priorTxID)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, false
// 	}
// 	priorCustomTokenTx, ok := priorTx.(*transaction.TxCustomToken)
// 	if !ok {
// 		Logger.log.Error(errors.New("Could not parse prior TxCustomToken."))
// 		return nil, false
// 	}

// 	priorMeta := priorCustomTokenTx.GetMetadata()
// 	if priorMeta == nil {
// 		Logger.log.Error(errors.New("No existed metadata in priorCustomTokenTx"))
// 		return nil, false
// 	}
// 	buySellResMeta, ok := priorMeta.(*metadata.BuySellResponse)
// 	if !ok {
// 		Logger.log.Error(errors.New("Could not parse BuySellResponse metadata."))
// 		return nil, false
// 	}
// 	prevBlock := blockgen.chain.BestState[shardID].BestBlock
// 	if buySellResMeta.StartSellingAt+buySellResMeta.Maturity > uint64(prevBlock.Header.Height)+1 {
// 		Logger.log.Error("The token is not overdued yet.")
// 		return nil, false
// 	}
// 	// check remaining constants in GOV fund is enough or not
// 	buyBackReqMeta := buyBackReqTx.GetMetadata()
// 	buyBackReq, ok := buyBackReqMeta.(*metadata.BuyBackRequest)
// 	if !ok {
// 		Logger.log.Error(errors.New("Could not parse BuyBackRequest metadata."))
// 		return nil, false
// 	}
// 	buyBackValue := buyBackReq.Amount * buySellResMeta.BuyBackPrice
// 	if buyBackConsts+buyBackValue > prevBlock.Header.SalaryFund {
// 		return nil, false
// 	}
// 	buyBackFromInfo := &buyBackFromInfo{
// 		paymentAddress: buyBackReq.PaymentAddress,
// 		buyBackPrice:   buySellResMeta.BuyBackPrice,
// 		value:          buyBackReq.Amount,
// 		requestedTxID:  tx.Hash(),
// 	}
// 	return buyBackFromInfo, true
// }

// func (blockgen *BlkTmplGenerator) checkBuyFromGOVReqTx(
// 	shardID byte,
// 	tx metadata.Transaction,
// 	bondsSold uint64,
// ) (uint64, uint64, bool) {
// 	prevBlock := blockgen.chain.BestState[shardID].BestBlock
// 	sellingBondsParams := prevBlock.Header.GOVConstitution.GOVParams.SellingBonds
// 	if uint64(prevBlock.Header.Height)+1 > sellingBondsParams.StartSellingAt+sellingBondsParams.SellingWithin {
// 		return 0, 0, false
// 	}

// 	buySellReqMeta := tx.GetMetadata()
// 	req, ok := buySellReqMeta.(*metadata.BuySellRequest)
// 	if !ok {
// 		return 0, 0, false
// 	}

// 	if bondsSold+req.Amount > sellingBondsParams.BondsToSell { // run out of bonds for selling
// 		return 0, 0, false
// 	}
// 	return req.Amount * req.BuyPrice, req.Amount, true
// }

// func (blockgen *BlkTmplGenerator) checkBuyGOVTokensReqTx(
// 	shardID byte,
// 	tx metadata.Transaction,
// 	govTokensSold uint64,
// ) (uint64, uint64, bool) {
// 	prevBlock := blockgen.chain.BestState[shardID].BestBlock
// 	sellingGOVTokensParams := prevBlock.Header.GOVConstitution.GOVParams.SellingGOVTokens
// 	if uint64(prevBlock.Header.Height)+1 > sellingGOVTokensParams.StartSellingAt+sellingGOVTokensParams.SellingWithin {
// 		return 0, 0, false
// 	}

// 	buyGOVTokensReqMeta := tx.GetMetadata()
// 	req, ok := buyGOVTokensReqMeta.(*metadata.BuyGOVTokenRequest)
// 	if !ok {
// 		return 0, 0, false
// 	}

// 	if govTokensSold+req.Amount > sellingGOVTokensParams.GOVTokensToSell { // run out of gov tokens for selling
// 		return 0, 0, false
// 	}
// 	return req.Amount * req.BuyPrice, req.Amount, true
// }

// func (blockgen *BlkTmplGenerator) checkAndGroupTxs(
// 	sourceTxns []*metadata.TxDesc,
// 	shardID byte,
// 	privatekey *privacy.SpendingKey,
// ) (*txGroups, *accumulativeValues, []*buyBackFromInfo, error) {
// 	prevBlock := blockgen.chain.BestState[shardID].BestBlock
// 	blockHeight := prevBlock.Header.Height + 1

// 	txsToAdd := []metadata.Transaction{}
// 	txToRemove := []metadata.Transaction{}
// 	buySellReqTxs := []metadata.Transaction{}
// 	buyGOVTokensReqTxs := []metadata.Transaction{}
// 	issuingReqTxs := []metadata.Transaction{}
// 	updatingOracleBoardTxs := []metadata.Transaction{}
// 	multiSigsRegistrationTxs := []metadata.Transaction{}
// 	buyBackFromInfos := []*buyBackFromInfo{}
// 	bondsSold := uint64(0)
// 	govTokensSold := uint64(0)
// 	dcbTokensSold := uint64(0)
// 	incomeFromBonds := uint64(0)
// 	incomeFromGOVTokens := uint64(0)
// 	totalFee := uint64(0)
// 	buyBackCoins := uint64(0)
// 	bankPayoutAmount := uint64(0)
// 	govPayoutAmount := uint64(0)

// 	for _, txDesc := range sourceTxns {
// 		tx := txDesc.Tx
// 		txShardID, err := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
// 		if txShardID != shardID || err != nil {
// 			continue
// 		}
// 		// ValidateTransaction vote and propose transaction

// 		// TODO: 0xbunyip need to determine a tx is in privacy format or not
// 		if !tx.ValidateTxByItself(tx.IsPrivacy(), blockgen.chain.config.DataBase, blockgen.chain, shardID) {
// 			txToRemove = append(txToRemove, metadata.Transaction(tx))
// 			continue
// 		}

// 		meta := tx.GetMetadata()
// 		if meta != nil && !meta.ValidateBeforeNewBlock(tx, blockgen.chain, shardID) {
// 			txToRemove = append(txToRemove, metadata.Transaction(tx))
// 			continue
// 		}

// 		switch tx.GetMetadataType() {
// 		case metadata.BuyFromGOVRequestMeta:
// 			{
// 				income, soldAmt, addable := blockgen.checkBuyFromGOVReqTx(shardID, tx, bondsSold)
// 				if !addable {
// 					txToRemove = append(txToRemove, tx)
// 					continue
// 				}
// 				bondsSold += soldAmt
// 				incomeFromBonds += income
// 				buySellReqTxs = append(buySellReqTxs, tx)
// 			}
// 		case metadata.BuyGOVTokenRequestMeta:
// 			{
// 				income, soldAmt, addable := blockgen.checkBuyGOVTokensReqTx(shardID, tx, govTokensSold)
// 				if !addable {
// 					txToRemove = append(txToRemove, tx)
// 					continue
// 				}
// 				govTokensSold += soldAmt
// 				incomeFromGOVTokens += income
// 				buyGOVTokensReqTxs = append(buyGOVTokensReqTxs, tx)
// 			}
// 		case metadata.BuyBackRequestMeta:
// 			{
// 				buyBackFromInfo, addable := blockgen.checkBuyBackReqTx(shardID, tx, buyBackCoins)
// 				if !addable {
// 					txToRemove = append(txToRemove, tx)
// 					continue
// 				}
// 				buyBackCoins += (buyBackFromInfo.buyBackPrice + buyBackFromInfo.value)
// 				buyBackFromInfos = append(buyBackFromInfos, buyBackFromInfo)
// 			}
// 		case metadata.IssuingRequestMeta:
// 			{
// 				addable, newDCBTokensSold := blockgen.checkIssuingReqTx(shardID, tx, dcbTokensSold)
// 				dcbTokensSold = newDCBTokensSold
// 				if !addable {
// 					txToRemove = append(txToRemove, tx)
// 					continue
// 				}
// 				issuingReqTxs = append(issuingReqTxs, tx)
// 			}
// 		case metadata.UpdatingOracleBoardMeta:
// 			{
// 				updatingOracleBoardTxs = append(updatingOracleBoardTxs, tx)
// 			}
// 		case metadata.MultiSigsRegistrationMeta:
// 			{
// 				multiSigsRegistrationTxs = append(multiSigsRegistrationTxs, tx)
// 			}
// 		}

// 		totalFee += tx.GetTxFee()
// 		txsToAdd = append(txsToAdd, tx)
// 		if len(txsToAdd) == common.MaxTxsInBlock {
// 			break
// 		}
// 	}

// 	// TODO(@0xbunyip): cap #tx to common.MaxTxsInBlock
// 	// Process dividend payout for DCB if needed
// 	bankDivTxs, bankPayoutAmount, err := blockgen.processBankDividend(blockHeight, privatekey)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	for _, tx := range bankDivTxs {
// 		txsToAdd = append(txsToAdd, tx)
// 	}

// 	// Process dividend payout for GOV if needed
// 	govDivTxs, govPayoutAmount, err := blockgen.processGovDividend(blockHeight, privatekey)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	for _, tx := range govDivTxs {
// 		txsToAdd = append(txsToAdd, tx)
// 	}

// 	// Process crowdsale for DCB
// 	dcbSaleTxs, removableTxs := blockgen.processCrowdsale(sourceTxns, shardID, privatekey)
// 	for _, tx := range dcbSaleTxs {
// 		txsToAdd = append(txsToAdd, tx)
// 	}
// 	for _, tx := range removableTxs {
// 		txsToAdd = append(txsToAdd, tx)
// 	}

// 	// Build CMB responses
// 	cmbInitRefundTxs, err := blockgen.buildCMBRefund(sourceTxns, shardID, privatekey)
// 	if err != nil {
// 		return nil, nil, nil, err
// 	}
// 	for _, tx := range cmbInitRefundTxs {
// 		txsToAdd = append(txsToAdd, tx)
// 	}

// 	txGroups := &txGroups{
// 		txsToAdd:                 txsToAdd,
// 		txToRemove:               txToRemove,
// 		buySellReqTxs:            buySellReqTxs,
// 		buyGOVTokensReqTxs:       buyGOVTokensReqTxs,
// 		issuingReqTxs:            issuingReqTxs,
// 		updatingOracleBoardTxs:   updatingOracleBoardTxs,
// 		multiSigsRegistrationTxs: multiSigsRegistrationTxs,
// 	}
// 	accumulativeValues := &accumulativeValues{
// 		bondsSold:           bondsSold,
// 		govTokensSold:       govTokensSold,
// 		dcbTokensSold:       dcbTokensSold,
// 		incomeFromBonds:     incomeFromBonds,
// 		incomeFromGOVTokens: incomeFromGOVTokens,
// 		totalFee:            totalFee,
// 		buyBackCoins:        buyBackCoins,
// 		govPayoutAmount:     govPayoutAmount,
// 		bankPayoutAmount:    bankPayoutAmount,
// 	}
// 	return txGroups, accumulativeValues, buyBackFromInfos, nil
// }
