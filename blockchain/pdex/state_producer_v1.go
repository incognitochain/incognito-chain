package pdex

import (
	"encoding/base64"
	"encoding/json"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateProducerV1 struct {
	stateProducerBase
}

func (sp *stateProducerV1) crossPoolTrade(
	actions [][]string,
	beaconHeight uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	shares map[string]uint64,
) ([][]string, map[string]*rawdbv2.PDEPoolForPair, map[string]uint64, error) {
	res := [][]string{}

	// handle cross pool trade
	sortedTradableActions, untradableActions := sp.categorizeAndSortCrossPoolTradeInstsByFee(
		actions,
		beaconHeight,
		poolPairs,
	)
	tradableInsts, tradingFeeByPair := sp.buildInstsForSortedTradableActions(sortedTradableActions, beaconHeight, poolPairs)
	untradableInsts := buildInstsForUntradableActions(untradableActions)
	res = append(res, tradableInsts...)
	res = append(res, untradableInsts...)

	// calculate and build instruction for trading fees distribution
	tradingFeesDistInst := sp.buildInstForTradingFeesDist(beaconHeight, tradingFeeByPair, shares)
	if len(tradingFeesDistInst) > 0 {
		res = append(res, tradingFeesDistInst)
	}

	return res, poolPairs, shares, nil
}

func (sp *stateProducerV1) buildInstForTradingFeesDist(
	beaconHeight uint64,
	tradingFeeByPair map[string]uint64,
	shares map[string]uint64,
) []string {

	feesForContributorsByPair := []*tradingFeeForContributorByPair{}

	var keys []string
	for k := range tradingFeeByPair {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, sKey := range keys {
		feeAmt := tradingFeeByPair[sKey]
		allSharesByPair := []shareInfo{}
		totalSharesOfPair := big.NewInt(0)

		var shareKeys []string
		for shareKey := range shares {
			shareKeys = append(shareKeys, shareKey)
		}
		sort.Strings(shareKeys)
		for _, shareKey := range shareKeys {
			shareAmt := shares[shareKey]
			if strings.Contains(shareKey, sKey) {
				allSharesByPair = append(allSharesByPair, shareInfo{shareKey: shareKey, shareAmt: shareAmt})
				totalSharesOfPair.Add(totalSharesOfPair, new(big.Int).SetUint64(shareAmt))
			}
		}

		accumFees := big.NewInt(0)
		totalFees := new(big.Int).SetUint64(feeAmt)
		for idx, sInfo := range allSharesByPair {
			feeForContributor := big.NewInt(0)
			if idx == len(allSharesByPair)-1 {
				feeForContributor.Sub(totalFees, accumFees)
			} else {
				if totalSharesOfPair.Cmp(big.NewInt(0)) == 1 {
					feeForContributor.Mul(totalFees, new(big.Int).SetUint64(sInfo.shareAmt))
					feeForContributor.Div(feeForContributor, totalSharesOfPair)
				}
			}

			parts := strings.Split(sInfo.shareKey, "-")
			partsLen := len(parts)
			if partsLen < 5 {
				continue
			}
			feesForContributorsByPair = append(
				feesForContributorsByPair,
				&tradingFeeForContributorByPair{
					ContributorAddressStr: parts[partsLen-1],
					FeeAmt:                feeForContributor.Uint64(),
					Token2IDStr:           parts[partsLen-2],
					Token1IDStr:           parts[partsLen-3],
				},
			)
			accumFees.Add(accumFees, feeForContributor)
		}
	}
	if len(feesForContributorsByPair) == 0 {
		return []string{}
	}

	feesForContributorsByPairBytes, err := json.Marshal(feesForContributorsByPair)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while marshaling feesForContributorsByPair: %+v", err)
		return []string{}
	}
	return []string{
		strconv.Itoa(metadata.PDETradingFeesDistributionMeta),
		"",
		"",
		string(feesForContributorsByPairBytes),
	}
}

func (sp *stateProducerV1) buildAcceptedCrossPoolTradeInstruction(
	sequentialTrades []*tradeInfo,
	action metadata.PDECrossPoolTradeRequestAction,
	beaconHeight uint64,
	tradingFeeByPair map[string]uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) ([]string, error) {
	tradeAcceptedContents := []metadata.PDECrossPoolTradeAcceptedContent{}
	proportionalFee := action.Meta.TradingFee / uint64(len(sequentialTrades))
	for idx, tradeInf := range sequentialTrades {
		// update current pde state on mem
		pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeInf.tokenIDToBuyStr, tradeInf.tokenIDToSellStr))
		poolPair, _ := poolPairs[pairKey]

		poolPair.Token1PoolValue = tradeInf.newTokenPoolValueToBuy
		poolPair.Token2PoolValue = tradeInf.newTokenPoolValueToSell
		if poolPair.Token1IDStr == tradeInf.tokenIDToSellStr {
			poolPair.Token1PoolValue = tradeInf.newTokenPoolValueToSell
			poolPair.Token2PoolValue = tradeInf.newTokenPoolValueToBuy
		}

		// build trade accepted contents
		tradeAcceptedContent := metadata.PDECrossPoolTradeAcceptedContent{
			TraderAddressStr: action.Meta.TraderAddressStr,
			TxRandomStr:      action.Meta.TxRandomStr,
			TokenIDToBuyStr:  tradeInf.tokenIDToBuyStr,
			ReceiveAmount:    tradeInf.receiveAmount,
			Token1IDStr:      poolPair.Token1IDStr,
			Token2IDStr:      poolPair.Token2IDStr,
			ShardID:          action.ShardID,
			RequestedTxID:    action.TxReqID,
		}
		tradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
			Operator: "-",
			Value:    tradeInf.receiveAmount,
		}
		tradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
			Operator: "+",
			Value:    tradeInf.sellAmount,
		}
		if poolPair.Token1IDStr == tradeInf.tokenIDToSellStr {
			tradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
				Operator: "+",
				Value:    tradeInf.sellAmount,
			}
			tradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
				Operator: "-",
				Value:    tradeInf.receiveAmount,
			}
		}

		addingFee := proportionalFee
		if idx == len(sequentialTrades)-1 {
			addingFee = action.Meta.TradingFee - uint64(len(sequentialTrades)-1)*proportionalFee
		}
		tradeAcceptedContent.AddingFee = addingFee
		sKeyBytes, err := rawdbv2.BuildPDESharesKeyV2(beaconHeight, tradeInf.tokenIDToBuyStr, tradeInf.tokenIDToSellStr, "")
		if err != nil {
			Logger.log.Errorf("cannot build PDESharesKeyV2. Error: %v\n", err)
			return []string{}, err
		}

		sKey := string(sKeyBytes)
		tradingFeeByPair[sKey] += addingFee
		tradeAcceptedContents = append(tradeAcceptedContents, tradeAcceptedContent)
	}

	tradeAcceptedContentsBytes, err := json.Marshal(tradeAcceptedContents)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while marshaling pdeTradeAcceptedContents: %+v", err)
		return []string{}, nil
	}
	inst := []string{
		strconv.Itoa(metadata.PDECrossPoolTradeRequestMeta),
		strconv.Itoa(int(action.ShardID)),
		common.PDECrossPoolTradeAcceptedChainStatus,
		string(tradeAcceptedContentsBytes),
	}
	return inst, nil
}

