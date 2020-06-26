package transaction

import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"github.com/incognitochain/incognito-chain/wallet"
	"math"
	"math/big"
	"math/rand"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/utils"
)

type RandomCommitmentsProcessParam struct {
	usableInputCoins []coin.PlainCoin
	randNum          int
	stateDB          *statedb.StateDB
	shardID          byte
	tokenID          *common.Hash
}

func NewRandomCommitmentsProcessParam(usableInputCoins []coin.PlainCoin, randNum int, stateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) *RandomCommitmentsProcessParam {
	return &RandomCommitmentsProcessParam{
		tokenID:          tokenID,
		shardID:          shardID,
		stateDB:          stateDB,
		randNum:          randNum,
		usableInputCoins: usableInputCoins,
	}
}

// RandomCommitmentsProcess - process list commitments and useable tx to create
// a list commitment random which be used to create a proof for new tx
// result contains
// commitmentIndexs = [{1,2,3,4,myindex1,6,7,8}{9,10,11,12,13,myindex2,15,16}...]
// myCommitmentIndexs = [4, 13, ...]
func RandomCommitmentsProcess(param *RandomCommitmentsProcessParam) (commitmentIndexs []uint64, myCommitmentIndexs []uint64, commitments [][]byte) {
	if len(param.usableInputCoins) == 0 {
		return
	}
	if param.randNum == 0 {
		param.randNum = privacy.CommitmentRingSize // default
	}
	// loop to create list usable commitments from usableInputCoins
	listUsableCommitments := make(map[common.Hash][]byte)
	listUsableCommitmentsIndices := make([]common.Hash, len(param.usableInputCoins))
	// tick index of each usable commitment with full db commitments
	mapIndexCommitmentsInUsableTx := make(map[string]*big.Int)
	for i, in := range param.usableInputCoins {
		usableCommitment := in.GetCommitment().ToBytesS()
		commitmentInHash := common.HashH(usableCommitment)
		listUsableCommitments[commitmentInHash] = usableCommitment
		listUsableCommitmentsIndices[i] = commitmentInHash
		index, err := txDatabaseWrapper.getCommitmentIndex(param.stateDB, *param.tokenID, usableCommitment, param.shardID)
		if err != nil {
			Logger.Log.Error(err)
			return
		}
		commitmentInBase58Check := base58.Base58Check{}.Encode(usableCommitment, common.ZeroByte)
		mapIndexCommitmentsInUsableTx[commitmentInBase58Check] = index
	}
	// loop to random commitmentIndexs
	cpRandNum := (len(listUsableCommitments) * param.randNum) - len(listUsableCommitments)
	//fmt.Printf("cpRandNum: %d\n", cpRandNum)
	lenCommitment, err1 := txDatabaseWrapper.getCommitmentLength(param.stateDB, *param.tokenID, param.shardID)
	if err1 != nil {
		Logger.Log.Error(err1)
		return
	}
	if lenCommitment == nil {
		Logger.Log.Error(errors.New("Commitments is empty"))
		return
	}
	if lenCommitment.Uint64() == 1 && len(param.usableInputCoins) == 1 {
		commitmentIndexs = []uint64{0, 0, 0, 0, 0, 0, 0}
		temp := param.usableInputCoins[0].GetCommitment().ToBytesS()
		commitments = [][]byte{temp, temp, temp, temp, temp, temp, temp}
	} else {
		for i := 0; i < cpRandNum; i++ {
			for {
				lenCommitment, _ = txDatabaseWrapper.getCommitmentLength(param.stateDB, *param.tokenID, param.shardID)
				index, _ := common.RandBigIntMaxRange(lenCommitment)
				ok, err := txDatabaseWrapper.hasCommitmentIndex(param.stateDB, *param.tokenID, index.Uint64(), param.shardID)
				if ok && err == nil {
					temp, _ := txDatabaseWrapper.getCommitmentByIndex(param.stateDB, *param.tokenID, index.Uint64(), param.shardID)
					if _, found := listUsableCommitments[common.HashH(temp)]; !found {
						// random commitment not in commitments of usableinputcoin
						commitmentIndexs = append(commitmentIndexs, index.Uint64())
						commitments = append(commitments, temp)
						break
					}
				} else {
					continue
				}
			}
		}
	}
	// loop to insert usable commitments into commitmentIndexs for every group
	j := 0
	for _, commitmentInHash := range listUsableCommitmentsIndices {
		commitmentValue := listUsableCommitments[commitmentInHash]
		index := mapIndexCommitmentsInUsableTx[base58.Base58Check{}.Encode(commitmentValue, common.ZeroByte)]
		randInt := rand.Intn(param.randNum)
		i := (j * param.randNum) + randInt
		commitmentIndexs = append(commitmentIndexs[:i], append([]uint64{index.Uint64()}, commitmentIndexs[i:]...)...)
		commitments = append(commitments[:i], append([][]byte{commitmentValue}, commitments[i:]...)...)
		myCommitmentIndexs = append(myCommitmentIndexs, uint64(i)) // create myCommitmentIndexs
		j++
	}
	return commitmentIndexs, myCommitmentIndexs, commitments
}

