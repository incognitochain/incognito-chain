package pdex

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"sort"
	"strconv"
	"strings"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/utils"
)

type stateV1 struct {
	waitingContributions        map[string]*rawdbv2.PDEContribution
	deletedWaitingContributions map[string]*rawdbv2.PDEContribution
	poolPairs                   map[string]*rawdbv2.PDEPoolForPair
	shares                      map[string]uint64
	tradingFees                 map[string]uint64
}

func (s *stateV1) Version() uint {
	return BasicVersion
}

func (s *stateV1) Clone() State {
	var state State
	return state
}

func (s *stateV1) Update(env StateEnvironment) error {
	return nil
}

func (s *stateV1) BuildInstructions(env StateEnvironment) ([][]string, error) {
	instructions := [][]string{}

	// handle fee withdrawal
	tempInstructions, err := s.buildInstructionsForFeeWithdrawal(
		env.FeeWithdrawalActions(),
		env.BeaconHeight(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle trade
	tempInstructions, err = s.buildInstructionsForTrade(
		env.TradeActions(),
		env.BeaconHeight(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle cross pool trade
	tempInstructions, err = s.buildInstructionsForCrossPoolTrade(
		env.CrossPoolTradeActions(),
		env.BeaconHeight(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle withdrawal
	tempInstructions, err = s.buildInstructionsForWithdrawal(
		env.WithdrawalActions(),
		env.BeaconHeight(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle contribution
	tempInstructions, err = s.buildInstructionsForContribution(
		env.ContributionActions(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	// handle prv required contribution
	tempInstructions, err = s.buildInstructionsForPRVRequiredContribution(
		env.PRVRequiredContributionActions(),
	)
	if err != nil {
		return instructions, err
	}
	instructions = append(instructions, tempInstructions...)

	return instructions, nil
}

func (s *stateV1) buildInstructionsForCrossPoolTrade(
	actions [][]string,
	beaconHeight uint64,
) ([][]string, error) {
	res := [][]string{}

	// handle cross pool trade
	sortedTradableActions, untradableActions := s.categorizeAndSortCrossPoolTradeInstsByFee(
		actions,
		beaconHeight,
	)
	tradableInsts, tradingFeeByPair := s.buildInstsForSortedTradableActions(sortedTradableActions, beaconHeight)
	untradableInsts := buildInstsForUntradableActions(untradableActions)
	res = append(res, tradableInsts...)
	res = append(res, untradableInsts...)

	// calculate and build instruction for trading fees distribution
	tradingFeesDistInst := s.buildInstForTradingFeesDist(beaconHeight, tradingFeeByPair)
	if len(tradingFeesDistInst) > 0 {
		res = append(res, tradingFeesDistInst)
	}

	return res, nil
}

func (s *stateV1) buildInstForTradingFeesDist(
	beaconHeight uint64,
	tradingFeeByPair map[string]uint64,
) []string {

	feesForContributorsByPair := []*tradingFeeForContributorByPair{}
	shares := s.shares

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

func buildInstsForUntradableActions(
	actions []metadata.PDECrossPoolTradeRequestAction,
) [][]string {
	untradableInsts := [][]string{}
	for _, tradeAction := range actions {
		//Work-around solution for the sake of syncing mainnet
		//Todo: find better solution
		if _, err := metadata.AssertPaymentAddressAndTxVersion(tradeAction.Meta.TraderAddressStr, 1); err == nil {
			if len(tradeAction.Meta.SubTraderAddressStr) == 0 {
				tradeAction.Meta.SubTraderAddressStr = tradeAction.Meta.TraderAddressStr
			}
		}

		refundTradingFeeInst := buildCrossPoolTradeRefundInst(
			tradeAction.Meta.SubTraderAddressStr,
			tradeAction.Meta.SubTxRandomStr,
			common.PRVCoinID.String(),
			tradeAction.Meta.TradingFee,
			common.PDECrossPoolTradeFeeRefundChainStatus,
			tradeAction.ShardID,
			tradeAction.TxReqID,
		)
		if len(refundTradingFeeInst) > 0 {
			untradableInsts = append(untradableInsts, refundTradingFeeInst)
		}

		refundSellingTokenInst := buildCrossPoolTradeRefundInst(
			tradeAction.Meta.TraderAddressStr,
			tradeAction.Meta.TxRandomStr,
			tradeAction.Meta.TokenIDToSellStr,
			tradeAction.Meta.SellAmount,
			common.PDECrossPoolTradeSellingTokenRefundChainStatus,
			tradeAction.ShardID,
			tradeAction.TxReqID,
		)
		if len(refundSellingTokenInst) > 0 {
			untradableInsts = append(untradableInsts, refundSellingTokenInst)
		}
	}
	return untradableInsts
}

func (s *stateV1) buildInstsForSortedTradableActions(
	actions []metadata.PDECrossPoolTradeRequestAction,
	beaconHeight uint64,
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

		newInsts, err := s.buildInstructionsCrossPoolTrade(
			sequentialTrades,
			action,
			beaconHeight,
			tradingFeeByPair,
		)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		tradableInsts = append(tradableInsts, newInsts...)
	}
	return tradableInsts, tradingFeeByPair
}

func (s *stateV1) buildInstructionsCrossPoolTrade(
	sequentialTrades []*tradeInfo,
	action metadata.PDECrossPoolTradeRequestAction,
	beaconHeight uint64,
	tradingFeeByPair map[string]uint64,
) ([][]string, error) {
	res := [][]string{}

	should, err := s.shouldRefundCrossPoolTrade(sequentialTrades, action, beaconHeight)
	if err != nil {
		return res, err
	}
	if should {
		return refundForCrossPoolTrade(sequentialTrades, action)
	}

	inst, err := s.buildAcceptedCrossPoolTradeInstruction(
		sequentialTrades,
		action,
		beaconHeight,
		tradingFeeByPair,
	)
	if err != nil {
		return res, err
	}
	res = append(res, inst)
	return res, nil
}

func (s *stateV1) buildAcceptedCrossPoolTradeInstruction(
	sequentialTrades []*tradeInfo,
	action metadata.PDECrossPoolTradeRequestAction,
	beaconHeight uint64,
	tradingFeeByPair map[string]uint64,
) ([]string, error) {
	tradeAcceptedContents := []metadata.PDECrossPoolTradeAcceptedContent{}
	proportionalFee := action.Meta.TradingFee / uint64(len(sequentialTrades))
	for idx, tradeInf := range sequentialTrades {
		// update current pde state on mem
		pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeInf.tokenIDToBuyStr, tradeInf.tokenIDToSellStr))
		poolPair, _ := s.poolPairs[pairKey]

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

func buildCrossPoolTradeRefundInst(
	traderAddressStr string,
	txRandomStr string,
	tokenIDStr string,
	amount uint64,
	status string,
	shardID byte,
	txReqID common.Hash,
) []string {
	if amount == 0 {
		return []string{}
	}
	refundCrossPoolTrade := metadata.PDERefundCrossPoolTrade{
		TraderAddressStr: traderAddressStr,
		TxRandomStr:      txRandomStr,
		TokenIDStr:       tokenIDStr,
		Amount:           amount,
		ShardID:          shardID,
		TxReqID:          txReqID,
	}
	refundCrossPoolTradeBytes, _ := json.Marshal(refundCrossPoolTrade)
	return []string{
		strconv.Itoa(metadata.PDECrossPoolTradeRequestMeta),
		strconv.Itoa(int(shardID)),
		status,
		string(refundCrossPoolTradeBytes),
	}
}

// build refund instructions for an unsuccessful trade
// note: only refund if amount > 0
func refundForCrossPoolTrade(
	sequentialTrades []*tradeInfo,
	action metadata.PDECrossPoolTradeRequestAction,
) ([][]string, error) {

	refundTradingFeeInst := buildCrossPoolTradeRefundInst(
		action.Meta.SubTraderAddressStr,
		action.Meta.SubTxRandomStr,
		common.PRVIDStr,
		action.Meta.TradingFee,
		common.PDECrossPoolTradeFeeRefundChainStatus,
		action.ShardID,
		action.TxReqID,
	)
	refundSellingTokenInst := buildCrossPoolTradeRefundInst(
		action.Meta.TraderAddressStr,
		action.Meta.TxRandomStr,
		sequentialTrades[0].tokenIDToSellStr,
		sequentialTrades[0].sellAmount,
		common.PDECrossPoolTradeSellingTokenRefundChainStatus,
		action.ShardID,
		action.TxReqID,
	)

	refundInsts := [][]string{}
	if len(refundTradingFeeInst) > 0 {
		refundInsts = append(refundInsts, refundTradingFeeInst)
	}
	if len(refundSellingTokenInst) > 0 {
		refundInsts = append(refundInsts, refundSellingTokenInst)
	}
	return refundInsts, nil
}

func (s *stateV1) categorizeAndSortCrossPoolTradeInstsByFee(
	actions [][]string,
	beaconHeight uint64,
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
			!s.havePoolPair(beaconHeight, tokenIDToSell, tokenIDToBuy)) ||
			(!isTradingPairContainsPRV(tokenIDToSell, tokenIDToBuy) &&
				(!s.havePoolPair(beaconHeight, prvIDStr, tokenIDToSell) ||
					!s.havePoolPair(beaconHeight, prvIDStr, tokenIDToBuy))) {
			untradableActions = append(untradableActions, crossPoolTradeRequestAction)
			continue
		}
		tradableActions = append(tradableActions, crossPoolTradeRequestAction)
	}
	// sort tradable actions by trading fee
	sort.SliceStable(tradableActions, func(i, j int) bool {
		firstTradingFee, firstSellAmount := s.prepareInfoForSorting(
			beaconHeight,
			tradableActions[i],
		)
		secondTradingFee, secondSellAmount := s.prepareInfoForSorting(
			beaconHeight,
			tradableActions[j],
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

func (s *stateV1) prepareInfoForSorting(
	beaconHeight uint64,
	action metadata.PDECrossPoolTradeRequestAction,
) (uint64, uint64) {
	prvIDStr := common.PRVCoinID.String()
	tradeMeta := action.Meta
	sellAmount := tradeMeta.SellAmount
	tradingFee := tradeMeta.TradingFee
	if tradeMeta.TokenIDToSellStr == prvIDStr {
		return tradingFee, sellAmount
	}
	poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, prvIDStr, tradeMeta.TokenIDToSellStr))
	poolPair, _ := s.poolPairs[poolPairKey]
	sellAmount, _, _, _ = calcTradeValue(poolPair, tradeMeta.TokenIDToSellStr, sellAmount)
	return tradingFee, sellAmount
}

func calcTradeValue(
	poolPair *rawdbv2.PDEPoolForPair,
	tokenIDStrToSell string,
	sellAmount uint64,
) (uint64, uint64, uint64, error) {
	tokenPoolValueToBuy := poolPair.Token1PoolValue
	tokenPoolValueToSell := poolPair.Token2PoolValue
	if poolPair.Token1IDStr == tokenIDStrToSell {
		tokenPoolValueToSell = poolPair.Token1PoolValue
		tokenPoolValueToBuy = poolPair.Token2PoolValue
	}
	invariant := big.NewInt(0)
	invariant.Mul(new(big.Int).SetUint64(tokenPoolValueToSell), new(big.Int).SetUint64(tokenPoolValueToBuy))
	newTokenPoolValueToSell := big.NewInt(0)
	newTokenPoolValueToSell.Add(new(big.Int).SetUint64(tokenPoolValueToSell), new(big.Int).SetUint64(sellAmount))

	newTokenPoolValueToBuy := big.NewInt(0).Div(invariant, newTokenPoolValueToSell).Uint64()
	modValue := big.NewInt(0).Mod(invariant, newTokenPoolValueToSell)
	if modValue.Cmp(big.NewInt(0)) != 0 {
		newTokenPoolValueToBuy++
	}
	if tokenPoolValueToBuy <= newTokenPoolValueToBuy {
		return uint64(0), uint64(0), uint64(0), errors.New("tokenPoolValueToBuy <= newTokenPoolValueToBuy")
	}
	return tokenPoolValueToBuy - newTokenPoolValueToBuy, newTokenPoolValueToBuy, newTokenPoolValueToSell.Uint64(), nil
}

func (s *stateV1) havePoolPair(
	beaconHeight uint64,
	token1IDStr string,
	token2IDStr string,
) bool {
	poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, token1IDStr, token2IDStr))
	poolPair, found := s.poolPairs[poolPairKey]
	if !found || poolPair == nil || poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
		return false
	}
	return true
}

func (s *stateV1) buildInstructionsForFeeWithdrawal(
	actions [][]string,
	beaconHeight uint64,
) ([][]string, error) {
	res := [][]string{}
	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
			return utils.EmptyStringMatrix, nil
		}
		var feeWithdrawalRequestAction metadata.PDEFeeWithdrawalRequestAction
		err = json.Unmarshal(contentBytes, &feeWithdrawalRequestAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde fee withdrawal request action: %+v", err)
			return utils.EmptyStringMatrix, nil
		}
		wdMeta := feeWithdrawalRequestAction.Meta
		tradingFeeKeyBytes, err := rawdbv2.BuildPDETradingFeeKey(
			beaconHeight,
			wdMeta.WithdrawalToken1IDStr,
			wdMeta.WithdrawalToken2IDStr,
			wdMeta.WithdrawerAddressStr,
		)
		if err != nil {
			Logger.log.Errorf("cannot build PDETradingFeeKey for address: %v. Error: %v\n", wdMeta.WithdrawerAddressStr, err)
			return utils.EmptyStringMatrix, err
		}
		tradingFeeKey := string(tradingFeeKeyBytes)
		withdrawableFee, found := s.tradingFees[tradingFeeKey]
		if !found || withdrawableFee < wdMeta.WithdrawalFeeAmt {
			rejectedInst := []string{
				strconv.Itoa(metadata.PDEFeeWithdrawalRequestMeta),
				strconv.Itoa(int(feeWithdrawalRequestAction.ShardID)),
				common.PDEFeeWithdrawalRejectedChainStatus,
				contentStr,
			}
			res = append(res, rejectedInst)
			continue
		}
		s.tradingFees[tradingFeeKey] -= wdMeta.WithdrawalFeeAmt
		acceptedInst := []string{
			strconv.Itoa(metadata.PDEFeeWithdrawalRequestMeta),
			strconv.Itoa(int(feeWithdrawalRequestAction.ShardID)),
			common.PDEFeeWithdrawalAcceptedChainStatus,
			contentStr,
		}
		res = append(res, acceptedInst)
	}
	return res, nil
}

func (s *stateV1) buildInstructionsForTrade(
	actions [][]string,
	beaconHeight uint64,
) ([][]string, error) {
	res := [][]string{}

	// handle trade
	sortedTradesActions := s.sortTradeInstsByFee(
		actions,
		beaconHeight,
	)
	for _, tradeAction := range sortedTradesActions {
		should, receiveAmount, err := s.shouldRefundTradeAction(tradeAction, beaconHeight)
		if err != nil {
			continue
		}
		if should {
			actionContentBytes, err := json.Marshal(tradeAction)
			if err != nil {
				return utils.EmptyStringMatrix, err
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
		inst, err := s.buildAcceptedTradeInstruction(tradeAction, beaconHeight, receiveAmount)
		if err != nil {
			continue
		}
		res = append(res, inst)
	}

	return res, nil
}

func (s *stateV1) shouldRefundCrossPoolTrade(
	sequentialTrades []*tradeInfo,
	action metadata.PDECrossPoolTradeRequestAction,
	beaconHeight uint64,
) (bool, error) {
	if s.poolPairs == nil || len(s.poolPairs) == 0 {
		return true, nil
	}

	amt := sequentialTrades[0].sellAmount
	for _, tradeInf := range sequentialTrades {
		tradeInf.sellAmount = amt
		pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeInf.tokenIDToBuyStr, tradeInf.tokenIDToSellStr))
		pdePoolPair, _ := s.poolPairs[pairKey]
		newAmt, newTokenPoolValueToBuy, newTokenPoolValueToSell, err := calcTradeValue(pdePoolPair, tradeInf.tokenIDToSellStr, amt)
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

func (s *stateV1) buildAcceptedTradeInstruction(
	action metadata.PDETradeRequestAction,
	beaconHeight, receiveAmount uint64,
) ([]string, error) {

	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, action.Meta.TokenIDToBuyStr, action.Meta.TokenIDToSellStr))
	poolPair := s.poolPairs[pairKey]

	pdeTradeAcceptedContent := metadata.PDETradeAcceptedContent{
		TraderAddressStr: action.Meta.TraderAddressStr,
		TxRandomStr:      action.Meta.TxRandomStr,
		TokenIDToBuyStr:  action.Meta.TokenIDToBuyStr,
		ReceiveAmount:    receiveAmount,
		Token1IDStr:      poolPair.Token1IDStr,
		Token2IDStr:      poolPair.Token2IDStr,
		ShardID:          action.ShardID,
		RequestedTxID:    action.TxReqID,
	}
	pdeTradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "-",
		Value:    receiveAmount,
	}
	pdeTradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
		Operator: "+",
		Value:    action.Meta.SellAmount + action.Meta.TradingFee,
	}
	if poolPair.Token1IDStr == action.Meta.TokenIDToSellStr {
		pdeTradeAcceptedContent.Token1PoolValueOperation = metadata.TokenPoolValueOperation{
			Operator: "+",
			Value:    action.Meta.SellAmount + action.Meta.TradingFee,
		}
		pdeTradeAcceptedContent.Token2PoolValueOperation = metadata.TokenPoolValueOperation{
			Operator: "-",
			Value:    receiveAmount,
		}
	}
	pdeTradeAcceptedContentBytes, err := json.Marshal(pdeTradeAcceptedContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while marshaling pdeTradeAcceptedContent: %+v", err)
		return utils.EmptyStringArray, err
	}
	return []string{
		strconv.Itoa(metadata.PDETradeRequestMeta),
		strconv.Itoa(int(action.ShardID)),
		common.PDETradeAcceptedChainStatus,
		string(pdeTradeAcceptedContentBytes),
	}, nil
}

func (s *stateV1) sortTradeInstsByFee(
	actions [][]string,
	beaconHeight uint64,
) []metadata.PDETradeRequestAction {
	// TODO: @tin improve here for v2 by sorting only with fee not necessary with poolPairs sort
	tradesByPairs := make(map[string][]metadata.PDETradeRequestAction)

	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade action: %+v", err)
			continue
		}
		pdeTradeReqAction := metadata.PDETradeRequestAction{}
		err = json.Unmarshal(contentBytes, &pdeTradeReqAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade action: %+v", err)
			continue
		}
		tradeMeta := pdeTradeReqAction.Meta
		poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeMeta.TokenIDToBuyStr, tradeMeta.TokenIDToSellStr))
		tradesByPairs[poolPairKey] = append(tradesByPairs[poolPairKey], pdeTradeReqAction)
	}

	notExistingPairTradeActions := []metadata.PDETradeRequestAction{}
	sortedExistingPairTradeActions := []metadata.PDETradeRequestAction{}

	var ppKeys []string
	for k := range tradesByPairs {
		ppKeys = append(ppKeys, k)
	}
	sort.Strings(ppKeys)
	for _, poolPairKey := range ppKeys {
		tradeActions := tradesByPairs[poolPairKey]
		poolPair, found := s.poolPairs[poolPairKey]
		if !found || poolPair == nil || poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
			notExistingPairTradeActions = append(notExistingPairTradeActions, tradeActions...)
			continue
		}

		// sort trade actions by trading fee
		sort.Slice(tradeActions, func(i, j int) bool {
			// comparing a/b to c/d is equivalent with comparing a*d to c*b
			firstItemProportion := big.NewInt(0)
			firstItemProportion.Mul(
				new(big.Int).SetUint64(tradeActions[i].Meta.TradingFee),
				new(big.Int).SetUint64(tradeActions[j].Meta.SellAmount),
			)
			secondItemProportion := big.NewInt(0)
			secondItemProportion.Mul(
				new(big.Int).SetUint64(tradeActions[j].Meta.TradingFee),
				new(big.Int).SetUint64(tradeActions[i].Meta.SellAmount),
			)
			return firstItemProportion.Cmp(secondItemProportion) == 1
		})
		sortedExistingPairTradeActions = append(sortedExistingPairTradeActions, tradeActions...)
	}
	return append(sortedExistingPairTradeActions, notExistingPairTradeActions...)
}

