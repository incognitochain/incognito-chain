package transaction

import "github.com/ninjadotorg/constant/common"

type CrowSaleRequestTx struct {
	*RequestInfo
	*Tx
}

type RequestInfo struct {
	Amount   uint64
	BuyPrice uint64 // in Constant unit
}

// CreateTxLoanRequest
// senderKey and paymentInfo is for paying fee
func CreateCrowSaleRequestTx(
	feeArgs FeeArgs,
	requestInfo *RequestInfo,
) (*CrowSaleRequestTx, error) {
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

	CrowSaleRequestTx := &CrowSaleRequestTx{
		RequestInfo: requestInfo,
		Tx:          tx,
	}
	return CrowSaleRequestTx, nil
}

func (tx *CrowSaleRequestTx) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of collateral data
	record += string(tx.Amount)
	record += string(tx.BuyPrice)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *CrowSaleRequestTx) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.Tx.ValidateTransaction() {
		return false
	}
	return true
}

func (tx *CrowSaleRequestTx) GetType() string {
	return tx.Tx.Type
}

func (tx *CrowSaleRequestTx) GetTxVirtualSize() uint64 {
	// TODO: calculate
	return 0
}

func (tx *CrowSaleRequestTx) GetSenderAddrLastByte() byte {
	return tx.Tx.AddressLastByte
}

func (tx *CrowSaleRequestTx) GetTxFee() uint64 {
	return tx.Tx.Fee
}
