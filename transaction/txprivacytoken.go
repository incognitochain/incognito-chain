package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/key"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

type TxPrivacyTokenInitParams struct {
	senderKey          *privacy.PrivateKey
	paymentInfo        []*privacy.PaymentInfo
	inputCoin          []coin.PlainCoin
	feeNativeCoin      uint64
	tokenParams        *CustomTokenPrivacyParamTx
	transactionStateDB *statedb.StateDB
	bridgeStateDB      *statedb.StateDB
	metaData           metadata.Metadata
	hasPrivacyCoin     bool
	hasPrivacyToken    bool
	shardID            byte
	info               []byte
}

func NewTxPrivacyTokenInitParams(senderKey *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoin []coin.PlainCoin,
	feeNativeCoin uint64,
	tokenParams *CustomTokenPrivacyParamTx,
	transactionStateDB *statedb.StateDB,
	metaData metadata.Metadata,
	hasPrivacyCoin bool,
	hasPrivacyToken bool,
	shardID byte,
	info []byte,
	bridgeStateDB *statedb.StateDB) *TxPrivacyTokenInitParams {
	params := &TxPrivacyTokenInitParams{
		shardID:            shardID,
		paymentInfo:        paymentInfo,
		metaData:           metaData,
		transactionStateDB: transactionStateDB,
		bridgeStateDB:      bridgeStateDB,
		feeNativeCoin:      feeNativeCoin,
		hasPrivacyCoin:     hasPrivacyCoin,
		hasPrivacyToken:    hasPrivacyToken,
		inputCoin:          inputCoin,
		senderKey:          senderKey,
		tokenParams:        tokenParams,
		info:               info,
	}
	return params
}

type TxCustomTokenPrivacy struct {
	TxBase                                // inherit from normal tx of P(supporting privacy) with a high fee to ensure that tx could contain a big data of privacy for token
	TxPrivacyTokenData TxPrivacyTokenData `json:"TxTokenPrivacyData"` // supporting privacy format
	// private field, not use for json parser, only use as temp variable
	cachedHash *common.Hash // cached hash data of tx
}

// ========== CHECK FUNCTION ===========

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) CheckAuthorizedSender([]byte) (bool, error) {
	return false, errors.New("TxCustomTokenPrivacy does not has CheckAuthorizedSender")
}

// =================== PARSING JSON FUNCTIONS ===================

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) UnmarshalJSON(data []byte) error {
	tx := TxBase{}
	err := json.Unmarshal(data, &tx)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenPRVJsonError, err)
	}
	temp := &struct {
		TxTokenPrivacyData TxPrivacyTokenData
	}{}
	err = json.Unmarshal(data, &temp)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenJsonError, err)
	}
	TxTokenPrivacyDataJson, err := json.MarshalIndent(temp.TxTokenPrivacyData, "", "\t")
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(UnexpectedError, err)
	}
	err = json.Unmarshal(TxTokenPrivacyDataJson, &txCustomTokenPrivacy.TxPrivacyTokenData)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenJsonError, err)
	}
	txCustomTokenPrivacy.TxBase = tx

	// TODO: hotfix, remove when fixed this issue
	if tx.Metadata != nil && tx.Metadata.GetType() == 81 {
		if txCustomTokenPrivacy.TxPrivacyTokenData.Amount == 37772966455153490 {
			txCustomTokenPrivacy.TxPrivacyTokenData.Amount = 37772966455153487
		}
	}

	return nil
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) String() string {
	// get hash of tx
	record := txCustomTokenPrivacy.TxBase.Hash().String()
	// add more hash of tx custom token data privacy
	tokenPrivacyDataHash, _ := txCustomTokenPrivacy.TxPrivacyTokenData.Hash()
	record += tokenPrivacyDataHash.String()
	if txCustomTokenPrivacy.Metadata != nil {
		record += string(txCustomTokenPrivacy.Metadata.Hash()[:])
	}
	return record
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) JSONString() string {
	data, err := json.MarshalIndent(txCustomTokenPrivacy, "", "\t")
	if err != nil {
		Logger.Log.Error(err)
		return ""
	}
	return string(data)
}

// =================== FUNCTIONS THAT GET STUFF ===================

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
func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetTxActualSize() uint64 {
	normalTxSize := txCustomTokenPrivacy.TxBase.GetTxActualSize()
	tokenDataSize := uint64(0)
	tokenDataSize += txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(txCustomTokenPrivacy.TxPrivacyTokenData.PropertyName))
	tokenDataSize += uint64(len(txCustomTokenPrivacy.TxPrivacyTokenData.PropertySymbol))
	tokenDataSize += uint64(len(txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID))
	tokenDataSize += 4 // for TxPrivacyTokenData.Type
	tokenDataSize += 8 // for TxPrivacyTokenData.Amount
	meta := txCustomTokenPrivacy.Metadata
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return normalTxSize + uint64(math.Ceil(float64(tokenDataSize)/1024))
}

