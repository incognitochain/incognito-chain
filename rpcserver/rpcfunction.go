package rpcserver

import (
	"log"
	"encoding/json"
	"github.com/internet-cash/prototype/rpcserver/jsonrpc"
)

type commandHandler func(*RpcServer, interface{}, <-chan struct{}) (interface{}, error)

var RpcHandler = map[string]commandHandler{
	"dosomething":       handleDoSomething,
	"getblockchaininfo": handleGetBlockChainInfo,
	"createtransaction": handleCreateTransaction,
}

// Commands that are available to a limited user
var RpcLimited = map[string]struct{}{

}

func handleGetBlockChainInfo(self *RpcServer, params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	result := jsonrpc.GetBlockChainInfoResponse{
		Chain:  self.Config.ChainParams.Name,
		Blocks: len(self.Config.Chain.Blocks),
	}
	return result, nil
}

func handleDoSomething(self *RpcServer, params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	result := make(map[string]string)
	result["param"] = string(params.([]json.RawMessage)[0])
	return result, nil
}

func handleCreateTransaction(self *RpcServer, params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	log.Println(params)
	return nil, nil
}
