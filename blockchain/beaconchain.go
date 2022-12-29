package blockchain

import (
	"encoding/json"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/incognitochain/incognito-chain/consensus_v2/consensustypes"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"

	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
	"github.com/incognitochain/incognito-chain/portal/portalv4"
)

type BeaconChain struct {
	multiView multiview.MultiView

	BlockGen            *BlockGenerator
	Blockchain          *BlockChain
	BlockStorage        *BlockStorage
	hashHistory         *lru.Cache
	ChainName           string
	Ready               bool //when has peerstate
	committeesInfoCache *lru.Cache
	insertLock          sync.Mutex
}

func NewBeaconChain(multiView multiview.MultiView, blockGen *BlockGenerator, blockchain *BlockChain, chainName string) *BeaconChain {
	committeeInfoCache, _ := lru.New(100)
	cfg := config.Config()
	ffPath := path.Join(cfg.DataDir, cfg.DatabaseDir, "beacon", "blockstorage")
	bs := NewBlockStorage(blockchain.GetBeaconChainDatabase(), ffPath, -1, false)
	chain := &BeaconChain{
		multiView:           multiView,
		BlockGen:            blockGen,
		BlockStorage:        bs,
		Blockchain:          blockchain,
		ChainName:           chainName,
		committeesInfoCache: committeeInfoCache,
	}
	return chain
}

func (chain *BeaconChain) GetAllViewHash() (res []common.Hash) {
	for _, v := range chain.multiView.GetAllViewsWithBFS() {
		res = append(res, *v.GetHash())
	}
	return
}

func (chain *BeaconChain) GetDatabase() incdb.Database {
	return chain.Blockchain.GetBeaconChainDatabase()
}

func (chain *BeaconChain) GetMultiView() multiview.MultiView {
	return chain.multiView
}

func (chain *BeaconChain) CloneMultiView() multiview.MultiView {
	return chain.multiView.Clone()
}