func (sp *stateProducerV1) buildInstructionsCrossPoolTrade(
	sequentialTrades []*tradeInfo,
	action metadata.PDECrossPoolTradeRequestAction,
	beaconHeight uint64,
	tradingFeeByPair map[string]uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) ([][]string, error) {
	res := [][]string{}

	should, err := sp.shouldRefundCrossPoolTrade(sequentialTrades, action, beaconHeight, poolPairs)
	if err != nil {
		return res, err
	}
	if should {
		return refundForCrossPoolTrade(sequentialTrades, action)
	}

	inst, err := sp.buildAcceptedCrossPoolTradeInstruction(
		sequentialTrades,
		action,
		beaconHeight,
		tradingFeeByPair,
		poolPairs,
	)
	if err != nil {
		return res, err
	}
	res = append(res, inst)
	return res, nil
}

func (sp *stateProducerV1) shouldRefundCrossPoolTrade(
	sequentialTrades []*tradeInfo,
	action metadata.PDECrossPoolTradeRequestAction,
	beaconHeight uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) (bool, error) {
	if poolPairs == nil || len(poolPairs) == 0 {
		return true, nil
	}

	amt := sequentialTrades[0].sellAmount
	for _, tradeInf := range sequentialTrades {
		tradeInf.sellAmount = amt
		pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeInf.tokenIDToBuyStr, tradeInf.tokenIDToSellStr))
		poolPair, _ := poolPairs[pairKey]
		newAmt, newTokenPoolValueToBuy, newTokenPoolValueToSell, err := calcTradeValue(poolPair, tradeInf.tokenIDToSellStr, amt)
		if err != nil {
			return true, err
		}
		amt = newAmt
		tradeInf.newTokenPoolValueToBuy = newTokenPoolValueToBuy
		tradeInf.newTokenPoolValueToSell = newTokenPoolValueToSell
		tradeInf.receiveAmount = amt
	}

	if action.Meta.MinAcceptableAmount > amt {
		return true, nil
	}
	return false, nil
}

