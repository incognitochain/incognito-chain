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

// getAssetPrice returns price stored in Oracle
func (self *BestStateBeacon) getAssetPrice(assetID common.Hash) uint64 {
	price := uint64(0)
	if common.IsBondAsset(&assetID) {
		if self.StabilityInfo.Oracle.Bonds != nil {
			price = self.StabilityInfo.Oracle.Bonds[assetID.String()]
		}
	} else {
		oracle := self.StabilityInfo.Oracle
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
func (self *BestStateBeacon) GetSaleData(saleID []byte) (*params.SaleData, error) {
	key := getSaleDataKeyBeacon(saleID)
	if value, ok := self.Params[key]; ok {
		return parseSaleDataValueBeacon(value)
	}
	return nil, errors.Errorf("SaleID not exist: %x", saleID)
}

func (self *BestStateBeacon) GetLatestDividendProposal(forDCB bool) (id, amount uint64) {
	key := ""
	if forDCB {
		key = getDCBDividendKeyBeacon()
	} else {
		// TODO: GOV
	}
	dividendAmounts := []uint64{}
	if value, ok := self.Params[key]; ok {
		dividendAmounts, _ = parseDividendValueBeacon(value)
		if len(dividendAmounts) > 0 {
			id = uint64(len(dividendAmounts))
			amount = dividendAmounts[len(dividendAmounts)-1]
		}
	}
	return id, amount
}

func (self *BestStateBeacon) GetDividendAggregatedInfo(dividendID uint64, tokenID *common.Hash) (uint64, uint64, error) {
	key := getDividendAggregatedKeyBeacon(dividendID, tokenID)
	if value, ok := self.Params[key]; ok {
		totalTokenOnAllShards, cstToPayout := parseDividendAggregatedValueBeacon(value)
		value = getDividendAggregatedValueBeacon(totalTokenOnAllShards, cstToPayout)
		return totalTokenOnAllShards, cstToPayout, nil
	} else {
		return 0, 0, errors.Errorf("Aggregated dividend info not found for id %d, tokenID %x", dividendID, tokenID.String())
	}

}
