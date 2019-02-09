package blockchain

import (
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

func buildPaymentForCoin(
	txRequest *transaction.TxCustomToken,
	amount uint64,
	saleID []byte,
	producerPrivateKey *privacy.SpendingKey,
	db database.DatabaseInterface,
) (*transaction.Tx, error) {
	// Mint and send Constant
	meta := txRequest.Metadata.(*metadata.CrowdsaleRequest)
	metaPay := &metadata.CrowdsalePayment{
		RequestedTxID: &common.Hash{},
		SaleID:        make([]byte, len(saleID)),
	}
	hash := txRequest.Hash()
	metaPay.Type = metadata.CrowdsalePaymentMeta
	copy(metaPay.RequestedTxID[:], hash[:])
	copy(metaPay.SaleID, saleID)
	metaPayList := []metadata.Metadata{metaPay}

	fmt.Printf("[db] build CST payment: %d\n", amount)

	amounts := []uint64{amount}
	txs, err := transaction.BuildCoinbaseTxs([]*privacy.PaymentAddress{&meta.PaymentAddress}, amounts, producerPrivateKey, db, metaPayList) // There's only one tx in txs
	if err != nil {
		return nil, err
	}
	return txs[0], nil
}

func transferTxToken(
	tokenAmount uint64,
	unspentTxTokenOuts []transaction.TxTokenVout,
	tokenID common.Hash,
	receiverPk []byte,
	meta metadata.Metadata,
) (*transaction.TxCustomToken, int, error) {
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

	txToken := &transaction.TxCustomToken{
		TxTokenData: transaction.TxTokenData{
			Type:       transaction.CustomTokenTransfer,
			Amount:     sumTokens,
			PropertyID: tokenID,
			Vins:       txTokenIns,
			Vouts:      txTokenOuts,
		},
	}
	txToken.Metadata = meta
	txToken.Type = common.TxCustomTokenType
	return txToken, usedID, nil
}

func mintTxToken(
	tokenAmount uint64,
	tokenID common.Hash,
	receiverPk []byte,
	meta metadata.Metadata,
) *transaction.TxCustomToken {
	txTokenOuts := []transaction.TxTokenVout{
		transaction.TxTokenVout{
			PaymentAddress: privacy.PaymentAddress{Pk: receiverPk}, // TODO(@0xbunyip): send to payment address
			Value:          tokenAmount,
		},
	}
	txToken := &transaction.TxCustomToken{
		TxTokenData: transaction.TxTokenData{
			Type:       transaction.CustomTokenInit,
			Amount:     tokenAmount,
			PropertyID: tokenID,
			Vins:       []transaction.TxTokenVin{},
			Vouts:      txTokenOuts,
		},
	}
	txToken.Metadata = meta
	txToken.Type = common.TxCustomTokenType
	return txToken
}

func buildPaymentForToken(
	txRequest *transaction.Tx,
	tokenAmount uint64,
	tokenID common.Hash,
	unspentTokenMap map[string]([]transaction.TxTokenVout),
	saleID []byte,
	mint bool,
) (*transaction.TxCustomToken, error) {
	var txToken *transaction.TxCustomToken
	var err error
	unspentTxTokenOuts := unspentTokenMap[tokenID.String()]
	usedID := -1
	if len(txRequest.Proof.InputCoins) == 0 {
		return nil, errors.New("Found no sender in request tx")
	}
	pubkey := txRequest.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()

	// Create metadata for crowdsale payment
	metaPay := &metadata.CrowdsalePayment{
		RequestedTxID: &common.Hash{},
		SaleID:        make([]byte, len(saleID)),
	}
	hash := txRequest.Hash()
	copy(metaPay.RequestedTxID[:], hash[:])
	copy(metaPay.SaleID, saleID)
	metaPay.Type = metadata.CrowdsalePaymentMeta

	// Build txcustomtoken
	if mint {
		txToken = mintTxToken(tokenAmount, tokenID, pubkey, metaPay)
	} else {
		fmt.Printf("[db] transferTxToken with unspentTxTokenOuts && tokenAmount:\n%+v\n%d\n", unspentTxTokenOuts, tokenAmount)
		txToken, usedID, err = transferTxToken(tokenAmount, unspentTxTokenOuts, tokenID, pubkey, metaPay)
		if err != nil {
			return nil, err
		}
	}

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
	shardID byte,
	saleID []byte,
	producerPrivateKey *privacy.SpendingKey,
) (metadata.Transaction, error) {
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	saleData := saleDataMap[string(saleID)]
	metaReq := tx.GetMetadata().(*metadata.CrowdsaleRequest)
	priceLimit := metaReq.PriceLimit
	sellingAsset := saleData.SellingAsset
	buyingAsset := saleData.BuyingAsset

	if sellingAsset.IsEqual(&common.ConstantID) {
		txToken := tx.(*transaction.TxCustomToken)
		txResponse, err = buildPaymentForCoin(
			txToken,
			sentAmount,
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
			fmt.Printf("[db] unspentTxTokenOuts: %+v\n%v\n", unspentTxTokenOuts, err)
			if err == nil {
				unspentTokenMap[sellingAsset.String()] = unspentTxTokenOuts
			} else {
				unspentTokenMap[sellingAsset.String()] = []transaction.TxTokenVout{}
			}
		}

		mint := false // Mint DCB token, transfer bonds
		txNormal := tx.(*transaction.Tx)
		txResponse, err = buildPaymentForToken(
			txNormal,
			sentAmount,
			sellingAsset,
			unspentTokenMap,
			saleData.SaleID,
			mint,
		)
	}
	return txResponse, err
}

// processCrowdsaleRequest gets sale data and creates a CrowdsalePayment for a request
func (blockgen *BlkTmplGenerator) processCrowdsaleRequest(
	tx metadata.Transaction,
	txsPayment []metadata.Transaction,
	txsToRemove []metadata.Transaction,
	saleDataMap map[string]*params.SaleData,
	unspentTokenMap map[string][]transaction.TxTokenVout,
	shardID byte,
	producerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, []metadata.Transaction) {
	fmt.Printf("[db] inside txsPayment addr: %p\n", &txsPayment)
	// Create corresponding payment to send selling asset
	meta := tx.GetMetadata()
	metaRequest, ok := meta.(*metadata.CrowdsaleRequest)
	if !ok {
		txsToRemove = append(txsToRemove, tx)
		return txsPayment, txsToRemove
	}
	if _, ok := saleDataMap[string(metaRequest.SaleID)]; !ok {
		saleData, err := blockgen.chain.GetCrowdsaleData(metaRequest.SaleID)
		if err != nil {
			Logger.log.Error(err)
			txsToRemove = append(txsToRemove, tx)
			return txsPayment, txsToRemove
		}

		saleDataMap[string(metaRequest.SaleID)] = saleData
	}

	// Skip payment if either selling or buying asset is offchain (needs confirmation)
	saleData := saleDataMap[string(metaRequest.SaleID)]
	if common.IsOffChainAsset(&saleData.SellingAsset) || common.IsOffChainAsset(&saleData.BuyingAsset) {
		fmt.Println("[db] crowdsale offchain asset")
		return txsPayment, txsToRemove
	}

	txPayment, err := blockgen.buildPaymentForCrowdsale(
		tx,
		saleDataMap,
		unspentTokenMap,
		shardID,
		metaRequest.SaleID,
		producerPrivateKey,
	)
	fmt.Printf("[db] txpayment err: %v\n", err)
	if err != nil {
		Logger.log.Error(err)
		txsToRemove = append(txsToRemove, tx)
	} else if txPayment != nil {
		txsPayment = append(txsPayment, txPayment)
		fmt.Printf("[db] len(txsPayment) after append: %d\n", len(txsPayment))
		fmt.Printf("[db] after append txsPayment addr: %p\n", &txsPayment)
	}
	return txsPayment, txsToRemove
}

// processCrowdsale finds all CrowdsaleRequests and creates Payments for them
func (blockgen *BlkTmplGenerator) processCrowdsale(
	sourceTxns []*metadata.TxDesc,
	shardID byte,
	producerPrivateKey *privacy.SpendingKey,
) ([]metadata.Transaction, []metadata.Transaction) {
	txsToRemove := []metadata.Transaction{}
	txsPayment := []metadata.Transaction{}
	fmt.Printf("[db] outside txsPayment addr: %p\n", &txsPayment)

	// Get unspent bond tx to spend if needed
	unspentTokenMap := map[string]([]transaction.TxTokenVout){}
	saleDataMap := map[string]*params.SaleData{}
	for _, txDesc := range sourceTxns {
		switch txDesc.Tx.GetMetadataType() {
		case metadata.CrowdsaleRequestMeta:
			{
				txsPayment, txsToRemove = blockgen.processCrowdsaleRequest(
					txDesc.Tx,
					txsPayment,
					txsToRemove,
					saleDataMap,
					unspentTokenMap,
					shardID,
					producerPrivateKey,
				)
				fmt.Printf("[db] len(txsPayment) after process: %d\n", len(txsPayment))
			}
		}
	}
	txsPayment = txsPayment[:cap(txsPayment)]
	fmt.Printf("[db] process crowdsale len(txsPayment): %d\n", len(txsPayment))
	fmt.Printf("[db] process crowdsale len(txsToRemove): %d\n", len(txsToRemove))
	return txsPayment, txsToRemove
}

func generateCrowdsalePaymentInstruction(paymentAddress privacy.PaymentAddress, amount uint64, assetID common.Hash) ([][]string, error) {
	inst := &CrowdsalePaymentInstruction{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		AssetID:        AssetID,
	}
	instStr, err := inst.String()
	if err != nil {
		return nil, err
	}
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbPk := keyWalletDCBAccount.KeySet.PaymentAddress.Pk
	dcbShardID := common.GetShardIDFromLastByte(dcbPk[len(dcbPk)-1])
	return [][]string{strconv.Itoa(metadata.CrowdsalePaymentMeta), strconv.Itoa(dcbShardID), instStr}, nil
}
