package rpcserver

import (
	"time"
)
func (wsServer *WsServer) handleSubcribeNewBlock(params interface{}, subcription string, result chan interface{}, error chan *RPCError) {
	for i := 0; i < 10; i++ {
		result <- i
		<-time.Tick(1 * time.Second)
	}
	close(result)
}
