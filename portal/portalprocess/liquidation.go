package portalprocess

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/incognitochain/incognito-chain/basemeta"
	"github.com/incognitochain/incognito-chain/portal"
	pCommon "github.com/incognitochain/incognito-chain/portal/common"
	pMeta "github.com/incognitochain/incognito-chain/portal/metadata"
	"math/big"
	"sort"
	"strconv"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/wallet"
)

func buildTopPercentileExchangeRatesLiquidationInst(
	custodianAddress string,
	metaType int,
	status string,
	topPercentile map[string]pMeta.LiquidateTopPercentileExchangeRatesDetail,
	remainUnlockAmounts map[string]uint64,
) []string {
	tpContent := pMeta.PortalLiquidateTopPercentileExchangeRatesContent{
		CustodianAddress:   custodianAddress,
		MetaType:           metaType,
		Status:             status,
		TP:                 topPercentile,
		RemainUnlockAmount: remainUnlockAmounts,
	}
	tpContentBytes, _ := json.Marshal(tpContent)
	return []string{
		strconv.Itoa(metaType),
		"-1",
		status,
		string(tpContentBytes),
	}
}



//// CheckAndBuildInstForCustodianLiquidation checks and builds liquidation instructions
//// when custodians didn't return public token to users after timeout
//func CheckAndBuildInstForCustodianLiquidation(
//	bc basemeta.ChainRetriever,
//	beaconHeight uint64,
//	shardHeights map[byte]uint64,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams,
//) ([][]string, error) {
//	insts := [][]string{}
//
//	// get exchange rate
//	exchangeRate := currentPortalState.FinalExchangeRatesState
//	if exchangeRate == nil {
//		Logger.log.Errorf("[CheckAndBuildInstForCustodianLiquidation] Error when get exchange rate")
//		return insts, nil
//	}
//
//	liquidatedByExchangeRate := false
//
//	sortedMatchedRedeemReqKeys := make([]string, 0)
//	for key := range currentPortalState.MatchedRedeemRequests {
//		sortedMatchedRedeemReqKeys = append(sortedMatchedRedeemReqKeys, key)
//	}
//	sort.Strings(sortedMatchedRedeemReqKeys)
//	for _, redeemReqKey := range sortedMatchedRedeemReqKeys {
//		redeemReq := currentPortalState.MatchedRedeemRequests[redeemReqKey]
//		if bc.CheckBlockTimeIsReached(beaconHeight, redeemReq.GetBeaconHeight(), shardHeights[redeemReq.ShardID()], redeemReq.ShardHeight(), portalParams.TimeOutCustodianReturnPubToken) {
//			// get shardId of redeemer
//			redeemerKey, err := wallet.Base58CheckDeserialize(redeemReq.GetRedeemerAddress())
//			if err != nil {
//				Logger.log.Errorf("[CheckAndBuildInstForCustodianLiquidation] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
//					redeemReq.GetUniqueRedeemID(), err)
//				continue
//			}
//			shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])
//
//			// get tokenID from redeemTokenID
//			tokenID := redeemReq.GetTokenID()
//			metaType := basemeta.PortalLiquidateCustodianMeta
//			unlockCollateralsForUser := make(map[string]uint64)
//
//			liquidatedCustodians := make([]*statedb.MatchingRedeemCustodianDetail, 0)
//			for _, matchCusDetail := range redeemReq.GetCustodians() {
//				//Logger.log.Errorf("matchCusDetail.GetIncognitoAddress(): %v\n", matchCusDetail.GetIncognitoAddress())
//				custodianStateKey := statedb.GenerateCustodianStateObjectKey(matchCusDetail.GetIncognitoAddress()).String()
//				//Logger.log.Errorf("custodianStateKey: %v\n", custodianStateKey)
//
//				// determine meta type
//				lockedCollaterals := currentPortalState.CustodianPoolState[custodianStateKey].GetLockedTokenCollaterals()
//				if lockedCollaterals != nil && len(lockedCollaterals[tokenID]) != 0 {
//					metaType = basemeta.PortalLiquidateCustodianMetaV3
//				}
//
//				// calculate liquidated amount and remain unlocked amount for custodian
//				var liquidatedAmount, remainUnlockAmount uint64
//				var liquidatedAmounts, remainUnlockAmounts map[string]uint64
//				if metaType == basemeta.PortalLiquidateCustodianMeta {
//					// return value in prv
//					liquidatedAmount, remainUnlockAmount, err = CalUnlockCollateralAmountAfterLiquidation(
//						currentPortalState,
//						custodianStateKey,
//						matchCusDetail.GetAmount(),
//						tokenID,
//						exchangeRate,
//						portalParams)
//				} else {
//					// return value in usdt
//					liquidatedAmount, remainUnlockAmount, liquidatedAmounts, remainUnlockAmounts, err = CalUnlockCollateralAmountAfterLiquidationV3(
//						currentPortalState,
//						custodianStateKey,
//						matchCusDetail.GetAmount(),
//						tokenID,
//						portalParams)
//					if len(liquidatedAmounts) > 0 {
//						for i, v := range liquidatedAmounts {
//							unlockCollateralsForUser[i] += v
//						}
//					}
//				}
//				if err != nil {
//					Logger.log.Errorf("[CheckAndBuildInstForCustodianLiquidation] Error when calculating unlock collateral amount %v\n: ", err)
//					inst := buildCustodianRunAwayLiquidationInst(
//						redeemReq.GetUniqueRedeemID(),
//						redeemReq.GetTokenID(),
//						matchCusDetail.GetAmount(),
//						0,
//						0,
//						liquidatedAmounts,
//						remainUnlockAmounts,
//						redeemReq.GetRedeemerAddress(),
//						matchCusDetail.GetIncognitoAddress(),
//						liquidatedByExchangeRate,
//						metaType,
//						shardID,
//						common.PortalLiquidateCustodianFailedChainStatus,
//					)
//					insts = append(insts, inst)
//					continue
//				}
//
//				// update custodian state
//				custodianState := currentPortalState.CustodianPoolState[custodianStateKey]
//				if metaType == basemeta.PortalLiquidateCustodianMeta {
//					err = updateCustodianStateAfterLiquidateCustodian(custodianState, liquidatedAmount, remainUnlockAmount, tokenID)
//				} else {
//					err = updateCustodianStateAfterLiquidateCustodianV3(custodianState, liquidatedAmount, remainUnlockAmount, liquidatedAmounts, remainUnlockAmounts, tokenID)
//				}
//				if err != nil {
//					Logger.log.Errorf("[CheckAndBuildInstForCustodianLiquidation] Error when updating custodian state %v\n: ", err)
//					inst := buildCustodianRunAwayLiquidationInst(
//						redeemReq.GetUniqueRedeemID(),
//						redeemReq.GetTokenID(),
//						matchCusDetail.GetAmount(),
//						liquidatedAmount,
//						remainUnlockAmount,
//						liquidatedAmounts,
//						remainUnlockAmounts,
//						redeemReq.GetRedeemerAddress(),
//						matchCusDetail.GetIncognitoAddress(),
//						liquidatedByExchangeRate,
//						metaType,
//						shardID,
//						common.PortalLiquidateCustodianFailedChainStatus,
//					)
//					insts = append(insts, inst)
//					continue
//				}
//
//				// remove matching custodian from matching custodians list in waiting redeem request
//				liquidatedCustodians = append(liquidatedCustodians, matchCusDetail)
//
//				// build instruction
//				inst := buildCustodianRunAwayLiquidationInst(
//					redeemReq.GetUniqueRedeemID(),
//					redeemReq.GetTokenID(),
//					matchCusDetail.GetAmount(),
//					liquidatedAmount,
//					remainUnlockAmount,
//					liquidatedAmounts,
//					remainUnlockAmounts,
//					redeemReq.GetRedeemerAddress(),
//					matchCusDetail.GetIncognitoAddress(),
//					liquidatedByExchangeRate,
//					metaType,
//					shardID,
//					common.PortalLiquidateCustodianSuccessChainStatus,
//				)
//				insts = append(insts, inst)
//			}
//
//			// create proof to liquidate runaway custodian
//			if metaType == basemeta.PortalLiquidateCustodianMetaV3 && len(unlockCollateralsForUser) > 0 {
//				liquidatedBigIntAmounts := make(map[string]*big.Int)
//				for tokenLiquidateId, tokenAmount := range unlockCollateralsForUser {
//					amountBN := big.NewInt(0).SetUint64(tokenAmount)
//					if bytes.Equal(common.FromHex(tokenLiquidateId), common.FromHex(common.EthAddrStr)) {
//						// Convert Gwei to Wei for Ether
//						amountBN = amountBN.Mul(amountBN, big.NewInt(1000000000))
//					}
//					liquidatedBigIntAmounts[tokenLiquidateId] = amountBN
//				}
//				confirmInst := buildConfirmWithdrawCollateralInstV3(
//					basemeta.PortalLiquidateRunAwayCustodianConfirmMetaV3,
//					shardID,
//					redeemReq.GetRedeemerAddress(),
//					redeemReq.GetRedeemerExternalAddress(),
//					liquidatedBigIntAmounts,
//					redeemReq.GetTxReqID(),
//					beaconHeight+1,
//				)
//				insts = append(insts, confirmInst)
//			}
//			updatedCustodians := currentPortalState.MatchedRedeemRequests[redeemReqKey].GetCustodians()
//			for _, cus := range liquidatedCustodians {
//				updatedCustodians, _ = removeCustodianFromMatchingRedeemCustodians(
//					updatedCustodians, cus.GetIncognitoAddress())
//			}
//
//			// remove redeem request from waiting redeem requests list
//			currentPortalState.MatchedRedeemRequests[redeemReqKey].SetCustodians(updatedCustodians)
//			if len(currentPortalState.MatchedRedeemRequests[redeemReqKey].GetCustodians()) == 0 {
//				deleteMatchedRedeemRequest(currentPortalState, redeemReqKey)
//			}
//		}
//	}
//
//	return insts, nil
//}

//func CheckAndBuildInstForExpiredWaitingPortingRequest(
//	bc basemeta.ChainRetriever,
//	beaconHeight uint64,
//	shardHeights map[byte]uint64,
//	currentPortalState *CurrentPortalState,
//	portalParams portal.PortalParams,
//) ([][]string, error) {
//	insts := [][]string{}
//	sortedWaitingPortingReqKeys := make([]string, 0)
//	for key := range currentPortalState.WaitingPortingRequests {
//		sortedWaitingPortingReqKeys = append(sortedWaitingPortingReqKeys, key)
//	}
//	sort.Strings(sortedWaitingPortingReqKeys)
//	for _, portingReqKey := range sortedWaitingPortingReqKeys {
//		portingReq := currentPortalState.WaitingPortingRequests[portingReqKey]
//		if bc.CheckBlockTimeIsReached(beaconHeight, portingReq.BeaconHeight(), shardHeights[portingReq.ShardID()], portingReq.ShardHeight(), portalParams.TimeOutWaitingPortingRequest) {
//			inst, err := buildInstForExpiredPortingReqByPortingID(
//				beaconHeight, currentPortalState, portingReqKey, portingReq, false)
//			if err != nil {
//				Logger.log.Errorf("[checkAndBuildInstForExpiredWaitingPortingRequest] Error when build instruction for expired porting request %v\n", err)
//				continue
//			}
//			insts = append(insts, inst...)
//		}
//	}
//
//	return insts, nil
//}

func checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	liquidatedCustodianState *statedb.CustodianState,
	tokenID string,
	portalParams portal.PortalParams,
) ([][]string, error) {
	insts := [][]string{}

	sortedWaitingRedeemReqKeys := make([]string, 0)
	for key := range currentPortalState.WaitingRedeemRequests {
		sortedWaitingRedeemReqKeys = append(sortedWaitingRedeemReqKeys, key)
	}
	sort.Strings(sortedWaitingRedeemReqKeys)
	for _, redeemReqKey := range sortedWaitingRedeemReqKeys {
		redeemReq := currentPortalState.WaitingRedeemRequests[redeemReqKey]
		if redeemReq.GetTokenID() != tokenID {
			continue
		}
		for _, matchCustodian := range redeemReq.GetCustodians() {
			if matchCustodian.GetIncognitoAddress() != liquidatedCustodianState.GetIncognitoAddress() {
				continue
			}

			// reject waiting redeem request, return ptoken and redeem fee for users
			// update custodian state (return holding public token amount)
			err := UpdateCustodianStateAfterRejectRedeemRequestByLiquidation(currentPortalState, redeemReq, beaconHeight)
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate] Error when updating custodian state %v - RedeemID %v\n: ",
					err, redeemReq.GetUniqueRedeemID())
				break
			}

			// remove redeem request from waiting redeem requests list
			deleteWaitingRedeemRequest(currentPortalState, redeemReqKey)

			// get shardId of redeemer
			redeemerKey, err := wallet.Base58CheckDeserialize(redeemReq.GetRedeemerAddress())
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
					redeemReq.GetUniqueRedeemID(), err)
				break
			}
			shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

			// build instruction
			inst := buildRedeemRequestInst(
				redeemReq.GetUniqueRedeemID(),
				redeemReq.GetTokenID(),
				redeemReq.GetRedeemAmount(),
				redeemReq.GetRedeemerAddress(),
				redeemReq.GetRedeemerRemoteAddress(),
				redeemReq.GetRedeemFee(),
				redeemReq.GetCustodians(),
				basemeta.PortalRedeemRequestMetaV3,
				shardID,
				common.Hash{},
				common.PortalRedeemReqCancelledByLiquidationChainStatus,
				redeemReq.ShardHeight(),
				redeemReq.GetRedeemerExternalAddress(),
			)
			insts = append(insts, inst)
			break
		}
	}

	return insts, nil
}


