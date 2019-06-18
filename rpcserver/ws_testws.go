package rpcserver

import (
	"time"
)
type Result struct {
	Result interface{}
	Subcription string
}
func (wsServer *WsServer) handleSubcribeNewBlock(params interface{}, subcription string, result chan interface{}, error chan *RPCError) {
	Logger.log.Info("Handle Subcribe New Block", params, subcription)
	for i := 0; i < 10; i++ {
		result <- Result{Result:i, Subcription:subcription}
		<-time.Tick(1 * time.Second)
	}
	close(result)
}
