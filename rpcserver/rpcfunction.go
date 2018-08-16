package rpcserver

import "log"

var RpcHandler = map[string]interface{}{
	"dosomething":       handleDoSomething,
	"createtransaction": handleCreateTransaction,
}

// Commands that are available to a limited user
var RpcLimited = map[string]struct{}{

}

func handleDoSomething(self *RpcServer, params interface{}, closeChan <-chan struct{}) {
	log.Println(params)
}

func handleCreateTransaction(self *RpcServer, params interface{}, closeChan <-chan struct{}) {
	log.Println(params)
}
