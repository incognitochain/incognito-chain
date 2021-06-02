package blockchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/portal"
	"github.com/incognitochain/incognito-chain/pubsub"
	"reflect"
	"sort"
	"strconv"
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
	for _, v := range chain.multiView.GetAllViewsWithBFS(chain.multiView.GetFinalView()) {
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

type BeaconProducingFlow struct {
	version            int
	producer           string
	round              int
	startTime          int64
	newBlock           *types.BeaconBlock
	copiedCurView      *BeaconBestState
	nextBlockEpoch     uint64
	confirmShardBlocks map[byte][]*types.ShardBlock
}

func (chain *BeaconChain) getDataBeforeBlockProducing(buildView *BeaconBestState,
	version int, producer string,
	round int, startTime int64) (*BeaconProducingFlow, error) {
	blockchain := chain.Blockchain
	createFlow := &BeaconProducingFlow{
		version:   version,
		producer:  producer,
		round:     round,
		startTime: startTime,
	}

	//clone curView
	copiedCurView := NewBeaconBestState()
	if err := copiedCurView.cloneBeaconBestStateFrom(buildView); err != nil {
		return nil, err
	}
	createFlow.copiedCurView = copiedCurView
	createFlow.nextBlockEpoch, _ = blockchain.GetEpochNextHeight(copiedCurView.BeaconHeight)

	//get shard blocks to confirm
	createFlow.confirmShardBlocks = blockchain.GetShardBlockForBeaconProducer(copiedCurView.BestShardHeight)

	return createFlow, nil
}

func (chain *BeaconChain) buildBlock(createFlow *BeaconProducingFlow) error {
	curView := createFlow.copiedCurView
	blockchain := chain.Blockchain
	newBeaconBlock := types.NewBeaconBlock()
	createFlow.newBlock = newBeaconBlock
	Logger.log.Infof("New Beacon Block, height %+v, epoch %+v", curView.BeaconHeight+1, createFlow.nextBlockEpoch)
	newBeaconBlock.Header = types.NewBeaconHeader(
		createFlow.version,
		curView.BeaconHeight+1,
		createFlow.nextBlockEpoch,
		createFlow.round,
		createFlow.startTime,
		createFlow.copiedCurView.BestBlockHash,
		createFlow.copiedCurView.ConsensusAlgorithm,
		createFlow.producer,
		createFlow.producer,
	)

	portalParams := portal.GetPortalParams()
	instructions, shardStates, err := blockchain.GenerateBeaconBlockBody(
		newBeaconBlock,
		curView,
		*portalParams,
		createFlow.confirmShardBlocks,
	)
	if err != nil {
		return NewBlockChainError(GenerateInstructionError, err)
	}
	newBeaconBlock.Body = types.NewBeaconBody(shardStates, instructions)

	// Process new block with new view
	_, hashes, _, incurredInstructions, err := curView.updateBeaconBestState(newBeaconBlock, blockchain)
	if err != nil {
		return err
	}

	instructions = append(instructions, incurredInstructions...)
	newBeaconBlock.Body.SetInstructions(instructions)
	if len(newBeaconBlock.Body.Instructions) != 0 {
		Logger.log.Info("Beacon Produce: Beacon Instruction", newBeaconBlock.Body.Instructions)
	}

	// calculate hash
	tempInstructionArr := []string{}
	for _, strs := range instructions {
		tempInstructionArr = append(tempInstructionArr, strs...)
	}
	instructionHash, err := generateHashFromStringArray(tempInstructionArr)
	if err != nil {
		return NewBlockChainError(GenerateInstructionHashError, err)
	}
	shardStatesHash, err := generateHashFromShardState(shardStates, curView.CommitteeEngineVersion())
	if err != nil {
		return NewBlockChainError(GenerateShardStateError, err)
	}

	// Instruction merkle root
	flattenInsts, err := FlattenAndConvertStringInst(instructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, err)
	}

	// add hash to header
	newBeaconBlock.Header.AddBeaconHeaderHash(
		instructionHash,
		shardStatesHash,
		GetKeccak256MerkleRoot(flattenInsts),
		hashes.BeaconCommitteeAndValidatorHash,
		hashes.BeaconCandidateHash,
		hashes.ShardCandidateHash,
		hashes.ShardCommitteeAndValidatorHash,
		hashes.AutoStakeHash,
	)

	if createFlow.version >= 2 {
		createFlow.newBlock.Header.Proposer = createFlow.producer
		createFlow.newBlock.Header.ProposeTime = createFlow.startTime
	}

	return nil
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
	createFlow, err := chain.getDataBeforeBlockProducing(buildView.(*BeaconBestState), version, proposer, round, startTime)
	if err != nil {
		return nil, err
	}

	if err := chain.buildBlock(createFlow); err != nil {
		return nil, err
	}

	return createFlow.newBlock, nil
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

func (chain *BeaconChain) commitAndStore(flow *BeaconValidationFlow) error {
	if err := flow.nextView.beaconCommitteeEngine.Commit(flow.newBeaconCommitteeHashes, flow.committeeChange); err != nil {
		return err
	}

	if err := chain.Blockchain.processStoreBeaconBlock(flow.copiedCurView, flow.nextView, flow.beaconBlock, flow.committeeChange); err != nil {
		return err
	}
	return nil
}

func (chain *BeaconChain) InsertBlock(block types.BlockInterface, validationMode int) error {
	blockHeight := block.GetHeight()
	blockHash := block.Hash().String()
	//get required object for validation
	Logger.log.Infof("BEACON | Begin insert block height %+v - hash %+v, get required data for validate", blockHeight, blockHash)

	chain.insertLock.Lock()
	defer chain.insertLock.Unlock()

	validationFlow, err := chain.getDataBeforeBlockValidation(block.(*types.BeaconBlock), validationMode, false)
	if err != nil {
		return err
	}

	if err = chain.ValidateAndProcessBlock(validationFlow); err != nil {
		return err
	}

	//store and commit
	Logger.log.Infof("BEACON | Commit and Store block height %+v - hash %+v", blockHeight, blockHash)
	if err = chain.commitAndStore(validationFlow); err != nil {
		return err
	}

	Logger.log.Infof("BEACON | Finish insert block height %+v - hash %+v", blockHeight, blockHash)

	//broadcast after successfully insert
	chain.Blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewShardblockTopic, block))
	chain.Blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardBeststateTopic, validationFlow.nextView))
	return nil
}

