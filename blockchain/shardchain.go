package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/pubsub"
	"os"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/multiview"
)

type ShardChain struct {
	shardID   int
	multiView *multiview.MultiView

	BlockGen    *BlockGenerator
	Blockchain  *BlockChain
	hashHistory *lru.Cache
	ChainName   string
	Ready       bool

	insertLock sync.Mutex
}

func NewShardChain(shardID int, multiView *multiview.MultiView, blockGen *BlockGenerator, blockchain *BlockChain, chainName string) *ShardChain {
	return &ShardChain{shardID: shardID, multiView: multiView, BlockGen: blockGen, Blockchain: blockchain, ChainName: chainName}
}

func (chain *ShardChain) GetDatabase() incdb.Database {
	return chain.Blockchain.GetShardChainDatabase(byte(chain.shardID))
}

func (chain *ShardChain) GetFinalView() multiview.View {
	return chain.multiView.GetFinalView()
}

func (chain *ShardChain) GetBestView() multiview.View {
	return chain.multiView.GetBestView()
}

func (chain *ShardChain) GetViewByHash(hash common.Hash) multiview.View {
	return chain.multiView.GetViewByHash(hash)
}

func (chain *ShardChain) GetBestState() *ShardBestState {
	return chain.multiView.GetBestView().(*ShardBestState)
}

func (s *ShardChain) GetEpoch() uint64 {
	return s.GetBestState().Epoch
}

func (s *ShardChain) InsertBatchBlock([]types.BlockInterface) (int, error) {
	panic("implement me")
}

func (s *ShardChain) GetCrossShardState() map[byte]uint64 {

	res := make(map[byte]uint64)
	for index, key := range s.GetBestState().BestCrossShard {
		res[index] = key
	}
	return res
}

func (s *ShardChain) GetAllViewHash() (res []common.Hash) {
	for _, v := range s.multiView.GetAllViewsWithBFS() {
		res = append(res, *v.GetHash())
	}
	return
}

func (s *ShardChain) GetBestViewHeight() uint64 {
	return s.CurrentHeight()
}

func (s *ShardChain) GetFinalViewHeight() uint64 {
	return s.GetFinalView().GetHeight()
}

func (s *ShardChain) GetBestViewHash() string {
	return s.GetBestState().BestBlockHash.String()
}

func (s *ShardChain) GetFinalViewHash() string {
	return s.GetBestState().Hash().String()
}
func (chain *ShardChain) GetLastBlockTimeStamp() int64 {
	return chain.GetBestState().BestBlock.Header.Timestamp
}

func (chain *ShardChain) GetMinBlkInterval() time.Duration {
	return chain.GetBestState().BlockInterval
}

func (chain *ShardChain) GetMaxBlkCreateTime() time.Duration {
	return chain.GetBestState().BlockMaxCreateTime
}

func (chain *ShardChain) IsReady() bool {
	return chain.Ready
}

func (chain *ShardChain) SetReady(ready bool) {
	chain.Ready = ready
}

func (chain *ShardChain) CurrentHeight() uint64 {
	return chain.GetBestState().BestBlock.Header.Height
}

func (chain *ShardChain) GetCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	return append(result, chain.GetBestState().shardCommitteeEngine.GetShardCommittee()...)
}

func (chain *ShardChain) GetLastCommittee() []incognitokey.CommitteePublicKey {
	v := chain.multiView.GetViewByHash(*chain.GetBestView().GetPreviousHash())
	if v == nil {
		return nil
	}
	result := []incognitokey.CommitteePublicKey{}
	return append(result, v.GetCommittee()...)
}

func (chain *ShardChain) GetCommitteeByHeight(h uint64) ([]incognitokey.CommitteePublicKey, error) {
	bcStateRootHash := chain.Blockchain.GetBeaconBestState().ConsensusStateDBRootHash
	bcDB := chain.Blockchain.GetBeaconChainDatabase()
	bcStateDB, err := statedb.NewWithPrefixTrie(bcStateRootHash, statedb.NewDatabaseAccessWarper(bcDB))
	if err != nil {
		return nil, err
	}
	return statedb.GetOneShardCommittee(bcStateDB, byte(chain.shardID)), nil
}

func (chain *ShardChain) GetPendingCommittee() []incognitokey.CommitteePublicKey {
	result := []incognitokey.CommitteePublicKey{}
	return append(result, chain.GetBestState().shardCommitteeEngine.GetShardSubstitute()...)
}

