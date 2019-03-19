package blockchain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/metadata/frombeaconins"
	"github.com/constant-money/constant-chain/metadata/fromshardins"
	"github.com/constant-money/constant-chain/privacy"
	"github.com/constant-money/constant-chain/transaction"
	"github.com/pkg/errors"
)

type accumulativeValues struct {
	bondsSold            uint64
	govTokensSold        uint64
	incomeFromBonds      uint64
	incomeFromGOVTokens  uint64
	dcbTokensSoldByUSD   uint64
	dcbTokensSoldByETH   uint64
	constantsBurnedByETH uint64
	buyBackCoins         uint64
	totalFee             uint64
	totalSalary          uint64
	totalRefundAmt       uint64
	totalOracleRewards   uint64
	saleDataMap          map[string]*component.SaleData
}

func isGOVFundEnough(
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
	expense uint64,
) bool {
	govFund := beaconBestState.StabilityInfo.SalaryFund
	income := accumulativeValues.incomeFromBonds + accumulativeValues.incomeFromGOVTokens + accumulativeValues.totalFee
	totalExpensed := accumulativeValues.buyBackCoins + accumulativeValues.totalSalary + accumulativeValues.totalRefundAmt + accumulativeValues.totalOracleRewards
	return (govFund + income - expense - totalExpensed) > 0
}

// build actions from txs and ins at shard
func buildStabilityActions(
	txs []metadata.Transaction,
	bc *BlockChain,
	shardID byte,
	producerAddress *privacy.PaymentAddress,
	shardBlockHeight uint64,
	beaconBlocks []*BeaconBlock,
) [][]string {
	actions := [][]string{}
	for _, tx := range txs {
		meta := tx.GetMetadata()
		if meta != nil {
			actionPairs, err := meta.BuildReqActions(tx, bc, shardID)
			if err != nil {
				continue
			}
			actions = append(actions, actionPairs...)
		}
	}

	// build salary update action
	totalFee := getShardBlockFee(txs)
	totalSalary := getShardBlockSalary(txs, bc.BestState.Beacon)
	if totalFee != 0 || totalSalary != 0 {
		salaryUpdateActions, _ := createShardBlockSalaryUpdateAction(totalSalary, totalFee, producerAddress, shardBlockHeight)
		actions = append(actions, salaryUpdateActions...)
	}

	//Add response instruction
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {

			shardToProcess, err := strconv.Atoi(l[1])
			if err != nil {
				panic("err")
				return nil
			}
			if shardToProcess == int(shardID) {
				metaType, err := strconv.Atoi(l[0])
				if err != nil {
					panic("err")
					return nil
				}
				var newIns []string
				switch metaType {
				case component.AcceptDCBProposalIns:
					acceptProposalIns := frombeaconins.AcceptProposalIns{}
					err := json.Unmarshal([]byte(l[2]), &acceptProposalIns)
					if err != nil {
						panic(err)
					}
					txID := acceptProposalIns.TxID
					_, _, _, txProposal, err := bc.GetTransactionByHash(&txID)
					metaProposal := txProposal.GetMetadata().(*metadata.SubmitDCBProposalMetadata)
					newIns, err = fromshardins.NewNewDCBConstitutionIns(
						metaProposal.SubmitProposalInfo,
						metaProposal.DCBParams,
						acceptProposalIns.Voter,
					).GetStringFormat()
					if err != nil {
						panic(err)
					}
				case component.AcceptGOVProposalIns:
					acceptProposalIns := frombeaconins.AcceptProposalIns{}
					err := json.Unmarshal([]byte(l[2]), &acceptProposalIns)
					if err != nil {
						panic(err)
					}
					txID := acceptProposalIns.TxID
					_, _, _, txProposal, err := bc.GetTransactionByHash(&txID)
					metaProposal := txProposal.GetMetadata().(*metadata.SubmitGOVProposalMetadata)
					if err != nil {
						panic(err)
					}
					newIns, err = fromshardins.NewNewGOVConstitutionIns(
						metaProposal.SubmitProposalInfo,
						metaProposal.GOVParams,
						acceptProposalIns.Voter,
					).GetStringFormat()
					if err != nil {
						panic(err)
					}
				}
				actions = append(actions, newIns)
			}
		}
	}

	return actions
}

