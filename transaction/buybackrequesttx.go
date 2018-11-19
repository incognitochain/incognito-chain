package transaction

import (
	"github.com/ninjadotorg/constant/common"
)

type BuyBackRequestTx struct {
	*BuyBackRequestInfo
	*Tx // fee
	// TODO: signature?
}

type BuyBackRequestInfo struct {
	BuyBackFromTxID *common.Hash
	VoutIndex       int
}

// CreateBuyBackRequestTx
// senderKey and paymentInfo is for paying fee
func CreateBuyBackRequestTx(
	feeArgs FeeArgs,
	buyBackRequestInfo *BuyBackRequestInfo,
) (*BuyBackRequestTx, error) {
	// Create tx for fee &
	tx, err := CreateTx(
		feeArgs.SenderKey,
		feeArgs.PaymentInfo,
		feeArgs.Rts,
		feeArgs.UsableTx,
		feeArgs.Commitments,
		feeArgs.Fee,
		feeArgs.SenderChainID,
		false,
	)
	if err != nil {
		return nil, err
	}

	buyBackRequestTx := &BuyBackRequestTx{
		BuyBackRequestInfo: buyBackRequestInfo,
		Tx:                 tx,
	}
	buyBackRequestTx.Type = common.TxBuyBackRequest
	return buyBackRequestTx, nil
}

func (tx *BuyBackRequestTx) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()
	record += tx.BuyBackFromTxID.String()
	record += string(tx.VoutIndex)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *BuyBackRequestTx) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}
	return true
}

func (tx *BuyBackRequestTx) GetType() string {
	return tx.Tx.Type
}

func (tx *BuyBackRequestTx) GetTxVirtualSize() uint64 {
	// TODO: calculate
	return 0
}

func (tx *BuyBackRequestTx) GetSenderAddrLastByte() byte {
	return tx.Tx.AddressLastByte
}

func (tx *BuyBackRequestTx) GetTxFee() uint64 {
	return tx.Tx.Fee
}
