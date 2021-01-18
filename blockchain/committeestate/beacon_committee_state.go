package committeestate

import (
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeState interface {
	Version() int
	BeaconCommittee() []incognitokey.CommitteePublicKey
	ShardCommittee() map[byte][]incognitokey.CommitteePublicKey
	ShardSubstitute() map[byte][]incognitokey.CommitteePublicKey
	ShardCommonPool() []incognitokey.CommitteePublicKey
	CandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey
	CandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey
	NumberOfAssignedCandidates() int
	AutoStake() map[string]bool
	RewardReceiver() map[string]privacy.PaymentAddress
	StakingTx() map[string]common.Hash
	Mu() *sync.RWMutex
	AllCandidateSubstituteCommittees() []string
	IsEmpty() bool
	Hash() (*BeaconCommitteeStateHash, error)
	Reset()
	SetBeaconCommittees([]incognitokey.CommitteePublicKey)
	SetNumberOfAssignedCandidates(int)
	SwapRule() SwapRule
	UnassignedCommonPool() []string
	AllSubstituteCommittees() []string
	SetSwapRule(SwapRule)
	SyncPool() map[byte][]incognitokey.CommitteePublicKey
}

//fromB and toB need to be different from null
func cloneBeaconCommitteeStateFromTo(fromB, toB BeaconCommitteeState) {
	if fromB == nil {
		return
	}
	switch fromB.Version() {
	case SELF_SWAP_SHARD_VERSION:
		toB.(*BeaconCommitteeStateV1).cloneFrom(*fromB.(*BeaconCommitteeStateV1))
	case SLASHING_VERSION:
		toB.(*BeaconCommitteeStateV2).cloneFrom(*fromB.(*BeaconCommitteeStateV2))
	case DCS_VERSION:
		toB.(*BeaconCommitteeStateV3).cloneFrom(*fromB.(*BeaconCommitteeStateV3))
	case STATE_TEST_VERSION:
		toB = fromB
	}
}

func cloneBeaconCommitteeStateFrom(state BeaconCommitteeState) BeaconCommitteeState {
	if state == nil {
		return nil
	}
	var res BeaconCommitteeState
	switch state.Version() {
	case SELF_SWAP_SHARD_VERSION:
		res = state.(*BeaconCommitteeStateV1).clone()
	case SLASHING_VERSION:
		res = state.(*BeaconCommitteeStateV2).clone()
	case DCS_VERSION:
		res = state.(*BeaconCommitteeStateV3).clone()
	case STATE_TEST_VERSION:
		res = state
	}
	return res
}
