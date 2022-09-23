package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/incognitochain/incognito-chain/multiview"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
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

type ShardRootHash struct {
	ConsensusStateDBRootHash   common.Hash
	TransactionStateDBRootHash common.Hash
	FeatureStateDBRootHash     common.Hash
	RewardStateDBRootHash      common.Hash
	SlashStateDBRootHash       common.Hash
}

type ShardBestState struct {
	blockChain                       *BlockChain
	BestBlockHash                    common.Hash       `json:"BestBlockHash"` // hash of block.
	BestBlock                        *types.ShardBlock `json:"BestBlock"`     // block data
	BestBeaconHash                   common.Hash       `json:"BestBeaconHash"`
	BeaconHeight                     uint64            `json:"BeaconHeight"`
	ShardID                          byte              `json:"ShardID"`
	Epoch                            uint64            `json:"Epoch"`
	ShardHeight                      uint64            `json:"ShardHeight"`
	MaxShardCommitteeSize            int               `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize            int               `json:"MinShardCommitteeSize"`
	NumberOfFixedShardBlockValidator int               `json:"NumberOfFixedValidator"`
	ShardProposerIdx                 int               `json:"ShardProposerIdx"`
	BestCrossShard                   map[byte]uint64   `json:"BestCrossShard"`         // Best cross shard block by heigh
	NumTxns                          uint64            `json:"NumTxns"`                // The number of txns in the block.
	TotalTxns                        uint64            `json:"TotalTxns"`              // The total number of txns in the chain.
	TotalTxnsExcludeSalary           uint64            `json:"TotalTxnsExcludeSalary"` // for testing and benchmark
	ActiveShards                     int               `json:"ActiveShards"`
	ConsensusAlgorithm               string            `json:"ConsensusAlgorithm"`

	TriggeredFeature map[string]uint64 `json:"TriggeredFeature"`
	// Number of blocks produced by producers in epoch
	NumOfBlocksByProducers  map[string]uint64 `json:"NumOfBlocksByProducers"`
	BlockInterval           time.Duration
	BlockMaxCreateTime      time.Duration
	MetricBlockHeight       uint64
	MaxTxsPerBlockRemainder int64
	TSManager               TSManager

	//================================ StateDB Method
	// block height => root hash
	consensusStateDB           *statedb.StateDB
	ConsensusStateDBRootHash   common.Hash
	transactionStateDB         *statedb.StateDB
	TransactionStateDBRootHash common.Hash
	featureStateDB             *statedb.StateDB
	FeatureStateDBRootHash     common.Hash
	rewardStateDB              *statedb.StateDB
	RewardStateDBRootHash      common.Hash
	slashStateDB               *statedb.StateDB
	SlashStateDBRootHash       common.Hash
	shardCommitteeState        committeestate.ShardCommitteeState
	ShardRebuildRootHash       *ShardRebuildRootHash
}

type ShardRebuildRootHash struct {
	ConsensusStateDBRootHash   *statedb.RebuildInfo
	TransactionStateDBRootHash *statedb.RebuildInfo
	FeatureStateDBRootHash     *statedb.RebuildInfo
	RewardStateDBRootHash      *statedb.RebuildInfo
	SlashStateDBRootHash       *statedb.RebuildInfo
}

func (shardBestState *ShardBestState) CalculateTimeSlot(t int64) int64 {
	return shardBestState.TSManager.calculateTimeslot(t)
}

func (shardBestState *ShardBestState) GetCurrentTimeSlot() int64 {
	return shardBestState.TSManager.getCurrentTS()
}

func (shardBestState *ShardBestState) GetCopiedConsensusStateDB() *statedb.StateDB {
	return shardBestState.consensusStateDB.Copy()
}

//for test only
func (shardBestState *ShardBestState) SetTransactonDB(h common.Hash, txDB *statedb.StateDB) {
	shardBestState.transactionStateDB = txDB
	shardBestState.TransactionStateDBRootHash = h
	return
}

func (shardBestState *ShardBestState) GetCopiedTransactionStateDB() *statedb.StateDB {
	return shardBestState.transactionStateDB.Copy()
}

func (shardBestState *ShardBestState) GetCopiedFeatureStateDB() *statedb.StateDB {
	return shardBestState.featureStateDB.Copy()
}

func (shardBestState *ShardBestState) GetShardRewardStateDB() *statedb.StateDB {
	return shardBestState.rewardStateDB.Copy()
}

func (shardBestState *ShardBestState) GetHash() *common.Hash {
	return shardBestState.BestBlock.Hash()
}

func (shardBestState *ShardBestState) GetPreviousHash() *common.Hash {
	return &shardBestState.BestBlock.Header.PreviousBlockHash
}

func (shardBestState *ShardBestState) GetPreviousBlockCommittee(db incdb.Database) ([]incognitokey.CommitteePublicKey, error) {
	return getOneShardCommitteeFromBeaconDB(db, shardBestState.ShardID, *shardBestState.GetPreviousHash())
}

func (shardBestState *ShardBestState) GetHeight() uint64 {
	return shardBestState.BestBlock.GetHeight()
}

func (shardBestState *ShardBestState) GetEpoch() uint64 {
	return shardBestState.Epoch
}

func (shardBestState *ShardBestState) GetBlockTime() int64 {
	return shardBestState.BestBlock.Header.Timestamp
}

func (shardBestState *ShardBestState) CommitteeFromBlock() common.Hash {
	return shardBestState.BestBlock.Header.CommitteeFromBlock
}

func NewShardBestState() *ShardBestState {
	return &ShardBestState{}
}
func NewShardBestStateWithShardID(shardID byte) *ShardBestState {
	return &ShardBestState{ShardID: shardID}
}
func NewBestStateShardWithConfig(shardID byte, shardCommitteeState committeestate.ShardCommitteeState) *ShardBestState {
	bestStateShard := NewShardBestStateWithShardID(shardID)
	err := bestStateShard.BestBlockHash.SetBytes(make([]byte, 32))
	if err != nil {
		panic(err)
	}
	err = bestStateShard.BestBeaconHash.SetBytes(make([]byte, 32))
	if err != nil {
		panic(err)
	}
	bestStateShard.BestBlock = nil
	bestStateShard.MaxShardCommitteeSize = config.Param().CommitteeSize.MaxShardCommitteeSize
	bestStateShard.MinShardCommitteeSize = config.Param().CommitteeSize.MinShardCommitteeSize
	bestStateShard.NumberOfFixedShardBlockValidator = config.Param().CommitteeSize.NumberOfFixedShardBlockValidator
	bestStateShard.ActiveShards = config.Param().ActiveShards
	bestStateShard.BestCrossShard = make(map[byte]uint64)
	bestStateShard.ShardHeight = 1
	bestStateShard.BeaconHeight = 1
	bestStateShard.BlockInterval = config.Param().BlockTime.MinShardBlockInterval
	bestStateShard.BlockMaxCreateTime = config.Param().BlockTime.MaxShardBlockCreation
	bestStateShard.shardCommitteeState = shardCommitteeState
	return bestStateShard
}

func (blockchain *BlockChain) GetBestStateShard(shardID byte) *ShardBestState {
	if blockchain.ShardChain[int(shardID)].multiView.GetBestView() == nil {
		return nil
	}
	return blockchain.ShardChain[int(shardID)].multiView.GetBestView().(*ShardBestState)
}

func (shardBestState *ShardBestState) InitStateRootHash(db incdb.Database, lastView *ShardBestState) error {
	var err error
	var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
	shardBestState.consensusStateDB, err = statedb.NewWithPrefixTrie(shardBestState.ConsensusStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}

	shardBestState.transactionStateDB, err = statedb.InitBatchCommit("tx", dbAccessWarper, shardBestState.ShardRebuildRootHash.TransactionStateDBRootHash, lastView.transactionStateDB)
	if err != nil {
		return err
	}

	shardBestState.featureStateDB, err = statedb.NewWithPrefixTrie(shardBestState.FeatureStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}

	shardBestState.rewardStateDB, err = statedb.NewWithPrefixTrie(shardBestState.RewardStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}

	shardBestState.slashStateDB, err = statedb.NewWithPrefixTrie(shardBestState.SlashStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}
	if shardBestState.ShardRebuildRootHash != nil {
		shardBestState.ShardRebuildRootHash = new(ShardRebuildRootHash)
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
	for _, value := range shardBestState.shardCommitteeState.GetShardCommittee() {
		valueBytes, err := value.Bytes()
		if err != nil {
			return nil
		}
		res = append(res, valueBytes...)
	}
	for _, value := range shardBestState.shardCommitteeState.GetShardSubstitute() {
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
	return common.HashH(shardBestState.GetBytes())
}

func (shardBestState *ShardBestState) SetMaxShardCommitteeSize(maxShardCommitteeSize int) bool {
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

func (shardBestState ShardBestState) GetBeaconHash() common.Hash {
	return shardBestState.BestBeaconHash
}

func (shardBestState ShardBestState) GetShardID() byte {
	return shardBestState.ShardID
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
	shardBestState.shardCommitteeState = target.shardCommitteeState.Clone()
	shardBestState.BestBlock = target.BestBlock
	shardBestState.blockChain = target.blockChain
	return nil
}

func (shardBestState *ShardBestState) GetStakingTx() map[string]string {
	m := make(map[string]string)
	return m
}

func (shardBestState *ShardBestState) GetCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	return append(result, shardBestState.shardCommitteeState.GetShardCommittee()...)
}

// GetProposerByTimeSlot return proposer by timeslot from current committee of shard view
func (shardBestState *ShardBestState) GetProposerByTimeSlot(
	ts int64,
	version int,
) (incognitokey.CommitteePublicKey, int) {
	id := GetProposerByTimeSlot(ts, shardBestState.GetProposerLength())
	return shardBestState.GetShardCommittee()[id], id
}

func (shardBestState *ShardBestState) GetBlock() types.BlockInterface {
	return shardBestState.BestBlock
}

func (shardBestState *ShardBestState) ReplaceBlock(replaceBlock types.BlockInterface) {
	shardBestState.BestBlock = replaceBlock.(*types.ShardBlock)
}

func (shardBestState *ShardBestState) GetShardCommittee() []incognitokey.CommitteePublicKey {
	return shardBestState.shardCommitteeState.GetShardCommittee()
}

func (shardBestState *ShardBestState) GetShardPendingValidator() []incognitokey.CommitteePublicKey {
	return shardBestState.shardCommitteeState.GetShardSubstitute()
}

func (shardBestState *ShardBestState) ListShardPrivacyTokenAndPRV() []common.Hash {
	tokenIDs := []common.Hash{}
	tokenStates := statedb.ListPrivacyToken(shardBestState.GetCopiedTransactionStateDB())
	for k := range tokenStates {
		tokenIDs = append(tokenIDs, k)
	}
	return tokenIDs
}

func (shardBestState *ShardBestState) GetBlockVersion() int {
	return shardBestState.BestBlock.GetVersion()
}

func (blockchain *BlockChain) GetShardRootsHash(shardBestState *ShardBestState, shardID byte, height uint64) (*ShardRootHash, error) {
	h, err := blockchain.GetShardBlockHashByHeight(blockchain.ShardChain[shardID].GetFinalView(), blockchain.ShardChain[shardID].GetBestView(), height)
	if err != nil {
		return nil, err
	}
	data, err := rawdbv2.GetShardRootsHash(blockchain.GetShardChainDatabase(shardID), shardID, *h)
	if err != nil {
		return nil, err
	}
	sRH := &ShardRootHash{}
	err = json.Unmarshal(data, sRH)
	return sRH, err
}

func InitShardCommitteeState(
	version int,
	consensusStateDB *statedb.StateDB,
	shardHeight uint64,
	shardID byte,
	block *types.ShardBlock,
	bc *BlockChain) committeestate.ShardCommitteeState {
	var err error
	committees := statedb.GetOneShardCommittee(consensusStateDB, shardID)
	if version == committeestate.SELF_SWAP_SHARD_VERSION {
		shardPendingValidators := statedb.GetOneShardSubstituteValidator(consensusStateDB, shardID)
		shardCommitteeState := committeestate.NewShardCommitteeStateV1WithValue(committees, shardPendingValidators)
		return shardCommitteeState
	}
	if shardHeight != 1 {
		committees, err = bc.getShardCommitteeFromBeaconHash(block.Header.CommitteeFromBlock, shardID)
		if err != nil {
			Logger.log.Error(NewBlockChainError(InitShardStateError, err))
			panic(err)
		}
	}
	switch version {
	case committeestate.STAKING_FLOW_V2:
		return committeestate.NewShardCommitteeStateV2WithValue(
			committees,
		)
	case committeestate.STAKING_FLOW_V3:
		return committeestate.NewShardCommitteeStateV3WithValue(
			committees,
		)
	default:
		panic("shardBestState.CommitteeState not a valid version to init")
	}
}

//ShardCommitteeEngine : getter of shardCommitteeState ...
func (shardBestState *ShardBestState) ShardCommitteeEngine() committeestate.ShardCommitteeState {
	return shardBestState.shardCommitteeState
}

//CommitteeEngineVersion ...
func (shardBestState *ShardBestState) CommitteeStateVersion() int {
	return shardBestState.shardCommitteeState.Version()
}

func (shardBestState *ShardBestState) NewShardCommitteeStateEnvironmentWithValue(
	shardBlock *types.ShardBlock,
	bc *BlockChain,
	beaconInstructions [][]string,
	tempCommittees []string,
	genesisBeaconHash common.Hash) *committeestate.ShardCommitteeStateEnvironment {
	return &committeestate.ShardCommitteeStateEnvironment{
		BeaconHeight:                 shardBestState.BeaconHeight,
		Epoch:                        bc.GetEpochByHeight(shardBestState.BeaconHeight),
		EpochBreakPointSwapNewKey:    config.Param().ConsensusParam.EpochBreakPointSwapNewKey,
		BeaconInstructions:           beaconInstructions,
		MaxShardCommitteeSize:        shardBestState.MaxShardCommitteeSize,
		NumberOfFixedBlockValidators: shardBestState.NumberOfFixedShardBlockValidator,
		MinShardCommitteeSize:        shardBestState.MinShardCommitteeSize,
		Offset:                       config.Param().SwapCommitteeParam.Offset,
		ShardBlockHash:               shardBestState.BestBlockHash,
		ShardHeight:                  shardBestState.ShardHeight,
		ShardID:                      shardBestState.ShardID,
		StakingTx:                    make(map[string]string),
		SwapOffset:                   config.Param().SwapCommitteeParam.SwapOffset,
		Txs:                          shardBlock.Body.Transactions,
		ShardInstructions:            shardBlock.Body.Instructions,
		CommitteesFromBlock:          shardBlock.Header.CommitteeFromBlock,
		CommitteesFromBeaconView:     tempCommittees,
		GenesisBeaconHash:            genesisBeaconHash,
	}
}

// tryUpgradeCommitteeState only allow
// Upgrade to v2 if and only if current version is 1 and beacon height == staking flow v2 height
// Upgrade to v3 if and only if current version is 2 and beacon height == staking flow v3 height
// @NOTICE: DO NOT UPDATE IN BLOCK WITH SWAP INSTRUCTION
func (shardBestState *ShardBestState) tryUpgradeCommitteeState(bc *BlockChain) error {

	if shardBestState.BeaconHeight >= config.Param().ConsensusParam.BlockProducingV3Height {
		err := shardBestState.checkAndUpgradeStakingFlowV3Config()
		if err != nil {
			return err
		}
	}

	if shardBestState.BeaconHeight != config.Param().ConsensusParam.StakingFlowV2Height &&
		shardBestState.BeaconHeight != config.Param().ConsensusParam.StakingFlowV3Height {
		return nil
	}
	if shardBestState.BeaconHeight == config.Param().ConsensusParam.StakingFlowV3Height {
		if shardBestState.CommitteeStateVersion() != committeestate.STAKING_FLOW_V2 {
			return nil
		}
		if shardBestState.CommitteeStateVersion() == committeestate.STAKING_FLOW_V3 {
			return nil
		}
	}
	if shardBestState.BeaconHeight == config.Param().ConsensusParam.StakingFlowV2Height {
		if shardBestState.CommitteeStateVersion() != committeestate.SELF_SWAP_SHARD_VERSION {
			return nil
		}
		if shardBestState.CommitteeStateVersion() == committeestate.STAKING_FLOW_V2 {
			return nil
		}
	}

	var committeeFromBlock common.Hash
	var committees []incognitokey.CommitteePublicKey
	var err error

	if shardBestState.BeaconHeight == config.Param().ConsensusParam.StakingFlowV2Height &&
		committeeFromBlock.IsZeroValue() {
		committees = shardBestState.GetCommittee()
	} else {
		committeeFromBlock = shardBestState.BestBlock.CommitteeFromBlock()
		committees, err = bc.getShardCommitteeFromBeaconHash(committeeFromBlock, shardBestState.ShardID)
		if err != nil {
			return err
		}
	}

	if shardBestState.BeaconHeight == config.Param().ConsensusParam.StakingFlowV2Height {
		shardBestState.shardCommitteeState = committeestate.NewShardCommitteeStateV2WithValue(
			committees,
		)
	}

	if shardBestState.BeaconHeight == config.Param().ConsensusParam.StakingFlowV3Height {
		shardBestState.shardCommitteeState = committeestate.NewShardCommitteeStateV3WithValue(
			committees,
		)
	}

	Logger.log.Infof("SHARDID %+v | Shard Height %+v, UPGRADE Shard Committee State from V1 to V2", shardBestState.ShardID, shardBestState.ShardHeight)
	return nil
}

func (ShardBestState *ShardBestState) checkAndUpgradeStakingFlowV3Config() error {

	if err := ShardBestState.checkBlockProducingV3Config(); err != nil {
		return NewBlockChainError(UpgradeShardCommitteeStateError, err)
	}

	if err := ShardBestState.upgradeBlockProducingV3Config(); err != nil {
		return NewBlockChainError(UpgradeShardCommitteeStateError, err)
	}

	return nil
}

func (shardBestState *ShardBestState) checkBlockProducingV3Config() error {

	shardCommitteeSize := len(shardBestState.GetShardCommittee())
	if shardCommitteeSize < SFV3_MinShardCommitteeSize {
		return fmt.Errorf("shard %+v | current committee length %+v can not upgrade to staking flow v3, "+
			"minimum required committee size is 8", shardBestState.ShardID, shardCommitteeSize)
	}

	return nil
}

func (shardBestState *ShardBestState) upgradeBlockProducingV3Config() error {

	if shardBestState.MinShardCommitteeSize < SFV3_MinShardCommitteeSize {
		shardBestState.MinShardCommitteeSize = SFV3_MinShardCommitteeSize
		Logger.log.Infof("SHARD %+v | Set shardBestState.MinShardCommitteeSize from %+v to %+v ",
			shardBestState.ShardID, shardBestState.MinShardCommitteeSize, SFV3_MinShardCommitteeSize)
	}

	if shardBestState.NumberOfFixedShardBlockValidator < SFV3_MinShardCommitteeSize {
		shardBestState.NumberOfFixedShardBlockValidator = SFV3_MinShardCommitteeSize
		Logger.log.Infof("SHARD %+v | Set shardBestState.NumberOfFixedShardBlockValidator from %+v to %+v ",
			shardBestState.ShardID, shardBestState.NumberOfFixedShardBlockValidator, SFV3_MinShardCommitteeSize)
	}

	if shardBestState.MaxShardCommitteeSize < SFV3_MinShardCommitteeSize {
		shardBestState.MaxShardCommitteeSize = SFV3_MinShardCommitteeSize
		Logger.log.Infof("SHARD %+v | Set shardBestState.MaxShardCommitteeSize from %+v to %+v ",
			shardBestState.ShardID, shardBestState.MaxShardCommitteeSize, SFV3_MinShardCommitteeSize)
	}

	return nil
}

func (shardBestState *ShardBestState) verifyCommitteeFromBlock(
	blockchain *BlockChain,
	shardBlock *types.ShardBlock,
	committees []incognitokey.CommitteePublicKey,
) error {
	committeeFinalViewBlock, _, err := blockchain.GetBeaconBlockByHash(shardBlock.Header.CommitteeFromBlock)
	if err != nil {
		return err
	}
	if !shardBestState.CommitteeFromBlock().IsZeroValue() {
		newCommitteesPubKeys, _ := incognitokey.CommitteeKeyListToString(committees)
		oldCommitteesPubKeys, _ := incognitokey.CommitteeKeyListToString(shardBestState.GetCommittee())
		temp := committeestate.DifferentElementStrings(oldCommitteesPubKeys, newCommitteesPubKeys)
		if len(temp) != 0 {
			oldCommitteeFromBlock, _, err := blockchain.GetBeaconBlockByHash(shardBestState.CommitteeFromBlock())
			if err != nil {
				return err
			}

			if oldCommitteeFromBlock.Header.Height >= committeeFinalViewBlock.Header.Height {
				return NewBlockChainError(WrongBlockHeightError,
					fmt.Errorf("Height of New Shard Block's Committee From Block %+v is smaller than current Committee From Block View %+v",
						committeeFinalViewBlock.Header.Hash(), oldCommitteeFromBlock.Header.Hash()))
			}
		}
	}
	return nil
}

func (x *ShardBestState) CompareCommitteeFromBlock(_y multiview.View) int {
	//if equal
	y := _y.(*ShardBestState)
	if x.CommitteeFromBlock().String() == y.CommitteeFromBlock().String() {
		return 0
	}
	//if not equal
	xCommitteeBlock, _, err := x.blockChain.GetBeaconBlockByHash(x.BestBlock.Header.CommitteeFromBlock)
	if err != nil {
		Logger.log.Error("Cannot find committee from block!")
		return 0
	}
	yCommitteeBlock, _, err := x.blockChain.GetBeaconBlockByHash(y.BestBlock.Header.CommitteeFromBlock)
	if err != nil {
		Logger.log.Error("Cannot find committee from block!")
		return 0
	}
	if xCommitteeBlock.GetHeight() > yCommitteeBlock.GetHeight() {
		return 1
	}
	return -1
}

func (curView *ShardBestState) getUntriggerFeature() []string {
	unTriggerFeatures := []string{}
	for f, _ := range config.Param().AutoEnableFeature {
		if curView.TriggeredFeature == nil || curView.TriggeredFeature[f] == 0 {
			unTriggerFeatures = append(unTriggerFeatures, f)
		}
	}
	return unTriggerFeatures
}

// Output:
// 1. Full committee
// 2. signing committee
// 3. error
func (shardBestState *ShardBestState) getSigningCommittees(
	shardBlock *types.ShardBlock, bc *BlockChain,
) ([]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	if shardBlock.Header.CommitteeFromBlock.IsZeroValue() {
		return shardBestState.GetShardCommittee(), shardBestState.GetShardCommittee(), nil
	}

	if shardBlock.Header.Version == types.BFT_VERSION {
		return shardBestState.GetShardCommittee(), shardBestState.GetShardCommittee(), nil
	}
	if shardBlock.Header.Version >= types.MULTI_VIEW_VERSION && shardBlock.Header.Version <= types.LEMMA2_VERSION || shardBlock.Header.Version >= types.INSTANT_FINALITY_VERSION_V2 {
		committees, err := bc.getShardCommitteeForBlockProducing(shardBlock.CommitteeFromBlock(), shardBlock.Header.ShardID)
		if err != nil {
			return []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, err
		}
		signingCommittees := incognitokey.DeepCopy(committees)
		return committees, signingCommittees, nil
	}
	if shardBlock.Header.Version >= types.BLOCK_PRODUCINGV3_VERSION && shardBlock.Header.Version <= types.INSTANT_FINALITY_VERSION {
		committees, err := bc.getShardCommitteeForBlockProducing(shardBlock.CommitteeFromBlock(), shardBlock.Header.ShardID)
		if err != nil {
			return []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, err
		}
		timeSlot := shardBestState.CalculateTimeSlot(shardBlock.Header.ProposeTime)
		// timeSlot := common.CalculateTimeSlot(shardBlock.Header.ProposeTime)
		_, proposerIndex := GetProposer(
			timeSlot,
			committees,
			shardBestState.GetProposerLength(),
		)
		signingCommitteeV3 := FilterSigningCommitteeV3(
			committees,
			proposerIndex)
		return committees, signingCommitteeV3, nil
	}
	panic("shardBestState.CommitteeState is not a valid version")
}

func GetProposer(
	ts int64, committees []incognitokey.CommitteePublicKey,
	lenProposers int) (incognitokey.CommitteePublicKey, int) {
	id := GetProposerByTimeSlot(ts, lenProposers)
	return committees[id], id
}

func GetProposerByTimeSlot(ts int64, committeeLen int) int {
	id := int(ts) % committeeLen
	return id
}

// //GetSubsetIDFromProposerTime for block producing v3 only
// func GetSubsetIDFromProposerTime(proposerTime int64, validators int) int {
// 	proposerIndex := GetProposerByTimeSlot(common.CalculateTimeSlot(proposerTime), validators)
// 	subsetID := GetSubsetID(proposerIndex)
// 	return subsetID
// }

//TODO
func GetSubsetIDFromProposerTimeV2(proposerTimeSlot int64, validators int) int {
	proposerIndex := GetProposerByTimeSlot(proposerTimeSlot, validators)
	subsetID := GetSubsetID(proposerIndex)
	return subsetID
}

func GetSubsetID(proposerIndex int) int {
	return proposerIndex % MaxSubsetCommittees
}

// GetSubsetIDByKey compare based on consensus mining key
func GetSubsetIDByKey(fullCommittees []incognitokey.CommitteePublicKey, miningKey string, consensusName string) (int, int) {
	for i, v := range fullCommittees {
		if v.GetMiningKeyBase58(consensusName) == miningKey {
			return i, i % MaxSubsetCommittees
		}
	}

	return -1, -1
}

func FilterSigningCommitteeV3StringValue(fullCommittees []string, proposerIndex int) []string {
	signingCommittees := []string{}
	subsetID := GetSubsetID(proposerIndex)
	for i, v := range fullCommittees {
		if (i % MaxSubsetCommittees) == subsetID {
			signingCommittees = append(signingCommittees, v)
		}
	}
	return signingCommittees
}

func FilterSigningCommitteeV3(fullCommittees []incognitokey.CommitteePublicKey, proposerIndex int) []incognitokey.CommitteePublicKey {
	signingCommittees := []incognitokey.CommitteePublicKey{}
	subsetID := GetSubsetID(proposerIndex)
	for i, v := range fullCommittees {
		if (i % MaxSubsetCommittees) == subsetID {
			signingCommittees = append(signingCommittees, v)
		}
	}
	return signingCommittees
}

func getConfirmedCommitteeHeightFromBeacon(bc *BlockChain, shardBlock *types.ShardBlock) (uint64, error) {

	if shardBlock.Header.CommitteeFromBlock.IsZeroValue() {
		return shardBlock.Header.BeaconHeight, nil
	}

	_, beaconHeight, err := bc.GetBeaconBlockByHash(shardBlock.Header.CommitteeFromBlock)
	if err != nil {
		return 0, err
	}

	return beaconHeight, nil
}

func (curView *ShardBestState) GetProposerLength() int {
	return curView.NumberOfFixedShardBlockValidator
}
