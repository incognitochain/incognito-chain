package pdex

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

type stateProcessorV1 struct {
	stateProcessorBase
}

func (sp *stateProcessorV1) contribution(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	waitingContributions map[string]*rawdbv2.PDEContribution,
	deletedWaitingContributions map[string]*rawdbv2.PDEContribution,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	shares map[string]uint64,
) (
	map[string]*rawdbv2.PDEContribution,
	map[string]*rawdbv2.PDEContribution,
	map[string]*rawdbv2.PDEPoolForPair,
	map[string]uint64,
	error,
) {
	if len(inst) != 4 {
		err := fmt.Errorf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
	}
	contributionStatus := inst[2]

	switch contributionStatus {
	case common.PDEContributionWaitingChainStatus:
		var waitingContribution metadata.PDEWaitingContribution
		err := json.Unmarshal([]byte(inst[3]), &waitingContribution)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling content string of pde waiting contribution instruction: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
		}
		waitingContribPairKey := string(rawdbv2.BuildWaitingPDEContributionKey(beaconHeight, waitingContribution.PDEContributionPairID))
		waitingContributions[waitingContribPairKey] = &rawdbv2.PDEContribution{
			ContributorAddressStr: waitingContribution.ContributorAddressStr,
			TokenIDStr:            waitingContribution.TokenIDStr,
			Amount:                waitingContribution.ContributedAmount,
			TxReqID:               waitingContribution.TxReqID,
		}

		contribStatus := metadata.PDEContributionStatus{
			Status: byte(common.PDEContributionWaitingStatus),
		}

		contribStatusBytes, _ := json.Marshal(contribStatus)
		err = statedb.TrackPDEContributionStatus(
			stateDB,
			rawdbv2.PDEContributionStatusPrefix,
			[]byte(waitingContribution.PDEContributionPairID),
			contribStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde waiting contribution status: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
		}

	case common.PDEContributionRefundChainStatus:
		var refundContribution metadata.PDERefundContribution
		err := json.Unmarshal([]byte(inst[3]), &refundContribution)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling content string of pde refund contribution instruction: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
		}
		waitingContribPairKey := string(rawdbv2.BuildWaitingPDEContributionKey(beaconHeight, refundContribution.PDEContributionPairID))
		existingWaitingContribution, found := waitingContributions[waitingContribPairKey]
		if found {
			deletedWaitingContributions[waitingContribPairKey] = existingWaitingContribution
			delete(waitingContributions, waitingContribPairKey)
		}
		contribStatus := metadata.PDEContributionStatus{
			Status: byte(common.PDEContributionRefundStatus),
		}
		contribStatusBytes, _ := json.Marshal(contribStatus)
		err = statedb.TrackPDEContributionStatus(
			stateDB,
			rawdbv2.PDEContributionStatusPrefix,
			[]byte(refundContribution.PDEContributionPairID),
			contribStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde refund contribution status: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
		}

	case common.PDEContributionMatchedChainStatus:
		var matchedContribution metadata.PDEMatchedContribution
		err := json.Unmarshal([]byte(inst[3]), &matchedContribution)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling content string of pde matched contribution instruction: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
		}
		waitingContribPairKey := string(rawdbv2.BuildWaitingPDEContributionKey(beaconHeight, matchedContribution.PDEContributionPairID))
		existingWaitingContribution, found := waitingContributions[waitingContribPairKey]
		if !found || existingWaitingContribution == nil {
			err := fmt.Errorf("ERROR: could not find out existing waiting contribution with unique pair id: %s", matchedContribution.PDEContributionPairID)
			Logger.log.Error(err)
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
		}
		incomingWaitingContribution := &rawdbv2.PDEContribution{
			ContributorAddressStr: matchedContribution.ContributorAddressStr,
			TokenIDStr:            matchedContribution.TokenIDStr,
			Amount:                matchedContribution.ContributedAmount,
			TxReqID:               matchedContribution.TxReqID,
		}
		updateWaitingContributionPairToPool(
			beaconHeight,
			existingWaitingContribution,
			incomingWaitingContribution,
			poolPairs,
			shares,
		)
		deletedWaitingContributions[waitingContribPairKey] = existingWaitingContribution
		delete(waitingContributions, waitingContribPairKey)
		contribStatus := metadata.PDEContributionStatus{
			Status: byte(common.PDEContributionAcceptedStatus),
		}
		contribStatusBytes, _ := json.Marshal(contribStatus)
		err = statedb.TrackPDEContributionStatus(
			stateDB,
			rawdbv2.PDEContributionStatusPrefix,
			[]byte(matchedContribution.PDEContributionPairID),
			contribStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde accepted contribution status: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
		}

	case common.PDEContributionMatchedNReturnedChainStatus:
		var matchedNReturnedContrib metadata.PDEMatchedNReturnedContribution
		err := json.Unmarshal([]byte(inst[3]), &matchedNReturnedContrib)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling content string of pde matched and returned contribution instruction: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
		}
		waitingContribPairKey := string(rawdbv2.BuildWaitingPDEContributionKey(beaconHeight, matchedNReturnedContrib.PDEContributionPairID))
		waitingContribution, found := waitingContributions[waitingContribPairKey]
		if found && waitingContribution != nil {
			incomingWaitingContribution := &rawdbv2.PDEContribution{
				ContributorAddressStr: matchedNReturnedContrib.ContributorAddressStr,
				TokenIDStr:            matchedNReturnedContrib.TokenIDStr,
				Amount:                matchedNReturnedContrib.ActualContributedAmount,
				TxReqID:               matchedNReturnedContrib.TxReqID,
			}
			existingWaitingContribution := &rawdbv2.PDEContribution{
				ContributorAddressStr: waitingContribution.ContributorAddressStr,
				TokenIDStr:            waitingContribution.TokenIDStr,
				Amount:                matchedNReturnedContrib.ActualWaitingContribAmount,
				TxReqID:               waitingContribution.TxReqID,
			}
			updateWaitingContributionPairToPool(
				beaconHeight,
				existingWaitingContribution,
				incomingWaitingContribution,
				poolPairs,
				shares,
			)
			deletedWaitingContributions[waitingContribPairKey] = waitingContribution
			delete(waitingContributions, waitingContribPairKey)
		}
		pdeStatusContentBytes, err := statedb.GetPDEContributionStatus(
			stateDB,
			rawdbv2.PDEContributionStatusPrefix,
			[]byte(matchedNReturnedContrib.PDEContributionPairID),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while getting pde contribution status: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
		}
		if len(pdeStatusContentBytes) == 0 {
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, nil
		}

		var contribStatus metadata.PDEContributionStatus
		err = json.Unmarshal(pdeStatusContentBytes, &contribStatus)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde contribution status: %+v", err)
			return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
		}

		if contribStatus.Status != byte(common.PDEContributionMatchedNReturnedStatus) {
			contribStatus := metadata.PDEContributionStatus{
				Status:             byte(common.PDEContributionMatchedNReturnedStatus),
				TokenID1Str:        matchedNReturnedContrib.TokenIDStr,
				Contributed1Amount: matchedNReturnedContrib.ActualContributedAmount,
				Returned1Amount:    matchedNReturnedContrib.ReturnedContributedAmount,
			}
			contribStatusBytes, _ := json.Marshal(contribStatus)
			err := statedb.TrackPDEContributionStatus(
				stateDB,
				rawdbv2.PDEContributionStatusPrefix,
				[]byte(matchedNReturnedContrib.PDEContributionPairID),
				contribStatusBytes,
			)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while tracking pde contribution status: %+v", err)
				return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
			}
		} else {
			var contribStatus metadata.PDEContributionStatus
			err := json.Unmarshal(pdeStatusContentBytes, &contribStatus)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while unmarshaling pde contribution status: %+v", err)
				return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
			}
			contribStatus.TokenID2Str = matchedNReturnedContrib.TokenIDStr
			contribStatus.Contributed2Amount = matchedNReturnedContrib.ActualContributedAmount
			contribStatus.Returned2Amount = matchedNReturnedContrib.ReturnedContributedAmount
			contribStatusBytes, _ := json.Marshal(contribStatus)
			err = statedb.TrackPDEContributionStatus(
				stateDB,
				rawdbv2.PDEContributionStatusPrefix,
				[]byte(matchedNReturnedContrib.PDEContributionPairID),
				contribStatusBytes,
			)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while tracking pde contribution status: %+v", err)
				return waitingContributions, deletedWaitingContributions, poolPairs, shares, err
			}
		}
	}
	return waitingContributions, deletedWaitingContributions, poolPairs, shares, nil
}

