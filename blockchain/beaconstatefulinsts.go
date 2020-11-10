package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
)

// build instructions at beacon chain before syncing to shards
func (blockchain *BlockChain) collectStatefulActions(
	shardBlockInstructions [][]string,
) [][]string {
	// stateful instructions are dependently processed with results of instructioins before them in shards2beacon blocks
	statefulInsts := [][]string{}
	for _, inst := range shardBlockInstructions {
		if len(inst) < 2 {
			continue
		}
		if inst[0] == SetAction || inst[0] == StakeAction || inst[0] == SwapAction || inst[0] == RandomAction || inst[0] == AssignAction {
			continue
		}

		metaType, err := strconv.Atoi(inst[0])
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		switch metaType {
		case metadata.IssuingRequestMeta,
			metadata.IssuingETHRequestMeta,
			metadata.PDEContributionMeta,
			metadata.PDETradeRequestMeta,
			metadata.PDEWithdrawalRequestMeta,
			metadata.PDEFeeWithdrawalRequestMeta,
			metadata.PDEPRVRequiredContributionRequestMeta,
			metadata.PDECrossPoolTradeRequestMeta,
			metadata.PortalCustodianDepositMeta,
			metadata.PortalUserRegisterMeta,
			metadata.PortalUserRequestPTokenMeta,
			metadata.PortalExchangeRatesMeta,
			metadata.RelayingBNBHeaderMeta,
			metadata.RelayingBTCHeaderMeta,
			metadata.PortalCustodianWithdrawRequestMeta,
			metadata.PortalRedeemRequestMeta,
			metadata.PortalRequestUnlockCollateralMeta,
			metadata.PortalRequestUnlockCollateralMetaV3,
			metadata.PortalLiquidateCustodianMeta,
			metadata.PortalLiquidateCustodianMetaV3,
			metadata.PortalRequestWithdrawRewardMeta,
			metadata.PortalRedeemFromLiquidationPoolMeta,
			metadata.PortalCustodianTopupMetaV2,
			metadata.PortalCustodianTopupResponseMeta,
			metadata.PortalReqMatchingRedeemMeta,
			metadata.PortalTopUpWaitingPortingRequestMeta,
			metadata.PortalCustodianDepositMetaV3,
			metadata.PortalCustodianWithdrawRequestMetaV3,
			metadata.PortalRedeemFromLiquidationPoolMetaV3,
			metadata.PortalCustodianTopupMetaV3,
			metadata.PortalTopUpWaitingPortingRequestMetaV3:
			statefulInsts = append(statefulInsts, inst)

		default:
			continue
		}
	}
	return statefulInsts
}

func groupPDEActionsByShardID(
	pdeActionsByShardID map[byte][][]string,
	action []string,
	shardID byte,
) map[byte][][]string {
	_, found := pdeActionsByShardID[shardID]
	if !found {
		pdeActionsByShardID[shardID] = [][]string{action}
	} else {
		pdeActionsByShardID[shardID] = append(pdeActionsByShardID[shardID], action)
	}
	return pdeActionsByShardID
}

