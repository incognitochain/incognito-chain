package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/incognitochain/incognito-chain/wallet"
)

func (httpServer *HttpServer) handleCreateRawTxWithWithdrawRewardReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// get metadata from params
	metaParam, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	tokenIDStr, ok := metaParam["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token ID string is invalid"))
	}
	paymentAddStr, ok := metaParam["PaymentAddress"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("payment address string is invalid"))
	}
	txVersion, ok := metaParam["TxVersion"].(float64)
	if !ok {
		txVersion = 0
	}
	meta, err := metadata.NewWithDrawRewardRequest(
		tokenIDStr,
		paymentAddStr,
		1,
		metadata.WithDrawRewardRequestMeta,
	)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	if txVersion == 1 {
		meta.PaymentAddress.OTAPublic = nil
	}

	param := map[string]interface{}{}
	param["PaymentAddress"] = paymentAddStr
	param["TokenID"] = tokenIDStr
	param["Version"] = common.SALARY_VER_FIX_HASH
	arrayParams[4] = interface{}(param)
	return httpServer.createRawTxWithMetadata(
		arrayParams,
		closeChan,
		metadata.NewWithDrawRewardRequestFromRPC,
	)
}

func (httpServer *HttpServer) handleCreateAndSendWithDrawTransaction(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithWithdrawRewardReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	result := jsonresult.NewCreateTransactionResult(nil, sendResult.(jsonresult.CreateTransactionResult).TxID, nil, sendResult.(jsonresult.CreateTransactionResult).ShardID)
	return result, nil
}

// handleGetRewardAmount - Get the reward amount of a payment address with all existed token
func (httpServer *HttpServer) handleGetRewardAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}

	paymentAddress, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("payment address is invalid"))
	}
	mode := 0 //0: committee reward, 1: delegate reward, 2: total reward
	if len(arrayParams) == 2 {
		modeInput, ok := arrayParams[1].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("mode is invalid"))
		}
		mode = int(modeInput)
	}
	switch mode {
	default:
		rewardAmount, err := httpServer.blockService.GetRewardAmount(paymentAddress)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.GetRewardAmountError, err)
		}
		return rewardAmount, nil
	case 1:
		keyWallet, err := wallet.Base58CheckDeserialize(paymentAddress)
		if err != nil {
			panic(1)
		}
		receiverAddr := keyWallet.KeySet.PaymentAddress
		rewardAmount, err := httpServer.GetBlockchain().GetDelegationRewardAmount(httpServer.GetBlockchain().GetBeaconBestState().GetBeaconConsensusStateDB(), receiverAddr.Pk)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.GetRewardAmountError, err)
		}
		return rewardAmount, nil
	}
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
