package transaction

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
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
	Proof                *privacy.Proof
	PubKeyLastByteSender byte

	// Metadata, optional
	Metadata metadata.Metadata

	// private field, not use for json parser, only use as temp variable
	sigPrivKey       []byte       // is ALWAYS private property of struct, if privacy: 64 bytes, and otherwise, 32 bytes
	cachedHash       *common.Hash // cached hash data of tx
	cachedActualSize *uint64      // cached actualsize data for tx
}

func (tx *Tx) UnmarshalJSON(data []byte) error {
	type Alias Tx
	temp := &struct {
		Metadata interface{}
		*Alias
	}{
		Alias: (*Alias)(tx),
	}
	err := json.Unmarshal(data, &temp)
	if err != nil {
		Logger.log.Error("UnmarshalJSON tx", string(data))
		return NewTransactionErr(UnexpectedError, err)
	}
	meta, parseErr := metadata.ParseMetadata(temp.Metadata)
	if parseErr != nil {
		Logger.log.Error(parseErr)
		return parseErr
	}
	tx.SetMetadata(meta)

	return nil
}

type TxPrivacyInitParams struct {
	senderSK    *privacy.PrivateKey
	paymentInfo []*privacy.PaymentInfo
	inputCoins  []*privacy.InputCoin
	fee         uint64
	hasPrivacy  bool
	db          database.DatabaseInterface
	tokenID     *common.Hash // default is nil -> use for prv coin
	metaData    metadata.Metadata
	info        []byte // 512 bytes
}

func NewTxPrivacyInitParams(senderSK *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoins []*privacy.InputCoin,
	fee uint64,
	hasPrivacy bool,
	db database.DatabaseInterface,
	tokenID *common.Hash, // default is nil -> use for prv coin
	metaData metadata.Metadata,
	info []byte) *TxPrivacyInitParams {
	params := &TxPrivacyInitParams{
		db:          db,
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
	Logger.log.Debugf("len(inputCoins), fee, hasPrivacy: %d, %d, %v\n", len(params.inputCoins), params.fee, params.hasPrivacy)
	if len(params.inputCoins) == 0 && params.fee == 0 && !params.hasPrivacy {
		Logger.log.Debugf("len(inputCoins) == 0 && fee == 0 && !hasPrivacy\n")
		tx.sigPrivKey = *params.senderSK
		err := signTx(tx)
		if err != nil {
			Logger.log.Error(errors.New(fmt.Sprintf("Cannot sign tx %v\n", err)))
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
		sumInputValue += coin.CoinDetails.GetValue()
	}
	Logger.log.Debugf("sumInputValue: %d\n", sumInputValue)

	overBalance := int64(sumInputValue - sumOutputValue - params.fee)
	// Check if sum of input coins' value is at least sum of output coins' value and tx fee
	if overBalance < 0 {
		return NewTransactionErr(WrongInputError, errors.New(fmt.Sprintf("input value less than output value. sumInputValue=%d sumOutputValue=%d fee=%d", sumInputValue, sumOutputValue, params.fee)))
	}
	// Create a new payment to sender's pk where amount is overBalance if > 0
	if overBalance > 0 {
		// Should not check error because have checked before
		senderFullKey, _ := parseSenderFullKey(params)
		changePaymentInfo := new(privacy.PaymentInfo)
		changePaymentInfo.Amount = uint64(overBalance)
		changePaymentInfo.PaymentAddress = senderFullKey.PaymentAddress
		params.paymentInfo = append(params.paymentInfo, changePaymentInfo)
	}

	return nil
}

func parseLastByteSender(params *TxPrivacyInitParams) (byte, error) {
	senderFullKey, err := parseSenderFullKey(params)
	if err != nil {
		return 0, err
	}
	return senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1], nil
}

func parseSenderFullKey(params *TxPrivacyInitParams) (*incognitokey.KeySet, error) {
	senderFullKey := incognitokey.KeySet{}
	err := senderFullKey.InitFromPrivateKey(params.senderSK)
	if err != nil {
		Logger.log.Error(errors.New(fmt.Sprintf("Can not import Private key for sender keyset from %+v", params.senderSK)))
		return nil, NewTransactionErr(PrivateKeySenderInvalidError, err)
	}
	return &senderFullKey, nil
}

func initializeTxAndParams(tx *Tx, params *TxPrivacyInitParams) error {
	var err error

	// Tx: initialize some values
	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}
	tx.Fee = params.fee
	tx.Version = txVersion
	tx.Type = common.TxNormalType
	tx.Metadata = params.metaData
	// tx.Proof = &privacy.Proof{}
	tx.Info, err = getTxInfo(params.info)
	if err != nil {
		return err
	}
	tx.PubKeyLastByteSender, err = parseLastByteSender(params)
	if err != nil {
		return err
	}
	if isNonPrivacy, err := tx.isNonPrivacyNonInput(params); isNonPrivacy {
		return err
	}

	// Params: update balance if overbalance
	if err = updateParamsWhenOverBalance(params); err != nil {
		return err
	}
	return nil
}

