package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/privacy"
	"github.com/big0t/constant-chain/wallet"
	"strconv"
)

type TxTokenPrivacyData struct {
	TxNormal       Tx          // used for privacy functionality
	PropertyID     common.Hash // = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string

	Type     int    // action type
	Mintable bool   // default false
	Amount   uint64 // init amount
}

func (txTokenPrivacyData TxTokenPrivacyData) String() string {
	record := txTokenPrivacyData.PropertyName
	record += txTokenPrivacyData.PropertySymbol
	record += fmt.Sprintf("%d", txTokenPrivacyData.Amount)
	if txTokenPrivacyData.TxNormal.Proof != nil {
		for _, out := range txTokenPrivacyData.TxNormal.Proof.OutputCoins {
			record += string(out.CoinDetails.PublicKey.Compress())
			record += strconv.FormatUint(out.CoinDetails.Value, 10)
		}
		for _, in := range txTokenPrivacyData.TxNormal.Proof.InputCoins {
			if in.CoinDetails.PublicKey != nil {
				record += string(in.CoinDetails.PublicKey.Compress())
			}
			if in.CoinDetails.Value > 0 {
				record += strconv.FormatUint(in.CoinDetails.Value, 10)
			}
		}
	}
	return record
}

func (txTokenPrivacyData TxTokenPrivacyData) JSONString() string {
	data, err := json.MarshalIndent(txTokenPrivacyData, "", "\t")
	if err != nil {
		Logger.log.Error(err)
		return ""
	}
	return string(data)
}

// Hash - return hash of custom token data, be used as Token ID
func (txTokenPrivacyData TxTokenPrivacyData) Hash() (*common.Hash, error) {
	hash := common.DoubleHashH([]byte(txTokenPrivacyData.String()))
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
func CreateCustomTokenPrivacyReceiverArray(data interface{}) ([]*privacy.PaymentInfo, int64) {
	result := []*privacy.PaymentInfo{}
	voutsAmount := int64(0)
	receivers := data.(map[string]interface{})
	for key, value := range receivers {
		keyWallet, _ := wallet.Base58CheckDeserialize(key)
		keySet := keyWallet.KeySet
		temp := &privacy.PaymentInfo{
			PaymentAddress: keySet.PaymentAddress,
			Amount:         uint64(value.(float64)),
		}
		result = append(result, temp)
		voutsAmount += int64(temp.Amount)
	}
	return result, voutsAmount
}
