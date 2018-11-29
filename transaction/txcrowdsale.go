package transaction

import (
	"bytes"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/wallet"
)

type TxBuySellDCBResponse struct {
	*TxCustomToken // fee + amount to pay for bonds/constant
	RequestedTxID  *common.Hash
	SaleID         []byte
}

func BuildResponseForCoin(txRequest *TxBuySellRequest, bondID []byte, rt []byte, chainID byte, bondPrices map[string]uint64, saleID []byte, dcbAddress string) (*TxBuySellDCBResponse, error) {
	// Mint and send Constant
	pks := [][]byte{txRequest.PaymentAddress.Pk[:], txRequest.PaymentAddress.Pk[:]}
	tks := [][]byte{txRequest.PaymentAddress.Tk[:], txRequest.PaymentAddress.Tk[:]}

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
		SaleID:        saleID,
	}
	return txResponse, nil
}

func BuildResponseForBond(txRequest *TxBuySellRequest, bondID []byte, rt []byte, chainID byte, bondPrices map[string]uint64, unspentTxTokenOuts []TxTokenVout, saleID []byte, dcbAddress string) (*TxBuySellDCBResponse, []TxTokenVout, error) {
	accountDCB, _ := wallet.Base58CheckDeserialize(dcbAddress)
	// Get amount of Constant user sent
	value := uint64(0)
	userPk := txRequest.Tx.JSPubKey
	for _, desc := range txRequest.Tx.Descs {
		for _, note := range desc.Note {
			if bytes.Equal(note.Apk[:], accountDCB.KeySet.PaymentAddress.Pk) {
				value += note.Value
			}
		}
	}
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
		accountDCB, _ := wallet.Base58CheckDeserialize(dcbAddress)
		txTokenOuts = append(txTokenOuts, TxTokenVout{
			PaymentAddress: accountDCB.KeySet.PaymentAddress,
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
		SaleID:        saleID,
	}
	return txResponse, unspentTxTokenOuts[usedID:], nil
}

func (tx *TxBuySellDCBResponse) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()
	record += string(tx.RequestedTxID[:])
	record += string(tx.SaleID)

	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxBuySellDCBResponse) ValidateTransaction() bool {
	// validate for customtoken tx
	if !tx.TxCustomToken.ValidateTransaction() {
		return false
	}
	// TODO(@0xbunyip): check if there's a corresponding request in the same block
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
