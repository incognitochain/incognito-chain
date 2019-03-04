package rpcserver

import (
	"errors"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/mempool"
)

/*
handleGetShardToBeaconPoolState - RPC get shard to beacon pool state
*/
func (rpcServer RpcServer) handleGetShardToBeaconPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// if rpcServer.config.BlockChain.BestState.Beacon == nil {
	// 	return nil, NewRPCError(ErrUnexpected, errors.New("Best State beacon not existed"))
	// }
	shardToBeaconPool := mempool.GetShardToBeaconPool()
	if shardToBeaconPool == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardToBeaconPool.GetAllPendingBlockHeight()
	// result.BestBlock = nil
	return result, nil
}

/*
handleGetCrossShardPoolState - RPC get cross shard pool state
*/
func (rpcServer RpcServer) handleGetCrossShardPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// get params
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key params"))
	}
	shardID := byte(paramsArray[0].(int))

	result := mempool.GetCrossShardPool(shardID).GetAllBlockHeight()
	// if !ok || result == nil {
	// 	return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard given by ID not existed"))
	// }
	// result.BestShardBlock = nil
	return result, nil
}

func (rpcServer RpcServer) handleGetBeaconPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// if rpcServer.config.BlockChain.BestState.Beacon == nil {
	// 	return nil, NewRPCError(ErrUnexpected, errors.New("Best State beacon not existed"))
	// }
	beaconPool := mempool.GetBeaconPool()
	if beaconPool == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Beacon Pool not init"))
	}
	result := beaconPool.GetAllBlockHeight()
	// result.BestBlock = nil
	return result, nil
}

func (rpcServer RpcServer) handleGetShardPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// if rpcServer.config.BlockChain.BestState.Beacon == nil {
	// 	return nil, NewRPCError(ErrUnexpected, errors.New("Best State beacon not existed"))
	// }
	// get params
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key params"))
	}
	shardID := byte(paramsArray[0].(float64))

	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardPool.GetAllBlockHeight()
	// result.BestBlock = nil
	return result, nil
}

func (rpcServer RpcServer) handleGetShardPoolLatestValidHeight(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// if rpcServer.config.BlockChain.BestState.Beacon == nil {
	// 	return nil, NewRPCError(ErrUnexpected, errors.New("Best State beacon not existed"))
	// }
	// get params
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key params"))
	}
	shardID := byte(paramsArray[0].(float64))

	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardPool.GetLatestValidBlockHeight()
	// result.BestBlock = nil
	return result, nil
}
