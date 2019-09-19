package rpcserver

import (
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) handleGetProducersBlackList(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	producersBlackList, err := httpServer.databaseService.GetProducersBlackList()
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return producersBlackList, nil
}
