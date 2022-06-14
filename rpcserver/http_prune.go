package rpcserver

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) handlePrune(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
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
	result, err := httpServer.blockService.Prune(d.ShardIDs)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.PruneError, err)
	}
	return result, nil
}
