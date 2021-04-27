package tests

import (
	"fmt"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"github.com/incognitochain/incognito-chain/testsuite/rpcclient"
	"testing"
)

func Test_Swap(t *testing.T) {
	sim := testsuite.NewStandaloneSimulation("newsim", testsuite.Config{
		ChainParam: testsuite.NewChainParam(testsuite.ID_TESTNET2).SetActiveShardNumber(2).SetMaxShardCommitteeSize(5),
		DisableLog: true,
	})
	sim.GenerateBlock().NextRound()
	miner1 := sim.NewAccountFromShard(0)

	sim.RPC.API_SendTxPRV(sim.GenesisAccount.PrivateKey, map[string]uint64{miner1.PaymentAddress: 1000000}, -1, true)
	for i := 0; i < 4; i++ {
		sim.GenerateBlock().NextRound()
	}

	sim.GenerateBlock().NextRound()
	minerCm1, _ := account.BuildCommitteePubkey(miner1.PrivateKey, sim.GenesisAccount.PaymentAddress)
	stake1 := rpcclient.StakingTxParam{
		Name:         "miner1",
		CommitteeKey: minerCm1,
		BurnAddr:     sim.GetBlockchain().GetBurningAddress(sim.GetBlockchain().BeaconChain.GetFinalViewHeight()),
		StakerPrk:    sim.GenesisAccount.PrivateKey,
		MinerPrk:     miner1.PrivateKey,
		RewardAddr:   miner1.PaymentAddress,
		StakeShard:   true,
		AutoRestake:  true,
	}

	stakeList := []rpcclient.StakingTxParam{stake1}
	_, err := sim.RPC.API_SendTxStaking(stake1)
	if err != nil {
		panic(err)
	}
	monitorPool := func(oldLen1 int, oldLen2 int, oldLen3 int) (bool, int, int, int) {
		len1 := len(sim.GetBlockchain().BeaconChain.GetShardsWaitingList())
		len2 := 0
		len3 := 0
		for _, sCommittee := range sim.GetBlockchain().BeaconChain.GetShardsPendingList()["bls"] {
			len2 += len(sCommittee)
		}
		for _, sCommittee := range sim.GetBlockchain().BeaconChain.GetAllCommittees()[sim.GetBlockchain().BeaconChain.GetConsensusType()] {
			len3 += len(sCommittee)
		}
		if (oldLen1 != len1) || (oldLen2 != len2) || (oldLen3 != len3) {
			return true, len1, len2, len3
		}
		return false, len1, len2, len3
	}

	viewPool := func() {
		waitingPool := []string{}
		pendingPool := []string{}
		committeePool := []string{}
		for _, stake := range stakeList {
			role, _ := sim.RPC.API_GetPublicKeyRole(stake.CommitteeKey.GetMiningKeyBase58("bls"))
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
	for i := 0; i < 100; i++ {
		sim.GenerateBlock().NextRound()
		isChange, l1, l2, l3 = monitorPool(l1, l2, l3)
		if isChange {
			fmt.Println("\n----------------------------------")
			fmt.Println("Beacon Epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
			fmt.Println("Beacon Height", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetHeight())
			viewPool()
			fmt.Println("----------------------------------")
			isChange = false
		}
	}
	// create miner1 unstake information
	unstake1 := rpcclient.StopStakingParam{
		BurnAddr:  sim.GetBlockchain().GetBurningAddress(sim.GetBlockchain().BeaconChain.GetFinalViewHeight()),
		StakerPrk: sim.GenesisAccount.PrivateKey,
		MinerPrk:  miner1.PrivateKey,
	}

	// send stop staking tx
	if _, err := sim.RPC.API_SendTxStopAutoStake(unstake1); err != nil {
		panic(err)
	}
	fmt.Println("Stopstake staker1 at epoch", sim.GetBlockchain().BeaconChain.GetBestView().GetBlock().GetCurrentEpoch())
	// generate block for tx to add to chain
	sim.GenerateBlock().NextRound()

	// create miner2 stake information
	miner2 := sim.NewAccountFromShard(0)
	minerCm2, _ := account.BuildCommitteePubkey(miner2.PrivateKey, sim.GenesisAccount.PaymentAddress)
	stake2 := rpcclient.StakingTxParam{
		Name:         "miner2",
		CommitteeKey: minerCm2,
		BurnAddr:     sim.GetBlockchain().GetBurningAddress(sim.GetBlockchain().BeaconChain.GetFinalViewHeight()),
		StakerPrk:    sim.GenesisAccount.PrivateKey,
		MinerPrk:     miner2.PrivateKey,
		RewardAddr:   miner2.PaymentAddress,
		StakeShard:   true,
		AutoRestake:  true,
	}
	stakeList = append(stakeList, stake2)

	// send miner2 staking information
	_, err = sim.RPC.API_SendTxStaking(stake2)
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
			viewPool()
			fmt.Println("----------------------------------")
			isChange = false
		}
	}

	_, err = sim.RPC.API_SendTxWithdrawReward(miner1.PrivateKey, miner1.PaymentAddress)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 5; i++ {
		sim.GenerateBlock().NextRound()
	}

	bl1, _ := sim.RPC.API_GetBalance(miner1)
	fmt.Println(bl1)
}
