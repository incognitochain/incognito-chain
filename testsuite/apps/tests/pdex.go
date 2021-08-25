package main

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/config"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
)

func Test_PDE() {
	node := testsuite.InitChainParam(func() {
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

	acc1 := node.NewAccountFromShard(0)
	node.RPC.API_SubmitKey(node.GenesisAccount.PrivateKey)
	node.RPC.API_SubmitKey(acc1.PrivateKey)
	node.GenerateBlock().NextRound()
	node.GenerateBlock().NextRound()

	_, err := node.RPC.API_SendTxPRV(node.GenesisAccount.PrivateKey, map[string]uint64{
		acc1.PaymentAddress: 100000000,
	}, -1, false)
	if err != nil {
		panic(err)
	}
	// for i := 0; i < 2; i++ {
	node.GenerateBlock().NextRound()
	node.GenerateBlock().NextRound()
	node.RPC.ShowBalance(node.GenesisAccount)
	node.RPC.ShowBalance(acc1)

	// }
	// blx, _ := sim.RPC.API_GetBalance(acc1)
	// fmt.Println("ACC1", blx)
	// return
	//Create custom token
	result1, err := node.RPC.API_SendTxCreateCustomToken(node.GenesisAccount.PrivateKey, node.GenesisAccount.PaymentAddress, true, "pTest", "TES", 30000000000)
	if err != nil {
		panic(err)
	}
	fmt.Println(result1.TokenName, result1.TokenID)
	for i := 0; i < 50; i++ {
		node.GenerateBlock().NextRound()
	}

	bl0, _ := node.RPC.API_GetBalance(node.GenesisAccount)
	fmt.Println(bl0)

	// burnAddr := sim.GetBlockchain().GetBurningAddress(sim.GetBlockchain().BeaconChain.GetFinalViewHeight())
	// fmt.Println(burnAddr)
	result2, err := node.RPC.API_SendTxWithPTokenContributionV2(node.GenesisAccount, result1.TokenID, 300000000, "testPAIR")
	if err != nil {
		panic(err)
	}

	r2Bytes, _ := json.Marshal(result2)
	fmt.Println(string(r2Bytes))

	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
	}

	_, err = node.RPC.API_SendTxWithPRVContributionV2(node.GenesisAccount, 100000000000, "testPAIR")
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
		node.Pause()
	}

	r, err := node.RPC.API_GetPDEState(float64(node.GetBlockchain().GetBeaconBestState().BeaconHeight))
	fmt.Println(r)
	if err != nil {
		panic(err)
	}
	rBytes, _ := json.Marshal(r)
	fmt.Println(string(rBytes))
	fmt.Println("XXXXXXXXXXXXX")
	_, err = node.RPC.API_SendTxWithPRVCrossPoolTradeReq(acc1, result1.TokenID, 1000000, 1)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
		node.Pause()
	}
	fmt.Println("YYYYYYYYYYYY")
	_, err = node.RPC.API_SendTxWithPTokenCrossPoolTradeReq(node.GenesisAccount, result1.TokenID, "0000000000000000000000000000000000000000000000000000000000000004", 1000000000, 1)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()
	}
	fmt.Println("------------------------------------------------------------")
	bl, _ := node.RPC.API_GetBalance(node.GenesisAccount)
	fmt.Println("ICO", bl)
	fmt.Println("------------------------------------------------------------")
	bl1, _ := node.RPC.API_GetBalance(acc1)
	fmt.Println("ACC1", bl1)

	fmt.Println("------------------------------------------------------------")
	r2, err := node.RPC.API_GetPDEState(float64(node.GetBlockchain().GetBeaconBestState().BeaconHeight))
	if err != nil {
		panic(err)
	}
	rBytes2, _ := json.Marshal(r2)
	fmt.Println(string(rBytes2))

}