// Init - init value for tx from inputcoin(old output coin from old tx)
// create new outputcoin and build privacy proof
// if not want to create a privacy tx proof, set hashPrivacy = false
// database is used like an interface which use to query info from db in building tx
func (tx *Tx) Init(params *TxPrivacyInitParams) error {
	Logger.log.Debugf("CREATING TX........\n")
	if err := validateTxInit(params); err != nil {
		return err
	}

	// Execution time
	start := time.Now()

	// Init tx and params (tx and params will be changed)
	if err := initializeTxAndParams(tx, params); err != nil {
		return err
	}

	// Calculate execution time for creating payment proof
	startPrivacy := time.Now()

	if err := proveAndSignVersionSwitcher(tx, params); err != nil {
		return err
	}

	elapsedPrivacy := time.Since(startPrivacy)
	elapsed := time.Since(start)
	Logger.log.Debugf("Creating payment proof time %s", elapsedPrivacy)
	Logger.log.Debugf("Successfully Creating normal tx %+v in %s time", *tx.Hash(), elapsed)
	return nil
}

// verifySigTx - verify signature on tx
func (tx *Tx) verifySigTx() (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be an signed one"))
	}

	var err error
	res := false

	/****** verify Schnorr signature *****/
	// prepare Public key for verification
	verifyKey := new(privacy.SchnorrPublicKey)
	sigPublicKey, err := new(privacy.Point).FromBytesS(tx.SigPubKey)

	if err != nil {
		Logger.log.Error(err)
		return false, NewTransactionErr(DecompressSigPubKeyError, err)
	}
	verifyKey.Set(sigPublicKey)

	// convert signature from byte array to SchnorrSign
	signature := new(privacy.SchnSignature)
	err = signature.SetBytes(tx.Sig)
	if err != nil {
		Logger.log.Error(err)
		return false, NewTransactionErr(InitTxSignatureFromBytesError, err)
	}

	// verify signature
	/*Logger.log.Debugf(" VERIFY SIGNATURE ----------- HASH: %v\n", tx.Hash()[:])
	if tx.Proof != nil {
		Logger.log.Debugf(" VERIFY SIGNATURE ----------- TX Proof bytes before verifing the signature: %v\n", tx.Proof.Bytes())
	}
	Logger.log.Debugf(" VERIFY SIGNATURE ----------- TX meta: %v\n", tx.Metadata)*/
	res = verifyKey.Verify(signature, tx.Hash()[:])

	return res, nil
}

// ValidateTransaction returns true if transaction is valid:
// - Verify tx signature
// - Verify the payment proof
func (tx *Tx) ValidateTransaction(hasPrivacy bool, db database.DatabaseInterface, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	//hasPrivacy = false
	Logger.log.Debugf("VALIDATING TX........\n")
	// start := time.Now()
	// Verify tx signature
	if tx.GetType() == common.TxRewardType {
		return tx.ValidateTxSalary(db)
	}
	if tx.GetType() == common.TxReturnStakingType {
		return tx.ValidateTxReturnStaking(db), nil
	}

	return verifierVersionSwitcher(tx, hasPrivacy, db, shardID, tokenID, isBatch, isNewTransaction)
}

