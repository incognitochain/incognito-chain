package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/testsuite/apps/remotetests"
	"time"
)

func NormalScenarioTest(nodeManager remotetests.NodeManager) {
	fmt.Println("Normal Test: Checking having only 1 view in multiview ... ")
	beaconNode := nodeManager.BeaconNode[0]
	shardNode0 := nodeManager.ShardFixNode[0][0]
	shardNode1 := nodeManager.ShardFixNode[1][0]
	cnt := 0
	for cnt < 10 {
		cnt++
		fmt.Println("Test", cnt)
		go func() {
			v, _ := beaconNode.RPCClient.GetAllViewDetail(-1)
			if len(v) > 1 {
				panic("More than 1 view")
			}
			v, _ = shardNode0.RPCClient.GetAllViewDetail(-1)
			if len(v) > 1 {
				panic("More than 1 view")
			}
			v, _ = shardNode1.RPCClient.GetAllViewDetail(-1)
			if len(v) > 1 {
				panic("More than 1 view")
			}
		}()
		go func() {
			v, _ := beaconNode.RPCClient.GetAllViewDetail(0)
			if len(v) > 1 {
				panic("More than 1 view")
			}
			v, _ = shardNode0.RPCClient.GetAllViewDetail(-1)
			if len(v) > 1 {
				panic("More than 1 view")
			}
		}()
		go func() {
			v, _ := beaconNode.RPCClient.GetAllViewDetail(1)
			if len(v) > 1 {
				panic("More than 1 view")
			}
			v, _ = shardNode1.RPCClient.GetAllViewDetail(-1)
			if len(v) > 1 {
				panic("More than 1 view")
			}
		}()
		time.Sleep(time.Second * 10)
	}
	fmt.Println("Done test! ")
}
