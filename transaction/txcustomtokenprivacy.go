package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
	"github.com/incognitochain/incognito-chain/wallet"
)

// TxCustomTokenPrivacy is class tx which is inherited from P tx(supporting privacy) for fee
// and contain data(with supporting privacy format) to support issuing and transfer a custom token(token from end-user, look like erc-20)
// Dev or end-user can use this class tx to create an token type which use personal purpose
// TxCustomTokenPrivacy is an advance format of TxCustomToken
// so that user need to spend a lot fee to create this class tx
type TxCustomTokenPrivacy struct {
	Tx                                    // inherit from normal tx of P(supporting privacy) with a high fee to ensure that tx could contain a big data of privacy for token
	TxTokenPrivacyData TxTokenPrivacyData // supporting privacy format

	cachedHash *common.Hash // cached hash data of tx
}

func (txObj *TxCustomTokenPrivacy) UnmarshalJSON(data []byte) error {
	tx := Tx{}
	err := json.Unmarshal(data, &tx)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	temp := &struct {
		TxTokenPrivacyData interface{}
	}{}
	err = json.Unmarshal(data, &temp)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	TxTokenPrivacyDataJson, _ := json.MarshalIndent(temp.TxTokenPrivacyData, "", "\t")
	_ = json.Unmarshal(TxTokenPrivacyDataJson, &txObj.TxTokenPrivacyData)
	txObj.Tx = tx
	return nil
}

func (tx *TxCustomTokenPrivacy) String() string {
	// get hash of tx
	record := tx.Tx.Hash().String()

	// add more hash of tx custom token data privacy
	tokenPrivacyDataHash, _ := tx.TxTokenPrivacyData.Hash()
	record += tokenPrivacyDataHash.String()
	if tx.Metadata != nil {
		record += string(tx.Metadata.Hash()[:])
	}
	return record
}

func (txObj TxCustomTokenPrivacy) JSONString() string {
	data, err := json.MarshalIndent(txObj, "", "\t")
	if err != nil {
		Logger.log.Error(err)
		return ""
	}
	return string(data)
}

// Hash returns the hash of all fields of the transaction
func (tx *TxCustomTokenPrivacy) Hash() *common.Hash {
	if tx.cachedHash != nil {
		return tx.cachedHash
	}
	// final hash
	hash := common.HashH([]byte(tx.String()))
	return &hash
}

// GetTxActualSize computes the virtual size of a given transaction
// size of this tx = (normal TxNormal size) + (custom token data size)
func (tx *TxCustomTokenPrivacy) GetTxActualSize() uint64 {
	normalTxSize := tx.Tx.GetTxActualSize()

	tokenDataSize := uint64(0)
	tokenDataSize += tx.TxTokenPrivacyData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(tx.TxTokenPrivacyData.PropertyName))
	tokenDataSize += uint64(len(tx.TxTokenPrivacyData.PropertySymbol))
	tokenDataSize += uint64(len(tx.TxTokenPrivacyData.PropertyID))
	tokenDataSize += 4 // for TxTokenPrivacyData.Type
	tokenDataSize += 8 // for TxTokenPrivacyData.Amount

	meta := tx.Metadata
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return normalTxSize + uint64(math.Ceil(float64(tokenDataSize)/1024))
}