func (s *stateV1) shouldRefundTradeAction(
	action metadata.PDETradeRequestAction,
	beaconHeight uint64,
) (bool, uint64, error) {

	if len(s.poolPairs) == 0 || s.poolPairs == nil {
		return true, 0, nil
	}

	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, action.Meta.TokenIDToBuyStr, action.Meta.TokenIDToSellStr))
	poolPair, found := s.poolPairs[pairKey]
	if !found || poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
		return true, 0, nil
	}

	receiveAmt, newTokenPoolValueToBuy, tempNewTokenPoolValueToSell, err := calcTradeValue(poolPair, action.Meta.TokenIDToSellStr, action.Meta.SellAmount)
	if err != nil {
		return true, 0, nil
	}
	if action.Meta.MinAcceptableAmount > receiveAmt {
		return true, 0, nil
	}

	// update current pde state on mem
	newTokenPoolValueToSell := new(big.Int).SetUint64(tempNewTokenPoolValueToSell)
	fee := action.Meta.TradingFee
	newTokenPoolValueToSell.Add(newTokenPoolValueToSell, new(big.Int).SetUint64(fee))

	poolPair.Token1PoolValue = newTokenPoolValueToBuy
	poolPair.Token2PoolValue = newTokenPoolValueToSell.Uint64()
	if poolPair.Token1IDStr == action.Meta.TokenIDToSellStr {
		poolPair.Token1PoolValue = newTokenPoolValueToSell.Uint64()
		poolPair.Token2PoolValue = newTokenPoolValueToBuy
	}

	return false, receiveAmt, nil
}

