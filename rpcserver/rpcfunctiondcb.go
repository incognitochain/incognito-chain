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

func (rpcServer RpcServer) handleCreateRawVoteDCBBoardTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	params = setBuildRawBurnTransactionParams(params, FeeVote)
	return rpcServer.createRawCustomTokenTxWithMetadata(params, closeChan, metadata.NewVoteDCBBoardMetadataFromRPC)
}

func (rpcServer RpcServer) handleCreateAndSendVoteDCBBoardTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawVoteDCBBoardTransaction,
		RpcServer.handleSendRawCustomTokenTransaction,
	)
}

func (rpcServer RpcServer) handleCreateRawSubmitDCBProposalTransaction(
	params interface{},
	closeChan <-chan struct{},
) (interface{}, *RPCError) {
	params, err := rpcServer.buildParamsSubmitDCBProposal(params)
	if err != nil {
		return nil, err
	}
	return rpcServer.createRawTxWithMetadata(
		params,
		closeChan,
		metadata.NewSubmitDCBProposalMetadataFromRPC,
	)
}

func (rpcServer RpcServer) handleCreateAndSendSubmitDCBProposalTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateRawSubmitDCBProposalTransaction,
		RpcServer.handleSendRawTransaction,
	)
}
