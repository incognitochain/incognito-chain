package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"math/big"
	"time"

	"github.com/incognitochain/incognito-chain/metadata"
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

type TxVersion1 struct {
	TxBase
}

// ========== CHECK FUNCTION ===========

func (tx *TxVersion1) CheckAuthorizedSender(publicKey []byte) (bool, error) {
	sigPubKey := tx.GetSigPubKey()
	if bytes.Equal(sigPubKey, publicKey) {
		return true, nil
	} else {
		return false, nil
	}
}

// ========== NORMAL INIT FUNCTIONS ==========

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
				ok1, err := checkSNDerivatorExistence(tokenID, sndOut, stateDB)
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

func (tx *TxVersion1) Init(paramsInterface interface{}) error {
	params, ok := paramsInterface.(*TxPrivacyInitParams)
	if !ok {
		return errors.New("params of tx Init is not TxPrivacyInitParam")
	}

	Logger.Log.Debugf("CREATING TX........\n")
	if err := validateTxParams(params); err != nil {
		return err
	}

	// Init tx and params (tx and params will be changed)
	if err := tx.initializeTxAndParams(params); err != nil {
		return err
	}

	// Check if this tx is nonPrivacyNonInput
	// Case 1: tx ptoken transfer with ptoken fee
	// Case 2: tx Reward
	if check, err := tx.isNonPrivacyNonInput(params); check {
		return err
	}

	if err := tx.prove(params); err != nil {
		return err
	}
	return nil
}

