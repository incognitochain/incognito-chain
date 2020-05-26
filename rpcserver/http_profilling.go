package rpcserver

import (
	"github.com/incognitochain/incognito-chain/metrics"
	"github.com/incognitochain/incognito-chain/metrics/exp"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"log"
	"os"
	"runtime/pprof"
)

func (httpServer *HttpServer) handleStartProfiling(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var f, err = os.OpenFile("/data/profiling.prof", os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	return nil, nil
}

func (httpServer *HttpServer) handleStopProfiling(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	pprof.StopCPUProfile()
	return nil, nil
}

func (httpServer *HttpServer) handleExportMetrics(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	exporter := exp.NewExp(metrics.DefaultRegistry)
	return exporter.Export(), nil
}
