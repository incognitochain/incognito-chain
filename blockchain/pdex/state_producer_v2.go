package pdex

import (
	"errors"

	"github.com/incognitochain/incognito-chain/metadata"
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
		metaData, ok := tx.GetMetadata().(*metadata.PDEV3AddLiquidity)
		if !ok {
			return res, errors.New("Can not parse add liquidity metadata")
		}
		Logger.log.Info(metaData)
		Logger.log.Info(shardID)
		Logger.log.Info(txReqID)
	}

	return [][]string{}, nil
}

func (sp *stateProducerV2) modifyParams(
	actions [][]string,
	beaconHeight uint64,
	params Params,
) ([][]string, error) {
	return [][]string{}, nil
}
