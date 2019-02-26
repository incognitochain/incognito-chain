package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/pkg/errors"
)

func (bsb *BestStateBeacon) processStabilityInstruction(inst []string) error {
	if len(inst) < 2 {
		return nil // Not error, just not stability instruction
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

	case strconv.Itoa(metadata.DividendSubmitMeta):
		return bsb.processDividendSubmitInstruction(inst)

	case strconv.Itoa(metadata.CrowdsalePaymentMeta):
		return bsb.processCrowdsalePaymentInstruction(inst)

	case strconv.Itoa(metadata.BuyFromGOVRequestMeta):
		return bsb.processBuyFromGOVReqInstruction(inst)

	case strconv.Itoa(metadata.BuyBackRequestMeta):
		return bsb.processBuyBackReqInstruction(inst)

	case strconv.Itoa(metadata.BuyGOVTokenRequestMeta):
		return bsb.processBuyGOVTokenReqInstruction(inst)

	case strconv.Itoa(metadata.IssuingRequestMeta):
		return bsb.processIssuingReqInstruction(inst)

	case strconv.Itoa(metadata.ContractingRequestMeta):
		return bsb.processContractingReqInstruction(inst)

	case strconv.Itoa(metadata.ShardBlockSalaryRequestMeta):
		return bsb.processSalaryUpdateInstruction(inst)
	}
	return nil
}

func (bsb *BestStateBeacon) processSalaryUpdateInstruction(inst []string) error {
	stabilityInfo := bsb.StabilityInfo
	shardBlockSalaryInfoStr := inst[3]
	var shardBlockSalaryInfo ShardBlockSalaryInfo
	err := json.Unmarshal([]byte(shardBlockSalaryInfoStr), &shardBlockSalaryInfo)
	if err != nil {
		return err
	}

	instType := inst[2]
	if instType == "fundNotEnough" {
		stabilityInfo.SalaryFund += shardBlockSalaryInfo.ShardBlockFee
		return nil
	}
	// accepted
	stabilityInfo.SalaryFund -= shardBlockSalaryInfo.ShardBlockSalary
	stabilityInfo.SalaryFund += shardBlockSalaryInfo.ShardBlockFee
	return nil
}

func (bsb *BestStateBeacon) processContractingReqInstruction(inst []string) error {
	instType := inst[2]
	if instType == "refund" {
		return nil
	}
	// accepted
	cInfoStr := inst[3]
	var cInfo ContractingInfo
	err := json.Unmarshal([]byte(cInfoStr), &cInfo)
	if err != nil {
		return err
	}
	if bytes.Equal(cInfo.CurrencyType[:], common.USDAssetID[:]) {
		// no need to update BestStateBeacon
		return nil
	}
	// burn const by crypto
	stabilityInfo := bsb.StabilityInfo
	spendReserveData := stabilityInfo.DCBConstitution.DCBParams.SpendReserveData
	if spendReserveData == nil {
		return nil
	}
	reserveData, existed := spendReserveData[cInfo.CurrencyType]
	if !existed {
		return nil
	}
	reserveData.Amount -= cInfo.BurnedConstAmount
	return nil
}

func (bsb *BestStateBeacon) processIssuingReqInstruction(inst []string) error {
	instType := inst[2]
	if instType == "refund" {
		return nil
	}
	// accepted
	iInfoStr := inst[3]
	var iInfo IssuingInfo
	err := json.Unmarshal([]byte(iInfoStr), &iInfo)
	if err != nil {
		return err
	}
	stabilityInfo := bsb.StabilityInfo
	raiseReserveData := stabilityInfo.DCBConstitution.DCBParams.RaiseReserveData
	if raiseReserveData == nil {
		return nil
	}
	reserveData, existed := raiseReserveData[iInfo.CurrencyType]
	if !existed {
		return nil
	}
	reserveData.Amount -= iInfo.Amount
	return nil
}

