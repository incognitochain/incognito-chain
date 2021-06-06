package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
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
func initGenesisShardCommitteeStateV3(env ShardCommitteeStateEnvironment) *ShardCommitteeStateV3 {
	s2 := initGenesisShardCommitteeStateV2(env)
	s3 := &ShardCommitteeStateV3{
		ShardCommitteeStateV2: *s2,
	}
	return s3
}

func (s ShardCommitteeStateV3) Version() int {
	return DCS_VERSION
}
