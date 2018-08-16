package rpcserver

type commandHandler func(*RpcServer, interface{}, <-chan struct{}) (interface{}, error)

var RpcHandler = map[string]commandHandler{
	"dosomething":       handleDoSomething,
	"createtransaction": handleCreateTransaction,
}

// Commands that are available to a limited user
var RpcLimited = map[string]struct{}{

}

func handleDoSomething(self *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return nil, nil
}

func handleCreateTransaction(self *RpcServer, cmd interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return nil, nil
}
