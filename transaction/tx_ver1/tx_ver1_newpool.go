package tx_ver1

import (
	"fmt"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/transaction/utils"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

// VerifySigTx - verify signature on tx
func (tx *Tx) VerifySigTx() (bool, error) {
	// check input transaction
	if tx.Sig == nil || tx.SigPubKey == nil {
		return false, utils.NewTransactionErr(utils.UnexpectedError, errors.New("input transaction must be an signed one"))
	}

	var err error
	res := false

	/****** verify Schnorr signature *****/
	// prepare Public key for verification
	verifyKey := new(privacy.SchnorrPublicKey)
	sigPublicKey, err := new(privacy.Point).FromBytesS(tx.SigPubKey)

	if err != nil {
		utils.Logger.Log.Error(err)
		return false, utils.NewTransactionErr(utils.DecompressSigPubKeyError, err)
	}
	verifyKey.Set(sigPublicKey)

	// convert signature from byte array to SchnorrSign
	signature := new(privacy.SchnSignature)
	err = signature.SetBytes(tx.Sig)
	if err != nil {
		utils.Logger.Log.Error(err)
		return false, utils.NewTransactionErr(utils.InitTxSignatureFromBytesError, err)
	}

	// verify signature
	/*Logger.log.Debugf(" VERIFY SIGNATURE ----------- HASH: %v\n", tx.Hash()[:])
	if tx.Proof != nil {
		Logger.log.Debugf(" VERIFY SIGNATURE ----------- TX Proof bytes before verifing the signature: %v\n", tx.Proof.Bytes())
	}
	Logger.log.Debugf(" VERIFY SIGNATURE ----------- TX meta: %v\n", tx.Metadata)*/
	res = verifyKey.Verify(signature, tx.Hash()[:])
	if !res {
		err = fmt.Errorf("Verify signature of tx %v failed", tx.Hash().String())
		utils.Logger.Log.Error(err)
	}

	return res, err
}

// CheckCMExistence returns true if cm exists in cm list
func (tx Tx) CheckCMExistence(cm []byte, stateDB *statedb.StateDB, shardID byte, tokenID *common.Hash) (bool, error) {
	ok, err := statedb.HasCommitment(stateDB, *tokenID, cm, shardID)
	return ok, err
}

func (tx Tx) ValidateDoubleSpendWithBlockChain(
	stateDB *statedb.StateDB,
) (bool, error) {
	tokenID := &common.Hash{}
	err := tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return false, err
	}
	shardID := byte(tx.GetValidationEnv().ShardID())

	if tx.Proof != nil {
		for _, txInput := range tx.Proof.GetInputCoins() {
			serialNumber := txInput.GetKeyImage().ToBytesS()
			ok, err := statedb.HasSerialNumber(stateDB, *tokenID, serialNumber, shardID)
			if ok || err != nil {
				return false, fmt.Errorf("double spend")
			}
		}
		for i, txOutput := range tx.Proof.GetOutputCoins() {
			if ok, err := checkSNDerivatorExistence(tokenID, txOutput.GetSNDerivator(), stateDB); ok || err != nil {
				if err != nil {
					utils.Logger.Log.Error(err)
				}
				utils.Logger.Log.Errorf("snd existed: %d\n", i)
				return false, utils.NewTransactionErr(utils.SndExistedError, err, fmt.Sprintf("snd existed: %d\n", i))
			}
		}

	}
	return true, nil
}

