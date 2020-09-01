package blockchain

import (
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/blockchain/committeestate"
	"github.com/incognitochain/incognito-chain/blockchain/types"
	"github.com/incognitochain/incognito-chain/instruction"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/incognitokey"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/privacy"
	"github.com/incognitochain/incognito-chain/transaction"
)

// NewBlockShard Create New block Shard:
//	1. Identify Beacon State for this Shard Block: Beacon Hash & Beacon Height & Epoch
//		+ Get Beacon Block (B) from Beacon Best State (from Beacon Chain of Shard Node)
//		+ Beacon Block (B) must have the same epoch With New Shard Block (S):
//		+ If Beacon Block (B) have different height previous shard block PS (previous of S)
//		Then Beacon Block (B) epoch greater than Shard Block (S) epoch exact 1 value
//		BUT This only works if Shard Best State have the Beacon Height divisible by epoch
//		+ Ex: 1 epoch has 50 block
//		Example 1:
//			shard block with
//				height 10,
//				epoch: 1,
//				beacon block height: 49
//			then shard block with
//				height 11 must have
//				epoch: 1,
//				beacon block height: must be 49 or 50
//		Example 2:
//			shard block with
//				height 10,
//				epoch: 1,
//				beacon block height: 50
//			then shard block with
//				height is 11 can have 2 option:
//				a. epoch: 1, if beacon block height remain 50
//				b. epoch: 2, and beacon block must in range from 51-100
//				Can have beacon block with height > 100
//	2. Build Shard Block Body:
//		a. Get Cross Transaction from other shard && Build Cross Shard Tx Custom Token Transaction (if exist)
//		b. Get Transactions for New Block
//		c. Process Assign Instructions from Beacon Blocks
//		c. Generate Instructions
//	3. Build Shard Block Essential Data for Header
//	4. Update Cloned ShardBestState with New Shard Block
//	5. Create Root Hash from New Shard Block and updated Clone Shard Beststate Data
func (blockchain *BlockChain) NewBlockShard(curView *ShardBestState, version int, proposer string, round int, start time.Time) (*types.ShardBlock, error) {
	var (
		transactionsForNewBlock = make([]metadata.Transaction, 0)
		totalTxsFee             = make(map[common.Hash]uint64)
		newShardBlock           = types.NewShardBlock()
		shardInstructions       = [][]string{}
		isOldBeaconHeight       = false
		tempPrivateKey          = blockchain.config.BlockGen.createTempKeyset()
		shardBestState          = NewShardBestState()
		beaconHeight            = blockchain.BeaconChain.GetFinalView().GetHeight()
		shardID                 = curView.ShardID
	)
	Logger.log.Criticalf("⛏ Creating Shard Block %+v", curView.ShardHeight+1)

	currentPendingValidators := curView.GetShardPendingValidator()

	currentCommitteePubKeys, err := incognitokey.
		CommitteeKeyListToString(curView.GetCommittee())
	if err != nil {
		return nil, err
	}

	//========Verify newShardBlock with previous best state
	// Get Beststate of previous newShardBlock == previous best state
	// Clone best state value into new variable
	// // startStep := time.Now()
	if err := shardBestState.cloneShardBestStateFrom(curView); err != nil {
		return nil, err
	}

	//==========Fetch Beacon Blocks============
	// // startStep = time.Now()
	BLogger.log.Infof("Producing block: %d", curView.ShardHeight+1)
	if beaconHeight-shardBestState.BeaconHeight > MAX_BEACON_BLOCK {
		beaconHeight = shardBestState.BeaconHeight + MAX_BEACON_BLOCK
	}

	beaconHash, err := rawdbv2.GetFinalizedBeaconBlockHashByIndex(blockchain.GetBeaconChainDatabase(), beaconHeight)
	if err != nil {
		fmt.Println("cannot find beacon block ", beaconHeight)
		return nil, NewBlockChainError(FetchBeaconBlockHashError, err)
	}

	beaconBlockBytes, err := rawdbv2.GetBeaconBlockByHash(blockchain.GetBeaconChainDatabase(), *beaconHash)
	if err != nil {
		return nil, err
	}

	beaconBlock := types.BeaconBlock{}
	err = json.Unmarshal(beaconBlockBytes, &beaconBlock)
	if err != nil {
		return nil, err
	}
	epoch := beaconBlock.Header.Epoch
	if epoch-shardBestState.Epoch >= 1 {
		beaconHeight = shardBestState.Epoch * blockchain.config.ChainParams.Epoch
		newBeaconHash, err := rawdbv2.GetFinalizedBeaconBlockHashByIndex(blockchain.GetBeaconChainDatabase(), beaconHeight)
		if err != nil {
			return nil, NewBlockChainError(FetchBeaconBlockHashError, err)
		}
		copy(beaconHash[:], newBeaconHash.GetBytes())
		epoch = shardBestState.Epoch + 1
	}
	Logger.log.Infof("Get Beacon Block With Height %+v, Shard BestState %+v", beaconHeight, shardBestState.BeaconHeight)
	//Fetch beacon block from height
	beaconBlocks, err := FetchBeaconBlockFromHeight(blockchain, shardBestState.BeaconHeight+1, beaconHeight)
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
	crossTransactions := blockchain.config.BlockGen.getCrossShardData(shardID, shardBestState.BeaconHeight, beaconHeight)
	Logger.log.Critical("Cross Transaction: ", crossTransactions)
	// Get Transaction for new block
	// // startStep = time.Now()
	blockCreationLeftOver := curView.BlockMaxCreateTime.Nanoseconds() - time.Since(start).Nanoseconds()
	txsToAddFromBlock, err := blockchain.config.BlockGen.getTransactionForNewBlock(curView, &tempPrivateKey, shardID, beaconBlocks, blockCreationLeftOver, beaconHeight)
	if err != nil {
		return nil, err
	}
	transactionsForNewBlock = append(transactionsForNewBlock, txsToAddFromBlock...)
	// build txs with metadata
	transactionsForNewBlock, err = blockchain.BuildResponseTransactionFromTxsWithMetadata(curView, transactionsForNewBlock, &tempPrivateKey, shardID)
	// process instruction from beacon block

	beaconInstructions, _, err := blockchain.
		preProcessInstructionFromBeacon(beaconBlocks, shardBestState.ShardID)
	if err != nil {
		return nil, err
	}

	shardPendingValidatorStr, err := incognitokey.
		CommitteeKeyListToString(currentPendingValidators)
	if err != nil {
		return nil, err
	}

	env := committeestate.
		NewShardEnvBuilder().
		BuildBeaconInstructions(beaconInstructions).
		BuildShardID(curView.ShardID).
		Build()

	committeeChange, err := curView.shardCommitteeEngine.ProcessInstructionFromBeacon(env)
	if err != nil {
		return nil, err
	}

	curView.shardCommitteeEngine.AbortUncommittedShardState()

	currentPendingValidators, err = updateCommiteesWithAddedAndRemovedListValidator(currentPendingValidators,
		committeeChange.ShardSubstituteAdded[curView.ShardID],
		committeeChange.ShardSubstituteRemoved[curView.ShardID])

	shardPendingValidatorStr, err = incognitokey.CommitteeKeyListToString(currentPendingValidators)
	if err != nil {
		return nil, NewBlockChainError(ProcessInstructionFromBeaconError, err)
	}

	shardInstructions, _, _, err = blockchain.generateInstruction(curView, shardID,
		beaconHeight, isOldBeaconHeight, beaconBlocks, beaconInstructions,
		shardPendingValidatorStr, currentCommitteePubKeys)
	if err != nil {
		return nil, NewBlockChainError(GenerateInstructionError, err)
	}

	if len(shardInstructions) != 0 {
		Logger.log.Info("Shard Producer: Instruction", shardInstructions)
	}

	newShardBlock.BuildShardBlockBody(shardInstructions, crossTransactions, transactionsForNewBlock)
	//==========Build Essential Header Data=========
	// producer key
	producerKey := proposer
	producerPubKeyStr := proposer

	for _, tx := range newShardBlock.Body.Transactions {
		totalTxsFee[*tx.GetTokenID()] += tx.GetTxFee()
		txType := tx.GetType()
		if txType == common.TxCustomTokenPrivacyType {
			txCustomPrivacy := tx.(*transaction.TxCustomTokenPrivacy)
			totalTxsFee[*txCustomPrivacy.GetTokenID()] = txCustomPrivacy.GetTxFeeToken()
		}
	}
	newShardBlock.Header = types.ShardHeader{
		Producer:          producerKey, //committeeMiningKeys[producerPosition],
		ProducerPubKeyStr: producerPubKeyStr,
		ShardID:           shardID,
		Version:           version,
		PreviousBlockHash: shardBestState.BestBlockHash,
		Height:            shardBestState.ShardHeight + 1,
		Round:             round,
		Epoch:             epoch,
		CrossShardBitMap:  CreateCrossShardByteArray(newShardBlock.Body.Transactions, shardID),
		BeaconHeight:      beaconHeight,
		BeaconHash:        *beaconHash,
		TotalTxsFee:       totalTxsFee,
		ConsensusType:     curView.ConsensusAlgorithm,
	}
	//============Update Shard BestState=============
	// startStep = time.Now()
	newShardBestState, hashes, _, err := shardBestState.updateShardBestState(blockchain, newShardBlock, beaconBlocks)
	if err != nil {
		return nil, err
	}
	curView.shardCommitteeEngine.AbortUncommittedShardState()
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
	txInstructions, err := CreateShardInstructionsFromTransactionAndInstruction(newShardBlock.Body.Transactions, blockchain, shardID)
	if err != nil {
		return nil, err
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
		return nil, NewBlockChainError(InstructionsHashError, err)
	}

	stakingTxRoot, err := generateHashFromMapStringString(newShardBestState.StakingTx.Data())
	if err != nil {
		return nil, NewBlockChainError(StakingTxHashError, err)
	}
	// Instruction merkle root
	flattenTxInsts, err := FlattenAndConvertStringInst(txInstructions)
	if err != nil {
		return nil, NewBlockChainError(FlattenAndConvertStringInstError, fmt.Errorf("Instruction from Tx: %+v", err))
	}
	flattenInsts, err := FlattenAndConvertStringInst(shardInstructions)
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
	newShardBlock.Header.CommitteeRoot = hashes.ShardCommitteeHash
	newShardBlock.Header.PendingValidatorRoot = hashes.ShardSubstituteHash
	newShardBlock.Header.StakingTxRoot = stakingTxRoot
	newShardBlock.Header.Timestamp = start.Unix()
	copy(newShardBlock.Header.InstructionMerkleRoot[:], instMerkleRoot)
	return newShardBlock, nil
}

