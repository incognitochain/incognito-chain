package rpcserver

import (
	"encoding/json"
	"fmt"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/common/base58"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/rpcserver/jsonresult"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/constant-money/constant-chain/wire"
)

type metaConstructorType func(map[string]interface{}) (metadata.Metadata, error)

// var metaConstructors = map[string]metaConstructorType{
// 	CreateAndSendIssuingRequest:     metadata.NewIssuingRequestFromMap,
// 	CreateAndSendContractingRequest: metadata.NewContractingRequestFromMap,
// }

func (rpcServer RpcServer) createRawTxWithMetadata(params interface{}, closeChan <-chan struct{}, metaConstructorType metaConstructorType) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	metaRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	meta, errCons := metaConstructorType(metaRaw)
	if errCons != nil {
		return nil, NewRPCError(ErrUnexpected, errCons)
	}

	_, errParseKey := rpcServer.GetKeySetFromPrivateKeyParams(arrayParams[0].(string))
	if err := common.CheckError(errCons, errParseKey); err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	tx, err := rpcServer.buildRawTransaction(params, meta)
	if err != nil {
		return nil, err
	}
	byteArrays, errMarshal := json.Marshal(tx)
	if errMarshal != nil {
		// return hex for a new tx
		return nil, NewRPCError(ErrUnexpected, errMarshal)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) createRawCustomTokenTxWithMetadata(params interface{}, closeChan <-chan struct{}, metaConstructorType metaConstructorType) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	metaRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	meta, errCons := metaConstructorType(metaRaw)
	if errCons != nil {
		return nil, NewRPCError(ErrUnexpected, errCons)
	}
	tx, err := rpcServer.buildRawCustomTokenTransaction(params, meta)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	fmt.Printf("sigPubKey after build: %v\n", tx.SigPubKey)
	byteArrays, errMarshal := json.Marshal(tx)
	if errMarshal != nil {
		// return hex for a new tx
		return nil, NewRPCError(ErrUnexpected, errMarshal)
	}
	fmt.Printf("Created raw tx: %+v\n", tx)
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (rpcServer RpcServer) sendRawTxWithMetadata(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.Tx{}
	err = json.Unmarshal(rawTxBytes, &tx)
	// fmt.Printf("[db] sendRawTx received tx: %+v\n", tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, _, err := rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdTx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	err = rpcServer.config.Server.PushMessageToAll(txMsg)
	if err == nil {
		rpcServer.config.TxMemPool.MarkFowardedTransaction(*tx.Hash())
	}
	rpcServer.config.TxMemPool.MarkFowardedTransaction(*tx.Hash())
	result := jsonresult.CreateTransactionResult{
		TxID: tx.Hash().String(),
	}
	return result, nil
}

func (rpcServer RpcServer) sendRawCustomTokenTxWithMetadata(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.TxCustomToken{}
	err = json.Unmarshal(rawTxBytes, &tx)
	fmt.Printf("%+v\n", tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, _, err := rpcServer.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTxToken).Transaction = &tx
	err = rpcServer.config.Server.PushMessageToAll(txMsg)
	if err == nil {
		rpcServer.config.TxMemPool.MarkFowardedTransaction(*tx.Hash())
	}
	rpcServer.config.TxMemPool.MarkFowardedTransaction(*tx.Hash())
	result := jsonresult.CreateTransactionResult{
		TxID: tx.Hash().String(),
	}
	return result, nil
}

func (rpcServer RpcServer) createAndSendTxWithMetadata(params interface{}, closeChan <-chan struct{}, createHandler, sendHandler commandHandler) (interface{}, *RPCError) {
	data, err := createHandler(rpcServer, params, closeChan)
	fmt.Printf("err create handler: %v\n", err)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	return sendHandler(rpcServer, newParam, closeChan)
}
