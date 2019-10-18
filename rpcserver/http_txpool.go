package rpcserver

import (
	"errors"
	"github.com/incognitochain/incognito-chain/common"
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

/*
handleGetPendingTxsInBlockgen - RPC returns all transaction ids in blockgen
*/
func (httpServer *HttpServer) handleGetPendingTxsInBlockgen(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetPendingTxsInBlockgen params: %+v", params)
	result := jsonresult.NewGetPendingTxsInBlockgenResult(httpServer.config.Blockgen.GetPendingTxsV2())
	Logger.log.Debugf("handleGetPendingTxsInBlockgen result: %+v", result)
	return result, nil
}

func (httpServer *HttpServer) handleGetNumberOfTxsInMempool(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleGetNumberOfTxsInMempool params: %+v", params)
	result := httpServer.txMemPoolService.GetNumberOfTxsInMempool()
	Logger.log.Debugf("handleGetNumberOfTxsInMempool result: %+v", result)
	return result, nil
}

/*
handleMempoolEntry - RPC fetch a specific transaction from the mempool
*/
func (httpServer *HttpServer) handleMempoolEntry(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	Logger.log.Debugf("handleMempoolEntry params: %+v", params)
	// Param #1: hash string of tx(tx id)
	txIDParam, ok := params.(string)
	if !ok || txIDParam == "" {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("transaction id is invalid"))
	}

	txInPool, shardID, err := httpServer.txMemPoolService.MempoolEntry(txIDParam)
	if err != nil {
		return nil, err
	}

	tx, errM := jsonresult.NewTransactionDetail(txInPool, nil, 0, 0, shardID)
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
	arrays := common.InterfaceSlice(params)
	if arrays == nil {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, errors.New("param is invalid"))
	}

	result := []bool{}
	for _, txHashString := range arrays {
		txHash, ok := txHashString.(string)
		if ok {
			t, _ := httpServer.txMemPoolService.RemoveTxInMempool(txHash)
			result = append(result, t)
		} else {
			result = append(result, false)
		}
	}
	return result, nil
}