func (chain *ShardChain) GetCommitteeSize() int {
	return len(chain.GetBestState().shardCommitteeEngine.GetShardCommittee())
}

func (chain *ShardChain) GetPubKeyCommitteeIndex(pubkey string) int {
	for index, key := range chain.GetBestState().shardCommitteeEngine.GetShardCommittee() {
		if key.GetMiningKeyBase58(chain.GetBestState().ConsensusAlgorithm) == pubkey {
			return index
		}
	}
	return -1
}

func (chain *ShardChain) GetLastProposerIndex() int {
	return chain.GetBestState().ShardProposerIdx
}

func (chain *ShardChain) CreateNewBlock(
	version int, proposer string, round int, startTime int64,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash) (types.BlockInterface, error) {
	Logger.log.Infof("Begin Start New Block Shard %+v", time.Now())
	newBlock, err := chain.Blockchain.NewBlockShard(
		chain.GetBestState(),
		version, proposer, round,
		startTime, committees, committeeViewHash)
	Logger.log.Infof("Finish New Block Shard %+v", time.Now())
	if err != nil {
		Logger.log.Error(err)
		return nil, err
	}
	if version >= 2 {
		newBlock.Header.Proposer = proposer
		newBlock.Header.ProposeTime = startTime
	}

	Logger.log.Infof("Finish Create New Block")
	return newBlock, nil
}

func (chain *ShardChain) CreateNewBlockFromOldBlock(
	oldBlock types.BlockInterface,
	proposer string, startTime int64,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash,
) (types.BlockInterface, error) {
	b, _ := json.Marshal(oldBlock)
	newBlock := new(types.ShardBlock)
	json.Unmarshal(b, &newBlock)

	newBlock.Header.Proposer = proposer
	newBlock.Header.ProposeTime = startTime

	return newBlock, nil
}

func (chain *ShardChain) validateBlockSignaturesWithCurrentView(block types.BlockInterface, curView *ShardBestState, committee []incognitokey.CommitteePublicKey) (err error) {
	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerPosition(block.(*types.ShardBlock),
		curView.ShardProposerIdx, committee,
		curView.MinShardCommitteeSize); err != nil {
		return err
	}

	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}

	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committee); err != nil {
		return err
	}
	return nil
}

func (chain *ShardChain) ValidateBlockSignatures(block types.BlockInterface, committee []incognitokey.CommitteePublicKey) error {

	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(block, chain.GetConsensusType()); err != nil {
		return err
	}

	if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(block, committee); err != nil {
		return err
	}
	return nil
}

func (chain *ShardChain) getCommitteeFromBlock(shardBlock *types.ShardBlock, curView *ShardBestState) (committee []incognitokey.CommitteePublicKey, err error) {
	if curView.shardCommitteeEngine.Version() == committeestate.SELF_SWAP_SHARD_VERSION ||
		shardBlock.Header.CommitteeFromBlock.IsZeroValue() {
		committee = curView.GetShardCommittee()
	} else {
		committee, err = chain.Blockchain.GetShardCommitteeFromBeaconHash(shardBlock.Header.CommitteeFromBlock, byte(shardBlock.GetShardID()))
		if err != nil {
			return nil, err
		}
	}
	return committee, err
}

type ShardValidationFlow struct {
	validationMode  int
	blockchain      *BlockChain
	curView         *ShardBestState
	nextView        *ShardBestState
	block           *types.ShardBlock
	beaconBlocks    []*types.BeaconBlock
	blockCommittees []incognitokey.CommitteePublicKey

	shardCommitteeHashes *committeestate.ShardCommitteeStateHash
	committeeChange      *committeestate.CommitteeChange
}

