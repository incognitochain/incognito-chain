package blockchain

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
)

type BeaconChain struct {
	multiView *multiview.MultiView

	BlockGen    *BlockGenerator
	Blockchain  *BlockChain
	hashHistory *lru.Cache
	ChainName   string
	Ready       bool //when has peerstate

	committeeCache *lru.Cache
	insertLock     sync.Mutex
}

func NewBeaconChain(multiView *multiview.MultiView, blockGen *BlockGenerator, blockchain *BlockChain, chainName string) *BeaconChain {
	return &BeaconChain{multiView: multiView, BlockGen: blockGen, Blockchain: blockchain, ChainName: chainName}
}

func (chain *BeaconChain) GetAllViewHash() (res []common.Hash) {
	for _, v := range chain.multiView.GetAllViewsWithBFS() {
		res = append(res, *v.GetHash())
	}
	return
}

func (chain *BeaconChain) CloneMultiView() *multiview.MultiView {
	return chain.multiView.Clone()
}

func (chain *BeaconChain) SetMultiView(multiView *multiview.MultiView) {
	chain.multiView = multiView
}

func (chain *BeaconChain) GetDatabase() incdb.Database {
	return chain.Blockchain.GetBeaconChainDatabase()
}

func (chain *BeaconChain) GetBestView() multiview.View {
	return chain.multiView.GetBestView()
}

func (chain *BeaconChain) GetFinalView() multiview.View {
	return chain.multiView.GetFinalView()
}

func (chain *BeaconChain) GetFinalViewState() *BeaconBestState {
	return chain.multiView.GetFinalView().(*BeaconBestState)
}

func (chain *BeaconChain) GetViewByHash(hash common.Hash) multiview.View {
	if chain.multiView.GetViewByHash(hash) == nil {
		return nil
	}
	return chain.multiView.GetViewByHash(hash)
}

func (s *BeaconChain) GetShardBestViewHash() map[byte]common.Hash {
	return s.GetBestView().(*BeaconBestState).GetBestShardHash()
}

func (s *BeaconChain) GetShardBestViewHeight() map[byte]uint64 {
	return s.GetBestView().(*BeaconBestState).GetBestShardHeight()
}

func (s *BeaconChain) GetCurrentCrossShardHeightToShard(sid byte) map[byte]uint64 {

	res := make(map[byte]uint64)
	for fromShard, toShardStatus := range s.GetBestView().(*BeaconBestState).LastCrossShardState {
		for toShard, currentHeight := range toShardStatus {
			if toShard == sid {
				res[fromShard] = currentHeight
			}
		}
	}
	return res
}

func (s *BeaconChain) GetEpoch() uint64 {
	return s.GetBestView().(*BeaconBestState).Epoch
}

func (s *BeaconChain) GetBestViewHeight() uint64 {
	return s.GetBestView().(*BeaconBestState).BeaconHeight
}

func (s *BeaconChain) GetFinalViewHeight() uint64 {
	return s.GetFinalView().(*BeaconBestState).BeaconHeight
}

func (s *BeaconChain) GetBestViewHash() string {
	return s.GetBestView().(*BeaconBestState).BestBlockHash.String()
}

func (s *BeaconChain) GetFinalViewHash() string {
	return s.GetBestView().(*BeaconBestState).BestBlockHash.String()
}

func (chain *BeaconChain) GetLastBlockTimeStamp() int64 {
	return chain.multiView.GetBestView().(*BeaconBestState).BestBlock.Header.Timestamp
}

func (chain *BeaconChain) GetMinBlkInterval() time.Duration {
	return chain.multiView.GetBestView().(*BeaconBestState).BlockInterval
}

func (chain *BeaconChain) GetMaxBlkCreateTime() time.Duration {
	return chain.multiView.GetBestView().(*BeaconBestState).BlockMaxCreateTime
}

func (chain *BeaconChain) IsReady() bool {
	return chain.Ready
}

func (chain *BeaconChain) SetReady(ready bool) {
	chain.Ready = ready
}

func (chain *BeaconChain) CurrentHeight() uint64 {
	return chain.multiView.GetBestView().(*BeaconBestState).BestBlock.Header.Height
}

func (chain *BeaconChain) GetCommittee() []incognitokey.CommitteePublicKey {
	return chain.multiView.GetBestView().(*BeaconBestState).GetBeaconCommittee()
}

