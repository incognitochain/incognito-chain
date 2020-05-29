package transaction

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"

	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/wallet"
)

type Tx struct {
	// Basic data, required
	Version  int8   `json:"Version"`
	Type     string `json:"Type"` // Transaction type
	LockTime int64  `json:"LockTime"`
	Fee      uint64 `json:"Fee"` // Fee applies: always consant
	Info     []byte // 512 bytes
	// Sign and Privacy proof, required
	SigPubKey            []byte `json:"SigPubKey, omitempty"` // 33 bytes
	Sig                  []byte `json:"Sig, omitempty"`       //
	Proof                privacy.Proof
	PubKeyLastByteSender byte
	// Metadata, optional
	Metadata metadata.Metadata
	// private field, not use for json parser, only use as temp variable
	sigPrivKey       []byte       // is ALWAYS private property of struct, if privacy: 64 bytes, and otherwise, 32 bytes
	cachedHash       *common.Hash // cached hash data of tx
	cachedActualSize *uint64      // cached actualsize data for tx
}

func parseProof(p interface{}, ver int8) (privacy.Proof, error) {
	// If transaction is nonPrivacyNonInput then we do not have proof, so parse it as nil
	if p == nil {
		return nil, nil
	}
	proofInBytes, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	var res privacy.Proof
	if ver == txVersion1 {
		res = &zkp.PaymentProof{}
	} else if ver == txVersion2 {
		res = &privacy_v2.PaymentProofV2{}
	} else if ver == txConversionVersion12 {
		res = &privacy_v2.ConversionProofVer1ToVer2{}
	} else {
		return nil, errors.New("ParseProof: Tx.Version is not 1 or 2 or -1")
	}
	res.Init()
	err = json.Unmarshal(proofInBytes, res)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (tx *Tx) UnmarshalJSON(data []byte) error {
	// For rolling version
	type Alias Tx
	temp := &struct {
		Metadata interface{}
		Proof    interface{}
		*Alias
	}{
		Alias: (*Alias)(tx),
	}
	err := json.Unmarshal(data, temp)
	if err != nil {
		Logger.Log.Error("UnmarshalJSON tx", string(data))
		return NewTransactionErr(UnexpectedError, err)
	}
	meta, parseErr := metadata.ParseMetadata(temp.Metadata)
	if parseErr != nil {
		Logger.Log.Error(parseErr)
		return parseErr
	}
	tx.SetMetadata(meta)
	proof, proofErr := parseProof(temp.Proof, tx.Version)
	if proofErr != nil {
		Logger.Log.Error(proofErr)
		return proofErr
	}
	tx.SetProof(proof)

	return nil
}

type TxPrivacyInitParams struct {
	senderSK    *privacy.PrivateKey
	paymentInfo []*privacy.PaymentInfo
	inputCoins  []coin.PlainCoin
	fee         uint64
	hasPrivacy  bool
	stateDB     *statedb.StateDB
	tokenID     *common.Hash // default is nil -> use for prv coin
	metaData    metadata.Metadata
	info        []byte // 512 bytes
}

func NewTxPrivacyInitParams(senderSK *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoins []coin.PlainCoin,
	fee uint64,
	hasPrivacy bool,
	stateDB *statedb.StateDB,
	tokenID *common.Hash, // default is nil -> use for prv coin
	metaData metadata.Metadata,
	info []byte) *TxPrivacyInitParams {
	params := &TxPrivacyInitParams{
		stateDB:     stateDB,
		tokenID:     tokenID,
		hasPrivacy:  hasPrivacy,
		inputCoins:  inputCoins,
		fee:         fee,
		metaData:    metaData,
		paymentInfo: paymentInfo,
		senderSK:    senderSK,
		info:        info,
	}
	return params
}

func validateTxInit(params *TxPrivacyInitParams) error {
	if len(params.inputCoins) > 255 {
		return NewTransactionErr(InputCoinIsVeryLargeError, nil, strconv.Itoa(len(params.inputCoins)))
	}
	if len(params.paymentInfo) > 254 {
		return NewTransactionErr(PaymentInfoIsVeryLargeError, nil, strconv.Itoa(len(params.paymentInfo)))
	}
	limitFee := uint64(0)
	estimateTxSizeParam := NewEstimateTxSizeParam(len(params.inputCoins), len(params.paymentInfo),
		params.hasPrivacy, nil, nil, limitFee)
	if txSize := EstimateTxSize(estimateTxSizeParam); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	if params.tokenID == nil {
		// using default PRV
		params.tokenID = &common.Hash{}
		err := params.tokenID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return NewTransactionErr(TokenIDInvalidError, err, params.tokenID.String())
		}
	}
	return nil
}