func (chain *ShardChain) validateBlockHeader(flow *ShardValidationFlow) error {
	chain.Blockchain.verifyPreProcessingShardBlock(flow.curView, flow.block, flow.beaconBlocks, false, flow.blockCommittees)
	shardBestState := flow.curView
	committees := flow.blockCommittees
	blockchain := flow.blockchain
	shardBlock := flow.block

	if shardBestState.shardCommitteeEngine.Version() == committeestate.SLASHING_VERSION {
		if !shardBestState.CommitteeFromBlock().IsZeroValue() {
			newCommitteesPubKeys, _ := incognitokey.CommitteeKeyListToString(committees)
			oldCommitteesPubKeys, _ := incognitokey.CommitteeKeyListToString(shardBestState.GetCommittee())
			//Logger.log.Infof("new Committee %+v \n old Committees %+v", newCommitteesPubKeys, oldCommitteesPubKeys)
			temp := common.DifferentElementStrings(oldCommitteesPubKeys, newCommitteesPubKeys)
			if len(temp) != 0 {
				oldBeaconBlock, _, err := blockchain.GetBeaconBlockByHash(shardBestState.CommitteeFromBlock())
				if err != nil {
					return err
				}
				newBeaconBlock, _, err := blockchain.GetBeaconBlockByHash(shardBlock.Header.CommitteeFromBlock)
				if err != nil {
					return err
				}
				if oldBeaconBlock.Header.Height >= newBeaconBlock.Header.Height {
					return NewBlockChainError(WrongBlockHeightError,
						fmt.Errorf("Height of New Shard Block's Committee From Block %+v is smaller than current Committee From Block View %+v",
							newBeaconBlock.Header.Hash(), oldBeaconBlock.Header.Hash()))
				}
			}
		}
	}

	// check with current final best state
	// shardBlock can only be insert if it match the current best state
	if !shardBestState.BestBlockHash.IsEqual(&shardBlock.Header.PreviousBlockHash) {
		return NewBlockChainError(ShardBestStateNotCompatibleError, fmt.Errorf("Current Best Block Hash %+v, New Shard Block %+v, Previous Block Hash of New Block %+v", shardBestState.BestBlockHash, shardBlock.Header.Height, shardBlock.Header.PreviousBlockHash))
	}
	if shardBestState.ShardHeight+1 != shardBlock.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Shard Block height of new Shard Block should be %+v, but get %+v", shardBestState.ShardHeight+1, shardBlock.Header.Height))
	}
	if shardBlock.Header.BeaconHeight < shardBestState.BeaconHeight {
		return NewBlockChainError(ShardBestStateBeaconHeightNotCompatibleError, fmt.Errorf("Shard Block contain invalid beacon height, current beacon height %+v but get %+v ", shardBestState.BeaconHeight, shardBlock.Header.BeaconHeight))
	}

	return nil
}

func (chain *ShardChain) validateBlockBody(flow *ShardValidationFlow) error {
	shardID := flow.curView.ShardID
	shardBlock := flow.block
	curView := flow.curView

	//validate transaction
	if err := flow.blockchain.verifyTransactionFromNewBlock(shardID, shardBlock.Body.Transactions, int64(curView.BeaconHeight), curView); err != nil {
		return NewBlockChainError(TransactionFromNewBlockError, err)
	}

	//TODO: validate cross shard transaction hash (only check hash)
	//TODO: validate cross shard transaction content (when beacon chain not confirm, we need to validate its content)

	return nil
}

func (chain *ShardChain) getDataBeforeBlockValidation(shardBlock *types.ShardBlock, validationMode int) (*ShardValidationFlow, error) {
	blockHash := shardBlock.Header.Hash()
	blockHeight := shardBlock.Header.Height
	shardID := shardBlock.Header.ShardID
	preHash := shardBlock.Header.PreviousBlockHash
	blockchain := chain.Blockchain

	validationFlow := new(ShardValidationFlow)
	validationFlow.block = shardBlock
	validationFlow.validationMode = validationMode

	//check if view is committed
	checkView := chain.GetViewByHash(blockHash)
	if checkView != nil {
		return nil, NewBlockChainError(ShardBlockAlreadyExist, fmt.Errorf("View already exists"))
	}

	//get current view that block link to
	preView := chain.GetViewByHash(preHash)
	if preView == nil {
		ctx, _ := context.WithTimeout(context.Background(), DefaultMaxBlockSyncTime)
		blockchain.config.Syncker.SyncMissingShardBlock(ctx, "", shardID, preHash)
		return nil, NewBlockChainError(InsertShardBlockError, fmt.Errorf("ShardBlock %v link to wrong view (%s)", blockHeight, preHash.String()))
	}
	curView := preView.(*ShardBestState)
	validationFlow.curView = curView

	previousBeaconHeight := curView.BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain, previousBeaconHeight+1, shardBlock.Header.BeaconHeight)
	validationFlow.beaconBlocks = beaconBlocks
	if err != nil {
		return nil, NewBlockChainError(FetchBeaconBlocksError, err)
	}

	committee, err := chain.getCommitteeFromBlock(shardBlock, curView)
	validationFlow.blockCommittees = committee
	if err != nil {
		return nil, NewBlockChainError(CommitteeFromBlockNotFoundError, err)
	}

	//TODO: get cross shard block (when beacon chain not confirm, we need to validate its content)
	return validationFlow, nil
}

