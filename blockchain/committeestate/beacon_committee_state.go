package committeestate

import (
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeState interface {
	Version() int
	BeaconCommittee() []incognitokey.CommitteePublicKey
	ShardCommittee() map[byte][]incognitokey.CommitteePublicKey
	ShardSubstitute() map[byte][]incognitokey.CommitteePublicKey
	ShardCommonPool() []incognitokey.CommitteePublicKey
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
	ProcessStakeInstruction(*instruction.StakeInstruction, *CommitteeChange) (*CommitteeChange, error)
	ProcessStopAutoStakeInstruction(*instruction.StopAutoStakeInstruction, *BeaconCommitteeStateEnvironment, *CommitteeChange, BeaconCommitteeState) *CommitteeChange
	ProcessAssignWithRandomInstruction(int64, int, *CommitteeChange, BeaconCommitteeState) *CommitteeChange
	ProcessSwapShardInstruction(*instruction.SwapShardInstruction, *BeaconCommitteeStateEnvironment, *CommitteeChange, *instruction.ReturnStakeInstruction, BeaconCommitteeState) (*CommitteeChange, *instruction.ReturnStakeInstruction, error)
	ProcessUnstakeInstruction(*instruction.UnstakeInstruction, *BeaconCommitteeStateEnvironment, *CommitteeChange, *instruction.ReturnStakeInstruction, BeaconCommitteeState) (*CommitteeChange, *instruction.ReturnStakeInstruction, error)
	SyncPool() map[byte][]incognitokey.CommitteePublicKey
	SetSwapRule(SwapRule)
}

func cloneBeaconCommitteeStateFrom(state BeaconCommitteeState) BeaconCommitteeState {
	var res BeaconCommitteeState
	if state != nil {
		switch state.Version() {
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
