package pdex

import (
	"errors"

	v3 "github.com/incognitochain/incognito-chain/blockchain/pdex/v3utils"
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
		instStr := waitingInstruction.StringArr()
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

func (sp *stateProducerV2) trade(
	txs []metadata.Transaction,
	beaconHeight uint64,
	pairs map[string]PoolPairState,
) ([][]string, map[string]PoolPairState, error) {
	result := [][]string{}
	var tradeRequests []metadataPdexV3.TradeRequest

	for _, tx := range txs {
		item, ok := tx.GetMetadata().(*metadataPdexV3.TradeRequest)
		if !ok {
			return result, pairs, errors.New("Can not parse add liquidity metadata")
		}
		tradeRequests = append(tradeRequests, *item)
	}

	// TODO: sort
	// tradeRequests := sortByFee(
	// 	tradeRequests,
	// 	beaconHeight,
	// 	pairs,
	// )

	for _, currentTrade := range tradeRequests {
		// line up the trade path
		reserves, tradePath, tradeDirections, err := getRelevantReserves(currentTrade.TokenToSell, currentTrade.TradePath, pairs)
		if err != nil {
			return [][]string{}, pairs, err
		}
		refunded, currentInst, changedReserves, err := v3.AcceptOrRefundTrade(currentTrade.SellAmount, reserves, tradeDirections, nil)
		if err != nil {
			return [][]string{}, pairs, err
		}
		_, _, _ = tradePath, refunded, changedReserves
		result = append(result, currentInst)
	}

	return result, pairs, nil
}

func (sp *stateProducerV2) addOrder(
	txs []metadata.Transaction,
	beaconHeight uint64,
	pairs map[string]PoolPairState,
) ([][]string, map[string]PoolPairState, error) {
	result := [][]string{}
	var orderRequests []metadataPdexV3.AddOrderRequest

	for _, tx := range txs {
		item, ok := tx.GetMetadata().(*metadataPdexV3.AddOrderRequest)
		if !ok {
			return result, pairs, errors.New("Can not parse add liquidity metadata")
		}
		orderRequests = append(orderRequests, *item)
	}

	// TODO: sort
	// orderRequests := sortByFee(
	// 	orderRequests,
	// 	beaconHeight,
	// 	pairs,
	// )

	for _, currentOrderReq := range orderRequests {
		_ = currentOrderReq
		// result = append(result, currentInst)
	}

	return result, pairs, nil
}
