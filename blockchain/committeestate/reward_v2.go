package committeestate

import (
	"fmt"
	"github.com/incognitochain/incognito-chain/common"
)

//SplitReward ...
// TODO: @tin rewrite
func (b *BeaconCommitteeEngineV2) SplitReward(
	env *BeaconCommitteeStateEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error) {

	hasValue := false
	devPercent := uint64(env.DAOPercent)
	totalRewardForShard := env.TotalRewardForShard
	totalRewardForBeacon := env.TotalRewardForBeacon
	totalRewardForIncDAO := env.TotalRewardForIncDAO
	totalRewardForCustodian := env.TotalRewardForCustodian
	lenBeaconCommittees := uint64(len(b.finalBeaconCommitteeStateV2.beaconCommittee))
	lenShardCommittees := uint64(len(b.finalBeaconCommitteeStateV2.shardCommittee[env.ShardID]))
	beaconAndShardCommitteesSize := lenShardCommittees + 2*lenBeaconCommittees/uint64(env.ActiveShards)

	for key, value := range totalRewardForShard {
		totalRewardForDAOAndCustodians := uint64(devPercent) * value / 100
		totalRewardForShardAndBeaconValidators := value - totalRewardForDAOAndCustodians
		totalRewardForBeacon[key] += totalRewardForShardAndBeaconValidators - lenShardCommittees*totalRewardForShardAndBeaconValidators/beaconAndShardCommitteesSize
		Logger.log.Infof("[test-salary] totalRewardForDAOAndCustodians tokenID %v - %v\n", key.String(), totalRewardForDAOAndCustodians)
		if env.IsSplitRewardForCustodian {
			totalRewardForCustodian[key] += env.PercentCustodianReward * totalRewardForDAOAndCustodians / 100
			totalRewardForIncDAO[key] += totalRewardForDAOAndCustodians - totalRewardForCustodian[key]
		} else {
			totalRewardForIncDAO[key] += totalRewardForDAOAndCustodians
		}
		totalRewardForShard[key] = value - totalRewardForBeacon[key] - totalRewardForDAOAndCustodians
		if !hasValue {
			hasValue = true
		}
	}

	if !hasValue {
		return nil, nil, nil, nil, NewCommitteeStateError(ErrNotEnoughReward, fmt.Errorf("have no reward value"))
	}

	return totalRewardForBeacon, totalRewardForShard, totalRewardForIncDAO, totalRewardForCustodian, nil
}
