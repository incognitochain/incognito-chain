package rpcserver

import "github.com/ninjadotorg/constant/rpcserver/jsonresult"

func (self RpcServer) handleGetBondTypes(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	tempRes1 := jsonresult.GetBondTypeResult{
		BondID:         []byte("12345abc"),
		StartSellingAt: 123,
		Maturity:       300,
		BuyBackPrice:   100000,
	}
	tempRes2 := jsonresult.GetBondTypeResult{
		BondID:         []byte("12345xyz"),
		StartSellingAt: 95,
		Maturity:       200,
		BuyBackPrice:   200000,
	}
	return []jsonresult.GetBondTypeResult{tempRes1, tempRes2}, nil
}

func (self RpcServer) handleGetGOVParams(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	govParam := self.config.BlockChain.BestState[0].BestBlock.Header.GOVConstitution.GOVParams
	return govParam, nil
}

func (self RpcServer) handleGetGOVConstitution(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	constitution := self.config.BlockChain.BestState[0].BestBlock.Header.GOVConstitution
	return constitution, nil
}

func (self RpcServer) handleGetListGOVBoard(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.GOVGovernor.GOVBoardPubKeys, nil
}
