package main

import (
	"log"

	"github.com/incognitochain/incognito-chain/bootnode/server"
)

var (
	cfg *config
)

func main() {
	// Show Version at startup.
	log.Printf("Version %s\n", Version)

	// load config
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
	server := &server.RpcServer{}
	// Init RPC server in golang
	err = server.Init(&rpcConfig)

	if err != nil {
		log.Println("Init bootnode error", err.Error())
		return
	}
	log.Printf("Start server with config \n %+v", server.Config)
	for {
		// Start server and listen request from rpc client
		server.Start()
	}
}
