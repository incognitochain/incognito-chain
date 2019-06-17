package rpcserver

import (
	"errors"
	"sort"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/mempool"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

/*
handleGetShardToBeaconPoolState - RPC get shard to beacon pool state
*/
func (httpServer *HttpServer) handleGetShardToBeaconPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetShardToBeaconPoolState params: %+v", params)
	shardToBeaconPool := mempool.GetShardToBeaconPool()
	if shardToBeaconPool == nil {
		Logger.log.Infof("handleGetShardToBeaconPoolState result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardToBeaconPool.GetAllBlockHeight()
	Logger.log.Infof("handleGetShardToBeaconPoolState result: %+v", result)
	return result, nil
}

/*
handleGetCrossShardPoolState - RPC get cross shard pool state
*/
func (httpServer *HttpServer) handleGetCrossShardPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetCrossShardPoolState params: %+v", params)
	// get component
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		Logger.log.Infof("handleGetCrossShardPoolState result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Infof("handleGetCrossShardPoolState result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	shardID := byte(shardIDTemp)

	result := mempool.GetCrossShardPool(shardID).GetAllBlockHeight()
	Logger.log.Infof("handleGetCrossShardPoolState result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetNextCrossShard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetNextCrossShard params: %+v", params)
	// get component
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 3 {
		Logger.log.Infof("handleGetNextCrossShard result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	fromShardTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Infof("handleGetNextCrossShard result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	fromShard := byte(fromShardTemp)
	toShardTemp, ok := paramsArray[1].(float64)
	if !ok {
		Logger.log.Infof("handleGetNextCrossShard result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	toShard := byte(toShardTemp)
	startHeightTemp, ok := paramsArray[2].(float64)
	if !ok {
		Logger.log.Infof("handleGetNextCrossShard result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	startHeight := uint64(startHeightTemp)

	result := mempool.GetCrossShardPool(toShard).GetNextCrossShardHeight(fromShard, toShard, startHeight)
	Logger.log.Infof("handleGetNextCrossShard result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetBeaconPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBeaconPoolState params: %+v", params)
	beaconPool := mempool.GetBeaconPool()
	if beaconPool == nil {
		Logger.log.Infof("handleGetBeaconPoolState result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Beacon Pool not init"))
	}
	result := beaconPool.GetAllBlockHeight()
	Logger.log.Infof("handleGetBeaconPoolState result: %+v", result)
	return result, nil
}

type Blocks struct {
	Pending []uint64
	Valid   []uint64
	Latest  uint64
}

func (httpServer *HttpServer) handleGetShardPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetShardPoolState params: %+v", params)
	// get params
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key params"))
	}
	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Infof("handleGetShardPoolState result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	shardID := byte(shardIDTemp)

	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		Logger.log.Infof("handleGetShardPoolState result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardPool.GetAllBlockHeight()
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	temp := Blocks{Valid: shardPool.GetValidBlockHeight(), Pending: shardPool.GetPendingBlockHeight(), Latest: shardPool.GetShardState()}
	Logger.log.Infof("handleGetShardPoolState result: %+v", temp)
	return temp, nil
}

func (httpServer *HttpServer) handleGetShardPoolLatestValidHeight(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetShardPoolLatestValidHeight params: %+v", params)
	// get params
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		Logger.log.Infof("handleGetShardPoolLatestValidHeight result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key params"))
	}
	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Infof("handleGetShardPoolLatestValidHeight result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key params"))
	}
	shardID := byte(shardIDTemp)

	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		Logger.log.Infof("handleGetShardPoolLatestValidHeight result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardPool.GetLatestValidBlockHeight()
	Logger.log.Infof("handleGetShardPoolLatestValidHeight result: %+v", result)
	return result, nil
}

//==============Version 2================
/*
handleGetShardToBeaconPoolState - RPC get shard to beacon pool state
*/
func (httpServer *HttpServer) handleGetShardToBeaconPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetShardToBeaconPoolStateV2 params: %+v", params)
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) != 0 {
		Logger.log.Infof("handleGetShardToBeaconPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("wrong format params"))
	}
	shardToBeaconPool := mempool.GetShardToBeaconPool()
	if shardToBeaconPool == nil {
		Logger.log.Infof("handleGetShardToBeaconPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard to Beacon Pool not init"))
	}
	allBlockHeight := shardToBeaconPool.GetAllBlockHeight()
	allLatestBlockHeight := shardToBeaconPool.GetLatestValidPendingBlockHeight()
	shardToBeaconPoolResult := jsonresult.ShardToBeaconPoolResult{}
	shardToBeaconPoolResult.ValidBlockHeight = make([]jsonresult.BlockHeights, len(allBlockHeight))
	shardToBeaconPoolResult.PendingBlockHeight = make([]jsonresult.BlockHeights, len(allBlockHeight))
	index := 0
	for shardID, blockHeights := range allBlockHeight {
		latestBlockHeight := allLatestBlockHeight[shardID]
		shardToBeaconPoolResult.PendingBlockHeight[index].ShardID = shardID
		for _, blockHeight := range blockHeights {
			if blockHeight <= latestBlockHeight {
				shardToBeaconPoolResult.ValidBlockHeight[index].BlockHeightList = append(shardToBeaconPoolResult.ValidBlockHeight[index].BlockHeightList, blockHeight)
			} else {
				shardToBeaconPoolResult.PendingBlockHeight[index].BlockHeightList = append(shardToBeaconPoolResult.PendingBlockHeight[index].BlockHeightList, blockHeight)
			}
		}
		index++
	}
	Logger.log.Infof("handleGetShardToBeaconPoolStateV2 result: %+v", shardToBeaconPoolResult)
	return shardToBeaconPoolResult, nil
}

/*
handleGetCrossShardPoolState - RPC get cross shard pool state
*/
func (httpServer *HttpServer) handleGetCrossShardPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetCrossShardPoolStateV2 params: %+v", params)
	var index = 0
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) != 1 {
		Logger.log.Infof("handleGetCrossShardPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("wrong format params"))
	}
	shardID := byte(paramsArray[0].(float64))

	crossShardPool := mempool.GetCrossShardPool(shardID)
	if crossShardPool == nil {
		Logger.log.Infof("handleGetCrossShardPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Cross Shard Pool not init"))
	}
	allValidBlockHeight := crossShardPool.GetValidBlockHeight()
	allPendingBlockHeight := crossShardPool.GetPendingBlockHeight()

	crossShardPoolResult := jsonresult.CrossShardPoolResult{}
	crossShardPoolResult.ValidBlockHeight = make([]jsonresult.BlockHeights, len(allValidBlockHeight))
	crossShardPoolResult.PendingBlockHeight = make([]jsonresult.BlockHeights, len(allPendingBlockHeight))
	index = 0
	for shardID, blockHeights := range allValidBlockHeight {
		crossShardPoolResult.ValidBlockHeight[index].ShardID = shardID
		crossShardPoolResult.ValidBlockHeight[index].BlockHeightList = blockHeights
	}
	index = 0
	for shardID, blockHeights := range allPendingBlockHeight {
		crossShardPoolResult.PendingBlockHeight[index].ShardID = shardID
		crossShardPoolResult.PendingBlockHeight[index].BlockHeightList = blockHeights
	}
	Logger.log.Infof("handleGetCrossShardPoolStateV2 result: %+v", crossShardPoolResult)
	return crossShardPoolResult, nil
}

