package blockchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/pubsub"
	"os"
	"sort"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/txpool"

	"github.com/incognitochain/incognito-chain/common"
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

	TxPool      txpool.TxPool
	TxsVerifier txpool.TxVerifier

	insertLock sync.Mutex
}

func NewShardChain(
	shardID int,
	multiView *multiview.MultiView,
	blockGen *BlockGenerator,
	blockchain *BlockChain,
	chainName string,
	tp txpool.TxPool,
	tv txpool.TxVerifier,
) *ShardChain {
	return &ShardChain{
		shardID:     shardID,
		multiView:   multiView,
		BlockGen:    blockGen,
		Blockchain:  blockchain,
		ChainName:   chainName,
		TxPool:      tp,
		TxsVerifier: tv,
	}
}

func (chain *ShardChain) CloneMultiView() *multiview.MultiView {
	return chain.multiView.Clone()
}

func (chain *ShardChain) SetMultiView(multiView *multiview.MultiView) {
	chain.multiView = multiView
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

type ShardProducingFlow struct {
	curView                 *ShardBestState
	nextView                *ShardBestState
	newBlock                *types.ShardBlock
	version                 int
	proposer                string
	round                   int
	startTime               int64
	blockCommittees         []incognitokey.CommitteePublicKey
	committeeViewHash       common.Hash
	beaconBlocks            []*types.BeaconBlock
	newShardCommitteeHashes *committeestate.ShardCommitteeStateHash
	isOldBeaconHeight       bool
	processBeaconBlock      types.BeaconBlock
	crossTransactions       map[byte][]types.CrossTransaction
	transactionsForNewBlock []metadata.Transaction
}

func (chain *ShardChain) getDataBeforeBlockProducing(buildView *ShardBestState, version int, proposer string, round int, startTime int64,
	committeeViewHash common.Hash) (*ShardProducingFlow, error) {
	blockchain := chain.Blockchain

	blockCommittees, err := chain.getCommitteeFromBlock(committeeViewHash, chain.GetBestState())

	createFlow := &ShardProducingFlow{
		nextView:          nil,
		newBlock:          types.NewShardBlock(),
		version:           version,
		proposer:          proposer,
		round:             round,
		startTime:         startTime,
		blockCommittees:   blockCommittees,
		committeeViewHash: committeeViewHash,
		isOldBeaconHeight: false,
	}
	createFlow.curView = NewShardBestState()
	if err := createFlow.curView.cloneShardBestStateFrom(buildView); err != nil {
		return nil, err
	}
	curView := createFlow.curView

	beaconProcessView := blockchain.BeaconChain.GetFinalView().(*BeaconBestState)
	beaconProcessHeight := beaconProcessView.GetHeight()
	if beaconProcessHeight-curView.BeaconHeight > MAX_BEACON_BLOCK {
		beaconProcessHeight = curView.BeaconHeight + MAX_BEACON_BLOCK
	}

	if curView.shardCommitteeEngine.Version() == committeestate.SELF_SWAP_SHARD_VERSION {
		if err != nil {
			return nil, err
		}
		if beaconProcessHeight > blockchain.config.ChainParams.StakingFlowV2Height {
			beaconProcessHeight = blockchain.config.ChainParams.StakingFlowV2Height
		}
	}

	//if no new beacon block,
	if beaconProcessHeight == curView.BeaconHeight {
		//return error if this is slashing version
		if curView.shardCommitteeEngine.Version() == committeestate.SLASHING_VERSION {
			if beaconProcessHeight <= curView.BeaconHeight {
				Logger.log.Info("Waiting For Beacon Produce Block beaconProcessHeight %+v shardBestState.BeaconHeight %+v",
					beaconProcessHeight, curView.BeaconHeight)
				return nil, errors.New("Waiting For Beacon Produce Block")
			}
		}

		//continue with the old beacon block
		createFlow.isOldBeaconHeight = true
		beaconBlock, _, err := blockchain.GetBeaconBlockByHash(curView.BestBeaconHash)
		if err != nil {
			return nil, err
		}
		createFlow.processBeaconBlock = *beaconBlock
		createFlow.beaconBlocks = []*types.BeaconBlock{}
	} else {
		//get new beacon blocks to process
		beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain, curView.BeaconHeight+1, beaconProcessHeight)
		if err != nil {
			return nil, err
		}
		createFlow.beaconBlocks = beaconBlocks
		createFlow.processBeaconBlock = *beaconBlocks[len(beaconBlocks)-1]
	}

	//cross shard
	tempPrivateKey := blockchain.config.BlockGen.createTempKeyset(startTime)
	createFlow.crossTransactions = blockchain.config.BlockGen.getCrossShardData(curView)

	// Get Transaction for new block
	blockCreationLeftOver := time.Duration(curView.BlockMaxCreateTime.Nanoseconds() - startTime)
	txsToAddFromBlock, err := blockchain.config.BlockGen.getTransactionForNewBlock(curView, &tempPrivateKey, curView.ShardID, createFlow.beaconBlocks, blockCreationLeftOver, createFlow.processBeaconBlock.Header.Height)
	if err != nil {
		return nil, err
	}
	transactionsForNewBlock := []metadata.Transaction{}
	transactionsForNewBlock = append(transactionsForNewBlock, txsToAddFromBlock...)
	// build txs with metadata
	transactionsForNewBlock, err = blockchain.BuildResponseTransactionFromTxsWithMetadata(curView, transactionsForNewBlock, &tempPrivateKey, curView.ShardID)
	createFlow.transactionsForNewBlock = transactionsForNewBlock

	return createFlow, nil
}