func (bsb *BestStateBeacon) processBuyGOVTokenReqInstruction(inst []string) error {
	instType := inst[2]
	if instType == "refund" {
		return nil
	}
	// accepted
	contentStr := inst[3]
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return err
	}
	var buyGOVTokenReqAction BuyGOVTokenReqAction
	err = json.Unmarshal(contentBytes, &buyGOVTokenReqAction)
	if err != nil {
		return err
	}
	md := buyGOVTokenReqAction.Meta
	stabilityInfo := bsb.StabilityInfo
	sellingGOVTokensParams := stabilityInfo.GOVConstitution.GOVParams.SellingGOVTokens
	if sellingGOVTokensParams != nil {
		sellingGOVTokensParams.GOVTokensToSell -= md.Amount
		stabilityInfo.SalaryFund += (md.Amount + md.BuyPrice)
	}
	return nil
}

func (bsb *BestStateBeacon) processBuyBackReqInstruction(inst []string) error {
	instType := inst[2]
	if instType == "refund" {
		return nil
	}
	// accepted
	buyBackInfoStr := inst[3]
	var buyBackInfo BuyBackInfo
	err := json.Unmarshal([]byte(buyBackInfoStr), &buyBackInfo)
	if err != nil {
		return err
	}
	bsb.StabilityInfo.SalaryFund -= (buyBackInfo.Value + buyBackInfo.BuyBackPrice)
	return nil
}

func (bsb *BestStateBeacon) processBuyFromGOVReqInstruction(inst []string) error {
	instType := inst[2]
	if instType == "refund" {
		return nil
	}
	// accepted
	contentStr := inst[3]
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		return err
	}
	var buySellReqAction BuySellReqAction
	err = json.Unmarshal(contentBytes, &buySellReqAction)
	if err != nil {
		return err
	}
	md := buySellReqAction.Meta
	stabilityInfo := bsb.StabilityInfo
	sellingBondsParams := stabilityInfo.GOVConstitution.GOVParams.SellingBonds
	if sellingBondsParams != nil {
		sellingBondsParams.BondsToSell -= md.Amount
		stabilityInfo.SalaryFund += (md.Amount + md.BuyPrice)
	}
	return nil
}

func (bsb *BestStateBeacon) processLoanRequestInstruction(inst []string) error {
	fmt.Printf("[db] beaconProcess found inst: %+v\n", inst)
	loanID, txHash, err := metadata.ParseLoanRequestActionValue(inst[2])
	if err != nil {
		fmt.Printf("[db] parse err: %+v\n", err)
		return err
	}
	// Check if no loan request with the same id existed
	key := getLoanRequestKeyBeacon(loanID)
	if _, ok := bsb.Params[key]; ok {
		fmt.Printf("[db] LoanID existed: %t %x\n", ok, key)
		return errors.Errorf("LoanID already existed: %x", loanID)
	}

	// Save loan request on beacon shard
	value := txHash.String()
	bsb.Params[key] = value
	fmt.Printf("[db] procLoanReqInst success\n")
	return nil
}

