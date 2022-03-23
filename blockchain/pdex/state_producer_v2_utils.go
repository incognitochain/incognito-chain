package pdex

import (
	"errors"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
)

func (sp *stateProducerV2) validateContributions(
	contribution0, contribution1 rawdbv2.Pdexv3Contribution, // waiting and incoming contributions
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

	if contribution0.UseNft() && contribution1.UseNft() {
		if contribution0.NftID().String() != contribution1.NftID().String() {
			return errors.New("contribution 0 and contribution 1 need to be same nftID")
		}
	} else if contribution0.UseAccessOTANewLP() && contribution1.UseAccessOTANewLP() {
		if accessOTAReceiver, found := contribution0.OtaReceivers()[common.PdexAccessCoinID]; found {
			if accessOTAReceiver != contribution0.OtaReceiver() {
				return errors.New("otaReceiver need to be in otaReceivers")
			}
		} else {
			return errors.New("otaReceivers need to have pdexAccessCoin receiver")
		}
		if contribution0.OtaReceiver() != contribution1.OtaReceiver() {
			return errors.New("contribution 0 and contribution 1 need to be same otaReceiver")
		}
	} else if contribution0.UseAccessOTAOldLP() && contribution1.UseAccessOTAOldLP() {
		if contribution0.NftID().String() != contribution1.NftID().String() {
			return errors.New("contribution 0 and contribution 1 need to be same accessID")
		}
	} else {
		return errors.New("contribution 0 and contribution 1 need to use same access option")
	}

	return nil
}
