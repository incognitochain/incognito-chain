package blockchain

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database/lvdb"
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

func getStabilityInfoByHeight(blockchain *BlockChain, beaconHeight uint64) (*StabilityInfo, error) {
	stabilityInfoBytes, dbErr := blockchain.config.DataBase.FetchStabilityInfoByHeight(beaconHeight)
	if dbErr != nil {
		return nil, dbErr
	}
	if len(stabilityInfoBytes) == 0 { // not found
		return nil, nil
	}
	var stabilityInfo StabilityInfo
	unmarshalErr := json.Unmarshal(stabilityInfoBytes, &stabilityInfo)
	if unmarshalErr != nil {
		return nil, unmarshalErr
	}
	return &stabilityInfo, nil
}

func isGOVFundEnough(
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
	expense uint64,
) bool {
	govFund := beaconBestState.StabilityInfo.SalaryFund
	income := accumulativeValues.incomeFromBonds + accumulativeValues.incomeFromGOVTokens + accumulativeValues.totalFee
	totalExpensed := accumulativeValues.buyBackCoins + accumulativeValues.totalSalary + accumulativeValues.totalRefundAmt + accumulativeValues.totalOracleRewards
	return govFund+income > expense+totalExpensed
}

// build actions from txs and ins at shard
func buildStabilityActions(
	txs []metadata.Transaction,
	bc *BlockChain,
	shardID byte,
	producerAddress *privacy.PaymentAddress,
	shardBlockHeight uint64,
	beaconBlocks []*BeaconBlock,
	beaconHeight uint64,
) ([][]string, error) {
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
	totalSalary, err := getShardBlockSalary(txs, bc, beaconHeight)
	if err != nil {
		return nil, err
	}

	if totalFee != 0 || totalSalary != 0 {
		salaryUpdateActions, _ := createShardBlockSalaryUpdateAction(totalSalary, totalFee, producerAddress, shardBlockHeight)
		actions = append(actions, salaryUpdateActions...)
	}

	//Add response instruction
	for _, beaconBlock := range beaconBlocks {
		for _, l := range beaconBlock.Body.Instructions {
			shardToProcess, err := strconv.Atoi(l[1])
			if err != nil {
				continue
			}
			if shardToProcess == int(shardID) {
				metaType, err := strconv.Atoi(l[0])
				if err != nil {
					return nil, err
				}
				var newIns []string
				if metaType != 37 {
					fmt.Println("[voting] - instructions metaType: ", metaType, component.AcceptDCBProposalIns)
				}
				switch metaType {
				case component.AcceptDCBProposalIns:
					acceptProposalIns := frombeaconins.AcceptProposalIns{}
					err := json.Unmarshal([]byte(l[2]), &acceptProposalIns)
					if err != nil {
						fmt.Println("[voting] - error 1 ", err.Error())
						return nil, err
					}
					txID := acceptProposalIns.TxID
					_, _, _, txProposal, err := bc.GetTransactionByHash(&txID)
					metaProposal := txProposal.GetMetadata().(*metadata.SubmitDCBProposalMetadata)
					newIns, err = fromshardins.NewNewDCBConstitutionIns(
						metaProposal.SubmitProposalInfo,
						metaProposal.DCBParams,
						acceptProposalIns.Voters,
					).GetStringFormat()
					if err != nil {
						fmt.Println("[voting] - error 2 ", err.Error())
						return nil, err
					}
					fmt.Println("[voting] - new instructions AcceptProposalIns: ", newIns)
				case component.AcceptGOVProposalIns:
					acceptProposalIns := frombeaconins.AcceptProposalIns{}
					err := json.Unmarshal([]byte(l[2]), &acceptProposalIns)
					if err != nil {
						return nil, err
					}
					txID := acceptProposalIns.TxID
					_, _, _, txProposal, err := bc.GetTransactionByHash(&txID)
					metaProposal := txProposal.GetMetadata().(*metadata.SubmitGOVProposalMetadata)
					if err != nil {
						fmt.Println("[voting] - error 1 ", err.Error())
						return nil, err
					}
					newIns, err = fromshardins.NewNewGOVConstitutionIns(
						metaProposal.SubmitProposalInfo,
						metaProposal.GOVParams,
						acceptProposalIns.Voters,
					).GetStringFormat()
					if err != nil {
						fmt.Println("[voting] - error 2 ", err.Error())
						return nil, err
					}
					fmt.Println("[voting] - new instructions AcceptProposalIns: ", newIns)
				}
				actions = append(actions, newIns)
			}
		}
	}

	return actions, nil
}

