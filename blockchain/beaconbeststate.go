package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"sort"
	"strconv"
	"sync"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
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

var bestStateBeacon *BestStateBeacon //singleton object

type BestStateBeacon struct {
	BestBlockHash                          common.Hash          `json:"BestBlockHash"`     // The hash of the block.
	PrevBestBlockHash                      common.Hash          `json:"PrevBestBlockHash"` // The hash of the block.
	BestBlock                              *BeaconBlock         `json:"BestBlock"`         // The block.
	BestShardHash                          map[byte]common.Hash `json:"BestShardHash"`
	BestShardHeight                        map[byte]uint64      `json:"BestShardHeight"`
	Epoch                                  uint64               `json:"Epoch"`
	BeaconHeight                           uint64               `json:"BeaconHeight"`
	BeaconProposerIdx                      int                  `json:"BeaconProposerIdx"`
	BeaconCommittee                        []string             `json:"BeaconCommittee"`
	BeaconPendingValidator                 []string             `json:"BeaconPendingValidator"`
	CandidateShardWaitingForCurrentRandom  []string             `json:"CandidateShardWaitingForCurrentRandom"` // snapshot shard candidate list, waiting to be shuffled in this current epoch
	CandidateBeaconWaitingForCurrentRandom []string             `json:"CandidateBeaconWaitingForCurrentRandom"`
	CandidateShardWaitingForNextRandom     []string             `json:"CandidateShardWaitingForNextRandom"` // shard candidate list, waiting to be shuffled in next epoch
	CandidateBeaconWaitingForNextRandom    []string             `json:"CandidateBeaconWaitingForNextRandom"`
	ShardCommittee                         map[byte][]string    `json:"ShardCommittee"`        // current committee and validator of all shard
	ShardPendingValidator                  map[byte][]string    `json:"ShardPendingValidator"` // pending candidate waiting for swap to get in committee of all shard
	CurrentRandomNumber                    int64                `json:"CurrentRandomNumber"`
	CurrentRandomTimeStamp                 int64                `json:"CurrentRandomTimeStamp"` // random timestamp for this epoch
	IsGetRandomNumber                      bool                 `json:"IsGetRandomNumber"`
	Params                                 map[string]string    `json:"Params,omitempty"`
	StabilityInfo                          StabilityInfo        `json:"StabilityInfo"`
	BeaconCommitteeSize                    int                  `json:"BeaconCommitteeSize"`
	ShardCommitteeSize                     int                  `json:"ShardCommitteeSize"`
	ActiveShards                           int                  `json:"ActiveShards"`
	// cross shard state for all the shard. from shardID -> to crossShard shardID -> last height
	// e.g 1 -> 2 -> 3 // shard 1 send cross shard to shard 2 at  height 3
	// e.g 1 -> 3 -> 2 // shard 1 send cross shard to shard 3 at  height 2
	LastCrossShardState map[byte]map[byte]uint64 `json:"LastCrossShardState"`

	ShardHandle map[byte]bool `json:"ShardHandle"` // lock sync.RWMutex
	lockMu      sync.RWMutex
}

type StabilityInfo struct {
	SalaryFund uint64 // use to pay salary for miners(block producer or current leader) in chain
	BankFund   uint64 // for DBank

	GOVConstitution GOVConstitution // component which get from governance for network
	DCBConstitution DCBConstitution

	// BOARD
	DCBGovernor DCBGovernor
	GOVGovernor GOVGovernor

	// Price feeds through Oracle
	Oracle component.Oracle
}

func (si StabilityInfo) GetBytes() []byte {
	return common.GetBytes(si)
}

func (bestStateBeacon *BestStateBeacon) GetBestShardHeight() map[byte]uint64 {
	res := make(map[byte]uint64)
	for index, element := range bestStateBeacon.BestShardHeight {
		res[index] = element
	}
	return res
}

func (bestStateBeacon *BestStateBeacon) GetBestHeightOfShard(shardID byte) uint64 {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()
	return bestStateBeacon.BestShardHeight[shardID]
}

func (bsb *BestStateBeacon) GetCurrentShard() byte {
	for shardID, isCurrent := range bsb.ShardHandle {
		if isCurrent {
			return shardID
		}
	}
	return 0
}

func SetBestStateBeacon(beacon *BestStateBeacon) {
	bestStateBeacon = beacon
}

