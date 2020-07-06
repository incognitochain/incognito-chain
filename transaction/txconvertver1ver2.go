package transaction

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/serialnumbernoprivacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/privacy/coin"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

// ================ TX NORMAL CONVERSION =================

type TxConvertVer1ToVer2InitParams struct {
	senderSK    *privacy.PrivateKey
	paymentInfo []*privacy.PaymentInfo
	inputCoins  []coin.PlainCoin
	fee         uint64
	stateDB     *statedb.StateDB
	tokenID     *common.Hash // default is nil -> use for prv coin
	metaData    metadata.Metadata
	info        []byte // 512 bytes
}

func NewTxConvertVer1ToVer2InitParams(senderSK *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoins []coin.PlainCoin,
	fee uint64,
	stateDB *statedb.StateDB,
	tokenID *common.Hash, // default is nil -> use for prv coin
	metaData metadata.Metadata,
	info []byte) *TxConvertVer1ToVer2InitParams {

	return  &TxConvertVer1ToVer2InitParams{
		stateDB:     stateDB,
		tokenID:     tokenID,
		inputCoins:  inputCoins,
		fee:         fee,
		metaData:    metaData,
		paymentInfo: paymentInfo,
		senderSK:    senderSK,
		info:        info,
	}
}

func validateTxConvertVer1ToVer2Params (params *TxConvertVer1ToVer2InitParams) error {
	if len(params.inputCoins) > 255 {
		return NewTransactionErr(InputCoinIsVeryLargeError, nil, strconv.Itoa(len(params.inputCoins)))
	}
	if len(params.paymentInfo) > 254 {
		return NewTransactionErr(PaymentInfoIsVeryLargeError, nil, strconv.Itoa(len(params.paymentInfo)))
	}
	limitFee := uint64(0)
	estimateTxSizeParam := NewEstimateTxSizeParam(len(params.inputCoins), len(params.paymentInfo),
		false, nil, nil, limitFee)
	if txSize := EstimateTxSize(estimateTxSizeParam); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	sumInput, sumOutput := uint64(0), uint64(0)
	for _, c := range params.inputCoins {
		if c.GetVersion() != 1 {
			err := errors.New("TxConversion should only have inputCoins version 1")
			return NewTransactionErr(InvalidInputCoinVersionErr, err)
		}

		//Verify if input coins have been concealed
		if c.GetRandomness() == nil || c.GetSNDerivator() == nil || c.GetPublicKey() == nil || c.GetCommitment() == nil {
			err := errors.New("input coins should not be concealed")
			return NewTransactionErr(InvalidInputCoinVersionErr, err)
		}
		sumInput += c.GetValue()
	}
	for _, c := range params.paymentInfo {
		sumOutput += c.Amount
	}
	if sumInput != sumOutput + params.fee {
		err := errors.New("TxConversion's sum input coin and output coin (with fee) is not the same")
		return NewTransactionErr(SumInputCoinsAndOutputCoinsError, err)
	}

	if params.tokenID == nil {
		// using default PRV
		params.tokenID = &common.Hash{}
		if err := params.tokenID.SetBytes(common.PRVCoinID[:]); err != nil {
			return NewTransactionErr(TokenIDInvalidError, err, params.tokenID.String())
		}
	}
	return nil
}

func initializeTxConversion(tx *TxVersion2, params *TxConvertVer1ToVer2InitParams) error {
	var err error
	// Get Keyset from param
	senderKeySet :=  incognitokey.KeySet{}
	if err:= senderKeySet.InitFromPrivateKey(params.senderSK); err != nil {
		Logger.Log.Errorf("Cannot parse Private Key. Err %v", err)
		return NewTransactionErr(PrivateKeySenderInvalidError, err)
	}

	// Tx: initialize some values
	tx.Fee = params.fee
	tx.Version = TxConversionVersion12Number
	tx.Type = common.TxConversionType
	tx.Metadata = params.metaData
	tx.PubKeyLastByteSender = senderKeySet.PaymentAddress.Pk[len(senderKeySet.PaymentAddress.Pk)-1]

	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}
	if tx.Info, err = getTxInfo(params.info); err != nil {
		return err
	}
	return nil
}

