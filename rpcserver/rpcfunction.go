package rpcserver

import "log"

var RpcHandler = map[string]interface{}{
	"dosomething":       handleDoSomething,
	"createtransaction": handleCreateTransaction,
}

// Commands that are available to a limited user
var RpcLimited = map[string]struct{}{

}

func handleDoSomething(self *RpcServer, cmd interface{}, closeChan <-chan struct{}) {
	log.Println(cmd)
}

func handleCreateTransaction(self *RpcServer, cmd interface{}, closeChan <-chan struct{}) {
	log.Println(cmd)
}