func (sp *stateProcessorV1) trade(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) (map[string]*rawdbv2.PDEPoolForPair, error) {
	if len(inst) != 4 {
		err := fmt.Errorf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		return poolPairs, err
	}
	if inst[2] == common.PDETradeRefundChainStatus {
		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade instruction: %+v", err)
			return poolPairs, err
		}
		var pdeTradeReqAction metadata.PDETradeRequestAction
		err = json.Unmarshal(contentBytes, &pdeTradeReqAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade instruction: %+v", err)
			return poolPairs, err
		}
		err = statedb.TrackPDEStatus(
			stateDB,
			rawdbv2.PDETradeStatusPrefix,
			pdeTradeReqAction.TxReqID[:],
			byte(common.PDETradeRefundStatus),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde refund trade status: %+v", err)
		}
		return poolPairs, err
	}
	var tradeAcceptedContent metadata.PDETradeAcceptedContent
	err := json.Unmarshal([]byte(inst[3]), &tradeAcceptedContent)
	if err != nil {
		Logger.log.Errorf("WARNING: an error occured while unmarshaling PDETradeAcceptedContent: %+v", err)
		return poolPairs, err
	}

	poolForPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeAcceptedContent.Token1IDStr, tradeAcceptedContent.Token2IDStr))
	poolForPair, found := poolPairs[poolForPairKey]
	if !found || poolForPair == nil {
		err := fmt.Errorf("WARNING: could not find out pdePoolForPair with token ids: %s & %s", tradeAcceptedContent.Token1IDStr, tradeAcceptedContent.Token2IDStr)
		Logger.log.Error(err)
		return poolPairs, err
	}

	if tradeAcceptedContent.Token1PoolValueOperation.Operator == "+" {
		poolForPair.Token1PoolValue += tradeAcceptedContent.Token1PoolValueOperation.Value
		poolForPair.Token2PoolValue -= tradeAcceptedContent.Token2PoolValueOperation.Value
	} else {
		poolForPair.Token1PoolValue -= tradeAcceptedContent.Token1PoolValueOperation.Value
		poolForPair.Token2PoolValue += tradeAcceptedContent.Token2PoolValueOperation.Value
	}
	err = statedb.TrackPDEStatus(
		stateDB,
		rawdbv2.PDETradeStatusPrefix,
		tradeAcceptedContent.RequestedTxID[:],
		byte(common.PDETradeAcceptedStatus),
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde accepted trade status: %+v", err)
	}
	return poolPairs, err
}