// Don't change to pointer. Because if not splitting to function like this (put this in Init func), the performance still be the same. If change to pointer it could be wrong.
func getTxInfo(paramInfo []byte) ([]byte, error) {
	if lenTxInfo := len(paramInfo); lenTxInfo > MaxSizeInfo {
		return []byte{}, NewTransactionErr(ExceedSizeInfoTxError, nil)
	}
	return paramInfo, nil
}

func (tx *Tx) isNonPrivacyNonInput(params *TxPrivacyInitParams) (bool, error) {
	Logger.Log.Debugf("len(inputCoins), fee, hasPrivacy: %d, %d, %v\n", len(params.inputCoins), params.fee, params.hasPrivacy)
	if len(params.inputCoins) == 0 && params.fee == 0 && !params.hasPrivacy {
		Logger.Log.Debugf("len(inputCoins) == 0 && fee == 0 && !hasPrivacy\n")
		tx.sigPrivKey = *params.senderSK
		err := signTx(tx)
		if err != nil {
			Logger.Log.Error(errors.New(fmt.Sprintf("Cannot sign tx %v\n", err)))
			return true, NewTransactionErr(SignTxError, err)
		}
		return true, nil
	}
	return false, nil
}

func updateParamsWhenOverBalance(params *TxPrivacyInitParams) error {
	// Calculate sum of all output coins' value
	sumOutputValue := uint64(0)
	for _, p := range params.paymentInfo {
		sumOutputValue += p.Amount
	}

	// Calculate sum of all input coins' value
	sumInputValue := uint64(0)
	for _, coin := range params.inputCoins {
		sumInputValue += coin.GetValue()
	}
	Logger.Log.Debugf("sumInputValue: %d\n", sumInputValue)

	overBalance := int64(sumInputValue - sumOutputValue - params.fee)
	// Check if sum of input coins' value is at least sum of output coins' value and tx fee
	if overBalance < 0 {
		return NewTransactionErr(WrongInputError, errors.New(fmt.Sprintf("input value less than output value. sumInputValue=%d sumOutputValue=%d fee=%d", sumInputValue, sumOutputValue, params.fee)))
	}
	// Create a new payment to sender's pk where amount is overBalance if > 0
	if overBalance > 0 {
		// Should not check error because have checked before
		senderFullKey, _ := parseSenderFullKey(params.senderSK)
		changePaymentInfo := new(privacy.PaymentInfo)
		changePaymentInfo.Amount = uint64(overBalance)
		changePaymentInfo.PaymentAddress = senderFullKey.PaymentAddress
		params.paymentInfo = append(params.paymentInfo, changePaymentInfo)
	}

	return nil
}

func parseLastByteSender(senderSK *privacy.PrivateKey) (byte, error) {
	senderFullKey, err := parseSenderFullKey(senderSK)
	if err != nil {
		return 0, err
	}
	return senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1], nil
}

func parseSenderFullKey(senderSK *privacy.PrivateKey) (*incognitokey.KeySet, error) {
	senderFullKey := incognitokey.KeySet{}
	err := senderFullKey.InitFromPrivateKey(senderSK)
	if err != nil {
		Logger.Log.Error(errors.New(fmt.Sprintf("Can not import Private key for sender keyset from %+v", senderSK)))
		return nil, NewTransactionErr(PrivateKeySenderInvalidError, err)
	}
	return &senderFullKey, nil
}

func getTxVersionFromCoins(inputCoins []coin.PlainCoin) (int8, error) {
	// If this is nonPrivacyNonInputCoins (maybe)
	if len(inputCoins) == 0 {
		return currentTxVersion, nil
	}
	check := [3]bool{}
	for i := 0; i < len(inputCoins); i += 1 {
		check[inputCoins[i].GetVersion()] = true
	}

	// If inputCoins contain 2 versions
	if check[1] && check[2] {
		return 0, errors.New("Cannot get tx version because there are 2 versions of input coins")
	}

	// If somehow no version is checked???
	if !check[1] && !check[2] {
		return 0, errors.New("Cannot get tx version, something is wrong with coins.version, it should be 1 or 2 only")
	}

	var res int8
	if check[1] {
		res = 1
	} else {
		res = 2
	}
	return res, nil
}

