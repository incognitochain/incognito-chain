package tx_ver2

import (
	"bytes"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2/mlsag"
	"github.com/incognitochain/incognito-chain/transaction/tx_generic"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

func (tx *Tx) ValidateSanityDataByItSelf() (bool, error) {
	if tx.Proof == nil {
		return false, fmt.Errorf("tx Privacy Ver 2 must have proof")
	}
	ok, err := tx.TxBase.ValidateSanityDataWithMetadata()
	if (!ok) || (err != nil) {
		return false, err
	}
	if check, err := tx.TxBase.ValidateSanityDataByItSelf(); !check || err != nil {
		utils.Logger.Log.Errorf("Cannot check sanity of version, size, proof, type and info: err %v", err)
		return false, err
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

func (tx *Tx) ValidateTxCorrectness(transactionStateDB *statedb.StateDB) (bool, error) {
	txEnv := tx.GetValidationEnv()
	sID := byte(txEnv.ShardID())
	tokenID := txEnv.TokenID()
	switch tx.GetType() {
	case common.TxRewardType, common.TxReturnStakingType:
		valid, err := tx.ValidateTxSalary(transactionStateDB)
		return valid, err
	case common.TxConversionType:
		valid, err := validateConversionVer1ToVer2(tx, transactionStateDB, sID, &tokenID)
		return valid, err
	}
	validSig, err := tx.VerifySigTx(transactionStateDB)
	if !validSig {
		if err != nil {
			utils.Logger.Log.Errorf("Error verifying signature ver2 with tx hash %s: %+v", tx.Hash().String(), err)
			return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, err)
		}
		utils.Logger.Log.Errorf("FAILED VERIFICATION SIGNATURE ver2 with tx hash %s", tx.Hash().String())
		return false, utils.NewTransactionErr(utils.VerifyTxSigFailError, fmt.Errorf("FAILED VERIFICATION SIGNATURE ver2 with tx hash %s", tx.Hash().String()))
	}

	if valid, err := tx.Proof.VerifyV2(txEnv, tx.Fee); !valid {
		if err != nil {
			utils.Logger.Log.Error(err)
		}
		return false, utils.NewTransactionErr(utils.TxProofVerifyFailError, err, tx.Hash().String())
	}
	utils.Logger.Log.Debugf("SUCCESSED VERIFICATION PAYMENT PROOF ")
	return true, nil

}

func (tx *Tx) VerifySigTx(transactionStateDB *statedb.StateDB) (bool, error) {
	vEnv := tx.GetValidationEnv()
	tokenID := vEnv.TokenID()
	sID := byte(vEnv.ShardID())
	if vEnv.HasCA() {
		return tx.verifySigCA(transactionStateDB, sID, &tokenID, false)
	}
	return tx.verifySig(transactionStateDB, sID, &tokenID, false)
}

// Retrieve ring from database using sigpubkey and last column commitment (last column = sumOutputCoinCommitment + fee)
func getRingFromSigPubKeyAndLastColumnCommitmentV2(txEnv metadata.ValidationEnviroment, sumOutputsWithFee *privacy.Point, transactionStateDB *statedb.StateDB) (*mlsag.Ring, error) {
	txSigPubKey := new(SigPubKey)
	if err := txSigPubKey.SetBytes(txEnv.SigPubKey()); err != nil {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("error when parsing bytes of txSigPubKey %v", err))
	}
	indexes := txSigPubKey.Indexes
	OTAData := txEnv.DBData()
	n := len(indexes)
	if n == 0 {
		return nil, fmt.Errorf("cannot get ring from Indexes: Indexes is empty")
	}
	m := len(indexes[0])
	if m*n != len(OTAData) {
		return nil, fmt.Errorf("cached OTA data not match with indexes")
	}

	ring := make([][]*privacy.Point, n)
	for i := 0; i < n; i++ {
		sumCommitment := new(privacy.Point).Identity()
		sumCommitment.Sub(sumCommitment, sumOutputsWithFee)
		row := make([]*privacy.Point, m+1)
		for j := 0; j < m; j++ {
			randomCoinBytes := OTAData[i*m+j]
			randomCoin := new(privacy.CoinV2)
			if err := randomCoin.SetBytes(randomCoinBytes); err != nil {
				utils.Logger.Log.Errorf("Set coin Byte error %v ", err)
				return nil, err
			}
			row[j] = randomCoin.GetPublicKey()
			sumCommitment.Add(sumCommitment, randomCoin.GetCommitment())
		}
		row[m] = new(privacy.Point).Set(sumCommitment)
		ring[i] = row
	}
	return mlsag.NewRing(ring), nil
}

// Retrieve ring from database using sigpubkey and last column commitment (last column = sumOutputCoinCommitment + fee)
func (tx *Tx) LoadData(transactionStateDB *statedb.StateDB) error {
	txEnv := tx.GetValidationEnv()
	switch tx.GetType() {
	case common.TxRewardType, common.TxReturnStakingType:
		return nil
	case common.TxConversionType:
		return checkInputInDB(tx, transactionStateDB)
	}
	if txEnv.TxAction() == common.TxActInit {
		return nil
	}
	txSigPubKey := new(SigPubKey)
	//utils.Logger.Log.Infof("tx val env %v %v %v %v", tx.Hash().String(), txEnv.IsPrivacy(), tx.GetType(), txEnv.TxAction())
	if err := txSigPubKey.SetBytes(txEnv.SigPubKey()); err != nil {
		return utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("error when parsing bytes of txSigPubKey %v", err))
	}
	indexes := txSigPubKey.Indexes
	n := len(indexes)
	if n == 0 {
		return fmt.Errorf("cannot get ring from Indexes: Indexes is empty")
	}

	m := len(indexes[0])
	data := make([][]byte, m*n)
	tokenID := txEnv.TokenID()
	shardID := txEnv.ShardID()
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			index := indexes[i][j]
			randomCoinBytes, err := statedb.GetOTACoinByIndex(transactionStateDB, tokenID, index.Uint64(), byte(shardID))
			if err != nil {
				utils.Logger.Log.Errorf("Get random onetimeaddresscoin error %v ", err)
				return err
			}
			data[i*m+j] = randomCoinBytes
			randomCoin := new(privacy.CoinV2)
			if err := randomCoin.SetBytes(randomCoinBytes); err != nil {
				utils.Logger.Log.Errorf("Set coin Byte error %v ", err)
				return err
			}
		}
	}
	tx.SetValidationEnv(tx_generic.WithDBData(txEnv, data))
	return nil
}

