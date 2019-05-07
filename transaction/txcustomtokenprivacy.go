package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/privacy"
	zkp "github.com/constant-money/constant-chain/privacy/zeroknowledge"
)

// TxCustomTokenPrivacy is class tx which is inherited from constant tx(supporting privacy) for fee
// and contain data(with supporting privacy format) to support issuing and transfer a custom token(token from end-user, look like erc-20)
// Dev or end-user can use this class tx to create an token type which use personal purpose
// TxCustomTokenPrivacy is an advance format of TxCustomToken
// so that user need to spend a lot fee to create this class tx
type TxCustomTokenPrivacy struct {
	Tx                                    // inherit from normal tx of constant(supporting privacy) with a high fee to ensure that tx could contain a big data of privacy for token
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
	fee uint64,
	tokenParams *CustomTokenPrivacyParamTx,
	//listCustomTokens map[common.Hash]TxCustomTokenPrivacy,
	db database.DatabaseInterface,
	hasPrivacyConst bool,
	shardID byte,
	//listCustomTokenCrossShard map[common.Hash]bool,
) *TransactionError {
	var err error
	// init data for tx constant for fee
	normalTx := Tx{}
	err = normalTx.Init(senderKey,
		paymentInfo,
		inputCoin,
		fee,
		hasPrivacyConst,
		db,
		nil,
		nil)
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
			//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
			newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), shardID))
			fmt.Println("INIT Tx Custom Token Privacy/ newHashInitToken", newHashInitToken)
			// validate PropertyID is the only one
			//for customTokenID := range listCustomTokens {
			//	if newHashInitToken.String() == customTokenID.String() {
			//		fmt.Println("INIT Tx Custom Token Privacy/ Existed", customTokenID, customTokenID.String() == newHashInitToken.String())
			//		return NewTransactionErr(UnexpectedErr, errors.New("this token is existed in network"))
			//	}
			//}
			existed := db.PrivacyCustomTokenIDExisted(&newHashInitToken)
			if existed {
				Logger.log.Error("INIT Tx Custom Token Privacy is Existed", newHashInitToken)
				return NewTransactionErr(UnexpectedErr, errors.New("this token is existed in network"))
			}
			//for key, _ := range listCustomTokenCrossShard {
			//	if newHashInitToken.String() == key.String() {
			//		fmt.Println("INIT Tx Custom Token Privacy/ Existed", key, key.String() == newHashInitToken.String())
			//		return NewTransactionErr(UnexpectedErr, errors.New("this token is existed in network via cross shard"))
			//	}
			//}
			existed = db.PrivacyCustomTokenIDCrossShardExisted(&newHashInitToken)
			if existed {
				Logger.log.Error("INIT Tx Custom Token Privacy is Existed(crossshard)", newHashInitToken)
				return NewTransactionErr(UnexpectedErr, errors.New("this token is existed in network via cross shard"))
			}
			txCustomToken.TxTokenPrivacyData.PropertyID = newHashInitToken
			Logger.log.Infof("A new token privacy wil be issued with ID: %+v", txCustomToken.TxTokenPrivacyData.PropertyID.String())
		}
	case CustomTokenTransfer:
		{
			handled = true
			// make a transfering for privacy custom token
			// fee always 0 and reuse function of normal tx for custom token ID
			temp := Tx{}
			propertyID, _ := common.Hash{}.NewHashFromStr(tokenParams.PropertyID)
			//if _, ok := listCustomTokens[*propertyID]; !ok {
			//	if _, ok := listCustomTokenCrossShard[*propertyID]; !ok {
			//		return NewTransactionErr(UnexpectedErr, errors.New("invalid Token ID"))
			//	}
			//}
			existed := db.PrivacyCustomTokenIDExisted(propertyID)
			existedCross := db.PrivacyCustomTokenIDCrossShardExisted(propertyID)
			if !existed && !existedCross {
				return NewTransactionErr(UnexpectedErr, errors.New("invalid Token ID"))
			}
			Logger.log.Infof("Token %+v wil be transfered with", propertyID)
			txCustomToken.TxTokenPrivacyData = TxTokenPrivacyData{
				Type:           tokenParams.TokenTxType,
				PropertyName:   tokenParams.PropertyName,
				PropertySymbol: tokenParams.PropertySymbol,
				PropertyID:     *propertyID,
			}
			err := temp.Init(senderKey,
				tokenParams.Receiver,
				tokenParams.TokenInput,
				0,
				true,
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
	poolSerialNumbers := mr.GetSerialNumbers()
	err := tx.validateDoubleSpendTxWithCurrentMempool(poolSerialNumbers)
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
	hasPrivacy bool,
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever,
	shardID byte,
) bool {
	if customTokenTx.TxTokenPrivacyData.Type == CustomTokenInit {
		return true
	}
	constantTokenID := &common.Hash{}
	constantTokenID.SetBytes(common.ConstantID[:])
	ok := customTokenTx.ValidateTransaction(hasPrivacy, db, shardID, constantTokenID)
	if !ok {
		return false
	}

	if customTokenTx.Metadata != nil {
		return customTokenTx.Metadata.ValidateMetadataByItself()
	}
	return true
}

func (customTokenTx *TxCustomTokenPrivacy) ValidateTransaction(hasPrivacy bool, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) bool {
	if customTokenTx.Tx.ValidateTransaction(hasPrivacy, db, shardID, tokenID) {
		if customTokenTx.TxTokenPrivacyData.Type == CustomTokenInit {
			return customTokenTx.TxTokenPrivacyData.TxNormal.ValidateTransaction(false, db, shardID, &customTokenTx.TxTokenPrivacyData.PropertyID)
		} else {
			return customTokenTx.TxTokenPrivacyData.TxNormal.ValidateTransaction(true, db, shardID, &customTokenTx.TxTokenPrivacyData.PropertyID)
		}
	}
	return false
}

func (tx *TxCustomTokenPrivacy) GetProof() *zkp.PaymentProof {
	return tx.Proof
}

func (tx *TxCustomTokenPrivacy) VerifyMinerCreatedTxBeforeGettingInBlock(
	insts [][]string,
	instsUsed []int,
	shardID byte,
	bcr metadata.BlockchainRetriever,
	accumulatedData *component.UsedInstData,
) (bool, error) {
	return true, nil
}