func (tx TxCustomTokenPrivacy) GetTxPrivacyTokenActualSize() uint64 {
	tokenDataSize := uint64(0)
	tokenDataSize += tx.TxPrivacyTokenData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(tx.TxPrivacyTokenData.PropertyName))
	tokenDataSize += uint64(len(tx.TxPrivacyTokenData.PropertySymbol))
	tokenDataSize += uint64(len(tx.TxPrivacyTokenData.PropertyID))
	tokenDataSize += 4 // for TxPrivacyTokenData.Type
	tokenDataSize += 8 // for TxPrivacyTokenData.Amount

	meta := tx.TxPrivacyTokenData.TxNormal.Metadata
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return uint64(math.Ceil(float64(tokenDataSize) / 1024))
}

// Get SigPubKey of ptoken
func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetSigPubKey() []byte {
	return txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.SigPubKey
}

// GetTxFeeToken - return Token Fee use to pay for privacy token Tx
func (tx TxCustomTokenPrivacy) GetTxFeeToken() uint64 {
	return tx.TxPrivacyTokenData.TxNormal.Fee
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetTokenID() *common.Hash {
	return &txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	pubkeys, amounts := txCustomTokenPrivacy.GetReceivers()
	if len(pubkeys) == 0 {
		Logger.Log.Error("GetTransferData receive 0 output, it should has exactly 1 output")
		return false, nil, 0, &common.PRVCoinID
	}
	if len(pubkeys) > 1 {
		Logger.Log.Error("GetTransferData receiver: More than 1 receiver")
		return false, nil, 0, &txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID
	}
	return true, pubkeys[0], amounts[0], &txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID
}

// CalculateBurnAmount - get tx value for pToken
func (txCustomTokenPrivacy TxCustomTokenPrivacy) CalculateTxValue() uint64 {
	if txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof == nil {
		return 0
	}
	proof := txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof
	if proof.GetOutputCoins() == nil || len(proof.GetOutputCoins()) == 0 {
		return 0
	}
	if proof.GetInputCoins() == nil || len(proof.GetInputCoins()) == 0 { // coinbase tx
		txValue := uint64(0)
		for _, outCoin := range proof.GetOutputCoins() {
			txValue += outCoin.GetValue()
		}
		return txValue
	}

	if txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.IsPrivacy() {
		return 0
	}

	senderPKBytes := proof.GetInputCoins()[0].GetPublicKey().ToBytesS()
	txValue := uint64(0)
	for _, outCoin := range proof.GetOutputCoins() {
		outPKBytes := outCoin.GetPublicKey().ToBytesS()
		if bytes.Equal(senderPKBytes, outPKBytes) {
			continue
		}
		txValue += outCoin.GetValue()
	}
	return txValue
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) ListSerialNumbersHashH() []common.Hash {
	tx := txCustomTokenPrivacy.TxBase
	result := []common.Hash{}
	if tx.Proof != nil {
		for _, d := range tx.Proof.GetInputCoins() {
			hash := common.HashH(d.GetKeyImage().ToBytesS())
			result = append(result, hash)
		}
	}
	customTokenPrivacy := txCustomTokenPrivacy.TxPrivacyTokenData
	if customTokenPrivacy.TxNormal.Proof != nil {
		for _, d := range customTokenPrivacy.TxNormal.Proof.GetInputCoins() {
			hash := common.HashH(d.GetKeyImage().ToBytesS())
			result = append(result, hash)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result
}

// GetTxFee - return fee PRV of Tx which contain privacy token Tx
func (tx TxCustomTokenPrivacy) GetTxFee() uint64 {
	return tx.TxBase.GetTxFee()
}

// ========== NORMAL INIT FUNCTIONS ==========

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) Init(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*TxPrivacyTokenInitParams)
	if !ok {
		return errors.New("Cannot init TxCustomTokenPrivacy because params is not correct")
	}
	// init data for tx PRV for fee
	txPrivacyParams := NewTxPrivacyInitParams(
		params.senderKey,
		params.paymentInfo,
		params.inputCoin,
		params.feeNativeCoin,
		params.hasPrivacyCoin,
		params.transactionStateDB,
		nil,
		params.metaData,
		params.info,
	)
	normalTx, err := NewTxPrivacyFromParams(txPrivacyParams)
	if err != nil {
		Logger.Log.Errorf("Cannot create tx from params, error %v", err)
		return NewTransactionErr(PrivacyTokenInitFeeParamsError, err)
	}
	if err = normalTx.Init(txPrivacyParams); err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	normalTx.SetType(common.TxCustomTokenPrivacyType)
	txCustomTokenPrivacy.TxBase = NewTxBaseFromMetadataTx(normalTx)

	// check tx size
	limitFee := uint64(0)
	estimateTxSizeParam := NewEstimateTxSizeParam(len(params.inputCoin), len(params.paymentInfo),
		params.hasPrivacyCoin, nil, params.tokenParams, limitFee)
	if txSize := EstimateTxSize(estimateTxSizeParam); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	// check action type and create privacy custom toke data
	var handled = false
	// Add token data component
	switch params.tokenParams.TokenTxType {
	case CustomTokenInit:
		// case init a new privacy custom token
		{
			handled = true
			txCustomTokenPrivacy.TxPrivacyTokenData = TxPrivacyTokenData{
				Type:           params.tokenParams.TokenTxType,
				PropertyName:   params.tokenParams.PropertyName,
				PropertySymbol: params.tokenParams.PropertySymbol,
				Amount:         params.tokenParams.Amount,
			}

			temp := TxVersion2{}
			temp.Version = txVersion2Number
			temp.Type = common.TxNormalType

			// Amount, Randomness, SharedRandom is transparency until we call concealData
			message := []byte{}

			// Set Info
			if len(params.tokenParams.Receiver[0].Message) > 0 {
				if len(params.tokenParams.Receiver[0].Message) > coin.MaxSizeInfoCoin {
					return NewTransactionErr(ExceedSizeInfoOutCoinError, nil)
				}
				message = params.tokenParams.Receiver[0].Message
			}
			tempPaymentInfo := key.InitPaymentInfo(params.tokenParams.Receiver[0].PaymentAddress, params.tokenParams.Amount, message)
			c, errCoin := coin.NewCoinFromPaymentInfo(tempPaymentInfo)
			if errCoin != nil {
				Logger.Log.Errorf("Cannot create new coin based on payment info err %v", errCoin)
				return errCoin
			}
			tempOutputCoin := make([]coin.Coin, 1)
			tempOutputCoin[0] = c
			proof := new(privacy.ProofV2)
			proof.Init()
			if err = proof.SetOutputCoins(tempOutputCoin); err != nil {
				Logger.Log.Errorf("Init customPrivacyToken cannot set outputCoins")
				return err
			}

			temp.Proof = proof
			if temp.PubKeyLastByteSender, err = params.inputCoin[0].GetShardID(); err != nil {
				return NewTransactionErr(GetShardIDByPublicKeyError, err)
			}

			// sign Tx
			temp.sigPrivKey = *params.senderKey
			temp.Sig, temp.SigPubKey, err = signNoPrivacy(params.senderKey, temp.Hash()[:])
			if err != nil {
				Logger.Log.Error(errors.New("can't sign this tx"))
				return NewTransactionErr(SignTxError, err)
			}
			temp.SigPubKey = params.tokenParams.Receiver[0].PaymentAddress.Pk

			txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal = NewTxBaseFromMetadataTx(&temp)
			hashInitToken, err := txCustomTokenPrivacy.TxPrivacyTokenData.Hash()
			if err != nil {
				Logger.Log.Error(errors.New("can't hash this token data"))
				return NewTransactionErr(UnexpectedError, err)
			}

			if params.tokenParams.Mintable {
				propertyID, err := common.Hash{}.NewHashFromStr(params.tokenParams.PropertyID)
				if err != nil {
					return NewTransactionErr(TokenIDInvalidError, err, propertyID.String())
				}
				txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID = *propertyID
				txCustomTokenPrivacy.TxPrivacyTokenData.Mintable = true
			} else {
				//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
				newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), params.shardID))
				Logger.Log.Debug("New Privacy Token %+v ", newHashInitToken)
				existed := txDatabaseWrapper.privacyTokenIDExisted(params.transactionStateDB, newHashInitToken)
				if existed {
					Logger.Log.Error("INIT Tx Custom Token Privacy is Existed", newHashInitToken)
					return NewTransactionErr(TokenIDExistedError, errors.New("this token is existed in network"))
				}
				txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID = newHashInitToken
				Logger.Log.Debugf("A new token privacy wil be issued with ID: %+v", txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID.String())
			}
		}
	case CustomTokenTransfer:
		{
			handled = true
			// make a transfering for privacy custom token
			// fee always 0 and reuse function of normal tx for custom token ID
			propertyID, _ := common.Hash{}.NewHashFromStr(params.tokenParams.PropertyID)
			existed := txDatabaseWrapper.privacyTokenIDExisted(params.transactionStateDB, *propertyID)
			if !existed {
				isBridgeToken := false
				allBridgeTokensBytes, err := txDatabaseWrapper.getAllBridgeTokens(params.bridgeStateDB)
				if err != nil {
					return NewTransactionErr(TokenIDExistedError, err)
				}
				if len(allBridgeTokensBytes) > 0 {
					var allBridgeTokens []*rawdbv2.BridgeTokenInfo
					err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
					if err != nil {
						return NewTransactionErr(TokenIDExistedError, err)
					}
					for _, bridgeTokens := range allBridgeTokens {
						if propertyID.IsEqual(bridgeTokens.TokenID) {
							isBridgeToken = true
							break
						}
					}
				}
				if !isBridgeToken {
					return NewTransactionErr(TokenIDExistedError, errors.New("invalid Token ID"))
				}
			}
			Logger.Log.Debugf("Token %+v wil be transfered with", propertyID)
			txCustomTokenPrivacy.TxPrivacyTokenData = TxPrivacyTokenData{
				Type:           params.tokenParams.TokenTxType,
				PropertyName:   params.tokenParams.PropertyName,
				PropertySymbol: params.tokenParams.PropertySymbol,
				PropertyID:     *propertyID,
				Mintable:       params.tokenParams.Mintable,
			}
			params := NewTxPrivacyInitParams(
				params.senderKey,
				params.tokenParams.Receiver,
				params.tokenParams.TokenInput,
				params.tokenParams.Fee,
				params.hasPrivacyToken,
				params.transactionStateDB,
				propertyID,
				nil,
				nil,
			)
			tx, err := NewTxPrivacyFromParams(params)
			if err != nil {
				Logger.Log.Errorf("Cannot init NewTxPrivacyInitParams: params has error %v", err)
				return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
			}
			if err = tx.Init(params); err != nil {
				return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
			}
			txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal = NewTxBaseFromMetadataTx(tx)
		}
	}
	if !handled {
		return NewTransactionErr(PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	}
	return nil
}

