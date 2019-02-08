package blockchain

import (
	"bytes"
	"strconv"

	"github.com/ninjadotorg/constant/metadata"
	"github.com/pkg/errors"
)

func (bsb *BestStateBeacon) processStabilityInstruction(inst []string) error {
	if len(inst) < 2 {
		return nil // Not error, just not loan instruction
	}
	switch inst[0] {
	case strconv.Itoa(metadata.LoanRequestMeta):
		return bsb.processLoanRequestInstruction(inst)
	case strconv.Itoa(metadata.LoanResponseMeta):
		return bsb.processLoanResponseInstruction(inst)
	case strconv.Itoa(metadata.LoanPaymentMeta):
		return bsb.processLoanPaymentInstruction(inst)

	case strconv.Itoa(metadata.AcceptDCBProposalMeta):
		return bsb.processAcceptDCBProposalInstruction(inst)
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

func (bsb *BestStateBeacon) processLoanPaymentInstruction(inst []string) error {
	_, amount, err := metadata.ParseLoanPaymentActionValue(inst[1])
	if err != nil {
		return err
	}

	// Update fund of DCB
	bsb.StabilityInfo.BankFund += amount
	return nil
}

func (bsb *BestStateBeacon) processLoanPaymentInstruction(inst []string) error {
	_, amount, err := metadata.ParseLoanPaymentActionValue(inst[1])
	if err != nil {
		return err
	}

	// Update fund of DCB
	bsb.StabilityInfo.BankFund += amount
	return nil
}

func (bsb *BestStateBeacon) processLoanPaymentInstruction(inst []string) error {
	_, amount, err := metadata.ParseLoanPaymentActionValue(inst[1])
	if err != nil {
		return err
	}

	// Update fund of DCB
	bsb.StabilityInfo.BankFund += amount
	return nil
}

func (bsb *BestStateBeacon) processAcceptDCBProposalInstruction(inst []string) error {
	// TODO(@0xjackalope): process other dcb params here
	dcbParams, err := metadata.ParseAcceptDCBProposalMetadataActionValue(inst[1])
	if err != nil {
		return err
	}
	// Store saledata in db
	for _, data := range dcbParams.ListSaleData {
		key := getSaleDataKeyBeacon(data.SaleID)
		if _, ok := bsb.Params[key]; ok {
			// TODO(@0xbunyip): support update crowdsale data
			continue
		}
		value := getSaleDataValueBeacon(data)
		bsb.Params[key] = value
	}
	return nil
}
