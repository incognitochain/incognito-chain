package rpcserver

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

/*
handleGetMempoolInfo - RPC returns information about the node's current txs memory pool
*/
func (httpServer *HttpServer) handleGetMempoolInfo(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetMempoolInfo params: %+v", params)
	result := jsonresult.NewGetMempoolInfo(httpServer.config.TxMemPool)
	Logger.log.Debugf("handleGetMempoolInfo result: %+v", result)
	return result, nil
}

/*
handleGetRawMempool - RPC returns all transaction ids in memory pool as a json array of string transaction ids
Hint: use getmempoolentry to fetch a specific transaction from the mempool.
*/
func (httpServer *HttpServer) handleGetRawMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetRawMempool params: %+v", params)
	result := jsonresult.NewGetRawMempoolResult(*httpServer.config.TxMemPool)
	Logger.log.Debugf("handleGetRawMempool result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetNumberOfTxsInMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetNumberOfTxsInMempool params: %+v", params)
	result := len(httpServer.config.TxMemPool.ListTxs())
	Logger.log.Debugf("handleGetNumberOfTxsInMempool result: %+v", result)
	return result, nil
}

/*
handleMempoolEntry - RPC fetch a specific transaction from the mempool
*/
func (httpServer *HttpServer) handleMempoolEntry(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleMempoolEntry params: %+v", params)
	// Param #1: hash string of tx(tx id)
	if params == nil {
		params = ""
	}
	txID, err := common.Hash{}.NewHashFromStr(params.(string))
	if err != nil {
		Logger.log.Debugf("handleMempoolEntry result: nil %+v", err)
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	txInPool, err := httpServer.config.TxMemPool.GetTx(txID)
	if err != nil {
		Logger.log.Error(err)
		return nil, rpcservice.NewRPCError(rpcservice.GeTxFromPoolError, err)
	}
	shardIDTemp := common.GetShardIDFromLastByte(txInPool.GetSenderAddrLastByte())
	tx, errM := jsonresult.NewTransactionDetail(txInPool, nil, 0, 0, shardIDTemp)
	if errM != nil {
		Logger.log.Error(errM)
		return nil, rpcservice.NewRPCError(rpcservice.UnexpectedError, errM)
	}
	tx.IsInMempool = true
	Logger.log.Debugf("handleMempoolEntry result: %+v", tx)
	return tx, nil
}

// handleRemoveTxInMempool - try to remove tx from tx mempool
func (httpServer *HttpServer) handleRemoveTxInMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	if params == nil {
		params = ""
	}
	txID, err := common.Hash{}.NewHashFromStr(params.(string))
	if err != nil {
		Logger.log.Debugf("handleMempoolEntry result: nil %+v", err)
		return false, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, err)
	}

	tempTx, err := httpServer.config.TxMemPool.GetTx(txID)
	if err != nil {
		return false, rpcservice.NewRPCError(rpcservice.GeTxFromPoolError, err)
	}
	httpServer.config.TxMemPool.RemoveTx([]metadata.Transaction{tempTx}, false)
	httpServer.config.TxMemPool.TriggerCRemoveTxs(tempTx)
	return true, nil
}
