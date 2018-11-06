package main

import (
	"github.com/ninjadotorg/constant/blockchain"
)

// activeNetParams is a pointer to the parameters specific to the
// currently active network.
var activeNetParams = &mainNetParams

// params is used to group parameters for various networks such as the main
// network and test networks.
type params struct {
	*blockchain.Params
	rpcPort string
}

var mainNetParams = params{
	Params:  &blockchain.MainNetParams,
	rpcPort: MainnetRpcServerPort,
}

var testNetParams = params{
	Params:  &blockchain.TestNetParams,
	rpcPort: TestnetRpcServerPort,
}

// netName returns the name used when referring to a coin network.
func netName(chainParams *params) string {
	return chainParams.Name
}
