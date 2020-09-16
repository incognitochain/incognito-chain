package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wallet"
)

// beacon build instruction for portal liquidation when custodians run away - don't send public tokens back to users.
func buildCustodianRunAwayLiquidationInst(
	redeemID string,
	tokenID string,
	redeemPubTokenAmount uint64,
	mintedCollateralAmount uint64,
	remainUnlockAmountForCustodian uint64,
	redeemerIncAddrStr string,
	custodianIncAddrStr string,
	liquidatedByExchangeRate bool,
	metaType int,
	shardID byte,
	status string,
) []string {
	liqCustodianContent := metadata.PortalLiquidateCustodianContent{
		UniqueRedeemID:                 redeemID,
		TokenID:                        tokenID,
		RedeemPubTokenAmount:           redeemPubTokenAmount,
		LiquidatedCollateralAmount:     mintedCollateralAmount,
		RemainUnlockAmountForCustodian: remainUnlockAmountForCustodian,
		RedeemerIncAddressStr:          redeemerIncAddrStr,
		CustodianIncAddressStr:         custodianIncAddrStr,
		LiquidatedByExchangeRate:       liquidatedByExchangeRate,
		ShardID:                        shardID,
	}
	liqCustodianContentBytes, _ := json.Marshal(liqCustodianContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(liqCustodianContentBytes),
	}
}

