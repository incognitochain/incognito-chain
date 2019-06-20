package rpcserver

import (
	"log"
	"os"
	"runtime/pprof"
)

func (httpServer *HttpServer) handleStartProfiling(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var f, err = os.Create("/data/profiling.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	return nil, nil
}

func (httpServer *HttpServer) handleStopProfiling(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	pprof.StopCPUProfile()
	return nil, nil
}