func (sp *stateProducerV1) buildInstsForSortedTradableActions(
	actions []metadata.PDECrossPoolTradeRequestAction,
	beaconHeight uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) ([][]string, map[string]uint64) {
	prvIDStr := common.PRVCoinID.String()
	tradableInsts := [][]string{}
	tradingFeeByPair := make(map[string]uint64)
	for _, action := range actions {
		tradeMeta := action.Meta
		var sequentialTrades []*tradeInfo
		if isTradingPairContainsPRV(tradeMeta.TokenIDToSellStr, tradeMeta.TokenIDToBuyStr) { // direct trade
			sequentialTrades = []*tradeInfo{
				&tradeInfo{
					tokenIDToBuyStr:  tradeMeta.TokenIDToBuyStr,
					tokenIDToSellStr: tradeMeta.TokenIDToSellStr,
					sellAmount:       tradeMeta.SellAmount,
				},
			}
		} else { // cross pool trade
			sequentialTrades = []*tradeInfo{
				&tradeInfo{
					tokenIDToBuyStr:  prvIDStr,
					tokenIDToSellStr: tradeMeta.TokenIDToSellStr,
					sellAmount:       tradeMeta.SellAmount,
				},
				&tradeInfo{
					tokenIDToBuyStr:  tradeMeta.TokenIDToBuyStr,
					tokenIDToSellStr: prvIDStr,
					sellAmount:       uint64(0),
				},
			}
		}

		//Work-around solution for the sake of syncing mainnet
		//Todo: find better solution
		if _, err := metadata.AssertPaymentAddressAndTxVersion(action.Meta.TraderAddressStr, 1); err == nil {
			if len(action.Meta.SubTraderAddressStr) == 0 {
				action.Meta.SubTraderAddressStr = action.Meta.TraderAddressStr
			}
		}

		newInsts, err := sp.buildInstructionsCrossPoolTrade(
			sequentialTrades,
			action,
			beaconHeight,
			tradingFeeByPair,
			poolPairs,
		)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		tradableInsts = append(tradableInsts, newInsts...)
	}
	return tradableInsts, tradingFeeByPair
}

