package blockchain

import (
	"bytes"
	"errors"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
)

// getTxTokenValue converts total tokens in a tx to Constant
func getTxTokenValue(tokenData transaction.TxTokenData, tokenID []byte, pk []byte, prices map[string]uint64) (uint64, uint64, error) {
	amount := uint64(0)
	if bytes.Equal(tokenData.PropertyID[:], tokenID) {
		for _, vout := range tokenData.Vouts {
			if bytes.Equal(vout.PaymentAddress.Pk[:], pk) {
				amount += vout.Value
			}
		}
	}
	if price, ok := prices[string(tokenID)]; ok {
		return amount, amount * price, nil
	}
	return 0, 0, errors.New("Miss price")
}

// getTxValue converts total Constants in a tx to another token
func getTxValue(tx *transaction.Tx, tokenID []byte, pk []byte, prices map[string]uint64) (uint64, uint64) {
	if tx.Proof == nil || len(tx.Proof.OutputCoins) == 0 {
		return 0, 0
	}
	// Get amount of Constant user sent
	value := uint64(0)
	for _, coin := range tx.Proof.OutputCoins {
		if bytes.Equal(coin.CoinDetails.PublicKey.Compress(), pk) {
			value += coin.CoinDetails.Value
		}
	}
	assetPrice := prices[string(tokenID)]
	amounts := value / assetPrice
	return value, amounts
}

