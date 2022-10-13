package blockchain

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/common/base58"
	"github.com/incognitochain/incognito-chain/config"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/metadata"
	metadataBridge "github.com/incognitochain/incognito-chain/metadata/bridge"
	metadataCommon "github.com/incognitochain/incognito-chain/metadata/common"
	"github.com/incognitochain/incognito-chain/privacy/key"

	"github.com/incognitochain/incognito-chain/pubsub"
)

// VerifyPreSignShardBlock Verify Shard Block Before Signing
// Used for PBFT consensus
// this block doesn't have full information (incomplete block)
func (blockchain *BlockChain) VerifyPreSignShardBlock(
	shardBlock *types.ShardBlock,
	signingCommittees []incognitokey.CommitteePublicKey,
	committees []incognitokey.CommitteePublicKey,
	shardID byte,
) error {
	//get view that block link to
	preHash := shardBlock.Header.PreviousBlockHash
	view := blockchain.ShardChain[int(shardID)].GetViewByHash(preHash)
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if view == nil {
		blockchain.config.Syncker.SyncMissingShardBlock(ctx, "", shardID, preHash)
	}
	var checkShardUntilTimeout = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return errors.New(fmt.Sprintf("ShardBlock %v link to wrong view (%s)", shardBlock.GetHeight(), preHash.String()))
			default:
				view = blockchain.ShardChain[int(shardID)].GetViewByHash(preHash)
				if view != nil {
					return nil
				}
				time.Sleep(time.Second)
			}
		}
	}
	if err := checkShardUntilTimeout(ctx); err != nil {
		return err
	}

	curView := view.(*ShardBestState)
	Logger.log.Infof("SHARD %+v | Verify ShardBlock for signing process %d, with hash %+v", shardID, shardBlock.Header.Height, *shardBlock.Hash())
	shardBestState := NewShardBestState()
	if err := shardBestState.cloneShardBestStateFrom(curView); err != nil {
		return err
	}

	// fetch beacon blocks
	previousBeaconHeight := shardBestState.BeaconHeight
	var checkBeaconUntilTimeout = func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return errors.New("Wait for beacon timeout")
			default:
				if shardBlock.Header.BeaconHeight <= blockchain.BeaconChain.GetFinalView().GetHeight() {
					return nil
				}
				time.Sleep(time.Second)
			}
		}
	}
	ctx, _ = context.WithTimeout(context.Background(), time.Second*5)
	if checkBeaconUntilTimeout(ctx) != nil {
		return errors.New(fmt.Sprintf("Beacon %d not ready, latest is %d", shardBlock.Header.BeaconHeight, blockchain.GetBeaconBestState().BeaconHeight))
	}

	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain, previousBeaconHeight+1, shardBlock.Header.BeaconHeight)
	if err != nil {
		Logger.log.Error("err:", err)
		return err
	}

	//========Get Committes For Processing Block
	if shardBestState.shardCommitteeState.Version() == committeestate.SELF_SWAP_SHARD_VERSION {
		committees = shardBestState.GetShardCommittee()
	}
	//

	//========Verify shardBlock only
	if err := blockchain.verifyPreProcessingShardBlock(
		shardBestState, shardBlock, beaconBlocks,
		shardID, true, signingCommittees); err != nil {
		Logger.log.Error("err:", err)
		return err
	}
	//========Verify shardBlock with previous best state

	// Verify shardBlock with previous best state
	// DO NOT verify agg signature in this function
	if err := shardBestState.verifyBestStateWithShardBlock(blockchain, shardBlock, signingCommittees, committees); err != nil {
		Logger.log.Error("err:", err)
		return err
	}
	//========updateShardBestState best state with new shardBlock
	newBeststate, hashes, _, err := shardBestState.updateShardBestState(blockchain, shardBlock, beaconBlocks, committees)
	if err != nil {
		return err
	}
	//========Post verififcation: verify new beaconstate with corresponding shardBlock
	if err := newBeststate.verifyPostProcessingShardBlock(shardBlock, shardID, hashes); err != nil {
		Logger.log.Error("err:", err)
		return err
	}
	Logger.log.Infof("SHARD %+v | Block %d, with hash %+v is VALID for 🖋 signing", shardID, shardBlock.GetHeight(), shardBlock.Hash().String())
	return nil
}

