package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/incdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/incognitokey"
)

func (blockGenerator *BlockGenerator) NewBlockShardV2(shardID byte, round int, crossShards map[byte]uint64, beaconHeight uint64, start time.Time) (*ShardBlock, error) {
	var (
		transactionsForNewBlock = make([]metadata.Transaction, 0)
		totalTxsFee             = make(map[common.Hash]uint64)
		newShardBlock           = NewShardBlock()
		instructions            = [][]string{}
		isOldBeaconHeight       = false
		tempPrivateKey          = blockGenerator.createTempKeyset()
		shardBestState          = NewShardBestState()
	)
	Logger.log.Criticalf("⛏ Creating Shard Block %+v", blockGenerator.chain.BestState.Shard[shardID].ShardHeight+1)
	shardPendingValidator, err := incognitokey.CommitteeKeyListToString(blockGenerator.chain.BestState.Shard[shardID].ShardPendingValidator)
	if err != nil {
		return nil, err
	}
	currentCommitteePubKeys, err := incognitokey.CommitteeKeyListToString(blockGenerator.chain.BestState.Shard[shardID].ShardCommittee)
	if err != nil {
		return nil, err
	}
	//========Verify newShardBlock with previous best state
	// Get Beststate of previous newShardBlock == previous best state
	// Clone best state value into new variable
	if err := shardBestState.cloneShardBestStateFrom(blockGenerator.chain.BestState.Shard[shardID]); err != nil {
		return nil, err
	}
	//==========Fetch Beacon Blocks============
	BLogger.log.Infof("Producing block: %d", blockGenerator.chain.BestState.Shard[shardID].ShardHeight+1)
	if beaconHeight-shardBestState.BeaconHeight > MAX_BEACON_BLOCK {
		beaconHeight = shardBestState.BeaconHeight + MAX_BEACON_BLOCK
	}
	beaconHashes, err := rawdbv2.GetBeaconBlockHashByIndex(blockGenerator.chain.GetDatabase(), beaconHeight)
	if err != nil {
		return nil, err
	}
	beaconHash := beaconHashes[0]
	beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(blockGenerator.chain.GetDatabase(), beaconHash)
	if err != nil {
		return nil, err
	}
	beaconBlock := BeaconBlock{}
	err = json.Unmarshal(beaconBlockBytes, &beaconBlock)
	if err != nil {
		return nil, err
	}
	epoch := beaconBlock.Header.Epoch
	if epoch-shardBestState.Epoch >= 1 {
		beaconHeight = shardBestState.Epoch * blockGenerator.chain.config.ChainParams.Epoch
		newBeaconHashes, err := rawdbv2.GetBeaconBlockHashByIndex(blockGenerator.chain.GetDatabase(), beaconHeight)
		if err != nil {
			return nil, err
		}
		newBeaconHash := newBeaconHashes[0]
		copy(beaconHash[:], newBeaconHash.GetBytes())
		epoch = shardBestState.Epoch + 1
	}
	Logger.log.Infof("Get Beacon Block With Height %+v, Shard BestState %+v", beaconHeight, shardBestState.BeaconHeight)
	//Fetch beacon block from height
	beaconBlocks, err := FetchBeaconBlockFromHeightV2(blockGenerator.chain.GetDatabase(), shardBestState.BeaconHeight+1, beaconHeight)
	if err != nil {
		return nil, err
	}
	// this  beacon height is already seen by shard best state
	if beaconHeight == shardBestState.BeaconHeight {
		isOldBeaconHeight = true
	}
	//==========Build block body============
	// Get Transaction For new Block
	// Get Cross output coin from other shard && produce cross shard transaction
	crossTransactions := blockGenerator.getCrossShardDataV2(shardID, shardBestState.BeaconHeight, beaconHeight, crossShards)
	Logger.log.Critical("Cross Transaction: ", crossTransactions)
	// Get Transaction for new block
	blockCreationLeftOver := blockGenerator.chain.BestState.Shard[shardID].BlockMaxCreateTime.Nanoseconds() - time.Since(start).Nanoseconds()
	txsToAddFromBlock, err := blockGenerator.getTransactionForNewBlockV2(&tempPrivateKey, shardID, blockGenerator.chain.GetDatabase(), beaconBlocks, blockCreationLeftOver, beaconHeight)
	if err != nil {
		return nil, err
	}
	transactionsForNewBlock = append(transactionsForNewBlock, txsToAddFromBlock...)
	// build txs with metadata
	txsWithMetadata, err := blockGenerator.chain.BuildResponseTransactionFromTxsWithMetadataV2(transactionsForNewBlock, &tempPrivateKey, shardID)
	if err != nil {
		return nil, err
	}
	transactionsForNewBlock = append(transactionsForNewBlock, txsWithMetadata...)
	// process instruction from beacon block
	shardPendingValidator, _, _ = blockGenerator.chain.processInstructionFromBeaconV2(beaconBlocks, shardID, newCommitteeChange())
	// Create Instruction
	instructions, _, _, err = blockGenerator.chain.generateInstruction(shardID, beaconHeight, isOldBeaconHeight, beaconBlocks, shardPendingValidator, currentCommitteePubKeys)
	if err != nil {
		return nil, NewBlockChainError(GenerateInstructionError, err)
	}
	if len(instructions) != 0 {
		Logger.log.Info("Shard Producer: Instruction", instructions)
	}
	newShardBlock.BuildShardBlockBody(instructions, crossTransactions, transactionsForNewBlock)
	//==========Build Essential Header Data=========
	// producer key
	producerPosition := (blockGenerator.chain.BestState.Shard[shardID].ShardProposerIdx + round) % len(currentCommitteePubKeys)
	producerKey, err := blockGenerator.chain.BestState.Shard[shardID].ShardCommittee[producerPosition].ToBase58()
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	producerPubKeyStr, err := blockGenerator.chain.BestState.Shard[shardID].ShardCommittee[producerPosition].ToBase58()
	if err != nil {
		return nil, NewBlockChainError(ConvertCommitteePubKeyToBase58Error, fmt.Errorf("Failed to convert pub key of producer to base58 string in shard block %+v of shardID %+v", newShardBlock.Header.Height, newShardBlock.Header.ShardID))
	}
	for _, tx := range newShardBlock.Body.Transactions {
		totalTxsFee[*tx.GetTokenID()] += tx.GetTxFee()
		txType := tx.GetType()
		if txType == common.TxCustomTokenPrivacyType {
			txCustomPrivacy := tx.(*transaction.TxCustomTokenPrivacy)
			totalTxsFee[*txCustomPrivacy.GetTokenID()] = txCustomPrivacy.GetTxFeeToken()
		}
	}
	newShardBlock.Header = ShardHeader{
		Producer:          producerKey, //committeeMiningKeys[producerPosition],
		ProducerPubKeyStr: producerPubKeyStr,
		ShardID:           shardID,
		Version:           SHARD_BLOCK_VERSION,
		PreviousBlockHash: shardBestState.BestBlockHash,
		Height:            shardBestState.ShardHeight + 1,
		Round:             round,
		Epoch:             epoch,
		CrossShardBitMap:  CreateCrossShardByteArray(newShardBlock.Body.Transactions, shardID),
		BeaconHeight:      beaconHeight,
		BeaconHash:        beaconHash,
		TotalTxsFee:       totalTxsFee,
		ConsensusType:     blockGenerator.chain.BestState.Shard[shardID].ConsensusAlgorithm,
	}
	//============Update Shard BestState=============
	if err := shardBestState.updateShardBestStateV2(blockGenerator.chain, newShardBlock, beaconBlocks, newCommitteeChange()); err != nil {
		return nil, err
	}
	//============Build Header=============
	// Build Root Hash for Header
	merkleRoots := Merkle{}.BuildMerkleTreeStore(newShardBlock.Body.Transactions)
	merkleRoot := &common.Hash{}
	if len(merkleRoots) > 0 {
		merkleRoot = merkleRoots[len(merkleRoots)-1]
	}
	crossTransactionRoot, err := CreateMerkleCrossTransaction(newShardBlock.Body.CrossTransactions)
	if err != nil {
		return nil, err
	}
	txInstructions, err := CreateShardInstructionsFromTransactionAndInstruction(newShardBlock.Body.Transactions, blockGenerator.chain, shardID)
	if err != nil {
		return nil, err
	}
	totalInstructions := []string{}
	for _, value := range txInstructions {
		totalInstructions = append(totalInstructions, value...)
	}
	for _, value := range instructions {
		totalInstructions = append(totalInstructions, value...)
	}
	instructionsHash, err := generateHashFromStringArray(totalInstructions)
	if err != nil {
		return nil, NewBlockChainError(InstructionsHashError, err)
	}
	tempShardCommitteePubKeys, err := incognitokey.CommitteeKeyListToString(shardBestState.ShardCommittee)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	committeeRoot, err := generateHashFromStringArray(tempShardCommitteePubKeys)
	if err != nil {
		return nil, NewBlockChainError(CommitteeHashError, err)
	}
	tempShardPendintValidator, err := incognitokey.CommitteeKeyListToString(shardBestState.ShardPendingValidator)
	if err != nil {
		return nil, NewBlockChainError(UnExpectedError, err)
	}
	pendingValidatorRoot, err := generateHashFromStringArray(tempShardPendintValidator)
	if err != nil {
		return nil, NewBlockChainError(PendingValidatorRootError, err)
	}
	stakingTxRoot, err := generateHashFromMapStringString(shardBestState.StakingTx)
	if err != nil {
		return nil, NewBlockChainError(StakingTxHashError, err)
	}
	// Instruction merkle root
	flattenTxInsts, err := FlattenAndConvertStringInst(txInstructions)
	if err != nil {
		return nil, NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from Tx: %+v", err))
	}
	flattenInsts, err := FlattenAndConvertStringInst(instructions)
	if err != nil {
		return nil, NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from block body: %+v", err))
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
	newShardBlock.Header.CommitteeRoot = committeeRoot
	newShardBlock.Header.PendingValidatorRoot = pendingValidatorRoot
	newShardBlock.Header.StakingTxRoot = stakingTxRoot
	newShardBlock.Header.Timestamp = start.Unix()
	copy(newShardBlock.Header.InstructionMerkleRoot[:], instMerkleRoot)
	return newShardBlock, nil
}

func (blockGenerator *BlockGenerator) getCrossShardDataV2(toShard byte, lastBeaconHeight uint64, currentBeaconHeight uint64, crossShards map[byte]uint64) map[byte][]CrossTransaction {
	crossTransactions := make(map[byte][]CrossTransaction)
	// get cross shard block
	allCrossShardBlock := blockGenerator.crossShardPool[toShard].GetValidBlock(crossShards)
	// Get Cross Shard Block
	for fromShard, crossShardBlock := range allCrossShardBlock {
		sort.SliceStable(crossShardBlock[:], func(i, j int) bool {
			return crossShardBlock[i].Header.Height < crossShardBlock[j].Header.Height
		})
		indexs := []int{}
		startHeight := blockGenerator.chain.BestState.Shard[toShard].BestCrossShard[fromShard]
		for index, crossShardBlock := range crossShardBlock {
			Logger.log.Critical("index cross shard block", index, crossShardBlock)
			if crossShardBlock.Header.Height <= startHeight {
				break
			}
			nextHeight, err := rawdbv2.GetCrossShardNextHeight(blockGenerator.chain.GetDatabase(), fromShard, toShard, startHeight)
			if err != nil {
				break
			}
			if nextHeight != crossShardBlock.Header.Height {
				continue
			}
			startHeight = nextHeight
			beaconHeight, err := blockGenerator.chain.FindBeaconHeightForCrossShardBlockV2(crossShardBlock.Header.BeaconHeight, crossShardBlock.Header.ShardID, crossShardBlock.Header.Height)
			if err != nil {
				Logger.log.Errorf("%+v", err)
				break
			}
			consensusStateRootHash, ok := blockGenerator.chain.BestState.Beacon.GetConsensusStateRootHash(beaconHeight)
			if !ok {
				Logger.log.Errorf("Can't found ConsensusStateRootHash of beacon height %+v ", beaconHeight)
				break
			}
			consensusStateDB, err := statedb.NewWithPrefixTrie(consensusStateRootHash, statedb.NewDatabaseAccessWarper(blockGenerator.chain.GetDatabase()))
			if err != nil {
				Logger.log.Error(err)
				break
			}
			shardCommittee := statedb.GetOneShardCommittee(consensusStateDB, crossShardBlock.Header.ShardID)
			err = crossShardBlock.VerifyCrossShardBlock(blockGenerator.chain, shardCommittee)
			if err != nil {
				Logger.log.Error(err)
				break
			}
			indexs = append(indexs, index)
		}
		for _, index := range indexs {
			blk := crossShardBlock[index]
			crossTransaction := CrossTransaction{
				OutputCoin:       blk.CrossOutputCoin,
				TokenPrivacyData: blk.CrossTxTokenPrivacyData,
				BlockHash:        *blk.Hash(),
				BlockHeight:      blk.Header.Height,
			}
			crossTransactions[blk.Header.ShardID] = append(crossTransactions[blk.Header.ShardID], crossTransaction)
		}
	}
	for _, crossTransaction := range crossTransactions {
		sort.SliceStable(crossTransaction[:], func(i, j int) bool {
			return crossTransaction[i].BlockHeight < crossTransaction[j].BlockHeight
		})
	}
	return crossTransactions
}

//FindBeaconHeightForCrossShardBlockV2 Find Beacon Block with compatible shard states of cross shard block
func (blockchain *BlockChain) FindBeaconHeightForCrossShardBlockV2(beaconHeight uint64, fromShardID byte, crossShardBlockHeight uint64) (uint64, error) {
	for {
		beaconBlocks, err := blockchain.GetBeaconBlockByHeightV2(beaconHeight)
		if err != nil {
			return 0, NewBlockChainError(FetchBeaconBlockError, err)
		}
		beaconBlock := beaconBlocks[0]
		if shardStates, ok := beaconBlock.Body.ShardState[fromShardID]; ok {
			for _, shardState := range shardStates {
				if shardState.Height == crossShardBlockHeight {
					return beaconBlock.Header.Height, nil
				}
			}
		}
		beaconHeight += 1
	}
}

func (blockchain *BlockChain) processInstructionFromBeaconV2(beaconBlocks []*BeaconBlock, shardID byte, committeeChange *committeeChange) ([]string, []string, map[string]string) {
	newShardPendingValidator := []string{}
	shardPendingValidator, err := incognitokey.CommitteeKeyListToString(blockchain.BestState.Shard[shardID].ShardPendingValidator)
	if err != nil {
		panic(err)
	}
	assignInstructions := [][]string{}
	stakingTx := make(map[string]string)
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			// Process Assign Instruction
			if l[0] == AssignAction && l[2] == "shard" {
				if strings.Compare(l[3], strconv.Itoa(int(shardID))) == 0 {
					tempNewShardPendingValidator := strings.Split(l[1], ",")
					shardPendingValidator = append(shardPendingValidator, tempNewShardPendingValidator...)
					newShardPendingValidator = append(newShardPendingValidator, tempNewShardPendingValidator...)
					assignInstructions = append(assignInstructions, l)
				}
			}
			// Get Staking Tx
			// assume that stake instruction already been validated by beacon committee
			if l[0] == StakeAction && l[2] == "beacon" {
				beacon := strings.Split(l[1], ",")
				newBeaconCandidates := []string{}
				newBeaconCandidates = append(newBeaconCandidates, beacon...)
				if len(l) == 6 {
					for i, v := range strings.Split(l[3], ",") {
						txHash, err := common.Hash{}.NewHashFromStr(v)
						if err != nil {
							continue
						}
						txShardID, _, _, _, err := blockchain.GetTransactionByHash(*txHash)
						if err != nil {
							continue
						}
						if txShardID != shardID {
							continue
						}
						// if transaction belong to this shard then add to shard beststate
						stakingTx[newBeaconCandidates[i]] = v
					}
				}
			}
			if l[0] == StakeAction && l[2] == "shard" {
				shard := strings.Split(l[1], ",")
				newShardCandidates := []string{}
				newShardCandidates = append(newShardCandidates, shard...)
				if len(l) == 6 {
					for i, v := range strings.Split(l[3], ",") {
						txHash, err := common.Hash{}.NewHashFromStr(v)
						if err != nil {
							continue
						}
						txShardID, _, _, _, err := blockchain.GetTransactionByHash(*txHash)
						if err != nil {
							continue
						}
						if txShardID != shardID {
							continue
						}
						// if transaction belong to this shard then add to shard beststate
						stakingTx[newShardCandidates[i]] = v
					}
				}
			}
		}
	}
	return shardPendingValidator, newShardPendingValidator, stakingTx
}

