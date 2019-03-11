package main

import (
	"log"

	"github.com/big0t/constant-chain/bootnode/server"
)

var (
	cfg *config
)

func main() {
	// Show version at startup.
	log.Printf("Version %s\n", Version)

	// load config
	tcfg, err := loadConfig()
	if err != nil {
		log.Println("Parse config error", err.Error())
		return
	}
	cfg = tcfg

	rpcConfig := server.RpcServerConfig{
		Port: cfg.RPCPort,
	}
	server := &server.RpcServer{}
	err = server.Init(&rpcConfig)
	if err != nil {
		log.Println("Init bootnode error", err.Error())
		return
	}
	log.Printf("Start server with config \n %+v", server.Config)
	for {
		server.Start()
	}
}
