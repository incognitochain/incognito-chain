package pdex

import (
	"fmt"
	"strconv"

	v2 "github.com/incognitochain/incognito-chain/blockchain/pdex/v2utils"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	metadataPdexv3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
	"github.com/incognitochain/incognito-chain/privacy"
)

type TxBuilderV2 struct {
}

func (txBuilder *TxBuilderV2) Build(
	metaType int,
	inst []string,
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
	transactionStateDB *statedb.StateDB,
	featureStateDB *statedb.StateDB,
) (metadataCommon.Transaction, error) {

	var tx metadataCommon.Transaction
	var err error

	switch metaType {
	case metadataCommon.Pdexv3TradeRequestMeta:
		switch inst[1] {
		case strconv.Itoa(metadataPdexv3.TradeAcceptedStatus):
			action := instruction.Action{Content: &metadataPdexv3.AcceptedTrade{}}
			err := action.FromStringSlice(inst)
			if err != nil {
				return nil, err
			}
			tx, err = v2.TradeAcceptTx(action, producerPrivateKey, shardID, transactionStateDB)
		case strconv.Itoa(metadataPdexv3.TradeRefundedStatus):
			action := instruction.Action{Content: &metadataPdexv3.RefundedTrade{}}
			err := action.FromStringSlice(inst)
			if err != nil {
				return nil, err
			}
			tx, err = v2.TradeRefundTx(action, producerPrivateKey, shardID, transactionStateDB)

		default:
			return nil, fmt.Errorf("Invalid status %s from instruction", inst[1])
		}
	}

	return tx, err
}