// build instructions at beacon chain before syncing to shards
func (blkTmpGen *BlkTmplGenerator) buildStabilityInstructions(
	shardID byte,
	shardBlockInstructions [][]string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	instructions := [][]string{}
	//Add Voting instruction
	// step 3 hyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
	votingInstruction, err := blkTmpGen.chain.generateVotingInstructionWOIns(0)
	if err != nil {
		return nil, NewBlockChainError(BeaconError, err)
	}
	instructions = append(instructions, votingInstruction...)

	for _, inst := range shardBlockInstructions {
		fmt.Printf("[db] beaconProducer found inst: %s\n", inst[0])
		// TODO: will improve the condition later
		if inst[0] == StakeAction || inst[0] == SwapAction || inst[0] == RandomAction {
			continue
		}
		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			return [][]string{}, err
		}
		contentStr := inst[1]
		newInst := [][]string{}
		switch metaType {
		case metadata.LoanRequestMeta, metadata.LoanResponseMeta, metadata.LoanWithdrawMeta, metadata.LoanPaymentMeta, metadata.DividendSubmitMeta:
			newInst, err = buildPassThroughInstruction(metaType, contentStr)

		case metadata.BuyFromGOVRequestMeta:
			newInst, err = buildInstructionsForBuyBondsFromGOVReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.BuyGOVTokenRequestMeta:
			newInst, err = buildInstructionsForBuyGOVTokensReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.CrowdsaleRequestMeta:
			newInst, err = buildInstructionsForCrowdsaleRequest(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.BuyBackRequestMeta:
			newInst, err = buildInstructionsForBuyBackBondsReq(shardID, contentStr, beaconBestState, accumulativeValues, blkTmpGen.chain)

		case metadata.IssuingRequestMeta:
			newInst, err = buildInstructionsForIssuingReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.ContractingRequestMeta:
			newInst, err = buildInstructionsForContractingReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.ShardBlockSalaryRequestMeta:
			newInst, err = buildInstForShardBlockSalaryReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case component.NewDCBConstitutionIns:
			newInst, err = buildUpdateConstitutionIns(inst[2], common.DCBBoard)

		case component.NewGOVConstitutionIns:
			newInst, err = buildUpdateConstitutionIns(inst[2], common.GOVBoard)

		case component.VoteBoardIns:
			err = blkTmpGen.chain.AddVoteBoard(inst[2])

		case component.VoteProposalIns:
			err = blkTmpGen.chain.AddVoteProposal(inst[2])
		default:
			continue
		}
		if err != nil {
			return [][]string{}, err
		}
		instructions = append(instructions, newInst...)
	}
	// update component in beststate
	return instructions, nil
}

func buildUpdateConstitutionIns(inst string, boardType common.BoardType) ([][]string, error) {
	var newInst []string
	if boardType == common.DCBBoard {
		newConstitutionIns, err := fromshardins.NewNewDCBConstitutionInsFromStr(inst)
		if err != nil {
			return nil, err
		}
		newInst, err = frombeaconins.NewUpdateDCBConstitutionIns(
			newConstitutionIns.SubmitProposalInfo,
			newConstitutionIns.DCBParams,
			newConstitutionIns.Voter,
		).GetStringFormat()
		if err != nil {
			return nil, err
		}
	} else {
		newConstitutionIns, err := fromshardins.NewNewGOVConstitutionInsFromStr(inst)
		if err != nil {
			return nil, err
		}
		newInst, err = frombeaconins.NewUpdateGOVConstitutionIns(
			newConstitutionIns.SubmitProposalInfo,
			newConstitutionIns.GOVParams,
			newConstitutionIns.Voter,
		).GetStringFormat()
		if err != nil {
			return nil, err
		}
	}
	return [][]string{newInst}, nil
}

func (blockgen *BlkTmplGenerator) buildLoanResponseTx(tx metadata.Transaction, producerPrivateKey *privacy.SpendingKey) (metadata.Transaction, error) {
	// Get loan request
	withdrawMeta := tx.GetMetadata().(*metadata.LoanWithdraw)
	meta, err := blockgen.chain.GetLoanRequestMeta(withdrawMeta.LoanID)
	if err != nil {
		return nil, err
	}

	// Build loan unlock tx
	unlockMeta := &metadata.LoanUnlock{
		LoanID:       make([]byte, len(withdrawMeta.LoanID)),
		MetadataBase: metadata.MetadataBase{Type: metadata.LoanUnlockMeta},
	}
	copy(unlockMeta.LoanID, withdrawMeta.LoanID)
	unlockMetaList := []metadata.Metadata{unlockMeta}
	amounts := []uint64{meta.LoanAmount}
	txNormals, err := transaction.BuildCoinbaseTxs([]*privacy.PaymentAddress{meta.ReceiveAddress}, amounts, producerPrivateKey, blockgen.chain.GetDatabase(), unlockMetaList)
	if err != nil {
		return nil, errors.Errorf("Error building unlock tx for loan id %x", withdrawMeta.LoanID)
	}
	fmt.Printf("[db] success build LoanUnlock with amount %d\n", meta.LoanAmount)
	return txNormals[0], nil
}

