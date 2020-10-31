package testsuite

import (
	"fmt"
	"testing"

	"github.com/incognitochain/incognito-chain/blockchain"
	"github.com/incognitochain/incognito-chain/common"
	F "github.com/incognitochain/incognito-chain/devframework"
	"github.com/incognitochain/incognito-chain/transaction"
)

func Test_SendTX(t *testing.T) {
	sim := F.NewStandaloneSimulation("sim1", F.Config{
		ShardNumber: 2,
	})
	sim.GenerateBlock().NextRound()
	acc1 := sim.NewAccountFromShard(1)
	acc2 := sim.NewAccountFromShard(0)
	_, err := sim.API_CreateTransaction(sim.IcoAccount, acc1, 1000, acc2, 3000)
	if err != nil {
		panic(err)
	}
	sim.GenerateBlock(F.Hook{
		//Create: func(chainID int, doCreate func(time time.Time) (common.BlockInterface, error)) {
		//
		//	doCreate(time.Now())
		//},
		Validation: func(chainID int, block common.BlockInterface, doValidation func(common.BlockInterface) error) {
			fmt.Println("PreValidation block", 0)
			if chainID == 0 && block.GetHeight() == 3 {
				newShardBlock := block.(*blockchain.ShardBlock)
				newShardBlock.Body.Transactions = append(newShardBlock.Body.Transactions, newShardBlock.Body.Transactions[0])
				totalTxsFee := make(map[common.Hash]uint64)
				for _, tx := range newShardBlock.Body.Transactions {
					totalTxsFee[*tx.GetTokenID()] += tx.GetTxFee()
					txType := tx.GetType()
					if txType == common.TxCustomTokenPrivacyType {
						txCustomPrivacy := tx.(*transaction.TxCustomTokenPrivacy)
						totalTxsFee[*txCustomPrivacy.GetTokenID()] = txCustomPrivacy.GetTxFeeToken()
					}
				}
				newShardBlock.Header.TotalTxsFee = totalTxsFee
				merkleRoots := blockchain.Merkle{}.BuildMerkleTreeStore(newShardBlock.Body.Transactions)
				merkleRoot := &common.Hash{}
				if len(merkleRoots) > 0 {
					merkleRoot = merkleRoots[len(merkleRoots)-1]
				}
				crossTransactionRoot, err := blockchain.CreateMerkleCrossTransaction(newShardBlock.Body.CrossTransactions)
				if err != nil {
					fmt.Println(err)
				}
				_, shardTxMerkleData := blockchain.CreateShardTxRoot(newShardBlock.Body.Transactions)
				newShardBlock.Header.TxRoot = *merkleRoot
				newShardBlock.Header.ShardTxRoot = shardTxMerkleData[len(shardTxMerkleData)-1]
				newShardBlock.Header.CrossTransactionRoot = *crossTransactionRoot
			}

			err = doValidation(block)
			fmt.Println("PostValidation block", 0, err)
		},
		Insert: func(chainID int, block common.BlockInterface, doInsert func(common.BlockInterface) error) {
			doInsert(block)
			if chainID == 0 {
				bl1, _ := sim.GetBalance(sim.IcoAccount)
				fmt.Println(bl1)
				bl2, _ := sim.GetBalance(acc1)
				fmt.Println(bl2)
				bl3, _ := sim.GetBalance(staker2)
				fmt.Println(bl3)
				fmt.Printf("%+v", block.(*blockchain.ShardBlock).Body)

			}

		},
	})
	sim.NextRound()

	for i := 0; i < 10; i++ {
		sim.GenerateBlock(F.Hook{
			Insert: func(chainID int, block common.BlockInterface, doInsert func(common.BlockInterface) error) {
				if chainID == -1 {
					fmt.Printf("%+v %+v", block.(*blockchain.BeaconBlock).GetHeight(), block.(*blockchain.BeaconBlock).Body.ShardState)

					doInsert(block)

				} else {
					doInsert(block)
				}

			},
		}).NextRound()
	}

	balance, _ := sim.GetBalance(acc1)
	fmt.Printf("%+v", balance)

}

