package transaction

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/wallet"
	"math"
	"sort"
	"strconv"
	"time"
)

type TxBase struct {
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
	cachedActualSize *uint64      // cached actualsize data for tx
}

// Function that choose which version to create metadata Transaction
func NewTxPrivacyFromParams(params *TxPrivacyInitParams) (metadata.Transaction, error) {
	version, err := getTxVersionFromCoins(params.inputCoins);
	if err != nil {
		Logger.Log.Errorf("Cannot get version from params")
		return nil, err
	}
	if version == txVersion1Number {
		return &TxVersion1{}, nil
	} else if version == txVersion2Number {
		return &TxVersion2{}, nil
	}
	return nil, errors.New("Version is not 1 or 2, cannot NewTxPrivacyFromParams")
}

func NewTxPrivacyFromVersionNumber(version int8) (metadata.Transaction, error) {
	if version == txVersion1Number {
		return &TxVersion1{}, nil
	} else if version == txVersion2Number {
		return &TxVersion2{}, nil
	}
	return nil, errors.New("Version is not 1 or 2, cannot NewTxPrivacyFromParams")
}

// This function copies values from TxBase to metadata.Transaction
// It does not copy sigPrivKey because it is private field
func NewTransactionFromTxBase(tx TxBase) (metadata.Transaction, error) {
	metaTx, err := NewTxPrivacyFromVersionNumber(tx.GetVersion())
	if err != nil {
		return nil, err
	}
	metaTx.SetVersion(tx.GetVersion())
	metaTx.SetType(tx.GetType())
	metaTx.SetLockTime(tx.GetLockTime())
	metaTx.SetTxFee(tx.GetTxFee())
	metaTx.SetInfo(tx.GetInfo())
	metaTx.SetSigPubKey(tx.GetSigPubKey())
	metaTx.SetSig(tx.GetSig())
	metaTx.SetProof(tx.GetProof())
	metaTx.SetGetSenderAddrLastByte(tx.GetSenderAddrLastByte())
	metaTx.SetMetadata(tx.GetMetadata())
	return metaTx, nil
}

