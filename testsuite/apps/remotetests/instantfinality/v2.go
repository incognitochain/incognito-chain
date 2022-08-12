package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/consensus_v2/blsbft"
	devframework "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/apps/remotetests"
	"time"
)

func InstantFinalityV2(nodeManager remotetests.NodeManager) {
	fmt.Println("Checking instant finality work")
	//beaconNode := nodeManager.BeaconNode[0]
	//cnt := 0
	//beaconNode.RPCClient.SetConsensusRule(devframework.ConsensusRule{CreateRule: "nocreate"})
	//for cnt < 10 {
	//	cnt++
	//	fmt.Println("Test", cnt)
	//	v, _ := beaconNode.RPCClient.GetAllViewDetail(-1)
	//	if len(v) > 1 {
	//		panic("More than 1 view")
	//	}
	//	fmt.Println("view", v[0].Height, v[0].ProduceTime, v[0].ProposeTime)
	//	time.Sleep(time.Second * 10)
	//}
	//beaconNode.RPCClient.SetConsensusRule(devframework.ConsensusRule{CreateRule: blsbft.CREATE_RULE_NORMAL})

	time.Sleep(time.Second * 10)
	fmt.Println("Checking POLC work")
	nodeManager.BeaconNode[1].RPCClient.SetConsensusRule(devframework.ConsensusRule{VoteRule: "novote"})
	nodeManager.BeaconNode[2].RPCClient.SetConsensusRule(devframework.ConsensusRule{VoteRule: "novote"})
	nodeManager.BeaconNode[3].RPCClient.SetConsensusRule(devframework.ConsensusRule{VoteRule: "novote"})
	cnt := 0
	for cnt < 10 {
		cnt++
		time.Sleep(time.Second * 10)

	}

	nodeManager.BeaconNode[1].RPCClient.SetConsensusRule(devframework.ConsensusRule{VoteRule: blsbft.VOTE_RULE_VOTE})
	nodeManager.BeaconNode[2].RPCClient.SetConsensusRule(devframework.ConsensusRule{VoteRule: blsbft.VOTE_RULE_VOTE})
	nodeManager.BeaconNode[3].RPCClient.SetConsensusRule(devframework.ConsensusRule{VoteRule: blsbft.VOTE_RULE_VOTE})

	fmt.Println("Done test! ")
}
