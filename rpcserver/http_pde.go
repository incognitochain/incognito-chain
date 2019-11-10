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
	pdeContributionPairID := data["PDEContributionPairID"].(string)
	contributorAddressStr := data["ContributorAddressStr"].(string)
	contributedAmount := uint64(data["ContributedAmount"].(float64))
	tokenIDStr := data["TokenIDStr"].(string)
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

	pdeContributionPairID := tokenParamsRaw["PDEContributionPairID"].(string)
	contributorAddressStr := tokenParamsRaw["ContributorAddressStr"].(string)
	contributedAmount := uint64(tokenParamsRaw["ContributedAmount"].(float64))
	tokenIDStr := tokenParamsRaw["TokenIDStr"].(string)
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
	tokenIDToBuyStr := data["TokenIDToBuyStr"].(string)
	tokenIDToSellStr := data["TokenIDToSellStr"].(string)
	sellAmount := uint64(data["SellAmount"].(float64))
	traderAddressStr := data["TraderAddressStr"].(string)
	minAcceptableAmount := uint64(data["MinAcceptableAmount"].(float64))
	tradingFee := uint64(data["TradingFee"].(float64))
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
	tokenIDToBuyStr := tokenParamsRaw["TokenIDToBuyStr"].(string)
	tokenIDToSellStr := tokenParamsRaw["TokenIDToSellStr"].(string)
	sellAmount := uint64(tokenParamsRaw["SellAmount"].(float64))
	traderAddressStr := tokenParamsRaw["TraderAddressStr"].(string)
	minAcceptableAmount := uint64(tokenParamsRaw["MinAcceptableAmount"].(float64))
	tradingFee := uint64(tokenParamsRaw["TradingFee"].(float64))
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
	withdrawerAddressStr := data["WithdrawerAddressStr"].(string)
	withdrawalToken1IDStr := data["WithdrawalToken1IDStr"].(string)
	withdrawalToken2IDStr := data["WithdrawalToken2IDStr"].(string)
	withdrawalShareAmt := uint64(data["WithdrawalShareAmt"].(float64))
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
