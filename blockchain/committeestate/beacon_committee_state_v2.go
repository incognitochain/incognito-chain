package committeestate

import (
	"sync"

	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateV2 struct {
	beaconCommitteeStateBase
}

func NewBeaconCommitteeStateV2() *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		beaconCommitteeStateBase: beaconCommitteeStateBase{
			shardCommittee:  make(map[byte][]incognitokey.CommitteePublicKey),
			shardSubstitute: make(map[byte][]incognitokey.CommitteePublicKey),
			autoStake:       make(map[string]bool),
			rewardReceiver:  make(map[string]privacy.PaymentAddress),
			stakingTx:       make(map[string]common.Hash),
			mu:              new(sync.RWMutex),
		},
	}
}

func NewBeaconCommitteeStateV2WithValue(
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	shardCommonPool []incognitokey.CommitteePublicKey,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	swapRule SwapRule,
) *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		beaconCommitteeStateBase: beaconCommitteeStateBase{
			beaconCommittee:            beaconCommittee,
			shardCommittee:             shardCommittee,
			shardSubstitute:            shardSubstitute,
			shardCommonPool:            shardCommonPool,
			numberOfAssignedCandidates: numberOfAssignedCandidates,
			autoStake:                  autoStake,
			rewardReceiver:             rewardReceiver,
			stakingTx:                  stakingTx,
			swapRule:                   swapRule,
			mu:                         new(sync.RWMutex),
		},
	}
}

func (b *BeaconCommitteeStateV2) Version() int {
	return SLASHING_VERSION
}
