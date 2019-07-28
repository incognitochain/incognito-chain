package main

import (
	"github.com/incognitochain/incognito-chain/rpcserver"
)

type command func(*Client, ...interface{}) (interface{}, *rpcserver.RPCError)

var Command = map[string]command{
	createAndSendTransaction: (*Client).createAndSendTransaction,
	getBalanceByPrivatekey:   (*Client).getBalanceByPrivatekey,
	getTransactionByHash:     (*Client).getTransactionByHash,
	getBlockChainInfo:        (*Client).getBlockChainInfo,
}