func (blockgen *BlkTmplGenerator) buildStabilityResponseTxsFromInstructions(
	beaconBlocks []*BeaconBlock,
	producerPrivateKey *privacy.SpendingKey,
	shardID byte,
) ([]metadata.Transaction, error) {
	// TODO(@0xbunyip): refund bonds in multiple blocks since many refund instructions might come at once and UTXO picking order is not perfect
	unspentTokens := map[string]([]transaction.TxTokenVout){}
	resTxs := []metadata.Transaction{}
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			// TODO: will improve the condition later
			if l[0] == StakeAction || l[0] == "swap" || l[0] == RandomAction {
				continue
			}
			if len(l) <= 2 {
				continue
			}
			shardToProcess, err := strconv.Atoi(l[1])
			if err == nil && shardToProcess == int(shardID) {
				metaType, err := strconv.Atoi(l[0])
				if err != nil {
					return nil, err
				}
				fmt.Printf("[db] shard build Resp from inst: %+v\n", l)
				Logger.log.Warn("Metadata type:", metaType, "\n")
				switch metaType {
				case metadata.CrowdsalePaymentMeta:
					paymentInst, err := ParseCrowdsalePaymentInstruction(l[2])
					if err != nil {
						return nil, err
					}
					tx, err := blockgen.buildPaymentForCrowdsale(paymentInst, unspentTokens, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, tx)

				case metadata.BuyFromGOVRequestMeta:
					contentStr := l[3]
					sellingBondsParamsStr := l[4]
					txs, err := blockgen.buildBuyBondsFromGOVRes(l[2], contentStr, sellingBondsParamsStr, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)

				case metadata.BuyGOVTokenRequestMeta:
					contentStr := l[3]
					txs, err := blockgen.buildBuyGOVTokensRes(l[2], contentStr, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)

				case metadata.BuyBackRequestMeta:
					buyBackInfoStr := l[3]
					txs, err := blockgen.buildBuyBackRes(l[2], buyBackInfoStr, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)

					// todo @0xjackalope move meta to ins?
				case metadata.SendBackTokenVoteBoardFailMeta:
					Logger.log.Info(metadata.SendBackTokenVoteBoardFailMeta, "\n")
					sendBackTokenVoteFail := frombeaconins.TxSendBackTokenVoteFailIns{}
					err := json.Unmarshal([]byte(l[2]), &sendBackTokenVoteFail)
					if err != nil {
						Logger.log.Error("Here, why?", err.Error(), "\n")
						return nil, err
					}

					txs, err1 := sendBackTokenVoteFail.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase, blockgen.chain, shardID)

					if err1 != nil {
						Logger.log.Error("Here, why?????????", err1, reflect.TypeOf(err1), reflect.ValueOf(err1), "\n")
						return nil, err1
					}
					resTxs = append(resTxs, txs)
				case metadata.ShareRewardOldDCBBoardMeta, metadata.ShareRewardOldGOVBoardMeta:
					Logger.log.Info(metadata.ShareRewardOldDCBBoardMeta, metadata.ShareRewardOldGOVBoardMeta, "\n")
					shareRewardOldBoard := frombeaconins.TxShareRewardOldBoardMetadataIns{}
					err := json.Unmarshal([]byte(l[2]), &shareRewardOldBoard)
					if err != nil {
						return nil, err
					}
					txs, err := shareRewardOldBoard.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs)

				case metadata.IssuingRequestMeta:
					issuingInfoStr := l[3]
					txs, err := blockgen.buildIssuingRes(l[2], issuingInfoStr, producerPrivateKey, shardID)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)

				case metadata.ContractingRequestMeta:
					contractingInfoStr := l[3]
					txs, err := blockgen.buildContractingRes(l[2], contractingInfoStr, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)

				case metadata.ShardBlockSalaryRequestMeta:
					salaryReqInfoStr := l[3]
					txs, err := blockgen.buildSalaryRes(l[2], salaryReqInfoStr, producerPrivateKey)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, txs...)
				}
			}
		}
	}
	return resTxs, nil
}

