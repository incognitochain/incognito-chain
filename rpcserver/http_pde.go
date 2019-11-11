package rpcserver

import (
	"encoding/json"
	"errors"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) handleCreateRawTxWithPRVContribution(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// get meta data from params
	data, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	pdeContributionPairID, ok := data["PDEContributionPairID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	contributorAddressStr, ok := data["ContributorAddressStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	contributedAmountData, ok := data["ContributedAmount"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	contributedAmount := uint64(contributedAmountData)
	tokenIDStr, ok := data["TokenIDStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	meta, _ := metadata.NewPDEContribution(
		pdeContributionPairID,
		contributorAddressStr,
		contributedAmount,
		tokenIDStr,
		metadata.PDEContributionMeta,
	)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta, *httpServer.config.Database)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPRVContribution(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPRVContribution(params, closeChan)
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

func (httpServer *HttpServer) handleCreateRawTxWithPTokenContribution(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	if len(arrayParams) >= 7 {
		hasPrivacyToken := int(arrayParams[6].(float64)) > 0
		if hasPrivacyToken {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("The privacy mode must be disabled"))
		}
	}
	tokenParamsRaw := arrayParams[4].(map[string]interface{})

	pdeContributionPairID, ok := tokenParamsRaw["PDEContributionPairID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	contributorAddressStr, ok := tokenParamsRaw["ContributorAddressStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	contributedAmountData, ok := tokenParamsRaw["ContributedAmount"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	contributedAmount := uint64(contributedAmountData)
	tokenIDStr := tokenParamsRaw["TokenIDStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	meta, _ := metadata.NewPDEContribution(
		pdeContributionPairID,
		contributorAddressStr,
		contributedAmount,
		tokenIDStr,
		metadata.PDEContributionMeta,
	)

	customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyCustomTokenTransaction(params, meta, *httpServer.config.Database)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	byteArrays, err2 := json.Marshal(customTokenTx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            customTokenTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPTokenContribution(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPTokenContribution(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	// sendResult, err1 := httpServer.handleSendRawCustomTokenTransaction(newParam, closeChan)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithPRVTradeReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// get meta data from params
	data, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	tokenIDToBuyStr, ok := data["TokenIDToBuyStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	tokenIDToSellStr, ok := data["TokenIDToSellStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	sellAmount := uint64(data["SellAmount"].(float64))
	traderAddressStr, ok := data["TraderAddressStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	minAcceptableAmountData, ok := data["MinAcceptableAmount"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	minAcceptableAmount := uint64(minAcceptableAmountData)
	tradingFeeData, ok := data["TradingFee"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	tradingFee := uint64(tradingFeeData)
	meta, _ := metadata.NewPDETradeRequest(
		tokenIDToBuyStr,
		tokenIDToSellStr,
		sellAmount,
		minAcceptableAmount,
		tradingFee,
		traderAddressStr,
		metadata.PDETradeRequestMeta,
	)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta, *httpServer.config.Database)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPRVTradeReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPRVTradeReq(params, closeChan)
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

func (httpServer *HttpServer) handleCreateRawTxWithPTokenTradeReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	if len(arrayParams) >= 7 {
		hasPrivacyToken := int(arrayParams[6].(float64)) > 0
		if hasPrivacyToken {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("The privacy mode must be disabled"))
		}
	}
	tokenParamsRaw := arrayParams[4].(map[string]interface{})

	tokenIDToBuyStr, ok := tokenParamsRaw["TokenIDToBuyStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}

	tokenIDToSellStr, ok := tokenParamsRaw["TokenIDToSellStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}

	sellAmountData, ok := tokenParamsRaw["SellAmount"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	sellAmount := uint64(sellAmountData)

	traderAddressStr, ok := tokenParamsRaw["TraderAddressStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}

	minAcceptableAmountData, ok := tokenParamsRaw["MinAcceptableAmount"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	minAcceptableAmount := uint64(minAcceptableAmountData)

	tradingFeeData, ok := tokenParamsRaw["TradingFee"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	tradingFee := uint64(tradingFeeData)

	meta, _ := metadata.NewPDETradeRequest(
		tokenIDToBuyStr,
		tokenIDToSellStr,
		sellAmount,
		minAcceptableAmount,
		tradingFee,
		traderAddressStr,
		metadata.PDETradeRequestMeta,
	)

	customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyCustomTokenTransaction(params, meta, *httpServer.config.Database)
	if rpcErr != nil {
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	byteArrays, err2 := json.Marshal(customTokenTx)
	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            customTokenTx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithPTokenTradeReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithPTokenTradeReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	// sendResult, err1 := httpServer.handleSendRawCustomTokenTransaction(newParam, closeChan)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithWithdrawalReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)

	// get meta data from params
	data, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}

	withdrawerAddressStr, ok := data["WithdrawerAddressStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}

	withdrawalToken1IDStr, ok := data["WithdrawalToken1IDStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}

	withdrawalToken2IDStr, ok := data["WithdrawalToken2IDStr"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}

	withdrawalShareAmtData, ok := data["WithdrawalShareAmt"].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	withdrawalShareAmt := uint64(withdrawalShareAmtData)

	meta, _ := metadata.NewPDEWithdrawalRequest(
		withdrawerAddressStr,
		withdrawalToken1IDStr,
		withdrawalToken2IDStr,
		withdrawalShareAmt,
		metadata.PDEWithdrawalRequestMeta,
	)

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta, *httpServer.config.Database)
	if err1 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	byteArrays, err2 := json.Marshal(tx)
	if err2 != nil {
		Logger.log.Error(err1)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithWithdrawalReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithWithdrawalReq(params, closeChan)
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

func (httpServer *HttpServer) handleGetPDEState(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	data := arrayParams[0].(map[string]interface{})
	beaconHeight := uint64(data["BeaconHeight"].(float64))
	pdeState, err := blockchain.InitCurrentPDEStateFromDB(httpServer.config.BlockChain.GetDatabase(), beaconHeight)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.GetPDEStateError, err)
	}
	return pdeState, nil
}
