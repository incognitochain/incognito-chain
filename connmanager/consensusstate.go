package connmanager

import "sync"

// ConnState can be either pending, established, disconnected or failed.  When
// a new connection is requested, it is attempted and categorized as
// established or failed depending on the connection result.  An established
// connection which was disconnected is categorized as disconnected.

type ConsensusState struct {
	sync.Mutex
	Role            string
	CurrentShard    *byte
	BeaconCommittee []string
	ShardCommittee  map[byte][]string
	UserPbk         string
	Committee       map[string]byte
	ShardNumber     int
}

func (consensusState *ConsensusState) rebuild() {
	consensusState.Committee = make(map[string]byte)
	for shard, committees := range consensusState.ShardCommittee {
		for _, committee := range committees {
			consensusState.Committee[committee] = shard
		}
	}
}

func (consensusState *ConsensusState) GetBeaconCommittee() []string {
	consensusState.Lock()
	defer consensusState.Unlock()
	ret := make([]string, len(consensusState.BeaconCommittee))
	copy(ret, consensusState.BeaconCommittee)
	return ret
}

func (consensusState *ConsensusState) GetShardCommittee(shard byte) []string {
	consensusState.Lock()
	defer consensusState.Unlock()
	committee, ok := consensusState.ShardCommittee[shard]
	if ok {
		ret := make([]string, len(committee))
		copy(ret, committee)
		return ret
	}
	return make([]string, 0)
}

func (consensusState *ConsensusState) GetCommittee() map[string]byte {
	consensusState.Lock()
	defer consensusState.Unlock()
	ret := make(map[string]byte)
	for k, v := range consensusState.Committee {
		ret[k] = v
	}
	return ret
}