func (chain *ShardChain) updateHeaderRootHash(flow *ShardProducingFlow) error {
	newShardBlock := flow.newBlock
	shardInstructions := flow.newBlock.Body.Instructions
	blockchain := chain.Blockchain
	shardID := byte(chain.shardID)

	merkleRoots := Merkle{}.BuildMerkleTreeStore(newShardBlock.Body.Transactions)
	merkleRoot := &common.Hash{}
	if len(merkleRoots) > 0 {
		merkleRoot = merkleRoots[len(merkleRoots)-1]
	}
	crossTransactionRoot, err := CreateMerkleCrossTransaction(newShardBlock.Body.CrossTransactions)
	if err != nil {
		return err
	}
	txInstructions, err := CreateShardInstructionsFromTransactionAndInstruction(newShardBlock.Body.Transactions, blockchain, shardID, newShardBlock.Header.Height)
	if err != nil {
		return err
	}

	totalInstructions := []string{}
	for _, value := range txInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	for _, value := range shardInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	instructionsHash, err := generateHashFromStringArray(totalInstructions)
	if err != nil {
		return NewBlockChainError(InstructionsHashError, err)
	}

	// Instruction merkle root
	flattenTxInsts, err := FlattenAndConvertStringInst(txInstructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from Tx: %+v", err))
	}
	flattenInsts, err := FlattenAndConvertStringInst(shardInstructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from block body: %+v", err))
	}
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be preserved in shardprocess
	instMerkleRoot := GetKeccak256MerkleRoot(insts)
	// shard tx root
	_, shardTxMerkleData := CreateShardTxRoot(newShardBlock.Body.Transactions)
	// Add Root Hash To Header
	newShardBlock.Header.TxRoot = *merkleRoot
	newShardBlock.Header.ShardTxRoot = shardTxMerkleData[len(shardTxMerkleData)-1]
	newShardBlock.Header.CrossTransactionRoot = *crossTransactionRoot
	newShardBlock.Header.InstructionsRoot = instructionsHash
	newShardBlock.Header.CommitteeRoot = flow.newShardCommitteeHashes.ShardCommitteeHash
	newShardBlock.Header.PendingValidatorRoot = flow.newShardCommitteeHashes.ShardSubstituteHash
	newShardBlock.Header.StakingTxRoot = common.Hash{}
	newShardBlock.Header.Timestamp = flow.startTime
	copy(newShardBlock.Header.InstructionMerkleRoot[:], instMerkleRoot)

	if flow.version >= 2 {
		newShardBlock.Header.Proposer = flow.proposer
		newShardBlock.Header.ProposeTime = flow.startTime
	}
	return nil
}

