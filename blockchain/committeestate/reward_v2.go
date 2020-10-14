package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/pkg/errors"
)

//SplitReward ...
func (b *BeaconCommitteeEngineV2) SplitReward(
	env *BeaconCommitteeStateEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error) {

	hasValue := false
	devPercent := uint64(env.DAOPercent)
	totalReward := env.TotalReward
	rewardForBeacon := map[common.Hash]uint64{}
	rewardForShard := map[common.Hash]uint64{}
	rewardForIncDAO := map[common.Hash]uint64{}
	rewardForCustodian := map[common.Hash]uint64{}
	lenBeaconCommittees := uint64(len(b.finalBeaconCommitteeStateV2.beaconCommittee))
	lenShardCommittees := uint64(len(b.finalBeaconCommitteeStateV2.shardCommittee[env.ShardID]))
	beaconAndShardCommitteesSize := lenShardCommittees + 2*lenBeaconCommittees/uint64(env.ActiveShards)
	for key, value := range totalReward {
		totalRewardForDAOAndCustodians := uint64(devPercent) * value / 100
		totalRewardForShardAndBeaconValidators := value - totalRewardForDAOAndCustodians
		rewardForBeacon[key] = totalRewardForShardAndBeaconValidators - lenShardCommittees*totalRewardForShardAndBeaconValidators/beaconAndShardCommitteesSize
		Logger.log.Infof("[test-salary] totalRewardForDAOAndCustodians tokenID %v - %v\n", key.String(), totalRewardForDAOAndCustodians)
		if env.IsSplitRewardForCustodian {
			rewardForCustodian[key] = env.PercentCustodianReward * totalRewardForDAOAndCustodians / 100
			rewardForIncDAO[key] = totalRewardForDAOAndCustodians - rewardForCustodian[key]
		} else {
			rewardForIncDAO[key] = totalRewardForDAOAndCustodians
		}
		rewardForShard[key] = value - rewardForBeacon[key] - totalRewardForDAOAndCustodians
		if !hasValue {
			hasValue = true
		}
	}
	if !hasValue {
		return nil, nil, nil, nil, errors.New("Not enough reward")
	}
	return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
}
