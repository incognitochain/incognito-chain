package committeestate

import (
	"github.com/incognitochain/incognito-chain/instruction"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeState interface {
	Version() uint
	Clone() BeaconCommitteeState

	GetBeaconCommittee() []incognitokey.CommitteePublicKey
	GetBeaconSubstitute() []incognitokey.CommitteePublicKey
	GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey
	GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey
	GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey
	GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey
	GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey
	GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey
	GetUncommittedCommittee() map[byte][]incognitokey.CommitteePublicKey
	GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey
	GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey
	GetAutoStaking() map[string]bool
	GetStakingTx() map[string]common.Hash
	GetRewardReceiver() map[string]privacy.PaymentAddress
	GetAllCandidateSubstituteCommittee() []string
	GetShardCommonPool() []incognitokey.CommitteePublicKey

	UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
		*BeaconCommitteeStateHash,
		*CommitteeChange,
		[][]string,
		error)
	InitCommitteeState(env *BeaconCommitteeStateEnvironment)

	GenerateAllSwapShardInstructions(env *BeaconCommitteeStateEnvironment) ([]*instruction.SwapShardInstruction, error)

	ActiveShards() int
	AssignInstructions(env *BeaconCommitteeStateEnvironment) []*instruction.AssignInstruction
	SyncingValidators() map[byte][]incognitokey.CommitteePublicKey
	IsSwapTime(uint64, uint64) bool
	Upgrade(*BeaconCommitteeStateEnvironment) BeaconCommitteeState

	NumberOfAssignedCandidates() int

	Hash() (*BeaconCommitteeStateHash, error)
	Reset()
	SetBeaconCommittees([]incognitokey.CommitteePublicKey)
	SetNumberOfAssignedCandidates(int)
	SwapRule() SwapRule
	UnassignedCommonPool() []string
	AllSubstituteCommittees() []string
	SetSwapRule(SwapRule)
	SyncPool() map[byte][]incognitokey.CommitteePublicKey

	Mu() *sync.RWMutex
}

//fromB and toB need to be different from null
func cloneBeaconCommitteeStateFromTo(fromB, toB BeaconCommitteeState) {
	if fromB == nil {
		return
	}

	/*Logger.log.Infof("[dcs] fromB 0: %p \n", fromB)*/
	//Logger.log.Infof("[dcs] toB 0: %p \n", toB)

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
	/*Logger.log.Infof("[dcs] fromB 1: %p \n", fromB)*/
	/*Logger.log.Infof("[dcs] toB 1: %p \n", toB)*/
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
