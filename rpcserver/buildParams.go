package rpcserver

import (
	"encoding/json"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
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

func (rpcServer *RpcServer) buildParamsSealLv2VoteProposal(params interface{}) (interface{}, *RPCError) {
	params = setBuildRawBurnTransactionParams(params, FeeVoteProposal)
	arrayParams := common.InterfaceSlice(params)
	data := arrayParams[len(arrayParams)-1].(map[string]interface{})
	newData := make(map[string]interface{})

	lv3TxID, err := common.NewHashFromStr(data["Lv3TxID"].(string))
	firstPrivateKey := []byte(data["FirstPrivateKey"].(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	_, _, _, lv3Tx, err := rpcServer.config.BlockChain.GetTransactionByHash(lv3TxID)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	SealLv3Data, err := GetSealLv3Data(lv3Tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	paymentAddresses := GetLockerPaymentAddresses(lv3Tx)
	SealLv2Data := common.Decrypt(SealLv3Data, firstPrivateKey)

	newData["SealLv2Data"] = SealLv2Data
	newData["PaymentAddresses"] = paymentAddresses
	newData["Lv3TxID"] = *lv3TxID
	arrayParams[len(arrayParams)-1] = newData
	return arrayParams, nil
}

func (rpcServer *RpcServer) buildParamsSealLv3VoteProposal(params interface{}) (interface{}, *RPCError) {
	params = setBuildRawBurnTransactionParams(params, FeeVoteProposal)
	return params, nil
}

func (rpcServer RpcServer) buildParamsSealLv1VoteProposal(
	params interface{},
) (interface{}, *RPCError) {
	params = setBuildRawBurnTransactionParams(params, FeeVoteProposal)
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)
	data := arrayParams[NParams-1].(map[string]interface{})
	newData := make(map[string]interface{})

	boardType := metadata.NewBoardTypeFromString(data["BoardType"].(string))

	secondPrivateKey := []byte(data["SecondPrivateKey"].(string))

	lv3TxID, err1 := common.NewHashFromStr(data["Lv3TxID"].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	lv2TxID, err1 := common.NewHashFromStr(data["Lv2TxID"].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	_, _, _, lv2tx, _ := rpcServer.config.BlockChain.GetTransactionByHash(lv2TxID)
	SealLv2Data, err1 := GetSealLv2Data(lv2tx)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	_, _, _, lv3tx, _ := rpcServer.config.BlockChain.GetTransactionByHash(lv3TxID)
	paymentAddresses := GetLockerPaymentAddresses(lv3tx)
	sealLv1Data := common.Decrypt(SealLv2Data, secondPrivateKey)

	newData["BoardType"] = boardType
	newData["SealLv1Data"] = sealLv1Data
	newData["PaymentAddresses"] = paymentAddresses
	newData["Lv2TxID"] = lv2TxID
	newData["Lv3TxID"] = lv3TxID
	arrayParams[NParams-1] = newData

	return arrayParams, nil
}

func (rpcServer RpcServer) buildParamsNormalVoteProposalFromOwner(
	params interface{},
) (interface{}, *RPCError) {
	params = setBuildRawBurnTransactionParams(params, FeeVoteProposal)
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)
	data := arrayParams[len(arrayParams)-1].(map[string]interface{})
	newData := make(map[string]interface{})

	boardType := metadata.NewBoardTypeFromString(data["BoardType"].(string))

	lv3TxID, err1 := common.NewHashFromStr(data["Lv3TxID"].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	_, _, _, lv3tx, _ := rpcServer.config.BlockChain.GetTransactionByHash(lv3TxID)
	paymentAddresses := GetLockerPaymentAddresses(lv3tx)

	newData["BoardType"] = boardType
	newData["VoteProposalData"] = data["VoteProposalData"]
	newData["PaymentAddresses"] = paymentAddresses
	newData["Lv3TxID"] = lv3TxID

	arrayParams[NParams-1] = newData
	return arrayParams, nil
}

func (rpcServer RpcServer) buildParamsNormalVoteProposalFromSealer(
	params interface{},
) (interface{}, *RPCError) {
	params = setBuildRawBurnTransactionParams(params, FeeVoteProposal)
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)
	data := arrayParams[len(arrayParams)-1].(map[string]interface{})
	newData := make(map[string]interface{})

	boardType := metadata.NewBoardTypeFromString(data["BoardType"].(string))

	lv3TxID, err1 := common.NewHashFromStr(data["Lv3TxID"].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	_, _, _, lv3tx, _ := rpcServer.config.BlockChain.GetTransactionByHash(lv3TxID)
	paymentAddresses := GetLockerPaymentAddresses(lv3tx)

	lv1TxID, err1 := common.NewHashFromStr(data["Lv1TxID"].(string))
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}
	_, _, _, lv1tx, _ := rpcServer.config.BlockChain.GetTransactionByHash(lv1TxID)
	SealLv1Data, err1 := GetSealLv1Data(lv1tx)
	if err1 != nil {
		return nil, NewRPCError(ErrUnexpected, err1)
	}

	thirdPrivateKey := []byte(data["ThirdPrivateKey"].(string))

	normalVoteProposalData := common.Decrypt(SealLv1Data, thirdPrivateKey)
	voteProposalDataTemp := metadata.NewVoteProposalDataFromBytes(normalVoteProposalData)
	voteProposalDataByte, err := json.Marshal(voteProposalDataTemp)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	newData["BoardType"] = boardType
	newData["VoteProposalData"] = voteProposalDataByte
	newData["PaymentAddresses"] = paymentAddresses
	newData["Lv1TxID"] = lv1TxID
	newData["Lv3TxID"] = lv3TxID

	arrayParams[NParams-1] = newData
	return arrayParams, nil
}
