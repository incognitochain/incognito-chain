package transaction

import (
	"github.com/ninjadotorg/constant/common"
)

type TxBuySellDCBResponse struct {
	*TxCustomToken // fee + amount to pay for bonds/constant
	RequestedTxID  *common.Hash
	SaleID         []byte
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
