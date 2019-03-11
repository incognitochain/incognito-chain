package rpcserver

import (
	"math/big"
	"strconv"

	"github.com/big0t/constant-chain/common"
	"github.com/big0t/constant-chain/metadata"
	"github.com/pkg/errors"
)

func (rpcServer RpcServer) handleCreateIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := metaConstructors[CreateAndSendIssuingRequest]
	return rpcServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawTxWithMetadata(params, closeChan)
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
	return rpcServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (rpcServer RpcServer) handleSendContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return rpcServer.sendRawTxWithMetadata(params, closeChan)
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
	amount, redeem, status, err := (*rpcServer.config.Database).GetContractingInfo(*reqTxID)
	if err != nil {
		return nil, NewRPCError(ErrRPCInternal, err)
	}

	// Convert redeem asset units
	_, _, _, txReq, err := rpcServer.config.BlockChain.GetTransactionByHash(reqTxID)
	if err != nil {
		return nil, NewRPCError(ErrRPCInternal, err)
	}
	meta := txReq.GetMetadata().(*metadata.ContractingRequest)
	redeemStr := ""
	if common.IsUSDAsset(&meta.CurrencyType) {
		redeemStr = strconv.FormatUint(redeem, 10)
	} else {
		// Convert from milliether to wei
		redeemBig := big.NewInt(int64(redeem))
		redeemBig = redeemBig.Mul(redeemBig, big.NewInt(common.WeiToMilliEtherRatio))
		redeemStr = redeemBig.String()
	}
	result := map[string]interface{}{
		"Status": status,
		"Amount": amount,
		"Redeem": redeemStr,
	}
	return result, nil
}

// handleConvertETHToDCBTokenAmount receives amount of ETH (in Wei) and returns number of DCB Tokens at current price
func (rpcServer RpcServer) handleConvertETHToDCBTokenAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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

	// Calculate #DCB tokens
	dcbTokenAmount := depositedAmount * oracle.ETH / oracle.DCBToken
	return dcbTokenAmount, nil
}

//  handleConvertCSTToETHAmount receives amount of CST and returns number of ETH (in Wei) at current price
func (rpcServer RpcServer) handleConvertCSTToETHAmount(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	amount := uint64(arrayParams[0].(float64))
	oracle := rpcServer.config.BlockChain.BestState.Beacon.StabilityInfo.Oracle
	milliEtherAmount := amount * oracle.Constant / oracle.ETH

	// Convert MilliEther to Wei
	etherAmount := big.NewInt(int64(milliEtherAmount))
	etherAmount = etherAmount.Mul(etherAmount, big.NewInt(common.WeiToMilliEtherRatio))
	return etherAmount.String(), nil
}
