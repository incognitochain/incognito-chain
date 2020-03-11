package metadata

type PortalRewardContent struct {
	BeaconHeight uint64
	Receivers map[string]uint64   // CustodianIncAddr : reward amount in prv
}

func NewPortalReward(beaconHeight uint64, receivers map[string]uint64) (*PortalRewardContent, error) {
	return &PortalRewardContent {
		BeaconHeight: beaconHeight,
		Receivers:    receivers,
	}, nil
}

