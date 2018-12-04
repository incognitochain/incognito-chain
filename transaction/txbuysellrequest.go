package transaction

/*import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
)

type TxBuySellRequest struct {
	*RequestInfo
	*TxCustomToken // fee + amount to pay for buying bonds/govs
	// TODO: signature?
}

type RequestInfo struct {
	PaymentAddress privacy.PaymentAddress
	AssetType      common.Hash // token id (note: for bond, this one is just bond token id prefix)
	Amount         uint64
	BuyPrice       uint64 // in Constant unit

	SaleID []byte // only when requesting to DCB
}

// TxCreateBuySellRequest
// senderKey and paymentInfo is for paying fee
func CreateBuySellRequestTx(
	senderKey *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*TxNormal,
	commitments map[byte]([][]byte),
	fee uint64,
	senderChainID byte,
	tokenParams *CustomTokenParamTx,
	listCustomTokens map[common.Hash]TxCustomToken,
	requestInfo *RequestInfo,
) (*TxBuySellRequest, error) {
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

	txbuySellRequest := &TxBuySellRequest{
		RequestInfo:   requestInfo,
		TxCustomToken: tx,
	}
	txbuySellRequest.Type = common.TxBuyFromGOVRequest
	return txbuySellRequest, nil
}

func (tx *TxBuySellRequest) Hash() *common.Hash {
	// get hash of tx
	record := tx.TxNormal.Hash().String()

	record += tx.AssetType.String()
	record += string(tx.Amount)
	record += string(tx.BuyPrice)

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxBuySellRequest) ValidateTransaction() bool {
	// validate for normal tx
	if !tx.TxCustomToken.ValidateTransaction() {
		return false
	}
	return true
}

func (tx *TxBuySellRequest) GetType() string {
	return tx.TxNormal.Type
}

func (tx *TxBuySellRequest) GetTxVirtualSize() uint64 {
	// TODO: calculate
	return 0
}

func (tx *TxBuySellRequest) GetSenderAddrLastByte() byte {
	return tx.TxNormal.AddressLastByte
}

func (tx *TxBuySellRequest) GetTxFee() uint64 {
	return tx.TxNormal.Fee
}*/
