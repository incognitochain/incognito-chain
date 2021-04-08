package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate/finishsync"
	"github.com/incognitochain/incognito-chain/blockchain/signaturecounter"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/privacy"
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
	BestBlockHash           common.Hash          `json:"BestBlockHash"`         // The hash of the block.
	PreviousBestBlockHash   common.Hash          `json:"PreviousBestBlockHash"` // The hash of the block.
	BestBlock               types.BeaconBlock    `json:"BestBlock"`             // The block.
	BestShardHash           map[byte]common.Hash `json:"BestShardHash"`
	BestShardHeight         map[byte]uint64      `json:"BestShardHeight"`
	Epoch                   uint64               `json:"Epoch"`
	BeaconHeight            uint64               `json:"BeaconHeight"`
	BeaconProposerIndex     int                  `json:"BeaconProposerIndex"`
	CurrentRandomNumber     int64                `json:"CurrentRandomNumber"`
	CurrentRandomTimeStamp  int64                `json:"CurrentRandomTimeStamp"` // random timestamp for this epoch
	IsGetRandomNumber       bool                 `json:"IsGetRandomNumber"`
	Params                  map[string]string    `json:"Params,omitempty"`
	MaxBeaconCommitteeSize  int                  `json:"MaxBeaconCommitteeSize"`
	MinBeaconCommitteeSize  int                  `json:"MinBeaconCommitteeSize"`
	MaxShardCommitteeSize   int                  `json:"MaxShardCommitteeSize"`
	MinShardCommitteeSize   int                  `json:"MinShardCommitteeSize"`
	ActiveShards            int                  `json:"ActiveShards"`
	ConsensusAlgorithm      string               `json:"ConsensusAlgorithm"`
	ShardConsensusAlgorithm map[byte]string      `json:"ShardConsensusAlgorithm"`
	// key: public key of committee, value: payment address reward receiver
	beaconCommitteeState    committeestate.BeaconCommitteeState
	missingSignatureCounter signaturecounter.IMissingSignatureCounter
	finishSyncManager       *finishsync.FinishSyncManager
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
	return beaconBestState
}
func NewBeaconBestStateWithConfig(netparam *Params,
	beaconCommitteeEngine committeestate.BeaconCommitteeState,
) *BeaconBestState {
	beaconBestState := NewBeaconBestState()
	beaconBestState.BestBlockHash.SetBytes(make([]byte, 32))
	beaconBestState.BestShardHeight = make(map[byte]uint64)
	beaconBestState.BestShardHash = make(map[byte]common.Hash)
	beaconBestState.BeaconHeight = 0
	beaconBestState.Params = make(map[string]string)
	beaconBestState.CurrentRandomNumber = -1
	beaconBestState.MaxBeaconCommitteeSize = netparam.MaxBeaconCommitteeSize
	beaconBestState.MinBeaconCommitteeSize = netparam.MinBeaconCommitteeSize
	beaconBestState.MaxShardCommitteeSize = netparam.MaxShardCommitteeSize
	beaconBestState.MinShardCommitteeSize = netparam.MinShardCommitteeSize
	beaconBestState.ActiveShards = netparam.ActiveShards
	beaconBestState.LastCrossShardState = make(map[byte]map[byte]uint64)
	beaconBestState.BlockInterval = netparam.MinBeaconBlockInterval
	beaconBestState.BlockMaxCreateTime = netparam.MaxBeaconBlockCreation
	beaconBestState.beaconCommitteeState = beaconCommitteeEngine
	return beaconBestState
}

func (curView *BeaconBestState) SetMissingSignatureCounter(missingSignatureCounter signaturecounter.IMissingSignatureCounter) {
	curView.missingSignatureCounter = missingSignatureCounter
}

func (bc *BlockChain) GetBeaconBestState() *BeaconBestState {
	return bc.BeaconChain.multiView.GetBestView().(*BeaconBestState)
}

func (bc *BlockChain) GetChain(cid int) Chain {
	if cid == -1 {
		return bc.BeaconChain
	} else {
		return bc.ShardChain[cid]
	}
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

func (beaconBestState *BeaconBestState) GetProposerByTimeSlot(ts int64, version int) (incognitokey.CommitteePublicKey, int) {
	id := GetProposerByTimeSlot(ts, beaconBestState.MinBeaconCommitteeSize)
	return beaconBestState.GetBeaconCommittee()[id], id
}

func (beaconBestState *BeaconBestState) GetBlock() types.BlockInterface {
	return &beaconBestState.BestBlock
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

func (beaconBestState *BeaconBestState) SyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	return beaconBestState.beaconCommitteeState.GetSyncingValidators()
}

//CommitteeEngineVersion ...
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
	beaconBestState.finishSyncManager = target.finishSyncManager.Clone()

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
	return beaconBestState.missingSignatureCounter.MissingSignature()
}

