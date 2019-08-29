package main

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/rpcserver"
	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

func (client *Client) getBlockChainInfo() (*jsonresult.GetBlockChainInfoResult, *rpcserver.RPCError) {
	result := &jsonresult.GetBlockChainInfoResult{}
	res, rpcError := makeRPCRequest(client, getBlockChainInfo, []string{})
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return result, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	return result, res.Error
}
func (client *Client) createAndSendTransaction(params []interface{}) (*jsonresult.CreateTransactionResult, *rpcserver.RPCError) {
	result := &jsonresult.CreateTransactionResult{}
	res, rpcError := makeRPCRequest(client, createAndSendTransaction, params...)
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return result, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	return result, nil
}
func (client *Client) getBalanceByPrivatekey(params string) (uint64, *rpcserver.RPCError) {
	var result interface{}
	res, rpcError := makeRPCRequest(client, getBalanceByPrivatekey, params)
	if rpcError != nil {
		return 0, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return 0, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	return uint64(result.(float64)), nil
}
func (client *Client) getTransactionByHash(params string) (interface{}, *rpcserver.RPCError) {
	result := &jsonresult.TransactionDetail{}
	res, rpcError := makeRPCRequest(client, getTransactionByHash, params)
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return result, rpcserver.NewRPCError(rpcserver.NetworkError, err)
	}
	return result, nil
}