/*
Top percentile (TP): 150 (TP150), 130 (TP130), 120 (TP120)
if TP down, we are need liquidation custodian and notify to custodians (or users)
*/
func buildInstForLiquidationTopPercentileExchangeRates(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams) ([][]string, error) {
	if len(currentPortalState.CustodianPoolState) <= 0 {
		return [][]string{}, nil
	}

	insts := [][]string{}
	exchangeRate := currentPortalState.FinalExchangeRatesState
	if exchangeRate == nil {
		Logger.log.Errorf("Final exchange rate is empty")
		return [][]string{}, nil
	}

	custodianPoolState := currentPortalState.CustodianPoolState
	sortedCustodianStateKeys := make([]string, 0)
	for key := range custodianPoolState {
		sortedCustodianStateKeys = append(sortedCustodianStateKeys, key)
	}
	sort.Strings(sortedCustodianStateKeys)

	for _, custodianKey := range sortedCustodianStateKeys {
		custodianState := custodianPoolState[custodianKey]
		tpRatios, err := calAndCheckTPRatio(currentPortalState, custodianState, exchangeRate, portalParams)
		if err != nil {
			Logger.log.Errorf("Error when calculating and checking tp ratio %v", err)
			continue
		}

		liquidationRatios := map[string]pMeta.LiquidateTopPercentileExchangeRatesDetail{}

		// reject waiting redeem requests that matching with liquidated custodians
		if len(tpRatios) > 0 {
			sortedTPRatioKeys := make([]string, 0)
			for key := range tpRatios {
				sortedTPRatioKeys = append(sortedTPRatioKeys, key)
			}
			sort.Strings(sortedTPRatioKeys)
			for _, pTokenID := range sortedTPRatioKeys {
				tpRatioDetail := tpRatios[pTokenID]
				if tpRatioDetail.HoldAmountFreeCollateral > 0 {
					// check and build instruction for waiting redeem request
					instsFromRedeemRequest, err := checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate(
						beaconHeight,
						currentPortalState,
						custodianState,
						pTokenID,
						portalParams,
					)
					if err != nil {
						Logger.log.Errorf("Error when check and build instruction from redeem request %v\n", err)
						continue
					}
					if len(instsFromRedeemRequest) > 0 {
						Logger.log.Infof("There is % tpRatioDetail instructions for tp exchange rate for redeem request", len(instsFromRedeemRequest))
						insts = append(insts, instsFromRedeemRequest...)
					}
				}
			}

			remainUnlockAmounts := map[string]uint64{}
			for _, pTokenID := range sortedTPRatioKeys {
				if tpRatios[pTokenID].TPKey == int(portalParams.TP130) {
					liquidationRatios[pTokenID] = tpRatios[pTokenID]
					continue
				}

				liquidatedPubToken := GetTotalHoldPubTokenAmountExcludeMatchedRedeemReqs(currentPortalState, custodianState, pTokenID)
				if liquidatedPubToken <= 0 {
					continue
				}

				// calculate liquidated amount and remain unlocked amount for custodian
				liquidatedAmountInPRV, remainUnlockAmount, err := CalUnlockCollateralAmountAfterLiquidation(
					currentPortalState,
					custodianKey,
					liquidatedPubToken,
					pTokenID,
					exchangeRate,
					portalParams)
				if err != nil {
					Logger.log.Errorf("Error when calculating unlock collateral amount %v - tokenID %v - Custodian address %v\n",
						err, pTokenID, custodianState.GetIncognitoAddress())
					continue
				}

				remainUnlockAmounts[pTokenID] += remainUnlockAmount
				liquidationRatios[pTokenID] = pMeta.LiquidateTopPercentileExchangeRatesDetail{
					TPKey:                    tpRatios[pTokenID].TPKey,
					TPValue:                  tpRatios[pTokenID].TPValue,
					HoldAmountFreeCollateral: liquidatedAmountInPRV,
					HoldAmountPubToken:       liquidatedPubToken,
				}
			}

			if len(liquidationRatios) > 0 {
				//update current portal state
				updateCurrentPortalStateOfLiquidationExchangeRates(currentPortalState, custodianKey, custodianState, liquidationRatios, remainUnlockAmounts)
				inst := buildTopPercentileExchangeRatesLiquidationInst(
					custodianState.GetIncognitoAddress(),
					basemeta.PortalLiquidateTPExchangeRatesMeta,
					common.PortalLiquidateTPExchangeRatesSuccessChainStatus,
					liquidationRatios,
					remainUnlockAmounts,
				)
				insts = append(insts, inst)
			}
		}
	}

	return insts, nil
}

func BuildInstForLiquidationByExchangeRatesV3(
	bc basemeta.ChainRetriever,
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams) ([][]string, error) {
	if currentPortalState == nil {
		Logger.log.Errorf("[LIQUIDATIONBYRATES] Current portal state is null")
		return [][]string{}, nil
	}
	if len(currentPortalState.CustodianPoolState) == 0 {
		return [][]string{}, nil
	}
	exchangeRate := currentPortalState.FinalExchangeRatesState
	if exchangeRate == nil {
		Logger.log.Errorf("[LIQUIDATIONBYRATES] Final exchange rate is empty")
		return [][]string{}, nil
	}

	insts := [][]string{}
	custodianPoolState := currentPortalState.CustodianPoolState
	sortedCustodianStateKeys := make([]string, 0)
	for key := range custodianPoolState {
		sortedCustodianStateKeys = append(sortedCustodianStateKeys, key)
	}
	sort.Strings(sortedCustodianStateKeys)

	for _, custodianKey := range sortedCustodianStateKeys {
		custodianState := custodianPoolState[custodianKey]
		tpRatios, remainUnlockColalterals, rejectedWRedeemIDs, err := calAndCheckLiquidationRatioV3(currentPortalState, custodianState, exchangeRate, portalParams)
		if err != nil {
			Logger.log.Errorf("Error when calculating and checking tp ratio %v", err)
			continue
		}

		if len(tpRatios) == 0 {
			continue
		}

		// reject waiting redeem requests that matching with liquidated custodians
		for _, wRedeemId := range rejectedWRedeemIDs {
			inst, err := buildInstRejectRedeemRequestByLiquidationExchangeRate(beaconHeight, currentPortalState, wRedeemId)
			if err != nil {
				Logger.log.Errorf("[LIQUIDATIONBYRATES] Error when building instruction reject redeem request ID %v - %v", wRedeemId, err)
				continue
			}
			insts = append(insts, inst)
		}

		// update current portal state after liquidation custodianKey
		updateCurrentPortalStateAfterLiquidationByRatesV3(currentPortalState, custodianKey, tpRatios, remainUnlockColalterals)
		inst := buildLiquidationByExchangeRateInstV3(
			custodianState.GetIncognitoAddress(),
			basemeta.PortalLiquidateByRatesMetaV3,
			common.PortalLiquidateTPExchangeRatesSuccessChainStatus,
			tpRatios,
			remainUnlockColalterals)
		insts = append(insts, inst)
	}

	return insts, nil
}

/* =======
Portal Redeem From Liquidation Pool Processor
======= */

type portalRedeemFromLiquidationPoolProcessor struct {
	*portalInstProcessor
}

func (p *portalRedeemFromLiquidationPoolProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalRedeemFromLiquidationPoolProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRedeemFromLiquidationPoolProcessor) prepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildRedeemFromLiquidationPoolInst(
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
	totalPTokenReceived uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	redeemRequestContent := pMeta.PortalRedeemLiquidateExchangeRatesContent{
		TokenID:               tokenID,
		RedeemAmount:          redeemAmount,
		RedeemerIncAddressStr: incAddressStr,
		TxReqID:               txReqID,
		ShardID:               shardID,
		TotalPTokenReceived:   totalPTokenReceived,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

func (p *portalRedeemFromLiquidationPoolProcessor) buildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal redeem liquidate exchange rate action: %+v", err)
		return [][]string{}, nil
	}
	var actionData pMeta.PortalRedeemLiquidateExchangeRatesAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal redeem liquidate exchange rate action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildRedeemFromLiquidationPoolInst(
		meta.TokenID,
		meta.RedeemAmount,
		meta.RedeemerIncAddressStr,
		0,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		common.PortalRedeemFromLiquidationPoolRejectedChainStatus,
	)

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null.")
		// need to mint ptoken to user
		return [][]string{rejectInst}, nil
	}

	//get exchange rates
	exchangeRatesState := currentPortalState.FinalExchangeRatesState
	if exchangeRatesState == nil {
		Logger.log.Errorf("exchange rates not found")
		return [][]string{rejectInst}, nil
	}

	//check redeem amount
	liquidateExchangeRatesKey := statedb.GeneratePortalLiquidationPoolObjectKey()
	liquidateExchangeRates, ok := currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()]

	if !ok {
		Logger.log.Errorf("Liquidate exchange rates not found")
		return [][]string{rejectInst}, nil
	}

	liquidateByTokenID, ok := liquidateExchangeRates.Rates()[meta.TokenID]

	if !ok {
		Logger.log.Errorf("Liquidate exchange rates not found")
		return [][]string{rejectInst}, nil
	}

	totalPrv, err := calTotalLiquidationByExchangeRates(meta.RedeemAmount, liquidateByTokenID)

	if err != nil {
		Logger.log.Errorf("Calculate total liquidation error %v", err)
		return [][]string{rejectInst}, nil
	}

	if totalPrv > liquidateByTokenID.CollateralAmount || liquidateByTokenID.CollateralAmount <= 0 {
		Logger.log.Errorf("amout free collateral not enough, need prv %v != hold amount free collateral %v", totalPrv, liquidateByTokenID.CollateralAmount)
		return [][]string{rejectInst}, nil
	}

	liquidateExchangeRates.Rates()[meta.TokenID] = statedb.LiquidationPoolDetail{
		CollateralAmount: liquidateByTokenID.CollateralAmount - totalPrv,
		PubTokenAmount:   liquidateByTokenID.PubTokenAmount - meta.RedeemAmount,
	}

	currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()] = liquidateExchangeRates

	inst := buildRedeemFromLiquidationPoolInst(
		meta.TokenID,
		meta.RedeemAmount,
		meta.RedeemerIncAddressStr,
		totalPrv,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		common.PortalRedeemFromLiquidationPoolSuccessChainStatus,
	)
	return [][]string{inst}, nil
}

func (p *portalRedeemFromLiquidationPoolProcessor) processInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData pMeta.PortalRedeemLiquidateExchangeRatesContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalRedeemFromLiquidationPoolSuccessChainStatus {
		liquidateExchangeRatesKey := statedb.GeneratePortalLiquidationPoolObjectKey()
		liquidateExchangeRates, ok := currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()]

		if !ok {
			Logger.log.Errorf("Liquidate exchange rates not found")
			return nil
		}

		liquidateByTokenID, ok := liquidateExchangeRates.Rates()[actionData.TokenID]
		if !ok {
			Logger.log.Errorf("Liquidate exchange rates not found")
			return nil
		}

		totalPrv := actionData.TotalPTokenReceived

		liquidateExchangeRates.Rates()[actionData.TokenID] = statedb.LiquidationPoolDetail{
			CollateralAmount: liquidateByTokenID.CollateralAmount - totalPrv,
			PubTokenAmount:   liquidateByTokenID.PubTokenAmount - actionData.RedeemAmount,
		}

		currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()] = liquidateExchangeRates

		Logger.log.Infof("Redeem Liquidation: Amount refund to user amount ptoken %v, amount prv %v", actionData.RedeemAmount, totalPrv)

		redeem := pMeta.NewRedeemLiquidateExchangeRatesStatus(
			actionData.TxReqID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RedeemAmount,
			common.PortalRedeemFromLiquidationPoolSuccessStatus,
			totalPrv,
		)

		contentStatusBytes, _ := json.Marshal(redeem)
		err = statedb.StoreRedeemRequestFromLiquidationPoolByTxIDStatus(
			stateDB,
			actionData.TxReqID.String(),
			contentStatusBytes,
		)

		if err != nil {
			Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
			return nil
		}

		// update bridge/portal token info
		incTokenID, err := common.Hash{}.NewHashFromStr(actionData.TokenID)
		if err != nil {
			Logger.log.Errorf("ERROR: Can not new hash from porting incTokenID: %+v", err)
			return nil
		}
		updatingInfo, found := updatingInfoByTokenID[*incTokenID]
		if found {
			updatingInfo.DeductAmt += actionData.RedeemAmount
		} else {
			updatingInfo = basemeta.UpdatingInfo{
				CountUpAmt:      0,
				DeductAmt:       actionData.RedeemAmount,
				TokenID:         *incTokenID,
				ExternalTokenID: nil,
				IsCentralized:   false,
			}
		}
		updatingInfoByTokenID[*incTokenID] = updatingInfo
	} else if reqStatus == common.PortalRedeemFromLiquidationPoolRejectedChainStatus {
		redeem := pMeta.NewRedeemLiquidateExchangeRatesStatus(
			actionData.TxReqID,
			actionData.TokenID,
			actionData.RedeemerIncAddressStr,
			actionData.RedeemAmount,
			common.PortalRedeemFromLiquidationPoolRejectedStatus,
			0,
		)

		contentStatusBytes, _ := json.Marshal(redeem)
		err = statedb.StoreRedeemRequestFromLiquidationPoolByTxIDStatus(
			stateDB,
			actionData.TxReqID.String(),
			contentStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
			return nil
		}
	}

	return nil
}