func InitConversion(tx *TxVersion2, params *TxConvertVer1ToVer2InitParams) error {
	// validate again
	if err := validateTxConvertVer1ToVer2Params(params); err != nil {
		return err
	}
	if err := initializeTxConversion(tx, params); err != nil {
		return err
	}
	if err := proveConversion(tx, params); err != nil {
		return err
	}
	return nil
}

func getOutputcoinsFromPaymentInfo(paymentInfos []*privacy.PaymentInfo, tokenID *common.Hash,  statedb *statedb.StateDB) ([]*coin.CoinV2, error) {
	var err error
	c := make([]*coin.CoinV2, len(paymentInfos))

	for i := 0; i < len(paymentInfos); i += 1 {
		c[i], err = newCoinUniqueOTABasedOnPaymentInfo(paymentInfos[i], tokenID, statedb)
		if err != nil {
			Logger.Log.Errorf("TxConversion cannot create new coin unique OTA, got error %v", err)
			return nil, err
		}
	}
	return c, nil
}

func proveConversion(tx *TxVersion2, params *TxConvertVer1ToVer2InitParams) error {
	inputCoins := params.inputCoins
	outputCoins, err := getOutputcoinsFromPaymentInfo(params.paymentInfo, params.tokenID, params.stateDB)
	if err != nil {
		Logger.Log.Errorf("TxConversion cannot get output coins from payment info got error %v", err)
		return err
	}
	lenInputs := len(inputCoins)
	serialnumberWitness := make([]*serialnumbernoprivacy.SNNoPrivacyWitness, lenInputs)
	for i := 0; i < len(inputCoins); i++ {
		/***** Build witness for proving that serial number is derived from the committed derivator *****/
		serialnumberWitness[i] = new(serialnumbernoprivacy.SNNoPrivacyWitness)
		serialnumberWitness[i].Set(inputCoins[i].GetKeyImage(), inputCoins[i].GetPublicKey(),
			inputCoins[i].GetSNDerivator(), new(privacy.Scalar).FromBytesS(*params.senderSK))
	}
	tx.Proof, err = privacy_v2.ProveConversion(inputCoins, outputCoins, serialnumberWitness)
	if err != nil {
		Logger.Log.Errorf("Error in privacy_v2.Prove, error %v ", err)
		return err
	}
	//tx.sigPrivKey = []byte{}
	//randSK := big.NewInt(0)
	//tx.sigPrivKey = append(*params.senderSK, randSK.Bytes()...) //CHECK THIS! Why setting tx.sigPrivKey?

	// sign tx
	if tx.Sig, tx.SigPubKey, err = signNoPrivacy(params.senderSK, tx.Hash()[:]); err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(SignTxError, err)
	}
	return nil
}

