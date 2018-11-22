package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
)

type TxBuySellRequest struct {
	*RequestInfo
	*Tx // fee + amount to pay for buying bonds/govs
	// TODO: signature?
}

type RequestInfo struct {
	PaymentAddress privacy.PaymentAddress
	AssetType      common.Hash // token id
	Amount         uint64
	BuyPrice       uint64 // in Constant unit
}

// TxCreateBuySellRequest
// senderKey and paymentInfo is for paying fee
func TxCreateBuySellRequest(
	feeArgs FeeArgs,
	requestInfo *RequestInfo,
) (*TxBuySellRequest, error) {
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

	txbuySellRequest := &TxBuySellRequest{
		RequestInfo: requestInfo,
		Tx:          tx,
	}
	txbuySellRequest.Type = common.TxBuyFromGOVRequest
	return txbuySellRequest, nil
}

func (tx *TxBuySellRequest) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	record += tx.AssetType.String()
	record += string(tx.Amount)
	record += string(tx.BuyPrice)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxBuySellRequest) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}
	return true
}

func (tx *TxBuySellRequest) GetType() string {
	return tx.Tx.Type
}

func (tx *TxBuySellRequest) GetTxVirtualSize() uint64 {
	// TODO: calculate
	return 0
}

func (tx *TxBuySellRequest) GetSenderAddrLastByte() byte {
	return tx.Tx.AddressLastByte
}

func (tx *TxBuySellRequest) GetTxFee() uint64 {
	return tx.Tx.Fee
}
