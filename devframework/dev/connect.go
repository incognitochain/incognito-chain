package main

import (
	"github.com/incognitochain/incognito-chain/devframework"
)

func main() {
	_ = devframework.NewIncognitoNode("devnode", 1, nil, "45.56.115.6:9330", true)

	// sim.OnReceive(F.MSG_BLOCK_BEACON, func(msg interface{}) {
	// 	//process 1st listenner
	// 	fmt.Println("1 process receive", msg)
	// })
	// sim.OnInserted(F.BLK_BEACON, func(msg interface{}) {
	// 	//process 2nd listenner
	// 	fmt.Println("2 process receive", msg)
	// })
	// sim.Pause()

	select {}
}