func (s *stateV1) buildInstructionsForWithdrawal(
	actions [][]string,
	beaconHeight uint64,
) ([][]string, error) {
	res := [][]string{}

	for _, action := range actions {
		contentStr := action[1]
		contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while decoding content string of pde withdrawal action: %+v", err)
			return utils.EmptyStringMatrix, err
		}
		var withdrawalRequestAction metadata.PDEWithdrawalRequestAction
		err = json.Unmarshal(contentBytes, &withdrawalRequestAction)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while unmarshaling pde withdrawal request action: %+v", err)
			return [][]string{}, err
		}
		wdMeta := withdrawalRequestAction.Meta
		deductingAmounts := s.deductAmounts(
			wdMeta,
			beaconHeight,
		)

		if deductingAmounts == nil {
			inst := []string{
				strconv.Itoa(metadata.PDEWithdrawalRequestMeta),
				strconv.Itoa(int(withdrawalRequestAction.ShardID)),
				common.PDEWithdrawalRejectedChainStatus,
				contentStr,
			}
			return [][]string{inst}, nil
		}

		inst, err := buildWithdrawalAcceptedInst(
			withdrawalRequestAction,
			deductingAmounts.Token1IDStr,
			deductingAmounts.PoolValue1,
			deductingAmounts.Shares,
		)
		if err != nil {
			return [][]string{}, nil
		}
		res = append(res, inst)
		inst, err = buildWithdrawalAcceptedInst(
			withdrawalRequestAction,
			deductingAmounts.Token2IDStr,
			deductingAmounts.PoolValue2,
			0,
		)
		if err != nil {
			return [][]string{}, nil
		}
		res = append(res, inst)
	}

	return res, nil
}