/*
handleGetShardPoolState - RPC get shard block in pool
*/
func (httpServer *HttpServer) handleGetShardPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetShardPoolStateV2 params: %+v", params)
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		Logger.log.Infof("handleGetShardPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key params"))
	}
	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Infof("handleGetShardPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("shardID is invalid"))
	}
	shardID := byte(shardIDTemp)

	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		Logger.log.Infof("handleGetShardPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardPool.GetAllBlockHeight()
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	temp := Blocks{Valid: shardPool.GetValidBlockHeight(), Pending: shardPool.GetPendingBlockHeight(), Latest: shardPool.GetShardState()}
	Logger.log.Infof("handleGetShardPoolStateV2 result: %+v", temp)
	return temp, nil
}

func (httpServer *HttpServer) handleGetBeaconPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleGetBeaconPoolStateV2 params: %+v", params)
	beaconPool := mempool.GetBeaconPool()
	if beaconPool == nil {
		Logger.log.Infof("handleGetBeaconPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Beacon Pool not init"))
	}
	result := Blocks{Valid: beaconPool.GetValidBlockHeight(), Pending: beaconPool.GetPendingBlockHeight(), Latest: beaconPool.GetBeaconState()}
	Logger.log.Infof("handleGetBeaconPoolStateV2 result: %+v", result)
	return result, nil
}
