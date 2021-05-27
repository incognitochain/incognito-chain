package main

import "fmt"

func Test_PortalV4() {
	node := InitSimTestnetv2()
	node.RPC.ShowBalance(node.GenesisAccount)
	node.RPC.API_SubmitKey(node.GenesisAccount.PrivateKey)
	err := node.RPC.API_CreateConvertCoinVer1ToVer2Transaction(node.GenesisAccount.PrivateKey)
	if err != nil {
		fmt.Println(err)
	}
	node.RPC.ShowBalance(node.GenesisAccount)
	mempool, err := node.RPC.Client.GetMempoolInfo()
	fmt.Printf("%+v\n", mempool)
	node.GenerateBlock().NextRound()
	node.GetBlockchain().GetConfig().TxPool.EmptyPool()
	fmt.Printf("%+v\n", mempool)
	node.GenerateBlock().NextRound()
	fmt.Printf("%+v\n", mempool)

	//node.GenerateBlock().NextRound()
	//node.RPC.ShowBalance(node.GenesisAccount)
	//node.GenerateBlock().NextRound()
	//node.RPC.ShowBalance(node.GenesisAccount)
}
