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

func buildResponseForCoin(txRequest *transaction.TxCustomToken, bondID []byte, rt []byte, chainID byte, bondPrices map[string]uint64, saleID []byte, dcbAddress string) (*transaction.TxCustomToken, error) {
	// Mint and send Constant
	meta := txRequest.Metadata.(*metadata.CrowdsaleRequest)
	pks := [][]byte{meta.PaymentAddress.Pk[:], meta.PaymentAddress.Pk[:]}
	tks := [][]byte{meta.PaymentAddress.Tk[:], meta.PaymentAddress.Tk[:]}

	// Get value of the bonds that user sent
	bonds := uint64(0)
	accountDCB, _ := wallet.Base58CheckDeserialize(dcbAddress)
	for _, vout := range txRequest.TxTokenData.Vouts {
		if bytes.Equal(vout.PaymentAddress.Pk[:], accountDCB.KeySet.PaymentAddress.Pk) {
			bonds += vout.Value
		}
	}
	bondPrice := bondPrices[string(bondID)]
	amounts := []uint64{bonds * bondPrice, 0} // TODO(@0xbunyip): use correct unit of price and value here
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

func buildResponseForBond(txRequest *transaction.TxCustomToken, bondID []byte, rt []byte, chainID byte, bondPrices map[string]uint64, unspentTxTokenOuts []transaction.TxTokenVout, saleID []byte, dcbAddress string) (*transaction.TxCustomToken, []transaction.TxTokenVout, error) {
	// TODO:@bunnyip  need to double check here
	// accountDCB, _ := wallet.Base58CheckDeserialize(dcbAddress)
	// Get amount of Constant user sent
	value := uint64(0)
	// userPk := txRequest.Tx.JSPubKey
	userPk := txRequest.Tx.SigPubKey
	// for _, desc := range txRequest.Tx.Descs {
	// 	for _, note := range desc.Note {
	// 		if bytes.Equal(note.Apk[:], accountDCB.KeySet.PaymentAddress.Pk) {
	// 			value += note.Value
	// 		}
	// 	}
	// }
	bondPrice := bondPrices[string(bondID)]
	bonds := value / bondPrice
	sumBonds := uint64(0)
	usedID := 0
	for _, out := range unspentTxTokenOuts {
		usedID += 1
		sumBonds += out.Value
		if sumBonds >= bonds {
			break
		}
	}

	if sumBonds < bonds {
		return nil, unspentTxTokenOuts, fmt.Errorf("Not enough bond to pay")
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
			PaymentAddress: privacy.PaymentAddress{Pk: userPk}, // TODO(@0xbunyip): send to payment address
			Value:          bonds,
		},
	}
	if sumBonds > bonds {
		accountDCB, _ := wallet.Base58CheckDeserialize(dcbAddress)
		txTokenOuts = append(txTokenOuts, transaction.TxTokenVout{
			PaymentAddress: accountDCB.KeySet.PaymentAddress,
			Value:          sumBonds - bonds,
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
	return txToken, unspentTxTokenOuts[usedID:], nil
}
