package blockchain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
	"github.com/ninjadotorg/constant/transaction"
)

type dividendPair struct {
	DividendID uint64
	ForDCB     bool
}

func (bc *BlockChain) processLoanPayment(tx metadata.Transaction) error {
	_, _, value := tx.GetUniqueReceiver()
	meta := tx.GetMetadata().(*metadata.LoanPayment)
	principle, interest, deadline, err := bc.config.DataBase.GetLoanPayment(meta.LoanID)
	requestMeta, err := bc.GetLoanRequestMeta(meta.LoanID)
	if err != nil {
		return err
	}
	fmt.Printf("[db]pid: %d, %d, %d\n", principle, interest, deadline)

	// Pay interest
	interestPerTerm := metadata.GetInterestPerTerm(principle, requestMeta.Params.InterestRate)
	shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	height := bc.GetChainHeight(shardID)
	totalInterest := metadata.GetTotalInterest(
		principle,
		interest,
		requestMeta.Params.InterestRate,
		requestMeta.Params.Maturity,
		deadline,
		height,
	)
	fmt.Printf("[db]perTerm, totalInt: %d, %d\n", interestPerTerm, totalInterest)
	termInc := uint64(0)
	if value <= totalInterest { // Pay all to cover interest
		if interestPerTerm > 0 {
			if value >= interest {
				termInc = 1 + uint64((value-interest)/interestPerTerm)
				interest = interestPerTerm - (value-interest)%interestPerTerm
			} else {
				interest -= value
			}
		}
	} else { // Pay enough to cover interest, the rest go to principle
		if value-totalInterest > principle {
			principle = 0
		} else {
			principle -= value - totalInterest
		}
		if totalInterest >= interest { // This payment pays for interest
			if interestPerTerm > 0 {
				termInc = 1 + uint64((totalInterest-interest)/interestPerTerm)
				interest = interestPerTerm
			}
		}
	}
	fmt.Printf("termInc: %d\n", termInc)
	deadline = deadline + termInc*requestMeta.Params.Maturity

	return bc.config.DataBase.StoreLoanPayment(meta.LoanID, principle, interest, deadline)
}

func (bc *BlockChain) ProcessLoanForBlock(block *ShardBlock) error {
	for _, tx := range block.Body.Transactions {
		switch tx.GetMetadataType() {
		case metadata.LoanUnlockMeta:
			{
				// Update loan payment info after withdrawing Constant
				tx := tx.(*transaction.Tx)
				meta := tx.GetMetadata().(*metadata.LoanUnlock)
				fmt.Printf("Found tx %x of type loan unlock\n", tx.Hash()[:])
				fmt.Printf("LoanID: %x\n", meta.LoanID)
				requestMeta, err := bc.GetLoanRequestMeta(meta.LoanID)
				if err != nil {
					return err
				}
				principle := requestMeta.LoanAmount
				interest := metadata.GetInterestPerTerm(principle, requestMeta.Params.InterestRate)
				err = bc.config.DataBase.StoreLoanPayment(meta.LoanID, principle, interest, uint64(block.Header.Height))
				fmt.Printf("principle: %d\ninterest: %d\nblock: %d\n", principle, interest, uint64(block.Header.Height))
				if err != nil {
					return err
				}
			}
		case metadata.LoanPaymentMeta:
			{
				bc.processLoanPayment(tx)
			}
		}
	}
	return nil
}