// return bool indicates whether we should continue "Init" function or not
func initializeTxAndParams(tx *Tx, params *TxPrivacyInitParams) (bool, error) {
	var err error

	// Tx: initialize some values
	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}
	tx.Fee = params.fee
	tx.Type = common.TxNormalType
	tx.Metadata = params.metaData

	if tx.Version, err = getTxVersionFromCoins(params.inputCoins); err != nil {
		return false, err
	}
	if tx.Info, err = getTxInfo(params.info); err != nil {
		return false, err
	}
	if tx.PubKeyLastByteSender, err = parseLastByteSender(params.senderSK); err != nil {
		return false, err
	}
	// we should stop if it isNonPrivacyNonInput or have error
	if isNonPrivacy, err := tx.isNonPrivacyNonInput(params); err != nil || isNonPrivacy {
		return !isNonPrivacy, err
	}

	// Params: update balance if overbalance
	if err = updateParamsWhenOverBalance(params); err != nil {
		return false, err
	}
	return true, nil
}

func (tx *Tx) ShouldSignMetaData() bool {
	if tx.GetMetadata() == nil {
		return false
	}
	meta := tx.GetMetadata()
	return meta.ShouldSignMetaData()
}

// Init - init value for tx from inputcoin(old output coin from old tx)
// create new outputcoin and build privacy proof
// if not want to create a privacy tx proof, set hashPrivacy = false
// database is used like an interface which use to query info from db in building tx
func (tx *Tx) Init(params *TxPrivacyInitParams) error {
	Logger.Log.Debugf("CREATING TX........\n")
	if err := validateTxInit(params); err != nil {
		return err
	}

	// Execution time
	start := time.Now()

	// Init tx and params (tx and params will be changed)
	if toContinue, err := initializeTxAndParams(tx, params); err != nil || !toContinue {
		return err
	}

	// Calculate execution time for creating payment proof
	startPrivacy := time.Now()

	// Prover will prove and sign based on version
	prover := newTxVersionSwitcher(tx.Version)
	if err := prover.Prove(tx, params); err != nil {
		return err
	}

	elapsedPrivacy := time.Since(startPrivacy)
	elapsed := time.Since(start)
	Logger.Log.Debugf("Creating payment proof time %s", elapsedPrivacy)
	Logger.Log.Debugf("Successfully Creating normal tx %+v in %s time", *tx.Hash(), elapsed)
	return nil
}

// ValidateTransaction returns true if transaction is valid:
// - Verify tx signature
// - Verify the payment proof
func (tx *Tx) ValidateTransaction(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	//hasPrivacy = false
	Logger.Log.Debugf("VALIDATING TX........\n")
	if tx.GetType() == common.TxRewardType {
		return tx.ValidateTxSalary(transactionStateDB)
	}
	if tx.GetType() == common.TxReturnStakingType {
		return tx.ValidateTxReturnStaking(transactionStateDB), nil
	}
	if tx.Version == txConversionVersion12 {
		return validateConversionVer1ToVer2(tx, transactionStateDB, shardID, tokenID)
	}

	// Verifier will verify based on tx.Version
	verifier := newTxVersionSwitcher(tx.Version)
	return verifier.Verify(tx, hasPrivacy, transactionStateDB, bridgeStateDB, shardID, tokenID, isBatch, isNewTransaction)
}

func (tx Tx) StringWithoutMetadataSig() string {
	record := strconv.Itoa(int(tx.Version))
	record += strconv.FormatInt(tx.LockTime, 10)
	record += strconv.FormatUint(tx.Fee, 10)
	if tx.Proof != nil {
		record += base64.StdEncoding.EncodeToString(tx.Proof.Bytes())
	}
	if tx.Metadata != nil {
		var metadataHash *common.Hash
		if meta, ok := tx.Metadata.(*metadata.StopAutoStakingMetadata); ok {
			metadataHash = meta.HashWithoutSig()
		} else {
			metadataHash = tx.Metadata.Hash()
		}
		record += metadataHash.String()
	}
	record += string(tx.Info)
	return record
}