func Test_StakeFlow1(t *testing.T) {
	F.DisableLog(true)
	sim := F.NewStandaloneSimulation("sim2", F.Config{
		ShardNumber: 2,
	})
	sim.GenerateBlock().NextRound()
	staker1 := sim.NewAccountFromShard(1)
	staker2 := sim.NewAccountFromShard(0)

	stake1 := F.StakingTxParam{
		SenderPrk:   sim.IcoAccount.PrivateKey,
		MinerPrk:    staker1.PrivateKey,
		RewardAddr:  staker1.PaymentAddress,
		StakeShard:  true,
		AutoRestake: true,
	}
	stake2 := F.StakingTxParam{
		SenderPrk:   sim.IcoAccount.PrivateKey,
		MinerPrk:    staker2.PrivateKey,
		RewardAddr:  staker2.PaymentAddress,
		StakeShard:  true,
		AutoRestake: true,
	}
	_, err := sim.CreateTxStaking(stake1)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 120; i++ {
		sim.GenerateBlock().NextRound()
	}
	_, err = sim.CreateTxStaking(stake2)
	if err != nil {
		panic(err)
	}
	fmt.Println("----------------------------------")
	// fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).CandidateShardWaitingForCurrentRandom)
	// fmt.Println()
	// fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).CandidateShardWaitingForNextRandom)
	// fmt.Println()
	// fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).ShardPendingValidator)
	// fmt.Println()
	fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
	fmt.Println()
	fmt.Println(len(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).ShardCommittee[0]))
	fmt.Println()
	fmt.Println(len(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).ShardCommittee[1]))
	fmt.Println("----------------------------------")

	acc3 := sim.NewAccountFromShard(1)
	acc4 := sim.NewAccountFromShard(0)
	_, err = sim.CreateTransaction(sim.IcoAccount, acc3, 1000000, acc4, 3000000, staker1, 10000)
	if err != nil {
		panic(err)
	}
	sim.GenerateBlock().NextRound()
	sim.GenerateBlock().NextRound()

	bl1, _ := sim.GetBalance(staker1)
	fmt.Println("staker1 bl:", bl1)
	_, err = sim.CreateTransaction(acc4, acc3, 100000, sim.IcoAccount, 100000)
	if err != nil {
		panic(err)
	}
	for i := 0; i < 100; i++ {
		sim.GenerateBlock().NextRound()
	}

	if result, err := sim.GetRewardAmount(staker1.PaymentAddress); err != nil {
		panic(err)
	} else {
		fmt.Println("staker1", result)
	}

	unstake1 := F.StopStakingParam{
		SenderPrk: sim.IcoAccount.PrivateKey,
		MinerPrk:  staker1.PrivateKey,
	}

	if result, err := sim.CreateTxStopAutoStake(unstake1); err != nil {
		panic(err)
	} else {
		fmt.Println(result)
	}
	for i := 0; i < 100; i++ {
		sim.GenerateBlock().NextRound()
	}
	_, err = sim.WithdrawReward(staker1.PrivateKey, staker1.PaymentAddress)
	if err != nil {
		panic(err)
	}
	fmt.Println("----------------------------------")
	// fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).CandidateShardWaitingForCurrentRandom)
	// fmt.Println()
	// fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).CandidateShardWaitingForNextRandom)
	// fmt.Println()
	// fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).ShardPendingValidator)
	// fmt.Println()
	fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
	fmt.Println()
	fmt.Println(len(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).ShardCommittee[0]))
	fmt.Println()
	fmt.Println(len(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).ShardCommittee[1]))

	fmt.Println("----------------------------------")
	for i := 0; i < 100; i++ {
		sim.GenerateBlock().NextRound()
	}

	if result, err := sim.GetRewardAmount(staker1.PaymentAddress); err != nil {
		panic(err)
	} else {
		fmt.Println("staker1", result)
	}
	if result, err := sim.GetRewardAmount(staker2.PaymentAddress); err != nil {
		panic(err)
	} else {
		fmt.Println("staker2", result)
	}

	bl1, _ = sim.GetBalance(staker1)
	fmt.Println("staker1 bl:", bl1)

	fmt.Println("----------------------------------")
	// fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).CandidateShardWaitingForCurrentRandom)
	// fmt.Println()
	// fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).CandidateShardWaitingForNextRandom)
	// fmt.Println()
	// fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).ShardPendingValidator)
	// fmt.Println()
	fmt.Println(sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
	fmt.Println()
	fmt.Println(len(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).ShardCommittee[0]))
	fmt.Println()
	fmt.Println(len(sim.GetBlockchain().BeaconChain.GetBestView().(*blockchain.BeaconBestState).ShardCommittee[1]))
	fmt.Println("----------------------------------")

	return
}
