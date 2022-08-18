package rpcserver

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/pkg/errors"
)

func (httpServer *HttpServer) handleCreateIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	constructor := metaConstructors[createAndSendIssuingRequest]
	return httpServer.createRawTxWithMetadata(params, closeChan, constructor)
}

func (httpServer *HttpServer) handleCreateIssuingRequestV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateIssuingRequest(params, closeChan)
}

func (httpServer *HttpServer) handleSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleSendRawTransaction(params, closeChan)
}

// handleCreateAndSendIssuingRequest for user to buy Constant (using USD) or BANK token (using USD/ETH) from DCB
func (httpServer *HttpServer) handleCreateAndSendIssuingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.createAndSendTxWithMetadata(
		params,
		closeChan,
		(*HttpServer).handleCreateIssuingRequest,
		(*HttpServer).handleSendIssuingRequest,
	)
}

// handleCreateAndSendIssuingRequest for user to buy Constant (using USD) or BANK token (using USD/ETH) from DCB
func (httpServer *HttpServer) handleCreateAndSendIssuingRequestV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateAndSendIssuingRequest(params, closeChan)
}

func (httpServer *HttpServer) handleCreateRawTxWithContractingReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	// check privacy mode param
	if len(arrayParams) > 6 {
		privacyTemp, ok := arrayParams[6].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("The privacy mode must be valid"))
		}
		hasPrivacyToken := int(privacyTemp) > 0
		if hasPrivacyToken {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("The privacy mode must be disabled"))
		}
	}

	senderPrivateKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token param is invalid"))
	}

	tokenReceivers, ok := tokenParamsRaw["TokenReceivers"].(interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token receivers is invalid"))
	}

	tokenID, ok := tokenParamsRaw["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token ID is invalid"))
	}

	var txVersion int8
	tmpVersionParam, ok := tokenParamsRaw["TxVersion"]
	if !ok {
		txVersion = 2
	} else {
		tmpVersion, ok := tmpVersionParam.(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("txVersion must be a float64"))
		}
		txVersion = int8(tmpVersion)
	}

	meta, err := rpcservice.NewContractingRequestMetadata(senderPrivateKeyParam, tokenReceivers, tokenID)
	if err != nil {
		return nil, err
	}
	if txVersion == 1 {
		meta.BurnerAddress.OTAPublic = nil
	}

	customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyCustomTokenTransaction(params, meta)
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

func (httpServer *HttpServer) handleCreateAndSendContractingRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithContractingReq(params, closeChan)
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

func (httpServer *HttpServer) handleCreateAndSendContractingRequestV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateAndSendContractingRequest(params, closeChan)
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningRequestMetaV2,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningReq(params, closeChan)
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

func (httpServer *HttpServer) handleCreateAndSendBurningRequestV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateAndSendBurningRequestV2(params, closeChan)
}

func (httpServer *HttpServer) handleCreateRawTxWithIssuingEVMReq(params interface{}, closeChan <-chan struct{}, metatype int) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	// get meta data from params
	data, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	meta, err := metadataBridge.NewIssuingEVMRequestFromMap(data, 0, metatype)
	if err != nil {
		rpcErr := rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
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

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingETHReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingETHRequestMeta)
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

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingETHReqV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateAndSendTxWithIssuingETHReq(params, closeChan)
}

func (httpServer *HttpServer) handleCheckETHHashIssued(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	data := arrayParams[0].(map[string]interface{})

	issued, err := httpServer.blockService.CheckETHHashIssued(data)
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return issued, nil
}

func (httpServer *HttpServer) handleCheckBSCHashIssued(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	data := arrayParams[0].(map[string]interface{})

	issued, err := httpServer.blockService.CheckBSCHashIssued(data)
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return issued, nil
}

func (httpServer *HttpServer) handleCheckPrvPeggingHashIssued(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	data := arrayParams[0].(map[string]interface{})

	issued, err := httpServer.blockService.CheckPRVPeggingHashIssued(data)
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return issued, nil
}

func (httpServer *HttpServer) handleGetAllBridgeTokens(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	allBridgeTokens, err := httpServer.blockService.GetAllBridgeTokens()
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return allBridgeTokens, nil
}

func (httpServer *HttpServer) handleGetAllBridgeTokensByHeight(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	if len(arrayParams) == 0 {
		return httpServer.handleGetAllBridgeTokens(params, closeChan)
	}
	height, ok := arrayParams[0].(float64)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse height param, must be a float"))
	}

	allBridgeTokens, err := httpServer.blockService.GetAllBridgeTokensByHeight(uint64(height))
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return allBridgeTokens, nil
}

func (httpServer *HttpServer) handleGetETHHeaderByHash(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	ethBlockHash := arrayParams[0].(string)

	ethHeader, err := rpcservice.GetETHHeaderByHash(ethBlockHash)
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return ethHeader, nil
}

func (httpServer *HttpServer) handleGetBridgeReqWithStatus(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}
	data := arrayParams[0].(map[string]interface{})

	status, err := httpServer.blockService.GetBridgeReqWithStatus(data["TxReqID"].(string))
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	return status, nil
}

