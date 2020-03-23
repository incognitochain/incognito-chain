package transaction

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
)

type TxVersion1 struct{}

func parseCommitments(params *TxPrivacyInitParams, shardID byte) (*[]uint64, *[]uint64, error) {
	var commitmentIndexs []uint64   // array index random of commitments in db
	var myCommitmentIndexs []uint64 // index in array index random of commitment in db

	if params.hasPrivacy {
		randomParams := NewRandomCommitmentsProcessParam(params.inputCoins, privacy.CommitmentRingSize, params.stateDB, shardID, params.tokenID)
		commitmentIndexs, myCommitmentIndexs, _ = RandomCommitmentsProcess(randomParams)

		// Check number of list of random commitments, list of random commitment indices
		if len(commitmentIndexs) != len(params.inputCoins)*privacy.CommitmentRingSize {
			return nil, nil, NewTransactionErr(RandomCommitmentError, nil)
		}

		if len(myCommitmentIndexs) != len(params.inputCoins) {
			return nil, nil, NewTransactionErr(RandomCommitmentError, errors.New("number of list my commitment indices must be equal to number of input coins"))
		}
	}
	return &commitmentIndexs, &myCommitmentIndexs, nil
}

func parseCommitmentProving(params *TxPrivacyInitParams, shardID byte, commitmentIndexsPtr *[]uint64) (*[]*privacy.Point, error) {
	commitmentIndexs := *commitmentIndexsPtr
	commitmentProving := make([]*privacy.Point, len(commitmentIndexs))
	for i, cmIndex := range commitmentIndexs {
		temp, err := statedb.GetCommitmentByIndex(params.stateDB, *params.tokenID, cmIndex, shardID)
		if err != nil {
			Logger.log.Error(errors.New(fmt.Sprintf("can not get commitment from index=%d shardID=%+v", cmIndex, shardID)))
			return nil, NewTransactionErr(CanNotGetCommitmentFromIndexError, err, cmIndex, shardID)
		}
		commitmentProving[i] = new(privacy.Point)
		commitmentProving[i], err = commitmentProving[i].FromBytesS(temp)
		if err != nil {
			Logger.log.Error(errors.New(fmt.Sprintf("can not get commitment from index=%d shardID=%+v value=%+v", cmIndex, shardID, temp)))
			return nil, NewTransactionErr(CanNotDecompressCommitmentFromIndexError, err, cmIndex, shardID, temp)
		}
	}
	return &commitmentProving, nil
}

func parseSndOut(params *TxPrivacyInitParams) *[]*privacy.Scalar {
	ok := true
	sndOuts := make([]*privacy.Scalar, 0)
	for ok {
		for i := 0; i < len(params.paymentInfo); i++ {
			sndOut := privacy.RandomScalar()
			for {
				ok1, err := CheckSNDerivatorExistence(params.tokenID, sndOut, params.stateDB)
				if err != nil {
					Logger.log.Error(err)
				}
				// if sndOut existed, then re-random it
				if ok1 {
					sndOut = privacy.RandomScalar()
				} else {
					break
				}
			}
			sndOuts = append(sndOuts, sndOut)
		}

		// if sndOuts has two elements that have same value, then re-generates it
		ok = privacy.CheckDuplicateScalarArray(sndOuts)
		if ok {
			sndOuts = make([]*privacy.Scalar, 0)
		}
	}
	return &sndOuts
}

