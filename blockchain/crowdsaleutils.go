package blockchain

import (
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/cashec"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wallet"
	"github.com/pkg/errors"
)

func buildPaymentForCoin(
	receiverAddress privacy.PaymentAddress,
	amount uint64,
	saleID []byte,
	producerPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
) (*transaction.Tx, error) {
	// Mint and send Constant
	metaPay := &metadata.CrowdsalePayment{
		SaleID: make([]byte, len(saleID)),
	}
	metaPay.Type = metadata.CrowdsalePaymentMeta
	copy(metaPay.SaleID, saleID)
	metaPayList := []metadata.Metadata{metaPay}

	amounts := []uint64{amount}
	txs, err := transaction.BuildCoinbaseTxs(
		[]*privacy.PaymentAddress{&receiverAddress},
		amounts,
		producerPrivateKey,
		db,
		metaPayList,
	)
	if err != nil {
		return nil, err
	}
	return txs[0], nil
}

func transferTxToken(
	tokenAmount uint64,
	unspentTxTokenOuts []transaction.TxTokenVout,
	tokenID common.Hash,
	receiverAddress privacy.PaymentAddress,
	meta metadata.Metadata,
	producerPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	shardID byte,
) (*transaction.TxCustomToken, int, error) {
	sumTokens := uint64(0)
	usedID := 0
	// Choose input token UTXO
	for _, out := range unspentTxTokenOuts {
		usedID += 1
		sumTokens += out.Value
		if sumTokens >= tokenAmount {
			break
		}
	}

	if sumTokens < tokenAmount {
		return nil, 0, errors.New("not enough tokens to pay in this block")
	}

	// Build list of inputs and outputs
	txTokenIns := []transaction.TxTokenVin{}
	for i := 0; i < usedID; i += 1 {
		out := unspentTxTokenOuts[i]

		// Sign dummy signature using miner's key
		keySet := &cashec.KeySet{PrivateKey: *producerPrivateKey}
		signature, err := keySet.Sign(out.Hash()[:])
		if err != nil {
			return nil, 0, err
		}

		item := transaction.TxTokenVin{
			PaymentAddress:  out.PaymentAddress,
			TxCustomTokenID: out.GetTxCustomTokenID(),
			VoutIndex:       out.GetIndex(),
			Signature:       base58.Base58Check{}.Encode(signature, 0),
		}

		// No need for signature to spend tokens in DCB's account
		txTokenIns = append(txTokenIns, item)
	}
	txTokenOuts := []transaction.TxTokenVout{
		transaction.TxTokenVout{
			PaymentAddress: receiverAddress,
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

	// Build token params
	tokenParams := &transaction.CustomTokenParamTx{
		PropertyID:  tokenID.String(),
		TokenTxType: transaction.CustomTokenTransfer,
		Amount:      sumTokens,
		Receiver:    txTokenOuts,
		Mintable:    true,
	}
	tokenParams.SetVins(txTokenIns)
	tokenParams.SetVinsAmount(sumTokens)

	// Build TxCustomToken from token params
	txToken := &transaction.TxCustomToken{}
	err := txToken.Init(
		producerPrivateKey,
		nil,
		nil,
		0,
		tokenParams,
		db,
		meta,
		false,
		shardID,
	)
	if err != nil {
		return nil, 0, err
	}

	//txToken := &transaction.TxCustomToken{
	//	TxTokenData: transaction.TxTokenData{
	//		Type:       transaction.CustomTokenTransfer,
	//		Amount:     sumTokens,
	//		PropertyID: tokenID,
	//		Vins:       txTokenIns,
	//		Vouts:      txTokenOuts,
	//	},
	//}
	//txToken.Metadata = meta
	//txToken.Type = common.TxCustomTokenType
	return txToken, usedID, nil
}

func mintTxToken(
	tokenAmount uint64,
	tokenID common.Hash,
	receiverAddress privacy.PaymentAddress,
	meta metadata.Metadata,
) *transaction.TxCustomToken {
	txTokenOuts := []transaction.TxTokenVout{
		transaction.TxTokenVout{
			PaymentAddress: receiverAddress,
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
	receiverAddress privacy.PaymentAddress,
	tokenAmount uint64,
	tokenID common.Hash,
	unspentTokens map[string]([]transaction.TxTokenVout),
	saleID []byte,
	mint bool,
	producerPrivateKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	shardID byte,
) (*transaction.TxCustomToken, error) {
	var txToken *transaction.TxCustomToken
	var err error
	unspentTxTokenOuts := unspentTokens[tokenID.String()]
	usedID := -1

	// Create metadata for crowdsale payment
	metaPay := &metadata.CrowdsalePayment{
		SaleID: make([]byte, len(saleID)),
	}
	copy(metaPay.SaleID, saleID)
	metaPay.Type = metadata.CrowdsalePaymentMeta

	// Build txcustomtoken
	if mint {
		txToken = mintTxToken(tokenAmount, tokenID, receiverAddress, metaPay)
	} else {
		// fmt.Printf("[db] transferTxToken with unspentTxTokenOuts && tokenAmount: %+v %d\n", unspentTxTokenOuts, tokenAmount)
		txToken, usedID, err = transferTxToken(
			tokenAmount,
			unspentTxTokenOuts,
			tokenID,
			receiverAddress,
			metaPay,
			producerPrivateKey,
			db,
			shardID,
		)
		if err != nil {
			return nil, err
		}
	}

	// Update list of token available for next request
	if usedID >= 0 && !mint {
		unspentTokens[tokenID.String()] = unspentTxTokenOuts[usedID:]
	}
	return txToken, nil
}

// buildPaymentForCrowdsale builds CrowdsalePayment tx sending either CST or Token
func (blockgen *BlkTmplGenerator) buildPaymentForCrowdsale(
	inst string,
	unspentTokens map[string]([]transaction.TxTokenVout),
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) ([]metadata.Transaction, error) {
	paymentInst, err := component.ParseCrowdsalePaymentInstruction(inst)
	if err != nil {
		return nil, err
	}
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	saleID := paymentInst.SaleID
	assetID := &paymentInst.AssetID

	var txResponse metadata.Transaction
	if common.IsConstantAsset(assetID) {
		txResponse, err = buildPaymentForCoin(
			paymentInst.PaymentAddress,
			paymentInst.Amount,
			saleID,
			producerPrivateKey,
			blockgen.chain.GetDatabase(),
		)
	} else if common.IsBondAsset(assetID) {
		// Get unspent token UTXO to send to user
		if _, ok := unspentTokens[assetID.String()]; !ok {
			unspentTxTokenOuts, err := blockgen.chain.GetUnspentTxCustomTokenVout(keyWalletDCBAccount.KeySet, assetID)
			// fmt.Printf("[db] unspentTxTokenOuts: %+v\n%v\n", unspentTxTokenOuts, err)
			if err == nil {
				unspentTokens[assetID.String()] = unspentTxTokenOuts
			} else {
				unspentTokens[assetID.String()] = []transaction.TxTokenVout{}
			}
		}

		mint := false // Mint DCB token, transfer bonds
		txResponse, err = buildPaymentForToken(
			paymentInst.PaymentAddress,
			paymentInst.Amount,
			*assetID,
			unspentTokens,
			saleID,
			mint,
			producerPrivateKey,
			blockgen.chain.GetDatabase(),
			shardID,
		)
	}
	if err != nil {
		return nil, err
	}
	return []metadata.Transaction{txResponse}, err
}

func generateCrowdsalePaymentInstruction(
	paymentAddress privacy.PaymentAddress,
	amount uint64,
	assetID *common.Hash,
	saleID []byte,
	sentAmount uint64,
	updateSale bool,
) ([][]string, error) {
	inst := &component.CrowdsalePaymentInstruction{
		PaymentAddress: paymentAddress,
		Amount:         amount,
		AssetID:        *assetID,
		SaleID:         saleID,
		SentAmount:     sentAmount,
		UpdateSale:     updateSale,
	}
	instStr, err := inst.String()
	if err != nil {
		return nil, err
	}
	keyWalletDCBAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
	dcbPk := keyWalletDCBAccount.KeySet.PaymentAddress.Pk
	dcbShardID := common.GetShardIDFromLastByte(dcbPk[len(dcbPk)-1])
	paymentInst := []string{
		strconv.Itoa(metadata.CrowdsalePaymentMeta),
		strconv.Itoa(int(dcbShardID)),
		instStr,
	}
	return [][]string{paymentInst}, nil
}
