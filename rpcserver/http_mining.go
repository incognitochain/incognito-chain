package rpcserver

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/pkg/errors"
)

/*
handleGetMiningInfo - RPC returns various mining-related info
*/
func (httpServer *HttpServer) handleGetMiningInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Debugf("handleGetMiningInfo params: %+v", params)
	if !httpServer.config.IsMiningNode || httpServer.config.MiningPubKeyB58 == "" {
		return jsonresult.GetMiningInfoResult{
			IsCommittee: false,
		}, nil
	}

	result := jsonresult.GetMiningInfoResult{}
	result.IsCommittee = true
	result.PoolSize = httpServer.config.TxMemPool.Count()
	result.Chain = httpServer.config.ChainParams.Name
	result.IsEnableMining = httpServer.config.Server.IsEnableMining()
	result.BeaconHeight = httpServer.config.BlockChain.BestState.Beacon.BeaconHeight

	role, shardID := httpServer.config.BlockChain.BestState.Beacon.GetPubkeyRole(httpServer.config.MiningPubKeyB58, 0)
	result.Role = role
	if role == common.SHARD_ROLE {
		result.ShardHeight = httpServer.config.BlockChain.BestState.Shard[shardID].ShardHeight
		result.CurrentShardBlockTx = len(httpServer.config.BlockChain.BestState.Shard[shardID].BestBlock.Body.Transactions)
		result.ShardID = int(shardID)
	} else if role == common.VALIDATOR_ROLE || role == common.PROPOSER_ROLE || role == common.PENDING_ROLE {
		result.ShardID = -1
	}
	Logger.log.Debugf("handleGetMiningInfo result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleEnableMining(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("EnableParam empty"))
	}
	enableParam, ok := arrayParams[0].(bool)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("EnableParam component invalid"))
	}
	return httpServer.config.Server.EnableMining(enableParam), nil
}

func (httpServer *HttpServer) handleGetChainMiningStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Chain ID empty"))
	}
	chainIDParam, ok := arrayParams[0].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Chain ID component invalid"))
	}
	return httpServer.config.Server.GetChainMiningStatus(int(chainIDParam)), nil
}