func estimateTxSizeOfInitTokenSalary(publicKey []byte, amount uint64, coinName string, coinID *common.Hash) uint64 {
	receiver := &privacy.PaymentInfo{
		Amount:         amount,
		PaymentAddress: privacy.PaymentAddress{
			Pk: publicKey,
			Tk: []byte{},
		},
	}
	var propertyID [common.HashSize]byte
	copy(propertyID[:], coinID[:])
	propID := common.Hash(propertyID)
	tokenParams := &CustomTokenPrivacyParamTx{
		PropertyID:     propID.String(),
		PropertyName:   coinName,
		PropertySymbol: coinName,
		Amount:         amount,
		TokenTxType:    CustomTokenInit,
		Receiver:       []*privacy.PaymentInfo{receiver},
		TokenInput:     []coin.PlainCoin{},
		Mintable:       true,
	}
	estimateTxSizeParam := NewEstimateTxSizeParam(0,0,false, nil, tokenParams, uint64(0))
	return EstimateTxSize(estimateTxSizeParam)
}
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) InitTxTokenSalary(otaCoin *coin.CoinV2, privKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata, coinID *common.Hash, coinName string) error {
	// init data for tx PRV for fee
	txPrivacyParams := NewTxPrivacyInitParams(
		privKey, []*privacy.PaymentInfo{}, nil, 0, false, stateDB, nil, metaData, nil,
	)
	normalTx, err := NewTxPrivacyFromParams(txPrivacyParams)
	if err != nil {
		Logger.Log.Errorf("Cannot create tx from params, error %v", err)
		return NewTransactionErr(PrivacyTokenInitFeeParamsError, err)
	}
	if err = normalTx.Init(txPrivacyParams); err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	normalTx.SetType(common.TxCustomTokenPrivacyType)
	txCustomTokenPrivacy.TxBase = NewTxBaseFromMetadataTx(normalTx)
	// check tx size
	publicKeyBytes := otaCoin.GetPublicKey().ToBytesS()
	if txSize := estimateTxSizeOfInitTokenSalary(publicKeyBytes, otaCoin.GetValue(), coinName, coinID); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}
	// check action type and create privacy custom toke data
	txCustomTokenPrivacy.TxPrivacyTokenData = TxPrivacyTokenData{
		Type:           CustomTokenInit,
		PropertyName:   coinName,
		PropertySymbol: coinName,
		Amount:         otaCoin.GetValue(),
	}
	tempOutputCoin := []coin.Coin{otaCoin}
	proof := new(privacy.ProofV2)
	proof.Init()
	if err = proof.SetOutputCoins(tempOutputCoin); err != nil {
		Logger.Log.Errorf("Init customPrivacyToken cannot set outputCoins")
		return err
	}
	temp := TxVersion2{}
	temp.Version = txVersion2Number
	temp.Type = common.TxNormalType
	temp.Proof = proof
	temp.PubKeyLastByteSender = publicKeyBytes[len(publicKeyBytes)-1]
	// sign Tx
	temp.sigPrivKey = *privKey
	if temp.Sig, _, err = signNoPrivacy(privKey, temp.Hash()[:]); err != nil {
		Logger.Log.Error(errors.New("can't sign this tx"))
		return NewTransactionErr(SignTxError, err)
	}
	temp.SigPubKey = otaCoin.GetPublicKey().ToBytesS()
	var propertyID [common.HashSize]byte
	copy(propertyID[:], coinID[:])
	txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID = propertyID
	txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal = NewTxBaseFromMetadataTx(&temp)
	txCustomTokenPrivacy.TxPrivacyTokenData.Mintable = true
	return nil
}

