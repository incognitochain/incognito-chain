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
	"github.com/incognitochain/incognito-chain/portal/portalv4"
)

type BeaconChain struct {
	multiView multiview.MultiView

	BlockGen            *BlockGenerator
	Blockchain          *BlockChain
	hashHistory         *lru.Cache
	ChainName           string
	Ready               bool //when has peerstate
	committeesInfoCache *lru.Cache

	insertLock sync.Mutex
}

func NewBeaconChain(multiView multiview.MultiView, blockGen *BlockGenerator, blockchain *BlockChain, chainName string) *BeaconChain {
	committeeInfoCache, _ := lru.New(100)
	return &BeaconChain{
		multiView:           multiView,
		BlockGen:            blockGen,
		Blockchain:          blockchain,
		ChainName:           chainName,
		committeesInfoCache: committeeInfoCache,
	}
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
	version int, proposer string, round int, startTime int64,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	//wait a little bit, for shard
	waitTime := common.TIMESLOT / 5
	time.Sleep(time.Duration(waitTime) * time.Second)

	newBlock, err := chain.Blockchain.NewBlockBeacon(chain.GetBestView().(*BeaconBestState), version, proposer, round, startTime)
	if err != nil {
		return nil, err
	}
	if version >= 2 {
		newBlock.Header.Proposer = proposer
		newBlock.Header.ProposeTime = startTime
	}
	if version >= types.LEMMA2_VERSION {
		previousBlock, err := chain.GetBlockByHash(newBlock.Header.PreviousBlockHash)
		if err != nil {
			return nil, err
		}
		previousProposeTimeSlot := common.CalculateTimeSlot(previousBlock.GetProposeTime())
		currentTimeSlot := common.CalculateTimeSlot(newBlock.Header.Timestamp)

		if newBlock.Header.Timestamp == newBlock.Header.ProposeTime &&
			newBlock.Header.Producer == newBlock.Header.Proposer &&
			previousProposeTimeSlot+1 == currentTimeSlot {
			newBlock.Header.FinalityHeight = newBlock.Header.Height - 1
		} else {
			newBlock.Header.FinalityHeight = 0
		}
	}

	return newBlock, nil
}