func (tx Tx) String() string {
	record := strconv.Itoa(int(tx.Version))
	record += strconv.FormatInt(tx.LockTime, 10)
	record += strconv.FormatUint(tx.Fee, 10)
	if tx.Proof != nil {
		//tmp := base58.Base58Check{}.Encode(tx.Proof.Bytes(), 0x00)
		record += base64.StdEncoding.EncodeToString(tx.Proof.Bytes())
		// fmt.Printf("Proof check base 58: %v\n",tmp)
	}
	if tx.Metadata != nil {
		metadataHash := tx.Metadata.Hash()
		record += metadataHash.String()
		//Logger.log.Debugf("\n\n\n\n test metadata after hashing: %v\n", metadataHash.GetBytes())
	}
	record += string(tx.Info)
	return record
}

func (tx *Tx) Hash() *common.Hash {
	//if tx.cachedHash != nil {
	//	return tx.cachedHash
	//}
	inBytes := []byte(tx.String())
	hash := common.HashH(inBytes)
	//tx.cachedHash = &hash
	return &hash
}

func (tx Tx) GetSenderAddrLastByte() byte {
	return tx.PubKeyLastByteSender
}

func (tx *Tx) CheckAuthorizedSender(publicKey []byte) (bool, error) {
	checker := newTxVersionSwitcher(tx.Version)
	return checker.CheckAuthorizedSender(tx, publicKey)
}

func (tx Tx) GetTxFee() uint64 {
	return tx.Fee
}

func (tx Tx) GetTxFeeToken() uint64 {
	return uint64(0)
}

// GetTxActualSize computes the actual size of a given transaction in kilobyte
func (tx Tx) GetTxActualSize() uint64 {
	//txBytes, _ := json.Marshal(tx)
	//txSizeInByte := len(txBytes)
	//
	//return uint64(math.Ceil(float64(txSizeInByte) / 1024))
	if tx.cachedActualSize != nil {
		return *tx.cachedActualSize
	}
	sizeTx := uint64(1)                // int8
	sizeTx += uint64(len(tx.Type) + 1) // string
	sizeTx += uint64(8)                // int64
	sizeTx += uint64(8)

	sigPubKey := uint64(len(tx.SigPubKey))
	sizeTx += sigPubKey
	sig := uint64(len(tx.Sig))
	sizeTx += sig
	if tx.Proof != nil {
		proof := uint64(len(tx.Proof.Bytes()))
		sizeTx += proof
	}

	sizeTx += uint64(1)
	info := uint64(len(tx.Info))
	sizeTx += info

	meta := tx.Metadata
	if meta != nil {
		metaSize := meta.CalculateSize()
		sizeTx += metaSize
	}
	result := uint64(math.Ceil(float64(sizeTx) / 1024))
	tx.cachedActualSize = &result
	return *tx.cachedActualSize
}

// GetType returns the type of the transaction
func (tx Tx) GetType() string {
	return tx.Type
}

