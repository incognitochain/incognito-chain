package beststate

import (
	"encoding/json"

	"github.com/incognitochain/incognito-chain/incognitokey"
)

//BeaconPreCommitteeInfo ...
type BeaconPreCommitteeInfo struct {
	BeaconPendingValidator                 []incognitokey.CommitteePublicKey `json:"BeaconPendingValidator"`
	CandidateBeaconWaitingForCurrentRandom []incognitokey.CommitteePublicKey `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateBeaconWaitingForNextRandom    []incognitokey.CommitteePublicKey `json:"CandidateBeaconWaitingForNextRandom"`
}

//ShardPreCommitteeInfo ...
type ShardPreCommitteeInfo struct {
	CandidateShardWaitingForCurrentRandom []incognitokey.CommitteePublicKey          `json:"CandidateShardWaitingForCurrentRandom"`
	CandidateShardWaitingForNextRandom    []incognitokey.CommitteePublicKey          `json:"CandidateShardWaitingForNextRandom"`
	ShardPendingValidator                 map[byte][]incognitokey.CommitteePublicKey `json:"ShardPendingValidator"`
}

func (beaconPreCommitteeInfo *BeaconPreCommitteeInfo) MarshalJSON() ([]byte, error) {
	type Alias BeaconPreCommitteeInfo
	b, err := json.Marshal(&struct {
		*Alias
	}{
		(*Alias)(beaconPreCommitteeInfo),
	})
	if err != nil {
		return nil, err
	}
	return b, err
}

func (shardPreCommitteeInfo *ShardPreCommitteeInfo) MarshalJSON() ([]byte, error) {
	type Alias ShardPreCommitteeInfo
	b, err := json.Marshal(&struct {
		*Alias
	}{
		(*Alias)(shardPreCommitteeInfo),
	})
	if err != nil {
		return nil, err
	}
	return b, err
}
