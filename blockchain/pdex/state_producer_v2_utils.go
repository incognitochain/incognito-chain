package pdex

import (
	"errors"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/utils"
)

func (sp *stateProducerV2) validateContributions(
	contribution0, contribution1 rawdbv2.Pdexv3Contribution,
) error {
	if contribution0.TokenID().String() == contribution1.TokenID().String() {
		return errors.New("contribution 0 and contribution 1 need to be same tokenID")
	}
	if contribution0.Amplifier() != contribution1.Amplifier() {
		return errors.New("contribution 0 and contribution 1 need to be same amplifier")
	}
	if contribution0.PoolPairID() != contribution1.PoolPairID() {
		return errors.New("contribution 0 and contribution 1 need to be same poolPairID")
	}
	if contribution0.OtaReceiver() != utils.EmptyString && contribution1.OtaReceiver() != utils.EmptyString {
		if contribution0.NftID().String() != contribution1.NftID().String() {
			return errors.New("contribution 0 and contribution 1 need to be same nftID")
		}
	}
	if contribution0.OtaReceiver() == utils.EmptyString && contribution1.OtaReceiver() == utils.EmptyString {
		if contribution0.AccessOTA() == utils.EmptyString && contribution1.AccessOTA() == utils.EmptyString {
			if contribution0.NftID().String() != contribution1.NftID().String() {
				return errors.New("contribution 0 and contribution 1 need to be same accessID")
			}
		}
	}
	return nil
}
