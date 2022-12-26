package jsonresult

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type CommiteeState struct {
	Root             string            `json:"root"`
	AutoStaking      map[string]bool   `json:"autoStaking"`
	StakingTx        map[string]string `json:"stakingTx"`
	RewardReceivers  map[string]string `json:"rewardReceivers"`
	Committee        map[int][]string  `json:"committee"`
	CurrentCandidate []string          `json:"currentCandidate"`
	NextCandidate    []string          `json:"nextCandidate"`
	BeaconWaiting    []string          `json:"beaconWaiting"`
	Syncing          map[int][]string  `json:"syncing"`
	Substitute       map[int][]string  `json:"substitute"`
	DelegateList     map[string]string `json:"delegateList"`
	//ShardStakerInfos  map[string]*statedb.ShardStakerInfo  `json:"shardStakerInfos"`
	BeaconStakerInfos map[string]*statedb.BeaconStakerInfo `json:"beaconStakerInfos"`
}

func (cs *CommiteeState) IsDiffFrom(target *CommiteeState) bool {
	if cs.Root != target.Root {
		return true
	}
	return false
}

func (cs *CommiteeState) Print() {
	b, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

// Filter for testing only
func (cs *CommiteeState) Filter(fixedCommiteesNodes map[int][]string, fixedRewardReceivers []string) {
	for i, v := range fixedCommiteesNodes {
		cs.Committee[i] = cs.Committee[i][len(v):]
		for _, value := range v {
			delete(cs.AutoStaking, value)
			delete(cs.StakingTx, value)
		}
	}

	for _, v := range fixedRewardReceivers {
		delete(cs.RewardReceivers, v)
	}

	cs.DelegateList = map[string]string{}
	cs.AutoStaking = map[string]bool{}
	cs.RewardReceivers = map[string]string{}
	cs.StakingTx = map[string]string{}

	for i, v := range cs.Committee {
		t := []string{}
		for _, value := range v {
			t = append(t, value[len(value)-5:])
		}
		cs.Committee[i] = t
	}

	for i, v := range cs.Substitute {
		t := []string{}
		for _, value := range v {
			t = append(t, value[len(value)-5:])
		}
		cs.Substitute[i] = t
	}

	for i, v := range cs.Syncing {
		t := []string{}
		for _, value := range v {
			t = append(t, value[len(value)-5:])
		}
		cs.Syncing[i] = t
	}

	t := []string{}
	for _, value := range cs.BeaconWaiting {
		t = append(t, value[len(value)-5:])
	}
	cs.BeaconWaiting = t

	t = []string{}
	for _, value := range cs.NextCandidate {
		t = append(t, value[len(value)-5:])
	}
	cs.NextCandidate = t

	t = []string{}
	for _, value := range cs.CurrentCandidate {
		t = append(t, value[len(value)-5:])
	}
	cs.CurrentCandidate = t
}
