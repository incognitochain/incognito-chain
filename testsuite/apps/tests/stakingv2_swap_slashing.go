package main

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func Test_Stakingv2() {
	node := testsuite.InitChainParam(func() {
		config.Param().ActiveShards = 2
		config.Param().BCHeightBreakPointNewZKP = 1
		config.Param().BCHeightBreakPointPrivacyV2 = 2
		config.Param().BeaconHeightBreakPointBurnAddr = 1
		config.Param().ConsensusParam.EnableSlashingHeightV2 = 2
		config.Param().ConsensusParam.StakingFlowV2Height = 5
		config.Param().ConsensusParam.AssignRuleV3Height = 5
		config.Param().ConsensusParam.ConsensusV2Epoch = 1
		config.Param().EpochParam.NumberOfBlockInEpoch = 20
		config.Param().EpochParam.RandomTime = 10
		config.Config().LimitFee = 0
	})

	//stake node
	stakers := []account.Account{}
	for i := 0; i < 22; i++ {
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
