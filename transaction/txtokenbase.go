// Because txprivacytoken version 1 had a bug
// txprivacytoken in later version will not use the same base with txtokenversion1
// So we duplicate some code from ver1 to ver2 and not use any embedding

package transaction

import (
	"errors"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"math"
	"sort"
	"strconv"
)

type TxPrivacyTokenData struct {
	TxNormal TxBase
	PropertyID     common.Hash // = hash of TxCustomTokenprivacy data
	PropertyName   string
	PropertySymbol string

	Type     int    // action type
	Mintable bool   // default false
	Amount   uint64 // init amount
}

func (TxPrivacyTokenData TxPrivacyTokenData) GetPropertyID() common.Hash { return TxPrivacyTokenData.PropertyID }

func (TxPrivacyTokenData *TxPrivacyTokenData) SetPropertyID(propID common.Hash)  { TxPrivacyTokenData.PropertyID = propID }

func (TxPrivacyTokenData TxPrivacyTokenData) GetPropertyName() string { return TxPrivacyTokenData.PropertyName }

func (TxPrivacyTokenData *TxPrivacyTokenData) SetPropertyName(propertyName string) { TxPrivacyTokenData.PropertyName = propertyName }

func (TxPrivacyTokenData TxPrivacyTokenData) GetPropertySymbol() string { return TxPrivacyTokenData.PropertySymbol }

func (TxPrivacyTokenData *TxPrivacyTokenData) SetPropertySymbol(propertySymbol string)  { TxPrivacyTokenData.PropertySymbol = propertySymbol }

func (TxPrivacyTokenData TxPrivacyTokenData) GetType() int { return TxPrivacyTokenData.Type }

func (TxPrivacyTokenData *TxPrivacyTokenData) SetType(t int) { TxPrivacyTokenData.Type = t }

func (TxPrivacyTokenData TxPrivacyTokenData) IsMintable() bool { return TxPrivacyTokenData.Mintable }

func (TxPrivacyTokenData *TxPrivacyTokenData) SetMintable(mintable bool) { TxPrivacyTokenData.Mintable = mintable }

func (TxPrivacyTokenData TxPrivacyTokenData) GetAmount() uint64 { return TxPrivacyTokenData.Amount }

func (TxPrivacyTokenData *TxPrivacyTokenData) SetAmount(amount uint64) { TxPrivacyTokenData.Amount = amount }

func (TxPrivacyTokenData TxPrivacyTokenData) String() string {
	record := TxPrivacyTokenData.PropertyName
	record += TxPrivacyTokenData.PropertySymbol
	record += fmt.Sprintf("%d", TxPrivacyTokenData.Amount)
	if TxPrivacyTokenData.TxNormal.Proof != nil {
		inputCoins := TxPrivacyTokenData.TxNormal.GetProof().GetInputCoins()
		outputCoins := TxPrivacyTokenData.TxNormal.GetProof().GetOutputCoins()
		for _, out := range outputCoins {
			publicKeyBytes := []byte{}
			if out.GetPublicKey() != nil {
				publicKeyBytes = out.GetPublicKey().ToBytesS()
			}
			record += string(publicKeyBytes)
			record += strconv.FormatUint(out.GetValue(), 10)
		}
		for _, in := range inputCoins {
			publicKeyBytes := []byte{}
			if in.GetPublicKey() != nil {
				publicKeyBytes = in.GetPublicKey().ToBytesS()
			}
			record += string(publicKeyBytes)
			record += strconv.FormatUint(in.GetValue(), 10)
		}
	}
	return record
}

// Hash - return hash of custom token data, be used as Token ID
func (TxPrivacyTokenData TxPrivacyTokenData) Hash() (*common.Hash, error) {
	point := operation.HashToPoint([]byte(TxPrivacyTokenData.String()))
	hash := new(common.Hash)
	err := hash.SetBytes(point.ToBytesS())
	if err != nil {
		return nil, err
	}
	return hash, nil
}

