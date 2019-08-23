package connmanager

import "sync"

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.

type ConsensusState struct {
	sync.Mutex
	role             string
	currentShard     *byte
	beaconCommittee  []string          // list public key of beacon committee
	committeeByShard map[byte][]string // map[shardID] = list CommitteePublicKeyBase58CheckStr of shard
	userPublicKey    string            // in base58check encode format
	shardByCommittee map[string]byte   // store conversion of ShardCommittee data map[CommitteePublicKeyBase58CheckStr] = shardID
	shardNumber      int
}

// rebuild - convert CommitteeByShard to ShardByCommittee
func (consensusState *ConsensusState) rebuild() {
	consensusState.shardByCommittee = make(map[string]byte)
	for shard, committees := range consensusState.committeeByShard {
		for _, committee := range committees {
			consensusState.shardByCommittee[committee] = shard
		}
	}
}

// getBeaconCommittee - return BeaconCommittee
func (consensusState *ConsensusState) getBeaconCommittee() []string {
	consensusState.Lock()
	defer consensusState.Unlock()
	lenBeaconCommittee := len(consensusState.beaconCommittee)
	beaconCommittee := make([]string, lenBeaconCommittee)
	if lenBeaconCommittee > 0 {
		copy(beaconCommittee, consensusState.beaconCommittee)
	}
	return beaconCommittee
}

// getCommitteeByShard - return CommitteeByShard
func (consensusState *ConsensusState) getCommitteeByShard(shard byte) []string {
	consensusState.Lock()
	defer consensusState.Unlock()
	if committee, ok := consensusState.committeeByShard[shard]; ok {
		lenCommittee := len(committee)
		shardCommittee := make([]string, lenCommittee)
		if lenCommittee > 0 {
			copy(shardCommittee, committee)
		}
		return shardCommittee
	}
	return make([]string, 0)
}

// getShardByCommittee - return list [commitee public key] = shardID
func (consensusState *ConsensusState) getShardByCommittee() map[string]byte {
	consensusState.Lock()
	defer consensusState.Unlock()
	listCommittee := make(map[string]byte)
	for publicKey, shardID := range consensusState.shardByCommittee {
		listCommittee[publicKey] = shardID
	}
	return listCommittee
}