// InsertShardBlock Insert Shard Block into blockchain
// this block must have full information (complete block)
func (blockchain *BlockChain) InsertShardBlock(shardBlock *types.ShardBlock, shouldValidate bool) error {
	//startTimeInsertShardBlock := time.Now()
	blockHash := shardBlock.Header.Hash()
	blockHeight := shardBlock.Header.Height
	shardID := shardBlock.Header.ShardID
	preHash := shardBlock.Header.PreviousBlockHash

	if config.Config().IsFullValidation {
		shouldValidate = true
	}

	Logger.log.Infof("SHARD %+v | InsertShardBlock %+v with hash %+v Prev hash: %+v", shardID, blockHeight, blockHash, preHash)
	blockchain.ShardChain[int(shardID)].insertLock.Lock()
	defer blockchain.ShardChain[int(shardID)].insertLock.Unlock()
	//check if view is committed
	checkView := blockchain.ShardChain[int(shardID)].GetViewByHash(blockHash)
	if checkView != nil {
		Logger.log.Errorf("SHARD %+v | Block %+v, hash %+v already inserted", shardID, blockHeight, blockHash)
		return nil
	}
	if ok := checkLimitTxAction(false, map[int]int{}, shardBlock); !ok {
		return errors.Errorf("Total txs of this block %v %v shard %v is large than limit", shardBlock.GetHeight(), shardBlock.Hash().String(), shardBlock.GetShardID())
	}
	//get view that block link to
	preView := blockchain.ShardChain[int(shardID)].GetViewByHash(preHash)
	if preView == nil {
		ctx, cancel := context.WithTimeout(context.Background(), DefaultMaxBlockSyncTime)
		defer cancel()
		blockchain.config.Syncker.ReceiveBlock(shardBlock, "", "")
		blockchain.config.Syncker.SyncMissingShardBlock(ctx, "", shardID, preHash)
		return NewBlockChainError(InsertShardBlockError, fmt.Errorf("ShardBlock %v link to wrong view (%s)", blockHeight, preHash.String()))
	}
	curView := NewShardBestState()
	err := curView.cloneShardBestStateFrom(preView.(*ShardBestState))
	if err != nil {
		return NewBlockChainError(InsertShardBlockError, fmt.Errorf("Cannot clone shard view"))
	}

	if blockHeight != curView.ShardHeight+1 {
		return NewBlockChainError(InsertShardBlockError, fmt.Errorf("Not expected height, current view height %+v, incomminÏg block height %+v", curView.ShardHeight, blockHeight))
	}

	// fetch beacon blocks
	previousBeaconHeight := curView.BeaconHeight
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain, previousBeaconHeight+1, shardBlock.Header.BeaconHeight)
	if err != nil {

		return NewBlockChainError(FetchBeaconBlocksError, err)
	}
	signingCommittees := []incognitokey.CommitteePublicKey{}
	committees := []incognitokey.CommitteePublicKey{}
	committees, signingCommittees, err = curView.getSigningCommittees(shardBlock, blockchain)

	if err != nil {
		return err
	}
	if curView.CommitteeStateVersion() != committeestate.SELF_SWAP_SHARD_VERSION {
		beaconHeight := curView.BeaconHeight
		for _, v := range beaconBlocks {
			if v.GetHeight() >= beaconHeight {
				beaconHeight = v.GetHeight()
			}
		}
		if beaconHeight <= curView.BeaconHeight {
			Logger.log.Info("Waiting For Beacon Produce Block beaconHeight %+v curView.BeaconHeight %+v",
				beaconHeight, curView.BeaconHeight)
			return NewBlockChainError(WrongBlockHeightError, errors.New("Waiting For Beacon Produce Block"))
		}
		if err := curView.verifyCommitteeFromBlock(blockchain, shardBlock, committees); err != nil {
			return err
		}
	}
	committeesStr, _ := incognitokey.CommitteeKeyListToString(signingCommittees)
	if err := blockchain.config.ConsensusEngine.ValidateBlockCommitteSig(shardBlock, signingCommittees); err != nil {
		Logger.log.Errorf("Validate block %v shard %v with committee %v return error %v", shardBlock.GetHeight(), shardBlock.GetShardID(), committeesStr, err)
		return err
	}

	if shouldValidate {
		Logger.log.Infof("SHARD %+v | Verify Pre Processing, block height %+v with hash %+v", shardID, blockHeight, blockHash)
		if err := blockchain.verifyPreProcessingShardBlock(curView, shardBlock, beaconBlocks, shardID, false, signingCommittees); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Pre Processing, block height %+v with hash %+v", shardID, blockHeight, blockHash)
	}

	if shouldValidate {
		// Verify block with previous best state
		Logger.log.Infof("SHARD %+v | Verify BestState With Shard Block, block height %+v with hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
		if err := curView.verifyBestStateWithShardBlock(blockchain, shardBlock, signingCommittees, committees); err != nil {
			return err
		}
	} else {
		Logger.log.Debugf("SHARD %+v | SKIP Verify Best State With Shard Block, Shard Block Height %+v with hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
	}
	Logger.log.Debugf("SHARD %+v | Update ShardBestState, block height %+v with hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)

	//only validate all tx if we have env variable FULL_VALIDATION = 1
	if config.Config().IsFullValidation {
		Logger.log.Infof("SHARD %+v | Verify Transaction From Block 🔍 %+v, total %v txs, block height %+v with hash %+v, beaconHash %+v",
			shardID, blockHeight, len(shardBlock.Body.Transactions), shardBlock.Header.Height, shardBlock.Hash().String(), shardBlock.Header.BeaconHash)

		st := time.Now()
		if err := blockchain.verifyTransactionFromNewBlock(shardID, shardBlock.Body.Transactions, curView.BestBeaconHash, curView); err != nil {
			return NewBlockChainError(TransactionFromNewBlockError, err)
		}
		if len(shardBlock.Body.Transactions) > 0 {
			Logger.log.Infof("SHARD %+v | Validate %v txs of block %v cost %v", shardID, len(shardBlock.Body.Transactions), shardBlock.GetHeight(), time.Since(st))
		}
	}

	newBestState, hashes, committeeChange, err := curView.updateShardBestState(blockchain, shardBlock, beaconBlocks, committees)
	if err != nil {
		return err
	}

	//========Post verification: verify new beaconstate with corresponding block
	if shouldValidate {
		Logger.log.Debugf("SHARD %+v | Verify Post Processing, block height %+v with hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash)
		if err = newBestState.verifyPostProcessingShardBlock(shardBlock, shardID, hashes); err != nil {
			return err
		}
	} else {
		Logger.log.Infof("SHARD %+v | SKIP Verify Post Processing, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	}

	Logger.log.Infof("SHARD %+v | Update Beacon Instruction, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	err = blockchain.processSalaryInstructions(curView, newBestState.rewardStateDB, newBestState.consensusStateDB, beaconBlocks, newBestState.BeaconHeight, shardID)
	if err != nil {
		return err
	}

	Logger.log.Infof("SHARD %+v | Store New Shard Block And Update Data, block height %+v with hash %+v \n", shardID, blockHeight, blockHash)
	err = blockchain.processStoreShardBlock(newBestState, shardBlock, committeeChange, beaconBlocks)
	if err != nil {
		return err
	}

	blockchain.config.Server.InsertNewShardView(newBestState)
	blockchain.removeOldDataAfterProcessingShardBlock(shardBlock, shardID)
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.NewShardblockTopic, shardBlock))
	go blockchain.config.PubSubManager.PublishMessage(pubsub.NewMessage(pubsub.ShardBeststateTopic, newBestState))
	Logger.log.Infof("SHARD %+v | Finish Insert new block %d, with hash %+v 🔗, "+
		"Found 🔎 %+v transactions, "+
		"%+v cross shard transactions, "+
		"%+v instruction",
		shardBlock.Header.ShardID, shardBlock.Header.Height, blockHash,
		len(shardBlock.Body.Transactions), len(shardBlock.Body.CrossTransactions), len(shardBlock.Body.Instructions))
	return nil
}

// verifyPreProcessingShardBlock DOES NOT verify new block with best state
// DO NOT USE THIS with GENESIS BLOCK
// Verification condition:
//   - Producer Address is not empty
//   - ShardID: of received block same shardID
//   - Version: shard block version is one of pre-defined versions
//   - Parent (previous) block must be found in database ( current block point to an exist block in database )
//   - Height: parent block height + 1
//   - epoch: blockHeight % epoch ? Parent epoch + 1 : Current epoch
//   - Timestamp: block timestamp must be greater than previous block timestamp
//   - TransactionRoot: rebuild transaction root from txs in block and compare with transaction root in header
//   - ShardTxRoot: rebuild shard transaction root from txs in block and compare with shard transaction root in header
//   - CrossOutputCoinRoot: rebuild cross shard output root from cross output coin in block and compare with cross shard output coin
//   - cross output coin must be re-created (from cross shard block) if verify block for signing
//   - InstructionRoot: rebuild instruction root from instructions and txs in block and compare with instruction root in header
//   - instructions must be re-created (from beacon block and instruction) if verify block for signing
//   - InstructionMerkleRoot: rebuild instruction root from instructions and txs in block and compare with instruction root in header
//   - TotalTxFee: calculate tx token fee from all transaction in block then compare with header
//   - CrossShars: Verify Cross Shard Bitmap
//   - BeaconHeight: fetch beacon hash using beacon height in current shard block
//   - BeaconHash: compare beacon hash in database with beacon hash in shard block
//   - Verify swap instruction
//   - Validate transaction created from miner via instruction
//   - Validate Response Transaction From Transaction with Metadata
//   - ALL Transaction in block: see in verifyTransactionFromNewBlock
func (blockchain *BlockChain) verifyPreProcessingShardBlock(curView *ShardBestState,
	shardBlock *types.ShardBlock, beaconBlocks []*types.BeaconBlock,
	shardID byte, isPreSign bool, committees []incognitokey.CommitteePublicKey) error {
	startTimeVerifyPreProcessingShardBlock := time.Now()
	Logger.log.Debugf("SHARD %+v | Begin verifyPreProcessingShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash().String())
	if shardBlock.Header.ShardID != shardID {
		return NewBlockChainError(WrongShardIDError, fmt.Errorf("Expect receive shardBlock from Shard ID %+v but get %+v", shardID, shardBlock.Header.ShardID))
	}

	beaconHeight := curView.BeaconHeight
	for _, v := range beaconBlocks {
		if v.GetHeight() >= beaconHeight {
			beaconHeight = v.GetHeight()
		}
	}
	if curView.CommitteeStateVersion() != committeestate.SELF_SWAP_SHARD_VERSION {
		if beaconHeight <= curView.BeaconHeight {
			Logger.log.Info("Waiting For Beacon Produce Block beaconHeight %+v curView.BeaconHeight %+v",
				beaconHeight, curView.BeaconHeight)
			return NewBlockChainError(WrongBlockHeightError, errors.New("Waiting For Beacon Produce Block"))
		}
	}

	if shardBlock.Header.Height > curView.ShardHeight+1 {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Expect shardBlock height %+v but get %+v", curView.ShardHeight+1, shardBlock.Header.Height))
	}
	// Verify parent hash exist or not
	previousBlockHash := shardBlock.Header.PreviousBlockHash
	previousShardBlockData, err := rawdbv2.GetShardBlockByHash(blockchain.GetShardChainDatabase(shardID), previousBlockHash)
	if err != nil {
		return NewBlockChainError(FetchPreviousBlockError, err)
	}

	previousShardBlock := types.ShardBlock{}
	err = json.Unmarshal(previousShardBlockData, &previousShardBlock)
	if err != nil {
		return NewBlockChainError(UnmashallJsonShardBlockError, err)
	}
	// Verify shardBlock height with parent shardBlock
	if previousShardBlock.Header.Height+1 != shardBlock.Header.Height {
		return NewBlockChainError(WrongBlockHeightError, fmt.Errorf("Expect receive shardBlock height %+v but get %+v", previousShardBlock.Header.Height+1, shardBlock.Header.Height))
	}
	// Verify timestamp with parent shardBlock
	if shardBlock.Header.Timestamp <= previousShardBlock.Header.Timestamp {
		return NewBlockChainError(WrongTimestampError, fmt.Errorf("Expect receive shardBlock has timestamp must be greater than %+v but get %+v", previousShardBlock.Header.Timestamp, shardBlock.Header.Timestamp))
	}

	if shardBlock.GetVersion() >= types.MULTI_VIEW_VERSION && curView.BestBlock.GetProposeTime() > 0 && curView.CalculateTimeSlot(shardBlock.Header.ProposeTime) <= curView.CalculateTimeSlot(curView.BestBlock.GetProposeTime()) {
		return NewBlockChainError(WrongTimeslotError, fmt.Errorf("Propose timeslot must be greater than last propose timeslot (but get %v <= %v) ", curView.CalculateTimeSlot(shardBlock.Header.ProposeTime), curView.CalculateTimeSlot(curView.BestBlock.GetProposeTime())))
	}

	// Verify transaction root
	txMerkleTree := types.Merkle{}.BuildMerkleTreeStore(shardBlock.Body.Transactions)
	txRoot := &common.Hash{}
	if len(txMerkleTree) > 0 {
		txRoot = txMerkleTree[len(txMerkleTree)-1]
	}
	if !bytes.Equal(shardBlock.Header.TxRoot.GetBytes(), txRoot.GetBytes()) &&
		(config.Param().Net == config.LocalNet || config.Param().Net != config.TestnetNet || (shardBlock.Header.Height != 487260 && shardBlock.Header.Height != 487261 && shardBlock.Header.Height != 494144)) {
		return NewBlockChainError(TransactionRootHashError, fmt.Errorf("Expect transaction root hash %+v but get %+v", shardBlock.Header.TxRoot.String(), txRoot.String()))
	}

	// Verify ShardTx Root
	_, shardTxMerkleData := types.CreateShardTxRoot(shardBlock.Body.Transactions)
	shardTxRoot := shardTxMerkleData[len(shardTxMerkleData)-1]
	if !bytes.Equal(shardBlock.Header.ShardTxRoot.GetBytes(), shardTxRoot.GetBytes()) {
		return NewBlockChainError(ShardTransactionRootHashError, fmt.Errorf("Expect shard transaction root hash %+v but get %+v", shardBlock.Header.ShardTxRoot, shardTxRoot))
	}
	// Verify crossTransaction coin
	if !VerifyMerkleCrossTransaction(shardBlock.Body.CrossTransactions, shardBlock.Header.CrossTransactionRoot) {
		return NewBlockChainError(CrossShardTransactionRootHashError, fmt.Errorf("Expect cross shard transaction root hash %+v", shardBlock.Header.CrossTransactionRoot))
	}
	// Verify Action
	txInstructions, _, err := CreateShardInstructionsFromTransactionAndInstruction(shardBlock.Body.Transactions, blockchain, shardID, shardBlock.Header.Height, shardBlock.Header.BeaconHeight, false)
	if err != nil {
		Logger.log.Error(err)
		return NewBlockChainError(ShardIntructionFromTransactionAndInstructionError, err)
	}

	if !isPreSign {
		totalInstructions := []string{}
		for _, value := range txInstructions {
			totalInstructions = append(totalInstructions, value...)
		}
		for _, value := range shardBlock.Body.Instructions {
			totalInstructions = append(totalInstructions, value...)
		}
		if hash, ok := verifyHashFromStringArray(totalInstructions, shardBlock.Header.InstructionsRoot); !ok {
			return NewBlockChainError(
				InstructionsHashError,
				fmt.Errorf("Expect instruction hash to be %+v but get %+v at block %+v hash %+v", hash, shardBlock.Header.InstructionsRoot, shardBlock.Header.Height, shardBlock.Hash().String()),
				//fmt.Errorf("Expect instruction hash to be %+v but get %+v at block %+v hash %+v", shardBlock.Header.InstructionsRoot, hash, shardBlock.Header.Height, shardBlock.Hash().String()),
			)
		}
	}

	totalTxsFee := curView.shardCommitteeState.BuildTotalTxsFeeFromTxs(shardBlock.Body.Transactions)

	tokenIDsfromTxs := make([]common.Hash, 0)
	for tokenID := range totalTxsFee {
		tokenIDsfromTxs = append(tokenIDsfromTxs, tokenID)
	}
	sort.Slice(tokenIDsfromTxs, func(i int, j int) bool {
		res, _ := tokenIDsfromTxs[i].Cmp(&tokenIDsfromTxs[j])
		return res == -1
	})
	tokenIDsfromBlock := make([]common.Hash, 0)
	for tokenID := range shardBlock.Header.TotalTxsFee {
		tokenIDsfromBlock = append(tokenIDsfromBlock, tokenID)
	}
	sort.Slice(tokenIDsfromBlock, func(i int, j int) bool {
		res, _ := tokenIDsfromBlock[i].Cmp(&tokenIDsfromBlock[j])
		return res == -1
	})
	if len(tokenIDsfromTxs) != len(tokenIDsfromBlock) {
		return NewBlockChainError(WrongBlockTotalFeeError, fmt.Errorf("Expect Total Fee to be equal, From Txs %+v, From Block %+v", len(tokenIDsfromTxs), len(tokenIDsfromBlock)))
	}
	for i, tokenID := range tokenIDsfromTxs {
		if !tokenIDsfromTxs[i].IsEqual(&tokenIDsfromBlock[i]) {
			return NewBlockChainError(WrongBlockTotalFeeError, fmt.Errorf("Expect Total Fee to be equal, From Txs %+v, From Block %+v", tokenIDsfromTxs[i], tokenIDsfromBlock[i]))
		}
		if totalTxsFee[tokenID] != shardBlock.Header.TotalTxsFee[tokenID] {
			return NewBlockChainError(WrongBlockTotalFeeError, fmt.Errorf("Expect Total Fee to be equal, From Txs %+v, From Block Header %+v", totalTxsFee[tokenID], shardBlock.Header.TotalTxsFee[tokenID]))
		}
	}
	// Verify Cross Shards
	crossShards, err := CreateCrossShardByteArray(shardBlock.Body.Transactions, shardID)
	if err != nil {
		return err
	}
	if len(crossShards) != len(shardBlock.Header.CrossShardBitMap) {
		return NewBlockChainError(CrossShardBitMapError, fmt.Errorf("Expect number of cross shardID is %+v but get %+v", len(shardBlock.Header.CrossShardBitMap), len(crossShards)))
	}
	for index := range crossShards {
		if crossShards[index] != shardBlock.Header.CrossShardBitMap[index] {
			return NewBlockChainError(CrossShardBitMapError, fmt.Errorf("Expect Cross Shard Bitmap of shardID %+v is %+v but get %+v", index, shardBlock.Header.CrossShardBitMap[index], crossShards[index]))
		}
	}
	// Check if InstructionMerkleRoot is the root of merkle tree containing all instructions in this shardBlock
	flattenTxInsts, err := FlattenAndConvertStringInst(txInstructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from Tx: %+v", err))
	}
	flattenInsts, err := FlattenAndConvertStringInst(shardBlock.Body.Instructions)
	if err != nil {
		return NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from shardBlock body: %+v", err))
	}
	insts := append(flattenTxInsts, flattenInsts...) // Order of instructions must be the same as when creating new shard shardBlock
	root := types.GetKeccak256MerkleRoot(insts)
	if !bytes.Equal(root, shardBlock.Header.InstructionMerkleRoot[:]) {
		return NewBlockChainError(InstructionMerkleRootError, fmt.Errorf("Expect transaction merkle root to be %+v but get %+v", shardBlock.Header.InstructionMerkleRoot, string(root)))
	}
	//Get beacon hash by height in db
	//If hash not found then fail to verify
	beaconHash, err := rawdbv2.GetFinalizedBeaconBlockHashByIndex(blockchain.GetBeaconChainDatabase(), shardBlock.Header.BeaconHeight)
	if err != nil {
		return NewBlockChainError(FetchBeaconBlockHashError, err)
	}

	//Hash in db must be equal to hash in shard shardBlock
	newHash, err := common.Hash{}.NewHash(shardBlock.Header.BeaconHash.GetBytes())
	if err != nil {
		return NewBlockChainError(HashError, err)
	}

	if !newHash.IsEqual(beaconHash) {
		return NewBlockChainError(BeaconBlockNotCompatibleError, fmt.Errorf("Expect beacon shardBlock hash to be %+v but get %+v", beaconHash.String(), newHash.String()))
	}

	// Verify response transactions
	instsForValidations := [][]string{}
	instsForValidations = append(instsForValidations, shardBlock.Body.Instructions...)
	for _, beaconBlock := range beaconBlocks {
		instsForValidations = append(instsForValidations, beaconBlock.Body.Instructions...)
	}
	invalidTxs, err := blockchain.verifyMinerCreatedTxBeforeGettingInBlock(instsForValidations, shardBlock.Body.Transactions, shardID)
	if err != nil {
		return NewBlockChainError(TransactionCreatedByMinerError, err)
	}
	if len(invalidTxs) > 0 {
		return NewBlockChainError(TransactionCreatedByMinerError, fmt.Errorf("There are %d invalid txs", len(invalidTxs)))
	}
	err = blockchain.ValidateResponseTransactionFromTxsWithMetadata(shardBlock)
	if err != nil {
		return NewBlockChainError(ResponsedTransactionWithMetadataError, err)
	}
	err = blockchain.ValidateResponseTransactionFromBeaconInstructions(curView, shardBlock, beaconBlocks, shardID)
	if err != nil {
		return NewBlockChainError(ResponsedTransactionWithMetadataError, err)
	}
	shardVerifyPreprocesingTimer.UpdateSince(startTimeVerifyPreProcessingShardBlock)
	// Get cross shard shardBlock from pool
	if isPreSign {
		err := blockchain.verifyPreProcessingShardBlockForSigning(curView, shardBlock, beaconBlocks, txInstructions, shardID, committees)
		if err != nil {
			return err
		}
	}
	Logger.log.Debugf("SHARD %+v | Finish verifyPreProcessingShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash().String())
	return err
}

// VerifyPreProcessingShardBlockForSigning verify shard block before a validator signs new shard block
//   - Verify Transactions In New Block
//   - Generate Instruction (from beacon), create instruction root and compare instruction root with instruction root in header
//   - Get Cross Output Data from cross shard block (shard pool) and verify cross transaction hash
//   - Get Cross Tx Custom Token from cross shard block (shard pool) then verify
func (blockchain *BlockChain) verifyPreProcessingShardBlockForSigning(curView *ShardBestState,
	shardBlock *types.ShardBlock, beaconBlocks []*types.BeaconBlock,
	txInstructions [][]string, shardID byte, committees []incognitokey.CommitteePublicKey) error {
	var err error
	var isOldBeaconHeight = false
	instructions := [][]string{}
	beaconInstructions := [][]string{}
	startTimeVerifyPreProcessingShardBlockForSigning := time.Now()
	// Verify Transaction
	//get beacon height from shard block
	// beaconHeight := shardBlock.Header.BeaconHeight
	Logger.log.Infof("SHARD %+v | Verify Transaction From Block 🔍 %+v, total %v txs, block height %+v with hash %+v, beaconHash %+v", shardID, len(shardBlock.Body.Transactions), shardBlock.Header.Height, shardBlock.Hash().String(), shardBlock.Header.BeaconHash)
	st := time.Now()
	if err := blockchain.verifyTransactionFromNewBlock(shardID, shardBlock.Body.Transactions, curView.BestBeaconHash, curView); err != nil {
		return NewBlockChainError(TransactionFromNewBlockError, err)
	}
	if len(shardBlock.Body.Transactions) > 0 {
		Logger.log.Infof("SHARD %+v | Validate %v txs of block %v cost %v", shardID, len(shardBlock.Body.Transactions), shardBlock.GetHeight(), time.Since(st))
	}
	// Verify Instruction

	shardCommittee, err := incognitokey.CommitteeKeyListToString(committees)
	if err != nil {
		return err
	}

	currentPendingValidators := curView.GetShardPendingValidator()
	shardPendingValidatorStr, _ := incognitokey.CommitteeKeyListToString(currentPendingValidators)

	beaconInstructions, _, err = blockchain.extractInstructionsFromBeacon(beaconBlocks, curView.ShardID)
	if err != nil {
		return err
	}

	/*env := committeestate.*/
	//NewShardEnvBuilder().
	//BuildShardID(curView.ShardID).
	//BuildBeaconInstructions(beaconInstructions).
	//BuildNumberOfFixedBlockValidators(blockchain.config.ChainParams.NumberOfShardFixedBlockValidators).
	//BuildShardHeight(curView.ShardHeight).
	//Build()

	//committeeChange, err := curView.shardCommitteeState.ProcessInstructionFromBeacon(env)
	//if err != nil {
	//return err
	/*}*/

	if curView.BeaconHeight == shardBlock.Header.BeaconHeight {
		isOldBeaconHeight = true
	}

	if curView.shardCommitteeState.Version() == committeestate.SELF_SWAP_SHARD_VERSION {
		env := committeestate.NewShardCommitteeStateEnvironmentForAssignInstruction(
			beaconInstructions,
			curView.ShardID,
			curView.NumberOfFixedShardBlockValidator,
			shardBlock.Header.Height,
		)

		assignInstructionProcessor := curView.shardCommitteeState.(committeestate.AssignInstructionProcessor)
		addedSubstitutes := assignInstructionProcessor.ProcessAssignInstructions(env)

		currentPendingValidators, err = updateCommitteesWithAddedAndRemovedListValidator(currentPendingValidators,
			addedSubstitutes)
		if err != nil {
			return NewBlockChainError(ProcessInstructionFromBeaconError, err)
		}

		shardPendingValidatorStr, _ = incognitokey.CommitteeKeyListToString(currentPendingValidators)

	}

	instructions, _, shardCommittee, err = blockchain.generateInstruction(curView, shardID,
		shardBlock.Header.BeaconHeight, isOldBeaconHeight, beaconBlocks,
		shardPendingValidatorStr, shardCommittee)
	if err != nil {
		return NewBlockChainError(GenerateInstructionError, err)
	}
	///
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
		return NewBlockChainError(InstructionsHashError, fmt.Errorf("Expect instruction hash to be %+v but %+v", shardBlock.Header.InstructionsRoot, hash))
	}
	toShard := shardID
	var toShardAllCrossShardBlock = make(map[byte][]*types.CrossShardBlock)

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
		return NewBlockChainError(CrossShardBlockError, fmt.Errorf("Unable to get required crossShard blocks from pool in time"))
	}
	for sid, v := range crossShardBlksFromPool {
		heightList := make([]uint64, len(v))
		for i, b := range v {
			toShardAllCrossShardBlock[sid] = append(toShardAllCrossShardBlock[sid], b.(*types.CrossShardBlock))
			heightList[i] = b.(*types.CrossShardBlock).GetHeight()
		}
		Logger.log.Infof("Shard %v, GetCrossShardBlocksForShardValidator from shard %v: %v", toShard, sid, heightList)

	}
	for fromShard, crossTransactions := range shardBlock.Body.CrossTransactions {
		toShardCrossShardBlocks := toShardAllCrossShardBlock[fromShard]
		// if !ok {

		// 	//blockchain.Synker.SyncBlkCrossShard(false, false, []common.Hash{}, heights, fromShard, shardID, "")
		// 	return NewBlockChainError(CrossShardBlockError, fmt.Errorf("Cross Shard Block From Shard %+v Not Found in Pool", fromShard))
		// }
		sort.SliceStable(toShardCrossShardBlocks[:], func(i, j int) bool {
			return toShardCrossShardBlocks[i].Header.Height < toShardCrossShardBlocks[j].Header.Height
		})
		//startHeight := blockchain.GetBestStateShard(toShard).BestCrossShard[fromShard]
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
	shardVerifyPreprocesingForPreSignTimer.UpdateSince(startTimeVerifyPreProcessingShardBlockForSigning)
	return nil
}

// verifyBestStateWithShardBlock will verify the validation of a block with some best state in cache or current best state
// Get beacon state of this block
// For example, new blockHeight is 91 then beacon state of this block must have height 90
// OR new block has previous has is beacon best block hash
//   - New Shard Block has parent (previous) hash is current shard state best block hash (compatible with current beststate)
//   - New Shard Block Height must be compatible with best shard state
//   - New Shard Block has beacon must higher or equal to beacon height of shard best state
func (shardBestState *ShardBestState) verifyBestStateWithShardBlock(blockchain *BlockChain,
	shardBlock *types.ShardBlock,
	signingCommittees []incognitokey.CommitteePublicKey,
	committees []incognitokey.CommitteePublicKey) error {
	startTimeVerifyBestStateWithShardBlock := time.Now()
	Logger.log.Debugf("SHARD %+v | Begin VerifyBestStateWithShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash().String())
	//verify producer via index

	produceTimeSlot := shardBestState.CalculateTimeSlot(shardBlock.GetProduceTime())
	proposeTimeSlot := shardBestState.CalculateTimeSlot(shardBlock.GetProposeTime())
	if err := blockchain.config.ConsensusEngine.ValidateProducerPosition(shardBlock,
		shardBestState.ShardProposerIdx, committees, shardBestState.GetProposerLength(), produceTimeSlot, proposeTimeSlot); err != nil {
		return err
	}
	if err := blockchain.config.ConsensusEngine.ValidateProducerSig(shardBlock, common.BlsConsensus); err != nil {
		return err
	}

	if shardBestState.shardCommitteeState.Version() != committeestate.SELF_SWAP_SHARD_VERSION {
		if err := shardBestState.verifyCommitteeFromBlock(blockchain, shardBlock, committees); err != nil {
			return err
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
	shardVerifyWithBestStateTimer.UpdateSince(startTimeVerifyBestStateWithShardBlock)
	Logger.log.Debugf("SHARD %+v | Finish VerifyBestStateWithShardBlock Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash().String())
	return nil
}

// updateShardBestState beststate with new shard block:
//   - New Previous Shard BlockHash
//   - New BestShardBlockHash
//   - New BestBeaconHash
//   - New Best Shard Block
//   - New Best Shard Height
//   - New Beacon Height
//   - ShardProposerIdx of new shard block
//   - Execute stake instruction, store staking transaction (if exist)
//   - Execute assign instruction, add new pending validator (if exist)
//   - Execute swap instruction, swap pending validator and committee (if exist)
func (oldBestState *ShardBestState) updateShardBestState(blockchain *BlockChain,
	shardBlock *types.ShardBlock,
	beaconBlocks []*types.BeaconBlock,
	committees []incognitokey.CommitteePublicKey) (
	*ShardBestState, *committeestate.ShardCommitteeStateHash, *committeestate.CommitteeChange, error) {
	var (
		err error
	)

	startTimeUpdateShardBestState := time.Now()
	Logger.log.Debugf("SHARD %+v | Begin update Beststate with new Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash().String())
	shardBestState := NewShardBestState()
	if err := shardBestState.cloneShardBestStateFrom(oldBestState); err != nil {
		return nil, nil, nil, err
	}
	shardBestState.BestBlockHash = *shardBlock.Hash()
	shardBestState.BestBeaconHash = shardBlock.Header.BeaconHash
	shardBestState.BestBlock = shardBlock
	shardBestState.BestBlockHash = *shardBlock.Hash()
	shardBestState.ShardHeight = shardBlock.Header.Height
	shardBestState.Epoch = shardBlock.Header.Epoch
	shardBestState.BeaconHeight = shardBlock.Header.BeaconHeight
	shardBestState.TotalTxns += uint64(len(shardBlock.Body.Transactions))
	shardBestState.NumTxns = uint64(len(shardBlock.Body.Transactions))
	shardBestState.MaxTxsPerBlockRemainder = int64(config.Param().TransactionInBlockParam.Lower)
	if shardBlock.Header.Height == 1 {
		shardBestState.ShardProposerIdx = 0
	} else {
		for i, v := range committees {
			b58Str, _ := v.ToBase58()
			if b58Str == shardBlock.Header.Producer {
				shardBestState.ShardProposerIdx = i
				break
			}
		}
	}
	//updateShardBestState best cross shard
	for shardID, crossShardBlock := range shardBlock.Body.CrossTransactions {
		shardBestState.BestCrossShard[shardID] = crossShardBlock[len(crossShardBlock)-1].BlockHeight
	}
	temp := 0
	for _, tx := range shardBlock.Body.Transactions {
		//detect transaction that's not salary
		if !tx.IsSalaryTx() {
			temp++
		}
	}
	if shardBlock.Header.Version >= types.INSTANT_FINALITY_VERSION_V2 {
		maxTxsReminder := oldBestState.MaxTxsPerBlockRemainder - int64(len(shardBlock.Body.Transactions))
		if maxTxsReminder > 0 {
			if shardBestState.MaxTxsPerBlockRemainder+maxTxsReminder >= 10000 {
				shardBestState.MaxTxsPerBlockRemainder = 10000
			} else {
				shardBestState.MaxTxsPerBlockRemainder += maxTxsReminder
			}
		}
	}

	shardBestState.TotalTxnsExcludeSalary += uint64(temp)

	//update trigger feature
	var beaconBlockContainTriggerFeature *types.BeaconBlock
	for _, beaconBlock := range beaconBlocks {
		for _, inst := range beaconBlock.Body.Instructions {
			if inst[0] == instruction.ENABLE_FEATURE {
				enableFeatures, err := instruction.ValidateAndImportEnableFeatureInstructionFromString(inst)
				if err != nil {
					return nil, nil, nil, err
				}
				if shardBestState.TriggeredFeature == nil {
					shardBestState.TriggeredFeature = make(map[string]uint64)
				}
				for _, feature := range enableFeatures.Features {
					if common.IndexOfStr(feature, shardBestState.getUntriggerFeature()) != -1 {
						shardBestState.TriggeredFeature[feature] = shardBlock.GetHeight()
						beaconBlockContainTriggerFeature = beaconBlock
					} else { //cannot find feature in untrigger feature lists(not have or already trigger cases -> unexpected condition)
						Logger.log.Warnf("This source code does not contain new feature or already trigger the feature! Feature:" + feature)
						return nil, nil, nil, NewBlockChainError(OutdatedCodeError, errors.New("Expected having feature "+feature))
					}

				}
			}
		}
	}

	//checkpoint timeslot
	curTS := shardBestState.CalculateTimeSlot(shardBlock.GetProposeTime())
	for feature, _ := range config.Param().BlockTimeParam {
		if triggerHeight, ok := shardBestState.TriggeredFeature[feature]; ok {
			if triggerHeight == shardBlock.GetHeight() {
				//align shard timeslot to be middle of beacon timeslot
				alignTime := beaconBlockContainTriggerFeature.GetProposeTime() + (config.Param().BlockTimeParam[feature] / 2)
				for alignTime < shardBlock.GetProposeTime() {
					alignTime += config.Param().BlockTimeParam[feature]
				}
				//endtime is current propose time
				//starttime is new align time
				Logger.log.Infof("Align shard timeslot: end in %v, start from %v", shardBlock.GetProposeTime(), alignTime)
				shardBestState.TSManager.updateNewAnchor(shardBlock.GetProposeTime(), alignTime, curTS, int(config.Param().BlockTimeParam[feature]), feature, triggerHeight)
			}
		}
	}
	shardBestState.TSManager.updateCurrentInfo(shardBlock.GetVersion(), curTS, shardBlock.GetProposeTime())

	//update committee
	beaconInstructions, _, err := blockchain.
		extractInstructionsFromBeacon(beaconBlocks, shardBestState.ShardID)
	if err != nil {
		return nil, nil, nil, err
	}

	tempCommittees, _ := incognitokey.CommitteeKeyListToString(committees)
	env := shardBestState.NewShardCommitteeStateEnvironmentWithValue(
		shardBlock,
		blockchain,
		beaconInstructions,
		tempCommittees,
		common.Hash{},
	)

	hashes, committeeChange, err := shardBestState.shardCommitteeState.UpdateCommitteeState(env)
	if err != nil {
		return nil, nil, nil, NewBlockChainError(UpgradeShardCommitteeStateError, err)
	}

	confirmedBeaconHeightCommittee, err := getConfirmedCommitteeHeightFromBeacon(blockchain, shardBlock)
	if err != nil {
		return nil, nil, nil, NewBlockChainError(UpgradeShardCommitteeStateError, fmt.Errorf("get confirmed committee height from beacon, %+v", err))
	}

	newMaxCommitteeSize := GetMaxCommitteeSize(shardBestState.MaxShardCommitteeSize,
		shardBestState.TriggeredFeature, shardBlock.GetHeight())
	if newMaxCommitteeSize != shardBestState.MaxShardCommitteeSize {
		Logger.log.Infof("SHARD %+v | Shard Height %+v, hash %+v, Confirmed Beacon Height %+v, found new max committee size %+v",
			shardBlock.Header.ShardID, shardBlock.Header.Height, *shardBlock.Hash(), confirmedBeaconHeightCommittee, newMaxCommitteeSize)
		shardBestState.MaxShardCommitteeSize = newMaxCommitteeSize
	}

	shardUpdateBestStateTimer.UpdateSince(startTimeUpdateShardBestState)
	Logger.log.Debugf("SHARD %+v | Finish update Beststate with new Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, *shardBlock.Hash())
	return shardBestState, hashes, committeeChange, nil
}

func (shardBestState *ShardBestState) initShardBestState(
	blockchain *BlockChain,
	db incdb.Database,
	genesisShardBlock *types.ShardBlock,
	genesisBeaconBlock *types.BeaconBlock,
) error {

	shardBestState.BestBeaconHash = genesisBeaconBlock.Header.Hash()
	shardBestState.BestBlock = genesisShardBlock
	shardBestState.BestBlockHash = *genesisShardBlock.Hash()
	shardBestState.ShardHeight = genesisShardBlock.Header.Height
	shardBestState.Epoch = genesisShardBlock.Header.Epoch
	shardBestState.BeaconHeight = genesisShardBlock.Header.BeaconHeight
	shardBestState.TotalTxns += uint64(len(genesisShardBlock.Body.Transactions))
	shardBestState.NumTxns = uint64(len(genesisShardBlock.Body.Transactions))
	shardBestState.ShardProposerIdx = 0

	shardBestState.ConsensusAlgorithm = common.BlsConsensus
	shardBestState.NumOfBlocksByProducers = make(map[string]uint64)
	shardBestState.MaxTxsPerBlockRemainder = int64(config.Param().TransactionInBlockParam.Lower)

	// Get all beaconInstructions from beacon here
	beaconInstructions, _, err := blockchain.
		extractInstructionsFromBeacon([]*types.BeaconBlock{genesisBeaconBlock}, shardBestState.ShardID)
	if err != nil {
		return err
	}

	env := shardBestState.NewShardCommitteeStateEnvironmentWithValue(
		genesisShardBlock,
		blockchain,
		beaconInstructions,
		[]string{},
		genesisBeaconBlock.Header.Hash(),
	)

	shardBestState.shardCommitteeState = committeestate.InitGenesisShardCommitteeState(
		1,
		config.Param().ConsensusParam.StakingFlowV2Height,
		config.Param().ConsensusParam.StakingFlowV3Height,
		config.Param().ConsensusParam.StakingFlowV4Height,
		env)

	if config.Param().ConsensusParam.BlockProducingV3Height == shardBestState.BeaconHeight {
		if err := shardBestState.checkAndUpgradeStakingFlowV3Config(); err != nil {
			return err
		}
	}

	//statedb===========================START
	dbAccessWarper := statedb.NewDatabaseAccessWarper(db)
	shardBestState.consensusStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	shardBestState.transactionStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	shardBestState.featureStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	shardBestState.rewardStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	shardBestState.slashStateDB, err = statedb.NewWithPrefixTrie(common.EmptyRoot, dbAccessWarper)
	if err != nil {
		return err
	}
	shardBestState.ConsensusStateDBRootHash = common.EmptyRoot
	shardBestState.SlashStateDBRootHash = common.EmptyRoot
	shardBestState.RewardStateDBRootHash = common.EmptyRoot
	shardBestState.FeatureStateDBRootHash = common.EmptyRoot
	shardBestState.TransactionStateDBRootHash = common.EmptyRoot
	//statedb===========================END
	return nil
}

// verifyPostProcessingShardBlock
//   - commitee root
//   - pending validator root
func (shardBestState *ShardBestState) verifyPostProcessingShardBlock(shardBlock *types.ShardBlock, shardID byte,
	hashes *committeestate.ShardCommitteeStateHash) error {
	if !hashes.ShardCommitteeHash.IsEqual(&shardBlock.Header.CommitteeRoot) {
		return NewBlockChainError(ShardCommitteeRootHashError, fmt.Errorf("Expect %+v but get %+v", shardBlock.Header.CommitteeRoot, hashes.ShardCommitteeHash))
	}

	if !hashes.ShardSubstituteHash.IsEqual(&shardBlock.Header.PendingValidatorRoot) {
		return NewBlockChainError(ShardPendingValidatorRootHashError, fmt.Errorf("Expect %+v but get %+v", shardBlock.Header.PendingValidatorRoot, hashes.ShardSubstituteHash))
	}

	startTimeVerifyPostProcessingShardBlock := time.Now()
	Logger.log.Debugf("SHARD %+v | Begin VerifyPostProcessing Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash())
	//if hash, isOk := verifyHashFromMapStringString(shardBestState.StakingTx, shardBlock.Header.StakingTxRoot); !isOk {
	//	return NewBlockChainError(ShardPendingValidatorRootHashError, fmt.Errorf("Expect shard staking root hash to be %+v but get %+v", shardBlock.Header.StakingTxRoot, hash))
	//}
	shardVerifyPostProcessingTimer.UpdateSince(startTimeVerifyPostProcessingShardBlock)
	Logger.log.Debugf("SHARD %+v | Finish VerifyPostProcessing Block with height %+v at hash %+v", shardBlock.Header.ShardID, shardBlock.Header.Height, shardBlock.Hash().String())
	return nil
}

// Verify Transaction with these condition:
//  1. Validate tx version
//  2. Validate fee with tx size
//  3. Validate type of tx
//  4. Validate with other txs in block:
//     - Normal Transaction:
//     - Custom Tx:
//     4.1 Validate Init Custom Token
//  5. Validate sanity data of tx
//  6. Validate data in tx: privacy proof, metadata,...
//  7. Validate tx with blockchain: douple spend, ...
//  8. Check tx existed in block
//  9. Not accept a salary tx
//  10. Check duplicate staker public key in block
//  11. Check duplicate Init Custom Token in block
func (blockchain *BlockChain) verifyTransactionFromNewBlock(
	shardID byte,
	txs []metadata.Transaction,
	beaconHash common.Hash,
	curView *ShardBestState,
) error {
	if len(txs) == 0 {
		return nil
	}
	st := time.Now()
	// isEmpty := blockchain.config.TempTxPool.EmptyPool()
	// if !isEmpty {
	// 	panic("TempTxPool Is not Empty")
	// }
	// defer blockchain.config.TempTxPool.EmptyPool()
	_, beaconHeight, err := blockchain.GetBeaconBlockByHash(beaconHash)
	if err != nil {
		Logger.log.Errorf("Can not get beacon view state for new block err: %+v, get from beacon hash %v", err, beaconHash.String())
		return err
	}
	st = time.Now()
	err = blockchain.verifyTransactionIndividuallyFromNewBlock(shardID, txs, beaconHeight, beaconHash, curView)
	Logger.log.Infof("[validatetxs] verifyTransactionIndividuallyFromNewBlock cost %v", time.Since(st))
	return err
}
func hasCommitteeRelatedTx(txs ...metadata.Transaction) bool {
	for _, tx := range txs {
		if tx.GetMetadata() != nil {
			switch tx.GetMetadata().GetType() {
			case metadata.BeaconStakingMeta, metadata.ShardStakingMeta, metadata.StopAutoStakingMeta, metadata.UnStakingMeta:
				return true
			}
		}
	}
	return false
}

func (blockchain *BlockChain) verifyTransactionIndividuallyFromNewBlock(shardID byte, txs []metadata.Transaction, beaconHeight uint64, beaconHash common.Hash, curView *ShardBestState) error {
	if blockchain.config.usingNewPool {
		bView, err := blockchain.GetBeaconViewStateDataFromBlockHash(beaconHash, hasCommitteeRelatedTx(txs...), false, false)
		if err != nil {
			Logger.log.Errorf("Can not get beacon view state for new block err: %+v, get from beacon hash %v", err, beaconHash.String())
			return err
		}
		ok, err := blockchain.ShardChain[shardID].TxsVerifier.FullValidateTransactions(
			blockchain,
			curView,
			bView,
			txs,
		)
		if !ok || (err != nil) {
			return NewBlockChainError(TransactionFromNewBlockError, err)
		}
	} else {
		isEmpty := blockchain.config.TempTxPool.EmptyPool()
		if !isEmpty {
			panic("TempTxPool Is not Empty")
		}
		defer blockchain.config.TempTxPool.EmptyPool()
		listTxs := []metadata.Transaction{}
		txDB := curView.GetCopiedTransactionStateDB()
		whiteListTxs := blockchain.WhiteListTx()
		jsb, _ := json.Marshal(txs)
		Logger.log.Info("all transactions from block:", string(jsb))
		for _, tx := range txs {
			tx.LoadData(txDB)
			if ok := whiteListTxs[tx.Hash().String()]; ok {
				return nil
			}
			if tx.IsSalaryTx() {
				_, err := blockchain.config.TempTxPool.MaybeAcceptSalaryTransactionForBlockProducing(shardID, tx, int64(beaconHeight), curView)
				if err != nil {
					return err
				}
			} else {
				listTxs = append(listTxs, tx)
			}
		}
		_, err := blockchain.config.TempTxPool.MaybeAcceptBatchTransactionForBlockProducing(shardID, listTxs, int64(beaconHeight), curView)
		if err != nil {
			Logger.log.Errorf("Batching verify transactions from new block err: %+v\n Trying verify one by one", err)
			for index, tx := range listTxs {
				if blockchain.config.TempTxPool.HaveTransaction(tx.Hash()) {
					continue
				}
				_, err1 := blockchain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx, int64(beaconHeight), curView)
				if err1 != nil {
					Logger.log.Errorf("One by one verify txs at index %d error: %+v", index, err1)
					return NewBlockChainError(TransactionFromNewBlockError, fmt.Errorf("Transaction %+v, index %+v get %+v ", *tx.Hash(), index, err1))
				}
			}
		}
	}

	return nil
}

// processStoreShardBlock Store All information after Insert
//   - Shard Block
//   - Shard Best State
//   - Store tokenInit transactions (with metadata: InitTokenRequestMeta, IssuingRequestMeta, IssuingETHRequestMeta)
//   - Transaction => UTXO, serial number, snd, commitment
//   - Cross Output Coin => UTXO, snd, commmitment
//   - Store transaction metadata:
//   - Withdraw Metadata
//   - Store incoming cross shard block
//   - Store Burning Confirmation
//   - Update Mempool fee estimator
func (blockchain *BlockChain) processStoreShardBlock(
	newShardState *ShardBestState,
	shardBlock *types.ShardBlock,
	committeeChange *committeestate.CommitteeChange,
	beaconBlocks []*types.BeaconBlock,
) error {
	startTimeProcessStoreShardBlock := time.Now()
	shardID := shardBlock.Header.ShardID
	blockHeight := shardBlock.Header.Height
	blockHash := shardBlock.Header.Hash()
	err := blockchain.storeTokenInitInstructions(newShardState.transactionStateDB, beaconBlocks)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, fmt.Errorf("storeTokenInitInstructions error: %v", err))
	}
	// newShardState.
	if err = blockchain.processStoreAllShardStakerInfo(newShardState, beaconBlocks); err != nil {
		return err
	}

	Logger.log.Infof("SHARD %+v | Process store block height %+v at hash %+v", shardBlock.Header.ShardID, blockHeight, *shardBlock.Hash())
	if len(shardBlock.Body.CrossTransactions) != 0 {
		Logger.log.Critical("processStoreShardBlock/CrossTransactions	", shardBlock.Body.CrossTransactions)
	}

	if blockHeight == 1 {
		Logger.log.Infof("Genesis block of shard %v: %v, #txs: %v\n", shardID, blockHash.String(), len(shardBlock.Body.Transactions))
	}

	if err := blockchain.CreateAndSaveTxViewPointFromBlock(shardBlock, newShardState.transactionStateDB); err != nil {
		return NewBlockChainError(FetchAndStoreTransactionError, err)
	}
	listTxHashes := []string{}
	for index, tx := range shardBlock.Body.Transactions {
		Logger.log.Infof("Process storing tx %v, index %x, shard %v, height %v, blockHash %v\n", tx.Hash().String(), index, shardID, blockHeight, blockHash.String())
		listTxHashes = append(listTxHashes, tx.Hash().String())
		if err := rawdbv2.StoreTransactionIndex(blockchain.GetShardChainDatabase(shardID), *tx.Hash(), shardBlock.Header.Hash(), index); err != nil {
			return NewBlockChainError(FetchAndStoreTransactionError, err)
		}
		// Process Transaction Metadata
		metaType := tx.GetMetadataType()
		if metaType == metadata.WithDrawRewardResponseMeta {
			isMinted, mintCoin, coinID, err := tx.GetTxMintData()
			if err != nil || !isMinted {
				return NewBlockChainError(RemoveCommitteeRewardError, err)
			}

			if tx.GetVersion() == 1 {
				err = statedb.RemoveCommitteeReward(newShardState.rewardStateDB, mintCoin.GetPublicKey().ToBytesS(), mintCoin.GetValue(), *coinID)
				if err != nil {
					return NewBlockChainError(RemoveCommitteeRewardError, err)
				}
			} else {
				md, ok := tx.GetMetadata().(*metadata.WithDrawRewardResponse)
				if !ok {
					return NewBlockChainError(RemoveCommitteeRewardError, fmt.Errorf("cannot parse withdraw reward response metadata for tx %v", tx.Hash().String()))
				}

				err = statedb.RemoveCommitteeReward(newShardState.rewardStateDB, md.RewardPublicKey, mintCoin.GetValue(), *coinID)
				if err != nil {
					return NewBlockChainError(RemoveCommitteeRewardError, err)
				}
			}
		}
		Logger.log.Debug("Transaction in block with hash", blockHash, "and index", index)
	}
	if blockchain.UsingNewPool() {
		if len(listTxHashes) > 0 {
			if blockchain.ShardChain[shardID].TxPool.IsRunning() {
				blockchain.ShardChain[shardID].TxPool.RemoveTxs(listTxHashes)
			}
		}
	}
	// Store Incomming Cross Shard
	if err := blockchain.CreateAndSaveCrossTransactionViewPointFromBlock(shardBlock, newShardState.transactionStateDB); err != nil {
		return NewBlockChainError(FetchAndStoreCrossTransactionError, err)
	}
	// Save result of BurningConfirm instruction to get proof later
	metas := []string{ // Burning v1: sig on both beacon and bridge
		strconv.Itoa(metadata.BurningConfirmMeta),
		strconv.Itoa(metadata.BurningConfirmForDepositToSCMeta),
	}
	err = blockchain.storeBurningConfirm(newShardState.featureStateDB, shardBlock.Body.Instructions, shardBlock.Header.Height, metas)
	if err != nil {
		return NewBlockChainError(StoreBurningConfirmError, err)
	}
	// Update bridge issuancstore sharde request status
	err = blockchain.updateBridgeIssuanceStatus(newShardState.featureStateDB, shardBlock)
	if err != nil {
		return NewBlockChainError(UpdateBridgeIssuanceStatusError, err)
	}

	// call FeeEstimator for processing recent blocks
	if feeEstimator, ok := blockchain.config.FeeEstimator[shardBlock.Header.ShardID]; ok && time.Since(time.Unix(shardBlock.GetProduceTime(), 0)).Seconds() < 15*60 {
		go func(fe FeeEstimator) {
			err := fe.RegisterBlock(shardBlock)
			if err != nil {
				Logger.log.Debug(NewBlockChainError(RegisterEstimatorFeeError, err))
			}
		}(feeEstimator)
	}
	if len(committeeChange.ShardCommitteeAdded[shardID]) > 0 || len(committeeChange.ShardSubstituteAdded[shardID]) > 0 {
		err = statedb.StoreOneShardCommittee(newShardState.consensusStateDB, shardID, committeeChange.ShardCommitteeAdded[shardID])
		if err != nil {
			return NewBlockChainError(StoreShardBlockError, err)
		}
		err = statedb.StoreOneShardSubstitutesValidator(newShardState.consensusStateDB, shardID, committeeChange.ShardSubstituteAdded[shardID])
		if err != nil {
			return NewBlockChainError(StoreShardBlockError, fmt.Errorf("can't get ConsensusStateRootHash of height %+v ,error %+v", newShardState.GetHeight(), err))
		}
	}
	err = statedb.ReplaceOneShardCommittee(newShardState.consensusStateDB, shardID, committeeChange.ShardCommitteeReplaced[shardID])
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = statedb.DeleteOneShardCommittee(newShardState.consensusStateDB, shardID, committeeChange.ShardCommitteeRemoved[shardID])
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = statedb.DeleteOneShardSubstitutesValidator(newShardState.consensusStateDB, shardID, committeeChange.ShardSubstituteRemoved[shardID])
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}

	// consensus root hash
	consensusRootHash, err := newShardState.consensusStateDB.Commit(true) // Store data to memory
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = newShardState.consensusStateDB.Database().TrieDB().Commit(consensusRootHash, false) // Save data to disk database
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}

	newShardState.ConsensusStateDBRootHash = consensusRootHash
	// transaction root hash
	transactionRootHash, err := newShardState.transactionStateDB.Commit(true)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = newShardState.transactionStateDB.Database().TrieDB().Commit(transactionRootHash, false)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	newShardState.TransactionStateDBRootHash = transactionRootHash
	// feature root hash
	featureRootHash, err := newShardState.featureStateDB.Commit(true)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = newShardState.featureStateDB.Database().TrieDB().Commit(featureRootHash, false)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	newShardState.FeatureStateDBRootHash = featureRootHash
	// reward root hash
	rewardRootHash, err := newShardState.rewardStateDB.Commit(true)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = newShardState.rewardStateDB.Database().TrieDB().Commit(rewardRootHash, false)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	newShardState.RewardStateDBRootHash = rewardRootHash
	// slash root hash
	slashRootHash, err := newShardState.slashStateDB.Commit(true)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	err = newShardState.slashStateDB.Database().TrieDB().Commit(slashRootHash, false)
	if err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}
	newShardState.consensusStateDB.ClearObjects()
	newShardState.transactionStateDB.ClearObjects()
	newShardState.featureStateDB.ClearObjects()
	newShardState.rewardStateDB.ClearObjects()
	newShardState.slashStateDB.ClearObjects()

	batchData := blockchain.GetShardChainDatabase(shardID).NewBatch()
	sRH := ShardRootHash{
		ConsensusStateDBRootHash:   consensusRootHash,
		FeatureStateDBRootHash:     featureRootHash,
		RewardStateDBRootHash:      rewardRootHash,
		SlashStateDBRootHash:       slashRootHash,
		TransactionStateDBRootHash: transactionRootHash,
	}

	if err := rawdbv2.StoreShardRootsHash(batchData, shardID, blockHash, sRH); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}

	//statedb===========================END
	if err := rawdbv2.StoreShardBlock(batchData, blockHash, shardBlock); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}

	err = newShardState.tryUpgradeCommitteeState(blockchain)
	if err != nil {
		panic(NewBlockChainError(-11111, fmt.Errorf("Upgrade Committe Engine Error, %+v", err)))
	}

	simulatedMultiView := blockchain.ShardChain[shardBlock.Header.ShardID].multiView.SimulateAddView(newShardState)
	err = blockchain.BackupShardViews(batchData, shardBlock.Header.ShardID, simulatedMultiView)

	storeBlock := simulatedMultiView.GetExpectedFinalView().GetBlock()
	//traverse back to final view
	if shardBlock.GetVersion() < types.INSTANT_FINALITY_VERSION {
		oldFinalView := blockchain.ShardChain[shardID].multiView.GetFinalView()
		for {
			if oldFinalView != nil && storeBlock.GetHeight() <= oldFinalView.GetHeight() {
				break
			}
			err := rawdbv2.StoreFinalizedShardBlockHashByIndex(batchData, shardID, storeBlock.GetHeight(), *storeBlock.Hash())
			if err != nil {
				return NewBlockChainError(StoreBeaconBlockError, err)
			}

			if storeBlock.GetHeight() == 1 {
				break
			}

			prevHash := storeBlock.GetPrevHash()
			preView := blockchain.ShardChain[shardID].multiView.GetViewByHash(prevHash)
			if preView == nil {
				storeBlock, _, err = blockchain.GetShardBlockByHashWithShardID(prevHash, shardID)
				if err != nil {
					panic("Database is corrupt")
				}
			} else {
				storeBlock = preView.GetBlock()
			}
		}
	} else { //instant finality
		blockchain.storeFinalizeShardBlockByBeaconView(batchData, shardID, *simulatedMultiView.GetExpectedFinalView().GetHash())
	}

	if err != nil {
		panic("Backup shard view error")
	}

	if err := batchData.Write(); err != nil {
		return NewBlockChainError(StoreShardBlockError, err)
	}

	//add view
	isSuccess := blockchain.ShardChain[shardBlock.Header.ShardID].AddView(newShardState)
	if !isSuccess {
		return NewBlockChainError(StoreShardBlockError, err)
	}

	txDB := simulatedMultiView.GetBestView().(*ShardBestState).GetCopiedTransactionStateDB()
	//TODO: @hy check this txDB only use  to verify incoming tx
	blockchain.ShardChain[shardBlock.Header.ShardID].TxsVerifier.UpdateTransactionStateDB(txDB)

	if !config.Config().ForceBackup {
		return nil
	}

	backupPoint := false
	for _, beaconBlock := range beaconBlocks {
		if blockchain.IsLastBeaconHeightInEpoch(beaconBlock.GetHeight() + 1) {
			backupPoint = true
		}
	}

	if backupPoint {
		err := blockchain.GetShardChainDatabase(newShardState.ShardID).Backup(fmt.Sprintf("../../backup/shard%d/%d", newShardState.ShardID, newShardState.Epoch))
		if err != nil {
			blockchain.GetShardChainDatabase(newShardState.ShardID).RemoveBackup(fmt.Sprintf("../../backup/shard%d/%d", newShardState.ShardID, newShardState.Epoch))
		}
	}

	shardStoreBlockTimer.UpdateSince(startTimeProcessStoreShardBlock)
	Logger.log.Infof("SHARD %+v | 🔎 %d transactions in block height %+v \n", shardBlock.Header.ShardID, len(shardBlock.Body.Transactions), blockHeight)
	return nil
}

func (blockchain *BlockChain) storeFinalizeShardBlockByBeaconView(db incdb.KeyValueWriter, shardID byte, finalizedBlockHash common.Hash) error {
	finalizedBlockView := blockchain.ShardChain[shardID].multiView.GetViewByHash(finalizedBlockHash)
	if finalizedBlockView == nil {
		finalizedBlockView = blockchain.ShardChain[shardID].multiView.GetExpectedFinalView()
	}

	finalizedBlock := finalizedBlockView.GetBlock()
	for {
		_, err := rawdbv2.GetFinalizedShardBlockHashByIndex(blockchain.GetShardChainDatabase(shardID), shardID, finalizedBlock.GetHeight())
		if err == nil { //already insert
			break
		}
		confirmHash, err := rawdbv2.GetBeaconConfirmInstantFinalityShardBlock(blockchain.GetBeaconChainDatabase(), shardID, finalizedBlock.GetHeight())
		if err == nil && confirmHash.String() == finalizedBlock.Hash().String() {
			Logger.log.Info("============== StoreFinalizedShardBlockHashByIndex", shardID, finalizedBlock.GetHeight(), finalizedBlock.Hash().String())
			blockchain.ShardChain[shardID].multiView.FinalizeView(*confirmHash)
			err = rawdbv2.StoreFinalizedShardBlockHashByIndex(db, shardID, finalizedBlock.GetHeight(), *finalizedBlock.Hash())
			if err != nil {
				return NewBlockChainError(StoreBeaconBlockError, err)
			}
		}
		prevHash := finalizedBlock.GetPrevHash()
		preView := blockchain.ShardChain[shardID].multiView.GetViewByHash(prevHash)
		if preView == nil {
			finalizedBlock, _, err = blockchain.GetShardBlockByHashWithShardID(prevHash, shardID)
			if err != nil {
				panic("Database is corrupt")
			}
		} else {
			finalizedBlock = preView.GetBlock()
		}
	}
	return nil
}

// removeOldDataAfterProcessingShardBlock remove outdate data from pool and beststate
//   - Remove Staking TX in Shard BestState from instruction
//   - Set Shard State for removing old Shard Block in Pool
//   - Remove Old Cross Shard Block
//   - Remove Init Tokens ID in Mempool
//   - Remove Candiates in Mempool
//   - Remove Transaction in Mempool and Block Generator
func (blockchain *BlockChain) removeOldDataAfterProcessingShardBlock(shardBlock *types.ShardBlock, shardID byte) {
	go func() {
		//Remove Candidate In pool
		candidates := []string{}
		for _, tx := range shardBlock.Body.Transactions {
			if blockchain.config.IsBlockGenStarted {
				blockchain.config.CRemovedTxs <- tx
			}
			if tx.GetMetadata() != nil {
				if tx.GetMetadata().GetType() == metadata.ShardStakingMeta || tx.GetMetadata().GetType() == metadata.BeaconStakingMeta {
					stakingMetadata, ok := tx.GetMetadata().(*metadata.StakingMetadata)
					if !ok {
						continue
					}
					candidates = append(candidates, stakingMetadata.CommitteePublicKey)
				}
			}
		}
		go blockchain.config.TxPool.RemoveCandidateList(candidates)
		//Remove tx out of pool
		go blockchain.config.TxPool.RemoveTx(shardBlock.Body.Transactions, true)
	}()
}

func (blockchain *BlockChain) GetShardCommitteeFromBeaconHash(
	committeeFromBlock common.Hash, shardID byte) ([]incognitokey.CommitteePublicKey, error) {
	_, _, err := blockchain.GetBeaconBlockByHash(committeeFromBlock)
	if err != nil {
		return []incognitokey.CommitteePublicKey{}, NewBlockChainError(CommitteeFromBlockNotFoundError, err)
	}

	bRH, err := GetBeaconRootsHashByBlockHash(blockchain.GetBeaconChainDatabase(), committeeFromBlock)
	if err != nil {
		return []incognitokey.CommitteePublicKey{}, NewBlockChainError(CommitteeFromBlockNotFoundError, err)
	}

	stateDB, err := statedb.NewWithPrefixTrie(
		bRH.ConsensusStateDBRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetBeaconChainDatabase()))
	if err != nil {
		return []incognitokey.CommitteePublicKey{}, NewBlockChainError(CommitteeFromBlockNotFoundError, err)
	}
	committees := statedb.GetOneShardCommittee(stateDB, shardID)

	return committees, nil
}

// storeTokenInitInstructions tries to store new tokens when they are initialized. There are 3 ways to init a token:
//  1. InitTokenRequestMeta - for user-customized tokens
//  2. IssuingRequestMeta - for centralized bridge tokens
//  3. IssuingETHRequestMeta - for decentralized bridge tokens
func (blockchain *BlockChain) storeTokenInitInstructions(stateDB *statedb.StateDB, beaconBlocks []*types.BeaconBlock) error {
	for _, block := range beaconBlocks {
		instructions := block.Body.Instructions

		for _, l := range instructions {
			if len(l) < 4 {
				continue
			}
			if instruction.IsConsensusInstruction(l[0]) {
				continue
			}

			metaType, err := strconv.Atoi(l[0])
			if err != nil {
				return err
			}
			switch metaType {
			case metadata.InitTokenRequestMeta:
				if len(l) == 4 && l[2] == "accepted" {
					acceptedContent, err := metadata.ParseInitTokenInstAcceptedContent(l[3])
					if err != nil {
						Logger.log.Errorf("ParseInitTokenInstAcceptedContent(%v) error: %v\n", l[3], err)
						return err
					}

					if existed := statedb.PrivacyTokenIDExisted(stateDB, acceptedContent.TokenID); existed {
						msgStr := fmt.Sprintf("init token %v existed, something might be wrong", acceptedContent.TokenID.String())
						Logger.log.Infof(msgStr + "\n")
						return fmt.Errorf(msgStr)
					}

					err = statedb.StorePrivacyToken(stateDB, acceptedContent.TokenID, acceptedContent.TokenName,
						acceptedContent.TokenSymbol, statedb.InitToken, true, acceptedContent.Amount, []byte{}, acceptedContent.RequestedTxID,
					)
					if err != nil {
						Logger.log.Errorf("StorePrivacyToken error: %v\n", err)
						return err
					}

					Logger.log.Infof("store init token %v succeeded\n", acceptedContent.TokenID.String())
				}

			case metadata.IssuingETHRequestMeta, metadata.IssuingBSCRequestMeta,
				metadata.IssuingPRVERC20RequestMeta, metadata.IssuingPRVBEP20RequestMeta,
				metadata.IssuingPLGRequestMeta, metadata.IssuingFantomRequestMeta:
				if len(l) >= 4 && l[2] == "accepted" {
					acceptedContent, err := metadataBridge.ParseEVMIssuingInstAcceptedContent(l[3])
					if err != nil {
						Logger.log.Errorf("ParseEVMIssuingInstAcceptedContent(%v) error: %v\n", l[3], err)
						return err
					}
					if existed := statedb.PrivacyTokenIDExisted(stateDB, acceptedContent.IncTokenID); existed {
						Logger.log.Infof("eth-issued token %v existed, skip storing this token\n", acceptedContent.IncTokenID.String())
						continue
					}

					err = statedb.StorePrivacyToken(stateDB, acceptedContent.IncTokenID, "",
						"", statedb.BridgeToken, true, acceptedContent.IssuingAmount, []byte{}, acceptedContent.TxReqID,
					)
					if err != nil {
						Logger.log.Errorf("StorePrivacyToken error: %v\n", err)
						return err
					}

					Logger.log.Infof("store eth-isssued token %v succeeded\n", acceptedContent.IncTokenID.String())
				}
			case metadataCommon.IssuingUnifiedTokenRequestMeta:
				if len(l) >= 4 && l[2] == "accepted" {
					acceptedContent, err := metadataBridge.ParseShieldReqInstAcceptedContent(l[3])
					if err != nil {
						Logger.log.Errorf("ParseShieldReqInstAcceptedContent(%v) error: %v\n", l[3], err)
						return err
					}
					if existed := statedb.PrivacyTokenIDExisted(stateDB, acceptedContent.UnifiedTokenID); existed {
						Logger.log.Infof("issued token %v existed, skip storing this token\n", acceptedContent.UnifiedTokenID.String())
						continue
					}
					mintAmt := uint64(0)
					for _, data := range acceptedContent.Data {
						mintAmt = mintAmt + data.ShieldAmount + data.Reward
						if mintAmt < data.ShieldAmount+data.Reward {
							Logger.log.Errorf("StorePrivacyToken out of range minted amount tokenID %v, txId %v",
								acceptedContent.UnifiedTokenID, acceptedContent.TxReqID)
							return fmt.Errorf("StorePrivacyToken out of range minted amount tokenID %v, txId %v",
								acceptedContent.UnifiedTokenID, acceptedContent.TxReqID)
						}
					}

					err = statedb.StorePrivacyToken(stateDB, acceptedContent.UnifiedTokenID, "",
						"", statedb.BridgeToken, true, mintAmt, []byte{}, acceptedContent.TxReqID,
					)
					if err != nil {
						Logger.log.Errorf("StorePrivacyToken error: %v\n", err)
						return err
					}
					Logger.log.Infof("store unified-isssued token %v succeeded\n", acceptedContent.UnifiedTokenID.String())
				}

			case metadata.IssuingRequestMeta:
				if len(l) >= 4 && l[2] == "accepted" {
					acceptedContent, err := metadata.ParseIssuingInstAcceptedContent(l[3])
					if err != nil {
						Logger.log.Errorf("ParseIssuingInstAcceptedContent(%v) error: %v\n", l[3], err)
						return err
					}

					if existed := statedb.PrivacyTokenIDExisted(stateDB, acceptedContent.IncTokenID); existed {
						Logger.log.Infof("issued token %v existed, skip storing this token\n", acceptedContent.IncTokenID.String())
						continue
					}

					err = statedb.StorePrivacyToken(stateDB, acceptedContent.IncTokenID, acceptedContent.IncTokenName,
						acceptedContent.IncTokenName, statedb.BridgeToken, true, acceptedContent.DepositedAmount, []byte{}, acceptedContent.TxReqID,
					)
					if err != nil {
						Logger.log.Errorf("StorePrivacyToken error: %v\n", err)
						return err
					}

					Logger.log.Infof("store issued token %v succeeded\n", acceptedContent.IncTokenID.String())
				}
			case metadata.IssuingReshieldResponseMeta:
				if len(l) >= 4 && l[2] == "accepted" {
					inst := metadataCommon.NewInstruction()
					if err := inst.FromStringSlice(l); err != nil {
						Logger.log.Errorf("Parse IssuingReshield(%v) error: %v", l[3], err)
						return err
					}
					contentBytes, err := base64.StdEncoding.DecodeString(inst.Content)
					if err != nil {
						Logger.log.Errorf("Parse IssuingReshield(%v) error: %v", l[3], err)
						return err
					}
					var acceptedContent metadataBridge.AcceptedReshieldRequest
					err = json.Unmarshal(contentBytes, &acceptedContent)
					if err != nil {
						Logger.log.Errorf("Parse IssuingReshield(%v) error: %v", l[3], err)
						return err
					}
					shieldTokenID := acceptedContent.ReshieldData.IncTokenID
					if acceptedContent.UnifiedTokenID != nil {
						shieldTokenID = *acceptedContent.UnifiedTokenID
					}
					if existed := statedb.PrivacyTokenIDExisted(stateDB, shieldTokenID); existed {
						Logger.log.Infof("eth-reshield token %v existed, skip storing this token", shieldTokenID.String())
						continue
					}

					err = statedb.StorePrivacyToken(stateDB, shieldTokenID, "",
						"", statedb.BridgeToken, true, acceptedContent.ReshieldData.ShieldAmount, []byte{}, acceptedContent.TxReqID,
					)
					if err != nil {
						Logger.log.Errorf("StorePrivacyToken error: %v", err)
						return err
					}

					Logger.log.Infof("store eth-reshield token %v succeeded", shieldTokenID.String())
				}
			}
		}
	}

	return nil
}

func (blockchain *BlockChain) processStoreAllShardStakerInfo(
	sBestState *ShardBestState,
	beaconBlocks []*types.BeaconBlock,
) error {
	shardStakerDelegate := map[string]string{}
	shardStakerRewardReceivers := map[string]key.PaymentAddress{}
	shardIsCommittee := map[string]bool{}
	shardReDelegate := map[string]string{}
	outPKList := []string{}

	sConsensusDB := sBestState.consensusStateDB
	currentAllStaker, has, err := statedb.GetAllShardStakersInfo(sConsensusDB)
	if err != nil {
		return err
	}
	sID := sBestState.GetShardID()
	neededUpdate := false
	for _, beaconBlock := range beaconBlocks {
		if blockchain.IsFirstBeaconHeightInEpoch(beaconBlock.GetBeaconHeight()) {
			BLogger.log.Infof("teststore Start Store All Shard Staker, sH %v ", sBestState.ShardHeight)
			neededUpdate = true
			bcBestState := blockchain.GetBeaconBestState()
			beaconConsensusRootHash, err := blockchain.GetBeaconConsensusRootHash(bcBestState, bcBestState.GetHeight())
			if err != nil {
				return fmt.Errorf("Beacon Consensus Root Hash of Height %+v not found ,error %+v", bcBestState.GetHeight(), err)
			}
			beaconConsensusStateDB, err := statedb.NewWithPrefixTrie(beaconConsensusRootHash, statedb.NewDatabaseAccessWarper(blockchain.GetBeaconChainDatabase()))
			if err != nil {
				return fmt.Errorf("init beacon consensus statedb return error %v", err)
			}

			for _, inst := range beaconBlock.Body.Instructions {
				if inst[0] == instruction.SWAP_SHARD_ACTION {
					swapShardInstruction := instruction.ImportSwapShardInstructionFromString(inst)
					fmt.Printf("shard %v process list in of shard %v, bc %v\n", sID, swapShardInstruction.ChainID, beaconBlock.Header.Height)
					newCommittees := swapShardInstruction.InPublicKeys
					for _, pkStr := range newCommittees {
						infor, has, err := statedb.GetStakerInfo(beaconConsensusStateDB, pkStr)
						BLogger.log.Infof("teststore Add new committee pk %v, delegator %v ", sBestState.ShardHeight, infor.Delegate())
						if err != nil {
							return err
						}
						if !has {
							return errors.Errorf("Can not found this staker %v", pkStr)
						}
						rewReceiverPk := infor.RewardReceiver().Pk
						pkNew := base58.Base58Check{}.Encode(infor.RewardReceiver().Pk, 0)
						if common.GetShardIDFromLastByte(rewReceiverPk[len(rewReceiverPk)-1]) != sID {
							BLogger.log.Infof("testxx shard %v Skip pk %v, true sID %v", sID, pkNew, common.GetShardIDFromLastByte(rewReceiverPk[len(rewReceiverPk)-1]))
							continue
						}
						BLogger.log.Infof("testxx shard %v received pk %v", sID, pkNew)
						shardStakerDelegate[pkStr] = infor.Delegate()
						shardStakerRewardReceivers[pkStr] = infor.RewardReceiver()
						shardIsCommittee[pkStr] = true
					}
					outCommittees := swapShardInstruction.OutPublicKeys
					for _, pkStr := range outCommittees {
						shardIsCommittee[pkStr] = false
					}
					continue
				}
				if inst[0] == instruction.RETURN_ACTION {
					returnStakingIns, err := instruction.ValidateAndImportReturnStakingInstructionFromString(inst)
					if err != nil {
						Logger.log.Errorf("SKIP Return staking instruction %+v, error %+v", returnStakingIns, err)
						continue
					}
					outPKList = append(outPKList, returnStakingIns.PublicKeys...)
				}
			}
			slashed := statedb.GetSlashingCommittee(beaconConsensusStateDB, beaconBlock.Header.Epoch-1)
			for _, v := range slashed {
				outPKList = append(outPKList, v...)
			}
		}
	}
	if has {
		curStakerDelegate := currentAllStaker.MapDelegate()
		for sPK, newDelegate := range shardReDelegate {
			if _, ok := curStakerDelegate[sPK]; ok {
				curStakerDelegate[sPK] = newDelegate
				if !neededUpdate {
					neededUpdate = true
				}
			}
		}
		if !neededUpdate {
			return nil
		}
		curIsCommittee := currentAllStaker.IsCommittee()
		for k, v := range shardIsCommittee {
			if _, ok := curIsCommittee[k]; (ok) || (v) {
				curIsCommittee[k] = v
			}
		}
		for k, v := range shardStakerDelegate {
			curStakerDelegate[k] = v
		}
		curStakerRewardReceiver := currentAllStaker.RewardReceiver()

		for k, v := range shardStakerRewardReceivers {
			curStakerRewardReceiver[k] = v
		}
		for _, slashedKey := range outPKList {
			delete(curStakerDelegate, slashedKey)
			delete(curStakerRewardReceiver, slashedKey)
			delete(curIsCommittee, slashedKey)
		}
		return statedb.StoreAllShardStakersInfo(sConsensusDB, curStakerDelegate, curIsCommittee, curStakerRewardReceiver)

	} else {
		if neededUpdate {
			BLogger.log.Infof("teststore Cannot get all staker info, shard height %v", sBestState.ShardHeight)
			for _, slashedKey := range outPKList {
				delete(shardStakerDelegate, slashedKey)
				delete(shardStakerRewardReceivers, slashedKey)
				delete(shardIsCommittee, slashedKey)
			}
		}
		return statedb.StoreAllShardStakersInfo(sConsensusDB, shardStakerDelegate, shardIsCommittee, shardStakerRewardReceivers)
	}
}

// func (blockchain *BlockChain) processInitAllShardStakerInfo(sBestState *ShardBestState) error {
// 	shardStakerDelegate := map[string]string{}
// 	shardStakerRewardReceivers := map[string]key.PaymentAddress{}
// 	shardHasCredit := map[string]bool{}
// 	shardReDelegate := map[string]string{}
// 	sConsensusDB := sBestState.consensusStateDB
// 	currentAllStaker, has, err := statedb.GetAllShardStakersInfo(sConsensusDB)
// 	if err != nil {
// 		return err
// 	}
// 	sID := sBestState.GetShardID()
// 	neededUpdate := false
// 	for _, beaconBlock := range beaconBlocks {
// 		for _, inst := range beaconBlock.Body.Instructions {
// 			if inst[0] == instruction.STAKE_ACTION {
// 				stakeInstruction := instruction.ImportInitStakeInstructionFromString(inst)
// 				for index, receiverPayment := range stakeInstruction.RewardReceiverStructs {
// 					if common.GetShardIDFromLastByte(receiverPayment.Pk[len(receiverPayment.Pk)-1]) == sID {
// 						if !neededUpdate {
// 							neededUpdate = true
// 						}
// 						stakerPK := stakeInstruction.PublicKeys[index]
// 						stakerDelegate := stakeInstruction.DelegateList[index]
// 						shardStakerDelegate[stakerPK] = stakerDelegate
// 						shardStakerRewardReceivers[stakerPK] = receiverPayment
// 						shardHasCredit[stakerPK] = false
// 					}
// 				}
// 			}
// 			if inst[0] == instruction.RE_DELEGATE {
// 				redelegateInstruction, err := instruction.ValidateAndImportReDelegateInstructionFromString(inst)
// 				if err != nil {
// 					Logger.log.Errorf("SKIP stop auto stake instruction %+v, error %+v", inst, err)
// 					continue
// 				}
// 				for index, shardPublicKey := range redelegateInstruction.CommitteePublicKeys {
// 					shardReDelegate[shardPublicKey] = redelegateInstruction.DelegateList[index]
// 				}
// 			}
// 		}
// 	}
// 	if has {
// 		curStakerDelegate := currentAllStaker.MapDelegate()
// 		for sPK, newDelegate := range shardReDelegate {
// 			if _, ok := curStakerDelegate[sPK]; ok {
// 				curStakerDelegate[sPK] = newDelegate
// 				if !neededUpdate {
// 					neededUpdate = true
// 				}
// 			}
// 		}
// 		if !neededUpdate {
// 			return nil
// 		}
// 		for k, v := range shardStakerDelegate {
// 			curStakerDelegate[k] = v
// 		}
// 		curStakerRewardReceiver := currentAllStaker.RewardReceiver()
// 		curHasCredit := currentAllStaker.HasCredit()
// 		for k, v := range shardHasCredit {
// 			curHasCredit[k] = v
// 		}
// 		for k, v := range shardStakerRewardReceivers {
// 			curStakerRewardReceiver[k] = v
// 		}

// 	}
// 	return statedb.StoreAllShardStakersInfo(sConsensusDB, shardStakerDelegate, shardHasCredit, shardStakerRewardReceivers)
// }