func (sp *stateProcessorV1) crossPoolTrade(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) (map[string]*rawdbv2.PDEPoolForPair, error) {
	if len(inst) != 4 {
		err := fmt.Errorf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		return poolPairs, err
	}
	if inst[2] == common.PDECrossPoolTradeFeeRefundChainStatus ||
		inst[2] == common.PDECrossPoolTradeSellingTokenRefundChainStatus {
		contentBytes := []byte(inst[3])
		var refundCrossPoolTrade metadata.PDERefundCrossPoolTrade
		err := json.Unmarshal(contentBytes, &refundCrossPoolTrade)
		if err != nil {
			err := fmt.Errorf("ERROR: an error occured while unmarshaling pde refund cross pool trade instruction: %+v", err)
			Logger.log.Error(err)
			return poolPairs, err
		}
		err = statedb.TrackPDEStatus(
			stateDB,
			rawdbv2.PDETradeStatusPrefix,
			refundCrossPoolTrade.TxReqID[:],
			byte(common.PDECrossPoolTradeRefundStatus),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde refund trade status: %+v", err)
		}
		return poolPairs, err
	}
	// trade accepted
	var tradeAcceptedContents []metadata.PDECrossPoolTradeAcceptedContent
	err := json.Unmarshal([]byte(inst[3]), &tradeAcceptedContents)
	if err != nil {
		err := fmt.Errorf("WARNING: an error occured while unmarshaling PDETradeAcceptedContents: %+v", err)
		Logger.log.Error(err)
		return poolPairs, err
	}

	if len(tradeAcceptedContents) == 0 {
		Logger.log.Error("WARNING: There is no pde cross pool trade accepted content.")
		return poolPairs, nil
	}

	for _, tradeAcceptedContent := range tradeAcceptedContents {
		poolForPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeAcceptedContent.Token1IDStr, tradeAcceptedContent.Token2IDStr))
		poolForPair, found := poolPairs[poolForPairKey]
		if !found || poolForPair == nil {
			err := fmt.Errorf("WARNING: could not find out pdePoolForPair with token ids: %s & %s", tradeAcceptedContent.Token1IDStr, tradeAcceptedContent.Token2IDStr)
			Logger.log.Error(err)
			return poolPairs, err
		}

		if tradeAcceptedContent.Token1PoolValueOperation.Operator == "+" {
			poolForPair.Token1PoolValue += tradeAcceptedContent.Token1PoolValueOperation.Value
			poolForPair.Token2PoolValue -= tradeAcceptedContent.Token2PoolValueOperation.Value
		} else {
			poolForPair.Token1PoolValue -= tradeAcceptedContent.Token1PoolValueOperation.Value
			poolForPair.Token2PoolValue += tradeAcceptedContent.Token2PoolValueOperation.Value
		}
	}
	err = statedb.TrackPDEStatus(
		stateDB,
		rawdbv2.PDETradeStatusPrefix,
		tradeAcceptedContents[0].RequestedTxID[:],
		byte(common.PDECrossPoolTradeAcceptedStatus),
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde accepted trade status: %+v", err)
	}
	return poolPairs, err
}

