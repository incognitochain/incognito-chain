package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

type BeaconCommitteeState interface {
	Version() int
	Clone() BeaconCommitteeState

	GetBeaconCommittee() []incognitokey.CommitteePublicKey
	GetBeaconSubstitute() []incognitokey.CommitteePublicKey
	GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey
	GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey
	GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey
	GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey
	GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey
	GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey
	GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey
	GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey
	GetAutoStaking() map[string]bool
	GetStakingTx() map[string]common.Hash
	GetRewardReceiver() map[string]privacy.PaymentAddress
	GetAllCandidateSubstituteCommittee() []string

	GetSyncPool() map[byte][]incognitokey.CommitteePublicKey
	GetShardCommonPool() []incognitokey.CommitteePublicKey
	GetSyncingValidators() map[byte][]incognitokey.CommitteePublicKey

	UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
		*BeaconCommitteeStateHash,
		*CommitteeChange,
		[][]string,
		error)
	//TODO: refactor to non-method function
	InitCommitteeState(env *BeaconCommitteeStateEnvironment)
	Hash() (*BeaconCommitteeStateHash, error)

	ActiveShards() int
	IsSwapTime(uint64, uint64) bool
	Upgrade(*BeaconCommitteeStateEnvironment) BeaconCommitteeState
}

type BeaconCommitteeStateProcessor interface {
}

type AssignInstructionsGenerator interface {
	GenerateAssignInstructions(env *BeaconCommitteeStateEnvironment) []*instruction.AssignInstruction
}

type SwapShardInstructionsGenerator interface {
	GenerateAllSwapShardInstructions(env *BeaconCommitteeStateEnvironment) ([]*instruction.SwapShardInstruction, error)
}

type SplitRewardRuleProcessor interface {
	Process(*BeaconCommitteeStateEnvironment) (map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error)
}

//ShardCommitteeEngine :
type ShardCommitteeEngine interface {
	Version() int
	Clone() ShardCommitteeEngine
	Commit(*ShardCommitteeStateHash) error
	AbortUncommittedShardState()
	UpdateCommitteeState(env ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash,
		*CommitteeChange, error)
	InitCommitteeState(env ShardCommitteeStateEnvironment)
	GetShardCommittee() []incognitokey.CommitteePublicKey
	GetShardSubstitute() []incognitokey.CommitteePublicKey
	CommitteeFromBlock() common.Hash
	ProcessInstructionFromBeacon(env ShardCommitteeStateEnvironment) (*CommitteeChange, error)
	GenerateSwapInstruction(env ShardCommitteeStateEnvironment) (*instruction.SwapInstruction, []string, []string, error)
	BuildTotalTxsFeeFromTxs(txs []metadata.Transaction) map[common.Hash]uint64
}

type RewardSplitRule interface {
}
