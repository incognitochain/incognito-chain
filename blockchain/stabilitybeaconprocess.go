package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/metadata/frombeaconins"

	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/pkg/errors"
)

func (bsb *BestStateBeacon) processStabilityInstruction(inst []string) error {
	if inst[0] == InitAction {
		// init data for network
		var err error
		switch inst[1] {
		case salaryPerTx:
			{
				bsb.StabilityInfo.GOVConstitution.GOVParams.SalaryPerTx, err = strconv.ParseUint(inst[2], 10, 64)
				if err != nil {
					return err
				}
			}
		case basicSalary:
			{
				bsb.StabilityInfo.GOVConstitution.GOVParams.BasicSalary, err = strconv.ParseUint(inst[2], 10, 64)
				if err != nil {
					return err
				}
			}
		case salaryFund:
			{
				bsb.StabilityInfo.SalaryFund, err = strconv.ParseUint(inst[2], 10, 64)
				if err != nil {
					return err
				}
			}
		case feePerTxKb:
			{
				bsb.StabilityInfo.GOVConstitution.GOVParams.FeePerKbTx, err = strconv.ParseUint(inst[2], 10, 64)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}
	if len(inst) < 2 {
		return nil // Not error, just not stability instruction
	}
	switch inst[0] {
	case strconv.Itoa(metadata.LoanRequestMeta):
		return bsb.processLoanRequestInstruction(inst)
	case strconv.Itoa(metadata.LoanResponseMeta):
		return bsb.processLoanResponseInstruction(inst)
	case strconv.Itoa(component.AcceptDCBBoardIns):
		acceptDCBBoardIns := frombeaconins.AcceptDCBBoardIns{}
		err := json.Unmarshal([]byte(inst[2]), &acceptDCBBoardIns)
		if err != nil {
			return err
		}
		err = bsb.UpdateDCBBoard(acceptDCBBoardIns)
		if err != nil {
			return err
		}
	case strconv.Itoa(component.AcceptGOVBoardIns):
		acceptGOVBoardIns := frombeaconins.AcceptGOVBoardIns{}
		err := json.Unmarshal([]byte(inst[2]), &acceptGOVBoardIns)
		if err != nil {
			return err
		}
		err = bsb.UpdateGOVBoard(acceptGOVBoardIns)
		if err != nil {
			return err
		}
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
	stabilityInfo := &bsb.StabilityInfo
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

func (bsb *BestStateBeacon) processUpdateDCBProposalInstruction(ins frombeaconins.UpdateDCBConstitutionIns) error {
	dcbParams := ins.DCBParams
	//todo @0xjackalope: update new Constitution
	oldConstitution := bsb.StabilityInfo.DCBConstitution
	bsb.StabilityInfo.DCBConstitution = DCBConstitution{
		ConstitutionInfo: ConstitutionInfo{
			ConstitutionIndex:  oldConstitution.ConstitutionIndex + 1,
			StartedBlockHeight: bsb.BestBlock.Header.Height,
			ExecuteDuration:    ins.SubmitProposalInfo.ExecuteDuration,
			Explanation:        ins.SubmitProposalInfo.Explanation,
			Voter:              ins.Voter,
		},
		CurrentDCBNationalWelfare: GetOracleDCBNationalWelfare(),
		DCBParams:                 dcbParams,
	}

	// Store saledata in state

	for _, data := range dcbParams.ListSaleData {
		key := getSaleDataKeyBeacon(data.SaleID)
		if _, ok := bsb.Params[key]; ok {
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
			var err error
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

func (bsb *BestStateBeacon) processUpdateGOVProposalInstruction(ins frombeaconins.UpdateGOVConstitutionIns) error {
	panic("")
	return nil
}

func (bsb *BestStateBeacon) processDividendSubmitInstruction(inst []string) error {
	fmt.Printf("[db] beaconProcess found inst: %+v\n", inst)
	ds, err := metadata.ParseDividendSubmitActionValue(inst[2])
	if err != nil {
		fmt.Printf("[db] err parse divsub: %v\n", err)
		return err
	}

	// Save number of token for this shard
	key := getDividendSubmitKeyBeacon(ds.ShardID, ds.DividendID, ds.TokenID)
	value := getDividendSubmitValueBeacon(ds.TotalTokenAmount)
	bsb.Params[key] = value

	// If enough shard submitted token amounts, aggregate total and save to component to initiate dividend payments
	totalTokenOnAllShards := uint64(0)
	for i := byte(0); i <= byte(255); i++ {
		key := getDividendSubmitKeyBeacon(i, ds.DividendID, ds.TokenID)
		if value, ok := bsb.Params[key]; ok {
			shardTokenAmount := parseDividendSubmitValueBeacon(value)
			totalTokenOnAllShards += shardTokenAmount
		} else {
			fmt.Printf("[db] no divSub for: %d %d %x\n", i, ds.DividendID, ds.TokenID)
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
	fmt.Printf("[db] updated dividend: %d %d %d\n", totalTokenOnAllShards, cstToPayout, bsb.StabilityInfo.BankFund)
	return nil
}

func (bsb *BestStateBeacon) processCrowdsalePaymentInstruction(inst []string) error {
	fmt.Printf("[db] beaconProcess found inst: %+v\n", inst)
	// All shard update bsb, only DCB shard creates payment txs
	paymentInst, err := ParseCrowdsalePaymentInstruction(inst[2])
	if err != nil {
		return err
	}
	if paymentInst.UpdateSale {
		saleData, err := bsb.GetSaleData(paymentInst.SaleID)
		if err != nil {
			fmt.Printf("[db] error get sale data: %+v\n", err)
			return err
		}
		saleData.BuyingAmount -= paymentInst.SentAmount
		saleData.SellingAmount -= paymentInst.Amount

		key := getSaleDataKeyBeacon(paymentInst.SaleID)
		bsb.Params[key] = getSaleDataValueBeacon(saleData)
		fmt.Printf("[db] updated crowdsale: %s\n", bsb.Params[key])
	}
	return nil
}

func (bc *BlockChain) processLoanWithdrawInstruction(inst []string) error {
	loanID, principle, interest, err := metadata.ParseLoanWithdrawActionValue(inst[2])
	if err != nil {
		fmt.Printf("[db] parse err: %+v\n", err)
		return err
	}
	beaconHeight := bc.BestState.Beacon.BeaconHeight
	return bc.config.DataBase.StoreLoanPayment(loanID, principle, interest, beaconHeight)
}

func (bc *BlockChain) processLoanPayment(
	loanID []byte,
	amountSent uint64,
	interestRate uint64,
	maturity uint64,
	beaconHeight uint64,
) error {
	principle, interest, deadline, err := bc.config.DataBase.GetLoanPayment(loanID)
	if err != nil {
		return err
	}
	fmt.Printf("[db] pid: %d, %d, %d\n", principle, interest, deadline)

	// Update BANK fund
	interestPaid := metadata.CalculateInterestPaid(amountSent, principle, interest, deadline, interestRate, maturity, beaconHeight)
	bc.BestState.Beacon.StabilityInfo.BankFund += interestPaid // a little bit hacky, update fund here instead of in bsb.Update

	// Pay interest
	interestPerTerm := metadata.GetInterestPerTerm(principle, interestRate)
	totalInterest := metadata.GetTotalInterest(
		principle,
		interest,
		interestRate,
		maturity,
		deadline,
		beaconHeight,
	)
	fmt.Printf("[db] perTerm, totalInt: %d, %d\n", interestPerTerm, totalInterest)
	termInc := uint64(0)
	if amountSent <= totalInterest { // Pay all to cover interest
		if interestPerTerm > 0 {
			if amountSent >= interest {
				termInc = 1 + uint64((amountSent-interest)/interestPerTerm)
				interest = interestPerTerm - (amountSent-interest)%interestPerTerm
			} else {
				interest -= amountSent
			}
		}
	} else { // Pay enough to cover interest, the rest go to principle
		if amountSent-totalInterest > principle {
			principle = 0
		} else {
			principle -= amountSent - totalInterest
		}
		if totalInterest >= interest { // This payment pays for interest
			if interestPerTerm > 0 {
				termInc = 1 + uint64((totalInterest-interest)/interestPerTerm)
				interest = interestPerTerm
			}
		}
	}
	fmt.Printf("[db] termInc: %d\n", termInc)
	deadline = deadline + termInc*maturity

	return bc.config.DataBase.StoreLoanPayment(loanID, principle, interest, deadline)
}

func (bc *BlockChain) processLoanPaymentInstruction(inst []string) error {
	loanID, amountSent, interestRate, maturity, err := metadata.ParseLoanPaymentActionValue(inst[2])
	if err != nil {
		fmt.Printf("[db] parse err: %+v\n", err)
		return err
	}
	beaconHeight := bc.BestState.Beacon.BeaconHeight

	// Update loan payment info and BANK fund
	return bc.processLoanPayment(loanID, amountSent, interestRate, maturity, beaconHeight)
}

func (bc *BlockChain) processBeaconOnlyInstructions(block *BeaconBlock) error {
	for _, inst := range block.Body.Instructions {
		switch inst[0] {
		case strconv.Itoa(metadata.LoanWithdrawMeta):
			err := bc.processLoanWithdrawInstruction(inst)
			if err != nil {
				return err
			}

		case strconv.Itoa(metadata.LoanPaymentMeta):
			err := bc.processLoanPaymentInstruction(inst)
			if err != nil {
				return err
			}
		case strconv.Itoa(component.UpdateDCBConstitutionIns):
			return bc.processUpdateDCBConstitutionIns(inst)
		case strconv.Itoa(component.UpdateGOVConstitutionIns):
			return bc.processUpdateGOVConstitutionIns(inst)
		}
	}
	return nil
}