func buildPaymentForCoin(
	txRequest *transaction.TxCustomToken,
	amount uint64,
	saleID []byte,
	producerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (*transaction.TxCustomToken, error) {
	// Mint and send Constant
	meta := txRequest.Metadata.(*metadata.ReserveRequest)
	metaPay := &metadata.ReservePayment{
		RequestedTxID: &common.Hash{},
		SaleID:        make([]byte, len(saleID)),
	}
	hash := txRequest.Hash()
	copy(metaPay.RequestedTxID[:], hash[:])
	copy(metaPay.SaleID, saleID)
	metaPayList := []metadata.Metadata{metaPay}

	amounts := []uint64{amount}
	txs, err := transaction.BuildCoinbaseTxs([]*privacy.PaymentAddress{&meta.PaymentAddress}, amounts, producerPrivateKey, db, metaPayList) // There's only one tx in txs
	if err != nil {
		return nil, err
	}

	txToken := &transaction.TxCustomToken{
		Tx:          *(txs[0]),
		TxTokenData: transaction.TxTokenData{},
	}
	return txToken, nil
}

func transferTxToken(tokenAmount uint64, unspentTxTokenOuts []transaction.TxTokenVout, tokenID, receiverPk []byte) (*transaction.TxCustomToken, int, error) {
	sumTokens := uint64(0)
	usedID := 0
	for _, out := range unspentTxTokenOuts {
		usedID += 1
		sumTokens += out.Value
		if sumTokens >= tokenAmount {
			break
		}
	}

	if sumTokens < tokenAmount {
		return nil, 0, errors.New("Not enough tokens to pay in this block")
	}

	txTokenIns := []transaction.TxTokenVin{}
	for i := 0; i < usedID; i += 1 {
		out := unspentTxTokenOuts[i]
		item := transaction.TxTokenVin{
			PaymentAddress:  out.PaymentAddress,
			TxCustomTokenID: out.GetTxCustomTokenID(),
			VoutIndex:       out.GetIndex(),
		}

		// No need for signature to spend tokens in DCB's account
		txTokenIns = append(txTokenIns, item)
	}
	txTokenOuts := []transaction.TxTokenVout{
		transaction.TxTokenVout{
			PaymentAddress: privacy.PaymentAddress{Pk: receiverPk}, // TODO(@0xbunyip): send to payment address
			Value:          tokenAmount,
		},
	}
	if sumTokens > tokenAmount {
		keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		txTokenOuts = append(txTokenOuts, transaction.TxTokenVout{
			PaymentAddress: keyWalletDCBAccount.KeySet.PaymentAddress,
			Value:          sumTokens - tokenAmount,
		})
	}

	propertyID := common.Hash{}
	copy(propertyID[:], tokenID)
	txToken := &transaction.TxCustomToken{
		TxTokenData: transaction.TxTokenData{
			Type:       transaction.CustomTokenTransfer,
			Amount:     sumTokens,
			PropertyID: propertyID,
			Vins:       txTokenIns,
			Vouts:      txTokenOuts,
		},
	}
	return txToken, usedID, nil
}

func mintTxToken(tokenAmount uint64, tokenID, receiverPk []byte) *transaction.TxCustomToken {
	txTokenOuts := []transaction.TxTokenVout{
		transaction.TxTokenVout{
			PaymentAddress: privacy.PaymentAddress{Pk: receiverPk}, // TODO(@0xbunyip): send to payment address
			Value:          tokenAmount,
		},
	}
	propertyID := common.Hash{}
	copy(propertyID[:], tokenID)
	txToken := &transaction.TxCustomToken{
		TxTokenData: transaction.TxTokenData{
			Type:       transaction.CustomTokenInit,
			Amount:     tokenAmount,
			PropertyID: propertyID,
			Vins:       []transaction.TxTokenVin{},
			Vouts:      txTokenOuts,
		},
	}
	return txToken
}

func buildPaymentForToken(
	txRequest *transaction.TxCustomToken,
	tokenAmount uint64,
	tokenID []byte,
	unspentTokenMap map[string]([]transaction.TxTokenVout),
	saleID []byte,
	mint bool,
) (*transaction.TxCustomToken, error) {
	var txToken *transaction.TxCustomToken
	var err error
	unspentTxTokenOuts := unspentTokenMap[string(tokenID)]
	usedID := -1
	if len(txRequest.Tx.Proof.InputCoins) == 0 {
		return nil, errors.New("Found no sender in request tx")
	}
	pubkey := txRequest.Tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()

	if mint {
		txToken = mintTxToken(tokenAmount, tokenID, pubkey)
	} else {
		txToken, usedID, err = transferTxToken(tokenAmount, unspentTxTokenOuts, tokenID, pubkey)
		if err != nil {
			return nil, err
		}
	}

	metaPay := &metadata.ReservePayment{
		RequestedTxID: &common.Hash{},
		SaleID:        make([]byte, len(saleID)),
	}
	hash := txRequest.Hash()
	copy(metaPay.RequestedTxID[:], hash[:])
	copy(metaPay.SaleID, saleID)
	txToken.Metadata = metaPay

	// Update list of token available for next request
	if usedID >= 0 && !mint {
		unspentTokenMap[string(tokenID)] = unspentTxTokenOuts[usedID:]
	}
	return txToken, nil
}

func (blockgen *BlkTmplGenerator) buildPaymentForReserve(
	tx *transaction.TxCustomToken,
	saleDataMap map[string]*params.SaleData,
	unspentTokenMap map[string]([]transaction.TxTokenVout),
	chainID byte,
	saleID []byte,
	producerPrivateKey *privacy.SpendingKey,
) (*transaction.TxCustomToken, error) {
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbPk := keyWalletDCBAccount.KeySet.PaymentAddress.Pk
	saleData := saleDataMap[string(saleID)]

	// Get price for asset
	prices := make(map[string]uint64)
	if blockgen.chain.BestState[chainID].BestBlock.Header.Oracle != nil && len(blockgen.chain.BestState[chainID].BestBlock.Header.Oracle.Bonds) > 0 {
		prices = blockgen.chain.BestState[chainID].BestBlock.Header.Oracle.Bonds
	}
	if len(prices) == 0 {
		return nil, errors.New("Missing bonds data in block")
	}
	// TODO(@0xbunyip): validate sale data in proposal to admit only valid pair of assets
	txResponse := &transaction.TxCustomToken{}
	err := errors.New("Incorrect assets for crowdsale")
	sellingAsset := saleData.SellingAsset

	if sellingAsset.IsEqual(&common.ConstantID) {
		tokenAmount, valuesInConstant, err := getTxTokenValue(tx.TxTokenData, saleData.BuyingAsset[:], dcbPk, prices)
		if err != nil {
			return nil, err
		}
		if tokenAmount > saleData.BuyingAmount || valuesInConstant > saleData.SellingAmount {
			// User sent too many token, reject request
			return nil, errors.New("Reserve reached limit")
		}
		// Update amount of buying/selling asset of the crowdsale
		saleData.BuyingAmount -= tokenAmount
		saleData.SellingAmount -= valuesInConstant
		txResponse, err = buildPaymentForCoin(
			tx,
			valuesInConstant,
			saleData.SaleID,
			producerPrivateKey,
			blockgen.chain.GetDatabase(),
		)
		if err != nil {
			return nil, err
		}

	} else if common.IsBondAsset(&sellingAsset) || common.IsDCBTokenAsset(&sellingAsset) {
		// Get unspent token UTXO to send to user
		if _, ok := unspentTokenMap[string(sellingAsset[:])]; !ok {
			unspentTxTokenOuts, err := blockgen.chain.GetUnspentTxCustomTokenVout(keyWalletDCBAccount.KeySet, &sellingAsset)
			if err == nil {
				unspentTokenMap[string(sellingAsset[:])] = unspentTxTokenOuts
			} else {
				unspentTokenMap[string(sellingAsset[:])] = []transaction.TxTokenVout{}
			}
		}

		// Calculate amount of token to send
		sentAmount := uint64(0)
		tokensToSend := uint64(0)
		if common.IsOffChainAsset(&saleData.BuyingAsset) {
			meta, ok := tx.GetMetadata().(*metadata.ReserveResponse)
			if !ok {
				return nil, errors.New("Error parsing crowdsale response")
			}
			_, _, _, txReq, _ := blockgen.chain.GetTransactionByHash(meta.RequestedTxID)
			metaReq, ok := txReq.GetMetadata().(*metadata.ReserveRequest)
			if !ok {
				return nil, errors.New("Error getting crowdsale request")
			}
			if metaReq.Amount.IsUint64() {
				sentAmount = metaReq.Amount.Uint64() // TODO(@0xbunyip): support buy and sell amount as big.Int
				tokenPrice := blockgen.chain.BestState[0].BestBlock.Header.Oracle.Bonds[string(saleData.SellingAsset[:])]
				tokensToSend = sentAmount * metaReq.AssetPrice / tokenPrice
			}

		} else {
			sentAmount, tokensToSend = getTxValue(&tx.Tx, sellingAsset[:], dcbPk, prices)
		}

		mint := bytes.Equal(sellingAsset[:], common.DCBTokenID[:]) // Mint DCB token, transfer bonds
		if sentAmount > saleData.BuyingAmount || tokensToSend > saleData.SellingAmount {
			return nil, errors.New("Reserve reached limit")
		}
		saleData.BuyingAmount -= sentAmount
		saleData.SellingAmount -= tokensToSend
		txResponse, err = buildPaymentForToken(
			tx,
			tokensToSend,
			sellingAsset[:],
			unspentTokenMap,
			saleData.SaleID,
			mint,
		)
	}
	return txResponse, err
}

func (blockgen *BlkTmplGenerator) processReserveResponse(
	tx metadata.Transaction,
	txsPayment []*transaction.TxCustomToken,
	txsToRemove []metadata.Transaction,
	saleDataMap map[string]*params.SaleData,
	unspentTokenMap map[string][]transaction.TxTokenVout,
	chainID byte,
	producerPrivateKey *privacy.SpendingKey,
	respCounter map[string]int,
) {
	// Create corresponding response to send selling asset
	// Get buying and selling asset from current sale
	meta := tx.GetMetadata()
	if meta == nil {
		txsToRemove = append(txsToRemove, tx)
		return
	}
	metaResponse, ok := meta.(*metadata.ReserveResponse)
	if !ok {
		txsToRemove = append(txsToRemove, tx)
		return
	}
	_, _, _, txReq, err := blockgen.chain.GetTransactionByHash(metaResponse.RequestedTxID)
	if err != nil {
		return
	}
	metaRequest, ok := txReq.GetMetadata().(*metadata.ReserveRequest)
	if !ok {
		return
	}

	if _, ok := saleDataMap[string(metaRequest.SaleID)]; !ok {
		saleData, err := blockgen.chain.GetReserveData(metaRequest.SaleID)
		if err != nil {
			txsToRemove = append(txsToRemove, tx)
			return
		}

		saleDataMap[string(metaRequest.SaleID)] = saleData
	}

	// Create payment if number of response is exactly enough
	txHashes, err := blockgen.chain.GetReserveTxs(metaResponse.RequestedTxID[:])
	count := 0
	for _, txHash := range txHashes {
		hash, _ := (&common.Hash{}).NewHash(txHash)
		_, _, _, txOld, _ := blockgen.chain.GetTransactionByHash(hash)
		if txOld.GetMetadataType() == metadata.ReserveResponseMeta {
			count += 1
		}
	}

	// TODO(@0xbunyip): use separate crowdsale and loan response requirement
	responseRequired := blockgen.chain.GetDCBParams().MinLoanResponseRequire
	respFound := count + respCounter[string(metaResponse.RequestedTxID[:])] + 1
	if respFound == int(responseRequired) {
		txRequest, _ := tx.(*transaction.TxCustomToken)
		txPayment, err := blockgen.buildPaymentForReserve(
			txRequest,
			saleDataMap,
			unspentTokenMap,
			chainID,
			metaRequest.SaleID,
			producerPrivateKey,
		)
		if err != nil {
			txsToRemove = append(txsToRemove, tx)
			return
		} else if txPayment != nil {
			txsPayment = append(txsPayment, txPayment)
		}
	}
	respCounter[string(metaResponse.RequestedTxID[:])] += 1
}

func (blockgen *BlkTmplGenerator) processReserveRequest(
	tx metadata.Transaction,
	txsPayment []*transaction.TxCustomToken,
	txsToRemove []metadata.Transaction,
	saleDataMap map[string]*params.SaleData,
	unspentTokenMap map[string][]transaction.TxTokenVout,
	chainID byte,
	producerPrivateKey *privacy.SpendingKey,
) {
	// Create corresponding payment to send selling asset
	meta := tx.GetMetadata()
	if meta == nil {
		txsToRemove = append(txsToRemove, tx)
		return
	}
	metaRequest, ok := meta.(*metadata.ReserveRequest)
	if !ok {
		txsToRemove = append(txsToRemove, tx)
		return
	}
	if _, ok := saleDataMap[string(metaRequest.SaleID)]; !ok {
		saleData, err := blockgen.chain.GetReserveData(metaRequest.SaleID)
		if err != nil {
			Logger.log.Error(err)
			txsToRemove = append(txsToRemove, tx)
			return
		}

		saleDataMap[string(metaRequest.SaleID)] = saleData
	}

	// Skip payment if either selling or buying asset is offchain (needs confirmation)
	saleData := saleDataMap[string(metaRequest.SaleID)]
	if common.IsOffChainAsset(&saleData.SellingAsset) || common.IsOffChainAsset(&saleData.BuyingAsset) {
		// Save asset price if the either buying or selling asset is offchain
		// TODO(@0xbunyip): get price of offchain asset instead of bonds here
		assetID := []byte{}
		if common.IsOffChainAsset(&saleData.SellingAsset) {
			assetID = saleData.SellingAsset[:]
		} else {
			assetID = saleData.BuyingAsset[:]
		}
		metaRequest.AssetPrice = blockgen.chain.BestState[0].BestBlock.Header.Oracle.Bonds[string(assetID)]
		return
	}

	txRequest, _ := tx.(*transaction.TxCustomToken)
	txPayment, err := blockgen.buildPaymentForReserve(
		txRequest,
		saleDataMap,
		unspentTokenMap,
		chainID,
		metaRequest.SaleID,
		producerPrivateKey,
	)
	if err != nil {
		Logger.log.Error(err)
		txsToRemove = append(txsToRemove, tx)
	} else if txPayment != nil {
		txsPayment = append(txsPayment, txPayment)
	}
}

func (blockgen *BlkTmplGenerator) processReserve(
	sourceTxns []*metadata.TxDesc,
	chainID byte,
	producerPrivateKey *privacy.SpendingKey,
) ([]*transaction.TxCustomToken, []metadata.Transaction) {
	txsToRemove := []metadata.Transaction{}
	txsPayment := []*transaction.TxCustomToken{}

	// Get unspent bond tx to spend if needed
	unspentTokenMap := map[string]([]transaction.TxTokenVout){}
	saleDataMap := map[string]*params.SaleData{}
	respCounter := map[string]int{}
	for _, txDesc := range sourceTxns {
		switch txDesc.Tx.GetMetadataType() {
		case metadata.ReserveRequestMeta:
			{
				blockgen.processReserveRequest(
					txDesc.Tx,
					txsPayment,
					txsToRemove,
					saleDataMap,
					unspentTokenMap,
					chainID,
					producerPrivateKey,
				)
			}
		case metadata.ReserveResponseMeta:
			{
				blockgen.processReserveResponse(
					txDesc.Tx,
					txsPayment,
					txsToRemove,
					saleDataMap,
					unspentTokenMap,
					chainID,
					producerPrivateKey,
					respCounter,
				)
			}
		}

	}
	return txsPayment, txsToRemove
}