func (sp *stateProcessorV1) withdrawal(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	shares map[string]uint64,
) (
	map[string]*rawdbv2.PDEPoolForPair,
	map[string]uint64,
	error,
) {
	if len(inst) != 4 {
		err := fmt.Errorf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		return poolPairs, shares, err
	}
	if inst[2] == common.PDEWithdrawalRejectedChainStatus {
		contentBytes, err := base64.StdEncoding.DecodeString(inst[3])
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
			return poolPairs, shares, err
		}
		var withdrawalRequestAction metadata.PDEWithdrawalRequestAction
		err = json.Unmarshal(contentBytes, &withdrawalRequestAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde withdrawal request action: %+v", err)
			return poolPairs, shares, err
		}
		err = statedb.TrackPDEStatus(
			stateDB,
			rawdbv2.PDEWithdrawalStatusPrefix,
			withdrawalRequestAction.TxReqID[:],
			byte(common.PDEWithdrawalRejectedStatus),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde rejected withdrawal status: %+v", err)
		}
		return poolPairs, shares, err
	}

	var wdAcceptedContent metadata.PDEWithdrawalAcceptedContent
	err := json.Unmarshal([]byte(inst[3]), &wdAcceptedContent)
	if err != nil {
		Logger.log.Errorf("WARNING: an error occured while unmarshaling PDEWithdrawalAcceptedContent: %+v", err)
		return poolPairs, shares, err
	}

	// update pde pool pair
	poolForPairKey := string(rawdbv2.BuildPDEPoolForPairKey(
		beaconHeight,
		wdAcceptedContent.PairToken1IDStr,
		wdAcceptedContent.PairToken2IDStr,
	))
	poolForPair, found := poolPairs[poolForPairKey]
	if !found || poolForPair == nil {
		Logger.log.Errorf("WARNING: could not find out pdePoolForPair with token ids: %s & %s", wdAcceptedContent.PairToken1IDStr, wdAcceptedContent.PairToken2IDStr)
		return poolPairs, shares, err
	}
	if poolForPair.Token1IDStr == wdAcceptedContent.WithdrawalTokenIDStr {
		poolForPair.Token1PoolValue -= wdAcceptedContent.DeductingPoolValue
	} else {
		poolForPair.Token2PoolValue -= wdAcceptedContent.DeductingPoolValue
	}

	// update pde shares
	shares = sp.deductSharesForWithdrawal(
		beaconHeight,
		wdAcceptedContent.PairToken1IDStr,
		wdAcceptedContent.PairToken2IDStr,
		wdAcceptedContent.WithdrawerAddressStr,
		wdAcceptedContent.DeductingShares,
		shares,
	)

	err = statedb.TrackPDEStatus(
		stateDB,
		rawdbv2.PDEWithdrawalStatusPrefix,
		wdAcceptedContent.TxReqID[:],
		byte(common.PDEWithdrawalAcceptedStatus),
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde accepted withdrawal status: %+v", err)
	}
	return poolPairs, shares, err
}

