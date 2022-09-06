package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/config"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

/*
Testsuite to test beacon staking and delegation
*/
func Test_Shard_Staking_With_Delegation() {
	cfg := testsuite.Config{
		DataDir: "./data/",
		Network: testsuite.ID_TESTNET2,
		ResetDB: true,
	}

	node := testsuite.InitChainParam(cfg, func() {
		config.Param().ActiveShards = 2
		config.Param().BCHeightBreakPointNewZKP = 1
		config.Param().BCHeightBreakPointPrivacyV2 = 2
		config.Param().BeaconHeightBreakPointBurnAddr = 1
		config.Param().ConsensusParam.EnableSlashingHeightV2 = 2
		config.Param().ConsensusParam.StakingFlowV2Height = 5
		config.Param().ConsensusParam.AssignRuleV3Height = 10
		config.Param().ConsensusParam.StakingFlowV3Height = 15
		config.Param().CommitteeSize.MaxShardCommitteeSize = 16
		config.Param().CommitteeSize.MinShardCommitteeSize = 4
		config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 4
		config.Param().ConsensusParam.ConsensusV2Epoch = 1
		config.Param().EpochParam.NumberOfBlockInEpoch = 20
		config.Param().EpochParam.RandomTime = 10
		config.Param().ConsensusParam.EpochBreakPointSwapNewKey = []uint64{1e9}
		config.Config().LimitFee = 0
		config.Param().PDexParams.Pdexv3BreakPointHeight = 1e9
		config.Param().TxPoolVersion = 0
	}, func(node *testsuite.NodeEngine) {})

	//send PRV and stake
	stakers := []account.Account{}
	for i := 0; i < 40; i++ {
		acc := node.NewAccountFromShard(0)
		node.RPC.API_SubmitKey(acc.PrivateKey)
		node.SendPRV(node.GenesisAccount, acc, 1e14)
		node.GenerateBlock().NextRound()
		node.GenerateBlock().NextRound()
		node.RPC.Stake(acc)
		fmt.Println("send PRV and stake", acc.Name)
		node.GenerateBlock().NextRound()
		stakers = append(stakers, acc)
	}

	//check pool status
	for {
		node.GenerateBlock().NextRound()
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		if epoch > 20 {
			break
		}

		if height%20 == 1 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			node.ShowAccountPosition(stakers)
			//TODO: check account
		}
	}

	for {

		node.SendFinishSync(stakers, 0)
		node.SendFinishSync(stakers, 1)
		for i, s := range stakers {
			if i > 5 {
				node.SendFeatureStat([]*account.Account{&s}, []string{"TestFeature"})
			}
		}

		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		if epoch > 26 {
			break
		}

		//epoch := currentBeaconBlock.GetCurrentEpoch()
		if height%20 == 1 || height%20 == 11 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			node.ShowAccountPosition(stakers)

		}
		node.GenerateBlock().NextRound()

	}
	node.Pause()

}
