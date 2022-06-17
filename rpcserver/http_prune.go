package rpcserver

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
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
	type data struct {
		ShardIDs []byte `json:"ShardIDs"`
	}
	d := data{}
	b, err := json.Marshal(arrayParams[0])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	err = json.Unmarshal(b, &d)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}
	s, ok := httpServer.synkerService.Synker.ShardSyncProcess[0]
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.PruneError, fmt.Errorf("Not found shard sync process"))
	}
	var result interface{}
	result, s.Pruner, err = httpServer.blockService.Prune(d.ShardIDs, s.Pruner)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.PruneError, err)
	}
	return result, nil
}
