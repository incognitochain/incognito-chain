package main

import (
	"time"
)

// Config constant
const (
	defaultTimeout = 10 * time.Second
)
const (
	getTransactionByHash     = "gettransactionbyhash"
	createAndSendTransaction = "createandsendtransaction"
	getBalanceByPrivatekey   = "getbalancebyprivatekey"
	getBlockChainInfo        = "getblockchaininfo"
)
