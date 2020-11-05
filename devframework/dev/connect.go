package main

import (
	"github.com/incognitochain/incognito-chain/devframework"
	"github.com/incognitochain/incognito-chain/devframework/mock"
)

func main() {
	consensus := &mock.Consensus{}
	config := devframework.HighwayConnectionConfig{
		"127.0.0.1",
		19876,
		"2.0.0",
		"45.56.115.6:9330",
		"",
		consensus,
	}
	conn := devframework.NewHighwayConnection(config)
	conn.ConnectHighway()
	select {}
}
