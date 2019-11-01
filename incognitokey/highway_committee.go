package incognitokey

import (
	"encoding/json"
)

type ChainCommittee struct {
	Epoch             uint64
	BeaconCommittee   []CommitteeKeyString
	AllShardCommittee map[byte][]CommitteeKeyString
	AllShardPending   map[byte][]CommitteeKeyString
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