func (sp *stateProducerV1) categorizeAndSortCrossPoolTradeInstsByFee(
	actions [][]string,
	beaconHeight uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) (
	[]metadata.PDECrossPoolTradeRequestAction,
	[]metadata.PDECrossPoolTradeRequestAction,
) {
	prvIDStr := common.PRVCoinID.String()
	tradableActions := []metadata.PDECrossPoolTradeRequestAction{}
	untradableActions := []metadata.PDECrossPoolTradeRequestAction{}

	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade action: %+v", err)
			continue
		}
		var crossPoolTradeRequestAction metadata.PDECrossPoolTradeRequestAction
		err = json.Unmarshal(contentBytes, &crossPoolTradeRequestAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde cross pool trade request action: %+v", err)
			continue
		}
		tradeMeta := crossPoolTradeRequestAction.Meta
		tokenIDToSell := tradeMeta.TokenIDToSellStr
		tokenIDToBuy := tradeMeta.TokenIDToBuyStr
		if (isTradingPairContainsPRV(tokenIDToSell, tokenIDToBuy) &&
			!isExistedInPoolPair(poolPairs, beaconHeight, tokenIDToSell, tokenIDToBuy)) ||
			(!isTradingPairContainsPRV(tokenIDToSell, tokenIDToBuy) &&
				(!isExistedInPoolPair(poolPairs, beaconHeight, prvIDStr, tokenIDToSell) ||
					!isExistedInPoolPair(poolPairs, beaconHeight, prvIDStr, tokenIDToBuy))) {
			untradableActions = append(untradableActions, crossPoolTradeRequestAction)
			continue
		}
		tradableActions = append(tradableActions, crossPoolTradeRequestAction)
	}
	// sort tradable actions by trading fee
	sort.SliceStable(tradableActions, func(i, j int) bool {
		firstTradingFee, firstSellAmount := sp.prepareInfoForSorting(
			beaconHeight,
			tradableActions[i],
			poolPairs,
		)
		secondTradingFee, secondSellAmount := sp.prepareInfoForSorting(
			beaconHeight,
			tradableActions[j],
			poolPairs,
		)
		// comparing a/b to c/d is equivalent with comparing a*d to c*b
		firstItemProportion := big.NewInt(0)
		firstItemProportion.Mul(
			new(big.Int).SetUint64(firstTradingFee),
			new(big.Int).SetUint64(secondSellAmount),
		)
		secondItemProportion := big.NewInt(0)
		secondItemProportion.Mul(
			new(big.Int).SetUint64(secondTradingFee),
			new(big.Int).SetUint64(firstSellAmount),
		)
		return firstItemProportion.Cmp(secondItemProportion) == 1
	})
	return tradableActions, untradableActions
}

func (sp *stateProducerV1) prepareInfoForSorting(
	beaconHeight uint64,
	action metadata.PDECrossPoolTradeRequestAction,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) (uint64, uint64) {
	prvIDStr := common.PRVCoinID.String()
	tradeMeta := action.Meta
	sellAmount := tradeMeta.SellAmount
	tradingFee := tradeMeta.TradingFee
	if tradeMeta.TokenIDToSellStr == prvIDStr {
		return tradingFee, sellAmount
	}
	poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, prvIDStr, tradeMeta.TokenIDToSellStr))
	poolPair, _ := poolPairs[poolPairKey]
	sellAmount, _, _, _ = calcTradeValue(poolPair, tradeMeta.TokenIDToSellStr, sellAmount)
	return tradingFee, sellAmount
}

func (sp *stateProducerV1) withdrawal(
	actions [][]string,
	beaconHeight uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	shares map[string]uint64,
) ([][]string, map[string]*rawdbv2.PDEPoolForPair, map[string]uint64, error) {
	res := [][]string{}

	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
			return utils.EmptyStringMatrix, poolPairs, shares, err
		}
		var withdrawalRequestAction metadata.PDEWithdrawalRequestAction
		err = json.Unmarshal(contentBytes, &withdrawalRequestAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde withdrawal request action: %+v", err)
			return [][]string{}, poolPairs, shares, err
		}
		wdMeta := withdrawalRequestAction.Meta
		deductingAmounts := sp.deductAmounts(
			wdMeta,
			beaconHeight,
			poolPairs,
			shares,
		)

		if deductingAmounts == nil {
			inst := []string{
				strconv.Itoa(metadata.PDEWithdrawalRequestMeta),
				strconv.Itoa(int(withdrawalRequestAction.ShardID)),
				common.PDEWithdrawalRejectedChainStatus,
				contentStr,
			}
			return [][]string{inst}, poolPairs, shares, nil
		}

		inst, err := buildWithdrawalAcceptedInst(
			withdrawalRequestAction,
			deductingAmounts.Token1IDStr,
			deductingAmounts.PoolValue1,
			deductingAmounts.Shares,
		)
		if err != nil {
			return [][]string{}, poolPairs, shares, nil
		}
		res = append(res, inst)
		inst, err = buildWithdrawalAcceptedInst(
			withdrawalRequestAction,
			deductingAmounts.Token2IDStr,
			deductingAmounts.PoolValue2,
			0,
		)
		if err != nil {
			return [][]string{}, poolPairs, shares, nil
		}
		res = append(res, inst)
	}

	return res, poolPairs, shares, nil
}