type portalRedeemFromLiquidationPoolProcessorV3 struct {
	*portalInstProcessor
}

func (p *portalRedeemFromLiquidationPoolProcessorV3) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalRedeemFromLiquidationPoolProcessorV3) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalRedeemFromLiquidationPoolProcessorV3) prepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildRedeemFromLiquidationPoolInstV3(
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
	extAddressStr string,
	mintedPRVCollateral uint64,
	unlockedTokenCollaterals map[string]uint64,
	metaType int,
	shardID byte,
	txReqID common.Hash,
	status string,
) []string {
	redeemRequestContent := pMeta.PortalRedeemFromLiquidationPoolContentV3{
		TokenID:                  tokenID,
		RedeemAmount:             redeemAmount,
		RedeemerIncAddressStr:    incAddressStr,
		RedeemerExtAddressStr:    extAddressStr,
		TxReqID:                  txReqID,
		ShardID:                  shardID,
		MintedPRVCollateral:      mintedPRVCollateral,
		UnlockedTokenCollaterals: unlockedTokenCollaterals,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

func (p *portalRedeemFromLiquidationPoolProcessorV3) buildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	Logger.log.Errorf("===================== Starting producer redeem from liquidation pool v3 ....")
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal redeem liquidate exchange rate action: %+v", err)
		return [][]string{}, nil
	}
	var actionData pMeta.PortalRedeemFromLiquidationPoolActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal redeem liquidate exchange rate action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildRedeemFromLiquidationPoolInstV3(
		meta.TokenID,
		meta.RedeemAmount,
		meta.RedeemerIncAddressStr,
		meta.RedeemerExtAddressStr,
		0,
		map[string]uint64{},
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		common.PortalRedeemFromLiquidationPoolRejectedChainStatus,
	)

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null.")
		// need to mint ptoken to user
		return [][]string{rejectInst}, nil
	}

	//get exchange rates
	exchangeRatesState := currentPortalState.FinalExchangeRatesState
	if exchangeRatesState == nil {
		Logger.log.Errorf("exchange rates not found")
		return [][]string{rejectInst}, nil
	}
	exchangeTool := NewPortalExchangeRateTool(exchangeRatesState, portalParams.SupportedCollateralTokens)

	// check liquidation pool
	liquidateExchangeRatesKey := statedb.GeneratePortalLiquidationPoolObjectKey()
	liquidateExchangeRates, ok := currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()]
	if !ok || liquidateExchangeRates == nil || liquidateExchangeRates.Rates() == nil {
		Logger.log.Errorf("Liquidation pool not found")
		return [][]string{rejectInst}, nil
	}

	liquidationInfoByPortalTokenID, ok := liquidateExchangeRates.Rates()[meta.TokenID]
	if !ok || liquidationInfoByPortalTokenID.PubTokenAmount == 0 {
		Logger.log.Errorf("Liquidation for portalTokenID %v is empty", meta.TokenID)
		return [][]string{rejectInst}, nil
	}

	// calculate minted PRV collateral and unlocked token collaterals from liquidation pool
	mintedPRVCollateral, unlockedTokenCollaterals, err := calUnlockedCollateralRedeemFromLiquidationPoolV3(meta.RedeemAmount, liquidationInfoByPortalTokenID, *exchangeTool)
	if err != nil {
		Logger.log.Errorf("Calculate total liquidation error %v", err)
		return [][]string{rejectInst}, nil
	}

	// update liquidation pool
	UpdateLiquidationPoolAfterRedeemFrom(
		currentPortalState, liquidateExchangeRates, meta.TokenID, meta.RedeemAmount,
		mintedPRVCollateral, unlockedTokenCollaterals)

	inst := buildRedeemFromLiquidationPoolInstV3(
		meta.TokenID,
		meta.RedeemAmount,
		meta.RedeemerIncAddressStr,
		meta.RedeemerExtAddressStr,
		mintedPRVCollateral,
		unlockedTokenCollaterals,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		common.PortalRedeemFromLiquidationPoolSuccessChainStatus,
	)
	insts := [][]string{inst}

	if len(unlockedTokenCollaterals) > 0 {
		unlockedTokens := map[string]*big.Int{}
		for tokenID, amount := range unlockedTokenCollaterals {
			amountBN := big.NewInt(0).SetUint64(amount)
			// Convert amount to big.Int to get bytes later
			if bytes.Equal(common.FromHex(tokenID), common.FromHex(common.EthAddrStr)) {
				// Convert Gwei to Wei for Ether
				amountBN = amountBN.Mul(amountBN, big.NewInt(1000000000))
			}
			unlockedTokens[tokenID] = amountBN
		}

		confirmInst := buildConfirmWithdrawCollateralInstV3(
			basemeta.PortalRedeemFromLiquidationPoolConfirmMetaV3,
			shardID,
			meta.RedeemerIncAddressStr,
			meta.RedeemerExtAddressStr,
			unlockedTokens,
			actionData.TxReqID,
			beaconHeight+1,
		)
		insts = append(insts, confirmInst)
	}

	Logger.log.Errorf("===================== Build instructions for producer redeem from liquidation pool v3 successfully....")
	Logger.log.Errorf("insts: %+v, %+v", insts)
	return insts, nil
}

func (p *portalRedeemFromLiquidationPoolProcessorV3) processInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}

	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData pMeta.PortalRedeemFromLiquidationPoolContentV3
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]

	status := byte(common.PortalRedeemFromLiquidationPoolRejectedStatus)
	if reqStatus == common.PortalRedeemFromLiquidationPoolSuccessChainStatus {
		liquidateExchangeRatesKey := statedb.GeneratePortalLiquidationPoolObjectKey()
		liquidateExchangeRates := currentPortalState.LiquidationPool[liquidateExchangeRatesKey.String()]

		UpdateLiquidationPoolAfterRedeemFrom(
			currentPortalState, liquidateExchangeRates, actionData.TokenID, actionData.RedeemAmount,
			actionData.MintedPRVCollateral, actionData.UnlockedTokenCollaterals)
		status = byte(common.PortalRedeemFromLiquidationPoolSuccessStatus)

		// update bridge/portal token info
		incTokenID, err := common.Hash{}.NewHashFromStr(actionData.TokenID)
		if err != nil {
			Logger.log.Errorf("ERROR: Can not new hash from porting incTokenID: %+v", err)
			return nil
		}
		updatingInfo, found := updatingInfoByTokenID[*incTokenID]
		if found {
			updatingInfo.DeductAmt += actionData.RedeemAmount
		} else {
			updatingInfo = basemeta.UpdatingInfo{
				CountUpAmt:      0,
				DeductAmt:       actionData.RedeemAmount,
				TokenID:         *incTokenID,
				ExternalTokenID: nil,
				IsCentralized:   false,
			}
		}
		updatingInfoByTokenID[*incTokenID] = updatingInfo
	}

	// store db status
	redeem := pMeta.PortalRedeemFromLiquidationPoolStatusV3{
		TokenID:                  actionData.TokenID,
		RedeemAmount:             actionData.RedeemAmount,
		RedeemerIncAddressStr:    actionData.RedeemerIncAddressStr,
		RedeemerExtAddressStr:    actionData.RedeemerExtAddressStr,
		TxReqID:                  actionData.TxReqID,
		MintedPRVCollateral:      actionData.MintedPRVCollateral,
		UnlockedTokenCollaterals: actionData.UnlockedTokenCollaterals,
		Status:                   status,
	}

	contentStatusBytes, _ := json.Marshal(redeem)
	err = statedb.StoreRedeemRequestFromLiquidationPoolByTxIDStatusV3(
		stateDB,
		actionData.TxReqID.String(),
		contentStatusBytes,
	)

	if err != nil {
		Logger.log.Errorf("Store redeem liquidate exchange rates error %v\n", err)
		return nil
	}

	return nil
}

/* =======
Portal Custodian Topup Processor
======= */

type portalCustodianTopupProcessor struct {
	*portalInstProcessor
}

func (p *portalCustodianTopupProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalCustodianTopupProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalCustodianTopupProcessor) prepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildPortalCustodianTopupInst(
	pTokenId string,
	incogAddress string,
	depositedAmount uint64,
	freeCollateralAmount uint64,
	status string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	redeemRequestContent := pMeta.PortalLiquidationCustodianDepositContentV2{
		PTokenId:             pTokenId,
		IncogAddressStr:      incogAddress,
		DepositedAmount:      depositedAmount,
		FreeCollateralAmount: freeCollateralAmount,
		TxReqID:              txReqID,
		ShardID:              shardID,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

func (p *portalCustodianTopupProcessor) buildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal liquidation custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData pMeta.PortalLiquidationCustodianDepositActionV2
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal liquidation custodian deposit action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildPortalCustodianTopupInst(
		meta.PTokenId,
		meta.IncogAddressStr,
		meta.DepositedAmount,
		meta.FreeCollateralAmount,
		common.PortalCustodianTopupRejectedChainStatus,
		meta.Type,
		shardID,
		actionData.TxReqID,
	)

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	custodianStateKey := statedb.GenerateCustodianStateObjectKey(meta.IncogAddressStr)
	custodian, ok := currentPortalState.CustodianPoolState[custodianStateKey.String()]
	if !ok {
		Logger.log.Errorf("Custodian not found")
		return [][]string{rejectInst}, nil
	}

	lockedAmountCollateral := custodian.GetLockedAmountCollateral()
	if _, ok := lockedAmountCollateral[meta.PTokenId]; !ok {
		Logger.log.Errorf("PToken not found")
		return [][]string{rejectInst}, nil
	}

	totalHoldPubTokenAmount := GetTotalHoldPubTokenAmount(currentPortalState, custodian, meta.PTokenId)
	if totalHoldPubTokenAmount <= 0 {
		Logger.log.Errorf("Holding public token amount is zero, don't need to top up")
		return [][]string{rejectInst}, nil
	}

	if meta.FreeCollateralAmount > custodian.GetFreeCollateral() {
		Logger.log.Errorf("Free collateral topup amount is greater than free collateral of custodian's state")
		return [][]string{rejectInst}, nil
	}

	_, err = UpdateCustodianAfterTopup(currentPortalState, custodian, meta.PTokenId, meta.DepositedAmount, meta.FreeCollateralAmount, common.PRVIDStr)
	if err != nil {
		Logger.log.Errorf("Update custodians state error : %+v", err)
		return [][]string{rejectInst}, nil
	}
	inst := buildPortalCustodianTopupInst(
		meta.PTokenId,
		meta.IncogAddressStr,
		meta.DepositedAmount,
		meta.FreeCollateralAmount,
		common.PortalCustodianTopupSuccessChainStatus,
		meta.Type,
		shardID,
		actionData.TxReqID,
	)
	return [][]string{inst}, nil
}

func (p *portalCustodianTopupProcessor) processInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData pMeta.PortalLiquidationCustodianDepositContentV2
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Error when unmarshaling portal liquidation custodian deposit content %v - %v", instructions[3], err)
		return nil
	}

	depositStatus := instructions[2]

	if depositStatus == common.PortalCustodianTopupSuccessChainStatus {
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		custodian, ok := currentPortalState.CustodianPoolState[custodianStateKeyStr]
		if !ok {
			Logger.log.Errorf("Custodian not found")
			return nil
		}

		_, err = UpdateCustodianAfterTopup(currentPortalState, custodian, actionData.PTokenId, actionData.DepositedAmount, actionData.FreeCollateralAmount, common.PRVIDStr)
		if err != nil {
			Logger.log.Errorf("Update custodians state error : %+v", err)
			return nil
		}

		newLiquidationCustodianDeposit := pMeta.NewLiquidationCustodianDepositStatusV2(
			actionData.TxReqID,
			actionData.IncogAddressStr,
			actionData.PTokenId,
			actionData.DepositedAmount,
			actionData.FreeCollateralAmount,
			common.PortalCustodianTopupSuccessStatus,
		)

		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
		err = statedb.StoreCustodianTopupStatus(
			stateDB,
			actionData.TxReqID.String(),
			contentStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit error %v", err)
			return nil
		}
	} else if depositStatus == common.PortalCustodianTopupRejectedChainStatus {
		newLiquidationCustodianDeposit := pMeta.NewLiquidationCustodianDepositStatusV2(
			actionData.TxReqID,
			actionData.IncogAddressStr,
			actionData.PTokenId,
			actionData.DepositedAmount,
			actionData.FreeCollateralAmount,
			common.PortalCustodianTopupRejectedStatus,
		)

		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
		err = statedb.StoreCustodianTopupStatus(
			stateDB,
			actionData.TxReqID.String(),
			contentStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit error %v", err)
			return nil
		}
	}

	return nil
}

