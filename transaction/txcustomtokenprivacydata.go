package transaction

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/ninjadotorg/constant/privacy-protocol"
)

type TxTokenPrivacyData struct {
	TxNormal       Tx          // used for privacy functionality
	PropertyID     common.Hash // = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string

	Type   int    // action type
	Amount uint64 // init amount
}

// Hash - return hash of token data, be used as Token ID
func (self TxTokenPrivacyData) Hash() (*common.Hash, error) {
	record := self.PropertyName + self.PropertySymbol + fmt.Sprintf("%d", self.Amount) + self.TxNormal.Hash().String()
	hash := common.DoubleHashH([]byte(record))
	return &hash, nil
}

// CustomTokenParamTx - use for rpc request json body
type CustomTokenPrivacyParamTx struct {
	PropertyID     string                 `json:"TokenID"`
	PropertyName   string                 `json:"TokenName"`
	PropertySymbol string                 `json:"TokenSymbol"`
	Amount         uint64                 `json:"TokenAmount"`
	TokenTxType    int                    `json:"TokenTxType"`
	Receiver       []*privacy.PaymentInfo `json:"TokenReceiver"`
	TokenInput     []*privacy.InputCoin   `json:"TokenInput"`
}

// CreateCustomTokenReceiverArray - parse data frm rpc request to create a list vout for preparing to create a custom token tx
// data interface is a map[paymentt-address]{transferring-amount}
func CreateCustomTokenPrivacyReceiverArray(data interface{}) []*privacy.PaymentInfo {
	result := []*privacy.PaymentInfo{}
	receivers := data.(map[string]interface{})
	for key, value := range receivers {
		key, _ := wallet.Base58CheckDeserialize(key)
		temp := &privacy.PaymentInfo{
			PaymentAddress: key.KeySet.PaymentAddress,
			Amount:         uint64(value.(float64)),
		}
		result = append(result, temp)
	}
	return result
}
