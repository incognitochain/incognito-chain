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

type ShardCommitteeForBlockProducing struct {
	hash         common.Hash
	beaconHeight uint64
	committees   []incognitokey.CommitteePublicKey
	shardID      byte
}

func NewTempCommitteeInfo() *ShardCommitteeForBlockProducing {
	return &ShardCommitteeForBlockProducing{}
}

func NewTempCommitteeInfoWithValue(
	hash common.Hash,
	committees []incognitokey.CommitteePublicKey,
	shardID byte,
	beaconHeight uint64,
) *ShardCommitteeForBlockProducing {
	return &ShardCommitteeForBlockProducing{
		hash:         hash,
		beaconHeight: beaconHeight,
		committees:   committees,
		shardID:      shardID,
	}
}

func (tempCommitteeInfo *ShardCommitteeForBlockProducing) ShardID() byte {
	return tempCommitteeInfo.shardID
}

func (tempCommitteeInfo *ShardCommitteeForBlockProducing) BeaconHeight() uint64 {
	return tempCommitteeInfo.beaconHeight
}

func (tempCommitteeInfo *ShardCommitteeForBlockProducing) Committees() []incognitokey.CommitteePublicKey {
	return tempCommitteeInfo.committees
}

func (tempCommitteeInfo *ShardCommitteeForBlockProducing) Hash() common.Hash {
	return tempCommitteeInfo.hash
}