func GetBestStateBeacon() *BestStateBeacon {
	if bestStateBeacon != nil {
		return bestStateBeacon
	}
	bestStateBeacon = &BestStateBeacon{}
	return bestStateBeacon
}

func InitBestStateBeacon(netparam *Params) *BestStateBeacon {
	if bestStateBeacon == nil {
		bestStateBeacon = GetBestStateBeacon()
	}
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
	bestStateBeacon.BeaconCommitteeSize = netparam.BeaconCommitteeSize
	bestStateBeacon.ShardCommitteeSize = netparam.ShardCommitteeSize
	bestStateBeacon.ActiveShards = netparam.ActiveShards
	bestStateBeacon.LastCrossShardState = make(map[byte]map[byte]uint64)
	return bestStateBeacon
}
func (bestStateBeacon *BestStateBeacon) GetBytes() []byte {
	var keys []int
	var keyStrs []string
	res := []byte{}
	res = append(res, bestStateBeacon.BestBlockHash.GetBytes()...)
	res = append(res, bestStateBeacon.PrevBestBlockHash.GetBytes()...)
	res = append(res, bestStateBeacon.BestBlock.Hash().GetBytes()...)
	res = append(res, bestStateBeacon.BestBlock.Header.PrevBlockHash.GetBytes()...)
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
	EpochBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(EpochBytes, bestStateBeacon.Epoch)
	res = append(res, EpochBytes...)
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, bestStateBeacon.BeaconHeight)
	res = append(res, heightBytes...)
	res = append(res, []byte(strconv.Itoa(bestStateBeacon.BeaconProposerIdx))...)
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

	randomNumBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(randomNumBytes, uint64(bestStateBeacon.CurrentRandomNumber))
	res = append(res, randomNumBytes...)

	randomTimeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(randomTimeBytes, uint64(bestStateBeacon.CurrentRandomTimeStamp))
	res = append(res, randomTimeBytes...)

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

	//TODO: @stability
	//res = append(res, bestStateBeacon.StabilityInfo.GetBytes()...)
	//return common.HashH(res)

	keys = []int{}
	for k := range bestStateBeacon.ShardHandle {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		shardHandleItem := bestStateBeacon.ShardHandle[byte(shardID)]
		if shardHandleItem {
			res = append(res, []byte("true")...)
		} else {
			res = append(res, []byte("false")...)
		}
	}
	res = append(res, []byte(strconv.Itoa(bestStateBeacon.BeaconCommitteeSize))...)
	res = append(res, []byte(strconv.Itoa(bestStateBeacon.ShardCommitteeSize))...)
	res = append(res, []byte(strconv.Itoa(bestStateBeacon.ActiveShards))...)

	keys = []int{}
	for k := range bestStateBeacon.LastCrossShardState {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, fromShard := range keys {
		fromShardMap := bestStateBeacon.LastCrossShardState[byte(fromShard)]
		newKeys := []int{}
		for k := range fromShardMap {
			newKeys = append(newKeys, int(k))
		}
		sort.Ints(newKeys)
		for _, toShard := range newKeys {
			value := fromShardMap[byte(toShard)]
			valueBytes := make([]byte, 8)
			binary.LittleEndian.PutUint64(valueBytes, value)
			res = append(res, valueBytes...)
		}
	}
	return res
}
func (bestStateBeacon *BestStateBeacon) Hash() common.Hash {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()
	return common.HashH(bestStateBeacon.GetBytes())
}

// Get role of a public key base on best state beacond
// return node-role, <shardID>
func (bestStateBeacon *BestStateBeacon) GetPubkeyRole(pubkey string, round int) (string, byte) {
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
		tmpID := (bestStateBeacon.BeaconProposerIdx + round) % len(bestStateBeacon.BeaconCommittee)
		if found == tmpID {
			return common.PROPOSER_ROLE, 0
		}
		return common.VALIDATOR_ROLE, 0
	}

	found = common.IndexOfStr(pubkey, bestStateBeacon.BeaconPendingValidator)
	if found > -1 {
		return common.PENDING_ROLE, 0
	}

	return common.EmptyString, 0
}

