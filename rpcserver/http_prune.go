package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/pruner"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
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
		httpServer.Pruner.ForwardCh <- ec
	}
	type Result struct {
		Message string `json:"Message"`
	}
	return Result{Message: "Success"}, nil
}
