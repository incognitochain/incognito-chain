package main

import (
	"fmt"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func Test_Swap_v3() {
	chainParam := testsuite.NewChainParam(testsuite.ID_TESTNET2).
		SetActiveShardNumber(2).
		SetMaxShardCommitteeSize(5).
		SetBCHeightBreakPointNewZKP(1)
	chainParam.BeaconHeightBreakPointBurnAddr = 1

	sim := testsuite.NewStandaloneSimulation("newsim", testsuite.Config{
		ChainParam: chainParam,
	})
	sim.GenerateBlock().NextRound()

	//send PRV same shard, and crossshard
	miner0_1 := sim.NewAccountFromShard(0)
	miner1_1 := sim.NewAccountFromShard(1)
	sim.SendPRV(sim.GenesisAccount, miner0_1, 1e14, miner1_1, 2e14)
	for i := 0; i < 4; i++ {
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
	for i := 0; i < 60; i++ {
		acc := sim.NewAccountFromShard(0)
		sim.SendPRV(sim.GenesisAccount, acc, 1e14)
		sim.GenerateBlock().NextRound()
		sim.RPC.Stake(acc)
		sim.GenerateBlock().NextRound()
		stakers = append(stakers, acc)
	}
	sim.GenerateBlock().NextRound()

	for {
		pool, index := sim.TrackAccount(stakers[0])
		fmt.Println(pool, index)
		sim.Pause()
		sim.GenerateBlock().NextRound()
	}

}
