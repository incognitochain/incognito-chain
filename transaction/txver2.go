package transaction

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database"
)

type TxVersion2 struct{}

func (*TxVersion2) Prove(tx *Tx, params *TxPrivacyInitParams) error {
	return nil
}

func (*TxVersion2) Verify(tx *Tx, hasPrivacy bool, db database.DatabaseInterface, shardID byte, tokenID *common.Hash, isBatch bool, isNewTransaction bool) (bool, error) {
	return true, nil
}