func (blockgen *BlkTmplGenerator) buildStabilityResponseTxsAtShardOnly(txs []metadata.Transaction, producerPrivateKey *privacy.SpendingKey) ([]metadata.Transaction, error) {
	respTxs := []metadata.Transaction{}
	removeIds := []int{}
	multisigsRegTxs := []metadata.Transaction{}
	for i, tx := range txs {
		var respTx metadata.Transaction
		var err error

		switch tx.GetMetadataType() {
		case metadata.LoanWithdrawMeta:
			respTx, err = blockgen.buildLoanResponseTx(tx, producerPrivateKey)
		case metadata.MultiSigsRegistrationMeta:
			multisigsRegTxs = append(multisigsRegTxs, tx)
		}

		if err != nil {
			// Remove this tx if cannot create corresponding response
			removeIds = append(removeIds, i)
		} else if respTx != nil {
			respTxs = append(respTxs, respTx)
		}
	}

	err := blockgen.registerMultiSigsAddresses(multisigsRegTxs)
	if err != nil {
		return nil, err
	}

	// TODO(@0xbunyip): remove tx from txsToAdd?
	return respTxs, nil
}

func (chain *BlockChain) AddVoteBoard(inst string) error {
	newInst, err := fromshardins.NewVoteBoardInsFromStr(inst)
	if err != nil {
		return err
	}
	boardType := newInst.BoardType
	voteAmount := newInst.AmountOfVote
	voterPayment := newInst.VoterPaymentAddress
	governor := chain.GetGovernor(boardType)
	boardIndex := governor.GetBoardIndex() + 1
	err1 := chain.GetDatabase().AddVoteBoard(
		boardType,
		boardIndex,
		voterPayment,
		newInst.CandidatePaymentAddress,
		voteAmount,
	)
	if err1 != nil {
		return err1
	}
	return nil
}

func (chain *BlockChain) AddVoteProposal(inst string) error {
	newInst, err := fromshardins.NewNormalVoteProposalInsFromStr(inst)
	if err != nil {
		return err
	}
	boardType := newInst.BoardType
	// step 4 hyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
	nextConstitutionIndex := chain.GetConstitution(boardType).GetConstitutionIndex() + 1
	err = chain.GetDatabase().AddVoteProposalDB(
		boardType,
		nextConstitutionIndex,
		newInst.VoteProposal.VoterPayment.Bytes(),
		newInst.VoteProposal.ProposalTxID.GetBytes(),
	)
	if err != nil {
		return err
	}
	return nil
}

// func (blockgen *BlkTmplGenerator) buildIssuingResTxs(
// 	chainID byte,
// 	issuingReqTxs []metadata.Transaction,
// 	privatekey *privacy.SpendingKey,
// ) ([]metadata.Transaction, error) {
// 	prevBlock := blockgen.chain.BestState[chainID].BestBlock
// 	oracleParams := prevBlock.Header.Oracle