func (tx Tx) ValidateSanityDataByItSelf() (bool, error) {
	switch tx.Type {
	case common.TxNormalType, common.TxRewardType, common.TxCustomTokenPrivacyType, common.TxReturnStakingType: //is valid
	default:
		return false, utils.NewTransactionErr(utils.RejectTxType, fmt.Errorf("wrong tx type with %s", tx.Type))
	}

	// check info field
	if len(tx.Info) > 512 {
		return false, utils.NewTransactionErr(utils.RejectTxInfoSize, fmt.Errorf("wrong tx info length %d bytes, only support info with max length <= %d bytes", len(tx.Info), 512))
	}

	if tx.Metadata != nil {
		metaType := tx.Metadata.GetType()
		txType := tx.GetValidationEnv().TxType()
		if !metadata.IsAvailableMetaInTxType(metaType, txType) {
			return false, errors.Errorf("Not mismatch Type, txType: %v, metadataType %v", txType, metaType)
		}
	}

	// check tx size
	actualTxSize := tx.GetTxActualSize()
	if actualTxSize > common.MaxTxSize {
		//fmt.Print(actualTxSize, common.MaxTxSize)
		return false, utils.NewTransactionErr(utils.RejectTxSize, fmt.Errorf("tx size %d kB is too large", actualTxSize))
	}

	//check version
	if tx.Version > utils.TxVersion2Number {
		return false, utils.NewTransactionErr(utils.RejectTxVersion, fmt.Errorf("tx version is %d. Wrong version tx. Only support for version <= %d", tx.Version, utils.CurrentTxVersion))
	}
	// check LockTime before now
	if int64(tx.LockTime) > time.Now().Unix() {
		return false, utils.NewTransactionErr(utils.RejectInvalidLockTime, fmt.Errorf("wrong tx locktime %d", tx.LockTime))
	}

	if len(tx.SigPubKey) != common.SigPubKeySize {
		return false, utils.NewTransactionErr(utils.RejectTxPublickeySigSize, fmt.Errorf("wrong tx Sig PK size %d", len(tx.SigPubKey)))
	}

	metaData := tx.GetMetadata()
	proof := tx.GetProof()
	if metaData != nil {
		if !metaData.ValidateMetadataByItself() {
			return false, errors.Errorf("Metadata is not valid")
		}
	}

	if (proof == nil) || ((len(proof.GetInputCoins()) == 0) && (len(proof.GetOutputCoins()) == 0)) {
		if metaData == nil {
			utils.Logger.Log.Errorf("[invalidtxsanity] This tx %v has no proof, but metadata is nil", tx.Hash().String())
		} else {
			metaType := metaData.GetType()
			if !metadata.NoInputNoOutput(metaType) {
				utils.Logger.Log.Errorf("[invalidtxsanity] This tx %v has no proof, but metadata is invalid, metadata type %v", tx.Hash().String(), metaType)
			}
		}
	} else {
		if len(proof.GetInputCoins()) == 0 {
			if (metaData == nil) && (tx.GetValidationEnv().TxAction() != common.TxActInit) {
				return false, utils.NewTransactionErr(utils.RejectTxType, fmt.Errorf("This tx %v has no input, but metadata is nil", tx.Hash().String()))
			} else {
				if metaData != nil {
					metaType := metaData.GetType()
					if !metadata.NoInputHasOutput(metaType) {
						return false, utils.NewTransactionErr(utils.RejectTxType, fmt.Errorf("This tx %v has no proof, but metadata is invalid, metadata type %v", tx.Hash().String(), metaType))
					}
				}
			}
		}
		// if len(proof.GetOutputCoins()) == 0 {
		// 	if metaData != nil {
		// 		metaType := metaData.GetType()
		// 		if !metadata.HasInputNoOutput(metaType) {
		// 			return false, utils.NewTransactionErr(RejectTxType, fmt.Errorf("This tx %v has no proof, but metadata is invalid, metadata type %v", tx.Hash().String(), metaType))
		// 		}
		// 	}
		// }
		// check sanity of Proof
		validateSanityOfProof, err := tx.validateSanityDataOfProofV2()
		if err != nil || !validateSanityOfProof {
			return false, err
		}
	}

	return true, nil
}

func (tx *Tx) ValidateSanityDataWithBlockchain(
	chainRetriever metadata.ChainRetriever,
	shardViewRetriever metadata.ShardViewRetriever,
	beaconViewRetriever metadata.BeaconViewRetriever,
	beaconHeight uint64,
) (
	bool,
	error,
) {
	// Validate SND???
	// Validate DoubleSpend???
	if tx.Metadata != nil {
		utils.Logger.Log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := tx.Metadata.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, tx)
		utils.Logger.Log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	return true, nil
}

