package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func Test_Swap_v3() {
	chainParam := testsuite.NewChainParam(testsuite.ID_MAINNET)
	chainParam.ActiveShards = 2
	chainParam.BCHeightBreakPointNewZKP = 1
	chainParam.BeaconHeightBreakPointBurnAddr = 1

	chainParam.StakingFlowV2Height = 1
	chainParam.Epoch = 20
	chainParam.RandomTime = 10
	common.TIMESLOT = chainParam.Timeslot
	sim := testsuite.NewStandaloneSimulation("newsim", testsuite.Config{
		ChainParam:       chainParam,
		ConsensusVersion: 3,
	})

	//send PRV same shard, and crossshard
	miner0_1 := sim.NewAccountFromShard(0)
	miner1_1 := sim.NewAccountFromShard(1)
	sim.SendPRV(sim.GenesisAccount, miner0_1, 1e14, miner1_1, 2e14)
	for i := 0; i < 10; i++ {
		sim.GenerateBlock().NextRound()
	}
	if sim.ShowBalance(miner0_1)["PRV"] != 1e14 {
		panic("Cannot receive prv")
	}
	if sim.ShowBalance(miner1_1)["PRV"] != 2e14 {
		panic("Cannot receive prv")
	}

	//stake node
	stakers := []account.Account{}
	for i := 0; i < 50; i++ {
		acc := sim.NewAccountFromShard(0)
		sim.SendPRV(sim.GenesisAccount, acc, 1e14)
		sim.GenerateBlock().NextRound()
		sim.GenerateBlock().NextRound()
		fmt.Println("send", acc.Name, len(sim.GetBlockchain().GetChain(0).GetBestView().GetBlock().(*types.ShardBlock).Body.Transactions))
		stakers = append(stakers, acc)

	}
	for i := 0; i < len(stakers); i++ {
		acc := stakers[i]
		sim.RPC.Stake(acc)
		fmt.Println("stake", acc.Name)
		sim.GenerateBlock().NextRound()
	}
	sim.GenerateBlock().NextRound()

	for {
		height := sim.GetBlockchain().BeaconChain.CurrentHeight()
		sim.GenerateBlock().NextRound()
		if height%20 == 0 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", sim.GetBlockchain().BeaconChain.CurrentHeight(), sim.GetBlockchain().BeaconChain.GetEpoch())
			sim.ShowAccountPosition(stakers)
		}
	}
}
