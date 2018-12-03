package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"errors"
	"fmt"
	"github.com/ninjadotorg/constant/wallet"
)

type TxTokenPrivacyData struct {
	PropertyID     common.Hash // = hash of TxTokenData data
	PropertyName   string
	PropertySymbol string

	Type   int    // action type
	Amount uint64 // init amount
	Descs  []JoinSplitDesc `json:"Descs"`
}

// Hash - return hash of token data, be used as Token ID
func (self TxTokenPrivacyData) Hash() (*common.Hash, error) {
	if self.Descs == nil {
		return nil, errors.New("Privacy data is empty")
	}
	record := self.PropertyName + self.PropertySymbol + fmt.Sprintf("%d", self.Amount)
	for _, out := range self.Descs {
		record += out.toString()
	}
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash, nil
}

// CustomTokenParamTx - use for rpc request json body
type CustomTokenPrivacyParamTx struct {
	PropertyID     string        `json:"TokenID"`
	PropertyName   string        `json:"TokenName"`
	PropertySymbol string        `json:"TokenSymbol"`
	Amount         uint64        `json:"TokenAmount"`
	TokenTxType    int           `json:"TokenTxType"`
	Receiver       []TxTokenVout `json:"TokenReceiver"`
}

// CreateCustomTokenReceiverArray - parse data frm rpc request to create a list vout for preparing to create a custom token tx
// data interface is a map[paymentt-address]{transferring-amount}
func CreateCustomTokenPrivacyReceiverArray(data interface{}) []TxTokenVout {
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