// func (tx *Tx) ValidateSanityDataWithBlockchain(

func (tx *Tx) LoadCommitment(
	db *statedb.StateDB,
) error {
	prf := tx.Proof
	if prf == nil {
		meta := tx.GetMetadata()
		if meta == nil {
			return errors.Errorf("This tx has no proof and not a tx for pay fee or tx with metadata")
		}
		if meta != nil {
			if metadata.NoInputNoOutput(meta.GetType()) || metadata.NoInputHasOutput(meta.GetType()) {
				return nil
			} else {
				return errors.Errorf("Invalid tx")
			}
		}
	}
	tokenID := tx.GetTokenID()
	utils.Logger.Log.Infof("[debugtxs] %v %v\n", tx, tx.GetValidationEnv())
	if tx.GetValidationEnv().IsPrivacy() {
		proofV1, ok := prf.(*privacy.ProofV1)
		if !ok {
			return fmt.Errorf("cannot cast payment proofV1")
		}
		return proofV1.LoadCommitmentFromStateDB(db, tokenID, byte(tx.GetValidationEnv().ShardID()))
	} else {
		for _, iCoin := range prf.GetInputCoins() {
			ok, err := tx.CheckCMExistence(
				iCoin.GetCommitment().ToBytesS(),
				db,
				byte(tx.GetValidationEnv().ShardID()),
				tokenID,
			)
			if !ok || err != nil {
				if err != nil {
					utils.Logger.Log.Error(err)
				}
				return utils.NewTransactionErr(utils.InputCommitmentIsNotExistedError, err)
			}
		}
	}
	return nil
}

func (tx *Tx) ValidateTxCorrectness(
// transactionStateDB *statedb.StateDB,
) (
	bool,
	error,
) {
	if ok, err := tx.VerifySigTx(); (!ok) || (err != nil) {
		utils.Logger.Log.Errorf("Validate tx %v return %v error %v", tx.Hash().String(), ok, err)
		return ok, err
	}

	utils.Logger.Log.Debugf("VALIDATING TX........\n")

	var valid bool
	var err error

	if tx.Proof != nil {
		proofV1, ok := tx.Proof.(*privacy.ProofV1)
		if !ok {
			return false, utils.NewTransactionErr(utils.TxProofVerifyFailError, fmt.Errorf("cannot cast proofV1 for tx %v", tx.Hash().String()))
		}
		valid, err = proofV1.VerifyV2(tx.GetValidationEnv(), tx.SigPubKey, tx.Fee)
		if !valid {
			if err != nil {
				utils.Logger.Log.Error(err)
			}
			return false, utils.NewTransactionErr(utils.TxProofVerifyFailError, err, tx.Hash().String())
		} else {
			utils.Logger.Log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
		}
	}

	return true, nil
}

func (tx Tx) checkDuplicateInput() (bool, error) {
	if len(tx.Proof.GetInputCoins()) > 255 {
		return false, fmt.Errorf("Input coins in tx are very large:" + strconv.Itoa(len(tx.Proof.GetInputCoins())))
	}

	// check doubling a input coin in tx
	serialNumbers := make(map[[privacy.Ed25519KeySize]byte]struct{})
	for i, inCoin := range tx.Proof.GetInputCoins() {
		snBytes := inCoin.GetKeyImage().ToBytes()
		if _, ok := serialNumbers[snBytes]; ok {
			utils.Logger.Log.Errorf("Double input in tx - txId %v - index %v", tx.Hash().String(), i)
			return false, fmt.Errorf("double input in tx")
		}
		serialNumbers[snBytes] = struct{}{}
	}
	return true, nil
}

