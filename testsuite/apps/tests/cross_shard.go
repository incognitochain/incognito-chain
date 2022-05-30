package main

import (
	"github.com/incognitochain/incognito-chain/config"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
)

func Test_CrossShard() {
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

	//create account
	miner0_1 := node.NewAccountFromShard(0)
	node.RPC.API_SubmitKey(miner0_1.PrivateKey)

	miner1_1 := node.NewAccountFromShard(1)
	node.RPC.API_SubmitKey(miner1_1.PrivateKey)

	//generate fork
	node.GenerateFork2Branch(0, func() {
		node.SendPRV(node.GenesisAccount, miner0_1, 1e14, miner1_1, 2e14)
	})

	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
	}

	node.ShowBalance(node.GenesisAccount)

	if node.ShowBalance(miner0_1)["PRV"] != 1e14 {
		panic("Cannot receive prv")
	}
	if node.ShowBalance(miner1_1)["PRV"] != 2e14 {
		panic("Cannot receive prv")
	}
}