// Interface for all type of transaction
type TxTokenInterface interface {
	// GET/SET FUNCTION
	GetTxBase() TxBase
	SetTxBase(TxBase)
	GetTxPrivacyTokenData() TxPrivacyTokenData
	SetTxPrivacyTokenData(TxPrivacyTokenData)

	// =================== FUNCTIONS THAT GET STUFF AND REQUIRE SOME CODING ===================
	GetTxActualSize() uint64
	GetReceivers() ([][]byte, []uint64)
	GetTransferData() (bool, []byte, uint64, *common.Hash)

	GetReceiverData() ([]*privacy.Point, []*coin.TxRandom, []uint64, error)
	GetTxMintData() (bool, []byte, []byte,  uint64, *common.Hash, error)
	GetTxBurnData(metadata.ChainRetriever, uint64) (bool, []byte, uint64, *common.Hash, error)

	GetMetadataFromVinsTx(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) (metadata.Metadata, error)
	ListSerialNumbersHashH() []common.Hash
	String() string
	Hash() *common.Hash
	CalculateTxValue() uint64

	// =================== FUNCTION THAT CHECK STUFFS  ===================
	CheckTxVersion(int8) bool
	CheckAuthorizedSender([]byte) (bool, error)
	ShouldSignMetaData() bool
	IsSalaryTx() bool
	IsPrivacy() bool
	IsCoinsBurning(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) bool

	// =================== FUNCTIONS THAT VALIDATE STUFFS ===================
	ValidateTxSalary(*statedb.StateDB) (bool, error)
	ValidateTxWithCurrentMempool(metadata.MempoolRetriever) error
	ValidateSanityData(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) (bool, error)
	ValidateTxWithBlockChain(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error
	ValidateDoubleSpendWithBlockchain(byte, *statedb.StateDB, *common.Hash) error
	ValidateTxByItself(bool, *statedb.StateDB, *statedb.StateDB, metadata.ChainRetriever, byte, bool, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) (bool, error)
	ValidateType() bool
	ValidateTransaction(bool, *statedb.StateDB, *statedb.StateDB, byte, *common.Hash, bool, bool) (bool, error)
	VerifyMinerCreatedTxBeforeGettingInBlock([]metadata.Transaction, []int, [][]string, []int, byte, metadata.ChainRetriever, *metadata.AccumulatedValues, metadata.ShardViewRetriever, metadata.BeaconViewRetriever) (bool, error)

	// Init Transaction, the input should be params such as: TxPrivacyInitParams
	Init(interface{}) error
	// Verify the init function above, which verify zero knowledge proof and signatures
	Verify(bool, *statedb.StateDB, *statedb.StateDB, byte, *common.Hash, bool, bool) (bool, error)
}

type TxTokenBase struct {
	TxBase
	TxPrivacyTokenData TxPrivacyTokenData
}

func newTxTokenFromParams(params *TxPrivacyTokenInitParams) (TxTokenInterface, error) {
	version, err := getTxVersionFromCoins(params.inputCoin)
	if err != nil {
		return nil, err
	}
	return newTxTokenFromVersionNumber(version)
}

func newTxTokenFromVersionNumber(version int8) (TxTokenInterface, error) {
	if version == txVersion1Number {
		return new(TxTokenVersion1), nil
	} else if version == txVersion2Number {
		return new(TxTokenVersion2), nil
	}
	return nil, errors.New("Version is not 1 or 2, cannot NewTxPrivacyFromParams")
}

// This function copies values from TxTokenBase to TxTokenInterface
// It does not copy sigPrivKey because it is private field
func newTxTokenInterfaceFromTxTokenBase(tx *TxTokenBase) (TxTokenInterface, error) {
	txInterface, err := newTxTokenFromVersionNumber(tx.GetVersion())
	if err != nil {
		return nil, err
	}
	txInterface.SetTxBase(tx.GetTxBase())
	txInterface.SetTxPrivacyTokenData(tx.GetTxPrivacyTokenData())
	return txInterface, nil
}

// This function copies values from TxTokenInterface to TxTokenBase
func newTxTokenBaseFromTxTokenInterface(tx TxTokenInterface) *TxTokenBase {
	txTokenBase := new(TxTokenBase)
	txTokenBase.SetTxBase(tx.GetTxBase())
	txTokenBase.SetTxPrivacyTokenData(tx.GetTxPrivacyTokenData())
	return txTokenBase
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

func (txToken TxTokenBase) GetTxBase() TxBase { return txToken.TxBase }
func (txToken *TxTokenBase) SetTxBase(tx TxBase) { txToken.TxBase = tx }
func (txToken TxTokenBase) GetTxPrivacyTokenData() TxPrivacyTokenData { return txToken.TxPrivacyTokenData }
func (txToken *TxTokenBase) SetTxPrivacyTokenData(data TxPrivacyTokenData)  { txToken.TxPrivacyTokenData = data }

// ========== CHECK FUNCTION ===========

func (txToken *TxTokenBase) CheckAuthorizedSender([]byte) (bool, error) {
	return false, errors.New("TxTokenBase does not has CheckAuthorizedSender")
}

// =================== PARSING JSON FUNCTIONS ===================

func (txToken *TxTokenBase) UnmarshalJSON(data []byte) error {
	tx := TxBase{}
	err := json.Unmarshal(data, &tx)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenPRVJsonError, err)
	}
	temp := &struct {
		TxPrivacyTokenData TxPrivacyTokenData
	}{}
	err = json.Unmarshal(data, &temp)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenJsonError, err)
	}
	TxTokenPrivacyDataJson, err := json.MarshalIndent(temp.TxPrivacyTokenData, "", "\t")
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(UnexpectedError, err)
	}
	err = json.Unmarshal(TxTokenPrivacyDataJson, &txToken.TxPrivacyTokenData)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(PrivacyTokenJsonError, err)
	}
	txToken.TxBase = tx

	// TODO: hotfix, remove when fixed this issue
	if tx.Metadata != nil && tx.Metadata.GetType() == 81 {
		if txToken.TxPrivacyTokenData.Amount == 37772966455153490 {
			txToken.TxPrivacyTokenData.Amount = 37772966455153487
		}
	}
	return nil
}

