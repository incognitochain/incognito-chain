package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/privacy"
)

//BeaconCommitteeEngine :
type BeaconCommitteeEngine interface {
	Clone() BeaconCommitteeEngine
	Version() uint
	GetBeaconHeight() uint64
	GetBeaconHash() common.Hash
	GetBeaconCommittee() []incognitokey.CommitteePublicKey
	GetBeaconSubstitute() []incognitokey.CommitteePublicKey
	GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey
	GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey
	GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey
	GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey
	GetOneShardCommittee(shardID byte) []incognitokey.CommitteePublicKey
	GetShardCommittee() map[byte][]incognitokey.CommitteePublicKey
	GetOneShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey
	GetShardSubstitute() map[byte][]incognitokey.CommitteePublicKey
	GetAutoStaking() map[string]bool
	GetStakingTx() map[string]common.Hash
	GetRewardReceiver() map[string]privacy.PaymentAddress
	GetAllCandidateSubstituteCommittee() []string
	Commit(*BeaconCommitteeStateHash) error
	AbortUncommittedBeaconState()
	UpdateCommitteeState(env *BeaconCommitteeStateEnvironment) (
		*BeaconCommitteeStateHash,
		*CommitteeChange,
		[][]string,
		error)
	InitCommitteeState(env *BeaconCommitteeStateEnvironment)
	GenerateAssignInstruction(rand int64, assignOffset int, activeShards int) ([]*instruction.AssignInstruction, []string, map[byte][]string)
	GenerateAllSwapShardInstructions(env *BeaconCommitteeStateEnvironment) ([]*instruction.SwapShardInstruction, error)
	BuildIncurredInstructions(env *BeaconCommitteeStateEnvironment) ([][]string, error)
}

//ShardCommitteeEngine :
type ShardCommitteeEngine interface {
	// Clone() ShardCommitteeEngine
	Commit(*ShardCommitteeStateHash) error
	AbortUncommittedShardState()
	UpdateCommitteeState(env ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash,
		*CommitteeChange, error)
	InitCommitteeState(env ShardCommitteeStateEnvironment)
	GetShardCommittee(shardID byte) []incognitokey.CommitteePublicKey
	GetShardSubstitute(shardID byte) []incognitokey.CommitteePublicKey
	ProcessInstructionFromBeacon(env ShardCommitteeStateEnvironment) (*CommitteeChange, error)
	ProcessInstructionFromShard(env ShardCommitteeStateEnvironment) (*CommitteeChange, error)
	GenerateSwapInstruction(env ShardCommitteeStateEnvironment) (*instruction.SwapInstruction, []string, []string, error)
}
