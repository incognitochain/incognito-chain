package main

import (
	"github.com/ninjadotorg/cash-prototype/bootnode/server"
	"log"
)

var (
	cfg *config
)

func main() {
	// Show version at startup.
	log.Printf("Version %s\n", "1")

	// load config
	tcfg, err := loadConfig()
	if err != nil {
		Logger.log.Info("Parse config error", err.Error())
		return
	}
	cfg = tcfg

	rpcConfig := server.RpcServerConfig{
		Port: cfg.RPCPort,
	}
	server := &server.RpcServer{}
	err = server.Init(&rpcConfig)
	if err != nil {
		Logger.log.Info("Init bootnode error", err.Error())
		return
	}
	Logger.log.Info("Start server")
	server.Start()
}