func parseOutputCoins(params *TxPrivacyInitParams) (*[]*privacy.OutputCoin, error) {
	sndOuts := *parseSndOut(params)
	outputCoins := make([]*privacy.OutputCoin, len(params.paymentInfo))
	for i, pInfo := range params.paymentInfo {
		outputCoins[i] = new(privacy.OutputCoin)
		outputCoins[i].CoinDetails = new(privacy.CoinV1)
		outputCoins[i].CoinDetails.SetValue(pInfo.Amount)
		if len(pInfo.Message) > 0 {
			if len(pInfo.Message) > privacy.MaxSizeInfoCoin {
				return nil, NewTransactionErr(ExceedSizeInfoOutCoinError, nil)
			}
		}
		outputCoins[i].CoinDetails.SetInfo(pInfo.Message)

		PK, err := new(privacy.Point).FromBytesS(pInfo.PaymentAddress.Pk)
		if err != nil {
			Logger.log.Error(errors.New(fmt.Sprintf("can not decompress public key from %+v", pInfo.PaymentAddress)))
			return nil, NewTransactionErr(DecompressPaymentAddressError, err, pInfo.PaymentAddress)
		}
		outputCoins[i].CoinDetails.SetPublicKey(PK)
		outputCoins[i].CoinDetails.SetSNDerivator(sndOuts[i])
	}
	return &outputCoins, nil
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

	tx.Proof, err = witness.Prove(params.hasPrivacy, params.paymentInfo[i].PaymentAddress.Tk)
	if err.(*errhandler.PrivacyError) != nil {
		Logger.log.Error(err)
		jsonParam, _ := json.MarshalIndent(paymentWitnessParam, common.EmptyString, "  ")
		return NewTransactionErr(WithnessProveError, err, params.hasPrivacy, string(jsonParam))
	}

	// set private key for signing tx
	if params.hasPrivacy {
		randSK := witness.GetRandSecretKey()
		tx.sigPrivKey = append(*params.senderSK, randSK.ToBytesS()...)
	} else {
		tx.sigPrivKey = []byte{}
		randSK := big.NewInt(0)
		tx.sigPrivKey = append(*params.senderSK, randSK.Bytes()...)
	}

	// sign tx
	err = signTx(tx)
	if err != nil {
		Logger.log.Error(err)
		return NewTransactionErr(SignTxError, err)
	}
	return nil
}

// signTx - signs tx
func signTx(tx *Tx) error {
	//Check input transaction
	if tx.Sig != nil {
		return NewTransactionErr(UnexpectedError, errors.New("input transaction must be an unsigned one"))
	}

	/****** using Schnorr signature *******/
	// sign with sigPrivKey
	// prepare private key for Schnorr
	sk := new(privacy.Scalar).FromBytesS(tx.sigPrivKey[:common.BigIntSize])
	r := new(privacy.Scalar).FromBytesS(tx.sigPrivKey[common.BigIntSize:])
	sigKey := new(privacy.SchnorrPrivateKey)
	sigKey.Set(sk, r)

	// save public key for verification signature tx
	tx.SigPubKey = sigKey.GetPublicKey().GetPublicKey().ToBytesS()

	// signing
	if Logger.log != nil {
		Logger.log.Debugf(tx.Hash().String())
	}
	signature, err := sigKey.Sign(tx.Hash()[:])
	if err != nil {
		return err
	}

	// convert signature to byte array
	tx.Sig = signature.Bytes()

	return nil
}

// ValidateTransaction returns true if transaction is valid:
// - Verify tx signature
// - Verify the payment proof
func (*TxVersion1) Verify(tx *Tx, hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
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

	if tx.Proof != nil {
		if tokenID == nil {
			tokenID = &common.Hash{}
			err := tokenID.SetBytes(common.PRVCoinID[:])
			if err != nil {
				Logger.log.Error(err)
				return false, NewTransactionErr(TokenIDInvalidError, err, tokenID.String())
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
				if ok, err := CheckSNDerivatorExistence(tokenID, tx.Proof.GetOutputCoins()[i].CoinDetails.GetSNDerivator(), transactionStateDB); ok || err != nil {
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
				ok, err := tx.CheckCMExistence(tx.Proof.GetInputCoins()[i].CoinDetails.GetCoinCommitment().ToBytesS(), transactionStateDB, shardID, tokenID)
				if !ok || err != nil {
					if err != nil {
						Logger.log.Error(err)
					}
					return false, NewTransactionErr(InputCommitmentIsNotExistedError, err)
				}
			}
		}
		// Verify the payment proof
		valid, err = tx.Proof.Verify(hasPrivacy, tx.SigPubKey, tx.Fee, transactionStateDB, shardID, tokenID, isBatch)
		if !valid {
			if err != nil {
				Logger.log.Error(err)
			}
			Logger.log.Error("FAILED VERIFICATION PAYMENT PROOF")
			err1, ok := err.(*privacy.PrivacyError)
			if ok {
				// parse error detail
				if err1.Code == privacy.ErrCodeMessage[privacy.VerifyOneOutOfManyProofFailedErr].Code {
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
		} else {
			Logger.log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
		}
	}
	//@UNCOMMENT: metrics time
	//elapsed := time.Since(start)
	//Logger.log.Debugf("Validation normal tx %+v in %s time \n", *tx.Hash(), elapsed)

	return true, nil
}
