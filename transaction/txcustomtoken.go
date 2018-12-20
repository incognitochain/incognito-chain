package transaction

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/ninjadotorg/constant/cashec"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/database"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy-protocol"
	"github.com/ninjadotorg/constant/wallet"
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

func (customTokentx *TxCustomToken) validateDoubleSpendCustomTokenOnTx(
	txInBlock metadata.Transaction,
) error {
	temp := txInBlock.(*TxCustomToken)
	for _, vin := range temp.TxTokenData.Vins {
		for _, item := range customTokentx.TxTokenData.Vins {
			if vin.TxCustomTokenID.String() == item.TxCustomTokenID.String() {
				if vin.VoutIndex == item.VoutIndex {
					return errors.New("Double spend")
				}
			}
		}
	}
	return nil
}

func (customTokenTx *TxCustomToken) ValidateTxWithCurrentMempool(
	mr metadata.MempoolRetriever,
) error {
	if customTokenTx.Type == common.TxSalaryType {
		return errors.New("Can not receive a salary tx from other node, this is a violation")
	}

	normalTx := customTokenTx.Tx
	err := normalTx.ValidateTxWithCurrentMempool(mr)
	if err != nil {
		return err
	}
	txsInMem := mr.GetTxsInMem()
	for _, txInMem := range txsInMem {
		if txInMem.Tx.GetType() == common.TxCustomTokenType {
			err := customTokenTx.validateDoubleSpendCustomTokenOnTx(txInMem.Tx)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (customTokenTx *TxCustomToken) validateDoubleSpendCustomTokenWithBlockchain(
	bcr metadata.BlockchainRetriever,
) error {
	listTxs, err := bcr.GetCustomTokenTxs(&customTokenTx.TxTokenData.PropertyID)
	if err != nil {
		return err
	}

	if len(listTxs) == 0 {
		if customTokenTx.TxTokenData.Type != CustomTokenInit {
			return errors.New("Not exist tx for this ")
		}
	}

	if len(listTxs) > 0 {
		for _, txInBlocks := range listTxs {
			err := customTokenTx.validateDoubleSpendCustomTokenOnTx(txInBlocks)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (customTokenTx *TxCustomToken) ValidateTxWithBlockChain(
	bcr metadata.BlockchainRetriever,
	chainID byte,
	db database.DatabaseInterface,
) error {
	if customTokenTx.GetType() == common.TxSalaryType {
		return nil
	}
	if customTokenTx.Metadata != nil {
		isContinued, err := customTokenTx.Metadata.ValidateTxWithBlockChain(customTokenTx, bcr, chainID, db)
		if err != nil || !isContinued {
			return err
		}
	}

	// TODO: add validate signs for multisig tx

	err := customTokenTx.Tx.ValidateConstDoubleSpendWithBlockchain(bcr, chainID, db)
	if err != nil {
		return err
	}
	return customTokenTx.validateDoubleSpendCustomTokenWithBlockchain(bcr)
}

func (txCustomToken *TxCustomToken) validateCustomTokenTxSanityData(bcr metadata.BlockchainRetriever) (bool, error) {
	ok, err := txCustomToken.Tx.ValidateSanityData(bcr)
	if err != nil || !ok {
		return ok, err
	}
	vins := txCustomToken.TxTokenData.Vins
	zeroHash := common.Hash{}
	for _, vin := range vins {
		if len(vin.PaymentAddress.Pk) == 0 {
			return false, errors.New("Wrong input transaction")
		}
		// TODO: @0xbunyip - should move logic below to BuySellDCBResponse metadata's logic
		// dbcAccount, _ := wallet.Base58CheckDeserialize(common.DCBAddress)
		// if bytes.Equal(vin.PaymentAddress.Pk, dbcAccount.KeySet.PaymentAddress.Pk) {
		// 	if !allowToUseDCBFund {
		// 		return false, errors.New("Cannot use DCB's fund here")
		// 	}
		// }
		if vin.Signature == "" {
			return false, errors.New("Wrong signature")
		}
		if vin.TxCustomTokenID.String() == zeroHash.String() {
			return false, errors.New("Wrong input transaction")
		}
	}
	vouts := txCustomToken.TxTokenData.Vouts
	for _, vout := range vouts {
		if len(vout.PaymentAddress.Pk) == 0 {
			return false, errors.New("Wrong input transaction")
		}
		if vout.Value == 0 {
			return false, errors.New("Wrong input transaction")
		}
	}
	return true, nil
}

func (customTokenTx *TxCustomToken) ValidateSanityData(bcr metadata.BlockchainRetriever) (bool, error) {
	if customTokenTx.Metadata != nil {
		isContinued, ok, err := customTokenTx.Metadata.ValidateSanityData(bcr, customTokenTx)
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	return customTokenTx.validateCustomTokenTxSanityData(bcr)
}

// ValidateTransaction - validate inheritance data from normal tx to check privacy and double spend for fee and transfer by constant
// if pass normal tx validation, it continue check signature on (vin-vout) custom token data
func (tx *TxCustomToken) ValidateTransaction(hasPrivacy bool, db database.DatabaseInterface, chainID byte, tokenID *common.Hash) bool {
	// validate for normal tx
	if tx.Tx.ValidateTransaction(hasPrivacy, db, chainID, tokenID) {
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
			signature, _, _ := base58.Base58Check{}.Decode(vin.Signature)
			ok, err := keySet.Verify(data[:], signature)
			if err != nil {
				return false
			}
			return ok
		}
		return true
	}
	return false
}

func (customTokenTx *TxCustomToken) getListUTXOFromTxCustomToken(
	bcr metadata.BlockchainRetriever,
) bool {
	data := make(map[common.Hash]TxCustomToken)
	for _, vin := range customTokenTx.TxTokenData.Vins {
		_, _, _, utxo, err := bcr.GetTransactionByHash(&vin.TxCustomTokenID)
		if err != nil {
			// Logger.log.Error(err)
			return false
		}
		data[vin.TxCustomTokenID] = *(utxo.(*TxCustomToken))
	}
	if len(data) == 0 {
		// Logger.log.Error(errors.New("Can not find any utxo for TxCustomToken"))
		return false
	}
	customTokenTx.SetListUtxo(data)
	return true
}

func (customTokenTx *TxCustomToken) ValidateTxByItself(
	hasPrivacy bool,
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever,
	chainID byte,
) bool {
	if customTokenTx.TxTokenData.Type == CustomTokenInit {
		return true
	}
	ok := customTokenTx.getListUTXOFromTxCustomToken(bcr)
	if !ok {
		return false
	}
	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	ok = customTokenTx.ValidateTransaction(hasPrivacy, db, chainID, constantTokenID)
	if !ok {
		return false
	}
	if customTokenTx.Metadata != nil {
		return customTokenTx.Metadata.ValidateMetadataByItself()
	}
	return true
}

// Hash returns the hash of all fields of the transaction
func (tx TxCustomToken) Hash() *common.Hash {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of txtokendata
	txTokenDataHash, _ := tx.TxTokenData.Hash()
	record += txTokenDataHash.String()
	if tx.Metadata != nil {
		record += string(tx.Metadata.Hash()[:])
	}

	// final hash
	hash := common.DoubleHashH([]byte(record))
	return &hash
}

// GetTxActualSize computes the virtual size of a given transaction
// size of this tx = (normal TxNormal size) + (custom token data size)
func (tx *TxCustomToken) GetTxActualSize() uint64 {
	normalTxSize := tx.Tx.GetTxActualSize()

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
func (txCustomToken *TxCustomToken) Init(senderKey *privacy.SpendingKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoin []*privacy.InputCoin,
	fee uint64,
	tokenParams *CustomTokenParamTx,
	listCustomTokens map[common.Hash]TxCustomToken,
) error {
	// create normal txCustomToken
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
		if refundTokenAmount > 0 {
			VoutsTemp = append(VoutsTemp, TxTokenVout{
				PaymentAddress: tokenParams.vins[0].PaymentAddress,
				Value:          refundTokenAmount,
			})
		}
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

func (tx *TxCustomToken) GetAmountOfVote() uint64 {
	sum := uint64(0)
	for _, vout := range tx.TxTokenData.Vouts {
		voteAccount, _ := wallet.Base58CheckDeserialize(common.VoteAddress)
		pubKey := string(voteAccount.KeySet.PaymentAddress.Pk)
		if string(vout.PaymentAddress.Pk) == string(pubKey) {
			sum += vout.Value
		}
	}
	return sum
}

func (tx *TxCustomToken) IsPrivacy() bool {
	return false
}

func (tx *TxCustomToken) ValidateType() bool {
	return tx.Type == common.TxCustomTokenType
}
