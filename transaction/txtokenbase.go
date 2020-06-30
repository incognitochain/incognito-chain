// Because txprivacytoken version 1 had a bug
// txprivacytoken in later version will not use the same base with txtokenversion1
// So we duplicate some code from ver1 to ver2 and not use any embedding

package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
)

func NewTransactionTokenFromParams(params *TxPrivacyTokenInitParams) (TxTokenInterface, error) {
	inputCoins := params.inputCoin
	ver, err := getTxVersionFromCoins(inputCoins)
	if err != nil {
		return nil, err
	}

	if ver == 1 {
		return new(TxTokenVersion1), nil
	} else if ver == 2 {
		return new(TxTokenVersion2), nil
	}
	return nil, errors.New("Something is wrong when NewTransactionFromParams")
}

type Tx = metadata.Transaction

type TxTokenBase struct {
	Tx
	TxTokenData TxTokenData `json:"TxTokenPrivacyData"`
}

func GetTxTokenDataFromTransaction(tx metadata.Transaction) *TxTokenData {
	if tx.GetType() != common.TxCustomTokenPrivacyType && tx.GetType() != common.TxTokenConversionType {
		return nil
	}
	if tx.GetVersion() == TxVersion1Number {
		txTemp := tx.(*TxTokenVersion1)
		return &txTemp.TxTokenData
	} else if tx.GetVersion() == TxVersion2Number || tx.GetVersion() == TxConversionVersion12Number {
		txTemp := tx.(*TxTokenVersion2)
		return &txTemp.TxTokenData
	}
	return nil
}

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

// CustomTokenParamTx - use for rpc request json body
type CustomTokenPrivacyParamTx struct {
	PropertyID     string                 `json:"TokenID"`
	PropertyName   string                 `json:"TokenName"`
	PropertySymbol string                 `json:"TokenSymbol"`
	Amount         uint64                 `json:"TokenAmount"`
	TokenTxType    int                    `json:"TokenTxType"`
	Receiver       []*privacy.PaymentInfo `json:"TokenReceiver"`
	TokenInput     []coin.PlainCoin       `json:"TokenInput"`
	Mintable       bool                   `json:"TokenMintable"`
	Fee            uint64                 `json:"TokenFee"`
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

// ========== Get/Set FUNCTION ============

func (txToken TxTokenBase) GetTxBase() metadata.Transaction    { return txToken.Tx }
func (txToken *TxTokenBase) SetTxBase(tx metadata.Transaction) { txToken.Tx = tx }
func (txToken TxTokenBase) GetTxPrivacyTokenData() TxTokenData {
	return txToken.TxTokenData
}
func (txToken *TxTokenBase) SetTxPrivacyTokenData(data TxTokenData) {
	txToken.TxTokenData = data
}

// ========== CHECK FUNCTION ===========

func (txToken *TxTokenBase) CheckAuthorizedSender([]byte) (bool, error) {
	return false, errors.New("TxTokenBase does not has CheckAuthorizedSender")
}

// =================== PARSING JSON FUNCTIONS ===================

//func (txToken TxTokenBase) MarshalJSON() ([]byte, error) {
//	type TemporaryTxToken struct {
//		TxBase
//		TxPrivacyTokenData TxTokenData `json:"TxTokenPrivacyData"`
//	}
//	tempTx := TemporaryTxToken{}
//	tempTx.TxPrivacyTokenData = txToken.GetTxPrivacyTokenData()
//	tx := txToken.GetTxBase()
//	tempTx.TxBase.SetVersion(tx.GetVersion())
//	tempTx.TxBase.SetType(tx.GetType())
//	tempTx.TxBase.SetLockTime(tx.GetLockTime())
//	tempTx.TxBase.SetTxFee(tx.GetTxFee())
//	tempTx.TxBase.SetInfo(tx.GetInfo())
//	tempTx.TxBase.SetSigPubKey(tx.GetSigPubKey())
//	tempTx.TxBase.SetSig(tx.GetSig())
//	tempTx.TxBase.SetProof(tx.GetProof())
//	tempTx.TxBase.SetGetSenderAddrLastByte(tx.GetSenderAddrLastByte())
//	tempTx.TxBase.SetMetadata(tx.GetMetadata())
//	tempTx.TxBase.SetGetSenderAddrLastByte(tx.GetSenderAddrLastByte())
//
//	return json.Marshal(tempTx)
//}

func (txToken TxTokenBase) MarshalJSON() ([]byte, error) {
	type TemporaryTxToken struct {
		TxBase
		TxPrivacyTokenData TxTokenData `json:"TxTokenPrivacyData"`
	}
	tempTx := TemporaryTxToken{}
	tempTx.TxPrivacyTokenData = txToken.GetTxPrivacyTokenData()
	tx := txToken.GetTxBase()
	tempTx.TxBase.SetVersion(tx.GetVersion())
	tempTx.TxBase.SetType(tx.GetType())
	tempTx.TxBase.SetLockTime(tx.GetLockTime())
	tempTx.TxBase.SetTxFee(tx.GetTxFee())
	tempTx.TxBase.SetInfo(tx.GetInfo())
	tempTx.TxBase.SetSigPubKey(tx.GetSigPubKey())
	tempTx.TxBase.SetSig(tx.GetSig())
	tempTx.TxBase.SetProof(tx.GetProof())
	tempTx.TxBase.SetGetSenderAddrLastByte(tx.GetSenderAddrLastByte())
	tempTx.TxBase.SetMetadata(tx.GetMetadata())
	tempTx.TxBase.SetGetSenderAddrLastByte(tx.GetSenderAddrLastByte())

	return json.Marshal(tempTx)
}

func (txToken *TxTokenBase) UnmarshalJSON(data []byte) error {
	var err error
	if txToken.Tx, err = NewTransactionFromJsonBytes(data); err != nil {
		return err
	}
	temp := &struct {
		TxTokenData TxTokenData `json:"TxTokenPrivacyData"`
	}{}
	err = json.Unmarshal(data, &temp)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenJsonError, err)
	}
	TxTokenDataJson, err := json.MarshalIndent(temp.TxTokenData, "", "\t")
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(UnexpectedError, err)
	}
	err = json.Unmarshal(TxTokenDataJson, &txToken.TxTokenData)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenJsonError, err)
	}

	// TODO: hotfix, remove when fixed this issue
	if txToken.Tx.GetMetadata() != nil && txToken.Tx.GetMetadata().GetType() == 81 {
		if txToken.TxTokenData.Amount == 37772966455153490 {
			txToken.TxTokenData.Amount = 37772966455153487
		}
	}
	return nil
}

