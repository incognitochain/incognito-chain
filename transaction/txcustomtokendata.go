package transaction

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
)

type BuyBackInfo struct {
	StartSellingAt uint32
	Maturity       uint32
	BuyBackPrice   uint64 // in Constant unit
}

type BuySellResponse struct {
	BuyBackInfo *BuyBackInfo
	AssetID     string // only bond for now - encoded string of compound values (Maturity + BuyBackPrice + StartSellingAt) from SellingBonds param
}

// TxTokenVin ...
type TxTokenVin struct {
	TxCustomTokenID common.Hash
	VoutIndex       int
	Signature       string
	PaymentAddress  privacy.PaymentAddress // use to verify signature of pre-utxo of token
}

func (self TxTokenVin) Hash() *common.Hash {
	record := common.EmptyString
	record += self.TxCustomTokenID.String()
	record += fmt.Sprintf("%d", self.VoutIndex)
	record += self.Signature
	record += base58.Base58Check{}.Encode(self.PaymentAddress.Pk[:], 0)
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// TxTokenVout ...
type TxTokenVout struct {
	Value          uint64
	PaymentAddress privacy.PaymentAddress // public key of receiver

	BondID          string // Temporary
	index           int
	txCustomTokenID common.Hash
	BuySellResponse *BuySellResponse
}

func (self TxTokenVout) Hash() *common.Hash {
	record := common.EmptyString
	record += fmt.Sprintf("%d", self.Value)
	record += base58.Base58Check{}.Encode(self.PaymentAddress.Pk[:], 0)
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (self *TxTokenVout) SetIndex(index int) {
	self.index = index
}

func (self TxTokenVout) GetIndex() int {
	return self.index
}

func (self *TxTokenVout) SetTxCustomTokenID(txCustomTokenID common.Hash) {
	self.txCustomTokenID = txCustomTokenID
}

func (self TxTokenVout) GetTxCustomTokenID() common.Hash {
	return self.txCustomTokenID
}

// TxTokenData ...
type TxTokenData struct {
	PropertyID     common.Hash // = hash of TxTokenData data
	PropertyName   string
	PropertySymbol string

	Type   int // action type
	Amount uint64
	Vins   []TxTokenVin
	Vouts  []TxTokenVout
}

func (self TxTokenData) Hash() (*common.Hash, error) {
	if self.Vouts == nil {
		return nil, errors.New("Vout is empty")
	}
	record := self.PropertyName + self.PropertySymbol + fmt.Sprintf("%d", self.Amount)
	for _, out := range self.Vouts {
		record += string(out.PaymentAddress.Pk[:])
	}
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash, nil
}

// CustomTokenParamTx - use for rpc request json body
type CustomTokenParamTx struct {
	PropertyID     string        `json:"TokenID"`
	PropertyName   string        `json:"TokenName"`
	PropertySymbol string        `json:"TokenSymbol"`
	Amount         uint64        `json:"TokenAmount"`
	TokenTxType    int           `json:"TokenTxType"`
	Receiver       []TxTokenVout `json:"TokenReceiver"`

	vins       []TxTokenVin
	vinsAmount uint64
}

func (self *CustomTokenParamTx) SetVins(vins []TxTokenVin) {
	self.vins = vins
}

func (self *CustomTokenParamTx) SetVinsAmount(vinsAmount uint64) {
	self.vinsAmount = vinsAmount
}

// CreateCustomTokenReceiverArray ...
func CreateCustomTokenReceiverArray(data interface{}) []TxTokenVout {
	result := []TxTokenVout{}
	receivers := data.(map[string]interface{})
	for key, value := range receivers {
		key, _ := wallet.Base58CheckDeserialize(key)
		temp := TxTokenVout{
			PaymentAddress: key.KeySet.PaymentAddress,
			Value:          uint64(value.(float64)),
		}
		result = append(result, temp)
	}
	return result
}
