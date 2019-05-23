package rpcserver

// func (rpcServer RpcServer) handleCreateContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
// 	constructor := metaConstructors[CreateAndSendContractingRequest]
// 	return rpcServer.createRawTxWithMetadata(params, closeChan, constructor)
// }

// func (rpcServer RpcServer) handleSendContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
// 	return rpcServer.sendRawTxWithMetadata(params, closeChan)
// }

// // handleCreateAndSendContractingRequest for user to sell Constant and receive either USD or ETH
// func (rpcServer RpcServer) handleCreateAndSendContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
// 	return rpcServer.createAndSendTxWithMetadata(
// 		params,
// 		closeChan,
// 		RpcServer.handleCreateContractingRequest,
// 		RpcServer.handleSendContractingRequest,
// 	)
// }

// handleGetIssuingStatus returns status accept/refund of a reserve issuing tx
// func (rpcServer RpcServer) handleGetIssuingStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
// 	arrayParams := common.InterfaceSlice(params)
// 	reqTxID, err := common.NewHashFromStr(arrayParams[0].(string))
// 	if err != nil {
// 		return nil, NewRPCError(ErrRPCParse, err)
// 	}
// 	amount, status, err := (*rpcServer.config.Database).GetIssuingInfo(*reqTxID)
// 	if err != nil {
// 		return nil, NewRPCError(ErrRPCInternal, err)
// 	}
// 	result := map[string]interface{}{
// 		"Status": status,
// 		"Amount": amount,
// 	}
// 	return result, nil
// }

// // handleGetContractingStatus returns status accept/refund of a reserve contracting tx
// func (rpcServer RpcServer) handleGetContractingStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
// 	arrayParams := common.InterfaceSlice(params)
// 	reqTxID, err := common.NewHashFromStr(arrayParams[0].(string))
// 	if err != nil {
// 		return nil, NewRPCError(ErrRPCParse, err)
// 	}
// 	amount, redeem, status, err := (*rpcServer.config.Database).GetContractingInfo(*reqTxID)
// 	if err != nil {
// 		return nil, NewRPCError(ErrRPCInternal, err)
// 	}

// 	// Convert redeem asset units
// 	_, _, _, txReq, err := rpcServer.config.BlockChain.GetTransactionByHash(reqTxID)
// 	if err != nil {
// 		return nil, NewRPCError(ErrRPCInternal, err)
// 	}
// 	meta := txReq.GetMetadata().(*metadata.ContractingRequest)
// 	redeemStr := ""
// 	if common.IsUSDAsset(&meta.CurrencyType) {
// 		redeemStr = strconv.FormatUint(redeem, 10)
// 	} else {
// 		// Convert from milliether to wei
// 		redeemBig := big.NewInt(int64(redeem))
// 		redeemBig = redeemBig.Mul(redeemBig, big.NewInt(common.WeiToMilliEtherRatio))
// 		redeemStr = redeemBig.String()
// 	}
// 	result := map[string]interface{}{
// 		"Status": status,
// 		"Amount": amount,
// 		"Redeem": redeemStr,
// 	}
// 	return result, nil
// }