// getTransactionForNewBlockV2 get transaction for new block
// 1. Get pending transaction from blockgen
// 2. Keep valid tx & Removed error tx
// 3. Build response Transaction For Shard
// 4. Build response Transaction For Beacon
// 5. Return valid transaction from pending, response transactions from shard and beacon
func (blockGenerator *BlockGenerator) getTransactionForNewBlockV2(privatekey *privacy.PrivateKey, shardID byte, db incdb.Database, beaconBlocks []*BeaconBlock, blockCreation int64, beaconHeight uint64) ([]metadata.Transaction, error) {
	txsToAdd, txToRemove, _ := blockGenerator.getPendingTransaction(shardID, beaconBlocks, blockCreation, beaconHeight)
	if len(txsToAdd) == 0 {
		Logger.log.Info("Creating empty block...")
	}
	go blockGenerator.txPool.RemoveTx(txToRemove, false)
	var responseTxsBeacon []metadata.Transaction
	var errInstructions [][]string
	var cError chan error
	cError = make(chan error)
	go func() {
		var err error
		responseTxsBeacon, errInstructions, err = blockGenerator.buildResponseTxsFromBeaconInstructionsV2(beaconBlocks, privatekey, shardID)
		cError <- err
	}()
	nilCount := 0
	for {
		err := <-cError
		if err != nil {
			return nil, err
		}
		nilCount++
		if nilCount == 1 {
			break
		}
	}
	txsToAdd = append(txsToAdd, responseTxsBeacon...)
	if len(errInstructions) > 0 {
		Logger.log.Error("List error instructions, which can not create tx", errInstructions)
	}
	return txsToAdd, nil
}