/* =======
Portal Custodian Topup For Waiting Porting Request Processor
======= */

type portalTopupWaitingPortingReqProcessor struct {
	*portalInstProcessor
}

func (p *portalTopupWaitingPortingReqProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalTopupWaitingPortingReqProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalTopupWaitingPortingReqProcessor) prepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildTopUpWaitingPortingInst(
	portingID string,
	pTokenID string,
	incogAddress string,
	depositedAmount uint64,
	freeCollateralAmount uint64,
	status string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	topUpWaitingPortingReqContent := pMeta.PortalTopUpWaitingPortingRequestContent{
		PortingID:            portingID,
		PTokenID:             pTokenID,
		IncogAddressStr:      incogAddress,
		DepositedAmount:      depositedAmount,
		FreeCollateralAmount: freeCollateralAmount,
		TxReqID:              txReqID,
		ShardID:              shardID,
	}
	topUpWaitingPortingReqContentBytes, _ := json.Marshal(topUpWaitingPortingReqContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(topUpWaitingPortingReqContentBytes),
	}
}

func (p *portalTopupWaitingPortingReqProcessor) buildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal top up waiting porting content: %+v", err)
		return [][]string{}, nil
	}
	var actionData pMeta.PortalTopUpWaitingPortingRequestAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshaling portal top up waiting porting action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildTopUpWaitingPortingInst(
		meta.PortingID,
		meta.PTokenID,
		meta.IncogAddressStr,
		meta.DepositedAmount,
		meta.FreeCollateralAmount,
		common.PortalTopUpWaitingPortingRejectedChainStatus,
		meta.Type,
		shardID,
		actionData.TxReqID,
	)

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	custodianStateKey := statedb.GenerateCustodianStateObjectKey(meta.IncogAddressStr)
	custodian, ok := currentPortalState.CustodianPoolState[custodianStateKey.String()]
	if !ok {
		Logger.log.Errorf("Custodian not found")
		return [][]string{rejectInst}, nil
	}

	waitingPortingRequestKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(meta.PortingID)
	waitingPortingReq, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()]
	if !ok || waitingPortingReq == nil || waitingPortingReq.TokenID() != meta.PTokenID {
		Logger.log.Errorf("Waiting porting request with portingID (%s) not found", meta.PortingID)
		return [][]string{rejectInst}, nil
	}
	isMatchPorting := false
	for _, cus := range waitingPortingReq.Custodians() {
		if cus.IncAddress == meta.IncogAddressStr {
			isMatchPorting = true
			break
		}
	}
	if !isMatchPorting {
		Logger.log.Errorf("Custodian %v is not the matching custodian in portingID ", meta.IncogAddressStr, meta.PortingID)
		return [][]string{rejectInst}, nil
	}

	if meta.FreeCollateralAmount > custodian.GetFreeCollateral() {
		Logger.log.Errorf("Free collateral topup amount is greater than free collateral of custodian's state")
		return [][]string{rejectInst}, nil
	}

	err = UpdateCustodianAfterTopupWaitingPorting(currentPortalState, waitingPortingReq, custodian, meta.PTokenID, meta.DepositedAmount, meta.FreeCollateralAmount, common.PRVIDStr)
	if err != nil {
		Logger.log.Errorf("Update portal state error: %+v", err)
		return [][]string{rejectInst}, nil
	}

	inst := buildTopUpWaitingPortingInst(
		meta.PortingID,
		meta.PTokenID,
		meta.IncogAddressStr,
		meta.DepositedAmount,
		meta.FreeCollateralAmount,
		common.PortalTopUpWaitingPortingSuccessChainStatus,
		meta.Type,
		shardID,
		actionData.TxReqID,
	)
	return [][]string{inst}, nil
}

func (p *portalTopupWaitingPortingReqProcessor) processInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	var actionData pMeta.PortalTopUpWaitingPortingRequestContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Error when unmarshaling portal top up waiting porting action %v - %v", instructions[3], err)
		return nil
	}

	depositStatus := instructions[2]
	if depositStatus == common.PortalTopUpWaitingPortingRejectedChainStatus {
		topUpWaitingPortingReq := pMeta.NewPortalTopUpWaitingPortingRequestStatus(
			actionData.TxReqID,
			actionData.PortingID,
			actionData.IncogAddressStr,
			actionData.PTokenID,
			actionData.DepositedAmount,
			actionData.FreeCollateralAmount,
			common.PortalTopUpWaitingPortingRejectedStatus,
		)
		statusContentBytes, _ := json.Marshal(topUpWaitingPortingReq)
		err = statedb.StoreCustodianTopupWaitingPortingStatus(
			stateDB,
			actionData.TxReqID.String(),
			statusContentBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while storing waiting porting top up error %v", err)
		}
	} else if depositStatus == common.PortalTopUpWaitingPortingSuccessChainStatus {
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr)
		custodian, ok := currentPortalState.CustodianPoolState[custodianStateKey.String()]
		if !ok {
			Logger.log.Errorf("Custodian not found")
			return nil
		}

		waitingPortingRequestKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.PortingID)
		waitingPortingReq, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()]
		if !ok || waitingPortingReq == nil || waitingPortingReq.TokenID() != actionData.PTokenID {
			Logger.log.Errorf("Waiting porting request with portingID (%s) not found", actionData.PortingID)
			return nil
		}

		err = UpdateCustodianAfterTopupWaitingPorting(currentPortalState, waitingPortingReq, custodian, actionData.PTokenID, actionData.DepositedAmount, actionData.FreeCollateralAmount, common.PRVIDStr)
		if err != nil {
			Logger.log.Errorf("Update portal state error: %+v", err)
			return nil
		}

		topUpWaitingPortingReq := pMeta.NewPortalTopUpWaitingPortingRequestStatus(
			actionData.TxReqID,
			actionData.PortingID,
			actionData.IncogAddressStr,
			actionData.PTokenID,
			actionData.DepositedAmount,
			actionData.FreeCollateralAmount,
			common.PortalTopUpWaitingPortingSuccessStatus,
		)
		statusContentBytes, _ := json.Marshal(topUpWaitingPortingReq)
		err = statedb.StoreCustodianTopupWaitingPortingStatus(
			stateDB,
			actionData.TxReqID.String(),
			statusContentBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while storing waiting porting top up error %v", err)
			return nil
		}

		// update state of porting request by portingID
		newPortingRequestState := pMeta.NewPortingRequestStatus(
			waitingPortingReq.UniquePortingID(),
			waitingPortingReq.TxReqID(),
			waitingPortingReq.TokenID(),
			waitingPortingReq.PorterAddress(),
			waitingPortingReq.Amount(),
			waitingPortingReq.Custodians(),
			waitingPortingReq.PortingFee(),
			common.PortalPortingReqWaitingStatus,
			beaconHeight+1,
			waitingPortingReq.ShardHeight(),
			waitingPortingReq.ShardID(),
		)
		newPortingRequestStatusBytes, _ := json.Marshal(newPortingRequestState)
		err = statedb.StorePortalPortingRequestStatus(
			stateDB,
			waitingPortingReq.UniquePortingID(),
			newPortingRequestStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
			return nil
		}
	}
	return nil
}

/* =======
Portal Custodian Topup Processor v3
======= */

type portalCustodianTopupProcessorV3 struct {
	*portalInstProcessor
}

func (p *portalCustodianTopupProcessorV3) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalCustodianTopupProcessorV3) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalCustodianTopupProcessorV3) prepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian topup action v3: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while decoding content string of portal custodian topup action v3: %+v", err)
	}
	var actionData pMeta.PortalLiquidationCustodianDepositActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action v3: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action v3: %+v", err)
	}
	meta := actionData.Meta
	if meta.DepositAmount > 0 {
		// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
		// so must build unique external tx as combination of chain name and block hash and tx index.
		uniqExternalTxID := pCommon.GetUniqExternalTxID(common.ETHChainName, meta.BlockHash, meta.TxIndex)
		isSubmitted, err := statedb.IsPortalExternalTxHashSubmitted(stateDB, uniqExternalTxID)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while checking eth tx submitted: %+v", err)
			return nil, fmt.Errorf("ERROR: an error occured while checking eth tx submitted: %+v", err)
		}

		optionalData := make(map[string]interface{})
		optionalData["isSubmitted"] = isSubmitted
		optionalData["uniqExternalTxID"] = uniqExternalTxID
		return optionalData, nil
	}
	return nil, nil
}

