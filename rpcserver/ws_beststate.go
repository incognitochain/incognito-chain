package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"reflect"
)

func (wsServer *WsServer) handleSubcribeShardBestState(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe Shard Beststate", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := NewRPCError(ErrRPCInvalidParams, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	shardID := byte(arrayParams[0].(float64))
	subId, subChan, err := wsServer.config.PubsubManager.RegisterNewSubcriber(pubsub.ShardBeststateTopic)
	if err != nil {
		err := NewRPCError(ErrSubcribe, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subcribe Beacon Beststate Block")
		wsServer.config.PubsubManager.Unsubcribe(pubsub.ShardBeststateTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				bestStateShard, ok := msg.Value.(*blockchain.BestStateShard)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BestStateShard, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				if bestStateShard.ShardID != shardID {
					continue
				}
				cResult <- RpcSubResult{Result: *bestStateShard, Error: nil}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubcribe Shard Beststate"}}
				return
			}
		}
	}
}

func (wsServer *WsServer) handleSubcribeBeaconBestState(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe Beacon Beststate", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 0 {
		err := NewRPCError(ErrRPCInvalidParams, errors.New("Methods should only contain NO params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	subId, subChan, err := wsServer.config.PubsubManager.RegisterNewSubcriber(pubsub.BeaconBeststateTopic)
	if err != nil {
		err := NewRPCError(ErrSubcribe, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subcribe Beacon Beststate Block")
		wsServer.config.PubsubManager.Unsubcribe(pubsub.BeaconBeststateTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				bestStateBeacon, ok := msg.Value.(*blockchain.BestStateBeacon)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BestStateBeacon, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				cResult <- RpcSubResult{Result: *bestStateBeacon, Error: nil}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubcribe Beacon Beststate"}}
				return
			}
		}
	}
}
