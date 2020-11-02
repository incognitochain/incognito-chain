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
				bl3, _ := sim.GetBalance(acc2)
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
		ShardNumber: 1,
	})
	sim.GenerateBlock().NextRound()
	staker1 := sim.NewAccountFromShard(0)
	stakerCm1, _ := staker1.BuildCommitteePubkey(staker1.PaymentAddress)
	stake1 := F.StakingTxParam{
		Name:         "staker1",
		CommitteeKey: stakerCm1,
		SenderPrk:    sim.IcoAccount.PrivateKey,
		MinerPrk:     staker1.PrivateKey,
		RewardAddr:   staker1.PaymentAddress,
		StakeShard:   true,
		AutoRestake:  true,
	}

	staker2 := sim.NewAccountFromShard(0)
	stakerCm2, _ := staker2.BuildCommitteePubkey(staker2.PaymentAddress)
	stake2 := F.StakingTxParam{
		Name:         "staker2",
		CommitteeKey: stakerCm2,
		SenderPrk:    sim.IcoAccount.PrivateKey,
		MinerPrk:     staker2.PrivateKey,
		RewardAddr:   staker2.PaymentAddress,
		StakeShard:   true,
		AutoRestake:  true,
	}

	stakeList := []F.StakingTxParam{stake1, stake2}
	_, err := sim.API_CreateTxStaking(stake1)
	if err != nil {
		panic(err)
	}
	monitorPool := func(oldLen1 int, oldLen2 int, oldLen3 int) (bool, int, int, int) {
		len1 := len(sim.GetBlockchain().BeaconChain.GetShardsWaitingList())
		len2 := len(sim.GetBlockchain().BeaconChain.GetShardsPendingList())
		len3 := 0
		for _, sCommittee := range sim.GetBlockchain().BeaconChain.GetAllCommittees()[sim.GetBlockchain().BeaconChain.GetConsensusType()] {
			len3 += len(sCommittee)
		}
		if oldLen1 != len1 || oldLen2 != len2 || oldLen3 != len3 {
			return true, len1, len2, len3
		}
		return false, len1, len2, len3
	}

	viewPool := func() {
		waitingPool := []string{}
		pendingPool := []string{}
		committeePool := []string{}
		for _, stake := range stakeList {
			role, _ := sim.GetPubkeyState(stake.CommitteeKey)
			switch role {
			case "waiting":
				waitingPool = append(waitingPool, stake.Name)
			case "pending":
				pendingPool = append(pendingPool, stake.Name)
			case "committee":
				committeePool = append(committeePool, stake.Name)
			}
		}

		fmt.Println("Waiting Pool:", waitingPool)
		fmt.Println("Pending Pool:", pendingPool)
		fmt.Println("Committee Pool:", committeePool)
	}

	_, l1, l2, l3 := monitorPool(0, 0, 0)
	isChange := false
	for i := 0; i < 40; i++ {
		sim.GenerateBlock().NextRound()
		isChange, l1, l2, l3 = monitorPool(l1, l2, l3)
		if isChange {
			fmt.Println("\n----------------------------------")
			fmt.Println("Beacon Epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
			fmt.Println("Beacon Height", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetHeight())
			sim.GetBlockchain().BeaconChain.GetAllCommittees()
			viewPool()
			fmt.Println("----------------------------------")
			isChange = false
		}
	}

	for i := 0; i < 20; i++ {
		sim.GenerateBlock().NextRound()
		isChange, l1, l2, l3 = monitorPool(l1, l2, l3)
		if isChange {
			fmt.Println("\n----------------------------------")
			fmt.Println("Beacon Epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
			fmt.Println("Beacon Height", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetHeight())
			sim.GetBlockchain().BeaconChain.GetAllCommittees()
			viewPool()
			fmt.Println("----------------------------------")
			isChange = false
		}
	}

	for i := 0; i < 50; i++ {
		sim.GenerateBlock().NextRound()
		isChange, l1, l2, l3 = monitorPool(l1, l2, l3)
		if isChange {
			fmt.Println("\n----------------------------------")
			fmt.Println("Beacon Epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
			fmt.Println("Beacon Height", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetHeight())
			sim.GetBlockchain().BeaconChain.GetAllCommittees()
			viewPool()
			fmt.Println("----------------------------------")
			isChange = false
		}
	}

	// fmt.Println("\n----------------------------------")
	// fmt.Println("Beacon Epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
	// fmt.Println("Beacon Height", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetHeight())
	// sim.GetBlockchain().BeaconChain.GetAllCommittees()
	// viewPool()
	// if result, err := sim.GetRewardAmount(staker1.PaymentAddress); err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("REWARD staker1", result)
	// 	bl1, _ := sim.GetBalance(staker1)
	// 	fmt.Println("BALANCE staker1:", bl1)
	// }
	// fmt.Println("----------------------------------")

	unstake1 := F.StopStakingParam{
		SenderPrk: sim.IcoAccount.PrivateKey,
		MinerPrk:  staker1.PrivateKey,
	}
	if _, err := sim.CreateTxStopAutoStake(unstake1); err != nil {
		panic(err)
	}
	fmt.Println("Stopstake staker1 at epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())

	sim.GenerateBlock().NextRound()

	_, err = sim.API_CreateTxStaking(stake2)
	if err != nil {
		panic(err)
	}
	fmt.Println("Stake staker2 at epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())

	for i := 0; i < 100; i++ {
		sim.GenerateBlock().NextRound()
		isChange, l1, l2, l3 = monitorPool(l1, l2, l3)
		if isChange {
			fmt.Println("\n----------------------------------")
			fmt.Println("Beacon Epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
			fmt.Println("Beacon Height", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetHeight())
			sim.GetBlockchain().BeaconChain.GetAllCommittees()
			viewPool()
			fmt.Println("----------------------------------")
			isChange = false
		}
	}

	_, err = sim.WithdrawReward(staker1.PrivateKey, staker1.PaymentAddress)
	if err != nil {
		panic(err)
	}

	// fmt.Println("\n----------------------------------")
	// fmt.Println("Beacon Epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
	// fmt.Println("Beacon Height", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetHeight())
	// sim.GetBlockchain().BeaconChain.GetAllCommittees()
	// viewPool()
	// if result, err := sim.GetRewardAmount(staker1.PaymentAddress); err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("REWARD staker1", result)
	// 	bl1, _ := sim.GetBalance(staker1)
	// 	fmt.Println("BALANCE staker1:", bl1)
	// }
	// fmt.Println("----------------------------------")

	for i := 0; i < 300; i++ {
		sim.GenerateBlock().NextRound()
		isChange, l1, l2, l3 = monitorPool(l1, l2, l3)
		if isChange {
			fmt.Println("\n----------------------------------")
			fmt.Println("Beacon Epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
			fmt.Println("Beacon Height", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetHeight())
			sim.GetBlockchain().BeaconChain.GetAllCommittees()
			viewPool()
			fmt.Println("----------------------------------")
			isChange = false
		}
	}

	// fmt.Println("\n----------------------------------")
	// fmt.Println("Beacon Epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
	// fmt.Println("Beacon Height", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetHeight())
	// sim.GetBlockchain().BeaconChain.GetAllCommittees()
	// viewPool()
	// if result, err := sim.GetRewardAmount(staker1.PaymentAddress); err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("REWARD staker1", result)
	// 	bl1, _ := sim.GetBalance(staker1)
	// 	fmt.Println("BALANCE staker1:", bl1)
	// }

	// if result, err := sim.GetRewardAmount(staker2.PaymentAddress); err != nil {
	// 	panic(err)
	// } else {
	// 	fmt.Println("REWARD staker2", result)
	// 	bl1, _ := sim.GetBalance(staker2)
	// 	fmt.Println("BALANCE staker2:", bl1)
	// }
	// fmt.Println("----------------------------------")
	fmt.Println()
	return
}
