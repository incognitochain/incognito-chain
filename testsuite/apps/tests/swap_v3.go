package main

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func InitSimMainnet() *testsuite.NodeEngine {

	node := testsuite.NewStandaloneSimulation("newsim", testsuite.Config{
		Network: testsuite.ID_MAINNET,
		ResetDB: true,
	})
	config.Param().ActiveShards = 2
	config.Param().BCHeightBreakPointNewZKP = 1
	config.Param().BeaconHeightBreakPointBurnAddr = 2
	config.Param().ConsensusParam.StakingFlowV2Height = 1
	config.Param().EpochParam.NumberOfBlockInEpoch = 20
	config.Param().EpochParam.RandomTime = 10

	node.Init()

	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
	}
	return node
}

func InitSimTestnet() *testsuite.NodeEngine {

	node := testsuite.NewStandaloneSimulation("newsim", testsuite.Config{
		Network: testsuite.ID_TESTNET,
		ResetDB: true,
	})
	config.Param().ActiveShards = 2
	config.Param().BCHeightBreakPointNewZKP = 1
	config.Param().BeaconHeightBreakPointBurnAddr = 1
	config.Param().ConsensusParam.StakingFlowV2Height = 1
	config.Param().EpochParam.NumberOfBlockInEpoch = 20
	config.Param().EpochParam.RandomTime = 10
	node.Init()
	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
	}
	return node
}

func InitSimTestnetv2() *testsuite.NodeEngine {
	node := testsuite.NewStandaloneSimulation("newsim", testsuite.Config{
		Network: testsuite.ID_TESTNET2,
		ResetDB: true,
	})
	config.Param().ActiveShards = 2
	config.Param().BCHeightBreakPointNewZKP = 1
	config.Param().BeaconHeightBreakPointBurnAddr = 2
	config.Param().ConsensusParam.StakingFlowV2Height = 1
	config.Param().EpochParam.NumberOfBlockInEpoch = 20
	config.Param().EpochParam.RandomTime = 10

	node.Init()

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
	node := InitSimTestnetv2()
	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
	}

	node.GenerateFork2Branch(-1, func() {})
	printAllView(node.GetBlockchain().BeaconChain)

}

func Test_CrossShard() {
	node := InitSimTestnetv2()

	//send PRV same shard, and crossshard
	miner0_1 := node.NewAccountFromShard(0)
	miner1_1 := node.NewAccountFromShard(1)
	node.GenerateFork2Branch(0, func() {
		node.SendPRV(node.GenesisAccount, miner0_1, 1e14, miner1_1, 2e14)
	})
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

		fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
		shardIDs := []int{-1}
		shardIDs = append(shardIDs, node.GetBlockchain().GetShardIDs()...)
		consensusStateDB := node.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).GetBeaconConsensusStateDB()
		_, substituteValidator, nextEpochShardCandidate, currentEpochShardCandidate, _, _, _, _, _ := statedb.GetAllCandidateSubstituteCommittee(consensusStateDB, shardIDs)
		str, _ := incognitokey.CommitteeKeyListToString(currentEpochShardCandidate)
		fmt.Println("currentEpochShardCandidate", str)
		str, _ = incognitokey.CommitteeKeyListToString(nextEpochShardCandidate)
		fmt.Println("nextEpochShardCandidate", str)

		substituteValidatorStr := make(map[int][]string)
		for shardID, v := range substituteValidator {
			tempV, _ := incognitokey.CommitteeKeyListToString(v)
			substituteValidatorStr[shardID] = tempV
		}
		fmt.Println("substituteValidator", substituteValidatorStr)
		if node.GetBlockchain().BeaconChain.CurrentHeight() == 71 {
			node.Pause()
		}
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
func Test_PDE() {
	sim := InitSimTestnet()
	acc1 := sim.NewAccountFromShard(0)
	sim.RPC.API_SubmitKey(sim.GenesisAccount.PrivateKey)
	sim.RPC.API_SubmitKey(acc1.PrivateKey)
	sim.GenerateBlock().NextRound()
	sim.GenerateBlock().NextRound()

	_, err := sim.RPC.API_SendTxPRV(sim.GenesisAccount.PrivateKey, map[string]uint64{
		acc1.PaymentAddress: 100000000,
	}, -1, false)
	if err != nil {
		panic(err)
	}
	// for i := 0; i < 2; i++ {
	sim.GenerateBlock().NextRound()
	sim.GenerateBlock().NextRound()
	sim.RPC.ShowBalance(sim.GenesisAccount)
	sim.RPC.ShowBalance(acc1)
	sim.Pause()
	// }
	// blx, _ := sim.RPC.API_GetBalance(acc1)
	// fmt.Println("ACC1", blx)
	// return
	//Create custom token
	result1, err := sim.RPC.API_SendTxCreateCustomToken(sim.GenesisAccount.PrivateKey, sim.GenesisAccount.PaymentAddress, true, "pTest", "TES", 30000000000)
	if err != nil {
		panic(err)
	}
	fmt.Println(result1.TokenName, result1.TokenID)
	for i := 0; i < 50; i++ {
		sim.GenerateBlock().NextRound()
	}

	bl0, _ := sim.RPC.API_GetBalance(sim.GenesisAccount)
	fmt.Println(bl0)

	// burnAddr := sim.GetBlockchain().GetBurningAddress(sim.GetBlockchain().BeaconChain.GetFinalViewHeight())
	// fmt.Println(burnAddr)
	result2, err := sim.RPC.API_SendTxWithPTokenContributionV2(sim.GenesisAccount, result1.TokenID, 300000000, "testPAIR")
	if err != nil {
		panic(err)
	}

	r2Bytes, _ := json.Marshal(result2)
	fmt.Println(string(r2Bytes))

	for i := 0; i < 10; i++ {
		sim.GenerateBlock().NextRound()
	}

	_, err = sim.RPC.API_SendTxWithPRVContributionV2(sim.GenesisAccount, 100000000000, "testPAIR")
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		sim.GenerateBlock().NextRound()
		sim.Pause()
	}

	r, err := sim.RPC.API_GetPDEState(float64(sim.GetBlockchain().GetBeaconBestState().BeaconHeight))
	fmt.Println(r)
	if err != nil {
		panic(err)
	}
	rBytes, _ := json.Marshal(r)
	fmt.Println(string(rBytes))
	fmt.Println("XXXXXXXXXXXXX")
	_, err = sim.RPC.API_SendTxWithPRVCrossPoolTradeReq(acc1, result1.TokenID, 1000000, 1)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		sim.GenerateBlock().NextRound()
		sim.Pause()
	}
	fmt.Println("YYYYYYYYYYYY")
	_, err = sim.RPC.API_SendTxWithPTokenCrossPoolTradeReq(sim.GenesisAccount, result1.TokenID, "0000000000000000000000000000000000000000000000000000000000000004", 1000000000, 1)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		sim.GenerateBlock().NextRound()
	}
	fmt.Println("------------------------------------------------------------")
	bl, _ := sim.RPC.API_GetBalance(sim.GenesisAccount)
	fmt.Println("ICO", bl)
	fmt.Println("------------------------------------------------------------")
	bl1, _ := sim.RPC.API_GetBalance(acc1)
	fmt.Println("ACC1", bl1)

	fmt.Println("------------------------------------------------------------")
	r2, err := sim.RPC.API_GetPDEState(float64(sim.GetBlockchain().GetBeaconBestState().BeaconHeight))
	if err != nil {
		panic(err)
	}
	rBytes2, _ := json.Marshal(r2)
	fmt.Println(string(rBytes2))

}
