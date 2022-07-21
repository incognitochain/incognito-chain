package committeestate

import (
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
)

func GetRewardSplitRule(blockVersion int) SplitRewardRuleProcessor {
	if blockVersion >= types.INSTANT_FINALITY_VERSION_V2 {
		return RewardSplitRuleV2{}
	}
	if blockVersion >= types.BLOCK_PRODUCINGV3_VERSION {
		return RewardSplitRuleV3{}
	}

	return RewardSplitRuleV1{}
}

type RewardSplitRuleV1 struct{}

func (r RewardSplitRuleV1) SplitReward(env *SplitRewardEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error) {
	devPercent := uint64(env.DAOPercent)
	allCoinTotalReward := env.TotalReward
	rewardForBeacon := map[common.Hash]uint64{}
	rewardForShard := map[common.Hash]uint64{}
	rewardForIncDAO := map[common.Hash]uint64{}
	rewardForCustodian := map[common.Hash]uint64{}

	if len(allCoinTotalReward) == 0 {
		Logger.log.Info("Beacon Height %+v, 😭 found NO reward", env.BeaconHeight)
		return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
	}

	for key, totalReward := range allCoinTotalReward {
		rewardForBeacon[key] += 2 * ((100 - devPercent) * totalReward) / ((uint64(env.ActiveShards) + 2) * 100)
		totalRewardForDAOAndCustodians := uint64(devPercent) * totalReward / uint64(100)

		Logger.log.Infof("[test-salary] totalRewardForDAOAndCustodians tokenID %v - %v\n",
			key.String(), totalRewardForDAOAndCustodians)

		if env.IsSplitRewardForCustodian {
			rewardForCustodian[key] += uint64(env.PercentCustodianReward) * totalRewardForDAOAndCustodians / uint64(100)
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians - rewardForCustodian[key]
		} else {
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians
		}

		rewardForShard[key] = totalReward - (rewardForBeacon[key] + totalRewardForDAOAndCustodians)
	}

	return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
}

type RewardSplitRuleV2 struct{}

func (r RewardSplitRuleV2) SplitReward(env *SplitRewardEnvironment) (map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error) {
	devPercent := uint64(env.DAOPercent)
	allCoinTotalReward := env.TotalReward
	rewardForBeacon := map[common.Hash]uint64{}
	rewardForShard := map[common.Hash]uint64{}
	rewardForIncDAO := map[common.Hash]uint64{}
	rewardForCustodian := map[common.Hash]uint64{}
	lenBeaconCommittees := uint64(len(env.BeaconCommittee))
	lenShardCommittees := uint64(len(env.ShardCommittee[env.ShardID]))
	if len(allCoinTotalReward) == 0 {
		Logger.log.Info("Beacon Height %+v, 😭 found NO reward", env.BeaconHeight)
		return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
	}
	for key, totalReward := range allCoinTotalReward {
		totalRewardForDAOAndCustodians := devPercent * totalReward / 100
		totalRewardForShardAndBeaconValidators := totalReward - totalRewardForDAOAndCustodians
		shardWeight := float64(lenShardCommittees)
		beaconWeight := 2 * float64(lenBeaconCommittees) / float64(len(env.ShardCommittee))
		totalValidatorWeight := shardWeight + beaconWeight
		rewardForShard[key] = uint64(shardWeight * float64(totalRewardForShardAndBeaconValidators) / totalValidatorWeight)
		Logger.log.Infof("[test-salary] totalRewardForDAOAndCustodians tokenID %v - %v\n",
			key.String(), totalRewardForDAOAndCustodians)
		if env.IsSplitRewardForCustodian {
			rewardForCustodian[key] += env.PercentCustodianReward * totalRewardForDAOAndCustodians / 100
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians - rewardForCustodian[key]
		} else {
			rewardForIncDAO[key] += totalRewardForDAOAndCustodians
		}
		rewardForBeacon[key] += totalReward - (rewardForShard[key] + totalRewardForDAOAndCustodians)
	}
	return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
}

type RewardSplitRuleV3 struct {
}

func (r RewardSplitRuleV3) SplitReward(env *SplitRewardEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error) {
	devPercent := uint64(env.DAOPercent)
	allCoinTotalReward := env.TotalReward // total reward for shard subset
	rewardForBeacon := map[common.Hash]uint64{}
	rewardForShardSubset := map[common.Hash]uint64{}
	rewardForIncDAO := map[common.Hash]uint64{}
	rewardForCustodian := map[common.Hash]uint64{}
	beaconCommitteeSize := uint64(len(env.BeaconCommittee))
	numberOfShard := len(env.ShardCommittee)

	shardSubsetCommitteeSize := uint64(len(env.ShardCommittee[env.ShardID]) / int(env.MaxSubsetCommittees))
	// Plus 1 for subset 0 if shard_committee_size id odd
	if len(env.ShardCommittee[env.ShardID])%int(env.MaxSubsetCommittees) != 0 {
		if (env.SubsetID % env.MaxSubsetCommittees) == 0 {
			shardSubsetCommitteeSize += uint64(len(env.ShardCommittee[env.ShardID]) % int(env.MaxSubsetCommittees))
		}
	}

	shardSubsetWeight := float64(shardSubsetCommitteeSize)
	beaconWeight := float64(beaconCommitteeSize) / float64(numberOfShard)
	totalValidatorWeight := shardSubsetWeight + beaconWeight

	if len(allCoinTotalReward) == 0 {
		Logger.log.Info("Beacon Height %+v, 😭 found NO reward", env.BeaconHeight)
		return rewardForBeacon, rewardForShardSubset, rewardForIncDAO, rewardForCustodian, nil
	}

	for coinID, totalReward := range allCoinTotalReward {
		totalRewardForDAOAndCustodians := devPercent * totalReward / 100
		totalRewardForShardAndBeaconValidators := totalReward - totalRewardForDAOAndCustodians

		rewardForShardSubset[coinID] = uint64(shardSubsetWeight * float64(totalRewardForShardAndBeaconValidators) / totalValidatorWeight)
		Logger.log.Infof("totalRewardForDAOAndCustodians tokenID %v - %v\n", coinID.String(), totalRewardForDAOAndCustodians)

		if env.IsSplitRewardForCustodian {
			rewardForCustodian[coinID] += env.PercentCustodianReward * totalRewardForDAOAndCustodians / 100
			rewardForIncDAO[coinID] += totalRewardForDAOAndCustodians - rewardForCustodian[coinID]
		} else {
			rewardForIncDAO[coinID] += totalRewardForDAOAndCustodians
		}
		rewardForBeacon[coinID] += totalReward - (rewardForShardSubset[coinID] + totalRewardForDAOAndCustodians)
	}
	return rewardForBeacon, rewardForShardSubset, rewardForIncDAO, rewardForCustodian, nil
}
