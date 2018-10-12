package main

import (
	"github.com/ninjadotorg/cash-prototype/blockchain"
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

// netName returns the name used when referring to a coin network.  At the
// time of writing, btcd currently places blocks for testnet version 3 in the
// data and log directory "testnet", which does not match the Name field of the
// chaincfg parameters.  This function can be used to override this directory
// name as "testnet" when the passed active network matches wire.TestNet.
//
// A proper upgrade to move the data and log directories for this network to
// "testnet3" is planned for the future, at which point this function can be
// removed and the network parameter's name used instead.
func netName(chainParams *params) string {
	return chainParams.Name
}
