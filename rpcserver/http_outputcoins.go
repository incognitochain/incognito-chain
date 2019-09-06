package rpcserver

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

//handleListUnspentOutputCoins - use private key to get all tx which contains output coin of account
// by private key, it return full tx outputcoin with amount and receiver address in txs
//component:
//Parameter #1—the minimum number of confirmations an output must have
//Parameter #2—the maximum number of confirmations an output may have
//Parameter #3—the list priv-key which be used to view utxo
//
func (httpServer *HttpServer) handleListUnspentOutputCoins(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleListUnspentOutputCoins params: %+v", params)

	// get component
	paramsArray := common.InterfaceSlice(params)
	var min int
	var max int
	if len(paramsArray) > 0 && paramsArray[0] != nil {
		min = int(paramsArray[0].(float64))
	}
	if len(paramsArray) > 1 && paramsArray[1] != nil {
		max = int(paramsArray[1].(float64))
	}
	_ = min
	_ = max
	listKeyParams := common.InterfaceSlice(paramsArray[2])
	result, err := httpServer.outputCoinService.ListUnspentOutputCoinsByKey(listKeyParams)
	if err != nil {
		return nil, err
	}

	Logger.log.Debugf("handleListUnspentOutputCoins result: %+v", result)
	return result, nil
}

//handleListOutputCoins - use readonly key to get all tx which contains output coin of account
// by private key, it return full tx outputcoin with amount and receiver address in txs
//component:
//Parameter #1—the minimum number of confirmations an output must have
//Parameter #2—the maximum number of confirmations an output may have
//Parameter #3—the list paymentaddress-readonlykey which be used to view list outputcoin
//Parameter #4 - optional - token id - default prv coin
func (httpServer *HttpServer) handleListOutputCoins(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleListOutputCoins params: %+v", params)

	// get component
	paramsArray := common.InterfaceSlice(params)
	if len(paramsArray) < 1 {
		Logger.log.Debugf("handleListOutputCoins result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("invalid list Key component"))
	}
	minTemp, ok := paramsArray[0].(float64)
	if !ok {
		Logger.log.Debugf("handleListOutputCoins result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("invalid list Key component"))
	}
	min := int(minTemp)
	maxTemp, ok := paramsArray[1].(float64)
	if !ok {
		Logger.log.Debugf("handleListOutputCoins result: %+v", nil)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("invalid list Key component"))
	}
	max := int(maxTemp)
	_ = min
	_ = max
	//#3: list key component
	listKeyParams := common.InterfaceSlice(paramsArray[2])

	//#4: optional token type - default prv coin
	tokenID := &common.Hash{}
	err := tokenID.SetBytes(common.PRVCoinID[:])
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TokenIsInvalidError, err)
	}
	if len(paramsArray) > 3 {
		var err1 error
		tokenID, err1 = common.Hash{}.NewHashFromStr(paramsArray[3].(string))
		if err1 != nil {
			Logger.log.Debugf("handleListOutputCoins result: %+v, err: %+v", nil, err1)
			return nil, rpcservice.NewRPCError(rpcservice.ListCustomTokenNotFoundError, err1)
		}
	}
	result, err1 := httpServer.outputCoinService.ListOutputCoinsByKey(listKeyParams, *tokenID)
	if err != nil {
		return nil, err1
	}
	Logger.log.Debugf("handleListOutputCoins result: %+v", result)
	return result, nil
}
