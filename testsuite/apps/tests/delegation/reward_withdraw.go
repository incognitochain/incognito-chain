package delegation

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func TestRewardWithdraw() error {
	node := InitDelegationTest()

	//stake shard
	stakers := []account.Account{}
	bcPKStructs := node.GetBlockchain().GetBeaconBestState().GetBeaconCommittee()
	bcPKStrs, _ := incognitokey.CommitteeKeyListToString(bcPKStructs)

	for i := 0; i < 3; i++ {
		acc := node.NewAccountFromShard(0)
		node.RPC.API_SubmitKey(acc.PrivateKey)
		node.SendPRV(node.GenesisAccount, acc, 1e14)
		node.GenerateBlock().NextRound()
		tx, err := node.RPC.StakeNew(acc, bcPKStrs[1], true)
		fmt.Printf("Staker %+v create tx stake shard %v, err %+v\n", acc.Name, tx.TxID, err)
		node.GenerateBlock().NextRound()
		stakers = append(stakers, acc)
	}

	for {
		node.GenerateBlock().NextRound()
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		node.SendFinishSync(stakers, 0)
		node.SendFinishSync(stakers, 1)

		if height%20 == 1 {
			res, err := node.RPC.API_GetRewardAmount(node.GetAllAccounts()[0].PaymentAddress)
			fmt.Println(epoch, res, err)
			res, err = node.RPC.API_GetRewardAmount(stakers[0].PaymentAddress)
			fmt.Println(res, err)
			node.Pause()
		}

	}

	return nil
}
