package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func InitSimMainnet() *testsuite.NodeEngine {
	chainParam := testsuite.NewChainParam(testsuite.ID_MAINNET)
	chainParam.ActiveShards = 2
	chainParam.BCHeightBreakPointNewZKP = 1
	chainParam.BeaconHeightBreakPointBurnAddr = 1

	chainParam.StakingFlowV2Height = 1
	chainParam.Epoch = 20
	chainParam.RandomTime = 10
	common.TIMESLOT = chainParam.Timeslot
	node := testsuite.NewStandaloneSimulation("newsim", testsuite.Config{
		ChainParam:       chainParam,
		ConsensusVersion: 3,
	})
	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
	}
	return node
}

func Test_Shard_Stall_v3() {
	node := InitSimMainnet()
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

func Test_CrossShard() {
	node := InitSimMainnet()

	//send PRV same shard, and crossshard
	miner0_1 := node.NewAccountFromShard(0)
	miner1_1 := node.NewAccountFromShard(1)
	node.SendPRV(node.GenesisAccount, miner0_1, 1e14, miner1_1, 2e14)
	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
	}
	if node.ShowBalance(miner0_1)["PRV"] != 1e14 {
		panic("Cannot receive prv")
	}
	if node.ShowBalance(miner1_1)["PRV"] != 2e14 {
		panic("Cannot receive prv")
	}
}
func Test_Swap_v3() {
	node := InitSimMainnet()

	//stake node
	stakers := []account.Account{}
	for i := 0; i < 22; i++ {
		acc := node.NewAccountFromShard(0)
		node.SendPRV(node.GenesisAccount, acc, 1e14)
		node.GenerateBlock().NextRound()
		node.GenerateBlock().NextRound()

		fmt.Println("send", acc.Name)
		stakers = append(stakers, acc)

	}
	for i := 0; i < len(stakers); i++ {
		acc := stakers[i]
		node.RPC.Stake(acc)
		fmt.Println("stake", acc.Name)
		node.GenerateBlock().NextRound()
	}
	node.GenerateBlock().NextRound()

	for {
		height := node.GetBlockchain().BeaconChain.CurrentHeight()
		node.GenerateBlock().NextRound()
		if height%20 == 0 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			fmt.Printf("Shard 0 Height %v %v\n", node.GetBlockchain().GetChain(0).GetBestView().GetHeight(), node.GetBlockchain().GetChain(0).GetBestView().GetBlock().CommitteeFromBlock().String())
			node.ShowAccountPosition(stakers)
		}
	}
}
