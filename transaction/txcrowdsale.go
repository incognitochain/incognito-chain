package transaction

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/voting"
)

type TxBuySellDCBResponse struct {
	*TxCustomToken // fee + amount to pay for bonds/constant
	RequestedTxID  *common.Hash
}

func BuildTxBuySellDCBResponse(txRequest *TxBuySellRequest, saleData *voting.SaleData) (*TxBuySellDCBResponse, error) {
	if saleData.SellingAsset == common.AssetTypeCoin {
		// Mint and send Constant

	} else if saleData.SellingAsset == common.AssetTypeBond {
		// Send bond from DCB's account
	} else {
		return nil, fmt.Errorf("Selling asset of crowdsale is invalid: %s", saleData.SellingAsset)
	}
	return nil, nil
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
