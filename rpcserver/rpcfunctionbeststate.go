package rpcserver

import (
	"errors"

	"github.com/ninjadotorg/constant/common"
)

/*
handleGetBeaconBestState - RPC get beacon best state
*/
func (rpcServer RpcServer) handleGetBeaconBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	if rpcServer.config.BlockChain.BestState.Beacon == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State beacon not existed"))
	}

	result := *rpcServer.config.BlockChain.BestState.Beacon
	result.BestBlock = nil
	return result, nil
}

/*
handleGetShardBestState - RPC get shard best state
*/
func (rpcServer RpcServer) handleGetShardBestState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Shard ID params empty"))
	}
	shardIdParam, ok := arrayParams[0].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Shard ID params invalid"))
	}
	shardID := byte(shardIdParam)
	if rpcServer.config.BlockChain.BestState.Shard == nil || len(rpcServer.config.BlockChain.BestState.Shard) <= 0 {
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard not existed"))
	}
	result, ok := rpcServer.config.BlockChain.BestState.Shard[shardID]
	if !ok || result == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard given by ID not existed"))
	}
	result.BestBlock = nil
	return result, nil
}