func buildPortalCustodianTopupV3(
	incogAddress string,
	portalTokenId string,
	collateralTokenID string,
	depositedAmount uint64,
	freeCollateralAmount uint64,
	uniqExternalTxID []byte,
	status string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	redeemRequestContent := pMeta.PortalLiquidationCustodianDepositContentV3{
		IncogAddressStr:           incogAddress,
		PortalTokenID:             portalTokenId,
		CollateralTokenID:         collateralTokenID,
		DepositAmount:             depositedAmount,
		FreeTokenCollateralAmount: freeCollateralAmount,
		UniqExternalTxID:          uniqExternalTxID,
		TxReqID:                   txReqID,
		ShardID:                   shardID,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

func (p *portalCustodianTopupProcessorV3) buildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal liquidation custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData pMeta.PortalLiquidationCustodianDepositActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal liquidation custodian deposit action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildPortalCustodianTopupV3(
		meta.IncogAddressStr,
		meta.PortalTokenID,
		meta.CollateralTokenID,
		meta.DepositAmount,
		meta.FreeTokenCollateralAmount,
		nil,
		common.PortalCustodianTopupRejectedChainStatus,
		meta.Type,
		shardID,
		actionData.TxReqID,
	)
	rejectInst2 := []string{}

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	custodianStateKey := statedb.GenerateCustodianStateObjectKey(meta.IncogAddressStr)
	custodian, ok := currentPortalState.CustodianPoolState[custodianStateKey.String()]
	if !ok {
		Logger.log.Errorf("Custodian not found")
		return [][]string{rejectInst}, nil
	}

	// check total hold public tokens
	totalHoldPubTokenAmount := GetTotalHoldPubTokenAmount(currentPortalState, custodian, meta.PortalTokenID)
	if totalHoldPubTokenAmount <= 0 {
		Logger.log.Errorf("Holding public token amount is zero, don't need to top up")
		return [][]string{rejectInst}, nil
	}

	// check free token collaterals
	freeTokenCollaterals := custodian.GetFreeTokenCollaterals()
	if meta.FreeTokenCollateralAmount > 0 && freeTokenCollaterals == nil {
		Logger.log.Errorf("Free token collaterals of custodian's state is nil")
		return [][]string{rejectInst}, nil
	}
	if meta.FreeTokenCollateralAmount > custodian.GetFreeTokenCollaterals()[meta.CollateralTokenID] {
		Logger.log.Errorf("Free token collateral topup amount is greater than free token collateral of custodian's state")
		return [][]string{rejectInst}, nil
	}

	// check deposit amount and deposit proof
	uniqExternalTxID := []byte{}
	if meta.DepositAmount > 0 {
		// check uniqExternalTxID from optionalData which get from statedb
		if optionalData == nil {
			Logger.log.Errorf("Topup v3: optionalData is null")
			return [][]string{rejectInst}, nil
		}
		uniqExternalTxID, ok = optionalData["uniqExternalTxID"].([]byte)
		if !ok || len(uniqExternalTxID) == 0 {
			Logger.log.Errorf("Topup v3: optionalData uniqExternalTxID is invalid")
			return [][]string{rejectInst}, nil
		}

		// reject instruction with uniqExternalTxID
		rejectInst2 = buildPortalCustodianTopupV3(
			meta.IncogAddressStr,
			meta.PortalTokenID,
			meta.CollateralTokenID,
			meta.DepositAmount,
			meta.FreeTokenCollateralAmount,
			uniqExternalTxID,
			common.PortalCustodianTopupRejectedChainStatus,
			meta.Type,
			shardID,
			actionData.TxReqID,
		)

		isExist, ok := optionalData["isSubmitted"].(bool)
		if !ok {
			Logger.log.Errorf("Topup v3: optionalData isSubmitted is invalid")
			return [][]string{rejectInst2}, nil
		}
		if isExist {
			Logger.log.Errorf("Topup v3: Unique external id exist in db %v", uniqExternalTxID)
			return [][]string{rejectInst2}, nil
		}

		// verify proof and parse receipt
		ethReceipt, err := pCommon.VerifyProofAndParseReceipt(meta.BlockHash, meta.TxIndex, meta.ProofStrs)
		if err != nil {
			Logger.log.Errorf("Topup v3: Verify eth proof error: %+v", err)
			return [][]string{rejectInst2}, nil
		}
		if ethReceipt == nil {
			Logger.log.Errorf("Topup v3: The eth proof's receipt could not be null.")
			return [][]string{rejectInst2}, nil
		}

		logMap, err := pCommon.PickAndParseLogMapFromReceiptByContractAddr(ethReceipt, bc.GetPortalETHContractAddrStr(beaconHeight), "Deposit")
		if err != nil {
			Logger.log.Errorf("WARNING: an error occured while parsing log map from receipt: ", err)
			return [][]string{rejectInst2}, nil
		}
		if logMap == nil {
			Logger.log.Errorf("WARNING: could not find log map out from receipt")
			return [][]string{rejectInst2}, nil
		}

		// parse info from log map and validate info
		custodianIncAddr, externalTokenIDStr, depositAmount, err := pMeta.ParseInfoFromLogMap(logMap)
		if err != nil {
			Logger.log.Errorf("Topup v3: Error when parsing info from log map : %+v", err)
			return [][]string{rejectInst2}, nil
		}
		externalTokenIDStr = common.Remove0xPrefix(externalTokenIDStr)

		if externalTokenIDStr != meta.CollateralTokenID {
			Logger.log.Errorf("Topup v3: Collateral token id in meta %v is different from in deposit proof %+v", meta.CollateralTokenID, externalTokenIDStr)
			return [][]string{rejectInst2}, nil
		}

		if custodianIncAddr != meta.IncogAddressStr {
			Logger.log.Errorf("Topup v3: Custodian incognito address in meta %v is different from in deposit proof %+v", meta.IncogAddressStr, custodianIncAddr)
			return [][]string{rejectInst2}, nil
		}

		if depositAmount != meta.DepositAmount {
			Logger.log.Errorf("Topup v3: Deposit amount in meta %v is different from in deposit proof %+v", meta.DepositAmount, depositAmount)
			return [][]string{rejectInst2}, nil
		}
	}

	_, err = UpdateCustodianAfterTopup(currentPortalState, custodian, meta.PortalTokenID, meta.DepositAmount, meta.FreeTokenCollateralAmount, meta.CollateralTokenID)
	if err != nil {
		Logger.log.Errorf("Topup v3: Update custodian state error %+v", err)
		if len(rejectInst2) > 0 {
			return [][]string{rejectInst2}, nil
		}
		return [][]string{rejectInst}, nil
	}

	inst := buildPortalCustodianTopupV3(
		meta.IncogAddressStr,
		meta.PortalTokenID,
		meta.CollateralTokenID,
		meta.DepositAmount,
		meta.FreeTokenCollateralAmount,
		uniqExternalTxID,
		common.PortalCustodianTopupSuccessChainStatus,
		meta.Type,
		shardID,
		actionData.TxReqID,
	)
	return [][]string{inst}, nil
}

func (p *portalCustodianTopupProcessorV3) processInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData pMeta.PortalLiquidationCustodianDepositContentV3
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Error when unmarshaling portal liquidation custodian deposit content %v - %v", instructions[3], err)
		return nil
	}

	depositStatus := instructions[2]

	if depositStatus == common.PortalCustodianTopupSuccessChainStatus {
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr)
		custodianStateKeyStr := custodianStateKey.String()
		custodian, ok := currentPortalState.CustodianPoolState[custodianStateKeyStr]
		if !ok {
			Logger.log.Errorf("Process custodian topop v3 error: Custodian not found")
			return nil
		}

		_, err = UpdateCustodianAfterTopup(currentPortalState, custodian, actionData.PortalTokenID, actionData.DepositAmount, actionData.FreeTokenCollateralAmount, actionData.CollateralTokenID)
		if !ok {
			Logger.log.Errorf("Process custodian topop v3 error: %+v", err)
			return nil
		}

		newLiquidationCustodianDeposit := pMeta.NewLiquidationCustodianDepositStatus3(
			actionData.IncogAddressStr,
			actionData.PortalTokenID,
			actionData.CollateralTokenID,
			actionData.DepositAmount,
			actionData.FreeTokenCollateralAmount,
			actionData.UniqExternalTxID,
			actionData.TxReqID,
			common.PortalCustodianTopupSuccessStatus,
		)

		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
		err = statedb.StoreCustodianTopupStatusV3(
			stateDB,
			actionData.TxReqID.String(),
			contentStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit v3 error %v", err)
			return nil
		}

		// store uniq external tx
		err := statedb.InsertPortalExternalTxHashSubmitted(stateDB, actionData.UniqExternalTxID)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking uniq external tx id: %+v", err)
			return nil
		}
	} else if depositStatus == common.PortalCustodianTopupRejectedChainStatus {
		newLiquidationCustodianDeposit := pMeta.NewLiquidationCustodianDepositStatus3(
			actionData.IncogAddressStr,
			actionData.PortalTokenID,
			actionData.CollateralTokenID,
			actionData.DepositAmount,
			actionData.FreeTokenCollateralAmount,
			actionData.UniqExternalTxID,
			actionData.TxReqID,
			common.PortalCustodianTopupRejectedStatus,
		)

		contentStatusBytes, _ := json.Marshal(newLiquidationCustodianDeposit)
		err = statedb.StoreCustodianTopupStatusV3(
			stateDB,
			actionData.TxReqID.String(),
			contentStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation custodian deposit v3 error %v", err)
			return nil
		}
	}

	return nil
}

/* =======
Portal Custodian Topup Processor v3
======= */

type portalTopupWaitingPortingReqProcessorV3 struct {
	*portalInstProcessor
}

func (p *portalTopupWaitingPortingReqProcessorV3) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalTopupWaitingPortingReqProcessorV3) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalTopupWaitingPortingReqProcessorV3) prepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while decoding content string of portal custodian topup action v3: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while decoding content string of portal custodian topup action v3: %+v", err)
	}
	var actionData pMeta.PortalLiquidationCustodianDepositActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action v3: %+v", err)
		return nil, fmt.Errorf("ERROR: an error occured while unmarshal portal custodian deposit action v3: %+v", err)
	}
	meta := actionData.Meta
	if meta.DepositAmount > 0 {
		// NOTE: since TxHash from constructedReceipt is always '0x0000000000000000000000000000000000000000000000000000000000000000'
		// so must build unique external tx as combination of chain name and block hash and tx index.
		uniqExternalTxID := pCommon.GetUniqExternalTxID(common.ETHChainName, meta.BlockHash, meta.TxIndex)
		isSubmitted, err := statedb.IsPortalExternalTxHashSubmitted(stateDB, uniqExternalTxID)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while checking eth tx submitted: %+v", err)
			return nil, fmt.Errorf("ERROR: an error occured while checking eth tx submitted: %+v", err)
		}

		optionalData := make(map[string]interface{})
		optionalData["isSubmitted"] = isSubmitted
		optionalData["uniqExternalTxID"] = uniqExternalTxID
		return optionalData, nil
	}
	return nil, nil
}

func buildPortalTopupWaitingPortingInstV3(
	incogAddress string,
	portalTokenId string,
	collateralTokenID string,
	depositedAmount uint64,
	freeCollateralAmount uint64,
	uniqExternalTxID []byte,
	portingID string,
	status string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	redeemRequestContent := pMeta.PortalTopUpWaitingPortingRequestContentV3{
		IncogAddressStr:           incogAddress,
		PortalTokenID:             portalTokenId,
		CollateralTokenID:         collateralTokenID,
		DepositAmount:             depositedAmount,
		FreeTokenCollateralAmount: freeCollateralAmount,
		UniqExternalTxID:          uniqExternalTxID,
		PortingID:                 portingID,
		TxReqID:                   txReqID,
		ShardID:                   shardID,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

func (p *portalTopupWaitingPortingReqProcessorV3) buildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal custodian topup waiting porting action v3: %+v", err)
		return [][]string{}, nil
	}
	var actionData pMeta.PortalTopUpWaitingPortingRequestActionV3
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal custodian topup waiting porting action v3: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildPortalTopupWaitingPortingInstV3(
		meta.IncogAddressStr,
		meta.PortalTokenID,
		meta.CollateralTokenID,
		meta.DepositAmount,
		meta.FreeTokenCollateralAmount,
		nil,
		meta.PortingID,
		common.PortalTopUpWaitingPortingRejectedChainStatus,
		meta.Type,
		shardID,
		actionData.TxReqID,
	)
	rejectInst2 := []string{}

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null.")
		return [][]string{rejectInst}, nil
	}

	custodianStateKey := statedb.GenerateCustodianStateObjectKey(meta.IncogAddressStr)
	custodian, ok := currentPortalState.CustodianPoolState[custodianStateKey.String()]
	if !ok {
		Logger.log.Errorf("Custodian not found")
		return [][]string{rejectInst}, nil
	}

	// check waiting porting
	waitingPortingRequestKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(meta.PortingID)
	waitingPortingReq, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()]
	if !ok || waitingPortingReq == nil || waitingPortingReq.TokenID() != meta.PortalTokenID {
		Logger.log.Errorf("Waiting porting request with portingID (%s) not found", meta.PortingID)
		return [][]string{rejectInst}, nil
	}
	isMatchPorting := false
	for _, cus := range waitingPortingReq.Custodians() {
		if cus.IncAddress == meta.IncogAddressStr {
			isMatchPorting = true
			break
		}
	}
	if !isMatchPorting {
		Logger.log.Errorf("Custodian %v is not the matching custodian in portingID ", meta.IncogAddressStr, meta.PortingID)
		return [][]string{rejectInst}, nil
	}

	// check free token collaterals
	freeTokenCollaterals := custodian.GetFreeTokenCollaterals()
	if meta.FreeTokenCollateralAmount > 0 && freeTokenCollaterals == nil {
		Logger.log.Errorf("Free token collaterals of custodian's state is nil")
		return [][]string{rejectInst}, nil
	}
	if meta.FreeTokenCollateralAmount > custodian.GetFreeTokenCollaterals()[meta.CollateralTokenID] {
		Logger.log.Errorf("Free token collateral topup amount is greater than free token collateral of custodian's state")
		return [][]string{rejectInst}, nil
	}

	// check deposit amount and deposit proof
	uniqExternalTxID := []byte{}
	if meta.DepositAmount > 0 {
		// check uniqExternalTxID from optionalData which get from statedb
		if optionalData == nil {
			Logger.log.Errorf("Topup v3: optionalData is null")
			return [][]string{rejectInst}, nil
		}
		uniqExternalTxID, ok = optionalData["uniqExternalTxID"].([]byte)
		if !ok || len(uniqExternalTxID) == 0 {
			Logger.log.Errorf("Topup v3: optionalData uniqExternalTxID is invalid")
			return [][]string{rejectInst}, nil
		}

		// reject instruction with uniqExternalTxID
		rejectInst2 = buildPortalTopupWaitingPortingInstV3(
			meta.IncogAddressStr,
			meta.PortalTokenID,
			meta.CollateralTokenID,
			meta.DepositAmount,
			meta.FreeTokenCollateralAmount,
			uniqExternalTxID,
			meta.PortingID,
			common.PortalTopUpWaitingPortingRejectedChainStatus,
			meta.Type,
			shardID,
			actionData.TxReqID,
		)

		isExist, ok := optionalData["isSubmitted"].(bool)
		if !ok {
			Logger.log.Errorf("Topup waiting porting v3: optionalData isSubmitted is invalid")
			return [][]string{rejectInst2}, nil
		}
		if isExist {
			Logger.log.Errorf("Topup waiting porting v3: Unique external id exist in db %v", uniqExternalTxID)
			return [][]string{rejectInst2}, nil
		}

		// verify proof and parse receipt
		ethReceipt, err := pCommon.VerifyProofAndParseReceipt(meta.BlockHash, meta.TxIndex, meta.ProofStrs)
		if err != nil {
			Logger.log.Errorf("Topup waiting porting v3: Verify eth proof error: %+v", err)
			return [][]string{rejectInst2}, nil
		}
		if ethReceipt == nil {
			Logger.log.Errorf("Topup waiting porting v3: The eth proof's receipt could not be null.")
			return [][]string{rejectInst2}, nil
		}

		logMap, err := pCommon.PickAndParseLogMapFromReceiptByContractAddr(ethReceipt, portalParams.PortalETHContractAddressStr, "Deposit")
		if err != nil {
			Logger.log.Errorf("WARNING: an error occured while parsing log map from receipt: ", err)
			return [][]string{rejectInst2}, nil
		}
		if logMap == nil {
			Logger.log.Errorf("WARNING: could not find log map out from receipt")
			return [][]string{rejectInst2}, nil
		}

		// parse info from log map and validate info
		custodianIncAddr, externalTokenIDStr, depositAmount, err := pMeta.ParseInfoFromLogMap(logMap)
		if err != nil {
			Logger.log.Errorf("Topup waiting porting v3: Error when parsing info from log map : %+v", err)
			return [][]string{rejectInst2}, nil
		}
		externalTokenIDStr = common.Remove0xPrefix(externalTokenIDStr)

		if externalTokenIDStr != meta.CollateralTokenID {
			Logger.log.Errorf("Topup waiting porting v3: Collateral token id in meta %v is different from in deposit proof %+v", meta.CollateralTokenID, externalTokenIDStr)
			return [][]string{rejectInst2}, nil
		}

		if custodianIncAddr != meta.IncogAddressStr {
			Logger.log.Errorf("Topup waiting porting v3: Custodian incognito address in meta %v is different from in deposit proof %+v", meta.IncogAddressStr, custodianIncAddr)
			return [][]string{rejectInst2}, nil
		}

		if depositAmount != meta.DepositAmount {
			Logger.log.Errorf("Topup waiting porting v3: Custodian incognito address in meta %v is different from in deposit proof %+v", meta.IncogAddressStr, custodianIncAddr)
			return [][]string{rejectInst2}, nil
		}
	}

	err = UpdateCustodianAfterTopupWaitingPorting(currentPortalState, waitingPortingReq, custodian, meta.PortalTokenID, meta.DepositAmount, meta.FreeTokenCollateralAmount, meta.CollateralTokenID)
	if err != nil {
		Logger.log.Errorf("Topup waiting porting v3: Update custodian state error %+v", err)
		if len(rejectInst2) > 0 {
			return [][]string{rejectInst2}, nil
		}
		return [][]string{rejectInst}, nil
	}

	inst := buildPortalTopupWaitingPortingInstV3(
		meta.IncogAddressStr,
		meta.PortalTokenID,
		meta.CollateralTokenID,
		meta.DepositAmount,
		meta.FreeTokenCollateralAmount,
		uniqExternalTxID,
		meta.PortingID,
		common.PortalTopUpWaitingPortingSuccessChainStatus,
		meta.Type,
		shardID,
		actionData.TxReqID,
	)
	return [][]string{inst}, nil
}

