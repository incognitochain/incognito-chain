package transaction

/*
import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"errors"
)

// TxCustomTokenPrivacy is class tx which is inherited from constant tx(supporting privacy) for fee
// and contain data(with supporting privacy format) to support issuing and transfer a custom token(token from end-user, look like erc-20)
// Dev or end-user can use this class tx to create an token type which use personal purpose
// TxCustomTokenPrivacy is an advance format of TxCustomToken
// so that user need to spend a lot fee to create this class tx
type TxCustomTokenPrivacy struct {
	TxNormal                              // inherit from normal tx of constant(supporting privacy) with a high fee to ensure that tx could contain a big data of privacy for token
	TxTokenPrivacyData TxTokenPrivacyData // supporting privacy format
}

// Hash returns the hash of all fields of the transaction
func (tx TxCustomTokenPrivacy) Hash() *common.Hash {
	// get hash of tx
	record := tx.TxNormal.Hash().String()

	// add more hash of tx custom token data privacy
	txTokenPtivacyDataHash, _ := tx.TxTokenPrivacyData.Hash()
	record += txTokenPtivacyDataHash.String()

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxCustomTokenPrivacy) ValidateTransaction() bool {
	// validate for normal tx
	if tx.TxNormal.ValidateTransaction() {
		// TODO
		return true
	}
	return false
}

// GetTxVirtualSize computes the virtual size of a given transaction
// size of this tx = (normal TxNormal size) + (custom token data size)
func (tx *TxCustomTokenPrivacy) GetTxVirtualSize() uint64 {
	normalTxSize := tx.TxNormal.GetTxVirtualSize()

	tokenDataSize := uint64(0)

	tokenDataSize += uint64(len(tx.TxTokenPrivacyData.PropertyName))
	tokenDataSize += uint64(len(tx.TxTokenPrivacyData.PropertyName))
	tokenDataSize += uint64(len(tx.TxTokenPrivacyData.PropertyID))
	tokenDataSize += 4 // for TxTokenData.Type

	// TODO for txcustomtokendata

	return normalTxSize + tokenDataSize
}

// CreateTxCustomToken ...
func CreateTxCustomTokenPrivacy(senderKey *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	rts map[byte]*common.Hash,
	usableTx map[byte][]*TxNormal,
	commitments map[byte]([][]byte),
	fee uint64,
	senderChainID byte,
	tokenParams *CustomTokenParamTx,
	listCustomTokens map[common.Hash]TxCustomToken,
) (*TxCustomTokenPrivacy, error) {
	// create normal txCustomToken
	normalTx, err := CreateTx(senderKey, paymentInfo, rts, usableTx, commitments, fee, senderChainID, true)
	if err != nil {
		return nil, err
	}
	// override txCustomToken type
	normalTx.Type = common.TxCustomTokenType

	txCustomToken := &TxCustomTokenPrivacy{
		TxNormal:           *normalTx,
		TxTokenPrivacyData: TxTokenPrivacyData{},
	}

	var handled = false

	// Add token data params
	switch tokenParams.TokenTxType {
	case CustomTokenInit:
		{
			handled = true
			txCustomToken.TxTokenPrivacyData = TxTokenPrivacyData{
				Type:           tokenParams.TokenTxType,
				PropertyName:   tokenParams.PropertyName,
				PropertySymbol: tokenParams.PropertySymbol,
				Amount:         tokenParams.Amount,
			}
			// TODO create descs
			hashInitToken, err := txCustomToken.TxTokenPrivacyData.Hash()
			if err != nil {
				return nil, errors.New("Can't handle this TokenTxType")
			}
			// validate PropertyID is the only one
			for customTokenID := range listCustomTokens {
				if hashInitToken.String() == customTokenID.String() {
					return nil, errors.New("This token is existed in network")
				}
			}
			txCustomToken.TxTokenPrivacyData.PropertyID = *hashInitToken

		}
	case CustomTokenTransfer:
		handled = true
		paymentTokenAmount := uint64(0)
		for _, receiver := range tokenParams.Receiver {
			paymentTokenAmount += receiver.Value
		}
		refundTokenAmount := tokenParams.vinsAmount - paymentTokenAmount
		txCustomToken.TxTokenPrivacyData = TxTokenPrivacyData{
			Type:           tokenParams.TokenTxType,
			PropertyName:   tokenParams.PropertyName,
			PropertySymbol: tokenParams.PropertySymbol,
			Descs:          nil,
		}
		_ = refundTokenAmount
		propertyID, _ := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
		txCustomToken.TxTokenPrivacyData.PropertyID = *propertyID
		// TODO create descs
	}

	if handled != true {
		return nil, errors.New("Can't handle this TokenTxType")
	}
	return txCustomToken, nil
}
*/