func (tx *TxVersion1) initializePaymentWitnessParam(params *TxPrivacyInitParams) (*zkp.PaymentWitnessParam, error) {
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
		PrivateKey:              new(operation.Scalar).FromBytesS(*params.senderSK),
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

func (tx *TxVersion1) proveAndSignCore(params *TxPrivacyInitParams, paymentWitnessParamPtr *zkp.PaymentWitnessParam) error {
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
	signErr := tx.sign()
	if signErr != nil {
		Logger.Log.Error(err)
		return NewTransactionErr(SignTxError, err)
	}
	return nil
}

func (tx *TxVersion1) prove(params *TxPrivacyInitParams) error {
	// Prepare paymentWitness params
	paymentWitnessParamPtr, err := tx.initializePaymentWitnessParam(params)
	if err != nil {
		return err
	}
	return tx.proveAndSignCore(params, paymentWitnessParamPtr)
}

func (tx *TxVersion1) initializePaymentWitnessParamASM(params *TxPrivacyInitParamsForASM) (*zkp.PaymentWitnessParam, error) {
	// create SNDs for output coins
	txParams := &params.txParam
	outputCoins, err := parseOutputCoins(txParams.paymentInfo, txParams.tokenID, txParams.stateDB)
	if err != nil {
		return nil, err
	}
	// get list of commitments for proving one-out-of-many from commitmentIndexs
	shardID := common.GetShardIDFromLastByte(tx.PubKeyLastByteSender)
	commitmentProving := make([]*operation.Point, len(params.commitmentBytes))
	for i, cmBytes := range params.commitmentBytes {
		commitmentProving[i] = new(operation.Point)
		commitmentProving[i], err = commitmentProving[i].FromBytesS(cmBytes)
		if err != nil {
			Logger.Log.Error(errors.New(fmt.Sprintf("ASM: Can not get commitment from index=%d shardID=%+v value=%+v", params.commitmentIndices[i], shardID, cmBytes)))
			return nil, NewTransactionErr(CanNotDecompressCommitmentFromIndexError, err, params.commitmentIndices[i], shardID, cmBytes)
		}
	}

	// prepare witness for proving
	paymentWitnessParam := zkp.PaymentWitnessParam{
		HasPrivacy:              params.txParam.hasPrivacy,
		PrivateKey:              new(operation.Scalar).FromBytesS(*params.txParam.senderSK),
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

func (tx *TxVersion1) sign() error {
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

func (tx *TxVersion1) proveASM(params *TxPrivacyInitParamsForASM) error {
	paymentWitnessParamPtr, err := tx.initializePaymentWitnessParamASM(params)
	if err != nil {
		return err
	}
	return tx.proveAndSignCore(&params.txParam, paymentWitnessParamPtr)
}

// ========== NORMAL VERIFY FUNCTIONS ==========

func validateSndFromOutputCoin(outputCoins []*coin.CoinV1) error {
	sndOutputs := make([]*operation.Scalar, len(outputCoins))
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

func (tx *TxVersion1) verifySig() (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, NewTransactionErr(UnexpectedError, errors.New("input transaction must be an signed one"))
	}

	/****** verify Schnorr signature *****/
	// prepare Public key for verification
	verifyKey := new(privacy.SchnorrPublicKey)
	sigPublicKey, err := new(operation.Point).FromBytesS(tx.SigPubKey)
	if err != nil {
		Logger.Log.Error(err)
		return false, NewTransactionErr(DecompressSigPubKeyError, err)
	}
	verifyKey.Set(sigPublicKey)

	// convert signature from byte array to SchnorrSign
	signature := new(privacy.SchnSignature)
	if err = signature.SetBytes(tx.Sig); err != nil {
		Logger.Log.Error(err)
		return false, NewTransactionErr(InitTxSignatureFromBytesError, err)
	}
	res := verifyKey.Verify(signature, tx.Hash()[:])
	return res, nil
}

func (tx *TxVersion1) Verify(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	var err error
	if valid, err := tx.verifySig(); !valid {
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
			if ok, err := checkSNDerivatorExistence(tokenID, outputCoins[i].GetSNDerivator(), transactionStateDB); ok || err != nil {
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
			ok, err := statedb.HasCommitment(transactionStateDB, *tokenID, inputCoins[i].GetCommitment().ToBytesS(), shardID)
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

// ========== SALARY FUNCTIONS: INIT AND VALIDATE  ==========

func (tx *TxVersion1) InitTxSalary(salary uint64, receiverAddr *privacy.PaymentAddress, privKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata,
) error {
	tx.Version = txVersion1Number
	tx.Type = common.TxRewardType
	if tx.LockTime == 0 {
		tx.LockTime = time.Now().Unix()
	}

	var err error
	// create new output coins with info: Pk, value, input, randomness, last byte pk, coin commitment
	tx.Proof = new(zkp.PaymentProof)
	tempOutputCoin := make([]*coin.CoinV1, 1)
	tempOutputCoin[0] = new(coin.CoinV1).Init()
	publicKey, err := new(operation.Point).FromBytesS(receiverAddr.Pk)
	if err != nil {
		return err
	}
	tempOutputCoin[0].CoinDetails.SetPublicKey(publicKey)
	tempOutputCoin[0].CoinDetails.SetValue(salary)
	tempOutputCoin[0].CoinDetails.SetRandomness(privacy.RandomScalar())

	sndOut := privacy.RandomScalar()
	for {
		tokenID := &common.Hash{}
		err := tokenID.SetBytes(common.PRVCoinID[:])
		if err != nil {
			return NewTransactionErr(TokenIDInvalidError, err, tokenID.String())
		}
		ok, err := checkSNDerivatorExistence(tokenID, sndOut, stateDB)
		if err != nil {
			return NewTransactionErr(SndExistedError, err)
		}
		if ok {
			sndOut = privacy.RandomScalar()
		} else {
			break
		}
	}
	tempOutputCoin[0].CoinDetails.SetSNDerivator(sndOut)
	if err = tempOutputCoin[0].CoinDetails.CommitAll(); err != nil {
		return NewTransactionErr(CommitOutputCoinError, err)
	}
	tx.Proof.SetOutputCoins(coin.ArrayCoinV1ToCoin(tempOutputCoin))
	tx.PubKeyLastByteSender = receiverAddr.Pk[len(receiverAddr.Pk)-1]

	// sign Tx
	tx.SigPubKey = receiverAddr.Pk
	tx.sigPrivKey = *privKey
	tx.SetMetadata(metaData)
	err = tx.sign()
	if err != nil {
		return NewTransactionErr(SignTxError, err)
	}
	return nil
}

func (tx TxVersion1) ValidateTxSalary(
	db *statedb.StateDB,
) (bool, error) {
	// verify signature
	valid, err := tx.verifySig()
	if !valid {
		if err != nil {
			Logger.Log.Debugf("Error verifying signature of tx: %+v", err)
			return false, NewTransactionErr(VerifyTxSigFailError, err)
		}
		return false, nil
	}

	// check whether output coin's input exists in input list or not
	tokenID := &common.Hash{}
	err = tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, NewTransactionErr(TokenIDInvalidError, err, tokenID.String())
	}
	if ok, err := checkSNDerivatorExistence(tokenID, tx.Proof.GetOutputCoins()[0].GetSNDerivator(), db); ok || err != nil {
		return false, err
	}

	// check output coin's coin commitment is calculated correctly
	coin := tx.Proof.GetOutputCoins()[0]
	shardID, err := coin.GetShardID()
	if err != nil {
		Logger.Log.Errorf("Cannot get shardID from coin %v", err)
		return false, err
	}

	cmTmp2 := new(privacy.Point)
	cmTmp2.Add(coin.GetPublicKey(), new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenValueIndex], new(privacy.Scalar).FromUint64(uint64(coin.GetValue()))))
	cmTmp2.Add(cmTmp2, new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenSndIndex], coin.GetSNDerivator()))
	cmTmp2.Add(cmTmp2, new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenShardIDIndex], new(privacy.Scalar).FromUint64(uint64(shardID))))
	cmTmp2.Add(cmTmp2, new(privacy.Point).ScalarMult(privacy.PedCom.G[privacy.PedersenRandomnessIndex], coin.GetRandomness()))

	ok := operation.IsPointEqual(cmTmp2, tx.Proof.GetOutputCoins()[0].GetCommitment())
	if !ok {
		return ok, NewTransactionErr(UnexpectedError, errors.New("check output coin's coin commitment isn't calculated correctly"))
	}
	return ok, nil
}

func (tx TxVersion1) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	check, err := tx.TxBase.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight)
	if err != nil {
		Logger.Log.Errorf("TxVersion1 error when ValidateSanityDataInterface")
		return false, err
	}
	if !check {
		Logger.Log.Errorf("TxVersion1 ValidateSanityData got fail check")
		return false, errors.New("TxVersion1 ValidateSanityData got fail check")
	}
	if len(tx.SigPubKey) != common.SigPubKeySize {
		return false, NewTransactionErr(RejectTxPublickeySigSize, fmt.Errorf("wrong tx Sig PK size %d", len(tx.SigPubKey)))
	}
	return true, nil
}