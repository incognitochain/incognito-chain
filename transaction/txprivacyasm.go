package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	zkp "github.com/incognitochain/incognito-chain/privacy/zeroknowledge"
)

type TxPrivacyInitParamsForASM struct {
	txParam             TxPrivacyInitParams
	commitmentIndices   []uint64
	commitmentBytes     [][]byte
	myCommitmentIndices []uint64
	sndOutputs          []*privacy.Scalar
}

func NewTxPrivacyInitParamsForASM(
	senderSK *privacy.PrivateKey,
	paymentInfo []*privacy.PaymentInfo,
	inputCoins []*privacy.InputCoin,
	fee uint64,
	hasPrivacy bool,
	tokenID *common.Hash, // default is nil -> use for prv coin
	metaData metadata.Metadata,
	info []byte,
	commitmentIndices []uint64,
	commitmentBytes [][]byte,
	myCommitmentIndices []uint64,
	sndOutputs []*privacy.Scalar) *TxPrivacyInitParamsForASM {

	txParam := TxPrivacyInitParams{
		senderSK:    senderSK,
		paymentInfo: paymentInfo,
		inputCoins:  inputCoins,
		fee:         fee,
		hasPrivacy:  hasPrivacy,
		tokenID:     tokenID,
		metaData:    metaData,
		info:        info,
	}
	params := &TxPrivacyInitParamsForASM{
		txParam:             txParam,
		commitmentIndices:   commitmentIndices,
		commitmentBytes:     commitmentBytes,
		myCommitmentIndices: myCommitmentIndices,
		sndOutputs:          sndOutputs,
	}
	return params
}

func (param *TxPrivacyInitParamsForASM) SetMetaData(meta metadata.Metadata) {
	param.txParam.metaData = meta
}