func (txToken TxTokenBase) String() string {
	// get hash of tx
	record := txToken.Tx.Hash().String()
	// add more hash of tx custom token data privacy
	tokenPrivacyDataHash, _ := txToken.TxTokenData.Hash()
	record += tokenPrivacyDataHash.String()

	meta := txToken.GetMetadata()
	if meta != nil {
		record += string(meta.Hash()[:])
	}
	return record
}

func (txToken TxTokenBase) JSONString() string {
	data, err := json.MarshalIndent(txToken, "", "\t")
	if err != nil {
		Logger.Log.Error(err)
		return ""
	}
	return string(data)
}

// =================== FUNCTIONS THAT GET STUFF ===================

// Hash returns the hash of all fields of the transaction
func (txToken *TxTokenBase) Hash() *common.Hash {
	hash := common.HashH([]byte(txToken.String()))
	return &hash
}

// GetTxActualSize computes the virtual size of a given transaction
// size of this tx = (normal TxNormal size) + (custom token data size)
func (txToken TxTokenBase) GetTxActualSize() uint64 {
	normalTxSize := txToken.Tx.GetTxActualSize()
	tokenDataSize := uint64(0)
	tokenDataSize += txToken.TxTokenData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(txToken.TxTokenData.PropertyName))
	tokenDataSize += uint64(len(txToken.TxTokenData.PropertySymbol))
	tokenDataSize += uint64(len(txToken.TxTokenData.PropertyID))
	tokenDataSize += 4 // for TxPrivacyTokenDataVersion1.Type
	tokenDataSize += 8 // for TxPrivacyTokenDataVersion1.Amount
	meta := txToken.GetMetadata()
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return normalTxSize + uint64(math.Ceil(float64(tokenDataSize)/1024))
}

func (txToken TxTokenBase) GetTxPrivacyTokenActualSize() uint64 {
	tokenDataSize := uint64(0)
	tokenDataSize += txToken.TxTokenData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(txToken.TxTokenData.PropertyName))
	tokenDataSize += uint64(len(txToken.TxTokenData.PropertySymbol))
	tokenDataSize += uint64(len(txToken.TxTokenData.PropertyID))
	tokenDataSize += 4 // for TxPrivacyTokenDataVersion1.Type
	tokenDataSize += 8 // for TxPrivacyTokenDataVersion1.Amount

	meta := txToken.TxTokenData.TxNormal.GetMetadata()
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return uint64(math.Ceil(float64(tokenDataSize) / 1024))
}

