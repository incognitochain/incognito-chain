package tx_ver2

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (tx *Tx) LoadCommitment(*statedb.StateDB) error {
	return nil
}

func (tx Tx) ValidateDoubleSpendWithBlockChain(*statedb.StateDB) (bool, error) {
	return true, nil
}

func (tx Tx) ValidateSanityDataByItSelf() (bool, error) {
	return true, nil
}

func (tx *Tx) ValidateSanityDataWithBlockchain(metadata.ChainRetriever, metadata.ShardViewRetriever, metadata.BeaconViewRetriever, uint64) (bool, error) {
	return true, nil
}

func (tx *Tx) ValidateTxCorrectness() (bool, error) {
	return true, nil
}

func (tx *Tx) VerifySigTx() (bool, error) {
	return true, nil
}