// This function copies values from metadata.Transaction to TxBase
// It does not copy sigPrivKey because it is private field
func NewTxBaseFromMetadataTx(tx metadata.Transaction) TxBase {
	var txBase TxBase
	txBase.SetVersion(tx.GetVersion())
	txBase.SetType(tx.GetType())
	txBase.SetLockTime(tx.GetLockTime())
	txBase.SetTxFee(tx.GetTxFee())
	txBase.SetInfo(tx.GetInfo())
	txBase.SetSigPubKey(tx.GetSigPubKey())
	txBase.SetSig(tx.GetSig())
	txBase.SetProof(tx.GetProof())
	txBase.SetGetSenderAddrLastByte(tx.GetSenderAddrLastByte())
	txBase.SetMetadata(tx.GetMetadata())
	return txBase
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

// Don't change to pointer. Because if not splitting to function like this (put this in Init func), the performance still be the same. If change to pointer it could be wrong.
func getTxInfo(paramInfo []byte) ([]byte, error) {
	if lenTxInfo := len(paramInfo); lenTxInfo > MaxSizeInfo {
		return []byte{}, NewTransactionErr(ExceedSizeInfoTxError, nil)
	}
	return paramInfo, nil
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
func (tx *TxBase) initializeTxAndParams(params *TxPrivacyInitParams) error {
	var err error

	// Tx: initialize some values
	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}
	tx.Fee = params.fee
	tx.Type = common.TxNormalType
	tx.Metadata = params.metaData

	if tx.Version, err = getTxVersionFromCoins(params.inputCoins); err != nil {
		return err
	}
	if tx.Info, err = getTxInfo(params.info); err != nil {
		return err
	}
	if tx.PubKeyLastByteSender, err = parseLastByteSender(params.senderSK); err != nil {
		return err
	}
	// Params: update balance if overbalance
	if err = updateParamsWhenOverBalance(params); err != nil {
		return err
	}
	return nil
}

// =================== PARSING JSON FUNCTIONS ===================

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
	if ver == txVersion1Number {
		res = &zkp.PaymentProof{}
	} else if ver == txVersion2Number {
		res = &privacy_v2.PaymentProofV2{}
	} else if ver == txConversionVersion12Number {
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

func (tx *TxBase) UnmarshalJSON(data []byte) error {
	// For rolling version
	type Alias TxBase
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

// =================== GET/SET FUNCTIONS ===================

func (tx TxBase) GetVersion() int8 { return tx.Version }

func (tx *TxBase) SetVersion(version int8) { tx.Version = version }

func (tx TxBase) GetMetadataType() int {
	if tx.Metadata != nil {
		return tx.Metadata.GetType()
	}
	return metadata.InvalidMeta
}

func (tx TxBase) GetType() string { return tx.Type }

func (tx *TxBase) SetType(t string) { tx.Type = t }

func (tx TxBase) GetLockTime() int64 { return tx.LockTime }

func (tx *TxBase) SetLockTime(locktime int64) { tx.LockTime = locktime }

func (tx TxBase) GetSenderAddrLastByte() byte { return tx.PubKeyLastByteSender }

func (tx *TxBase) SetGetSenderAddrLastByte(b byte) { tx.PubKeyLastByteSender = b }

func (tx TxBase) GetTxFee() uint64 { return tx.Fee }

func (tx *TxBase) SetTxFee(fee uint64) { tx.Fee = fee }

func (tx TxBase) GetTxFeeToken() uint64 { return uint64(0) }

func (tx TxBase) GetInfo() []byte { return tx.Info }

func (tx *TxBase) SetInfo(info []byte) { tx.Info = info }

func (tx TxBase) GetSigPubKey() []byte { return tx.SigPubKey }

func (tx *TxBase) SetSigPubKey(sigPubkey []byte) { tx.SigPubKey = sigPubkey }

func (tx TxBase) GetSig() []byte { return tx.Sig }

func (tx *TxBase) SetSig(sig []byte) { tx.Sig = sig }

func (tx TxBase) GetProof() privacy.Proof { return tx.Proof }

func (tx *TxBase) SetProof(proof privacy.Proof) { tx.Proof = proof }

func (tx TxBase) GetTokenID() *common.Hash { return &common.PRVCoinID }

func (tx TxBase) GetMetadata() metadata.Metadata { return tx.Metadata }

func (tx *TxBase) SetMetadata(meta metadata.Metadata) { tx.Metadata = meta }

// =================== FUNCTIONS THAT GET STUFF AND REQUIRE SOME CODING ===================

func (tx TxBase) CheckAuthorizedSender(publicKey []byte) (bool, error) {
	transaction, err := NewTransactionFromTxBase(tx)
	if err != nil {
		Logger.Log.Errorf("Cannot create new transaction from txBase")
		return false, err
	}
	return transaction.CheckAuthorizedSender(publicKey)
}

func (tx TxBase) GetTxActualSize() uint64 {
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

func (tx TxBase) GetReceivers() ([][]byte, []uint64) {
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

func (tx TxBase) GetSender() []byte {
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

func (tx TxBase) GetTransferData() (bool, []byte, uint64, *common.Hash) {
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

func (tx TxBase) GetAndCheckBurningReceiver(retriever metadata.ChainRetriever, blockHeight uint64) (bool, []byte, uint64) {
	pubkeys, amounts := tx.GetReceivers()
	if len(pubkeys) > 2 {
		Logger.Log.Error("GetAndCheckBurning receiver: More than 2 receivers")
		return false, nil, 0
	}

	burnAccount, err := wallet.Base58CheckDeserialize(retriever.GetBurningAddress(blockHeight))
	if err != nil {
		return false, nil, 0
	}
	burnPaymentAddress := burnAccount.KeySet.PaymentAddress

	hasBurning, pubkey, amount := false, []byte{}, uint64(0)
	for i, pk := range pubkeys {
		if bytes.Equal(burnPaymentAddress.Pk, pk) {
			hasBurning = true
			pubkey = pk
			amount += amounts[i]
		}
	}
	return hasBurning, pubkey, amount
}

// implement this func if needed
func (tx TxBase) GetMetadataFromVinsTx(bcr metadata.ChainRetriever, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever) (metadata.Metadata, error) {
	// implement this func if needed
	return nil, nil
}

func (tx TxBase) ListSerialNumbersHashH() []common.Hash {
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

func (tx TxBase) String() string {
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

func (tx TxBase) Hash() *common.Hash {
	inBytes := []byte(tx.String())
	hash := common.HashH(inBytes)
	return &hash
}

func (tx TxBase) CalculateTxValue() uint64 {
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

// =================== FUNCTION THAT CHECK STUFFS  ===================

func (tx TxBase) CheckTxVersion(maxTxVersion int8) bool {
	return !(tx.Version > maxTxVersion)
}

func (tx *TxBase) isNonPrivacyNonInput(params *TxPrivacyInitParams) (bool, error) {
	var err error
	Logger.Log.Debugf("len(inputCoins), fee, hasPrivacy: %d, %d, %v\n", len(params.inputCoins), params.fee, params.hasPrivacy)
	if len(params.inputCoins) == 0 && params.fee == 0 && !params.hasPrivacy {
		Logger.Log.Debugf("len(inputCoins) == 0 && fee == 0 && !hasPrivacy\n")
		tx.sigPrivKey = *params.senderSK
		if tx.Sig, tx.SigPubKey, err = signNoPrivacy(params.senderSK, tx.Hash()[:]); err != nil {
			Logger.Log.Error(errors.New(fmt.Sprintf("Cannot sign tx %v\n", err)))
			return true, NewTransactionErr(SignTxError, err)
		}
		return true, nil
	}
	return false, nil
}

func (tx *TxBase) ShouldSignMetaData() bool {
	if tx.GetMetadata() == nil {
		return false
	}
	meta := tx.GetMetadata()
	return meta.ShouldSignMetaData()
}

func (tx TxBase) IsSalaryTx() bool {
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

func (tx TxBase) IsPrivacy() bool {
	// In the case of NonPrivacyNonInput, we do not have proof
	if tx.Proof == nil {
		return false
	}
	return tx.Proof.IsPrivacy()
}

func (tx TxBase) IsCoinsBurning(bcr metadata.ChainRetriever, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) bool {
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

// =================== FUNCTIONS THAT VALIDATE STUFFS ===================

func (tx TxBase) ValidateTxSalary(db *statedb.StateDB) (bool, error) {
	transaction, err := NewTransactionFromTxBase(tx)
	if err != nil {
		Logger.Log.Errorf("Cannot create new transaction from txBase")
		return false, err
	}
	return transaction.ValidateTxSalary(db)
}

func (tx TxBase) Verify(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	transaction, err := NewTransactionFromTxBase(tx)
	if err != nil {
		Logger.Log.Errorf("Cannot create new transaction from txBase")
		return false, err
	}
	return transaction.Verify(hasPrivacy, transactionStateDB, bridgeStateDB, shardID, tokenID, isBatch, isNewTransaction)
}

func (tx TxBase) ValidateTransaction(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	if tx.GetType() == common.TxRewardType {
		return tx.ValidateTxSalary(transactionStateDB)
	}
	if tx.GetType() == common.TxReturnStakingType {
		return tx.ValidateTxReturnStaking(transactionStateDB), nil
	}
	if tx.Version == txConversionVersion12Number {
		return validateConversionVer1ToVer2(&tx, transactionStateDB, shardID, tokenID)
	}
	transaction, err := NewTransactionFromTxBase(tx)
	if err != nil {
		Logger.Log.Errorf("Cannot create new transaction from txBase")
		return false, err
	}
	return transaction.Verify(hasPrivacy, transactionStateDB, bridgeStateDB, shardID, tokenID, isBatch, isNewTransaction)
}

func (tx TxBase) ValidateTxWithCurrentMempool(mr metadata.MempoolRetriever) error {
	if tx.Proof == nil {
		return nil
	}
	temp := make(map[common.Hash]interface{})
	for _, desc := range tx.Proof.GetInputCoins() {
		hash := common.HashH(desc.GetKeyImage().ToBytesS())
		temp[hash] = nil
	}
	poolSerialNumbersHashH := mr.GetSerialNumbersHashH()
	for _, listSerialNumbers := range poolSerialNumbersHashH {
		for _, serialNumberHash := range listSerialNumbers {
			if _, ok := temp[serialNumberHash]; ok {
				return errors.New("double spend")
			}
		}
	}
	return nil
}

func (tx TxBase) ValidateDoubleSpendWithBlockchain(shardID byte, stateDB *statedb.StateDB, tokenID *common.Hash) error {
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
		ok, err := txDatabaseWrapper.hasSerialNumber(stateDB, *prvCoinID, serialNumber, shardID)
		if ok || err != nil {
			return errors.New("double spend")
		}
	}
	return nil
}

func (tx TxBase) ValidateType() bool {
	return tx.Type == common.TxNormalType || tx.Type == common.TxRewardType || tx.Type == common.TxReturnStakingType
}

func (tx TxBase) ValidateTxReturnStaking(stateDB *statedb.StateDB) bool { return true }

func (tx TxBase) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []metadata.Transaction, txsUsed []int, insts [][]string, instsUsed []int, shardID byte, bcr metadata.ChainRetriever, accumulatedValues *metadata.AccumulatedValues, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever) (bool, error) {
	if tx.IsPrivacy() {
		return true, nil
	}
	proof := tx.GetProof()
	meta := tx.GetMetadata()

	inputCoins := make([]coin.PlainCoin, 0)
	outputCoins := make([]coin.Coin, 0)
	if tx.GetProof() != nil {
		inputCoins = tx.GetProof().GetInputCoins()
		outputCoins = tx.GetProof().GetOutputCoins()
	}
	if proof != nil && len(inputCoins) == 0 && len(outputCoins) > 0 { // coinbase tx
		if meta == nil {
			return false, nil
		}
		if !meta.IsMinerCreatedMetaType() {
			return false, nil
		}
	}
	if meta != nil {
		ok, err := meta.VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock, txsUsed, insts, instsUsed, shardID, &tx, bcr, accumulatedValues, nil, nil)
		if err != nil {
			Logger.Log.Error(err)
			return false, NewTransactionErr(VerifyMinerCreatedTxBeforeGettingInBlockError, err)
		}
		return ok, nil
	}
	return true, nil
}


func (tx TxBase) ValidateTxByItself(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, chainRetriever metadata.ChainRetriever, shardID byte, isNewTransaction bool, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever) (bool, error) {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, err
	}
	ok, err := tx.ValidateTransaction(hasPrivacy, transactionStateDB, bridgeStateDB, shardID, prvCoinID, false, isNewTransaction)
	if !ok {
		return false, err
	}
	meta := tx.GetMetadata()
	if meta != nil {
		if hasPrivacy {
			return false, errors.New("Metadata can not exist in not privacy tx")
		}
		validateMetadata := meta.ValidateMetadataByItself()
		if validateMetadata {
			return validateMetadata, nil
		} else {
			return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid"))
		}
	}
	return true, nil
}

func (tx TxBase) ValidateTxWithBlockChain(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error {
	if tx.GetType() == common.TxRewardType || tx.GetType() == common.TxReturnStakingType {
		return nil
	}
	meta := tx.GetMetadata()
	if meta != nil {
		metaTx, err := NewTransactionFromTxBase(tx)
		if err != nil {
			return err
		}
		isContinued, err := meta.ValidateTxWithBlockChain(metaTx, chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB)
		fmt.Printf("[transactionStateDB] validate metadata with blockchain: %d %h %t %v\n", tx.GetMetadataType(), tx.Hash(), isContinued, err)
		if err != nil {
			Logger.Log.Errorf("[db] validate metadata with blockchain: %d %s %t %v\n", tx.GetMetadataType(), tx.Hash().String(), isContinued, err)
			return NewTransactionErr(RejectTxMedataWithBlockChain, fmt.Errorf("validate metadata of tx %s with blockchain error %+v", tx.Hash().String(), err))
		}
		if !isContinued {
			return nil
		}
	}
	return tx.ValidateDoubleSpendWithBlockchain(shardID, stateDB, nil)
}

func (tx TxBase) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	meta := tx.GetMetadata()
	Logger.Log.Debugf("\n\n\n START Validating sanity data of metadata %+v\n\n\n", meta)
	if meta != nil {
		Logger.Log.Debug("tx.Metadata.ValidateSanityData")
		txMeta, err := NewTransactionFromTxBase(tx)
		if err != nil {
			return false, err
		}
		isContinued, ok, err := meta.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, txMeta)
		Logger.Log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	Logger.Log.Debugf("\n\n\n END sanity data of metadata%+v\n\n\n")
	//check version
	if tx.GetVersion() > txVersion2Number {
		return false, NewTransactionErr(RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version <= %d", tx.GetVersion(), currentTxVersion))
	}
	// check LockTime before now
	if int64(tx.GetLockTime()) > time.Now().Unix() {
		return false, NewTransactionErr(RejectInvalidLockTime, fmt.Errorf("wrong tx locktime %d", tx.GetLockTime()))
	}

	// check tx size
	actualTxSize := tx.GetTxActualSize()
	if actualTxSize > common.MaxTxSize {
		return false, NewTransactionErr(RejectTxSize, fmt.Errorf("tx size %d kB is too large", actualTxSize))
	}

	// check sanity of Proof
	if tx.GetProof() != nil {
		ok, err := tx.GetProof().ValidateSanity()
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
	switch tx.GetType() {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenPrivacyType, common.TxReturnStakingType: //is valid
	default:
		return false, NewTransactionErr(RejectTxType, fmt.Errorf("wrong tx type with %s", tx.GetType()))
	}

	//if txN.Type != common.TxNormalType && txN.Type != common.TxRewardType && txN.Type != common.TxCustomTokenType && txN.Type != common.TxCustomTokenPrivacyType { // only 1 byte
	//	return false, errors.New("wrong tx type")
	//}

	// check info field
	info := tx.GetInfo()
	if len(info) > 512 {
		return false, NewTransactionErr(RejectTxInfoSize, fmt.Errorf("wrong tx info length %d bytes, only support info with max length <= %d bytes", len(info), 512))
	}

	return true, nil
}

// ============= Init Transaction, the input should be params such as: TxPrivacyInitParams ============
func (tx *TxBase) Init(paramsInterface interface{}) error {
	transaction, err := NewTransactionFromTxBase(*tx)
	if err != nil {
		Logger.Log.Errorf("Cannot create new transaction from txBase")
		return err
	}
	return transaction.Init(paramsInterface)
}