func (sp *stateProducerV1) deductAmounts(
	wdMeta metadata.PDEWithdrawalRequest,
	beaconHeight uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	shares map[string]uint64,
) *deductingAmountsByWithdrawal {
	var res *deductingAmountsByWithdrawal
	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(
		beaconHeight,
		wdMeta.WithdrawalToken1IDStr, wdMeta.WithdrawalToken2IDStr,
	))
	poolPair, found := poolPairs[pairKey]
	if !found || poolPair == nil {
		return res
	}
	shareForWithdrawerKeyBytes, err := rawdbv2.BuildPDESharesKeyV2(
		beaconHeight,
		wdMeta.WithdrawalToken1IDStr,
		wdMeta.WithdrawalToken2IDStr,
		wdMeta.WithdrawerAddressStr,
	)
	if err != nil {
		Logger.log.Errorf("cannot build PDESharesKeyV2 for address: %v. Error: %v\n", wdMeta.WithdrawerAddressStr, err)
		return res
	}

	shareForWithdrawerKey := string(shareForWithdrawerKeyBytes)
	currentSharesForWithdrawer, found := shares[shareForWithdrawerKey]
	if !found || currentSharesForWithdrawer == 0 {
		return res
	}

	totalSharesForPairPrefixBytes, err := rawdbv2.BuildPDESharesKeyV2(
		beaconHeight,
		wdMeta.WithdrawalToken1IDStr, wdMeta.WithdrawalToken2IDStr, "",
	)
	if err != nil {
		Logger.log.Errorf("cannot build PDESharesKeyV2. Error: %v\n", err)
		return res
	}

	totalSharesForPairPrefix := string(totalSharesForPairPrefixBytes)
	totalSharesForPair := big.NewInt(0)

	for shareKey, shareAmt := range shares {
		if strings.Contains(shareKey, totalSharesForPairPrefix) {
			totalSharesForPair.Add(totalSharesForPair, new(big.Int).SetUint64(shareAmt))
		}
	}
	if totalSharesForPair.Cmp(big.NewInt(0)) == 0 {
		return res
	}
	wdSharesForWithdrawer := wdMeta.WithdrawalShareAmt
	if wdSharesForWithdrawer > currentSharesForWithdrawer {
		wdSharesForWithdrawer = currentSharesForWithdrawer
	}
	if wdSharesForWithdrawer == 0 {
		return res
	}

	res = &deductingAmountsByWithdrawal{}
	deductingPoolValueToken1 := big.NewInt(0)
	deductingPoolValueToken1.Mul(new(big.Int).SetUint64(poolPair.Token1PoolValue), new(big.Int).SetUint64(wdSharesForWithdrawer))
	deductingPoolValueToken1.Div(deductingPoolValueToken1, totalSharesForPair)
	if poolPair.Token1PoolValue < deductingPoolValueToken1.Uint64() {
		poolPair.Token1PoolValue = 0
	} else {
		poolPair.Token1PoolValue -= deductingPoolValueToken1.Uint64()
	}
	res.Token1IDStr = poolPair.Token1IDStr
	res.PoolValue1 = deductingPoolValueToken1.Uint64()

	deductingPoolValueToken2 := big.NewInt(0)
	deductingPoolValueToken2.Mul(new(big.Int).SetUint64(poolPair.Token2PoolValue), new(big.Int).SetUint64(wdSharesForWithdrawer))
	deductingPoolValueToken2.Div(deductingPoolValueToken2, totalSharesForPair)
	if poolPair.Token2PoolValue < deductingPoolValueToken2.Uint64() {
		poolPair.Token2PoolValue = 0
	} else {
		poolPair.Token2PoolValue -= deductingPoolValueToken2.Uint64()
	}
	res.Token2IDStr = poolPair.Token2IDStr
	res.PoolValue2 = deductingPoolValueToken2.Uint64()

	if shares[shareForWithdrawerKey] < wdSharesForWithdrawer {
		shares[shareForWithdrawerKey] = 0
	} else {
		shares[shareForWithdrawerKey] -= wdSharesForWithdrawer
	}
	res.Shares = wdSharesForWithdrawer

	return res
}

