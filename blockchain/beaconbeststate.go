package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/incognitochain/incognito-chain/blockchain/btc"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/metadata"

	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/incognitokey"
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

type BeaconBestState struct {
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
	MaxBeaconCommitteeSize                 int                  `json:"MaxBeaconCommitteeSize"`
	MinBeaconCommitteeSize                 int                  `json:"MinBeaconCommitteeSize"`
	MaxShardCommitteeSize                  int                  `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize                  int                  `json:"MinShardCommitteeSize"`
	ActiveShards                           int                  `json:"ActiveShards"`
	// cross shard state for all the shard. from shardID -> to crossShard shardID -> last height
	// e.g 1 -> 2 -> 3 // shard 1 send cross shard to shard 2 at  height 3
	// e.g 1 -> 3 -> 2 // shard 1 send cross shard to shard 3 at  height 2
	LastCrossShardState map[byte]map[byte]uint64 `json:"LastCrossShardState"`
	ShardHandle         map[byte]bool            `json:"ShardHandle"` // lock sync.RWMutex
	lockMu              sync.RWMutex
	randomClient        btc.RandomClient
}

var beaconBestState *BeaconBestState

func NewBeaconBestState() *BeaconBestState {
	return &BeaconBestState{
		BestShardHash:         make(map[byte]common.Hash),
		BestShardHeight:       make(map[byte]uint64),
		ShardCommittee:        make(map[byte][]string),
		ShardPendingValidator: make(map[byte][]string),
		Params:                make(map[string]string),
		LastCrossShardState:   make(map[byte]map[byte]uint64),
	}
}
func NewBeaconBestStateWithConfig(netparam *Params) *BeaconBestState {
	if beaconBestState == nil {
		beaconBestState = GetBeaconBestState()
	}
	beaconBestState.BestBlockHash.SetBytes(make([]byte, 32))
	beaconBestState.BestBlockHash.SetBytes(make([]byte, 32))
	beaconBestState.BestShardHash = make(map[byte]common.Hash)
	beaconBestState.BestShardHeight = make(map[byte]uint64)
	beaconBestState.BeaconHeight = 0
	beaconBestState.BeaconCommittee = []string{}
	beaconBestState.BeaconPendingValidator = []string{}
	beaconBestState.CandidateShardWaitingForCurrentRandom = []string{}
	beaconBestState.CandidateBeaconWaitingForCurrentRandom = []string{}
	beaconBestState.CandidateShardWaitingForNextRandom = []string{}
	beaconBestState.CandidateBeaconWaitingForNextRandom = []string{}
	beaconBestState.ShardCommittee = make(map[byte][]string)
	beaconBestState.ShardPendingValidator = make(map[byte][]string)
	beaconBestState.Params = make(map[string]string)
	beaconBestState.CurrentRandomNumber = -1
	beaconBestState.MaxBeaconCommitteeSize = netparam.MaxBeaconCommitteeSize
	beaconBestState.MinBeaconCommitteeSize = netparam.MinBeaconCommitteeSize
	beaconBestState.MaxShardCommitteeSize = netparam.MaxShardCommitteeSize
	beaconBestState.MinShardCommitteeSize = netparam.MinShardCommitteeSize
	beaconBestState.ActiveShards = netparam.ActiveShards
	beaconBestState.LastCrossShardState = make(map[byte]map[byte]uint64)
	return beaconBestState
}
func SetBeaconBestState(beacon *BeaconBestState) {
	beaconBestState = beacon
}

func GetBeaconBestState() *BeaconBestState {
	if beaconBestState != nil {
		return beaconBestState
	}
	beaconBestState = NewBeaconBestState()
	return beaconBestState
}

func (beaconBestState *BeaconBestState) InitRandomClient(randomClient btc.RandomClient) {
	beaconBestState.randomClient = randomClient
}

func (beaconBestState *BeaconBestState) MarshalJSON() ([]byte, error) {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()

	type Alias BeaconBestState
	b, err := json.Marshal(&struct {
		*Alias
	}{
		(*Alias)(beaconBestState),
	})
	if err != nil {
		Logger.log.Error(err)
	}
	return b, err
}

func (beaconBestState *BeaconBestState) SetBestShardHeight(shardID byte, height uint64) {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	beaconBestState.BestShardHeight[shardID] = height
}

func (beaconBestState *BeaconBestState) GetBestShardHeight() map[byte]uint64 {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	res := make(map[byte]uint64)
	for index, element := range beaconBestState.BestShardHeight {
		res[index] = element
	}
	return res
}

func (beaconBestState *BeaconBestState) GetBestHeightOfShard(shardID byte) uint64 {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	return beaconBestState.BestShardHeight[shardID]
}

func (beaconBestState *BeaconBestState) GetAShardCommittee(shardID byte) []string {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	return beaconBestState.ShardCommittee[shardID]
}

func (beaconBestState *BeaconBestState) GetShardCommittee() (res map[byte][]string) {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	res = make(map[byte][]string)
	for index, element := range beaconBestState.ShardCommittee {
		res[index] = element
	}
	return res
}

func (beaconBestState *BeaconBestState) GetAShardPendingValidator(shardID byte) []string {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	return beaconBestState.ShardPendingValidator[shardID]
}

func (beaconBestState *BeaconBestState) GetShardPendingValidator() (res map[byte][]string) {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	res = make(map[byte][]string)
	for index, element := range beaconBestState.ShardPendingValidator {
		res[index] = element
	}
	return res
}

func (beaconBestState *BeaconBestState) GetCurrentShard() byte {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	for shardID, isCurrent := range beaconBestState.ShardHandle {
		if isCurrent {
			return shardID
		}
	}
	return 0
}

func (beaconBestState *BeaconBestState) SetMaxShardCommitteeSize(maxShardCommitteeSize int) bool {
	beaconBestState.lockMu.Lock()
	defer beaconBestState.lockMu.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if maxShardCommitteeSize < MinCommitteeSize {
		return false
	}
	// max committee size can't be lower than current min committee size
	if maxShardCommitteeSize >= beaconBestState.MinShardCommitteeSize {
		beaconBestState.MaxShardCommitteeSize = maxShardCommitteeSize
		return true
	}
	return false
}

func (beaconBestState *BeaconBestState) SetMinShardCommitteeSize(minShardCommitteeSize int) bool {
	beaconBestState.lockMu.Lock()
	defer beaconBestState.lockMu.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if minShardCommitteeSize < MinCommitteeSize {
		return false
	}
	// min committee size can't be greater than current min committee size
	if minShardCommitteeSize <= beaconBestState.MaxShardCommitteeSize {
		beaconBestState.MinShardCommitteeSize = minShardCommitteeSize
		return true
	}
	return false
}

func (beaconBestState *BeaconBestState) SetMaxBeaconCommitteeSize(maxBeaconCommitteeSize int) bool {
	beaconBestState.lockMu.Lock()
	defer beaconBestState.lockMu.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if maxBeaconCommitteeSize < MinCommitteeSize {
		return false
	}
	// max committee size can't be lower than current min committee size
	if maxBeaconCommitteeSize >= beaconBestState.MinBeaconCommitteeSize {
		beaconBestState.MaxBeaconCommitteeSize = maxBeaconCommitteeSize
		return true
	}
	return false
}

func (beaconBestState *BeaconBestState) SetMinBeaconCommitteeSize(minBeaconCommitteeSize int) bool {
	beaconBestState.lockMu.Lock()
	defer beaconBestState.lockMu.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if minBeaconCommitteeSize < MinCommitteeSize {
		return false
	}
	// min committee size can't be greater than current min committee size
	if minBeaconCommitteeSize <= beaconBestState.MaxBeaconCommitteeSize {
		beaconBestState.MinBeaconCommitteeSize = minBeaconCommitteeSize
		return true
	}
	return false
}
func (beaconBestState *BeaconBestState) CheckCommitteeSize() error {
	if beaconBestState.MaxBeaconCommitteeSize < MinCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect max beacon size %+v equal or greater than min size %+v", beaconBestState.MaxBeaconCommitteeSize, MinCommitteeSize))
	}
	if beaconBestState.MinBeaconCommitteeSize < MinCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect min beacon size %+v equal or greater than min size %+v", beaconBestState.MinBeaconCommitteeSize, MinCommitteeSize))
	}
	if beaconBestState.MaxShardCommitteeSize < MinCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect max shard size %+v equal or greater than min size %+v", beaconBestState.MaxShardCommitteeSize, MinCommitteeSize))
	}
	if beaconBestState.MinShardCommitteeSize < MinCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect min shard size %+v equal or greater than min size %+v", beaconBestState.MinShardCommitteeSize, MinCommitteeSize))
	}
	if beaconBestState.MaxBeaconCommitteeSize < beaconBestState.MinBeaconCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect Max beacon size is higher than min beacon size but max is %+v and min is %+v", beaconBestState.MaxBeaconCommitteeSize, beaconBestState.MinBeaconCommitteeSize))
	}
	if beaconBestState.MaxShardCommitteeSize < beaconBestState.MinShardCommitteeSize {
		return NewBlockChainError(CommitteeOrValidatorError, fmt.Errorf("Expect Max beacon size is higher than min beacon size but max is %+v and min is %+v", beaconBestState.MaxBeaconCommitteeSize, beaconBestState.MinBeaconCommitteeSize))
	}
	return nil
}

func (beaconBestState *BeaconBestState) GetBytes() []byte {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	var keys []int
	var keyStrs []string
	res := []byte{}
	res = append(res, beaconBestState.BestBlockHash.GetBytes()...)
	res = append(res, beaconBestState.PrevBestBlockHash.GetBytes()...)
	res = append(res, beaconBestState.BestBlock.Hash().GetBytes()...)
	res = append(res, beaconBestState.BestBlock.Header.PreviousBlockHash.GetBytes()...)
	for k := range beaconBestState.BestShardHash {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		hash := beaconBestState.BestShardHash[byte(shardID)]
		res = append(res, hash.GetBytes()...)
	}
	keys = []int{}
	for k := range beaconBestState.BestShardHeight {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		height := beaconBestState.BestShardHeight[byte(shardID)]
		res = append(res, byte(height))
	}
	EpochBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(EpochBytes, beaconBestState.Epoch)
	res = append(res, EpochBytes...)
	heightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(heightBytes, beaconBestState.BeaconHeight)
	res = append(res, heightBytes...)
	res = append(res, []byte(strconv.Itoa(beaconBestState.BeaconProposerIdx))...)
	for _, value := range beaconBestState.BeaconCommittee {
		res = append(res, []byte(value)...)
	}
	for _, value := range beaconBestState.BeaconPendingValidator {
		res = append(res, []byte(value)...)
	}
	for _, value := range beaconBestState.CandidateBeaconWaitingForCurrentRandom {
		res = append(res, []byte(value)...)
	}
	for _, value := range beaconBestState.CandidateBeaconWaitingForNextRandom {
		res = append(res, []byte(value)...)
	}
	for _, value := range beaconBestState.CandidateShardWaitingForCurrentRandom {
		res = append(res, []byte(value)...)
	}
	for _, value := range beaconBestState.CandidateShardWaitingForNextRandom {
		res = append(res, []byte(value)...)
	}
	keys = []int{}
	for k := range beaconBestState.ShardCommittee {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range beaconBestState.ShardCommittee[byte(shardID)] {
			res = append(res, []byte(value)...)
		}
	}
	keys = []int{}
	for k := range beaconBestState.ShardPendingValidator {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		for _, value := range beaconBestState.ShardPendingValidator[byte(shardID)] {
			res = append(res, []byte(value)...)
		}
	}

	randomNumBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(randomNumBytes, uint64(beaconBestState.CurrentRandomNumber))
	res = append(res, randomNumBytes...)

	randomTimeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(randomTimeBytes, uint64(beaconBestState.CurrentRandomTimeStamp))
	res = append(res, randomTimeBytes...)

	if beaconBestState.IsGetRandomNumber {
		res = append(res, []byte("true")...)
	} else {
		res = append(res, []byte("false")...)
	}
	for k := range beaconBestState.Params {
		keyStrs = append(keyStrs, k)
	}
	sort.Strings(keyStrs)
	for _, key := range keyStrs {
		res = append(res, []byte(beaconBestState.Params[key])...)
	}

	keys = []int{}
	for k := range beaconBestState.ShardHandle {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		shardHandleItem := beaconBestState.ShardHandle[byte(shardID)]
		if shardHandleItem {
			res = append(res, []byte("true")...)
		} else {
			res = append(res, []byte("false")...)
		}
	}
	res = append(res, []byte(strconv.Itoa(beaconBestState.MaxBeaconCommitteeSize))...)
	res = append(res, []byte(strconv.Itoa(beaconBestState.MinBeaconCommitteeSize))...)
	res = append(res, []byte(strconv.Itoa(beaconBestState.MaxShardCommitteeSize))...)
	res = append(res, []byte(strconv.Itoa(beaconBestState.MinShardCommitteeSize))...)
	res = append(res, []byte(strconv.Itoa(beaconBestState.ActiveShards))...)

	keys = []int{}
	for k := range beaconBestState.LastCrossShardState {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, fromShard := range keys {
		fromShardMap := beaconBestState.LastCrossShardState[byte(fromShard)]
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
func (beaconBestState *BeaconBestState) Hash() common.Hash {
	return common.HashH(beaconBestState.GetBytes())
}

// Get role of a public key base on best state beacond
// return node-role, <shardID>
func (beaconBestState *BeaconBestState) GetPubkeyRole(pubkey string, round int) (string, byte) {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	for shardID, pubkeyArr := range beaconBestState.ShardPendingValidator {
		found := common.IndexOfStr(pubkey, pubkeyArr)
		if found > -1 {
			return common.SHARD_ROLE, shardID
		}
	}

	for shardID, pubkeyArr := range beaconBestState.ShardCommittee {
		found := common.IndexOfStr(pubkey, pubkeyArr)
		if found > -1 {
			return common.SHARD_ROLE, shardID
		}
	}

	found := common.IndexOfStr(pubkey, beaconBestState.BeaconCommittee)
	if found > -1 {
		tmpID := (beaconBestState.BeaconProposerIdx + round) % len(beaconBestState.BeaconCommittee)
		if found == tmpID {
			return common.PROPOSER_ROLE, 0
		}
		return common.VALIDATOR_ROLE, 0
	}

	found = common.IndexOfStr(pubkey, beaconBestState.BeaconPendingValidator)
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
	beaconBestState := BeaconBestState{}
	if err := json.Unmarshal(prevBST, &beaconBestState); err != nil {
		return err
	}

	blkHash := block.Header.Hash()
	producerPk := base58.Base58Check{}.Encode(block.Header.ProducerAddress.Pk, common.ZeroByte)
	err = incognitokey.ValidateDataB58(producerPk, block.ProducerSig, blkHash.GetBytes())
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
	if block.Header.Version != BEACON_BLOCK_VERSION {
		return NewBlockChainError(WrongVersionError, errors.New("Version should be :"+strconv.Itoa(BEACON_BLOCK_VERSION)))
	}
	prevBlockHash := block.Header.PreviousBlockHash
	// Verify parent hash exist or not
	parentBlockBytes, err := blockchain.config.DataBase.FetchBeaconBlock(prevBlockHash)
	if err != nil {
		return NewBlockChainError(DatabaseError, err)
	}
	parentBlock := NewBeaconBlock()
	json.Unmarshal(parentBlockBytes, &parentBlock)
	// Verify block height with parent block
	if parentBlock.Header.Height+1 != block.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(block.Header.Height+1))))
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
	beaconBestState := BeaconBestState{}
	if err := json.Unmarshal(prevBST, &beaconBestState); err != nil {
		return err
	}

	blockchain.config.BeaconPool.SetBeaconState(beaconBestState.BeaconHeight)
	blockchain.config.ShardToBeaconPool.SetShardState(blockchain.BestState.Beacon.GetBestShardHeight())

	if err := blockchain.config.DataBase.DeleteCommitteeByHeight(currentBestStateBlk.Header.Height); err != nil {
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

	for _, inst := range currentBestStateBlk.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		if inst[0] == SetAction || inst[0] == StakeAction || inst[0] == RandomAction || inst[0] == SwapAction || inst[0] == AssignAction {
			continue
		}
		var err error
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}
		switch metaType {
		case metadata.AcceptedBlockRewardInfoMeta:
			acceptedBlkRewardInfo, err := metadata.NewAcceptedBlockRewardInfoFromStr(inst[2])
			if err != nil {
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
			Logger.log.Infof("TxsFee in Epoch: %+v of shardID: %+v:\n", currentBestStateBlk.Header.Epoch, acceptedBlkRewardInfo.ShardID)
			for key, value := range acceptedBlkRewardInfo.TxsFee {
				Logger.log.Infof("===> TokenID:%+v: Amount: %+v\n", key, value)
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
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}

	if err := blockchain.config.DataBase.StorePrevBestState(tempMarshal, true, 0); err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}

	for _, inst := range block.Body.Instructions {
		if len(inst) < 2 {
			continue // Not error, just not bridge instruction
		}
		if inst[0] == SetAction || inst[0] == StakeAction || inst[0] == RandomAction || inst[0] == SwapAction || inst[0] == AssignAction {
			continue
		}

		var err error
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			continue
		}

		switch metaType {
		case metadata.AcceptedBlockRewardInfoMeta:
			acceptedBlkRewardInfo, err := metadata.NewAcceptedBlockRewardInfoFromStr(inst[2])
			if err != nil {
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
			for key, _ := range acceptedBlkRewardInfo.TxsFee {
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
	return nil
}

func (beaconBestState *BeaconBestState) GetShardCandidate() []string {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	return append(beaconBestState.CandidateShardWaitingForCurrentRandom, beaconBestState.CandidateShardWaitingForNextRandom...)
}
func (beaconBestState *BeaconBestState) GetBeaconCandidate() []string {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	return append(beaconBestState.CandidateBeaconWaitingForCurrentRandom, beaconBestState.CandidateBeaconWaitingForNextRandom...)
}
func (beaconBestState *BeaconBestState) GetBeaconCommittee() []string {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	return beaconBestState.BeaconCommittee
}
func (beaconBestState *BeaconBestState) GetBeaconPendingValidator() []string {
	beaconBestState.lockMu.RLock()
	defer beaconBestState.lockMu.RUnlock()
	return beaconBestState.BeaconPendingValidator
}
func (shardBestState *ShardBestState) cloneBeaconBestState(target *ShardBestState) error {
	tempMarshal, err := json.Marshal(target)
	if err != nil {
		return NewBlockChainError(MashallJsonShardBestStateError, fmt.Errorf("Shard Best State %+v get %+v", target.ShardHeight, err))
	}
	err = json.Unmarshal(tempMarshal, shardBestState)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBestStateError, fmt.Errorf("Clone Shard Best State %+v get %+v", target.ShardHeight, err))
	}
	if reflect.DeepEqual(*shardBestState, ShardBestState{}) {
		return NewBlockChainError(CloneShardBestStateError, fmt.Errorf("Shard Best State %+v clone failed", target.ShardHeight))
	}
	return nil
}
