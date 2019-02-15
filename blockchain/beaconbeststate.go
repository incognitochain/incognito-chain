package blockchain

import (
	"sort"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/pkg/errors"
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
	BestBlockHash   common.Hash          `json:"BestBlockHash"` // The hash of the block.
	BestBlock       *BeaconBlock         `json:"BestBlock"`     // The block.
	BestShardHash   map[byte]common.Hash `json:"BestShardHash"`
	BestShardHeight map[byte]uint64      `json:"BestShardHeight"`
	// New field
	//TODO: calculate hash
	AllShardState map[byte][]ShardState `json:"AllShardState"`

	BeaconEpoch            uint64   `json:"BeaconEpoch"`
	BeaconHeight           uint64   `json:"BeaconHeight"`
	BeaconProposerIdx      int      `json:"BeaconProposerIdx"`
	BeaconCommittee        []string `json:"BeaconCommittee"`
	BeaconPendingValidator []string `json:"BeaconPendingValidator"`

	// assigned candidate
	// function as a snapshot list, waiting for random
	CandidateShardWaitingForCurrentRandom  []string `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateBeaconWaitingForCurrentRandom []string `json:"CandidateBeaconWaitingForCurrentRandom"`

	// assigned candidate
	CandidateShardWaitingForNextRandom  []string `json:"CandidateShardWaitingForNextRandom"`
	CandidateBeaconWaitingForNextRandom []string `json:"CandidateBeaconWaitingForNextRandom"`

	// ShardCommittee && ShardPendingValidator will be verify from shardBlock
	// validator of shards
	ShardCommittee map[byte][]string `json:"ShardCommittee"`
	// pending validator of shards
	ShardPendingValidator map[byte][]string `json:"ShardPendingValidator"`

	// UnassignBeaconCandidate []strings
	// UnassignShardCandidate  []string

	CurrentRandomNumber int64 `json:"CurrentRandomNumber"`
	// random timestamp for this epoch
	CurrentRandomTimeStamp int64 `json:"CurrentRandomTimeStamp"`
	IsGetRandomNumber      bool  `json:"IsGetRandomNumber"`

	Params        map[string]string `json:"Params,omitempty"`
	StabilityInfo StabilityInfo     `json:"StabilityInfo"`

	// lock sync.RWMutex
	ShardHandle map[byte]bool `json:"ShardHandle"`
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

func (bsb *BestStateBeacon) GetCurrentShard() byte {
	for shardID, isCurrent := range bsb.ShardHandle {
		if isCurrent {
			return shardID
		}
	}
	return 0
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

func (bestStateBeacon *BestStateBeacon) Hash() common.Hash {
	var keys []int
	var keyStrs []string
	res := []byte{}
	res = append(res, bestStateBeacon.BestBlockHash.GetBytes()...)
	res = append(res, bestStateBeacon.BestBlock.Hash().GetBytes()...)
	res = append(res, bestStateBeacon.BestBlock.Hash().GetBytes()...)

	for k := range bestStateBeacon.BestShardHash {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		hash := bestStateBeacon.BestShardHash[byte(shardID)]
		res = append(res, hash.GetBytes()...)
	}
	keys = []int{}
	for k := range bestStateBeacon.BestShardHeight {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		height := bestStateBeacon.BestShardHeight[byte(shardID)]
		res = append(res, byte(height))
	}
	for _, value := range bestStateBeacon.BeaconCommittee {
		res = append(res, []byte(value)...)
	}
	for _, value := range bestStateBeacon.BeaconPendingValidator {
		res = append(res, []byte(value)...)
	}
	for _, value := range bestStateBeacon.CandidateBeaconWaitingForCurrentRandom {
		res = append(res, []byte(value)...)
	}
	for _, value := range bestStateBeacon.CandidateBeaconWaitingForNextRandom {
		res = append(res, []byte(value)...)
	}
	for _, value := range bestStateBeacon.CandidateShardWaitingForCurrentRandom {
		res = append(res, []byte(value)...)
	}
	for _, value := range bestStateBeacon.CandidateShardWaitingForNextRandom {
		res = append(res, []byte(value)...)
	}
	keys = []int{}
	for k := range bestStateBeacon.ShardCommittee {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range bestStateBeacon.ShardCommittee[byte(shardID)] {
			res = append(res, []byte(value)...)
		}
	}
	keys = []int{}
	for k := range bestStateBeacon.ShardPendingValidator {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range bestStateBeacon.ShardPendingValidator[byte(shardID)] {
			res = append(res, []byte(value)...)
		}
	}
	res = append(res, byte(bestStateBeacon.CurrentRandomNumber))
	res = append(res, byte(bestStateBeacon.CurrentRandomTimeStamp))
	if bestStateBeacon.IsGetRandomNumber {
		res = append(res, []byte("true")...)
	} else {
		res = append(res, []byte("false")...)
	}
	for k := range bestStateBeacon.Params {
		keyStrs = append(keyStrs, k)
	}
	sort.Strings(keyStrs)
	for _, key := range keyStrs {
		res = append(res, []byte(bestStateBeacon.Params[key])...)
	}
	res = append(res, bestStateBeacon.StabilityInfo.GetBytes()...)
	return common.DoubleHashH(res)
}

// Get role of a public key base on best state beacond
// return node-role, <shardID>
// TODO: Role name should be write in common as constant value
func (bestStateBeacon *BestStateBeacon) GetPubkeyRole(pubkey string) (string, byte) {

	for shardID, pubkeyArr := range bestStateBeacon.ShardPendingValidator {
		found := common.IndexOfStr(pubkey, pubkeyArr)
		if found > -1 {
			return common.SHARD_ROLE, shardID
		}
	}

	for shardID, pubkeyArr := range bestStateBeacon.ShardCommittee {
		found := common.IndexOfStr(pubkey, pubkeyArr)
		if found > -1 {
			return common.SHARD_ROLE, shardID
		}
	}

	found := common.IndexOfStr(pubkey, bestStateBeacon.BeaconCommittee)
	if found > -1 {
		tmpID := (bestStateBeacon.BeaconProposerIdx + 1) % len(bestStateBeacon.BeaconCommittee)
		if found == tmpID {
			return common.BEACON_PROPOSER_ROLE, 0
		}
		return common.BEACON_VALIDATOR_ROLE, 0
	}

	found = common.IndexOfStr(pubkey, bestStateBeacon.BeaconPendingValidator)
	if found > -1 {
		return common.BEACON_PENDING_ROLE, 0
	}

	return common.EmptyString, 0
}

// getAssetPrice returns price stored in Oracle
func (bestStateBeacon *BestStateBeacon) getAssetPrice(assetID common.Hash) uint64 {
	price := uint64(0)
	if common.IsBondAsset(&assetID) {
		if bestStateBeacon.StabilityInfo.Oracle.Bonds != nil {
			price = bestStateBeacon.StabilityInfo.Oracle.Bonds[assetID.String()]
		}
	} else {
		oracle := bestStateBeacon.StabilityInfo.Oracle
		if assetID.IsEqual(&common.ConstantID) {
			price = oracle.Constant
		} else if assetID.IsEqual(&common.DCBTokenID) {
			price = oracle.DCBToken
		} else if assetID.IsEqual(&common.GOVTokenID) {
			price = oracle.GOVToken
		} else if assetID.IsEqual(&common.ETHAssetID) {
			price = oracle.ETH
		} else if assetID.IsEqual(&common.BTCAssetID) {
			price = oracle.BTC
		}
	}
	return price
}

// GetSaleData returns latest data of a crowdsale
func (bestStateBeacon *BestStateBeacon) GetSaleData(saleID []byte) (*params.SaleData, error) {
	key := getSaleDataKeyBeacon(saleID)
	if value, ok := bestStateBeacon.Params[key]; ok {
		return parseSaleDataValueBeacon(value)
	}
	return nil, errors.Errorf("SaleID not exist: %x", saleID)
}
