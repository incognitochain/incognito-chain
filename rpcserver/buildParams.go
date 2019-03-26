package rpcserver

import (
	"github.com/constant-money/constant-chain/common"
)

func (rpcServer *RpcServer) buildParamsSubmitDCBProposal(params interface{}) (interface{}, *RPCError) {
	params = setBuildRawBurnTransactionParams(params, FeeSubmitProposal)
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)

	data := arrayParams[NParams-1].(map[string]interface{})
	tmp, err := rpcServer.GetPaymentAddressFromPrivateKeyParams(arrayParams[0].(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	data["PaymentAddress"] = tmp
	arrayParams[NParams-1] = data

	return params, nil
}

func (rpcServer *RpcServer) buildParamsSubmitGOVProposal(params interface{}) (interface{}, *RPCError) {
	params = setBuildRawBurnTransactionParams(params, FeeSubmitProposal)
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)

	data := arrayParams[NParams-1].(map[string]interface{})
	tmp, err := rpcServer.GetPaymentAddressFromPrivateKeyParams(arrayParams[0].(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	data["PaymentAddress"] = tmp
	arrayParams[NParams-1] = data

	return params, nil
}

func (rpcServer RpcServer) buildParamsVoteProposal(
	params interface{},
) (interface{}, *RPCError) {
	params = setBuildRawBurnTransactionParams(params, FeeVote)
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)
	data := arrayParams[len(arrayParams)-1].(map[string]interface{})
	newData := make(map[string]interface{})

	newData["BoardType"] = data["BoardType"]

	arrayParams[NParams-1] = newData
	return arrayParams, nil
}
