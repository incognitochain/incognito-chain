package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"github.com/incognitochain/incognito-chain/wallet"
)

// TxCustomToken is class tx which is inherited from PRV tx(supporting privacy) for fee
// and contain data(vin, vout) to support issuing and transfer a custom token(token from end-user, look like erc-20)
// Dev or end-user can use this class tx to create an token type which use personal purpose
type TxCustomToken struct {
	Tx                      // inherit from normal tx of PRV(supporting privacy)
	TxTokenData TxTokenData // vin - vout format

	// Template data variable to process logic
	listUtxo   map[common.Hash]TxCustomToken
	cachedHash *common.Hash // cached hash data of tx
}

func (txObj *TxCustomToken) UnmarshalJSON(data []byte) error {
	tx := Tx{}
	err := json.Unmarshal(data, &tx)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	temp := &struct {
		TxTokenData interface{}
	}{}
	err = json.Unmarshal(data, &temp)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	txTokenDataJson, _ := json.MarshalIndent(temp.TxTokenData, "", "\t")
	_ = json.Unmarshal(txTokenDataJson, &txObj.TxTokenData)
	txObj.Tx = tx
	return nil
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
					return NewTransactionErr(DoubleSpend, nil)
				}
			}
		}
	}
	return nil
}

func (customTokenTx *TxCustomToken) ValidateTxWithCurrentMempool(
	mr metadata.MempoolRetriever,
) error {
	if customTokenTx.Type == common.TxRewardType || customTokenTx.Type == common.TxReturnStakingType {
		return errors.New("can not receive a salary tx | return staking tx from other node, this is a violation")
	}

	normalTx := customTokenTx.Tx
	err := normalTx.ValidateTxWithCurrentMempool(mr)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	txsInMem := mr.GetTxsInMem()
	for _, txInMem := range txsInMem {
		if txInMem.Tx.GetType() == common.TxCustomTokenType {
			err := customTokenTx.validateDoubleSpendCustomTokenOnTx(txInMem.Tx)
			if err != nil {
				return NewTransactionErr(UnexpectedErr, err)
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
		return NewTransactionErr(UnexpectedErr, err)
	}

	if len(listTxs) == 0 {
		if customTokenTx.TxTokenData.Type != CustomTokenInit && customTokenTx.TxTokenData.Type != CustomTokenCrossShard {
			return NewTransactionErr(TxNotExist, nil)
		}
	}

	if len(listTxs) > 0 {
		for _, txInBlocks := range listTxs {
			err := customTokenTx.validateDoubleSpendCustomTokenOnTx(txInBlocks)
			if err != nil {
				return NewTransactionErr(UnexpectedErr, err)
			}
		}
	}
	return nil
}

func (customTokenTx *TxCustomToken) ValidateTxWithBlockChain(
	bcr metadata.BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) error {
	if customTokenTx.GetType() == common.TxRewardType {
		return NewTransactionErr(UnexpectedErr, errors.New("Wrong salary tx"))
	}
	if customTokenTx.GetType() == common.TxReturnStakingType {
		return NewTransactionErr(UnexpectedErr, errors.New("Wrong return staking tx"))
	}

	if customTokenTx.Metadata != nil {
		isContinued, err := customTokenTx.Metadata.ValidateTxWithBlockChain(customTokenTx, bcr, shardID, db)
		if err != nil {
			return NewTransactionErr(UnexpectedErr, err)
		}
		if !isContinued {
			return nil
		}
	}

	err := customTokenTx.Tx.ValidateConstDoubleSpendWithBlockchain(bcr, shardID, db)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	return customTokenTx.validateDoubleSpendCustomTokenWithBlockchain(bcr)
}

func (txCustomToken *TxCustomToken) validateCustomTokenTxSanityData(bcr metadata.BlockchainRetriever) (bool, error) {
	ok, err := txCustomToken.Tx.ValidateSanityData(bcr)
	if err != nil || !ok {
		return ok, NewTransactionErr(UnexpectedErr, err)
	}
	vins := txCustomToken.TxTokenData.Vins
	zeroHash := common.Hash{}
	for _, vin := range vins {
		if len(vin.PaymentAddress.Pk) == 0 {
			return false, NewTransactionErr(WrongInput, nil)
		}
		if vin.Signature == "" {
			return false, NewTransactionErr(WrongSig, nil)
		}
		if vin.TxCustomTokenID.String() == zeroHash.String() {
			return false, NewTransactionErr(WrongInput, nil)
		}
	}
	vouts := txCustomToken.TxTokenData.Vouts
	for _, vout := range vouts {
		if len(vout.PaymentAddress.Pk) == 0 {
			return false, NewTransactionErr(WrongInput, nil)
		}
		if vout.Value == 0 {
			return false, NewTransactionErr(WrongInput, nil)
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
	result, err := customTokenTx.validateCustomTokenTxSanityData(bcr)
	if err == nil {
		return result, nil
	} else {
		return result, NewTransactionErr(UnexpectedErr, err)
	}
}

// ValidateTransaction - validate inheritance data from normal tx to check privacy and double spend for fee and transfer by PRV
// if pass normal tx validation, it continue check signature on (vin-vout) custom token data
func (tx *TxCustomToken) ValidateTransaction(hasPrivacy bool, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) (bool, error) {
	// validate for normal tx
	ok, err := tx.Tx.ValidateTransaction(hasPrivacy, db, shardID, tokenID)
	if ok {
		if len(tx.listUtxo) == 0 {
			return false, errors.New("Len listUtxo is 0")
		}
		if len(tx.TxTokenData.Vins) == 0 {
			return false, errors.New("Len Vins is 0")
		}
		totalVinAmount := uint64(0)
		for _, vin := range tx.TxTokenData.Vins {
			keySet := incognitokey.KeySet{}
			keySet.PaymentAddress = vin.PaymentAddress

			// get data from utxo
			utxo := tx.listUtxo[vin.TxCustomTokenID]
			vout := utxo.TxTokenData.Vouts[vin.VoutIndex]
			data := vout.Hash() // hash of vout in utxo
			signature, _, _ := base58.Base58Check{}.Decode(vin.Signature)
			ok, err := keySet.Verify(data[:], signature)
			if err != nil || !ok {
				return false, err
			}
			totalVinAmount += vout.Value
		}
		totalVoutAmount := uint64(0)
		for _, vout := range tx.TxTokenData.Vouts {
			totalVoutAmount += vout.Value
		}
		if totalVinAmount != totalVoutAmount {
			return false, errors.New("Vin amount <> Vout amount")
		}
		return true, nil
	}
	return false, err
}

func (customTokenTx *TxCustomToken) getListUTXOFromTxCustomToken(
	bcr metadata.BlockchainRetriever,
) bool {
	data := make(map[common.Hash]TxCustomToken)
	for _, vin := range customTokenTx.TxTokenData.Vins {
		_, _, _, utxo, err := bcr.GetTransactionByHash(vin.TxCustomTokenID)
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
	shardID byte,
) (bool, error) {
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	if customTokenTx.TxTokenData.Type == CustomTokenInit {
		if ok, err := customTokenTx.Tx.ValidateTransaction(hasPrivacy, db, shardID, prvCoinID); !ok {
			return false, err
		}
		if len(customTokenTx.TxTokenData.Vouts) != 1 {
			return false, errors.New("Length of Vouts != 1")
		}
		if len(customTokenTx.TxTokenData.Vins) != 0 && customTokenTx.TxTokenData.Vins != nil {
			return false, errors.New("Length of Vins != 0 and Vins is nil")
		}
		return true, nil
	}
	//Process CustomToken CrossShard
	if customTokenTx.TxTokenData.Type == CustomTokenCrossShard {
		if ok, err := customTokenTx.Tx.ValidateTransaction(hasPrivacy, db, shardID, prvCoinID); !ok {
			return false, err
		}
		if len(customTokenTx.listUtxo) != 0 {
			return false, errors.New("Length listUtxo != 0")
		}
		if len(customTokenTx.TxTokenData.Vins) != 0 {
			return false, errors.New("Length Vins != 0")
		}
		return true, nil
	}

	//Process CustomToken Transfer
	ok := customTokenTx.getListUTXOFromTxCustomToken(bcr)
	if !ok {
		return false, errors.New("getListUTXOFromTxCustomToken")
	}
	if ok, err := customTokenTx.ValidateTransaction(hasPrivacy, db, shardID, prvCoinID); !ok {
		return false, err
	}
	if customTokenTx.Metadata != nil {
		validateMetadata := customTokenTx.Metadata.ValidateMetadataByItself()
		if !validateMetadata {
			return validateMetadata, errors.New("Metadata is invalid")
		}
		return validateMetadata, nil
	}
	return true, nil
}

func (tx TxCustomToken) String() string {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of txtokendata
	txTokenDataHash, _ := tx.TxTokenData.Hash()
	record += txTokenDataHash.String()
	if tx.Metadata != nil {
		record += tx.Metadata.Hash().String()
	}
	return record
}

func (txObj TxCustomToken) JSONString() string {
	data, err := json.MarshalIndent(txObj, "", "\t")
	if err != nil {
		Logger.log.Error(err)
		return ""
	}
	return string(data)
}

// Hash returns the hash of all fields of the transaction
func (tx TxCustomToken) Hash() *common.Hash {
	if tx.cachedHash != nil {
		return tx.cachedHash
	}
	// final hash
	hash := common.HashH([]byte(tx.String()))
	return &hash
}

// GetTxActualSize computes the virtual size of a given transaction
// size of this tx = (normal TxNormal size) + (custom token data size)
func (tx *TxCustomToken) GetTxActualSize() uint64 {
	normalTxSize := tx.Tx.GetTxActualSize()

	tokenDataSize := uint64(0)

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

	// calculate metadata size if any
	meta := tx.Metadata
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return normalTxSize + uint64(math.Ceil(float64(tokenDataSize)/1024))
}

// CreateTxCustomToken ...
func (txCustomToken *TxCustomToken) Init(senderKey *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoin []*privacy.InputCoin,
	fee uint64,
	tokenParams *CustomTokenParamTx,
	db database.DatabaseInterface,
	metaData metadata.Metadata,
	hasPrivacyCoin bool,
	shardID byte,
) *TransactionError {
	var err error
	// create normal txCustomToken
	normalTx := Tx{}
	err = normalTx.Init(senderKey,
		paymentInfo,
		inputCoin,
		fee,
		hasPrivacyCoin,
		db,
		nil,
		metaData)
	if err.(*TransactionError) != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	// override txCustomToken type
	normalTx.Type = common.TxCustomTokenType

	txCustomToken.Tx = normalTx
	txCustomToken.TxTokenData = TxTokenData{}

	var handled = false

	// Add token data component
	switch tokenParams.TokenTxType {
	case CustomTokenCrossShard:
		{
			handled = true
			propertyID, err := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			if err != nil {
				return NewTransactionErr(UnexpectedErr, err)
			}
			txCustomToken.TxTokenData = TxTokenData{
				PropertyID:     *propertyID,
				Type:           tokenParams.TokenTxType,
				PropertyName:   tokenParams.PropertyName,
				PropertySymbol: tokenParams.PropertySymbol,
				Vins:           nil,
				Vouts:          nil,
				Amount:         tokenParams.Amount,
			}
			txCustomToken.TxTokenData.Vouts = tokenParams.Receiver
		}
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
			if tokenParams.Mintable {
				propertyID, err := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
				if err != nil {
					return NewTransactionErr(UnexpectedErr, err)
				}
				txCustomToken.TxTokenData.PropertyID = *propertyID
				txCustomToken.TxTokenData.Mintable = true

			} else {
				hashInitToken, err := txCustomToken.TxTokenData.Hash()
				if err != nil {
					return NewTransactionErr(WrongTokenTxType, err)
				}
				//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
				newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), shardID))
				fmt.Println("INIT Tx Custom Token/ newHashInitToken", newHashInitToken)
				//for customTokenID := range listCustomTokens {
				//	if newHashInitToken.String() == customTokenID.String() {
				//		fmt.Println("INIT Tx Custom Token/ Existed", customTokenID, customTokenID.String() == newHashInitToken.String())
				//		return NewTransactionErr(CustomTokenExisted, nil)
				//	}
				//}
				existed := db.CustomTokenIDExisted(newHashInitToken)
				if existed {
					Logger.log.Error("INIT Tx Custom Token is Existed", newHashInitToken)
					return NewTransactionErr(CustomTokenExisted, nil)
				}
				txCustomToken.TxTokenData.PropertyID = newHashInitToken
			}
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
			Mintable:       tokenParams.Mintable,
		}
		propertyID, _ := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
		//if _, ok := listCustomTokens[*propertyID]; !ok {
		//	return NewTransactionErr(UnexpectedErr, errors.New("invalid Token ID"))
		//}
		existed := db.CustomTokenIDExisted(*propertyID)
		if !existed {
			return NewTransactionErr(UnexpectedErr, errors.New("invalid Token ID"))
		}
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
		if refundTokenAmount < 0 {
			return NewTransactionErr(WrongInput, errors.New("input value less than output value"))
		}
		txCustomToken.TxTokenData.Vouts = VoutsTemp
	}
	txCustomToken.Type = common.TxCustomTokenType

	if !handled {
		return NewTransactionErr(WrongTokenTxType, nil)
	}
	return nil
}

func (tx *TxCustomToken) GetTxCustomTokenSignature(keyset incognitokey.KeySet) ([]byte, error) {
	buff := new(bytes.Buffer)
	json.NewEncoder(buff).Encode(tx)
	return keyset.Sign(buff.Bytes())
}

func (tx *TxCustomToken) IsPrivacy() bool {
	return false
}

func (tx *TxCustomToken) ValidateType() bool {
	return tx.Type == common.TxCustomTokenType
}

func (tx *TxCustomToken) GetProof() *zkp.PaymentProof {
	return tx.Proof
}

func (tx *TxCustomToken) GetSender() []byte {
	if len(tx.TxTokenData.Vins) == 0 {
		return nil
	}
	return tx.TxTokenData.Vins[0].PaymentAddress.Pk[:]
}

func (tx *TxCustomToken) GetTokenReceivers() ([][]byte, []uint64) {
	pubkeys := [][]byte{}
	amounts := []uint64{}
	for _, vout := range tx.TxTokenData.Vouts {
		added := false
		coinPubKey := vout.PaymentAddress.Pk
		for i, key := range pubkeys {
			if bytes.Equal(coinPubKey, key) {
				added = true
				amounts[i] += vout.Value
				break
			}
		}
		if !added {
			pubkeys = append(pubkeys, coinPubKey)
			amounts = append(amounts, vout.Value)
		}
	}
	return pubkeys, amounts
}

func (tx *TxCustomToken) GetTokenUniqueReceiver() (bool, []byte, uint64) {
	sender := []byte{}
	if len(tx.TxTokenData.Vins) > 0 {
		sender = tx.TxTokenData.Vins[0].PaymentAddress.Pk
	}
	pubkeys, amounts := tx.GetTokenReceivers()
	pubkey := []byte{}
	amount := uint64(0)
	count := 0
	for i, pk := range pubkeys {
		if !bytes.Equal(pk, sender) {
			pubkey = pk
			amount = amounts[i]
			count += 1
		}
	}
	return count == 1, pubkey, amount
}

func (tx *TxCustomToken) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	unique, pk, amount := tx.GetTokenUniqueReceiver()
	return unique, pk, amount, &tx.TxTokenData.PropertyID
}