func (tx Tx) ListSerialNumbersHashH() []common.Hash {
	result := []common.Hash{}
	if tx.Proof != nil {
		for _, d := range tx.Proof.GetInputCoins() {
			hash := common.HashH(d.GetKeyImage().ToBytesS())
			result = append(result, hash)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result
}

// CheckCMExistence returns true if cm exists in cm list
func (tx Tx) CheckCMExistence(cm []byte, stateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (bool, error) {
	ok, err := statedb.HasCommitment(stateDB, *tokenID, cm, shardID)
	return ok, err
}

func (tx Tx) CheckTxVersion(maxTxVersion int8) bool {
	return !(tx.Version > maxTxVersion)
}

//func (tx Tx) CheckTransactionFee(minFeePerKbTx uint64) bool {
//	if tx.IsSalaryTx() {
//		return true
//	}
//	if tx.Metadata != nil {
//		return tx.Metadata.CheckTransactionFee(&tx, minFeePerKbTx)
//	}
//	fullFee := minFeePerKbTx * tx.GetTxActualSize()
//	return tx.Fee >= fullFee
//}

func (tx Tx) IsSalaryTx() bool {
	// Check normal tx(not an action tx)
	if tx.GetType() != common.TxRewardType {
		return false
	}
	// Check serialNumber in every Descs
	if len(tx.Proof.GetInputCoins()) == 0 {
		return true
	}
	return false
}

func (tx Tx) GetSender() []byte {
	if tx.Proof == nil || len(tx.Proof.GetInputCoins()) == 0 {
		return nil
	}
	if tx.IsPrivacy() {
		return nil
	}
	if len(tx.Proof.GetInputCoins()) == 0 || tx.Proof.GetInputCoins()[0] == nil {
		return nil
	}
	return tx.Proof.GetInputCoins()[0].GetPublicKey().ToBytesS()
}

func (tx Tx) GetReceivers() ([][]byte, []uint64) {
	pubkeys := [][]byte{}
	amounts := []uint64{}
	if tx.Proof != nil && len(tx.Proof.GetOutputCoins()) > 0 {
		for _, coin := range tx.Proof.GetOutputCoins() {
			added := false
			coinPubKey := coin.GetPublicKey().ToBytesS()
			for i, key := range pubkeys {
				if bytes.Equal(coinPubKey, key) {
					added = true
					amounts[i] += coin.GetValue()
					break
				}
			}
			if !added {
				pubkeys = append(pubkeys, coinPubKey)
				amounts = append(amounts, coin.GetValue())
			}
		}
	}
	return pubkeys, amounts
}

func (tx Tx) GetUniqueReceiver() (bool, []byte, uint64) {
	sender := []byte{} // Empty byte slice for coinbase tx
	if tx.Proof != nil && len(tx.Proof.GetInputCoins()) > 0 && !tx.IsPrivacy() {
		sender = tx.Proof.GetInputCoins()[0].GetPublicKey().ToBytesS()
	}
	pubkeys, amounts := tx.GetReceivers()
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

func (tx Tx) GetAndCheckBurningReceiver() (bool, []byte, uint64) {
	pubkeys, amounts := tx.GetReceivers()
	if len(pubkeys) > 2 {
		Logger.Log.Error("GetAndCheckBurning receiver: More than 2 receivers")
		return false, nil, 0
	}
	hasBurning, pubkey, amount := false, []byte{}, uint64(0)
	for i, pk := range pubkeys {
		if wallet.IsPublicKeyBurningAddress(pk) {
			hasBurning = true
			pubkey = pk
			amount += amounts[i]
		}
	}
	return hasBurning, pubkey, amount
}

func (tx Tx) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	pubkeys, amounts := tx.GetReceivers()
	if len(pubkeys) == 0 {
		Logger.Log.Error("GetTransferData receiver 0 output, it should has exactly 1 output")
		return false, nil, 0, &common.PRVCoinID
	}
	if len(pubkeys) > 1 {
		Logger.Log.Error("GetTransferData receiver: More than 1 receiver")
		return false, nil, 0, &common.PRVCoinID
	}
	return true, pubkeys[0], amounts[0], &common.PRVCoinID
}

func (tx Tx) GetTokenReceivers() ([][]byte, []uint64) {
	return nil, nil
}

func (tx Tx) GetTokenUniqueReceiver() (bool, []byte, uint64) {
	return false, nil, 0
}

func (tx Tx) validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH map[common.Hash][]common.Hash) error {
	if tx.Proof == nil {
		return nil
	}
	temp := make(map[common.Hash]interface{})
	for _, desc := range tx.Proof.GetInputCoins() {
		hash := common.HashH(desc.GetKeyImage().ToBytesS())
		temp[hash] = nil
	}

	for _, listSerialNumbers := range poolSerialNumbersHashH {
		for _, serialNumberHash := range listSerialNumbers {
			if _, ok := temp[serialNumberHash]; ok {
				return errors.New("double spend")
			}
		}
	}
	return nil
}

func (tx Tx) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	return tx.validateDoubleSpendTxWithCurrentMempool(poolSerialNumbersHashH)
}

// ValidateDoubleSpend - check double spend for any transaction type
func (tx Tx) ValidateDoubleSpendWithBlockchain(
	bcr metadata.BlockchainRetriever,
	shardID byte,
	stateDB *statedb.StateDB,
	tokenID *common.Hash,
) error {

	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return err
	}
	if tokenID != nil {
		err := prvCoinID.SetBytes(tokenID.GetBytes())
		if err != nil {
			return err
		}
	}
	if tx.Proof == nil {
		return nil
	}
	inputCoins := tx.Proof.GetInputCoins()
	for i := 0; i < len(inputCoins); i++ {
		serialNumber := inputCoins[i].GetKeyImage().ToBytesS()
		ok, err := statedb.HasSerialNumber(stateDB, *prvCoinID, serialNumber, shardID)
		if ok || err != nil {
			return errors.New("double spend")
		}
	}
	return nil
}