type EstimateTxSizeParam struct {
	numInputCoins            int
	numPayments              int
	hasPrivacy               bool
	metadata                 metadata.Metadata
	privacyCustomTokenParams *CustomTokenPrivacyParamTx
	limitFee                 uint64
}

func NewEstimateTxSizeParam(numInputCoins, numPayments int,
	hasPrivacy bool, metadata metadata.Metadata,
	privacyCustomTokenParams *CustomTokenPrivacyParamTx,
	limitFee uint64) *EstimateTxSizeParam {
	estimateTxSizeParam := &EstimateTxSizeParam{
		numInputCoins:            numInputCoins,
		numPayments:              numPayments,
		hasPrivacy:               hasPrivacy,
		limitFee:                 limitFee,
		metadata:                 metadata,
		privacyCustomTokenParams: privacyCustomTokenParams,
	}
	return estimateTxSizeParam
}

// EstimateTxSize returns the estimated size of the tx in kilobyte
func EstimateTxSize(estimateTxSizeParam *EstimateTxSizeParam) uint64 {
	sizeVersion := uint64(1)  // int8
	sizeType := uint64(5)     // string, max : 5
	sizeLockTime := uint64(8) // int64
	sizeFee := uint64(8)      // uint64
	sizeInfo := uint64(512)

	sizeSigPubKey := uint64(common.SigPubKeySize)
	sizeSig := uint64(common.SigNoPrivacySize)
	if estimateTxSizeParam.hasPrivacy {
		sizeSig = uint64(common.SigPrivacySize)
	}

	sizeProof := uint64(0)
	if estimateTxSizeParam.numInputCoins != 0 || estimateTxSizeParam.numPayments != 0 {
		sizeProof = utils.EstimateProofSize(estimateTxSizeParam.numInputCoins, estimateTxSizeParam.numPayments, estimateTxSizeParam.hasPrivacy)
	} else {
		if estimateTxSizeParam.limitFee > 0 {
			sizeProof = utils.EstimateProofSize(1, 1, estimateTxSizeParam.hasPrivacy)
		}
	}

	sizePubKeyLastByte := uint64(1)

	sizeMetadata := uint64(0)
	if estimateTxSizeParam.metadata != nil {
		sizeMetadata += estimateTxSizeParam.metadata.CalculateSize()
	}

	sizeTx := sizeVersion + sizeType + sizeLockTime + sizeFee + sizeInfo + sizeSigPubKey + sizeSig + sizeProof + sizePubKeyLastByte + sizeMetadata

	// size of privacy custom token  data
	if estimateTxSizeParam.privacyCustomTokenParams != nil {
		customTokenDataSize := uint64(0)

		customTokenDataSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertyID))
		customTokenDataSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertySymbol))
		customTokenDataSize += uint64(len(estimateTxSizeParam.privacyCustomTokenParams.PropertyName))

		customTokenDataSize += 8 // for amount
		customTokenDataSize += 4 // for TokenTxType

		customTokenDataSize += uint64(1) // int8 version
		customTokenDataSize += uint64(5) // string, max : 5 type
		customTokenDataSize += uint64(8) // int64 locktime
		customTokenDataSize += uint64(8) // uint64 fee

		customTokenDataSize += uint64(64) // info

		customTokenDataSize += uint64(common.SigPubKeySize)  // sig pubkey
		customTokenDataSize += uint64(common.SigPrivacySize) // sig

		// Proof
		customTokenDataSize += utils.EstimateProofSize(len(estimateTxSizeParam.privacyCustomTokenParams.TokenInput), len(estimateTxSizeParam.privacyCustomTokenParams.Receiver), true)

		customTokenDataSize += uint64(1) //PubKeyLastByte

		sizeTx += customTokenDataSize
	}

	return uint64(math.Ceil(float64(sizeTx) / 1024))
}

