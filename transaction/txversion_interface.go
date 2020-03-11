package transaction

// What needs to know when Tx Bridge with Privacy package?
// Only 2 main things:
// "Prove" and "Verify" these rules:
// - For each input conceal our real input by getting random inputs (ring signature).
// - Ensure sum input = sum output (pedersen commitment)
// - Ensure all output is non-negative (bulletproofs, aggregaterangeproof)

// Ver 1:
// Prove:
// - Prove the input is oneofmany with other random inputs (with sum input = output by Pedersen)
// - Prove the non-negative with bulletproofs (aggregaterangeproof)
// - Sign the above proofs

// Ver 2:
// Prove:
// - Prove the non-negative with bulletproofs (aggregaterangeproof)
// - Prove the input is one of many with other random inputs plus sum input = output using MLSAG. (it also provides signature).

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
	"github.com/incognitochain/incognito-chain/privacy"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
)

type TxVersionSwitcher interface {
	Prove(tx *Tx, params *TxPrivacyInitParams) error
	Verify() (bool, error)
}

type TxVersion1 struct{}

func (*TxVersion1) Prove(tx *Tx, params *TxPrivacyInitParams) error {
	// Prepare paymentWitness params
	paymentWitnessParamPtr, err := initializePaymentWitnessParam(tx, params)
	if err != nil {
		return err
	}
	paymentWitnessParam := *paymentWitnessParamPtr

	witness := new(zkp.PaymentWitness)
	err = witness.Init(paymentWitnessParam)
	if err.(*errhandler.PrivacyError) != nil {
		Logger.log.Error(err)
		jsonParam, _ := json.MarshalIndent(paymentWitnessParam, common.EmptyString, "  ")
		return NewTransactionErr(InitWithnessError, err, string(jsonParam))
	}

	tx.Proof, err = witness.Prove(params.hasPrivacy)
	if err.(*errhandler.PrivacyError) != nil {
		Logger.log.Error(err)
		jsonParam, _ := json.MarshalIndent(paymentWitnessParam, common.EmptyString, "  ")
		return NewTransactionErr(WithnessProveError, err, params.hasPrivacy, string(jsonParam))
	}

	// set private key for signing tx
	if params.hasPrivacy {
		randSK := witness.GetRandSecretKey()
		tx.sigPrivKey = append(*params.senderSK, randSK.ToBytesS()...)

		// encrypt coin details (Randomness)
		// hide information of output coins except coin commitments, public key, snDerivators
		for i := 0; i < len(tx.Proof.GetOutputCoins()); i++ {
			err = tx.Proof.GetOutputCoins()[i].Encrypt(params.paymentInfo[i].PaymentAddress.Tk)
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
		tx.sigPrivKey = append(*params.senderSK, randSK.Bytes()...)
	}

	// sign tx
	err = tx.signTx()
	if err != nil {
		Logger.log.Error(err)
		return NewTransactionErr(SignTxError, err)
	}
	return nil
}

func (*TxVersion1) Verify() (bool, error) {
	return true, nil
}

// This payment witness currently use one out of many
func initializePaymentWitnessParam(tx *Tx, params *TxPrivacyInitParams) (*zkp.PaymentWitnessParam, error) {
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)

	// get list of commitments for proving one-out-of-many from commitmentIndexs
	commitmentIndexs, myCommitmentIndexs, err := parseCommitments(params, shardID)
	if err != nil {
		return nil, err
	}
	commitmentProving, err := parseCommitmentProving(params, shardID, commitmentIndexs)
	if err != nil {
		return nil, err
	}
	outputCoins, err := parseOutputCoins(params)
	if err != nil {
		return nil, err
	}
	// prepare witness for proving
	paymentWitnessParam := zkp.PaymentWitnessParam{
		HasPrivacy:              params.hasPrivacy,
		PrivateKey:              new(privacy.Scalar).FromBytesS(*params.senderSK),
		InputCoins:              params.inputCoins,
		OutputCoins:             *outputCoins,
		PublicKeyLastByteSender: tx.PubKeyLastByteSender,
		Commitments:             *commitmentProving,
		CommitmentIndices:       *commitmentIndexs,
		MyCommitmentIndices:     *myCommitmentIndexs,
		Fee:                     params.fee,
	}
	return &paymentWitnessParam, nil
}

