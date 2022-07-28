package main

import (
	"github.com/incognitochain/incognito-chain/config"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
)

func Test_Shard_Stall_v3() {
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
	node.Pause()
	node.PrintChainInfo([]int{0, 1})
	node.ApplyChain(0).GenerateBlock().NextRound()
	node.ApplyChain(0).GenerateBlock().NextRound()
	node.ApplyChain(0).GenerateBlock().NextRound()
	node.ApplyChain(0).GenerateBlock().NextRound()
	node.PrintChainInfo([]int{0, 1})
	node.Pause()
	node.GenerateBlock().NextRound()
	node.Pause()
	node.GenerateBlock().NextRound()
	node.PrintChainInfo([]int{0, 1})
	for i := 0; i < 100; i++ {
		node.ApplyChain(-1).GenerateBlock().NextRound()

	}
	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
		node.PrintChainInfo([]int{0, 1})
	}
}