func (beaconBestState *BeaconBestState) GetMissingSignaturePenalty() map[string]signaturecounter.Penalty {
	return beaconBestState.missingSignatureCounter.GetAllSlashingPenalty()
}

func (beaconBestState *BeaconBestState) GetAllCommitteeValidatorCandidate() (map[byte][]incognitokey.CommitteePublicKey, map[byte][]incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, []incognitokey.CommitteePublicKey, error) {
	SC := make(map[byte][]incognitokey.CommitteePublicKey)
	SPV := make(map[byte][]incognitokey.CommitteePublicKey)
	for shardID, committee := range beaconBestState.GetShardCommittee() {
		SC[shardID] = append([]incognitokey.CommitteePublicKey{}, committee...)
	}
	for shardID, Substitute := range beaconBestState.GetShardPendingValidator() {
		SPV[shardID] = append([]incognitokey.CommitteePublicKey{}, Substitute...)
	}
	BC := beaconBestState.beaconCommitteeState.GetBeaconCommittee()
	BPV := beaconBestState.beaconCommitteeState.GetBeaconSubstitute()
	CBWFCR := beaconBestState.beaconCommitteeState.GetCandidateBeaconWaitingForCurrentRandom()
	CBWFNR := beaconBestState.beaconCommitteeState.GetCandidateBeaconWaitingForNextRandom()
	CSWFCR := beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForCurrentRandom()
	CSWFNR := beaconBestState.beaconCommitteeState.GetCandidateShardWaitingForNextRandom()
	return SC, SPV, BC, BPV, CBWFCR, CBWFNR, CSWFCR, CSWFNR, nil
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
	params *Params,
	beaconInstructions [][]string,
	isFoundRandomInstruction bool,
	isBeaconRandomTime bool,
) *committeestate.BeaconCommitteeStateEnvironment {
	slashingPenalty := make(map[string]signaturecounter.Penalty)
	if beaconBestState.BeaconHeight != 1 &&
		beaconBestState.CommitteeStateVersion() == committeestate.SLASHING_VERSION &&
		params.EnableSlashingStakingFlowV2 >= beaconBestState.BeaconHeight {
		slashingPenalty = beaconBestState.missingSignatureCounter.GetAllSlashingPenalty()
	} else {
		slashingPenalty = make(map[string]signaturecounter.Penalty)
	}
	return &committeestate.BeaconCommitteeStateEnvironment{
		BeaconHeight:                     beaconBestState.BeaconHeight,
		BeaconHash:                       beaconBestState.BestBlockHash,
		Epoch:                            beaconBestState.Epoch,
		BeaconInstructions:               beaconInstructions,
		EpochBreakPointSwapNewKey:        params.EpochBreakPointSwapNewKey,
		AssignOffset:                     params.AssignOffset,
		RandomNumber:                     beaconBestState.CurrentRandomNumber,
		IsFoundRandomNumber:              isFoundRandomInstruction,
		IsBeaconRandomTime:               isBeaconRandomTime,
		ActiveShards:                     beaconBestState.ActiveShards,
		MinShardCommitteeSize:            beaconBestState.MinShardCommitteeSize,
		ConsensusStateDB:                 beaconBestState.consensusStateDB,
		NumberOfFixedShardBlockValidator: NumberOfFixedShardBlockValidators,
		MaxShardCommitteeSize:            params.MaxShardCommitteeSize,
		MissingSignaturePenalty:          slashingPenalty,
		StakingV3Height:                  params.StakingFlowV3,
		EpochLengthV1:                    params.Epoch,
	}
}

func (beaconBestState BeaconBestState) NewBeaconCommitteeStateEnvironment(
	params *Params,
) *committeestate.BeaconCommitteeStateEnvironment {
	return &committeestate.BeaconCommitteeStateEnvironment{
		BeaconHeight:                     beaconBestState.BeaconHeight,
		BeaconHash:                       beaconBestState.BestBlockHash,
		Epoch:                            beaconBestState.Epoch,
		RandomNumber:                     beaconBestState.CurrentRandomNumber,
		ActiveShards:                     beaconBestState.ActiveShards,
		MinShardCommitteeSize:            beaconBestState.MinShardCommitteeSize,
		ConsensusStateDB:                 beaconBestState.consensusStateDB,
		MaxShardCommitteeSize:            params.MaxShardCommitteeSize,
		NumberOfFixedShardBlockValidator: NumberOfFixedShardBlockValidators,
		MissingSignaturePenalty:          beaconBestState.missingSignatureCounter.GetAllSlashingPenalty(),
		StakingV3Height:                  params.StakingFlowV3,
		EpochLengthV1:                    params.Epoch,
	}
}