// =================== FUNCTION THAT CHECK STUFFS  ===================

// IsCoinsBurning - checking this is a burning pToken
func (txCustomTokenPrivacy TxCustomTokenPrivacy) IsCoinsBurning(bcr metadata.ChainRetriever, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) bool {
	// get proof of pToken
	proof := txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof
	if proof == nil || len(proof.GetOutputCoins()) == 0 {
		return false
	}
	return txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.IsCoinsBurning(bcr, retriever, viewRetriever, beaconHeight)
	////  validate receiver with burning address
	//senderPKBytes := []byte{}
	//if len(proof.GetInputCoins()) > 0 {
	//	senderPKBytes = proof.GetInputCoins()[0].GetPublicKey().ToBytesS()
	//}
	//
	////get burning address
	//burningAddress := bcr.GetBurningAddress(beaconHeight)
	//keyWalletBurningAccount, err := wallet.Base58CheckDeserialize(burningAddress)
	//if err != nil {
	//	Logger.Log.Errorf("Can not deserialize burn address: %v\n", burningAddress)
	//	return false
	//}
	//
	//keysetBurningAccount := keyWalletBurningAccount.KeySet
	//paymentAddressBurningAccount := keysetBurningAccount.PaymentAddress
	//for _, outCoin := range proof.GetOutputCoins() {
	//	outPKBytes := outCoin.GetPublicKey().ToBytesS()
	//	if !bytes.Equal(senderPKBytes, outPKBytes) && !bytes.Equal(outPKBytes, paymentAddressBurningAccount.Pk[:]) {
	//		return false
	//	}
	//}
	//return true
}