// 	issuingResTxs := []metadata.Transaction{}
// 	for _, issuingReqTx := range issuingReqTxs {
// 		meta := issuingReqTx.GetMetadata()
// 		issuingReq, ok := meta.(*metadata.IssuingRequest)
// 		if !ok {
// 			return []metadata.Transaction{}, errors.New("Could not parse IssuingRequest metadata.")
// 		}
// 		if issuingReq.AssetType == common.DCBTokenID {
// 			issuingRes := metadata.NewIssuingResponse(*issuingReqTx.Hash(), metadata.IssuingResponseMeta)
// 			dcbTokenPrice := uint64(1)
// 			if oracleParams.DCBToken != 0 {
// 				dcbTokenPrice = oracleParams.DCBToken
// 			}
// 			issuingAmt := issuingReq.DepositedAmount / dcbTokenPrice
// 			txTokenVout := transaction.TxTokenVout{
// 				Value:          issuingAmt,
// 				PaymentAddress: issuingReq.ReceiverAddress,
// 			}
// 			txTokenData := transaction.TxTokenData{
// 				Type:       transaction.CustomTokenInit,
// 				Amount:     issuingAmt,
// 				PropertyID: common.Hash(common.DCBTokenID),
// 				Vins:       []transaction.TxTokenVin{},
// 				Vouts:      []transaction.TxTokenVout{txTokenVout},
// 				// PropertyName:   "",
// 				// PropertySymbol: coinbaseTxType,
// 			}
// 			resTx := &transaction.TxCustomToken{
// 				TxTokenData: txTokenData,
// 			}
// 			resTx.Type = common.TxCustomTokenType
// 			resTx.SetMetadata(issuingRes)
// 			issuingResTxs = append(issuingResTxs, resTx)
// 			continue
// 		}
// 		if issuingReq.AssetType == common.ConstantID {
// 			constantPrice := uint64(1)
// 			if oracleParams.Constant != 0 {
// 				constantPrice = oracleParams.Constant
// 			}
// 			issuingAmt := issuingReq.DepositedAmount / constantPrice
// 			issuingRes := metadata.NewIssuingResponse(*issuingReqTx.Hash(), metadata.IssuingResponseMeta)
// 			resTx := new(transaction.Tx)
// 			err := resTx.InitTxSalary(issuingAmt, &issuingReq.ReceiverAddress, privatekey, blockgen.chain.GetDatabase(), issuingRes)
// 			if err != nil {
// 				return []metadata.Transaction{}, err
// 			}
// 			issuingResTxs = append(issuingResTxs, resTx)
// 		}
// 	}
// 	return issuingResTxs, nil
// }

// func calculateAmountOfRefundTxs(
// 	smallTxHashes []*common.Hash,
// 	addresses []*privacy.PaymentAddress,
// 	estimatedRefundAmt uint64,
// 	remainingFund uint64,
// 	db database.DatabaseInterface,
// 	privatekey *privacy.SpendingKey,
// ) ([]metadata.Transaction, uint64) {
// 	amt := uint64(0)
// 	if estimatedRefundAmt <= remainingFund {
// 		amt = estimatedRefundAmt
// 	} else {
// 		amt = remainingFund
// 	}
// 	actualRefundAmt := amt / uint64(len(addresses))
// 	var refundTxs []metadata.Transaction
// 	for i := 0; i < len(addresses); i++ {
// 		addr := addresses[i]
// 		refundMeta := metadata.NewRefund(*smallTxHashes[i], metadata.RefundMeta)
// 		refundTx := new(transaction.Tx)
// 		err := refundTx.InitTxSalary(actualRefundAmt, addr, privatekey, db, refundMeta)
// 		if err != nil {
// 			Logger.log.Error(err)
// 			continue
// 		}
// 		refundTxs = append(refundTxs, refundTx)
// 	}
// 	return refundTxs, amt
// }

// func (blockgen *BlkTmplGenerator) buildRefundTxs(
// 	chainID byte,
// 	remainingFund uint64,
// 	privatekey *privacy.SpendingKey,
// ) ([]metadata.Transaction, uint64) {
// 	if remainingFund <= 0 {
// 		Logger.log.Info("GOV fund is not enough for refund.")
// 		return []metadata.Transaction{}, 0
// 	}
// 	prevBlock := blockgen.chain.BestState[chainID].BestBlock
// 	header := prevBlock.Header
// 	govParams := header.GOVConstitution.GOVParams
// 	refundInfo := govParams.RefundInfo
// 	if refundInfo == nil {
// 		Logger.log.Info("Refund info is not existed.")
// 		return []metadata.Transaction{}, 0
// 	}
// 	lookbackBlockHeight := header.Height - common.RefundPeriod
// 	if lookbackBlockHeight < 0 {
// 		return []metadata.Transaction{}, 0
// 	}
// 	lookbackBlock, err := blockgen.chain.GetBlockByBlockHeight(lookbackBlockHeight, chainID)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return []metadata.Transaction{}, 0
// 	}
// 	addresses := []*privacy.PaymentAddress{}
// 	smallTxHashes := []*common.Hash{}
// 	estimatedRefundAmt := uint64(0)
// 	for _, tx := range lookbackBlock.Transactions {
// 		if tx.GetType() != common.TxNormalType {
// 			continue
// 		}
// 		lookbackTx, ok := tx.(*transaction.Tx)
// 		if !ok {
// 			continue
// 		}
// 		addr := lookbackTx.GetSenderAddress()
// 		if addr == nil {
// 			continue
// 		}
// 		txValue := lookbackTx.CalculateTxValue()
// 		if txValue > refundInfo.ThresholdToLargeTx {
// 			continue
// 		}
// 		addresses = append(addresses, addr)
// 		smallTxHashes = append(smallTxHashes, tx.Hash())
// 		estimatedRefundAmt += refundInfo.RefundAmount
// 	}
// 	if len(addresses) == 0 {
// 		return []metadata.Transaction{}, 0
// 	}
// 	refundTxs, totalRefundAmt := calculateAmountOfRefundTxs(
// 		smallTxHashes,
// 		addresses,
// 		estimatedRefundAmt,
// 		remainingFund,
// 		blockgen.chain.GetDatabase(),
// 		privatekey,
// 	)
// 	return refundTxs, totalRefundAmt
// }