// Get SigPubKey of ptoken
func (txToken TxTokenBase) GetSigPubKey() []byte {
	return txToken.TxTokenData.TxNormal.GetSigPubKey()
}

// GetTxFeeToken - return Token Fee use to pay for privacy token Tx
func (txToken TxTokenBase) GetTxFeeToken() uint64 {
	return txToken.TxTokenData.TxNormal.GetTxFee()
}

func (txToken TxTokenBase) GetTokenID() *common.Hash {
	return &txToken.TxTokenData.PropertyID
}

func (txToken TxTokenBase) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	pubkeys, amounts := txToken.TxTokenData.TxNormal.GetReceivers()
	if len(pubkeys) == 0 {
		Logger.Log.Error("GetTransferData receive 0 output, it should has exactly 1 output")
		return false, nil, 0, &txToken.TxTokenData.PropertyID
	}
	if len(pubkeys) > 1 {
		Logger.Log.Error("GetTransferData receiver: More than 1 receiver")
		return false, nil, 0, &txToken.TxTokenData.PropertyID
	}
	return true, pubkeys[0], amounts[0], &txToken.TxTokenData.PropertyID
}

func (txToken TxTokenBase) GetTxMintData() (bool, coin.Coin, *common.Hash, error) {
	tx := txToken.TxTokenData.TxNormal
	isMinted, outputCoin, _, err := tx.GetTxMintData()
	return isMinted, outputCoin, &txToken.TxTokenData.PropertyID, err
}

func (txToken TxTokenBase) GetTxBurnData() (bool, coin.Coin, *common.Hash, error) {
	tx := txToken.TxTokenData.TxNormal
	isBurned, outputCoin, _, err := tx.GetTxBurnData()
	return isBurned, outputCoin, &txToken.TxTokenData.PropertyID, err
}

