package rpcserver

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/transaction"
	"github.com/incognitochain/incognito-chain/wire"
)

type metaConstructorType func(map[string]interface{}) (metadata.Metadata, error)

var metaConstructors = map[string]metaConstructorType{
	createAndSendIssuingRequest: metadata.NewIssuingRequestFromMap,
	// createAndSendContractingRequest: metadata.NewContractingRequestFromMap,
}

func (httpServer *HttpServer) createRawTxWithMetadata(params interface{}, closeChan <-chan struct{}, metaConstructorType metaConstructorType) (interface{}, *rpcservice.RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	metaRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	meta, errCons := metaConstructorType(metaRaw)

	_, _, errParseKey := rpcservice.GetKeySetFromPrivateKeyParams(arrayParams[0].(string))
	if err := common.CheckError(errCons, errParseKey); err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	tx, err := httpServer.txService.BuildRawTransaction(params, meta)
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

func (httpServer *HttpServer) createRawCustomTokenTxWithMetadata(params interface{}, closeChan <-chan struct{}, metaConstructorType metaConstructorType) (interface{}, *rpcservice.RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	metaRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	meta, errCons := metaConstructorType(metaRaw)
	if errCons != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errCons)
	}
	tx, err := httpServer.txService.BuildRawCustomTokenTransaction(params, meta)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}
	fmt.Printf("sigPubKey after build: %v\n", tx.SigPubKey)
	byteArrays, errMarshal := json.Marshal(tx)
	if errMarshal != nil {
		// return hex for a new tx
		return nil, rpcservice.NewRPCError(rpcservice.JsonError, errMarshal)
	}
	fmt.Printf("Created raw tx: %+v\n", tx)
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (httpServer *HttpServer) sendRawTxWithMetadata(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	tx := transaction.Tx{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.JsonError, err)
	}

	hash, _, err := httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TxPoolRejectTxError, err)
	}

	Logger.log.Debugf("there is hash of transaction: %s\n", hash.String())

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	err = httpServer.config.Server.PushMessageToAll(txMsg)
	if err == nil {
		httpServer.config.TxMemPool.MarkForwardedTransaction(*tx.Hash())
	}
	httpServer.config.TxMemPool.MarkForwardedTransaction(*tx.Hash())
	result := jsonresult.CreateTransactionResult{
		TxID: tx.Hash().String(),
	}
	return result, nil
}

func (httpServer *HttpServer) sendRawCustomTokenTxWithMetadata(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	tx := transaction.TxNormalToken{}
	err = json.Unmarshal(rawTxBytes, &tx)
	fmt.Printf("%+v\n", tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.JsonError, err)
	}

	hash, _, err := httpServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.TxPoolRejectTxError, err)
	}

	Logger.log.Debugf("there is hash of transaction: %s\n", hash.String())

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, err)
	}

	txMsg.(*wire.MessageTxToken).Transaction = &tx
	err = httpServer.config.Server.PushMessageToAll(txMsg)
	if err == nil {
		httpServer.config.TxMemPool.MarkForwardedTransaction(*tx.Hash())
	}
	httpServer.config.TxMemPool.MarkForwardedTransaction(*tx.Hash())
	result := jsonresult.CreateTransactionResult{
		TxID: tx.Hash().String(),
	}
	return result, nil
}

func (httpServer *HttpServer) createAndSendTxWithMetadata(params interface{}, closeChan <-chan struct{}, createHandler, sendHandler httpHandler) (interface{}, *rpcservice.RPCError) {
	data, err := createHandler(httpServer, params, closeChan)
	fmt.Printf("err create handler: %v\n", err)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	return sendHandler(httpServer, newParam, closeChan)
}