func (tx Tx) ValidateTxWithBlockChain(
	bcr metadata.BlockchainRetriever,
	shardID byte,
	stateDB *statedb.StateDB,
) error {
	if tx.GetType() == common.TxRewardType || tx.GetType() == common.TxReturnStakingType {
		return nil
	}
	if tx.Metadata != nil {
		isContinued, err := tx.Metadata.ValidateTxWithBlockChain(&tx, bcr, shardID, stateDB)
		fmt.Printf("[transactionStateDB] validate metadata with blockchain: %d %h %t %v\n", tx.GetMetadataType(), tx.Hash(), isContinued, err)
		if err != nil {
			Logger.Log.Errorf("[db] validate metadata with blockchain: %d %s %t %v\n", tx.GetMetadataType(), tx.Hash().String(), isContinued, err)
			return NewTransactionErr(RejectTxMedataWithBlockChain, fmt.Errorf("validate metadata of tx %s with blockchain error %+v", tx.Hash().String(), err))
		}
		if !isContinued {
			return nil
		}
	}
	return tx.ValidateDoubleSpendWithBlockchain(bcr, shardID, stateDB, nil)
}

func (tx Tx) ValidateSanityData(bcr metadata.BlockchainRetriever, beaconHeight uint64) (bool, error) {
	Logger.Log.Debugf("\n\n\n START Validating sanity data of metadata %+v\n\n\n", tx.Metadata)
	if tx.Metadata != nil {
		Logger.Log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := tx.Metadata.ValidateSanityData(bcr, &tx, beaconHeight)
		Logger.Log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	Logger.Log.Debugf("\n\n\n END sanity data of metadata%+v\n\n\n")
	//check version
	if tx.Version > txVersion2 {
		return false, NewTransactionErr(RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version <= %d", tx.Version, currentTxVersion))
	}
	// check LockTime before now
	if int64(tx.LockTime) > time.Now().Unix() {
		return false, NewTransactionErr(RejectInvalidLockTime, fmt.Errorf("wrong tx locktime %d", tx.LockTime))
	}

	// check tx size
	actualTxSize := tx.GetTxActualSize()
	if actualTxSize > common.MaxTxSize {
		//fmt.Print(actualTxSize, common.MaxTxSize)
		return false, NewTransactionErr(RejectTxSize, fmt.Errorf("tx size %d kB is too large", actualTxSize))
	}

	// check sanity of Proof
	if tx.Proof != nil {
		ok, err := tx.Proof.ValidateSanity()
		if !ok || err != nil {
			s := ""
			if !ok {
				s = "ValidateSanity Proof got error: ok = false\n"
			} else {
				s = fmt.Sprintf("ValidateSanity Proof got error: error %s\n", err.Error())
			}
			return false, errors.New(s)
		}
	}

	// check Type is normal or salary tx
	switch tx.Type {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenPrivacyType, common.TxReturnStakingType: //is valid
	default:
		return false, NewTransactionErr(RejectTxType, fmt.Errorf("wrong tx type with %s", tx.Type))
	}

	//if txN.Type != common.TxNormalType && txN.Type != common.TxRewardType && txN.Type != common.TxCustomTokenType && txN.Type != common.TxCustomTokenPrivacyType { // only 1 byte
	//	return false, errors.New("wrong tx type")
	//}

	// check info field
	if len(tx.Info) > 512 {
		return false, NewTransactionErr(RejectTxInfoSize, fmt.Errorf("wrong tx info length %d bytes, only support info with max length <= %d bytes", len(tx.Info), 512))
	}

	if tx.Version == 1 {
		if len(tx.SigPubKey) != common.SigPubKeySize {
			return false, NewTransactionErr(RejectTxPublickeySigSize, fmt.Errorf("wrong tx Sig PK size %d", len(tx.SigPubKey)))
		}
	}

	return true, nil
}

func (tx Tx) ValidateTxByItself(
	hasPrivacy bool,
	transactionStateDB *statedb.StateDB,
	bridgeStateDB *statedb.StateDB,
	bcr metadata.BlockchainRetriever,
	shardID byte,
	isNewTransaction bool,
) (bool, error) {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, err
	}
	ok, err := tx.ValidateTransaction(hasPrivacy, transactionStateDB, bridgeStateDB, shardID, prvCoinID, false, isNewTransaction)
	if !ok {
		return false, err
	}
	if tx.Metadata != nil {
		if hasPrivacy {
			return false, errors.New("Metadata can not exist in not privacy tx")
		}
		validateMetadata := tx.Metadata.ValidateMetadataByItself()
		if validateMetadata {
			return validateMetadata, nil
		} else {
			return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid"))
		}
	}
	return true, nil
}

// GetMetadataType returns the type of underlying metadata if is existed
func (tx Tx) GetMetadataType() int {
	if tx.Metadata != nil {
		return tx.Metadata.GetType()
	}
	return metadata.InvalidMeta
}

// GetMetadata returns metadata of tx is existed
func (tx Tx) GetMetadata() metadata.Metadata {
	return tx.Metadata
}

// SetMetadata sets metadata to tx
func (tx *Tx) SetMetadata(meta metadata.Metadata) {
	tx.Metadata = meta
}

func (tx *Tx) SetProof(proof privacy.Proof) {
	tx.Proof = proof
}

// GetMetadata returns metadata of tx is existed
func (tx Tx) GetInfo() []byte {
	return tx.Info
}

func (tx Tx) GetLockTime() int64 {
	return tx.LockTime
}

func (tx Tx) GetSigPubKey() []byte {
	return tx.SigPubKey
}

func (tx Tx) GetProof() privacy.Proof {
	return tx.Proof
}

func (tx Tx) IsPrivacy() bool {
	// In the case of NonPrivacyNonInput, we do not have proof
	if tx.Proof == nil {
		return false
	}
	return tx.Proof.IsPrivacy()
}

func (tx Tx) ValidateType() bool {
	return tx.Type == common.TxNormalType || tx.Type == common.TxRewardType || tx.Type == common.TxReturnStakingType
}

func (tx Tx) IsCoinsBurning(bcr metadata.BlockchainRetriever, beaconHeight uint64) bool {
	if tx.Proof == nil || len(tx.Proof.GetOutputCoins()) == 0 {
		return false
	}
	senderPKBytes := []byte{}
	outputCoins := tx.Proof.GetOutputCoins()
	inputCoins := tx.Proof.GetInputCoins()

	if len(inputCoins) > 0 {
		senderPKBytes = inputCoins[0].GetPublicKey().ToBytesS()
	}
	//get burning address
	burningAddress := bcr.GetBurningAddress(beaconHeight)
	keyWalletBurningAccount, err := wallet.Base58CheckDeserialize(burningAddress)
	if err != nil {
		return false
	}
	keysetBurningAccount := keyWalletBurningAccount.KeySet
	paymentAddressBurningAccount := keysetBurningAccount.PaymentAddress
	for _, outCoin := range outputCoins {
		outPKBytes := outCoin.GetPublicKey().ToBytesS()
		if !bytes.Equal(senderPKBytes, outPKBytes) && !bytes.Equal(outPKBytes, paymentAddressBurningAccount.Pk[:]) {
			return false
		}
	}
	return true
}

func (tx Tx) CalculateTxValue() uint64 {
	if tx.Proof == nil {
		return 0
	}

	outputCoins := tx.Proof.GetOutputCoins()
	inputCoins := tx.Proof.GetInputCoins()
	if outputCoins == nil || len(outputCoins) == 0 {
		return 0
	}
	if inputCoins == nil || len(inputCoins) == 0 { // coinbase tx
		txValue := uint64(0)
		for _, outCoin := range outputCoins {
			txValue += outCoin.GetValue()
		}
		return txValue
	}

	senderPKBytes := inputCoins[0].GetPublicKey().ToBytesS()
	txValue := uint64(0)
	for _, outCoin := range outputCoins {
		outPKBytes := outCoin.GetPublicKey().ToBytesS()
		if bytes.Equal(senderPKBytes, outPKBytes) {
			continue
		}
		txValue += outCoin.GetValue()
	}
	return txValue
}


// InitTxSalary
// Init salary transaction
// #1 - salary:
// #2 - receiverAddr:
// #3 - privKey:
// #4 - snDerivators:
func (tx *Tx) InitTxSalary(otaCoin *coin.CoinV2, privateKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata) error {
	tx.Version = txVersion2
	tx.Type = common.TxRewardType

	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}

	tempOutputCoin := make([]coin.Coin, 1)
	tempOutputCoin[0] = otaCoin

	proof := new(privacy.ProofV2)
	proof.Init()
	proof.SetOutputCoins(tempOutputCoin)
	tx.Proof = proof

	publicKeyBytes := otaCoin.GetPublicKey().ToBytesS()
	tx.PubKeyLastByteSender = publicKeyBytes[len(publicKeyBytes) - 1]

	// sign Tx using ver1 schnorr
	tx.sigPrivKey = *privateKey
	tx.SetMetadata(metaData)
	if err := signTx(tx); err != nil {
		return NewTransactionErr(SignTxError, err)
	}
	return nil
}