// func (blockgen *BlkTmplGenerator) processLoan(sourceTxns []*metadata.TxDesc, producerPrivateKey *privacy.SpendingKey) (uint64, []metadata.Transaction, []metadata.Transaction) {
// 	amount := uint64(0)
// 	loanUnlockTxs := []metadata.Transaction{}
// 	removableTxs := []metadata.Transaction{}
// 	for _, txDesc := range sourceTxns {
// 		if txDesc.Tx.GetMetadataType() == metadata.LoanPaymentMeta {
// 			paymentAmount, err := blockgen.calculateInterestPaid(txDesc.Tx)
// 			if err != nil {
// 				removableTxs = append(removableTxs, txDesc.Tx)
// 				continue
// 			}
// 			amount += paymentAmount
// 		} else if txDesc.Tx.GetMetadataType() == metadata.LoanWithdrawMeta {
//          DONE
// 			withdrawMeta := txDesc.Tx.GetMetadata().(*metadata.LoanWithdraw)
// 			meta, err := blockgen.chain.GetLoanRequestMeta(withdrawMeta.LoanID)
// 			if err != nil {
// 				removableTxs = append(removableTxs, txDesc.Tx)
// 				continue
// 			}

// 			unlockMeta := &metadata.LoanUnlock{
// 				LoanID:       make([]byte, len(withdrawMeta.LoanID)),
// 				MetadataBase: metadata.MetadataBase{Type: metadata.LoanUnlockMeta},
// 			}
// 			copy(unlockMeta.LoanID, withdrawMeta.LoanID)
// 			uplockMetaList := []metadata.Metadata{unlockMeta}
// 			amounts := []uint64{meta.LoanAmount}
// 			txNormals, err := transaction.BuildCoinbaseTxs([]*privacy.PaymentAddress{meta.ReceiveAddress}, amounts, producerPrivateKey, blockgen.chain.GetDatabase(), uplockMetaList)
// 			if err != nil {
// 				removableTxs = append(removableTxs, txDesc.Tx)
// 				continue
// 			}
// 			loanUnlockTxs = append(loanUnlockTxs, txNormals[0]) // There's only one tx
// 		}
// 	}
// 	return amount, loanUnlockTxs, removableTxs
// }

// func (blockgen *BlkTmplGenerator) buildBuyBackResponseTxs(
// 	buyBackFromInfos []*buyBackFromInfo,
// 	chainID byte,
// 	privatekey *privacy.SpendingKey,
// ) ([]metadata.Transaction, error) {
// 	if len(buyBackFromInfos) == 0 {
// 		return []metadata.Transaction{}, nil
// 	}

// 	// prevBlock := blockgen.chain.BestState[chainID].BestBlock
// 	var buyBackResTxs []metadata.Transaction
// 	for _, buyBackFromInfo := range buyBackFromInfos {
// 		buyBackAmount := buyBackFromInfo.value * buyBackFromInfo.buyBackPrice
// 		buyBackRes := metadata.NewBuyBackResponse(*buyBackFromInfo.requestedTxID, metadata.BuyBackResponseMeta)
// 		buyBackResTx := new(transaction.Tx)
// 		err := buyBackResTx.InitTxSalary(buyBackAmount, &buyBackFromInfo.paymentAddress, privatekey, blockgen.chain.GetDatabase(), buyBackRes)
// 		if err != nil {
// 			return []metadata.Transaction{}, err
// 		}
// 		buyBackResTxs = append(buyBackResTxs, buyBackResTx)
// 	}
// 	return buyBackResTxs, nil
// }

