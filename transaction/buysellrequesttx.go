package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
)

type BuySellRequestTx struct {
	*RequestInfo
	*Tx
}

type RequestInfo struct {
	PaymentAddress privacy.PaymentAddress
	AssetType      string
	Amount         uint64
	BuyPrice       uint64 // in Constant unit
}

// CreateBuySellRequestTx
// senderKey and paymentInfo is for paying fee
func CreateBuySellRequestTx(
	feeArgs FeeArgs,
	requestInfo *RequestInfo,
) (*BuySellRequestTx, error) {
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

	BuySellRequestTx := &BuySellRequestTx{
		RequestInfo: requestInfo,
		Tx:          tx,
	}
	return BuySellRequestTx, nil
}

func (tx *BuySellRequestTx) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	record += tx.AssetType
	record += string(tx.Amount)
	record += string(tx.BuyPrice)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *BuySellRequestTx) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}
	return true
}

func (tx *BuySellRequestTx) GetType() string {
	return tx.Tx.Type
}

func (tx *BuySellRequestTx) GetTxVirtualSize() uint64 {
	// TODO: calculate
	return 0
}

func (tx *BuySellRequestTx) GetSenderAddrLastByte() byte {
	return tx.Tx.AddressLastByte
}

func (tx *BuySellRequestTx) GetTxFee() uint64 {
	return tx.Tx.Fee
}
