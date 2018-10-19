package main

import (
	"github.com/ninjadotorg/cash/bootnode/server"
	"log"
)

var (
	cfg *config
)

func main() {
	// Show version at startup.
	log.Printf("Version %s\n", "0.0.1")

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
	log.Println("Start server")
	server.Start()
}
