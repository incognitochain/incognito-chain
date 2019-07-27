package main

import (
	"github.com/incognitochain/incognito-chain/rpcserver"
)

type command func(*Client, ...interface{}) (interface{}, *rpcserver.RPCError)
var Command = map[string]command {
	createAndSendTransaction: (*Client).createAndSendTransaction,
	getBlockChainInfo: (*Client).getBlockChainInfo,
}