func (blockchain *BlockChain) buildStatefulInstructions(
	stateDB *statedb.StateDB,
	statefulActionsByShardID map[byte][][]string,
	beaconHeight uint64,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	portalParams PortalParams) [][]string {
	currentPDEState, err := InitCurrentPDEStateFromDB(stateDB, beaconHeight-1)
	if err != nil {
		Logger.log.Error(err)
	}

	currentPortalState, err := InitCurrentPortalStateFromDB(stateDB)
	if err != nil {
		Logger.log.Error(err)
	}

	pm := NewPortalManager()
	relayingHeaderState, err := blockchain.InitRelayingHeaderChainStateFromDB()
	if err != nil {
		Logger.log.Error(err)
	}

	accumulatedValues := &metadata.AccumulatedValues{
		UniqETHTxsUsed:   [][]byte{},
		DBridgeTokenPair: map[string][]byte{},
		CBridgeTokens:    []*common.Hash{},
	}
	instructions := [][]string{}

	// pde instructions
	pdeContributionActionsByShardID := map[byte][][]string{}
	pdePRVRequiredContributionActionsByShardID := map[byte][][]string{}
	pdeTradeActionsByShardID := map[byte][][]string{}
	pdeCrossPoolTradeActionsByShardID := map[byte][][]string{}
	pdeWithdrawalActionsByShardID := map[byte][][]string{}
	pdeFeeWithdrawalActionsByShardID := map[byte][][]string{}

	var keys []int
	for k := range statefulActionsByShardID {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, value := range keys {
		shardID := byte(value)
		actions := statefulActionsByShardID[shardID]
		for _, action := range actions {
			metaType, err := strconv.Atoi(action[0])
			if err != nil {
				continue
			}
			contentStr := action[1]
			newInst := [][]string{}
			switch metaType {
			case metadata.IssuingRequestMeta:
				newInst, err = blockchain.buildInstructionsForIssuingReq(stateDB, contentStr, shardID, metaType, accumulatedValues)

			case metadata.IssuingETHRequestMeta:
				newInst, err = blockchain.buildInstructionsForIssuingETHReq(stateDB, contentStr, shardID, metaType, accumulatedValues)

			case metadata.PDEContributionMeta:
				pdeContributionActionsByShardID = groupPDEActionsByShardID(
					pdeContributionActionsByShardID,
					action,
					shardID,
				)
			case metadata.PDEPRVRequiredContributionRequestMeta:
				pdePRVRequiredContributionActionsByShardID = groupPDEActionsByShardID(
					pdePRVRequiredContributionActionsByShardID,
					action,
					shardID,
				)
			case metadata.PDETradeRequestMeta:
				pdeTradeActionsByShardID = groupPDEActionsByShardID(
					pdeTradeActionsByShardID,
					action,
					shardID,
				)
			case metadata.PDECrossPoolTradeRequestMeta:
				pdeCrossPoolTradeActionsByShardID = groupPDEActionsByShardID(
					pdeCrossPoolTradeActionsByShardID,
					action,
					shardID,
				)
			case metadata.PDEWithdrawalRequestMeta:
				pdeWithdrawalActionsByShardID = groupPDEActionsByShardID(
					pdeWithdrawalActionsByShardID,
					action,
					shardID,
				)
			case metadata.PDEFeeWithdrawalRequestMeta:
				pdeFeeWithdrawalActionsByShardID = groupPDEActionsByShardID(
					pdeFeeWithdrawalActionsByShardID,
					action,
					shardID,
				)
			case metadata.PortalCustodianDepositMeta:
				pm.portalInstructions[metadata.PortalCustodianDepositMeta].putAction(action, shardID)
			case metadata.PortalUserRegisterMeta:
				pm.portalInstructions[metadata.PortalUserRegisterMeta].putAction(action, shardID)
			case metadata.PortalUserRequestPTokenMeta:
				pm.portalInstructions[metadata.PortalUserRequestPTokenMeta].putAction(action, shardID)
			case metadata.PortalExchangeRatesMeta:
				pm.portalInstructions[metadata.PortalExchangeRatesMeta].putAction(action, shardID)
			case metadata.PortalCustodianWithdrawRequestMeta:
				pm.portalInstructions[metadata.PortalCustodianWithdrawRequestMeta].putAction(action, shardID)

			case metadata.PortalRedeemRequestMeta:
				pm.portalInstructions[metadata.PortalRedeemRequestMeta].putAction(action, shardID)
			case metadata.PortalRequestUnlockCollateralMeta, metadata.PortalRequestUnlockCollateralMetaV3:
				pm.portalInstructions[metadata.PortalRequestUnlockCollateralMeta].putAction(action, shardID)
			case metadata.PortalRequestWithdrawRewardMeta:
				pm.portalInstructions[metadata.PortalRequestWithdrawRewardMeta].putAction(action, shardID)
			case metadata.PortalRedeemFromLiquidationPoolMetaV3:
				pm.portalInstructions[metadata.PortalRedeemFromLiquidationPoolMetaV3].putAction(action, shardID)
			case metadata.PortalCustodianTopupMetaV2:
				pm.portalInstructions[metadata.PortalCustodianTopupMetaV2].putAction(action, shardID)
			case metadata.PortalReqMatchingRedeemMeta:
				pm.portalInstructions[metadata.PortalReqMatchingRedeemMeta].putAction(action, shardID)
			case metadata.PortalTopUpWaitingPortingRequestMeta:
				pm.portalInstructions[metadata.PortalTopUpWaitingPortingRequestMeta].putAction(action, shardID)
			case metadata.PortalCustodianTopupMetaV3:
				pm.portalInstructions[metadata.PortalCustodianTopupMetaV3].putAction(action, shardID)
			case metadata.PortalTopUpWaitingPortingRequestMetaV3:
				pm.portalInstructions[metadata.PortalTopUpWaitingPortingRequestMetaV3].putAction(action, shardID)

			case metadata.PortalCustodianDepositMetaV3:
				pm.portalInstructions[metadata.PortalCustodianDepositMetaV3].putAction(action, shardID)
			case metadata.PortalCustodianWithdrawRequestMetaV3:
				pm.portalInstructions[metadata.PortalCustodianWithdrawRequestMetaV3].putAction(action, shardID)

			case metadata.RelayingBNBHeaderMeta:
				pm.relayingChains[metadata.RelayingBNBHeaderMeta].putAction(action)
			case metadata.RelayingBTCHeaderMeta:
				pm.relayingChains[metadata.RelayingBTCHeaderMeta].putAction(action)
			default:
				continue
			}
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	pdeInsts, err := blockchain.handlePDEInsts(
		beaconHeight-1, currentPDEState,
		pdeContributionActionsByShardID,
		pdePRVRequiredContributionActionsByShardID,
		pdeTradeActionsByShardID,
		pdeCrossPoolTradeActionsByShardID,
		pdeWithdrawalActionsByShardID,
		pdeFeeWithdrawalActionsByShardID,
	)

	if err != nil {
		Logger.log.Error(err)
		return instructions
	}
	if len(pdeInsts) > 0 {
		instructions = append(instructions, pdeInsts...)
	}

	// handle portal instructions
	portalInsts, err := blockchain.handlePortalInsts(
		stateDB,
		beaconHeight-1,
		currentPortalState,
		rewardForCustodianByEpoch,
		portalParams,
		pm,
	)

	if err != nil {
		Logger.log.Error(err)
		return instructions
	}
	if len(portalInsts) > 0 {
		instructions = append(instructions, portalInsts...)
	}

	// handle relaying instructions
	relayingInsts := blockchain.handleRelayingInsts(relayingHeaderState, pm)
	if len(relayingInsts) > 0 {
		instructions = append(instructions, relayingInsts...)
	}

	return instructions
}

func isTradingFairContainsPRV(
	tokenIDToSellStr string,
	tokenIDToBuyStr string,
) bool {
	return tokenIDToSellStr == common.PRVCoinID.String() || tokenIDToBuyStr == common.PRVCoinID.String()
}

func isPoolPairExisting(
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
	token1IDStr string,
	token2IDStr string,
) bool {
	poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, token1IDStr, token2IDStr))
	poolPair, found := currentPDEState.PDEPoolPairs[poolPairKey]
	if !found || poolPair == nil || poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
		return false
	}
	return true
}

