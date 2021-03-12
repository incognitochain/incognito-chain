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

	SplitReward(*BeaconCommitteeStateEnvironment) (map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error)
}