// Init -  build normal tx component and privacy custom token data
func (txCustomToken *TxCustomTokenPrivacy) Init(senderKey *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoin []*privacy.InputCoin,
	feeNativeCoin uint64,
	tokenParams *CustomTokenPrivacyParamTx,
	db database.DatabaseInterface,
	metaData metadata.Metadata,
	hasPrivacyCoin bool,
	hasPrivacyToken bool,
	shardID byte,
) *TransactionError {
	var err error
	// init data for tx PRV for fee
	normalTx := Tx{}
	err = normalTx.Init(senderKey,
		paymentInfo,
		inputCoin,
		feeNativeCoin,
		hasPrivacyCoin,
		db,
		nil,
		metaData)
	if err.(*TransactionError) != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	// override TxCustomTokenPrivacyType type
	normalTx.Type = common.TxCustomTokenPrivacyType
	txCustomToken.Tx = normalTx

	// check action type and create privacy custom toke data
	var handled = false
	// Add token data component
	switch tokenParams.TokenTxType {
	case CustomTokenInit:
		// case init a new privacy custom token
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
			temp.Proof.OutputCoins[0].CoinDetails.PublicKey = new(privacy.EllipticPoint)
			err := temp.Proof.OutputCoins[0].CoinDetails.PublicKey.Decompress(tokenParams.Receiver[0].PaymentAddress.Pk)
			if err != nil {
				return NewTransactionErr(UnexpectedErr, err)
			}
			temp.Proof.OutputCoins[0].CoinDetails.Randomness = privacy.RandScalar()

			sndOut := privacy.RandScalar()
			temp.Proof.OutputCoins[0].CoinDetails.SNDerivator = sndOut

			// create coin commitment
			temp.Proof.OutputCoins[0].CoinDetails.CommitAll()
			// get last byte
			temp.PubKeyLastByteSender = tokenParams.Receiver[0].PaymentAddress.Pk[len(tokenParams.Receiver[0].PaymentAddress.Pk)-1]

			// sign Tx
			temp.SigPubKey = tokenParams.Receiver[0].PaymentAddress.Pk
			temp.sigPrivKey = *senderKey
			err = temp.signTx()
			if err != nil {
				return NewTransactionErr(UnexpectedErr, errors.New("can't handle this TokenTxType"))
			}

			txCustomToken.TxTokenPrivacyData.TxNormal = temp
			hashInitToken, err := txCustomToken.TxTokenPrivacyData.Hash()
			if err != nil {
				return NewTransactionErr(UnexpectedErr, errors.New("can't handle this TokenTxType"))
			}

			if tokenParams.Mintable {
				propertyID, err := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
				if err != nil {
					return NewTransactionErr(UnexpectedErr, err)
				}
				txCustomToken.TxTokenPrivacyData.PropertyID = *propertyID
				txCustomToken.TxTokenPrivacyData.Mintable = true
			} else {
				//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
				newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), shardID))
				fmt.Println("INIT Tx Custom Token Privacy/ newHashInitToken", newHashInitToken)

				existed := db.PrivacyCustomTokenIDExisted(newHashInitToken)
				if existed {
					Logger.log.Error("INIT Tx Custom Token Privacy is Existed", newHashInitToken)
					return NewTransactionErr(UnexpectedErr, errors.New("this token is existed in network"))
				}
				existed = db.PrivacyCustomTokenIDCrossShardExisted(newHashInitToken)
				if existed {
					Logger.log.Error("INIT Tx Custom Token Privacy is Existed(crossshard)", newHashInitToken)
					return NewTransactionErr(UnexpectedErr, errors.New("this token is existed in network via cross shard"))
				}
				txCustomToken.TxTokenPrivacyData.PropertyID = newHashInitToken
				Logger.log.Infof("A new token privacy wil be issued with ID: %+v", txCustomToken.TxTokenPrivacyData.PropertyID.String())
			}
		}
	case CustomTokenTransfer:
		{
			handled = true
			// make a transfering for privacy custom token
			// fee always 0 and reuse function of normal tx for custom token ID
			temp := Tx{}
			propertyID, _ := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			existed := db.PrivacyCustomTokenIDExisted(*propertyID)
			existedCross := db.PrivacyCustomTokenIDCrossShardExisted(*propertyID)
			if !existed && !existedCross {
				return NewTransactionErr(UnexpectedErr, errors.New("invalid Token ID"))
			}
			Logger.log.Infof("Token %+v wil be transfered with", propertyID)
			txCustomToken.TxTokenPrivacyData = TxTokenPrivacyData{
				Type:           tokenParams.TokenTxType,
				PropertyName:   tokenParams.PropertyName,
				PropertySymbol: tokenParams.PropertySymbol,
				PropertyID:     *propertyID,
				Mintable:       tokenParams.Mintable,
			}
			err := temp.Init(senderKey,
				tokenParams.Receiver,
				tokenParams.TokenInput,
				tokenParams.Fee,
				hasPrivacyToken,
				db,
				propertyID,
				nil,
			)
			if err != nil {
				return err
			}
			txCustomToken.TxTokenPrivacyData.TxNormal = temp
		}
	}

	if !handled {
		return NewTransactionErr(UnexpectedErr, errors.New("can't handle this TokenTxType"))
	}
	return nil
}

func (tx *TxCustomTokenPrivacy) ValidateType() bool {
	return tx.Type == common.TxCustomTokenPrivacyType
}

func (tx *TxCustomTokenPrivacy) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	err := tx.validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	return nil
}

func (tx *TxCustomTokenPrivacy) ValidateTxWithBlockChain(
	bcr metadata.BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) error {
	err := tx.ValidateConstDoubleSpendWithBlockchain(bcr, shardID, db)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	return nil
}

func (tx *TxCustomTokenPrivacy) ValidateSanityData(bcr metadata.BlockchainRetriever) (bool, error) {
	result, err := tx.validateNormalTxSanityData()
	if err != nil {
		return result, NewTransactionErr(UnexpectedErr, err)
	}
	return result, nil
}

