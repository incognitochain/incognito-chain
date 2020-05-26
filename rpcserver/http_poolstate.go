package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) handleGetNextCrossShard(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// get component
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) < 3 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 3 elements"))
	}

	fromShardTemp, ok := paramsArray[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("from shard id param is invalid"))
	}
	fromShard := byte(fromShardTemp)

	toShardTemp, ok := paramsArray[1].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("to shard id param is invalid"))
	}
	toShard := byte(toShardTemp)

	startHeightTemp, ok := paramsArray[2].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("start height param is invalid"))
	}
	startHeight := uint64(startHeightTemp)

	result := httpServer.poolStateService.GetNextCrossShard(fromShard, toShard, startHeight)
	return result, nil
}

func (httpServer *HttpServer) handleGetBeaconPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result, err := httpServer.poolStateService.GetBeaconPoolState()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return result, nil
}

func (httpServer *HttpServer) handleGetShardPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// get params
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}

	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shard id param is invalid"))
	}
	shardID := byte(shardIDTemp)

	shardPool, err := httpServer.poolStateService.GetShardPoolState(shardID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("shard to Beacon Pool not init"))
	}

	result := jsonresult.NewBlocksFromShardPool(shardPool)
	return result, nil
}

func (httpServer *HttpServer) handleGetShardPoolLatestValidHeight(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	// get params
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}

	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shard id param is invalid"))
	}
	shardID := byte(shardIDTemp)

	result, err := httpServer.poolStateService.GetShardPoolLatestValidHeight(shardID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("shard to Beacon Pool not init"))
	}

	return result, nil
}

//==============Version 2================
/*
handleGetShardToBeaconPoolState - RPC get shard to beacon pool state
*/
func (httpServer *HttpServer) handleGetShardToBeaconPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) != 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("wrong format params"))
	}

	allBlockHeight, allLatestBlockHeight, err := httpServer.poolStateService.GetShardToBeaconPoolStateV2()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	shardToBeaconPoolResult := jsonresult.NewShardToBeaconPoolResult(allBlockHeight, allLatestBlockHeight)
	return shardToBeaconPoolResult, nil
}

/*
handleGetCrossShardPoolState - RPC get cross shard pool state
*/
func (httpServer *HttpServer) handleGetCrossShardPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {

	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("wrong format params"))
	}

	shardIDParam, ok := paramsArray[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shard id param is invalid"))
	}
	shardID := byte(shardIDParam)

	allValidBlockHeight, allPendingBlockHeight, err := httpServer.poolStateService.GetCrossShardPoolStateV2(shardID)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	crossShardPoolResult := jsonresult.NewCrossShardPoolResult(allValidBlockHeight, allPendingBlockHeight)

	return crossShardPoolResult, nil
}

/*
handleGetShardPoolState - RPC get shard block in pool
*/
func (httpServer *HttpServer) handleGetShardPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}

	shardIDTemp, ok := paramsArray[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("shardID is invalid"))
	}
	shardID := byte(shardIDTemp)

	shardPool, err := httpServer.poolStateService.GetShardPoolStateV2(shardID)
	if err == nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	result := jsonresult.NewBlocksFromShardPool(shardPool)
	return result, nil
}

func (httpServer *HttpServer) handleGetBeaconPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	beaconPool, err := httpServer.poolStateService.GetBeaconPoolStateV2()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewBlocksFromBeaconPool(beaconPool)
	return result, nil
}