func calcTradeValue(
	pdePoolPair *rawdbv2.PDEPoolForPair,
	tokenIDStrToSell string,
	sellAmount uint64,
) (uint64, uint64, uint64) {
	tokenPoolValueToBuy := pdePoolPair.Token1PoolValue
	tokenPoolValueToSell := pdePoolPair.Token2PoolValue
	if pdePoolPair.Token1IDStr == tokenIDStrToSell {
		tokenPoolValueToSell = pdePoolPair.Token1PoolValue
		tokenPoolValueToBuy = pdePoolPair.Token2PoolValue
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
		return uint64(0), uint64(0), uint64(0)
	}
	return tokenPoolValueToBuy - newTokenPoolValueToBuy, newTokenPoolValueToBuy, newTokenPoolValueToSell.Uint64()
}

func prepareInfoForSorting(
	currentPDEState *CurrentPDEState,
	beaconHeight uint64,
	tradeAction metadata.PDECrossPoolTradeRequestAction,
) (uint64, uint64) {
	prvIDStr := common.PRVCoinID.String()
	tradeMeta := tradeAction.Meta
	sellAmount := tradeMeta.SellAmount
	tradingFee := tradeMeta.TradingFee
	if tradeMeta.TokenIDToSellStr == prvIDStr {
		return tradingFee, sellAmount
	}
	poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, prvIDStr, tradeMeta.TokenIDToSellStr))
	poolPair, _ := currentPDEState.PDEPoolPairs[poolPairKey]
	sellAmount, _, _ = calcTradeValue(poolPair, tradeMeta.TokenIDToSellStr, sellAmount)
	return tradingFee, sellAmount
}