func (tx Tx) checkDuplicateOutput() (bool, error) {
	if len(tx.Proof.GetOutputCoins()) > 255 {
		return false, fmt.Errorf("Output coins in tx are very large:" + strconv.Itoa(len(tx.Proof.GetOutputCoins())))
	}

	sndOutputs := make(map[[privacy.Ed25519KeySize]byte]struct{})
	for i, output := range tx.Proof.GetOutputCoins() {
		sndBytes := output.GetSNDerivator().ToBytes()
		if _, ok := sndOutputs[sndBytes]; ok {
			utils.Logger.Log.Errorf("Double output in tx - txId %v - index %v", tx.Hash().String(), i)
			return false, utils.NewTransactionErr(utils.DuplicatedOutputSndError, fmt.Errorf("Duplicate output coins' snd\n"))
		}
		sndOutputs[sndBytes] = struct{}{}
	}
	return true, nil
}

func (tx Tx) validateInputPrivacy() (bool, error) {
	tmpProof := tx.Proof
	prf, ok := tmpProof.(*privacy.ProofV1)
	if !ok {
		return false, fmt.Errorf("cannot cast payment proof v1")
	}

	cmInputSK := prf.GetCommitmentInputSecretKey()
	if !cmInputSK.PointValid() {
		return false, fmt.Errorf("validate sanity ComInputSK of proof failed")
	}
	sigPubKeyPoint, err := new(privacy.Point).FromBytesS(tx.GetSigPubKey())
	if err != nil {
		utils.Logger.Log.Errorf("SigPubKey is invalid - txId %v", tx.Hash().String())
		return false, fmt.Errorf("SigPubKey is invalid")
	}
	if !privacy.IsPointEqual(cmInputSK, sigPubKeyPoint) {
		utils.Logger.Log.Errorf("SigPubKey is not equal to commitment of private key - txId %v", tx.Hash().String())
		return false, fmt.Errorf("SigPubKey is not equal to commitment of private key")
	}
	if !prf.GetCommitmentInputShardID().PointValid() {
		return false, fmt.Errorf("validate sanity ComInputShardID of proof failed")
	}
	cmInputSNDs := prf.GetCommitmentInputSND()
	cmInputValue := prf.GetCommitmentInputValue()
	if (len(cmInputSNDs) != len(cmInputValue)) || (len(cmInputSNDs) != len(prf.GetInputCoins())) {
		return false, errors.Errorf("Len Commitment input SND %v and Commitment input value %v and len input coins %v is not equal", len(cmInputSNDs), len(cmInputValue), len(prf.GetInputCoins()))
	}
	for i, iCoin := range prf.GetInputCoins() {
		if !iCoin.GetKeyImage().PointValid() {
			return false, fmt.Errorf("validate sanity Serial number of input coin failed")
		}
		if iCoin.GetCommitment() != nil {
			return false, fmt.Errorf("CoinCommitment of input coin is not nil")
		}
		if iCoin.GetPublicKey() != nil {
			return false, fmt.Errorf("PublicKey of input coin is not nil")
		}
		if iCoin.GetRandomness() != nil {
			return false, fmt.Errorf("Randomness of input coin is not nil")
		}
		if iCoin.GetSNDerivator() != nil {
			return false, fmt.Errorf("SNDerivator of input coin is not nil")
		}

		if !cmInputValue[i].PointValid() {
			return false, fmt.Errorf("validate sanity ComInputValue of proof failed")
		}
		if !cmInputSNDs[i].PointValid() {
			return false, fmt.Errorf("validate sanity ComInputValue of proof failed")
		}
	}

	return true, nil
}