// ========== VALIDATE FUNCTIONS ===========

func (txCustomTokenPrivacy TxCustomTokenPrivacy) ValidateType() bool {
	return txCustomTokenPrivacy.Type == common.TxCustomTokenPrivacyType
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	err := txCustomTokenPrivacy.validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(DoubleSpendError, err)
	}
	// TODO: will move this to mempool process
	if txCustomTokenPrivacy.TxPrivacyTokenData.Type == CustomTokenInit && txCustomTokenPrivacy.GetMetadata() == nil {
		initTokenID := txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID
		txsInMem := mr.GetTxsInMem()
		for _, tx := range txsInMem {
			// try parse to TxCustomTokenPrivacy
			var privacyTokenTx, ok = tx.Tx.(*TxCustomTokenPrivacy)
			if ok && privacyTokenTx.TxPrivacyTokenData.Type == CustomTokenInit && privacyTokenTx.GetMetadata() == nil {
				// check > 1 tx init token by the same token ID
				if privacyTokenTx.TxPrivacyTokenData.PropertyID.IsEqual(&initTokenID) {
					return NewTransactionErr(TokenIDInvalidError, fmt.Errorf("had already tx for initing token ID %s in pool", privacyTokenTx.TxPrivacyTokenData.PropertyID.String()), privacyTokenTx.TxPrivacyTokenData.PropertyID.String())
				}
			}
		}
	}

	return nil
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH map[common.Hash][]common.Hash) error {
	// check proof of PRV and pToken
	if txCustomTokenPrivacy.Proof == nil && txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof == nil {
		return errors.New("empty tx")
	}

	// collect serial number for PRV
	temp := make(map[common.Hash]interface{})
	if txCustomTokenPrivacy.Proof != nil {
		for _, desc := range txCustomTokenPrivacy.Proof.GetInputCoins() {
			hash := common.HashH(desc.GetKeyImage().ToBytesS())
			temp[hash] = nil
		}
	}
	// collect serial number for pToken
	if txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof != nil {
		for _, desc := range txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.Proof.GetInputCoins() {
			hash := common.HashH(desc.GetKeyImage().ToBytesS())
			temp[hash] = nil
		}
	}

	// check with pool serial number in mempool
	for _, listSerialNumbers := range poolSerialNumbersHashH {
		for _, serialNumberHash := range listSerialNumbers {
			if _, ok := temp[serialNumberHash]; ok {
				return errors.New("double spend")
			}
		}
	}
	return nil
}

func (txCustomTokenPrivacy TxCustomTokenPrivacy) ValidateTxWithBlockChain(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error {
	err := txCustomTokenPrivacy.ValidateDoubleSpendWithBlockchain(shardID, stateDB, nil)
	if err != nil {
		return NewTransactionErr(InvalidDoubleSpendPRVError, err)
	}
	err = txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.ValidateDoubleSpendWithBlockchain(shardID, stateDB, txCustomTokenPrivacy.GetTokenID())
	if err != nil {
		return NewTransactionErr(InvalidDoubleSpendPrivacyTokenError, err)
	}
	return nil
}

// ValidateSanityData - validate sanity data of PRV and pToken
func (txCustomTokenPrivacy TxCustomTokenPrivacy) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	meta := txCustomTokenPrivacy.TxBase.Metadata
	if meta != nil {
		isContinued, ok, err := meta.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, &txCustomTokenPrivacy)
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}

	// validate sanity data for PRV
	//result, err := txCustomTokenPrivacy.Tx.validateNormalTxSanityData()
	result, err := txCustomTokenPrivacy.TxBase.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if err != nil {
		return result, NewTransactionErr(InvalidSanityDataPRVError, err)
	}
	// validate sanity for pToken

	//result, err = txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.validateNormalTxSanityData()
	result, err = txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if err != nil {
		return result, NewTransactionErr(InvalidSanityDataPrivacyTokenError, err)
	}
	return result, nil
}

