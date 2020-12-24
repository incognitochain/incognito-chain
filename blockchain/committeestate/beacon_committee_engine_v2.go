package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
)

type BeaconCommitteeEngineV2 struct {
	beaconCommitteeEngineBase
}

func NewBeaconCommitteeEngineV2(
	beaconHeight uint64,
	beaconHash common.Hash,
	finalBeaconCommitteeStateV2 *BeaconCommitteeStateV2) *BeaconCommitteeEngineV2 {
	Logger.log.Infof("Init Beacon Committee Engine V2, %+v", beaconHeight)
	return &BeaconCommitteeEngineV2{
		beaconCommitteeEngineBase: beaconCommitteeEngineBase{
			beaconHeight:     beaconHeight,
			beaconHash:       beaconHash,
			finalState:       finalBeaconCommitteeStateV2,
			uncommittedState: NewBeaconCommitteeStateV2(),
		},
	}
}

//Version :
func (engine BeaconCommitteeEngineV2) Version() uint {
	return SLASHING_VERSION
}
