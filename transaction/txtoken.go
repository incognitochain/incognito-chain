package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy/client"
	"github.com/ninjadotorg/constant/wallet"
	"fmt"
	"github.com/pkg/errors"
)

// TxTokenVin ...
type TxTokenVin struct {
	TxCustomTokenID common.Hash
	VoutIndex       int
	Signature       string
	PaymentAddress  client.PaymentAddress
}

func (self TxTokenVin) Hash() *common.Hash {
	record := common.EmptyString
	record += self.TxCustomTokenID.String()
	record += fmt.Sprintf("%d", self.VoutIndex)
	record += self.Signature
	record += base58.Base58Check{}.Encode(self.PaymentAddress.Apk[:], 0)
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// TxTokenVout ...
type TxTokenVout struct {
	Value          uint64
	PaymentAddress client.PaymentAddress

	index           int
	txCustomTokenID common.Hash
}

func (self TxTokenVout) Hash() *common.Hash {
	record := common.EmptyString
	record += fmt.Sprintf("%d", self.Value)
	record += base58.Base58Check{}.Encode(self.PaymentAddress.Apk[:], 0)
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

// TxToken ...
type TxToken struct {
	PropertyID     common.Hash // = hash of TxToken data
	PropertyName   string
	PropertySymbol string
	Type           int
	Amount         uint64
	Vins           []TxTokenVin
	Vouts          []TxTokenVout
}

func (self TxToken) Hash() (*common.Hash, error) {
	if (self.Vouts == nil) {
		return nil, errors.New("Vout is empty")
	}
	record := self.PropertyName + self.PropertyName + fmt.Sprintf("%d", self.Amount)
	for _, out := range self.Vouts {
		record += string(out.PaymentAddress.Apk[:])
	}
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash, nil
}

// CustomTokenParamTx - use for rpc request json body
type CustomTokenParamTx struct {
	PropertyName   string      `json:"TokenName"`
	PropertySymbol string      `json:"TokenSymbol"`
	Amount         uint64      `json:"TokenAmount"`
	TokenTxType    int         `json:"TokenTxType"`
	Receiver       TxTokenVout `json:"TokenReceiver"`

	vins []TxTokenVin
}

func (self *CustomTokenParamTx) SetVins(vins []TxTokenVin) {
	self.vins = vins
}

// CreateCustomTokenReceiverArray ...
func CreateCustomTokenReceiverArray(data interface{}) TxTokenVout {
	temp := data.(map[string]interface{})
	paymentAddress := temp["PaymentAddress"].(string)
	key, _ := wallet.Base58CheckDeserialize(paymentAddress)
	resultItem := TxTokenVout{
		PaymentAddress: key.KeySet.PaymentAddress,
		Value:          uint64(temp["Value"].(float64)),
	}
	return resultItem
}
