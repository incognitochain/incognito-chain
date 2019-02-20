package rpcserver

import (
	"errors"
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

	result := mempool.GetCrossShardPoolState()
	// if !ok || result == nil {
	// 	return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard given by ID not existed"))
	// }
	// result.BestShardBlock = nil
	return result, nil
}
