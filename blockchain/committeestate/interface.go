package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
)

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

type RewardSplitRule interface {
}
