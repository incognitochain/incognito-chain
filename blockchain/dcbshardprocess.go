package blockchain

import (
	"fmt"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/transaction"
)

func (self *BlockChain) ProcessLoanPayment(tx metadata.Transaction) error {
	_, _, value := tx.GetUniqueReceiver()
	meta := tx.GetMetadata().(*metadata.LoanPayment)
	principle, interest, deadline, err := self.config.DataBase.GetLoanPayment(meta.LoanID)
	requestMeta, err := self.GetLoanRequestMeta(meta.LoanID)
	if err != nil {
		return err
	}
	fmt.Printf("[db]pid: %d, %d, %d\n", principle, interest, deadline)

	// Pay interest
	interestPerTerm := metadata.GetInterestPerTerm(principle, requestMeta.Params.InterestRate)
	shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
	height := self.GetChainHeight(shardID)
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

	return self.config.DataBase.StoreLoanPayment(meta.LoanID, principle, interest, deadline)
}

func (self *BlockChain) ProcessLoanForBlock(block *ShardBlock) error {
	for _, tx := range block.Body.Transactions {
		switch tx.GetMetadataType() {
		case metadata.LoanUnlockMeta:
			{
				// Update loan payment info after withdrawing Constant
				tx := tx.(*transaction.Tx)
				meta := tx.GetMetadata().(*metadata.LoanUnlock)
				fmt.Printf("Found tx %x of type loan unlock\n", tx.Hash()[:])
				fmt.Printf("LoanID: %x\n", meta.LoanID)
				requestMeta, err := self.GetLoanRequestMeta(meta.LoanID)
				if err != nil {
					return err
				}
				principle := requestMeta.LoanAmount
				interest := metadata.GetInterestPerTerm(principle, requestMeta.Params.InterestRate)
				self.config.DataBase.StoreLoanPayment(meta.LoanID, principle, interest, uint64(block.Header.Height))
				fmt.Printf("principle: %d\ninterest: %d\nblock: %d\n", principle, interest, uint64(block.Header.Height))
			}
		case metadata.LoanPaymentMeta:
			{
				self.ProcessLoanPayment(tx)
			}
		}
	}
	return nil
}

func (self *BlockChain) ProcessDividendForBlock(block *ShardBlock) error {
	for _, tx := range block.Body.Transactions {
		switch tx.GetMetadataType() {
		case metadata.DividendSubmitMeta:
			{
				// Store current list of token holders to local state
				tx := tx.(*transaction.Tx)
				meta := tx.GetMetadata().(*metadata.DividendSubmit)
				_, holders, amounts, err := self.GetAmountPerAccount(meta.TokenID)
				if err != nil {
					return err
				}
				forDCB := meta.TokenID.IsEqual(&common.DCBTokenID)
				err = self.config.DataBase.StoreDividendReceiversForID(meta.DividendID, forDCB, holders, amounts)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

//func (self *BlockChain) UpdateDividendPayout(block *Block) error {
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
//					vouts, err := self.GetUnspentTxCustomTokenVout(keySet, meta.TokenID)
//					if err != nil {
//						return err
//					}
//					for _, vout := range vouts {
//						txHash := vout.GetTxCustomTokenID()
//						err := self.config.DataBase.UpdateRewardAccountUTXO(meta.TokenID, keySet.PaymentAddress.Pk, &txHash, vout.GetIndex())
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

// func (self *BlockChain) ProcessCMBTxs(block *Block) error {
// 	for _, tx := range block.Transactions {
// 		switch tx.GetMetadataType() {
// 		case metadata.CMBInitRequestMeta:
// 			{
// 				err := self.processCMBInitRequest(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBInitResponseMeta:
// 			{
// 				err := self.processCMBInitResponse(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBInitRefundMeta:
// 			{
// 				err := self.processCMBInitRefund(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBDepositSendMeta:
// 			{
// 				err := self.processCMBDepositSend(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBWithdrawRequestMeta:
// 			{
// 				err := self.processCMBWithdrawRequest(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		case metadata.CMBWithdrawResponseMeta:
// 			{
// 				err := self.processCMBWithdrawResponse(tx)
// 				if err != nil {
// 					return err
// 				}
// 			}
// 		}
// 	}

// 	// Penalize late response for cmb withdraw request
// 	return self.findLateWithdrawResponse(uint64(block.Header.Height))
// }