func (tx *TxCustomToken) GetMetadataFromVinsTx(bcr metadata.BlockchainRetriever) (metadata.Metadata, error) {
	vins := tx.TxTokenData.Vins
	if len(vins) == 0 {
		return nil, nil
	}
	prevTxID := vins[0].TxCustomTokenID
	_, _, _, prevTx, err := bcr.GetTransactionByHash(prevTxID)
	if err != nil {
		return nil, err
	}
	if prevTx == nil {
		return nil, nil
	}
	return prevTx.GetMetadata(), nil
}

func (tx *TxCustomToken) CalculateTxValue() uint64 {
	vins := tx.TxTokenData.Vins
	vouts := tx.TxTokenData.Vouts
	if len(vins) == 0 { // coinbase tx
		txValue := uint64(0)
		for _, vout := range vouts {
			txValue += vout.Value
		}
		return txValue
	}

	senderPk := vins[0].PaymentAddress.Pk
	txValue := uint64(0)
	for _, vout := range vouts {
		if bytes.Equal(vout.PaymentAddress.Pk[:], senderPk[:]) {
			continue
		}
		txValue += vout.Value
	}
	return txValue
}

func (tx *TxCustomToken) IsCoinsBurning() bool {
	vins := tx.TxTokenData.Vins
	vouts := tx.TxTokenData.Vouts
	if len(vins) == 0 || len(vouts) == 0 {
		return false
	}
	senderPk := vins[0].PaymentAddress.Pk
	keyWalletBurningAccount, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	keysetBurningAccount := keyWalletBurningAccount.KeySet
	paymentAddressBurningAccount := keysetBurningAccount.PaymentAddress
	for _, vout := range vouts {
		outPKBytes := vout.PaymentAddress.Pk
		if !bytes.Equal(senderPk[:], outPKBytes[:]) && !bytes.Equal(outPKBytes[:], paymentAddressBurningAccount.Pk[:]) {
			return false
		}
	}
	return true
}

func (tx *TxCustomToken) GetTokenID() *common.Hash {
	return &tx.TxTokenData.PropertyID
}

func (tx *TxCustomToken) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []metadata.Transaction,
	txsUsed []int,
	insts [][]string,
	instsUsed []int,
	shardID byte,
	bcr metadata.BlockchainRetriever,
) (bool, error) {
	if !tx.TxTokenData.Mintable {
		return true, nil
	}
	meta := tx.Metadata
	if meta == nil {
		Logger.log.Error("Mintable custom token must contain metadata")
		return false, nil
	}
	// TODO: uncomment below as we have fully validation for all tx/meta types in order to check strictly miner created tx
	if !meta.IsMinerCreatedMetaType() {
		return false, nil
	}
	return meta.VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock, txsUsed, insts, instsUsed, shardID, tx, bcr)
}
