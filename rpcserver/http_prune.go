package rpcserver

import (
	"encoding/json"
	"errors"

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
	type Temp struct {
		Data map[byte]rpcservice.PruneData `json:"Data"`
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
	/*s, ok := httpServer.synkerService.Synker.ShardSyncProcess[0]*/
	/*if !ok {*/
	/*return nil, rpcservice.NewRPCError(rpcservice.PruneError, fmt.Errorf("Not found shard sync process"))*/
	/*}*/
	var result interface{}
	/*result, s.Pruner, err = httpServer.blockService.Prune(t.Data, s.Pruner)*/
	/*if err != nil {*/
	/*return nil, rpcservice.NewRPCError(rpcservice.PruneError, err)*/
	/*}*/
	return result, nil
}
