package transaction

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/wallet"
	"github.com/pkg/errors"
	"strconv"
)

// TxTokenVin - vin format for custom token data
// It look like vin format of bitcoin
type TxTokenVin struct {
	TxCustomTokenID common.Hash            // TxNormal-id(or hash) of before tx, which is used as a input for current tx as a pre-utxo
	VoutIndex       int                    // index in vouts array of before TxNormal-id
	Signature       string                 // Signature to verify owning before tx(pre-utxo)
	PaymentAddress  privacy.PaymentAddress // use to verify signature of pre-utxo of token
}

// Hash - return hash data of TxTokenVin
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

// TxTokenVout - vout format for custom token data
// It look like vout format of bitcoin
type TxTokenVout struct {
	Value          uint64                 // Amount to transfer
	PaymentAddress privacy.PaymentAddress // public key of receiver

	// temp variable to determine position of itself in vouts arrays of tx which contain itself
	index int
	// temp variable to know what is id of tx which contain itself
	txCustomTokenID common.Hash
	// BuySellResponse *BuySellResponse
}

// Hash - return hash data of TxTokenVout
func (self TxTokenVout) Hash() *common.Hash {
	record := common.EmptyString
	record += fmt.Sprintf("%d", self.Value)
	record += base58.Base58Check{}.Encode(self.PaymentAddress.Pk[:], 0)
	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// Set index temp variable
func (self *TxTokenVout) SetIndex(index int) {
	self.index = index
}

// Get index temp variable
func (self TxTokenVout) GetIndex() int {
	return self.index
}

// Set tx id temp variable
func (self *TxTokenVout) SetTxCustomTokenID(txCustomTokenID common.Hash) {
	self.txCustomTokenID = txCustomTokenID
}

// Get tx id temp variable
func (self TxTokenVout) GetTxCustomTokenID() common.Hash {
	return self.txCustomTokenID
}

// TxTokenData - main struct which contain vin and vout array for transferring or issuing custom token
// of course, it also contain token metadata: name, symbol, id(hash of token data)
type TxTokenData struct {
	PropertyID     common.Hash // = hash of TxTokenData data
	PropertyName   string
	PropertySymbol string

	Type   int    // action type [init, transfer]
	Amount uint64 // init amount
	Vins   []TxTokenVin
	Vouts  []TxTokenVout
}

// Hash - return hash of token data, be used as Token ID
func (self TxTokenData) Hash() (*common.Hash, error) {
	if self.Vouts == nil {
		return nil, errors.New("Vout is empty")
	}
	record := self.PropertyName + self.PropertySymbol + fmt.Sprintf("%d", self.Amount)
	if len(self.Vins) > 0 {
		for _, in := range self.Vins {
			record += in.TxCustomTokenID.String()
			record += strconv.Itoa(in.VoutIndex)
			record += base58.Base58Check{}.Encode(in.PaymentAddress.Pk, 0x00)
			record += in.Signature
		}
	}
	for _, out := range self.Vouts {
		record += string(out.PaymentAddress.Pk[:])
		record += strconv.FormatUint(out.Value, 10)
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

	// temp variable to process coding
	vins       []TxTokenVin
	vinsAmount uint64
}

func (self *CustomTokenParamTx) SetVins(vins []TxTokenVin) {
	self.vins = vins
}

func (self *CustomTokenParamTx) SetVinsAmount(vinsAmount uint64) {
	self.vinsAmount = vinsAmount
}

// CreateCustomTokenReceiverArray - parse data frm rpc request to create a list vout for preparing to create a custom token tx
// data interface is a map[paymentt-address]{transferring-amount}
func CreateCustomTokenReceiverArray(data interface{}) ([]TxTokenVout, int64) {
	result := []TxTokenVout{}
	voutsAmount := int64(0)
	receivers := data.(map[string]interface{})
	for key, value := range receivers {
		key, _ := wallet.Base58CheckDeserialize(key)
		temp := TxTokenVout{
			PaymentAddress: key.KeySet.PaymentAddress,
			Value:          uint64(value.(float64)),
		}
		result = append(result, temp)
		voutsAmount += int64(temp.Value)
	}
	return result, voutsAmount
}