// ValidateTxByItself - validate tx by itself, check signature, proof,... and metadata
func (txCustomTokenPrivacy TxCustomTokenPrivacy) ValidateTxByItself(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, chainRetriever metadata.ChainRetriever, shardID byte, isNewTransaction bool, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever) (bool, error) {
	// no need to check for tx init token
	if txCustomTokenPrivacy.TxPrivacyTokenData.Type == CustomTokenInit {
		return txCustomTokenPrivacy.TxBase.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, false, isNewTransaction)
	}
	// check for proof, signature ...
	if ok, err := txCustomTokenPrivacy.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, false, isNewTransaction); !ok {
		return false, err
	}
	// check for metadata
	if txCustomTokenPrivacy.Metadata != nil {
		validateMetadata := txCustomTokenPrivacy.Metadata.ValidateMetadataByItself()
		if !validateMetadata {
			return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid"))
		}
		return validateMetadata, nil
	}
	return true, nil
}

func (txCustomTokenPrivacy *TxCustomTokenPrivacy) ValidateTransaction(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	// validate for PRV
	ok, err := txCustomTokenPrivacy.TxBase.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, isBatch, isNewTransaction)
	if ok {
		// validate for pToken
		tokenID := txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID
		if txCustomTokenPrivacy.TxPrivacyTokenData.Type == CustomTokenInit {
			if txCustomTokenPrivacy.Type == common.TxRewardType && txCustomTokenPrivacy.TxPrivacyTokenData.Mintable {
				isBridgeCentralizedToken, _ := txDatabaseWrapper.isBridgeTokenExistedByType(bridgeStateDB, tokenID, true)
				isBridgeDecentralizedToken, _ := txDatabaseWrapper.isBridgeTokenExistedByType(bridgeStateDB, tokenID, false)
				if isBridgeCentralizedToken || isBridgeDecentralizedToken {
					return true, nil
				}
				return false, nil
			} else {
				// check exist token
				if txDatabaseWrapper.privacyTokenIDExisted(transactionStateDB, tokenID) {
					return false, nil
				}
				return true, nil
			}
		} else {
			if err != nil {
				Logger.Log.Errorf("Cannot create txPrivacyFromVersionNumber from TxPrivacyTokenData, err %v", err)
				return false, err
			}
			return txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.ValidateTransaction(
				txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal.IsPrivacy(),
				transactionStateDB, bridgeStateDB, shardID, &tokenID, isBatch, isNewTransaction)
		}
	}
	return false, err
}

// GetProof - return proof PRV of tx
func (txCustomTokenPrivacy TxCustomTokenPrivacy) GetProof() privacy.Proof {
	return txCustomTokenPrivacy.Proof
}

// VerifyMinerCreatedTxBeforeGettingInBlock
func (txCustomTokenPrivacy TxCustomTokenPrivacy) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []metadata.Transaction, txsUsed []int, insts [][]string, instsUsed []int, shardID byte, bcr metadata.ChainRetriever, accumulatedValues *metadata.AccumulatedValues, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever) (bool, error) {
	if !txCustomTokenPrivacy.TxPrivacyTokenData.Mintable {
		return true, nil
	}
	meta := txCustomTokenPrivacy.Metadata
	if meta == nil {
		Logger.Log.Error("Mintable custom token must contain metadata")
		return false, nil
	}
	if !meta.IsMinerCreatedMetaType() {
		return false, nil
	}
	return meta.VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock, txsUsed, insts, instsUsed, shardID, &txCustomTokenPrivacy, bcr, accumulatedValues, nil, nil)
}

type TxPrivacyTokenInitParamsForASM struct {
	//senderKey       *privacy.PrivateKey
	//paymentInfo     []*privacy.PaymentInfo
	//inputCoin       []*coin.PlainCoinV1
	//feeNativeCoin   uint64
	//tokenParams     *CustomTokenPrivacyParamTx
	//transactionStateDB              database.DatabaseInterface
	//metaData        metadata.Metadata
	//hasPrivacyCoin  bool
	//hasPrivacyToken bool
	//shardID         byte
	//info            []byte

	txParam                           TxPrivacyTokenInitParams
	commitmentIndicesForNativeToken   []uint64
	commitmentBytesForNativeToken     [][]byte
	myCommitmentIndicesForNativeToken []uint64
	sndOutputsForNativeToken          []*privacy.Scalar

	commitmentIndicesForPToken   []uint64
	commitmentBytesForPToken     [][]byte
	myCommitmentIndicesForPToken []uint64
	sndOutputsForPToken          []*privacy.Scalar
}