func (tx Tx) ValidateTxReturnStaking(stateDB *statedb.StateDB) bool {
	return true
}

func (tx Tx) ValidateTxSalary(stateDB *statedb.StateDB) (bool, error) {
	// verify signature
	valid, err := verifySigTx(&tx)
	if !valid {
		if err != nil {
			Logger.Log.Debugf("Error verifying signature of tx: %+v", err)
			return false, NewTransactionErr(VerifyTxSigFailError, err)
		}
		return false, nil
	}
	// check whether output coin's input exists in input list or not
	tokenID := &common.Hash{}
	err = tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, NewTransactionErr(TokenIDInvalidError, err, tokenID.String())
	}

	// Check commitment
	outputCoins := tx.Proof.GetOutputCoins()
	coin := outputCoins[0].(*coin.CoinV2)
	cmpCommitment := operation.PedCom.CommitAtIndex(coin.GetAmount(), coin.GetRandomness(), operation.PedersenValueIndex)
	if !operation.IsPointEqual(cmpCommitment, coin.GetCommitment()) {
		return false, NewTransactionErr(UnexpectedError, errors.New("check output coin's coin commitment isn't calculated correctly"))
	}

	// Check shardID
	coinShardID, errShard := coin.GetShardID()
	if errShard != nil {
		errStr := fmt.Sprintf("error when getting coin shardID, err: %v", errShard)
		return false, NewTransactionErr(UnexpectedError, errors.New(errStr))
	}
	if coinShardID != tx.PubKeyLastByteSender {
		return false, NewTransactionErr(UnexpectedError, errors.New("output coin's shardID is different from tx pubkey last byte"))
	}
	return true, nil
}

