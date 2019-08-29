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
	result := jsonresult.NewGetMiningInfoResult(*httpServer.config.TxMemPool, *httpServer.config.BlockChain, httpServer.config.ConsensusEngine, *httpServer.config.ChainParams, httpServer.config.Server.IsEnableMining())
	Logger.log.Debugf("handleGetMiningInfo result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleEnableMining(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("EnableParam empty"))
	}
	enableParam, ok := arrayParams[0].(bool)
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("EnableParam component invalid"))
	}
	return httpServer.config.Server.EnableMining(enableParam), nil
}

func (httpServer *HttpServer) handleGetChainMiningStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("Chain ID empty"))
	}
	chainIDParam, ok := arrayParams[0].(float64)
	if !ok {
		return nil, NewRPCError(RPCInvalidParamsError, errors.New("Chain ID component invalid"))
	}
	return httpServer.config.Server.GetChainMiningStatus(int(chainIDParam)), nil
}
