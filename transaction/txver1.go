package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"math/big"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/coin"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/privacy/key"
	"github.com/incognitochain/incognito-chain/privacy/operation"
	"github.com/incognitochain/incognito-chain/privacy/privacy_util"
	zkp "github.com/incognitochain/incognito-chain/privacy/privacy_v1/zeroknowledge"
)

type TxVersion1 struct{}

func (*TxVersion1) CheckAuthorizedSender(tx *Tx, publicKey []byte) (bool, error) {
	sigPubKey := tx.GetSigPubKey()
	if bytes.Equal(sigPubKey, publicKey) {
		return true, nil
	} else {
		return false, nil
	}
}

func parseCommitments(params *TxPrivacyInitParams, shardID byte) ([]uint64, []uint64, error) {
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
	return commitmentIndexs, myCommitmentIndexs, nil
}

func parseCommitmentProving(params *TxPrivacyInitParams, shardID byte, commitmentIndexs []uint64) ([]*operation.Point, error) {
	commitmentProving := make([]*privacy.Point, len(commitmentIndexs))
	for i, cmIndex := range commitmentIndexs {
		temp, err := statedb.GetCommitmentByIndex(params.stateDB, *params.tokenID, cmIndex, shardID)
		if err != nil {
			Logger.Log.Error(errors.New(fmt.Sprintf("can not get commitment from index=%d shardID=%+v", cmIndex, shardID)))
			return nil, NewTransactionErr(CanNotGetCommitmentFromIndexError, err, cmIndex, shardID)
		}
		commitmentProving[i] = new(privacy.Point)
		commitmentProving[i], err = commitmentProving[i].FromBytesS(temp)
		if err != nil {
			Logger.Log.Error(errors.New(fmt.Sprintf("can not get commitment from index=%d shardID=%+v value=%+v", cmIndex, shardID, temp)))
			return nil, NewTransactionErr(CanNotDecompressCommitmentFromIndexError, err, cmIndex, shardID, temp)
		}
	}
	return commitmentProving, nil
}

func generateSndOut(paymentInfo []*privacy.PaymentInfo, tokenID *common.Hash, stateDB *statedb.StateDB) []*operation.Scalar {
	ok := true
	sndOuts := make([]*operation.Scalar, 0)
	for ok {
		for i := 0; i < len(paymentInfo); i++ {
			sndOut := operation.RandomScalar()
			for {
				ok1, err := CheckSNDerivatorExistence(tokenID, sndOut, stateDB)
				if err != nil {
					Logger.Log.Error(err)
				}
				// if sndOut existed, then re-random it
				if ok1 {
					sndOut = operation.RandomScalar()
				} else {
					break
				}
			}
			sndOuts = append(sndOuts, sndOut)
		}

		// if sndOuts has two elements that have same value, then re-generates it
		ok = operation.CheckDuplicateScalarArray(sndOuts)
		if ok {
			sndOuts = make([]*operation.Scalar, 0)
		}
	}
	return sndOuts
}