func (sp *stateProducerV1) contribution(
	actions [][]string,
	beaconHeight uint64,
	isPRVRequired bool,
	metaType int,
	waitingContributions map[string]*rawdbv2.PDEContribution,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
	shares map[string]uint64,
) (
	[][]string,
	map[string]*rawdbv2.PDEContribution,
	map[string]*rawdbv2.PDEPoolForPair,
	map[string]uint64,
	error,
) {
	res := [][]string{}

	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
			return [][]string{}, waitingContributions, poolPairs, shares, err
		}
		var contributionAction metadata.PDEContributionAction
		err = json.Unmarshal(contentBytes, &contributionAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde contribution action: %+v", err)
			return [][]string{}, waitingContributions, poolPairs, shares, err
		}
		meta := contributionAction.Meta
		waitingContribPairKey := string(rawdbv2.BuildWaitingPDEContributionKey(beaconHeight, meta.PDEContributionPairID))
		waitingContribution, found := waitingContributions[waitingContribPairKey]
		if !found || waitingContribution == nil {
			waitingContributions[waitingContribPairKey] = &rawdbv2.PDEContribution{
				ContributorAddressStr: meta.ContributorAddressStr,
				TokenIDStr:            meta.TokenIDStr,
				Amount:                meta.ContributedAmount,
				TxReqID:               contributionAction.TxReqID,
			}
			inst := buildWaitingContributionInst(
				contributionAction,
				metaType,
			)
			res = append(res, inst)
			continue
		}

		if waitingContribution.TokenIDStr == meta.TokenIDStr ||
			waitingContribution.ContributorAddressStr != meta.ContributorAddressStr ||
			(isPRVRequired && waitingContribution.TokenIDStr != common.PRVIDStr && meta.TokenIDStr != common.PRVIDStr) {
			delete(waitingContributions, waitingContribPairKey)

			refundInst1 := buildRefundContributionInst(
				meta.PDEContributionPairID,
				meta.ContributorAddressStr,
				meta.ContributedAmount,
				meta.TokenIDStr,
				metaType,
				contributionAction.ShardID,
				contributionAction.TxReqID,
			)
			refundInst2 := buildRefundContributionInst(
				meta.PDEContributionPairID,
				waitingContribution.ContributorAddressStr,
				waitingContribution.Amount,
				waitingContribution.TokenIDStr,
				metaType,
				contributionAction.ShardID,
				waitingContribution.TxReqID,
			)

			res = append(res, refundInst1)
			res = append(res, refundInst2)
			continue
		}
		// contributed to 2 diff sides of a pair and its a first contribution of this pair
		poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, waitingContribution.TokenIDStr, meta.TokenIDStr))
		poolPair, found := poolPairs[poolPairKey]
		incomingWaitingContribution := &rawdbv2.PDEContribution{
			ContributorAddressStr: meta.ContributorAddressStr,
			TokenIDStr:            meta.TokenIDStr,
			Amount:                meta.ContributedAmount,
			TxReqID:               contributionAction.TxReqID,
		}

		if !found || poolPair == nil {
			delete(waitingContributions, waitingContribPairKey)
			err := updateWaitingContributionPairToPool(
				beaconHeight,
				waitingContribution,
				incomingWaitingContribution,
				poolPairs,
				shares,
			)
			if err != nil {
				return utils.EmptyStringMatrix, waitingContributions, poolPairs, shares, err
			}
			matchedInst := buildMatchedContributionInst(
				contributionAction,
				metaType,
			)
			res = append(res, matchedInst)
			continue
		}

		//isRightRatio(waitingContribution, incomingWaitingContribution, poolPair)
		actualWaitingContribAmt, returnedWaitingContribAmt, actualIncomingWaitingContribAmt, returnedIncomingWaitingContribAmt := computeActualContributedAmounts(
			waitingContribution,
			incomingWaitingContribution,
			poolPair,
		)

		if actualWaitingContribAmt == 0 || actualIncomingWaitingContribAmt == 0 {
			delete(waitingContributions, waitingContribPairKey)
			refundInst1 := buildRefundContributionInst(
				meta.PDEContributionPairID,
				meta.ContributorAddressStr,
				meta.ContributedAmount,
				meta.TokenIDStr,
				metaType,
				contributionAction.ShardID,
				contributionAction.TxReqID,
			)
			refundInst2 := buildRefundContributionInst(
				meta.PDEContributionPairID,
				waitingContribution.ContributorAddressStr,
				waitingContribution.Amount,
				waitingContribution.TokenIDStr,
				metaType,
				contributionAction.ShardID,
				waitingContribution.TxReqID,
			)
			res = append(res, refundInst1)
			res = append(res, refundInst2)
			continue
		}

		delete(waitingContributions, waitingContribPairKey)
		actualWaitingContrib := &rawdbv2.PDEContribution{
			ContributorAddressStr: waitingContribution.ContributorAddressStr,
			TokenIDStr:            waitingContribution.TokenIDStr,
			Amount:                actualWaitingContribAmt,
			TxReqID:               waitingContribution.TxReqID,
		}
		actualIncomingWaitingContrib := &rawdbv2.PDEContribution{
			ContributorAddressStr: meta.ContributorAddressStr,
			TokenIDStr:            meta.TokenIDStr,
			Amount:                actualIncomingWaitingContribAmt,
			TxReqID:               contributionAction.TxReqID,
		}
		updateWaitingContributionPairToPool(
			beaconHeight,
			actualWaitingContrib,
			actualIncomingWaitingContrib,
			poolPairs,
			shares,
		)

		matchedAndReturnedInst1 := buildMatchedAndReturnedContributionInst(
			meta.PDEContributionPairID,
			meta.ContributorAddressStr,
			actualIncomingWaitingContribAmt,
			returnedIncomingWaitingContribAmt,
			meta.TokenIDStr,
			metaType,
			contributionAction.ShardID,
			contributionAction.TxReqID,
			actualWaitingContribAmt,
		)
		matchedAndReturnedInst2 := buildMatchedAndReturnedContributionInst(
			meta.PDEContributionPairID,
			waitingContribution.ContributorAddressStr,
			actualWaitingContribAmt,
			returnedWaitingContribAmt,
			waitingContribution.TokenIDStr,
			metaType,
			contributionAction.ShardID,
			waitingContribution.TxReqID,
			0,
		)
		res = append(res, matchedAndReturnedInst1)
		res = append(res, matchedAndReturnedInst2)
	}

	return res, waitingContributions, poolPairs, shares, nil
}

