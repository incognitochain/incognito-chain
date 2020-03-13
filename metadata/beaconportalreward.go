package metadata

import "github.com/incognitochain/incognito-chain/database/lvdb"

type PortalRewardContent struct {
	BeaconHeight uint64
	Rewards      []*lvdb.PortalRewardInfo
}

func NewPortalReward(beaconHeight uint64, receivers []*lvdb.PortalRewardInfo) (*PortalRewardContent, error) {
	return &PortalRewardContent{
		BeaconHeight: beaconHeight,
		Rewards:      receivers,
	}, nil
}