func (param *TxPrivacyTokenInitParamsForASM) SetMetaData(meta metadata.Metadata) {
	param.txParam.metaData = meta
}

func NewTxPrivacyTokenInitParamsForASM(
	senderKey *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoin []coin.PlainCoin,
	feeNativeCoin uint64,
	tokenParams *CustomTokenPrivacyParamTx,
	metaData metadata.Metadata,
	hasPrivacyCoin bool,
	hasPrivacyToken bool,
	shardID byte,
	info []byte,
	commitmentIndicesForNativeToken []uint64,
	commitmentBytesForNativeToken [][]byte,
	myCommitmentIndicesForNativeToken []uint64,
	sndOutputsForNativeToken []*privacy.Scalar,

	commitmentIndicesForPToken []uint64,
	commitmentBytesForPToken [][]byte,
	myCommitmentIndicesForPToken []uint64,
	sndOutputsForPToken []*privacy.Scalar) *TxPrivacyTokenInitParamsForASM {

	txParam := NewTxPrivacyTokenInitParams(senderKey, paymentInfo, inputCoin, feeNativeCoin, tokenParams, nil, metaData, hasPrivacyCoin, hasPrivacyToken, shardID, info, nil)
	params := &TxPrivacyTokenInitParamsForASM{
		txParam:                           *txParam,
		commitmentIndicesForNativeToken:   commitmentIndicesForNativeToken,
		commitmentBytesForNativeToken:     commitmentBytesForNativeToken,
		myCommitmentIndicesForNativeToken: myCommitmentIndicesForNativeToken,
		sndOutputsForNativeToken:          sndOutputsForNativeToken,

		commitmentIndicesForPToken:   commitmentIndicesForPToken,
		commitmentBytesForPToken:     commitmentBytesForPToken,
		myCommitmentIndicesForPToken: myCommitmentIndicesForPToken,
		sndOutputsForPToken:          sndOutputsForPToken,
	}
	return params
}

