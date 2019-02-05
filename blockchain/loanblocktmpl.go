package blockchain

import (
	"bytes"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/pkg/errors"
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
	shardID := common.GetShardIDFromLastByte(tx.GetSenderAddrLastByte())
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

func (bsb *BestStateBeacon) processLoanInstruction(inst []string) error {
	if len(inst) < 2 {
		return nil // Not error, just not loan instruction
	}
	switch inst[0] {
	case strconv.Itoa(metadata.LoanRequestMeta):
		return bsb.processLoanRequestInstruction(inst)
	case strconv.Itoa(metadata.LoanResponseMeta):
		return bsb.processLoanResponseInstruction(inst)
	}
	return nil
}

func (bsb *BestStateBeacon) processLoanRequestInstruction(inst []string) error {
	loanID, txHash, err := metadata.ParseLoanRequestActionValue(inst[1])
	if err != nil {
		return err
	}
	// Check if no loan request with the same id existed
	key := getLoanRequestKeyBeacon(loanID)
	if _, ok := bsb.Params[key]; ok {
		return errors.Errorf("LoanID already existed: %x", loanID)
	}

	// Save loan request on beacon shard
	value := txHash.String()
	bsb.Params[key] = value
	return nil
}

func (bsb *BestStateBeacon) processLoanResponseInstruction(inst []string) error {
	loanID, sender, resp, err := metadata.ParseLoanResponseActionValue(inst[1])
	if err != nil {
		return err
	}

	// For safety, beacon shard checks if loan request existed
	key := getLoanRequestKeyBeacon(loanID)
	if _, ok := bsb.Params[key]; !ok {
		return errors.Errorf("LoanID not existed: %x", loanID)
	}

	// Get current list of responses
	lrds := []*LoanRespData{}
	key = getLoanResponseKeyBeacon(loanID)
	if value, ok := bsb.Params[key]; ok {
		lrds, err = parseLoanResponseValueBeacon(value)
		if err != nil {
			return err
		}
	}

	// Check if same member doesn't respond twice
	for _, resp := range lrds {
		if bytes.Equal(resp.SenderPubkey, sender) {
			return errors.Errorf("Sender %x already responded to loanID %x", sender, loanID)
		}
	}

	// Update list of responses
	lrd := &LoanRespData{
		SenderPubkey: sender,
		Response:     resp,
	}
	lrds = append(lrds, lrd)
	value := getLoanResponseValueBeacon(lrds)
	bsb.Params[key] = value
	return nil
}