func (beaconBestState *BeaconBestState) initCommitteeEngine(bc *BlockChain) {
	Logger.log.Infof("Init Beacon Committee Engine V1, %+v", beaconBestState.BeaconHeight)
	shardIDs := []int{statedb.BeaconChainID}
	for i := 0; i < beaconBestState.ActiveShards; i++ {
		shardIDs = append(shardIDs, i)
	}
	currentValidator, substituteValidator, nextEpochShardCandidate, currentEpochShardCandidate,
		_, _, syncingValidators,
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
		bc.config.ChainParams.StakingFlowV2,
		bc.config.ChainParams.StakingFlowV3)

	shardCommonPool := []incognitokey.CommitteePublicKey{}
	numberOfAssignedCandidates := 0
	var swapRule committeestate.SwapRuleProcessor

	if version != committeestate.SELF_SWAP_SHARD_VERSION {
		shardCommonPool = nextEpochShardCandidate
		swapRuleEnv := committeestate.NewBeaconCommitteeStateEnvironmentForSwapRule(
			beaconBestState.BeaconHeight, bc.config.ChainParams.StakingFlowV3,
		)
		swapRule = committeestate.SwapRuleByEnv(swapRuleEnv)

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
				NumberOfFixedShardBlockValidators,
				bc.config.ChainParams.MinShardCommitteeSize,
				swapRule,
			)
		} else {
			numberOfAssignedCandidates = 0
		}
	}
	committeeState := committeestate.NewBeaconCommitteeState(
		version, beaconCommittee, shardCommittee, shardSubstitute, shardCommonPool,
		numberOfAssignedCandidates, autoStaking, rewardReceivers, stakingTx, syncingValidators,
		swapRule, nextEpochShardCandidate, currentEpochShardCandidate,
	)

	beaconBestState.beaconCommitteeState = committeeState
}

func initMissingSignatureCounter(bc *BlockChain, curView *BeaconBestState, beaconBlock *types.BeaconBlock) error {
	committees := curView.GetShardCommitteeFlattenList()
	missingSignatureCounter := signaturecounter.NewDefaultSignatureCounter(committees)
	curView.SetMissingSignatureCounter(missingSignatureCounter)

	firstBeaconHeightOfEpoch := bc.GetFirstBeaconHeightInEpoch(curView.Epoch)
	tempBeaconBlock := beaconBlock
	tempBeaconHeight := beaconBlock.Header.Height
	allShardStates := make(map[byte][]types.ShardState)

	for tempBeaconHeight >= firstBeaconHeightOfEpoch {
		for shardID, shardStates := range tempBeaconBlock.Body.ShardState {
			allShardStates[shardID] = append(allShardStates[shardID], shardStates...)
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

	return curView.countMissingSignature(bc, allShardStates)
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

func (beaconBestState *BeaconBestState) upgradeCommitteeEngine(bc *BlockChain) {
	env := committeestate.NewBeaconCommitteeStateEnvironmentForUpgrading(
		beaconBestState.BeaconHeight,
		bc.config.ChainParams.StakingFlowV3,
		beaconBestState.BestBlockHash,
	)

	committeeEngine := beaconBestState.beaconCommitteeState.Upgrade(env)
	beaconBestState.beaconCommitteeState = committeeEngine
}

func (beaconBestState *BeaconBestState) ShouldSendFinishSyncMessage(committeePublicKeys []string, shardID byte) bool {
	if len(committeePublicKeys) == 0 {
		return false
	}
	syncingValidators := beaconBestState.beaconCommitteeState.GetSyncingValidators()[shardID]
	currentSelfSyncingValidators := make(map[string]bool)
	count := 0
	for _, committeePublicKey := range committeePublicKeys {
		currentSelfSyncingValidators[committeePublicKey] = true
	}

	for _, v := range syncingValidators {
		if count == len(committeePublicKeys) {
			return true
		}
		key := v.GetMiningKeyBase58(common.BlsConsensus)
		if currentSelfSyncingValidators[key] {
			count++
		}
	}

	if count == len(committeePublicKeys) {
		return true
	}
	return false
}

func (beaconBestState *BeaconBestState) AddFinishedSyncValidators(committeePublicKeys []string, shardID byte) {
	beaconBestState.finishSyncManager.AddFinishedSyncValidators(
		committeePublicKeys,
		beaconBestState.beaconCommitteeState.GetSyncingValidators()[shardID],
		shardID,
	)
}

func (beaconBestState *BeaconBestState) RemoveFinishedSyncValidators(committeeChange *committeestate.CommitteeChange) {
	for shardID := 0; shardID < beaconBestState.beaconCommitteeState.GetNumberOfActiveShards(); shardID++ {
		committeePublicKeys, _ := incognitokey.CommitteeBase58KeyListToStruct(committeeChange.FinishedSyncValidators[byte(shardID)])
		beaconBestState.finishSyncManager.RemoveValidators(committeePublicKeys, byte(shardID))
	}
}
