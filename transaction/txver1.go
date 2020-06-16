package transaction

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v1/schnorr"
	"github.com/incognitochain/incognito-chain/wallet"
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

// ========== GET FUNCTION ============

func (tx TxVersion1) GetTxMintData() (bool, []byte, []byte, uint64, *common.Hash, error) {
	publicKeys, txRandoms, amounts, err := tx.GetReceiverData()
	if err != nil {
		Logger.Log.Error("GetTxMintData: Cannot get receiver data")
		return false, nil, nil, 0, nil, err
	}
	if len(publicKeys) != 1 {
		Logger.Log.Error("GetTxMintData : Should only have one receiver")
		return false, nil, nil, 0, nil, errors.New("Error Tx mint has more than one receiver")
	}
	if inputCoins := tx.Proof.GetInputCoins(); len(inputCoins) > 0 {
		return false, nil, nil, 0, nil, errors.New("Error this is not Tx mint")
	}
	if txRandoms == nil {
		return true, publicKeys[0].ToBytesS(), nil, amounts[0], &common.PRVCoinID, nil
	} else {
		return true, publicKeys[0].ToBytesS(), txRandoms[0].Bytes(), amounts[0], &common.PRVCoinID, nil
	}
}

func (tx *TxVersion1) GetReceiverData() ([]*privacy.Point, []*coin.TxRandom, []uint64, error) {
	pubkeys := make([]*privacy.Point, 0)
	amounts := []uint64{}

	if tx.Proof != nil && len(tx.Proof.GetOutputCoins()) > 0 {
		for _, coin := range tx.Proof.GetOutputCoins() {
			coinPubKey := coin.GetPublicKey()
			added := false
			for i, key := range pubkeys {
				if bytes.Equal(coinPubKey.ToBytesS(), key.ToBytesS()) {
					added = true
					amounts[i] += coin.GetValue()
					break
				}
			}
			if !added {
				pubkeys = append(pubkeys, coinPubKey)
				amounts = append(amounts, coin.GetValue())
			}
		}
	}
	return pubkeys, nil, amounts, nil
}

