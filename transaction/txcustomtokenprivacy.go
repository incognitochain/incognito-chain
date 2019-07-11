package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"

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

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) UnmarshalJSON(data []byte) error {
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
	_ = json.Unmarshal(TxTokenPrivacyDataJson, &txCustomTokenPrivacy.TxTokenPrivacyData)
	txCustomTokenPrivacy.Tx = tx
	return nil
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) String() string {
	// get hash of tx
	record := txCustomTokenPrivacy.Tx.Hash().String()

	// add more hash of tx custom token data privacy
	tokenPrivacyDataHash, _ := txCustomTokenPrivacy.TxTokenPrivacyData.Hash()
	record += tokenPrivacyDataHash.String()
	if txCustomTokenPrivacy.Metadata != nil {
		record += string(txCustomTokenPrivacy.Metadata.Hash()[:])
	}
	return record
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) JSONString() string {
	data, err := json.MarshalIndent(txCustomTokenPrivacy, "", "\t")
	if err != nil {
		Logger.log.Error(err)
		return ""
	}
	return string(data)
}

// Hash returns the hash of all fields of the transaction
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) Hash() *common.Hash {
	if txCustomTokenPrivacy.cachedHash != nil {
		return txCustomTokenPrivacy.cachedHash
	}
	// final hash
	hash := common.HashH([]byte(txCustomTokenPrivacy.String()))
	return &hash
}

// GetTxActualSize computes the virtual size of a given transaction
// size of this tx = (normal TxNormal size) + (custom token data size)
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) GetTxActualSize() uint64 {
	normalTxSize := txCustomTokenPrivacy.Tx.GetTxActualSize()

	tokenDataSize := uint64(0)
	tokenDataSize += txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(txCustomTokenPrivacy.TxTokenPrivacyData.PropertyName))
	tokenDataSize += uint64(len(txCustomTokenPrivacy.TxTokenPrivacyData.PropertySymbol))
	tokenDataSize += uint64(len(txCustomTokenPrivacy.TxTokenPrivacyData.PropertyID))
	tokenDataSize += 4 // for TxTokenPrivacyData.Type
	tokenDataSize += 8 // for TxTokenPrivacyData.Amount

	meta := txCustomTokenPrivacy.Metadata
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return normalTxSize + uint64(math.Ceil(float64(tokenDataSize)/1024))
}

