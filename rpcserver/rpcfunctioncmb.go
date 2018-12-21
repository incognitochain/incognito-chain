package rpcserver

func (self RpcServer) handleGetListCBBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.CMBGovernor.CMBBoardPubKeys, nil
}