func validateConversionVer1ToVer2(tx metadata.Transaction, statedb *statedb.StateDB, shardID byte, tokenID *common.Hash) (bool, error) {
	//Step to validate a ConversionVer1ToVer2 proof:
	//	- verify signature no privacy
	//	- verify if input coins have been spent (serial number already existed in database)
	//	- verify if output coins' OTA has been obtained
	//	- verify proofConversion

	if valid, err := verifySigNoPrivacy(tx.GetSig(), tx.GetSigPubKey(), tx.Hash()[:]); !valid {
		if err != nil {
			Logger.Log.Errorf("Error verifying signature conversion with tx hash %s: %+v \n", tx.Hash().String(), err)
			return false, NewTransactionErr(VerifyTxSigFailError, err)
		}
		Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE conversion with tx hash %s", tx.Hash().String())
		return false, NewTransactionErr(VerifyTxSigFailError, fmt.Errorf("FAILED VERIFICATION SIGNATURE ver1 with tx hash %s", tx.Hash().String()))
	}
	proofConversion, ok := tx.GetProof().(*privacy_v2.ConversionProofVer1ToVer2)
	if !ok {
		Logger.Log.Error("Error casting ConversionProofVer1ToVer2")
		return false, errors.New("Error casting ConversionProofVer1ToVer2")
	}

	//Verify that input coins have not been spent
	inputCoins := proofConversion.GetInputCoins()
	for i := 0; i < len(inputCoins); i++ {
		if ok, err := txDatabaseWrapper.hasSerialNumber(statedb, *tokenID, inputCoins[i].GetKeyImage().ToBytesS(), shardID); ok || err != nil {
			if err != nil {
				errStr := fmt.Sprintf("TxConversion database serialNumber got error: %v", err)
				return false, errors.New(errStr)
			}
			return false, errors.New("TxConversion found existing serialNumber in database error")
		}
	}

	//Verify that output coins' one-time-address has not been obtained
	outputCoins := proofConversion.GetOutputCoins()
	for i := 0; i < len(outputCoins); i++ {
		if ok, err := txDatabaseWrapper.hasOnetimeAddress(statedb, *tokenID, outputCoins[i].GetPublicKey().ToBytesS()); ok || err != nil {
			if err != nil {
				errStr := fmt.Sprintf("TxConversion database onetimeAddress got error: %v", err)
				return false, errors.New(errStr)
			}
			return false, errors.New("TxConversion found existing onetimeaddress in database error")
		}
	}

	//Verify the conversion proof
	valid, err := proofConversion.Verify(false, tx.GetSigPubKey(), tx.GetTxFee(), shardID, tokenID, false, nil)
	if !valid {
		if err != nil {
			Logger.Log.Error(err)
		}
		return false, NewTransactionErr(TxProofVerifyFailError, err, tx.Hash().String())
	}

	Logger.Log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF")
	return true, nil
}

// ================ TX TOKEN CONVERSION =================

type CustomTokenConversionParams struct {
	tokenID       *common.Hash
	tokenInputs   []coin.PlainCoin
	tokenPayments []*privacy.PaymentInfo
}

type TxTokenConvertVer1ToVer2InitParams struct {
	senderSK      *privacy.PrivateKey
	feeInputs     []coin.PlainCoin
	feePayments   []*privacy.PaymentInfo
	fee           uint64
	tokenParams   *CustomTokenConversionParams
	stateDB       *statedb.StateDB
	bridgeStateDB *statedb.StateDB
	metaData      metadata.Metadata
	info          []byte // 512 bytes
}

func NewTxTokenConvertVer1ToVer2InitParams(senderSK *privacy.PrivateKey,
	feeInputs []coin.PlainCoin,
	feePayments []*privacy.PaymentInfo,
	tokenInputs []coin.PlainCoin,
	tokenPayments []*privacy.PaymentInfo,
	fee uint64,
	stateDB *statedb.StateDB,
	bridgeStateDB *statedb.StateDB,
	tokenID *common.Hash, // tokenID of the conversion coin
	metaData metadata.Metadata,
	info []byte) *TxTokenConvertVer1ToVer2InitParams {

	tokenParams := &CustomTokenConversionParams{
		tokenID:       tokenID,
		tokenPayments: tokenPayments,
		tokenInputs:   tokenInputs,
	}
	return  &TxTokenConvertVer1ToVer2InitParams{
		stateDB:     stateDB,
		feeInputs:   feeInputs,
		fee:         fee,
		tokenParams: tokenParams,
		bridgeStateDB: bridgeStateDB,
		metaData:    metaData,
		feePayments: feePayments,
		senderSK:    senderSK,
		info:        info,
	}
}