// Used in Tx.Init
// For Tx to be formed correctly by using privacy package
func initBridgeTxWithPrivacy(tx *Tx, params *TxPrivacyInitParams) error {
	versionSwitcher := TxVersionSwitcher{}
	if tx.Version == 1 {
		versionSwitcher[0] = new(TxVersion1)
	}
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
	var valid bool
	var err error

	valid, err = tx.verifySigTx()
	if !valid {
		if err != nil {
			Logger.log.Errorf("Error verifying signature with tx hash %s: %+v \n", tx.Hash().String(), err)
			return false, NewTransactionErr(VerifyTxSigFailError, err)
		}
		Logger.log.Errorf("FAILED VERIFICATION SIGNATURE with tx hash %s", tx.Hash().String())
		return false, NewTransactionErr(VerifyTxSigFailError, fmt.Errorf("FAILED VERIFICATION SIGNATURE with tx hash %s", tx.Hash().String()))
	}

	if tx.Proof == nil {
		return true, nil
	}

	if tokenID == nil {
		tokenID = &common.Hash{}
		err := tokenID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			Logger.log.Error(err)
			return false, NewTransactionErr(TokenIDInvalidError, err)
		}
	}

	sndOutputs := make([]*privacy.Scalar, len(tx.Proof.GetOutputCoins()))
	for i := 0; i < len(tx.Proof.GetOutputCoins()); i++ {
		sndOutputs[i] = tx.Proof.GetOutputCoins()[i].CoinDetails.GetSNDerivator()
	}

	if privacy.CheckDuplicateScalarArray(sndOutputs) {
		Logger.log.Errorf("Duplicate output coins' snd\n")
		return false, NewTransactionErr(DuplicatedOutputSndError, errors.New("Duplicate output coins' snd\n"))
	}

	if isNewTransaction {
		for i := 0; i < len(tx.Proof.GetOutputCoins()); i++ {
			// Check output coins' SND is not exists in SND list (Database)
			if ok, err := CheckSNDerivatorExistence(tokenID, tx.Proof.GetOutputCoins()[i].CoinDetails.GetSNDerivator(), db); ok || err != nil {
				if err != nil {
					Logger.log.Error(err)
				}
				Logger.log.Errorf("snd existed: %d\n", i)
				return false, NewTransactionErr(SndExistedError, err, fmt.Sprintf("snd existed: %d\n", i))
			}
		}
	}

	if !hasPrivacy {
		// Check input coins' commitment is exists in cm list (Database)
		for i := 0; i < len(tx.Proof.GetInputCoins()); i++ {
			ok, err := tx.CheckCMExistence(tx.Proof.GetInputCoins()[i].CoinDetails.GetCoinCommitment().ToBytesS(), db, shardID, tokenID)
			if !ok || err != nil {
				if err != nil {
					Logger.log.Error(err)
				}
				return false, NewTransactionErr(InputCommitmentIsNotExistedError, err)
			}
		}
	}

	// Verify the payment proof
	valid, err = tx.Proof.Verify(hasPrivacy, tx.SigPubKey, tx.Fee, db, shardID, tokenID, isBatch)
	if !valid {
		if err != nil {
			Logger.log.Error(err)
		}
		Logger.log.Error("FAILED VERIFICATION PAYMENT PROOF")
		err1, ok := err.(*errhandler.PrivacyError)
		if ok {
			// parse error detail
			if err1.Code == errhandler.ErrCodeMessage[errhandler.VerifyOneOutOfManyProofFailedErr].Code {
				if isNewTransaction {
					return false, NewTransactionErr(VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
				} else {
					// for old txs which be get from sync block or validate new block
					if tx.LockTime <= ValidateTimeForOneoutOfManyProof {
						// only verify by sign on block because of issue #504(that mean we should pass old tx, which happen before this issue)
						return true, nil
					} else {
						return false, NewTransactionErr(VerifyOneOutOfManyProofFailedErr, err1, tx.Hash().String())
					}
				}
			}
		}
		return false, NewTransactionErr(TxProofVerifyFailError, err, tx.Hash().String())
	}
	//@UNCOMMENT: metrics time
	//elapsed := time.Since(start)
	//Logger.log.Debugf("Validation normal tx %+v in %s time \n", *tx.Hash(), elapsed)
	Logger.log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
	return true, nil
}