func (chain *BeaconChain) CheckExistedBlk(block types.BlockInterface) bool {
	blkHash := block.Hash()
	_, err := rawdbv2.GetBeaconBlockByHash(chain.Blockchain.GetBeaconChainDatabase(), *blkHash)
	return err == nil
}

func (chain *BeaconChain) InsertAndBroadcastBlock(block types.BlockInterface) error {
	go chain.Blockchain.config.Server.PushBlockToAll(block, "", true)
	if err := chain.InsertBlock(block.(*types.BeaconBlock), common.BASIC_VALIDATION); err != nil {
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

type BeaconValidationFlow struct {
	validationMode           int
	forSigning               bool
	beaconBlock              *types.BeaconBlock
	confirmShardBlocks       map[byte][]*types.ShardBlock
	blockCommittees          []incognitokey.CommitteePublicKey
	copiedCurView            *BeaconBestState
	nextView                 *BeaconBestState
	newBeaconCommitteeHashes *committeestate.BeaconCommitteeStateHash
	committeeChange          *committeestate.CommitteeChange
}

func checkValidShardState(curView *BeaconBestState, shardStates map[byte][]types.ShardState, allShardBlock map[byte][]*types.ShardBlock) error {
	lastShardHeight := make(map[byte]uint64)
	lastShardHash := make(map[byte]string)

	for sid, v := range curView.BestShardHeight {
		lastShardHeight[sid] = v
	}
	for sid, v := range curView.BestShardHash {
		lastShardHash[sid] = v.String()
	}

	for sid, shardState := range shardStates {
		listHeight := []uint64{}
		if len(allShardBlock[sid]) != len(shardState) {
			return fmt.Errorf("List Shard Block is not enough, shardID: %v shardBlockList: %v shardStates: %v", sid, len(allShardBlock[sid]), len(shardState))
		}
		for i, blockstate := range shardState {
			if blockstate.Height == 2 {
				continue
			}
			listHeight = append(listHeight, allShardBlock[sid][i].GetHeight())
			//check shard state height valid
			if lastShardHeight[sid] != allShardBlock[sid][i].GetHeight()-1 {
				return fmt.Errorf("Shard Height is not valid! Shard %v, bestShardHeight %v, requiredHeight: %v", sid, curView.BestShardHeight[sid], listHeight)
			}
			//check shard state hash valid
			if lastShardHash[sid] != allShardBlock[sid][i].GetPrevHash().String() {
				return fmt.Errorf("Prev shard Hash is not valid! Shard %v, height %v, hash: %v, prehash: %v", sid, blockstate.Height, blockstate.Hash.String(), lastShardHash[sid])
			}
			if blockstate.Hash.String() != allShardBlock[sid][i].Hash().String() {
				return fmt.Errorf("Shard Hash is not valid! Shard %v, height %v, stateHash: %v, shardHash: %v", sid, blockstate.Height, blockstate.Hash.String(), allShardBlock[sid][i].Hash().String())
			}
			lastShardHeight[sid] = allShardBlock[sid][i].GetHeight()
			lastShardHash[sid] = allShardBlock[sid][i].Hash().String()
		}
	}

	return nil
}

func (chain *BeaconChain) getDataBeforeBlockValidation(beaconBlock *types.BeaconBlock, validationMode int, forSigning bool) (*BeaconValidationFlow, error) {
	validationFlow := &BeaconValidationFlow{
		beaconBlock:    beaconBlock,
		validationMode: validationMode,
		forSigning:     forSigning,
	}

	//get previous block
	preHash := beaconBlock.Header.PreviousBlockHash
	view := chain.Blockchain.BeaconChain.GetViewByHash(preHash)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if view == nil {
		chain.Blockchain.config.Syncker.SyncMissingBeaconBlock(ctx, "", preHash)
		return nil, errors.New(fmt.Sprintf("BeaconBlock %v link to wrong view (%s)", beaconBlock.GetHeight(), preHash.String()))
	}
	curView := view.(*BeaconBestState)
	copiedCurView := NewBeaconBestState()
	err := copiedCurView.cloneBeaconBestStateFrom(curView)
	if err != nil {
		return nil, err
	}

	validationFlow.copiedCurView = copiedCurView
	validationFlow.blockCommittees = copiedCurView.GetBeaconCommittee()

	if validationFlow.forSigning {
		// get shard block to confirm
		allRequiredShardBlockHeight := make(map[byte][]uint64)
		for shardID, shardstates := range beaconBlock.Body.ShardState {
			heights := []uint64{}
			for _, state := range shardstates {
				heights = append(heights, state.Height)
			}
			sort.Slice(heights, func(i, j int) bool {
				return heights[i] < heights[j]
			})
			allRequiredShardBlockHeight[shardID] = heights
		}
		allShardBlocks, err := chain.Blockchain.GetShardBlocksForBeaconValidator(allRequiredShardBlockHeight)
		if err != nil {
			Logger.log.Error(err)
			return nil, NewBlockChainError(GetShardBlocksForBeaconProcessError, fmt.Errorf("Unable to get required shard block for beacon process."))
		}
		if err := checkValidShardState(copiedCurView, beaconBlock.Body.ShardState, allShardBlocks); err != nil {
			return nil, err
		}
		validationFlow.confirmShardBlocks = allShardBlocks
	}

	return validationFlow, nil
}

func (chain *BeaconChain) validateBlockSignaturesWithCurrentView(validationFlow *BeaconValidationFlow) error {
	beaconBlock := validationFlow.beaconBlock
	curView := validationFlow.copiedCurView
	committee := validationFlow.blockCommittees

	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerPosition(beaconBlock,
		curView.BeaconProposerIndex, committee,
		curView.MinBeaconCommitteeSize); err != nil {
		return err
	}

	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(beaconBlock, chain.GetConsensusType()); err != nil {
		return err
	}

	if !validationFlow.forSigning {
		if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(beaconBlock, committee); err != nil {
			return err
		}
	}

	return nil
}

func (chain *BeaconChain) validateBlockHeader(validationFlow *BeaconValidationFlow) error {
	beaconBlock := validationFlow.beaconBlock
	curView := validationFlow.copiedCurView
	blockChain := chain.Blockchain

	if err := blockChain.verifyPreProcessingBeaconBlock(beaconBlock, curView); err != nil {
		return err
	}

	if !curView.BestBlockHash.IsEqual(&beaconBlock.Header.PreviousBlockHash) {
		return NewBlockChainError(BeaconBestStateBestBlockNotCompatibleError, errors.New("previous us block should be :"+curView.BestBlockHash.String()))
	}
	if curView.BeaconHeight+1 != beaconBlock.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, errors.New("block height of new block should be :"+strconv.Itoa(int(beaconBlock.Header.Height+1))))
	}
	if blockChain.IsFirstBeaconHeightInEpoch(curView.BeaconHeight+1) && curView.Epoch+1 != beaconBlock.Header.Epoch {
		return NewBlockChainError(WrongEpochError, fmt.Errorf("Expect beacon block height %+v has epoch %+v but get %+v", beaconBlock.Header.Height, curView.Epoch+1, beaconBlock.Header.Epoch))
	}
	if !blockChain.IsFirstBeaconHeightInEpoch(curView.BeaconHeight+1) && curView.Epoch != beaconBlock.Header.Epoch {
		return NewBlockChainError(WrongEpochError, fmt.Errorf("Expect beacon block height %+v has epoch %+v but get %+v", beaconBlock.Header.Height, curView.Epoch, beaconBlock.Header.Epoch))
	}
	return nil
}

