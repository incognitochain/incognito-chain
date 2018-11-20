package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
)

type BuySellRequestTx struct {
	*RequestInfo
	*TxCustomToken // fee + amount to pay for buying bonds/govs
	// TODO: signature?
}

type RequestInfo struct {
	PaymentAddress privacy.PaymentAddress
	AssetType      string
	Amount         uint64
	BuyPrice       uint64 // in Constant unit

	SaleID []byte // only when requesting to DCB
}

// CreateBuySellRequestTx
// senderKey and paymentInfo is for paying fee
func CreateBuySellRequestTx(
	senderKey *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*Tx,
	commitments map[byte]([][]byte),
	fee uint64,
	senderChainID byte,
	tokenParams *CustomTokenParamTx,
	listCustomTokens map[common.Hash]TxCustomToken,
	requestInfo *RequestInfo,
) (*BuySellRequestTx, error) {
	// Create tx for fee &
	tx, err := CreateTxCustomToken(
		senderKey,
		paymentInfo,
		rts,
		usableTx,
		commitments,
		fee,
		senderChainID,
		tokenParams,
		listCustomTokens,
	)
	if err != nil {
		return nil, err
	}

	buySellRequestTx := &BuySellRequestTx{
		RequestInfo:   requestInfo,
		TxCustomToken: tx,
	}
	buySellRequestTx.Type = common.TxBuyFromGOVRequest
	return buySellRequestTx, nil
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
