package blockchain

import (
	"bytes"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wallet"
)

// getTxTokenValue converts total tokens in a tx to Constant
func getTxTokenValue(tokenData transaction.TxTokenData, tokenID []byte, pk []byte, prices map[string]uint64) uint64 {
	amount := uint64(0)
	if bytes.Equal(tokenData.PropertyID[:], tokenID) {
		for _, vout := range tokenData.Vouts {
			if bytes.Equal(vout.PaymentAddress.Pk[:], pk) {
				amount += vout.Value
			}
		}
	}
	return amount * prices[string(tokenID)]
}

// getTxValue converts total Constants in a tx to another token
func getTxValue(tx *transaction.Tx, tokenID []byte, pk []byte, prices map[string]uint64) uint64 {
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
	return amounts
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

func buildResponseForToken(
	txRequest *transaction.TxCustomToken,
	tokens uint64,
	tokenID []byte,
	rt []byte,
	chainID byte,
	unspentTokenMap map[string]([]transaction.TxTokenVout),
	saleID []byte,
	mint bool,
) (*transaction.TxCustomToken, error) {
	sumTokens := uint64(0)
	usedID := 0
	unspentTxTokenOuts := unspentTokenMap[string(tokenID)]
	for _, out := range unspentTxTokenOuts {
		usedID += 1
		sumTokens += out.Value
		if sumTokens >= tokens {
			break
		}
	}

	if sumTokens < tokens {
		return nil, fmt.Errorf("Not enough bond to pay")
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
			PaymentAddress: privacy.PaymentAddress{Pk: txRequest.Tx.JSPubKey}, // TODO(@0xbunyip): send to payment address
			Value:          tokens,
		},
	}
	if sumTokens > tokens {
		accountDCB, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		txTokenOuts = append(txTokenOuts, transaction.TxTokenVout{
			PaymentAddress: accountDCB.KeySet.PaymentAddress,
			Value:          sumTokens - tokens,
		})
	}

	metaRes := &metadata.CrowdsaleResponse{
		RequestedTxID: &common.Hash{},
		SaleID:        make([]byte, len(saleID)),
	}
	hash := txRequest.Hash()
	copy(metaRes.RequestedTxID[:], hash[:])
	copy(metaRes.SaleID, saleID)
	txToken := &transaction.TxCustomToken{
		TxTokenData: transaction.TxTokenData{
			Type:  transaction.CustomTokenTransfer,
			Vins:  txTokenIns,
			Vouts: txTokenOuts,
		},
	}
	txToken.Metadata = metaRes

	// Update list of token available for next request
	unspentTokenMap[string(tokenID)] = unspentTxTokenOuts[usedID:]
	return txToken, nil
}