// Retrieve ring from database using sigpubkey and last column commitment (last column = sumOutputCoinCommitment + fee)
func (tx *Tx) CheckData(transactionStateDB *statedb.StateDB) error {
	txEnv := tx.GetValidationEnv()
	switch tx.GetType() {
	case common.TxRewardType, common.TxReturnStakingType:
		return nil
	case common.TxConversionType:
		return checkInputInDB(tx, transactionStateDB)
	}
	if txEnv.TxAction() == common.TxActInit {
		return nil
	}
	txSigPubKey := new(SigPubKey)
	if err := txSigPubKey.SetBytes(txEnv.SigPubKey()); err != nil {
		errStr := fmt.Sprintf("Error when parsing bytes of txSigPubKey %v", err)
		return utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf(errStr))
	}
	indexes := txSigPubKey.Indexes
	n := len(indexes)
	if n == 0 {
		return fmt.Errorf("cannot get ring from Indexes: Indexes is empty")
	}
	m := len(indexes[0])
	data := txEnv.DBData()
	tokenID := txEnv.TokenID()
	shardID := txEnv.ShardID()
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			index := indexes[i][j]
			cachedRandomCoin := data[i*m+j]
			randomCoinBytes, err := statedb.GetOTACoinByIndex(transactionStateDB, tokenID, index.Uint64(), byte(shardID))
			if err != nil {
				utils.Logger.Log.Errorf("Get random onetimeaddresscoin error %v ", err)
				return err
			}
			if !bytes.Equal(cachedRandomCoin, randomCoinBytes) {
				return fmt.Errorf("cached OTA coin is invalid, tx %v", tx.Hash().String())
			}
		}
	}
	return nil
}
