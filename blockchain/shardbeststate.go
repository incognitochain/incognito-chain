package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/common"
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

type ShardBestState struct {
	BestBlockHash          common.Hash                       `json:"BestBlockHash"` // hash of block.
	BestBlock              *ShardBlock                       `json:"BestBlock"`     // block data
	BestBeaconHash         common.Hash                       `json:"BestBeaconHash"`
	BeaconHeight           uint64                            `json:"BeaconHeight"`
	ShardID                byte                              `json:"ShardID"`
	Epoch                  uint64                            `json:"Epoch"`
	ShardHeight            uint64                            `json:"ShardHeight"`
	MaxShardCommitteeSize  int                               `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize  int                               `json:"MinShardCommitteeSize"`
	ShardProposerIdx       int                               `json:"ShardProposerIdx"`
	ShardCommittee         []incognitokey.CommitteePublicKey `json:"ShardCommittee"`
	ShardPendingValidator  []incognitokey.CommitteePublicKey `json:"ShardPendingValidator"`
	BestCrossShard         map[byte]uint64                   `json:"BestCrossShard"` // Best cross shard block by heigh
	StakingTx              map[string]string                 `json:"StakingTx"`
	NumTxns                uint64                            `json:"NumTxns"`                // The number of txns in the block.
	TotalTxns              uint64                            `json:"TotalTxns"`              // The total number of txns in the chain.
	TotalTxnsExcludeSalary uint64                            `json:"TotalTxnsExcludeSalary"` // for testing and benchmark
	ActiveShards           int                               `json:"ActiveShards"`
	ConsensusAlgorithm     string                            `json:"ConsensusAlgorithm"`
	// Number of blocks produced by producers in epoch
	NumOfBlocksByProducers map[string]uint64 `json:"NumOfBlocksByProducers"`
	BlockInterval          time.Duration
	BlockMaxCreateTime     time.Duration
	MetricBlockHeight      uint64
	//================================ StateDB Method
	// block height => root hash
	consensusStateDB   *statedb.StateDB
	transactionStateDB *statedb.StateDB
	featureStateDB     *statedb.StateDB
	rewardStateDB      *statedb.StateDB
	slashStateDB       *statedb.StateDB
	lock               sync.RWMutex
}

func (shardBestState *ShardBestState) GetCopiedTransactionStateDB() *statedb.StateDB {
	return shardBestState.transactionStateDB.Copy()
}

func (shardBestState *ShardBestState) GetCopiedFeatureStateDB() *statedb.StateDB {
	return shardBestState.featureStateDB.Copy()
}

func (shardBestState *ShardBestState) GetCopiedRewardStateDB() *statedb.StateDB {
	return shardBestState.rewardStateDB.Copy()
}

var bestStateShardMap = make(map[byte]*ShardBestState)

func NewShardBestState() *ShardBestState {
	return &ShardBestState{}
}
func NewShardBestStateWithShardID(shardID byte) *ShardBestState {
	return &ShardBestState{ShardID: shardID}
}
func NewBestStateShardWithConfig(shardID byte, netparam *Params) *ShardBestState {
	bestStateShard := GetBestStateShard(shardID)
	err := bestStateShard.BestBlockHash.SetBytes(make([]byte, 32))
	if err != nil {
		panic(err)
	}
	err = bestStateShard.BestBeaconHash.SetBytes(make([]byte, 32))
	if err != nil {
		panic(err)
	}
	bestStateShard.BestBlock = nil
	bestStateShard.ShardCommittee = []incognitokey.CommitteePublicKey{}
	bestStateShard.MaxShardCommitteeSize = netparam.MaxShardCommitteeSize
	bestStateShard.MinShardCommitteeSize = netparam.MinShardCommitteeSize
	bestStateShard.ShardPendingValidator = []incognitokey.CommitteePublicKey{}
	bestStateShard.ActiveShards = netparam.ActiveShards
	bestStateShard.BestCrossShard = make(map[byte]uint64)
	bestStateShard.StakingTx = make(map[string]string)
	bestStateShard.ShardHeight = 1
	bestStateShard.BeaconHeight = 1
	bestStateShard.BlockInterval = netparam.MinShardBlockInterval
	bestStateShard.BlockMaxCreateTime = netparam.MaxShardBlockCreation
	return bestStateShard
}

func GetBestStateShard(shardID byte) *ShardBestState {
	if bestStateShard, ok := bestStateShardMap[shardID]; !ok {
		bestStateShardMap[shardID] = &ShardBestState{}
		bestStateShardMap[shardID].ShardID = shardID
		return bestStateShardMap[shardID]
	} else {
		return bestStateShard
	}
}

func SetBestStateShard(shardID byte, beststateShard *ShardBestState) {
	beststateShard.lock = GetBestStateShard(shardID).lock
	*GetBestStateShard(shardID) = *beststateShard
}
func (shardBestState *ShardBestState) InitStateRootHash(db incdb.Database, bc *BlockChain) error {
	var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
	if rootHash, err := bc.GetShardConsensusRootHash(db, shardBestState.ShardID, shardBestState.ShardHeight); err == nil {
		shardBestState.consensusStateDB, err = statedb.NewWithPrefixTrie(rootHash, dbAccessWarper)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	if rootHash, err := bc.GetShardTransactionRootHash(db, shardBestState.ShardID, shardBestState.ShardHeight); err == nil {
		shardBestState.transactionStateDB, err = statedb.NewWithPrefixTrie(rootHash, dbAccessWarper)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	if rootHash, err := bc.GetShardFeatureRootHash(db, shardBestState.ShardID, shardBestState.ShardHeight); err == nil {
		shardBestState.featureStateDB, err = statedb.NewWithPrefixTrie(rootHash, dbAccessWarper)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	if rootHash, err := bc.GetShardCommitteeRewardRootHash(db, shardBestState.ShardID, shardBestState.ShardHeight); err == nil {
		shardBestState.rewardStateDB, err = statedb.NewWithPrefixTrie(rootHash, dbAccessWarper)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	if rootHash, err := bc.GetShardSlashRootHash(db, shardBestState.ShardID, shardBestState.ShardHeight); err == nil {
		shardBestState.slashStateDB, err = statedb.NewWithPrefixTrie(rootHash, dbAccessWarper)
		if err != nil {
			return err
		}
	} else {
		return err
	}
	return nil
}

// Get role of a public key base on best state shard
func (shardBestState *ShardBestState) GetBytes() []byte {
	res := []byte{}
	res = append(res, shardBestState.BestBlockHash.GetBytes()...)
	res = append(res, shardBestState.BestBlock.Hash().GetBytes()...)
	res = append(res, shardBestState.BestBeaconHash.GetBytes()...)
	beaconHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(beaconHeightBytes, shardBestState.BeaconHeight)
	res = append(res, beaconHeightBytes...)
	res = append(res, shardBestState.ShardID)
	epochBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(epochBytes, shardBestState.Epoch)
	res = append(res, epochBytes...)
	shardHeightBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(shardHeightBytes, shardBestState.ShardHeight)
	res = append(res, shardHeightBytes...)
	shardCommitteeSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(shardCommitteeSizeBytes, uint32(shardBestState.MaxShardCommitteeSize))
	res = append(res, shardCommitteeSizeBytes...)
	minShardCommitteeSizeBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(minShardCommitteeSizeBytes, uint32(shardBestState.MinShardCommitteeSize))
	res = append(res, minShardCommitteeSizeBytes...)
	proposerIdxBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(proposerIdxBytes, uint32(shardBestState.ShardProposerIdx))
	res = append(res, proposerIdxBytes...)
	for _, value := range shardBestState.ShardCommittee {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	for _, value := range shardBestState.ShardPendingValidator {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	keys := []int{}
	for k := range shardBestState.BestCrossShard {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, shardID := range keys {
		value := shardBestState.BestCrossShard[byte(shardID)]
		valueBytes := make([]byte, 8)
		binary.LittleEndian.PutUint64(valueBytes, value)
		res = append(res, valueBytes...)
	}
	keystr := []string{}
	for _, k := range shardBestState.StakingTx {
		keystr = append(keystr, k)
	}
	sort.Strings(keystr)
	for _, key := range keystr {
		value := shardBestState.StakingTx[key]
		res = append(res, []byte(key)...)
		res = append(res, []byte(value)...)
	}
	numTxnsBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(numTxnsBytes, shardBestState.NumTxns)
	res = append(res, numTxnsBytes...)
	totalTxnsBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(totalTxnsBytes, shardBestState.TotalTxns)
	res = append(res, totalTxnsBytes...)
	activeShardsBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(activeShardsBytes, uint32(shardBestState.ActiveShards))
	res = append(res, activeShardsBytes...)
	return res
}

func (shardBestState *ShardBestState) Hash() common.Hash {
	shardBestState.lock.RLock()
	defer shardBestState.lock.RUnlock()
	return common.HashH(shardBestState.GetBytes())
}

func (shardBestState *ShardBestState) SetMaxShardCommitteeSize(maxShardCommitteeSize int) bool {
	shardBestState.lock.Lock()
	defer shardBestState.lock.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if maxShardCommitteeSize < MinCommitteeSize {
		return false
	}
	// max committee size can't be lower than current min committee size
	if maxShardCommitteeSize >= shardBestState.MinShardCommitteeSize {
		shardBestState.MaxShardCommitteeSize = maxShardCommitteeSize
		return true
	}
	return false
}

func (shardBestState *ShardBestState) SetMinShardCommitteeSize(minShardCommitteeSize int) bool {
	shardBestState.lock.Lock()
	defer shardBestState.lock.Unlock()
	// check input params, below MinCommitteeSize failed to acheive consensus
	if minShardCommitteeSize < MinCommitteeSize {
		return false
	}
	// min committee size can't be greater than current min committee size
	if minShardCommitteeSize <= shardBestState.MaxShardCommitteeSize {
		shardBestState.MinShardCommitteeSize = minShardCommitteeSize
		return true
	}
	return false
}

func (shardBestState *ShardBestState) GetPubkeyRole(pubkey string, round int) string {
	keyList, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(shardBestState.ShardCommittee, shardBestState.ConsensusAlgorithm)
	found := common.IndexOfStr(pubkey, keyList)
	if found > -1 {
		tmpID := (shardBestState.ShardProposerIdx + round) % len(keyList)
		if found == tmpID {
			return common.ProposerRole
		} else {
			return common.ValidatorRole
		}
	}

	keyList, _ = incognitokey.ExtractPublickeysFromCommitteeKeyList(shardBestState.ShardPendingValidator, shardBestState.ConsensusAlgorithm)
	found = common.IndexOfStr(pubkey, keyList)
	if found > -1 {
		return common.PendingRole
	}
	return common.EmptyString
}

//MarshalJSON - remember to use lock
func (shardBestState *ShardBestState) MarshalJSON() ([]byte, error) {
	type Alias ShardBestState
	b, err := json.Marshal(&struct {
		*Alias
	}{
		(*Alias)(shardBestState),
	})
	if err != nil {
		Logger.log.Error(err)
	}
	return b, err
}

func (shardBestState ShardBestState) GetShardHeight() uint64 {
	return shardBestState.ShardHeight
}

func (shardBestState ShardBestState) GetBeaconHeight() uint64 {
	return shardBestState.BeaconHeight
}

//cloneShardBestStateFrom - remember to use lock
func (shardBestState *ShardBestState) cloneShardBestStateFrom(target *ShardBestState) error {
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
	shardBestState.consensusStateDB = target.consensusStateDB.Copy()
	shardBestState.transactionStateDB = target.transactionStateDB.Copy()
	shardBestState.featureStateDB = target.featureStateDB.Copy()
	shardBestState.rewardStateDB = target.rewardStateDB.Copy()
	shardBestState.slashStateDB = target.slashStateDB.Copy()
	return nil
}

func (shardBestState *ShardBestState) GetStakingTx() map[string]string {
	shardBestState.lock.RLock()
	defer shardBestState.lock.RUnlock()
	m := make(map[string]string)
	for k, v := range shardBestState.StakingTx {
		m[k] = v
	}
	return m
}

func (shardBestState *ShardBestState) GetShardCommittee() []incognitokey.CommitteePublicKey {
	shardBestState.lock.RLock()
	defer shardBestState.lock.RUnlock()
	return shardBestState.ShardCommittee
}

func (shardBestState *ShardBestState) GetShardPendingValidator() []incognitokey.CommitteePublicKey {
	shardBestState.lock.RLock()
	defer shardBestState.lock.RUnlock()
	return shardBestState.ShardPendingValidator
}

func (blockchain *BlockChain) GetShardConsensusRootHash(db incdb.Database, shardID byte, height uint64) (common.Hash, error) {
	return rawdbv2.GetShardConsensusRootHash(db, shardID, height)
}

func (blockchain *BlockChain) GetShardCommitteeRewardRootHash(db incdb.Database, shardID byte, height uint64) (common.Hash, error) {
	return rawdbv2.GetShardCommitteeRewardRootHash(db, shardID, height)
}

func (blockchain *BlockChain) GetShardTransactionRootHash(db incdb.Database, shardID byte, height uint64) (common.Hash, error) {
	return rawdbv2.GetShardTransactionRootHash(db, shardID, height)
}

func (blockchain *BlockChain) GetShardFeatureRootHash(db incdb.Database, shardID byte, height uint64) (common.Hash, error) {
	return rawdbv2.GetShardFeatureRootHash(db, shardID, height)
}

func (blockchain *BlockChain) GetShardSlashRootHash(db incdb.Database, shardID byte, height uint64) (common.Hash, error) {
	return rawdbv2.GetShardSlashRootHash(db, shardID, height)
}