// GetAssetPrice returns price stored in Oracle
func (bestStateBeacon *BestStateBeacon) GetAssetPrice(assetID common.Hash) uint64 {
	price := uint64(0)
	if common.IsBondAsset(&assetID) {
		if bestStateBeacon.StabilityInfo.Oracle.Bonds != nil {
			price = bestStateBeacon.StabilityInfo.Oracle.Bonds[assetID.String()]
		}
	} else {
		oracle := bestStateBeacon.StabilityInfo.Oracle
		if common.IsConstantAsset(&assetID) {
			price = oracle.Constant
		} else if assetID.IsEqual(&common.DCBTokenID) {
			price = oracle.DCBToken
		} else if assetID.IsEqual(&common.GOVTokenID) {
			price = oracle.GOVToken
		} else if assetID.IsEqual(&common.ETHAssetID) {
			price = oracle.ETH
		} else if assetID.IsEqual(&common.BTCAssetID) {
			price = oracle.BTC
		} else if assetID.IsEqual(&common.USDAssetID) {
			price = 1 // Oracle's price is again USD itself
		}
	}
	return price
}

func (blockchain *BlockChain) ValidateBlockWithPrevBeaconBestState(block *BeaconBlock) error {
	return nil
}

//This only happen if user is a beacon committee member.
func (blockchain *BlockChain) RevertBeaconState() {
	//Steps:
	// 1. Restore current beststate to previous beststate
	// 2. Set beacon/shardtobeacon pool state
	// 3. Delete newly inserted block
	// 4. Delete data store by block
	// if err := blockchain.config.DataBase.StoreCommitteeByEpoch(block.Header.Height, blockchain.BestState.Beacon.ShardCommittee); err != nil {
	// 	return err
	// }
	// for shardID, shardStates := range block.Body.ShardState {
	// 	for _, shardState := range shardStates {
	// 		blockchain.config.DataBase.StoreAcceptedShardToBeacon(shardID, block.Header.Height, &shardState.Hash)
	// 	}
	// }
	// //=========Remove beacon block in pool
	// blockchain.config.BeaconPool.SetBeaconState(blockchain.BestState.Beacon.BeaconHeight)

	// //=========Remove shard to beacon block in pool
	// Logger.log.Info("Remove block from pool block with hash  ", *block.Hash(), block.Header.Height, blockchain.BestState.Beacon.BestShardHeight)
	// blockchain.config.ShardToBeaconPool.SetShardState(blockchain.BestState.Beacon.GetBestShardHeight())

	// lastCrossShardState := GetBestStateBeacon().LastCrossShardState
	// if block.Body.ShardState != nil {
	// 	for fromShard, shardBlocks := range block.Body.ShardState {
	// 		go func(fromShard byte, shardBlocks []ShardState) {
	// 			for _, shardBlock := range shardBlocks {
	// 				for _, toShard := range shardBlock.CrossShard {
	// 					if fromShard == toShard {
	// 						continue
	// 					}
	// 					if lastCrossShardState[fromShard] == nil {
	// 						lastCrossShardState[fromShard] = make(map[byte]uint64)
	// 					}
	// 					lastHeight := lastCrossShardState[fromShard][toShard] // get last cross shard height from shardID  to crossShardShardID
	// 					waitHeight := shardBlock.Height
	// 					// fmt.Println("StoreCrossShardNextHeight", fromShard, toShard, lastHeight, waitHeight)
	// 					blockchain.config.DataBase.StoreCrossShardNextHeight(fromShard, toShard, lastHeight, waitHeight)
	// 					//beacon process shard_to_beacon in order so cross shard next height also will be saved in order
	// 					//dont care overwrite this value
	// 					blockchain.config.DataBase.StoreCrossShardNextHeight(fromShard, toShard, waitHeight, 0)
	// 					if lastCrossShardState[fromShard] == nil {
	// 						lastCrossShardState[fromShard] = make(map[byte]uint64)
	// 					}
	// 					lastCrossShardState[fromShard][toShard] = waitHeight //update lastHeight to waitHeight
	// 				}
	// 			}
	// 			blockchain.config.CrossShardPool[fromShard].UpdatePool()
	// 		}(fromShard, shardBlocks)
	// 	}
	// }
}

func (blockchain *BlockChain) BackupCurrentBeaconState(block *BeaconBlock) error {
	//Steps:
	// 1. Backup beststate
	// 2. Backup data that will be modify by new block data
	tempMarshal, err := json.Marshal(blockchain.BestState.Beacon)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}

	if err := blockchain.config.DataBase.StorePrevBestState(tempMarshal, true, 0); err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	// blockchain.config.DataBase.BackupCommitteeByEpoch()
	// blockchain.config.DataBase.BackupCrossShardNextHeight()
	return nil
}
