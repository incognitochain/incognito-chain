package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
)

//ShardCommitteeState :
type ShardCommitteeState interface {
	Version() int
	Clone() ShardCommitteeState
	GetShardCommittee() []incognitokey.CommitteePublicKey
	GetShardSubstitute() []incognitokey.CommitteePublicKey

	UpdateCommitteeState(env *ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash,
		*CommitteeChange, error)

	BuildTotalTxsFeeFromTxs(txs []metadata.Transaction) map[common.Hash]uint64
}

type SwapInstructionGenerator interface {
	GenerateSwapInstructions(env *ShardCommitteeStateEnvironment) (*instruction.SwapInstruction, []string, []string, error)
}

type AssignInstructionProcessor interface {
	ProcessAssignInstructions(env *ShardCommitteeStateEnvironment) []incognitokey.CommitteePublicKey
}