//beaconBlockBody:
// - shardState : already checked during  get required data before validation
// - instruction: rebuild and check if generate same inst hash
func (chain *BeaconChain) validateBlockBody(validationFlow *BeaconValidationFlow, incurInstructions [][]string) error {
	curView := validationFlow.copiedCurView
	beaconBlock := validationFlow.beaconBlock

	//=============Verify Stake Public Key =>
	newBeaconCandidate, newShardCandidate := getStakingCandidate(*beaconBlock)
	if !reflect.DeepEqual(newBeaconCandidate, []string{}) {
		validBeaconCandidate := curView.GetValidStakers(newBeaconCandidate)
		if !reflect.DeepEqual(validBeaconCandidate, newBeaconCandidate) {
			return NewBlockChainError(CandidateError, errors.New("beacon candidate list is INVALID"))
		}
	}
	if !reflect.DeepEqual(newShardCandidate, []string{}) {
		validShardCandidate := curView.GetValidStakers(newShardCandidate)
		if !reflect.DeepEqual(validShardCandidate, newShardCandidate) {
			return NewBlockChainError(CandidateError, errors.New("shard candidate list is INVALID"))
		}
	}
	if validationFlow.forSigning {
		//rebuild instruction and check if generate same inst hash
		return chain.Blockchain.verifyPreProcessingBeaconBlockForSigning(curView, beaconBlock, incurInstructions, validationFlow.confirmShardBlocks)
	}

	return nil
}

