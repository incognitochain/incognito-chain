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

	lv3TxID, err1 := common.NewHashFromStr(data["Lv3TxID"].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	lv1TxID, err1 := common.NewHashFromStr(data["Lv1TxID"].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	newData["BoardType"] = data["BoardType"]
	newData["Lv1TxID"] = *lv1TxID
	newData["Lv3TxID"] = *lv3TxID

	arrayParams[NParams-1] = newData
	return arrayParams, nil
}