//this function for version 2
func (chain *BeaconChain) CreateNewBlockFromOldBlock(oldBlock types.BlockInterface, proposer string, startTime int64, isValidRePropose bool) (types.BlockInterface, error) {
	b, _ := json.Marshal(oldBlock)
	newBlock := new(types.BeaconBlock)
	json.Unmarshal(b, &newBlock)
	newBlock.Header.Proposer = proposer
	newBlock.Header.ProposeTime = startTime

	if newBlock.Header.Version >= types.LEMMA2_VERSION {
		if isValidRePropose {
			newBlock.Header.FinalityHeight = newBlock.Header.Height - 1
		} else {
			newBlock.Header.FinalityHeight = 0
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
	_, err := rawdbv2.GetBeaconBlockByHash(chain.Blockchain.GetBeaconChainDatabase(), *blkHash)
	return err == nil
}

func (chain *BeaconChain) InsertAndBroadcastBlock(block types.BlockInterface) error {
	go chain.Blockchain.config.Server.PushBlockToAll(block, "", true)
	if err := chain.Blockchain.InsertBeaconBlock(block.(*types.BeaconBlock), true); err != nil {
		Logger.log.Error(err)
		return err
	}
	return nil

}

//this get consensus data for all latest shard state
func (chain *BeaconChain) GetBlockConsensusData(blk types.BlockInterface) map[int]types.BlockConsensusData {
	consensusData := map[int]types.BlockConsensusData{}
	for sid, sChain := range chain.Blockchain.ShardChain {
		rawBlk, err := rawdbv2.GetShardBlockByHash(chain.Blockchain.GetShardChainDatabase(byte(sid)), *sChain.multiView.GetExpectedFinalView().GetHash())
		if err != nil {
			panic(err)
		}
		shardBlk := new(types.ShardBlock)
		err = json.Unmarshal(rawBlk, shardBlk)
		if err != nil {
			panic(err)
		}
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

//for replace signature index and finality height
func (chain *BeaconChain) ReplaceBlockConsensusData(consensusData types.BlockConsensusData) error {
	replaceBlockHash := consensusData.BlockHash
	//retrieve block from database and replace consensus field
	rawBlk, _ := rawdbv2.GetBeaconBlockByHash(chain.GetDatabase(), replaceBlockHash)
	beaconBlk := new(types.BeaconBlock)
	_ = json.Unmarshal(rawBlk, beaconBlk)
	beaconBlk.Header.Proposer = consensusData.Proposer
	beaconBlk.Header.ProposeTime = consensusData.ProposerTime
	beaconBlk.Header.FinalityHeight = consensusData.FinalityHeight
	beaconBlk.ValidationData = consensusData.ValidationData

	// validate block before rewrite to
	committees, err := chain.GetCommitteeV2(beaconBlk)
	if err != nil {
		return err
	}
	if err = chain.ValidateBlockSignatures(beaconBlk, committees); err != nil {
		return err
	}
	//rewrite to database
	if err = rawdbv2.StoreBeaconBlockByHash(chain.GetChainDatabase(), replaceBlockHash, beaconBlk); err != nil {
		return err
	}
	//update multiview
	chain.multiView.GetViewByHash(replaceBlockHash).ReplaceBlock(beaconBlk)
	return nil
}

func (chain *BeaconChain) ReplacePreviousValidationData(previousBlockHash common.Hash, newValidationData string) error {
	panic("this function is not supported on beacon chain")
}

func (chain *BeaconChain) InsertAndBroadcastBlockWithPrevValidationData(types.BlockInterface, string) error {
	panic("this function is not supported on beacon chain")
}
func (chain *BeaconChain) InsertWithPrevValidationData(types.BlockInterface, string) error {
	panic("this function is not supported on beacon chain")
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

func (chain *BeaconChain) ValidateBlockSignatures(block types.BlockInterface, committees []incognitokey.CommitteePublicKey) error {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		Logger.log.Info("[dcs] err:", err)
		return err
	}
	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committees); err != nil {
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

//CommitteesByShardID ...
func (chain *BeaconChain) CommitteesFromViewHashForShard(hash common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error) {
	var committees []incognitokey.CommitteePublicKey
	var err error
	res, has := chain.committeesInfoCache.Get(getCommitteeCacheKey(hash, shardID))
	if !has {
		committees, err = chain.Blockchain.GetShardCommitteeFromBeaconHash(hash, shardID)
		if err != nil {
			return committees, err
		}
		chain.committeesInfoCache.Add(getCommitteeCacheKey(hash, shardID), committees)
	} else {
		committees = res.([]incognitokey.CommitteePublicKey)
	}
	return committees, nil
}

func (chain *BeaconChain) GetSigningCommittees(
	proposerIndex int,
	committees []incognitokey.CommitteePublicKey,
	blockVersion int,
) []incognitokey.CommitteePublicKey {
	return append([]incognitokey.CommitteePublicKey{}, committees...)
}

func (chain *BeaconChain) GetCommitteeV2(block types.BlockInterface) ([]incognitokey.CommitteePublicKey, error) {
	committees := chain.multiView.GetBestView().(*BeaconBestState).GetBeaconCommittee()
	return committees, nil
}

func (chain *BeaconChain) CommitteeStateVersion() int {
	return chain.GetBestView().(*BeaconBestState).beaconCommitteeState.Version()
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

func (chain *BeaconChain) CommitteeEngineVersion() int {
	return chain.multiView.GetBestView().CommitteeStateVersion()
}

func getCommitteeCacheKey(hash common.Hash, shardID byte) string {
	return fmt.Sprintf("%s-%d", hash.String(), shardID)
}

func (chain *BeaconChain) StoreFinalityProof(block types.BlockInterface, finalityProof interface{}, reProposeSig interface{}) error {
	return nil
}