func parseOutputCoins(paymentInfo []*privacy.PaymentInfo, tokenID *common.Hash, stateDB *statedb.StateDB) ([]*coin.CoinV1, error) {
	sndOuts := generateSndOut(paymentInfo, tokenID, stateDB)
	outputCoins := make([]*coin.CoinV1, len(paymentInfo))
	for i, pInfo := range paymentInfo {
		outputCoins[i] = new(coin.CoinV1)
		outputCoins[i].CoinDetails = new(coin.PlainCoinV1)
		outputCoins[i].CoinDetails.SetValue(pInfo.Amount)
		if len(pInfo.Message) > 0 {
			if len(pInfo.Message) > privacy.MaxSizeInfoCoin {
				return nil, NewTransactionErr(ExceedSizeInfoOutCoinError, nil)
			}
		}
		outputCoins[i].CoinDetails.SetInfo(pInfo.Message)

		PK, err := new(operation.Point).FromBytesS(pInfo.PaymentAddress.Pk)
		if err != nil {
			Logger.Log.Error(errors.New(fmt.Sprintf("can not decompress public key from %+v", pInfo.PaymentAddress)))
			return nil, NewTransactionErr(DecompressPaymentAddressError, err, pInfo.PaymentAddress)
		}
		outputCoins[i].CoinDetails.SetPublicKey(PK)
		outputCoins[i].CoinDetails.SetSNDerivator(sndOuts[i])
	}
	return outputCoins, nil
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
	outputCoins, err := parseOutputCoins(params.paymentInfo, params.tokenID, params.stateDB)
	if err != nil {
		return nil, err
	}

	// prepare witness for proving
	paymentWitnessParam := zkp.PaymentWitnessParam{
		HasPrivacy:              params.hasPrivacy,
		PrivateKey:              new(privacy.Scalar).FromBytesS(*params.senderSK),
		InputCoins:              params.inputCoins,
		OutputCoins:             outputCoins,
		PublicKeyLastByteSender: tx.PubKeyLastByteSender,
		Commitments:             commitmentProving,
		CommitmentIndices:       commitmentIndexs,
		MyCommitmentIndices:     myCommitmentIndexs,
		Fee:                     params.fee,
	}
	return &paymentWitnessParam, nil
}

func proveAndSignCore(tx *Tx, params *TxPrivacyInitParams, paymentWitnessParamPtr *zkp.PaymentWitnessParam) error {
	paymentWitnessParam := *paymentWitnessParamPtr
	witness := new(zkp.PaymentWitness)
	err := witness.Init(paymentWitnessParam)
	if err != nil {
		Logger.Log.Error(err)
		jsonParam, _ := json.MarshalIndent(paymentWitnessParam, common.EmptyString, "  ")
		return NewTransactionErr(InitWithnessError, err, string(jsonParam))
	}

	paymentProof, err := witness.Prove(params.hasPrivacy, params.paymentInfo)
	if err != nil {
		Logger.Log.Error(err)
		jsonParam, _ := json.MarshalIndent(paymentWitnessParam, common.EmptyString, "  ")
		return NewTransactionErr(WithnessProveError, err, params.hasPrivacy, string(jsonParam))
	}
	tx.Proof = paymentProof

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
	signErr := signTx(tx)
	if signErr != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(SignTxError, err)
	}
	return nil
}

func (*TxVersion1) Prove(tx *Tx, params *TxPrivacyInitParams) error {
	// Prepare paymentWitness params
	paymentWitnessParamPtr, err := initializePaymentWitnessParam(tx, params)
	if err != nil {
		return err
	}
	return proveAndSignCore(tx, params, paymentWitnessParamPtr)
}

func initializePaymentWitnessParamASM(tx *Tx, params *TxPrivacyInitParamsForASM) (*zkp.PaymentWitnessParam, error) {
	// create SNDs for output coins
	txParams := &params.txParam
	outputCoins, err := parseOutputCoins(txParams.paymentInfo, txParams.tokenID, txParams.stateDB)
	if err != nil {
		return nil, err
	}
	// get list of commitments for proving one-out-of-many from commitmentIndexs
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)
	commitmentProving := make([]*privacy.Point, len(params.commitmentBytes))
	for i, cmBytes := range params.commitmentBytes {
		commitmentProving[i] = new(privacy.Point)
		commitmentProving[i], err = commitmentProving[i].FromBytesS(cmBytes)
		if err != nil {
			Logger.Log.Error(errors.New(fmt.Sprintf("ASM: Can not get commitment from index=%d shardID=%+v value=%+v", params.commitmentIndices[i], shardID, cmBytes)))
			return nil, NewTransactionErr(CanNotDecompressCommitmentFromIndexError, err, params.commitmentIndices[i], shardID, cmBytes)
		}
	}

	// prepare witness for proving
	paymentWitnessParam := zkp.PaymentWitnessParam{
		HasPrivacy:              params.txParam.hasPrivacy,
		PrivateKey:              new(privacy.Scalar).FromBytesS(*params.txParam.senderSK),
		InputCoins:              params.txParam.inputCoins,
		OutputCoins:             outputCoins,
		PublicKeyLastByteSender: tx.PubKeyLastByteSender,
		Commitments:             commitmentProving,
		CommitmentIndices:       params.commitmentIndices,
		MyCommitmentIndices:     params.myCommitmentIndices,
		Fee:                     params.txParam.fee,
	}
	return &paymentWitnessParam, nil
}