func (chain *BeaconChain) GetLastCommittee() []incognitokey.CommitteePublicKey {
	v := chain.multiView.GetViewByHash(*chain.GetBestView().GetPreviousHash())
	if v == nil {
		return nil
	}
	result := []incognitokey.CommitteePublicKey{}
	return append(result, v.GetCommittee()...)
}

func (chain *BeaconChain) GetCommitteeByHeight(h uint64) ([]incognitokey.CommitteePublicKey, error) {
	bcStateRootHash := chain.GetBestView().(*BeaconBestState).ConsensusStateDBRootHash
	bcDB := chain.Blockchain.GetBeaconChainDatabase()
	bcStateDB, err := statedb.NewWithPrefixTrie(bcStateRootHash, statedb.NewDatabaseAccessWarper(bcDB))
	if err != nil {
		return nil, err
	}
	return statedb.GetBeaconCommittee(bcStateDB), nil
}

func (chain *BeaconChain) GetPendingCommittee() []incognitokey.CommitteePublicKey {
	return chain.GetBestView().(*BeaconBestState).GetBeaconPendingValidator()
}

func (chain *BeaconChain) GetCommitteeSize() int {
	return len(chain.multiView.GetBestView().(*BeaconBestState).GetBeaconCommittee())
}

func (chain *BeaconChain) GetPubKeyCommitteeIndex(pubkey string) int {
	for index, key := range chain.multiView.GetBestView().(*BeaconBestState).GetBeaconCommittee() {
		if key.GetMiningKeyBase58(chain.multiView.GetBestView().(*BeaconBestState).ConsensusAlgorithm) == pubkey {
			return index
		}
	}
	return -1
}

func (chain *BeaconChain) GetLastProposerIndex() int {
	return chain.multiView.GetBestView().(*BeaconBestState).BeaconProposerIndex
}

func (chain *BeaconChain) CreateNewBlock(
	buildView multiview.View,
	version int, proposer string,
	round int, startTime int64,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	if buildView == nil {
		buildView = chain.GetBestView()
	}
	newBlock, err := chain.Blockchain.NewBlockBeacon(buildView.(*BeaconBestState), version, proposer, round, startTime)
	if err != nil {
		return nil, err
	}
	if version >= 2 {
		newBlock.Header.Proposer = proposer
		newBlock.Header.ProposeTime = startTime
	}

	return newBlock, nil
}

//this function for version 2
func (chain *BeaconChain) CreateNewBlockFromOldBlock(
	oldBlock types.BlockInterface, proposer string,
	startTime int64,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	b, _ := json.Marshal(oldBlock)
	newBlock := new(types.BeaconBlock)
	json.Unmarshal(b, &newBlock)
	newBlock.Header.Proposer = proposer
	newBlock.Header.ProposeTime = startTime
	return newBlock, nil
}

// TODO: change name
func (chain *BeaconChain) InsertBlock(block types.BlockInterface, validationMode int) error {
	if err := chain.Blockchain.InsertBeaconBlock(block.(*types.BeaconBlock), validationMode); err != nil {
		if err.Error() != "View already exists" {
			Logger.log.Error(err)
		}
		return err
	}
	return nil
}

func (chain *BeaconChain) CheckExistedBlk(block types.BlockInterface) bool {
	blkHash := block.Hash()
	_, err := rawdbv2.GetBeaconBlockByHash(chain.Blockchain.GetBeaconChainDatabase(), *blkHash)
	return err == nil
}

func (chain *BeaconChain) InsertAndBroadcastBlock(block types.BlockInterface) error {
	go chain.Blockchain.config.Server.PushBlockToAll(block, "", true)
	if err := chain.Blockchain.InsertBeaconBlock(block.(*types.BeaconBlock), common.BASIC_VALIDATION); err != nil {
		Logger.log.Info(err)
		return err
	}
	return nil

}

func (chain *BeaconChain) ReplacePreviousValidationData(previousBlockHash common.Hash, newValidationData string) error {
	panic("this function is not supported on beacon chain")
}

func (chain *BeaconChain) InsertAndBroadcastBlockWithPrevValidationData(types.BlockInterface, string) error {
	panic("this function is not supported on beacon chain")
}

