package connmanager

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCommitteeByShard(t *testing.T) {
	consensusState := ConsensusState{
		committeeByShard: make(map[byte][]string),
	}
	shardCommittee := make(map[byte][]string)
	shardCommittee[0] = []string{"a", "b"}
	shardCommittee[2] = []string{"c", "d"}

	for shardID, committee := range shardCommittee {
		consensusState.committeeByShard[shardID] = make([]string, len(committee))
		copy(consensusState.committeeByShard[shardID], committee)
	}
	commitee := consensusState.getCommitteeByShard(2)
	if len(commitee) == 0 {
		t.Error("Can not getCommitteeByShard")
	}
	assert.Equal(t, "c", commitee[0])
	assert.Equal(t, "d", commitee[1])
	commitee = consensusState.getCommitteeByShard(1)
	if len(commitee) > 0 {
		t.Error("Can not getCommitteeByShard")
	}
}

func TestConsensusState_GetBeaconCommittee(t *testing.T) {
	consensusState := ConsensusState{}
	beaconCommittee := []string{"a", "b"}

	consensusState.beaconCommittee = make([]string, len(beaconCommittee))
	copy(consensusState.beaconCommittee, beaconCommittee)
	committee := consensusState.getBeaconCommittee()
	if len(committee) == 0 {
		t.Error("Can not getBeaconCommittee")
	}
}

func TestConsensusState_GetShardByCommittee(t *testing.T) {
	consensusState := ConsensusState{
		shardByCommittee: make(map[string]byte),
		committeeByShard: make(map[byte][]string),
	}

	shardCommittee := make(map[byte][]string)
	shardCommittee[0] = []string{"a", "b"}
	shardCommittee[2] = []string{"c", "d"}
	for shardID, committee := range shardCommittee {
		consensusState.committeeByShard[shardID] = make([]string, len(committee))
		copy(consensusState.committeeByShard[shardID], committee)
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
