package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/pubsub"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"reflect"
)

func (wsServer *WsServer) handleSubcribeMempoolInfo(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	Logger.log.Info("Handle Subcribe Mempool Informantion", params, subcription)
	arrayParams := common.InterfaceSlice(params)
	if len(arrayParams) != 0 {
		err := NewRPCError(ErrRPCInvalidParams, errors.New("Methods should only contain NO params"))
		cResult <- RpcSubResult{Error: err}
		return
	}
	subId, subChan, err := wsServer.config.PubsubManager.RegisterNewSubcriber(pubsub.MempoolInfoTopic)
	if err != nil {
		err := NewRPCError(ErrSubcribe, err)
		cResult <- RpcSubResult{Error: err}
		return
	}
	defer func() {
		Logger.log.Info("Finish Subcribe Mempool Informantion")
		wsServer.config.PubsubManager.Unsubcribe(pubsub.MempoolInfoTopic, subId)
		close(cResult)
	}()
	for {
		select {
		case msg := <-subChan:
			{
				listTxs, ok := msg.Value.([]string)
				if !ok {
					Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted []string, have %+v", reflect.TypeOf(msg.Value))
					continue
				}
				cResult <- RpcSubResult{Result: listTxs, Error: nil}
			}
		case <-closeChan:
			{
				cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubcribe Mempool Info"}}
				return
			}
		}
	}
}

