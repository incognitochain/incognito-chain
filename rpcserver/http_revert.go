package rpcserver

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/pkg/errors"
)

func (httpServer *HttpServer) handleRevertBeacon(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info("handleRevertBeacon")
	err := httpServer.config.BlockChain.RevertBeaconState()
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return nil, nil
}

func (httpServer *HttpServer) handleRevertShard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Infof("handleRevertShard: %+v", params)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) < 1 {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Shard ID empty"))
	}
	shardIdParam, ok := arrayParams[0].(float64)
	if !ok {
		return nil, NewRPCError(ErrRPCInvalidParams, errors.New("Shard ID component invalid"))
	}
	shardID := byte(shardIdParam)
	err := httpServer.config.BlockChain.RevertShardState(shardID)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return nil, nil
}
