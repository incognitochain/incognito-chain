package rpcserver

import (
	"fmt"
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

func (httpServer *HttpServer) handleGetLatestBackup(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	paramArray, ok := params.([]interface{})
	//fmt.Println("handleGetLatestBackup", paramArray)
	if ok && len(paramArray) == 1 {

		chainName, ok := paramArray[0].(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("chainName is invalid"))
		}

		return httpServer.config.BlockChain.GetBeaconChainDatabase().LatestBackup(fmt.Sprintf("../../backup/%v", chainName)), nil
	}

	return 0, nil
}
