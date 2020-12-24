package committeestate

import (
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeStateV3 struct {
	beaconCommitteeStateBase
	syncPool map[byte][]incognitokey.CommitteePublicKey
}

func NewBeaconCommitteeStateV3() *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
		beaconCommitteeStateBase: beaconCommitteeStateBase{
			shardCommittee:  make(map[byte][]incognitokey.CommitteePublicKey),
			shardSubstitute: make(map[byte][]incognitokey.CommitteePublicKey),
			autoStake:       make(map[string]bool),
			rewardReceiver:  make(map[string]privacy.PaymentAddress),
			stakingTx:       make(map[string]common.Hash),
			mu:              new(sync.RWMutex),
		},
		syncPool: make(map[byte][]incognitokey.CommitteePublicKey),
	}
}

func NewBeaconCommitteeStateV3WithValue(
	beaconCommittee []incognitokey.CommitteePublicKey,
	shardCommittee map[byte][]incognitokey.CommitteePublicKey,
	shardSubstitute map[byte][]incognitokey.CommitteePublicKey,
	shardCommonPool []incognitokey.CommitteePublicKey,
	numberOfAssignedCandidates int,
	autoStake map[string]bool,
	rewardReceiver map[string]privacy.PaymentAddress,
	stakingTx map[string]common.Hash,
	syncPool map[byte][]incognitokey.CommitteePublicKey,
	swapRule SwapRule,
) *BeaconCommitteeStateV3 {
	return &BeaconCommitteeStateV3{
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
		syncPool: syncPool,
	}
}

func (b *BeaconCommitteeStateV3) Version() int {
	return DCS_VERSION
}

func (b *BeaconCommitteeStateV3) SyncPool() map[byte][]incognitokey.CommitteePublicKey {
	return b.syncPool
}

//ProcessAssignWithRandomInstruction process assign with random instruction
//@TODO: Override from parent function and handle to add validators to syncPool
func (b *BeaconCommitteeStateV3) ProcessAssignWithRandomInstruction(
	rand int64,
	activeShards int,
	committeeChange *CommitteeChange,
	oldState BeaconCommitteeState,
) *CommitteeChange {
	return nil
}