func (txToken TxTokenBase) String() string {
	// get hash of tx
	record := txToken.TxBase.Hash().String()
	// add more hash of tx custom token data privacy
	tokenPrivacyDataHash, _ := txToken.TxPrivacyTokenData.Hash()
	record += tokenPrivacyDataHash.String()
	if txToken.Metadata != nil {
		record += string(txToken.Metadata.Hash()[:])
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
	normalTxSize := txToken.TxBase.GetTxActualSize()
	tokenDataSize := uint64(0)
	tokenDataSize += txToken.TxPrivacyTokenData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(txToken.TxPrivacyTokenData.PropertyName))
	tokenDataSize += uint64(len(txToken.TxPrivacyTokenData.PropertySymbol))
	tokenDataSize += uint64(len(txToken.TxPrivacyTokenData.PropertyID))
	tokenDataSize += 4 // for TxPrivacyTokenDataVersion1.Type
	tokenDataSize += 8 // for TxPrivacyTokenDataVersion1.Amount
	meta := txToken.Metadata
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return normalTxSize + uint64(math.Ceil(float64(tokenDataSize)/1024))
}

func (txToken TxTokenBase) GetTxPrivacyTokenActualSize() uint64 {
	tokenDataSize := uint64(0)
	tokenDataSize += txToken.TxPrivacyTokenData.TxNormal.GetTxActualSize()
	tokenDataSize += uint64(len(txToken.TxPrivacyTokenData.PropertyName))
	tokenDataSize += uint64(len(txToken.TxPrivacyTokenData.PropertySymbol))
	tokenDataSize += uint64(len(txToken.TxPrivacyTokenData.PropertyID))
	tokenDataSize += 4 // for TxPrivacyTokenDataVersion1.Type
	tokenDataSize += 8 // for TxPrivacyTokenDataVersion1.Amount

	meta := txToken.TxPrivacyTokenData.TxNormal.Metadata
	if meta != nil {
		tokenDataSize += meta.CalculateSize()
	}

	return uint64(math.Ceil(float64(tokenDataSize) / 1024))
}

// Get SigPubKey of ptoken
func (txToken TxTokenBase) GetSigPubKey() []byte {
	return txToken.TxPrivacyTokenData.TxNormal.SigPubKey
}

// GetTxFeeToken - return Token Fee use to pay for privacy token Tx
func (txToken TxTokenBase) GetTxFeeToken() uint64 {
	return txToken.TxPrivacyTokenData.TxNormal.Fee
}

func (txToken TxTokenBase) GetTokenID() *common.Hash {
	return &txToken.TxPrivacyTokenData.PropertyID
}

func (txToken TxTokenBase) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	pubkeys, amounts := txToken.TxPrivacyTokenData.TxNormal.GetReceivers()
	if len(pubkeys) == 0 {
		Logger.Log.Error("GetTransferData receive 0 output, it should has exactly 1 output")
		return false, nil, 0, &txToken.TxPrivacyTokenData.PropertyID
	}
	if len(pubkeys) > 1 {
		Logger.Log.Error("GetTransferData receiver: More than 1 receiver")
		return false, nil, 0, &txToken.TxPrivacyTokenData.PropertyID
	}
	return true, pubkeys[0], amounts[0], &txToken.TxPrivacyTokenData.PropertyID
}

func (txToken TxTokenBase) GetTxMintData() (bool, []byte, []byte, uint64, *common.Hash, error) {
	tx := txToken.TxPrivacyTokenData.TxNormal
	isMinted, publicKey, txRandom, amount, _, err := tx.GetTxMintData()
	return isMinted, publicKey, txRandom, amount, &txToken.TxPrivacyTokenData.PropertyID, err
}

