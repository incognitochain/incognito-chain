package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
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

func initializeTxAndParamsASM(tx *Tx, params *TxPrivacyInitParamsForASM) error {
	txParams := &params.txParam
	err := initializeTxAndParams(tx, txParams)
	if txParams.hasPrivacy {
		// Check number of list of random commitments, list of random commitment indices
		if len(params.commitmentIndices) != len(params.txParam.inputCoins)*privacy.CommitmentRingSize {
			return NewTransactionErr(RandomCommitmentError, nil)
		}

		if len(params.myCommitmentIndices) != len(params.txParam.inputCoins) {
			return NewTransactionErr(RandomCommitmentError, errors.New("number of list my commitment indices must be equal to number of input coins"))
		}
	}
	return err
}

func (tx *Tx) InitForASM(params *TxPrivacyInitParamsForASM) error {
	var err error
	Logger.log.Debugf("CREATING TX........\n")
	txParams := &params.txParam
	if err := validateTxInit(txParams); err != nil {
		return err
	}

	// Execution time
	// start := time.Now()

	// Init tx and params (tx and params will be changed)
	if err := initializeTxAndParamsASM(tx, params); err != nil {
		return err
	}

	// create SNDs for output coins
	outputCoins, err := parseOutputCoins(txParams)
	if err != nil {
		return err
	}
	// get list of commitments for proving one-out-of-many from commitmentIndexs
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)
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
		OutputCoins:             *outputCoins,
		PublicKeyLastByteSender: tx.PubKeyLastByteSender,
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

	err = tx.signTx()
	if err != nil {
		Logger.log.Error(err)
		return NewTransactionErr(SignTxError, err)
	}

	// snProof := tx.Proof.GetSerialNumberProof()
	// for i := 0; i < len(snProof); i++ {
	// 	res, _ := snProof[i].Verify(nil)
	// 	println("Verify serial number proof: ", i, ": ", res)
	// }

	//elapsedPrivacy := time.Since(startPrivacy)
	//elapsed := time.Since(start)
	//Logger.log.Debugf("Creating payment proof time %s", elapsedPrivacy)
	//Logger.log.Debugf("Successfully Creating normal tx %+v in %s time", *tx.Hash(), elapsed)
	return nil
}