func (tx *Tx) InitForASM(params *TxPrivacyInitParamsForASM) error {

	//Logger.log.Debugf("CREATING TX........\n")
	tx.Version = txVersion
	var err error

	if len(params.txParam.inputCoins) > 255 {
		return NewTransactionErr(InputCoinIsVeryLargeError, nil, strconv.Itoa(len(params.txParam.inputCoins)))
	}

	if len(params.txParam.paymentInfo) > 254 {
		return NewTransactionErr(PaymentInfoIsVeryLargeError, nil, strconv.Itoa(len(params.txParam.paymentInfo)))
	}

	if params.txParam.tokenID == nil {
		// using default PRV
		params.txParam.tokenID = &common.Hash{}
		err := params.txParam.tokenID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return NewTransactionErr(TokenIDInvalidError, err, params.txParam.tokenID.GetBytes())
		}
	}

	// Calculate execution time
	//start := time.Now()

	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}

	// create sender's key set from sender's spending key
	senderFullKey := incognitokey.KeySet{}
	err = senderFullKey.InitFromPrivateKey(params.txParam.senderSK)
	if err != nil {
		Logger.log.Error(errors.New(fmt.Sprintf("Can not import Private key for sender keyset from %+v", params.txParam.senderSK)))
		return NewTransactionErr(PrivateKeySenderInvalidError, err)
	}
	// get public key last byte of sender
	pkLastByteSender := senderFullKey.PaymentAddress.Pk[len(senderFullKey.PaymentAddress.Pk)-1]

	// init info of tx
	tx.Info = []byte{}
	if len(params.txParam.info) > 0 {
		tx.Info = params.txParam.info
	}

	// set metadata
	tx.Metadata = params.txParam.metaData

	// set tx type
	tx.Type = common.TxNormalType
	//Logger.log.Debugf("len(inputCoins), fee, hasPrivacy: %d, %d, %v\n", len(params.inputCoins), params.fee, params.hasPrivacy)

	if len(params.txParam.inputCoins) == 0 && params.txParam.fee == 0 && !params.txParam.hasPrivacy {
		//Logger.log.Debugf("len(inputCoins) == 0 && fee == 0 && !hasPrivacy\n")
		tx.Fee = params.txParam.fee
		tx.sigPrivKey = *params.txParam.senderSK
		tx.PubKeyLastByteSender = pkLastByteSender
		err := tx.signTx()
		if err != nil {
			Logger.log.Error(errors.New(fmt.Sprintf("Cannot sign tx %v\n", err)))
			return NewTransactionErr(SignTxError, err)
		}
		return nil
	}

	shardID := common.GetShardIDFromLastByte(pkLastByteSender)

	if params.txParam.hasPrivacy {
		// Check number of list of random commitments, list of random commitment indices
		if len(params.commitmentIndices) != len(params.txParam.inputCoins)*privacy.CommitmentRingSize {
			return NewTransactionErr(RandomCommitmentError, nil)
		}

		if len(params.myCommitmentIndices) != len(params.txParam.inputCoins) {
			return NewTransactionErr(RandomCommitmentError, errors.New("number of list my commitment indices must be equal to number of input coins"))
		}
	}

	// Calculate execution time for creating payment proof
	//startPrivacy := time.Now()

	// Calculate sum of all output coins' value
	sumOutputValue := uint64(0)
	for _, p := range params.txParam.paymentInfo {
		sumOutputValue += p.Amount
	}

	// Calculate sum of all input coins' value
	sumInputValue := uint64(0)
	for _, coin := range params.txParam.inputCoins {
		sumInputValue += coin.CoinDetails.GetValue()
	}
	//Logger.log.Debugf("sumInputValue: %d\n", sumInputValue)

	// Calculate over balance, it will be returned to sender
	overBalance := int64(sumInputValue - sumOutputValue - params.txParam.fee)

	// Check if sum of input coins' value is at least sum of output coins' value and tx fee
	if overBalance < 0 {
		return NewTransactionErr(WrongInputError,
			errors.New(
				fmt.Sprintf("input value less than output value. sumInputValue=%d sumOutputValue=%d fee=%d",
					sumInputValue, sumOutputValue, params.txParam.fee)))
	}

	// if overBalance > 0, create a new payment info with pk is sender's pk and amount is overBalance
	if overBalance > 0 {
		changePaymentInfo := new(privacy.PaymentInfo)
		changePaymentInfo.Amount = uint64(overBalance)
		changePaymentInfo.PaymentAddress = senderFullKey.PaymentAddress
		params.txParam.paymentInfo = append(params.txParam.paymentInfo, changePaymentInfo)
	}

	// create new output coins
	outputCoins := make([]*privacy.OutputCoin, len(params.txParam.paymentInfo))

	// create SNDs for output coins
	sndOuts := params.sndOutputs

	// create new output coins with info: Pk, value, last byte of pk, snd
	for i, pInfo := range params.txParam.paymentInfo {
		outputCoins[i] = new(privacy.OutputCoin)
		outputCoins[i].CoinDetails = new(privacy.Coin)
		outputCoins[i].CoinDetails.SetValue(pInfo.Amount)
		if len(pInfo.Message) > 0 {
			if len(pInfo.Message) > privacy.MaxSizeInfoCoin {
				return NewTransactionErr(ExceedSizeInfoOutCoinError, nil)
			}
			outputCoins[i].CoinDetails.SetInfo(pInfo.Message)
		}

		PK, err := new(privacy.Point).FromBytesS(pInfo.PaymentAddress.Pk)
		if err != nil {
			Logger.log.Error(errors.New(fmt.Sprintf("can not decompress public key from %+v", pInfo.PaymentAddress)))
			return NewTransactionErr(DecompressPaymentAddressError, err, pInfo.PaymentAddress)
		}
		outputCoins[i].CoinDetails.SetPublicKey(PK)
		outputCoins[i].CoinDetails.SetSNDerivator(sndOuts[i])
	}

	// assign fee tx
	tx.Fee = params.txParam.fee

	// create zero knowledge proof of payment
	tx.Proof = &zkp.PaymentProof{}

	// get list of commitments for proving one-out-of-many from commitmentIndexs
	commitmentProving := make([]*privacy.Point, len(params.commitmentBytes))
	for i, cmBytes := range params.commitmentBytes {
		commitmentProving[i] = new(privacy.Point)
		commitmentProving[i], err = commitmentProving[i].FromBytesS(cmBytes)
		if err != nil {
			Logger.log.Error(errors.New(fmt.Sprintf("can not get commitment from index=%d shardID=%+v value=%+v", params.commitmentIndices[i], shardID, cmBytes)))
			return NewTransactionErr(CanNotDecompressCommitmentFromIndexError, err, params.commitmentIndices[i], shardID, cmBytes)
		}
	}

	// prepare witness for proving
	witness := new(zkp.PaymentWitness)
	paymentWitnessParam := zkp.PaymentWitnessParam{
		HasPrivacy:              params.txParam.hasPrivacy,
		PrivateKey:              new(privacy.Scalar).FromBytesS(*params.txParam.senderSK),
		InputCoins:              params.txParam.inputCoins,
		OutputCoins:             outputCoins,
		PublicKeyLastByteSender: pkLastByteSender,
		Commitments:             commitmentProving,
		CommitmentIndices:       params.commitmentIndices,
		MyCommitmentIndices:     params.myCommitmentIndices,
		Fee:                     params.txParam.fee,
	}
	err = witness.Init(paymentWitnessParam)
	if err.(*errhandler.PrivacyError) != nil {
		Logger.log.Error(err)
		jsonParam, _ := json.MarshalIndent(paymentWitnessParam, common.EmptyString, "  ")
		return NewTransactionErr(InitWithnessError, err, string(jsonParam))
	}

	tx.Proof, err = witness.Prove(params.txParam.hasPrivacy)
	if err.(*errhandler.PrivacyError) != nil {
		Logger.log.Error(err)
		jsonParam, _ := json.MarshalIndent(paymentWitnessParam, common.EmptyString, "  ")
		return NewTransactionErr(WithnessProveError, err, params.txParam.hasPrivacy, string(jsonParam))
	}

	//Logger.log.Debugf("DONE PROVING........\n")

	// set private key for signing tx
	if params.txParam.hasPrivacy {
		randSK := witness.GetRandSecretKey()
		tx.sigPrivKey = append(*params.txParam.senderSK, randSK.ToBytesS()...)

		// encrypt coin details (Randomness)
		// hide information of output coins except coin commitments, public key, snDerivators
		for i := 0; i < len(tx.Proof.GetOutputCoins()); i++ {
			err = tx.Proof.GetOutputCoins()[i].Encrypt(params.txParam.paymentInfo[i].PaymentAddress.Tk)
			if err.(*errhandler.PrivacyError) != nil {
				Logger.log.Error(err)
				return NewTransactionErr(EncryptOutputError, err)
			}
			tx.Proof.GetOutputCoins()[i].CoinDetails.SetSerialNumber(nil)
			tx.Proof.GetOutputCoins()[i].CoinDetails.SetValue(0)
			tx.Proof.GetOutputCoins()[i].CoinDetails.SetRandomness(nil)
		}

		// hide information of input coins except serial number of input coins
		for i := 0; i < len(tx.Proof.GetInputCoins()); i++ {
			tx.Proof.GetInputCoins()[i].CoinDetails.SetCoinCommitment(nil)
			tx.Proof.GetInputCoins()[i].CoinDetails.SetValue(0)
			tx.Proof.GetInputCoins()[i].CoinDetails.SetSNDerivator(nil)
			tx.Proof.GetInputCoins()[i].CoinDetails.SetPublicKey(nil)
			tx.Proof.GetInputCoins()[i].CoinDetails.SetRandomness(nil)
		}

	} else {
		tx.sigPrivKey = []byte{}
		randSK := big.NewInt(0)
		tx.sigPrivKey = append(*params.txParam.senderSK, randSK.Bytes()...)
	}

	// sign tx
	tx.PubKeyLastByteSender = pkLastByteSender
	err = tx.signTx()
	if err != nil {
		Logger.log.Error(err)
		return NewTransactionErr(SignTxError, err)
	}

	snProof := tx.Proof.GetSerialNumberProof()
	for i := 0; i < len(snProof); i++ {
		res, _ := snProof[i].Verify(nil)
		println("Verify serial number proof: ", i, ": ", res)
	}

	//elapsedPrivacy := time.Since(startPrivacy)
	//elapsed := time.Since(start)
	//Logger.log.Debugf("Creating payment proof time %s", elapsedPrivacy)
	//Logger.log.Debugf("Successfully Creating normal tx %+v in %s time", *tx.Hash(), elapsed)
	return nil
}
