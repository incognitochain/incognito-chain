package rpcserver

func (wsServer *WsServer) handleSubcribeMempoolInfo(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	// Logger.log.Info("Handle Subcribe Mempool Informantion", params, subcription)
	// arrayParams := common.InterfaceSlice(params)
	// if len(arrayParams) != 0 {
	// 	err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain NO params"))
	// 	cResult <- RpcSubResult{Error: err}
	// 	return
	// }
	// subId, subChan, err := wsServer.config.PubSubManager.RegisterNewSubscriber(pubsub.MempoolInfoTopic)
	// if err != nil {
	// 	err := rpcservice.NewRPCError(rpcservice.SubcribeError, err)
	// 	cResult <- RpcSubResult{Error: err}
	// 	return
	// }
	// defer func() {
	// 	Logger.log.Info("Finish Subcribe Mempool Informantion")
	// 	wsServer.config.PubSubManager.Unsubscribe(pubsub.MempoolInfoTopic, subId)
	// 	close(cResult)
	// }()
	// for {
	// 	select {
	// 	case msg := <-subChan:
	// 		{
	// 			listTxs, ok := msg.Value.([]string)
	// 			if !ok {
	// 				Logger.log.Errorf("Wrong Message Type from Pubsub Manager, wanted []string, have %+v", reflect.TypeOf(msg.Value))
	// 				continue
	// 			}
	// 			cResult <- RpcSubResult{Result: listTxs, Error: nil}
	// 		}
	// 	case <-closeChan:
	// 		{
	// 			cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Mempool Info"}}
	// 			return
	// 		}
	// 	}
	// }
}

func (wsServer *WsServer) handleSubscribeBeaconPoolBestState(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	// Logger.log.Info("Handle Subscribe Beacon Pool Beststate", params, subcription)
	// arrayParams := common.InterfaceSlice(params)
	// if len(arrayParams) != 0 {
	// 	err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain NO params"))
	// 	cResult <- RpcSubResult{Error: err}
	// 	return
	// }
	// beaconPool := mempool.GetBeaconPool()
	// if beaconPool == nil {
	// 	Logger.log.Error("Beacon pool not found")
	// 	return
	// }
	// defer func() {
	// 	Logger.log.Info("Finish Subscribe Beacon Pool Beststate")
	// 	close(cResult)
	// }()

	// for {
	// 	select {
	// 	case <-closeChan:
	// 		{
	// 			cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Mempool Info"}}
	// 			return
	// 		}
	// 	default:
	// 		{
	// 			result := jsonresult.Blocks{Valid: beaconPool.GetValidBlockHeight(), Pending: beaconPool.GetPendingBlockHeight(), Latest: beaconPool.GetBeaconState()}
	// 			cResult <- RpcSubResult{Result: result, Error: nil}
	// 			time.Sleep(1 * time.Second)
	// 		}
	// 	}
	// }
}

func (wsServer *WsServer) handleSubscribeShardPoolBeststate(params interface{}, subcription string, cResult chan RpcSubResult, closeChan <-chan struct{}) {
	// Logger.log.Info("Handle Subscribe Shard Pool Beststate", params, subcription)
	// arrayParams := common.InterfaceSlice(params)
	// if len(arrayParams) != 1 {
	// 	err := rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("Methods should only contain 1 params"))
	// 	cResult <- RpcSubResult{Error: err}
	// 	return
	// }
	// shardID := byte(arrayParams[0].(float64))
	// shardPool := mempool.GetShardPool(shardID)
	// if shardPool == nil {
	// 	Logger.log.Errorf("Shard pool SHARDID %+v not found\n", shardID)
	// 	return
	// }
	// defer func() {
	// 	Logger.log.Info("Finish Subscribe Shard Pool Beststate")
	// 	close(cResult)
	// }()
	// for {
	// 	select {

	// 	case <-closeChan:
	// 		{
	// 			cResult <- RpcSubResult{Result: jsonresult.UnsubcribeResult{Message: "Unsubscribe Mempool Info"}}
	// 			return
	// 		}
	// 	default:
	// 		{
	// 			result := jsonresult.Blocks{Valid: shardPool.GetValidBlockHeight(), Pending: shardPool.GetPendingBlockHeight(), Latest: shardPool.GetShardState()}
	// 			cResult <- RpcSubResult{Result: result, Error: nil}
	// 			time.Sleep(1 * time.Second)
	// 		}
	// 	}
	// }
}