func (tx TxVersion1) GetTxBurnData(retriever metadata.ChainRetriever, blockHeight uint64) (bool, []byte, uint64, *common.Hash, error) {
	pubkeys, _, amounts, err := tx.GetReceiverData()
	if err != nil {
		Logger.Log.Errorf("Cannot get receiver data, error %v", err)
		return false, nil, 0, nil, err
	}
	if len(pubkeys) > 2 {
		Logger.Log.Error("GetAndCheckBurning receiver: More than 2 receivers")
		return false, nil, 0, nil, err
	}

	burnAccount, err := wallet.Base58CheckDeserialize(retriever.GetBurningAddress(blockHeight))
	if err != nil {
		return false, nil, 0, nil, err
	}
	burnPaymentAddress := burnAccount.KeySet.PaymentAddress

	isBurned, pubkey, amount := false, []byte{}, uint64(0)
	for i, pk := range pubkeys {
		if bytes.Equal(burnPaymentAddress.Pk, pk.ToBytesS()) {
			isBurned = true
			pubkey = pk.ToBytesS()
			amount += amounts[i]
		}
	}
	return isBurned, pubkey, amount, &common.PRVCoinID, nil
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

// checkSNDerivatorExistence return true if snd exists in snDerivators list
func checkSNDerivatorExistence(tokenID *common.Hash, snd *privacy.Scalar, stateDB *statedb.StateDB) (bool, error) {
	ok, err := txDatabaseWrapper.hasSNDerivator(stateDB, *tokenID, snd.ToBytesS())
	if err != nil {
		return false, err
	}
	return ok, nil
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
		temp, err := txDatabaseWrapper.getCommitmentByIndex(params.stateDB, *params.tokenID, cmIndex, shardID)
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

func (tx TxVersion1) ValidateSanityData(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, beaconHeight uint64) (bool, error) {
	meta := tx.GetMetadata()
	Logger.Log.Debugf("\n\n\n START Validating sanity data of metadata %+v\n\n\n", meta)
	if meta != nil {
		Logger.Log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := meta.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, &tx)
		Logger.Log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	Logger.Log.Debugf("\n\n\n END sanity data of metadata%+v\n\n\n")
	//check version
	if tx.GetVersion() > TxVersion2Number {
		return false, NewTransactionErr(RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version <= %d", tx.GetVersion(), currentTxVersion))
	}
	// check LockTime before now
	if int64(tx.GetLockTime()) > time.Now().Unix() {
		return false, NewTransactionErr(RejectInvalidLockTime, fmt.Errorf("wrong tx locktime %d", tx.GetLockTime()))
	}

	// check tx size
	actualTxSize := tx.GetTxActualSize()
	if actualTxSize > common.MaxTxSize {
		return false, NewTransactionErr(RejectTxSize, fmt.Errorf("tx size %d kB is too large", actualTxSize))
	}

	// check sanity of Proof
	if tx.GetProof() != nil {
		ok, err := tx.GetProof().ValidateSanity()
		if !ok || err != nil {
			s := ""
			if !ok {
				s = "ValidateSanity Proof got error: ok = false\n"
			} else {
				s = fmt.Sprintf("ValidateSanity Proof got error: error %s\n", err.Error())
			}
			return false, errors.New(s)
		}
	}

	// check Type is normal or salary tx
	switch tx.GetType() {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenPrivacyType, common.TxReturnStakingType: //is valid
	default:
		return false, NewTransactionErr(RejectTxType, fmt.Errorf("wrong tx type with %s", tx.GetType()))
	}

	//if txN.Type != common.TxNormalType && txN.Type != common.TxRewardType && txN.Type != common.TxCustomTokenType && txN.Type != common.TxCustomTokenPrivacyType { // only 1 byte
	//	return false, errors.New("wrong tx type")
	//}

	// check info field
	info := tx.GetInfo()
	if len(info) > 512 {
		return false, NewTransactionErr(RejectTxInfoSize, fmt.Errorf("wrong tx info length %d bytes, only support info with max length <= %d bytes", len(info), 512))
	}

	if len(tx.SigPubKey) != common.SigPubKeySize {
		return false, NewTransactionErr(RejectTxPublickeySigSize, fmt.Errorf("wrong tx Sig PK size %d", len(tx.SigPubKey)))
	}

	return true, nil
}

func (tx TxVersion1) ValidateTxWithBlockChain(chainRetriever metadata.ChainRetriever, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever, shardID byte, stateDB *statedb.StateDB) error {
	if tx.GetType() == common.TxRewardType || tx.GetType() == common.TxReturnStakingType {
		return nil
	}
	meta := tx.GetMetadata()
	if meta != nil {
		isContinued, err := meta.ValidateTxWithBlockChain(&tx, chainRetriever, shardViewRetriever, beaconViewRetriever, shardID, stateDB)
		fmt.Printf("[transactionStateDB] validate metadata with blockchain: %d %h %t %v\n", tx.GetMetadataType(), tx.Hash(), isContinued, err)
		if err != nil {
			Logger.Log.Errorf("[db] validate metadata with blockchain: %d %s %t %v\n", tx.GetMetadataType(), tx.Hash().String(), isContinued, err)
			return NewTransactionErr(RejectTxMedataWithBlockChain, fmt.Errorf("validate metadata of tx %s with blockchain error %+v", tx.Hash().String(), err))
		}
		if !isContinued {
			return nil
		}
	}
	return tx.ValidateDoubleSpendWithBlockchain(shardID, stateDB, nil)
}

func (tx TxVersion1) ValidateTransaction(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	if tx.GetType() == common.TxRewardType {
		return tx.ValidateTxSalary(transactionStateDB)
	}
	if tx.GetType() == common.TxReturnStakingType {
		return tx.ValidateTxReturnStaking(transactionStateDB), nil
	}
	if tx.Version == TxConversionVersion12Number {
		return validateConversionVer1ToVer2(&tx, transactionStateDB, shardID, tokenID)
	}
	return tx.Verify(hasPrivacy, transactionStateDB, bridgeStateDB, shardID, tokenID, isBatch, isNewTransaction)
}

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
			commitmentBytes, err := txDatabaseWrapper.getCommitmentByIndex(transactionStateDB, *tokenID, index, shardID)
			if err != nil {
				Logger.Log.Errorf("GetCommitmentInDatabase: Error when getCommitmentByIndex from database", index, err)
				return nil, NewTransactionErr(GetCommitmentsInDatabaseError, err)
			}
			recheckIndex, err := txDatabaseWrapper.getCommitmentIndex(transactionStateDB, *tokenID, commitmentBytes, shardID)
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
			ok, err := txDatabaseWrapper.hasCommitment(transactionStateDB, *tokenID, inputCoins[i].GetCommitment().ToBytesS(), shardID)
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

func (tx TxVersion1) VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock []metadata.Transaction, txsUsed []int, insts [][]string, instsUsed []int, shardID byte, bcr metadata.ChainRetriever, accumulatedValues *metadata.AccumulatedValues, retriever metadata.ShardViewRetriever, viewRetriever metadata.BeaconViewRetriever) (bool, error) {
	if tx.IsPrivacy() {
		return true, nil
	}
	proof := tx.GetProof()
	meta := tx.GetMetadata()

	inputCoins := make([]coin.PlainCoin, 0)
	outputCoins := make([]coin.Coin, 0)
	if tx.GetProof() != nil {
		inputCoins = tx.GetProof().GetInputCoins()
		outputCoins = tx.GetProof().GetOutputCoins()
	}
	if proof != nil && len(inputCoins) == 0 && len(outputCoins) > 0 { // coinbase tx
		if meta == nil {
			return false, nil
		}
		if !meta.IsMinerCreatedMetaType() {
			return false, nil
		}
	}
	if meta != nil {
		ok, err := meta.VerifyMinerCreatedTxBeforeGettingInBlock(txsInBlock, txsUsed, insts, instsUsed, shardID, &tx, bcr, accumulatedValues, nil, nil)
		if err != nil {
			Logger.Log.Error(err)
			return false, NewTransactionErr(VerifyMinerCreatedTxBeforeGettingInBlockError, err)
		}
		return ok, nil
	}
	return true, nil
}

func (tx TxVersion1) ValidateTxByItself(hasPrivacy bool, transactionStateDB *statedb.StateDB, bridgeStateDB *statedb.StateDB, chainRetriever metadata.ChainRetriever, shardID byte, isNewTransaction bool, shardViewRetriever metadata.ShardViewRetriever, beaconViewRetriever metadata.BeaconViewRetriever) (bool, error) {
	prvCoinID := &common.Hash{}
	err := prvCoinID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, err
	}
	ok, err := tx.ValidateTransaction(hasPrivacy, transactionStateDB, bridgeStateDB, shardID, prvCoinID, false, isNewTransaction)
	if !ok {
		return false, err
	}
	meta := tx.GetMetadata()
	if meta != nil {
		if hasPrivacy {
			return false, errors.New("Metadata can not exist in not privacy tx")
		}
		validateMetadata := meta.ValidateMetadataByItself()
		if validateMetadata {
			return validateMetadata, nil
		} else {
			return validateMetadata, NewTransactionErr(UnexpectedError, errors.New("Metadata is invalid"))
		}
	}
	return true, nil
}

// ========== SALARY FUNCTIONS: INIT AND VALIDATE  ==========

func (tx *TxVersion1) InitTxSalary(salary uint64, receiverAddr *privacy.PaymentAddress, privKey *privacy.PrivateKey, stateDB *statedb.StateDB, metaData metadata.Metadata,
) error {
	tx.Version = TxVersion1Number
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