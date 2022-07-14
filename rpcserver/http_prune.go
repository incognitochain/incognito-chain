package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/pruner"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"sync"
)

func (httpServer *HttpServer) handlePrune(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	if !config.Config().AllowStatePrune {
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
		if int(shardID) > config.Param().ActiveShards {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("shardID is %v is invalid", shardID))
		}
		ec := pruner.ExtendedConfig{
			Config:  pruner.Config{ShouldPruneByHash: c.ShouldPruneByHash},
			ShardID: shardID,
		}
		httpServer.Pruner.TriggerCh <- ec
	}
	type Result struct {
		Message string `json:"Message"`
	}
	return Result{Message: "Success"}, nil
}

func (httpServer *HttpServer) getPruneState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	type Temp struct {
		Status       byte   `json:"Status"`
		PrunedHeight uint64 `json:"PrunedHeight"`
	}
	results := make(map[byte]Temp)
	for i := 0; i < common.MaxShardNumber; i++ {
		status, _ := rawdbv2.GetPruneStatus(httpServer.GetShardChainDatabase(byte(i)))
		prunedHeight, _ := rawdbv2.GetLastPrunedHeight(httpServer.GetShardChainDatabase(byte(i)))
		temp := Temp{
			Status:       status,
			PrunedHeight: prunedHeight,
		}
		results[byte(i)] = temp
	}
	return results, nil
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
