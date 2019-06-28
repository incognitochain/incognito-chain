package connmanager

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetCommitteeByShard(t *testing.T) {
	consensusState := ConsensusState{
		CommitteeByShard: make(map[byte][]string),
	}
	shardCommittee := make(map[byte][]string)
	shardCommittee[0] = []string{"a", "b"}
	shardCommittee[2] = []string{"c", "d"}

	for shardID, committee := range shardCommittee {
		consensusState.CommitteeByShard[shardID] = make([]string, len(committee))
		copy(consensusState.CommitteeByShard[shardID], committee)
	}
	commitee := consensusState.getCommitteeByShard(2)
	if len(commitee) == 0 {
		t.Error("Can not getCommitteeByShard")
	}
	commitee = consensusState.getCommitteeByShard(1)
	if len(commitee) > 0 {
		t.Error("Can not getCommitteeByShard")
	}
}

func TestConsensusState_GetBeaconCommittee(t *testing.T) {
	consensusState := ConsensusState{}
	beaconCommittee := []string{"a", "b"}

	consensusState.BeaconCommittee = make([]string, len(beaconCommittee))
	copy(consensusState.BeaconCommittee, beaconCommittee)
	committee := consensusState.getBeaconCommittee()
	if len(committee) == 0 {
		t.Error("Can not getBeaconCommittee")
	}
}

func TestConsensusState_GetShardByCommittee(t *testing.T) {
	consensusState := ConsensusState{
		ShardByCommittee: make(map[string]byte),
		CommitteeByShard: make(map[byte][]string),
	}

	shardCommittee := make(map[byte][]string)
	shardCommittee[0] = []string{"a", "b"}
	shardCommittee[2] = []string{"c", "d"}
	for shardID, committee := range shardCommittee {
		consensusState.CommitteeByShard[shardID] = make([]string, len(committee))
		copy(consensusState.CommitteeByShard[shardID], committee)
	}
	consensusState.rebuild()

	shardByCommittee := consensusState.getShardByCommittee()
	data, _ := json.Marshal(shardByCommittee)
	fmt.Println(string(data))
	if len(shardByCommittee) == 0 {
		t.Error("Can not getShardByCommittee")
	}
	if shardByCommittee["c"] != 2 {
		t.Error("Can not getShardByCommittee")
	}
}
