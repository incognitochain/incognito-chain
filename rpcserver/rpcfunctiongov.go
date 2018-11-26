package rpcserver

import "github.com/ninjadotorg/constant/rpcserver/jsonresult"

func (self RpcServer) handleGetBondTypes(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	tempRes1 := jsonresult.GetBondTypeResult{
		BondID:         "12345abc",
		StartSellingAt: 123,
		Maturity:       300,
		BuyBackPrice:   100000,
	}
	tempRes2 := jsonresult.GetBondTypeResult{
		BondID:         "12345xyz",
		StartSellingAt: 95,
		Maturity:       200,
		BuyBackPrice:   200000,
	}
	return []jsonresult.GetBondTypeResult{tempRes1, tempRes2}, nil
}
