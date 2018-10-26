package transaction

import (
	"encoding/base64"
	"fmt"
	"hash"
	"strconv"

	"github.com/ninjadotorg/cash/common"
	"github.com/ninjadotorg/cash/privacy/client"
)

// TxTokenVin ...
type TxTokenVin struct {
	Vout      int
	Signature string
	PubKey    string
}

// TxTokenVout ...
type TxTokenVout struct {
	Value       uint64
	ScripPubKey string
}

// TxToken ...
type TxToken struct {
	Version         byte
	Type            byte
	PropertyName    string
	PropertySymbol  string
	Vin             []TxTokenVin
	Vout            []TxTokenVout
	TxCustomTokenID hash.Hash
	Amount          float64
}

// TxCustomToken ...
type TxCustomToken struct {
	Tx

	TxToken TxToken
}

// CustomTokenParamTx ...
type CustomTokenParamTx struct {
	PropertyName    string  `json:"TokenName"`
	PropertySymbol  string  `json:"TokenSymbol"`
	TxCustomTokenID string  `json:"TokenHash"`
	Amount          float64 `json:"TokenAmount"`
	TokenTxType     float64 `json:"TokenTxType"`
}

// CreateEmptyCustomTokenTx - return an init custom token transaction
func CreateEmptyCustomTokenTx() (*TxCustomToken, error) {
	emptyTx, err := CreateEmptyTx(common.TxCustomTokenType)

	if err != nil {
		return nil, err
	}

	txToken := TxToken{}

	txCustomToken := &TxCustomToken{
		Tx:      *emptyTx,
		TxToken: txToken,
	}
	return txCustomToken, nil
}

// Hash returns the hash of all fields of the transaction
func (tx TxCustomToken) Hash() *common.Hash {
	record := strconv.Itoa(int(tx.Tx.Version))
	record += tx.Tx.Type
	record += strconv.FormatInt(tx.Tx.LockTime, 10)
	record += strconv.FormatUint(tx.Tx.Fee, 10)
	record += strconv.Itoa(len(tx.Tx.Descs))
	for _, desc := range tx.Tx.Descs {
		record += desc.toString()
	}
	record += string(tx.Tx.JSPubKey)
	// record += string(tx.JSSig)
	record += string(tx.Tx.AddressLastByte)
	record += tx.TxToken.PropertyName
	record += tx.TxToken.PropertySymbol
	record += fmt.Sprintf("%f", tx.TxToken.Amount)
	record += base64.URLEncoding.EncodeToString(tx.TxToken.TxCustomTokenID.Sum(nil))
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// ValidateTransaction ...
func (tx *TxCustomToken) ValidateTransaction() bool {
	if tx.Tx.ValidateTransaction() {
		return true
	}
	return false
}

// CreateTxCustomToken ...
func CreateTxCustomToken(senderKey *client.SpendingKey,
	paymentInfo []*client.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*Tx,
	nullifiers map[byte]([][]byte),
	commitments map[byte]([][]byte),
	fee uint64,
	assetType string,
	senderChainID byte,
	tokenParams *CustomTokenParamTx) (*TxCustomToken, error) {
	// fee = 0 // TODO remove this line

	tx, err := CreateEmptyCustomTokenTx()
	if err != nil {
		return nil, err
	}

	// TO-DO

	return tx, nil
}
