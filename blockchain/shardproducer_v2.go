package blockchain

import (
	"fmt"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
)

func (blockchain *BlockChain) NewBlockShard_V2(curView *ShardBestState, version int, producer string, round int, startTime int64) (newBlock *ShardBlock, err error) {
	processState := &ShardProcessState{
		curView:             curView,
		newView:             nil,
		blockchain:          blockchain,
		version:             version,
		producer:            producer,
		round:               round,
		newBlock:            NewShardBlock(),
		crossShardBlocks:    make(map[byte][]*CrossShardBlock),
		startTime:           time.Unix(startTime, 0),
		confirmBeaconHeight: 0,
	}
	Logger.log.Infof("PreProduceProcess", time.Now())
	if err := processState.PreProduceProcess(); err != nil {
		Logger.log.Info(err)
		return nil, err
	}

	Logger.log.Infof("BuildBody", time.Now())
	if err := processState.BuildBody(); err != nil {
		return nil, err
	}

	Logger.log.Infof("updateShardBestState", time.Now())
	processState.newView, err = processState.curView.updateShardBestState(blockchain, processState.newBlock, processState.beaconBlocks, newCommitteeChange())
	if err != nil {
		return nil, err
	}

	Logger.log.Infof("BuildHeader", time.Now())
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
	producer   string
	round      int

	//pre process state
	startTime           time.Time
	isOldBeaconHeight   bool
	epoch               int64
	confirmBeaconHeight uint64
	confirmBeaconHash   common.Hash

	newBlock         *ShardBlock
	crossShardBlocks map[byte][]*CrossShardBlock
	beaconBlocks     []*BeaconBlock
	txs              []metadata.Transaction
	instructions     [][]string
}

func (shardFlowState *ShardProcessState) PreProduceProcess() (err error) {
	//get s2b blocks
	for sid, v := range shardFlowState.blockchain.config.Syncker.GetCrossShardBlocksForShardProducer(shardFlowState.curView.ShardID) {
		for _, b := range v {
			shardFlowState.crossShardBlocks[sid] = append(shardFlowState.crossShardBlocks[sid], b.(*CrossShardBlock))
		}
	}
	if shardFlowState.confirmBeaconHeight-shardFlowState.curView.BeaconHeight > MAX_BEACON_BLOCK {
		shardFlowState.confirmBeaconHeight = shardFlowState.curView.BeaconHeight + MAX_BEACON_BLOCK
	}

	//get beacon blocks to confirm
	beaconFinalView := shardFlowState.blockchain.BeaconChain.GetFinalView().(*BeaconBestState)
	shardFlowState.confirmBeaconHeight = beaconFinalView.GetHeight()
	shardFlowState.confirmBeaconHash = *beaconFinalView.BestBlock.Hash()

	//if this beacon block is in different epoch, drop to the final block in current epoch
	epoch := beaconFinalView.BestBlock.Header.Epoch
	if epoch > shardFlowState.curView.Epoch {
		shardFlowState.confirmBeaconHeight = shardFlowState.curView.Epoch * shardFlowState.blockchain.config.ChainParams.Epoch
		newBeaconHashes, err := rawdbv2.GetBeaconBlockHashByIndex(shardFlowState.blockchain.GetDatabase(), shardFlowState.confirmBeaconHeight)
		if err != nil {
			fmt.Println("FAIL GetBeaconBlockHashByIndex", beaconFinalView.BestBlock.Header.Height, epoch, shardFlowState.confirmBeaconHeight, shardFlowState.curView.Epoch, shardFlowState.blockchain.config.ChainParams.Epoch)
			return err
		}
		newBeaconHash := newBeaconHashes[0]
		copy(shardFlowState.confirmBeaconHash[:], newBeaconHash.GetBytes())
		epoch = shardFlowState.curView.Epoch + 1
	}

	//Fetch beacon block from height to confirm beacon block
	shardFlowState.beaconBlocks, err = FetchBeaconBlockFromHeight(shardFlowState.blockchain.GetDatabase(), shardFlowState.curView.BeaconHeight+1, shardFlowState.confirmBeaconHeight)
	if err != nil {
		return err
	}

	// this  beacon height is already seen by shard best state
	if shardFlowState.confirmBeaconHeight == shardFlowState.curView.BeaconHeight {
		shardFlowState.isOldBeaconHeight = true
	}

	Logger.log.Infof("Get Beacon Block With Height %+v, Shard BestState %+v , isOldBeaconHeight: %v", shardFlowState.confirmBeaconHeight, shardFlowState.curView.BeaconHeight, shardFlowState.isOldBeaconHeight)
	return nil
}