func (chain *BeaconChain) validateNewState(validationFlow *BeaconValidationFlow) error {
	return validationFlow.nextView.verifyPostProcessingBeaconBlock(validationFlow.beaconBlock, validationFlow.newBeaconCommitteeHashes)
}

func (chain *BeaconChain) ValidateAndProcessBlock(validationFlow *BeaconValidationFlow) (err error) {
	blockchain := chain.Blockchain
	beaconBlock := validationFlow.beaconBlock
	copiedCurView := validationFlow.copiedCurView
	validationMode := validationFlow.validationMode
	blockHeight := beaconBlock.GetHeight()
	blockHash := beaconBlock.Hash().String()

	if validationMode >= common.BASIC_VALIDATION {
		Logger.log.Infof("BEACON | Validation block signature height %+v - hash %+v", blockHeight, blockHash)
		if err := chain.validateBlockSignaturesWithCurrentView(validationFlow); err != nil {
			return err
		}
	}

	//validate block content (header first)
	if validationMode >= common.BASIC_VALIDATION {
		Logger.log.Infof("BEACON | Validation block header height %+v - hash %+v", blockHeight, blockHash)
		if err := chain.validateBlockHeader(validationFlow); err != nil {
			return err
		}
	}

	//process block before validate body! (need inccurInstruction from CommiteeeEngine)
	Logger.log.Infof("BEACON | Process block feature height %+v - hash %+v", blockHeight, blockHash)
	var inccurInstruction [][]string
	validationFlow.nextView, validationFlow.newBeaconCommitteeHashes, validationFlow.committeeChange, inccurInstruction, err =
		copiedCurView.updateBeaconBestState(beaconBlock, blockchain)
	if err != nil {
		return err
	}

	if validationMode >= common.FULL_VALIDATION {
		Logger.log.Infof("BEACON | Validation block body height %+v - hash %+v", blockHeight, blockHash)
		if err := chain.validateBlockBody(validationFlow, inccurInstruction); err != nil {
			return err
		}
	}

	//validate new state
	if validationMode >= common.BASIC_VALIDATION {
		Logger.log.Infof("BEACON | Validate new state height %+v - hash %+v", blockHeight, blockHash)
		if err := chain.validateNewState(validationFlow); err != nil {
			return err
		}
	}
	return nil
}

func (chain *BeaconChain) ValidatePreSignBlock(block types.BlockInterface, committees []incognitokey.CommitteePublicKey) error {
	validationFlow, err := chain.getDataBeforeBlockValidation(block.(*types.BeaconBlock), common.FULL_VALIDATION, true)
	if err != nil {
		return err
	}

	return chain.ValidateAndProcessBlock(validationFlow)
}

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
	return chain.multiView.GetAllViewsWithBFS(chain.multiView.GetFinalView())
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