func (p *portalTopupWaitingPortingReqProcessorV3) processInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	var actionData pMeta.PortalTopUpWaitingPortingRequestContentV3
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Error when unmarshaling portal top up waiting porting action %v - %v", instructions[3], err)
		return nil
	}

	depositStatus := instructions[2]
	if depositStatus == common.PortalTopUpWaitingPortingRejectedChainStatus {
		topUpWaitingPortingReq := pMeta.NewPortalTopUpWaitingPortingRequestStatusV3(
			actionData.IncogAddressStr,
			actionData.PortalTokenID,
			actionData.CollateralTokenID,
			actionData.DepositAmount,
			actionData.FreeTokenCollateralAmount,
			actionData.PortingID,
			actionData.UniqExternalTxID,
			actionData.TxReqID,
			common.PortalTopUpWaitingPortingRejectedStatus,
		)
		statusContentBytes, _ := json.Marshal(topUpWaitingPortingReq)
		err = statedb.StoreCustodianTopupWaitingPortingStatusV3(
			stateDB,
			actionData.TxReqID.String(),
			statusContentBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while storing waiting porting top up error %v", err)
		}
	} else if depositStatus == common.PortalTopUpWaitingPortingSuccessChainStatus {
		custodianStateKey := statedb.GenerateCustodianStateObjectKey(actionData.IncogAddressStr)
		custodian, ok := currentPortalState.CustodianPoolState[custodianStateKey.String()]
		if !ok {
			Logger.log.Errorf("Custodian not found")
			return nil
		}

		waitingPortingRequestKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(actionData.PortingID)
		waitingPortingReq, ok := currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()]
		if !ok || waitingPortingReq == nil || waitingPortingReq.TokenID() != actionData.PortalTokenID {
			Logger.log.Errorf("Waiting porting request with portingID (%s) not found", actionData.PortingID)
			return nil
		}

		err = UpdateCustodianAfterTopupWaitingPorting(currentPortalState, waitingPortingReq, custodian, actionData.PortalTokenID, actionData.DepositAmount, actionData.FreeTokenCollateralAmount, actionData.CollateralTokenID)
		if err != nil {
			Logger.log.Errorf("Update portal state error: %+v", err)
			return nil
		}

		topUpWaitingPortingReq := pMeta.NewPortalTopUpWaitingPortingRequestStatusV3(
			actionData.IncogAddressStr,
			actionData.PortalTokenID,
			actionData.CollateralTokenID,
			actionData.DepositAmount,
			actionData.FreeTokenCollateralAmount,
			actionData.PortingID,
			actionData.UniqExternalTxID,
			actionData.TxReqID,
			common.PortalTopUpWaitingPortingSuccessStatus,
		)
		statusContentBytes, _ := json.Marshal(topUpWaitingPortingReq)
		err = statedb.StoreCustodianTopupWaitingPortingStatusV3(
			stateDB,
			actionData.TxReqID.String(),
			statusContentBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while storing waiting porting top up error %v", err)
		}

		// update state of porting request by portingID
		newPortingRequestState := pMeta.NewPortingRequestStatus(
			waitingPortingReq.UniquePortingID(),
			waitingPortingReq.TxReqID(),
			waitingPortingReq.TokenID(),
			waitingPortingReq.PorterAddress(),
			waitingPortingReq.Amount(),
			waitingPortingReq.Custodians(),
			waitingPortingReq.PortingFee(),
			common.PortalPortingReqWaitingStatus,
			beaconHeight+1,
			waitingPortingReq.ShardHeight(),
			waitingPortingReq.ShardID(),
		)
		newPortingRequestStatusBytes, _ := json.Marshal(newPortingRequestState)
		err = statedb.StorePortalPortingRequestStatus(
			stateDB,
			waitingPortingReq.UniquePortingID(),
			newPortingRequestStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item: %+v", err)
			return nil
		}

		// store uniq external tx
		err := statedb.InsertPortalExternalTxHashSubmitted(stateDB, actionData.UniqExternalTxID)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking uniq external tx id: %+v", err)
			return nil
		}
	}
	return nil
}


/* =======
Portal Liquidation Custodian Run Away Processor
======= */
type portalLiquidationCustodianRunAwayProcessor struct {
	*portalInstProcessor
}

func (p *portalLiquidationCustodianRunAwayProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalLiquidationCustodianRunAwayProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalLiquidationCustodianRunAwayProcessor) prepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// beacon build instruction for portal liquidation when custodians run away - don't send public tokens back to users.
func buildCustodianRunAwayLiquidationInst(
	redeemID string,
	tokenID string,
	redeemPubTokenAmount uint64,
	mintedCollateralAmount uint64,
	remainUnlockAmountForCustodian uint64,
	mintedCollateralAmounts map[string]uint64,
	remainUnlockAmountsForCustodian map[string]uint64,
	redeemerIncAddrStr string,
	custodianIncAddrStr string,
	liquidatedByExchangeRate bool,
	metaType int,
	shardID byte,
	status string,
) []string {
	liqCustodianContent := pMeta.PortalLiquidateCustodianContent{
		UniqueRedeemID:                 redeemID,
		TokenID:                        tokenID,
		RedeemPubTokenAmount:           redeemPubTokenAmount,
		LiquidatedCollateralAmount:     mintedCollateralAmount,
		RemainUnlockAmountForCustodian: remainUnlockAmountForCustodian,
		RedeemerIncAddressStr:          redeemerIncAddrStr,
		CustodianIncAddressStr:         custodianIncAddrStr,
		LiquidatedByExchangeRate:       liquidatedByExchangeRate,
		ShardID:                        shardID,
		// portal v3
		LiquidatedCollateralAmounts:     mintedCollateralAmounts,
		RemainUnlockAmountsForCustodian: remainUnlockAmountsForCustodian,
	}
	liqCustodianContentBytes, _ := json.Marshal(liqCustodianContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(liqCustodianContentBytes),
	}
}

func (p *portalLiquidationCustodianRunAwayProcessor) buildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	insts := [][]string{}

	// get exchange rate
	exchangeRate := currentPortalState.FinalExchangeRatesState
	if exchangeRate == nil {
		Logger.log.Errorf("[CheckAndBuildInstForCustodianLiquidation] Error when get exchange rate")
		return insts, nil
	}

	liquidatedByExchangeRate := false

	sortedMatchedRedeemReqKeys := make([]string, 0)
	for key := range currentPortalState.MatchedRedeemRequests {
		sortedMatchedRedeemReqKeys = append(sortedMatchedRedeemReqKeys, key)
	}
	sort.Strings(sortedMatchedRedeemReqKeys)
	for _, redeemReqKey := range sortedMatchedRedeemReqKeys {
		redeemReq := currentPortalState.MatchedRedeemRequests[redeemReqKey]
		if bc.CheckBlockTimeIsReached(beaconHeight, redeemReq.GetBeaconHeight(), shardHeights[redeemReq.ShardID()], redeemReq.ShardHeight(), portalParams.TimeOutCustodianReturnPubToken) {
			// get shardId of redeemer
			redeemerKey, err := wallet.Base58CheckDeserialize(redeemReq.GetRedeemerAddress())
			if err != nil {
				Logger.log.Errorf("[CheckAndBuildInstForCustodianLiquidation] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
					redeemReq.GetUniqueRedeemID(), err)
				continue
			}
			shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

			// get tokenID from redeemTokenID
			tokenID := redeemReq.GetTokenID()
			metaType := basemeta.PortalLiquidateCustodianMeta
			unlockCollateralsForUser := make(map[string]uint64)

			liquidatedCustodians := make([]*statedb.MatchingRedeemCustodianDetail, 0)
			for _, matchCusDetail := range redeemReq.GetCustodians() {
				//Logger.log.Errorf("matchCusDetail.GetIncognitoAddress(): %v\n", matchCusDetail.GetIncognitoAddress())
				custodianStateKey := statedb.GenerateCustodianStateObjectKey(matchCusDetail.GetIncognitoAddress()).String()
				//Logger.log.Errorf("custodianStateKey: %v\n", custodianStateKey)

				// determine meta type
				lockedCollaterals := currentPortalState.CustodianPoolState[custodianStateKey].GetLockedTokenCollaterals()
				if lockedCollaterals != nil && len(lockedCollaterals[tokenID]) != 0 {
					metaType = basemeta.PortalLiquidateCustodianMetaV3
				}

				// calculate liquidated amount and remain unlocked amount for custodian
				var liquidatedAmount, remainUnlockAmount uint64
				var liquidatedAmounts, remainUnlockAmounts map[string]uint64
				if metaType == basemeta.PortalLiquidateCustodianMeta {
					// return value in prv
					liquidatedAmount, remainUnlockAmount, err = CalUnlockCollateralAmountAfterLiquidation(
						currentPortalState,
						custodianStateKey,
						matchCusDetail.GetAmount(),
						tokenID,
						exchangeRate,
						portalParams)
				} else {
					// return value in usdt
					liquidatedAmount, remainUnlockAmount, liquidatedAmounts, remainUnlockAmounts, err = CalUnlockCollateralAmountAfterLiquidationV3(
						currentPortalState,
						custodianStateKey,
						matchCusDetail.GetAmount(),
						tokenID,
						portalParams)
					if len(liquidatedAmounts) > 0 {
						for i, v := range liquidatedAmounts {
							unlockCollateralsForUser[i] += v
						}
					}
				}
				if err != nil {
					Logger.log.Errorf("[CheckAndBuildInstForCustodianLiquidation] Error when calculating unlock collateral amount %v\n: ", err)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.GetUniqueRedeemID(),
						redeemReq.GetTokenID(),
						matchCusDetail.GetAmount(),
						0,
						0,
						liquidatedAmounts,
						remainUnlockAmounts,
						redeemReq.GetRedeemerAddress(),
						matchCusDetail.GetIncognitoAddress(),
						liquidatedByExchangeRate,
						metaType,
						shardID,
						common.PortalLiquidateCustodianFailedChainStatus,
					)
					insts = append(insts, inst)
					continue
				}

				// update custodian state
				custodianState := currentPortalState.CustodianPoolState[custodianStateKey]
				if metaType == basemeta.PortalLiquidateCustodianMeta {
					err = updateCustodianStateAfterLiquidateCustodian(custodianState, liquidatedAmount, remainUnlockAmount, tokenID)
				} else {
					err = updateCustodianStateAfterLiquidateCustodianV3(custodianState, liquidatedAmount, remainUnlockAmount, liquidatedAmounts, remainUnlockAmounts, tokenID)
				}
				if err != nil {
					Logger.log.Errorf("[CheckAndBuildInstForCustodianLiquidation] Error when updating custodian state %v\n: ", err)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.GetUniqueRedeemID(),
						redeemReq.GetTokenID(),
						matchCusDetail.GetAmount(),
						liquidatedAmount,
						remainUnlockAmount,
						liquidatedAmounts,
						remainUnlockAmounts,
						redeemReq.GetRedeemerAddress(),
						matchCusDetail.GetIncognitoAddress(),
						liquidatedByExchangeRate,
						metaType,
						shardID,
						common.PortalLiquidateCustodianFailedChainStatus,
					)
					insts = append(insts, inst)
					continue
				}

				// remove matching custodian from matching custodians list in waiting redeem request
				liquidatedCustodians = append(liquidatedCustodians, matchCusDetail)

				// build instruction
				inst := buildCustodianRunAwayLiquidationInst(
					redeemReq.GetUniqueRedeemID(),
					redeemReq.GetTokenID(),
					matchCusDetail.GetAmount(),
					liquidatedAmount,
					remainUnlockAmount,
					liquidatedAmounts,
					remainUnlockAmounts,
					redeemReq.GetRedeemerAddress(),
					matchCusDetail.GetIncognitoAddress(),
					liquidatedByExchangeRate,
					metaType,
					shardID,
					common.PortalLiquidateCustodianSuccessChainStatus,
				)
				insts = append(insts, inst)
			}

			// create proof to liquidate runaway custodian
			if metaType == basemeta.PortalLiquidateCustodianMetaV3 && len(unlockCollateralsForUser) > 0 {
				liquidatedBigIntAmounts := make(map[string]*big.Int)
				for tokenLiquidateId, tokenAmount := range unlockCollateralsForUser {
					amountBN := big.NewInt(0).SetUint64(tokenAmount)
					if bytes.Equal(common.FromHex(tokenLiquidateId), common.FromHex(common.EthAddrStr)) {
						// Convert Gwei to Wei for Ether
						amountBN = amountBN.Mul(amountBN, big.NewInt(1000000000))
					}
					liquidatedBigIntAmounts[tokenLiquidateId] = amountBN
				}
				confirmInst := buildConfirmWithdrawCollateralInstV3(
					basemeta.PortalLiquidateRunAwayCustodianConfirmMetaV3,
					shardID,
					redeemReq.GetRedeemerAddress(),
					redeemReq.GetRedeemerExternalAddress(),
					liquidatedBigIntAmounts,
					redeemReq.GetTxReqID(),
					beaconHeight+1,
				)
				insts = append(insts, confirmInst)
			}
			updatedCustodians := currentPortalState.MatchedRedeemRequests[redeemReqKey].GetCustodians()
			for _, cus := range liquidatedCustodians {
				updatedCustodians, _ = removeCustodianFromMatchingRedeemCustodians(
					updatedCustodians, cus.GetIncognitoAddress())
			}

			// remove redeem request from waiting redeem requests list
			currentPortalState.MatchedRedeemRequests[redeemReqKey].SetCustodians(updatedCustodians)
			if len(currentPortalState.MatchedRedeemRequests[redeemReqKey].GetCustodians()) == 0 {
				deleteMatchedRedeemRequest(currentPortalState, redeemReqKey)
			}
		}
	}

	return insts, nil
}