func (tx Tx) validateInputNoPrivacy() (bool, error) {
	tmpProof := tx.Proof
	prf, ok := tmpProof.(*privacy.ProofV1)
	if !ok {
		return false, fmt.Errorf("cannot cast payment proof v1")
	}

	inputCoins := prf.GetInputCoins()
	sigPubKeyPoint, err := new(privacy.Point).FromBytesS(tx.GetSigPubKey())
	if err != nil {
		utils.Logger.Log.Errorf("SigPubKey is invalid - txId %v", tx.Hash().String())
		return false, fmt.Errorf("SigPubKey is invalid")
	}
	for _, iCoin := range inputCoins {
		if !iCoin.GetKeyImage().PointValid() {
			return false, fmt.Errorf("validate sanity Serial number of input coin failed")
		}
		if !iCoin.GetCommitment().PointValid() {
			return false, fmt.Errorf("validate sanity CoinCommitment of input coin failed")
		}
		if !iCoin.GetPublicKey().PointValid() {
			return false, fmt.Errorf("validate sanity PublicKey of input coin failed")
		}
		if !iCoin.GetRandomness().ScalarValid() {
			return false, fmt.Errorf("validate sanity Randomness of input coin failed")
		}
		if !iCoin.GetSNDerivator().ScalarValid() {
			return false, fmt.Errorf("validate sanity SNDerivator of input coin failed")
		}
		if !privacy.IsPointEqual(iCoin.GetPublicKey(), sigPubKeyPoint) {
			utils.Logger.Log.Errorf("SigPubKey is not equal to public key of input coins - txId %v", tx.Hash().String())
			return false, fmt.Errorf("SigPubKey is not equal to public key of input coins")
		}
	}

	return true, nil
}

func (tx Tx) validateOutputPrivacy() (bool, error) {
	tmpProof := tx.Proof
	prf, ok := tmpProof.(*privacy.ProofV1)
	if !ok {
		return false, fmt.Errorf("cannot cast payment proof v1")
	}

	cmOutputValue := prf.GetCommitmentOutputValue()
	cmOutSNDs := prf.GetCommitmentOutputSND()
	cmOutSIDs := prf.GetCommitmentOutputShardID()

	for i, oCoin := range prf.GetOutputCoins() {
		if !oCoin.GetPublicKey().PointValid() {
			return false, fmt.Errorf("validate sanity Public key of output coin failed")
		}
		if !oCoin.GetCommitment().PointValid() {
			return false, fmt.Errorf("validate sanity Coin commitment of output coin failed")
		}
		if !oCoin.GetSNDerivator().ScalarValid() {
			return false, fmt.Errorf("validate sanity SNDerivator of output coin failed")
		}

		if oCoin.GetRandomness() != nil {
			return false, fmt.Errorf("Randomness of output coin is not nil")
		}

		if !cmOutSIDs[i].PointValid() {
			return false, fmt.Errorf("validate sanity ComOutputShardID of proof failed")
		}
		if !cmOutSNDs[i].PointValid() {
			return false, fmt.Errorf("validate sanity ComOutputSND of proof failed")
		}
		if !cmOutputValue[i].PointValid() {
			return false, fmt.Errorf("validate sanity ComOutputValue of proof failed")
		}
	}

	return true, nil
}

func (tx Tx) validateOutputNoPrivacy() (bool, error) {
	tmpProof := tx.Proof
	prf, ok := tmpProof.(*privacy.ProofV1)
	if !ok {
		return false, fmt.Errorf("cannot cast payment proof v1")
	}

	for _, oCoin := range prf.GetOutputCoins() {
		if !oCoin.GetCommitment().PointValid() {
			return false, fmt.Errorf("validate sanity CoinCommitment of output coin failed")
		}
		if !oCoin.GetPublicKey().PointValid() {
			return false, fmt.Errorf("validate sanity PublicKey of output coin failed")
		}
		if !oCoin.GetRandomness().ScalarValid() {
			return false, fmt.Errorf("validate sanity Randomness of output coin failed")
		}
		if !oCoin.GetSNDerivator().ScalarValid() {
			return false, fmt.Errorf("validate sanity SNDerivator of output coin failed")
		}
	}
	return true, nil
}

