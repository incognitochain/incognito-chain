package rpcserver

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/pkg/errors"
)

func createJSONResult(tx *transaction.Tx, meta metadata.Metadata) (interface{}, *RPCError) {
	if meta == nil {
		return nil, NewRPCError(ErrUnexpected, errors.Errorf("Invalid Metadata"))
	}
	tx.Metadata = meta
	byteArrays, marshalErr := json.Marshal(tx)
	if marshalErr != nil {
		Logger.log.Error(marshalErr)
		return nil, NewRPCError(ErrUnexpected, marshalErr)
	}
	result := jsonresult.CreateTransactionResult{
		TxID:            tx.Hash().String(),
		Base58CheckData: base58.Base58Check{}.Encode(byteArrays, 0x00),
	}
	return result, nil
}

func (self RpcServer) handleCreateAndSendTxWithCMBInitRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	normalTx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	// Req param #4: cmb init request
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbInitRequest := metadata.NewCMBInitRequest(paramsMap)
	return createJSONResult(normalTx, cmbInitRequest)
}

func (self RpcServer) handleCreateAndSendTxWithCMBInitResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	normalTx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	// Req param #4: cmb init response
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbInitResponse := metadata.NewCMBInitResponse(paramsMap)
	return createJSONResult(normalTx, cmbInitResponse)
}

func (self RpcServer) handleCreateAndSendTxWithCMBDepositContract(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	normalTx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	// Req param #4: cmb deposit contract
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbDepositContract := metadata.NewCMBDepositContract(paramsMap)
	return createJSONResult(normalTx, cmbDepositContract)
}

func (self RpcServer) handleCreateAndSendTxWithCMBDepositSend(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	normalTx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	// Req param #4: cmb deposit send
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbDepositSend := metadata.NewCMBDepositSend(paramsMap)
	return createJSONResult(normalTx, cmbDepositSend)
}

func (self RpcServer) handleCreateAndSendTxWithCMBWithdrawRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	normalTx, err := self.buildRawTransaction(params)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	// Req param #4: cmb withdraw request
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbWithdrawReq := metadata.NewCMBWithdrawRequest(paramsMap)
	return createJSONResult(normalTx, cmbWithdrawReq)
}
