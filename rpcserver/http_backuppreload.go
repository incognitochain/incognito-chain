package rpcserver

import (
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/pkg/errors"
)

func (httpServer *HttpServer) handleSetBackup(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramArray, ok := params.([]interface{})
	if ok && len(paramArray) == 1 {
		setBackup, ok := paramArray[0].(bool)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("set backup is invalid"))
		}
		httpServer.config.ChainParams.IsBackup = setBackup
		return setBackup, nil
	}
	return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("no param"))
}