func (tx Tx) validatePrivacyZKPSanityWithInput() (bool, error) {
	tmpProof := tx.Proof
	prf, ok := tmpProof.(*privacy.ProofV1)
	if !ok {
		return false, fmt.Errorf("cannot cast payment proof v1")
	}

	inputCoins := prf.GetInputCoins()
	oomProofs := prf.GetOneOfManyProof()
	snProofs := prf.GetSerialNumberProof()
	snNoPrivacyProofs := prf.GetSerialNumberNoPrivacyProof()
	if len(snNoPrivacyProofs) != 0 {
		return false, errors.Errorf("This is privacy tx, no NoPrivacy ZKP")
	}
	cmInputSNDs := prf.GetCommitmentInputSND()
	cmInputSK := prf.GetCommitmentInputSecretKey()
	if (len(inputCoins) != len(snProofs)) || (len(inputCoins) != len(oomProofs)) {
		return false, fmt.Errorf("the number of input coins must be equal to the number of serialnumber proofs and the number of one-of-many proofs")
	}
	for _, oomPrf := range oomProofs {
		if !oomPrf.ValidateSanity() {
			return false, fmt.Errorf("validate sanity One out of many proof failed")
		}
	}
	for i, snProof := range snProofs {
		if !snProof.ValidateSanity() {
			return false, fmt.Errorf("validate sanity Serial number proof failed")
		}
		if !privacy.IsPointEqual(snProof.GetComSK(), cmInputSK) {
			utils.Logger.Log.Errorf("ComSK in SNproof is not equal to commitment of private key - txId %v", tx.Hash().String())
			return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberPrivacyProofFailedErr, fmt.Errorf("comSK of SNProof %v is not comSK of input coins", i))
		}
		if !privacy.IsPointEqual(snProof.GetComInput(), cmInputSNDs[i]) {
			utils.Logger.Log.Errorf("cmSND in SNproof is not equal to commitment of input's SND - txId %v", tx.Hash().String())
			return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberPrivacyProofFailedErr, fmt.Errorf("cmSND in SNproof %v is not equal to commitment of input's SND", i))
		}
		if !privacy.IsPointEqual(inputCoins[i].GetKeyImage(), snProof.GetSN()) {
			utils.Logger.Log.Errorf("SN in SNProof is not equal to SN of input coin - txId %v", tx.Hash().String())
			return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberPrivacyProofFailedErr, fmt.Errorf("SN in SNProof %v is not equal to SN of input coin", i))
		}
	}
	return true, nil
}

func (tx Tx) validateNoPrivacyZKPSanityWithInput() (bool, error) {
	tmpProof := tx.Proof
	prf, ok := tmpProof.(*privacy.ProofV1)
	if !ok {
		return false, fmt.Errorf("cannot cast payment proof v1")
	}

	inputCoins := prf.GetInputCoins()
	snProofs := prf.GetSerialNumberNoPrivacyProof()
	oomProofs := prf.GetOneOfManyProof()
	snPrivacyProofs := prf.GetSerialNumberProof()
	if (len(snPrivacyProofs) != 0) || (len(oomProofs) != 0) {
		return false, errors.Errorf("This is tx no privacy, no privacy zkp")
	}
	for i, snPrf := range snProofs {
		if !snPrf.ValidateSanity() {
			return false, fmt.Errorf("validate sanity Serial number no privacy proof failed")
		}
		if !privacy.IsPointEqual(inputCoins[i].GetPublicKey(), snPrf.GetVKey()) {
			utils.Logger.Log.Errorf("VKey in SNNoPrivacyProof is not equal public key of sender - txId %v", tx.Hash().String())
			return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, fmt.Errorf("VKey of SNNoPrivacyProof %v is not public key of sender", i))
		}
		if !privacy.IsScalarEqual(inputCoins[i].GetSNDerivator(), snPrf.GetInput()) {
			utils.Logger.Log.Errorf("SND in SNNoPrivacyProof is not equal to input's SND - txId %v", tx.Hash().String())
			return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, fmt.Errorf("SND in SNNoPrivacyProof %v is not equal to input's SND", i))
		}
		if !privacy.IsPointEqual(inputCoins[i].GetKeyImage(), snPrf.GetOutput()) {
			utils.Logger.Log.Errorf("SN in SNNoPrivacyProof is not equal to SN in input coin - txId %v", tx.Hash().String())
			return false, errhandler.NewPrivacyErr(errhandler.VerifySerialNumberNoPrivacyProofFailedErr, fmt.Errorf("SN in SNNoPrivacyProof %v is not equal to SN in input coin", i))
		}
	}
	return true, nil
}

