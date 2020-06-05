package transaction

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/operation"

	"github.com/incognitochain/incognito-chain/privacy/coin"

	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

type TxPrivacyTokenData struct {
	TxNormal       TxBase      // used for privacy functionality
	PropertyID     common.Hash // = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string

	Type     int    // action type
	Mintable bool   // default false
	Amount   uint64 // init amount
}

func (txTokenPrivacyData TxPrivacyTokenData) String() string {
	record := txTokenPrivacyData.PropertyName
	record += txTokenPrivacyData.PropertySymbol
	record += fmt.Sprintf("%d", txTokenPrivacyData.Amount)
	if txTokenPrivacyData.TxNormal.Proof != nil {
		inputCoins := txTokenPrivacyData.TxNormal.GetProof().GetInputCoins()
		outputCoins := txTokenPrivacyData.TxNormal.GetProof().GetOutputCoins()
		for _, out := range outputCoins {
			publicKeyBytes := []byte{}
			if out.GetPublicKey() != nil {
				publicKeyBytes = out.GetPublicKey().ToBytesS()
			}
			record += string(publicKeyBytes)
			record += strconv.FormatUint(out.GetValue(), 10)
		}
		for _, in := range inputCoins {
			publicKeyBytes := []byte{}
			if in.GetPublicKey() != nil {
				publicKeyBytes = in.GetPublicKey().ToBytesS()
			}
			record += string(publicKeyBytes)
			record += strconv.FormatUint(in.GetValue(), 10)
		}
	}
	return record
}

func (txTokenPrivacyData TxPrivacyTokenData) JSONString() string {
	data, err := json.MarshalIndent(txTokenPrivacyData, "", "\t")
	if err != nil {
		Logger.Log.Error(err)
		return ""
	}
	return string(data)
}

// Hash - return hash of custom token data, be used as Token ID
func (txTokenPrivacyData TxPrivacyTokenData) Hash() (*common.Hash, error) {
	point := operation.HashToPoint([]byte(txTokenPrivacyData.String()))
	hash := new(common.Hash)
	err := hash.SetBytes(point.ToBytesS())
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// CustomTokenParamTx - use for rpc request json body
type CustomTokenPrivacyParamTx struct {
	PropertyID     string                 `json:"TokenID"`
	PropertyName   string                 `json:"TokenName"`
	PropertySymbol string                 `json:"TokenSymbol"`
	Amount         uint64                 `json:"TokenAmount"`
	TokenTxType    int                    `json:"TokenTxType"`
	Receiver       []*privacy.PaymentInfo `json:"TokenReceiver"`
	TokenInput     []coin.PlainCoin    `json:"TokenInput"`
	Mintable       bool                   `json:"TokenMintable"`
	Fee            uint64                 `json:"TokenFee"`
}

// CreateCustomTokenReceiverArray - parse data frm rpc request to create a list vout for preparing to create a custom token tx
// data interface is a map[paymentt-address]{transferring-amount}
func CreateCustomTokenPrivacyReceiverArray(dataReceiver interface{}) ([]*privacy.PaymentInfo, int64, error) {
	if dataReceiver == nil {
		return nil, 0, fmt.Errorf("data receiver is in valid")
	}
	result := []*privacy.PaymentInfo{}
	voutsAmount := int64(0)
	receivers, ok := dataReceiver.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("data receiver is in valid")
	}
	for key, value := range receivers {
		keyWallet, err := wallet.Base58CheckDeserialize(key)
		if err != nil {
			Logger.Log.Errorf("Invalid key in CreateCustomTokenPrivacyReceiverArray %+v", key)
			return nil, 0, err
		}
		keySet := keyWallet.KeySet
		temp := &privacy.PaymentInfo{
			PaymentAddress: keySet.PaymentAddress,
			Amount:         uint64(value.(float64)),
		}
		result = append(result, temp)
		voutsAmount += int64(temp.Amount)
	}
	return result, voutsAmount, nil
}

// CreateCustomTokenReceiverArray - parse data frm rpc request to create a list vout for preparing to create a custom token tx
// data interface is a map[paymentt-address]{transferring-amount}
func CreateCustomTokenPrivacyReceiverArrayV2(dataReceiver interface{}) ([]*privacy.PaymentInfo, int64, error) {
	if dataReceiver == nil {
		return nil, 0, fmt.Errorf("data receiver is in valid")
	}
	result := []*privacy.PaymentInfo{}
	voutsAmount := int64(0)
	receivers, ok := dataReceiver.(map[string]interface{})
	if !ok {
		return nil, 0, fmt.Errorf("data receiver is in valid")
	}
	for key, value := range receivers {
		keyWallet, err := wallet.Base58CheckDeserialize(key)
		if err != nil {
			Logger.Log.Errorf("Invalid key in CreateCustomTokenPrivacyReceiverArray %+v", key)
			return nil, 0, err
		}
		keySet := keyWallet.KeySet

		amount, err := common.AssertAndConvertStrToNumber(value)
		if err != nil {
			return nil, 0, err
		}

		temp := &privacy.PaymentInfo{
			PaymentAddress: keySet.PaymentAddress,
			Amount:         amount,
		}
		result = append(result, temp)
		voutsAmount += int64(temp.Amount)
	}
	return result, voutsAmount, nil
}
