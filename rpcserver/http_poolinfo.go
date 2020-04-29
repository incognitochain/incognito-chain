package rpcserver

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) hanldeGetBeaconPoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("hanldeGetBeaconPoolInfo params: %+v", params)
	blks := httpServer.synkerService.GetBeaconPoolInfo()
	result := jsonresult.NewPoolInfo(blks)
	Logger.log.Debugf("hanldeGetBeaconPoolInfo result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetShardToBeaconPoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetShardToBeaconPoolInfo params: %+v", params)
	blks := httpServer.synkerService.GetShardToBeaconPoolInfo()
	result := jsonresult.NewPoolInfo(blks)
	Logger.log.Debugf("handleGetShardToBeaconPoolInfo result: %+v", result)
	return result, nil
}
func (httpServer *HttpServer) hanldeGetShardPoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID component invalid"))
	}

	shardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID component invalid"))
	}
	Logger.log.Debugf("hanldeGetShardPoolInfo params: %+v", params)
	blks := httpServer.synkerService.GetShardPoolInfo(int(shardID))
	result := jsonresult.NewPoolInfo(blks)
	Logger.log.Debugf("handleGetShardPoolInfo result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) hanldeGetCrossShardPoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID invalid"))
	}

	shardID, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("ShardID component invalid"))
	}
	Logger.log.Debugf("hanldeGetCrossShardPoolInfo params: %+v", params)
	blks := httpServer.synkerService.GetCrossShardPoolInfo(int(shardID))
	result := jsonresult.NewPoolInfo(blks)
	Logger.log.Debugf("hanldeGetCrossShardPoolInfo result: %+v", result)
	return result, nil
}
