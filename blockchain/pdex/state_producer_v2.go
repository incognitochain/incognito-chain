package pdex

import (
	"errors"

	instruction "github.com/incognitochain/incognito-chain/instruction/pdexv3"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataPdexV3 "github.com/incognitochain/incognito-chain/metadata/pdexv3"
)

type stateProducerV2 struct {
	stateProducerBase
}

func (sp *stateProducerV2) addLiquidity(
	txs []metadata.Transaction,
	beaconHeight uint64,
) ([][]string, error) {
	res := [][]string{}
	for _, tx := range txs {
		shardID := byte(tx.GetValidationEnv().ShardID())
		txReqID := tx.Hash().String()
		metaData, ok := tx.GetMetadata().(*metadataPdexV3.AddLiquidity)
		if !ok {
			return res, errors.New("Can not parse add liquidity metadata")
		}
		waitingInstruction := instruction.NewWaitingAddLiquidityFromMetadata(*metaData, txReqID, shardID)
		instStr, err := waitingInstruction.StringArr()
		if err != nil {
			return res, err
		}
		res = append(res, instStr)
	}

	return res, nil
}

func (sp *stateProducerV2) modifyParams(
	actions [][]string,
	beaconHeight uint64,
	params Params,
) ([][]string, error) {
	return [][]string{}, nil
}
