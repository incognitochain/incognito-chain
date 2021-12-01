package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func Test_Auto_Enable() {
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
		config.Param().ConsensusParam.ConsensusV2Epoch = 1
		config.Param().EpochParam.NumberOfBlockInEpoch = 20
		config.Param().EpochParam.RandomTime = 10
		config.Param().ConsensusParam.EpochBreakPointSwapNewKey = []uint64{1e9}
		config.Config().LimitFee = 0
	})

	//stake node
	stakers := []account.Account{}
	for i := 0; i < 40; i++ {
		acc := node.NewAccountFromShard(0)
		node.RPC.API_SubmitKey(acc.PrivateKey)

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
		_, substituteValidator, nextEpochShardCandidate, currentEpochShardCandidate, _, _, syncingValidators, _, _, _ := statedb.GetAllCandidateSubstituteCommittee(consensusStateDB, shardIDs)
		str, _ := incognitokey.CommitteeKeyListToString(currentEpochShardCandidate)
		//fmt.Println("currentEpochShardCandidate", str)
		str, _ = incognitokey.CommitteeKeyListToString(nextEpochShardCandidate)
		_ = str
		//fmt.Println("nextEpochShardCandidate", str)
		substituteValidatorStr := make(map[int][]string)
		syncingValidatorStr := make(map[int][]string)

		//fmt.Println("syncingValidators", syncingValidators)
		for shardID, v := range syncingValidators {
			tempV, _ := incognitokey.CommitteeKeyListToString(v)
			syncingValidatorStr[int(shardID)] = tempV
		}
		for shardID, v := range substituteValidator {
			tempV, _ := incognitokey.CommitteeKeyListToString(v)
			substituteValidatorStr[shardID] = tempV
		}
		//fmt.Println("substituteValidator", substituteValidatorStr)
		//fmt.Println("syncingValidatorStr", syncingValidatorStr)
		if node.GetBlockchain().BeaconChain.CurrentHeight() == 71 {
			node.Pause()
		}
	}

	node.GenerateBlock().NextRound()

	for {
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		if epoch > 20 {
			break
		}

		if height%20 == 1 || height%20 == 11 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			node.ShowAccountPosition(stakers)
		}
		node.GenerateBlock().NextRound()
	}

	for {
		node.SendFinishSync(stakers, 0)
		node.SendFinishSync(stakers, 1)
		for i, s := range stakers {
			if i > 5 {
				node.SendFeatureStat([]account.Account{s}, []string{"TestFeature"})
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
			//fmt.Println(currentBeaconBlock.GetInstructions())
			node.ShowAccountPosition(stakers)
		}
		node.GenerateBlock().NextRound()

		shard0Block := node.GetBlockchain().GetChain(0).(testsuite.Chain).GetBestView().GetBlock().(*types.ShardBlock)
		if shard0Block.Header.BeaconHeight%20 == 1 {
			fmt.Println("shard0Block", shard0Block.Header.BeaconHeight, shard0Block.Body.Transactions, shard0Block.Body.Instructions, shard0Block.Body.CrossTransactions)
		}
	}

	for {
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		config.Param().AutoEnableFeature["TestFeature_2"] = config.AutoEnableFeature{
			1500,
			500,
			80,
		}

		if height > 600 {
			fmt.Println(config.Param().AutoEnableFeature)
			node.SendFeatureStat(stakers, []string{"TestFeature_2", "TestFeature"})
			node.SendFinishSync(stakers, 0)
			node.SendFinishSync(stakers, 1)
		}
		//epoch := currentBeaconBlock.GetCurrentEpoch()
		if height%20 == 1 || height%20 == 11 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			fmt.Println(currentBeaconBlock.GetInstructions())
			node.ShowAccountPosition(stakers)
		}

		node.GenerateBlock().NextRound()
	}
}
