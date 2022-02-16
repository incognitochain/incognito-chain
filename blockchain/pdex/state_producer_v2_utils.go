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
	contribution0UseOtaReceiver := contribution0.OtaReceiver() != utils.EmptyString
	contribution1UseOtaReceiver := contribution1.OtaReceiver() != utils.EmptyString
	if contribution0UseOtaReceiver && contribution1UseOtaReceiver {
		if contribution0.NftID().String() != contribution1.NftID().String() {
			return errors.New("contribution 0 and contribution 1 need to be same nftID")
		}
	} else if !contribution0UseOtaReceiver && !contribution1UseOtaReceiver {
		if contribution0.AccessOTA() == nil && contribution1.AccessOTA() == nil {
			if contribution0.NftID().String() != contribution1.NftID().String() {
				return errors.New("contribution 0 and contribution 1 need to be same accessID")
			}
		}
	} else {
		return errors.New("both contributions must share OtaReceiver format")
	}
	return nil
}
