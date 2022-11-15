package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/multiview"

	"github.com/incognitochain/incognito-chain/blockchain/bridgeagg"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/pdex"
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/consensus"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"
	portalprocessv4 "github.com/incognitochain/incognito-chain/portal/portalv4/portalprocess"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/syncker/finishsync"
)

const (
	MAX_COMMITTEE_SIZE_48_FEATURE = "maxcommitteesize48"
	INSTANT_FINALITY_FEATURE      = "instantfinality"
	BLOCKTIME_DEFAULT             = "blocktimedef"
	BLOCKTIME_20                  = "blocktime20"
	BLOCKTIME_10                  = "blocktime10"
	EPOCHV2                       = "epochparamv2"
	INSTANT_FINALITY_FEATURE_V2   = "instantfinalityv2"
	REDUCE_FIX_NODE               = "reduce_fix_node"
	REDUCE_FIX_NODE_V2            = "reduce_fix_node_v2"
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

type BeaconRootHash struct {
	ConsensusStateDBRootHash common.Hash
	FeatureStateDBRootHash   common.Hash
	RewardStateDBRootHash    common.Hash
	SlashStateDBRootHash     common.Hash
}

type BeaconBestState struct {
	BestBlockHash                    common.Hash          `json:"BestBlockHash"`         // The hash of the block.
	PreviousBestBlockHash            common.Hash          `json:"PreviousBestBlockHash"` // The hash of the block.
	BestBlock                        types.BeaconBlock    `json:"BestBlock"`             // The block.
	BestShardHash                    map[byte]common.Hash `json:"BestShardHash"`
	BestShardHeight                  map[byte]uint64      `json:"BestShardHeight"`
	Epoch                            uint64               `json:"Epoch"`
	BeaconHeight                     uint64               `json:"BeaconHeight"`
	BeaconProposerIndex              int                  `json:"BeaconProposerIndex"`
	CurrentRandomNumber              int64                `json:"CurrentRandomNumber"`
	CurrentRandomTimeStamp           int64                `json:"CurrentRandomTimeStamp"` // random timestamp for this epoch
	IsGetRandomNumber                bool                 `json:"IsGetRandomNumber"`
	MaxBeaconCommitteeSize           int                  `json:"MaxBeaconCommitteeSize"`
	MinBeaconCommitteeSize           int                  `json:"MinBeaconCommitteeSize"`
	MaxShardCommitteeSize            int                  `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize            int                  `json:"MinShardCommitteeSize"`
	ActiveShards                     int                  `json:"ActiveShards"`
	ConsensusAlgorithm               string               `json:"ConsensusAlgorithm"`
	ShardConsensusAlgorithm          map[byte]string      `json:"ShardConsensusAlgorithm"`
	NumberOfShardBlock               map[byte]uint        `json:"NumberOfShardBlock"`
	TriggeredFeature                 map[string]uint64    `json:"TriggeredFeature"`
	NumberOfFixedShardBlockValidator int                  `json:"NumberOfFixedShardBlockValidator"`
	RewardMinted                     uint64               `json:"RewardMinted"`
	TSManager                        TSManager
	ShardTSManager                   map[byte]*TSManager

	// key: public key of committee, value: payment address reward receiver
	beaconCommitteeState    committeestate.BeaconCommitteeState
	missingSignatureCounter signaturecounter.IMissingSignatureCounter
	// cross shard state for all the shard. from shardID -> to crossShard shardID -> last height
	// e.g 1 -> 2 -> 3 // shard 1 send cross shard to shard 2 at  height 3
	// e.g 1 -> 3 -> 2 // shard 1 send cross shard to shard 3 at  height 2
	LastCrossShardState map[byte]map[byte]uint64 `json:"LastCrossShardState"`
	ShardHandle         map[byte]bool            `json:"ShardHandle"` // lock sync.RWMutex
	// Number of blocks produced by producers in epoch
	BlockInterval      time.Duration
	BlockMaxCreateTime time.Duration
	//================================ StateDB Method
	// block height => root hash
	consensusStateDB         *statedb.StateDB
	ConsensusStateDBRootHash common.Hash
	rewardStateDB            *statedb.StateDB
	RewardStateDBRootHash    common.Hash
	featureStateDB           *statedb.StateDB
	FeatureStateDBRootHash   common.Hash
	slashStateDB             *statedb.StateDB
	SlashStateDBRootHash     common.Hash

	pdeStates              map[uint]pdex.State
	portalStateV3          *portalprocessv3.CurrentPortalState
	portalStateV4          *portalprocessv4.CurrentPortalStateV4
	bridgeAggManager       *bridgeagg.Manager
	LastBlockProcessBridge uint64
}

func (beaconBestState *BeaconBestState) TimeLeftOver(t int64) time.Duration {
	return beaconBestState.TSManager.timeLeftOver(t)
}

func (beaconBestState *BeaconBestState) PastHalfTimeslot(t int64) bool {
	return false
}

func (beaconBestState *BeaconBestState) CalculateTimeSlot(t int64) int64 {
	return beaconBestState.TSManager.calculateTimeslot(t)
}
func (beaconBestState *BeaconBestState) GetCurrentTimeSlot() int64 {
	return beaconBestState.TSManager.getCurrentTS()
}

func (beaconBestState *BeaconBestState) GetBeaconSlashStateDB() *statedb.StateDB {
	return beaconBestState.slashStateDB.Copy()
}

func (beaconBestState *BeaconBestState) GetBeaconFeatureStateDB() *statedb.StateDB {
	return beaconBestState.featureStateDB.Copy()
}

func (beaconBestState *BeaconBestState) GetBeaconRewardStateDB() *statedb.StateDB {
	return beaconBestState.rewardStateDB.Copy()
}

func (beaconBestState *BeaconBestState) GetBeaconConsensusStateDB() *statedb.StateDB {
	return beaconBestState.consensusStateDB.Copy()
}

// var beaconBestState *BeaconBestState

func NewBeaconBestState() *BeaconBestState {
	beaconBestState := new(BeaconBestState)
	beaconBestState.pdeStates = make(map[uint]pdex.State)
	beaconBestState.bridgeAggManager = bridgeagg.NewManager()
	beaconBestState.ShardTSManager = make(map[byte]*TSManager)
	return beaconBestState
}
func NewBeaconBestStateWithConfig(beaconCommitteeState committeestate.BeaconCommitteeState) *BeaconBestState {
	beaconBestState := NewBeaconBestState()
	beaconBestState.BestBlockHash.SetBytes(make([]byte, 32))
	beaconBestState.BestShardHeight = make(map[byte]uint64)
	beaconBestState.ShardTSManager = make(map[byte]*TSManager)
	beaconBestState.BestShardHash = make(map[byte]common.Hash)
	beaconBestState.BeaconHeight = 0
	beaconBestState.CurrentRandomNumber = -1
	beaconBestState.MaxBeaconCommitteeSize = config.Param().CommitteeSize.MaxBeaconCommitteeSize
	beaconBestState.MinBeaconCommitteeSize = config.Param().CommitteeSize.MinBeaconCommitteeSize
	beaconBestState.MaxShardCommitteeSize = config.Param().CommitteeSize.MaxShardCommitteeSize
	beaconBestState.MinShardCommitteeSize = config.Param().CommitteeSize.MinShardCommitteeSize
	beaconBestState.NumberOfFixedShardBlockValidator = config.Param().CommitteeSize.NumberOfFixedShardBlockValidator
	beaconBestState.ActiveShards = config.Param().ActiveShards
	beaconBestState.LastCrossShardState = make(map[byte]map[byte]uint64)
	beaconBestState.BlockInterval = config.Param().BlockTime.MinBeaconBlockInterval
	beaconBestState.BlockMaxCreateTime = config.Param().BlockTime.MaxBeaconBlockCreation
	beaconBestState.beaconCommitteeState = beaconCommitteeState
	return beaconBestState
}

func (curView *BeaconBestState) SetMissingSignatureCounter(missingSignatureCounter signaturecounter.IMissingSignatureCounter) {
	curView.missingSignatureCounter = missingSignatureCounter
}

func (bc *BlockChain) GetBeaconBestState() *BeaconBestState {
	return bc.BeaconChain.multiView.GetBestView().(*BeaconBestState)
}

func (beaconBestState *BeaconBestState) GetBeaconHeight() uint64 {
	return beaconBestState.BeaconHeight
}

func (beaconBestState *BeaconBestState) InitStateRootHash(bc *BlockChain) error {
	db := bc.GetBeaconChainDatabase()
	var err error
	var dbAccessWarper = statedb.NewDatabaseAccessWarper(db)
	beaconBestState.consensusStateDB, err = statedb.NewWithPrefixTrie(beaconBestState.ConsensusStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}
	beaconBestState.featureStateDB, err = statedb.NewWithPrefixTrie(beaconBestState.FeatureStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}
	beaconBestState.rewardStateDB, err = statedb.NewWithPrefixTrie(beaconBestState.RewardStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}
	beaconBestState.slashStateDB, err = statedb.NewWithPrefixTrie(beaconBestState.SlashStateDBRootHash, dbAccessWarper)
	if err != nil {
		return err
	}
	return nil
}

func (beaconBestState *BeaconBestState) MarshalJSON() ([]byte, error) {
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

func (beaconBestState *BeaconBestState) GetProducerIndexFromBlock(block *types.BeaconBlock) int {
	//TODO: revert his
	//return (beaconBestState.BeaconProposerIndex + block.Header.Round) % len(beaconBestState.BeaconCommittee)
	return 0
}

func (beaconBestState *BeaconBestState) SetBestShardHeight(shardID byte, height uint64) {

	beaconBestState.BestShardHeight[shardID] = height
}

func (beaconBestState *BeaconBestState) GetShardConsensusAlgorithm() map[byte]string {

	res := make(map[byte]string)
	for index, element := range beaconBestState.ShardConsensusAlgorithm {
		res[index] = element
	}
	return res
}

func (beaconBestState *BeaconBestState) GetBestShardHash() map[byte]common.Hash {
	res := make(map[byte]common.Hash)
	for index, element := range beaconBestState.BestShardHash {
		res[index] = element
	}
	return res
}

func (beaconBestState *BeaconBestState) GetBestShardHeight() map[byte]uint64 {

	res := make(map[byte]uint64)
	for index, element := range beaconBestState.BestShardHeight {
		res[index] = element
	}
	return res
}

func (beaconBestState *BeaconBestState) GetBestHeightOfShard(shardID byte) uint64 {

	return beaconBestState.BestShardHeight[shardID]
}

func (beaconBestState *BeaconBestState) GetAShardCommittee(shardID byte) []incognitokey.CommitteePublicKey {

	return beaconBestState.beaconCommitteeState.GetOneShardCommittee(shardID)
}

func (beaconBestState *BeaconBestState) GetShardCommittee() (res map[byte][]incognitokey.CommitteePublicKey) {
	res = make(map[byte][]incognitokey.CommitteePublicKey)
	for index, element := range beaconBestState.beaconCommitteeState.GetShardCommittee() {
		res[index] = element
	}
	return res
}

func (beaconBestState *BeaconBestState) GetShardCommitteeFlattenList() []string {
	committees := []string{}
	for _, committeeStructs := range beaconBestState.GetShardCommittee() {
		for _, committee := range committeeStructs {
			res, _ := committee.ToBase58()
			committees = append(committees, res)
		}
	}

	return committees
}

func (beaconBestState *BeaconBestState) getNewShardCommitteeFlattenList() []string {

	committees := []string{}
	for _, committeeStructs := range beaconBestState.beaconCommitteeState.GetShardCommittee() {
		for _, committee := range committeeStructs {
			res, _ := committee.ToBase58()
			committees = append(committees, res)
		}
	}

	return committees
}

func (beaconBestState *BeaconBestState) GetAShardPendingValidator(shardID byte) []incognitokey.CommitteePublicKey {

	return beaconBestState.beaconCommitteeState.GetOneShardSubstitute(shardID)
}

func (beaconBestState *BeaconBestState) GetShardPendingValidator() (res map[byte][]incognitokey.CommitteePublicKey) {

	res = make(map[byte][]incognitokey.CommitteePublicKey)
	for index, element := range beaconBestState.beaconCommitteeState.GetShardSubstitute() {
		res[index] = element
	}
	return res
}

func (beaconBestState *BeaconBestState) GetCurrentShard() byte {

	for shardID, isCurrent := range beaconBestState.ShardHandle {
		if isCurrent {
			return shardID
		}
	}
	return 0
}

func (beaconBestState *BeaconBestState) SetMaxShardCommitteeSize(maxShardCommitteeSize int) bool {
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

func (beaconBestState *BeaconBestState) Hash() common.Hash {
	return common.Hash{}
}

func (beaconBestState *BeaconBestState) GetShardCandidate() []incognitokey.CommitteePublicKey {
	current := beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForCurrentRandom()
	next := beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForNextRandom()
	return append(current, next...)
}

func (beaconBestState *BeaconBestState) GetBeaconCandidate() []incognitokey.CommitteePublicKey {
	current := beaconBestState.beaconCommitteeState.GetCandidateBeaconWaitingForCurrentRandom()
	next := beaconBestState.beaconCommitteeState.GetCandidateBeaconWaitingForNextRandom()
	return append(current, next...)
}
func (beaconBestState *BeaconBestState) GetBeaconCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	return append(result, beaconBestState.beaconCommitteeState.GetBeaconCommittee()...)
}

func (beaconBestState *BeaconBestState) GetCommittee() []incognitokey.CommitteePublicKey {
	committee := beaconBestState.GetBeaconCommittee()
	result := []incognitokey.CommitteePublicKey{}
	return append(result, committee...)
}

func (beaconBestState *BeaconBestState) GetProposerByTimeSlot(
	ts int64,
	version int,
) (incognitokey.CommitteePublicKey, int) {
	id := GetProposerByTimeSlot(ts, beaconBestState.MinBeaconCommitteeSize)
	return beaconBestState.GetBeaconCommittee()[id], id
}

func (beaconBestState *BeaconBestState) GetBlock() types.BlockInterface {
	return &beaconBestState.BestBlock
}

func (beaconBestState *BeaconBestState) ReplaceBlock(replaceBlock types.BlockInterface) {
	beaconBestState.BestBlock = *replaceBlock.(*types.BeaconBlock)
}

func (beaconBestState *BeaconBestState) GetBeaconPendingValidator() []incognitokey.CommitteePublicKey {
	return beaconBestState.beaconCommitteeState.GetBeaconSubstitute()
}

func (beaconBestState *BeaconBestState) GetRewardReceiver() map[string]privacy.PaymentAddress {
	return beaconBestState.beaconCommitteeState.GetRewardReceiver()
}

func (beaconBestState *BeaconBestState) GetAutoStaking() map[string]bool {
	return beaconBestState.beaconCommitteeState.GetAutoStaking()
}

func (beaconBestState *BeaconBestState) GetStakingTx() map[string]common.Hash {
	return beaconBestState.beaconCommitteeState.GetStakingTx()
}

func (beaconBestState *BeaconBestState) GetCandidateBeaconWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return beaconBestState.beaconCommitteeState.GetCandidateBeaconWaitingForCurrentRandom()
}

func (beaconBestState *BeaconBestState) GetCandidateBeaconWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return beaconBestState.beaconCommitteeState.GetCandidateBeaconWaitingForNextRandom()
}

func (beaconBestState *BeaconBestState) GetCandidateShardWaitingForCurrentRandom() []incognitokey.CommitteePublicKey {
	return beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForCurrentRandom()
}

func (beaconBestState *BeaconBestState) GetCandidateShardWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForNextRandom()
}

func (beaconBestState *BeaconBestState) GetSyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	return beaconBestState.beaconCommitteeState.GetSyncingValidators()
}
func (beaconBestState *BeaconBestState) GetSyncingValidatorsString() map[byte][]string {
	res := make(map[byte][]string)
	for shardID, validators := range beaconBestState.beaconCommitteeState.GetSyncingValidators() {
		res[shardID], _ = incognitokey.CommitteeKeyListToString(validators)
	}
	return res
}

// CommitteeStateVersion ...
func (beaconBestState *BeaconBestState) CommitteeStateVersion() int {
	return beaconBestState.beaconCommitteeState.Version()
}

func (beaconBestState *BeaconBestState) cloneBeaconBestStateFrom(target *BeaconBestState) error {
	tempMarshal, err := target.MarshalJSON()
	if err != nil {
		return NewBlockChainError(MashallJsonBeaconBestStateError, fmt.Errorf("Shard Best State %+v get %+v", beaconBestState.BeaconHeight, err))
	}
	err = json.Unmarshal(tempMarshal, beaconBestState)
	if err != nil {
		return NewBlockChainError(UnmashallJsonBeaconBestStateError, fmt.Errorf("Clone Shard Best State %+v get %+v", beaconBestState.BeaconHeight, err))
	}
	plainBeaconBestState := NewBeaconBestState()
	if reflect.DeepEqual(*beaconBestState, plainBeaconBestState) {
		return NewBlockChainError(CloneBeaconBestStateError, fmt.Errorf("Shard Best State %+v clone failed", beaconBestState.BeaconHeight))
	}
	beaconBestState.consensusStateDB = target.consensusStateDB.Copy()
	beaconBestState.featureStateDB = target.featureStateDB.Copy()
	beaconBestState.rewardStateDB = target.rewardStateDB.Copy()
	beaconBestState.slashStateDB = target.slashStateDB.Copy()
	beaconBestState.beaconCommitteeState = target.beaconCommitteeState.Clone()
	beaconBestState.missingSignatureCounter = target.missingSignatureCounter.Copy()

	if beaconBestState.pdeStates == nil {
		beaconBestState.pdeStates = make(map[uint]pdex.State)
	}
	for version, state := range target.pdeStates {
		if state != nil {
			beaconBestState.pdeStates[version] = state.Clone()
		}
	}
	if target.portalStateV3 != nil {
		beaconBestState.portalStateV3 = target.portalStateV3.Copy()
	}
	if target.portalStateV4 != nil {
		beaconBestState.portalStateV4 = target.portalStateV4.Copy()
	}
	if beaconBestState.bridgeAggManager != nil {
		beaconBestState.bridgeAggManager = target.bridgeAggManager.Clone()
	}

	return nil
}

func (beaconBestState *BeaconBestState) CloneBeaconBestStateFrom(target *BeaconBestState) error {
	return beaconBestState.cloneBeaconBestStateFrom(target)
}

func (beaconBestState *BeaconBestState) updateLastCrossShardState(shardStates map[byte][]types.ShardState) {
	lastCrossShardState := beaconBestState.LastCrossShardState
	for fromShard, shardBlocks := range shardStates {
		for _, shardBlock := range shardBlocks {
			for _, toShard := range shardBlock.CrossShard {
				if fromShard == toShard {
					continue
				}
				if lastCrossShardState[fromShard] == nil {
					lastCrossShardState[fromShard] = make(map[byte]uint64)
				}
				waitHeight := shardBlock.Height
				lastCrossShardState[fromShard][toShard] = waitHeight
			}
		}
	}
}
func (beaconBestState *BeaconBestState) UpdateLastCrossShardState(shardStates map[byte][]types.ShardState) {
	beaconBestState.updateLastCrossShardState(shardStates)
}

func (beaconBestState *BeaconBestState) GetAutoStakingList() map[string]bool {

	m := make(map[string]bool)
	for k, v := range beaconBestState.beaconCommitteeState.GetAutoStaking() {
		m[k] = v
	}
	return m
}

func (beaconBestState *BeaconBestState) GetAllCommitteeValidatorCandidateFlattenList() []string {
	return beaconBestState.getAllCommitteeValidatorCandidateFlattenList()
}

func (beaconBestState *BeaconBestState) getAllCommitteeValidatorCandidateFlattenList() []string {
	return beaconBestState.beaconCommitteeState.GetAllCandidateSubstituteCommittee()
}

func (beaconBestState *BeaconBestState) getAllCommitteeValidatorCandidateMap() map[string]struct{} {

	list := beaconBestState.beaconCommitteeState.GetAllCandidateSubstituteCommittee()

	m := make(map[string]struct{})
	for _, v := range list {
		m[v] = struct{}{}
	}

	return m
}

func (beaconBestState *BeaconBestState) GetHash() *common.Hash {
	return beaconBestState.BestBlock.Hash()
}

func (beaconBestState *BeaconBestState) GetPreviousHash() *common.Hash {
	return &beaconBestState.BestBlock.Header.PreviousBlockHash
}

func (beaconBestState *BeaconBestState) GetPreviousBlockCommittee(db incdb.Database) ([]incognitokey.CommitteePublicKey, error) {
	panic("not implement")
}

func (beaconBestState *BeaconBestState) GetHeight() uint64 {
	return beaconBestState.BestBlock.GetHeight()
}

func (beaconBestState *BeaconBestState) GetBlockTime() int64 {
	return beaconBestState.BestBlock.Header.Timestamp
}

func (beaconBestState *BeaconBestState) GetNumberOfMissingSignature() map[string]signaturecounter.MissingSignature {
	if beaconBestState.missingSignatureCounter == nil {
		return map[string]signaturecounter.MissingSignature{}
	}
	return beaconBestState.missingSignatureCounter.MissingSignature()
}

func (beaconBestState *BeaconBestState) GetMissingSignaturePenalty() map[string]signaturecounter.Penalty {
	if beaconBestState.missingSignatureCounter == nil {
		return map[string]signaturecounter.Penalty{}
	}
	slashingPenalty := make(map[string]signaturecounter.Penalty)

	if beaconBestState.BeaconHeight >= config.Param().ConsensusParam.EnableSlashingHeightV2 {

		expectedTotalBlock := beaconBestState.GetExpectedTotalBlock(beaconBestState.BestBlock.GetVersion())
		slashingPenalty = beaconBestState.missingSignatureCounter.GetAllSlashingPenaltyWithExpectedTotalBlock(expectedTotalBlock)
		Logger.log.Debug("Get Missing Signature with Slashing V2")
	} else if beaconBestState.BeaconHeight >= config.Param().ConsensusParam.EnableSlashingHeight {

		slashingPenalty = beaconBestState.missingSignatureCounter.GetAllSlashingPenaltyWithActualTotalBlock()
		Logger.log.Debug("Get Missing Signature with Slashing V1")
	}

	return slashingPenalty
}

func (beaconBestState *BeaconBestState) BridgeAggManager() *bridgeagg.Manager {
	return beaconBestState.bridgeAggManager
}

func (beaconBestState *BeaconBestState) PdeState(version uint) pdex.State {
	return beaconBestState.pdeStates[version]
}

func (beaconBestState *BeaconBestState) BlockHash() common.Hash {
	return beaconBestState.BestBlockHash
}

func (beaconBestState *BeaconBestState) GetAllCommitteeValidatorCandidate() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	sC := make(map[byte][]incognitokey.CommitteePublicKey)
	sPV := make(map[byte][]incognitokey.CommitteePublicKey)
	sSP := make(map[byte][]incognitokey.CommitteePublicKey)
	for shardID, committee := range beaconBestState.GetShardCommittee() {
		sC[shardID] = append([]incognitokey.CommitteePublicKey{}, committee...)
	}
	for shardID, Substitute := range beaconBestState.GetShardPendingValidator() {
		sPV[shardID] = append([]incognitokey.CommitteePublicKey{}, Substitute...)
	}
	for shardID, syncValidator := range beaconBestState.GetSyncingValidators() {
		sSP[shardID] = append([]incognitokey.CommitteePublicKey{}, syncValidator...)
	}
	bC := beaconBestState.beaconCommitteeState.GetBeaconCommittee()
	bPV := beaconBestState.beaconCommitteeState.GetBeaconSubstitute()
	cBWFCR := beaconBestState.beaconCommitteeState.GetCandidateBeaconWaitingForCurrentRandom()
	cBWFNR := beaconBestState.beaconCommitteeState.GetCandidateBeaconWaitingForNextRandom()
	cSWFCR := beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForCurrentRandom()
	cSWFNR := beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForNextRandom()
	return sC, sPV, sSP, bC, bPV, cBWFCR, cBWFNR, cSWFCR, cSWFNR, nil
}

func (beaconBestState *BeaconBestState) GetValidStakers(stakers []string) []string {
	for _, committees := range beaconBestState.GetShardCommittee() {
		committeesStr, err := incognitokey.CommitteeKeyListToString(committees)
		if err != nil {
			panic(err)
		}
		stakers = common.GetValidStaker(committeesStr, stakers)
	}
	for _, validators := range beaconBestState.GetShardPendingValidator() {
		validatorsStr, err := incognitokey.CommitteeKeyListToString(validators)
		if err != nil {
			panic(err)
		}
		stakers = common.GetValidStaker(validatorsStr, stakers)
	}
	beaconCommittee := beaconBestState.beaconCommitteeState.GetBeaconCommittee()
	beaconCommitteeStr, err := incognitokey.CommitteeKeyListToString(beaconCommittee)
	if err != nil {
		panic(err)
	}
	stakers = common.GetValidStaker(beaconCommitteeStr, stakers)
	beaconSubstitute := beaconBestState.beaconCommitteeState.GetBeaconSubstitute()
	beaconSubstituteStr, err := incognitokey.CommitteeKeyListToString(beaconSubstitute)
	if err != nil {
		panic(err)
	}
	stakers = common.GetValidStaker(beaconSubstituteStr, stakers)
	candidateBeaconWaitingForCurrentRandom := beaconBestState.beaconCommitteeState.GetCandidateBeaconWaitingForCurrentRandom()
	candidateBeaconWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(candidateBeaconWaitingForCurrentRandom)
	if err != nil {
		panic(err)
	}
	stakers = common.GetValidStaker(candidateBeaconWaitingForCurrentRandomStr, stakers)
	candidateBeaconWaitingForNextRandom := beaconBestState.beaconCommitteeState.GetCandidateBeaconWaitingForNextRandom()
	candidateBeaconWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(candidateBeaconWaitingForNextRandom)
	if err != nil {
		panic(err)
	}
	stakers = common.GetValidStaker(candidateBeaconWaitingForNextRandomStr, stakers)
	candidateShardWaitingForCurrentRandom := beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForCurrentRandom()
	candidateShardWaitingForCurrentRandomStr, err := incognitokey.CommitteeKeyListToString(candidateShardWaitingForCurrentRandom)
	if err != nil {
		panic(err)
	}
	stakers = common.GetValidStaker(candidateShardWaitingForCurrentRandomStr, stakers)
	candidateShardWaitingForNextRandom := beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForNextRandom()
	candidateShardWaitingForNextRandomStr, err := incognitokey.CommitteeKeyListToString(candidateShardWaitingForNextRandom)
	if err != nil {
		panic(err)
	}
	stakers = common.GetValidStaker(candidateShardWaitingForNextRandomStr, stakers)
	return stakers
}

func (beaconBestState *BeaconBestState) GetAllCommitteeValidatorCandidateFlattenListFromDatabase() ([]string, error) {
	return beaconBestState.beaconCommitteeState.GetAllCandidateSubstituteCommittee(), nil
}

func (beaconBestState *BeaconBestState) GetAllBridgeTokens() ([]common.Hash, error) {
	bridgeTokenIDs := []common.Hash{}
	allBridgeTokens := []*rawdbv2.BridgeTokenInfo{}
	bridgeStateDB := beaconBestState.GetBeaconFeatureStateDB()
	allBridgeTokensBytes, err := statedb.GetAllBridgeTokens(bridgeStateDB)
	if err != nil {
		return bridgeTokenIDs, err
	}
	err = json.Unmarshal(allBridgeTokensBytes, &allBridgeTokens)
	if err != nil {
		return bridgeTokenIDs, err
	}
	for _, bridgeTokenInfo := range allBridgeTokens {
		bridgeTokenIDs = append(bridgeTokenIDs, *bridgeTokenInfo.TokenID)
	}
	return bridgeTokenIDs, nil
}

func (beaconBestState BeaconBestState) NewBeaconCommitteeStateEnvironmentWithValue(
	beaconInstructions [][]string,
	isFoundRandomInstruction bool,
	isBeaconRandomTime bool,
) *committeestate.BeaconCommitteeStateEnvironment {
	slashingPenalty := make(map[string]signaturecounter.Penalty)
	if beaconBestState.BeaconHeight != 1 &&
		beaconBestState.CommitteeStateVersion() >= committeestate.STAKING_FLOW_V2 {
		if beaconBestState.BeaconHeight >= config.Param().ConsensusParam.EnableSlashingHeightV2 {
			expectedTotalBlock := beaconBestState.GetExpectedTotalBlock(beaconBestState.BestBlock.GetVersion())
			slashingPenalty = beaconBestState.missingSignatureCounter.GetAllSlashingPenaltyWithExpectedTotalBlock(expectedTotalBlock)
			Logger.log.Debug("Get Missing Signature with Slashing V2")
		} else if beaconBestState.BeaconHeight >= config.Param().ConsensusParam.EnableSlashingHeight {
			slashingPenalty = beaconBestState.missingSignatureCounter.GetAllSlashingPenaltyWithActualTotalBlock()
			Logger.log.Debug("Get Missing Signature with Slashing V1")
		}
	} else {
		slashingPenalty = make(map[string]signaturecounter.Penalty)
	}
	return &committeestate.BeaconCommitteeStateEnvironment{
		BeaconHeight:                     beaconBestState.BeaconHeight,
		BeaconHash:                       beaconBestState.BestBlockHash,
		Epoch:                            beaconBestState.Epoch,
		EpochLengthV1:                    config.Param().EpochParam.NumberOfBlockInEpoch,
		BeaconInstructions:               beaconInstructions,
		EpochBreakPointSwapNewKey:        config.Param().ConsensusParam.EpochBreakPointSwapNewKey,
		AssignOffset:                     config.Param().SwapCommitteeParam.AssignOffset,
		RandomNumber:                     beaconBestState.CurrentRandomNumber,
		IsFoundRandomNumber:              isFoundRandomInstruction,
		IsBeaconRandomTime:               isBeaconRandomTime,
		ActiveShards:                     beaconBestState.ActiveShards,
		MinShardCommitteeSize:            beaconBestState.MinShardCommitteeSize,
		ConsensusStateDB:                 beaconBestState.consensusStateDB,
		NumberOfFixedShardBlockValidator: beaconBestState.NumberOfFixedShardBlockValidator,
		MaxShardCommitteeSize:            beaconBestState.MaxShardCommitteeSize,
		MissingSignaturePenalty:          slashingPenalty,
		PreviousBlockHashes: &committeestate.BeaconCommitteeStateHash{
			BeaconCandidateHash:             beaconBestState.BestBlock.Header.BeaconCandidateRoot,
			BeaconCommitteeAndValidatorHash: beaconBestState.BestBlock.Header.BeaconCommitteeAndValidatorRoot,
			ShardCandidateHash:              beaconBestState.BestBlock.Header.ShardCandidateRoot,
			ShardCommitteeAndValidatorHash:  beaconBestState.BestBlock.Header.ShardCommitteeAndValidatorRoot,
			AutoStakeHash:                   beaconBestState.BestBlock.Header.AutoStakingRoot,
			ShardSyncValidatorsHash:         beaconBestState.BestBlock.Header.ShardSyncValidatorRoot,
		},
	}
}

func (beaconBestState BeaconBestState) NewBeaconCommitteeStateEnvironment() *committeestate.BeaconCommitteeStateEnvironment {
	slashingPenalty := make(map[string]signaturecounter.Penalty)
	if beaconBestState.BeaconHeight != 1 &&
		beaconBestState.CommitteeStateVersion() >= committeestate.STAKING_FLOW_V2 {
		if beaconBestState.BeaconHeight >= config.Param().ConsensusParam.EnableSlashingHeightV2 {
			expectedTotalBlock := beaconBestState.GetExpectedTotalBlock(beaconBestState.BestBlock.GetVersion())
			slashingPenalty = beaconBestState.missingSignatureCounter.GetAllSlashingPenaltyWithExpectedTotalBlock(expectedTotalBlock)
			Logger.log.Debug("Get Missing Signature with Slashing V2")
		} else if beaconBestState.BeaconHeight >= config.Param().ConsensusParam.EnableSlashingHeight {
			slashingPenalty = beaconBestState.missingSignatureCounter.GetAllSlashingPenaltyWithActualTotalBlock()
			Logger.log.Debug("Get Missing Signature with Slashing V1")
		}
	} else {
		slashingPenalty = make(map[string]signaturecounter.Penalty)
	}
	return &committeestate.BeaconCommitteeStateEnvironment{
		BeaconHeight:                     beaconBestState.BeaconHeight,
		BeaconHash:                       beaconBestState.BestBlockHash,
		Epoch:                            beaconBestState.Epoch,
		EpochLengthV1:                    config.Param().EpochParam.NumberOfBlockInEpoch,
		RandomNumber:                     beaconBestState.CurrentRandomNumber,
		ActiveShards:                     beaconBestState.ActiveShards,
		MinShardCommitteeSize:            beaconBestState.MinShardCommitteeSize,
		ConsensusStateDB:                 beaconBestState.consensusStateDB,
		MaxShardCommitteeSize:            beaconBestState.MaxShardCommitteeSize,
		NumberOfFixedShardBlockValidator: beaconBestState.NumberOfFixedShardBlockValidator,
		MissingSignaturePenalty:          slashingPenalty,
		StakingV3Height:                  config.Param().ConsensusParam.StakingFlowV3Height,
		StakingV2Height:                  config.Param().ConsensusParam.StakingFlowV2Height,
		AssignRuleV3Height:               config.Param().ConsensusParam.AssignRuleV3Height,
	}
}

func (beaconBestState *BeaconBestState) restoreCommitteeState(bc *BlockChain) error {
	Logger.log.Infof("Init Beacon Committee State %+v", beaconBestState.BeaconHeight)
	shardIDs := []int{statedb.BeaconChainID}
	for i := 0; i < beaconBestState.ActiveShards; i++ {
		shardIDs = append(shardIDs, i)
	}
	currentValidator, substituteValidator, nextEpochShardCandidate,
		currentEpochShardCandidate, _, _, syncingValidators,
		rewardReceivers, autoStaking, stakingTx := statedb.GetAllCandidateSubstituteCommittee(beaconBestState.consensusStateDB, shardIDs)
	beaconCommittee := currentValidator[statedb.BeaconChainID]
	delete(currentValidator, statedb.BeaconChainID)
	delete(substituteValidator, statedb.BeaconChainID)
	shardCommittee := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range currentValidator {
		shardCommittee[byte(k)] = v
	}
	shardSubstitute := make(map[byte][]incognitokey.CommitteePublicKey)
	for k, v := range substituteValidator {
		shardSubstitute[byte(k)] = v
	}

	//init version of committeeState here
	version := committeestate.VersionByBeaconHeight(
		beaconBestState.BeaconHeight,
		config.Param().ConsensusParam.StakingFlowV2Height,
		config.Param().ConsensusParam.StakingFlowV3Height)

	shardCommonPool := []incognitokey.CommitteePublicKey{}
	numberOfAssignedCandidates := 0
	var swapRule committeestate.SwapRuleProcessor
	assignRule := committeestate.GetAssignRuleVersion(
		beaconBestState.BeaconHeight,
		config.Param().ConsensusParam.StakingFlowV2Height,
		config.Param().ConsensusParam.AssignRuleV3Height,
	)

	if version >= committeestate.STAKING_FLOW_V2 {
		shardCommonPool = nextEpochShardCandidate
		swapRule = committeestate.GetSwapRuleVersion(beaconBestState.BeaconHeight, config.Param().ConsensusParam.StakingFlowV3Height)

		if bc.IsEqualToRandomTime(beaconBestState.BeaconHeight) {
			var err error
			var randomTimeBeaconHash = beaconBestState.BestBlockHash

			tempRootHash, err := GetBeaconRootsHashByBlockHash(bc.GetBeaconChainDatabase(), randomTimeBeaconHash)
			if err != nil {
				panic(err)
			}
			dbWarper := statedb.NewDatabaseAccessWarper(bc.GetBeaconChainDatabase())
			consensusSnapshotTimeStateDB, _ := statedb.NewWithPrefixTrie(tempRootHash.ConsensusStateDBRootHash, dbWarper)
			snapshotCurrentValidator, snapshotSubstituteValidator, snapshotNextEpochShardCandidate,
				_, _, _, _, _, _, _ := statedb.GetAllCandidateSubstituteCommittee(consensusSnapshotTimeStateDB, shardIDs)
			snapshotShardCommonPool, _ := incognitokey.CommitteeKeyListToString(snapshotNextEpochShardCandidate)
			snapshotShardCommittee := make(map[byte][]string)
			snapshotShardSubstitute := make(map[byte][]string)
			delete(snapshotCurrentValidator, statedb.BeaconChainID)
			delete(snapshotSubstituteValidator, statedb.BeaconChainID)
			for k, v := range snapshotCurrentValidator {
				snapshotShardCommittee[byte(k)], _ = incognitokey.CommitteeKeyListToString(v)
			}
			for k, v := range snapshotSubstituteValidator {
				snapshotShardSubstitute[byte(k)], _ = incognitokey.CommitteeKeyListToString(v)
			}

			numberOfAssignedCandidates = committeestate.SnapshotShardCommonPoolV2(
				snapshotShardCommonPool,
				snapshotShardCommittee,
				snapshotShardSubstitute,
				beaconBestState.NumberOfFixedShardBlockValidator,
				beaconBestState.MinShardCommitteeSize,
				swapRule,
			)
		}
	}

	committeeState := committeestate.NewBeaconCommitteeState(
		version, beaconCommittee, shardCommittee, shardSubstitute, shardCommonPool,
		numberOfAssignedCandidates, autoStaking, rewardReceivers, stakingTx, syncingValidators,
		swapRule, nextEpochShardCandidate, currentEpochShardCandidate, assignRule,
	)
	beaconBestState.beaconCommitteeState = committeeState
	if err := beaconBestState.tryUpgradeConsensusRule(); err != nil {
		return err
	}
	return nil
}

func (beaconBestState *BeaconBestState) initMissingSignatureCounter(bc *BlockChain, beaconBlock *types.BeaconBlock) error {
	committees := beaconBestState.GetShardCommitteeFlattenList()
	missingSignatureCounter := signaturecounter.NewDefaultSignatureCounter(committees)
	beaconBestState.SetMissingSignatureCounter(missingSignatureCounter)

	firstBeaconHeightOfEpoch := GetFirstBeaconHeightInEpoch(beaconBestState.Epoch)
	tempBeaconBlock := beaconBlock
	tempBeaconHeight := beaconBlock.Header.Height
	allShardStates := make(map[byte][]types.ShardState)

	for tempBeaconHeight >= firstBeaconHeightOfEpoch {
		for shardID, shardStates := range tempBeaconBlock.Body.ShardState {
			allShardStates[shardID] = append(shardStates, allShardStates[shardID]...)
		}
		if tempBeaconHeight == 1 {
			break
		}
		previousBeaconBlock, _, err := bc.GetBeaconBlockByHash(tempBeaconBlock.Header.PreviousBlockHash)
		if err != nil {
			return err
		}
		tempBeaconBlock = previousBeaconBlock
		tempBeaconHeight--
	}

	return beaconBestState.countMissingSignature(bc, allShardStates)
}

func (beaconBestState *BeaconBestState) CandidateWaitingForNextRandom() []incognitokey.CommitteePublicKey {
	return beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForNextRandom()
}

func (bc *BlockChain) GetTotalStaker() (int, error) {
	// var beaconConsensusRootHash common.Hash
	bcBestState := bc.GetBeaconBestState()
	beaconConsensusRootHash, err := bc.GetBeaconConsensusRootHash(bcBestState, bcBestState.GetHeight())
	if err != nil {
		return 0, fmt.Errorf("Beacon Consensus Root Hash of Height %+v not found ,error %+v", bcBestState.GetHeight(), err)
	}
	beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(bc.GetBeaconChainDatabase()))
	if err != nil {
		return 0, fmt.Errorf("init beacon consensus statedb return error %v", err)
	}
	return statedb.GetAllStaker(beaconConsensusStateDB, bc.GetShardIDs()), nil
}

func (beaconBestState *BeaconBestState) tryUpgradeConsensusRule() error {

	if beaconBestState.BeaconHeight == config.Param().ConsensusParam.StakingFlowV2Height ||
		beaconBestState.BeaconHeight == config.Param().ConsensusParam.StakingFlowV3Height {
		if err := beaconBestState.tryUpgradeCommitteeState(); err != nil {
			return err
		}
	}

	if beaconBestState.BeaconHeight == config.Param().ConsensusParam.BlockProducingV3Height {
		if err := beaconBestState.checkBlockProducingV3Config(); err != nil {
			return err
		}
		if err := beaconBestState.upgradeBlockProducingV3Config(); err != nil {
			return err
		}
		Logger.log.Infof("Upgrade Block Producing V3, min %+v, max %+v",
			beaconBestState.MinShardCommitteeSize, beaconBestState.MaxShardCommitteeSize)
	}

	if beaconBestState.BeaconHeight == config.Param().ConsensusParam.AssignRuleV3Height {
		beaconBestState.upgradeAssignRuleV3()
	}

	return nil
}

// tryUpgradeCommitteeState only allow
// Upgrade to v2 if current version is 1 and beacon height == staking flow v2 height
// Upgrade to v3 if current version is 2 and beacon height == staking flow v3 height
func (beaconBestState *BeaconBestState) tryUpgradeCommitteeState() error {
	if beaconBestState.BeaconHeight != config.Param().ConsensusParam.StakingFlowV3Height &&
		beaconBestState.BeaconHeight != config.Param().ConsensusParam.StakingFlowV2Height {
		return nil
	}
	if beaconBestState.BeaconHeight == config.Param().ConsensusParam.StakingFlowV3Height {
		if beaconBestState.beaconCommitteeState.Version() != committeestate.STAKING_FLOW_V2 {
			return nil
		}
		if beaconBestState.beaconCommitteeState.Version() == committeestate.STAKING_FLOW_V3 {
			return nil
		}
	}
	if beaconBestState.BeaconHeight == config.Param().ConsensusParam.StakingFlowV2Height {
		if beaconBestState.beaconCommitteeState.Version() != committeestate.SELF_SWAP_SHARD_VERSION {
			return nil
		}
		if beaconBestState.beaconCommitteeState.Version() == committeestate.STAKING_FLOW_V2 {
			return nil
		}
	}

	Logger.log.Infof("Try Upgrade Staking Flow, current version %+v, beacon height %+v"+
		"Staking Flow v2 %+v, Staking Flow v3 %+v",
		beaconBestState.beaconCommitteeState.Version(), beaconBestState.BeaconHeight,
		config.Param().ConsensusParam.StakingFlowV2Height, config.Param().ConsensusParam.StakingFlowV3Height)

	env := committeestate.NewBeaconCommitteeStateEnvironmentForUpgrading(
		beaconBestState.BeaconHeight,
		config.Param().ConsensusParam.StakingFlowV2Height,
		config.Param().ConsensusParam.AssignRuleV3Height,
		config.Param().ConsensusParam.StakingFlowV3Height,
		beaconBestState.BestBlockHash,
	)

	committeeState := beaconBestState.beaconCommitteeState.Upgrade(env)
	beaconBestState.beaconCommitteeState = committeeState

	return nil
}

func (beaconBestState *BeaconBestState) checkBlockProducingV3Config() error {

	for shardID, shardCommittee := range beaconBestState.GetShardCommittee() {
		shardCommitteeSize := len(shardCommittee)
		if shardCommitteeSize < SFV3_MinShardCommitteeSize {
			return fmt.Errorf("shard %+v | current committee length %+v can not upgrade to staking flow v3, "+
				"minimum required committee size is 8", shardID, shardCommitteeSize)
		}
	}

	return nil
}

func (beaconBestState *BeaconBestState) upgradeBlockProducingV3Config() error {

	if beaconBestState.MinShardCommitteeSize < SFV3_MinShardCommitteeSize {
		beaconBestState.MinShardCommitteeSize = SFV3_MinShardCommitteeSize
		Logger.log.Infof("BEACON | Set beaconBestState.MinShardCommitteeSize from %+v to %+v ",
			beaconBestState.MinShardCommitteeSize, SFV3_MinShardCommitteeSize)
	}

	if beaconBestState.NumberOfFixedShardBlockValidator < SFV3_MinShardCommitteeSize {
		beaconBestState.NumberOfFixedShardBlockValidator = SFV3_MinShardCommitteeSize
		Logger.log.Infof("BEACON | Set beaconBestState.NumberOfFixedShardBlockValidator from %+v to %+v ",
			beaconBestState.NumberOfFixedShardBlockValidator, SFV3_MinShardCommitteeSize)
	}

	if beaconBestState.MaxShardCommitteeSize < SFV3_MinShardCommitteeSize {
		beaconBestState.MaxShardCommitteeSize = SFV3_MinShardCommitteeSize
		Logger.log.Infof("BEACON | Set beaconBestState.MaxShardCommitteeSize from %+v to %+v ",
			beaconBestState.MaxShardCommitteeSize, SFV3_MinShardCommitteeSize)
	}

	return nil
}

func (beaconBestState *BeaconBestState) ExtractPendingAndCommittee(validatorFromUserKeys []*consensus.Validator) ([]*consensus.Validator, []string) {
	if len(validatorFromUserKeys) == 0 {
		return []*consensus.Validator{}, []string{}
	}
	beaconValidators := beaconBestState.beaconCommitteeState.GetBeaconCommittee()
	shardValidators := beaconBestState.beaconCommitteeState.GetShardCommittee()
	shardSubstitutes := beaconBestState.beaconCommitteeState.GetShardSubstitute()
	userKeys := []*consensus.Validator{}
	validatorString := []string{}

	for _, v := range beaconValidators {
		blsKey := v.GetMiningKeyBase58(common.BlsConsensus)
		for _, userKey := range validatorFromUserKeys {
			if blsKey == userKey.MiningKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus) {
				userKeys = append(userKeys, userKey)
				temp, _ := v.ToBase58()
				validatorString = append(validatorString, temp)
				break
			}
		}
	}

	for _, validators := range shardValidators {
		for _, v := range validators {
			blsKey := v.GetMiningKeyBase58(common.BlsConsensus)
			for _, userKey := range validatorFromUserKeys {
				if blsKey == userKey.MiningKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus) {
					userKeys = append(userKeys, userKey)
					temp, _ := v.ToBase58()
					validatorString = append(validatorString, temp)
					break
				}
			}
		}
	}

	for _, validators := range shardSubstitutes {
		for _, v := range validators {
			blsKey := v.GetMiningKeyBase58(common.BlsConsensus)
			for _, userKey := range validatorFromUserKeys {
				if blsKey == userKey.MiningKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus) {
					userKeys = append(userKeys, userKey)
					temp, _ := v.ToBase58()
					validatorString = append(validatorString, temp)
					break
				}
			}
		}
	}
	return userKeys, validatorString
}

func (beaconBestState *BeaconBestState) ExtractAllFinishSyncingValidators(validatorFromUserKeys []*consensus.Validator) ([]*consensus.Validator, []string) {
	if len(validatorFromUserKeys) == 0 {
		return []*consensus.Validator{}, []string{}
	}
	finishedSyncUserKeys := []*consensus.Validator{}
	finishedSyncValidators := []string{}

	for sid := 0; sid < beaconBestState.ActiveShards; sid++ {
		syncingValidators := beaconBestState.beaconCommitteeState.GetSyncingValidators()[byte(sid)]
		for _, v := range syncingValidators {
			blsKey := v.GetMiningKeyBase58(common.BlsConsensus)
			for _, userKey := range validatorFromUserKeys {
				if blsKey == userKey.MiningKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus) {
					finishedSyncUserKeys = append(finishedSyncUserKeys, userKey)
					temp, _ := v.ToBase58()
					finishedSyncValidators = append(finishedSyncValidators, temp)
					break
				}
			}
		}
	}

	return finishedSyncUserKeys, finishedSyncValidators
}

func (beaconBestState *BeaconBestState) ExtractFinishSyncingValidators(validatorFromUserKeys []*consensus.Validator, shardID byte) ([]*consensus.Validator, []string) {
	if len(validatorFromUserKeys) == 0 {
		return []*consensus.Validator{}, []string{}
	}
	syncingValidators := beaconBestState.beaconCommitteeState.GetSyncingValidators()[shardID]
	finishedSyncUserKeys := []*consensus.Validator{}
	finishedSyncValidators := []string{}

	for _, v := range syncingValidators {
		blsKey := v.GetMiningKeyBase58(common.BlsConsensus)
		for _, userKey := range validatorFromUserKeys {
			if blsKey == userKey.MiningKey.GetPublicKey().GetMiningKeyBase58(common.BlsConsensus) {
				finishedSyncUserKeys = append(finishedSyncUserKeys, userKey)
				temp, _ := v.ToBase58()
				finishedSyncValidators = append(finishedSyncValidators, temp)
				break
			}
		}
	}
	return finishedSyncUserKeys, finishedSyncValidators
}

func (beaconBestState *BeaconBestState) removeFinishedSyncValidators(committeeChange *committeestate.CommitteeChange) {
	for shardID, finishSyncValidators := range committeeChange.FinishedSyncValidators {
		finishsync.DefaultFinishSyncMsgPool.RemoveValidators(finishSyncValidators, byte(shardID))
	}
}

func (beaconBestState *BeaconBestState) upgradeAssignRuleV3() {

	if beaconBestState.CommitteeStateVersion() == committeestate.STAKING_FLOW_V2 {
		if beaconBestState.beaconCommitteeState.AssignRuleVersion() == committeestate.ASSIGN_RULE_V2 {
			beaconBestState.beaconCommitteeState.(*committeestate.BeaconCommitteeStateV2).UpgradeAssignRuleV3()
			Logger.log.Infof("BEACON | Beacon Height %+v, UPGRADE Assign Rule from V2 to V3", beaconBestState.BeaconHeight)

		}
	}

}

func (b *BeaconBestState) GetExpectedTotalBlock(blockVersion int) map[string]uint {

	expectedTotalBlock := make(map[string]uint)
	expectedBlockForShards := b.CalculateExpectedTotalBlock(blockVersion)

	for shardID, committees := range b.GetShardCommittee() {
		expectedBlockForShard := expectedBlockForShards[shardID]
		for _, committee := range committees {
			temp, _ := committee.ToBase58()
			expectedTotalBlock[temp] = expectedBlockForShard
		}
	}

	return expectedTotalBlock
}

func (b *BeaconBestState) CalculateExpectedTotalBlock(blockVersion int) map[byte]uint {

	mean := uint(0)

	subsetNumberOfShardBlock := make(map[byte]uint)

	for shardID, numberOfBlock := range b.NumberOfShardBlock {
		if blockVersion >= types.BLOCK_PRODUCINGV3_VERSION && blockVersion < types.INSTANT_FINALITY_VERSION_V2 {
			subsetNumberOfShardBlock[shardID] = numberOfBlock / 2
		} else {
			subsetNumberOfShardBlock[shardID] = numberOfBlock
		}
	}

	for _, v := range subsetNumberOfShardBlock {
		mean += v
	}

	mean = mean / uint(len(subsetNumberOfShardBlock))

	expectedTotalBlock := make(map[byte]uint)
	for k, v := range subsetNumberOfShardBlock {
		if v <= mean {
			expectedTotalBlock[k] = mean
		} else {
			expectedTotalBlock[k] = v
		}
	}

	return expectedTotalBlock
}

func (beaconBestState *BeaconBestState) GetNonSlashingCommittee(committees []*statedb.StakerInfoSlashingVersion, epoch uint64, shardID byte) ([]*statedb.StakerInfoSlashingVersion, error) {

	if epoch >= beaconBestState.Epoch {
		return nil, fmt.Errorf("Can't get committee to pay salary because, BeaconBestState Epoch %+v is"+
			"equal to lower than want epoch %+v", beaconBestState.Epoch, epoch)
	}

	slashingCommittees := statedb.GetSlashingCommittee(beaconBestState.slashStateDB, epoch)

	return filterNonSlashingCommittee(committees, slashingCommittees[shardID]), nil
}

func (curView *BeaconBestState) getUntriggerFeature(afterCheckPoint bool) []string {
	unTriggerFeatures := []string{}
	for f, _ := range config.Param().AutoEnableFeature {
		if config.Param().AutoEnableFeature[f].MinTriggerBlockHeight == 0 {
			//skip default value
			continue
		}
		if curView.TriggeredFeature == nil || curView.TriggeredFeature[f] == 0 {
			if afterCheckPoint {
				if curView.BeaconHeight > uint64(config.Param().AutoEnableFeature[f].MinTriggerBlockHeight) {
					unTriggerFeatures = append(unTriggerFeatures, f)
				}
			} else {
				unTriggerFeatures = append(unTriggerFeatures, f)
			}

		}
	}
	return unTriggerFeatures
}

func filterNonSlashingCommittee(committees []*statedb.StakerInfoSlashingVersion, slashingCommittees []string) []*statedb.StakerInfoSlashingVersion {

	nonSlashingCommittees := []*statedb.StakerInfoSlashingVersion{}
	tempSlashingCommittees := make(map[string]struct{})

	for _, committee := range slashingCommittees {
		tempSlashingCommittees[committee] = struct{}{}
	}

	for _, committee := range committees {
		_, ok := tempSlashingCommittees[committee.CommitteePublicKey()]
		if !ok {
			nonSlashingCommittees = append(nonSlashingCommittees, committee)
		}
	}

	return nonSlashingCommittees
}

func GetMaxCommitteeSize(currentMaxShardCommittee int, triggerFeature map[string]uint64, blkHeight uint64) int {
	if triggerFeature[MAX_COMMITTEE_SIZE_48_FEATURE] != 0 && triggerFeature[MAX_COMMITTEE_SIZE_48_FEATURE] <= blkHeight {
		return 48
	}
	return currentMaxShardCommittee
}

func (curView *BeaconBestState) GetProposerLength() int {
	return curView.MinBeaconCommitteeSize
}

func (curView *BeaconBestState) GetShardProposerLength() int {
	return curView.NumberOfFixedShardBlockValidator
}

func (x *BeaconBestState) CompareCommitteeFromBlock(_y multiview.View) int {
	return 0
}
