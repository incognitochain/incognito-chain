package tx_ver2

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (txToken *TxToken) LoadCommitment(*statedb.StateDB) error {
	return nil
}

func (txToken TxToken) ValidateDoubleSpendWithBlockChain(*statedb.StateDB) (bool, error) {
	return true, nil
}

func (txToken TxToken) ValidateSanityDataByItSelf() (bool, error) {
	return true, nil
}

func (txToken *TxToken) ValidateSanityDataWithBlockchain(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) (bool, error) {
	return true, nil
}

func (txToken *TxToken) ValidateTxCorrectness() (bool, error) {
	return true, nil
}

func (txToken *TxToken) VerifySigTx() (bool, error) {
	return true, nil
}

func (txToken *TxToken) initEnv() metadata.ValidationEnviroment {
	return nil
}

func (txToken *TxToken) GetValidationEnv() metadata.ValidationEnviroment {
	return nil
}

func (txToken *TxToken) SetValidationEnv(metadata.ValidationEnviroment) {
	return
}