package transaction

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/privacy/privacy_v2"
)

type TxVersion2 struct{}

func (*TxVersion2) Prove(tx *Tx, params *TxPrivacyInitParams) error {
	outputCoins, err := parseOutputCoins(params)
	if err != nil {
		return err
	}
	inputCoins := params.inputCoins
	var conversion privacy.Proof
	conversion, err = privacy_v2.Prove(inputCoins, *outputCoins, params.hasPrivacy)
	tx.Proof = &conversion

	if err != nil {
		return err
	}
	return nil
}

func (*TxVersion2) Verify(tx *Tx, hasPrivacy bool, transactionStateDB *statedb.StateDB, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	return true, nil
}