func (bc *BlockChain) processDividendPayment(receiversToRemove map[dividendPair][][]byte) error {
	for pair, receivers := range receiversToRemove {
		// Get list of token holders left
		paymentAddresses, amounts, _, _ := bc.config.DataBase.GetDividendReceiversForID(pair.DividendID, pair.ForDCB)

		// Update list of token holders left
		addrNotPaid := []privacy.PaymentAddress{}
		amountsNotPaid := []uint64{}
		for i, addr := range paymentAddresses {
			remove := false
			for _, pubkey := range receivers {
				if bytes.Equal(pubkey, addr.Pk[:]) {
					remove = true
					break
				}
			}
			if !remove {
				addrNotPaid = append(addrNotPaid, addr)
				amountsNotPaid = append(amountsNotPaid, amounts[i])
			}
		}
		err := bc.config.DataBase.StoreDividendReceiversForID(pair.DividendID, pair.ForDCB, addrNotPaid, amountsNotPaid)
		if err != nil {
			return err
		}
	}
	return nil
}

func (bc *BlockChain) ProcessDividendForBlock(block *ShardBlock) error {
	receiversToRemove := map[dividendPair][][]byte{}
	for _, tx := range block.Body.Transactions {
		switch tx.GetMetadataType() {
		case metadata.DividendSubmitMeta:
			// Store current list of token holders to local state
			tx := tx.(*transaction.Tx)
			meta := tx.GetMetadata().(*metadata.DividendSubmit)
			_, holders, amounts, err := bc.GetAmountPerAccount(meta.TokenID)
			if err != nil {
				return err
			}
			forDCB := meta.TokenID.IsEqual(&common.DCBTokenID)
			err = bc.config.DataBase.StoreDividendReceiversForID(meta.DividendID, forDCB, holders, amounts)
			if err != nil {
				return err
			}

		case metadata.DividendPaymentMeta:
			_, receiver, _ := tx.GetUniqueReceiver()
			meta := tx.GetMetadata().(*metadata.DividendPayment)
			forDCB := meta.TokenID.IsEqual(&common.DCBTokenID)
			pair := dividendPair{
				DividendID: meta.DividendID,
				ForDCB:     forDCB,
			}
			receiversToRemove[pair] = append(receiversToRemove[pair], receiver)
		}
	}
	return bc.processDividendPayment(receiversToRemove)
}

