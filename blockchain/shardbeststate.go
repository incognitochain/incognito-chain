package blockchain

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/flatfile"
	"reflect"
	"sort"
	"time"

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

	// Number of blocks produced by producers in epoch
	NumOfBlocksByProducers map[string]uint64 `json:"NumOfBlocksByProducers"`
	BlockInterval          time.Duration
	BlockMaxCreateTime     time.Duration
	MetricBlockHeight      uint64
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
}

func (shardBestState *ShardBestState) GetCopiedConsensusStateDB() *statedb.StateDB {
	return shardBestState.consensusStateDB.Copy()
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

func (shardBestState *ShardBestState) InitStateRootHash(db incdb.Database, bc *BlockChain, isRepair bool) error {
	if isRepair {
		return nil
	}
	var err error
	var dbAccessWarper = statedb.NewDatabaseAccessWrapperWithConfig(db, bc.cacheConfig.trieJournalPath[int(shardBestState.ShardID)], bc.cacheConfig.trieJournalCacheSize)
	shardBestState.transactionStateDB, err = statedb.NewWithPrefixTrie(shardBestState.TransactionStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}
	shardBestState.consensusStateDB, err = statedb.NewWithPrefixTrie(shardBestState.ConsensusStateDBRootHash, dbAccessWarper)
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

	if shardBestState.BeaconHeight == config.Param().ConsensusParam.BlockProducingV3Height {
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
	switch shardBlock.Header.Version {
	case types.BFT_VERSION:
		return shardBestState.GetShardCommittee(), shardBestState.GetShardCommittee(), nil
	case types.MULTI_VIEW_VERSION, types.SHARD_SFV2_VERSION, types.SHARD_SFV3_VERSION, types.LEMMA2_VERSION:
		committees, err := bc.getShardCommitteeForBlockProducing(shardBlock.CommitteeFromBlock(), shardBlock.Header.ShardID)
		if err != nil {
			return []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, err
		}
		signingCommittees := incognitokey.DeepCopy(committees)
		return committees, signingCommittees, nil
	case types.BLOCK_PRODUCINGV3_VERSION:
		committees, err := bc.getShardCommitteeForBlockProducing(shardBlock.CommitteeFromBlock(), shardBlock.Header.ShardID)
		if err != nil {
			return []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, err
		}
		timeSlot := common.CalculateTimeSlot(shardBlock.Header.ProposeTime)
		_, proposerIndex := GetProposer(
			timeSlot,
			committees,
			shardBestState.GetProposerLength(),
		)
		signingCommitteeV3 := FilterSigningCommitteeV3(
			committees,
			proposerIndex)
		return committees, signingCommitteeV3, nil
	default:
		panic("shardBestState.CommitteeState is not a valid version")
	}
	return []incognitokey.CommitteePublicKey{}, []incognitokey.CommitteePublicKey{}, nil
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

//GetSubsetIDFromProposerTime for block producing v3 only
func GetSubsetIDFromProposerTime(proposerTime int64, validators int) int {
	proposerIndex := GetProposerByTimeSlot(common.CalculateTimeSlot(proposerTime), validators)
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

func (shardBestState *ShardBestState) CommitTrieToDisk(
	batch incdb.Batch,
	bc *BlockChain,
	sRH ShardRootHash,
	flatFileIndexes [][]int,
	isFinalizedBlock bool,
	newFinalBlock types.BlockInterface,
) error {

	consensusTrieDB := shardBestState.consensusStateDB.Database().TrieDB()
	transactionTrieDB := shardBestState.transactionStateDB.Database().TrieDB()
	featureTrieDB := shardBestState.featureStateDB.Database().TrieDB()
	rewardTrieDB := shardBestState.rewardStateDB.Database().TrieDB()
	slashTrieDB := shardBestState.slashStateDB.Database().TrieDB()

	if bc.cacheConfig.trieJournalPath != nil {
		if path := bc.cacheConfig.trieJournalPath[int(shardBestState.ShardID)]; path != "" {
			// the below disk layer is just one => save one time is enough
			transactionTrieDB.SaveCache(path)
		}
	}
	// use for archive mode or force to do so
	if shardBestState.ShardHeight == 1 || ShardSyncMode == ARCHIVE_SYNC_MODE {
		if err := shardBestState.commitTrieToDisk(
			bc, batch, types.BlockInterface(shardBestState.BestBlock), sRH, bc.config.FlatFileManager[int(shardBestState.ShardID)], flatFileIndexes); err != nil {
			return err
		}
	} else {
		// Full but not archive node, do proper garbage collection
		consensusTrieDB.Reference(sRH.ConsensusStateDBRootHash, common.Hash{})
		transactionTrieDB.Reference(sRH.TransactionStateDBRootHash, common.Hash{}) // metadata reference to keep trie alive
		featureTrieDB.Reference(sRH.FeatureStateDBRootHash, common.Hash{})
		rewardTrieDB.Reference(sRH.RewardStateDBRootHash, common.Hash{})
		slashTrieDB.Reference(sRH.SlashStateDBRootHash, common.Hash{})
		bc.cacheConfig.triegc.Push(sRH, -int64(shardBestState.ShardHeight))
		if current := shardBestState.ShardHeight; current >= bc.cacheConfig.blockTriesInMemory {
			var (
				nodes, imgs = transactionTrieDB.Size()
			)
			Logger.log.Debugf("SHARD %+v | Transaction Trie Cap. Nodes %+v, trieNodeLimit %+v, img %+v, trieImgLimit %+v",
				shardBestState.ShardID, nodes, bc.cacheConfig.trieNodeLimit, imgs, common.StorageSize(4*1024*1024))
			// all statedb object use the same low-level triedb

			if nodes > bc.cacheConfig.trieNodeLimit || imgs > bc.cacheConfig.trieImgsLimit {
				transactionTrieDB.Cap(bc.cacheConfig.trieNodeLimit - incdb.IdealBatchSize)
			}

			if isFinalizedBlock &&
				current > bc.cacheConfig.fullSyncPivot[shardBestState.ShardID] &&
				current-bc.cacheConfig.fullSyncPivot[shardBestState.ShardID] >= bc.cacheConfig.blockTriesInMemory+1 {
				if err := shardBestState.fullSyncCommitTrieToDisk(
					bc, batch,
					newFinalBlock,
					bc.config.FlatFileManager[int(shardBestState.ShardID)], flatFileIndexes); err != nil {
					return err
				}
			}

			chosen := current / bc.cacheConfig.blockTriesInMemory * bc.cacheConfig.blockTriesInMemory
			// Garbage collect anything below our required write retention
			// Dereference could takes time and block the insertion process
			for !bc.cacheConfig.triegc.Empty() {
				oldSRH, number := bc.cacheConfig.triegc.Pop()
				if uint64(-number) > chosen {
					bc.cacheConfig.triegc.Push(oldSRH, number)
					break
				}
				Logger.log.Debugf("SHARD %+v | Try Dereference, current %+v, chosen %+v, deref block %+v", shardBestState.ShardID, current, chosen, number)
				consensusTrieDB.Dereference(oldSRH.(ShardRootHash).ConsensusStateDBRootHash)
				transactionTrieDB.Dereference(oldSRH.(ShardRootHash).TransactionStateDBRootHash)
				featureTrieDB.Dereference(oldSRH.(ShardRootHash).FeatureStateDBRootHash)
				rewardTrieDB.Dereference(oldSRH.(ShardRootHash).RewardStateDBRootHash)
				slashTrieDB.Dereference(oldSRH.(ShardRootHash).SlashStateDBRootHash)
			}
			postNodes, postImgs := transactionTrieDB.Size()
			if nodes-postNodes > 0 || imgs-postImgs > 0 {
				Logger.log.Debugf("SHARD %+v | Success Dereference, current %+v, reduce nodes %+v, reduce imgs %+v", shardBestState.ShardID, current, nodes-postNodes, imgs-postImgs)
			}
		}
	}

	if err := rawdbv2.StoreShardRootsHash(batch, shardBestState.ShardID, shardBestState.BestBlockHash, sRH); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	return nil
}

func (shardBestState *ShardBestState) fullSyncCommitTrieToDisk(
	bc *BlockChain,
	batch incdb.Batch,
	blockToCommit types.BlockInterface,
	flatFileManager *flatfile.FlatFileManager,
	flatFileIndexes [][]int,
) error {

	sRH, err := GetShardRootsHashByBlockHash(bc.ShardChain[shardBestState.ShardID].GetChainDatabase(),
		shardBestState.ShardID, *blockToCommit.Hash())
	if err != nil {
		return err
	}

	return shardBestState.commitTrieToDisk(bc, batch, blockToCommit, *sRH, flatFileManager, flatFileIndexes)
}

func (shardBestState *ShardBestState) commitTrieToDisk(
	bc *BlockChain,
	batch incdb.Batch,
	blockToCommit types.BlockInterface,
	sRH ShardRootHash,
	flatFileManager *flatfile.FlatFileManager,
	flatFileIndexes [][]int,
) error {

	var err error

	err = shardBestState.consensusStateDB.Database().TrieDB().Commit(sRH.ConsensusStateDBRootHash, false, nil) // Save data to disk database
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = shardBestState.slashStateDB.Database().TrieDB().Commit(sRH.SlashStateDBRootHash, false, nil)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = shardBestState.transactionStateDB.Database().TrieDB().Commit(sRH.TransactionStateDBRootHash, false, nil)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = shardBestState.featureStateDB.Database().TrieDB().Commit(sRH.FeatureStateDBRootHash, false, nil)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = shardBestState.rewardStateDB.Database().TrieDB().Commit(sRH.RewardStateDBRootHash, false, nil)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}

	bc.cacheConfig.fullSyncPivot[shardBestState.ShardID] = blockToCommit.GetHeight()

	if err := rawdbv2.StoreLatestPivotBlock(batch, shardBestState.ShardID, *blockToCommit.Hash()); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}

	truncateLastIndex(flatFileManager, flatFileIndexes)

	shardBestState.consensusStateDB.ClearObjects()
	shardBestState.transactionStateDB.ClearObjects()
	shardBestState.featureStateDB.ClearObjects()
	shardBestState.rewardStateDB.ClearObjects()
	shardBestState.slashStateDB.ClearObjects()
	Logger.log.Infof("SHARD %+v | Finish commit Trie to disk, height %+v, hash %+v", shardBestState.ShardID, blockToCommit.GetHeight(), *blockToCommit.Hash())

	return nil
}

func truncateLastIndex(flatFileManager *flatfile.FlatFileManager, indexes [][]int) {
	// truncate old files
	lastIndex := 0
	for _, v := range indexes {
		if len(v) != 0 {
			if lastIndex < v[len(v)-1] {
				lastIndex = v[len(v)-1]
			}
		}
	}
	if lastIndex != 0 {
		//TODO: hung test without truncate
		//err := flatFileManager.Truncate(lastIndex)
		//if err != nil {
		//	Logger.log.Errorf("StoreShardBlockError, truncate flatfile with last index %+v, error %+v", lastIndex, err)
		//}
	}
}

func (curView *ShardBestState) GetProposerLength() int {
	return curView.NumberOfFixedShardBlockValidator
}
