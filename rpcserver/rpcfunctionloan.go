package rpcserver

import (
	"encoding/hex"
	"encoding/json"
	"github.com/ninjadotorg/constant/transaction"
	"github.com/ninjadotorg/constant/wire"
	"github.com/ninjadotorg/constant/common"
)

func (self RpcServer) handleCreateRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {

	return nil, nil
}

func (self RpcServer) handleSendRawLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {
	Logger.log.Info(params)
	arrayParams := common.InterfaceSlice(params)
	hexRawTx := arrayParams[0].(string)
	rawTxBytes, err := hex.DecodeString(hexRawTx)

	if err != nil {
		return nil, err
	}
	tx := transaction.TxLoanRequest{}
	//tx := transaction.TxCustomToken{}
	// Logger.log.Info(string(rawTxBytes))
	err = json.Unmarshal(rawTxBytes, &tx)
	if err != nil {
		return nil, err
	}

	hash, txDesc, err := self.config.TxMemPool.MaybeAcceptTransaction(&tx)
	if err != nil {
		return nil, err
	}

	Logger.log.Infof("there is hash of transaction: %s\n", hash.String())
	Logger.log.Infof("there is priority of transaction in pool: %d", txDesc.StartingPriority)

	// broadcast message
	txMsg, err := wire.MakeEmptyMessage(wire.CmdCustomToken)
	if err != nil {
		return nil, err
	}

	txMsg.(*wire.MessageTx).Transaction = &tx
	self.config.Server.PushMessageToAll(txMsg)

	return tx.Hash(), nil
}

func (self RpcServer) handleCreateAndSendLoanRequest(params interface{}, closeChan <-chan struct{}) (interface{}, error) {



	return nil, nil
}