func (p *portalLiquidationCustodianRunAwayProcessor) processInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	// unmarshal instructions content
	var actionData pMeta.PortalLiquidateCustodianContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalLiquidateCustodianSuccessChainStatus {
		// update custodian state
		Logger.log.Infof("[processPortalLiquidateCustodian] actionData.CustodianIncAddressStr = %s in beaconHeight=%d", actionData.CustodianIncAddressStr, beaconHeight)
		cusStateKey := statedb.GenerateCustodianStateObjectKey(actionData.CustodianIncAddressStr)
		cusStateKeyStr := cusStateKey.String()
		custodianState, ok := currentPortalState.CustodianPoolState[cusStateKeyStr]
		if !ok {
			Logger.log.Errorf("[processPortalLiquidateCustodian] cusStateKeyStr %s can not found", cusStateKeyStr)
			return nil
		}

		if instructions[0] == strconv.Itoa(basemeta.PortalLiquidateCustodianMeta) {
			err = updateCustodianStateAfterLiquidateCustodian(custodianState, actionData.LiquidatedCollateralAmount, actionData.RemainUnlockAmountForCustodian, actionData.TokenID)
		} else {
			err = updateCustodianStateAfterLiquidateCustodianV3(custodianState, actionData.LiquidatedCollateralAmount, actionData.RemainUnlockAmountForCustodian, actionData.LiquidatedCollateralAmounts, actionData.RemainUnlockAmountsForCustodian, actionData.TokenID)
		}

		if err != nil {
			Logger.log.Errorf("[processPortalLiquidateCustodian] Error when update custodian state after liquidation %v", err)
			return nil
		}

		// remove matching custodian from matching custodians list in matched redeem request
		matchedRedeemReqKey := statedb.GenerateMatchedRedeemRequestObjectKey(actionData.UniqueRedeemID)
		matchedRedeemReqKeyStr := matchedRedeemReqKey.String()

		updatedCustodians, err := removeCustodianFromMatchingRedeemCustodians(
			currentPortalState.MatchedRedeemRequests[matchedRedeemReqKeyStr].GetCustodians(), actionData.CustodianIncAddressStr)
		currentPortalState.MatchedRedeemRequests[matchedRedeemReqKeyStr].SetCustodians(updatedCustodians)
		if err != nil {
			Logger.log.Errorf("[processPortalLiquidateCustodian] Error when removing custodian from matching custodians %v", err)
			return nil
		}

		// remove redeem request from matched redeem requests list
		if len(currentPortalState.MatchedRedeemRequests[matchedRedeemReqKeyStr].GetCustodians()) == 0 {
			deleteMatchedRedeemRequest(currentPortalState, matchedRedeemReqKeyStr)
			statedb.DeleteMatchedRedeemRequest(stateDB, actionData.UniqueRedeemID)

			// update status of redeem request with redeemID to liquidated status
			err = updateRedeemRequestStatusByRedeemId(actionData.UniqueRedeemID, common.PortalRedeemReqLiquidatedStatus, stateDB)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occurred while updating redeem request status by redeemID: %+v", err)
				return nil
			}
		}

		// track liquidation custodian status by redeemID and custodian address into DB
		custodianLiquidationTrackData := pMeta.PortalLiquidateCustodianStatus{
			Status:                          common.PortalLiquidateCustodianSuccessStatus,
			UniqueRedeemID:                  actionData.UniqueRedeemID,
			TokenID:                         actionData.TokenID,
			RedeemPubTokenAmount:            actionData.RedeemPubTokenAmount,
			LiquidatedCollateralAmount:      actionData.LiquidatedCollateralAmount,
			RemainUnlockAmountForCustodian:  actionData.RemainUnlockAmountForCustodian,
			LiquidatedCollateralAmounts:     actionData.LiquidatedCollateralAmounts,
			RemainUnlockAmountsForCustodian: actionData.RemainUnlockAmountsForCustodian,
			RedeemerIncAddressStr:           actionData.RedeemerIncAddressStr,
			CustodianIncAddressStr:          actionData.CustodianIncAddressStr,
			LiquidatedByExchangeRate:        actionData.LiquidatedByExchangeRate,
			ShardID:                         actionData.ShardID,
			LiquidatedBeaconHeight:          beaconHeight + 1,
		}
		custodianLiquidationTrackDataBytes, _ := json.Marshal(custodianLiquidationTrackData)
		err = statedb.StorePortalLiquidationCustodianRunAwayStatus(
			stateDB,
			actionData.UniqueRedeemID,
			actionData.CustodianIncAddressStr,
			custodianLiquidationTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}

	} else if reqStatus == common.PortalLiquidateCustodianFailedChainStatus {
		// track liquidation custodian status by redeemID and custodian address into DB
		custodianLiquidationTrackData := pMeta.PortalLiquidateCustodianStatus{
			Status:                          common.PortalLiquidateCustodianFailedStatus,
			UniqueRedeemID:                  actionData.UniqueRedeemID,
			TokenID:                         actionData.TokenID,
			RedeemPubTokenAmount:            actionData.RedeemPubTokenAmount,
			LiquidatedCollateralAmount:      actionData.LiquidatedCollateralAmount,
			RemainUnlockAmountForCustodian:  actionData.RemainUnlockAmountForCustodian,
			LiquidatedCollateralAmounts:     actionData.LiquidatedCollateralAmounts,
			RemainUnlockAmountsForCustodian: actionData.RemainUnlockAmountsForCustodian,
			RedeemerIncAddressStr:           actionData.RedeemerIncAddressStr,
			CustodianIncAddressStr:          actionData.CustodianIncAddressStr,
			LiquidatedByExchangeRate:        actionData.LiquidatedByExchangeRate,
			ShardID:                         actionData.ShardID,
			LiquidatedBeaconHeight:          beaconHeight + 1,
		}
		custodianLiquidationTrackDataBytes, _ := json.Marshal(custodianLiquidationTrackData)
		err = statedb.StorePortalLiquidationCustodianRunAwayStatus(
			stateDB,
			actionData.UniqueRedeemID,
			actionData.CustodianIncAddressStr,
			custodianLiquidationTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking liquidation custodian: %+v", err)
			return nil
		}
	}

	return nil
}

/* =======
Portal Liquidation Custodian Run Away Processor
======= */
type portalExpiredWaitingPortingProcessor struct {
	*portalInstProcessor
}

func (p *portalExpiredWaitingPortingProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalExpiredWaitingPortingProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalExpiredWaitingPortingProcessor) prepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

// beacon build instruction for expired waiting porting request - user doesn't send public token to custodian after requesting
func buildExpiredWaitingPortingReqInst(
	portingID string,
	expiredByLiquidation bool,
	metaType int,
	shardID byte,
	status string,
) []string {
	liqCustodianContent := pMeta.PortalExpiredWaitingPortingReqContent{
		UniquePortingID:      portingID,
		ExpiredByLiquidation: expiredByLiquidation,
		ShardID:              shardID,
	}
	liqCustodianContentBytes, _ := json.Marshal(liqCustodianContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(liqCustodianContentBytes),
	}
}

func buildInstForExpiredPortingReqByPortingID(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	portingReqKey string,
	portingReq *statedb.WaitingPortingRequest,
	expiredByLiquidation bool) ([][]string, error) {
	insts := [][]string{}

	//get shardId
	shardID := portingReq.ShardID()

	// get tokenID from redeemTokenID
	tokenID := portingReq.TokenID()

	// update custodian state in matching custodians list (holding public tokens, locked amount)
	for _, matchCusDetail := range portingReq.Custodians() {
		cusStateKey := statedb.GenerateCustodianStateObjectKey(matchCusDetail.IncAddress)
		cusStateKeyStr := cusStateKey.String()
		custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
		if custodianState == nil {
			Logger.log.Errorf("[checkAndBuildInstForExpiredWaitingPortingRequest] Error when get custodian state with key %v\n: ", cusStateKey)
			continue
		}
		err := updateCustodianStateAfterExpiredPortingReq(custodianState, matchCusDetail.LockedAmountCollateral, matchCusDetail.LockedTokenCollaterals, tokenID)
		if err != nil {
			return insts, err
		}
	}

	// remove waiting porting request from waiting list
	delete(currentPortalState.WaitingPortingRequests, portingReqKey)

	// build instruction
	inst := buildExpiredWaitingPortingReqInst(
		portingReq.UniquePortingID(),
		expiredByLiquidation,
		basemeta.PortalExpiredWaitingPortingReqMeta,
		shardID,
		common.PortalExpiredWaitingPortingReqSuccessChainStatus,
	)
	insts = append(insts, inst)

	return insts, nil
}

func (p *portalExpiredWaitingPortingProcessor) buildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	insts := [][]string{}
	sortedWaitingPortingReqKeys := make([]string, 0)
	for key := range currentPortalState.WaitingPortingRequests {
		sortedWaitingPortingReqKeys = append(sortedWaitingPortingReqKeys, key)
	}
	sort.Strings(sortedWaitingPortingReqKeys)
	for _, portingReqKey := range sortedWaitingPortingReqKeys {
		portingReq := currentPortalState.WaitingPortingRequests[portingReqKey]
		if bc.CheckBlockTimeIsReached(beaconHeight, portingReq.BeaconHeight(), shardHeights[portingReq.ShardID()], portingReq.ShardHeight(), portalParams.TimeOutWaitingPortingRequest) {
			inst, err := buildInstForExpiredPortingReqByPortingID(
				beaconHeight, currentPortalState, portingReqKey, portingReq, false)
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstForExpiredWaitingPortingRequest] Error when build instruction for expired porting request %v\n", err)
				continue
			}
			insts = append(insts, inst...)
		}
	}

	return insts, nil
}