// TODO Privacy, WILL DO THIS LATER BECAUSE IT IS ASM
// Init -  build normal tx component and privacy custom token data
func (txCustomTokenPrivacy *TxCustomTokenPrivacy) InitForASM(params *TxPrivacyTokenInitParamsForASM, serverTime int64) error {
	//var err error
	//initTokenParamsASM := NewTxPrivacyInitParamsForASM(
	//	params.txParam.senderKey,
	//	params.txParam.paymentInfo,
	//	params.txParam.inputCoin,
	//	params.txParam.feeNativeCoin,
	//	params.txParam.hasPrivacyCoin,
	//	nil,
	//	params.txParam.metaData,
	//	params.txParam.info,
	//	params.commitmentIndicesForNativeToken,
	//	params.commitmentBytesForNativeToken,
	//	params.myCommitmentIndicesForNativeToken,
	//	params.sndOutputsForNativeToken,
	//)
	//normalTx :=
	//err = normalTx.InitForASM()
	//if err != nil {
	//	return NewTransactionErr(PrivacyTokenInitPRVError, err)
	//}
	//
	//// override TxCustomTokenPrivacyType type
	//normalTx.Type = common.TxCustomTokenPrivacyType
	//txCustomTokenPrivacy.Tx = normalTx
	//
	//// check action type and create privacy custom toke data
	//var handled = false
	//// Add token data component
	//switch params.txParam.tokenParams.TokenTxType {
	//case CustomTokenInit:
	//	// case init a new privacy custom token
	//	{
	//		handled = true
	//		txCustomTokenPrivacy.TxPrivacyTokenData = TxPrivacyTokenData{
	//			Type:           params.txParam.tokenParams.TokenTxType,
	//			PropertyName:   params.txParam.tokenParams.PropertyName,
	//			PropertySymbol: params.txParam.tokenParams.PropertySymbol,
	//			Amount:         params.txParam.tokenParams.Amount,
	//		}
	//
	//		// issue token with data of privacy
	//		temp := Tx{}
	//		temp.Type = common.TxNormalType
	//		temp.Proof = privacy.NewProofWithVersion(txCustomTokenPrivacy.Version)
	//		temp.Proof.Init()
	//		tempOutputCoin := make([]coin.Coin, 1)
	//		c := new(coin.CoinV1)
	//		c.CoinDetails.SetValue(params.txParam.tokenParams.Amount)
	//		PK, err := new(privacy.Point).FromBytesS(params.txParam.tokenParams.Receiver[0].PaymentAddress.Pk)
	//		if err != nil {
	//			return NewTransactionErr(DecompressPaymentAddressError, err)
	//		}
	//		c.CoinDetails.SetPublicKey(PK)
	//		c.CoinDetails.SetRandomness(privacy.RandomScalar())
	//
	//		// set info coin for output coin
	//		if len(params.txParam.tokenParams.Receiver[0].Message) > 0 {
	//			if len(params.txParam.tokenParams.Receiver[0].Message) > privacy.MaxSizeInfoCoin {
	//				return NewTransactionErr(ExceedSizeInfoOutCoinError, nil)
	//			}
	//			c.CoinDetails.SetInfo(params.txParam.tokenParams.Receiver[0].Message)
	//		}
	//		sndOut := privacy.RandomScalar()
	//		c.CoinDetails.SetSNDerivator(sndOut)
	//		err = c.CoinDetails.CommitAll()
	//		if err != nil {
	//			return NewTransactionErr(CommitOutputCoinError, err)
	//		}
	//
	//		tempOutputCoin[0] = c
	//		if err = temp.Proof.SetOutputCoins(tempOutputCoin); err != nil {
	//			Logger.Log.Errorf("TxPrivacyToken InitASM cannot set output coins: err %v", err)
	//			return err
	//		}
	//
	//		// get last byte
	//		temp.PubKeyLastByteSender = params.txParam.tokenParams.Receiver[0].PaymentAddress.Pk[len(params.txParam.tokenParams.Receiver[0].PaymentAddress.Pk)-1]
	//
	//		// sign Tx
	//		temp.SigPubKey = params.txParam.tokenParams.Receiver[0].PaymentAddress.Pk
	//		temp.sigPrivKey = *params.txParam.senderKey
	//		err = signTx(&temp)
	//		if err != nil {
	//			Logger.Log.Error(errors.New("can't sign this tx"))
	//			return NewTransactionErr(SignTxError, err)
	//		}
	//
	//		txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal = temp
	//		hashInitToken, err := txCustomTokenPrivacy.TxPrivacyTokenData.Hash()
	//		if err != nil {
	//			Logger.Log.Error(errors.New("can't hash this token data"))
	//			return NewTransactionErr(UnexpectedError, err)
	//		}
	//
	//		if params.txParam.tokenParams.Mintable {
	//			propertyID, err := common.Hash{}.NewHashFromStr(params.txParam.tokenParams.PropertyID)
	//			if err != nil {
	//				return NewTransactionErr(TokenIDInvalidError, err, propertyID.String())
	//			}
	//			txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID = *propertyID
	//			txCustomTokenPrivacy.TxPrivacyTokenData.Mintable = true
	//		} else {
	//			//NOTICE: @merman update PropertyID calculated from hash of tokendata and shardID
	//			newHashInitToken := common.HashH(append(hashInitToken.GetBytes(), params.txParam.shardID))
	//			txCustomTokenPrivacy.TxPrivacyTokenData.PropertyID = newHashInitToken
	//		}
	//	}
	//case CustomTokenTransfer:
	//	{
	//		handled = true
	//		// make a transfering for privacy custom token
	//		// fee always 0 and reuse function of normal tx for custom token ID
	//		temp := Tx{}
	//		propertyID, _ := common.Hash{}.NewHashFromStr(params.txParam.tokenParams.PropertyID)
	//		txCustomTokenPrivacy.TxPrivacyTokenData = TxPrivacyTokenData{
	//			Type:           params.txParam.tokenParams.TokenTxType,
	//			PropertyName:   params.txParam.tokenParams.PropertyName,
	//			PropertySymbol: params.txParam.tokenParams.PropertySymbol,
	//			PropertyID:     *propertyID,
	//			Mintable:       params.txParam.tokenParams.Mintable,
	//		}
	//		//err := temp.InitForASM(NewTxPrivacyInitParamsForASM(
	//		//	params.txParam.senderKey,
	//		//	params.txParam.tokenParams.Receiver,
	//		//	params.txParam.tokenParams.TokenInput,
	//		//	params.txParam.tokenParams.Fee,
	//		//	params.txParam.hasPrivacyToken,
	//		//	propertyID,
	//		//	nil,
	//		//	params.txParam.info,
	//		//	params.commitmentIndicesForPToken,
	//		//	params.commitmentBytesForPToken,
	//		//	params.myCommitmentIndicesForPToken,
	//		//	params.sndOutputsForPToken,
	//		//), serverTime)
	//		err := temp.InitForASM(NewTxPrivacyInitParamsForASM(
	//			params.txParam.senderKey,
	//			params.txParam.tokenParams.Receiver,
	//			params.txParam.tokenParams.TokenInput,
	//			params.txParam.tokenParams.Fee,
	//			params.txParam.hasPrivacyToken,
	//			propertyID,
	//			nil,
	//			params.txParam.info,
	//			params.commitmentIndicesForPToken,
	//			params.commitmentBytesForPToken,
	//			params.myCommitmentIndicesForPToken,
	//			params.sndOutputsForPToken,
	//		))
	//		if err != nil {
	//			return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
	//		}
	//		txCustomTokenPrivacy.TxPrivacyTokenData.TxNormal = temp
	//	}
	//}
	//
	//if !handled {
	//	return NewTransactionErr(PrivacyTokenTxTypeNotHandleError, errors.New("can't handle this TokenTxType"))
	//}
	return nil
}
