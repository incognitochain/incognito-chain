package blockchain

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
)

// getTxTokenValue converts total tokens in a tx to Constant
func getTxTokenValue(tokenData transaction.TxTokenData, tokenID []byte, pk []byte, prices map[string]uint64) (uint64, uint64) {
	amount := uint64(0)
	if bytes.Equal(tokenData.PropertyID[:], tokenID) {
		for _, vout := range tokenData.Vouts {
			if bytes.Equal(vout.PaymentAddress.Pk[:], pk) {
				amount += vout.Value
			}
		}
	}
	return amount, amount * prices[string(tokenID)]
}

// getTxValue converts total Constants in a tx to another token
func getTxValue(tx *transaction.Tx, tokenID []byte, pk []byte, prices map[string]uint64) (uint64, uint64) {
	// Get amount of Constant user sent
	value := uint64(0)
	for _, desc := range tx.Descs {
		for _, note := range desc.Note {
			if bytes.Equal(note.Apk[:], pk) {
				value += note.Value
			}
		}
	}
	assetPrice := prices[string(tokenID)]
	amounts := value / assetPrice
	return value, amounts
}

func buildResponseForCoin(
	txRequest *transaction.TxCustomToken,
	amount uint64,
	rt []byte,
	chainID byte,
	saleID []byte,
) (*transaction.TxCustomToken, error) {
	// Mint and send Constant
	meta := txRequest.Metadata.(*metadata.CrowdsaleRequest)
	pks := [][]byte{meta.PaymentAddress.Pk[:], meta.PaymentAddress.Pk[:]}
	tks := [][]byte{meta.PaymentAddress.Tk[:], meta.PaymentAddress.Tk[:]}

	// Get value of the bonds that user sent
	amounts := []uint64{amount, 0}
	tx, err := buildCoinbaseTx(pks, tks, amounts, rt, chainID)
	if err != nil {
		return nil, err
	}
	metaRes := &metadata.CrowdsaleResponse{
		RequestedTxID: &common.Hash{},
		SaleID:        make([]byte, len(saleID)),
	}
	hash := txRequest.Hash()
	copy(metaRes.RequestedTxID[:], hash[:])
	copy(metaRes.SaleID, saleID)
	txToken := &transaction.TxCustomToken{
		Tx:          *tx,
		TxTokenData: transaction.TxTokenData{},
	}
	txToken.Metadata = metaRes
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
		return nil, 0, fmt.Errorf("Not enough tokens to pay in this block")
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
		accountDCB, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		txTokenOuts = append(txTokenOuts, transaction.TxTokenVout{
			PaymentAddress: accountDCB.KeySet.PaymentAddress,
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

func buildResponseForToken(
	txRequest *transaction.TxCustomToken,
	tokenAmount uint64,
	tokenID []byte,
	rt []byte,
	chainID byte,
	unspentTokenMap map[string]([]transaction.TxTokenVout),
	saleID []byte,
	mint bool,
) (*transaction.TxCustomToken, error) {
	unspentTxTokenOuts := unspentTokenMap[string(tokenID)]
	var txToken *transaction.TxCustomToken
	usedID := -1
	err := errors.New("")
	if mint {
		txToken = mintTxToken(tokenAmount, tokenID, txRequest.Tx.JSPubKey)
	} else {
		txToken, usedID, err = transferTxToken(tokenAmount, unspentTxTokenOuts, tokenID, txRequest.Tx.JSPubKey)
		if err != nil {
			return nil, err
		}
	}

	metaRes := &metadata.CrowdsaleResponse{
		RequestedTxID: &common.Hash{},
		SaleID:        make([]byte, len(saleID)),
	}
	hash := txRequest.Hash()
	copy(metaRes.RequestedTxID[:], hash[:])
	copy(metaRes.SaleID, saleID)
	txToken.Metadata = metaRes

	// Update list of token available for next request
	if usedID >= 0 && !mint {
		unspentTokenMap[string(tokenID)] = unspentTxTokenOuts[usedID:]
	}
	return txToken, nil
}
