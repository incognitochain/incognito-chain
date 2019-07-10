package transaction

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

// TxTokenVin - vin format for custom token data
// It look like vin format of bitcoin
type TxTokenVin struct {
	TxCustomTokenID common.Hash            // TxNormal-id(or hash) of before tx, which is used as a input for current tx as a pre-utxo
	VoutIndex       int                    // index in vouts array of before TxNormal-id
	Signature       string                 // Signature to verify owning before tx(pre-utxo)
	PaymentAddress  privacy.PaymentAddress // use to verify signature of pre-utxo of token
}

func (txObj TxTokenVin) String() string {
	record := ""
	record += txObj.TxCustomTokenID.String()
	record += fmt.Sprintf("%d", txObj.VoutIndex)
	record += txObj.Signature
	record += base58.Base58Check{}.Encode(txObj.PaymentAddress.Pk[:], 0)
	return record
}

func (txObj TxTokenVin) JSONString() string {
	data, err := json.MarshalIndent(txObj, "", "\t")
	if err != nil {
		Logger.log.Error(err)
		return ""
	}
	return string(data)
}

// Hash - return hash data of TxTokenVin
func (txObj TxTokenVin) Hash() *common.Hash {
	// final hash
	hash := common.HashH([]byte(txObj.String()))
	return &hash
}

// TxTokenVout - vout format for custom token data
// It look like vout format of bitcoin
type TxTokenVout struct {
	Value          uint64                 // Amount to transfer
	PaymentAddress privacy.PaymentAddress // payment address of receiver

	// temp variable to determine position of itself in vouts arrays of tx which contain ittxObj
	index int
	// temp variable to know what is id of tx which contain itself
	txCustomTokenID common.Hash
}

func (txObj TxTokenVout) String() string {
	record := ""
	record += fmt.Sprintf("%d", txObj.Value)
	record += base58.Base58Check{}.Encode(txObj.PaymentAddress.Pk[:], 0)
	return record
}

func (txObj TxTokenVout) JSONString() string {
	data, err := json.MarshalIndent(txObj, "", "\t")
	if err != nil {
		Logger.log.Error(err)
		return ""
	}
	return string(data)
}

// Hash - return hash data of TxTokenVout
func (txObj TxTokenVout) Hash() *common.Hash {
	// final hash
	hash := common.HashH([]byte(txObj.String()))
	return &hash
}

// Set index temp variable
func (txObj *TxTokenVout) SetIndex(index int) {
	txObj.index = index
}

// Get index temp variable
func (txObj TxTokenVout) GetIndex() int {
	return txObj.index
}

// Set tx id temp variable
func (txObj *TxTokenVout) SetTxCustomTokenID(txCustomTokenID common.Hash) {
	txObj.txCustomTokenID = txCustomTokenID
}

// Get tx id temp variable
func (txObj TxTokenVout) GetTxCustomTokenID() common.Hash {
	return txObj.txCustomTokenID
}

// TxTokenData - main struct which contain vin and vout array for transferring or issuing custom token
// of course, it also contain token metadata: name, symbol, id(hash of token data)
type TxTokenData struct {
	PropertyID     common.Hash // = hash of TxTokenData data
	PropertyName   string
	PropertySymbol string
	Type           int    // action type [init, transfer, crossShard (used only for crossShard msg)]
	Mintable       bool   // can mine, default false
	Amount         uint64 // init amount
	Vins           []TxTokenVin
	Vouts          []TxTokenVout
}

func (txObj TxTokenData) String() string {
	record := txObj.PropertyName
	record += txObj.PropertySymbol
	record += fmt.Sprintf("%d", txObj.Amount)
	if len(txObj.Vins) > 0 {
		for _, in := range txObj.Vins {
			record += in.String()
		}
	}
	for _, out := range txObj.Vouts {
		record += out.String()
	}
	return record
}

func (txObj TxTokenData) JSONString() string {
	data, err := json.MarshalIndent(txObj, "", "\t")
	if err != nil {
		Logger.log.Error(err)
		return ""
	}
	return string(data)
}

// Hash - return hash of token data, be used as Token ID
func (txObj TxTokenData) Hash() (*common.Hash, error) {
	if txObj.Vouts == nil {
		return nil, errors.New("Vout is empty")
	}
	// final hash
	hash := common.HashH([]byte(txObj.String()))
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
	Mintable       bool          `json:"TokenMintable"`
	// temp variable to process coding
	vins       []TxTokenVin
	vinsAmount uint64
}

func (txObj *CustomTokenParamTx) SetVins(vins []TxTokenVin) {
	txObj.vins = vins
}

func (txObj *CustomTokenParamTx) SetVinsAmount(vinsAmount uint64) {
	txObj.vinsAmount = vinsAmount
}

// CreateCustomTokenReceiverArray - parse data frm rpc request to create a list vout for preparing to create a custom token tx
// data interface is a map[paymentt-address]{transferring-amount}
func CreateCustomTokenReceiverArray(data interface{}) ([]TxTokenVout, int64, error) {
	result := []TxTokenVout{}
	voutsAmount := int64(0)
	receivers := data.(map[string]interface{})
	for key, value := range receivers {
		keyWallet, err := wallet.Base58CheckDeserialize(key)
		if err != nil {
			return nil, 0, err
		}
		keySet := keyWallet.KeySet
		temp := TxTokenVout{
			PaymentAddress: keySet.PaymentAddress,
			Value:          uint64(value.(float64)),
		}
		result = append(result, temp)
		voutsAmount += int64(temp.Value)
	}
	return result, voutsAmount, nil
}