func (chain *ShardChain) processBlock(flow *ShardValidationFlow) (err error) {
	blockchain := flow.blockchain
	shardBlock := flow.block
	beaconBlocks := flow.beaconBlocks
	committees := flow.blockCommittees
	curView := flow.curView

	flow.nextView, flow.shardCommitteeHashes, flow.committeeChange, err = curView.updateShardBestState(blockchain, shardBlock, beaconBlocks, committees)
	if err != nil {
		return err
	}

	err = blockchain.processSalaryInstructions(flow.nextView.rewardStateDB, beaconBlocks, curView.ShardID)
	if err != nil {
		return err
	}

	return err
}

func (chain *ShardChain) validateNewState(flow *ShardValidationFlow) (err error) {
	if err = flow.nextView.verifyPostProcessingShardBlock(flow.block, byte(flow.curView.ShardID), flow.shardCommitteeHashes); err != nil {
		return err
	}
	return err
}

func (chain *ShardChain) commitAndStore(flow *ShardValidationFlow) (err error) {
	if err = flow.nextView.shardCommitteeEngine.Commit(flow.shardCommitteeHashes); err != nil {
		return err
	}

	if err = flow.blockchain.processStoreShardBlock(flow.nextView, flow.block, flow.committeeChange, flow.beaconBlocks); err != nil {
		return err
	}
	return err
}

func (chain *ShardChain) InsertBlock(block types.BlockInterface, validationMode int) error {

	blockchain := chain.Blockchain
	shardBlock := block.(*types.ShardBlock)
	blockHeight := shardBlock.Header.Height
	shardID := shardBlock.Header.ShardID
	blockHash := shardBlock.Hash().String()
	//update validation Mode if need
	fullValidation := os.Getenv("FULL_VALIDATION") //trigger full validation when sync network for rechecking code logic
	if fullValidation == "1" {
		validationMode = common.FULL_VALIDATION
	}

	//get required object for validation
	Logger.log.Infof("SHARD %+v | Begin insert block height %+v - hash %+v, get required data for validate", shardID, blockHeight, blockHash)
	validationFlow, err := chain.getDataBeforeBlockValidation(shardBlock, validationMode)
	validationFlow.blockchain = blockchain
	if err != nil {
		return err
	}

	//validation block signature with current view
	if validationMode == common.FULL_VALIDATION {
		Logger.log.Infof("SHARD %+v | Validation block signature height %+v - hash %+v", shardID, blockHeight, blockHash)
		if err := chain.validateBlockSignaturesWithCurrentView(block, validationFlow.curView, validationFlow.blockCommittees); err != nil {
			return err
		}
	}

	//validate block content
	Logger.log.Infof("SHARD %+v | Validation block header height %+v - hash %+v", shardID, blockHeight, blockHash)
	if err := chain.validateBlockHeader(validationFlow); err != nil {
		return err
	}

	if validationMode == common.FULL_VALIDATION {
		Logger.log.Infof("SHARD %+v | Validation block body height %+v - hash %+v", shardID, blockHeight, blockHash)
		if err := chain.validateBlockBody(validationFlow); err != nil {
			return err
		}
	}

	//process block
	Logger.log.Infof("SHARD %+v | Process block feature height %+v - hash %+v", shardID, blockHeight, blockHash)
	if err := chain.processBlock(validationFlow); err != nil {
		return err
	}

	//validate new state
	Logger.log.Infof("SHARD %+v | Validate new state height %+v - hash %+v", shardID, blockHeight, blockHash)
	if err = chain.validateNewState(validationFlow); err != nil {
		return err
	}

	//store and commit
	Logger.log.Infof("SHARD %+v | Commit and Store block height %+v - hash %+v", shardID, blockHeight, blockHash)
	if err = chain.commitAndStore(validationFlow); err != nil {
		return err
	}

	//broadcast after successfully insert
	blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewShardblockTopic, shardBlock))
	blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardBeststateTopic, validationFlow.nextView))
	return nil
}