func (s *stateV1) deductAmounts(
	wdMeta metadata.PDEWithdrawalRequest,
	beaconHeight uint64,
) *deductingAmountsByWithdrawal {
	var res *deductingAmountsByWithdrawal
	pairKey := string(rawdbv2.BuildPDEPoolForPairKey(
		beaconHeight,
		wdMeta.WithdrawalToken1IDStr, wdMeta.WithdrawalToken2IDStr,
	))
	poolPair, found := s.poolPairs[pairKey]
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
	currentSharesForWithdrawer, found := s.shares[shareForWithdrawerKey]
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

	for shareKey, shareAmt := range s.shares {
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

	if s.shares[shareForWithdrawerKey] < wdSharesForWithdrawer {
		s.shares[shareForWithdrawerKey] = 0
	} else {
		s.shares[shareForWithdrawerKey] -= wdSharesForWithdrawer
	}
	res.Shares = wdSharesForWithdrawer

	return res
}

func buildWithdrawalAcceptedInst(
	action metadata.PDEWithdrawalRequestAction,
	withdrawalTokenIDStr string,
	deductingPoolValue uint64,
	deductingShares uint64,
) ([]string, error) {
	wdAcceptedContent := metadata.PDEWithdrawalAcceptedContent{
		WithdrawalTokenIDStr: withdrawalTokenIDStr,
		WithdrawerAddressStr: action.Meta.WithdrawerAddressStr,
		DeductingPoolValue:   deductingPoolValue,
		DeductingShares:      deductingShares,
		PairToken1IDStr:      action.Meta.WithdrawalToken1IDStr,
		PairToken2IDStr:      action.Meta.WithdrawalToken2IDStr,
		TxReqID:              action.TxReqID,
		ShardID:              action.ShardID,
	}
	wdAcceptedContentBytes, err := json.Marshal(wdAcceptedContent)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while marshaling PDEWithdrawalAcceptedContent: %+v", err)
		return []string{}, nil
	}
	return []string{
		strconv.Itoa(metadata.PDEWithdrawalRequestMeta),
		strconv.Itoa(int(action.ShardID)),
		common.PDEWithdrawalAcceptedChainStatus,
		string(wdAcceptedContentBytes),
	}, nil
}

func (s *stateV1) buildInstructionsForContribution(actions [][]string) ([][]string, error) {
	res := [][]string{}
	return res, nil
}

func (s *stateV1) buildInstructionsForPRVRequiredContribution(actions [][]string) ([][]string, error) {
	res := [][]string{}
	return res, nil
}

func (s *stateV1) Upgrade(env StateEnvironment) State {
	var state State
	return state
}
