package blockchain

import (
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
)

func (blockgen *BlkTmplGenerator) calculateInterestPaid(tx metadata.Transaction) (uint64, error) {
	paymentMeta := tx.GetMetadata().(*metadata.LoanPayment)
	principle, interest, deadline, err := blockgen.chain.config.DataBase.GetLoanPayment(paymentMeta.LoanID)
	if err != nil {
		return 0, err
	}

	// Get loan params
	requestMeta, err := blockgen.chain.GetLoanRequestMeta(paymentMeta.LoanID)
	if err != nil {
		return 0, err
	}

	// Only keep interest
	_, _, amount := tx.GetUniqueReceiver() // Receiver is unique and is burn address
	shardID, _ := common.GetTxSenderChain(tx.GetSenderAddrLastByte())
	totalInterest := metadata.GetTotalInterest(
		principle,
		interest,
		requestMeta.Params.InterestRate,
		requestMeta.Params.Maturity,
		deadline,
		blockgen.chain.GetChainHeight(shardID),
	)
	interestPaid := amount
	if amount > totalInterest {
		interestPaid = totalInterest
	}
	return interestPaid, nil
}