func (sp *stateProcessorV1) feeWithdrawal(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	tradingFees map[string]uint64,
) (map[string]uint64, error) {
	if len(inst) != 4 {
		err := fmt.Errorf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		return tradingFees, err
	}
	contentStr := inst[3]
	contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of pde fee withdrawal action: %+v", err)
		return tradingFees, err
	}
	var feeWithdrawalRequestAction metadata.PDEFeeWithdrawalRequestAction
	err = json.Unmarshal(contentBytes, &feeWithdrawalRequestAction)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshaling pde fee withdrawal request action: %+v", err)
		return tradingFees, err
	}

	if inst[2] == common.PDEFeeWithdrawalRejectedChainStatus {
		err = statedb.TrackPDEStatus(
			stateDB,
			rawdbv2.PDEFeeWithdrawalStatusPrefix,
			feeWithdrawalRequestAction.TxReqID[:],
			byte(common.PDEFeeWithdrawalRejectedStatus),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde rejected withdrawal status: %+v", err)
		}
		return tradingFees, err
	}
	// fee withdrawal accepted
	wdMeta := feeWithdrawalRequestAction.Meta
	tradingFeeKeyBytes, err := rawdbv2.BuildPDETradingFeeKey(
		beaconHeight,
		wdMeta.WithdrawalToken1IDStr,
		wdMeta.WithdrawalToken2IDStr,
		wdMeta.WithdrawerAddressStr,
	)
	if err != nil {
		Logger.log.Errorf("cannot build PDETradingFeeKey for address: %v. Error: %v\n", wdMeta.WithdrawerAddressStr, err)
		return tradingFees, err
	}

	tradingFeeKey := string(tradingFeeKeyBytes)
	withdrawableFee, found := tradingFees[tradingFeeKey]
	if !found || withdrawableFee < wdMeta.WithdrawalFeeAmt {
		Logger.log.Warnf("WARN: Could not withdraw trading fee due to insufficient amount or not existed trading fee key (%s)", tradingFeeKey)
		err = statedb.TrackPDEStatus(
			stateDB,
			rawdbv2.PDEFeeWithdrawalStatusPrefix,
			feeWithdrawalRequestAction.TxReqID[:],
			byte(common.PDEFeeWithdrawalRejectedStatus),
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking pde rejected withdrawal status: %+v", err)
		}
		return tradingFees, err
	}
	tradingFees[tradingFeeKey] -= wdMeta.WithdrawalFeeAmt
	err = statedb.TrackPDEStatus(
		stateDB,
		rawdbv2.PDEFeeWithdrawalStatusPrefix,
		feeWithdrawalRequestAction.TxReqID[:],
		byte(common.PDEFeeWithdrawalAcceptedStatus),
	)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while tracking pde rejected withdrawal status: %+v", err)
	}
	return tradingFees, err
}

func (sp *stateProcessorV1) tradingFeesDistribution(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	inst []string,
	tradingFees map[string]uint64,
) (map[string]uint64, error) {
	if len(inst) != 4 {
		err := fmt.Errorf("Length of instruction is not valid expect %v but get %v", 4, len(inst))
		return tradingFees, err
	}
	var feesForContributorsByPair []*tradingFeeForContributorByPair
	err := json.Unmarshal([]byte(inst[3]), &feesForContributorsByPair)
	if err != nil {
		err := fmt.Errorf("ERROR: an error occured while unmarshaling trading fees for contribution by pair: %+v", err)
		Logger.log.Error(err)
		return tradingFees, err
	}

	for _, item := range feesForContributorsByPair {
		tradingFeeKeyBytes, err := rawdbv2.BuildPDETradingFeeKey(
			beaconHeight,
			item.Token1IDStr,
			item.Token2IDStr,
			item.ContributorAddressStr,
		)
		if err != nil {
			Logger.log.Errorf("cannot build PDETradingFeeKey for address: %v. Error: %v\n", item.ContributorAddressStr, err)
			return tradingFees, err
		}

		tradingFeeKey := string(tradingFeeKeyBytes)
		tradingFees[tradingFeeKey] += item.FeeAmt
	}
	return tradingFees, nil
}

func (sp *stateProcessorV1) deductSharesForWithdrawal(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
	withdrawerAddressStr string,
	amt uint64,
	shares map[string]uint64,
) map[string]uint64 {
	pdeShareKeyBytes, err := rawdbv2.BuildPDESharesKeyV2(beaconHeight, token1IDStr, token2IDStr, withdrawerAddressStr)
	if err != nil {
		Logger.log.Errorf("cannot find pdeShareKey for address: %v. Error: %v\n", withdrawerAddressStr, err)
		return shares
	}
	pdeShareKey := string(pdeShareKeyBytes)
	adjustingAmt := uint64(0)
	currentAmt, found := shares[pdeShareKey]
	if found && amt <= currentAmt {
		adjustingAmt = currentAmt - amt
	}
	shares[pdeShareKey] = adjustingAmt
	return shares
}
