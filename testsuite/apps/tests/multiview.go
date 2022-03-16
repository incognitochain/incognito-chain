package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
)

func printAllView(beaconChain *blockchain.BeaconChain) {
	views := beaconChain.GetAllView()
	bestView := beaconChain.GetBestView().GetHash().String()
	finalView := beaconChain.GetFinalView().GetHash().String()
	for _, v := range views {
		note := ""
		if v.GetHash().String() == bestView {
			note = "(B)"
		}
		if v.GetHash().String() == finalView {
			note = "(F)"
		}
		ts := common.CalculateTimeSlot(v.GetBlock().GetProposeTime())
		fmt.Printf("%v:%v%v:%v -> %v\n", v.GetHeight(), v.GetHash().String(), note, ts, v.GetPreviousHash().String())
	}
	fmt.Printf("=============================\n")
}

func Test_Multiview() {
	cfg := testsuite.Config{
		DataDir: "./data/",
		Network: testsuite.ID_TESTNET2,
		AppNode: false,
		ResetDB: true,
	}

	node := testsuite.InitChainParam(cfg, func() {
		config.Param().ActiveShards = 2
		config.Param().BCHeightBreakPointNewZKP = 1
		config.Param().BCHeightBreakPointPrivacyV2 = 2
		config.Param().BeaconHeightBreakPointBurnAddr = 1
		config.Param().ConsensusParam.EnableSlashingHeightV2 = 2
		config.Param().ConsensusParam.StakingFlowV2Height = 5
		config.Param().EpochParam.NumberOfBlockInEpoch = 20
		config.Param().EpochParam.RandomTime = 10
		config.Config().LimitFee = 0
	})
	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
	}

	node.GenerateFork2Branch(-1, func() {})
	printAllView(node.GetBlockchain().BeaconChain)

}
