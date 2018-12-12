package transaction

/*import (
	"github.com/ninjadotorg/constant/common"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

type TxBuyBackRequest struct {
	*BuyBackRequestInfo
	*TxNormal // fee
	// TODO: signature?
}

type BuyBackRequestInfo struct {
	BuyBackFromTxID *common.Hash
	VoutIndex       int
}

type FeeArgs struct {
	SenderKey     *privacy.SpendingKey
	PaymentInfo   []*privacy.PaymentInfo
	Rts           map[byte]*common.Hash
	UsableTx      map[byte][]*Tx
	Commitments   map[byte]([][]byte)
	Fee           uint64
	SenderChainID byte
}

// CreateTxBuyBackRequest
// senderKey and paymentInfo is for paying fee
func CreateTxBuyBackRequest(
	feeArgs FeeArgs,
	buyBackRequestInfo *BuyBackRequestInfo,
) (*TxBuyBackRequest, error) {
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

	txBuyBackRequest := &TxBuyBackRequest{
		BuyBackRequestInfo: buyBackRequestInfo,
		TxNormal:           tx,
	}
	txBuyBackRequest.Type = common.TxBuyBackRequest
	return txBuyBackRequest, nil
}

func (tx *TxBuyBackRequest) Hash() *common.Hash {
	// get hash of tx
	record := tx.TxNormal.Hash().String()
	record += tx.BuyBackFromTxID.String()
	record += string(tx.VoutIndex)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxBuyBackRequest) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.TxNormal.ValidateTransaction() {
		return false
	}
	return true
}

func (tx *TxBuyBackRequest) GetType() string {
	return tx.TxNormal.Type
}

func (tx *TxBuyBackRequest) GetTxVirtualSize() uint64 {
	// TODO: calculate
	return 0
}

func (tx *TxBuyBackRequest) GetSenderAddrLastByte() byte {
	return tx.TxNormal.AddressLastByte
}

func (tx *TxBuyBackRequest) GetTxFee() uint64 {
	return tx.TxNormal.Fee
}*/
