package rpcserver

import (
	"log"
	"os"
	"runtime/pprof"
)

func (rpcServer RpcServer) handleStartProfiling(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	var f, err = os.Create("/data/profiling.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	return nil, nil
}

func (rpcServer RpcServer) handleStopProfiling(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	pprof.StopCPUProfile()
	return nil, nil
}