func (tx Tx) String() string {
	record := strconv.Itoa(int(tx.Version))

	record += strconv.FormatInt(tx.LockTime, 10)
	record += strconv.FormatUint(tx.Fee, 10)
	if tx.Proof != nil {
		tmp := base64.StdEncoding.EncodeToString((*tx.Proof).Bytes())
		//tmp := base58.Base58Check{}.Encode(tx.Proof.Bytes(), 0x00)
		record += tmp
		// fmt.Printf("Proof check base 58: %v\n",tmp)
	}
	if tx.Metadata != nil {
		metadataHash := tx.Metadata.Hash()
		//Logger.log.Debugf("\n\n\n\n test metadata after hashing: %v\n", metadataHash.GetBytes())
		metadataStr := metadataHash.String()
		record += metadataStr
	}

	//TODO: To be uncomment
	// record += string(tx.Info)
	return record
}

func (tx *Tx) Hash() *common.Hash {
	if tx.cachedHash != nil {
		return tx.cachedHash
	}
	inBytes := []byte(tx.String())
	hash := common.HashH(inBytes)
	tx.cachedHash = &hash
	return &hash
}

func (tx Tx) GetSenderAddrLastByte() byte {
	return tx.PubKeyLastByteSender
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
		proof := uint64(len((*tx.Proof).Bytes()))
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
		for _, d := range (*tx.Proof).GetInputCoins() {
			hash := common.HashH(d.CoinDetails.GetSerialNumber().ToBytesS())
			result = append(result, hash)
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].String() < result[j].String()
	})
	return result
}

// CheckCMExistence returns true if cm exists in cm list
func (tx Tx) CheckCMExistence(cm []byte, db database.DatabaseInterface, shardID byte, tokenID *common.Hash) (bool, error) {
	ok, err := db.HasCommitment(*tokenID, cm, shardID)
	return ok, err
}

func (tx Tx) CheckTxVersion(maxTxVersion int8) bool {
	return !(tx.Version > maxTxVersion)
}

// func (tx Tx) CheckTransactionFee(minFeePerKbTx uint64) bool {
// 	if tx.IsSalaryTx() {
// 		return true
// 	}
// 	if tx.Metadata != nil {
// 		return tx.Metadata.CheckTransactionFee(&tx, minFeePerKbTx)
// 	}
// 	fullFee := minFeePerKbTx * tx.GetTxActualSize()
// 	return tx.Fee >= fullFee
// }

func (tx Tx) IsSalaryTx() bool {
	// Check normal tx(not an action tx)
	if tx.GetType() != common.TxRewardType {
		return false
	}
	// Check serialNumber in every Descs
	txProof := *tx.Proof
	if len(txProof.GetInputCoins()) == 0 {
		return true
	}
	return false
}

func (tx Tx) GetSender() []byte {
	txProof := *tx.Proof
	if tx.Proof == nil || len(txProof.GetInputCoins()) == 0 {
		return nil
	}
	if tx.IsPrivacy() {
		return nil
	}
	if len((*tx.Proof).GetInputCoins()) == 0 || (*tx.Proof).GetInputCoins()[0].CoinDetails == nil {
		return nil
	}
	return (*tx.Proof).GetInputCoins()[0].CoinDetails.GetPublicKey().ToBytesS()
}

func (tx Tx) GetReceivers() ([][]byte, []uint64) {
	pubkeys := [][]byte{}
	amounts := []uint64{}
	txProof := *tx.Proof
	if tx.Proof != nil && len(txProof.GetOutputCoins()) > 0 {
		for _, coin := range txProof.GetOutputCoins() {
			added := false
			coinPubKey := coin.CoinDetails.GetPublicKey().ToBytesS()
			for i, key := range pubkeys {
				if bytes.Equal(coinPubKey, key) {
					added = true
					amounts[i] += coin.CoinDetails.GetValue()
					break
				}
			}
			if !added {
				pubkeys = append(pubkeys, coinPubKey)
				amounts = append(amounts, coin.CoinDetails.GetValue())
			}
		}
	}
	return pubkeys, amounts
}