func (txToken TxTokenBase) GetTxBurnData(retriever metadata.ChainRetriever, blockHeight uint64) (bool, []byte, uint64, *common.Hash, error) {
	tx := txToken.TxPrivacyTokenData.TxNormal
	isBurned, pubkey, amount, _, err := tx.GetTxBurnData(retriever, blockHeight)
	return isBurned, pubkey, amount, &txToken.TxPrivacyTokenData.PropertyID, err
}

// CalculateBurnAmount - get tx value for pToken
func (txToken TxTokenBase) CalculateTxValue() uint64 {
	if txToken.TxPrivacyTokenData.TxNormal.Proof == nil {
		return 0
	}
	proof := txToken.TxPrivacyTokenData.TxNormal.Proof
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

	if txToken.TxPrivacyTokenData.TxNormal.IsPrivacy() {
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
	tx := txToken.TxBase
	result := []common.Hash{}
	if tx.Proof != nil {
		for _, d := range tx.Proof.GetInputCoins() {
			hash := common.HashH(d.GetKeyImage().ToBytesS())
			result = append(result, hash)
		}
	}
	customTokenPrivacy := txToken.TxPrivacyTokenData
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
func (txToken TxTokenBase) GetTxFee() uint64 {
	return txToken.TxBase.GetTxFee()
}

// ================== NORMAL INIT FUNCTIONS ===================

func (txToken *TxTokenBase) Init(paramsInterface interface{}) error {
	txTokenParams, ok := paramsInterface.(*TxPrivacyTokenInitParams)
	if !ok {
		return errors.New("params of tx Init is not TxPrivacyInitParam")
	}
	transactionToken, err := newTxTokenFromParams(txTokenParams)
	if err != nil {
		return err
	}
	err = transactionToken.Init(paramsInterface)

	// Copy value of transactionToken to txToken
	*txToken = *newTxTokenBaseFromTxTokenInterface(transactionToken)
	return err
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

func (txToken *TxTokenBase) InitTxTokenSalary(otaCoin *coin.CoinV2, privKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata, coinID *common.Hash, coinName string) error {
	var err error
	// init data for tx PRV for fee
	txPrivacyParams := NewTxPrivacyInitParams(
		privKey, []*privacy.PaymentInfo{}, nil, 0, false, stateDB, nil, metaData, nil,
	)
	txToken.TxBase = TxBase{}
	if err := txToken.TxBase.Init(txPrivacyParams); err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	txToken.TxBase.SetType(common.TxCustomTokenPrivacyType)
	// check tx size
	publicKeyBytes := otaCoin.GetPublicKey().ToBytesS()
	if txSize := estimateTxSizeOfInitTokenSalary(publicKeyBytes, otaCoin.GetValue(), coinName, coinID); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}
	// check action type and create privacy custom toke data
	txToken.TxPrivacyTokenData.SetType(CustomTokenInit)
	txToken.TxPrivacyTokenData.SetPropertyName(coinName)
	txToken.TxPrivacyTokenData.SetPropertySymbol(coinName)
	txToken.TxPrivacyTokenData.SetAmount(otaCoin.GetValue())

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
	// signOnMessage Tx
	temp.sigPrivKey = *privKey
	if temp.Sig, _, err = signNoPrivacy(privKey, temp.Hash()[:]); err != nil {
		Logger.Log.Error(errors.New("can't signOnMessage this tx"))
		return NewTransactionErr(SignTxError, err)
	}
	temp.SigPubKey = otaCoin.GetPublicKey().ToBytesS()
	var propertyID [common.HashSize]byte
	copy(propertyID[:], coinID[:])
	txToken.TxPrivacyTokenData.PropertyID = propertyID
	txToken.TxPrivacyTokenData.TxNormal = *NewTxBaseFromTransaction(&temp)
	txToken.TxPrivacyTokenData.Mintable = true
	return nil
}

// =================== FUNCTION THAT CHECK STUFFS  ===================

// IsCoinsBurning - checking this is a burning pToken
func (txToken TxTokenBase) IsCoinsBurning(bcr metadata.ChainRetriever, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) bool {
	// get proof of pToken
	proof := txToken.TxPrivacyTokenData.TxNormal.Proof
	if proof == nil || len(proof.GetOutputCoins()) == 0 {
		return false
	}
	return txToken.TxPrivacyTokenData.TxNormal.IsCoinsBurning(bcr, retriever, viewRetriever, beaconHeight)
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
	return txToken.Type == common.TxCustomTokenPrivacyType
}

func (txToken TxTokenBase) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	err := txToken.validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH)
	if err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(DoubleSpendError, err)
	}
	// TODO: will move this to mempool process
	if txToken.TxPrivacyTokenData.Type == CustomTokenInit && txToken.GetMetadata() == nil {
		initTokenID := txToken.TxPrivacyTokenData.PropertyID
		txsInMem := mr.GetTxsInMem()
		for _, tx := range txsInMem {
			// try parse to TxTokenBase
			var privacyTokenTx, ok = tx.Tx.(*TxTokenBase)
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

func (txToken TxTokenBase) validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH map[common.Hash][]common.Hash) error {
	// check proof of PRV and pToken
	if txToken.Proof == nil && txToken.TxPrivacyTokenData.TxNormal.Proof == nil {
		return errors.New("empty tx")
	}

	// collect serial number for PRV
	temp := make(map[common.Hash]interface{})
	if txToken.Proof != nil {
		for _, desc := range txToken.Proof.GetInputCoins() {
			hash := common.HashH(desc.GetKeyImage().ToBytesS())
			temp[hash] = nil
		}
	}
	// collect serial number for pToken
	if txToken.TxPrivacyTokenData.TxNormal.Proof != nil {
		for _, desc := range txToken.TxPrivacyTokenData.TxNormal.Proof.GetInputCoins() {
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
	err = txToken.TxPrivacyTokenData.TxNormal.ValidateDoubleSpendWithBlockchain(shardID, stateDB, txToken.GetTokenID())
	if err != nil {
		return NewTransactionErr(InvalidDoubleSpendPrivacyTokenError, err)
	}
	return nil
}

// ValidateSanityData - validate sanity data of PRV and pToken
func (txToken TxTokenBase) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	meta := txToken.TxBase.Metadata
	if meta != nil {
		isContinued, ok, err := meta.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, &txToken)
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}

	// validate sanity data for PRV
	//result, err := txToken.Tx.validateNormalTxSanityData()
	result, err := txToken.TxBase.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if err != nil {
		return result, NewTransactionErr(InvalidSanityDataPRVError, err)
	}
	// validate sanity for pToken

	//result, err = txToken.TxPrivacyTokenDataVersion1.TxNormal.validateNormalTxSanityData()
	result, err = txToken.TxPrivacyTokenData.TxNormal.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if err != nil {
		return result, NewTransactionErr(InvalidSanityDataPrivacyTokenError, err)
	}
	return result, nil
}

// ValidateTxByItself - validate tx by itself, check signature, proof,... and metadata
func (txToken TxTokenBase) ValidateTxByItself(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, chainRetriever metadata.ChainRetriever, shardID byte, isNewTransaction bool, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever) (bool, error) {
	if txToken.Metadata != nil {
		validateMetadata := txToken.Metadata.ValidateMetadataByItself()
		if !validateMetadata {
			return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid"))
		}
		return validateMetadata, nil
	}

	txInterface, err := newTxTokenInterfaceFromTxTokenBase(&txToken)
	if err != nil {
		return false, errors.New("Cannot create TxTokenInterface from txTokenBase")
	}
	return txInterface.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, nil, false, isNewTransaction)
}