func (customTokenTx *TxCustomTokenPrivacy) ValidateTxByItself(
	hasPrivacyCoin bool,
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever,
	shardID byte,
) (bool, error) {
	if customTokenTx.TxTokenPrivacyData.Type == CustomTokenInit {
		return true, nil
	}
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	if ok, err := customTokenTx.ValidateTransaction(hasPrivacyCoin, db, shardID, prvCoinID); !ok {
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

func (customTokenTx *TxCustomTokenPrivacy) ValidateTransaction(hasPrivacyCoin bool, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) (bool, error) {
	ok, error := customTokenTx.Tx.ValidateTransaction(hasPrivacyCoin, db, shardID, tokenID)
	if ok {
		if customTokenTx.TxTokenPrivacyData.Type == CustomTokenInit {
			return customTokenTx.TxTokenPrivacyData.TxNormal.ValidateTransaction(false, db, shardID, &customTokenTx.TxTokenPrivacyData.PropertyID)
		} else {
			return customTokenTx.TxTokenPrivacyData.TxNormal.ValidateTransaction(customTokenTx.TxTokenPrivacyData.TxNormal.IsPrivacy(), db, shardID, &customTokenTx.TxTokenPrivacyData.PropertyID)
		}
	}
	return false, error
}

func (tx *TxCustomTokenPrivacy) GetProof() *zkp.PaymentProof {
	return tx.Proof
}

func (tx *TxCustomTokenPrivacy) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []metadata.Transaction,
	txsUsed []int,
	insts [][]string,
	instsUsed []int,
	shardID byte,
	bcr metadata.BlockchainRetriever,
) (bool, error) {
	if !tx.TxTokenPrivacyData.Mintable {
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

func (tx *TxCustomTokenPrivacy) GetTokenReceivers() ([][]byte, []uint64) {
	pubkeys := [][]byte{}
	amounts := []uint64{}
	proof := tx.TxTokenPrivacyData.TxNormal.Proof
	if proof == nil {
		return pubkeys, amounts
	}
	for _, coin := range proof.OutputCoins {
		coinPubKey := coin.CoinDetails.PublicKey.Compress()
		added := false
		// coinPubKey := vout.PaymentAddress.Pk
		for i, key := range pubkeys {
			if bytes.Equal(coinPubKey, key) {
				added = true
				amounts[i] += coin.CoinDetails.Value
				break
			}
		}
		if !added {
			pubkeys = append(pubkeys, coinPubKey)
			amounts = append(amounts, coin.CoinDetails.Value)
		}
	}
	return pubkeys, amounts
}

func (tx *TxCustomTokenPrivacy) GetTokenUniqueReceiver() (bool, []byte, uint64) {
	sender := []byte{}
	proof := tx.TxTokenPrivacyData.TxNormal.Proof
	if proof == nil {
		return false, []byte{}, 0
	}
	if len(proof.InputCoins) > 0 && proof.InputCoins[0].CoinDetails != nil {
		sender = proof.InputCoins[0].CoinDetails.PublicKey.Compress()
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

func (tx *TxCustomTokenPrivacy) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	unique, pk, amount := tx.GetTokenUniqueReceiver()
	return unique, pk, amount, &tx.TxTokenPrivacyData.PropertyID
}

func (tx *TxCustomTokenPrivacy) IsCoinsBurning() bool {
	proof := tx.TxTokenPrivacyData.TxNormal.Proof
	if proof == nil || len(proof.OutputCoins) == 0 {
		return false
	}
	senderPKBytes := []byte{}
	if len(proof.InputCoins) > 0 {
		senderPKBytes = tx.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
	}
	keyWalletBurningAccount, _ := wallet.Base58CheckDeserialize(common.BurningAddress)
	keysetBurningAccount := keyWalletBurningAccount.KeySet
	paymentAddressBurningAccount := keysetBurningAccount.PaymentAddress
	for _, outCoin := range proof.OutputCoins {
		outPKBytes := outCoin.CoinDetails.PublicKey.Compress()
		if !bytes.Equal(senderPKBytes, outPKBytes) && !bytes.Equal(outPKBytes, paymentAddressBurningAccount.Pk[:]) {
			return false
		}
	}
	return true
}

func (tx *TxCustomTokenPrivacy) CalculateTxValue() uint64 {
	proof := tx.TxTokenPrivacyData.TxNormal.Proof
	if proof == nil {
		return 0
	}
	if proof.OutputCoins == nil || len(proof.OutputCoins) == 0 {
		return 0
	}
	if proof.InputCoins == nil || len(proof.InputCoins) == 0 { // coinbase tx
		txValue := uint64(0)
		for _, outCoin := range proof.OutputCoins {
			txValue += outCoin.CoinDetails.Value
		}
		return txValue
	}

	senderPKBytes := proof.InputCoins[0].CoinDetails.PublicKey.Compress()
	txValue := uint64(0)
	for _, outCoin := range proof.OutputCoins {
		outPKBytes := outCoin.CoinDetails.PublicKey.Compress()
		if bytes.Equal(senderPKBytes, outPKBytes) {
			continue
		}
		txValue += outCoin.CoinDetails.Value
	}
	return txValue
}

func (tx *TxCustomTokenPrivacy) GetSigPubKey() []byte {
	return tx.TxTokenPrivacyData.TxNormal.SigPubKey
}

func (tx *TxCustomTokenPrivacy) GetTxFeeToken() uint64 {
	return tx.TxTokenPrivacyData.TxNormal.Fee
}

func (tx *TxCustomTokenPrivacy) GetTokenID() *common.Hash {
	return &tx.TxTokenPrivacyData.PropertyID
}