func (tx Tx) GetUniqueReceiver() (bool, []byte, uint64) {
	sender := []byte{} // Empty byte slice for coinbase tx
	txProof := *tx.Proof
	if tx.Proof != nil && len(txProof.GetInputCoins()) > 0 && !tx.IsPrivacy() {
		sender = txProof.GetInputCoins()[0].CoinDetails.GetPublicKey().ToBytesS()
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

func (tx Tx) GetTransferData() (bool, []byte, uint64, *common.Hash) {
	unique, pk, amount := tx.GetUniqueReceiver()
	return unique, pk, amount, &common.PRVCoinID
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
	txProof := *tx.Proof
	temp := make(map[common.Hash]interface{})
	for _, desc := range txProof.GetInputCoins() {
		hash := common.HashH(desc.CoinDetails.GetSerialNumber().ToBytesS())
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
	db database.DatabaseInterface,
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
	txProof := *tx.Proof
	for i := 0; tx.Proof != nil && i < len(txProof.GetInputCoins()); i++ {
		serialNumber := txProof.GetInputCoins()[i].CoinDetails.GetSerialNumber().ToBytesS()
		ok, err := db.HasSerialNumber(*prvCoinID, serialNumber, shardID)
		if ok || err != nil {
			return errors.New("double spend")
		}
	}
	return nil
}

func (tx Tx) ValidateTxWithBlockChain(
	bcr metadata.BlockchainRetriever,
	shardID byte,
	db database.DatabaseInterface,
) error {
	if tx.GetType() == common.TxRewardType || tx.GetType() == common.TxReturnStakingType {
		return nil
	}
	if tx.Metadata != nil {
		isContinued, err := tx.Metadata.ValidateTxWithBlockChain(&tx, bcr, shardID, db)
		if err != nil {
			Logger.log.Errorf("[db] validate metadata with blockchain: %d %s %t %v\n", tx.GetMetadataType(), tx.Hash().String(), isContinued, err)
			return NewTransactionErr(RejectTxMedataWithBlockChain, fmt.Errorf("validate metadata of tx %s with blockchain error %+v", tx.Hash().String(), err))
		}
		if !isContinued {
			return nil
		}
	}
	return tx.ValidateDoubleSpendWithBlockchain(bcr, shardID, db, nil)
}

func (tx Tx) validateNormalTxSanityData() (bool, error) {
	//check version
	if tx.Version > txVersion {
		return false, NewTransactionErr(RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version >= %d", tx.Version, txVersion))
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
	validateSanityOfProof, err := tx.validateSanityDataOfProof()
	if err != nil || !validateSanityOfProof {
		return false, err
	}

	if len(tx.SigPubKey) != common.SigPubKeySize {
		return false, NewTransactionErr(RejectTxPublickeySigSize, fmt.Errorf("wrong tx Sig PK size %d", len(tx.SigPubKey)))
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

	return true, nil
}

func (txN Tx) validateSanityDataOfProof() (bool, error) {
	if txN.Proof == nil {
		return true, nil
	}
	var txProofVersion privacy.Proof
	if (*txN.Proof).GetVersion() == 1 {
		txProofVersion = new(privacy.ProofV1)
	} else {
		txProofVersion = new(privacy.ProofV2)
	}

	if txN.Proof != nil {
		txProof := *txN.Proof
		if len(txProof.GetInputCoins()) > 255 {
			return false, errors.New("Input coins in tx are very large:" + strconv.Itoa(len(txProof.GetInputCoins())))
		}

		if len(txProof.GetOutputCoins()) > 255 {
			return false, errors.New("Output coins in tx are very large:" + strconv.Itoa(len(txProof.GetOutputCoins())))
		}

		isPrivacy := true
		// check Privacy or not

		if txProof.GetAggregatedRangeProof() == nil || len(txProof.GetOneOfManyProof()) == 0 || len(txN.Proof.GetSerialNumberProof()) == 0 {
			isPrivacy = false
		}

		if isPrivacy {
			if !txN.Proof.GetAggregatedRangeProof().ValidateSanity() {
				return false, errors.New("validate sanity Aggregated range proof failed")
			}

			for i := 0; i < len(txN.Proof.GetOneOfManyProof()); i++ {
				if !txN.Proof.GetOneOfManyProof()[i].ValidateSanity() {
					return false, errors.New("validate sanity One out of many proof failed")
				}
			}
			for i := 0; i < len(txN.Proof.GetSerialNumberProof()); i++ {
				if !txN.Proof.GetSerialNumberProof()[i].ValidateSanity() {
					return false, errors.New("validate sanity Serial number proof failed")
				}
			}

			// check input coins with privacy
			for i := 0; i < len(txN.Proof.GetInputCoins()); i++ {
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetSerialNumber().PointValid() {
					return false, errors.New("validate sanity Serial number of input coin failed")
				}
			}
			// check output coins with privacy
			for i := 0; i < len(txN.Proof.GetOutputCoins()); i++ {
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetPublicKey().PointValid() {
					return false, errors.New("validate sanity Public key of output coin failed")
				}
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetCoinCommitment().PointValid() {
					return false, errors.New("validate sanity Coin commitment of output coin failed")
				}
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetSNDerivator().ScalarValid() {
					return false, errors.New("validate sanity SNDerivator of output coin failed")
				}
			}
			// check ComInputSK
			if !txN.Proof.GetCommitmentInputSecretKey().PointValid() {
				return false, errors.New("validate sanity ComInputSK of proof failed")
			}
			// check ComInputValue
			for i := 0; i < len(txN.Proof.GetCommitmentInputValue()); i++ {
				if !txN.Proof.GetCommitmentInputValue()[i].PointValid() {
					return false, errors.New("validate sanity ComInputValue of proof failed")
				}
			}
			//check ComInputSND
			for i := 0; i < len(txN.Proof.GetCommitmentInputSND()); i++ {
				if !txN.Proof.GetCommitmentInputSND()[i].PointValid() {
					return false, errors.New("validate sanity ComInputSND of proof failed")
				}
			}
			//check ComInputShardID
			if !txN.Proof.GetCommitmentInputShardID().PointValid() {
				return false, errors.New("validate sanity ComInputShardID of proof failed")
			}

			// check ComOutputShardID
			for i := 0; i < len(txN.Proof.GetCommitmentOutputShardID()); i++ {
				if !txN.Proof.GetCommitmentOutputShardID()[i].PointValid() {
					return false, errors.New("validate sanity ComOutputShardID of proof failed")
				}
			}
			//check ComOutputSND
			for i := 0; i < len(txN.Proof.GetCommitmentOutputShardID()); i++ {
				if !txN.Proof.GetCommitmentOutputShardID()[i].PointValid() {
					return false, errors.New("validate sanity ComOutputSND of proof failed")
				}
			}
			//check ComOutputValue
			for i := 0; i < len(txN.Proof.GetCommitmentOutputValue()); i++ {
				if !txN.Proof.GetCommitmentOutputValue()[i].PointValid() {
					return false, errors.New("validate sanity ComOutputValue of proof failed")
				}
			}
			if len(txN.Proof.GetCommitmentIndices()) != len(txN.Proof.GetInputCoins())*privacy.CommitmentRingSize {
				return false, errors.New("validate sanity CommitmentIndices of proof failed")

			}
		}

		if !isPrivacy {
			for i := 0; i < len(txN.Proof.GetSerialNumberNoPrivacyProof()); i++ {
				if !txN.Proof.GetSerialNumberNoPrivacyProof()[i].ValidateSanity() {
					return false, errors.New("validate sanity Serial number no privacy proof failed")
				}
			}
			// check input coins without privacy
			for i := 0; i < len(txN.Proof.GetInputCoins()); i++ {
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetCoinCommitment().PointValid() {
					return false, errors.New("validate sanity CoinCommitment of input coin failed")
				}
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetPublicKey().PointValid() {
					return false, errors.New("validate sanity PublicKey of input coin failed")
				}
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetSerialNumber().PointValid() {
					return false, errors.New("validate sanity Serial number of input coin failed")
				}
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetRandomness().ScalarValid() {
					return false, errors.New("validate sanity Randomness of input coin failed")
				}
				if !txN.Proof.GetInputCoins()[i].CoinDetails.GetSNDerivator().ScalarValid() {
					return false, errors.New("validate sanity SNDerivator of input coin failed")
				}

			}

			// check output coins without privacy
			for i := 0; i < len(txN.Proof.GetOutputCoins()); i++ {
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetCoinCommitment().PointValid() {
					return false, errors.New("validate sanity CoinCommitment of output coin failed")
				}
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetPublicKey().PointValid() {
					return false, errors.New("validate sanity PublicKey of output coin failed")
				}
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetRandomness().ScalarValid() {
					return false, errors.New("validate sanity Randomness of output coin failed")
				}
				if !txN.Proof.GetOutputCoins()[i].CoinDetails.GetSNDerivator().ScalarValid() {
					return false, errors.New("validate sanity SNDerivator of output coin failed")
				}
			}
		}
	}
	return true, nil
}

func (tx Tx) ValidateSanityData(bcr metadata.BlockchainRetriever, beaconHeight uint64) (bool, error) {
	Logger.log.Debugf("\n\n\n START Validating sanity data of metadata %+v\n\n\n", tx.Metadata)
	if tx.Metadata != nil {
		Logger.log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := tx.Metadata.ValidateSanityData(bcr, &tx, beaconHeight)
		Logger.log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	Logger.log.Debugf("\n\n\n END sanity data of metadata%+v\n\n\n")
	return tx.validateNormalTxSanityData()
}

func (tx Tx) ValidateTxByItself(
	hasPrivacy bool,
	db database.DatabaseInterface,
	bcr metadata.BlockchainRetriever,
	shardID byte,
	isNewTransaction bool,
) (bool, error) {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, err
	}
	ok, err := tx.ValidateTransaction(hasPrivacy, db, shardID, prvCoinID, false, isNewTransaction)
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

func (tx Tx) GetProof() *privacy.Proof {
	return tx.Proof
}

func (tx Tx) IsPrivacy() bool {
	if tx.Proof == nil || len(tx.Proof.GetOneOfManyProof()) == 0 {
		return false
	}
	return true
}

func (tx Tx) ValidateType() bool {
	return tx.Type == common.TxNormalType || tx.Type == common.TxRewardType || tx.Type == common.TxReturnStakingType
}

func (tx Tx) IsCoinsBurning(bcr metadata.BlockchainRetriever, beaconHeight uint64) bool {
	if tx.Proof == nil || len(tx.Proof.GetOutputCoins()) == 0 {
		return false
	}
	senderPKBytes := []byte{}
	if len(tx.Proof.GetInputCoins()) > 0 {
		senderPKBytes = tx.Proof.GetInputCoins()[0].CoinDetails.GetPublicKey().ToBytesS()
	}
	//get burning address
	burningAddress := bcr.GetBurningAddress(beaconHeight)
	keyWalletBurningAccount, err := wallet.Base58CheckDeserialize(burningAddress)
	if err != nil {
		return false
	}
	keysetBurningAccount := keyWalletBurningAccount.KeySet
	paymentAddressBurningAccount := keysetBurningAccount.PaymentAddress
	for _, outCoin := range tx.Proof.GetOutputCoins() {
		outPKBytes := outCoin.CoinDetails.GetPublicKey().ToBytesS()
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
	if tx.Proof.GetOutputCoins() == nil || len(tx.Proof.GetOutputCoins()) == 0 {
		return 0
	}
	if tx.Proof.GetInputCoins() == nil || len(tx.Proof.GetInputCoins()) == 0 { // coinbase tx
		txValue := uint64(0)
		for _, outCoin := range tx.Proof.GetOutputCoins() {
			txValue += outCoin.CoinDetails.GetValue()
		}
		return txValue
	}

	senderPKBytes := tx.Proof.GetInputCoins()[0].CoinDetails.GetPublicKey().ToBytesS()
	txValue := uint64(0)
	for _, outCoin := range tx.Proof.GetOutputCoins() {
		outPKBytes := outCoin.CoinDetails.GetPublicKey().ToBytesS()
		if bytes.Equal(senderPKBytes, outPKBytes) {
			continue
		}
		txValue += outCoin.CoinDetails.GetValue()
	}
	return txValue
}

// InitTxSalary
// Blockchain use this tx to pay a reward(salary) to miner of chain
// #1 - salary:
// #2 - receiverAddr:
// #3 - privKey:
// #4 - snDerivators:
func (tx *Tx) InitTxSalary(
	salary uint64,
	receiverAddr *privacy.PaymentAddress,
	privKey *privacy.PrivateKey,
	db database.DatabaseInterface,
	metaData metadata.Metadata,
) error {
	tx.Version = txVersion
	tx.Type = common.TxRewardType

	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}

	var err error
	// create new output coins with info: Pk, value, input, randomness, last byte pk, coin commitment
	tx.Proof = new(zkp.PaymentProof)
	tempOutputCoin := make([]*privacy.OutputCoin, 1)
	tempOutputCoin[0] = new(privacy.OutputCoin).Init()
	publicKey, err := new(privacy.Point).FromBytesS(receiverAddr.Pk)
	if err != nil {
		return err
	}
	tempOutputCoin[0].CoinDetails.SetPublicKey(publicKey)
	tempOutputCoin[0].CoinDetails.SetValue(salary)
	tempOutputCoin[0].CoinDetails.SetRandomness(privacy.RandomScalar())

	sndOut := privacy.RandomScalar()
	for {
		tokenID := &common.Hash{}
		err := tokenID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return NewTransactionErr(TokenIDInvalidError, err, tokenID.String())
		}
		ok, err := CheckSNDerivatorExistence(tokenID, sndOut, db)
		if err != nil {
			return NewTransactionErr(SndExistedError, err)
		}
		if ok {
			sndOut = privacy.RandomScalar()
		} else {
			break
		}
	}
	tempOutputCoin[0].CoinDetails.SetSNDerivator(sndOut)
	// create coin commitment
	err = tempOutputCoin[0].CoinDetails.CommitAll()
	if err != nil {
		return NewTransactionErr(CommitOutputCoinError, err)
	}
	tx.Proof.SetOutputCoins(tempOutputCoin)

	// get last byte
	tx.PubKeyLastByteSender = receiverAddr.Pk[len(receiverAddr.Pk)-1]

	// sign Tx
	tx.SigPubKey = receiverAddr.Pk
	tx.sigPrivKey = *privKey
	tx.SetMetadata(metaData)
	err = tx.signTx()
	if err != nil {
		return NewTransactionErr(SignTxError, err)
	}

	return nil
}

func (tx Tx) ValidateTxReturnStaking(db database.DatabaseInterface,
) bool {
	return true
}

func (tx Tx) ValidateTxSalary(
	db database.DatabaseInterface,
) (bool, error) {
	// verify signature
	valid, err := tx.verifySigTx()
	if !valid {
		if err != nil {
			Logger.log.Debugf("Error verifying signature of tx: %+v", err)
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
	if ok, err := CheckSNDerivatorExistence(tokenID, tx.Proof.GetOutputCoins()[0].CoinDetails.GetSNDerivator(), db); ok || err != nil {
		return false, err
	}

	// check output coin's coin commitment is calculated correctly
	coin := tx.Proof.GetOutputCoins()[0].CoinDetails
	shardID2 := common.GetShardIDFromLastByte(coin.GetPubKeyLastByte())
	cmTmp2 := new(privacy.Point)
	cmTmp2.Add(coin.GetPublicKey(), new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenValueIndex], new(privacy.Scalar).FromUint64(uint64(coin.GetValue()))))
	cmTmp2.Add(cmTmp2, new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenSndIndex], coin.GetSNDerivator()))
	cmTmp2.Add(cmTmp2, new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenShardIDIndex], new(privacy.Scalar).FromUint64(uint64(shardID2))))
	cmTmp2.Add(cmTmp2, new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenRandomnessIndex], coin.GetRandomness()))

	ok := privacy.IsPointEqual(cmTmp2, tx.Proof.GetOutputCoins()[0].CoinDetails.GetCoinCommitment())
	if !ok {
		return ok, NewTransactionErr(UnexpectedError, errors.New("check output coin's coin commitment isn't calculated correctly"))
	}
	return ok, nil
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
	if tx.Proof != nil && len(tx.Proof.GetInputCoins()) == 0 && len(tx.Proof.GetOutputCoins()) > 0 { // coinbase tx
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
			Logger.log.Error(err)
			return false, NewTransactionErr(VerifyMinerCreatedTxBeforeGettingInBlockError, err)
		}
		return ok, nil
	}
	return true, nil
}
