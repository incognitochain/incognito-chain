package blockchain

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) NewBlockShard_V2(curView *ShardBestState, version int, proposer string, shardID byte, round int, crossShards map[byte]uint64, beaconHeight uint64, start time.Time) (newBlock *ShardBlock, err error) {
	processState := &ShardProcessState{
		curView:          curView,
		newView:          nil,
		blockchain:       blockchain,
		version:          version,
		proposer:         proposer,
		round:            round,
		newBlock:         NewShardBlock(),
		crossShardBlocks: make(map[byte][]*CrossShardBlock),
		startTime:        start,
		maxBeaconHeight:  beaconHeight,
	}

	if err := processState.PreProduceProcess(); err != nil {
		return nil, err
	}

	if err := processState.BuildBody(); err != nil {
		return nil, err
	}

	processState.newView, err = processState.curView.updateShardBestState(blockchain, processState.newBlock, processState.beaconBlocks, newCommitteeChange())
	if err != nil {
		return nil, err
	}

	if err := processState.BuildHeader(); err != nil {
		return nil, err
	}

	return processState.newBlock, nil
}

type ShardProcessState struct {
	//init state
	curView    *ShardBestState
	newView    *ShardBestState
	blockchain *BlockChain
	version    int
	proposer   string
	round      int

	//pre process state
	startTime         time.Time
	isOldBeaconHeight bool
	epoch             int64
	maxBeaconHeight   uint64
	newBlock          *ShardBlock
	crossShardBlocks  map[byte][]*CrossShardBlock
	beaconBlocks      []*BeaconBlock
	txs               []metadata.Transaction
	instructions      [][]string
}

func (s *ShardProcessState) PreProduceProcess() error {
	//get s2b blocks
	for sid, v := range s.blockchain.config.Syncker.GetCrossShardBlocksForShardProducer(s.curView.ShardID) {
		for _, b := range v {
			s.crossShardBlocks[sid] = append(s.crossShardBlocks[sid], b.(*CrossShardBlock))
		}
	}
	if s.maxBeaconHeight-s.curView.BeaconHeight > MAX_BEACON_BLOCK {
		s.maxBeaconHeight = s.curView.BeaconHeight + MAX_BEACON_BLOCK
	}

	beaconHashes, err := rawdbv2.GetBeaconBlockHashByIndex(s.blockchain.GetDatabase(), s.maxBeaconHeight)
	if err != nil {
		return err
	}
	beaconHash := beaconHashes[0]
	beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(s.blockchain.GetDatabase(), beaconHash)
	if err != nil {
		return err
	}
	beaconBlock := BeaconBlock{}
	err = json.Unmarshal(beaconBlockBytes, &beaconBlock)
	if err != nil {
		return err
	}
	epoch := beaconBlock.Header.Epoch
	if epoch-s.curView.Epoch >= 1 {
		s.maxBeaconHeight = s.curView.Epoch * s.blockchain.config.ChainParams.Epoch
		newBeaconHashes, err := rawdbv2.GetBeaconBlockHashByIndex(s.blockchain.GetDatabase(), s.maxBeaconHeight)
		if err != nil {
			return err
		}
		newBeaconHash := newBeaconHashes[0]
		copy(beaconHash[:], newBeaconHash.GetBytes())
		epoch = s.curView.Epoch + 1
	}
	Logger.log.Infof("Get Beacon Block With Height %+v, Shard BestState %+v", s.maxBeaconHeight, s.curView.BeaconHeight)
	//Fetch beacon block from height
	beaconBlocks, err := FetchBeaconBlockFromHeight(s.blockchain.GetDatabase(), s.curView.BeaconHeight+1, s.maxBeaconHeight)
	if err != nil {
		return err
	}
	// this  beacon height is already seen by shard best state
	if s.maxBeaconHeight == s.curView.BeaconHeight {
		s.isOldBeaconHeight = true
	}
	s.beaconBlocks = beaconBlocks

	tempPrivateKey := s.blockchain.config.BlockGen.createTempKeyset()
	blockCreationLeftOver := s.curView.BlockMaxCreateTime.Nanoseconds() - time.Since(s.startTime).Nanoseconds()
	transactionsForNewBlock := make([]metadata.Transaction, 0)
	transactionsForNewBlock, _ = s.blockchain.config.BlockGen.getTransactionForNewBlock(s.curView, &tempPrivateKey, s.curView.ShardID, s.blockchain.GetDatabase(), beaconBlocks, blockCreationLeftOver, s.maxBeaconHeight)
	// build txs with metadata
	transactionsForNewBlock, err = s.blockchain.BuildResponseTransactionFromTxsWithMetadata(s.curView, transactionsForNewBlock, &tempPrivateKey, s.curView.ShardID)
	s.txs = transactionsForNewBlock

	return nil
}

