package tx_ver1

import (
	"fmt"
	"strconv"

	"github.com/incognitochain/incognito-chain/config"
	errhandler "github.com/incognitochain/incognito-chain/privacy/errorhandler"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"

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

func (tx *Tx) ValidateSanityDataByItSelf() (bool, error) {
	isMint, _, _, _ := tx.GetTxMintData()
	bHeight := tx.GetValidationEnv().BeaconHeight()
	afterUpgrade := bHeight >= config.Param().BCHeightBreakPointPrivacyV2
	if afterUpgrade && !isMint {
		return false, utils.NewTransactionErr(utils.RejectTxVersion, errors.New("old version is no longer supported"))
	}
	if len(tx.SigPubKey) != common.SigPubKeySize {
		return false, utils.NewTransactionErr(utils.RejectTxPublickeySigSize, fmt.Errorf("wrong tx Sig PK size %d", len(tx.SigPubKey)))
	}
	ok, err := tx.TxBase.ValidateSanityDataWithMetadata()
	if (!ok) || (err != nil) {
		return false, err
	}
	return tx.TxBase.ValidateSanityDataByItSelf()
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
	meta := tx.GetMetadata()
	if meta != nil {
		utils.Logger.Log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := meta.ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, tx)
		utils.Logger.Log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	return true, nil
}

func (tx *Tx) CheckData(
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
	txEnv := tx.GetValidationEnv()
	tokenID := tx.GetTokenID()
	data := txEnv.DBData()
	utils.Logger.Log.Infof("[debugtxs] %v %v", tx, txEnv)
	if txEnv.IsPrivacy() {
		proofV1, ok := prf.(*privacy.ProofV1)
		if !ok {
			return fmt.Errorf("cannot cast payment proofV1")
		}
		err := proofV1.CheckCommitmentWithStateDB(data, db, tokenID, byte(tx.GetValidationEnv().ShardID()))
		if err != nil {
			return err
		}
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

func (tx *Tx) LoadData(
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
	txEnv := tx.GetValidationEnv()
	tokenID := tx.GetTokenID()
	utils.Logger.Log.Infof("[debugtxs] %v %v", tx, txEnv)
	if txEnv.IsPrivacy() {
		proofV1, ok := prf.(*privacy.ProofV1)
		if !ok {
			return fmt.Errorf("cannot cast payment proofV1")
		}
		data, err := proofV1.LoadDataFromStateDB(db, tokenID, byte(tx.GetValidationEnv().ShardID()))
		if err != nil {
			return err
		}
		tx.SetValidationEnv(tx_generic.WithDBData(txEnv, data))
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
	transactionStateDB *statedb.StateDB,
) (
	bool,
	error,
) {
	if (tx.GetValidationEnv().TxType() != common.TxCustomTokenPrivacyType) || (tx.GetValidationEnv().TxAction() != common.TxActInit) {
		if ok, err := tx.VerifySigTx(); (!ok) || (err != nil) {
			utils.Logger.Log.Errorf("Validate tx %v return %v error %v", tx.Hash().String(), ok, err)
			return ok, err
		}
	}

	utils.Logger.Log.Debugf("VALIDATING TX........")

	var valid bool
	var err error

	if tx.Proof != nil {
		proofV1, ok := tx.Proof.(*privacy.ProofV1)
		if !ok {
			return false, utils.NewTransactionErr(utils.TxProofVerifyFailError, fmt.Errorf("cannot cast proofV1 for tx %v", tx.Hash().String()))
		}
		valid, err = proofV1.VerifyV2(tx.GetValidationEnv(), tx.Fee)
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
			return false, utils.NewTransactionErr(utils.DuplicatedOutputSndError, fmt.Errorf("Duplicate output coins' snd"))
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
