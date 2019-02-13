package blockchain

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/ninjadotorg/constant/blockchain/params"
	"github.com/ninjadotorg/constant/common"
	"github.com/ninjadotorg/constant/metadata"
	"github.com/ninjadotorg/constant/privacy"
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

	case strconv.Itoa(metadata.DividendSubmitMeta):
		return bsb.processDividendSubmitInstruction(inst)

	case strconv.Itoa(metadata.CrowdsalePaymentMeta):
		return bsb.processCrowdsalePaymentInstruction(inst)
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

func buildInstructionsForCrowdsaleRequest(
	shardID byte,
	contentStr string,
	beaconBestState *BestStateBeacon,
	accumulativeValues *accumulativeValues,
) ([][]string, error) {
	saleID, priceLimit, limitSell, paymentAddress, sentAmount, err := metadata.ParseCrowdsaleRequestActionValue(contentStr)
	if err != nil {
		return nil, err
	}
	key := getSaleDataKeyBeacon(saleID)

	// Get data of current crowdsale
	var saleData *params.SaleData
	ok := false
	if saleData, ok = accumulativeValues.saleDataMap[key]; !ok {
		if value, ok := beaconBestState.Params[key]; ok {
			saleData, _ = parseSaleDataValueBeacon(value)
		} else {
			return nil, errors.Errorf("SaleID not exist: %x", saleID)
		}
	}
	accumulativeValues.saleDataMap[key] = saleData

	// Skip payment if either selling or buying asset is offchain (needs confirmation)
	if common.IsOffChainAsset(&saleData.SellingAsset) || common.IsOffChainAsset(&saleData.BuyingAsset) {
		fmt.Println("[db] crowdsale offchain asset")
		return nil, nil
	}

	inst, err := buildPaymentInstructionForCrowdsale(
		priceLimit,
		limitSell,
		paymentAddress,
		sentAmount,
		beaconBestState,
		saleData,
	)
	if err != nil {
		return nil, err
	}
	return inst, nil
}

func buildPaymentInstructionForCrowdsale(
	priceLimit uint64,
	limitSell bool,
	paymentAddress privacy.PaymentAddress,
	sentAmount uint64,
	beaconBestState *BestStateBeacon,
	saleData *params.SaleData,
) ([][]string, error) {
	// Get price for asset
	buyingAsset := saleData.BuyingAsset
	sellingAsset := saleData.SellingAsset
	buyPrice := beaconBestState.getAssetPrice(buyingAsset)
	sellPrice := beaconBestState.getAssetPrice(sellingAsset)
	if buyPrice == 0 || sellPrice == 0 {
		buyPrice = saleData.DefaultBuyPrice
		sellPrice = saleData.DefaultSellPrice
		if buyPrice == 0 || sellPrice == 0 {
			return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
		}
	}
	fmt.Printf("[db] buy and sell price: %d %d\n", buyPrice, sellPrice)

	// Check if price limit is not violated
	if limitSell && sellPrice > priceLimit {
		fmt.Printf("Price limit violated: %d %d\n", sellPrice, priceLimit)
		return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
	} else if !limitSell && buyPrice < priceLimit {
		fmt.Printf("Price limit violated: %d %d\n", buyPrice, priceLimit)
		return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
	}

	// Calculate value of asset sent in request tx
	sentAssetValue := sentAmount * buyPrice // in USD

	// Number of asset must pay to user
	paymentAmount := sentAssetValue / sellPrice

	// Check if there's still enough asset to trade
	if sentAmount > saleData.BuyingAmount || paymentAmount > saleData.SellingAmount {
		fmt.Printf("Crowdsale reached limit\n")
		return generateCrowdsalePaymentInstruction(paymentAddress, sentAmount, buyingAsset, saleData.SaleID, 0, false) // refund
	}

	// Update amount of buying/selling asset of the crowdsale
	saleData.BuyingAmount -= sentAmount
	saleData.SellingAmount -= paymentAmount

	// Build instructions
	return generateCrowdsalePaymentInstruction(paymentAddress, paymentAmount, sellingAsset, saleData.SaleID, sentAmount, true)
}
