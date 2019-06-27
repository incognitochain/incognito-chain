package connmanager

import "sync"

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.

type ConsensusState struct {
	sync.Mutex
	Role             string
	CurrentShard     *byte
	BeaconCommittee  []string
	CommitteeByShard map[byte][]string // map[shardID] = list committeePubkeyBase58CheckStr of shard
	UserPublicKey    string            // in base58check encode format
	ShardByCommittee map[string]byte   // store conversion of ShardCommittee data map[committeePubkeyBase58CheckStr] = shardID
	ShardNumber      int
}

// rebuild - convert CommitteeByShard to ShardByCommittee
func (consensusState *ConsensusState) rebuild() {
	consensusState.ShardByCommittee = make(map[string]byte)
	for shard, committees := range consensusState.CommitteeByShard {
		for _, committee := range committees {
			consensusState.ShardByCommittee[committee] = shard
		}
	}
}

// getBeaconCommittee - return BeaconCommittee
func (consensusState *ConsensusState) getBeaconCommittee() []string {
	consensusState.Lock()
	defer consensusState.Unlock()
	ret := make([]string, len(consensusState.BeaconCommittee))
	copy(ret, consensusState.BeaconCommittee)
	return ret
}

// getCommitteeByShard - return CommitteeByShard
func (consensusState *ConsensusState) getCommitteeByShard(shard byte) []string {
	consensusState.Lock()
	defer consensusState.Unlock()
	committee, ok := consensusState.CommitteeByShard[shard]
	if ok {
		ret := make([]string, len(committee))
		copy(ret, committee)
		return ret
	}
	return make([]string, 0)
}

// getShardByCommittee - return list [commitee public key] = shardID
func (consensusState *ConsensusState) getShardByCommittee() map[string]byte {
	consensusState.Lock()
	defer consensusState.Unlock()
	ret := make(map[string]byte)
	for k, v := range consensusState.ShardByCommittee {
		ret[k] = v
	}
	return ret
}
