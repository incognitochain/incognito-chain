package committeestate

import (
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
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
	PropationPool() map[string]signaturecounter.Penalty
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
	Terms() map[string]uint64
}

func cloneBeaconCommitteeStateFrom(state BeaconCommitteeState) BeaconCommitteeState {
	var res BeaconCommitteeState
	if state != nil {
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
	}
	return res
}
