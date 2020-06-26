package transaction

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge/serialnumbernoprivacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"math/big"
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
		sumInput += c.GetValue()
		if c.GetVersion() != 1 {
			err := errors.New("TxConversion should only have inputCoins version 1")
			return NewTransactionErr(InvalidInputCoinVersionErr, err)
		}
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
	// Tx: initialize some values
	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}
	tx.Fee = params.fee
	tx.Version = TxConversionVersion12Number
	tx.Type = common.TxConversionType
	tx.Metadata = params.metaData
	if tx.Info, err = getTxInfo(params.info); err != nil {
		return err
	}
	if tx.PubKeyLastByteSender, err = parseLastByteSender(params.senderSK); err != nil {
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
	tx.sigPrivKey = []byte{}
	randSK := big.NewInt(0)
	tx.sigPrivKey = append(*params.senderSK, randSK.Bytes()...)

	// sign tx
	if tx.Sig, tx.SigPubKey, err = signNoPrivacy(params.senderSK, tx.Hash()[:]); err != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(SignTxError, err)
	}
	return nil
}

func validateConversionVer1ToVer2(tx metadata.Transaction, statedb *statedb.StateDB, shardID byte, tokenID *common.Hash) (bool, error) {
	if valid, err := verifySigNoPrivacy(tx.GetSig(), tx.GetSigPubKey(), tx.Hash()[:]); !valid {
		if err != nil {
			Logger.Log.Errorf("Error verifying signature conversion with tx hash %s: %+v \n", tx.Hash().String(), err)
			return false, NewTransactionErr(VerifyTxSigFailError, err)
		}
		Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE conversion with tx hash %s", tx.Hash().String())
		return false, NewTransactionErr(VerifyTxSigFailError, fmt.Errorf("FAILED VERIFICATION SIGNATURE ver1 with tx hash %s", tx.Hash().String()))
	}
	txConversion := tx.GetProof().(*privacy_v2.ConversionProofVer1ToVer2)
	valid, err := txConversion.Verify(false, tx.GetSigPubKey(), tx.GetTxFee(), shardID, tokenID, false, nil)
	if !valid {
		if err != nil {
			Logger.Log.Error(err)
		}
		return false, NewTransactionErr(TxProofVerifyFailError, err, tx.Hash().String())
	}
	Logger.Log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
	return true, nil
}

// ================ TX TOKEN CONVERSION =================

type CustomTokenConversionParams struct {
	tokenID        *common.Hash
	paymentInfo    []*privacy.PaymentInfo
	tokenInputs     []coin.PlainCoin
}

type TxTokenConvertVer1ToVer2InitParams struct {
	senderSK    	*privacy.PrivateKey
	feePayment 		[]*privacy.PaymentInfo
	feeInputs  		[]coin.PlainCoin
	fee         	uint64
	tokenParams 	*CustomTokenConversionParams
	stateDB     	*statedb.StateDB
	bridgeStateDB 	*statedb.StateDB
	metaData    	metadata.Metadata
	info        	[]byte // 512 bytes
}

func NewTxTokenConvertVer1ToVer2InitParams(senderSK *privacy.PrivateKey,
	feePayment []*privacy.PaymentInfo,
	paymentInfo []*privacy.PaymentInfo,
	tokenInputs []coin.PlainCoin,
	feeInputs []coin.PlainCoin,
	fee uint64,
	stateDB *statedb.StateDB,
	bridgeStateDB *statedb.StateDB,
	tokenID *common.Hash, // tokenID of the conversion coin
	metaData metadata.Metadata,
	info []byte) *TxTokenConvertVer1ToVer2InitParams {

	tokenParams := &CustomTokenConversionParams{
		tokenID:     tokenID,
		paymentInfo: paymentInfo,
		tokenInputs:  tokenInputs,
	}
	return  &TxTokenConvertVer1ToVer2InitParams{
		stateDB:     stateDB,
		feeInputs:  feeInputs,
		fee:         fee,
		tokenParams: tokenParams,
		metaData:    metaData,
		feePayment:  feePayment,
		senderSK:    senderSK,
		info:        info,
	}
}

func validateTxTokenConvertVer1ToVer2Params (params *TxTokenConvertVer1ToVer2InitParams) error {
	if len(params.feeInputs) > 255 {
		return NewTransactionErr(InputCoinIsVeryLargeError, nil, "feeInputs length = " + strconv.Itoa(len(params.feeInputs)))
	}
	if len(params.feePayment) > 255 {
		return NewTransactionErr(PaymentInfoIsVeryLargeError, nil,  "feePayment length = " + strconv.Itoa(len(params.feePayment)))
	}
	if len(params.tokenParams.paymentInfo) > 255 {
		return NewTransactionErr(InputCoinIsVeryLargeError, nil,  "paymentInfo length = " + strconv.Itoa(len(params.tokenParams.paymentInfo)))
	}
	if len(params.tokenParams.tokenInputs) > 254 {
		return NewTransactionErr(PaymentInfoIsVeryLargeError, nil,  "tokenInputs length = " + strconv.Itoa(len(params.tokenParams.tokenInputs)))
	}

	limitFee := uint64(0)
	estimateTxSizeParam := NewEstimateTxSizeParam(len(params.feeInputs), len(params.feePayment),
		false, nil, nil, limitFee)
	if txSize := EstimateTxSize(estimateTxSizeParam); txSize > common.MaxTxSize {
		return NewTransactionErr(ExceedSizeTx, nil, strconv.Itoa(int(txSize)))
	}

	sumInput, sumOutput := uint64(0), uint64(0)

	for _, c := range params.feeInputs {
		sumInput += c.GetValue()
		if c.GetVersion() != 1 {
			err := errors.New("TxConversion should only have inputCoins version 1")
			return NewTransactionErr(InvalidInputCoinVersionErr, err)
		}
	}
	for _, c := range params.feePayment {
		sumOutput += c.Amount
	}
	if sumInput != sumOutput + params.fee {
		err := errors.New("TxTokenConversion's sum input coin and output coin (with fee) is not the same")
		return NewTransactionErr(SumInputCoinsAndOutputCoinsError, err)
	}

	tokenParams := params.tokenParams
	if tokenParams.tokenID == nil {
		return NewTransactionErr(TokenIDInvalidError, errors.New("TxTokenConversion should have its tokenID not null"))
	}
	sumInput = uint64(0)
	for _, c := range tokenParams.tokenInputs {
		sumInput += c.GetValue()
	}
	if sumInput != tokenParams.paymentInfo[0].Amount {
		return NewTransactionErr(SumInputCoinsAndOutputCoinsError, errors.New("TxTokenConversion should have its tokenID not null"))
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

	txParams := NewTxPrivacyInitParams(
		params.senderSK,
		params.tokenParams.paymentInfo,
		params.tokenParams.tokenInputs,
		0, false, params.stateDB,
		params.tokenParams.tokenID, nil, nil,
	)
	txNormal := new(TxVersion2)
	if err := txNormal.Init(txParams); err != nil {
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
		params.senderSK, params.feePayment, params.feeInputs, params.fee,
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