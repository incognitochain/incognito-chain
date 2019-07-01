package rpcserver

import (
	"time"
)

type Result struct {
	Result      interface{}
	Subcription string
}

func (wsServer *WsServer) handleTestSubcribe(params interface{}, subcription string, result chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe New Block", params, subcription)
	for i := 0; i < 10; i++ {
		result <- RpcSubResult{Result: i, Error: nil}
		<-time.Tick(1 * time.Second)
	}
	close(result)
}
