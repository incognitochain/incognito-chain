package rpcserver

func (rpcServer RpcServer) handleCreateRawTradeActivation(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendTradeActivation]
	return rpcServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendRawTradeActivation(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawTxWithMetadata(params, closeChan)
}

func (rpcServer RpcServer) handleCreateAndSendTradeActivation(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawTradeActivation,
		RpcServer.handleSendRawTradeActivation,
	)
}
