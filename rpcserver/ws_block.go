package rpcserver

import (
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"reflect"
)

func (wsServer *WsServer) handleSubscribeNewShardBlock(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subscribe New Block", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 1 {
		err := NewRPCError(ErrRPCInvalidParams, errors.New("Methods should only contain 1 params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	shardID := byte(arrayParams[0].(float64))
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewShardblockTopic)
	if err != nil {
		err := NewRPCError(ErrSubcribe, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe New Shard Block ShardID ", shardID)
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewShardblockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				shardBlock, ok := msg.Value.(*blockchain.ShardBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.ShardBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				if shardBlock.Header.ShardID != shardID {
					continue
				}
				blockResult := jsonresult.GetBlockResult{}
				blockBytes, err := json.Marshal(shardBlock)
				if err != nil {
					cResult <- RpcSubResult{Error: NewRPCError(ErrUnexpected, err)}
					return
				}
				blockResult.Init(shardBlock, uint64(len(blockBytes)))
				cResult <- RpcSubResult{Result: blockResult, Error: nil}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe New Shard Block"}}
				return
			}
		}
	}
}

func (wsServer *WsServer) handleSubscribeNewBeaconBlock(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subscribe New Block", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 0 {
		err := NewRPCError(ErrRPCInvalidParams, errors.New("Methods should only contain NO params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.NewBeaconBlockTopic)
	if err != nil {
		err := NewRPCError(ErrSubcribe, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subscribe New Beacon Block")
		wsServer.config.PubSubManager.Unsubscribe(pubsub.NewBeaconBlockTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				beaconBlock, ok := msg.Value.(*blockchain.BeaconBlock)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted *blockchain.BeaconBlock, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				blockBeaconResult := jsonresult.GetBlocksBeaconResult{}
				blockBytes, err := json.Marshal(beaconBlock)
				if err != nil {
					cResult <- RpcSubResult{Error: NewRPCError(ErrUnexpected, err)}
					return
				}
				blockBeaconResult.Init(beaconBlock, uint64(len(blockBytes)))
				cResult <- RpcSubResult{Result: blockBeaconResult, Error: nil}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe New Beacon Block"}}
				return
			}
		}
	}
}
