package testsuite

import (
	"encoding/json"
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
	miner1 := sim.NewAccountFromShard(0)
	minerCm1, _ := miner1.BuildCommitteePubkey(sim.IcoAccount.PaymentAddress)
	stake1 := F.StakingTxParam{
		Name:         "staker1",
		CommitteeKey: minerCm1,
		StakerPrk:    sim.IcoAccount.PrivateKey,
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
	viewPool()
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

func Test_PDEFlow(t *testing.T) {
	F.DisableLog(true)
	sim := F.NewStandaloneSimulation("sim3", F.Config{
		ShardNumber: 1,
	})
	acc1 := sim.NewAccountFromShard(0)
	_, err := sim.API_CreateTransaction(sim.IcoAccount, acc1, 100000000)
	if err != nil {
		panic(err)
	}
	sim.GenerateBlock().NextRound()

	//Create custom token
	result1, err := sim.API_CreateAndSendPrivacyCustomTokenTransaction(sim.IcoAccount.PrivateKey, nil, 5, 1, map[string]interface{}{
		"Privacy":     true,
		"TokenID":     "",
		"TokenName":   "pLAM",
		"TokenSymbol": "LAM",
		"TokenFee":    float64(0),
		"TokenTxType": float64(0),
		"TokenAmount": float64(30000000000),
		"TokenReceivers": map[string]interface{}{
			sim.IcoAccount.PaymentAddress: float64(30000000000),
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(result1.TokenID)
	for i := 0; i < 50; i++ {
		sim.GenerateBlock().NextRound()
	}

	burnAddr := sim.GetBlockchain().GetBurningAddress(sim.GetBlockchain().BeaconChain.GetFinalViewHeight())
	fmt.Println(burnAddr)
	result2, err := sim.API_CreateAndSendTxWithPTokenContributionV2(sim.IcoAccount.PrivateKey, nil, -1, 0, map[string]interface{}{
		"Privacy":     true,
		"TokenID":     result1.TokenID,
		"TokenTxType": float64(1),
		"TokenName":   "",
		"TokenSymbol": "",
		"TokenAmount": "300000000",
		"TokenReceivers": map[string]interface{}{
			burnAddr: "300000000",
		},
		"TokenFee":              "0",
		"PDEContributionPairID": "testPAIR",
		"ContributorAddressStr": sim.IcoAccount.PaymentAddress,
		"ContributedAmount":     "300000000",
		"TokenIDStr":            result1.TokenID,
	})
	if err != nil {
		panic(err)
	}

	r2Bytes, _ := json.Marshal(result2)
	fmt.Println(string(r2Bytes))

	for i := 0; i < 10; i++ {
		sim.GenerateBlock().NextRound()
	}

	_, err = sim.API_CreateAndSendTxWithPRVContributionV2(sim.IcoAccount.PrivateKey, map[string]interface{}{burnAddr: "100000000000"}, -1, 0, map[string]interface{}{
		"PDEContributionPairID": "testPAIR",
		"ContributorAddressStr": sim.IcoAccount.PaymentAddress,
		"ContributedAmount":     "100000000000",
		"TokenIDStr":            "0000000000000000000000000000000000000000000000000000000000000004",
	})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		sim.GenerateBlock().NextRound()
	}

	r, err := sim.API_GetPDEState(float64(sim.GetBlockchain().GetBeaconBestState().BeaconHeight))
	if err != nil {
		panic(err)
	}
	rBytes, _ := json.Marshal(r)
	fmt.Println(string(rBytes))

	_, err = sim.API_CreateAndSendTxWithPRVCrossPoolTradeReq(acc1.PrivateKey, map[string]interface{}{burnAddr: "1000000"}, -1, -1, map[string]interface{}{
		"TokenIDToBuyStr":     result1.TokenID,
		"TokenIDToSellStr":    "0000000000000000000000000000000000000000000000000000000000000004",
		"SellAmount":          "1000000",
		"MinAcceptableAmount": "1",
		"TradingFee":          "0",
		"TraderAddressStr":    acc1.PaymentAddress,
	})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		sim.GenerateBlock().NextRound()
	}

	_, err = sim.API_CreateAndSendTxWithPTokenCrossPoolTradeReq(sim.IcoAccount.PrivateKey, map[string]interface{}{burnAddr: "1"}, -1, 0, map[string]interface{}{
		"Privacy":     true,
		"TokenID":     result1.TokenID,
		"TokenTxType": float64(1),
		"TokenName":   "",
		"TokenSymbol": "",
		"TokenAmount": "1000000000",
		"TokenReceivers": map[string]interface{}{
			burnAddr: "1000000000",
		},
		"TokenFee":            "0",
		"TokenIDToBuyStr":     "0000000000000000000000000000000000000000000000000000000000000004",
		"TokenIDToSellStr":    result1.TokenID,
		"SellAmount":          "1000000000",
		"MinAcceptableAmount": "1",
		"TradingFee":          "1",
		"TraderAddressStr":    sim.IcoAccount.PaymentAddress,
	})
	if err != nil {
		panic(err)
	}
	for i := 0; i < 10; i++ {
		sim.GenerateBlock().NextRound()
	}
	fmt.Println("------------------------------------------------------------")
	bl, _ := sim.GetBalance(sim.IcoAccount)
	fmt.Println("ICO", bl)
	fmt.Println("------------------------------------------------------------")
	bl1, _ := sim.GetBalance(acc1)
	fmt.Println("ACC1", bl1)

	fmt.Println("------------------------------------------------------------")
	r2, err := sim.API_GetPDEState(float64(sim.GetBlockchain().GetBeaconBestState().BeaconHeight))
	if err != nil {
		panic(err)
	}
	rBytes2, _ := json.Marshal(r2)
	fmt.Println(string(rBytes2))
}
