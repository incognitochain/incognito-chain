package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
)

//BeaconCommitteeEngine :
type BeaconCommitteeEngine interface {
	Version() uint
	Clone() BeaconCommitteeEngine
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
	GetUncommittedCommittee() map[byte][]incognitokey.CommitteePublicKey
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
	GenerateAssignInstruction(rand int64, assignOffset int, activeShards int, beaconHeight uint64) []*instruction.AssignInstruction
	GenerateAllSwapShardInstructions(env *BeaconCommitteeStateEnvironment) ([]*instruction.SwapShardInstruction, error)
	SplitReward(*BeaconCommitteeStateEnvironment) (map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, map[common.Hash]uint64, error)
	ActiveShards() int
	GenerateAssignSyncInstructions(env *BeaconCommitteeStateEnvironment) ([]*instruction.AssignSyncInstruction, error)
}

//ShardCommitteeEngine :
type ShardCommitteeEngine interface {
	Version() uint
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

type StakeRule interface {
}