// Init -  build normal tx component and privacy custom token data
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) Init(senderKey *privacy.PrivateKey,
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
	txCustomTokenPrivacy.Tx = normalTx

	// check action type and create privacy custom toke data
	var handled = false
	// Add token data component
	switch tokenParams.TokenTxType {
	case CustomTokenInit:
		// case init a new privacy custom token
		{
			handled = true
			txCustomTokenPrivacy.TxTokenPrivacyData = TxTokenPrivacyData{
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

			txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal = temp
			hashInitToken, err := txCustomTokenPrivacy.TxTokenPrivacyData.Hash()
			if err != nil {
				return NewTransactionErr(UnexpectedErr, errors.New("can't handle this TokenTxType"))
			}

			if tokenParams.Mintable {
				propertyID, err := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
				if err != nil {
					return NewTransactionErr(UnexpectedErr, err)
				}
				txCustomTokenPrivacy.TxTokenPrivacyData.PropertyID = *propertyID
				txCustomTokenPrivacy.TxTokenPrivacyData.Mintable = true
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
				txCustomTokenPrivacy.TxTokenPrivacyData.PropertyID = newHashInitToken
				Logger.log.Infof("A new token privacy wil be issued with ID: %+v", txCustomTokenPrivacy.TxTokenPrivacyData.PropertyID.String())
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
			txCustomTokenPrivacy.TxTokenPrivacyData = TxTokenPrivacyData{
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
			txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal = temp
		}
	}

	if !handled {
		return NewTransactionErr(UnexpectedErr, errors.New("can't handle this TokenTxType"))
	}
	return nil
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) ValidateType() bool {
	return txCustomTokenPrivacy.Type == common.TxCustomTokenPrivacyType
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	err := txCustomTokenPrivacy.validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	return nil
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) ValidateTxWithBlockChain(
	bcr metadata.BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) error {
	err := txCustomTokenPrivacy.ValidateConstDoubleSpendWithBlockchain(bcr, shardID, db)
	if err != nil {
		return NewTransactionErr(UnexpectedErr, err)
	}
	return nil
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) ValidateSanityData(bcr metadata.BlockchainRetriever) (bool, error) {
	result, err := txCustomTokenPrivacy.validateNormalTxSanityData()
	if err != nil {
		return result, NewTransactionErr(UnexpectedErr, err)
	}
	return result, nil
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) ValidateTxByItself(
	hasPrivacyCoin bool,
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever,
	shardID byte,
) (bool, error) {
	if txCustomTokenPrivacy.TxTokenPrivacyData.Type == CustomTokenInit {
		return true, nil
	}
	prvCoinID := &common.Hash{}
	prvCoinID.SetBytes(common.PRVCoinID[:])
	if ok, err := txCustomTokenPrivacy.ValidateTransaction(hasPrivacyCoin, db, shardID, prvCoinID); !ok {
		return false, err
	}

	if txCustomTokenPrivacy.Metadata != nil {
		validateMetadata := txCustomTokenPrivacy.Metadata.ValidateMetadataByItself()
		if !validateMetadata {
			return validateMetadata, errors.New("Metadata is invalid")
		}
		return validateMetadata, nil
	}
	return true, nil
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) ValidateTransaction(hasPrivacyCoin bool, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) (bool, error) {
	ok, error := txCustomTokenPrivacy.Tx.ValidateTransaction(hasPrivacyCoin, db, shardID, tokenID)
	if ok {
		if txCustomTokenPrivacy.TxTokenPrivacyData.Type == CustomTokenInit {
			return txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.ValidateTransaction(false, db, shardID, &txCustomTokenPrivacy.TxTokenPrivacyData.PropertyID)
		} else {
			return txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.ValidateTransaction(txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.IsPrivacy(), db, shardID, &txCustomTokenPrivacy.TxTokenPrivacyData.PropertyID)
		}
	}
	return false, error
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) GetProof() *zkp.PaymentProof {
	return txCustomTokenPrivacy.Proof
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []metadata.Transaction,
	txsUsed []int,
	insts [][]string,
	instsUsed []int,
	shardID byte,
	bcr metadata.BlockchainRetriever,
) (bool, error) {
	if !txCustomTokenPrivacy.TxTokenPrivacyData.Mintable {
		return true, nil
	}
	meta := txCustomTokenPrivacy.Metadata
	if meta == nil {
		Logger.log.Error("Mintable custom token must contain metadata")
		return false, nil
	}
	// TODO: uncomment below as we have fully validation for all tx/meta types in order to check strictly miner created tx
	if !meta.IsMinerCreatedMetaType() {
		return false, nil
	}
	return meta.VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock, txsUsed, insts, instsUsed, shardID, txCustomTokenPrivacy, bcr)
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) GetTokenReceivers() ([][]byte, []uint64) {
	pubkeys := [][]byte{}
	amounts := []uint64{}
	proof := txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.Proof
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

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) GetTokenUniqueReceiver() (bool, []byte, uint64) {
	sender := []byte{}
	proof := txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.Proof
	if proof == nil {
		return false, []byte{}, 0
	}
	if len(proof.InputCoins) > 0 && proof.InputCoins[0].CoinDetails != nil {
		sender = proof.InputCoins[0].CoinDetails.PublicKey.Compress()
	}
	pubkeys, amounts := txCustomTokenPrivacy.GetTokenReceivers()
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

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	unique, pk, amount := txCustomTokenPrivacy.GetTokenUniqueReceiver()
	return unique, pk, amount, &txCustomTokenPrivacy.TxTokenPrivacyData.PropertyID
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) IsCoinsBurning() bool {
	proof := txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.Proof
	if proof == nil || len(proof.OutputCoins) == 0 {
		return false
	}
	senderPKBytes := []byte{}
	if len(proof.InputCoins) > 0 {
		senderPKBytes = txCustomTokenPrivacy.Proof.InputCoins[0].CoinDetails.PublicKey.Compress()
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

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) CalculateTxValue() uint64 {
	proof := txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.Proof
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
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) ListSerialNumbersHashH() []common.Hash {
	tx := txCustomTokenPrivacy.Tx
	result := []common.Hash{}
	if tx.Proof != nil {
		for _, d := range tx.Proof.InputCoins {
			hash := common.HashH(d.CoinDetails.SerialNumber.Compress())
			result = append(result, hash)
		}
	}
	customTokenPrivacy := txCustomTokenPrivacy.TxTokenPrivacyData
	if customTokenPrivacy.TxNormal.Proof != nil {
		for _, d := range tx.Proof.InputCoins {
			hash := common.HashH(d.CoinDetails.SerialNumber.Compress())
			result = append(result, hash)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result
}
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) GetSigPubKey() []byte {
	return txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.SigPubKey
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) GetTxFeeToken() uint64 {
	return txCustomTokenPrivacy.TxTokenPrivacyData.TxNormal.Fee
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) GetTokenID() *common.Hash {
	return &txCustomTokenPrivacy.TxTokenPrivacyData.PropertyID
}
