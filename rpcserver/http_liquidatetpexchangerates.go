package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) handleGetLiquidationTpExchangeRates(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// get meta data from params
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata param is invalid"))
	}

	beaconHeight, ok := data["BeaconHeight"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata BeaconHeight is invalid"))
	}

	_, err := httpServer.config.BlockChain.GetBeaconBlockByHeight(uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetTpExchangeRatesLiquidationError, err)
	}

	custodianAddress, ok := data["CustodianAddress"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata CustodianAddress is invalid"))
	}

	pTokenID, ok := data["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata TokenID is invalid"))
	}

	if !common.IsPortalExchangeRateToken(pTokenID) {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata TokenID is not support"))
	}

	result, err := httpServer.portal.GetLiquidateTpExchangeRates(uint64(beaconHeight), custodianAddress, pTokenID, httpServer.blockService, *httpServer.config.Database)

	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetTpExchangeRatesLiquidationError, err)
	}

	return result, nil
}

func (httpServer *HttpServer) handleGetLiquidationExchangeRates(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// get meta data from params
	data, ok := arrayParams[0].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata param is invalid"))
	}

	beaconHeight, ok := data["BeaconHeight"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata BeaconHeight is invalid"))
	}

	_, err := httpServer.config.BlockChain.GetBeaconBlockByHeight(uint64(beaconHeight))
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetExchangeRatesLiquidationError, err)
	}

	pTokenID, ok := data["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata TokenID is invalid"))
	}

	if !common.IsPortalExchangeRateToken(pTokenID) {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata TokenID is not support"))
	}

	result, err := httpServer.portal.GetLiquidateExchangeRates(uint64(beaconHeight), pTokenID, httpServer.blockService, *httpServer.config.Database)

	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetExchangeRatesLiquidationError, err)
	}

	return result, nil
}
