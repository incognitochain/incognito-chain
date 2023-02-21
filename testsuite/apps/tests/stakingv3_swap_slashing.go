package main

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func Test_Stakingv3() {
	cfg := testsuite.Config{
		DataDir: "./data/",
		Network: testsuite.ID_LOCAL,
		ResetDB: true,
	}

	node := testsuite.InitChainParam(cfg, func() {
		config.Param().ActiveShards = 2
		config.Param().BCHeightBreakPointNewZKP = 1
		config.Param().BCHeightBreakPointPrivacyV2 = 2
		config.Param().BeaconHeightBreakPointBurnAddr = 1
		config.Param().ConsensusParam.EnableSlashingHeightV2 = 1
		config.Param().ConsensusParam.StakingFlowV2Height = 1
		config.Param().ConsensusParam.AssignRuleV3Height = 1
		config.Param().ConsensusParam.StakingFlowV3Height = 1
		config.Param().CommitteeSize.MaxShardCommitteeSize = 16
		config.Param().CommitteeSize.MaxBeaconCommitteeSize = 8
		config.Param().CommitteeSize.MinShardCommitteeSize = 8
		config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 8
		config.Param().CommitteeSize.NumberOfFixedShardBlockValidatorV2 = 8
		config.Param().ConsensusParam.ConsensusV2Epoch = 1
		config.Param().EpochParam.NumberOfBlockInEpoch = 10
		config.Param().EpochParam.RandomTime = 5
		config.Param().ConsensusParam.EpochBreakPointSwapNewKey = []uint64{1e9}
		config.Config().LimitFee = 0
		config.Param().PDexParams.Pdexv3BreakPointHeight = 1e9
		config.Param().TxPoolVersion = 0
		config.Param().MaxReward = 100000000000000
	}, func(ne *testsuite.NodeEngine) {})

	node.GenerateBlock().NextRound()
	node.GenerateBlock().NextRound()
	//stake node
	stakers := []account.Account{}
	for i := 0; i < 30; i++ {
		acc := node.NewAccountFromShard(0)
		node.RPC.API_SubmitKey(acc.PrivateKey)
		stakers = append(stakers, acc)
	}
	node.GenerateBlock().NextRound()
	node.GenerateBlock().NextRound()
	x, e := node.RPC.API_GetBalance(node.GenesisAccount)
	fmt.Printf("%+v - %+v\n", x, e)
	accs := []interface{}{}
	accs = append(accs, node.GenesisAccount)
	for i := 0; i < 30; i++ {
		acc := stakers[i]
		accs = append(accs, acc)
		accs = append(accs, 1e14)
		fmt.Println("send", acc.Name)

	}
	result, err := node.SendPRVToMultiAccs(accs)
	node.GenerateBlock().NextRound()
	node.GenerateBlock().NextRound()
	node.GenerateBlock().NextRound()
	node.GenerateBlock().NextRound()
	fmt.Printf("%+v --- %+v\n", result, err)
	node.Pause()
	for i := 0; i < len(stakers); i++ {
		acc := stakers[i]
		node.RPC.Stake(acc)
	}
	for i := 0; i < 10; i++ {
		node.GenerateBlock().NextRound()

		fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
		shardIDs := []int{-1}
		shardIDs = append(shardIDs, node.GetBlockchain().GetShardIDs()...)
		consensusStateDB := node.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).GetBeaconConsensusStateDB()
		_, substituteValidator, nextEpochShardCandidate, currentEpochShardCandidate, _, _, syncingValidators, _, _, _, _ := statedb.GetAllCandidateSubstituteCommittee(consensusStateDB, shardIDs)
		str, _ := incognitokey.CommitteeKeyListToString(currentEpochShardCandidate)
		fmt.Println("currentEpochShardCandidate", str)
		str, _ = incognitokey.CommitteeKeyListToString(nextEpochShardCandidate)
		fmt.Println("nextEpochShardCandidate", str)
		substituteValidatorStr := make(map[int][]string)
		syncingValidatorStr := make(map[int][]string)

		fmt.Println("syncingValidators", syncingValidators)
		for shardID, v := range syncingValidators {
			tempV, _ := incognitokey.CommitteeKeyListToString(v)
			syncingValidatorStr[int(shardID)] = tempV
		}
		for shardID, v := range substituteValidator {
			tempV, _ := incognitokey.CommitteeKeyListToString(v)
			substituteValidatorStr[shardID] = tempV
		}
		fmt.Println("substituteValidator", substituteValidatorStr)
		fmt.Println("syncingValidatorStr", syncingValidatorStr)
		if node.GetBlockchain().BeaconChain.CurrentHeight() >= 51 {
			node.Pause()
		}
	}

	node.GenerateBlock().NextRound()

	for i := 0; i < 10; i++ {
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()

		if height%20 == 1 || height%20 == 11 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			node.ShowAccountPosition(stakers)
		}
		node.GenerateBlock().NextRound()
	}

	for {
		node.SendFinishSync(stakers, 0)
		node.SendFinishSync(stakers, 1)
		node.SendFeatureStat(stakers, []string{})
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		if epoch > 40 {
			break
		}
		//epoch := currentBeaconBlock.GetCurrentEpoch()
		if height%20 == 1 || height%20 == 11 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			fmt.Println(currentBeaconBlock.GetInstructions())
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

		//epoch := currentBeaconBlock.GetCurrentEpoch()
		if height%20 == 1 || height%20 == 11 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			fmt.Println(currentBeaconBlock.GetInstructions())
			node.ShowAccountPosition(stakers)
		}

		committee2 := node.GetBlockchain().BeaconChain.GetAllCommittees()[common.BlsConsensus][common.GetShardChainKey(byte(0))]
		//node.PrintAccountNameFromCPK(committee2)

		node.ApplyChain(-1, 1).GenerateBlock()
		signIndex := testsuite.GenerateCommitteeIndex(len(committee2) - 1)
		valIndex := testsuite.ValidatorIndex{}
		for _, v := range signIndex {
			valIndex = append(valIndex, v)
		}
		// node.Pause()
		//fmt.Println(valIndex, signIndex)
		node.ApplyChain(0).GenerateBlock(valIndex)
		node.NextRound()

		shard0Block := node.GetBlockchain().GetChain(0).(testsuite.Chain).GetBestView().GetBlock().(*types.ShardBlock)
		if shard0Block.Header.BeaconHeight%20 == 1 {
			fmt.Println("shard0Block", shard0Block.Header.BeaconHeight, shard0Block.Body.Transactions, shard0Block.Body.Instructions, shard0Block.Body.CrossTransactions)
		}
	}
}
