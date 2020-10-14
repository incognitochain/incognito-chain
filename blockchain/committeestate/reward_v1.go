package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/pkg/errors"
)

//SplitReward ...
func (b *BeaconCommitteeEngineV1) SplitReward(
	env *BeaconCommitteeStateEnvironment) (
	map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error) {

	hasValue := false
	devPercent := uint64(env.DAOPercent)
	totalReward := env.TotalReward
	rewardForBeacon := map[common.Hash]uint64{}
	rewardForShard := map[common.Hash]uint64{}
	rewardForIncDAO := map[common.Hash]uint64{}
	rewardForCustodian := map[common.Hash]uint64{}
	for key, value := range totalReward {
		rewardForBeacon[key] = 2 * (uint64(100-devPercent) * value) / ((uint64(env.ActiveShards) + 2) * 100)
		totalRewardForDAOAndCustodians := uint64(devPercent) * value / uint64(100)
		Logger.log.Infof("[test-salary] totalRewardForDAOAndCustodians tokenID %v - %v\n", key.String(), totalRewardForDAOAndCustodians)
		if env.IsSplitRewardForCustodian {
			rewardForCustodian[key] = uint64(env.PercentCustodianReward) * totalRewardForDAOAndCustodians / uint64(100)
			rewardForIncDAO[key] = totalRewardForDAOAndCustodians - rewardForCustodian[key]
		} else {
			rewardForIncDAO[key] = totalRewardForDAOAndCustodians
		}
		rewardForShard[key] = value - (rewardForBeacon[key] + totalRewardForDAOAndCustodians)
		if !hasValue {
			hasValue = true
		}
	}
	if !hasValue {
		return nil, nil, nil, nil, errors.New("Not enough reward")
	}
	return rewardForBeacon, rewardForShard, rewardForIncDAO, rewardForCustodian, nil
}
