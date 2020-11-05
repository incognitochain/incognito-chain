package main

import (
	"fmt"
	F "github.com/incognitochain/incognito-chain/devframework"
)

func main() {
	sim := F.NewStandaloneSimulation("sim1", F.Config{
		ShardNumber: 2,
	})
	sim.ConnectNetwork()
	sim.Network.On(F.MSG_BLOCK_BEACON, func(msg interface{}) {
		fmt.Println(msg)
	})

	sim.Pause()
}
