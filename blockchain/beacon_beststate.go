package blockchain

import (
	"github.com/ninjadotorg/constant/common"
)

// BestState houses information about the current best block and other info
// related to the state of the main chain as it exists from the point of view of
// the current best block.
//
// The BestSnapshot method can be used to obtain access to this information
// in a concurrent safe manner and the data will not be changed out from under
// the caller when chain state changes occur as the function name implies.
// However, the returned snapshot must be treated as immutable since it is
// shared by all callers.
type BestStateBeacon struct {
	BestBlockHash common.Hash  // The hash of the block.
	BestBlock     *BeaconBlock // The block.
	BestShardHash []common.Hash

	BeaconEpoch       uint64
	BeaconHeight      uint64
	BeaconLeaderIndex uint64

	BeaconCommittee        []string
	BeaconPendingValidator []string

	// assigned candidate
	// function as a snapshot list, waiting for random
	CandidateShardWaitingForCurrentRandom  []string
	CandidateBeaconWaitingForCurrentRandom []string

	// assigned candidate
	CandidateShardWaitingForNextRandom  []string
	CandidateBeaconWaitingForNextRandom []string

	// validator of shards
	ShardCommittee map[byte][]string
	// pending validator of shards
	ShardPendingValidator map[byte][]string

	// UnassignBeaconCandidate []string
	// UnassignShardCandidate  []string

	CurrentRandomNumber int64
	// random timestamp for this epoch
	CurrentRandomTimeStamp int64
	// NextRandomNumber    int64

	Params map[string]string
}

func NewBestStateBeacon() *BestStateBeacon {
	bestStateBeacon := BestStateBeacon{}
	bestStateBeacon.BestBlockHash.SetBytes(make([]byte, 32))
	bestStateBeacon.BestBlock = nil
	bestStateBeacon.BestShardHash = []common.Hash{}
	bestStateBeacon.BeaconHeight = 0
	bestStateBeacon.BeaconCommittee = []string{}
	bestStateBeacon.BeaconPendingValidator = []string{}

	bestStateBeacon.CandidateShardWaitingForCurrentRandom = []string{}
	bestStateBeacon.CandidateBeaconWaitingForCurrentRandom = []string{}

	bestStateBeacon.CandidateShardWaitingForNextRandom = []string{}
	bestStateBeacon.CandidateBeaconWaitingForNextRandom = []string{}

	bestStateBeacon.ShardCommittee = make(map[byte][]string)
	bestStateBeacon.ShardPendingValidator = make(map[byte][]string)
	bestStateBeacon.Params = make(map[string]string)
	bestStateBeacon.CurrentRandomNumber = -1
	return &bestStateBeacon
}

// Get role of a public key base on best state beacond
// return node-role, <shardID>
func (self *BestStateBeacon) GetPubkeyRole(pubkey string) (string, byte) {

	for shardID, pubkeyArr := range self.ShardPendingValidator {
		found := common.IndexOfStr(pubkey, pubkeyArr)
		if found > -1 {
			return "shard", shardID
		}
	}

	for shardID, pubkeyArr := range self.ShardCommittee {
		found := common.IndexOfStr(pubkey, pubkeyArr)
		if found > -1 {
			return "shard", shardID
		}
	}

	found := common.IndexOfStr(pubkey, self.BeaconCommittee)
	if found > -1 {
		if uint64(found) == self.BeaconLeaderIndex+1 {
			return "beacon-leader", 0
		} else {
			return "beacon-validator", 0
		}

	}

	found = common.IndexOfStr(pubkey, self.BeaconPendingValidator)
	if found > -1 {
		return "beacon-pending", 0
	}

	return "", 0
}
