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
	info []byte) (*TxConvertVer1ToVer2InitParams, error) {

	params := &TxConvertVer1ToVer2InitParams{
		stateDB:     stateDB,
		tokenID:     tokenID,
		inputCoins:  inputCoins,
		fee:         fee,
		metaData:    metaData,
		paymentInfo: paymentInfo,
		senderSK:    senderSK,
		info:        info,
	}

	if err := validateTxConvertVer1ToVer2Params(params); err != nil {
		return nil, err
	}
	return params, nil
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

// return bool indicates whether we should continue "Init" function or not
func initializeTxConversion(tx *TxVersion2, params *TxConvertVer1ToVer2InitParams) error {
	var err error
	// Tx: initialize some values
	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}
	tx.Fee = params.fee
	tx.Version = TxConversionVersion12Number
	tx.Type = common.TxNormalType
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