func (chain *ShardChain) buildBlockWithoutHeaderRootHash(flow *ShardProducingFlow) error {
	blockchain := chain.Blockchain
	shardID := byte(chain.shardID)
	curView := flow.curView

	beaconBlocks := flow.beaconBlocks
	beaconProcessHeight := flow.processBeaconBlock.GetHeight()
	beaconProcessHash := flow.processBeaconBlock.Hash()

	// process instruction from beacon block
	beaconInstructions, _, err := blockchain.
		preProcessInstructionFromBeacon(beaconBlocks, curView.ShardID)
	if err != nil {
		return err
	}
	currentPendingValidators := curView.GetShardPendingValidator()
	shardPendingValidatorStr, err := incognitokey.
		CommitteeKeyListToString(currentPendingValidators)
	if err != nil {
		return err
	}

	env := committeestate.
		NewShardEnvBuilder().
		BuildBeaconInstructions(beaconInstructions).
		BuildShardID(curView.ShardID).
		BuildNumberOfFixedBlockValidators(blockchain.config.ChainParams.NumberOfFixedBlockValidators).
		BuildShardHeight(curView.ShardHeight).
		Build()

	committeeChange, err := curView.shardCommitteeEngine.ProcessInstructionFromBeacon(env)
	if err != nil {
		return err
	}
	curView.shardCommitteeEngine.AbortUncommittedShardState()

	currentPendingValidators, err = updateCommiteesWithAddedAndRemovedListValidator(currentPendingValidators,
		committeeChange.ShardSubstituteAdded[curView.ShardID],
		committeeChange.ShardSubstituteRemoved[curView.ShardID])

	blockCommitteesStr, err := incognitokey.CommitteeKeyListToString(flow.blockCommittees)
	if err != nil {
		return NewBlockChainError(ProcessInstructionFromBeaconError, err)
	}

	shardPendingValidatorStr, err = incognitokey.CommitteeKeyListToString(currentPendingValidators)
	if err != nil {
		return NewBlockChainError(ProcessInstructionFromBeaconError, err)
	}

	shardInstructions, _, _, err := blockchain.generateInstruction(curView, shardID,
		beaconProcessHeight, flow.isOldBeaconHeight, beaconBlocks,
		shardPendingValidatorStr, blockCommitteesStr)
	if err != nil {
		return NewBlockChainError(GenerateInstructionError, err)
	}

	if len(shardInstructions) != 0 {
		Logger.log.Info("Shard Producer: Instruction", shardInstructions)
	}

	flow.newBlock.BuildShardBlockBody(shardInstructions, flow.crossTransactions, flow.transactionsForNewBlock)
	totalTxsFee := curView.shardCommitteeEngine.BuildTotalTxsFeeFromTxs(flow.newBlock.Body.Transactions)

	flow.newBlock.Header = types.ShardHeader{
		Producer:           flow.proposer, //committeeMiningKeys[producerPosition],
		ProducerPubKeyStr:  flow.proposer,
		ShardID:            shardID,
		Version:            flow.version,
		PreviousBlockHash:  curView.BestBlockHash,
		Height:             curView.ShardHeight + 1,
		Round:              flow.round,
		Epoch:              flow.processBeaconBlock.GetCurrentEpoch(),
		CrossShardBitMap:   CreateCrossShardByteArray(flow.newBlock.Body.Transactions, shardID),
		BeaconHeight:       beaconProcessHeight,
		BeaconHash:         *beaconProcessHash,
		TotalTxsFee:        totalTxsFee,
		ConsensusType:      curView.ConsensusAlgorithm,
		CommitteeFromBlock: flow.committeeViewHash,
	}

	return nil
}

