package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"sync"
)

type ShardCommitteeStateV3 struct {
	ShardCommitteeStateV2
}

//NewShardCommitteeStateV3 is default constructor for ShardCommitteeStateV3 ...
//Output: pointer of ShardCommitteeStateV3 struct
func NewShardCommitteeStateV3() *ShardCommitteeStateV3 {
	return &ShardCommitteeStateV3{
		ShardCommitteeStateV2: *NewShardCommitteeStateV2(),
	}
}

//NewShardCommitteeStateV3WithValue is constructor for ShardCommitteeStateV3 with value
//Output: pointer of ShardCommitteeStateV3 struct with value
func NewShardCommitteeStateV3WithValue(
	shardCommittee []incognitokey.CommitteePublicKey,
	committeeFromBlockHash common.Hash,
) *ShardCommitteeStateV3 {
	res, _ := incognitokey.CommitteeKeyListToString(shardCommittee)

	return &ShardCommitteeStateV3{
		ShardCommitteeStateV2: ShardCommitteeStateV2{
			shardCommittee:     res,
			committeeFromBlock: committeeFromBlockHash,
			mu:                 new(sync.RWMutex),
		},
	}
}

//initGenesisShardCommitteeStateV3 init committee state at genesis block or anytime restore program
//	- call function processInstructionFromBeacon for process instructions received from beacon
//	- call function processShardBlockInstruction for process shard block instructions
func initGenesisShardCommitteeStateV3(env *ShardCommitteeStateEnvironment) *ShardCommitteeStateV3 {
	s2 := initGenesisShardCommitteeStateV2(env)
	s3 := &ShardCommitteeStateV3{
		ShardCommitteeStateV2: *s2,
	}
	return s3
}

func (s ShardCommitteeStateV3) Version() int {
	return DCS_VERSION
}

func (s ShardCommitteeStateV3) Clone() ShardCommitteeState {

	newS := NewShardCommitteeStateV3()
	newS.ShardCommitteeStateV2.shardCommittee = common.DeepCopyString(s.shardCommittee)
	newS.ShardCommitteeStateV2.committeeFromBlock = s.committeeFromBlock

	return newS
}

func (s ShardCommitteeStateV3) GetShardCommittee() []incognitokey.CommitteePublicKey {
	res, _ := incognitokey.CommitteeBase58KeyListToStruct(s.shardCommittee)
	return res
}

func (s ShardCommitteeStateV3) GetShardSubstitute() []incognitokey.CommitteePublicKey {
	return []incognitokey.CommitteePublicKey{}
}

func (s ShardCommitteeStateV3) GetCommitteeFromBlock() common.Hash {
	return s.committeeFromBlock
}

func (s ShardCommitteeStateV3) UpdateCommitteeState(env *ShardCommitteeStateEnvironment) (*ShardCommitteeStateHash, *CommitteeChange, error) {
	return s.ShardCommitteeStateV2.UpdateCommitteeState(env)
}

func (s ShardCommitteeStateV3) BuildTotalTxsFeeFromTxs(txs []metadata.Transaction) map[common.Hash]uint64 {
	return s.ShardCommitteeStateV2.BuildTotalTxsFeeFromTxs(txs)
}
