package transaction

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/pkg/errors"
)

func (tx TxCustomTokenPrivacy) ValidateDoubleSpendWithBlockChain(
	stateDB *statedb.StateDB,
) (bool, error) {
	tokenID := tx.GetTokenID()
	shardID := byte(tx.valEnv.ShardID())
	txNormal := tx.TxPrivacyTokenData.TxNormal
	if tokenID == nil {
		return false, errors.Errorf("TokenID of tx %v is not valid", tx.Hash().String())
	}
	if txNormal.Proof != nil {
		for _, txInput := range txNormal.Proof.GetInputCoins() {
			serialNumber := txInput.CoinDetails.GetSerialNumber().ToBytesS()
			ok, err := statedb.HasSerialNumber(stateDB, *tokenID, serialNumber, shardID)
			if ok || err != nil {
				return false, errors.New("double spend")
			}
		}
		for i, txOutput := range txNormal.Proof.GetOutputCoins() {
			if ok, err := CheckSNDerivatorExistence(tokenID, txOutput.CoinDetails.GetSNDerivator(), stateDB); ok || err != nil {
				if err != nil {
					Logger.log.Error(err)
				}
				Logger.log.Errorf("snd existed: %d\n", i)
				return false, NewTransactionErr(SndExistedError, err, fmt.Sprintf("snd existed: %d\n", i))
			}
		}
	}
	return tx.Tx.ValidateDoubleSpendWithBlockChain(stateDB)
}

func (tx TxCustomTokenPrivacy) ValidateSanityDataByItSelf() (bool, error) {
	if tx.GetType() != common.TxCustomTokenPrivacyType {
		return false, NewTransactionErr(InvalidSanityDataPrivacyTokenError, errors.New("txCustomTokenPrivacy.Tx should have type tp"))
	}
	if tx.TxPrivacyTokenData.TxNormal.GetType() != common.TxNormalType {
		return false, NewTransactionErr(InvalidSanityDataPrivacyTokenError, errors.New("txCustomTokenPrivacy.TxNormal should have type n"))
	}
	meta := tx.Tx.Metadata
	if meta != nil {
		if !metadata.IsAvailableMetaInTxType(meta.GetType(), tx.GetType()) {
			err := errors.Errorf("Not mismatch type, txtype %v, metadatatype %v", tx.GetType(), meta.GetType())
			Logger.log.Errorf("Validate tx %v return %v, error %v", tx.Hash().String(), err)
			return false, err
		}
	}

	if tx.TxPrivacyTokenData.TxNormal.GetMetadata() != nil {
		return false, errors.Errorf("This tx field is just used for send token, can not have metadata")
	}
	if tx.TxPrivacyTokenData.PropertyID.String() == common.PRVIDStr {
		return false, NewTransactionErr(InvalidSanityDataPrivacyTokenError, errors.New("TokenID must not be equal PRVID"))
	}

	if (tx.Tx.Proof != nil) && ((len(tx.Tx.Proof.GetInputCoins()) != 0) || (len(tx.Tx.Proof.GetOutputCoins()) != 0)) {
		ok, err := tx.Tx.ValidateSanityDataByItSelf()
		if !ok || err != nil {
			return ok, err
		}
	}

	ok, err := tx.TxPrivacyTokenData.TxNormal.ValidateSanityDataByItSelf()
	if !ok || err != nil {
		return ok, err
	}

	return true, nil
}

func (tx *TxCustomTokenPrivacy) ValidateSanityDataWithBlockchain(
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
		Logger.log.Debug("tx.Metadata.ValidateSanityData")
		isContinued, ok, err := tx.GetMetadata().ValidateSanityData(chainRetriever, shardViewRetriever, beaconViewRetriever, beaconHeight, tx)
		Logger.log.Debug("END tx.Metadata.ValidateSanityData")
		if err != nil || !ok || !isContinued {
			return ok, err
		}
	}
	return true, nil
}

// LoadCommitment do something
func (tx *TxCustomTokenPrivacy) LoadCommitment(
	db *statedb.StateDB,
) error {
	embededTx := tx.Tx
	normalTx := tx.TxPrivacyTokenData.TxNormal
	prf := embededTx.Proof
	if prf != nil {
		if embededTx.valEnv.IsPrivacy() {
			// tokenID := embededTx.GetTokenID()
			err := prf.LoadCommitmentFromStateDB(db, &common.PRVCoinID, byte(tx.valEnv.ShardID()))
			if err != nil {
				Logger.log.Error(err)
				return err
			}
		} else {
			for _, iCoin := range prf.GetInputCoins() {
				ok, err := tx.CheckCMExistence(
					iCoin.CoinDetails.GetCoinCommitment().ToBytesS(),
					db,
					byte(tx.valEnv.ShardID()),
					&common.PRVCoinID,
				)
				if !ok || err != nil {
					if err != nil {
						Logger.log.Error(err)
					}
					return NewTransactionErr(InputCommitmentIsNotExistedError, err)
				}
			}
		}
	}
	tokenID := tx.GetTokenID()
	if tx.TxPrivacyTokenData.Type == CustomTokenInit {
		if !tx.TxPrivacyTokenData.Mintable {
			// check exist token
			if statedb.PrivacyTokenIDExisted(db, *tokenID) {
				return errors.Errorf("Privacy Token ID is existed")
			}
		}
	}
	prf = normalTx.Proof
	if prf != nil {
		if normalTx.valEnv.IsPrivacy() {
			err := prf.LoadCommitmentFromStateDB(db, tokenID, byte(tx.valEnv.ShardID()))
			if err != nil {
				Logger.log.Error(err)
				return err
			}
		} else {
			for _, iCoin := range prf.GetInputCoins() {
				ok, err := tx.CheckCMExistence(
					iCoin.CoinDetails.GetCoinCommitment().ToBytesS(),
					db,
					byte(tx.valEnv.ShardID()),
					tokenID,
				)
				if !ok || err != nil {
					if err != nil {
						Logger.log.Error(err)
					}
					return NewTransactionErr(InputCommitmentIsNotExistedError, err)
				}
			}
		}
	} else {
		return errors.Errorf("Normal tx of Tx CustomeTokenPrivacy can not has no input no outputs")
	}
	return nil
}

func (tx *TxCustomTokenPrivacy) ValidateTxCorrectness(
// transactionStateDB *statedb.StateDB,
) (
	bool,
	error,
) {
	Logger.log.Debugf("Validate tx correctness %v, normal tx %v, embeded tx %v", tx.Hash().String(), tx.TxPrivacyTokenData.TxNormal.Hash().String(), tx.Tx.Hash().String())
	ok, err := tx.TxPrivacyTokenData.TxNormal.ValidateTxCorrectness()
	if (!ok) || (err != nil) {
		return ok, err
	}
	return tx.Tx.ValidateTxCorrectness()

}
