package rpcserver

import (
	"encoding/hex"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
)

// handleGetDCBParams - get dcb params
func (rpcServer RpcServer) handleGetDCBParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constitution := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.DCBConstitution
	dcbParam := constitution.DCBParams
	results := make(map[string]interface{})
	results["DCBParams"] = dcbParam
	results["ExecuteDuration"] = constitution.ExecuteDuration
	results["Explanation"] = constitution.Explanation
	return results, nil
}

// handleGetDCBConstitution - get dcb constitution
func (rpcServer RpcServer) handleGetDCBConstitution(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constitution := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.DCBConstitution
	return constitution, nil
}

// handleGetListDCBBoard - return list payment address of DCB board
func (rpcServer RpcServer) handleGetListDCBBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	res := ListPaymentAddressToListString(rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress)
	return res, nil
}

func (rpcServer RpcServer) handleAppendListDCBBoard(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	senderKey := arrayParams[0].(string)
	paymentAddress, _ := metadata.GetPaymentAddressFromSenderKeyParams(senderKey)
	rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress =
		append(rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress, *paymentAddress)
	res := ListPaymentAddressToListString(rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.DCBGovernor.BoardPaymentAddress)
	return res, nil
}

func ListPaymentAddressToListString(addresses []privacy.PaymentAddress) []string {
	res := make([]string, 0)
	for _, i := range addresses {
		pk := hex.EncodeToString(i.Pk)
		res = append(res, pk)
	}
	return res
}

func getAmountVote(receiversPaymentAddressParam map[string]interface{}) int64 {
	sumAmount := int64(0)
	for paymentAddressStr, amount := range receiversPaymentAddressParam {
		if paymentAddressStr == common.BurningAddress {
			sumAmount += int64(amount.(float64))
		}
	}
	return sumAmount
}

func (rpcServer RpcServer) handleCreateRawVoteDCBBoardCustomTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	return rpcServer.createRawCustomTokenTxWithMetadata(params, closeChan, metadata.NewVoteDCBBoardMetadataFromRPC)
}

func (rpcServer RpcServer) handleCreateAndSendVoteDCBBoardTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawVoteDCBBoardCustomTransaction,
		RpcServer.handleSendRawCustomTokenTransaction,
	)
}

func setBuildRawBurnTransactionParams(params interface{}, fee float64) interface{} {
	arrayParams := common.InterfaceSlice(params)
	x := make(map[string]interface{})
	x[common.BurningAddress] = fee
	arrayParams[1] = x
	return arrayParams
}

func (rpcServer RpcServer) handleCreateRawSubmitDCBProposalTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	NParams := len(arrayParams)

	metaParams := arrayParams[NParams-1].(map[string]interface{})
	tmp, err := rpcServer.GetPaymentAddressFromPrivateKeyParams(arrayParams[0].(string))
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	metaParams["PaymentAddress"] = tmp
	arrayParams[NParams-1] = metaParams

	params = setBuildRawBurnTransactionParams(arrayParams, FeeSubmitProposal)
	return rpcServer.createRawTxWithMetadata(params, closeChan, metadata.NewSubmitDCBProposalMetadataFromRPC)
}

func (rpcServer RpcServer) handleCreateAndSendSubmitDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawSubmitDCBProposalTransaction,
		RpcServer.handleSendRawTransaction,
	)
}
