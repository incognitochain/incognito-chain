package transaction

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"errors"
	"github.com/ninjadotorg/constant/privacy-protocol/zero-knowledge"
)

// TxCustomTokenPrivacy is class tx which is inherited from constant tx(supporting privacy) for fee
// and contain data(with supporting privacy format) to support issuing and transfer a custom token(token from end-user, look like erc-20)
// Dev or end-user can use this class tx to create an token type which use personal purpose
// TxCustomTokenPrivacy is an advance format of TxCustomToken
// so that user need to spend a lot fee to create this class tx
type TxCustomTokenPrivacy struct {
	TxNormal           Tx                 // inherit from normal tx of constant(supporting privacy) with a high fee to ensure that tx could contain a big data of privacy for token
	TxTokenPrivacyData TxTokenPrivacyData // supporting privacy format
}

// Hash returns the hash of all fields of the transaction
func (tx *TxCustomTokenPrivacy) Hash() *common.Hash {
	// get hash of tx
	record := tx.TxNormal.Hash().String()

	// add more hash of tx custom token data privacy
	tokenPrivacyDataHash, _ := tx.TxTokenPrivacyData.Hash()
	record += tokenPrivacyDataHash.String()

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

func (tx *TxCustomTokenPrivacy) ValidateTransaction(hasPrivacy bool, db database.DatabaseInterface, chainID byte) bool {
	// validate for normal tx
	if tx.TxNormal.ValidateTransaction(hasPrivacy, db, chainID) {
		// TODO
		return true
	}
	return false
}

// GetTxActualSize computes the virtual size of a given transaction
// size of this tx = (normal TxNormal size) + (custom token data size)
func (tx *TxCustomTokenPrivacy) GetTxActualSize() uint64 {
	normalTxSize := tx.TxNormal.GetTxActualSize()

	tokenDataSize := uint64(0)
	tokenDataSize += tx.TxTokenPrivacyData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(tx.TxTokenPrivacyData.PropertyName))
	tokenDataSize += uint64(len(tx.TxTokenPrivacyData.PropertySymbol))
	tokenDataSize += uint64(len(tx.TxTokenPrivacyData.PropertyID))
	tokenDataSize += 4 // for TxTokenPrivacyData.Type
	tokenDataSize += 8 // for TxTokenPrivacyData.Amount

	return normalTxSize + tokenDataSize
}

// CreateTxCustomToken ...
func (txCustomToken *TxCustomTokenPrivacy) Init(senderKey *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoin []*privacy.InputCoin,
	fee uint64,
	tokenParams *CustomTokenPrivacyParamTx,
	listCustomTokens map[common.Hash]TxCustomToken,
	db database.DatabaseInterface,
) (error) {

	// init data for tx constant for fee
	normalTx := Tx{}
	err := normalTx.Init(senderKey,
		paymentInfo,
		inputCoin,
		fee,
		false,
		nil,
		nil)
	if err != nil {
		return err
	}
	// override TxCustomTokenPrivacyType type
	normalTx.Type = common.TxCustomTokenPrivacyType
	txCustomToken.TxNormal = normalTx

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

			// issue token with data of privacy
			temp := Tx{}
			temp.Proof = new(zkp.PaymentProof)
			temp.Proof.OutputCoins = make([]*privacy.OutputCoin, 1)
			temp.Proof.OutputCoins[0] = new(privacy.OutputCoin)
			temp.Proof.OutputCoins[0].CoinDetails = new(privacy.Coin)
			temp.Proof.OutputCoins[0].CoinDetails.Value = tokenParams.Amount
			temp.Proof.OutputCoins[0].CoinDetails.PublicKey, err = privacy.DecompressKey(tokenParams.Receiver[0].PaymentAddress.Pk)
			if err != nil {
				return err
			}
			temp.Proof.OutputCoins[0].CoinDetails.Randomness = privacy.RandInt()

			sndOut := privacy.RandInt()
			for true {
				lastByte := tokenParams.Receiver[0].PaymentAddress.Pk[len(tokenParams.Receiver[0].PaymentAddress.Pk)-1]
				chainIdSender, err := common.GetTxSenderChain(lastByte)

				tokenID := &common.Hash{}
				tokenID.SetBytes(common.ConstantID[:])
				ok, err := CheckSNDerivatorExistence(tokenID, sndOut, chainIdSender, db)
				if err != nil {
					return err
				}
				if ok {
					sndOut = privacy.RandInt()
				} else {
					break
				}
			}

			temp.Proof.OutputCoins[0].CoinDetails.SNDerivator = sndOut

			// create coin commitment
			temp.Proof.OutputCoins[0].CoinDetails.CommitAll()
			// get last byte
			temp.PubKeyLastByteSender = tokenParams.Receiver[0].PaymentAddress.Pk[len(tokenParams.Receiver[0].PaymentAddress.Pk)-1]

			// sign Tx
			temp.SigPubKey = tokenParams.Receiver[0].PaymentAddress.Pk
			temp.sigPrivKey = *senderKey
			err = temp.SignTx(false)

			txCustomToken.TxTokenPrivacyData.TxNormal = temp
			hashInitToken, err := txCustomToken.TxTokenPrivacyData.Hash()
			if err != nil {
				return errors.New("Can't handle this TokenTxType")
			}
			// validate PropertyID is the only one
			for customTokenID := range listCustomTokens {
				if hashInitToken.String() == customTokenID.String() {
					return errors.New("This token is existed in network")
				}
			}
			txCustomToken.TxTokenPrivacyData.PropertyID = *hashInitToken
		}
	case CustomTokenTransfer:
		// create privacy tx for token
		temp := Tx{}
		propertyID, _ := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
		temp.Init(senderKey,
			tokenParams.Receiver,
			tokenParams.InputCoin,
			0,
			true,
			db,
			propertyID,
		)
		txCustomToken.TxTokenPrivacyData.TxNormal = temp
	}

	if handled != true {
		return errors.New("Can't handle this TokenTxType")
	}
	return nil
}
