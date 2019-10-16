package peerv2

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

type ChainCommittee struct {
	Epoch             uint64
	BeaconCommittee   []incognitokey.CommitteePublicKey
	AllShardCommittee map[byte][]incognitokey.CommitteePublicKey
	AllShardPending   map[byte][]incognitokey.CommitteePublicKey
}

func (cc *ChainCommittee) ToByte() ([]byte, error) {
	data, err := json.Marshal(cc)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func ChainCommitteeFromByte(data []byte) (*ChainCommittee, error) {
	cc := &ChainCommittee{}
	err := json.Unmarshal(data, cc)
	if err != nil {
		return nil, err
	}
	return cc, nil
}
