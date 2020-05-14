package metadata

import (
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type PortalRewardContent struct {
	BeaconHeight uint64
	Rewards      map[string]*statedb.PortalRewardInfo // custodian incognito address : reward infos
}

func NewPortalReward(beaconHeight uint64, rewardInfos map[string]*statedb.PortalRewardInfo) *PortalRewardContent {
	return &PortalRewardContent{
		BeaconHeight: beaconHeight,
		Rewards:      rewardInfos,
	}
}

type PortalTotalCustodianReward struct {
	Rewards map[string]uint64
}

func NewPortalTotalCustodianReward(rewards map[string]uint64) *PortalTotalCustodianReward {
	return &PortalTotalCustodianReward{
		Rewards: rewards,
	}
}