// getTransactionForNewBlock get transaction for new block
// 1. Get pending transaction from blockgen
// 2. Keep valid tx & Removed error tx
// 3. Build response Transaction For Shard
// 4. Build response Transaction For Beacon
// 5. Return valid transaction from pending, response transactions from shard and beacon
func (blockGenerator *BlockGenerator) getTransactionForNewBlock(curView *ShardBestState, privatekey *privacy.PrivateKey, shardID byte, beaconBlocks []*types.BeaconBlock, blockCreation int64, beaconHeight uint64) ([]metadata.Transaction, error) {
	txsToAdd, txToRemove, _ := blockGenerator.getPendingTransaction(shardID, beaconBlocks, blockCreation, beaconHeight, curView)
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
		responseTxsBeacon, errInstructions, err = blockGenerator.buildResponseTxsFromBeaconInstructions(curView, beaconBlocks, privatekey, shardID)
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
func (blockGenerator *BlockGenerator) buildResponseTxsFromBeaconInstructions(curView *ShardBestState, beaconBlocks []*types.BeaconBlock, producerPrivateKey *privacy.PrivateKey, shardID byte) ([]metadata.Transaction, [][]string, error) {
	responsedTxs := []metadata.Transaction{}
	responsedHashTxs := []common.Hash{} // capture hash of responsed tx
	errorInstructions := [][]string{}   // capture error instruction -> which instruction can not create tx
	beaconView := blockGenerator.chain.BeaconChain.GetFinalView().(*BeaconBestState)
	//TODO: Please check this logic again, why PDE, Bridge build from old beacon block but get info from beacon final view
	for _, beaconBlock := range beaconBlocks {
		for _, inst := range beaconBlock.Body.Instructions {
			if instruction.IsConsensusInstruction(inst[0]) {
				continue
			}
			if len(inst) <= 2 {
				continue
			}
			metaType, err := strconv.Atoi(inst[0])
			if err != nil {
				return nil, nil, err
			}
			var newTx metadata.Transaction
			switch metaType {
			case metadata.IssuingETHRequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildETHIssuanceTx(inst[3], producerPrivateKey, shardID, curView, beaconView)
				}
			case metadata.IssuingRequestMeta:
				if len(inst) >= 4 && inst[2] == "accepted" {
					newTx, err = blockGenerator.buildIssuanceTx(inst[3], producerPrivateKey, shardID, curView, beaconView)
				}
			case metadata.PDETradeRequestMeta:
				if len(inst) >= 4 {
					newTx, err = blockGenerator.buildPDETradeIssuanceTx(inst[2], inst[3], producerPrivateKey, shardID, curView, beaconView)
				}
			case metadata.PDECrossPoolTradeRequestMeta:
				if len(inst) >= 4 {
					newTx, err = blockGenerator.buildPDECrossPoolTradeIssuanceTx(inst[2], inst[3], producerPrivateKey, shardID, curView, beaconView)
				}
			case metadata.PDEWithdrawalRequestMeta:
				if len(inst) >= 4 && inst[2] == common.PDEWithdrawalAcceptedChainStatus {
					newTx, err = blockGenerator.buildPDEWithdrawalTx(inst[3], producerPrivateKey, shardID, curView, beaconView)
				}
			case metadata.PDEFeeWithdrawalRequestMeta:
				if len(inst) >= 4 && inst[2] == common.PDEFeeWithdrawalAcceptedChainStatus {
					newTx, err = blockGenerator.buildPDEFeeWithdrawalTx(inst[3], producerPrivateKey, shardID, curView, beaconView)
				}
			case metadata.PDEContributionMeta, metadata.PDEPRVRequiredContributionRequestMeta:
				if len(inst) >= 4 {
					if inst[2] == common.PDEContributionRefundChainStatus {
						newTx, err = blockGenerator.buildPDERefundContributionTx(inst[3], producerPrivateKey, shardID, curView, beaconView)
					} else if inst[2] == common.PDEContributionMatchedNReturnedChainStatus {
						newTx, err = blockGenerator.buildPDEMatchedNReturnedContributionTx(inst[3], producerPrivateKey, shardID, curView, beaconView)
					}
				}
			// portal
			case metadata.PortalUserRegisterMeta:
				if len(inst) >= 4 && inst[2] == common.PortalPortingRequestRejectedChainStatus {
					newTx, err = curView.buildPortalRefundPortingFeeTx(inst[3], producerPrivateKey, shardID)
				}
			case metadata.PortalCustodianDepositMeta:
				if len(inst) >= 4 && inst[2] == common.PortalCustodianDepositRefundChainStatus {
					newTx, err = curView.buildPortalRefundCustodianDepositTx(inst[3], producerPrivateKey, shardID)
				}
			case metadata.PortalUserRequestPTokenMeta:
				if len(inst) >= 4 && inst[2] == common.PortalReqPTokensAcceptedChainStatus {
					newTx, err = curView.buildPortalAcceptedRequestPTokensTx(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
				}
				//custodian withdraw
			case metadata.PortalCustodianWithdrawRequestMeta:
				if len(inst) >= 4 && inst[2] == common.PortalCustodianWithdrawRequestAcceptedStatus {
					newTx, err = curView.buildPortalCustodianWithdrawRequest(inst[3], producerPrivateKey, shardID)
				}
			case metadata.PortalRedeemRequestMeta:
				if len(inst) >= 4 && (inst[2] == common.PortalRedeemRequestRejectedChainStatus || inst[2] == common.PortalRedeemReqCancelledByLiquidationChainStatus) {
					newTx, err = curView.buildPortalRejectedRedeemRequestTx(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
				}
				//liquidation: redeem ptoken
			case metadata.PortalRedeemLiquidateExchangeRatesMeta:
				if len(inst) >= 4 {
					if inst[2] == common.PortalRedeemLiquidateExchangeRatesSuccessChainStatus {
						newTx, err = curView.buildPortalRedeemLiquidateExchangeRatesRequestTx(inst[3], producerPrivateKey, shardID)
					} else if inst[2] == common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus {
						newTx, err = curView.buildPortalRefundRedeemLiquidateExchangeRatesTx(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
					}
				}
			case metadata.PortalLiquidateCustodianMeta:
				if len(inst) >= 4 && inst[2] == common.PortalLiquidateCustodianSuccessChainStatus {
					newTx, err = curView.buildPortalLiquidateCustodianResponseTx(inst[3], producerPrivateKey, shardID)
				}
			case metadata.PortalRequestWithdrawRewardMeta:
				if len(inst) >= 4 && inst[2] == common.PortalReqWithdrawRewardAcceptedChainStatus {
					newTx, err = curView.buildPortalAcceptedWithdrawRewardTx(blockGenerator.chain.GetBeaconBestState(), inst[3], producerPrivateKey, shardID)
				}
				//liquidation: custodian deposit
			case metadata.PortalLiquidationCustodianDepositMeta:
				if len(inst) >= 4 && inst[2] == common.PortalLiquidationCustodianDepositRejectedChainStatus {
					newTx, err = curView.buildPortalLiquidationCustodianDepositReject(inst[3], producerPrivateKey, shardID)
				}
			case metadata.PortalLiquidationCustodianDepositMetaV2:
				if len(inst) >= 4 && inst[2] == common.PortalLiquidationCustodianDepositRejectedChainStatus {
					newTx, err = curView.buildPortalLiquidationCustodianDepositRejectV2(inst[3], producerPrivateKey, shardID)
				}
			//
			case metadata.PortalTopUpWaitingPortingRequestMeta:
				if len(inst) >= 4 && inst[2] == common.PortalTopUpWaitingPortingRejectedChainStatus {
					newTx, err = curView.buildPortalRejectedTopUpWaitingPortingTx(inst[3], producerPrivateKey, shardID)
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
					Logger.log.Error("Double tx from instruction", inst, string(data))
					errorInstructions = append(errorInstructions, inst)
					//continue
				}
				responsedTxs = append(responsedTxs, newTx)
				responsedHashTxs = append(responsedHashTxs, newTxHash)
			}
		}
	}
	returnStakingTxs, errIns, err := blockGenerator.chain.buildReturnStakingTxFromBeaconInstructions(
		curView,
		beaconBlocks,
		producerPrivateKey,
		shardID,
	)
	if err != nil {
		return nil, nil, err
	}
	responsedTxs = append(responsedTxs, returnStakingTxs...)
	errorInstructions = append(errorInstructions, errIns...)
	return responsedTxs, errorInstructions, nil
}

//	Generate Instruction generate instruction for new shard block
//	- Swap: at the end of beacon epoch
//	- Brigde: at the end of beacon epoch
//	Return params:
//	#1: instruction list
//	#2: shardpendingvalidator
//	#3: shardcommittee
//	#4: error
func (blockchain *BlockChain) generateInstruction(view *ShardBestState,
	shardID byte, beaconHeight uint64,
	isOldBeaconHeight bool, beaconBlocks []*types.BeaconBlock, beaconInstructions [][]string,
	shardPendingValidator []string, shardCommittee []string) ([][]string, []string, []string, error) {
	var (
		instructions                      = [][]string{}
		bridgeSwapConfirmInst             = []string{}
		swapOrConfirmShardSwapInstruction = []string{}
		confirmShardSwapInstruction       = &instruction.ConfirmShardSwapInstruction{}
		err                               error
	)
	// if this beacon height has been seen already then DO NOT generate any more instruction
	if beaconHeight%blockchain.config.ChainParams.Epoch == 0 && isOldBeaconHeight == false {
		backupShardCommittee := shardCommittee
		fixedProducerShardValidators := shardCommittee[:NumberOfFixedBlockValidators]
		shardCommittee = shardCommittee[NumberOfFixedBlockValidators:]
		Logger.log.Info("ShardPendingValidator", shardPendingValidator)
		Logger.log.Info("ShardCommittee", shardCommittee)
		Logger.log.Info("MaxShardCommitteeSize", view.MaxShardCommitteeSize)
		Logger.log.Info("ShardID", shardID)

		maxShardCommitteeSize := view.MaxShardCommitteeSize - NumberOfFixedBlockValidators
		var minShardCommitteeSize int
		if view.MinShardCommitteeSize-NumberOfFixedBlockValidators < 0 {
			minShardCommitteeSize = 0
		} else {
			minShardCommitteeSize = view.MinShardCommitteeSize - NumberOfFixedBlockValidators
		}
		if common.IndexOfUint64(beaconHeight/blockchain.config.ChainParams.Epoch, blockchain.config.ChainParams.EpochBreakPointSwapNewKey) > -1 {
			epoch := beaconHeight / blockchain.config.ChainParams.Epoch
			swapOrConfirmShardSwapInstruction, shardCommittee = CreateShardSwapActionForKeyListV2(blockchain.config.GenesisParams, backupShardCommittee, NumberOfFixedBlockValidators, blockchain.config.ChainParams.ActiveShards, shardID, epoch)
		} else {
			tempSwapInstruction := instruction.NewSwapInstruction()
			env := committeestate.NewShardEnvBuilder().
				BuildMaxShardCommitteeSize(maxShardCommitteeSize).
				BuildMinShardCommitteeSize(minShardCommitteeSize).
				BuildShardID(shardID).
				BuildOffset(blockchain.config.ChainParams.Offset).
				BuildSwapOffset(blockchain.config.ChainParams.SwapOffset).
				Build()
			tempSwapInstruction, shardPendingValidator, shardCommittee, err = view.shardCommitteeEngine.GenerateSwapInstruction(env)
			if err != nil {
				Logger.log.Error(err)
				return instructions, shardPendingValidator, shardCommittee, err
			}
			swapOrConfirmShardSwapInstruction = tempSwapInstruction.ToString()
			shardCommittee = append(fixedProducerShardValidators, shardCommittee...)
		}
		env := committeestate.NewShardEnvBuilder().
			BuildBeaconInstructions(beaconInstructions).
			BuildShardID(shardID).
			BuildNumberOfFixedBlockValidators(NumberOfFixedBlockValidators).
			Build()
		confirmShardSwapInstruction, shardCommittee, err = view.shardCommitteeEngine.GenerateConfirmShardSwapInstruction(env)
		if err != nil {
			return instructions, shardPendingValidator, shardCommittee, err
		}
		if !confirmShardSwapInstruction.IsEmpty() {
			swapOrConfirmShardSwapInstruction = confirmShardSwapInstruction.ToString()
		}
		// NOTE: shardCommittee must be finalized before building Bridge instruction here
		// shardCommittee must include all producers and validators in the right order
		// Generate instruction storing merkle root of validators pubkey and send to beacon
		bridgeID := byte(common.BridgeShardID)
		if shardID == bridgeID && committeeChanged(swapOrConfirmShardSwapInstruction) {
			blockHeight := view.ShardHeight + 1
			bridgeSwapConfirmInst, err = buildBridgeSwapConfirmInstruction(shardCommittee, blockHeight)
			if err != nil {
				BLogger.log.Error(err)
				return instructions, shardPendingValidator, shardCommittee, err
			}
			BLogger.log.Infof("Add Bridge swap inst in ShardID %+v block %d", shardID, blockHeight)
		}
	}

	if len(swapOrConfirmShardSwapInstruction) > 0 {
		instructions = append(instructions, swapOrConfirmShardSwapInstruction)
	}

	if len(bridgeSwapConfirmInst) > 0 {
		instructions = append(instructions, bridgeSwapConfirmInst)
		Logger.log.Infof("Build bridge swap confirm inst: %s \n", bridgeSwapConfirmInst)
	}
	// Pick BurningConfirm inst and save to bridge block
	bridgeID := byte(common.BridgeShardID)
	if shardID == bridgeID {
		prevBlock := view.BestBlock
		height := view.ShardHeight + 1
		confirmInsts := pickBurningConfirmInstruction(beaconBlocks, height)
		if len(confirmInsts) > 0 {
			bid := []uint64{}
			for _, b := range beaconBlocks {
				bid = append(bid, b.Header.Height)
			}
			Logger.log.Infof("Picked burning confirm inst: %s %d %v\n", confirmInsts, prevBlock.Header.Height+1, bid)
			instructions = append(instructions, confirmInsts...)
		}
	}

	return instructions, shardPendingValidator, shardCommittee, nil
}

// getCrossShardData get cross shard data from cross shard block
//  1. Get Cross Shard Block and Validate
//	  a. Get Valid Cross Shard Block from Cross Shard Pool
//	  b. Get Current Cross Shard State: Last Cross Shard Block From other Shard (FS) to this shard (TS) (Ex: last cross shard block from Shard 0 to Shard 1)
//	  c. Get Next Cross Shard Block Height from other Shard (FS) to this shard (TS)
//     + Using FetchCrossShardNextHeight function in Database to determine next block height
//	  d. Fetch Other Shard (FS) Committee at Next Cross Shard Block Height for Validation
//  2. Validate
//	  a. Get Next Cross Shard Height from Database
//	  b. Cross Shard Block Height is Next Cross Shard Height from Database (if miss Cross Shard Block according to beacon bytemap then stop discard the rest)
//	  c. Verify Cross Shard Block Signature
//  3. After validation:
//	  - Process valid block to extract:
//	   + Cross output coin
//	   + Cross Normal Token
func (blockGenerator *BlockGenerator) getCrossShardData(toShard byte, lastBeaconHeight uint64, currentBeaconHeight uint64) map[byte][]types.CrossTransaction {
	crossTransactions := make(map[byte][]types.CrossTransaction)
	// get cross shard block
	var allCrossShardBlock = make([][]*types.CrossShardBlock, blockGenerator.chain.config.ChainParams.ActiveShards)
	for sid, v := range blockGenerator.syncker.GetCrossShardBlocksForShardProducer(toShard) {
		for _, b := range v {
			allCrossShardBlock[sid] = append(allCrossShardBlock[sid], b.(*types.CrossShardBlock))
		}
	}
	// allCrossShardBlock => already short
	for _, crossShardBlock := range allCrossShardBlock {
		for _, blk := range crossShardBlock {
			crossTransaction := types.CrossTransaction{
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

/*
	Verify Transaction with these condition: defined in mempool.go
*/
func (blockGenerator *BlockGenerator) getPendingTransaction(
	shardID byte,
	beaconBlocks []*types.BeaconBlock,
	blockCreationTimeLeftOver int64,
	beaconHeight uint64,
	curView *ShardBestState,
) (txsToAdd []metadata.Transaction, txToRemove []metadata.Transaction, totalFee uint64) {
	spareTime := SpareTime * time.Millisecond
	maxBlockCreationTimeLeftTime := blockCreationTimeLeftOver - spareTime.Nanoseconds()
	startTime := time.Now()
	sourceTxns := blockGenerator.GetPendingTxsV2()
	var elasped int64
	Logger.log.Info("Number of transaction get from Block Generator: ", len(sourceTxns))
	isEmpty := blockGenerator.chain.config.TempTxPool.EmptyPool()
	if !isEmpty {
		return []metadata.Transaction{}, []metadata.Transaction{}, 0
	}
	currentSize := uint64(0)
	preparedTxForNewBlock := []metadata.Transaction{}
	for _, tx := range sourceTxns {
		tempSize := tx.GetTxActualSize()
		if currentSize+tempSize >= common.MaxBlockSize {
			break
		}
		preparedTxForNewBlock = append(preparedTxForNewBlock, tx)
		elasped = time.Since(startTime).Nanoseconds()
		if elasped >= maxBlockCreationTimeLeftTime {
			Logger.log.Info("Shard Producer/Elapsed, Break: ", elasped)
			break
		}
	}
	listBatchTxs := []metadata.Transaction{}
	for index, tx := range preparedTxForNewBlock {
		elasped = time.Since(startTime).Nanoseconds()
		if elasped >= maxBlockCreationTimeLeftTime {
			Logger.log.Info("Shard Producer/Elapsed, Break: ", elasped)
			break
		}
		listBatchTxs = append(listBatchTxs, tx)
		if ((index+1)%TransactionBatchSize == 0) || (index == len(preparedTxForNewBlock)-1) {
			tempTxDesc, err := blockGenerator.chain.config.TempTxPool.MaybeAcceptBatchTransactionForBlockProducing(shardID, listBatchTxs, int64(beaconHeight), curView)
			if err != nil {
				Logger.log.Errorf("SHARD %+v | Verify Batch Transaction for new block error %+v", shardID, err)
				for _, tx2 := range listBatchTxs {
					if blockGenerator.chain.config.TempTxPool.HaveTransaction(tx2.Hash()) {
						continue
					}
					txShardID := common.GetShardIDFromLastByte(tx2.GetSenderAddrLastByte())
					if txShardID != shardID {
						continue
					}
					tempTxDesc, err := blockGenerator.chain.config.TempTxPool.MaybeAcceptTransactionForBlockProducing(tx2, int64(beaconHeight), curView)
					if err != nil {
						txToRemove = append(txToRemove, tx2)
						continue
					}
					tempTx := tempTxDesc.Tx
					totalFee += tempTx.GetTxFee()
					tempSize := tempTx.GetTxActualSize()
					if currentSize+tempSize >= common.MaxBlockSize {
						break
					}
					currentSize += tempSize
					txsToAdd = append(txsToAdd, tempTx)
				}
			}
			for _, tempToAddTxDesc := range tempTxDesc {
				tempTx := tempToAddTxDesc.Tx
				totalFee += tempTx.GetTxFee()
				tempSize := tempTx.GetTxActualSize()
				if currentSize+tempSize >= common.MaxBlockSize {
					break
				}
				currentSize += tempSize
				txsToAdd = append(txsToAdd, tempTx)
			}
			// reset list batch and add new txs
			listBatchTxs = []metadata.Transaction{}

		} else {
			continue
		}
	}
	Logger.log.Criticalf(" 🔎 %+v transactions for New Block from pool \n", len(txsToAdd))
	blockGenerator.chain.config.TempTxPool.EmptyPool()
	return txsToAdd, txToRemove, totalFee
}

func (blockGenerator *BlockGenerator) createTempKeyset() privacy.PrivateKey {
	rand.Seed(time.Now().UnixNano())
	seed := make([]byte, 16)
	rand.Read(seed)
	return privacy.GeneratePrivateKey(seed)
}

//preProcessInstructionFromBeacon : preprcess for beacon instructions before move to handle it in committee state
// Store stakingtx address and return it back to outside
// Only process for instruction not stake instruction
func (blockchain *BlockChain) preProcessInstructionFromBeacon(
	beaconBlocks []*types.BeaconBlock,
	shardID byte) ([][]string, map[string]string, error) {

	instructions := [][]string{}
	stakingTx := make(map[string]string)
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			// Get Staking Tx
			// assume that stake instruction already been validated by beacon committee

			if l[0] != instruction.STAKE_ACTION {
				instructions = append(instructions, l)
				continue
			}

			if l[0] == instruction.STAKE_ACTION && l[2] == "beacon" {
				beacon := strings.Split(l[1], ",")
				newBeaconCandidates := []string{}
				newBeaconCandidates = append(newBeaconCandidates, beacon...)
				if len(l) == 6 {
					for i, v := range strings.Split(l[3], ",") {
						txHash, err := common.Hash{}.NewHashFromStr(v)
						if err != nil {
							continue
						}
						_, _, _, err = blockchain.GetTransactionByHashWithShardID(*txHash, shardID)
						if err != nil {
							continue
						}
						// if transaction belong to this shard then add to shard beststate
						stakingTx[newBeaconCandidates[i]] = v
					}
					// instructions = append(instructions, l)
				}
			}

			if l[0] == instruction.STAKE_ACTION && l[2] == "shard" {
				shard := strings.Split(l[1], ",")
				newShardCandidates := []string{}
				newShardCandidates = append(newShardCandidates, shard...)
				if len(l) == 6 {
					for i, v := range strings.Split(l[3], ",") {
						txHash, err := common.Hash{}.NewHashFromStr(v)
						if err != nil {
							continue
						}
						_, _, _, err = blockchain.GetTransactionByHashWithShardID(*txHash, shardID)
						if err != nil {
							continue
						}
						// if transaction belong to this shard then add to shard beststate
						stakingTx[newShardCandidates[i]] = v
					}
					instructions = append(instructions, l)
				}
			}
		}
	}

	return instructions, stakingTx, nil
}

// committeeChanged checks if swap instructions really changed the committee list
// (not just empty swap instruction)
func committeeChanged(swap []string) bool {
	if len(swap) < 3 {
		return false
	}
	in := swap[1]
	out := swap[2]
	return len(in) > 0 || len(out) > 0
}