// build instructions at beacon chain before syncing to shards
func (blockChain *BlockChain) buildStabilityInstructions(
	shardID byte,
	shardBlockInstructions [][]string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	instructions := [][]string{}

	fmt.Printf("[db] building stability instructions\n")
	for _, inst := range shardBlockInstructions {
		if len(inst) == 0 {
			continue
		}
		if inst[0] != "37" {
			fmt.Println("[voting] -----------------------> Instrucstion from shard to beacon ", inst)
			fmt.Printf("[db] beaconProducer found inst: %s\n", inst[0])
		}
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
		case metadata.LoanRequestMeta, metadata.LoanResponseMeta, metadata.LoanWithdrawMeta, metadata.LoanPaymentMeta:
			newInst, err = buildPassThroughInstruction(metaType, contentStr)

		case metadata.BuyFromGOVRequestMeta:
			newInst, err = buildInstructionsForBuyBondsFromGOVReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.BuyGOVTokenRequestMeta:
			newInst, err = buildInstructionsForBuyGOVTokensReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.CrowdsaleRequestMeta:
			newInst, err = buildInstructionsForCrowdsaleRequest(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.TradeActivationMeta:
			newInst, err = buildInstructionsForTradeActivation(shardID, contentStr)

		case metadata.BuyBackRequestMeta:
			newInst, err = buildInstructionsForBuyBackBondsReq(shardID, contentStr, beaconBestState, accumulativeValues, blockChain)

		case metadata.IssuingRequestMeta:
			newInst, err = buildInstructionsForIssuingReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.ContractingRequestMeta:
			newInst, err = buildInstructionsForContractingReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.ShardBlockSalaryRequestMeta:
			newInst, err = buildInstForShardBlockSalaryReq(shardID, contentStr, beaconBestState, accumulativeValues)

		case metadata.OracleFeedMeta:
			newInst, err = buildInstForOracleFeedReq(shardID, contentStr, beaconBestState)

		case metadata.UpdatingOracleBoardMeta:
			newInst, err = buildInstForUpdatingOracleBoardReq(shardID, contentStr, beaconBestState)

		case component.NewDCBConstitutionIns:
			fmt.Println("[voting]-[NewDCBConstitutionIns] " + inst[2])
			newInst, err = buildUpdateConstitutionIns(inst[2], common.DCBBoard)

		case component.NewGOVConstitutionIns:
			fmt.Println("[voting]-[NewGOVConstitutionIns] " + inst[2])
			newInst, err = buildUpdateConstitutionIns(inst[2], common.GOVBoard)

		case component.VoteBoardIns:
			fmt.Println("[voting]-[AddVoteBoard] " + inst[2])
			err = blockChain.AddVoteBoard(inst[2])

		case component.SubmitProposalIns:
			fmt.Println("[voting]-[AddSubmitProposal] " + inst[2])
			err = blockChain.AddSubmitProposal(inst[2])
		case component.VoteProposalIns:
			fmt.Println("[voting]-[VoteProposalIns] " + inst[2])
			err = blockChain.AddVoteProposal(inst[2])
		default:
			continue
		}
		if err != nil {
			Logger.log.Error(err)
			continue
		}

		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
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
			newConstitutionIns.Voters,
		).GetStringFormat()
		if err != nil {
			return nil, err
		}
		fmt.Println("[voting]-[buildUpdateConstitutionIns]- completed", newInst)
	} else {
		newConstitutionIns, err := fromshardins.NewNewGOVConstitutionInsFromStr(inst)
		if err != nil {
			return nil, err
		}
		newInst, err = frombeaconins.NewUpdateGOVConstitutionIns(
			newConstitutionIns.SubmitProposalInfo,
			newConstitutionIns.GOVParams,
			newConstitutionIns.Voters,
		).GetStringFormat()
		if err != nil {
			return nil, err
		}
	}
	return [][]string{newInst}, nil
}

