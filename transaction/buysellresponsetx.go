package transaction

import "github.com/ninjadotorg/constant/common"

type BuySellResponseTx struct {
	txId      *common.Hash
	ID        string
	Amount    uint64
	AssetType string
	Type      string // tx type

	BuyBackInfo *BuyBackInfo // bond only for now
}

type BuyBackInfo struct {
	Maturity     uint32
	BuyBackPrice uint64 // in Constant unit
}

// CreateBuySellResponseTx
func CreateBuySellResponseTx(
	id string,
	amount uint64,
	assetType string,
	txType string,
	buyBackInfo *BuyBackInfo,
) *BuySellResponseTx {

	BuySellResponseTx := &BuySellResponseTx{
		ID:          id,
		Amount:      amount,
		AssetType:   assetType,
		Type:        txType,
		BuyBackInfo: buyBackInfo,
	}
	return BuySellResponseTx
}

func (tx *BuySellResponseTx) SetTxID(txId *common.Hash) {
	tx.txId = txId
}

func (tx *BuySellResponseTx) GetTxID() *common.Hash {
	return tx.txId
}

func (tx *BuySellResponseTx) Hash() *common.Hash {
	record := tx.ID
	record += tx.AssetType
	record += tx.Type
	record += string(tx.Amount)
	if tx.BuyBackInfo != nil {
		record += string(tx.BuyBackInfo.BuyBackPrice)
		record += string(tx.BuyBackInfo.Maturity)
	}

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *BuySellResponseTx) ValidateTransaction() bool {
	// will validate against buy/sell request tx as consluding block
	return true
}

func (tx *BuySellResponseTx) GetType() string {
	return tx.Type
}

func (tx *BuySellResponseTx) GetTxVirtualSize() uint64 {
	// TODO: calculate
	return 0
}

func (tx *BuySellResponseTx) GetSenderAddrLastByte() byte {
	return 0
}

func (tx *BuySellResponseTx) GetTxFee() uint64 {
	return 0
}