type BuildCoinBaseTxByCoinIDParams struct {
	payToAddress       *privacy.PaymentAddress
	amount             uint64
	txRandom           *coin.TxRandom
	payByPrivateKey    *privacy.PrivateKey
	transactionStateDB *statedb.StateDB
	bridgeStateDB      *statedb.StateDB
	meta               *metadata.WithDrawRewardResponse
	coinID             common.Hash
	txType             int
	coinName           string
	shardID            byte
}

func NewBuildCoinBaseTxByCoinIDParams(payToAddress *privacy.PaymentAddress,
	amount uint64,
	payByPrivateKey *privacy.PrivateKey,
	stateDB *statedb.StateDB,
	meta *metadata.WithDrawRewardResponse,
	coinID common.Hash,
	txType int,
	coinName string,
	shardID byte,
	bridgeStateDB *statedb.StateDB) *BuildCoinBaseTxByCoinIDParams {
	params := &BuildCoinBaseTxByCoinIDParams{
		transactionStateDB: stateDB,
		bridgeStateDB:      bridgeStateDB,
		shardID:            shardID,
		meta:               meta,
		amount:             amount,
		coinID:             coinID,
		coinName:           coinName,
		payByPrivateKey:    payByPrivateKey,
		payToAddress:       payToAddress,
		txType:             txType,
	}
	return params
}

func calculateSumOutputsWithFee(outputCoins []coin.Coin, fee uint64) *operation.Point {
	sumOutputsWithFee := new(operation.Point).Identity()
	for i := 0; i < len(outputCoins); i += 1 {
		sumOutputsWithFee.Add(sumOutputsWithFee, outputCoins[i].GetCommitment())
	}
	feeCommitment := new(operation.Point).ScalarMult(
		operation.PedCom.G[operation.PedersenValueIndex],
		new(operation.Scalar).FromUint64(fee),
	)
	sumOutputsWithFee.Add(sumOutputsWithFee, feeCommitment)
	return sumOutputsWithFee
}

func newCoinUniqueOTABasedOnPaymentInfo(paymentInfo *privacy.PaymentInfo, tokenID *common.Hash, stateDB *statedb.StateDB) (*coin.CoinV2, error) {
	for {
		c, err := coin.NewCoinFromPaymentInfo(paymentInfo)
		if err != nil {
			Logger.Log.Errorf("Cannot parse coin based on payment info err: %v", err)
			return nil, err
		}
		// If previously created coin is burning address
		if wallet.IsPublicKeyBurningAddress(c.GetPublicKey().ToBytesS()) {
			return c, nil // No need to check db
		}
		// Onetimeaddress should be unique
		publicKeyBytes := c.GetPublicKey().ToBytesS()
		found, err := txDatabaseWrapper.hasOnetimeAddress(stateDB, *tokenID, publicKeyBytes)
		if err != nil {
			Logger.Log.Errorf("Cannot check public key existence in DB, err %v", err)
			return nil, err
		}
		if !found {
			return c, nil
		}
	}
}

func newCoinV2ArrayFromPaymentInfoArray(paymentInfo []*privacy.PaymentInfo, tokenID *common.Hash, stateDB *statedb.StateDB) ([]*coin.CoinV2, error) {
	outputCoins := make([]*coin.CoinV2, len(paymentInfo))
	for index, info := range paymentInfo {
		var err error
		outputCoins[index], err = newCoinUniqueOTABasedOnPaymentInfo(info, tokenID, stateDB)
		if err != nil {
			Logger.Log.Errorf("Cannot create coin with unique OTA, error: %v", err)
			return nil, err
		}
	}
	return outputCoins, nil
}

func BuildCoinBaseTxByCoinID(params *BuildCoinBaseTxByCoinIDParams) (metadata.Transaction, error) {
	otaCoin, err := coin.NewCoinFromAmountAndReceiver(params.amount, *params.payToAddress)
	params.meta.SetSharedRandom(otaCoin.GetSharedRandom().ToBytesS())

	if err != nil {
		Logger.Log.Errorf("Cannot get new coin from amount and receiver")
		return nil, err
	}
	switch params.txType {
	case NormalCoinType:
		tx := new(TxVersion2)
		err = tx.InitTxSalary(otaCoin, params.payByPrivateKey, params.transactionStateDB, params.meta)
		return tx, err
	case CustomTokenPrivacyType:
		var propertyID [common.HashSize]byte
		copy(propertyID[:], params.coinID[:])
		propID := common.Hash(propertyID)
		tx := new(TxTokenVersion2)
		err = tx.InitTxTokenSalary(otaCoin, params.payByPrivateKey, params.transactionStateDB,
			params.meta, &propID, params.coinName)
		return tx, err
	}
	return nil, nil
}