func (*TxVersion1) ProveASM(tx *Tx, params *TxPrivacyInitParamsForASM) error {
	paymentWitnessParamPtr, err := initializePaymentWitnessParamASM(tx, params)
	if err != nil {
		return err
	}
	return proveAndSignCore(tx, &params.txParam, paymentWitnessParamPtr)
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
	sk := new(operation.Scalar).FromBytesS(tx.sigPrivKey[:common.BigIntSize])
	r := new(operation.Scalar).FromBytesS(tx.sigPrivKey[common.BigIntSize:])
	sigKey := new(schnorr.SchnorrPrivateKey)
	sigKey.Set(sk, r)

	// save public key for verification signature tx
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

func validateSndFromOutputCoin(outputCoins []*coin.CoinV1) error {
	sndOutputs := make([]*privacy.Scalar, len(outputCoins))
	for i := 0; i < len(outputCoins); i++ {
		sndOutputs[i] = outputCoins[i].GetSNDerivator()
	}
	if operation.CheckDuplicateScalarArray(sndOutputs) {
		Logger.Log.Errorf("Duplicate output coins' snd\n")
		return NewTransactionErr(DuplicatedOutputSndError, errors.New("Duplicate output coins' snd\n"))
	}
	return nil
}

func getCommitmentsInDatabase(
	proof *privacy.ProofV1, hasPrivacy bool,
	pubKey key.PublicKey, transactionStateDB *statedb.StateDB,
	shardID byte, tokenID *common.Hash, isBatch bool) (*[][privacy_util.CommitmentRingSize]*operation.Point, error) {

	// verify for input coins
	oneOfManyProof := proof.GetOneOfManyProof()
	commitmentIndices := proof.GetCommitmentIndices()
	commitmentInputSND := proof.GetCommitmentInputSND()
	commitmentInputValue := proof.GetCommitmentInputValue()
	commitmentInputShardID := proof.GetCommitmentInputShardID()
	commitmentInputSecretKey := proof.GetCommitmentInputSecretKey()

	const sz int = privacy_util.CommitmentRingSize
	commitments := make([][sz]*operation.Point, len(oneOfManyProof))
	for i := 0; i < len(oneOfManyProof); i++ {
		cmInputSum := new(operation.Point).Add(commitmentInputSecretKey, commitmentInputValue[i])
		cmInputSum.Add(cmInputSum, commitmentInputSND[i])
		cmInputSum.Add(cmInputSum, commitmentInputShardID)

		for j := 0; j < sz; j++ {
			index := commitmentIndices[i*privacy_util.CommitmentRingSize+j]
			commitmentBytes, err := statedb.GetCommitmentByIndex(transactionStateDB, *tokenID, index, shardID)
			if err != nil {
				Logger.Log.Errorf("GetCommitmentInDatabase: Error when getCommitmentByIndex from database", index, err)
				return nil, NewTransactionErr(GetCommitmentsInDatabaseError, err)
			}
			recheckIndex, err := statedb.GetCommitmentIndex(transactionStateDB, *tokenID, commitmentBytes, shardID)
			if err != nil || recheckIndex.Uint64() != index {
				Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Error when getCommitmentIndex from database", index, err)
				return nil, NewTransactionErr(GetCommitmentsInDatabaseError, err)
			}

			commitments[i][j], err = new(operation.Point).FromBytesS(commitmentBytes)
			if err != nil {
				Logger.Log.Errorf("VERIFICATION PAYMENT PROOF: Cannot decompress commitment from database", index, err)
				return nil, errhandler.NewPrivacyErr(VerifyOneOutOfManyProofFailedErr, err)
			}
			commitments[i][j].Sub(commitments[i][j], cmInputSum)
		}
	}
	return &commitments, nil
}

// verifySigTx - verify signature on tx
func verifySigTx(tx *Tx) (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be an signed one"))
	}

	var err error
	res := false

	/****** verify Schnorr signature *****/
	// prepare Public key for verification
	verifyKey := new(privacy.SchnorrPublicKey)
	sigPublicKey, err := new(privacy.Point).FromBytesS(tx.SigPubKey)

	if err != nil {
		Logger.Log.Error(err)
		return false, NewTransactionErr(DecompressSigPubKeyError, err)
	}
	verifyKey.Set(sigPublicKey)

	// convert signature from byte array to SchnorrSign
	signature := new(privacy.SchnSignature)
	err = signature.SetBytes(tx.Sig)
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
	res = verifyKey.Verify(signature, tx.Hash()[:])

	return res, nil
}

