package main

//
//import (
//	"encoding/json"
//	"fmt"
//
//	"github.com/incognitochain/incognito-chain/blockchain/types"
//	"github.com/incognitochain/incognito-chain/common"
//	"github.com/incognitochain/incognito-chain/config"
//	"github.com/incognitochain/incognito-chain/rpcserver/jsonresult"
//	testsuite "github.com/incognitochain/incognito-chain/testsuite"
//	"github.com/incognitochain/incognito-chain/testsuite/rpcclient"
//)
//
//func getPdexv3State(node *testsuite.NodeEngine, print bool) *jsonresult.Pdexv3State {
//	res, err := node.RPC.API_GetPDESV3tate(node.GetBlockchain().BeaconChain.GetBestViewHeight())
//	if err != nil {
//		panic(err)
//	}
//	if print {
//		b, _ := json.MarshalIndent(res, "", "\t")
//		fmt.Printf("%+v", string(b))
//	}
//	return res
//}
//
//func generateBlock(node *testsuite.NodeEngine, blkNum int) {
//	for i := 0; i < blkNum; i++ {
//		node.GenerateBlock().NextRound()
//	}
//}
//func Test_PDE_v3() {
//	cfg := testsuite.Config{
//		DataDir: "./data/",
//		Network: testsuite.ID_TESTNET2,
//		ResetDB: true,
//	}
//
//	node := testsuite.InitChainParam(cfg, func() {
//		config.Param().ActiveShards = 2
//		config.Param().BCHeightBreakPointNewZKP = 1
//		config.Param().BCHeightBreakPointPrivacyV2 = 2
//		config.Param().BeaconHeightBreakPointBurnAddr = 1
//		config.Param().ConsensusParam.EnableSlashingHeightV2 = 2
//		config.Param().ConsensusParam.StakingFlowV2Height = 5
//		config.Param().EpochParam.NumberOfBlockInEpoch = 20
//		config.Param().EpochParam.RandomTime = 10
//		config.Config().LimitFee = 0
//		config.Param().PDexParams.Pdexv3BreakPointHeight = 2
//
//		//init pdex param
//		config.Param().PDexParams.Pdexv3BreakPointHeight = 2
//		config.Param().PDexParams.ProtocolFundAddress = "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN"
//		config.Param().PDexParams.AdminAddress = "12svfkP6w5UDJDSCwqH978PvqiqBxKmUnA9em9yAYWYJVRv7wuXY1qhhYpPAm4BDz2mLbFrRmdK3yRhnTqJCZXKHUmoi7NV83HCH2YFpctHNaDdkSiQshsjw2UFUuwdEvcidgaKmF3VJpY5f8RdN"
//		config.Param().PDexParams.Params.DefaultFeeRateBPS = 30
//		config.Param().PDexParams.Params.PRVDiscountPercent = 25
//		config.Param().PDexParams.Params.TradingProtocolFeePercent = 0
//		config.Param().PDexParams.Params.TradingStakingPoolRewardPercent = 10
//		config.Param().PDexParams.Params.StakingPoolsShare = map[string]uint{
//			"0000000000000000000000000000000000000000000000000000000000000004": 10,
//			"0000000000000000000000000000000000000000000000000000000000000006": 20,
//		}
//		config.Param().PDexParams.Params.MintNftRequireAmount = 1000000000
//		config.Param().PDexParams.Params.MaxOrdersPerNft = 10
//		config.Param().PDexParams.Params.AutoWithdrawOrderLimitAmount = 10
//	})
//
//	acc1 := node.NewAccountFromShard(0)
//	node.RPC.API_SubmitKey(node.GenesisAccount.PrivateKey)
//	node.RPC.API_SubmitKey(acc1.PrivateKey)
//	node.GenerateBlock().NextRound()
//	node.RPC.API_CreateConvertCoinVer1ToVer2Transaction(node.GenesisAccount.PrivateKey)
//	node.GenerateBlock().NextRound()
//	node.GenerateBlock().NextRound()
//	node.GenerateBlock().NextRound()
//	node.EmptyPool()
//	node.SendPRV(node.GenesisAccount, acc1, 1e9*1e6)
//	node.GenerateBlock().NextRound()
//	node.GenerateBlock().NextRound()
//	node.EmptyPool()
//	//init pToken
//	initTokenResponse, err := node.RPC.API_CreateAndSendTokenInitTransaction(node.GenesisAccount.PrivateKey, "pETH", "pETH", 100000000000000)
//	if err != nil {
//		panic(err)
//	}
//	newTokenID := initTokenResponse.TokenID
//	generateBlock(node, 3)
//
//	//Mint nft
//	err = node.RPC.API_MintNFT(node.GenesisAccount.PrivateKey)
//	if err != nil {
//		panic(err)
//	}
//	generateBlock(node, 5)
//	pdexState := getPdexv3State(node, false)
//
//	//add pool
//	nftID := ""
//	for k, _ := range *pdexState.NftIDs {
//		nftID = k
//		break
//	}
//	err = node.RPC.API_PdexV3AddLiquididty(node.GenesisAccount.PrivateKey, nftID, common.PRVCoinID.String(), "", "pair_hash", "10000", "20000")
//	if err != nil {
//		panic(err)
//	}
//	generateBlock(node, 5)
//	//getPdexv3State(node, true)
//
//	//contribute
//	err = node.RPC.API_PdexV3AddLiquididty(node.GenesisAccount.PrivateKey, nftID, newTokenID, "", "pair_hash", "40000", "20000")
//	if err != nil {
//		panic(err)
//	}
//	generateBlock(node, 4)
//	//getPdexv3State(node, true)
//
//	//add order
//	pdexState = getPdexv3State(node, false)
//	poolPairID := ""
//	for k, _ := range *pdexState.PoolPairs {
//		poolPairID = k
//		break
//	}
//
//	for i := 0; i < 20; i++ {
//		err = node.RPC.API_Pdexv3AddOrder(
//			node.GenesisAccount.PrivateKey, poolPairID, newTokenID, common.PRVIDStr, nftID, "40", "1",
//		)
//		if err != nil {
//			panic(err)
//		}
//		generateBlock(node, 3)
//	}
//	getPdexv3State(node, true)
//	//node.Pause()
//	//trade
//	err = node.RPC.API_Pdexv3Trade(acc1.PrivateKey, rpcclient.PdexV3TradeParam{
//		TradePath:           []string{poolPairID},
//		TokenToSell:         common.PRVIDStr,
//		TokenToBuy:          newTokenID,
//		SellAmount:          "4000",
//		MinAcceptableAmount: "100",
//		TradingFee:          10,
//		FeeInPRV:            true,
//	})
//	if err != nil {
//		panic(err)
//	}
//
//	for i := 0; i < 5; i++ {
//		generateBlock(node, 1)
//		blk := node.GetBlockchain().BeaconChain.GetBestView().GetBlock().(*types.BeaconBlock)
//		fmt.Println("Beacon instruction:", blk.GetHeight(), blk.Body.Instructions)
//
//		sblk := node.GetBlockchain().ShardChain[0].GetBestView().GetBlock().(*types.ShardBlock)
//		fmt.Println("Shard Instruction:", sblk.Header.BeaconHeight, sblk.Body.Transactions)
//
//		fmt.Println("====================")
//	}
//	node.Pause()
//	getPdexv3State(node, true)
//	node.ShowBalance(acc1)
//}