func (txToken TxTokenBase) ValidateTransaction(hasPrivacyCoin bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	txInterface, err := newTxTokenInterfaceFromTxTokenBase(&txToken)
	if err != nil {
		return false, err
	}
	return txInterface.ValidateTransaction(hasPrivacyCoin, transactionStateDB, bridgeStateDB, shardID, tokenID, isBatch, isNewTransaction)
}

// GetProof - return proof PRV of tx
func (txToken TxTokenBase) GetProof() privacy.Proof {
	return txToken.Proof
}

// VerifyMinerCreatedTxBeforeGettingInBlock
func (txToken TxTokenBase) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []metadata.Transaction, txsUsed []int, insts [][]string, instsUsed []int, shardID byte, bcr metadata.ChainRetriever, accumulatedValues *metadata.AccumulatedValues, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever) (bool, error) {
	if !txToken.TxPrivacyTokenData.Mintable {
		return true, nil
	}
	meta := txToken.Metadata
	if meta == nil {
		Logger.Log.Error("Mintable custom token must contain metadata")
		return false, nil
	}
	if !meta.IsMinerCreatedMetaType() {
		return false, nil
	}
	return meta.VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock, txsUsed, insts, instsUsed, shardID, &txToken, bcr, accumulatedValues, nil, nil)
}