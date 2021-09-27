package rpcserver

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"
)

func (httpServer *HttpServer) handleGetTransactionHashByDecoys(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
	var err error
	arrayParams := common.InterfaceSlice(params)
	if arrayParams == nil || len(arrayParams) < 1 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("param must be an array at least 1 element"))
	}

	paramList, ok := arrayParams[0].(map[string]interface{})
	if !ok || len(paramList) == 0 {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError,fmt.Errorf("paramList %v is not a map[string]interface{}", arrayParams[0]))
	}

	//Get decoyKey list
	decoyKey := "Decoys"
	if _, ok = paramList[decoyKey]; !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("%v not found in %v", decoyKey, paramList))
	}
	decoyInterface, ok := paramList[decoyKey].([]interface{})
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse public keys, not a []interface{}: %v", paramList[decoyKey]))
	}
	decoys := make([]uint64, 0)
	for _, pk := range decoyInterface {
		if tmp, ok := pk.(uint64); !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse public keys, %v is not a string", pk))
		} else {
			decoys = append(decoys, tmp)
		}
	}

	tokenKey := "TokenID"
	tokenID := &common.PRVCoinID
	if tokenParam, ok := paramList[tokenKey]; ok {
		tokenIDStr, ok := tokenParam.(string)
		if !ok {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse tokenID: %v", tokenParam))
		}

		tokenID, err = new(common.Hash).NewHashFromStr(tokenIDStr)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot decode tokenID %v", tokenIDStr))
		}
	}

	shardIDInterface, ok := paramList["shardID"]
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("%v not found in %v", "shardID", paramList))
	}
	shardID, ok := shardIDInterface.(byte)
	if !ok {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse shardID, %v is not a byte", shardIDInterface))
	}
	if shardID >= byte(common.MaxShardNumber) {
		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("invalid shardID %v", shardID))
	}
	db := httpServer.GetBlockchain().GetShardChainDatabase(shardID)

	result := make(map[uint64]map[string]uint64)
	for _, idx := range decoys {
		tmpRes, err := rawdbv2.GetTxByDecoyIndex(db, *tokenID, idx)
		if err != nil {
			return nil, rpcservice.NewRPCError(rpcservice.RPCInternalError, fmt.Errorf("cannot get result for shard %v, token %v, idx %v", shardID, tokenID.String(), idx))
		}
		result[idx] = tmpRes
	}

	return result, nil
}

//// handleGetCoinInfoByHashes handles the request for get input/output coin information by the transaction hashes.
//func (httpServer *HttpServer) handleGetCoinInfoByHashes(params interface{}, closeChan <-chan struct{}) (interface{}, *rpcservice.RPCError) {
//	arrayParams := common.InterfaceSlice(params)
//	if len(arrayParams) == 0 {
//		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("there is no param to proceed"))
//	}
//
//	paramList, ok := arrayParams[0].(map[string]interface{})
//	if !ok {
//		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("param must be a map[string]interface{}"))
//	}
//
//	//Get txHashList
//	txListKey := "TxHashList"
//	if _, ok = paramList[txListKey]; !ok {
//		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("%v not found in %v", txListKey, paramList))
//	}
//	txHashListInterface, ok := paramList[txListKey].([]interface{})
//	if !ok {
//		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse txHashes, not a []interface{}: %v", paramList[txListKey]))
//	}
//
//	txHashList := make([]string, 0)
//	for _, sn := range txHashListInterface {
//		if tmp, ok := sn.(string); !ok {
//			return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("cannot parse txHashes, %v is not a string", sn))
//		} else {
//			txHashList = append(txHashList, tmp)
//		}
//	}
//
//	if len(txHashList) > 100 {
//		return nil, rpcservice.NewRPCError(rpcservice.RPCInvalidParamsError, fmt.Errorf("support at most 100 txs, got %v", len(txHashList)))
//	}
//
//	return httpServer.txService.GetEncodedTransactionsByHashes(txHashList)
//}