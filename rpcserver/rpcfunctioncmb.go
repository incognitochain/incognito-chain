package rpcserver

import (
	"encoding/json"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/common/base58"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
)

func createJSONResult(tx *transaction.Tx) (interface{}, *RPCError) {
	// if meta == nil {
	// 	return nil, NewRPCError(ErrUnexpected, errors.Errorf("Invalid Metadata"))
	// }
	// tx.Metadata = meta
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
	// Req param #4: cmb init request
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbInitRequest := metadata.NewCMBInitRequest(paramsMap)
	normalTx, err := self.buildRawTransaction(params, cmbInitRequest)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return createJSONResult(normalTx)
}

func (self RpcServer) handleCreateAndSendTxWithCMBInitResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: cmb init response
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbInitResponse := metadata.NewCMBInitResponse(paramsMap)
	normalTx, err := self.buildRawTransaction(params, cmbInitResponse)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return createJSONResult(normalTx)
}

func (self RpcServer) handleCreateAndSendTxWithCMBDepositContract(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: cmb deposit contract
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbDepositContract := metadata.NewCMBDepositContract(paramsMap)
	normalTx, err := self.buildRawTransaction(params, cmbDepositContract)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return createJSONResult(normalTx)
}

func (self RpcServer) handleCreateAndSendTxWithCMBDepositSend(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: cmb deposit send
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbDepositSend := metadata.NewCMBDepositSend(paramsMap)
	normalTx, err := self.buildRawTransaction(params, cmbDepositSend)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return createJSONResult(normalTx)
}

func (self RpcServer) handleCreateAndSendTxWithCMBWithdrawRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	arrayParams := common.InterfaceSlice(params)
	// Req param #4: cmb withdraw request
	paramsMap := arrayParams[4].(map[string]interface{})
	cmbWithdrawReq := metadata.NewCMBWithdrawRequest(paramsMap)
	normalTx, err := self.buildRawTransaction(params, cmbWithdrawReq)
	if err != nil {
		Logger.log.Error(err)
		return nil, NewRPCError(ErrUnexpected, err)
	}
	return createJSONResult(normalTx)
}
