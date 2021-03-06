package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/config"
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/dataaccessobject/rawdbv2"
	"github.com/incognitochain/incognito-chain/instruction"
	"github.com/incognitochain/incognito-chain/portal"
	portalprocessv3 "github.com/incognitochain/incognito-chain/portal/portalv3/portalprocess"

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
		if instruction.IsConsensusInstruction(inst[0]) {
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
			metadata.IssuingBSCRequestMeta,
			metadata.PDEContributionMeta,
			metadata.PDETradeRequestMeta,
			metadata.PDEWithdrawalRequestMeta,
			metadata.PDEFeeWithdrawalRequestMeta,
			metadata.PDEPRVRequiredContributionRequestMeta,
			metadata.PDECrossPoolTradeRequestMeta,
			metadata.PortalCustodianDepositMeta,
			metadata.PortalRequestPortingMeta,
			metadata.PortalUserRequestPTokenMeta,
			metadata.PortalExchangeRatesMeta,
			metadata.PortalUnlockOverRateCollateralsMeta,
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
			metadata.PortalTopUpWaitingPortingRequestMetaV3,
			metadata.PortalRequestPortingMetaV3,
			metadata.PortalRedeemRequestMetaV3:
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
	beaconBestState *BeaconBestState,
	featureStateDB *statedb.StateDB,
	statefulActionsByShardID map[byte][][]string,
	beaconHeight uint64,
	rewardForCustodianByEpoch map[common.Hash]uint64,
	portalParams portal.PortalParams) [][]string {

	currentPDEState, err := InitCurrentPDEStateFromDB(featureStateDB, beaconBestState.pdeState, beaconHeight-1)
	if err != nil {
		Logger.log.Error(err)
	}

	pm := portal.NewPortalManager()
	currentPortalStateV3, err := portalprocessv3.InitCurrentPortalStateFromDB(featureStateDB)
	if err != nil {
		Logger.log.Error(err)
	}
	relayingHeaderState, err := blockchain.InitRelayingHeaderChainStateFromDB()
	if err != nil {
		Logger.log.Error(err)
	}

	accumulatedValues := &metadata.AccumulatedValues{
		UniqETHTxsUsed:   [][]byte{},
		UniqBSCTxsUsed:   [][]byte{},
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

			// group portal instructions
			isCollected := portal.CollectPortalInstructions(pm, metaType, action, shardID)
			if isCollected {
				continue
			}

			switch metaType {
			case metadata.IssuingRequestMeta:
				newInst, err = blockchain.buildInstructionsForIssuingReq(beaconBestState, featureStateDB, contentStr, shardID, metaType, accumulatedValues)

			case metadata.IssuingETHRequestMeta:
				var uniqTx []byte
				newInst, uniqTx, err = blockchain.buildInstructionsForIssuingBridgeReq(
					beaconBestState,
					featureStateDB,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqETHTxsUsed,
					config.Param().EthContractAddressStr,
					"",
					statedb.IsETHTxHashIssued,
				)
				if uniqTx != nil {
					accumulatedValues.UniqETHTxsUsed = append(accumulatedValues.UniqETHTxsUsed, uniqTx)
				}
			case metadata.IssuingBSCRequestMeta:
				var uniqTx []byte
				newInst, uniqTx, err = blockchain.buildInstructionsForIssuingBridgeReq(
					beaconBestState,
					featureStateDB,
					contentStr,
					shardID,
					metaType,
					accumulatedValues,
					accumulatedValues.UniqBSCTxsUsed,
					config.Param().BscContractAddressStr,
					common.BSCPrefix,
					statedb.IsBSCTxHashIssued,
				)
				if uniqTx != nil {
					accumulatedValues.UniqBSCTxsUsed = append(accumulatedValues.UniqBSCTxsUsed, uniqTx)
				}
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
	// include portal v3, portal relaying header chain
	portalInsts, err := blockchain.handlePortalInsts(
		featureStateDB,
		beaconHeight-1,
		currentPortalStateV3,
		relayingHeaderState,
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