// CalculateBurnAmount - get tx value for pToken
func (txToken TxTokenBase) CalculateTxValue() uint64 {
	proof := txToken.TxTokenData.TxNormal.GetProof()
	if proof == nil {
		return 0
	}
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

	if txToken.TxTokenData.TxNormal.IsPrivacy() {
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

func (txToken TxTokenBase) ListSerialNumbersHashH() []common.Hash {
	tx := txToken.Tx
	result := []common.Hash{}
	if tx.GetProof() != nil {
		for _, d := range tx.GetProof().GetInputCoins() {
			hash := common.HashH(d.GetKeyImage().ToBytesS())
			result = append(result, hash)
		}
	}
	customTokenPrivacy := txToken.TxTokenData
	if customTokenPrivacy.TxNormal.GetProof() != nil {
		for _, d := range customTokenPrivacy.TxNormal.GetProof().GetInputCoins() {
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
func (txToken TxTokenBase) GetTxFee() uint64 {
	return txToken.Tx.GetTxFee()
}

// ================== NORMAL INIT FUNCTIONS ===================

func estimateTxSizeOfInitTokenSalary(publicKey []byte, amount uint64, coinName string, coinID *common.Hash) uint64 {
	receiver := &privacy.PaymentInfo{
		Amount: amount,
		PaymentAddress: privacy.PaymentAddress{
			Pk: publicKey,
			Tk: []byte{},
		},
	}
	propString := common.TokenHashToString(coinID)
	tokenParams := &CustomTokenPrivacyParamTx{
		PropertyID:     propString,
		PropertyName:   coinName,
		PropertySymbol: coinName,
		Amount:         amount,
		TokenTxType:    CustomTokenInit,
		Receiver:       []*privacy.PaymentInfo{receiver},
		TokenInput:     []coin.PlainCoin{},
		Mintable:       true,
	}
	estimateTxSizeParam := NewEstimateTxSizeParam(0, 0, false, nil, tokenParams, uint64(0))
	return EstimateTxSize(estimateTxSizeParam)
}

// =================== FUNCTION THAT CHECK STUFFS  ===================

// IsCoinsBurning - checking this is a burning pToken
func (txToken TxTokenBase) IsCoinsBurning(bcr metadata.ChainRetriever, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) bool {
	// get proof of pToken
	proof := txToken.TxTokenData.TxNormal.GetProof()
	if proof == nil || len(proof.GetOutputCoins()) == 0 {
		return false
	}
	return txToken.TxTokenData.TxNormal.IsCoinsBurning(bcr, retriever, viewRetriever, beaconHeight)
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

func (txToken TxTokenBase) ValidateType() bool {
	return txToken.Tx.GetType() == common.TxCustomTokenPrivacyType
}

func (txToken TxTokenBase) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	err := txToken.validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(DoubleSpendError, err)
	}
	// TODO: will move this to mempool process
	if txToken.TxTokenData.Type == CustomTokenInit && txToken.GetMetadata() == nil {
		initTokenID := txToken.TxTokenData.PropertyID
		txsInMem := mr.GetTxsInMem()
		for _, tx := range txsInMem {
			// try parse to TxTokenBase
			var privacyTokenTx, ok = tx.Tx.(TxTokenInterface)
			txTokenData := privacyTokenTx.GetTxPrivacyTokenData()
			if ok && txTokenData.Type == CustomTokenInit && privacyTokenTx.GetMetadata() == nil {
				// check > 1 tx init token by the same token ID
				if txTokenData.PropertyID.IsEqual(&initTokenID) {
					return NewTransactionErr(TokenIDInvalidError, fmt.Errorf("had already tx for initing token ID %s in pool", txTokenData.PropertyID.String()), txTokenData.PropertyID.String())
				}
			}
		}
	}
	return nil
}

func (txToken TxTokenBase) validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH map[common.Hash][]common.Hash) error {
	// check proof of PRV and pToken
	if txToken.GetProof() == nil && txToken.TxTokenData.TxNormal.GetProof() == nil {
		return errors.New("empty tx")
	}

	// collect serial number for PRV
	temp := make(map[common.Hash]interface{})
	if txToken.GetProof() != nil {
		for _, desc := range txToken.GetProof().GetInputCoins() {
			hash := common.HashH(desc.GetKeyImage().ToBytesS())
			temp[hash] = nil
		}
	}
	// collect serial number for pToken
	txNormalProof := txToken.TxTokenData.TxNormal.GetProof()
	if txNormalProof != nil {
		for _, desc := range txNormalProof.GetInputCoins() {
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

func (txToken TxTokenBase) ValidateTxWithBlockChain(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error {
	err := txToken.ValidateDoubleSpendWithBlockchain(shardID, stateDB, nil)
	if err != nil {
		return NewTransactionErr(InvalidDoubleSpendPRVError, err)
	}
	err = txToken.TxTokenData.TxNormal.ValidateDoubleSpendWithBlockchain(shardID, stateDB, txToken.GetTokenID())
	if err != nil {
		return NewTransactionErr(InvalidDoubleSpendPrivacyTokenError, err)
	}
	return nil
}

// ValidateSanityData - validate sanity data of PRV and pToken
func (txToken TxTokenBase) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	meta := txToken.Tx.GetMetadata()
	if meta != nil {
		isContinued, ok, err := meta.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, &txToken)
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}

	// validate sanity data for PRV
	//result, err := txToken.Tx.validateNormalTxSanityData()
	result, err := txToken.Tx.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if err != nil {
		return result, NewTransactionErr(InvalidSanityDataPRVError, err)
	}
	// validate sanity for pToken

	//result, err = txToken.TxPrivacyTokenDataVersion1.TxNormal.validateNormalTxSanityData()
	result, err = txToken.TxTokenData.TxNormal.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if err != nil {
		return result, NewTransactionErr(InvalidSanityDataPrivacyTokenError, err)
	}
	return result, nil
}

// VerifyMinerCreatedTxBeforeGettingInBlock
func (txToken TxTokenBase) VerifyMinerCreatedTxBeforeGettingInBlock(mintData *metadata.MintData, shardID byte, bcr metadata.ChainRetriever, accumulatedValues *metadata.AccumulatedValues, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever) (bool, error) {
	if !txToken.TxTokenData.Mintable {
		return true, nil
	}
	meta := txToken.Tx.GetMetadata()
	if meta == nil {
		Logger.Log.Error("Mintable custom token must contain metadata")
		return false, nil
	}
	if !meta.IsMinerCreatedMetaType() {
		return false, nil
	}
	return meta.VerifyMinerCreatedTxBeforeGettingInBlock(mintData, shardID, &txToken, bcr, accumulatedValues, retriever, viewRetriever)
}