func (chain *ShardChain) CreateNewBlock(
	buildView multiview.View,
	version int, proposer string, round int, startTime int64,
	committees []incognitokey.CommitteePublicKey,
	committeeViewHash common.Hash) (types.BlockInterface, error) {
	Logger.log.Infof("Begin Start New Block Shard %+v", time.Now())
	if buildView == nil {
		buildView = chain.GetBestView()
	}
	linkedView := buildView.(*ShardBestState) //this should be pass from consensus package
	createFlow, err := chain.getDataBeforeBlockProducing(linkedView, version, proposer, round, startTime, committeeViewHash)
	if err != nil {
		return nil, err
	}

	if err := chain.buildBlockWithoutHeaderRootHash(createFlow); err != nil {
		return nil, err
	}

	createFlow.nextView, createFlow.newShardCommitteeHashes, _, err =
		linkedView.updateShardBestState(chain.Blockchain, createFlow.newBlock, createFlow.beaconBlocks, createFlow.blockCommittees)

	if err != nil {
		return nil, err
	}

	if err := chain.updateHeaderRootHash(createFlow); err != nil {
		return nil, err
	}

	return createFlow.newBlock, nil
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

func (chain *ShardChain) validateBlockSignaturesWithCurrentView(validationFlow *ShardValidationFlow) (err error) {
	shardBlock := validationFlow.block
	curView := validationFlow.curView
	committee := validationFlow.blockCommittees

	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerPosition(shardBlock,
		curView.ShardProposerIdx, committee,
		curView.MinShardCommitteeSize); err != nil {
		return err
	}

	if err := chain.Blockchain.config.ConsensusEngine.ValidateProducerSig(shardBlock, chain.GetConsensusType()); err != nil {
		return err
	}

	if !validationFlow.forSigning {
		if err := chain.Blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(shardBlock, committee); err != nil {
			return err
		}
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

func (chain *ShardChain) getCommitteeFromBlock(CommitteeFromBlock common.Hash, curView *ShardBestState) (committee []incognitokey.CommitteePublicKey, err error) {
	if curView.shardCommitteeEngine.Version() == committeestate.SELF_SWAP_SHARD_VERSION ||
		CommitteeFromBlock.IsZeroValue() {
		committee = curView.GetShardCommittee()
	} else {
		committee, err = chain.Blockchain.GetShardCommitteeFromBeaconHash(CommitteeFromBlock, byte(curView.GetShardID()))
		if err != nil {
			return nil, err
		}
	}
	return committee, err
}

type ShardValidationFlow struct {
	validationMode          int
	forSigning              bool
	curView                 *ShardBestState
	nextView                *ShardBestState
	block                   *types.ShardBlock
	beaconBlocks            []*types.BeaconBlock
	blockCommittees         []incognitokey.CommitteePublicKey
	crossShardBlockToAdd    map[byte][]*types.CrossShardBlock
	newShardCommitteeHashes *committeestate.ShardCommitteeStateHash
	committeeChange         *committeestate.CommitteeChange
}

func (chain *ShardChain) validateBlockHeader(flow *ShardValidationFlow) error {
	chain.Blockchain.verifyPreProcessingShardBlock(flow.curView, flow.block, flow.beaconBlocks, false, flow.blockCommittees)
	shardBestState := flow.curView
	committees := flow.blockCommittees
	blockchain := chain.Blockchain
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
	blockchain := chain.Blockchain
	beaconBlocks := flow.beaconBlocks

	//validate transaction
	if err := blockchain.verifyTransactionFromNewBlock(shardID, shardBlock.Body.Transactions, int64(curView.BeaconHeight), curView); err != nil {
		return NewBlockChainError(TransactionFromNewBlockError, err)
	}

	//validate instruction
	beaconInstructions := [][]string{}
	shardCommittee, err := incognitokey.CommitteeKeyListToString(flow.blockCommittees)
	if err != nil {
		return err
	}

	shardPendingValidator := curView.GetShardPendingValidator()
	shardPendingValidatorStr := []string{}

	if curView != nil {
		var err error
		shardPendingValidatorStr, err = incognitokey.
			CommitteeKeyListToString(shardPendingValidator)
		if err != nil {
			return err
		}
	}

	beaconInstructions, _, err = blockchain.
		preProcessInstructionFromBeacon(beaconBlocks, curView.ShardID)
	if err != nil {
		return err
	}

	env := committeestate.
		NewShardEnvBuilder().
		BuildShardID(curView.ShardID).
		BuildBeaconInstructions(beaconInstructions).
		BuildNumberOfFixedBlockValidators(blockchain.config.ChainParams.NumberOfFixedBlockValidators).
		BuildShardHeight(curView.ShardHeight).
		Build()

	committeeChange, err := curView.shardCommitteeEngine.ProcessInstructionFromBeacon(env)
	if err != nil {
		return err
	}
	curView.shardCommitteeEngine.AbortUncommittedShardState()

	instructions := [][]string{}
	isOldBeaconHeight := false
	if curView.BeaconHeight == shardBlock.Header.BeaconHeight {
		isOldBeaconHeight = true
	}

	shardPendingValidator, err = updateCommiteesWithAddedAndRemovedListValidator(shardPendingValidator,
		committeeChange.ShardSubstituteAdded[curView.ShardID],
		committeeChange.ShardSubstituteRemoved[curView.ShardID])

	if err != nil {
		return NewBlockChainError(ProcessInstructionFromBeaconError, err)
	}

	shardPendingValidatorStr, err = incognitokey.CommitteeKeyListToString(shardPendingValidator)
	if err != nil {
		return NewBlockChainError(ProcessInstructionFromBeaconError, err)
	}

	instructions, _, shardCommittee, err = blockchain.generateInstruction(curView, shardID,
		shardBlock.Header.BeaconHeight, isOldBeaconHeight, beaconBlocks,
		shardPendingValidatorStr, shardCommittee)
	if err != nil {
		return NewBlockChainError(GenerateInstructionError, err)
	}

	txInstructions, err := CreateShardInstructionsFromTransactionAndInstruction(shardBlock.Body.Transactions, blockchain, shardID, shardBlock.Header.Height)
	if err != nil {
		return err
	}

	totalInstructions := []string{}
	for _, value := range txInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	for _, value := range instructions {
		totalInstructions = append(totalInstructions, value...)
	}

	if len(totalInstructions) != 0 {
		Logger.log.Info("totalInstructions:", totalInstructions)
	}

	if hash, ok := verifyHashFromStringArray(totalInstructions, shardBlock.Header.InstructionsRoot); !ok {
		return NewBlockChainError(InstructionsHashError, fmt.Errorf("Expect instruction hash to be %+v but get %+v", shardBlock.Header.InstructionsRoot, hash))
	}

	//check crossshard output coin content
	//TODO: add check crossshard output coin content in mode beacon full validation, and beacon not confirm this block shard yet
	if flow.forSigning {
		toShardAllCrossShardBlock := flow.crossShardBlockToAdd
		for fromShard, crossTransactions := range shardBlock.Body.CrossTransactions {
			toShardCrossShardBlocks := toShardAllCrossShardBlock[fromShard]
			sort.SliceStable(toShardCrossShardBlocks[:], func(i, j int) bool {
				return toShardCrossShardBlocks[i].Header.Height < toShardCrossShardBlocks[j].Header.Height
			})
			isValids := 0
			for _, crossTransaction := range crossTransactions {
				for index, toShardCrossShardBlock := range toShardCrossShardBlocks {
					//Compare block height and block hash
					if crossTransaction.BlockHeight == toShardCrossShardBlock.Header.Height {
						compareCrossTransaction := types.CrossTransaction{
							TokenPrivacyData: toShardCrossShardBlock.CrossTxTokenPrivacyData,
							OutputCoin:       toShardCrossShardBlock.CrossOutputCoin,
							BlockHash:        *toShardCrossShardBlock.Hash(),
							BlockHeight:      toShardCrossShardBlock.Header.Height,
						}
						targetHash := crossTransaction.Hash()
						hash := compareCrossTransaction.Hash()
						if !hash.IsEqual(&targetHash) {
							return NewBlockChainError(CrossTransactionHashError, fmt.Errorf("Cross Output Coin From New Block %+v not compatible with cross shard block in pool %+v", targetHash, hash))
						}
						if true {
							toShardCrossShardBlocks = toShardCrossShardBlocks[index:]
							isValids++
							break
						}
					}
				}
			}
			if len(crossTransactions) != isValids {
				return NewBlockChainError(CrossShardBlockError, fmt.Errorf("Can't not verify all cross shard block from shard %+v", fromShard))
			}
		}
	}

	return nil
}

func (chain *ShardChain) getDataBeforeBlockValidation(shardBlock *types.ShardBlock, validationMode int, forSigning bool) (*ShardValidationFlow, error) {
	blockHash := shardBlock.Header.Hash()
	blockHeight := shardBlock.Header.Height
	shardID := shardBlock.Header.ShardID
	preHash := shardBlock.Header.PreviousBlockHash
	blockchain := chain.Blockchain

	validationFlow := new(ShardValidationFlow)
	validationFlow.block = shardBlock
	validationFlow.validationMode = validationMode
	validationFlow.forSigning = forSigning
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
		return nil, NewBlockChainError(FetchBeaconBlocksError, fmt.Errorf("Cannot fetch beacon block height %v hash %v", shardBlock.Header.BeaconHeight, shardBlock.Header.BeaconHash.String()))
	}

	committee, err := chain.getCommitteeFromBlock(shardBlock.Header.CommitteeFromBlock, curView)
	validationFlow.blockCommittees = committee
	if err != nil {
		return nil, NewBlockChainError(CommitteeFromBlockNotFoundError, err)
	}

	//TODO: get cross shard block (when beacon chain not confirm, we need to validate the cross shard output coin)
	if forSigning {
		toShard := shardID
		var toShardAllCrossShardBlock = make(map[byte][]*types.CrossShardBlock)
		validationFlow.crossShardBlockToAdd = toShardAllCrossShardBlock
		// blockchain.config.Syncker.GetCrossShardBlocksForShardValidator(toShard, list map[byte]common.Hash) map[byte][]interface{}
		crossShardRequired := make(map[byte][]uint64)
		for fromShard, crossTransactions := range shardBlock.Body.CrossTransactions {
			for _, crossTransaction := range crossTransactions {
				//fmt.Println("Crossshard from ", fromShard, crossTransaction.BlockHash)
				crossShardRequired[fromShard] = append(crossShardRequired[fromShard], crossTransaction.BlockHeight)
			}
		}
		crossShardBlksFromPool, err := blockchain.config.Syncker.GetCrossShardBlocksForShardValidator(curView, crossShardRequired)
		if err != nil {
			return nil, NewBlockChainError(CrossShardBlockError, fmt.Errorf("Unable to get required crossShard blocks from pool in time"))
		}
		for sid, v := range crossShardBlksFromPool {
			heightList := make([]uint64, len(v))
			for i, b := range v {
				toShardAllCrossShardBlock[sid] = append(toShardAllCrossShardBlock[sid], b.(*types.CrossShardBlock))
				heightList[i] = b.(*types.CrossShardBlock).GetHeight()
			}
			Logger.log.Infof("Shard %v, GetCrossShardBlocksForShardValidator from shard %v: %v", toShard, sid, heightList)
		}
	}
	return validationFlow, nil
}

func (chain *ShardChain) validateNewState(flow *ShardValidationFlow) (err error) {
	if err = flow.nextView.verifyPostProcessingShardBlock(flow.block, byte(flow.curView.ShardID), flow.newShardCommitteeHashes); err != nil {
		return err
	}
	return err
}

func (chain *ShardChain) commitAndStore(flow *ShardValidationFlow) (err error) {
	if err = flow.nextView.shardCommitteeEngine.Commit(flow.newShardCommitteeHashes); err != nil {
		return err
	}

	if err = chain.Blockchain.processStoreShardBlock(flow.nextView, flow.block, flow.committeeChange, flow.beaconBlocks); err != nil {
		return err
	}
	return err
}

func (chain *ShardChain) InsertBlock(block types.BlockInterface, validationMode int) (err error) {

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
	validationFlow, err := chain.getDataBeforeBlockValidation(shardBlock, validationMode, false)
	if err != nil {
		return err
	}

	if err = chain.ValidateAndProcessBlock(block, validationFlow); err != nil {
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

//TODO: no need to pass committee => refactor bftv3
func (chain *ShardChain) ValidatePreSignBlock(block types.BlockInterface, committee []incognitokey.CommitteePublicKey) error {
	validationFlow, err := chain.getDataBeforeBlockValidation(block.(*types.ShardBlock), common.FULL_VALIDATION, true)
	if err != nil {
		return err
	}
	return chain.ValidateAndProcessBlock(block, validationFlow)
}

func (chain *ShardChain) ValidateAndProcessBlock(block types.BlockInterface, validationFlow *ShardValidationFlow) (err error) {
	shardBlock := block.(*types.ShardBlock)
	blockHeight := shardBlock.Header.Height
	shardID := shardBlock.Header.ShardID
	blockHash := shardBlock.Hash().String()
	validationMode := validationFlow.validationMode

	//validation block signature with current view
	if validationMode >= common.BASIC_VALIDATION {
		Logger.log.Infof("SHARD %+v | Validation block signature height %+v - hash %+v", shardID, blockHeight, blockHash)
		if err := chain.validateBlockSignaturesWithCurrentView(validationFlow); err != nil {
			return err
		}
	}

	//validate block content
	if validationMode >= common.BASIC_VALIDATION {
		Logger.log.Infof("SHARD %+v | Validation block header height %+v - hash %+v", shardID, blockHeight, blockHash)
		if err := chain.validateBlockHeader(validationFlow); err != nil {
			return err
		}
	}

	if validationMode >= common.FULL_VALIDATION {
		Logger.log.Infof("SHARD %+v | Validation block body height %+v - hash %+v", shardID, blockHeight, blockHash)
		if err := chain.validateBlockBody(validationFlow); err != nil {
			return err
		}
	}

	//process block
	Logger.log.Infof("SHARD %+v | Process block feature height %+v - hash %+v", shardID, blockHeight, blockHash)
	validationFlow.nextView, validationFlow.newShardCommitteeHashes, validationFlow.committeeChange, err =
		validationFlow.curView.updateShardBestState(chain.Blockchain, shardBlock, validationFlow.beaconBlocks, validationFlow.blockCommittees)
	if err != nil {
		return err
	}

	//validate new state
	if validationMode >= common.BASIC_VALIDATION {
		Logger.log.Infof("SHARD %+v | Validate new state height %+v - hash %+v", shardID, blockHeight, blockHash)
		if err := chain.validateNewState(validationFlow); err != nil {
			return err
		}
	}

	return nil
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
