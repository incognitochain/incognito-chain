package blockchain

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/constant-money/constant-chain/blockchain/component"
	"github.com/constant-money/constant-chain/common"
	"github.com/constant-money/constant-chain/database"
	"github.com/constant-money/constant-chain/metadata"
	"github.com/constant-money/constant-chain/metadata/frombeaconins"
	"github.com/pkg/errors"
)

func (bsb *BestStateBeacon) processStabilityInstruction(inst []string, db database.DatabaseInterface) error {
	if inst[0] == InitAction {
		// init data for network
		switch inst[1] {
		case salaryFund:
			return bsb.updateSalaryFund(inst)
		case oracleInitialPrices:
			return bsb.updateOracleInitialPrices(inst)
		}
		return nil
	}
	if len(inst) < 2 {
		return nil // Not error, just not stability instruction
	}
	switch inst[0] {
	case strconv.Itoa(component.AcceptDCBBoardIns):
		fmt.Println("[ndh] - Accept DCB Board intruction", inst)
		acceptDCBBoardIns := frombeaconins.AcceptDCBBoardIns{}
		err := json.Unmarshal([]byte(inst[2]), &acceptDCBBoardIns)
		if err != nil {
			fmt.Println("[ndh] - Accept DCB Board intruction ERRORRRRRRRRRRRRRRR", err)
			return err
		}
		err = bsb.UpdateDCBBoard(acceptDCBBoardIns)
		if err != nil {
			fmt.Println("[ndh] - Accept DCB Board intruction ERRORRRRRRRRRRRRRRR2", err)
			return err
		}
	case strconv.Itoa(component.AcceptGOVBoardIns):
		fmt.Println("[ndh] - Accept GOV Board intruction", inst)
		acceptGOVBoardIns := frombeaconins.AcceptGOVBoardIns{}
		err := json.Unmarshal([]byte(inst[2]), &acceptGOVBoardIns)
		if err != nil {
			fmt.Println("[ndh] - Accept GOV Board intruction ERRORRRRRRRRRRRRRRR", err)
			return err
		}
		err = bsb.UpdateGOVBoard(acceptGOVBoardIns)
		if err != nil {
			fmt.Println("[ndh] - Accept GOV Board intruction ERRORRRRRRRRRRRRRRR2", err)
			return err
		}
	case strconv.Itoa(component.ShareRewardOldDCBBoardSupportterIns):
		ShareRewardOldDCBBoardSupportterIns := frombeaconins.ShareRewardOldBoardIns{}
		err := json.Unmarshal([]byte(inst[2]), &ShareRewardOldDCBBoardSupportterIns)
		if err != nil {
			return err
		}
		bsb.UpdateDCBFund(-int64(ShareRewardOldDCBBoardSupportterIns.AmountOfCoin))
	case strconv.Itoa(component.ShareRewardOldGOVBoardSupportterIns):
		ShareRewardOldGOVBoardSupportterIns := frombeaconins.ShareRewardOldBoardIns{}
		err := json.Unmarshal([]byte(inst[2]), &ShareRewardOldGOVBoardSupportterIns)
		if err != nil {
			return err
		}
		bsb.UpdateGOVFund(-int64(ShareRewardOldGOVBoardSupportterIns.AmountOfCoin))
	case strconv.Itoa(component.RewardDCBProposalSubmitterIns):
		rewardDCBProposalSubmitterIns := frombeaconins.RewardProposalSubmitterIns{}
		err := json.Unmarshal([]byte(inst[2]), &rewardDCBProposalSubmitterIns)
		if err != nil {
			return err
		}
		bsb.UpdateDCBFund(-int64(rewardDCBProposalSubmitterIns.Amount))
	case strconv.Itoa(component.RewardGOVProposalSubmitterIns):
		rewardGOVProposalSubmitterIns := frombeaconins.RewardProposalSubmitterIns{}
		err := json.Unmarshal([]byte(inst[2]), &rewardGOVProposalSubmitterIns)
		if err != nil {
			return err
		}
		bsb.UpdateGOVFund(-int64(rewardGOVProposalSubmitterIns.Amount))

	case strconv.Itoa(metadata.BuyFromGOVRequestMeta):
		return bsb.processBuyFromGOVReqInstruction(inst, db)

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

	case strconv.Itoa(metadata.UpdatingOracleBoardMeta):
		return bsb.processUpdatingOracleBoardInstruction(inst)
	}
	return nil
}

func (bsb *BestStateBeacon) updateSalaryFund(inst []string) error {
	salaryFund, err := strconv.ParseUint(inst[2], 10, 64)
	if err != nil {
		return err
	}
	bsb.StabilityInfo.SalaryFund = salaryFund
	return nil
}