// ValidateTransaction returns true if transaction is valid:
// - Verify tx signature
// - Verify the payment proof
func (*TxVersion1) Verify(tx *Tx, hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	var err error
	if valid, err := verifySigTx(tx); !valid {
		if err != nil {
			Logger.Log.Errorf("Error verifying signature ver1 with tx hash %s: %+v \n", tx.Hash().String(), err)
			return false, NewTransactionErr(VerifyTxSigFailError, err)
		}
		Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE ver1 with tx hash %s", tx.Hash().String())
		return false, NewTransactionErr(VerifyTxSigFailError, fmt.Errorf("FAILED VERIFICATION SIGNATURE ver1 with tx hash %s", tx.Hash().String()))
	}

	if tx.Proof == nil {
		return true, nil
	}

	tokenID, err = parseTokenID(tokenID)
	if err != nil {
		return false, err
	}
	inputCoins := tx.Proof.GetInputCoins()
	outputCoins := tx.Proof.GetOutputCoins()
	outputCoinsV2 := coin.ArrayCoinToCoinV1(outputCoins)
	if err := validateSndFromOutputCoin(outputCoinsV2); err != nil {
		return false, err
	}

	if isNewTransaction {
		for i := 0; i < len(outputCoins); i++ {
			// Check output coins' SND is not exists in SND list (Database)
			if ok, err := CheckSNDerivatorExistence(tokenID, outputCoins[i].GetSNDerivator(), transactionStateDB); ok || err != nil {
				if err != nil {
					Logger.Log.Error(err)
				}
				Logger.Log.Errorf("snd existed: %d\n", i)
				return false, NewTransactionErr(SndExistedError, err, fmt.Sprintf("snd existed: %d\n", i))
			}
		}
	}
	if !hasPrivacy {
		// Check input coins' commitment is exists in cm list (Database)
		for i := 0; i < len(inputCoins); i++ {
			ok, err := tx.CheckCMExistence(inputCoins[i].GetCommitment().ToBytesS(), transactionStateDB, shardID, tokenID)
			if !ok || err != nil {
				if err != nil {
					Logger.Log.Error(err)
				}
				return false, NewTransactionErr(InputCommitmentIsNotExistedError, err)
			}
		}
	}
	// Verify the payment proof
	txProofV1 := tx.Proof.(*privacy.ProofV1)
	commitments, err := getCommitmentsInDatabase(txProofV1, hasPrivacy, tx.SigPubKey, transactionStateDB, shardID, tokenID, isBatch)
	if err != nil {
		return false, err
	}

	if valid, err := tx.Proof.Verify(hasPrivacy, tx.SigPubKey, tx.Fee, shardID, tokenID, isBatch, commitments); !valid {
		if err != nil {
			Logger.Log.Error(err)
		}
		Logger.Log.Error("FAILED VERIFICATION PAYMENT PROOF")
		err1, ok := err.(*privacy.PrivacyError)
		if ok {
			// parse error detail
			if err1.Code == privacy.ErrCodeMessage[errhandler.VerifyOneOutOfManyProofFailedErr].Code {
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
	Logger.Log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
	return true, nil
}