func (blockgen *BlkTmplGenerator) buildLoanResponseTx(tx metadata.Transaction, producerPrivateKey *privacy.PrivateKey) (metadata.Transaction, error) {
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
	producerPrivateKey *privacy.PrivateKey,
	shardID byte,
) ([]metadata.Transaction, error) {
	// TODO(@0xbunyip): refund bonds in multiple blocks since many refund instructions might come at once and UTXO picking order is not perfect
	unspentTokens := map[string]([]transaction.TxTokenVout){}
	tradeActivated := map[string]bool{}
	resTxs := []metadata.Transaction{}
	for _, beaconBlock := range beaconBlocks {
		fmt.Println("[db] - beaconBlock[", beaconBlock.Header.Height, "]")
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
				if metaType != 37 {
					fmt.Println("[voting] - metaType: ", l)
				}
				if err != nil {
					return nil, err
				}
				Logger.log.Warn("Metadata type:", metaType, "\n")

				var tx metadata.Transaction
				txs := []metadata.Transaction{}
				switch metaType {
				case component.RewardDCBProposalSubmitterIns:
					fmt.Println("[voting]-RewardDCBProposalSubmitterIns")
					rewardProposalSubmitter := frombeaconins.RewardProposalSubmitterIns{}
					err := json.Unmarshal([]byte(l[2]), &rewardProposalSubmitter)
					if err != nil {
						return nil, err
					}
					tx, err := rewardProposalSubmitter.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase, common.DCBBoard)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, tx)
				case component.RewardGOVProposalSubmitterIns:
					fmt.Println("[voting]-RewardGOVProposalSubmitterIns")
					rewardProposalSubmitter := frombeaconins.RewardProposalSubmitterIns{}
					err := json.Unmarshal([]byte(l[2]), &rewardProposalSubmitter)
					if err != nil {
						return nil, err
					}
					tx, err := rewardProposalSubmitter.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase, common.GOVBoard)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, tx)
				case component.RewardDCBProposalVoterIns:
					fmt.Println("[voting]-RewardDCBProposalVoterIns")
					rewardProposalVoter := frombeaconins.RewardProposalVoterIns{}
					err := json.Unmarshal([]byte(l[2]), &rewardProposalVoter)
					if err != nil {
						return nil, err
					}
					tx, err := rewardProposalVoter.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase, common.DCBBoard)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, tx)
				case component.RewardGOVProposalVoterIns:
					fmt.Println("[voting]-RewardGOVProposalVoterIns")
					rewardProposalVoter := frombeaconins.RewardProposalVoterIns{}
					err := json.Unmarshal([]byte(l[2]), &rewardProposalVoter)
					if err != nil {
						return nil, err
					}
					tx, err := rewardProposalVoter.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase, common.GOVBoard)
					if err != nil {
						return nil, err
					}
					resTxs = append(resTxs, tx)
				case metadata.CrowdsalePaymentMeta:
					txs, err = blockgen.buildPaymentForCrowdsale(l[2], unspentTokens, producerPrivateKey)

				case metadata.TradeActivationMeta:
					txs, err = blockgen.buildTradeActivationTx(l[2], unspentTokens, producerPrivateKey, tradeActivated)

				case metadata.BuyFromGOVRequestMeta:
					contentStr := l[3]
					sellingBondsParamsStr := l[4]
					txs, err = blockgen.buildBuyBondsFromGOVRes(l[2], contentStr, sellingBondsParamsStr, producerPrivateKey)

				case metadata.BuyGOVTokenRequestMeta:
					contentStr := l[3]
					txs, err = blockgen.buildBuyGOVTokensRes(l[2], contentStr, producerPrivateKey)

				case metadata.BuyBackRequestMeta:
					buyBackInfoStr := l[3]
					prevBuySellResMetaStr := l[4]
					txs, err = blockgen.buildBuyBackRes(l[2], buyBackInfoStr, prevBuySellResMetaStr, producerPrivateKey)

					// todo @0xjackalope move meta to ins?
				case component.SendBackTokenVoteBoardFailIns:
					fmt.Println("[voting]-SendBackTokenVoteBoardFailIns")
					sendBackTokenVoteFail := frombeaconins.TxSendBackTokenVoteFailIns{}
					err := json.Unmarshal([]byte(l[2]), &sendBackTokenVoteFail)
					if err != nil {
						return nil, err
					}

					tx, err = sendBackTokenVoteFail.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase, blockgen.chain, shardID)
					fmt.Println("[voting]-SendBackTokenVoteBoardFailIns Ok, tx:", tx.GetMetadata())
					txs = append(txs, tx)

				case metadata.SendBackTokenToOldSupporterMeta:
					fmt.Println("[voting]-SendBackTokenToOldSupporterMeta")
					sendBackTokenToOldSupporter := frombeaconins.TxSendBackTokenToOldSupporterIns{}
					err := json.Unmarshal([]byte(l[2]), &sendBackTokenToOldSupporter)
					if err != nil {
						return nil, err
					}
					tx, err = sendBackTokenToOldSupporter.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase, blockgen.chain, shardID)

					if err != nil {
						return nil, err
					}
					fmt.Println("[voting]-SendBackTokenToOldSupporterMeta ok, tx:", tx.GetMetadata())
					txs = append(txs, tx)

				case component.ShareRewardOldDCBBoardSupportterIns, component.ShareRewardOldGOVBoardSupportterIns:
					fmt.Println("[voting]-ShareRewardOldDCBBoardSupportterIns ok, tx:", tx.GetMetadata())
					shareRewardOldBoard := frombeaconins.ShareRewardOldBoardIns{}
					err := json.Unmarshal([]byte(l[2]), &shareRewardOldBoard)
					if err != nil {
						return nil, err
					}
					tx, err = shareRewardOldBoard.BuildTransaction(producerPrivateKey, blockgen.chain.config.DataBase)
					txs = append(txs, tx)

				case metadata.IssuingRequestMeta:
					issuingInfoStr := l[3]
					txs, err = blockgen.buildIssuingRes(l[2], issuingInfoStr, producerPrivateKey, shardID)

				case metadata.ContractingRequestMeta:
					contractingInfoStr := l[3]
					txs, err = blockgen.buildContractingRes(l[2], contractingInfoStr, producerPrivateKey)

				case metadata.OracleRewardMeta:
					evaluationStr := l[3]
					txs, err = blockgen.buildOracleRewardTxs(evaluationStr, producerPrivateKey)

				case metadata.ShardBlockSalaryRequestMeta:
					salaryReqInfoStr := l[3]
					txs, err = blockgen.buildSalaryRes(l[2], salaryReqInfoStr, producerPrivateKey)
				}

				if err != nil {
					return nil, err
				}
				if len(txs) > 0 {
					resTxs = append(resTxs, txs...)
				}
			}
		}
	}
	return resTxs, nil
}

