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
	if httpServer.config.MiningKeys == "" {
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

	// role, shardID := httpServer.config.BlockChain.BestState.Beacon.GetPubkeyRole(httpServer.config.MiningPubKeyB58, 0)
	role, shardID := httpServer.config.ConsensusEngine.GetUserRole()
	result.Role = role

	switch shardID {
	case -2:
		result.ShardID = -2
		result.IsCommittee = false
	case -1:
		result.IsCommittee = true
	default:
		result.ShardHeight = httpServer.config.BlockChain.BestState.Shard[byte(shardID)].ShardHeight
		result.CurrentShardBlockTx = len(httpServer.config.BlockChain.BestState.Shard[byte(shardID)].BestBlock.Body.Transactions)
		result.ShardID = shardID
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
