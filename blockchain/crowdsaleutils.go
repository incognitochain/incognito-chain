package blockchain

import (
	"bytes"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
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
	meta := txRequest.Metadata.(*metadata.CrowdsaleRequest)
	metaPay := &metadata.CrowdsalePayment{
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
	txRequest *transaction.Tx,
	tokenAmount uint64,
	tokenID common.Hash,
	unspentTokenMap map[string]([]transaction.TxTokenVout),
	saleID []byte,
	mint bool,
	change uint64,
) (*transaction.TxCustomToken, error) {
	var txToken *transaction.TxCustomToken
	var err error
	unspentTxTokenOuts := unspentTokenMap[tokenID.String()]
	usedID := -1
	if len(txRequest.Proof.InputCoins) == 0 {
		return nil, errors.New("Found no sender in request tx")
	}
	pubkey := txRequest.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()

	if mint {
		txToken = mintTxToken(tokenAmount, tokenID, pubkey)
	} else {
		txToken, usedID, err = transferTxToken(tokenAmount, unspentTxTokenOuts, tokenID, pubkey)
		if err != nil {
			return nil, err
		}
	}

	metaPay := &metadata.CrowdsalePayment{
		RequestedTxID: &common.Hash{},
		SaleID:        make([]byte, len(saleID)),
	}
	hash := txRequest.Hash()
	copy(metaPay.RequestedTxID[:], hash[:])
	copy(metaPay.SaleID, saleID)
	txToken.Metadata = metaPay

	// Update list of token available for next request
	if usedID >= 0 && !mint {
		unspentTokenMap[tokenID.String()] = unspentTxTokenOuts[usedID:]
	}
	return txToken, nil
}

// buildPaymentForCrowdsale builds CrowdsalePayment tx sending either CST or Token
func (blockgen *BlkTmplGenerator) buildPaymentForCrowdsale(
	tx metadata.Transaction,
	saleDataMap map[string]*params.SaleData,
	unspentTokenMap map[string]([]transaction.TxTokenVout),
	chainID byte,
	saleID []byte,
	producerPrivateKey *privacy.SpendingKey,
) (*transaction.TxCustomToken, error) {
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbPk := keyWalletDCBAccount.KeySet.PaymentAddress.Pk
	saleData := saleDataMap[string(saleID)]
	metaReq := tx.Metadata.(*metadata.CrowdsaleRequest)
	priceLimit := metaReq.PriceLimit
	sellingAsset := saleData.SellingAsset
	buyingAsset := saleData.BuyingAsset

	// Get price for asset
	buyPrice := blockgen.getAssetPrice(chainID, buyingAsset)
	sellPrice := blockgen.getAssetPrice(chainID, sellingAsset)
	if buyPrice == 0 || sellPrice == 0 {
		return nil, errors.New("Missing price data in block")
	}
	// TODO(@0xbunyip): validate sale data in proposal to admit only valid pair of assets
	var txResponse *transaction.TxCustomToken
	err := errors.New("Incorrect assets for crowdsale")

	// Calculate amount of asset to trade for both sides
	// 1 user buy bond => sellingAsset == Bond && buyingAsset == CST
	// 2 user sell bond => sellingAsset == CST && buyingAsset == Bond
	requestAmount := metaReq.Amount
	buyingAmount := uint64(0)
	sellingAmount := uint64(0)
	if metaReq.LimitSellingAsset {
		// TODO(@0xbunyip): check for overflow
		sellingAmount = requestAmount // user requested buying (DCB selling) exactly this amount of asset
		buyingAmount = sellingAmount * sellPrice / buyPrice
	} else {
		buyingAmount = requestAmount // user requested selling (DCB buying) exactly this amount of asset
		sellingAmount = buyingAmount * buyPrice / sellPrice
	}

	// Calculate value of asset sent in request tx
	sentAmount := uint64(0)
	if buyingAsset.IsEqual(&common.ConstantID) {
		_, _, sentAmount = tx.GetUniqueReceiver()
	} else if common.IsBondAsset(&buyingAsset) {
		_, _, sentAmount = tx.GetTokenUniqueReceiver()
	}
	sentAssetValue := sentAmount * buyPrice // in USD

	if sellingAsset.IsEqual(&common.ConstantID) {
		tokenAmount, valuesInConstant, err := getTxTokenValue(tx.TxTokenData, saleData.BuyingAsset[:], dcbPk, prices)
		if err != nil {
			return nil, err
		}
		if tokenAmount > saleData.BuyingAmount || valuesInConstant > saleData.SellingAmount {
			// User sent too many token, reject request
			return nil, errors.New("Crowdsale reached limit")
		}
		if tokenAmount <= 0 || valuesInConstant <= 0 {
			return nil, errors.New("Values sent in request is too low")
		}

		// Check if price limit is not violated
		if valuesInConstant/tokenAmount > priceLimit {
			return nil, errors.New("Price limit violated")
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
	} else if common.IsBondAsset(&sellingAsset) {
		// Get unspent token UTXO to send to user
		if _, ok := unspentTokenMap[sellingAsset.String()]; !ok {
			unspentTxTokenOuts, err := blockgen.chain.GetUnspentTxCustomTokenVout(keyWalletDCBAccount.KeySet, &sellingAsset)
			if err == nil {
				unspentTokenMap[sellingAsset.String()] = unspentTxTokenOuts
			} else {
				unspentTokenMap[sellingAsset.String()] = []transaction.TxTokenVout{}
			}
		}

		// Check if there's still enough asset to trade
		if buyingAmount > saleData.BuyingAmount || sellingAmount > saleData.SellingAmount {
			return nil, errors.New("Crowdsale reached limit")
		}

		// Check if amount if asset sending is enough
		if buyingAmount > sentAmount {
			return nil, errors.New("Values sent in request is too low")
		}
		change := sentAmount - buyingAmount

		// Check if price limit is not violated
		if reqMeta.LimitSellingAsset && sellPrice > priceLimit {
			return nil, errors.Errorf("Price limit violated: %d %d", sellPrice, priceLimit)
		} else if !reqMeta.LimitSellingAsset && buyPrice < priceLimit {
			return nil, errors.Errorf("Price limit violated: %d %d", buyPrice, priceLimit)
		}

		mint := false // Mint DCB token, transfer bonds
		saleData.BuyingAmount -= buyingAmount
		saleData.SellingAmount -= sellingAmount
		txNormal := tx.(*transaction.Tx)
		txResponse, err = buildPaymentForToken(
			txNormal,
			tokensToSend,
			sellingAsset,
			unspentTokenMap,
			saleData.SaleID,
			mint,
			change,
		)
	}
	return txResponse, err
}

// processCrowdsaleRequest gets sale data and creates a CrowdsalePayment for a request
func (blockgen *BlkTmplGenerator) processCrowdsaleRequest(
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
	metaRequest, ok := meta.(*metadata.CrowdsaleRequest)
	if !ok {
		txsToRemove = append(txsToRemove, tx)
		return
	}
	if _, ok := saleDataMap[string(metaRequest.SaleID)]; !ok {
		saleData, err := blockgen.chain.GetCrowdsaleData(metaRequest.SaleID)
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
		return
	}

	txPayment, err := blockgen.buildPaymentForCrowdsale(
		tx,
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

// processCrowdsale finds all CrowdsaleRequests and creates Payments for them
func (blockgen *BlkTmplGenerator) processCrowdsale(
	sourceTxns []*metadata.TxDesc,
	chainID byte,
	producerPrivateKey *privacy.SpendingKey,
) ([]*transaction.TxCustomToken, []metadata.Transaction) {
	txsToRemove := []metadata.Transaction{}
	txsPayment := []*transaction.TxCustomToken{}

	// Get unspent bond tx to spend if needed
	unspentTokenMap := map[string]([]transaction.TxTokenVout){}
	saleDataMap := map[string]*params.SaleData{}
	for _, txDesc := range sourceTxns {
		switch txDesc.Tx.GetMetadataType() {
		case metadata.CrowdsaleRequestMeta:
			{
				blockgen.processCrowdsaleRequest(
					txDesc.Tx,
					txsPayment,
					txsToRemove,
					saleDataMap,
					unspentTokenMap,
					chainID,
					producerPrivateKey,
				)
			}
		}
	}
	return txsPayment, txsToRemove
}

func (blockgen *BlkTmplGenerator) getAssetPrice(chainID byte, assetID common.Hash) uint64 {
	price := uint64(0)
	if common.IsBondAsset(&assetID) {
		if blockgen.chain.BestState[chainID].BestBlock.Header.Oracle.Bonds != nil {
			price = blockgen.chain.BestState[chainID].BestBlock.Header.Oracle.Bonds[assetID.String()]
		}
	} else if blockgen.chain.BestState[chainID].BestBlock.Header.Oracle != nil {
		oracle := blockgen.chain.BestState[chainID].BestBlock.Header.Oracle
		if assetID.IsEqual(&common.ConstantID) {
			price = oracle.Constant
		} else if assetID.IsEqual(&common.DCBTokenID) {
			price = oracle.DCBToken
		} else if assetID.IsEqual(&common.GOVTokenID) {
			price = oracle.GOVToken
		} else if assetID.IsEqual(&common.ETHAssetID) {
			price = oracle.ETH
		} else if assetID.IsEqual(&common.BTCAssetID) {
			price = oracle.BTC
		}
	}
	return price
}
