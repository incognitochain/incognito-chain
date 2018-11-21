package transaction

import (
	"bytes"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

type TxBuySellDCBResponse struct {
	*TxCustomToken // fee + amount to pay for bonds/constant
	RequestedTxID  *common.Hash
}

func BuildResponseForCoin(txRequest *TxBuySellRequest, bondID string, rt []byte, chainID byte, bondPrices map[string]uint64) (*TxBuySellDCBResponse, error) {
	// Mint and send Constant
	pks := [][]byte{txRequest.PaymentAddress.Pk[:], txRequest.PaymentAddress.Pk[:]}
	tks := [][]byte{txRequest.PaymentAddress.Tk[:], txRequest.PaymentAddress.Tk[:]}

	// Get value of the bonds that user sent
	bonds := uint64(0)
	for _, vout := range txRequest.TxTokenData.Vouts {
		if bytes.Equal(vout.PaymentAddress.Pk[:], common.DCBAddress) {
			bonds += vout.Value
		}
	}
	bondPrice := bondPrices[bondID]
	amounts := []uint64{bonds * bondPrice, 0} // TODO(@0xbunyip): use correct unit of price and value here
	tx, err := BuildCoinbaseTx(pks, tks, amounts, rt, chainID, common.TxBuySellDCBResponse)
	if err != nil {
		return nil, err
	}
	txToken := &TxCustomToken{
		Tx:          *tx,
		TxTokenData: TxTokenData{},
	}
	txResponse := &TxBuySellDCBResponse{
		TxCustomToken: txToken,
		RequestedTxID: txRequest.Hash(),
	}
	return txResponse, nil
}

func BuildResponseForBond(txRequest *TxBuySellRequest, bondID string, rt []byte, chainID byte, bondPrices map[string]uint64, unspentTxTokenOuts []TxTokenVout) (*TxBuySellDCBResponse, []TxTokenVout, error) {
	// Get amount of Constant user sent
	value := uint64(0)
	userPk := privacy.PublicKey{}
	for _, desc := range txRequest.Tx.Descs {
		for _, note := range desc.Note {
			if bytes.Equal(note.Apk[:], common.DCBAddress) {
				value += note.Value
				userPk = note.Apk
			}
		}
	}
	bondPrice := bondPrices[bondID]
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

	txTokenIns := []TxTokenVin{}
	for i := 0; i < usedID; i += 1 {
		out := unspentTxTokenOuts[i]
		item := TxTokenVin{
			PaymentAddress:  out.PaymentAddress,
			TxCustomTokenID: out.GetTxCustomTokenID(),
			VoutIndex:       out.GetIndex(),
		}

		// No need for signature to spend tokens in DCB's account
		txTokenIns = append(txTokenIns, item)
	}
	txTokenOuts := []TxTokenVout{
		TxTokenVout{
			PaymentAddress: privacy.PaymentAddress{Pk: userPk}, // TODO(@0xbunyip): send to payment address
			Value:          bonds,
		},
	}
	if sumBonds > bonds {
		txTokenOuts = append(txTokenOuts, TxTokenVout{
			PaymentAddress: privacy.PaymentAddress{Pk: common.DCBAddress},
			Value:          sumBonds - bonds,
		})
	}

	txToken := &TxCustomToken{
		TxTokenData: TxTokenData{
			Type:  CustomTokenTransfer,
			Vins:  txTokenIns,
			Vouts: txTokenOuts,
		},
	}
	txResponse := &TxBuySellDCBResponse{
		TxCustomToken: txToken,
		RequestedTxID: txRequest.Hash(),
	}
	return txResponse, unspentTxTokenOuts[usedID:], nil
}

func (tx *TxBuySellDCBResponse) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()
	record += string(tx.RequestedTxID[:])

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxBuySellDCBResponse) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}
	return true
}

func (tx *TxBuySellDCBResponse) GetType() string {
	return tx.Tx.Type
}

func (tx *TxBuySellDCBResponse) GetTxVirtualSize() uint64 {
	// TODO: calculate
	return 0
}

func (tx *TxBuySellDCBResponse) GetSenderAddrLastByte() byte {
	return tx.Tx.AddressLastByte
}

func (tx *TxBuySellDCBResponse) GetTxFee() uint64 {
	return tx.Tx.Fee
}
