package rpcserver

// handleGetDCBParams - get dcb params
func (self RpcServer) handleGetDCBParams(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	dcbParam := self.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution.DCBParams
	return dcbParam, nil
}

// handleGetDCBConstitution - get dcb constitution
func (self RpcServer) handleGetDCBConstitution(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	constitution := self.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution
	return constitution, nil
}

// handleGetListDCBBoard - return list payment address of DCB board
func (self RpcServer) handleGetListDCBBoard(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.DCBGovernor.DCBBoardPubKeys, nil
}