func processBurningReq(
	burningMetaType int,
	params interface{},
	closeChan <-chan struct{},
	httpServer *HttpServer,
	isPRV bool,
) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	if len(arrayParams) > 6 {
		privacyTemp, ok := arrayParams[6].(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("The privacy mode must be valid "))
		}
		hasPrivacyToken := int(privacyTemp) > 0
		if hasPrivacyToken {
			return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errors.New("The privacy mode must be disabled"))
		}
	}

	senderPrivateKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}

	tokenParamsRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token param is invalid"))
	}

	tokenReceivers, ok := tokenParamsRaw["TokenReceivers"].(interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token receivers is invalid"))
	}

	tokenID, ok := tokenParamsRaw["TokenID"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token ID is invalid"))
	}

	tokenName, ok := tokenParamsRaw["TokenName"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("token name is invalid"))
	}

	remoteAddress, ok := tokenParamsRaw["RemoteAddress"].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("remote address is invalid"))
	}

	var txVersion int8
	tmpVersionParam, ok := tokenParamsRaw["TxVersion"]
	if !ok {
		txVersion = 2
	} else {
		tmpVersion, ok := tmpVersionParam.(float64)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("txVersion must be a float64"))
		}
		txVersion = int8(tmpVersion)
	}

	meta, err := rpcservice.NewBurningRequestMetadata(
		senderPrivateKeyParam,
		tokenReceivers,
		tokenID,
		tokenName,
		remoteAddress,
		burningMetaType,
		httpServer.GetBlockchain(),
		httpServer.GetBlockchain().BeaconChain.CurrentHeight(),
		txVersion,
		common.DefaultNetworkID,
		0,
		nil,
	)
	if err != nil {
		return nil, err
	}
	var byteArrays []byte
	var err2 error
	var txHash string
	if isPRV {
		// create new param to build raw tx from param interface
		createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
		if errNewParam != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
		}
		rawTx, rpcErr := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
		if rpcErr != nil {
			Logger.log.Error(rpcErr)
			return nil, rpcErr
		}
		byteArrays, err2 = json.Marshal(rawTx)
		txHash = rawTx.Hash().String()
	} else {
		customTokenTx, rpcErr := httpServer.txService.BuildRawPrivacyCustomTokenTransaction(params, meta)
		if rpcErr != nil {
			Logger.log.Error(rpcErr)
			return nil, rpcErr
		}
		byteArrays, err2 = json.Marshal(customTokenTx)
		txHash = customTokenTx.Hash().String()
	}

	if err2 != nil {
		Logger.log.Error(err2)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err2)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            txHash,
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningForDepositToSCReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningForDepositToSCRequestMetaV2,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningForDepositToSCRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningForDepositToSCReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateAndSendBurningForDepositToSCRequestV2(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return httpServer.handleCreateAndSendBurningForDepositToSCRequest(params, closeChan)
}

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingBSCReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingBSCRequestMeta)
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

func (httpServer *HttpServer) handleCreateRawTxWithBurningBSCReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPBSCRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningBSCRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningBSCReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningPRVERC20Req(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPRVERC20RequestMeta,
		params,
		closeChan,
		httpServer,
		true,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningPRVERC20Request(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningPRVERC20Req(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningPRVBEP20Req(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPRVBEP20RequestMeta,
		params,
		closeChan,
		httpServer,
		true,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningPRVBEP20Request(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningPRVBEP20Req(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingPRVERC20Req(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingPRVERC20RequestMeta)
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

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingPRVBEP20Req(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingPRVBEP20RequestMeta)
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

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingPLGReq(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingPLGRequestMeta)
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

func (httpServer *HttpServer) handleCreateRawTxWithBurningPLGReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPLGRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningPLGRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningPLGReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningPBSCForDepositToSCReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPBSCForDepositToSCRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningPBSCForDepositToSCRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningPBSCForDepositToSCReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningPLGForDepositToSCReq(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningPLGForDepositToSCRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningPLGForDepositToSCRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningPLGForDepositToSCReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingFTMReq(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingFantomRequestMeta)
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

func (httpServer *HttpServer) handleCreateRawTxWithBurningFantomReq(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningFantomRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningFTMRequest(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningFantomReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

func (httpServer *HttpServer) handleCreateRawTxWithBurningFantomForDepositToSCReq(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningFantomForDepositToSCRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningFTMForDepositToSCRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningFantomForDepositToSCReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}
	return sendResult, nil
}

// Aurora bridge
func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingAURORAReq(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMAuroraReq(params, closeChan, metadata.IssuingAuroraRequestMeta)
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

func (httpServer *HttpServer) handleCreateRawTxWithIssuingEVMAuroraReq(params interface{}, closeChan <-chan struct{}, metatype int) (interface{}, *rpcservice.RPCError) {
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	// get meta data from params
	data, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata is invalid"))
	}
	meta, err := metadataBridge.NewIssuingEVMAuroraRequestFromMap(data, 0, metatype)
	if err != nil {
		rpcErr := rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
		Logger.log.Error(rpcErr)
		return nil, rpcErr
	}

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err1 := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
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

func (httpServer *HttpServer) handleCreateRawTxWithBurningAuroraReq(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningAuroraRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningAURORARequest(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningAuroraReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}

// Avalanche bridge
func (httpServer *HttpServer) handleCreateAndSendTxWithIssuingAVAXReq(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithIssuingEVMReq(params, closeChan, metadata.IssuingAvaxRequestMeta)
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

func (httpServer *HttpServer) handleCreateRawTxWithBurningAvaxReq(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	return processBurningReq(
		metadata.BurningAvaxRequestMeta,
		params,
		closeChan,
		httpServer,
		false,
	)
}

func (httpServer *HttpServer) handleCreateAndSendBurningAVAXRequest(params interface{},
	closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	data, err := httpServer.handleCreateRawTxWithBurningAvaxReq(params, closeChan)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	sendResult, err1 := httpServer.handleSendRawPrivacyCustomTokenTransaction(newParam, closeChan)
	if err1 != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err1)
	}

	return sendResult, nil
}
