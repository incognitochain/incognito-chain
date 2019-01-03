package rpcserver

import (
	"encoding/json"
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wire"
)

type metaConstructor func(map[string]interface{}) (metadata.Metadata, error)

var constructors = map[string]metaConstructor{
	CreateAndSendLoanRequest:  metadata.NewLoanRequest,
	CreateAndSendLoanResponse: metadata.NewLoanResponse,
	CreateAndSendLoanWithdraw: metadata.NewLoanWithdraw,
	CreateAndSendLoanPayment:  metadata.NewLoanPayment,
}

func (self RpcServer) handleGetLoanParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution.DCBParams.LoanParams, nil
}

func (self RpcServer) createRawLoanTx(params interface{}, closeChan <-chan struct{}, metaConstructor metaConstructor) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	loanDataRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	loanMeta, errCons := metaConstructor(loanDataRaw)
	if errCons != nil {
		return nil, NewRPCError(ErrUnexpected, errCons)
	}
	tx, err := self.buildRawTransaction(params, loanMeta)
	fmt.Printf("sigPubKey after build: %v\n", tx.SigPubKey)
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

func (self RpcServer) sendRawLoanTx(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.Tx{}
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCLoanRequestToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	result := jsonresult.CreateTransactionResult{
		TxID: tx.Hash().String(),
	}
	return result, nil
}

func (self RpcServer) createAndSendLoanTx(params interface{}, closeChan <-chan struct{}, createHandler, sendHandler commandHandler) (interface{}, *RPCError) {
	data, err := createHandler(self, params, closeChan)
	fmt.Printf("err create handler: %v\n", err)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	return sendHandler(self, newParam, closeChan)
}

func (self RpcServer) handleCreateRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanRequest]
	return self.createRawLoanTx(params, closeChan, constructor)
}

func (self RpcServer) handleSendRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.sendRawLoanTx(params, closeChan)
}

func (self RpcServer) handleCreateAndSendLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanRequest,
		RpcServer.handleSendRawLoanRequest,
	)
}

func (self RpcServer) handleCreateRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanResponse]
	return self.createRawLoanTx(params, closeChan, constructor)
}

func (self RpcServer) handleSendRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.sendRawLoanTx(params, closeChan)
}

func (self RpcServer) handleCreateAndSendLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanResponse,
		RpcServer.handleSendRawLoanResponse,
	)
}

func (self RpcServer) handleCreateRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanWithdraw]
	return self.createRawLoanTx(params, closeChan, constructor)
}

func (self RpcServer) handleSendRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.sendRawLoanTx(params, closeChan)
}

func (self RpcServer) handleCreateAndSendLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanWithdraw,
		RpcServer.handleSendRawLoanWithdraw,
	)
}

func (self RpcServer) handleCreateRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	constructor := constructors[CreateAndSendLoanPayment]
	return self.createRawLoanTx(params, closeChan, constructor)
}

func (self RpcServer) handleSendRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.sendRawLoanTx(params, closeChan)
}

func (self RpcServer) handleCreateAndSendLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.createAndSendLoanTx(
		params,
		closeChan,
		RpcServer.handleCreateRawLoanPayment,
		RpcServer.handleSendRawLoanPayment,
	)
}
