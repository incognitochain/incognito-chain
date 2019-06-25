package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"github.com/incognitochain/incognito-chain/blockchain/btc"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"

	"github.com/incognitochain/incognito-chain/cashec"
	"github.com/incognitochain/incognito-chain/common/base58"
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

var bestStateBeacon *BestStateBeacon

type BestStateBeacon struct {
	BestBlockHash                          common.Hash          `json:"BestBlockHash"`     // The hash of the block.
	PrevBestBlockHash                      common.Hash          `json:"PrevBestBlockHash"` // The hash of the block.
	BestBlock                              BeaconBlock          `json:"BestBlock"`         // The block.
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
	BeaconCommitteeSize                    int                  `json:"BeaconCommitteeSize"`
	ShardCommitteeSize                     int                  `json:"ShardCommitteeSize"`
	ActiveShards                           int                  `json:"ActiveShards"`
	// cross shard state for all the shard. from shardID -> to crossShard shardID -> last height
	// e.g 1 -> 2 -> 3 // shard 1 send cross shard to shard 2 at  height 3
	// e.g 1 -> 3 -> 2 // shard 1 send cross shard to shard 3 at  height 2
	LastCrossShardState map[byte]map[byte]uint64 `json:"LastCrossShardState"`
	ShardHandle map[byte]bool `json:"ShardHandle"` // lock sync.RWMutex
	lockMu      sync.RWMutex
	randomClient btc.RandomClient
}
func (bestStateBeacon *BestStateBeacon) InitRandomClient (randomClient btc.RandomClient) {
	bestStateBeacon.randomClient = randomClient
}
func (bestStateBeacon *BestStateBeacon) MarshalJSON() ([]byte, error) {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()

	type Alias BestStateBeacon
	b, err := json.Marshal(&struct {
		*Alias
	}{
		(*Alias)(bestStateBeacon),
	})
	if err != nil {
		Logger.log.Error(err)
	}
	return b, err
}

