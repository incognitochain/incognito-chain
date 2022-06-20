package main

import (
	"github.com/incognitochain/incognito-chain/testsuite/apps/remotetests"
)

func main() {
	nodeManager := remotetests.NewRemoteNodeManager()
	//NormalScenarioTest(nodeManager)
	Lemma2ScenarioTest(nodeManager)
}