func (s *ShardProcessState) BuildBody() (err error) {
	// curView := s.curView
	//==========Build block body============
	// Get Transaction For new Block
	// Get Cross output coin from other shard && produce cross shard transaction
	crossTransactions := getCrossTxsFromCrossShardBlocks(s.crossShardBlocks)
	Logger.log.Critical("Cross Transaction: ", crossTransactions)

	// process instruction from beacon block
	shardPendingValidator, _, _ := s.blockchain.processInstructionFromBeacon(s.curView, s.beaconBlocks, s.curView.ShardID, newCommitteeChange())
	// Create Instruction
	currentCommitteePubKeys, err := incognitokey.CommitteeKeyListToString(s.curView.ShardCommittee)
	if err != nil {
		return err
	}
	s.instructions, _, _, err = s.blockchain.generateInstruction(s.curView, s.curView.ShardID, s.maxBeaconHeight, s.isOldBeaconHeight, s.beaconBlocks, shardPendingValidator, currentCommitteePubKeys)
	if err != nil {
		return NewBlockChainError(GenerateInstructionError, err)
	}
	if len(s.instructions) != 0 {
		Logger.log.Info("Shard Producer: Instruction", s.instructions)
	}
	s.newBlock.BuildShardBlockBody(s.instructions, crossTransactions, s.txs)
	return
}