func (shardFlowState *ShardProcessState) BuildBody() (err error) {
	// curView := shardFlowState.curView
	//==========Build block body============

	tempPrivateKey := shardFlowState.blockchain.config.BlockGen.createTempKeyset()
	blockCreationLeftOver := shardFlowState.curView.BlockMaxCreateTime.Nanoseconds() - time.Since(shardFlowState.startTime).Nanoseconds()
	transactionsForNewBlock := make([]metadata.Transaction, 0)
	transactionsForNewBlock, _ = shardFlowState.blockchain.config.BlockGen.getTransactionForNewBlock(shardFlowState.curView, &tempPrivateKey, shardFlowState.curView.ShardID, shardFlowState.blockchain.GetDatabase(), shardFlowState.beaconBlocks, blockCreationLeftOver, shardFlowState.confirmBeaconHeight)
	// build txs with metadata
	shardFlowState.txs, err = shardFlowState.blockchain.BuildResponseTransactionFromTxsWithMetadata(shardFlowState.curView, transactionsForNewBlock, &tempPrivateKey, shardFlowState.curView.ShardID)

	// Get Transaction For new Block
	// Get Cross output coin from other shard && produce cross shard transaction
	crossTransactions := getCrossTxsFromCrossShardBlocks(shardFlowState.crossShardBlocks)
	Logger.log.Critical("Cross Transaction: ", crossTransactions)

	// process instruction from beacon block
	shardPendingValidator, _, _ := shardFlowState.blockchain.processInstructionFromBeacon(shardFlowState.curView, shardFlowState.beaconBlocks, shardFlowState.curView.ShardID, newCommitteeChange())
	// Create Instruction
	currentCommitteePubKeys, err := incognitokey.CommitteeKeyListToString(shardFlowState.curView.ShardCommittee)
	if err != nil {
		return err
	}
	shardFlowState.instructions, _, _, err = shardFlowState.blockchain.generateInstruction(shardFlowState.curView, shardFlowState.curView.ShardID, shardFlowState.confirmBeaconHeight, shardFlowState.isOldBeaconHeight, shardFlowState.beaconBlocks, shardPendingValidator, currentCommitteePubKeys)
	if err != nil {
		return NewBlockChainError(GenerateInstructionError, err)
	}
	if len(shardFlowState.instructions) != 0 {
		Logger.log.Info("Shard Producer: Instruction", shardFlowState.instructions)
	}
	shardFlowState.newBlock.BuildShardBlockBody(shardFlowState.instructions, crossTransactions, shardFlowState.txs)
	return
}

