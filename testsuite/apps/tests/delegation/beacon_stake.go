package delegation

import (
	"fmt"

	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func TestBeaconStaking() {

	node := InitDelegationTest()
	fmt.Println("Genesis account balance")

	//send PRV and stake
	stakers := []account.Account{}
	// bcPKStructs := node.GetBlockchain().GetBeaconBestState().GetBeaconCommittee()
	// bcPKStrs, _ := incognitokey.CommitteeKeyListToString(bcPKStructs)
	for i := 0; i < 3; i++ {
		acc := node.NewAccountFromShard(0)
		node.RPC.API_SubmitKey(acc.PrivateKey)
		node.SendPRV(node.GenesisAccount, acc, 1e14)
		node.GenerateBlock().NextRound()
		node.GenerateBlock().NextRound()
		tx, err := node.RPC.StakeNewBeacon(acc)
		fmt.Printf("Staker %+v create tx stake beacon %v, err %+v\n", acc.Name, tx.TxID, err)
		node.GenerateBlock().NextRound()
		node.GenerateBlock().NextRound()
		fmt.Println("send PRV and stake", acc.Name)
		stakers = append(stakers, acc)
	}

	//check pool status
	for {
		node.GenerateBlock().NextRound()
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		if epoch > 6 {
			break
		}

		if height%20 == 1 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			node.ShowAccountPosition(stakers)
			// if height == 41 {
			// 	txHash, err := node.RPC.ReDelegate(stakers[0], bcPKStrs[10%len(bcPKStrs)])
			// 	fmt.Println(txHash.Base58CheckData)
			// 	fmt.Printf("Staker 0 redelegate to beacon %v; Redelegate tx: %+v, err %+v\n", 10%len(bcPKStrs), txHash.TxID, err)
			// }
			fmt.Println("Account info:")
			// node.ShowAccountStakeInfo(stakers)
			node.ShowBeaconCandidateInfo(stakers)
			// node.Pause()
		}
	}
	node.Pause()
	flagg := false
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
		if (epoch == 20) && (!flagg) {
			flagg = true
			tx, _ := node.RPC.UnStake(stakers[0])
			fmt.Println(tx)
			node.Pause()

		}
		if epoch > 26 {
			break
		}

		//epoch := currentBeaconBlock.GetCurrentEpoch()
		if height%20 == 1 || height%20 == 11 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			node.ShowAccountPosition(stakers)
			node.ShowBeaconCandidateInfo(stakers)
		}

		validatorIndex := make(testsuite.ValidatorIndex)
		validatorIndex[-1] = []int{0, 1, 3}
		node.GenerateBlock(validatorIndex).NextRound()

	}
	node.Pause()

}
