package tx_ver2

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction/utils"
)

func getDerivableInputFromSigPubKey(rawPubkey []byte, tokenID common.Hash, txHash *common.Hash, shardID byte, db *statedb.StateDB) (*privacy.Point, error) {
	ringPubkey := new(SigPubKey)
	if err := ringPubkey.SetBytes(rawPubkey); err != nil {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("invalid SigPubKey in tx %s, token %s", txHash.String(), tokenID.String()))
	}
	ringSize := len(ringPubkey.Indexes)
	if ringSize != 1 || len(ringPubkey.Indexes[0]) != 1 {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("cannot identify burn input in tx %s, token %s with ring size %d, input length %d", txHash.String(), tokenID.String(), ringSize, len(ringPubkey.Indexes[0])))
	}
	rawCoin, err := statedb.GetOTACoinByIndex(db, tokenID, ringPubkey.Indexes[0][0].Uint64(), shardID)
	if err != nil {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, err)
	}
	c := new(privacy.CoinV2)
	if err := c.SetBytes(rawCoin); err != nil {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, err)
	}
	p := c.GetPublicKey()
	if p == nil {
		return nil, utils.NewTransactionErr(utils.UnexpectedError, fmt.Errorf("coin from db must have public key"))
	}
	return p, nil
}
