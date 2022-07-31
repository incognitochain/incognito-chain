package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/pruner"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) handlePrune(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	if !config.Config().AllowStatePruneByRPC {
		return nil, rpcservice.NewRPCError(rpcservice.PruneError, errors.New("Node is not able to prune"))
	}
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Payload data is invalid"))
	}
	type Temp struct {
		Config map[byte]pruner.Config `json:"Config"`
	}
	t := Temp{}
	b, err := json.Marshal(arrayParams[0])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	err = json.Unmarshal(b, &t)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	for shardID, c := range t.Config {
		if int(shardID) >= common.MaxShardNumber {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("shardID is %v is invalid", shardID))
		}
		httpServer.Pruner.JobRquest[int(shardID)] = &pruner.Config{ShouldPruneByHash: c.ShouldPruneByHash}
	}
	type Result struct {
		Message string `json:"Message"`
	}
	return Result{Message: "Success"}, nil
}

func (httpServer *HttpServer) getPruneState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	report := httpServer.Pruner.Report()
	return report, nil
}

func (httpServer *HttpServer) checkPruneData(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	res := map[int]bool{}
	wg := sync.WaitGroup{}
	for i := 0; i < common.MaxShardNumber; i++ {
		wg.Add(1)
		go func(sid int) {
			if err := httpServer.GetBlockchain().GetBestStateShard(byte(sid)).GetCopiedTransactionStateDB().Recheck(); err != nil {
				res[sid] = false
			} else {
				res[sid] = true
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	fmt.Println("checkPruneData", res)
	return res, nil
}
