package rpcserver

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/rpcserver/bean"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

type metaConstructorType func(map[string]interface{}) (metadata.Metadata, error)

var metaConstructors = map[string]metaConstructorType{
	createAndSendIssuingRequest: metadata.NewIssuingRequestFromMap,
	// createAndSendContractingRequest: metadata.NewContractingRequestFromMap,
}

func (httpServer *HttpServer) createRawTxWithMetadata(params interface{}, closeChan <-chan struct{}, metaConstructorType metaConstructorType) (interface{}, *rpcservice.RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 5 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 5 elements"))
	}

	// param #5 get meta data param
	metaRaw, ok := arrayParams[4].(map[string]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("metadata param is invalid"))
	}

	privateKeyParam, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("private key is invalid"))
	}

	meta, errCons := metaConstructorType(metaRaw)
	_, _, errParseKey := rpcservice.GetKeySetFromPrivateKeyParams(privateKeyParam)
	if err := common.CheckError(errCons, errParseKey); err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	// create new param to build raw tx from param interface
	createRawTxParam, errNewParam := bean.NewCreateRawTxParam(params)
	if errNewParam != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errNewParam)
	}

	tx, err := httpServer.txService.BuildRawTransaction(createRawTxParam, meta)
	if err != nil {
		Logger.log.Errorf("\n\n\n\n\n\n\n createRawTxWithMetadata Error 0 %+v \n\n\n\n\n\n", err)
		return nil, err
	}
	byteArrays, errMarshal := json.Marshal(tx)
	if errMarshal != nil {
		Logger.log.Errorf("\n\n\n\n\n\n\n createRawTxWithMetadata Error %+v \n\n\n\n\n\n", errMarshal)
		return nil, rpcservice.NewRPCError(rpcservice.JsonError, errMarshal)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	Logger.log.Debugf("\n\n\n\n\n\n\n createRawTxWithMetadata OK \n\n\n\n\n\n")
	return result, nil
}

// not to be used
//func (httpServer *HttpServer) createRawCustomTokenTxWithMetadata(params interface{}, closeChan <-chan struct{}, metaConstructorType metaConstructorType) (interface{}, *rpcservice.RPCError) {
//	Logger.log.Info(params)
//	arrayParams := common.InterfaceSlice(params)
//	metaRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
//	meta, errCons := metaConstructorType(metaRaw)
//	if errCons != nil {
//		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errCons)
//	}
//	tx, err := httpServer.txService.BuildRawCustomTokenTransaction(params, meta)
//	if err != nil {
//		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
//	}
//	fmt.Printf("sigPubKey after build: %v\n", tx.SigPubKey)
//	byteArrays, errMarshal := json.Marshal(tx)
//	if errMarshal != nil {
//		// return hex for a new tx
//		return nil, rpcservice.NewRPCError(rpcservice.JsonError, errMarshal)
//	}
//	fmt.Printf("Created raw tx: %+v\n", tx)
//	result := jsonresult.CreateTransactionResult{
//		TxID:            tx.Hash().String(),
//		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
//	}
//	return result, nil
//}

/*func (httpServer *HttpServer) sendRawTxWithMetadata(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param must be an array at least 1 element"))
	}

	base58CheckData, ok := arrayParams[0].(string)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("base58CheckData param is invalid"))
	}

	txMsg, txHash, _, err := httpServer.txService.SendRawTransaction(base58CheckData)
	if err != nil {
		return nil, err
	}

	err2 := httpServer.config.Server.PushMessageToAll(txMsg)
	if err2 == nil {
		httpServer.config.TxMemPool.MarkForwardedTransaction(*txHash)
	}
	httpServer.config.TxMemPool.MarkForwardedTransaction(*txHash)
	result := jsonresult.CreateTransactionResult{
		TxID: txHash.String(),
	}
	return result, nil
}*/

func (httpServer *HttpServer) createAndSendTxWithMetadata(params interface{}, closeChan <-chan struct{}, createHandler, sendHandler httpHandler) (interface{}, *rpcservice.RPCError) {
	data, err := createHandler(httpServer, params, closeChan)
	Logger.log.Errorf("err create handler: %v\n", err)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	return sendHandler(httpServer, newParam, closeChan)
}