func (chain *ShardChain) InsertAndBroadcastBlock(block types.BlockInterface) error {

	go chain.Blockchain.config.Server.PushBlockToAll(block, "", false)

	if err := chain.InsertBlock(block, common.BASIC_VALIDATION); err != nil {
		return err
	}

	return nil
}

func (chain *ShardChain) CheckExistedBlk(block types.BlockInterface) bool {
	blkHash := block.Hash()
	_, err := rawdbv2.GetShardBlockByHash(chain.Blockchain.GetShardChainDatabase(byte(chain.shardID)), *blkHash)
	return err == nil
}

func (chain *ShardChain) ReplacePreviousValidationData(previousBlockHash common.Hash, newValidationData string) error {

	if err := chain.Blockchain.ReplacePreviousValidationData(previousBlockHash, newValidationData); err != nil {
		Logger.log.Error(err)
		return err
	}

	return nil
}

func (chain *ShardChain) InsertAndBroadcastBlockWithPrevValidationData(block types.BlockInterface, newValidationData string) error {

	go chain.Blockchain.config.Server.PushBlockToAll(block, newValidationData, false)

	if err := chain.InsertBlock(block, common.BASIC_VALIDATION); err != nil {
		return err
	}

	if err := chain.ReplacePreviousValidationData(block.GetPrevHash(), newValidationData); err != nil {
		return err
	}

	return nil
}

func (chain *ShardChain) GetActiveShardNumber() int {
	return 0
}

func (chain *ShardChain) GetChainName() string {
	return chain.ChainName
}

func (chain *ShardChain) GetConsensusType() string {
	return chain.GetBestState().ConsensusAlgorithm
}

func (chain *ShardChain) GetShardID() int {
	return chain.shardID
}

func (chain *ShardChain) IsBeaconChain() bool {
	return false
}

func (chain *ShardChain) UnmarshalBlock(blockString []byte) (types.BlockInterface, error) {
	var shardBlk types.ShardBlock
	err := json.Unmarshal(blockString, &shardBlk)
	if err != nil {
		return nil, err
	}
	return &shardBlk, nil
}

func (chain *ShardChain) ValidatePreSignBlock(block types.BlockInterface, committees []incognitokey.CommitteePublicKey) error {
	return chain.Blockchain.VerifyPreSignShardBlock(block.(*types.ShardBlock), committees, byte(block.(*types.ShardBlock).GetShardID()))
}

func (chain *ShardChain) GetAllView() []multiview.View {
	return chain.multiView.GetAllViewsWithBFS()
}

//CommitteesV2 get committees by block for shardChain
// Input block must be ShardBlock
func (chain *ShardChain) GetCommitteeV2(block types.BlockInterface) ([]incognitokey.CommitteePublicKey, error) {
	var err error
	var isShardView bool
	var shardView *ShardBestState
	shardView, isShardView = chain.GetViewByHash(block.GetPrevHash()).(*ShardBestState)
	if !isShardView {
		shardView = chain.GetBestState()
	}
	result := []incognitokey.CommitteePublicKey{}

	shardBlock, isShardBlock := block.(*types.ShardBlock)
	if !isShardBlock {
		return result, fmt.Errorf("Shard Chain NOT insert Shard Block Types")
	}
	if shardView.shardCommitteeEngine.Version() == committeestate.SELF_SWAP_SHARD_VERSION || shardBlock.Header.CommitteeFromBlock.IsZeroValue() {
		result = append(result, chain.GetBestState().shardCommitteeEngine.GetShardCommittee()...)
	} else if shardView.shardCommitteeEngine.Version() == committeestate.SLASHING_VERSION {
		result, err = chain.Blockchain.GetShardCommitteeFromBeaconHash(block.CommitteeFromBlock(), byte(chain.shardID))
		if err != nil {
			return result, err
		}
	}

	return result, nil
}

func (chain *ShardChain) CommitteeStateVersion() uint {
	return chain.GetBestState().shardCommitteeEngine.Version()
}

//BestViewCommitteeFromBlock ...
func (chain *ShardChain) BestViewCommitteeFromBlock() common.Hash {
	return chain.GetBestState().CommitteeFromBlock()
}

func (chain *ShardChain) GetChainDatabase() incdb.Database {
	return chain.Blockchain.GetShardChainDatabase(byte(chain.shardID))
}

func (chain *ShardChain) CommitteeEngineVersion() uint {
	return chain.multiView.GetBestView().CommitteeEngineVersion()
}
