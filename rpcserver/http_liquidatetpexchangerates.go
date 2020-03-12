package rpcserver

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"errors"
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

	tokenSymbol, ok := data["TokenSymbol"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata TokenSymbol is invalid"))
	}

	tokenSymbolExist, _ := common.SliceExists(metadata.PortalSupportedExchangeRatesSymbols, tokenSymbol)
	if !tokenSymbolExist {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata TokenSymbol is not support"))
	}

	result, err := httpServer.portal.GetLiquidateTpExchangeRates(uint64(beaconHeight), custodianAddress, tokenSymbol, httpServer.blockService, *httpServer.config.Database)

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

	tokenSymbol, ok := data["TokenSymbol"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata TokenSymbol is invalid"))
	}

	tokenSymbolExist, _ := common.SliceExists(metadata.PortalSupportedExchangeRatesSymbols, tokenSymbol)
	if !tokenSymbolExist {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata TokenSymbol is not support"))
	}

	result, err := httpServer.portal.GetLiquidateExchangeRates(uint64(beaconHeight), tokenSymbol, httpServer.blockService, *httpServer.config.Database)

	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetExchangeRatesLiquidationError, err)
	}

	return result, nil
}