func buildTopPercentileExchangeRatesLiquidationInst(
	custodianAddress string,
	metaType int,
	status string,
	topPercentile map[string]metadata.LiquidateTopPercentileExchangeRatesDetail,
	remainUnlockAmounts map[string]uint64,
) []string {
	tpContent := metadata.PortalLiquidateTopPercentileExchangeRatesContent{
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

// checkAndBuildInstForCustodianLiquidation checks and builds liquidation instructions
// when custodians didn't return public token to users after timeout
func (blockchain *BlockChain) checkAndBuildInstForCustodianLiquidation(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams,
) ([][]string, error) {
	insts := [][]string{}

	// get exchange rate
	exchangeRate := currentPortalState.FinalExchangeRatesState
	if exchangeRate == nil {
		Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when get exchange rate")
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
		if (beaconHeight+1)-redeemReq.GetBeaconHeight() >= blockchain.convertDurationTimeToBeaconBlocks(portalParams.TimeOutCustodianReturnPubToken) {
			// get shardId of redeemer
			redeemerKey, err := wallet.Base58CheckDeserialize(redeemReq.GetRedeemerAddress())
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
					redeemReq.GetUniqueRedeemID(), err)
				continue
			}
			shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

			// get tokenID from redeemTokenID
			tokenID := redeemReq.GetTokenID()

			liquidatedCustodians := make([]*statedb.MatchingRedeemCustodianDetail, 0)
			for _, matchCusDetail := range redeemReq.GetCustodians() {
				//Logger.log.Errorf("matchCusDetail.GetIncognitoAddress(): %v\n", matchCusDetail.GetIncognitoAddress())
				custodianStateKey := statedb.GenerateCustodianStateObjectKey(matchCusDetail.GetIncognitoAddress()).String()
				//Logger.log.Errorf("custodianStateKey: %v\n", custodianStateKey)
				// calculate liquidated amount and remain unlocked amount for custodian
				liquidatedAmount, remainUnlockAmount, err := CalUnlockCollateralAmountAfterLiquidation(
					currentPortalState,
					custodianStateKey,
					matchCusDetail.GetAmount(),
					tokenID,
					exchangeRate,
					portalParams)
				if err != nil {
					Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when calculating unlock collateral amount %v\n: ", err)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.GetUniqueRedeemID(),
						redeemReq.GetTokenID(),
						matchCusDetail.GetAmount(),
						0,
						0,
						redeemReq.GetRedeemerAddress(),
						matchCusDetail.GetIncognitoAddress(),
						liquidatedByExchangeRate,
						metadata.PortalLiquidateCustodianMeta,
						shardID,
						common.PortalLiquidateCustodianFailedChainStatus,
					)
					insts = append(insts, inst)
					continue
				}

				// update custodian state
				custodianState := currentPortalState.CustodianPoolState[custodianStateKey]
				err = updateCustodianStateAfterLiquidateCustodian(custodianState, liquidatedAmount, remainUnlockAmount, tokenID)
				if err != nil {
					Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when updating custodian state %v\n: ", err)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.GetUniqueRedeemID(),
						redeemReq.GetTokenID(),
						matchCusDetail.GetAmount(),
						liquidatedAmount,
						remainUnlockAmount,
						redeemReq.GetRedeemerAddress(),
						matchCusDetail.GetIncognitoAddress(),
						liquidatedByExchangeRate,
						metadata.PortalLiquidateCustodianMeta,
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
					redeemReq.GetRedeemerAddress(),
					matchCusDetail.GetIncognitoAddress(),
					liquidatedByExchangeRate,
					metadata.PortalLiquidateCustodianMeta,
					shardID,
					common.PortalLiquidateCustodianSuccessChainStatus,
				)
				insts = append(insts, inst)
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

// beacon build instruction for expired waiting porting request - user doesn't send public token to custodian after requesting
func buildExpiredWaitingPortingReqInst(
	portingID string,
	expiredByLiquidation bool,
	metaType int,
	shardID byte,
	status string,
) []string {
	liqCustodianContent := metadata.PortalExpiredWaitingPortingReqContent{
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

	//get shardId of redeemer
	redeemerKey, err := wallet.Base58CheckDeserialize(portingReq.PorterAddress())
	if err != nil {
		Logger.log.Errorf("[buildInstForExpiredPortingReqByPortingID] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
			portingReq.UniquePortingID, err)
		return insts, err
	}
	shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

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
		updateCustodianStateAfterExpiredPortingReq(custodianState, matchCusDetail.LockedAmountCollateral, tokenID)
	}

	// remove waiting porting request from waiting list
	delete(currentPortalState.WaitingPortingRequests, portingReqKey)

	// build instruction
	inst := buildExpiredWaitingPortingReqInst(
		portingReq.UniquePortingID(),
		expiredByLiquidation,
		metadata.PortalExpiredWaitingPortingReqMeta,
		shardID,
		common.PortalExpiredWaitingPortingReqSuccessChainStatus,
	)
	insts = append(insts, inst)

	return insts, nil
}

// convertDurationTimeToBeaconBlocks returns number of beacon blocks corresponding to duration time
func (blockchain *BlockChain) convertDurationTimeToBeaconBlocks(duration time.Duration) uint64 {
	return uint64(duration.Seconds() / blockchain.config.ChainParams.MinBeaconBlockInterval.Seconds())
}

func (blockchain *BlockChain) checkAndBuildInstForExpiredWaitingPortingRequest(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	portalParams PortalParams,
) ([][]string, error) {
	insts := [][]string{}
	sortedWaitingPortingReqKeys := make([]string, 0)
	for key := range currentPortalState.WaitingPortingRequests {
		sortedWaitingPortingReqKeys = append(sortedWaitingPortingReqKeys, key)
	}
	sort.Strings(sortedWaitingPortingReqKeys)
	for _, portingReqKey := range sortedWaitingPortingReqKeys {
		portingReq := currentPortalState.WaitingPortingRequests[portingReqKey]
		if (beaconHeight+1)-portingReq.BeaconHeight() >= blockchain.convertDurationTimeToBeaconBlocks(portalParams.TimeOutWaitingPortingRequest) {
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

func checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	liquidatedCustodianState *statedb.CustodianState,
	tokenID string,
	portalParams PortalParams,
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
				metadata.PortalRedeemRequestMeta,
				shardID,
				common.Hash{},
				common.PortalRedeemReqCancelledByLiquidationChainStatus,
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
	portalParams PortalParams) ([][]string, error) {
	if len(currentPortalState.CustodianPoolState) <= 0 {
		return [][]string{}, nil
	}

	insts := [][]string{}
	exchangeRate := currentPortalState.FinalExchangeRatesState
	if exchangeRate == nil {
		Logger.log.Errorf("Exchange key %+v rate not found")
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

		liquidationRatios := map[string]metadata.LiquidateTopPercentileExchangeRatesDetail{}

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
				liquidationRatios[pTokenID] = metadata.LiquidateTopPercentileExchangeRatesDetail{
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
					metadata.PortalLiquidateTPExchangeRatesMeta,
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

func (p *portalRedeemFromLiquidationPoolProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
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
	redeemRequestContent := metadata.PortalRedeemLiquidateExchangeRatesContent{
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
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal redeem liquidate exchange rate action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalRedeemLiquidateExchangeRatesAction
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

func (p *portalCustodianTopupProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
	return nil, nil
}

func buildLiquidationCustodianDepositInst(
	pTokenId string,
	incogAddress string,
	depositedAmount uint64,
	freeCollateralAmount uint64,
	status string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	redeemRequestContent := metadata.PortalLiquidationCustodianDepositContentV2{
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
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal liquidation custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalLiquidationCustodianDepositActionV2
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal liquidation custodian deposit action: %+v", err)
		return [][]string{}, nil
	}

	meta := actionData.Meta
	rejectInst := buildLiquidationCustodianDepositInst(
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
	custodian.SetTotalCollateral(custodian.GetTotalCollateral() + meta.DepositedAmount)
	topUpAmt := meta.DepositedAmount
	if meta.FreeCollateralAmount > 0 {
		topUpAmt += meta.FreeCollateralAmount
		custodian.SetFreeCollateral(custodian.GetFreeCollateral() - meta.FreeCollateralAmount)
	}
	lockedAmountCollateral[meta.PTokenId] += topUpAmt
	custodian.SetLockedAmountCollateral(lockedAmountCollateral)
	currentPortalState.CustodianPoolState[custodianStateKey.String()] = custodian
	inst := buildLiquidationCustodianDepositInst(
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

func (p *portalTopupWaitingPortingReqProcessor) prepareDataBeforeProcessing(stateDB *statedb.StateDB, contentStr string) (map[string]interface{}, error) {
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
	topUpWaitingPortingReqContent := metadata.PortalTopUpWaitingPortingRequestContent{
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
	bc *BlockChain,
	contentStr string,
	shardID byte,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
	portalParams PortalParams,
	optionalData map[string]interface{},
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal top up waiting porting content: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalTopUpWaitingPortingRequestAction
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

	if meta.FreeCollateralAmount > custodian.GetFreeCollateral() {
		Logger.log.Errorf("Free collateral topup amount is greater than free collateral of custodian's state")
		return [][]string{rejectInst}, nil
	}
	custodian.SetTotalCollateral(custodian.GetTotalCollateral() + meta.DepositedAmount)
	topUpAmt := meta.DepositedAmount
	if meta.FreeCollateralAmount > 0 {
		topUpAmt += meta.FreeCollateralAmount
		custodian.SetFreeCollateral(custodian.GetFreeCollateral() - meta.FreeCollateralAmount)
	}
	lockedAmountCollateral := custodian.GetLockedAmountCollateral()
	lockedAmountCollateral[meta.PTokenID] += topUpAmt
	custodian.SetLockedAmountCollateral(lockedAmountCollateral)
	custodiansByPortingID := waitingPortingReq.Custodians()
	for _, cus := range custodiansByPortingID {
		if cus.IncAddress == meta.IncogAddressStr {
			cus.LockedAmountCollateral += topUpAmt
		}
	}
	waitingPortingReq.SetCustodians(custodiansByPortingID)
	currentPortalState.CustodianPoolState[custodianStateKey.String()] = custodian
	currentPortalState.WaitingPortingRequests[waitingPortingRequestKey.String()] = waitingPortingReq
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