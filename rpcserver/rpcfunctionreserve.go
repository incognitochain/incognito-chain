package rpcserver

import (
	"math/big"

	"github.com/ninjadotorg/constant/common"
	"github.com/pkg/errors"
)

func (rpcServer RpcServer) handleCreateIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendIssuingRequest]
	return rpcServer.createRawCustomTokenTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawCustomTokenTxWithMetadata(params, closeChan)
}

// handleCreateAndSendIssuingRequest for user to buy Constant (using USD) or BANK token (using USD/ETH) from DCB
func (rpcServer RpcServer) handleCreateAndSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateIssuingRequest,
		RpcServer.handleSendIssuingRequest,
	)
}

func (rpcServer RpcServer) handleCreateContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendContractingRequest]
	return rpcServer.createRawCustomTokenTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawCustomTokenTxWithMetadata(params, closeChan)
}

// handleCreateAndSendContractingRequest for user to sell Constant and receive either USD or ETH
func (rpcServer RpcServer) handleCreateAndSendContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		RpcServer.handleCreateContractingRequest,
		RpcServer.handleSendContractingRequest,
	)
}

// handleGetIssuingStatus returns status accept/refund of a reserve issuing tx
func (rpcServer RpcServer) handleGetIssuingStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	reqTxID, err := common.NewHashFromStr(arrayParams[0].(string))
	if err != nil {
		return nil, NewRPCError(ErrRPCParse, err)
	}
	amount, status, err := (*rpcServer.config.Database).GetIssuingInfo(*reqTxID)
	if err != nil {
		return nil, NewRPCError(ErrRPCInternal, err)
	}
	result := map[string]interface{}{
		"Status": status,
		"Amount": amount,
	}
	return result, nil
}

// handleGetContractingStatus returns status accept/refund of a reserve contracting tx
func (rpcServer RpcServer) handleGetContractingStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	reqTxID, err := common.NewHashFromStr(arrayParams[0].(string))
	if err != nil {
		return nil, NewRPCError(ErrRPCParse, err)
	}
	amount, status, err := (*rpcServer.config.Database).GetContractingInfo(*reqTxID)
	if err != nil {
		return nil, NewRPCError(ErrRPCInternal, err)
	}
	result := map[string]interface{}{
		"Status": status,
		"Amount": amount,
	}
	return result, nil
}

// handleConvertToDCBTokenAmount receives amount of ETH (in Wei) and returns number of DCB Tokens for current price
func (rpcServer RpcServer) handleConvertToDCBTokenAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	amountStr := arrayParams[0].(string)
	// Convert amount to MilliEther
	amount := big.NewInt(0)
	amount, ok := amount.SetString(amountStr, 10)
	if !ok {
		return nil, NewRPCError(ErrRPCParse, errors.Errorf("Error parsing amount: %s", amountStr))
	}
	oracle := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.Oracle
	amount = amount.Quo(amount, big.NewInt(common.WeiToMilliEtherRatio))
	if !amount.IsUint64() {
		return nil, NewRPCError(ErrRPCParse, errors.New("Amount invalid"))
	}
	depositedAmount := amount.Uint64()
	dcbTokenAmount := depositedAmount * oracle.ETH / oracle.DCBToken
	return dcbTokenAmount, nil
}