func validateTxParams(params *TxPrivacyInitParams) error {
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

func parseTokenID(tokenID *common.Hash) (*common.Hash, error) {
	if tokenID == nil {
		result := new(common.Hash)
		err := result.SetBytes(common.PRVCoinID[:])
		if err != nil {
			Logger.Log.Error(err)
			return nil, NewTransactionErr(TokenIDInvalidError, err, tokenID.String())
		}
		return result, nil
	}
	return tokenID, nil
}

func verifySigNoPrivacy(sig []byte, sigPubKey []byte, hashedMessage []byte) (bool, error) {
	// check input transaction
	if sig == nil || sigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be an signed one"))
	}

	var err error
	/****** verify Schnorr signature *****/
	// prepare Public key for verification
	verifyKey := new(privacy.SchnorrPublicKey)
	sigPublicKey, err := new(operation.Point).FromBytesS(sigPubKey)

	if err != nil {
		Logger.Log.Error(err)
		return false, NewTransactionErr(DecompressSigPubKeyError, err)
	}
	verifyKey.Set(sigPublicKey)

	// convert signature from byte array to SchnorrSign
	signature := new(privacy.SchnSignature)
	err = signature.SetBytes(sig)
	if err != nil {
		Logger.Log.Error(err)
		return false, NewTransactionErr(InitTxSignatureFromBytesError, err)
	}

	// verify signature
	/*Logger.log.Debugf(" VERIFY SIGNATURE ----------- HASH: %v\n", tx.Hash()[:])
	if tx.Proof != nil {
		Logger.log.Debugf(" VERIFY SIGNATURE ----------- TX Proof bytes before verifing the signature: %v\n", tx.Proof.Bytes())
	}
	Logger.log.Debugf(" VERIFY SIGNATURE ----------- TX meta: %v\n", tx.Metadata)*/
	res := verifyKey.Verify(signature, hashedMessage)
	return res, nil
}

func signNoPrivacy(privKey *privacy.PrivateKey, hashedMessage []byte) (signatureBytes []byte, sigPubKey []byte, err error) {
	/****** using Schnorr signature *******/
	// sign with sigPrivKey
	// prepare private key for Schnorr
	sk := new(operation.Scalar).FromBytesS(*privKey)
	r := new(operation.Scalar).FromUint64(0)
	sigKey := new(schnorr.SchnorrPrivateKey)
	sigKey.Set(sk, r)
	signature, err := sigKey.Sign(hashedMessage)
	if err != nil {
		return nil, nil, err
	}

	signatureBytes = signature.Bytes()
	sigPubKey = sigKey.GetPublicKey().GetPublicKey().ToBytesS()
	return signatureBytes, sigPubKey, nil
}

// Used to parse json
type txJsonDataVersion struct {
	Version int8 `json:"Version"`
}

func NewTransactionFromJsonBytes(data []byte) (metadata.Transaction, error) {
	txJsonVersion := new(txJsonDataVersion)
	if err := json.Unmarshal(data, txJsonVersion); err != nil {
		return nil, err
	}
	switch txJsonVersion.Version {
	case int8(TxVersion1Number), int8(TxVersion0Number):
		tx := new(TxVersion1)
		if err := json.Unmarshal(data, tx); err != nil {
			return nil, err
		}
		return tx, nil
	case int8(TxVersion2Number), int8(TxConversionVersion12Number):
		tx := new(TxVersion2)
		if err := json.Unmarshal(data, tx); err != nil {
			return nil, err
		}
		return tx, nil
	default:
		return nil, errors.New("Cannot new transaction from json, version is wrong")
	}
}

type txTokenJsonDataVersion struct {
	Version int8 `json:"Version"`
}

func NewTransactionTokenFromJsonBytes(data []byte) (TxTokenInterface, error) {
	txJsonVersion := new(txJsonDataVersion)
	if err := json.Unmarshal(data, txJsonVersion); err != nil {
		return nil, err
	}
	if txJsonVersion.Version == TxVersion1Number {
		tx := new(TxTokenVersion1)
		if err := json.Unmarshal(data, tx); err != nil {
			return nil, err
		}
		return tx, nil
	} else if txJsonVersion.Version == TxVersion2Number || txJsonVersion.Version == TxConversionVersion12Number {
		tx := new(TxTokenVersion2)
		if err := json.Unmarshal(data, tx); err != nil {
			return nil, err
		}
		return tx, nil
	}
	return nil, errors.New("Cannot new transaction token from json, version is not 1 or 2 or -1")
}
