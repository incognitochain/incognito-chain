package transaction

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
)

type TxVersion2 struct{}

func generateMlsagRing(params *TxPrivacyInitParams, pi int, shardId int) (*mlsag.Ring, error) {
	outputCoinsPtr, err := parseOutputCoins(params)
	if err != nil {
		return nil, err
	}
	outputCoins := *outputCoinsPtr
	inputCoins := params.inputCoins

	ring := make([][]*operation.Point, privacy.RingSize)

	outputCommitments := new(operation.Point).Identity()
	for i := 0; i < len(outputCoins); i += 1 {
		outputCommitments.Add(outputCommitments, outputCoins[i].CoinDetails.GetCoinCommitment())
	}

	lenCommitment, err := statedb.GetCommitmentLength(params.stateDB, *params.tokenID, shardID)
	if err != nil {
		Logger.Log.Error(err)
		return nil, err
	}

	for i := 0; i < privacy.RingSize; i += 1 {
		sumInputs := new(operation.Point).Identity()
		row := make([]*operation.Point, len(inputCoins))
		if i == pi {
			for j := 0; j < len(inputCoins); j += 1 {
				privKey := new(operation.Scalar).FromBytesS(*key)
				row[i][j] = new(operation.Point).ScalarMultBase(privKey)
				sumInputs.Add(inputCoins[j].CoinDetails.GetCoinCommitment())
			}
		} else {
			for j := 0; j < len(inputCoins); j += 1 {
				row[i][j] = key
			}
		}
	}
}

// signTx - signs tx
func signTxVer2(tx *Tx, params *TxPrivacyInitParams) error {
	if tx.Sig != nil {
		return NewTransactionErr(UnexpectedError, errors.New("input transaction must be an unsigned one"))
	}

	message := (*tx.Proof).Bytes()
	var pi int = common.RandIntInterval(0, privacy.RingSize-1)
	ring := generateMlsagRing(params, pi)
	////////

	tx.SigPubKey = sigKey.GetPublicKey().GetPublicKey().ToBytesS()

	// signing
	if Logger.Log != nil {
		Logger.Log.Debugf(tx.Hash().String())
	}
	signature, err := sigKey.Sign(tx.Hash()[:])
	if err != nil {
		return err
	}

	// convert signature to byte array
	tx.Sig = signature.Bytes()

	return nil
}

func (*TxVersion2) Prove(tx *Tx, params *TxPrivacyInitParams) error {
	outputCoins, err := parseOutputCoins(params)
	if err != nil {
		return err
	}
	inputCoins := &params.inputCoins

	var conversion privacy.Proof
	conversion, err = privacy_v2.Prove(inputCoins, outputCoins, params.hasPrivacy)
	if err != nil {
		return err
	}
	tx.Proof = &conversion

	err = signTxVer2(tx, params)
	return err
}

func (*TxVersion2) Verify(tx *Tx, hasPrivacy bool, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	return true, nil
}
