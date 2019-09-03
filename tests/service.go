package main

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/rpcserver/rpcservice"

	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
)

func (client *Client) getBlockChainInfo() (*jsonresult.GetBlockChainInfoResult, *rpcservice.RPCError) {
	result := &jsonresult.GetBlockChainInfoResult{}
	res, rpcError := makeRPCRequest(client, getBlockChainInfo, []string{})
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return result, rpcservice.NewRPCError(rpcservice.NetworkError, err)
	}
	return result, res.Error
}
func (client *Client) createAndSendTransaction(params []interface{}) (*jsonresult.CreateTransactionResult, *rpcservice.RPCError) {
	result := &jsonresult.CreateTransactionResult{}
	res, rpcError := makeRPCRequest(client, createAndSendTransaction, params...)
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return result, rpcservice.NewRPCError(rpcservice.NetworkError, err)
	}
	return result, nil
}
func (client *Client) getBalanceByPrivatekey(params string) (uint64, *rpcservice.RPCError) {
	var result interface{}
	res, rpcError := makeRPCRequest(client, getBalanceByPrivatekey, params)
	if rpcError != nil {
		return 0, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return 0, rpcservice.NewRPCError(rpcservice.NetworkError, err)
	}
	return uint64(result.(float64)), nil
}
func (client *Client) getTransactionByHash(params string) (interface{}, *rpcservice.RPCError) {
	result := &jsonresult.TransactionDetail{}
	res, rpcError := makeRPCRequest(client, getTransactionByHash, params)
	if rpcError != nil {
		return result, rpcError
	}
	err := json.Unmarshal(res.Result, &result)
	if err != nil {
		return result, rpcservice.NewRPCError(rpcservice.NetworkError, err)
	}
	return result, nil
}
