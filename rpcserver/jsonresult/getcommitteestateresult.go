package jsonresult

import (
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
)

type CommiteeState struct {
	Root              string                               `json:"root"`
	AutoStaking       map[string]bool                      `json:"autoStaking"`
	StakingTx         map[string]string                    `json:"stakingTx"`
	RewardReceivers   map[string]string                    `json:"rewardReceivers"`
	Committee         map[int][]string                     `json:"committee"`
	CurrentCandidate  []string                             `json:"currentCandidate"`
	NextCandidate     []string                             `json:"nextCandidate"`
	Syncing           map[int][]string                     `json:"syncing"`
	Substitute        map[int][]string                     `json:"substitute"`
	DelegateList      map[string]string                    `json:"delegateList"`
	ShardStakerInfos  map[string]*statedb.ShardStakerInfo  `json:"shardStakerInfos"`
	BeaconStakerInfos map[string]*statedb.BeaconStakerInfo `json:"beaconStakerInfos"`
}

func (cs *CommiteeState) IsDiffFrom(target *CommiteeState) bool {
	if cs.Root != target.Root {
		return false
	}
	return true
}

func (cs *CommiteeState) Print() {
	b, err := json.MarshalIndent(cs, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