func (tx Tx) GetMetadataFromVinsTx(bcr metadata.BlockchainRetriever) (metadata.Metadata, error) {
	// implement this func if needed
	return nil, nil
}

func (tx Tx) GetTokenID() *common.Hash {
	return &common.PRVCoinID
}

func (tx Tx) VerifyMinerCreatedTxBeforeGettingInBlock(
	txsInBlock []metadata.Transaction,
	txsUsed []int,
	insts [][]string,
	instsUsed []int,
	shardID byte,
	bcr metadata.BlockchainRetriever,
	accumulatedValues *metadata.AccumulatedValues,
) (bool, error) {
	if tx.IsPrivacy() {
		return true, nil
	}
	meta := tx.Metadata

	inputCoins := tx.Proof.GetInputCoins()
	outputCoins := tx.Proof.GetOutputCoins()

	if tx.Proof != nil && len(inputCoins) == 0 && len(outputCoins) > 0 { // coinbase tx
		if meta == nil {
			return false, nil
		}
		if !meta.IsMinerCreatedMetaType() {
			return false, nil
		}
	}
	if meta != nil {
		ok, err := meta.VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock, txsUsed, insts, instsUsed, shardID, &tx, bcr, accumulatedValues)
		if err != nil {
			Logger.Log.Error(err)
			return false, NewTransactionErr(VerifyMinerCreatedTxBeforeGettingInBlockError, err)
		}
		return ok, nil
	}
	return true, nil
}

func (tx *Tx) HashWithoutMetadataSig() *common.Hash {
	inBytes := []byte(tx.StringWithoutMetadataSig())
	hash := common.HashH(inBytes)
	//tx.cachedHash = &hash
	return &hash
}