package committeestate

import (
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

type TempCommitteeInfo struct {
	hash         common.Hash
	beaconHeight uint64
	committees   []incognitokey.CommitteePublicKey
	shardID      byte
}

func NewTempCommitteeInfo() *TempCommitteeInfo {
	return &TempCommitteeInfo{}
}

func NewTempCommitteeInfoWithValue(
	hash common.Hash,
	committees []incognitokey.CommitteePublicKey,
	shardID byte,
	beaconHeight uint64,
) *TempCommitteeInfo {
	return &TempCommitteeInfo{
		hash:         hash,
		beaconHeight: beaconHeight,
		committees:   committees,
		shardID:      shardID,
	}
}

func (tempCommitteeInfo *TempCommitteeInfo) ShardID() byte {
	return tempCommitteeInfo.shardID
}

func (tempCommitteeInfo *TempCommitteeInfo) BeaconHeight() uint64 {
	return tempCommitteeInfo.beaconHeight
}

func (tempCommitteeInfo *TempCommitteeInfo) Committees() []incognitokey.CommitteePublicKey {
	return tempCommitteeInfo.committees
}

func (tempCommitteeInfo *TempCommitteeInfo) Hash() common.Hash {
	return tempCommitteeInfo.hash
}
