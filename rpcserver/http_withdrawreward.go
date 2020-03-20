package rpcserver

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
	"github.com/pkg/errors"
)

func (httpServer *HttpServer) handleCreateRawWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}
	arrayParams[1] = nil

	privateKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}
	keyWallet, err := wallet.Base58CheckDeserialize(privateKeyParam)
	if err != nil {
		return []byte{}, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New(fmt.Sprintf("Wrong privatekey %+v", err)))
	}
	keyWallet.KeySet.InitFromPrivateKeyByte(keyWallet.KeySet.PrivateKey)

	metaParam, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	tokenIDParam, ok := metaParam["TokenID"]
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token ID is invalid"))
	}

	param := map[string]interface{}{}
	param["PaymentAddress"] = keyWallet.Base58CheckSerialize(1)
	param["TokenID"] = tokenIDParam
	param["Version"] = 1
	if version, ok := metaParam["Version"]; ok {
		param["Version"] = version
	}
	arrayParams[4] = interface{}(param)
	return httpServer.createRawTxWithMetadata(
		arrayParams,
		closeChan,
		metadata.NewWithDrawRewardRequestFromRPC,
	)
}

func (httpServer *HttpServer) handleCreateAndSendWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		(*HttpServer).handleCreateRawWithDrawTransaction,
		(*HttpServer).handleSendRawTransaction,
	)
}

// handleGetRewardAmount - Get the reward amount of a payment address with all existed token
func (httpServer *HttpServer) handleGetRewardAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}

	paymentAddress, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("payment address is invalid"))
	}
	rewardAmount, err := httpServer.blockService.GetRewardAmount(paymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetRewardAmountError, err)
	}
	return rewardAmount, nil
}

// handleGetRewardAmount - Get the reward amount of a payment address with all existed token
func (httpServer *HttpServer) handleGetRewardAmountByPublicKey(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) != 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}

	paymentAddress, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("payment address is invalid"))
	}
	rewardAmount, err := httpServer.blockService.GetRewardAmountByPublicKey(paymentAddress)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetRewardAmountError, err)
	}
	return rewardAmount, nil
}

// handleListRewardAmount - Get the reward amount of all committee with all existed token
func (httpServer *HttpServer) handleListRewardAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	result, err := httpServer.blockService.ListRewardAmount()
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.ListCommitteeRewardError, err)
	}
	return result, nil
}