func (tx Tx) validatePrivacyZKPSanityWithOutput() (bool, error) {
	tmpProof := tx.Proof
	prf, ok := tmpProof.(*privacy.ProofV1)
	if !ok {
		return false, fmt.Errorf("cannot cast payment proof v1")
	}

	cmValueOfOutputCoins := prf.GetCommitmentOutputValue()
	rangeProof := prf.GetAggregatedRangeProof()
	if rangeProof == nil {
		return false, errors.Errorf("Invalid range proof, it can not be nil")
	}
	cmValueInBulletProof := rangeProof.GetCommitments()
	if len(cmValueOfOutputCoins) != len(cmValueInBulletProof) {
		return false, fmt.Errorf("invalid cmValues in Bullet proof")
	}

	for i, cmValue := range cmValueOfOutputCoins {
		if !privacy.IsPointEqual(cmValue, cmValueInBulletProof[i]) {
			utils.Logger.Log.Errorf("cmValue in Bulletproof is not equal to commitment of output's Value - txId %v", tx.Hash().String())
			return false, fmt.Errorf("cmValue %v in Bulletproof is not equal to commitment of output's Value", i)
		}
	}

	if !rangeProof.ValidateSanity() {
		return false, fmt.Errorf("validate sanity Aggregated range proof failed")
	}

	return true, nil
}

func (tx Tx) validateNoPrivacyZKPSanityWithOutput() (bool, error) {
	rangeProof := tx.Proof.GetAggregatedRangeProof()
	if rangeProof != nil {
		return false, errors.Errorf("This field must be nil")
	}
	return true, nil
}

func (tx Tx) validateSanityDataPrivacyProof() (bool, error) {
	if len(tx.Proof.GetInputCoins()) > 0 {
		ok, err := tx.checkDuplicateInput()
		if !ok || err != nil {
			return ok, err
		}
		ok, err = tx.validateInputPrivacy()
		if !ok || err != nil {
			return ok, err
		}
		ok, err = tx.validatePrivacyZKPSanityWithInput()
		if !ok || err != nil {
			return ok, err
		}
	}
	if len(tx.Proof.GetOutputCoins()) > 0 {
		ok, err := tx.checkDuplicateOutput()
		if !ok || err != nil {
			return ok, err
		}
		ok, err = tx.validateOutputPrivacy()
		if !ok || err != nil {
			return ok, err
		}
		ok, err = tx.validatePrivacyZKPSanityWithOutput()
		if !ok || err != nil {
			return ok, err
		}
	}
	return true, nil
}

func (tx Tx) validateSanityDataNoPrivacyProof() (bool, error) {
	if len(tx.Proof.GetInputCoins()) > 0 {
		ok, err := tx.checkDuplicateInput()
		if !ok || err != nil {
			return ok, err
		}
		ok, err = tx.validateInputNoPrivacy()
		if !ok || err != nil {
			return ok, err
		}
		ok, err = tx.validateNoPrivacyZKPSanityWithInput()
		if !ok || err != nil {
			return ok, err
		}
	}
	if len(tx.Proof.GetOutputCoins()) > 0 {
		ok, err := tx.checkDuplicateOutput()
		if !ok || err != nil {
			return ok, err
		}
		ok, err = tx.validateOutputNoPrivacy()
		if !ok || err != nil {
			return ok, err
		}
		ok, err = tx.validateNoPrivacyZKPSanityWithOutput()
		if !ok || err != nil {
			return ok, err
		}
	}
	return true, nil
}

// Todo decoupling this function
func (tx Tx) validateSanityDataOfProofV2() (bool, error) {
	if tx.Proof != nil {
		isPrivacy := tx.IsPrivacy()
		if isPrivacy {
			return tx.validateSanityDataPrivacyProof()
		}
		return tx.validateSanityDataNoPrivacyProof()
	}
	return false, nil
}