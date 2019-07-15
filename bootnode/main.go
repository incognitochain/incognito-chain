//+build !test

package main

import (
	"log"

	"github.com/incognitochain/incognito-chain/bootnode/server"
)

var (
	cfg *config
)

// Bootnode is a centralized rpc server, which be used for receive Ping method from incognito node
// and store in mem list connectable peers in incognito network
func main() {
	// Show Version at startup.
	log.Printf("Version %s\n", Version)

	// Load config
	tcfg, err := loadConfig()
	if err != nil {
		log.Println("Parse config error", err.Error())
		return
	}
	cfg = tcfg

	// create RPC config for RPC server
	rpcConfig := server.RpcServerConfig{
		Port: cfg.RPCPort,
	}

	// Init RPC Serer in golang
	rpcServer := &server.RpcServer{}
	log.Printf("Init rpcServer with config \n")
	rpcServer.Init(&rpcConfig)

	log.Printf("Start rpcServer with config \n %+v\n", rpcServer.Config)
	for {
		// Start rpcServer and listen request from rpc client
		rpcServer.Start()
	}
}
