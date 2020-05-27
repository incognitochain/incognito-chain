package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"reflect"
)

func (wsServer *WsServer) handleSubscribeShardBestState(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	shardID := byte(arrayParams[0].(float64))
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.ShardBeststateTopic)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Beacon Beststate Block")
		wsServer.config.PubSubManager.Unsubscribe(pubsub.ShardBeststateTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				bestStateShard, ok := msg.Value.(*blockchain.ShardBestState)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.ShardBestState, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				if bestStateShard.ShardID != shardID {
					continue
				}
				cResult <- RpcSubResult{Result: *bestStateShard, Error: nil}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Shard Beststate"}}
				return
			}
		}
	}
}

func (wsServer *WsServer) handleSubscribeBeaconBestState(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subscribe Beacon Beststate", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 0 {
		err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain NO params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.BeaconBeststateTopic)
	if err != nil {
		err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe Beacon Beststate Block")
		wsServer.config.PubSubManager.Unsubscribe(pubsub.BeaconBeststateTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				bestStateBeacon, ok := msg.Value.(*blockchain.BeaconBestState)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBestState, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				cResult <- RpcSubResult{Result: *bestStateBeacon, Error: nil}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Beacon Beststate"}}
				return
			}
		}
	}
}