func (p *portalExpiredWaitingPortingProcessor) processInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	if currentPortalState == nil {
		Logger.log.Errorf("current portal state is nil")
		return nil
	}
	if len(instructions) != 4 {
		return nil // skip the instruction
	}

	// unmarshal instructions content
	var actionData pMeta.PortalExpiredWaitingPortingReqContent
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Error when unmarshaling portal expired waiting porting content %v - %v", instructions[3], err)
		return nil
	}

	status := instructions[2]
	waitingPortingID := actionData.UniquePortingID

	if status == common.PortalExpiredWaitingPortingReqSuccessChainStatus {
		waitingPortingKey := statedb.GeneratePortalWaitingPortingRequestObjectKey(waitingPortingID)
		waitingPortingKeyStr := waitingPortingKey.String()
		waitingPortingReq := currentPortalState.WaitingPortingRequests[waitingPortingKeyStr]
		if waitingPortingReq == nil {
			Logger.log.Errorf("[processPortalExpiredPortingRequest] waiting porting req nil with key : %v", waitingPortingKey)
			return nil
		}

		// get tokenID from redeemTokenID
		tokenID := waitingPortingReq.TokenID()

		// update custodian state in matching custodians list (holding public tokens, locked amount)
		for _, matchCusDetail := range waitingPortingReq.Custodians() {
			cusStateKey := statedb.GenerateCustodianStateObjectKey(matchCusDetail.IncAddress)
			cusStateKeyStr := cusStateKey.String()
			custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
			if custodianState == nil {
				Logger.log.Errorf("[checkAndBuildInstForExpiredWaitingPortingRequest] Error when get custodian state with key %v\n: ", cusStateKey)
				continue
			}
			err = updateCustodianStateAfterExpiredPortingReq(custodianState, matchCusDetail.LockedAmountCollateral, matchCusDetail.LockedTokenCollaterals, tokenID)
			if err != nil {
				Logger.log.Errorf("ERROR: an error occured while update state for expired porting request: %+v", err)
				return nil
			}
		}

		// remove waiting porting request from waiting list
		delete(currentPortalState.WaitingPortingRequests, waitingPortingKeyStr)
		statedb.DeleteWaitingPortingRequest(stateDB, waitingPortingReq.UniquePortingID())

		// update status of porting ID  => expired/liquidated
		portingReqStatus := common.PortalPortingReqExpiredStatus
		if actionData.ExpiredByLiquidation {
			portingReqStatus = common.PortalPortingReqLiquidatedStatus
		}

		newPortingRequestStatus := pMeta.NewPortingRequestStatus(
			waitingPortingReq.UniquePortingID(),
			waitingPortingReq.TxReqID(),
			tokenID,
			waitingPortingReq.PorterAddress(),
			waitingPortingReq.Amount(),
			waitingPortingReq.Custodians(),
			waitingPortingReq.PortingFee(),
			portingReqStatus,
			waitingPortingReq.BeaconHeight(),
			waitingPortingReq.ShardHeight(),
			waitingPortingReq.ShardID())

		newPortingRequestStatusBytes, _ := json.Marshal(newPortingRequestStatus)
		err = statedb.StorePortalPortingRequestStatus(
			stateDB,
			actionData.UniquePortingID,
			newPortingRequestStatusBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store porting request item status: %+v", err)
			return nil
		}

		// track expired waiting porting request status by portingID into DB
		expiredPortingTrackData := pMeta.PortalExpiredWaitingPortingReqStatus{
			Status:               common.PortalExpiredPortingReqSuccessStatus,
			UniquePortingID:      waitingPortingID,
			ShardID:              actionData.ShardID,
			ExpiredByLiquidation: actionData.ExpiredByLiquidation,
			ExpiredBeaconHeight:  beaconHeight + 1,
		}
		expiredPortingTrackDataBytes, _ := json.Marshal(expiredPortingTrackData)
		err = statedb.StorePortalExpiredPortingRequestStatus(
			stateDB,
			waitingPortingID,
			expiredPortingTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking expired porting request: %+v", err)
			return nil
		}

	} else if status == common.PortalCustodianTopupRejectedChainStatus {
		// track expired waiting porting request status by portingID into DB
		expiredPortingTrackData := pMeta.PortalExpiredWaitingPortingReqStatus{
			Status:               common.PortalExpiredPortingReqFailedStatus,
			UniquePortingID:      waitingPortingID,
			ShardID:              actionData.ShardID,
			ExpiredByLiquidation: actionData.ExpiredByLiquidation,
			ExpiredBeaconHeight:  beaconHeight + 1,
		}
		expiredPortingTrackDataBytes, _ := json.Marshal(expiredPortingTrackData)
		err = statedb.StorePortalExpiredPortingRequestStatus(
			stateDB,
			waitingPortingID,
			expiredPortingTrackDataBytes,
		)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occured while tracking expired porting request: %+v", err)
			return nil
		}
	}

	return nil
}


/* =======
Portal Liquidation Custodian Run Away Processor
======= */
type portalLiquidationByRatesProcessor struct {
	*portalInstProcessor
}

func (p *portalLiquidationByRatesProcessor) getActions() map[byte][][]string {
	return p.actions
}

func (p *portalLiquidationByRatesProcessor) putAction(action []string, shardID byte) {
	_, found := p.actions[shardID]
	if !found {
		p.actions[shardID] = [][]string{action}
	} else {
		p.actions[shardID] = append(p.actions[shardID], action)
	}
}

func (p *portalLiquidationByRatesProcessor) prepareDataForBlockProducer(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildInstRejectRedeemRequestByLiquidationExchangeRate(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	wRedeemID string,
) ([]string, error) {
	redeemReqKey := statedb.GenerateWaitingRedeemRequestObjectKey(wRedeemID).String()
	redeemReq := currentPortalState.WaitingRedeemRequests[redeemReqKey]

	// reject waiting redeem request, return ptoken and redeem fee for users
	// update custodian state (return holding public token amount)
	err := UpdateCustodianStateAfterRejectRedeemRequestByLiquidation(currentPortalState, redeemReq, beaconHeight)
	if err != nil {
		Logger.log.Errorf("[buildInstRejectRedeemRequestByLiquidationExchangeRate] Error when updating custodian state %v - RedeemID %v\n: ",
			err, redeemReq.GetUniqueRedeemID())
		return []string{}, nil
	}

	// remove redeem request from waiting redeem requests list
	deleteWaitingRedeemRequest(currentPortalState, redeemReqKey)

	// get shardId of redeemer
	redeemerKey, err := wallet.Base58CheckDeserialize(redeemReq.GetRedeemerAddress())
	if err != nil {
		Logger.log.Errorf("[buildInstRejectRedeemRequestByLiquidationExchangeRate] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
			redeemReq.GetUniqueRedeemID(), err)
		return []string{}, nil
	}
	shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

	// build instruction
	inst := buildRedeemRequestInst(
		redeemReq.GetUniqueRedeemID(),
		redeemReq.GetTokenID(),
		redeemReq.GetRedeemAmount(),
		redeemReq.GetRedeemerAddress(),
		redeemReq.GetRedeemerRemoteAddress(),
		redeemReq.GetRedeemFee(),
		redeemReq.GetCustodians(),
		basemeta.PortalRedeemRequestMetaV3,
		shardID,
		common.Hash{},
		common.PortalRedeemReqCancelledByLiquidationChainStatus,
		redeemReq.ShardHeight(),
		redeemReq.GetRedeemerExternalAddress(),
	)

	return inst, nil
}

func buildLiquidationByExchangeRateInstV3(
	custodianAddress string,
	metaType int,
	status string,
	liquidationInfo map[string]pMeta.LiquidationByRatesDetailV3,
	remainUnlockCollaterals map[string]pMeta.RemainUnlockCollateral,
) []string {
	liquidationContent := pMeta.PortalLiquidationByRatesContentV3{
		CustodianIncAddress:     custodianAddress,
		Details:                 liquidationInfo,
		RemainUnlockCollaterals: remainUnlockCollaterals,
	}
	liquidationContentBytes, _ := json.Marshal(liquidationContent)
	return []string{
		strconv.Itoa(metaType),
		"-1",
		status,
		string(liquidationContentBytes),
	}
}

func (p *portalLiquidationByRatesProcessor) buildNewInsts(
	bc basemeta.ChainRetriever,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	shardHeights map[byte]uint64,
	portalParams portal.PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	if currentPortalState == nil {
		Logger.log.Errorf("[LIQUIDATIONBYRATES] Current portal state is null")
		return [][]string{}, nil
	}
	if len(currentPortalState.CustodianPoolState) == 0 {
		return [][]string{}, nil
	}
	exchangeRate := currentPortalState.FinalExchangeRatesState
	if exchangeRate == nil {
		Logger.log.Errorf("[LIQUIDATIONBYRATES] Final exchange rate is empty")
		return [][]string{}, nil
	}

	insts := [][]string{}
	custodianPoolState := currentPortalState.CustodianPoolState
	sortedCustodianStateKeys := make([]string, 0)
	for key := range custodianPoolState {
		sortedCustodianStateKeys = append(sortedCustodianStateKeys, key)
	}
	sort.Strings(sortedCustodianStateKeys)

	for _, custodianKey := range sortedCustodianStateKeys {
		custodianState := custodianPoolState[custodianKey]
		tpRatios, remainUnlockColalterals, rejectedWRedeemIDs, err := calAndCheckLiquidationRatioV3(currentPortalState, custodianState, exchangeRate, portalParams)
		if err != nil {
			Logger.log.Errorf("Error when calculating and checking tp ratio %v", err)
			continue
		}

		if len(tpRatios) == 0 {
			continue
		}

		// reject waiting redeem requests that matching with liquidated custodians
		for _, wRedeemId := range rejectedWRedeemIDs {
			inst, err := buildInstRejectRedeemRequestByLiquidationExchangeRate(beaconHeight, currentPortalState, wRedeemId)
			if err != nil {
				Logger.log.Errorf("[LIQUIDATIONBYRATES] Error when building instruction reject redeem request ID %v - %v", wRedeemId, err)
				continue
			}
			insts = append(insts, inst)
		}

		// update current portal state after liquidation custodianKey
		updateCurrentPortalStateAfterLiquidationByRatesV3(currentPortalState, custodianKey, tpRatios, remainUnlockColalterals)
		inst := buildLiquidationByExchangeRateInstV3(
			custodianState.GetIncognitoAddress(),
			basemeta.PortalLiquidateByRatesMetaV3,
			common.PortalLiquidateTPExchangeRatesSuccessChainStatus,
			tpRatios,
			remainUnlockColalterals)
		insts = append(insts, inst)
	}

	return insts, nil
}

func (p *portalLiquidationByRatesProcessor) processInsts(
	stateDB *statedb.StateDB,
	beaconHeight uint64,
	instructions []string,
	currentPortalState *CurrentPortalState,
	portalParams portal.PortalParams,
	updatingInfoByTokenID map[common.Hash]basemeta.UpdatingInfo,
) error {
	// unmarshal instructions content
	var actionData pMeta.PortalLiquidationByRatesContentV3
	err := json.Unmarshal([]byte(instructions[3]), &actionData)
	if err != nil {
		Logger.log.Errorf("Can not unmarshal instruction content %v - Error %v\n", instructions[3], err)
		return nil
	}

	cusStateKeyStr := statedb.GenerateCustodianStateObjectKey(actionData.CustodianIncAddress).String()
	custodianState, ok := currentPortalState.CustodianPoolState[cusStateKeyStr]
	if !ok || custodianState == nil {
		Logger.log.Errorf("Custodian not found")
		return nil
	}

	reqStatus := instructions[2]
	if reqStatus == common.PortalLiquidateTPExchangeRatesSuccessChainStatus {
		liquidationInfo := actionData.Details

		//update current portal state
		updateCurrentPortalStateAfterLiquidationByRatesV3(currentPortalState, cusStateKeyStr, liquidationInfo, actionData.RemainUnlockCollaterals)

		// store db
		status := pMeta.PortalLiquidationByRatesStatusV3{
			CustodianIncAddress: actionData.CustodianIncAddress,
			Details:             actionData.Details,
		}
		statusBytes, _ := json.Marshal(status)
		err = statedb.StoreLiquidationByExchangeRateStatusV3(
			stateDB,
			beaconHeight,
			custodianState.GetIncognitoAddress(),
			statusBytes)
		if err != nil {
			Logger.log.Errorf("ERROR: an error occurred while store liquidation by exchange rates v3 %v", err)
			return nil
		}
	}
	return nil
}