package server

type commandHandler func(RpcServer, interface{}, <-chan struct{}) (interface{}, error)

var RpcHandler = map[string]commandHandler{
	"ping": RpcServer.handlePing,
}

// Commands that are available to a limited user
var RpcLimited = map[string]struct{}{}

func (self RpcServer) handlePing(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	sParams := params.(map[string]string)
	ID := sParams["ID"]
	self.AddPeer(ID)

	result := make(map[string]interface{})
	result["peers"] = self.peers
	return result, nil
}