func (bsb *BestStateBeacon) updateOracleInitialPrices(inst []string) error {
	oracleInitialPricesStr := inst[2]
	var oracle component.Oracle
	err := json.Unmarshal([]byte(oracleInitialPricesStr), &oracle)
	if err != nil {
		return err
	}
	bsb.StabilityInfo.Oracle = oracle
	return nil
}

func (bsb *BestStateBeacon) UpdateDCBFund(amount int64) {
	t := int64(bsb.StabilityInfo.BankFund) + amount
	bsb.StabilityInfo.BankFund = uint64(t)
}

func (bsb *BestStateBeacon) UpdateGOVFund(amount int64) {
	t := int64(bsb.StabilityInfo.SalaryFund) + amount
	bsb.StabilityInfo.SalaryFund = uint64(t)
}

func (bsb *BestStateBeacon) processUpdatingOracleBoardInstruction(inst []string) error {
	instType := inst[2]
	if instType != "accepted" {
		return nil
	}
	// accepted
	updatingOracleBoardMetaStr := inst[3]
	var updatingOracleBoardMeta metadata.UpdatingOracleBoard
	err := json.Unmarshal([]byte(updatingOracleBoardMetaStr), &updatingOracleBoardMeta)
	if err != nil {
		return err
	}

	oraclePubKeys := bsb.StabilityInfo.GOVConstitution.GOVParams.OracleNetwork.OraclePubKeys
	action := updatingOracleBoardMeta.Action
	if action == metadata.Add {
		bsb.StabilityInfo.GOVConstitution.GOVParams.OracleNetwork.OraclePubKeys = append(oraclePubKeys, updatingOracleBoardMeta.OraclePubKeys...)
	} else if action == metadata.Remove {
		bsb.StabilityInfo.GOVConstitution.GOVParams.OracleNetwork.OraclePubKeys = removeOraclePubKeys(updatingOracleBoardMeta.OraclePubKeys, oraclePubKeys)
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
	stabilityInfo.SalaryFund += shardBlockSalaryInfo.ShardBlockFee
	if shardBlockSalaryInfo.ShardBlockSalary > stabilityInfo.SalaryFund {
		stabilityInfo.SalaryFund = 0
	} else {
		stabilityInfo.SalaryFund -= shardBlockSalaryInfo.ShardBlockSalary
		stabilityInfo.SalaryFund -= shardBlockSalaryInfo.BeaconBlockSalary
	}
	return nil
}

func (bsb *BestStateBeacon) processContractingReqInstruction(inst []string) error {
	instType := inst[2]
	if instType == "refund" {
		return nil
	}
	// accepted
	cInfoStr := inst[3]
	var cInfo component.ContractingInfo
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
	var iInfo component.IssuingInfo
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
	stabilityInfo := &bsb.StabilityInfo
	sellingGOVTokensParams := stabilityInfo.GOVConstitution.GOVParams.SellingGOVTokens
	if sellingGOVTokensParams != nil {
		sellingGOVTokensParams.GOVTokensToSell -= md.Amount
		stabilityInfo.SalaryFund += (md.Amount * md.BuyPrice)
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
	bsb.StabilityInfo.SalaryFund -= (buyBackInfo.Value * buyBackInfo.BuyBackPrice)
	return nil
}

func (bsb *BestStateBeacon) processBuyFromGOVReqInstruction(inst []string, db database.DatabaseInterface) error {
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
	stabilityInfo := &bsb.StabilityInfo
	sellingBondsParams := stabilityInfo.GOVConstitution.GOVParams.SellingBonds
	if sellingBondsParams != nil {
		sellingBondsParams.BondsToSell -= md.Amount
		sellingBondsParamsBytes, err := json.Marshal(sellingBondsParams)
		if err != nil {
			return err
		}
		bondID := sellingBondsParams.GetID()
		err = db.StoreSoldBondTypes(bondID, sellingBondsParamsBytes)
		if err != nil {
			return err
		}
		stabilityInfo.SalaryFund += (md.Amount * md.BuyPrice)
	}
	return nil
}

func (bsb *BestStateBeacon) processUpdateDCBProposalInstruction(ins frombeaconins.UpdateDCBConstitutionIns) error {
	dcbParams := ins.DCBParams
	oldConstitution := bsb.StabilityInfo.DCBConstitution
	fmt.Printf("[ndh] - - - - - - - - - - - - old Constitution Index %+v \n", oldConstitution)
	bsb.StabilityInfo.DCBConstitution = DCBConstitution{
		ConstitutionInfo: ConstitutionInfo{
			ConstitutionIndex:  oldConstitution.ConstitutionIndex + 1,
			StartedBlockHeight: bsb.BestBlock.Header.Height,
			ExecuteDuration:    ins.SubmitProposalInfo.ExecuteDuration,
			Explanation:        ins.SubmitProposalInfo.Explanation,
			Voters:             ins.Voters,
		},
		CurrentDCBNationalWelfare: GetOracleDCBNationalWelfare(),
		DCBParams:                 dcbParams,
	}
	if ins.SubmitProposalInfo.ConstitutionIndex == 0 {
		bsb.StabilityInfo.DCBConstitution.ConstitutionIndex = 0
	}
	return nil
}

func (bsb *BestStateBeacon) processKeepOldDCBProposalInstruction(ins frombeaconins.KeepOldProposalIns) error {
	if ins.BoardType != common.DCBBoard {
		return errors.New("Wrong board type!")
	}
	oldConstitution := bsb.StabilityInfo.DCBConstitution
	bsb.StabilityInfo.DCBConstitution = DCBConstitution{
		ConstitutionInfo: ConstitutionInfo{
			ConstitutionIndex:  oldConstitution.ConstitutionIndex + 1,
			StartedBlockHeight: bsb.BestBlock.Header.Height,
			ExecuteDuration:    oldConstitution.ExecuteDuration,
			Explanation:        oldConstitution.Explanation,
			Voters:             oldConstitution.Voters,
		},
		CurrentDCBNationalWelfare: GetOracleDCBNationalWelfare(),
		DCBParams:                 oldConstitution.DCBParams,
	}
	//TODO @dbchiem
	return nil
}

func (bsb *BestStateBeacon) processUpdateGOVProposalInstruction(ins frombeaconins.UpdateGOVConstitutionIns) error {
	oldConstitution := bsb.StabilityInfo.GOVConstitution
	fmt.Printf("[ndh] - - - - - - - - - - - - old Constitution Index %+v \n", oldConstitution)
	bsb.StabilityInfo.GOVConstitution = GOVConstitution{
		ConstitutionInfo: ConstitutionInfo{
			ConstitutionIndex:  oldConstitution.ConstitutionIndex + 1,
			StartedBlockHeight: bsb.BestBlock.Header.Height,
			ExecuteDuration:    ins.SubmitProposalInfo.ExecuteDuration,
			Explanation:        ins.SubmitProposalInfo.Explanation,
			Voters:             ins.Voters,
		},
		CurrentGOVNationalWelfare: GetOracleGOVNationalWelfare(),
		GOVParams:                 ins.GOVParams,
	}
	if ins.SubmitProposalInfo.ConstitutionIndex == 0 {
		bsb.StabilityInfo.GOVConstitution.ConstitutionIndex = 0
	}
	return nil
}

func (bsb *BestStateBeacon) processKeepOldGOVProposalInstruction(ins frombeaconins.KeepOldProposalIns) error {
	if ins.BoardType != common.GOVBoard {
		return errors.New("Wrong board type!")
	}
	oldConstitution := bsb.StabilityInfo.GOVConstitution
	fmt.Printf("[ndh] - - - - - - - - - - - - old Constitution Index %+v \n", oldConstitution)
	bsb.StabilityInfo.GOVConstitution = GOVConstitution{
		ConstitutionInfo: ConstitutionInfo{
			ConstitutionIndex:  oldConstitution.ConstitutionIndex + 1,
			StartedBlockHeight: bsb.BestBlock.Header.Height,
			ExecuteDuration:    oldConstitution.ExecuteDuration,
			Explanation:        oldConstitution.Explanation,
			Voters:             oldConstitution.Voters,
		},
		CurrentGOVNationalWelfare: GetOracleGOVNationalWelfare(),
		GOVParams:                 oldConstitution.GOVParams,
	}
	return nil
}

func (bc *BlockChain) updateDCBBuyBondInfo(bondID *common.Hash, bondAmount uint64, price uint64) error {
	amountAvail, cstPaid := bc.config.DataBase.GetDCBBondInfo(bondID)
	amountAvail += bondAmount
	cstPaid += price * bondAmount
	fmt.Println("[db] updateDCBBuyBond:", amountAvail, cstPaid)
	return bc.config.DataBase.StoreDCBBondInfo(bondID, amountAvail, cstPaid)
}

func (bc *BlockChain) updateDCBSellBondInfo(bondID *common.Hash, bondAmount uint64) error {
	amountAvail, cstPaid := bc.config.DataBase.GetDCBBondInfo(bondID)
	if amountAvail < bondAmount {
		return fmt.Errorf("invalid sell bond info update, amount available lower than payment: %d, %d", amountAvail, bondAmount)
	}

	avgPrice := uint64(0)
	if amountAvail > 0 {
		avgPrice = cstPaid / amountAvail
	}

	principleCovered := bondAmount * avgPrice
	if cstPaid < principleCovered {
		principleCovered = cstPaid
	}
	amountAvail -= bondAmount
	cstPaid -= principleCovered

	fmt.Println("[db] updateDCBSellBond:", amountAvail, cstPaid, principleCovered)
	return bc.config.DataBase.StoreDCBBondInfo(bondID, amountAvail, cstPaid)
}

func (bc *BlockChain) updateDCBSellBondProfit(bondID *common.Hash, soldValue, soldBonds uint64) uint64 {
	amountAvail, cstPaid := bc.config.DataBase.GetDCBBondInfo(bondID)
	avgPrice := uint64(0)
	if amountAvail > 0 {
		avgPrice = cstPaid / amountAvail
	}
	profit := uint64(0)
	if soldValue > avgPrice*soldBonds {
		profit = soldValue - avgPrice*soldBonds
	}
	bc.BestState.Beacon.StabilityInfo.BankFund += profit
	fmt.Println("[db] update DCB Profit:", profit, soldValue, soldBonds)
	return profit
}

func (bc *BlockChain) processCrowdsalePaymentInstruction(inst []string) error {
	// All shards update, only DCB shard creates payment txs
	fmt.Printf("[db] updateLocalState found inst: %+v\n", inst)
	paymentInst, err := ParseCrowdsalePaymentInstruction(inst[2])
	if err != nil || !paymentInst.UpdateSale {
		return err
	}

	sale, err := bc.GetSaleData(paymentInst.SaleID)
	if err != nil {
		return err
	}

	// Trading amount should always be lower or equal to sale amount
	bondAmount := paymentInst.Amount
	if sale.Buy {
		bondAmount = paymentInst.SentAmount
	}
	if sale.Amount < bondAmount {
		return fmt.Errorf("invalid crowdsale payment inst, reached limit: %d, %d", sale.Amount, bondAmount)
	}

	if sale.Buy {
		// Update average price per bond
		err = bc.updateDCBBuyBondInfo(sale.BondID, bondAmount, sale.Price)
	} else {
		// Add profit only when selling bonds
		bc.updateDCBSellBondProfit(&paymentInst.AssetID, paymentInst.SentAmount, bondAmount)

		// Update average price per bond
		err = bc.updateDCBSellBondInfo(sale.BondID, bondAmount)
	}

	if err != nil {
		return err
	}

	// Update crowdsale amount left
	sale.Amount -= bondAmount
	return bc.storeSaleData(sale)
}

func (bc *BlockChain) updateStabilityLocalState(block *BeaconBlock) error {
	for _, inst := range block.Body.Instructions {
		var err error
		switch inst[0] {
		case strconv.Itoa(component.UpdateDCBConstitutionIns):
			err = bc.processUpdateDCBConstitutionIns(inst)
		case strconv.Itoa(component.UpdateGOVConstitutionIns):
			err = bc.processUpdateGOVConstitutionIns(inst)

		case strconv.Itoa(component.KeepOldDCBProposalIns):
			err = bc.processKeepOldDCBConstitutionIns(inst)
		case strconv.Itoa(component.KeepOldGOVProposalIns):
			err = bc.processKeepOldGOVConstitutionIns(inst)

		case strconv.Itoa(metadata.CrowdsalePaymentMeta):
			err = bc.processCrowdsalePaymentInstruction(inst)
		}

		if err != nil {
			return err
		}
	}
	return bc.GetDatabase().AddConstantsPriceDB(bc.BestState.Beacon.StabilityInfo.DCBConstitution.ConstitutionIndex, bc.BestState.Beacon.StabilityInfo.Oracle.Constant)
}

func (bc *BlockChain) storeSaleData(saleData *component.SaleData) error {
	var saleRaw []byte
	var err error
	if saleRaw, err = json.Marshal(saleData); err == nil {
		err = bc.config.DataBase.StoreSaleData(saleData.SaleID, saleRaw)
	}
	return err
}

func (bc *BlockChain) storeListSaleData(dcbParams component.DCBParams) error {
	// Store saledata in state
	for _, data := range dcbParams.ListSaleData {
		if err := bc.storeSaleData(&data); err != nil {
			return err
		}
	}
	return nil
}
