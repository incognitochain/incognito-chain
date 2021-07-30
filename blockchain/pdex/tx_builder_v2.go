package pdex

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

type Pdexv3TxBuilder struct {
}

func (txBuilder *Pdexv3TxBuilder) Build(
	metaType int,
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
	featureStateDB *statedb.StateDB,
) (metadata.Transaction, error) {

	var tx metadata.Transaction
	var err error

	return tx, err
}
