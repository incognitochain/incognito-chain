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
/*func (httpServer *HttpServer) handleGetShardToBeaconPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetShardToBeaconPoolState params: %+v", params)
	shardToBeaconPool := mempool.GetShardToBeaconPool()
	if shardToBeaconPool == nil {
		Logger.log.Debugf("handleGetShardToBeaconPoolState result: %+v", nil)
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardToBeaconPool.GetAllBlockHeight()
	Logger.log.Debugf("handleGetShardToBeaconPoolState result: %+v", result)
	return result, nil
}*/

/*
handleGetCrossShardPoolState - RPC get cross shard pool state
*/
/*func (httpServer *HttpServer) handleGetCrossShardPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetCrossShardPoolState params: %+v", params)
	// get component
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		Logger.log.Debugf("handleGetCrossShardPoolState result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetCrossShardPoolState result: %+v", nil)
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	shardID := byte(shardIDTemp)

	result := mempool.GetCrossShardPool(shardID).GetAllBlockHeight()
	Logger.log.Debugf("handleGetCrossShardPoolState result: %+v", result)
	return result, nil
}*/

func (httpServer *HttpServer) handleGetNextCrossShard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetNextCrossShard params: %+v", params)
	// get component
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 3 {
		Logger.log.Debugf("handleGetNextCrossShard result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid list Key component"))
	}
	fromShardTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetNextCrossShard result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid list Key component"))
	}
	fromShard := byte(fromShardTemp)
	toShardTemp, ok := paramsArray[1].(float64)
	if !ok {
		Logger.log.Debugf("handleGetNextCrossShard result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid list Key component"))
	}
	toShard := byte(toShardTemp)
	startHeightTemp, ok := paramsArray[2].(float64)
	if !ok {
		Logger.log.Debugf("handleGetNextCrossShard result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid list Key component"))
	}
	startHeight := uint64(startHeightTemp)

	result := mempool.GetCrossShardPool(toShard).GetNextCrossShardHeight(fromShard, toShard, startHeight)
	Logger.log.Debugf("handleGetNextCrossShard result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetBeaconPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetBeaconPoolState params: %+v", params)
	beaconPool := mempool.GetBeaconPool()
	if beaconPool == nil {
		Logger.log.Debugf("handleGetBeaconPoolState result: %+v", nil)
		return nil, NewRPCError(UnexpectedError, errors.New("Beacon Pool not init"))
	}
	result := beaconPool.GetAllBlockHeight()
	Logger.log.Debugf("handleGetBeaconPoolState result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetShardPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetShardPoolState params: %+v", params)
	// get params
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid list Key params"))
	}
	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetShardPoolState result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid list Key component"))
	}
	shardID := byte(shardIDTemp)

	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		Logger.log.Debugf("handleGetShardPoolState result: %+v", nil)
		return nil, NewRPCError(UnexpectedError, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardPool.GetAllBlockHeight()
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	temp := jsonresult.NewBlocksFromShardPool(*shardPool)
	Logger.log.Debugf("handleGetShardPoolState result: %+v", temp)
	return temp, nil
}

func (httpServer *HttpServer) handleGetShardPoolLatestValidHeight(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetShardPoolLatestValidHeight params: %+v", params)
	// get params
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		Logger.log.Debugf("handleGetShardPoolLatestValidHeight result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid list Key params"))
	}
	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetShardPoolLatestValidHeight result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid list Key params"))
	}
	shardID := byte(shardIDTemp)

	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		Logger.log.Debugf("handleGetShardPoolLatestValidHeight result: %+v", nil)
		return nil, NewRPCError(UnexpectedError, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardPool.GetLatestValidBlockHeight()
	Logger.log.Debugf("handleGetShardPoolLatestValidHeight result: %+v", result)
	return result, nil
}

//==============Version 2================
/*
handleGetShardToBeaconPoolState - RPC get shard to beacon pool state
*/
func (httpServer *HttpServer) handleGetShardToBeaconPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetShardToBeaconPoolStateV2 params: %+v", params)
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) != 0 {
		Logger.log.Debugf("handleGetShardToBeaconPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("wrong format params"))
	}
	shardToBeaconPool := mempool.GetShardToBeaconPool()
	if shardToBeaconPool == nil {
		Logger.log.Debugf("handleGetShardToBeaconPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(UnexpectedError, errors.New("Shard to Beacon Pool not init"))
	}
	allBlockHeight := shardToBeaconPool.GetAllBlockHeight()
	allLatestBlockHeight := shardToBeaconPool.GetLatestValidPendingBlockHeight()
	shardToBeaconPoolResult := jsonresult.NewShardToBeaconPoolResult(allBlockHeight, allLatestBlockHeight)

	Logger.log.Debugf("handleGetShardToBeaconPoolStateV2 result: %+v", shardToBeaconPoolResult)
	return shardToBeaconPoolResult, nil
}

/*
handleGetCrossShardPoolState - RPC get cross shard pool state
*/
func (httpServer *HttpServer) handleGetCrossShardPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetCrossShardPoolStateV2 params: %+v", params)

	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) != 1 {
		Logger.log.Debugf("handleGetCrossShardPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("wrong format params"))
	}
	shardID := byte(paramsArray[0].(float64))

	crossShardPool := mempool.GetCrossShardPool(shardID)
	if crossShardPool == nil {
		Logger.log.Debugf("handleGetCrossShardPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(UnexpectedError, errors.New("Cross Shard Pool not init"))
	}
	allValidBlockHeight := crossShardPool.GetValidBlockHeight()
	allPendingBlockHeight := crossShardPool.GetPendingBlockHeight()
	crossShardPoolResult := jsonresult.NewCrossShardPoolResult(allValidBlockHeight, allPendingBlockHeight)

	Logger.log.Debugf("handleGetCrossShardPoolStateV2 result: %+v", crossShardPoolResult)
	return crossShardPoolResult, nil
}

/*
handleGetShardPoolState - RPC get shard block in pool
*/
func (httpServer *HttpServer) handleGetShardPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetShardPoolStateV2 params: %+v", params)
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		Logger.log.Debugf("handleGetShardPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("invalid list Key params"))
	}
	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Debugf("handleGetShardPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("shardID is invalid"))
	}
	shardID := byte(shardIDTemp)

	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		Logger.log.Debugf("handleGetShardPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(UnexpectedError, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardPool.GetAllBlockHeight()
	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})
	temp := jsonresult.NewBlocksFromShardPool(*shardPool)
	Logger.log.Debugf("handleGetShardPoolStateV2 result: %+v", temp)
	return temp, nil
}

func (httpServer *HttpServer) handleGetBeaconPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetBeaconPoolStateV2 params: %+v", params)
	beaconPool := mempool.GetBeaconPool()
	if beaconPool == nil {
		Logger.log.Debugf("handleGetBeaconPoolStateV2 result: %+v", nil)
		return nil, NewRPCError(UnexpectedError, errors.New("Beacon Pool not init"))
	}
	result := jsonresult.NewBlocksFromBeaconPool(*beaconPool)
	Logger.log.Debugf("handleGetBeaconPoolStateV2 result: %+v", result)
	return result, nil
}
