package rpcserver

import (
	"encoding/json"

	"github.com/constant-money/constant-chain/database/lvdb"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
)

func (rpcServer RpcServer) handleGetBridgeTokensAmounts(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	db := rpcServer.config.BlockChain.GetDatabase()
	tokensAmtsBytesArr, dbErr := db.GetBridgeTokensAmounts()
	if dbErr != nil {
		return nil, NewRPCError(ErrUnexpected, dbErr)
	}

	result := &jsonresult.GetBridgeTokensAmounts{
		BridgeTokensAmounts: make(map[string]jsonresult.GetBridgeTokensAmount),
	}
	for _, tokensAmtsBytes := range tokensAmtsBytesArr {
		var tokenWithAmount lvdb.TokenWithAmount
		err := json.Unmarshal(tokensAmtsBytes, &tokenWithAmount)
		if err != nil {
			return nil, NewRPCError(ErrUnexpected, dbErr)
		}
		tokenID := tokenWithAmount.TokenID
		result.BridgeTokensAmounts[tokenID.String()] = jsonresult.GetBridgeTokensAmount{
			TokenID: tokenWithAmount.TokenID,
			Amount:  tokenWithAmount.Amount,
		}
	}
	return result, nil
}

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