func (sp *stateProducerV1) trade(
	actions [][]string,
	beaconHeight uint64,
	poolPairs map[string]*rawdbv2.PDEPoolForPair,
) ([][]string, map[string]*rawdbv2.PDEPoolForPair, error) {
	res := [][]string{}

	// handle trade
	sortedTradesActions := sp.sortTradeInstsByFee(
		actions,
		beaconHeight,
		poolPairs,
	)
	for _, tradeAction := range sortedTradesActions {
		should, receiveAmount, err := shouldRefundTradeAction(tradeAction, beaconHeight, poolPairs)
		if err != nil {
			return utils.EmptyStringMatrix, poolPairs, err
		}
		if should {
			actionContentBytes, err := json.Marshal(tradeAction)
			if err != nil {
				return utils.EmptyStringMatrix, poolPairs, err
			}
			actionStr := base64.StdEncoding.EncodeToString(actionContentBytes)
			inst := []string{
				strconv.Itoa(metadata.PDETradeRequestMeta),
				strconv.Itoa(int(tradeAction.ShardID)),
				common.PDETradeRefundChainStatus,
				actionStr,
			}
			res = append(res, inst)
			continue
		}
		inst, err := buildAcceptedTradeInstruction(tradeAction, beaconHeight, receiveAmount, poolPairs)
		if err != nil {
			return utils.EmptyStringMatrix, poolPairs, err
		}
		res = append(res, inst)
	}

	return res, poolPairs, nil
}
