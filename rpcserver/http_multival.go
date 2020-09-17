package rpcserver

import (
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

// getMultiValKeyState = "getmultivalkeystate"
// addMultiValKey      = "addmultivakey"
// setMultiValKeyLimit = "setmultivalkeylimit"

func (httpServer *HttpServer) handleGetMultiValKeyState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.NewGetMiningInfoResult(*httpServer.config.TxMemPool, *httpServer.config.BlockChain, httpServer.config.ConsensusEngine, *httpServer.config.ChainParams, httpServer.config.Server.IsEnableMining())
	return result, nil
}
func (httpServer *HttpServer) handleAddMultiValKey(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.NewGetMiningInfoResult(*httpServer.config.TxMemPool, *httpServer.config.BlockChain, httpServer.config.ConsensusEngine, *httpServer.config.ChainParams, httpServer.config.Server.IsEnableMining())
	return result, nil
}
func (httpServer *HttpServer) handleSetMultiValKeyLimit(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result := jsonresult.NewGetMiningInfoResult(*httpServer.config.TxMemPool, *httpServer.config.BlockChain, httpServer.config.ConsensusEngine, *httpServer.config.ChainParams, httpServer.config.Server.IsEnableMining())
	return result, nil
}
