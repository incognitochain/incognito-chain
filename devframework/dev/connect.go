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
	sim.OnReceive(F.MSG_BLOCK_BEACON, func(msg interface{}) {
		//process 1st listenner
		fmt.Println("1 process receive", msg)
	})
	sim.OnInserted(F.BLK_BEACON, func(msg interface{}) {
		//process 2nd listenner
		fmt.Println("2 process receive", msg)
	})
	sim.Pause()
}
