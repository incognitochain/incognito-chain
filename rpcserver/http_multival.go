package rpcserver

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

// getMultiValKeyState = "getmultivalkeystate"
// addMultiValKey      = "addmultivakey"
// setMultiValKeyLimit = "setmultivalkeylimit"

func (httpServer *HttpServer) handleGetMultiValKeyState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	if httpServer.config.DisableAuth {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("Auth is no enable"))
	}
	states := httpServer.config.ConsensusEngine.GetAllValidatorKeyState()
	return states, nil
}

func (httpServer *HttpServer) handleAddMultiValKey(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	if httpServer.config.DisableAuth {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("Auth is no enable"))
	}
	var key string
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) > 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array of 1 element"))
	}
	if paramsArray[0] != nil {
		strParam, ok := paramsArray[0].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("key param is invalid"))
		}
		key = strParam
	}
	err := httpServer.config.ConsensusEngine.AddValidatorKey(key)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return "ok", nil
}

func (httpServer *HttpServer) handleSetMultiValKeyLimit(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	if httpServer.config.DisableAuth {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("Auth is no enable"))
	}
	var limit int
	paramsArray := common.InterfaceSlice(params)
	if paramsArray == nil || len(paramsArray) > 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array of 1 element"))
	}
	if paramsArray[0] != nil {
		intParam, ok := paramsArray[0].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("limit param is invalid"))
		}
		limit = int(intParam)
	}

	err := httpServer.config.ConsensusEngine.SetValidatorKeyLimit(limit)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	return "ok", nil
}
