package delegation

import (
	"fmt"

	"github.com/incognitochain/incognito-chain/config"
	testsuite "github.com/incognitochain/incognito-chain/testsuite"
	"github.com/incognitochain/incognito-chain/testsuite/account"
	"github.com/pkg/errors"
)

/*
Testsuite to test beacon staking and delegation
*/
func Test_Shard_Cycle_Without_Delegation() error {
	cfg := testsuite.Config{
		DataDir: "./data/",
		Network: testsuite.ID_TESTNET2,
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
		config.Param().ConsensusParam.StakingFlowV4Height = 2
		config.Param().CommitteeSize.MaxShardCommitteeSize = 8
		config.Param().CommitteeSize.MaxBeaconCommitteeSize = 8
		config.Param().CommitteeSize.MinShardCommitteeSize = 4
		config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 4
		config.Param().CommitteeSize.NumberOfFixedShardBlockValidatorV2 = 4
		config.Param().ConsensusParam.ConsensusV2Epoch = 1
		config.Param().EpochParam.NumberOfBlockInEpoch = 20
		config.Param().EpochParam.RandomTime = 10
		config.Param().ConsensusParam.EpochBreakPointSwapNewKey = []uint64{1e9}
		config.Config().LimitFee = 0
		config.Param().PDexParams.Pdexv3BreakPointHeight = 1e9
		config.Param().TxPoolVersion = 0
		config.Param().MaxReward = 100000000000000
	}, func(node *testsuite.NodeEngine) {})
	//send PRV and stake
	stakersShard := []account.Account{}
	delegateList := []string{}
	autoStaking := []bool{}
	prevMap := map[string]*testsuite.AccountInfo{}
	amountPerStaker := []uint64{}
	numberOfSStaker := 3
	for i := 0; i < numberOfSStaker; i++ {
		acc := node.NewAccountFromShard(0)
		stakersShard = append(stakersShard, acc)
		delegateList = append(delegateList, "")
		if i%2 == 0 {
			autoStaking = append(autoStaking, true)
		} else {
			autoStaking = append(autoStaking, false)
		}
		amountPerStaker = append(amountPerStaker, 1e14)
	}
	node.PreparePRVForTest(node.GenesisAccount, stakersShard, amountPerStaker)
	node.StakeNewShards(stakersShard, delegateList, autoStaking)
	startEpoch := node.GetBlockchain().GetBeaconBestState().Epoch
	maxEpoch := 10
	prevView := node.GetBlockchain().GetBeaconBestState()
	for {
		ver := node.GetBlockchain().GetBeaconBestState().CommitteeStateVersion()
		node.GenerateBlock().NextRound()
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		if height%10 == 1 {
			if currentBeaconBlock.GetCurrentEpoch()-startEpoch > uint64(maxEpoch) {
				return errors.Errorf("Something wrong, some staker is not moved to waiting/syncing list")
			}
			fmt.Printf("Consensus version %v\n", ver)
			prevMap = node.GetAccountPosition(stakersShard, prevView)
			node.ShowAccountsInfo(prevMap)
			for _, v := range prevMap {
				if v.Queue == testsuite.SHARD_NORMAL {
					prevView = node.GetBlockchain().GetBeaconBestState()
					continue
				}
			}
			break
		}
		prevView = node.GetBlockchain().GetBeaconBestState()
	}
	fmt.Println("Joining the committee and test vote, all shard validators at position 6 in committee will be slashed")
	node.Pause()
	// flagg := false
	for {

		node.SendFinishSync(stakersShard, 0)
		node.SendFinishSync(stakersShard, 1)
		for i, s := range stakersShard {
			if i > 5 {
				node.SendFeatureStat([]*account.Account{&s}, []string{"TestFeature"})
			}
		}

		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		if height%20 == 1 {
			prevMap = node.GetAccountPosition(stakersShard, prevView)
			node.ShowAccountsInfo(prevMap)
			node.GetStakerInfo(stakersShard)
		}
		prevView = node.GetBlockchain().GetBeaconBestState()
		if epoch > 30 {
			break
		}
		validatorIndex := make(testsuite.ValidatorIndex)
		validatorIndex[1] = []int{6}
		validatorIndex[0] = []int{6}
		node.GenerateBlock(validatorIndex).NextRound()

	}
	node.Pause()
	return nil
}

