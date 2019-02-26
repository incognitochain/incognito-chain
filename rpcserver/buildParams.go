package rpcserver

import "github.com/ninjadotorg/constant/common"

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