func (chain *BeaconChain) GetActiveShardNumber() int {
	return chain.multiView.GetBestView().(*BeaconBestState).ActiveShards
}

func (chain *BeaconChain) GetChainName() string {
	return chain.ChainName
}

func (chain *BeaconChain) ValidatePreSignBlock(block types.BlockInterface, committees []incognitokey.CommitteePublicKey) error {
	return chain.Blockchain.VerifyPreSignBeaconBlock(block.(*types.BeaconBlock), true)
}

// func (chain *BeaconChain) ValidateAndInsertBlock(block common.BlockInterface) error {
// 	var beaconBestState BeaconBestState
// 	beaconBlock := block.(*BeaconBlock)
// 	beaconBestState.cloneBeaconBestStateFrom(chain.multiView.GetBestView().(*BeaconBestState))
// 	producerPublicKey := beaconBlock.Header.Producer
// 	producerPosition := (beaconBestState.BeaconProposerIndex + beaconBlock.Header.Round) % len(beaconBestState.BeaconCommittee)
// 	tempProducer := beaconBestState.BeaconCommittee[producerPosition].GetMiningKeyBase58(beaconBestState.ConsensusAlgorithm)
// 	if strings.Compare(tempProducer, producerPublicKey) != 0 {
// 		return NewBlockChainError(BeaconBlockProducerError, fmt.Errorf("Expect Producer Public Key to be equal but get %+v From Index, %+v From Header", tempProducer, producerPublicKey))
// 	}
// 	if err := chain.ValidateBlockSignatures(block, beaconBestState.BeaconCommittee); err != nil {
// 		return err
// 	}
// 	return chain.Blockchain.InsertBeaconBlock(beaconBlock, false)
// }

func (chain *BeaconChain) ValidateBlockSignatures(block types.BlockInterface, committee []incognitokey.CommitteePublicKey) error {

	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}

	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committee); err != nil {
		return err
	}
	return nil
}

func (chain *BeaconChain) GetConsensusType() string {
	return chain.multiView.GetBestView().(*BeaconBestState).ConsensusAlgorithm
}

func (chain *BeaconChain) GetShardID() int {
	return -1
}

func (chain *BeaconChain) IsBeaconChain() bool {
	return true
}

func (chain *BeaconChain) GetAllCommittees() map[string]map[string][]incognitokey.CommitteePublicKey {
	var result map[string]map[string][]incognitokey.CommitteePublicKey
	result = make(map[string]map[string][]incognitokey.CommitteePublicKey)
	result[chain.multiView.GetBestView().(*BeaconBestState).ConsensusAlgorithm] = make(map[string][]incognitokey.CommitteePublicKey)
	result[chain.multiView.GetBestView().(*BeaconBestState).ConsensusAlgorithm][common.BeaconChainKey] = append([]incognitokey.CommitteePublicKey{}, chain.multiView.GetBestView().(*BeaconBestState).GetBeaconCommittee()...)
	for shardID, consensusType := range chain.multiView.GetBestView().(*BeaconBestState).GetShardConsensusAlgorithm() {
		if _, ok := result[consensusType]; !ok {
			result[consensusType] = make(map[string][]incognitokey.CommitteePublicKey)
		}
		result[consensusType][common.GetShardChainKey(shardID)] = append([]incognitokey.CommitteePublicKey{}, chain.multiView.GetBestView().(*BeaconBestState).GetAShardCommittee(shardID)...)
	}
	return result
}

func (chain *BeaconChain) GetBeaconPendingList() []incognitokey.CommitteePublicKey {
	var result []incognitokey.CommitteePublicKey
	result = append(result, chain.multiView.GetBestView().(*BeaconBestState).GetBeaconPendingValidator()...)
	return result
}

func (chain *BeaconChain) GetShardsCommitteeList() map[int][]incognitokey.CommitteePublicKey {
	result := make(map[int][]incognitokey.CommitteePublicKey)
	for shardID := 0; shardID < chain.GetActiveShardNumber(); shardID++ {
		result[shardID] = append([]incognitokey.CommitteePublicKey{}, chain.multiView.GetBestView().(*BeaconBestState).GetAShardCommittee(byte(shardID))...)
	}
	return result
}