func Test_Beacon_Cycle_Without_Delegation() error {
	cfg := testsuite.Config{
		DataDir: "./data/",
		Network: testsuite.ID_TESTNET2,
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
		config.Param().ConsensusParam.StakingFlowV4Height = 2
		config.Param().ConsensusParam.RequiredActiveTimes = 4
		config.Param().CommitteeSize.MaxShardCommitteeSize = 8
		config.Param().CommitteeSize.MaxBeaconCommitteeSize = 8
		config.Param().CommitteeSize.MinShardCommitteeSize = 4
		config.Param().CommitteeSize.NumberOfFixedShardBlockValidator = 4
		config.Param().CommitteeSize.NumberOfFixedShardBlockValidatorV2 = 4
		config.Param().ConsensusParam.ConsensusV2Epoch = 1
		config.Param().EpochParam.NumberOfBlockInEpoch = 20
		config.Param().EpochParam.RandomTime = 10
		config.Param().ConsensusParam.EpochBreakPointSwapNewKey = []uint64{1e9}
		config.Config().LimitFee = 0
		config.Param().PDexParams.Pdexv3BreakPointHeight = 1e9
		config.Param().TxPoolVersion = 0
		config.Param().MaxReward = 100000000000000
	}, func(node *testsuite.NodeEngine) {})
	//send PRV and stake
	stakersShard := []account.Account{}
	stakersBeacon := []account.Account{}
	stakedBeacons := map[string]interface{}{}
	delegateList := []string{}
	autoStaking := []bool{}
	prevMap := map[string]*testsuite.AccountInfo{}
	amountPerStaker := []uint64{}
	numberOfSStaker := 25
	for i := 0; i < numberOfSStaker; i++ {
		acc := node.NewAccountFromShard(0)
		stakersShard = append(stakersShard, acc)
		delegateList = append(delegateList, "")
		autoStaking = append(autoStaking, true)
		amountPerStaker = append(amountPerStaker, 900000000000000)
	}
	node.PreparePRVForTest(node.GenesisAccount, stakersShard, amountPerStaker)
	node.GetStakerInfo(stakersShard)
	node.StakeNewShards(stakersShard, delegateList, autoStaking)
	node.GetStakerInfo(stakersShard)
	startEpoch := node.GetBlockchain().GetBeaconBestState().Epoch
	maxEpoch := 10
	prevView := node.GetBlockchain().GetBeaconBestState()
	for _, acc := range stakersShard {
		node.RPC.API_SubmitKey(acc.PrivateKey)
	}
	for {
		ver := node.GetBlockchain().GetBeaconBestState().CommitteeStateVersion()
		node.GenerateBlock().NextRound()
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		if height%10 == 1 {
			node.GetStakerInfo(stakersShard)
			fmt.Printf("Consensus version %v\n", ver)
			if currentBeaconBlock.GetCurrentEpoch()-startEpoch > uint64(maxEpoch) {
				return errors.Errorf("Something wrong, some staker is not moved to waiting/syncing list")
			}
			prevMap = node.GetAccountPosition(stakersShard, prevView)
			node.ShowAccountsInfo(prevMap)
			for _, v := range prevMap {
				if v.Queue == testsuite.SHARD_NORMAL {
					prevView = node.GetBlockchain().GetBeaconBestState()
					continue
				}
			}
			break
		}
		prevView = node.GetBlockchain().GetBeaconBestState()

	}
	fmt.Println("Start staking beacon")
	node.Pause()

	for {

		node.SendFinishSync(stakersShard, 0)
		node.SendFinishSync(stakersShard, 1)
		for i, s := range stakersShard {
			if i > 5 {
				node.SendFeatureStat([]*account.Account{&s}, []string{"TestFeature"})
			}
		}

		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		if height%20 == 1 {
			prevMap = node.GetAccountPosition(stakersShard, prevView)
			node.ShowAccountsInfo(prevMap)
			stakerInfos := node.GetStakerInfo(stakersShard)
			for k, v := range stakerInfos {
				if _, ok := stakedBeacons[k]; (!ok) && (v.HasCredit) {
					txIDs, errs := node.StakeNewBeacons([]account.Account{*prevMap[k].Acc})
					if (errs[0] == nil) && (len(txIDs[0]) > 0) {
						stakedBeacons[k] = nil
					}
					stakersBeacon = append(stakersBeacon, *prevMap[k].Acc)
				}
			}
			node.ShowBeaconCandidateInfo(stakersShard, stakersBeacon, epoch)
		}
		prevView = node.GetBlockchain().GetBeaconBestState()
		if epoch > 26 {
			break
		}
		validatorIndex := make(testsuite.ValidatorIndex)
		node.GenerateBlock(validatorIndex).NextRound()

	}
	for {

		node.SendFinishSync(stakersShard, 0)
		node.SendFinishSync(stakersShard, 1)
		for i, s := range stakersShard {
			if i > 5 {
				node.SendFeatureStat([]*account.Account{&s}, []string{"TestFeature"})
			}
		}

		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()

		if height%20 == 1 {
			if epoch%5 == 0 {
				prevMap = node.GetAccountPosition(stakersShard, prevView)
				node.ShowAccountsInfo(prevMap)
				for _, acc := range stakersBeacon {
					tx, err := node.RPC.AddStake(acc, 1750000000000*5)
					if (err == nil) && (len(tx.TxID) > 0) {
						fmt.Printf("Add staking for acc %v, txID %v\n", acc.Name, tx.TxID)

					} else {
						fmt.Printf("Add staking for acc %v, err %v\n", acc.Name, err)
					}
				}
			}
			node.ShowBeaconCandidateInfo(stakersShard, stakersBeacon, epoch)
		}
		prevView = node.GetBlockchain().GetBeaconBestState()
		if epoch > 36 {
			break
		}

		validatorIndex := make(testsuite.ValidatorIndex)
		node.GenerateBlock(validatorIndex).NextRound()

	}
	unstakedBeacons := map[string]interface{}{}
	for {
		currentBeaconBlock := node.GetBlockchain().BeaconChain.GetBestView().GetBlock()
		height := currentBeaconBlock.GetHeight()
		epoch := currentBeaconBlock.GetCurrentEpoch()
		if height%20 == 1 {
			prevMap = node.GetAccountPosition(stakersShard, prevView)
			node.ShowAccountsInfo(prevMap)
			node.GetStakerInfo(stakersShard)
			for _, acc := range stakersBeacon {
				if _, ok := unstakedBeacons[acc.Name]; !ok {
					txIDs, errs := node.RPC.StopAutoStake(acc)
					if (errs == nil) && (len(txIDs.TxID) > 0) {
						fmt.Printf("Stop auto stake acc %v, txID %v\n", acc.Name, txIDs.TxID)
						unstakedBeacons[acc.Name] = nil
					} else {
						fmt.Printf("Stop auto stake acc %v, err %v\n", acc.Name, errs)
					}
				}
			}
			node.ShowBeaconCandidateInfo(stakersShard, stakersBeacon, epoch)
		}
		prevView = node.GetBlockchain().GetBeaconBestState()
		if epoch > 66 {
			break
		}
		validatorIndex := make(testsuite.ValidatorIndex)
		node.GenerateBlock(validatorIndex).NextRound()

	}
	node.Pause()
	return nil
}