func (shardFlowState *ShardProcessState) BuildHeader() (err error) {

	//======Build Header Essential Data=======
	curView := shardFlowState.curView
	epoch := shardFlowState.curView.Epoch
	totalTxsFee := make(map[common.Hash]uint64)
	if (shardFlowState.curView.ShardHeight+1)%shardFlowState.blockchain.config.ChainParams.Epoch == 1 {
		epoch = shardFlowState.curView.Epoch + 1
	}

	fmt.Println("Build header version", shardFlowState.version)
	if shardFlowState.version == 1 {
		committee := curView.GetShardCommittee()
		producerPosition := (curView.ShardProposerIdx + shardFlowState.round) % len(curView.ShardCommittee)
		shardFlowState.newBlock.Header.Producer, err = committee[producerPosition].ToBase58() // .GetMiningKeyBase58(common.BridgeConsensus)
		if err != nil {
			return err
		}
		shardFlowState.newBlock.Header.ProducerPubKeyStr, err = committee[producerPosition].ToBase58()
		if err != nil {
			Logger.log.Error(err)
			return NewBlockChainError(ConvertCommitteePubKeyToBase58Error, err)
		}
	} else {
		shardFlowState.newBlock.Header.Producer = shardFlowState.producer
		shardFlowState.newBlock.Header.ProducerPubKeyStr = shardFlowState.producer
	}

	newView := shardFlowState.newView
	shardFlowState.newBlock.Header.BeaconHeight = shardFlowState.confirmBeaconHeight
	shardFlowState.newBlock.Header.BeaconHash = shardFlowState.confirmBeaconHash

	//============Build Header=============
	// Build Root Hash for Header
	merkleRoots := Merkle{}.BuildMerkleTreeStore(shardFlowState.newBlock.Body.Transactions)
	merkleRoot := &common.Hash{}
	if len(merkleRoots) > 0 {
		merkleRoot = merkleRoots[len(merkleRoots)-1]
	}
	crossTransactionRoot, err := CreateMerkleCrossTransaction(shardFlowState.newBlock.Body.CrossTransactions)
	if err != nil {
		return err
	}
	txInstructions, err := CreateShardInstructionsFromTransactionAndInstruction(shardFlowState.newBlock.Body.Transactions, shardFlowState.blockchain, newView.ShardID)
	if err != nil {
		return err
	}
	totalInstructions := []string{}
	for _, value := range txInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	for _, value := range shardFlowState.instructions {
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
	flattenInsts, err := FlattenAndConvertStringInst(shardFlowState.instructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from block body: %+v", err))
	}
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be preserved in shardprocess
	instMerkleRoot := GetKeccak256MerkleRoot(insts)
	// shard tx root
	_, shardTxMerkleData := CreateShardTxRoot(shardFlowState.newBlock.Body.Transactions)

	shardFlowState.newBlock.Header.Version = shardFlowState.version
	shardFlowState.newBlock.Header.Height = shardFlowState.curView.ShardHeight + 1
	shardFlowState.newBlock.Header.ConsensusType = shardFlowState.curView.ConsensusAlgorithm
	shardFlowState.newBlock.Header.Timestamp = shardFlowState.startTime.Unix()
	shardFlowState.newBlock.Header.Height = curView.ShardHeight + 1
	shardFlowState.newBlock.Header.Epoch = epoch
	shardFlowState.newBlock.Header.ShardID = shardFlowState.curView.ShardID
	shardFlowState.newBlock.Header.CrossShardBitMap = CreateCrossShardByteArray(shardFlowState.newBlock.Body.Transactions, shardFlowState.curView.ShardID)
	shardFlowState.newBlock.Header.Round = shardFlowState.round
	shardFlowState.newBlock.Header.PreviousBlockHash = curView.BestBlockHash
	shardFlowState.newBlock.Header.TotalTxsFee = totalTxsFee

	shardFlowState.newBlock.Header.TxRoot = *merkleRoot
	shardFlowState.newBlock.Header.ShardTxRoot = shardTxMerkleData[len(shardTxMerkleData)-1]
	shardFlowState.newBlock.Header.CrossTransactionRoot = *crossTransactionRoot
	shardFlowState.newBlock.Header.InstructionsRoot = instructionsHash
	shardFlowState.newBlock.Header.CommitteeRoot = committeeRoot
	shardFlowState.newBlock.Header.PendingValidatorRoot = pendingValidatorRoot
	shardFlowState.newBlock.Header.StakingTxRoot = stakingTxRoot
	copy(shardFlowState.newBlock.Header.InstructionMerkleRoot[:], instMerkleRoot)

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