func (bestStateBeacon *BestStateBeacon) GetBestShardHeight() map[byte]uint64 {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()
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

func (bestStateBeacon *BestStateBeacon) GetAShardCommittee(shardID byte) []string {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()
	return bestStateBeacon.ShardCommittee[shardID]
}

func (bestStateBeacon *BestStateBeacon) GetShardCommittee() (res map[byte][]string) {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()
	res = make(map[byte][]string)
	for index, element := range bestStateBeacon.ShardCommittee {
		res[index] = element
	}
	return res
}

func (bestStateBeacon *BestStateBeacon) GetAShardPendingValidator(shardID byte) []string {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()
	return bestStateBeacon.ShardPendingValidator[shardID]
}

func (bestStateBeacon *BestStateBeacon) GetShardPendingValidator() (res map[byte][]string) {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()
	res = make(map[byte][]string)
	for index, element := range bestStateBeacon.ShardPendingValidator {
		res[index] = element
	}
	return res
}

func (bsb *BestStateBeacon) GetCurrentShard() byte {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()
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
	bestStateBeacon.BestBlockHash.SetBytes(make([]byte, 32))
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
	bestStateBeacon.BeaconCommitteeSize = netparam.BeaconCommitteeSize
	bestStateBeacon.ShardCommitteeSize = netparam.ShardCommitteeSize
	bestStateBeacon.ActiveShards = netparam.ActiveShards
	bestStateBeacon.LastCrossShardState = make(map[byte]map[byte]uint64)
	return bestStateBeacon
}
func (bestStateBeacon *BestStateBeacon) GetBytes() []byte {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()
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
	return common.HashH(bestStateBeacon.GetBytes())
}

// Get role of a public key base on best state beacond
// return node-role, <shardID>
func (bestStateBeacon *BestStateBeacon) GetPubkeyRole(pubkey string, round int) (string, byte) {
	bestStateBeacon.lockMu.RLock()
	defer bestStateBeacon.lockMu.RUnlock()
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

func (blockchain *BlockChain) ValidateBlockWithPrevBeaconBestState(block *BeaconBlock) error {
	prevBST, err := blockchain.config.DataBase.FetchPrevBestState(true, 0)
	if err != nil {
		return err
	}
	beaconBestState := BestStateBeacon{}
	if err := json.Unmarshal(prevBST, &beaconBestState); err != nil {
		return err
	}

	blkHash := block.Header.Hash()
	producerPk := base58.Base58Check{}.Encode(block.Header.ProducerAddress.Pk, common.ZeroByte)
	err = cashec.ValidateDataB58(producerPk, block.ProducerSig, blkHash.GetBytes())
	if err != nil {
		return NewBlockChainError(ProducerError, errors.New("Producer's sig not match"))
	}
	//verify producer
	producerPosition := (beaconBestState.BeaconProposerIdx + block.Header.Round) % len(beaconBestState.BeaconCommittee)
	tempProducer := beaconBestState.BeaconCommittee[producerPosition]
	if strings.Compare(tempProducer, producerPk) != 0 {
		return NewBlockChainError(ProducerError, errors.New("Producer should be should be :"+tempProducer))
	}
	//verify version
	if block.Header.Version != VERSION {
		return NewBlockChainError(VersionError, errors.New("Version should be :"+strconv.Itoa(VERSION)))
	}
	prevBlockHash := block.Header.PrevBlockHash
	// Verify parent hash exist or not
	parentBlockBytes, err := blockchain.config.DataBase.FetchBeaconBlock(prevBlockHash)
	if err != nil {
		return NewBlockChainError(DBError, err)
	}
	parentBlock := NewBeaconBlock()
	json.Unmarshal(parentBlockBytes, &parentBlock)
	// Verify block height with parent block
	if parentBlock.Header.Height+1 != block.Header.Height {
		return NewBlockChainError(BlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
	}

	return nil
}

//This only happen if user is a beacon committee member.
func (blockchain *BlockChain) RevertBeaconState() error {
	//Steps:
	// 1. Restore current beststate to previous beststate
	// 2. Set beacon/shardtobeacon pool state
	// 3. Delete newly inserted block
	// 4. Delete data store by block
	blockchain.chainLock.Lock()
	defer blockchain.chainLock.Unlock()
	currentBestState := blockchain.BestState.Beacon
	currentBestStateBlk := currentBestState.BestBlock

	prevBST, err := blockchain.config.DataBase.FetchPrevBestState(true, 0)
	if err != nil {
		return err
	}
	beaconBestState := BestStateBeacon{}
	if err := json.Unmarshal(prevBST, &beaconBestState); err != nil {
		return err
	}

	blockchain.config.BeaconPool.SetBeaconState(beaconBestState.BeaconHeight)
	blockchain.config.ShardToBeaconPool.SetShardState(blockchain.BestState.Beacon.GetBestShardHeight())

	if err := blockchain.config.DataBase.DeleteCommitteeByEpoch(currentBestStateBlk.Header.Height); err != nil {
		return err
	}

	for shardID, shardStates := range currentBestStateBlk.Body.ShardState {
		for _, shardState := range shardStates {
			blockchain.config.DataBase.DeleteAcceptedShardToBeacon(shardID, shardState.Hash)
		}
	}

	lastCrossShardState := beaconBestState.LastCrossShardState
	for fromShard, toShards := range lastCrossShardState {
		for toShard, height := range toShards {
			blockchain.config.DataBase.RestoreCrossShardNextHeights(fromShard, toShard, height)
		}
		blockchain.config.CrossShardPool[fromShard].UpdatePool()
	}

	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}
	for _, inst := range currentBestStateBlk.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		var err error
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			fmt.Printf("[ndh] error - - %+v\n", err)
			return err
		}
		switch metaType {
		case metadata.IssuingRequestMeta:
			updatingInfoByTokenID, err = blockchain.processIssuingReq(inst, updatingInfoByTokenID)
		case metadata.ContractingRequestMeta:
			updatingInfoByTokenID, err = blockchain.processContractingReq(inst, updatingInfoByTokenID)
		case metadata.AcceptedBlockRewardInfoMeta:
			fmt.Printf("[ndh] - - %+v\n", inst[2])
			acceptedBlkRewardInfo, err := metadata.NewAcceptedBlockRewardInfoFromStr(inst[2])
			if err != nil {
				fmt.Printf("[ndh] error1 - - %+v\n", err)
				return err
			}
			if val, ok := acceptedBlkRewardInfo.TxsFee[common.PRVCoinID]; ok {
				acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] = val + blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)
			} else {
				if acceptedBlkRewardInfo.TxsFee == nil {
					acceptedBlkRewardInfo.TxsFee = map[common.Hash]uint64{}
				}
				acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] = blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)
			}
			for key, value := range acceptedBlkRewardInfo.TxsFee {
				fmt.Printf("[ndh] - - - zzzzzzzzzzzzzzzzzzzzzzzz epoch %+v, shardID %+v %+v %+v\n", currentBestStateBlk.Header.Epoch, acceptedBlkRewardInfo.ShardID, key, value)
				err = blockchain.config.DataBase.RestoreShardRewardRequest(currentBestStateBlk.Header.Epoch, acceptedBlkRewardInfo.ShardID, key)
				if err != nil {
					return err
				}

			}
		}
		if err != nil {
			return err
		}
	}
	for tokenID, _ := range updatingInfoByTokenID {
		err := blockchain.config.DataBase.RestoreBridgedTokenByTokenID(tokenID)
		if err != nil {
			return err
		}
	}

	blockchain.config.DataBase.DeleteBeaconBlock(currentBestStateBlk.Header.Hash(), currentBestStateBlk.Header.Height)
	blockchain.BestState.Beacon = &beaconBestState
	if err := blockchain.StoreBeaconBestState(); err != nil {
		return err
	}
	return nil
}