func (blockgen *BlkTmplGenerator) buildStabilityResponseTxsAtShardOnly(txs []metadata.Transaction, producerPrivateKey *privacy.PrivateKey) ([]metadata.Transaction, error) {
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
	fmt.Println("[voting] - AddVoteBoard: ", inst)
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

func (chain *BlockChain) AddSubmitProposal(inst string) error {
	newInst, err := fromshardins.NewSubmitProposalInsFromStr(inst)
	fmt.Println("[voting] - AddSubmitProposal: ", inst)
	if err != nil {
		return err
	}
	boardType := newInst.BoardType
	submitter := newInst.SubmitProposal.SubmitterPayment
	err1 := chain.GetDatabase().AddSubmitProposalDB(
		boardType,
		newInst.SubmitProposal.ConstitutionIndex,
		newInst.SubmitProposal.ProposalTxID.GetBytes(),
		submitter.Bytes(),
	)
	if err1 != nil {
		return err1
	}
	return nil
}

func (chain *BlockChain) AddVoteProposal(inst string) error {
	fmt.Println("[voting] - AddVoteProposal: ", inst)
	newInst, err := fromshardins.NewNormalVoteProposalInsFromStr(inst)
	if err != nil {
		return err
	}
	// step 4 hyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
	err = chain.GetDatabase().AddVoteProposalDB(
		newInst.BoardType,
		newInst.VoteProposal.ConstitutionIndex,
		newInst.VoteProposal.VoterPayment.Bytes(),
		newInst.VoteProposal.ProposalTxID.GetBytes(),
	)
	gg := lvdb.ViewDBByPrefix(chain.config.DataBase, lvdb.VoteProposalPrefix)
	_ = gg

	if err != nil {
		return err
	}
	return nil
}
