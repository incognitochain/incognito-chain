package rpcserver

import (
	"errors"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/mempool"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
)

/*
handleGetShardToBeaconPoolState - RPC get shard to beacon pool state
*/
func (rpcServer RpcServer) handleGetShardToBeaconPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	shardToBeaconPool := mempool.GetShardToBeaconPool()
	if shardToBeaconPool == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard to Beacon Pool not init"))
	}
	result := shardToBeaconPool.GetAllBlockHeight()
	return result, nil
}

/*
handleGetCrossShardPoolState - RPC get cross shard pool state
*/
func (rpcServer RpcServer) handleGetCrossShardPoolState(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// get component
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	shardID := byte(paramsArray[0].(float64))

	result := mempool.GetCrossShardPool(shardID).GetAllBlockHeight()
	// if !ok || result == nil {
	// 	return nil, NewRPCError(ErrUnexpected, errors.New("Best State shard given by ID not existed"))
	// }
	// result.BestShardBlock = nil
	return result, nil
}

func (rpcServer RpcServer) handleGetNextCrossShard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	// get component
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 3 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("invalid list Key component"))
	}
	fromShard := byte(paramsArray[0].(float64))
	toShard := byte(paramsArray[1].(float64))
	startHeight := uint64(paramsArray[2].(float64))

	result := mempool.GetCrossShardPool(toShard).GetNextCrossShardHeight(fromShard, toShard, startHeight)
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

//==============Version 2================
/*
handleGetShardToBeaconPoolState - RPC get shard to beacon pool state
*/
func (rpcServer RpcServer) handleGetShardToBeaconPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) != 0 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("wrong format params"))
	}
	shardToBeaconPool := mempool.GetShardToBeaconPool()
	if shardToBeaconPool == nil {
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
	return shardToBeaconPoolResult, nil
}

/*
handleGetCrossShardPoolState - RPC get cross shard pool state
*/
func (rpcServer RpcServer) handleGetCrossShardPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var index = 0
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) != 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("wrong format params"))
	}
	shardID := byte(paramsArray[0].(float64))

	crossShardPool := mempool.GetCrossShardPool(shardID)
	if crossShardPool == nil {
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
	return crossShardPoolResult, nil
}

/*
handleGetShardPoolState - RPC get shard block in pool
*/
func (rpcServer RpcServer) handleGetShardPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) != 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("wrong format params"))
	}
	shardID := byte(paramsArray[0].(float64))
	shardPool := mempool.GetShardPool(shardID)
	if shardPool == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Shard Pool not init"))
	}
	blockHeights := shardPool.GetAllBlockHeight()
	latestBlockHeight := shardPool.GetLatestValidBlockHeight()
	shardBlockPoolResult := jsonresult.ShardBlockPoolResult{ShardID: shardID}
	for _, blockHeight := range blockHeights {
		if blockHeight <= latestBlockHeight {
			shardBlockPoolResult.ValidBlockHeight = append(shardBlockPoolResult.ValidBlockHeight, blockHeight)
		} else {
			shardBlockPoolResult.PendingBlockHeight = append(shardBlockPoolResult.PendingBlockHeight, blockHeight)
		}
	}
	return shardBlockPoolResult, nil
}

func (rpcServer RpcServer) handleGetBeaconPoolStateV2(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) != 0 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("wrong format params"))
	}
	beaconPool := mempool.GetBeaconPool()
	if beaconPool == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Beacon Pool not init"))
	}
	blockHeights := beaconPool.GetAllBlockHeight()
	latestBlockHeight := beaconPool.GetLatestValidBlockHeight()
	beaconBlockPoolResult := jsonresult.BeaconBlockPoolResult{
		ValidBlockHeight:   []uint64{},
		PendingBlockHeight: []uint64{},
	}
	for _, blockHeight := range blockHeights {
		if blockHeight <= latestBlockHeight {
			beaconBlockPoolResult.ValidBlockHeight = append(beaconBlockPoolResult.ValidBlockHeight, blockHeight)
		} else {
			beaconBlockPoolResult.PendingBlockHeight = append(beaconBlockPoolResult.PendingBlockHeight, blockHeight)
		}
	}
	return beaconBlockPoolResult, nil
}
