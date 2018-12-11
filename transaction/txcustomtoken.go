package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/database"
)

// TxCustomToken is class tx which is inherited from constant tx(supporting privacy) for fee
// and contain data(vin, vout) to support issuing and transfer a custom token(token from end-user, look like erc-20)
// Dev or end-user can use this class tx to create an token type which use personal purpose
// In particular of constant network, some special token (DCB token, GOV token, BOND token, ....) used this class tx to implement something
type TxCustomToken struct {
	Tx                      // inherit from normal tx of constant(supporting privacy)
	TxTokenData TxTokenData // vin - vout format
	BoardType   uint8       // 1: DCB, 2: GOV
	BoardSigns  map[string][]byte

	// Template data variable to process logic
	listUtxo map[common.Hash]TxCustomToken
}

// Set listUtxo, which is used to contain a list old TxCustomToken relate to itself
func (tx *TxCustomToken) SetListUtxo(data map[common.Hash]TxCustomToken) {
	tx.listUtxo = data
}

// Hash returns the hash of all fields of the transaction
func (tx TxCustomToken) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of txtokendata
	txTokenDataHash, _ := tx.TxTokenData.Hash()
	record += txTokenDataHash.String()

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// ValidateTransaction - validate inheritance data from normal tx to check privacy and double spend for fee and transfer by constant
// if pass normal tx validation, it continue check signature on (vin-vout) custom token data
func (tx *TxCustomToken) ValidateTransaction(hasPrivacy bool, db database.DatabaseInterface) bool {
	// validate for normal tx
	if tx.Tx.ValidateTransaction(hasPrivacy, db) {
		if len(tx.listUtxo) == 0 {
			return false
		}
		for _, vin := range tx.TxTokenData.Vins {
			keySet := cashec.KeySet{}
			keySet.PaymentAddress = vin.PaymentAddress

			// get data from utxo
			utxo := tx.listUtxo[vin.TxCustomTokenID]
			vout := utxo.TxTokenData.Vouts[vin.VoutIndex]
			data := vout.Hash() // hash of vout in utxo

			ok, err := keySet.Verify(data[:], []byte(vin.Signature))
			if err != nil {
				return false
			}
			return ok
		}
		return true
	}
	return false
}

// GetTxVirtualSize computes the virtual size of a given transaction
// size of this tx = (normal TxNormal size) + (custom token data size)
func (tx *TxCustomToken) GetTxVirtualSize() uint64 {
	normalTxSize := tx.Tx.GetTxVirtualSize()

	tokenDataSize := uint64(0)

	tokenDataSize += uint64(len(tx.TxTokenData.PropertyName))
	tokenDataSize += uint64(len(tx.TxTokenData.PropertyName))
	tokenDataSize += uint64(len(tx.TxTokenData.PropertyID))
	tokenDataSize += 4 // for TxTokenData.Type

	for _, vin := range tx.TxTokenData.Vins {
		tokenDataSize += uint64(len(vin.Signature))
		tokenDataSize += uint64(len(vin.TxCustomTokenID))
		tokenDataSize += 4 // for VoutIndex
		tokenDataSize += uint64(vin.PaymentAddress.Size())
	}

	for _, vout := range tx.TxTokenData.Vouts {
		tokenDataSize += 8 // for value
		tokenDataSize += uint64(vout.PaymentAddress.Size())
	}

	return normalTxSize + tokenDataSize
}

// CreateTxCustomToken ...
func (txCustomToken *TxCustomToken) CreateTxCustomToken(senderKey *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoin []*privacy.InputCoin,
	fee uint64,
	tokenParams *CustomTokenParamTx,
	listCustomTokens map[common.Hash]TxCustomToken,
) (error) {
	// create normal txCustomToken
	normalTx := Tx{}
	err := normalTx.CreateTx(senderKey,
		paymentInfo,
		inputCoin,
		fee,
		true,
		nil)
	if err != nil {
		return err
	}
	// override txCustomToken type
	normalTx.Type = common.TxCustomTokenType

	txCustomToken.Tx = normalTx
	txCustomToken.TxTokenData = TxTokenData{}

	var handled = false

	// Add token data params
	switch tokenParams.TokenTxType {
	case CustomTokenInit:
		{
			handled = true
			txCustomToken.TxTokenData = TxTokenData{
				Type:           tokenParams.TokenTxType,
				PropertyName:   tokenParams.PropertyName,
				PropertySymbol: tokenParams.PropertySymbol,
				Vins:           nil,
				Vouts:          nil,
				Amount:         tokenParams.Amount,
			}
			var VoutsTemp []TxTokenVout

			receiver := tokenParams.Receiver[0]
			receiverAmount := receiver.Value
			VoutsTemp = append(VoutsTemp, TxTokenVout{
				PaymentAddress: receiver.PaymentAddress,
				Value:          receiverAmount,
			})

			txCustomToken.TxTokenData.Vouts = VoutsTemp
			hashInitToken, err := txCustomToken.TxTokenData.Hash()
			if err != nil {
				return errors.New("Can't handle this TokenTxType")
			}
			// validate PropertyID is the only one
			for customTokenID := range listCustomTokens {
				if hashInitToken.String() == customTokenID.String() {
					return errors.New("This token is existed in network")
				}
			}
			txCustomToken.TxTokenData.PropertyID = *hashInitToken

		}
	case CustomTokenTransfer:
		handled = true
		paymentTokenAmount := uint64(0)
		for _, receiver := range tokenParams.Receiver {
			paymentTokenAmount += receiver.Value
		}
		refundTokenAmount := tokenParams.vinsAmount - paymentTokenAmount
		txCustomToken.TxTokenData = TxTokenData{
			Type:           tokenParams.TokenTxType,
			PropertyName:   tokenParams.PropertyName,
			PropertySymbol: tokenParams.PropertySymbol,
			Vins:           nil,
			Vouts:          nil,
		}
		propertyID, _ := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
		txCustomToken.TxTokenData.PropertyID = *propertyID
		txCustomToken.TxTokenData.Vins = tokenParams.vins
		var VoutsTemp []TxTokenVout
		for _, receiver := range tokenParams.Receiver {
			receiverAmount := receiver.Value
			VoutsTemp = append(VoutsTemp, TxTokenVout{
				PaymentAddress: receiver.PaymentAddress,
				Value:          receiverAmount,
			})
		}
		VoutsTemp = append(VoutsTemp, TxTokenVout{
			PaymentAddress: tokenParams.vins[0].PaymentAddress,
			Value:          refundTokenAmount,
		})
		txCustomToken.TxTokenData.Vouts = VoutsTemp
	}

	if handled != true {
		return errors.New("Can't handle this TokenTxType")
	}
	return nil
}

func (tx *TxCustomToken) GetTxCustomTokenSignature(keyset cashec.KeySet) ([]byte, error) {
	buff := new(bytes.Buffer)
	json.NewEncoder(buff).Encode(tx)
	return keyset.Sign(buff.Bytes())
}