// // buildBuySellResponsesTx
// // the tx is to distribute tokens (bond, gov, ...) to token requesters
// func (blockgen *BlkTmplGenerator) buildBuySellResponsesTx(
// 	buySellReqTxs []metadata.Transaction,
// 	sellingBondsParam *component.SellingBonds,
// ) ([]metadata.Transaction, error) {
// 	if len(buySellReqTxs) == 0 {
// 		return []metadata.Transaction{}, nil
// 	}
// 	var resTxs []metadata.Transaction
// 	for _, reqTx := range buySellReqTxs {
// 		resTx, err := buildSingleBuySellResponseTx(reqTx, sellingBondsParam)
// 		if err != nil {
// 			return []metadata.Transaction{}, err
// 		}
// 		resTxs = append(resTxs, resTx)
// 	}
// 	return resTxs, nil
// }

// func buildSingleBuySellResponseTx(
// 	buySellReqTx metadata.Transaction,
// 	sellingBondsParam *component.SellingBonds,
// ) (*transaction.TxCustomToken, error) {
// 	bondID := sellingBondsParam.GetID()
// 	buySellRes := metadata.NewBuySellResponse(
// 		*buySellReqTx.Hash(),
// 		sellingBondsParam.StartSellingAt,
// 		sellingBondsParam.Maturity,
// 		sellingBondsParam.BuyBackPrice,
// 		bondID[:],
// 		metadata.BuyFromGOVResponseMeta,
// 	)

// 	buySellReqMeta := buySellReqTx.GetMetadata()
// 	buySellReq, ok := buySellReqMeta.(*metadata.BuySellRequest)
// 	if !ok {
// 		return nil, errors.New("Could not assert BuySellRequest metadata.")
// 	}
// 	txTokenVout := transaction.TxTokenVout{
// 		Value:          buySellReq.Amount,
// 		PaymentAddress: buySellReq.PaymentAddress,
// 	}

// 	var propertyID [common.HashSize]byte
// 	copy(propertyID[:], bondID[:])
// 	txTokenData := transaction.TxTokenData{
// 		Type:       transaction.CustomTokenInit,
// 		Mintable:   true,
// 		Amount:     buySellReq.Amount,
// 		PropertyID: common.Hash(propertyID),
// 		Vins:       []transaction.TxTokenVin{},
// 		Vouts:      []transaction.TxTokenVout{txTokenVout},
// 	}
// 	txTokenData.PropertyName = txTokenData.PropertyID.String()
// 	txTokenData.PropertySymbol = txTokenData.PropertyID.String()

// 	resTx := &transaction.TxCustomToken{
// 		TxTokenData: txTokenData,
// 	}
// 	resTx.Type = common.TxCustomTokenType
// 	resTx.SetMetadata(buySellRes)
// 	return resTx, nil
// }

// func buildSingleBuyGOVTokensResTx(
// 	buyGOVTokensReqTx metadata.Transaction,
// 	sellingGOVTokensParams *component.SellingGOVTokens,
// ) (*transaction.TxCustomToken, error) {
// 	buyGOVTokensRes := metadata.NewResponseBase(
// 		*buyGOVTokensReqTx.Hash(),
// 		metadata.ResponseBaseMeta,
// 	)

// 	buyGOVTokensReqMeta := buyGOVTokensReqTx.GetMetadata()
// 	buyGOVTokensReq, ok := buyGOVTokensReqMeta.(*metadata.BuyGOVTokenRequest)
// 	if !ok {
// 		return nil, errors.New("Could not assert BuyGOVTokenRequest metadata.")
// 	}
// 	txTokenVout := transaction.TxTokenVout{
// 		Value:          buyGOVTokensReq.Amount,
// 		PaymentAddress: buyGOVTokensReq.BuyerAddress,
// 	}

// 	var propertyID [common.HashSize]byte
// 	copy(propertyID[:], buyGOVTokensReq.TokenID[:])
// 	txTokenData := transaction.TxTokenData{
// 		Type:       transaction.CustomTokenInit,
// 		Mintable:   true,
// 		Amount:     buyGOVTokensReq.Amount,
// 		PropertyID: common.Hash(propertyID),
// 		Vins:       []transaction.TxTokenVin{},
// 		Vouts:      []transaction.TxTokenVout{txTokenVout},
// 	}
// 	txTokenData.PropertyName = txTokenData.PropertyID.String()
// 	txTokenData.PropertySymbol = txTokenData.PropertyID.String()

// 	resTx := &transaction.TxCustomToken{
// 		TxTokenData: txTokenData,
// 	}
// 	resTx.Type = common.TxCustomTokenType
// 	resTx.SetMetadata(buyGOVTokensRes)
// 	return resTx, nil
// }

