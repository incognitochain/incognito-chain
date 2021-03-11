package committeestate

import (
	"github.com/incognitochain/incognito-chain/privacy"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type BeaconCommitteeStateV2 struct {
	beaconCommitteeStateSlashingBase
}

func NewBeaconCommitteeStateV2() *BeaconCommitteeStateV2 {
	return &BeaconCommitteeStateV2{
		beaconCommitteeStateSlashingBase: *NewBeaconCommitteeStateSlashingBase(),
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
		beaconCommitteeStateSlashingBase: *NewBeaconCommitteeStateSlashingBaseWithValue(
			beaconCommittee, shardCommittee, shardSubstitute,
			autoStake, rewardReceiver, stakingTx,
			shardCommonPool,
			numberOfAssignedCandidates, swapRule,
		),
	}
}

func (b *BeaconCommitteeStateV2) clone() *BeaconCommitteeStateV2 {
	res := NewBeaconCommitteeStateV2()
	res.beaconCommitteeStateSlashingBase = *b.beaconCommitteeStateSlashingBase.clone()
	return res
}

func (b *BeaconCommitteeStateV2) reset() {
	b.beaconCommitteeStateBase.reset()
}

func (b *BeaconCommitteeStateV2) cloneFrom(fromB BeaconCommitteeStateV2) {
	b.reset()
	b.beaconCommitteeStateSlashingBase.cloneFrom(fromB.beaconCommitteeStateSlashingBase)
}

//Upgrade check interface method for des
func (engine *BeaconCommitteeStateV2) Upgrade(env *BeaconCommitteeStateEnvironment) BeaconCommitteeState {
	beaconCommittee, shardCommittee, shardSubstitute,
		shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule := engine.getDataForUpgrading(env)

	committeeStateV3 := NewBeaconCommitteeStateV3WithValue(
		beaconCommittee,
		shardCommittee,
		shardSubstitute,
		shardCommonPool,
		numberOfAssignedCandidates,
		autoStake,
		rewardReceiver,
		stakingTx,
		map[byte][]incognitokey.CommitteePublicKey{},
		swapRule,
	)
	return committeeStateV3
}

func (b *BeaconCommitteeStateV2) getDataForUpgrading(env *BeaconCommitteeStateEnvironment) (
	[]incognitokey.CommitteePublicKey,
	map[byte][]incognitokey.CommitteePublicKey,
	map[byte][]incognitokey.CommitteePublicKey,
	[]incognitokey.CommitteePublicKey,
	int,
	map[string]bool,
	map[string]privacy.PaymentAddress,
	map[string]common.Hash,
	SwapRule,
) {
	beaconCommittee, shardCommittee, shardSubstitute,
		shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule := b.getDataForUpgrading(env)

	numberOfAssignedCandidates = b.NumberOfAssignedCandidates()
	shardCommonPool = make([]incognitokey.CommitteePublicKey, numberOfAssignedCandidates)
	copy(shardCommonPool, b.shardCommonPool)
	return beaconCommittee, shardCommittee, shardSubstitute, shardCommonPool, numberOfAssignedCandidates,
		autoStake, rewardReceiver, stakingTx, swapRule
}
