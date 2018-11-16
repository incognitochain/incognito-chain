package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy/client"
)

// TxCustomToken ...
type TxCustomToken struct {
	Tx
	TxTokenData TxTokenData
	BoardType   uint8 // 1: DCB, 2: GOV
	BoardSigns  map[string][]byte
}

// Hash returns the hash of all fields of the transaction
func (tx TxCustomToken) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of txtoken
	record += tx.TxTokenData.PropertyName
	record += tx.TxTokenData.PropertySymbol
	record += strconv.Itoa(tx.TxTokenData.Type)
	record += strconv.Itoa(int(tx.TxTokenData.Amount))

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// ValidateTransaction ...
func (tx *TxCustomToken) ValidateTransaction() bool {
	// validate for normal tx
	if tx.Tx.ValidateTransaction() {
		// validate for tx token
		// TODO, verify signature
		return true
	}
	return false
}

// GetTxVirtualSize computes the virtual size of a given transaction
func (tx *TxCustomToken) GetTxVirtualSize() uint64 {
	var sizeVersion uint64 = 1  // int8
	var sizeType uint64 = 8     // string
	var sizeLockTime uint64 = 8 // int64
	var sizeFee uint64 = 8      // uint64
	var sizeDescs = uint64(common.Max(1, len(tx.Tx.Descs))) * EstimateJSDescSize()
	var sizejSPubKey uint64 = 64      // [64]byte
	var sizejSSig uint64 = 64         // [64]byte
	var sizeTokenName uint64 = 64     // string
	var sizeTokenSymbol uint64 = 64   // string
	var sizeTokenHash uint64 = 64     // string
	var sizeTokenAmount uint64 = 64   // string
	var sizeTokenTxType uint64 = 64   // string
	var sizeTokenReceiver uint64 = 64 // string

	estimateTxSizeInByte := sizeVersion
	estimateTxSizeInByte += sizeType
	estimateTxSizeInByte += sizeLockTime
	estimateTxSizeInByte += sizeFee
	estimateTxSizeInByte += sizeDescs
	estimateTxSizeInByte += sizejSPubKey
	estimateTxSizeInByte += sizejSSig
	estimateTxSizeInByte += sizeTokenName
	estimateTxSizeInByte += sizeTokenSymbol
	estimateTxSizeInByte += sizeTokenHash
	estimateTxSizeInByte += sizeTokenAmount
	estimateTxSizeInByte += sizeTokenTxType
	estimateTxSizeInByte += sizeTokenReceiver
	return uint64(math.Ceil(float64(estimateTxSizeInByte) / 1024))
}

// BuildTxCustomToken ...
func buildTxCustomToken(
	senderKey *client.SpendingKey,
	paymentInfo []*client.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*Tx,
	commitments map[byte]([][]byte),
	fee uint64,
	senderChainID byte,
	tokenParams *CustomTokenParamTx,
	listCustomTokens map[common.Hash]TxCustomToken,
) (*TxCustomToken, error) {

	// create normal tx
	normalTx, err := CreateTx(senderKey, paymentInfo, rts, usableTx, commitments, fee, senderChainID, true)
	if err != nil {
		return nil, err
	}
	normalTx.Type = common.TxCustomTokenType
	// end
	tx := &TxCustomToken{
		Tx:          *normalTx,
		TxTokenData: TxTokenData{},
	}

	var handled = false

	// Add token data params
	switch tokenParams.TokenTxType {
	case CustomTokenInit:
		{
			handled = true
			tx.TxTokenData = TxTokenData{
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

			tx.TxTokenData.Vouts = VoutsTemp
			hashInitToken, err := tx.TxTokenData.Hash()
			if err != nil {
				return nil, errors.New("Can't handle this TokenTxType")
			}
			// validate PropertyID is the only one
			for customTokenID := range listCustomTokens {
				if hashInitToken.String() == customTokenID.String() {
					return nil, errors.New("This token is existed in network")
				}
			}
			tx.TxTokenData.PropertyID = *hashInitToken

		}
	case CustomTokenTransfer:
		handled = true
		paymentTokenAmount := uint64(0)
		for _, receiver := range tokenParams.Receiver {
			paymentTokenAmount += receiver.Value
		}
		refundTokenAmount := tokenParams.vinsAmount - paymentTokenAmount
		tx.TxTokenData = TxTokenData{
			Type:           tokenParams.TokenTxType,
			PropertyName:   tokenParams.PropertyName,
			PropertySymbol: tokenParams.PropertySymbol,
			Vins:           nil,
			Vouts:          nil,
		}
		propertyID, _ := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
		tx.TxTokenData.PropertyID = *propertyID
		tx.TxTokenData.Vins = tokenParams.vins
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
		tx.TxTokenData.Vouts = VoutsTemp
	}

	if handled != true {
		return nil, errors.New("Can't handle this TokenTxType")
	}
	return tx, nil
}

// CreateTxCustomToken ...
func CreateTxCustomToken(senderKey *client.SpendingKey,
	paymentInfo []*client.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*Tx,
	commitments map[byte]([][]byte),
	fee uint64,
	senderChainID byte,
	tokenParams *CustomTokenParamTx,
	listCustomTokens map[common.Hash]TxCustomToken,
) (*TxCustomToken, error) {

	tx, err := buildTxCustomToken(
		senderKey,
		paymentInfo,
		rts,
		usableTx,
		commitments,
		fee,
		senderChainID,
		tokenParams,
		listCustomTokens,
	)
	if err != nil {
		return nil, err
	}

	fmt.Printf("jspubkey: %x\n", tx.JSPubKey)
	fmt.Printf("jssig: %x\n", tx.JSSig)
	return tx, nil
}

func GetTxCustomTokenSignature(tx *TxCustomToken, keyset cashec.KeySet) ([]byte, error) {
	buff := new(bytes.Buffer)
	json.NewEncoder(buff).Encode(tx)
	return keyset.Sign(buff.Bytes())
}