// func (blockgen *BlkTmplGenerator) buildBuyGOVTokensResTxs(
// 	buyGOVTokensReqTxs []metadata.Transaction,
// 	sellingGOVTokensParams *component.SellingGOVTokens,
// ) ([]metadata.Transaction, error) {
// 	if len(buyGOVTokensReqTxs) == 0 {
// 		return []metadata.Transaction{}, nil
// 	}
// 	var resTxs []metadata.Transaction
// 	for _, reqTx := range buyGOVTokensReqTxs {
// 		resTx, err := buildSingleBuyGOVTokensResTx(reqTx, sellingGOVTokensParams)
// 		if err != nil {
// 			return []metadata.Transaction{}, err
// 		}
// 		resTxs = append(resTxs, resTx)
// 	}
// 	return resTxs, nil
// }

// func (blockgen *BlkTmplGenerator) buildResponseTxs(
// 	chainID byte,
// 	sourceTxns []*metadata.TxDesc,
// 	privatekey *privacy.SpendingKey,
// 	txGroups *txGroups,
// 	accumulativeValues *accumulativeValues,
// 	buyBackFromInfos []*buyBackFromInfo,
// ) (*txGroups, *accumulativeValues, map[string]uint64, error) {
// 	prevBlock := blockgen.chain.BestState[chainID].BestBlock
// 	// create buy/sell response txs to distribute bonds/govs to requesters
// 	buySellResTxs, err := blockgen.buildBuySellResponsesTx(
// 		txGroups.buySellReqTxs,
// 		blockgen.chain.BestState[14].BestBlock.Header.GOVConstitution.GOVParams.SellingBonds,
// 	)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, nil, nil, err
// 	}
// 	buyGOVTokensResTxs, err := blockgen.buildBuyGOVTokensResTxs(
// 		txGroups.buyGOVTokensReqTxs,
// 		blockgen.chain.BestState[14].BestBlock.Header.GOVConstitution.GOVParams.SellingGOVTokens,
// 	)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, nil, nil, err
// 	}

// 	// create buy-back response txs to distribute constants to buy-back requesters
// 	buyBackResTxs, err := blockgen.buildBuyBackResponseTxs(buyBackFromInfos, chainID, privatekey)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, nil, nil, err
// 	}

// 	oracleRewardTxs, totalOracleRewards, updatedOracleValues, err := blockgen.buildOracleRewardTxs(chainID, privatekey)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, nil, nil, err
// 	}

// 	// create refund txs
// 	currentSalaryFund := prevBlock.Header.SalaryFund
// 	remainingFund := currentSalaryFund + accumulativeValues.totalFee + accumulativeValues.incomeFromBonds + accumulativeValues.incomeFromGOVTokens - (accumulativeValues.totalSalary + accumulativeValues.buyBackCoins + totalOracleRewards)
// 	refundTxs, totalRefundAmt := blockgen.buildRefundTxs(chainID, remainingFund, privatekey)

// 	issuingResTxs, err := blockgen.buildIssuingResTxs(chainID, txGroups.issuingReqTxs, privatekey)
// 	if err != nil {
// 		Logger.log.Error(err)
// 		return nil, nil, nil, err
// 	}

// 	// Get loan payment amount to add to DCB fund
// 	loanPaymentAmount, unlockTxs, removableTxs := blockgen.processLoan(sourceTxns, privatekey)
// 	for _, tx := range removableTxs {
// 		txGroups.txToRemove = append(txGroups.txToRemove, tx)
// 	}
// 	txGroups.buySellResTxs = buySellResTxs
// 	txGroups.buyGOVTokensResTxs = buyGOVTokensResTxs
// 	txGroups.buyBackResTxs = buyBackResTxs
// 	txGroups.oracleRewardTxs = oracleRewardTxs
// 	txGroups.refundTxs = refundTxs
// 	txGroups.issuingResTxs = issuingResTxs
// 	txGroups.unlockTxs = unlockTxs

// 	accumulativeValues.totalOracleRewards = totalOracleRewards
// 	accumulativeValues.totalRefundAmt = totalRefundAmt
// 	accumulativeValues.loanPaymentAmount = loanPaymentAmount
// 	accumulativeValues.currentSalaryFund = currentSalaryFund
// 	return txGroups, accumulativeValues, updatedOracleValues, nil
// }