func (bc *BlockChain) ProcessVotingForBlock(block *ShardBlock) error {
	DCBHelper := DCBConstitutionHelper{}
	GOVHelper := GOVConstitutionHelper{}
	nextDCBConstitutionIndex := bc.BestState.Beacon.StabilityInfo.DCBConstitution.ConstitutionIndex + 1
	nextGOVConstitutionIndex := bc.BestState.Beacon.StabilityInfo.GOVConstitution.ConstitutionIndex + 1

	for _, tx := range block.Body.Transactions {
		switch tx.GetMetadataType() {
		case metadata.AcceptDCBBoardMeta:
			err := bc.BestState.Beacon.UpdateDCBBoard(tx)
			if err != nil {
				return err
			}
		case metadata.AcceptGOVBoardMeta:
			err := bc.BestState.Beacon.UpdateGOVBoard(tx)
			if err != nil {
				return err
			}
		case metadata.AcceptDCBProposalMeta:
			err := bc.BestState.Beacon.updateConstitution(tx, bc, DCBHelper)
			if err != nil {
				return err
			}
			underlieMetadata := tx.GetMetadata().(*metadata.AcceptDCBProposalMetadata)
			err1 := bc.config.DataBase.TakeVoteTokenFromWinner(common.DCBBoard, nextDCBConstitutionIndex, underlieMetadata.AcceptProposalMetadata.Voter.PaymentAddress, underlieMetadata.AcceptProposalMetadata.Voter.AmountOfVote)
			if err1 != nil {
				return err1
			}
			err2 := bc.config.DataBase.SetNewProposalWinningVoter(common.DCBBoard, nextDCBConstitutionIndex, underlieMetadata.AcceptProposalMetadata.Voter.PaymentAddress)
			if err2 != nil {
				return err2
			}
		case metadata.AcceptGOVProposalMeta:
			err := bc.BestState.Beacon.updateConstitution(tx, bc, GOVHelper)
			if err != nil {
				return err
			}
			underlieMetadata := tx.GetMetadata().(*metadata.AcceptGOVProposalMetadata)
			err1 := bc.config.DataBase.TakeVoteTokenFromWinner(
				common.GOVBoard,
				nextGOVConstitutionIndex,
				underlieMetadata.AcceptProposalMetadata.Voter.PaymentAddress,
				underlieMetadata.AcceptProposalMetadata.Voter.AmountOfVote,
			)
			if err1 != nil {
				return err1
			}
			err2 := bc.config.DataBase.SetNewProposalWinningVoter(common.GOVBoard, nextGOVConstitutionIndex, underlieMetadata.AcceptProposalMetadata.Voter.PaymentAddress)
			if err2 != nil {
				return err2
			}
		case metadata.RewardDCBProposalSubmitterMeta:
			bc.BestState.Beacon.UpdateDCBFund(tx)
		case metadata.RewardGOVProposalSubmitterMeta:
			bc.BestState.Beacon.UpdateGOVFund(tx)

		case metadata.VoteDCBBoardMeta:
			{
				txCustomToken := tx.(*transaction.TxCustomToken)
				voteAmount := txCustomToken.GetAmountOfVote()
				voteDCBBoardMetadata := txCustomToken.Metadata.(*metadata.VoteDCBBoardMetadata)
				boardIndex := bc.BestState.Beacon.StabilityInfo.DCBGovernor.BoardIndex + 1
				err := bc.config.DataBase.AddVoteBoard(
					common.DCBBoard,
					boardIndex,
					txCustomToken.TxTokenData.Vins[0].PaymentAddress,
					voteDCBBoardMetadata.VoteBoardMetadata.CandidatePaymentAddress,
					voteAmount,
				)
				if err != nil {
					return err
				}
			}
		case metadata.VoteGOVBoardMeta:
			{
				txCustomToken := tx.(*transaction.TxCustomToken)
				voteAmount := txCustomToken.GetAmountOfVote()
				voteDCBBoardMetadata := txCustomToken.Metadata.(*metadata.VoteDCBBoardMetadata)
				boardIndex := bc.BestState.Beacon.StabilityInfo.DCBGovernor.BoardIndex + 1
				err := bc.config.DataBase.AddVoteBoard(
					common.DCBBoard,
					boardIndex,
					txCustomToken.TxTokenData.Vins[0].PaymentAddress,
					voteDCBBoardMetadata.VoteBoardMetadata.CandidatePaymentAddress,
					voteAmount,
				)
				if err != nil {
					return err
				}
			}

		case metadata.SendInitDCBVoteTokenMeta:
			meta := tx.GetMetadata().(*metadata.SendInitDCBVoteTokenMetadata)
			err := bc.config.DataBase.SendInitVoteToken(
				common.DCBBoard,
				bc.BestState.Beacon.StabilityInfo.DCBGovernor.BoardIndex,
				meta.SendInitVoteTokenMetadata.ReceiverPaymentAddress,
				meta.SendInitVoteTokenMetadata.Amount,
			)
			if err != nil {
				return err
			}
		case metadata.SendInitGOVVoteTokenMeta:
			meta := tx.GetMetadata().(*metadata.SendInitGOVVoteTokenMetadata)
			err := bc.config.DataBase.SendInitVoteToken(
				common.GOVBoard,
				bc.BestState.Beacon.StabilityInfo.GOVGovernor.BoardIndex,
				meta.SendInitVoteTokenMetadata.ReceiverPaymentAddress,
				meta.SendInitVoteTokenMetadata.Amount,
			)
			if err != nil {
				return err
			}

			//vote proposal
		case metadata.SealedLv3DCBVoteProposalMeta:
			err := bc.config.DataBase.AddVoteLv3Proposal(common.DCBBoard, nextDCBConstitutionIndex, tx.GetMetadata().Hash())
			if err != nil {
				return err
			}
		case metadata.SealedLv2DCBVoteProposalMeta:
			underlieMetadata := tx.GetMetadata().(*metadata.SealedLv2DCBVoteProposalMetadata)
			err := bc.config.DataBase.AddVoteLv1or2Proposal(common.DCBBoard, nextDCBConstitutionIndex, &underlieMetadata.SealedLv2VoteProposalMetadata.PointerToLv3VoteProposal)
			if err != nil {
				return err
			}
		case metadata.SealedLv1DCBVoteProposalMeta:
			underlieMetadata := tx.GetMetadata().(*metadata.SealedLv1DCBVoteProposalMetadata)
			err := bc.config.DataBase.AddVoteLv1or2Proposal(common.DCBBoard, nextDCBConstitutionIndex, &underlieMetadata.SealedLv1VoteProposalMetadata.PointerToLv3VoteProposal)
			if err != nil {
				return err
			}
		case metadata.NormalDCBVoteProposalFromOwnerMeta:
			underlieMetadata := tx.GetMetadata().(*metadata.NormalDCBVoteProposalFromOwnerMetadata)
			err := bc.config.DataBase.AddVoteNormalProposalFromOwner(common.DCBBoard, nextDCBConstitutionIndex, &underlieMetadata.NormalVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal, underlieMetadata.NormalVoteProposalFromOwnerMetadata.VoteProposal.ToBytes())
			if err != nil {
				return err
			}
		case metadata.NormalDCBVoteProposalFromSealerMeta:
			underlieMetadata := tx.GetMetadata().(*metadata.NormalDCBVoteProposalFromSealerMetadata)
			err := bc.config.DataBase.AddVoteNormalProposalFromSealer(common.DCBBoard, nextDCBConstitutionIndex, &underlieMetadata.NormalVoteProposalFromSealerMetadata.PointerToLv3VoteProposal, underlieMetadata.NormalVoteProposalFromSealerMetadata.VoteProposal.ToBytes())
			if err != nil {
				return err
			}
		case metadata.SealedLv3GOVVoteProposalMeta:
			underlieMetadata := tx.GetMetadata().(*metadata.SealedLv3GOVVoteProposalMetadata)
			err := bc.config.DataBase.AddVoteLv3Proposal(common.GOVBoard, nextGOVConstitutionIndex, underlieMetadata.Hash())
			if err != nil {
				return err
			}
		case metadata.SealedLv2GOVVoteProposalMeta:
			underlieMetadata := tx.GetMetadata().(*metadata.SealedLv2GOVVoteProposalMetadata)
			err := bc.config.DataBase.AddVoteLv1or2Proposal(common.GOVBoard, nextGOVConstitutionIndex, &underlieMetadata.SealedLv2VoteProposalMetadata.PointerToLv3VoteProposal)
			if err != nil {
				return err
			}
		case metadata.SealedLv1GOVVoteProposalMeta:
			underlieMetadata := tx.GetMetadata().(*metadata.SealedLv1GOVVoteProposalMetadata)
			err := bc.config.DataBase.AddVoteLv1or2Proposal(common.GOVBoard, nextGOVConstitutionIndex, &underlieMetadata.SealedLv1VoteProposalMetadata.PointerToLv3VoteProposal)
			if err != nil {
				return err
			}
		case metadata.NormalGOVVoteProposalFromOwnerMeta:
			underlieMetadata := tx.GetMetadata().(*metadata.NormalGOVVoteProposalFromOwnerMetadata)
			err := bc.config.DataBase.AddVoteNormalProposalFromOwner(common.GOVBoard, nextGOVConstitutionIndex, &underlieMetadata.NormalVoteProposalFromOwnerMetadata.PointerToLv3VoteProposal, underlieMetadata.NormalVoteProposalFromOwnerMetadata.VoteProposal.ToBytes())
			if err != nil {
				return err
			}
		case metadata.NormalGOVVoteProposalFromSealerMeta:
			underlieMetadata := tx.GetMetadata().(*metadata.NormalGOVVoteProposalFromSealerMetadata)
			err := bc.config.DataBase.AddVoteNormalProposalFromSealer(common.GOVBoard, nextGOVConstitutionIndex, &underlieMetadata.NormalVoteProposalFromSealerMetadata.PointerToLv3VoteProposal, underlieMetadata.NormalVoteProposalFromSealerMetadata.VoteProposal.ToBytes())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (bc *BlockChain) StoreMetadataInstructions(inst []string, shardID byte) error {
	if len(inst) < 2 {
		return nil // Not error, just not stability instruction
	}
	switch inst[0] {
	case strconv.Itoa(metadata.IssuingResponseMeta):
		return bc.storeIssuingResponseInstruction(inst, shardID)
	case strconv.Itoa(metadata.ContractingReponseMeta):
		return bc.storeContractingResponseInstruction(inst, shardID)
	}
	return nil
}

func (bc *BlockChain) storeIssuingResponseInstruction(inst []string, shardID byte) error {
	if strconv.Itoa(int(shardID)) != inst[1] {
		return nil
	}

	issuingInfo := &IssuingInfo{}
	err := json.Unmarshal([]byte(inst[3]), issuingInfo)
	if err != nil {
		return err
	}

	instType := inst[2]
	return bc.config.DataBase.StoreIssuingInfo(issuingInfo.RequestedTxID, issuingInfo.Amount, instType)
}

func (bc *BlockChain) storeContractingResponseInstruction(inst []string, shardID byte) error {
	if strconv.Itoa(int(shardID)) != inst[1] {
		return nil
	}

	contractingInfo := &ContractingInfo{}
	err := json.Unmarshal([]byte(inst[3]), contractingInfo)
	if err != nil {
		return err
	}

	instType := inst[2]
	return bc.config.DataBase.StoreContractingInfo(contractingInfo.RequestedTxID, contractingInfo.BurnedConstAmount, instType)
}

//func (bc *BlockChain) UpdateDividendPayout(block *Block) error {
//	for _, tx := range block.Transactions {
//		switch tx.GetMetadataType() {
//		case metadata.DividendMeta:
//			{
//				tx := tx.(*transaction.Tx)
//				meta := tx.Metadata.(*metadata.Dividend)
//				if tx.Proof == nil {
//					return errors.New("Miss output in tx")
//				}
//				for _, _ = range tx.Proof.OutputCoins {
//					keySet := cashec.KeySet{
//						PaymentAddress: meta.PaymentAddress,
//					}
//					vouts, err := bc.GetUnspentTxCustomTokenVout(keySet, meta.TokenID)
//					if err != nil {
//						return err
//					}
//					for _, vout := range vouts {
//						txHash := vout.GetTxCustomTokenID()
//						err := bc.config.DataBase.UpdateRewardAccountUTXO(meta.TokenID, keySet.PaymentAddress.Pk, &txHash, vout.GetIndex())
//						if err != nil {
//							return err
//						}
//					}
//				}
//			}
//		}
//	}
//	return nil
//}

// func (bc *BlockChain) ProcessCMBTxs(block *Block) error {
// 	for _, tx := range block.Transactions {
// 		switch tx.GetMetadataType() {
// 		case metadata.CMBInitRequestMeta:
// 			{
// 				err := bc.processCMBInitRequest(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBInitResponseMeta:
// 			{
// 				err := bc.processCMBInitResponse(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBInitRefundMeta:
// 			{
// 				err := bc.processCMBInitRefund(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBDepositSendMeta:
// 			{
// 				err := bc.processCMBDepositSend(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBWithdrawRequestMeta:
// 			{
// 				err := bc.processCMBWithdrawRequest(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBWithdrawResponseMeta:
// 			{
// 				err := bc.processCMBWithdrawResponse(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	}

// 	// Penalize late response for cmb withdraw request
// 	return bc.findLateWithdrawResponse(uint64(block.Header.Height))
// }
