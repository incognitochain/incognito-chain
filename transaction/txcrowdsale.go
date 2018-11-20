package transaction

import (
	"github.com/ninjadotorg/constant/common"
	privacy "github.com/ninjadotorg/constant/privacy-protocol"
)

type SaleData struct {
	SaleID []byte // Unique id of the crowdsale to store in db
	BondID []byte // in case either base or quote asset is bond

	BuyingAsset  string
	SellingAsset string
	Price        uint64
}

type TxBuySellDCBResponse struct {
	*TxCustomToken // fee + amount to pay for bonds/constant
	RequestedTxID  *common.Hash
}

// CreateTxBuySellDCBResponse
func CreateTxBuySellDCBResponse(
	senderKey *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*Tx,
	commitments map[byte]([][]byte),
	fee uint64,
	senderChainID byte,
	tokenParams *CustomTokenParamTx,
	listCustomTokens map[common.Hash]TxCustomToken,
	requestedTxID *common.Hash,
) (*TxBuySellDCBResponse, error) {
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

	buySellResponseTx := &TxBuySellDCBResponse{
		TxCustomToken: tx,
	}
	buySellResponseTx.Type = common.TxBuySellDCBResponse
	return buySellResponseTx, nil
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