func (chain *BeaconChain) SetMultiView(multiView multiview.MultiView) {
	chain.multiView = multiView
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
	return s.GetFinalView().(*BeaconBestState).BestBlockHash.String()
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

func (chain *BeaconChain) GetPendingCommittee() []incognitokey.CommitteePublicKey {
	return chain.GetBestView().(*BeaconBestState).GetBeaconPendingValidator()
}

func (chain *BeaconChain) GetWaitingCommittee() []incognitokey.CommitteePublicKey {
	return chain.GetBestView().(*BeaconBestState).GetBeaconWaiting()
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

// this is call when create new block
func (chain *BeaconChain) ReplacePreviousValidationData(previousBlockHash common.Hash, previousProposeHash common.Hash, previousCommittees []incognitokey.CommitteePublicKey, newValidationData string) error {
	if hasBlock := chain.BlockStorage.IsExisted(previousBlockHash); !hasBlock {
		// This block is not inserted yet, no need to replace
		Logger.log.Errorf("Replace previous validation data fail! Cannot find find block in db " + previousBlockHash.String())
		return nil
	}

	blk, err := chain.GetBlockByHash(previousBlockHash)
	if err != nil {
		return NewBlockChainError(ReplacePreviousValidationDataError, err)
	}
	beaconBlock := blk.(*types.BeaconBlock)

	if !previousProposeHash.IsEqual(beaconBlock.ProposeHash()) {
		Logger.log.Errorf("Replace previous validation data fail! Propose hash not correct, data for" + previousProposeHash.String() + " got " + beaconBlock.ProposeHash().String())
		return nil
	}

	decodedOldValidationData, err := consensustypes.DecodeValidationData(beaconBlock.ValidationData)
	if err != nil {
		return NewBlockChainError(ReplacePreviousValidationDataError, err)
	}

	decodedNewValidationData, err := consensustypes.DecodeValidationData(newValidationData)
	if err != nil {
		return NewBlockChainError(ReplacePreviousValidationDataError, err)
	}

	if len(decodedNewValidationData.ValidatiorsIdx) > len(decodedOldValidationData.ValidatiorsIdx) {
		Logger.log.Infof("Beacon | Height %+v, Replace Previous ValidationData new number of signatures %+v (old %+v)",
			beaconBlock.Header.Height, len(decodedNewValidationData.ValidatiorsIdx), len(decodedOldValidationData.ValidatiorsIdx))
	} else {
		return nil
	}

	// validate block before rewrite to
	replaceBlockHash := *beaconBlock.Hash()
	beaconBlock.ValidationData = newValidationData

	if err != nil {
		return err
	}
	if err = chain.ValidateBlockSignatures(beaconBlock, previousCommittees, chain.GetBestView().GetProposerLength()); err != nil {
		return err
	}
	//rewrite to database
	if err = chain.BlockStorage.ReplaceBlock(beaconBlock); err != nil {
		return err
	}
	//update multiview
	view := chain.multiView.GetViewByHash(replaceBlockHash)
	if view != nil {
		view.ReplaceBlock(beaconBlock)
	} else {
		fmt.Println("Cannot find beacon view", replaceBlockHash.String())
	}
	return nil
}

func (chain *BeaconChain) CreateNewBlock(
	version int, proposer string, round int, startTime int64,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	//wait a little bit, for shard
	beaconBestView := chain.GetBestView().(*BeaconBestState)
	if version < types.ADJUST_BLOCKTIME_VERSION {
		time.Sleep(time.Duration(beaconBestView.GetCurrentTimeSlot()/5) * time.Second)
	}

	newBlock, err := chain.Blockchain.NewBlockBeacon(beaconBestView, version, proposer, round, startTime)
	if err != nil {
		return nil, err
	}
	if version >= 2 {
		newBlock.Header.Proposer = proposer
		newBlock.Header.ProposeTime = startTime
	}
	newBlock.Header.FinalityHeight = 0
	if version >= types.LEMMA2_VERSION {
		previousBlock, err := chain.GetBlockByHash(newBlock.Header.PreviousBlockHash)
		if err != nil {
			return nil, err
		}
		previousProposeTimeSlot := beaconBestView.CalculateTimeSlot(previousBlock.GetProposeTime())
		previousProduceTimeSlot := beaconBestView.CalculateTimeSlot(previousBlock.GetProduceTime())
		currentTimeSlot := beaconBestView.CalculateTimeSlot(newBlock.Header.Timestamp)

		// if previous block is finality or same produce/propose
		// and  block produced/proposed next block time
		if newBlock.Header.Timestamp == newBlock.Header.ProposeTime && newBlock.Header.Producer == newBlock.Header.Proposer && previousProposeTimeSlot+1 == currentTimeSlot {
			if version >= types.INSTANT_FINALITY_VERSION {
				if previousBlock.GetFinalityHeight() != 0 || previousProduceTimeSlot == previousProposeTimeSlot {
					newBlock.Header.FinalityHeight = newBlock.Header.Height
				}
			} else {
				newBlock.Header.FinalityHeight = newBlock.Header.Height - 1
			}
		}
	}

	return newBlock, nil
}

// this function for version 2
func (chain *BeaconChain) CreateNewBlockFromOldBlock(oldBlock types.BlockInterface, proposer string, startTime int64, isValidRePropose bool) (types.BlockInterface, error) {
	b, _ := json.Marshal(oldBlock)
	newBlock := new(types.BeaconBlock)
	json.Unmarshal(b, &newBlock)
	newBlock.Header.Proposer = proposer
	newBlock.Header.ProposeTime = startTime
	version := newBlock.Header.Version
	newBlock.Header.FinalityHeight = 0
	if version >= types.LEMMA2_VERSION {
		previousBlock, err := chain.GetBlockByHash(newBlock.Header.PreviousBlockHash)
		if err != nil {
			return nil, err
		}

		// if previous block is finality or same produce/propose
		// and valid lemma2
		if isValidRePropose {
			bestView := chain.GetBestView()
			previousProposeTimeSlot := bestView.CalculateTimeSlot(previousBlock.GetProposeTime())
			previousProduceTimeSlot := bestView.CalculateTimeSlot(previousBlock.GetProduceTime())
			if version >= types.INSTANT_FINALITY_VERSION {
				if previousBlock.GetFinalityHeight() != 0 || previousProduceTimeSlot == previousProposeTimeSlot {
					newBlock.Header.FinalityHeight = newBlock.Header.Height
				}
			} else {
				newBlock.Header.FinalityHeight = newBlock.Header.Height - 1
			}
		}
	}
	return newBlock, nil
}

// TODO: change name
func (chain *BeaconChain) InsertBlock(block types.BlockInterface, shouldValidate bool) error {
	if err := chain.Blockchain.InsertBeaconBlock(block.(*types.BeaconBlock), shouldValidate); err != nil {
		Logger.log.Error(err)
		return err
	}
	return nil
}

func (chain *BeaconChain) CheckExistedBlk(block types.BlockInterface) bool {
	blkHash := block.Hash()
	return chain.BlockStorage.IsExisted(*blkHash)
}

func (chain *BeaconChain) InsertAndBroadcastBlock(block types.BlockInterface) error {
	go chain.Blockchain.config.Server.PushBlockToAll(block, "", true)
	if err := chain.Blockchain.InsertBeaconBlock(block.(*types.BeaconBlock), true); err != nil {
		Logger.log.Error(err)
		return err
	}
	return nil

}

// this get consensus data for all latest shard state
func (chain *BeaconChain) GetBlockConsensusData() map[int]types.BlockConsensusData {
	consensusData := map[int]types.BlockConsensusData{}
	bestViewBlock := chain.multiView.GetBestView().GetBlock().(*types.BeaconBlock)
	consensusData[-1] = types.BlockConsensusData{
		BlockHash:      *bestViewBlock.Hash(),
		BlockHeight:    bestViewBlock.GetHeight(),
		FinalityHeight: bestViewBlock.GetFinalityHeight(),
		Proposer:       bestViewBlock.GetProposer(),
		ProposerTime:   bestViewBlock.GetProposeTime(),
		ValidationData: bestViewBlock.GetValidationField(),
	}

	for sid, sChain := range chain.Blockchain.ShardChain {
		shardBlk := sChain.multiView.GetExpectedFinalView().GetBlock().(*types.ShardBlock)
		consensusData[int(sid)] = types.BlockConsensusData{
			BlockHash:      *shardBlk.Hash(),
			BlockHeight:    shardBlk.GetHeight(),
			FinalityHeight: shardBlk.GetFinalityHeight(),
			Proposer:       shardBlk.GetProposer(),
			ProposerTime:   shardBlk.GetProposeTime(),
			ValidationData: shardBlk.ValidationData,
		}
	}
	return consensusData
}

func (chain *BeaconChain) InsertAndBroadcastBlockWithPrevValidationData(block types.BlockInterface, s string) error {
	go chain.Blockchain.config.Server.PushBlockToAll(block, "", true)

	return chain.InsertBlock(block, true)
}
func (chain *BeaconChain) InsertWithPrevValidationData(types.BlockInterface, string) error {
	panic("this function is not supported on beacon chain")
}

func (chain *BeaconChain) CollectTxs(view multiview.View) {
	return
}

func (chain *BeaconChain) GetBlockByHash(hash common.Hash) (types.BlockInterface, error) {
	block, _, err := chain.Blockchain.GetBeaconBlockByHash(hash)
	return block, err
}

func (chain *BeaconChain) GetActiveShardNumber() int {
	return chain.multiView.GetBestView().(*BeaconBestState).ActiveShards
}

func (chain *BeaconChain) GetChainName() string {
	return chain.ChainName
}

func (chain *BeaconChain) ValidatePreSignBlock(block types.BlockInterface, signingCommittees, committees []incognitokey.CommitteePublicKey) error {
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

func (chain *BeaconChain) ValidateBlockSignatures(block types.BlockInterface, committees []incognitokey.CommitteePublicKey, numOfFixNode int) error {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		Logger.log.Info("[dcs] err:", err)
		return err
	}
	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committees, numOfFixNode); err != nil {
		Logger.log.Info("[dcs] err:", err)
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

func (chain *BeaconChain) GetShardsPendingList() map[string]map[string][]incognitokey.CommitteePublicKey {
	var result map[string]map[string][]incognitokey.CommitteePublicKey
	result = make(map[string]map[string][]incognitokey.CommitteePublicKey)
	for shardID, consensusType := range chain.multiView.GetBestView().(*BeaconBestState).GetShardConsensusAlgorithm() {
		if _, ok := result[consensusType]; !ok {
			result[consensusType] = make(map[string][]incognitokey.CommitteePublicKey)
		}
		result[consensusType][common.GetShardChainKey(shardID)] = append([]incognitokey.CommitteePublicKey{}, chain.multiView.GetBestView().(*BeaconBestState).GetAShardPendingValidator(shardID)...)
	}
	return result
}

func (chain *BeaconChain) GetSyncingValidators() map[byte][]incognitokey.CommitteePublicKey {
	return chain.GetBestView().(*BeaconBestState).GetSyncingValidators()
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

func (chain *BeaconChain) GetProposerByTimeSlotFromCommitteeList(ts int64, committees []incognitokey.CommitteePublicKey) (incognitokey.CommitteePublicKey, int) {
	id := GetProposerByTimeSlot(ts, chain.GetBestView().(*BeaconBestState).MinBeaconCommitteeSize)
	return committees[id], id
}

func (chain *BeaconChain) GetPortalParamsV4(beaconHeight uint64) portalv4.PortalParams {
	return chain.Blockchain.GetPortalParamsV4(beaconHeight)
}

// CommitteesByShardID ...
var CommitteeFromBlockCache, _ = lru.New(500)

func (chain *BeaconChain) CommitteesFromViewHashForShard(hash common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error) {
	committees := []incognitokey.CommitteePublicKey{}
	var err error

	cache := CommitteeFromBlockCache
	cacheKey := fmt.Sprintf("%v-%v", shardID, hash.String())
	tempCommittees, ok := cache.Get(cacheKey)
	if ok {
		committees = tempCommittees.([]incognitokey.CommitteePublicKey)
		return committees, nil
	}

	committees, err = rawdbv2.GetCacheCommitteeFromBlock(chain.BlockStorage.blockStorageDB, hash, int(shardID))
	if len(committees) > 0 {
		cache.Add(cacheKey, committees)
		return committees, nil
	}

	committees, err = chain.Blockchain.GetShardCommitteeFromBeaconHash(hash, shardID)
	if len(committees) > 0 {
		cache.Add(cacheKey, committees)
		return committees, err
	}
	return committees, fmt.Errorf("Cannot find committee from shardID %v viewHash %v", shardID, hash.String())
}

func (chain *BeaconChain) GetSigningCommittees(
	proposerIndex int,
	committees []incognitokey.CommitteePublicKey,
	blockVersion int,
) []incognitokey.CommitteePublicKey {
	return append([]incognitokey.CommitteePublicKey{}, committees...)
}

func (chain *BeaconChain) GetCommitteeV2(block types.BlockInterface) ([]incognitokey.CommitteePublicKey, error) {
	if config.Param().FeatureVersion[BEACON_STAKING_FLOW_V4] != 0 && block.GetVersion() >= int(config.Param().FeatureVersion[BEACON_STAKING_FLOW_V4]) {
		height := block.GetHeight()
		beaconConsensusStateRootHash, err := chain.Blockchain.GetBeaconRootsHashFromBlockHeight(
			height - 1, //the previous height statedb store the committee of current height
		)
		if err != nil {
			return nil, err
		}
		stateDB, err := statedb.NewWithPrefixTrie(beaconConsensusStateRootHash.ConsensusStateDBRootHash,
			statedb.NewDatabaseAccessWarper(chain.GetDatabase()))
		if err != nil {
			return nil, err
		}
		//Review Just get committee, no need to restore
		return statedb.GetBeaconCommittee(stateDB), nil
		// stateV4 := committeestate.NewBeaconCommitteeStateV4()
		// err = stateV4.RestoreBeaconCommitteeFromDB(stateDB, chain.GetBestView().(*BeaconBestState).MinBeaconCommitteeSize, nil)
		// if err != nil {
		// 	return nil, err
		// }
		// return stateV4.GetBeaconCommittee(), nil
	} else {
		committees := chain.multiView.GetBestView().(*BeaconBestState).GetBeaconCommittee()
		return committees, nil
	}

}

func (chain *BeaconChain) GetCommitteeByHash(blockHash common.Hash, blockHeight uint64) ([]incognitokey.CommitteePublicKey, error) {
	viewByBlock := chain.multiView.GetViewByHash(blockHash)
	if viewByBlock != nil {
		return viewByBlock.GetCommittee(), nil
	}
	bcRootHash, err := chain.Blockchain.GetBeaconConsensusRootHash(chain.GetFinalViewState(), blockHeight)
	if err != nil {
		return nil, err
	}
	beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(bcRootHash, statedb.NewDatabaseAccessWarper(chain.GetChainDatabase()))
	if err != nil {
		Logger.log.Error("Cannot get beacon consensus statedb!,", err.Error())
		return nil, err
	}
	return statedb.GetBeaconCommittee(beaconConsensusStateDB), nil
}

func (chain *BeaconChain) CommitteeStateVersion() int {
	return chain.GetBestView().(*BeaconBestState).beaconCommitteeState.Version()
}

func (chain *BeaconChain) FinalView() multiview.View {
	return chain.GetFinalView()
}

// BestViewCommitteeFromBlock ...
func (chain *BeaconChain) BestViewCommitteeFromBlock() common.Hash {
	return common.Hash{}
}

func (chain *BeaconChain) GetChainDatabase() incdb.Database {
	return chain.Blockchain.GetBeaconChainDatabase()
}

func (chain *BeaconChain) CommitteeEngineVersion() int {
	return chain.multiView.GetBestView().CommitteeStateVersion()
}

func getCommitteeCacheKey(hash common.Hash, shardID byte) string {
	return fmt.Sprintf("%s-%d", hash.String(), shardID)
}

func (chain *BeaconChain) StoreFinalityProof(block types.BlockInterface, finalityProof interface{}, reProposeSig interface{}) error {
	return nil
}
