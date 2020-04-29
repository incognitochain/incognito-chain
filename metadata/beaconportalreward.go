package metadata

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type PortalRewardContent struct {
	BeaconHeight uint64
	Rewards      []*statedb.PortalRewardInfo
}

func NewPortalReward(beaconHeight uint64, receivers []*statedb.PortalRewardInfo) *PortalRewardContent {
	return &PortalRewardContent{
		BeaconHeight: beaconHeight,
		Rewards:      receivers,
	}
}

type PortalTotalCustodianReward struct {
	Rewards []*statedb.RewardInfoDetail
}

func NewPortalTotalCustodianReward(rewards []*statedb.RewardInfoDetail) *PortalTotalCustodianReward {
	return &PortalTotalCustodianReward{
		Rewards: rewards,
	}
}
