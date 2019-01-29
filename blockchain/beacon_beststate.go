package blockchain

import (
	"fmt"
	"sort"

	"github.com/ninjadotorg/constant/blockchain/params"
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
	BestBlockHash   common.Hash  // The hash of the block.
	BestBlock       *BeaconBlock // The block.
	BestShardHash   map[byte]common.Hash
	BestShardHeight map[byte]uint64
	// New field
	//TODO: calculate hash
	AllShardState map[byte][]ShardState

	BeaconEpoch            uint64
	BeaconHeight           uint64
	BeaconProposerIdx      int
	BeaconCommittee        []string
	BeaconPendingValidator []string

	// assigned candidate
	// function as a snapshot list, waiting for random
	CandidateShardWaitingForCurrentRandom  []string
	CandidateBeaconWaitingForCurrentRandom []string

	// assigned candidate
	CandidateShardWaitingForNextRandom  []string
	CandidateBeaconWaitingForNextRandom []string

	// ShardCommittee && ShardPendingValidator will be verify from shardBlock
	// validator of shards
	ShardCommittee map[byte][]string
	// pending validator of shards
	ShardPendingValidator map[byte][]string

	// UnassignBeaconCandidate []strings
	// UnassignShardCandidate  []string

	CurrentRandomNumber int64
	// random timestamp for this epoch
	CurrentRandomTimeStamp int64
	IsGetRandomNumber      bool

	Params        map[string]string
	StabilityInfo StabilityInfo

	// lock sync.RWMutex
}

type StabilityInfo struct {
	SalaryFund uint64 // use to pay salary for miners(block producer or current leader) in chain
	BankFund   uint64 // for DBank

	GOVConstitution GOVConstitution // params which get from governance for network
	DCBConstitution DCBConstitution

	// BOARD
	DCBGovernor DCBGovernor
	GOVGovernor GOVGovernor

	// Price feeds through Oracle
	Oracle params.Oracle
}

func (si StabilityInfo) GetBytes() []byte {
	return common.GetBytes(si)
}

func NewBestStateBeacon() *BestStateBeacon {
	bestStateBeacon := BestStateBeacon{}
	bestStateBeacon.BestBlockHash.SetBytes(make([]byte, 32))
	bestStateBeacon.BestBlock = nil
	bestStateBeacon.BestShardHash = make(map[byte]common.Hash)
	bestStateBeacon.BestShardHeight = make(map[byte]uint64)
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
	bestStateBeacon.StabilityInfo = StabilityInfo{}
	return &bestStateBeacon
}

func (self *BestStateBeacon) Hash() common.Hash {
	var keys []int
	var keyStrs []string
	res := []byte{}
	res = append(res, self.BestBlockHash.GetBytes()...)
	res = append(res, self.BestBlock.Hash().GetBytes()...)
	res = append(res, self.BestBlock.Hash().GetBytes()...)

	for k, _ := range self.BestShardHash {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		hash := self.BestShardHash[byte(shardID)]
		res = append(res, hash.GetBytes()...)
	}
	keys = []int{}
	for k, _ := range self.BestShardHeight {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		height := self.BestShardHeight[byte(shardID)]
		res = append(res, byte(height))
	}
	for _, value := range self.BeaconCommittee {
		res = append(res, []byte(value)...)
	}
	for _, value := range self.BeaconPendingValidator {
		res = append(res, []byte(value)...)
	}
	for _, value := range self.CandidateBeaconWaitingForCurrentRandom {
		res = append(res, []byte(value)...)
	}
	for _, value := range self.CandidateBeaconWaitingForNextRandom {
		res = append(res, []byte(value)...)
	}
	for _, value := range self.CandidateShardWaitingForCurrentRandom {
		res = append(res, []byte(value)...)
	}
	for _, value := range self.CandidateShardWaitingForNextRandom {
		res = append(res, []byte(value)...)
	}
	keys = []int{}
	for k, _ := range self.ShardCommittee {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range self.ShardCommittee[byte(shardID)] {
			res = append(res, []byte(value)...)
		}
	}
	keys = []int{}
	for k, _ := range self.ShardPendingValidator {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range self.ShardPendingValidator[byte(shardID)] {
			res = append(res, []byte(value)...)
		}
	}
	res = append(res, byte(self.CurrentRandomNumber))
	res = append(res, byte(self.CurrentRandomTimeStamp))
	if self.IsGetRandomNumber {
		res = append(res, []byte("true")...)
	} else {
		res = append(res, []byte("false")...)
	}
	for k, _ := range self.Params {
		keyStrs = append(keyStrs, k)
	}
	sort.Strings(keyStrs)
	for _, key := range keyStrs {
		res = append(res, []byte(self.Params[key])...)
	}
	res = append(res, self.StabilityInfo.GetBytes()...)
	return common.DoubleHashH(res)
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
		tmpID := (self.BeaconProposerIdx + 1) % len(self.BeaconCommittee)
		fmt.Println("Producer idx:", tmpID, self.BeaconCommittee)
		if found == tmpID {
			return "beacon-proposer", 0
		}
		return "beacon-validator", 0
	}

	found = common.IndexOfStr(pubkey, self.BeaconPendingValidator)
	if found > -1 {
		return "beacon-pending", 0
	}

	return "", 0
}
