package rpcserver

import (
	"encoding/json"
	"errors"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/rpcserver/jsonresult"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wire"
	"github.com/ninjadotorg/constant/common/base58"
)

func (self RpcServer) handleGetLoanParams(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	return self.config.BlockChain.BestState[0].BestBlock.Header.DCBConstitution.DCBParams.LoanParams, nil
}

func (self RpcServer) handleCreateRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	loanRequestRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	loanRequest := metadata.NewLoanRequest(loanRequestRaw)
	if loanRequest == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Loan data missing"))
	}
	tx.Metadata = loanRequest
	return tx, err
}

func (self RpcServer) handleSendRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckDate := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckDate)

	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}
	tx := transaction.Tx{}
	//tx := transaction.TxCustomToken{}
	// Logger.log.Info(string(rawTxBytes))
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

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawLoanRequest(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawLoanRequest(newParam, closeChan)
	return txId, err
}

func (self RpcServer) handleCreateRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	loanResponseRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	loanResponse := metadata.NewLoanResponse(loanResponseRaw)
	if loanResponse == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Loan data missing"))
	}
	tx.Metadata = loanResponse
	return tx, err
}

func (self RpcServer) handleSendRawLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCLoanResponseToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendLoanResponse(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawLoanResponse(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawLoanResponse(newParam, closeChan)
	return txId, err
}

func (self RpcServer) handleCreateRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	loanWithdrawRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	loanWithdraw := metadata.NewLoanWithdraw(loanWithdrawRaw)
	if loanWithdraw == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Loan data missing"))
	}
	tx.Metadata = loanWithdraw
	return tx, err
}

func (self RpcServer) handleSendRawLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
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
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCLoanWithdrawToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendLoanWithdraw(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawLoanWithdraw(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawLoanWithdraw(newParam, closeChan)
	return txId, err
}

func (self RpcServer) handleCreateRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	tx, err := self.buildRawTransaction(params)
	arrayParams := common.InterfaceSlice(params)
	loanPaymentRaw := arrayParams[len(arrayParams)-1].(map[string]interface{})
	loanPayment := metadata.NewLoanPayment(loanPaymentRaw)
	if loanPayment == nil {
		return nil, NewRPCError(ErrUnexpected, errors.New("Loan data missing"))
	}
	tx.Metadata = loanPayment
	return tx, err
}

func (self RpcServer) handleSendRawLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	base58CheckData := arrayParams[0].(string)
	rawTxBytes, _, err := base58.Base58Check{}.Decode(base58CheckData)

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
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCLoanPayToken)
	if err != nil {
		return nil, NewRPCError(ErrUnexpected, err)
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendLoanPayment(params interface{}, closeChan <-chan struct{}) (interface{}, *RPCError) {
	data, err := self.handleCreateRawLoanPayment(params, closeChan)
	if err != nil {
		return nil, err
	}
	tx := data.(jsonresult.CreateTransactionResult)
	base58CheckData := tx.Base58CheckData
	if err != nil {
		return nil, err
	}
	newParam := make([]interface{}, 0)
	newParam = append(newParam, base58CheckData)
	txId, err := self.handleSendRawLoanPayment(newParam, closeChan)
	return txId, err
}