func categorizeNSortPDECrossPoolTradeInstsByFee(
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
	pdeCrossPoolTradeActionsByShardID map[byte][][]string,
) ([]metadata.PDECrossPoolTradeRequestAction, []metadata.PDECrossPoolTradeRequestAction) {
	prvIDStr := common.PRVCoinID.String()
	tradableActions := []metadata.PDECrossPoolTradeRequestAction{}
	untradableActions := []metadata.PDECrossPoolTradeRequestAction{}
	var keys []int
	for k := range pdeCrossPoolTradeActionsByShardID {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, value := range keys {
		shardID := byte(value)
		actions := pdeCrossPoolTradeActionsByShardID[shardID]
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
			if (isTradingFairContainsPRV(tradeMeta.TokenIDToSellStr, tradeMeta.TokenIDToBuyStr) && !isPoolPairExisting(beaconHeight, currentPDEState, tradeMeta.TokenIDToSellStr, tradeMeta.TokenIDToBuyStr)) ||
				(!isTradingFairContainsPRV(tradeMeta.TokenIDToSellStr, tradeMeta.TokenIDToBuyStr) && (!isPoolPairExisting(beaconHeight, currentPDEState, prvIDStr, tradeMeta.TokenIDToSellStr) || !isPoolPairExisting(beaconHeight, currentPDEState, prvIDStr, tradeMeta.TokenIDToBuyStr))) {
				untradableActions = append(untradableActions, crossPoolTradeRequestAction)
				continue
			}
			tradableActions = append(tradableActions, crossPoolTradeRequestAction)
		}
	}

	// sort tradable actions by trading fee
	sort.SliceStable(tradableActions, func(i, j int) bool {
		firstTradingFee, firstSellAmount := prepareInfoForSorting(
			currentPDEState,
			beaconHeight,
			tradableActions[i],
		)
		secondTradingFee, secondSellAmount := prepareInfoForSorting(
			currentPDEState,
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

func sortPDETradeInstsByFee(
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
	pdeTradeActionsByShardID map[byte][][]string,
) []metadata.PDETradeRequestAction {
	tradesByPairs := make(map[string][]metadata.PDETradeRequestAction)

	var keys []int
	for k := range pdeTradeActionsByShardID {
		keys = append(keys, int(k))
	}
	sort.Ints(keys)
	for _, value := range keys {
		shardID := byte(value)
		actions := pdeTradeActionsByShardID[shardID]
		for _, action := range actions {
			contentStr := action[1]
			contentBytes, err := base64.StdEncoding.DecodeString(contentStr)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while decoding content string of pde trade action: %+v", err)
				continue
			}
			var pdeTradeReqAction metadata.PDETradeRequestAction
			err = json.Unmarshal(contentBytes, &pdeTradeReqAction)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while unmarshaling pde trade action: %+v", err)
				continue
			}
			tradeMeta := pdeTradeReqAction.Meta
			poolPairKey := string(rawdbv2.BuildPDEPoolForPairKey(beaconHeight, tradeMeta.TokenIDToBuyStr, tradeMeta.TokenIDToSellStr))
			tradesByPair, found := tradesByPairs[poolPairKey]
			if !found {
				tradesByPairs[poolPairKey] = []metadata.PDETradeRequestAction{pdeTradeReqAction}
			} else {
				tradesByPairs[poolPairKey] = append(tradesByPair, pdeTradeReqAction)
			}
		}
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
		poolPair, found := currentPDEState.PDEPoolPairs[poolPairKey]
		if !found || poolPair == nil {
			notExistingPairTradeActions = append(notExistingPairTradeActions, tradeActions...)
			continue
		}
		if poolPair.Token1PoolValue == 0 || poolPair.Token2PoolValue == 0 {
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

func (blockchain *BlockChain) handlePDEInsts(
	beaconHeight uint64,
	currentPDEState *CurrentPDEState,
	pdeContributionActionsByShardID map[byte][][]string,
	pdePRVRequiredContributionActionsByShardID map[byte][][]string,
	pdeTradeActionsByShardID map[byte][][]string,
	pdeCrossPoolTradeActionsByShardID map[byte][][]string,
	pdeWithdrawalActionsByShardID map[byte][][]string,
	pdeFeeWithdrawalActionsByShardID map[byte][][]string,
) ([][]string, error) {
	instructions := [][]string{}

	// handle fee withdrawal
	var feeWRKeys []int
	for k := range pdeFeeWithdrawalActionsByShardID {
		feeWRKeys = append(feeWRKeys, int(k))
	}
	sort.Ints(feeWRKeys)
	for _, value := range feeWRKeys {
		shardID := byte(value)
		actions := pdeFeeWithdrawalActionsByShardID[shardID]
		for _, action := range actions {
			contentStr := action[1]
			newInst, err := blockchain.buildInstructionsForPDEFeeWithdrawal(contentStr, shardID, metadata.PDEFeeWithdrawalRequestMeta, currentPDEState, beaconHeight)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	// handle trade
	sortedTradesActions := sortPDETradeInstsByFee(
		beaconHeight,
		currentPDEState,
		pdeTradeActionsByShardID,
	)
	for _, tradeAction := range sortedTradesActions {
		actionContentBytes, _ := json.Marshal(tradeAction)
		actionContentBase64Str := base64.StdEncoding.EncodeToString(actionContentBytes)
		newInst, err := blockchain.buildInstructionsForPDETrade(actionContentBase64Str, tradeAction.ShardID, metadata.PDETradeRequestMeta, currentPDEState, beaconHeight)
		if err != nil {
			Logger.log.Error(err)
			continue
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}

	// handle cross pool trade
	sortedTradableActions, untradableActions := categorizeNSortPDECrossPoolTradeInstsByFee(
		beaconHeight,
		currentPDEState,
		pdeCrossPoolTradeActionsByShardID,
	)
	tradableInsts, tradingFeeByPair := blockchain.buildInstsForSortedTradableActions(currentPDEState, beaconHeight, sortedTradableActions)
	untradableInsts := blockchain.buildInstsForUntradableActions(untradableActions)
	instructions = append(instructions, tradableInsts...)
	instructions = append(instructions, untradableInsts...)

	// calculate and build instruction for trading fees distribution
	tradingFeesDistInst := blockchain.buildInstForTradingFeesDist(currentPDEState, beaconHeight, tradingFeeByPair)
	if len(tradingFeesDistInst) > 0 {
		instructions = append(instructions, tradingFeesDistInst)
	}

	// handle withdrawal
	var wrKeys []int
	for k := range pdeWithdrawalActionsByShardID {
		wrKeys = append(wrKeys, int(k))
	}
	sort.Ints(wrKeys)
	for _, value := range wrKeys {
		shardID := byte(value)
		actions := pdeWithdrawalActionsByShardID[shardID]
		for _, action := range actions {
			contentStr := action[1]
			newInst, err := blockchain.buildInstructionsForPDEWithdrawal(contentStr, shardID, metadata.PDEWithdrawalRequestMeta, currentPDEState, beaconHeight)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	// handle contribution
	var ctKeys []int
	for k := range pdeContributionActionsByShardID {
		ctKeys = append(ctKeys, int(k))
	}
	sort.Ints(ctKeys)
	for _, value := range ctKeys {
		shardID := byte(value)
		actions := pdeContributionActionsByShardID[shardID]
		for _, action := range actions {
			contentStr := action[1]
			newInst, err := blockchain.buildInstructionsForPDEContribution(contentStr, shardID, metadata.PDEContributionMeta, currentPDEState, beaconHeight, false)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	// handle prv required contribution
	var prvRequiredContribKeys []int
	for k := range pdePRVRequiredContributionActionsByShardID {
		prvRequiredContribKeys = append(prvRequiredContribKeys, int(k))
	}
	sort.Ints(prvRequiredContribKeys)
	for _, value := range prvRequiredContribKeys {
		shardID := byte(value)
		actions := pdePRVRequiredContributionActionsByShardID[shardID]
		for _, action := range actions {
			contentStr := action[1]
			newInst, err := blockchain.buildInstructionsForPDEContribution(contentStr, shardID, metadata.PDEPRVRequiredContributionRequestMeta, currentPDEState, beaconHeight, true)
			if err != nil {
				Logger.log.Error(err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}
	return instructions, nil
}

func (blockchain *BlockChain) handlePortalInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	portalParams PortalParams,
	pm *portalManager,
) ([][]string, error) {
	instructions := [][]string{}

	oldMatchedRedeemRequests := cloneRedeemRequests(currentPortalState.MatchedRedeemRequests)

	// get shard height of all shards for producer
	shardHeights := map[byte]uint64{}
	for i := 0; i < common.MaxShardNumber; i++ {
		shardHeights[byte(i)] = blockchain.ShardChain[i].multiView.GetBestView().GetHeight()
	}

	// auto-liquidation portal instructions
	portalLiquidationInsts, err := blockchain.autoCheckAndCreatePortalLiquidationInsts(
		beaconHeight,
		shardHeights,
		currentPortalState,
		portalParams,
	)
	if err != nil {
		Logger.log.Error(err)
	}
	if len(portalLiquidationInsts) > 0 {
		instructions = append(instructions, portalLiquidationInsts...)
	}

	// producer portal instructions for actions from shards
	// sort metadata type map to make it consistent for every run
	var metaTypes []int
	for metaType := range pm.portalInstructions {
		metaTypes = append(metaTypes, metaType)
	}
	sort.Ints(metaTypes)
	for _, metaType := range metaTypes {
		actions := pm.portalInstructions[metaType]
		newInst, err := buildNewPortalInstsFromActions(
			actions,
			blockchain,
			stateDB,
			currentPortalState,
			beaconHeight,
			portalParams)

		if err != nil {
			Logger.log.Error(err)
		}
		if len(newInst) > 0 {
			instructions = append(instructions, newInst...)
		}
	}

	// check and create instruction for picking more custodians for timeout waiting redeem requests
	var pickCustodiansForRedeemReqInsts [][]string
	pickCustodiansForRedeemReqInsts, err = blockchain.checkAndPickMoreCustodianForWaitingRedeemRequest(
		beaconHeight,
		shardHeights,
		currentPortalState,
	)
	if err != nil {
		Logger.log.Error(err)
	}
	if len(pickCustodiansForRedeemReqInsts) > 0 {
		instructions = append(instructions, pickCustodiansForRedeemReqInsts...)
	}

	// get new matched redeem request at beacon height
	newMatchedRedeemReqIDs := getNewMatchedRedeemReqIDs(oldMatchedRedeemRequests, currentPortalState.MatchedRedeemRequests)

	// calculate rewards (include porting fee and redeem fee) for custodians and build instructions at beaconHeight
	portalRewardsInsts, err := blockchain.handlePortalRewardInsts(
		beaconHeight,
		currentPortalState,
		rewardForCustodianByEpoch,
		newMatchedRedeemReqIDs,
	)

	if err != nil {
		Logger.log.Error(err)
	}
	if len(portalRewardsInsts) > 0 {
		instructions = append(instructions, portalRewardsInsts...)
	}

	return instructions, nil
}

func (blockchain *BlockChain) autoCheckAndCreatePortalLiquidationInsts(
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams) ([][]string, error) {
	insts := [][]string{}

	// check there is any waiting porting request timeout
	expiredWaitingPortingInsts, err := blockchain.checkAndBuildInstForExpiredWaitingPortingRequest(beaconHeight, shardHeights, currentPortalState, portalParams)
	if err != nil {
		Logger.log.Errorf("Error when check and build custodian liquidation %v\n", err)
	}
	if len(expiredWaitingPortingInsts) > 0 {
		insts = append(insts, expiredWaitingPortingInsts...)
	}
	Logger.log.Infof("There are %v instruction for expired waiting porting in portal\n", len(expiredWaitingPortingInsts))

	// case 1: check there is any custodian doesn't send public tokens back to user after TimeOutCustodianReturnPubToken
	// get custodian's collateral to return user
	custodianLiqInsts, err := blockchain.checkAndBuildInstForCustodianLiquidation(beaconHeight, shardHeights, currentPortalState, portalParams)
	if err != nil {
		Logger.log.Errorf("Error when check and build custodian liquidation %v\n", err)
	}
	if len(custodianLiqInsts) > 0 {
		insts = append(insts, custodianLiqInsts...)
	}
	Logger.log.Infof("There are %v instruction for custodian liquidation in portal\n", len(custodianLiqInsts))

	// case 2: check collateral's value (locked collateral amount) drops below MinRatio
	exchangeRatesLiqInsts, err := blockchain.buildInstForLiquidationByExchangeRatesV3(beaconHeight, currentPortalState, portalParams)
	if err != nil {
		Logger.log.Errorf("Error when check and build exchange rates liquidation %v\n", err)
	}
	if len(exchangeRatesLiqInsts) > 0 {
		insts = append(insts, exchangeRatesLiqInsts...)
	}

	Logger.log.Infof("There are %v instruction for exchange rates liquidation in portal\n", len(exchangeRatesLiqInsts))

	return insts, nil
}

// handlePortalRewardInsts
// Build instructions portal reward for each beacon block
func (blockchain *BlockChain) handlePortalRewardInsts(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	newMatchedRedeemReqIDs []string,
) ([][]string, error) {
	instructions := [][]string{}

	// Build instructions portal reward for each beacon block
	portalRewardInsts, err := blockchain.buildPortalRewardsInsts(beaconHeight, currentPortalState, rewardForCustodianByEpoch, newMatchedRedeemReqIDs)
	if err != nil {
		Logger.log.Error(err)
	}
	if len(portalRewardInsts) > 0 {
		instructions = append(instructions, portalRewardInsts...)
	}

	return instructions, nil
}

func buildNewPortalInstsFromActions(
	p portalInstructionProcessor,
	bc *BlockChain,
	stateDB *statedb.StateDB,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams) ([][]string, error) {

	instructions := [][]string{}
	actions := p.getActions()
	var shardIDKeys []int
	for k := range actions {
		shardIDKeys = append(shardIDKeys, int(k))
	}

	sort.Ints(shardIDKeys)
	for _, value := range shardIDKeys {
		shardID := byte(value)
		actions := actions[shardID]
		for _, action := range actions {
			contentStr := action[1]
			optionalData, err := p.prepareDataBeforeProcessing(stateDB, contentStr)
			if err != nil {
				Logger.log.Errorf("Error when preparing data before processing instruction %+v", err)
				continue
			}
			newInst, err := p.buildNewInsts(
				bc,
				contentStr,
				shardID,
				currentPortalState,
				beaconHeight,
				portalParams,
				optionalData,
			)
			if err != nil {
				Logger.log.Errorf("Error when building new instructions : %v", err)
				continue
			}
			if len(newInst) > 0 {
				instructions = append(instructions, newInst...)
			}
		}
	}

	return instructions, nil
}