func (bsb *BestStateBeacon) processLoanResponseInstruction(inst []string) error {
	fmt.Printf("[db] beaconProcess found inst: %+v\n", inst)
	loanID, sender, resp, err := metadata.ParseLoanResponseActionValue(inst[2])
	if err != nil {
		fmt.Printf("[db] fail parse loan resp: %+v\n", err)
		return err
	}

	// For safety, beacon shard checks if loan request existed
	key := getLoanRequestKeyBeacon(loanID)
	if _, ok := bsb.Params[key]; !ok {
		fmt.Printf("[db] loanID not existed: %t %x\n", ok, loanID)
		return errors.Errorf("LoanID not existed: %x", loanID)
	}

	// Get current list of responses
	lrds := []*LoanRespData{}
	key = getLoanResponseKeyBeacon(loanID)
	if value, ok := bsb.Params[key]; ok {
		lrds, err = parseLoanResponseValueBeacon(value)
		if err != nil {
			fmt.Printf("[db] parseLoanResp err: %+v\n", err)
			return err
		}
	}

	// Check if same member doesn't respond twice
	for _, resp := range lrds {
		if bytes.Equal(resp.SenderPubkey, sender) {
			fmt.Printf("[db] same member: %x %x\n", resp.SenderPubkey, sender)
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
	fmt.Printf("[db] procLoanRespInst success\n")
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

	// Store saledata in state
	for _, data := range dcbParams.ListSaleData {
		key := getSaleDataKeyBeacon(data.SaleID)
		if _, ok := bsb.Params[key]; ok {
			// TODO(@0xbunyip): support update crowdsale data
			continue
		}
		value := getSaleDataValueBeacon(&data)
		bsb.Params[key] = value
	}

	// Store dividend payments if needed
	if dcbParams.DividendAmount > 0 {
		key := getDCBDividendKeyBeacon()
		dividendAmounts := []uint64{}
		if value, ok := bsb.Params[key]; ok {
			dividendAmounts, err = parseDividendValueBeacon(value)
			if err != nil {
				return err
			}
		}
		dividendAmounts = append(dividendAmounts, dcbParams.DividendAmount)
		value := getDividendValueBeacon(dividendAmounts)
		bsb.Params[key] = value
	}
	return nil
}

func (bsb *BestStateBeacon) processDividendSubmitInstruction(inst []string) error {
	ds, err := metadata.ParseDividendSubmitActionValue(inst[1])
	if err != nil {
		return err
	}

	// Save number of token for this shard
	key := getDividendSubmitKeyBeacon(ds.ShardID, ds.DividendID, ds.TokenID)
	value := getDividendSubmitValueBeacon(ds.TotalTokenAmount)
	bsb.Params[key] = value

	// If enough shard submitted token amounts, aggregate total and save to params to initiate dividend payments
	totalTokenOnAllShards := uint64(0)
	for i := byte(0); i <= byte(255); i++ {
		key := getDividendSubmitKeyBeacon(i, ds.DividendID, ds.TokenID)
		if value, ok := bsb.Params[key]; ok {
			shardTokenAmount := parseDividendSubmitValueBeacon(value)
			totalTokenOnAllShards += shardTokenAmount
		} else {
			return nil
		}
	}
	forDCB := ds.TokenID.IsEqual(&common.DCBTokenID)
	_, cstToPayout := bsb.GetLatestDividendProposal(forDCB)
	if forDCB && cstToPayout > bsb.StabilityInfo.BankFund {
		cstToPayout = bsb.StabilityInfo.BankFund
	} else if !forDCB && cstToPayout > bsb.StabilityInfo.SalaryFund {
		cstToPayout = bsb.StabilityInfo.SalaryFund
	}

	key = getDividendAggregatedKeyBeacon(ds.DividendID, ds.TokenID)
	value = getDividendAggregatedValueBeacon(totalTokenOnAllShards, cstToPayout)
	bsb.Params[key] = value

	// Update institution's fund
	if forDCB {
		bsb.StabilityInfo.BankFund -= cstToPayout
	} else {
		bsb.StabilityInfo.SalaryFund -= cstToPayout
	}
	return nil
}

func (bsb *BestStateBeacon) processCrowdsalePaymentInstruction(inst []string) error {
	// All shard update bsb, only DCB shard creates payment txs
	paymentInst, err := ParseCrowdsalePaymentInstruction(inst[2])
	if err != nil {
		return err
	}
	if paymentInst.UpdateSale {
		saleData, err := bsb.GetSaleData(paymentInst.SaleID)
		if err != nil {
			return err
		}
		saleData.BuyingAmount -= paymentInst.SentAmount
		saleData.SellingAmount -= paymentInst.Amount

		key := getSaleDataKeyBeacon(paymentInst.SaleID)
		bsb.Params[key] = getSaleDataValueBeacon(saleData)
	}
	return nil
}