func (blockchain *BlockChain) BackupCurrentBeaconState(block *BeaconBlock) error {
	//Steps:
	// 1. Backup beststate
	tempMarshal, err := json.Marshal(blockchain.BestState.Beacon)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBlockError, err)
	}

	if err := blockchain.config.DataBase.StorePrevBestState(tempMarshal, true, 0); err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	updatingInfoByTokenID := map[common.Hash]UpdatingInfo{}
	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		var err error
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			fmt.Printf("[ndh] error - - %+v\n", err)
			return err
		}

		switch metaType {
		case metadata.IssuingRequestMeta:
			updatingInfoByTokenID, err = blockchain.processIssuingReq(inst, updatingInfoByTokenID)
		case metadata.ContractingRequestMeta:
			updatingInfoByTokenID, err = blockchain.processContractingReq(inst, updatingInfoByTokenID)
		case metadata.AcceptedBlockRewardInfoMeta:
			fmt.Printf("[ndh] - - %+v\n", inst[2])
			acceptedBlkRewardInfo, err := metadata.NewAcceptedBlockRewardInfoFromStr(inst[2])
			if err != nil {
				fmt.Printf("[ndh] error1 - - %+v\n", err)
				return err
			}
			if val, ok := acceptedBlkRewardInfo.TxsFee[common.PRVCoinID]; ok {
				acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] = val + blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)
			} else {
				if acceptedBlkRewardInfo.TxsFee == nil {
					acceptedBlkRewardInfo.TxsFee = map[common.Hash]uint64{}
				}
				acceptedBlkRewardInfo.TxsFee[common.PRVCoinID] = blockchain.getRewardAmount(acceptedBlkRewardInfo.ShardBlockHeight)
			}
			for key, value := range acceptedBlkRewardInfo.TxsFee {
				fmt.Printf("[ndh] - - - zzzzzzzzzzzzzzzzzzzzzzzz epoch %+v, shardID %+v %+v %+v\n", block.Header.Epoch, acceptedBlkRewardInfo.ShardID, key, value)
				err = blockchain.config.DataBase.BackupShardRewardRequest(block.Header.Epoch, acceptedBlkRewardInfo.ShardID, key)
				if err != nil {
					return err
				}

			}
		}

		if err != nil {
			return err
		}
	}
	for tokenID, _ := range updatingInfoByTokenID {
		err := blockchain.config.DataBase.BackupBridgedTokenByTokenID(tokenID)
		if err != nil {
			return err
		}
	}

	return nil
}
