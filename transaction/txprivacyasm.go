package transaction

import (
	"errors"
	"github.com/incognitochain/incognito-chain/privacy/coin"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
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
	inputCoins []coin.PlainCoin,
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
	Logger.Log.Debugf("CREATING TX........\n")
	txParams := &params.txParam
	if err := validateTxInit(txParams); err != nil {
		return err
	}

	// Init tx and params (tx and params will be changed)
	if err := initializeTxAndParamsASM(tx, params); err != nil {
		return err
	}

	// Prove based on tx.Version
	prover := newTxVersionSwitcher(tx.Version)
	if err := prover.ProveASM(tx, params); err != nil {
		return err
	}

	return nil
}