func validateTxTokenConvertVer1ToVer2Params (params *TxTokenConvertVer1ToVer2InitParams) error {
	if len(params.feeInputs) > 255 {
		return errors.New("FeeInput is too large, feeInputs length = " + strconv.Itoa(len(params.feeInputs)))
	}
	if len(params.feePayments) > 255 {
		return errors.New("FeePayment is too large, feePayments length = " + strconv.Itoa(len(params.feePayments)))
	}
	if len(params.tokenParams.tokenPayments) > 255 {
		return errors.New("tokenPayments is too large, tokenPayments length = " + strconv.Itoa(len(params.tokenParams.tokenPayments)))
	}
	if len(params.tokenParams.tokenInputs) > 255 {
		return errors.New("tokenInputs length = " + strconv.Itoa(len(params.tokenParams.tokenInputs)))
	}

	limitFee := uint64(0)
	estimateTxSizeParam := NewEstimateTxSizeParam(len(params.feeInputs), len(params.feePayments),
		false, nil, nil, limitFee)
	if txSize := EstimateTxSize(estimateTxSizeParam); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	for _, c := range params.feeInputs {
		if c.GetVersion() != TxVersion2Number {
			return errors.New("TxConversion should only have fee input coins version 2")
		}
	}
	tokenParams := params.tokenParams
	if tokenParams.tokenID == nil {
		return NewTransactionErr(TokenIDInvalidError, errors.New("TxTokenConversion should have its tokenID not null"))
	}
	sumInput := uint64(0)
	for _, c := range tokenParams.tokenInputs {
		sumInput += c.GetValue()
	}
	if sumInput != tokenParams.tokenPayments[0].Amount {
		return NewTransactionErr(SumInputCoinsAndOutputCoinsError, errors.New("sumInput and sum TokenPayment amount is not equal"))
	}
	return nil
}

func (txToken *TxTokenVersion2) initTokenConversion(params *TxTokenConvertVer1ToVer2InitParams) error {
	txToken.TxTokenData.SetType(CustomTokenTransfer)
	txToken.TxTokenData.SetPropertyName("")
	txToken.TxTokenData.SetPropertySymbol("")
	txToken.TxTokenData.SetMintable(false)
	txToken.TxTokenData.SetPropertyID(*params.tokenParams.tokenID)

	existed := txDatabaseWrapper.privacyTokenIDExisted(params.stateDB, *params.tokenParams.tokenID)
	if !existed {
		if err := checkIsBridgeTokenID(params.bridgeStateDB, params.tokenParams.tokenID); err != nil {
			return err
		}
	}

	txConvertParams := NewTxConvertVer1ToVer2InitParams(
		params.senderSK,
		params.tokenParams.tokenPayments,
		params.tokenParams.tokenInputs,
		0,
		params.stateDB,
		params.tokenParams.tokenID,
		nil,
		params.info,
	)
	txNormal := new(TxVersion2)
	if err := InitConversion(txNormal, txConvertParams); err != nil {
		return NewTransactionErr(PrivacyTokenInitTokenDataError, err)
	}
	txToken.TxTokenData.TxNormal = txNormal
	return nil
}

func (txToken *TxTokenVersion2) initPRVFeeConversion(feeTx *TxVersion2, params *TxPrivacyInitParams) error {
	txTokenDataHash, err := txToken.TxTokenData.Hash()
	if err != nil {
		Logger.Log.Errorf("Cannot calculate txPrivacyTokenData Hash, err %v", err)
		return err
	}
	if err := feeTx.proveWithMessage(params, txTokenDataHash[:]); err != nil {
		return NewTransactionErr(PrivacyTokenInitPRVError, err)
	}
	// override TxCustomTokenPrivacyType type
	feeTx.SetVersion(TxVersion2Number)
	feeTx.SetType(common.TxTokenConversionType)
	txToken.Tx = feeTx

	return nil
}

func InitTokenConversion(txToken *TxTokenVersion2, params *TxTokenConvertVer1ToVer2InitParams) error {
	if err := validateTxTokenConvertVer1ToVer2Params(params); err != nil {
		return err
	}

	txPrivacyParams := NewTxPrivacyInitParams(
		params.senderSK, params.feePayments, params.feeInputs, params.fee,
		false, params.stateDB, nil, params.metaData, params.info,
	)
	if err := validateTxParams(txPrivacyParams); err != nil {
		return err
	}
	// Init tx and params (tx and params will be changed)
	tx := new(TxVersion2)
	if err := tx.initializeTxAndParams(txPrivacyParams); err != nil {
		return err
	}

	// Init Token first
	if err := txToken.initTokenConversion(params); err != nil {
		Logger.Log.Errorf("Cannot init token ver2: err %v", err)
		return err
	}

	// Init PRV Fee on the whole transaction
	if err := txToken.initPRVFeeConversion(tx, txPrivacyParams); err != nil {
		Logger.Log.Errorf("Cannot init token ver2: err %v", err)
		return err
	}

	return nil
}