func (s *ShardProcessState) BuildHeader() (err error) {

	//======Build Header Essential Data=======
	curView := s.curView
	epoch := s.curView.Epoch
	totalTxsFee := make(map[common.Hash]uint64)
	if (s.curView.ShardHeight+1)%s.blockchain.config.ChainParams.Epoch == 1 {
		epoch = s.curView.Epoch + 1
	}

	if s.version == 1 {
		committee := curView.GetShardCommittee()
		producerPosition := (curView.ShardProposerIdx + s.round) % len(curView.ShardCommittee)
		s.newBlock.Header.Producer, err = committee[producerPosition].ToBase58() // .GetMiningKeyBase58(common.BridgeConsensus)
		if err != nil {
			return err
		}
		s.newBlock.Header.ProducerPubKeyStr, err = committee[producerPosition].ToBase58()
		if err != nil {
			Logger.log.Error(err)
			return NewBlockChainError(ConvertCommitteePubKeyToBase58Error, err)
		}
	} else {
		s.newBlock.Header.Producer = s.proposer
		s.newBlock.Header.ProducerPubKeyStr = s.proposer
	}
	newView := s.newView
	//============Build Header=============
	// Build Root Hash for Header
	merkleRoots := Merkle{}.BuildMerkleTreeStore(s.newBlock.Body.Transactions)
	merkleRoot := &common.Hash{}
	if len(merkleRoots) > 0 {
		merkleRoot = merkleRoots[len(merkleRoots)-1]
	}
	crossTransactionRoot, err := CreateMerkleCrossTransaction(s.newBlock.Body.CrossTransactions)
	if err != nil {
		return err
	}
	txInstructions, err := CreateShardInstructionsFromTransactionAndInstruction(s.newBlock.Body.Transactions, s.blockchain, newView.ShardID)
	if err != nil {
		return err
	}
	totalInstructions := []string{}
	for _, value := range txInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	for _, value := range s.instructions {
		totalInstructions = append(totalInstructions, value...)
	}
	instructionsHash, err := generateHashFromStringArray(totalInstructions)
	if err != nil {
		return NewBlockChainError(InstructionsHashError, err)
	}
	tempShardCommitteePubKeys, err := incognitokey.CommitteeKeyListToString(newView.ShardCommittee)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	committeeRoot, err := generateHashFromStringArray(tempShardCommitteePubKeys)
	if err != nil {
		return NewBlockChainError(CommitteeHashError, err)
	}
	tempShardPendintValidator, err := incognitokey.CommitteeKeyListToString(newView.ShardPendingValidator)
	if err != nil {
		return NewBlockChainError(UnExpectedError, err)
	}
	pendingValidatorRoot, err := generateHashFromStringArray(tempShardPendintValidator)
	if err != nil {
		return NewBlockChainError(PendingValidatorRootError, err)
	}
	stakingTxRoot, err := generateHashFromMapStringString(newView.StakingTx)
	if err != nil {
		return NewBlockChainError(StakingTxHashError, err)
	}
	// Instruction merkle root
	flattenTxInsts, err := FlattenAndConvertStringInst(txInstructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from Tx: %+v", err))
	}
	flattenInsts, err := FlattenAndConvertStringInst(s.instructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from block body: %+v", err))
	}
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be preserved in shardprocess
	instMerkleRoot := GetKeccak256MerkleRoot(insts)
	// shard tx root
	_, shardTxMerkleData := CreateShardTxRoot(s.newBlock.Body.Transactions)

	s.newBlock.Header.Version = s.version
	s.newBlock.Header.Height = s.curView.ShardHeight + 1
	s.newBlock.Header.ConsensusType = s.curView.ConsensusAlgorithm
	s.newBlock.Header.Timestamp = time.Now().Unix()
	s.newBlock.Header.Height = curView.ShardHeight + 1
	s.newBlock.Header.Epoch = epoch
	s.newBlock.Header.ShardID = s.curView.ShardID
	s.newBlock.Header.CrossShardBitMap = CreateCrossShardByteArray(s.newBlock.Body.Transactions, s.curView.ShardID)
	s.newBlock.Header.Round = s.round
	s.newBlock.Header.PreviousBlockHash = curView.BestBlockHash
	s.newBlock.Header.TotalTxsFee = totalTxsFee

	s.newBlock.Header.TxRoot = *merkleRoot
	s.newBlock.Header.ShardTxRoot = shardTxMerkleData[len(shardTxMerkleData)-1]
	s.newBlock.Header.CrossTransactionRoot = *crossTransactionRoot
	s.newBlock.Header.InstructionsRoot = instructionsHash
	s.newBlock.Header.CommitteeRoot = committeeRoot
	s.newBlock.Header.PendingValidatorRoot = pendingValidatorRoot
	s.newBlock.Header.StakingTxRoot = stakingTxRoot
	copy(s.newBlock.Header.InstructionMerkleRoot[:], instMerkleRoot)
	return
}

func getCrossTxsFromCrossShardBlocks(crossShardBlocks map[byte][]*CrossShardBlock) map[byte][]CrossTransaction {
	crossTransactions := make(map[byte][]CrossTransaction)
	for _, crossShardBlock := range crossShardBlocks {
		for _, blk := range crossShardBlock {
			crossTransaction := CrossTransaction{
				OutputCoin:       blk.CrossOutputCoin,
				TokenPrivacyData: blk.CrossTxTokenPrivacyData,
				BlockHash:        *blk.Hash(),
				BlockHeight:      blk.Header.Height,
			}
			crossTransactions[blk.Header.ShardID] = append(crossTransactions[blk.Header.ShardID], crossTransaction)
		}
	}

	return crossTransactions
}
