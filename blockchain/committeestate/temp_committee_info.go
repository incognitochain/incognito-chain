package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

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
