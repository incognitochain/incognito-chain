package delegation

import (
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/testsuite/account"
)

func TestShardStakingAfterTriggerFeature() error {
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

	checkCorrectness := false

	for {
		node.GenerateBlock().NextRound()
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		node.SendFinishSync(stakers, 0)
		node.SendFinishSync(stakers, 1)

		if epoch > 6 {
			break
		}

		if height%20 == 1 {
			fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
			//TODO: check if shard validator delegation info
			//TODO: check beacon delegation info
			//TODO: check if shard start earning when join committee
			node.ShowAccountPosition(stakers)
			node.Pause()
		}
	}

	if !checkCorrectness {
		return errors.New("Testcase not pass")
	}
	return nil
}
