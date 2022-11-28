package delegation

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"github.com/pkg/errors"
)

func TestShardStakingWithDelegation() error {
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
		epoch := currentBeaconBlock.GetCurrentEpoch()
		node.SendFinishSync(stakers, 0)
		node.SendFinishSync(stakers, 1)

		if epoch == 6 {
			beaconDelegateInfo := node.GetBlockchain().GetBeaconBestState().GetCommitteeState().GetDelegateState()
			if len(beaconDelegateInfo) != 4 {
				return NewTestError(TestShardStakingWithDelegationError,
					fmt.Errorf("Miss or unnecessary record in beacon delegate info. Number of record: got %v expect %v", len(beaconDelegateInfo), 4))
			}

			for _, staker := range stakers {
				if _, ok := beaconDelegateInfo[bcPKStrs[1]].CurrentDelegatorsDetails[staker.SelfCommitteePubkey]; !ok {
					return NewTestError(TestShardStakingWithDelegationError,
						fmt.Errorf("Miss delegator in beacon delegate info"))
				}
			}

			for i := 0; i <= len(bcPKStrs); i++ {
				if i == 1 {
					if len(beaconDelegateInfo[bcPKStrs[i]].CurrentDelegatorsDetails) != 3 || beaconDelegateInfo[bcPKStrs[i]].CurrentDelegators != 3 {
						return NewTestError(TestShardStakingWithReDelegationError,
							fmt.Errorf("Not enough delegator in beacon delegate info"))
					}
				} else {
					if len(beaconDelegateInfo[bcPKStrs[i]].CurrentDelegatorsDetails) != 0 || beaconDelegateInfo[bcPKStrs[i]].CurrentDelegators != 0 {
						return NewTestError(TestShardStakingWithReDelegationError,
							fmt.Errorf(" beacon delegate info error"))
					}
				}
			}
			shardDelegateInfo := node.GetBlockchain().GetBeaconBestState().GetCommitteeState().GetDelegate()
			//TODO: beacon candidate should not be in this list
			if len(shardDelegateInfo) != 11 {
				return NewTestError(TestShardStakingWithDelegationError,
					fmt.Errorf("Miss or unnecessary record in shard delegate info. Number of record: got %v expect %v", len(shardDelegateInfo), 11))
			}

			for shardValidator, delegate := range shardDelegateInfo {
				inStaker := false
				for _, staker := range stakers {
					if staker.SelfCommitteePubkey == shardValidator {
						inStaker = true
						if delegate != bcPKStrs[1] {
							return NewTestError(TestShardStakingWithDelegationError,
								errors.New("Delegate info is not correct! Mismatch delegate info"))
						}
					}
				}
				if !inStaker {
					if delegate != "" {
						return NewTestError(TestShardStakingWithDelegationError,
							errors.New("Delegate info is not correct! Expect empty"))
					}
				}
			}
			break
		}

		//if height%20 == 1 {
		//	fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
		//}
	}

	return nil
}

func TestDelegationAfterStake() error {
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
		tx, err := node.RPC.StakeNew(acc, "", true)
		fmt.Printf("Staker %+v create tx stake shard %v, err %+v\n", acc.Name, tx.TxID, err)
		node.GenerateBlock().NextRound()
		stakers = append(stakers, acc)
	}

	for {
		node.GenerateBlock().NextRound()
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		node.SendFinishSync(stakers, 0)
		node.SendFinishSync(stakers, 1)

		if epoch == 4 {
			beaconDelegateInfo := node.GetBlockchain().GetBeaconBestState().GetCommitteeState().GetDelegateState()
			shardDelegateInfo := node.GetBlockchain().GetBeaconBestState().GetCommitteeState().GetDelegate()
			for _, acc := range stakers {
				tx, err := node.RPC.ReDelegate(acc, bcPKStrs[1])
				fmt.Printf("Staker %+v create tx stake shard %v, err %+v\n", acc.Name, tx.TxID, err)
			}
			for _, info := range beaconDelegateInfo {
				if len(info.CurrentDelegatorsDetails) != 0 || info.CurrentDelegators != 0 {
					return NewTestError(TestDelegationAfterStakeError, fmt.Errorf("Beacon Delegation Info is not correct"))
				}
			}

			for _, info := range shardDelegateInfo {
				if info != "" {
					return NewTestError(TestDelegationAfterStakeError, fmt.Errorf("Shard Delegate Info is not correct"))
				}
			}

			node.GenerateBlock().NextRound()
			node.GenerateBlock().NextRound()

			//TODO: how to check shard next delegate

		}
		if epoch == 5 {
			beaconDelegateInfo := node.GetBlockchain().GetBeaconBestState().GetCommitteeState().GetDelegateState()
			if len(beaconDelegateInfo) != 4 {
				return NewTestError(TestShardStakingWithDelegationError,
					fmt.Errorf("Miss or unnecessary record in beacon delegate info. Number of record: got %v expect %v", len(beaconDelegateInfo), 4))
			}
			for _, staker := range stakers {
				if _, ok := beaconDelegateInfo[bcPKStrs[1]].CurrentDelegatorsDetails[staker.SelfCommitteePubkey]; !ok {
					return NewTestError(TestShardStakingWithDelegationError,
						fmt.Errorf("Miss delegator in beacon delegate info"))
				}
			}

			for i := 0; i <= len(bcPKStrs); i++ {
				if i == 2 {
					if len(beaconDelegateInfo[bcPKStrs[i]].CurrentDelegatorsDetails) != 3 || beaconDelegateInfo[bcPKStrs[i]].CurrentDelegators != 3 {
						return NewTestError(TestShardStakingWithDelegationError,
							fmt.Errorf("Not enough delegator in beacon delegate info"))
					}
				} else {
					if len(beaconDelegateInfo[bcPKStrs[i]].CurrentDelegatorsDetails) != 0 || beaconDelegateInfo[bcPKStrs[i]].CurrentDelegators != 0 {
						return NewTestError(TestShardStakingWithDelegationError,
							fmt.Errorf(" beacon delegate info error"))
					}
				}
			}

			shardDelegateInfo := node.GetBlockchain().GetBeaconBestState().GetCommitteeState().GetDelegate()
			if len(shardDelegateInfo) != 11 {
				return NewTestError(TestShardStakingWithDelegationError,
					fmt.Errorf("Miss or unnecessary record in shard delegate info. Number of record: got %v expect %v", len(shardDelegateInfo), 11))
			}

			for shardValidator, delegate := range shardDelegateInfo {
				inStaker := false
				for _, staker := range stakers {
					if staker.SelfCommitteePubkey == shardValidator {
						inStaker = true
						if delegate != bcPKStrs[1] {
							return NewTestError(TestShardStakingWithDelegationError,
								errors.New("Delegate info is not correct! Mismatch delegate info"))
						}
					}
				}
				if !inStaker {
					if delegate != "" {
						return NewTestError(TestShardStakingWithDelegationError,
							errors.New("Delegate info is not correct! Expect empty"))
					}
				}
			}
			break
		}

		//if height%20 == 1 {
		//	fmt.Printf("\n======================================\nBeacon Height %v Epoch %v \n", node.GetBlockchain().BeaconChain.CurrentHeight(), node.GetBlockchain().BeaconChain.GetEpoch())
		//}
	}

	return nil
}