func (chain *BeaconChain) GetShardsPendingList() map[int][]incognitokey.CommitteePublicKey {
	result := make(map[int][]incognitokey.CommitteePublicKey)
	for shardID := 0; shardID < chain.GetActiveShardNumber(); shardID++ {
		result[shardID] = append([]incognitokey.CommitteePublicKey{}, chain.multiView.GetBestView().(*BeaconBestState).GetAShardPendingValidator(byte(shardID))...)
	}
	return result
}

func (chain *BeaconChain) GetShardsWaitingList() []incognitokey.CommitteePublicKey {
	var result []incognitokey.CommitteePublicKey
	result = append(result, chain.multiView.GetBestView().(*BeaconBestState).GetCandidateShardWaitingForNextRandom()...)
	result = append(result, chain.multiView.GetBestView().(*BeaconBestState).GetCandidateShardWaitingForCurrentRandom()...)
	return result
}

func (chain *BeaconChain) GetBeaconWaitingList() []incognitokey.CommitteePublicKey {
	var result []incognitokey.CommitteePublicKey
	result = append(result, chain.multiView.GetBestView().(*BeaconBestState).GetCandidateBeaconWaitingForNextRandom()...)
	result = append(result, chain.multiView.GetBestView().(*BeaconBestState).GetCandidateBeaconWaitingForCurrentRandom()...)
	return result
}

func (chain *BeaconChain) UnmarshalBlock(blockString []byte) (types.BlockInterface, error) {
	var beaconBlk types.BeaconBlock
	err := json.Unmarshal(blockString, &beaconBlk)
	if err != nil {
		return nil, err
	}
	return &beaconBlk, nil
}

func (chain *BeaconChain) GetAllView() []multiview.View {
	return chain.multiView.GetAllViewsWithBFS()
}

//CommitteesByShardID ...
func (chain *BeaconChain) CommitteesFromViewHashForShard(hash common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error) {
	var committees []incognitokey.CommitteePublicKey
	var err error
	res, has := chain.committeeCache.Get(getCommitteeCacheKey(hash, shardID))
	if !has {
		committees, err = chain.Blockchain.GetShardCommitteeFromBeaconHash(hash, shardID)
		if err != nil {
			return committees, err
		}
		chain.committeeCache.Add(getCommitteeCacheKey(hash, shardID), committees)
	} else {
		committees = res.([]incognitokey.CommitteePublicKey)
	}
	return committees, nil
}

func getCommitteeCacheKey(hash common.Hash, shardID byte) string {
	return fmt.Sprintf("%s-%d", hash.String(), shardID)
}

//ProposerByTimeSlot ...
func (chain *BeaconChain) ProposerByTimeSlot(
	shardID byte, ts int64,
	committees []incognitokey.CommitteePublicKey) incognitokey.CommitteePublicKey {

	//TODO: add recalculate proposer index here when swap committees
	// chainParamEpoch := chain.Blockchain.config.ChainParams.Epoch
	// id := -1
	// if ok, err := finalView.HasSwappedCommittee(shardID, chainParamEpoch); err == nil && ok {
	// 	id = 0
	// } else {
	// 	id = GetProposerByTimeSlot(ts, finalView.MinShardCommitteeSize)
	// }

	id := GetProposerByTimeSlot(ts, chain.Blockchain.GetBestStateShard(shardID).MinShardCommitteeSize)
	return committees[id]
}

func (chain *BeaconChain) GetCommitteeV2(block types.BlockInterface) ([]incognitokey.CommitteePublicKey, error) {
	return chain.multiView.GetBestView().(*BeaconBestState).GetBeaconCommittee(), nil
}

func (chain *BeaconChain) CommitteeStateVersion() uint {
	return chain.GetBestView().(*BeaconBestState).beaconCommitteeEngine.Version()
}

func (chain *BeaconChain) FinalView() multiview.View {
	return chain.GetFinalView()
}

//BestViewCommitteeFromBlock ...
func (chain *BeaconChain) BestViewCommitteeFromBlock() common.Hash {
	return common.Hash{}
}

func (chain *BeaconChain) GetChainDatabase() incdb.Database {
	return chain.Blockchain.GetBeaconChainDatabase()
}

func (chain *BeaconChain) CommitteeEngineVersion() uint {
	return chain.multiView.GetBestView().CommitteeEngineVersion()
}
