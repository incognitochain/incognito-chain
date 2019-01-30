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
	// beacon := rpcServer.config.BlockChain.BestState.Beacon
	// result := jsonresult.BeaconBestStateResult{}
	// result.BestBlockHash = beacon.BestBlockHash
	// result.BestShardHash = beacon.BestShardHash
	// result.BestShardHeight = beacon.BestShardHeight
	// result.AllShardState = beacon.AllShardState
	// result.BeaconEpoch = beacon.BeaconEpoch
	// result.BeaconHeight = beacon.BeaconHeight
	// result.BeaconProposerIdx = beacon.BeaconProposerIdx
	// result.BeaconCommittee = beacon.BeaconCommittee
	// result.BeaconPendingValidator = beacon.BeaconPendingValidator

	// result.CandidateShardWaitingForCurrentRandom = beacon.CandidateShardWaitingForCurrentRandom
	// result.CandidateBeaconWaitingForCurrentRandom = beacon.CandidateBeaconWaitingForCurrentRandom

	// result.CandidateShardWaitingForNextRandom = beacon.CandidateShardWaitingForNextRandom
	// result.CandidateBeaconWaitingForNextRandom = beacon.CandidateBeaconWaitingForNextRandom

	// result.ShardCommittee = beacon.ShardCommittee
	// result.ShardPendingValidator = beacon.ShardPendingValidator

	// result.CurrentRandomNumber = beacon.CurrentRandomNumber
	// result.CurrentRandomTimeStamp = beacon.CurrentRandomTimeStamp
	// result.IsGetRandomNumber = beacon.IsGetRandomNumber
	// result.Params = beacon.Params
	result := rpcServer.config.BlockChain.BestState.Beacon
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
	result.BestShardBlock = nil
	return result, nil
}