// buildResponseTxsFromBeaconInstructions builds response txs from beacon instructions
func (blockGenerator *BlockGenerator) buildResponseTxsFromBeaconInstructionsV2(beaconBlocks []*BeaconBlock, producerPrivateKey *privacy.PrivateKey, shardID byte) ([]metadata.Transaction, [][]string, error) {
	responsedTxs := []metadata.Transaction{}
	responsedHashTxs := []common.Hash{} // capture hash of responsed tx
	errorInstructions := [][]string{}   // capture error instruction -> which instruction can not create tx
	tempAutoStakingM := make(map[uint64]map[string]bool)
	for _, beaconBlock := range beaconBlocks {
		autoStaking, ok := tempAutoStakingM[beaconBlock.Header.Height]
		if !ok {
			consensusStateRootHash, ok := blockGenerator.chain.BestState.Beacon.GetConsensusStateRootHash(beaconBlock.Header.Height)
			if !ok {
				return []metadata.Transaction{}, errorInstructions, NewBlockChainError(FetchAutoStakingByHeightError, fmt.Errorf("can't get ConsensusStateRootHash of height %+v", beaconBlock.Header.Height))
			}
			consensusStateDB, err := statedb.NewWithPrefixTrie(consensusStateRootHash, statedb.NewDatabaseAccessWarper(blockGenerator.chain.GetDatabase()))
			if err != nil {
				return []metadata.Transaction{}, errorInstructions, NewBlockChainError(FetchAutoStakingByHeightError, err)
			}
			_, newAutoStaking := statedb.GetRewardReceiverAndAutoStaking(consensusStateDB, blockGenerator.chain.GetShardIDs())
			tempAutoStakingM[beaconBlock.Header.Height] = newAutoStaking
			autoStaking = newAutoStaking
		}
		for _, l := range beaconBlock.Body.Instructions {
			if l[0] == SwapAction {
				for _, outPublicKeys := range strings.Split(l[2], ",") {
					// If out public key has auto staking then ignore this public key
					if _, ok := autoStaking[outPublicKeys]; ok {
						continue
					}
					tx, err := blockGenerator.buildReturnStakingAmountTxV2(outPublicKeys, producerPrivateKey, shardID)
					if err != nil {
						Logger.log.Error(err)
						continue
					}
					txHash := *tx.Hash()
					if ok, _ := common.SliceExists(responsedHashTxs, txHash); ok {
						data, _ := json.Marshal(tx)
						Logger.log.Error("Double tx from instruction", l, string(data))
						errorInstructions = append(errorInstructions, l)
						continue
					}
					responsedTxs = append(responsedTxs, tx)
					responsedHashTxs = append(responsedHashTxs, txHash)
				}

			}
			if l[0] == StakeAction || l[0] == RandomAction || l[0] == AssignAction || l[0] == SwapAction {
				continue
			}
			if len(l) <= 2 {
				continue
			}
			metaType, err := strconv.Atoi(l[0])
			if err != nil {
				return nil, nil, err
			}
			var newTx metadata.Transaction
			switch metaType {
			case metadata.IssuingETHRequestMeta:
				if len(l) >= 4 && l[2] == "accepted" {
					newTx, err = blockGenerator.buildETHIssuanceTxV2(l[3], producerPrivateKey, shardID)
				}
			case metadata.IssuingRequestMeta:
				if len(l) >= 4 && l[2] == "accepted" {
					newTx, err = blockGenerator.buildIssuanceTxV2(l[3], producerPrivateKey, shardID)
				}
			case metadata.PDETradeRequestMeta:
				if len(l) >= 4 {
					newTx, err = blockGenerator.buildPDETradeIssuanceTxV2(l[2], l[3], producerPrivateKey, shardID)
				}
			case metadata.PDEWithdrawalRequestMeta:
				if len(l) >= 4 && l[2] == common.PDEWithdrawalAcceptedChainStatus {
					newTx, err = blockGenerator.buildPDEWithdrawalTxV2(l[3], producerPrivateKey, shardID)
				}
			case metadata.PDEContributionMeta:
				if len(l) >= 4 {
					if l[2] == common.PDEContributionRefundChainStatus {
						newTx, err = blockGenerator.buildPDERefundContributionTxV2(l[3], producerPrivateKey, shardID)
					} else if l[2] == common.PDEContributionMatchedNReturnedChainStatus {
						newTx, err = blockGenerator.buildPDEMatchedNReturnedContributionTxV2(l[3], producerPrivateKey, shardID)
					}
				}

			default:
				continue
			}
			if err != nil {
				return nil, nil, err
			}
			if newTx != nil {
				newTxHash := *newTx.Hash()
				if ok, _ := common.SliceExists(responsedHashTxs, newTxHash); ok {
					data, _ := json.Marshal(newTx)
					Logger.log.Error("Double tx from instruction", l, string(data))
					errorInstructions = append(errorInstructions, l)
					//continue
				}
				responsedTxs = append(responsedTxs, newTx)
				responsedHashTxs = append(responsedHashTxs, newTxHash)
			}
		}
	}
	return responsedTxs, errorInstructions, nil
}
