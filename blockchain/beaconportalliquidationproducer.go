package blockchain

import (
	"encoding/base64"
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/dataaccessobject/statedb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wallet"
	"math/big"
	"sort"
	"strconv"
	"time"
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
) []string {
	tpContent := metadata.PortalLiquidateTopPercentileExchangeRatesContent{
		CustodianAddress: custodianAddress,
		MetaType:         metaType,
		Status:           status,
		TP:               topPercentile,
	}
	tpContentBytes, _ := json.Marshal(tpContent)
	return []string{
		strconv.Itoa(metaType),
		"-1",
		status,
		string(tpContentBytes),
	}
}

func buildRedeemLiquidateExchangeRatesInst(
	tokenID string,
	redeemAmount uint64,
	incAddressStr string,
	remoteAddress string,
	redeemFee uint64,
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
		RemoteAddress:         remoteAddress,
		RedeemFee:             redeemFee,
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

func buildLiquidationCustodianDepositInst(
	pTokenId string,
	incogAddress string,
	depositedAmount uint64,
	freeCollateralSelected bool,
	status string,
	metaType int,
	shardID byte,
	txReqID common.Hash,
) []string {
	redeemRequestContent := metadata.PortalLiquidationCustodianDepositContent{
		PTokenId:               pTokenId,
		IncogAddressStr:        incogAddress,
		DepositedAmount:        depositedAmount,
		FreeCollateralSelected: freeCollateralSelected,
		TxReqID:                txReqID,
		ShardID:                shardID,
	}
	redeemRequestContentBytes, _ := json.Marshal(redeemRequestContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(redeemRequestContentBytes),
	}
}

// checkAndBuildInstForCustodianLiquidation checks and builds liquidation instructions
// when custodians didn't return public token to users after timeout
func (blockchain *BlockChain) checkAndBuildInstForCustodianLiquidation(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
) ([][]string, error) {
	insts := [][]string{}

	// get exchange rate
	exchangeRateKey := statedb.GeneratePortalFinalExchangeRatesStateObjectKey(beaconHeight)
	exchangeRate := currentPortalState.FinalExchangeRatesState[exchangeRateKey.String()]
	if exchangeRate == nil {
		Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when get exchange rate %v", exchangeRateKey.String())
		return insts, nil
	}

	liquidatedByExchangeRate := false

	sortedWaitingRedeemReqKeys := make([]string, 0)
	for key := range currentPortalState.WaitingRedeemRequests {
		sortedWaitingRedeemReqKeys = append(sortedWaitingRedeemReqKeys, key)
	}
	sort.Strings(sortedWaitingRedeemReqKeys)
	for _, redeemReqKey := range sortedWaitingRedeemReqKeys {
		redeemReq := currentPortalState.WaitingRedeemRequests[redeemReqKey]
		if (beaconHeight+1)-redeemReq.GetBeaconHeight() >= blockchain.convertPortalTimeOutToBeaconBlocks(common.PortalTimeOutCustodianReturnPubToken) {
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

			for _, matchCusDetail := range redeemReq.GetCustodians() {
				custodianStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, matchCusDetail.GetIncognitoAddress())
				// calculate liquidated amount and remain unlocked amount for custodian
				liquidatedAmount, remainUnlockAmount, err := CalUnlockCollateralAmountAfterLiquidation(
					currentPortalState,
					custodianStateKey.String(),
					matchCusDetail,
					tokenID,
					exchangeRate)
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
				cusStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, matchCusDetail.GetIncognitoAddress())
				cusStateKeyStr := cusStateKey.String()
				custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
				err = updateCustodianStateAfterLiquidateCustodian(custodianState, liquidatedAmount, remainUnlockAmount, tokenID)
				if err != nil {
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
				updatedCustodians, _ := removeCustodianFromMatchingRedeemCustodians(
					currentPortalState.WaitingRedeemRequests[redeemReqKey].GetCustodians(), matchCusDetail.GetIncognitoAddress())
				currentPortalState.WaitingRedeemRequests[redeemReqKey].SetCustodians(updatedCustodians)

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

			// remove redeem request from waiting redeem requests list
			if len(currentPortalState.WaitingRedeemRequests[redeemReqKey].GetCustodians()) == 0 {
				deleteWaitingRedeemRequest(currentPortalState, redeemReqKey)
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
		cusStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, matchCusDetail.IncAddress)
		cusStateKeyStr := cusStateKey.String()
		custodianState := currentPortalState.CustodianPoolState[cusStateKeyStr]
		if custodianState == nil {
			Logger.log.Errorf("[checkAndBuildInstForExpiredWaitingPortingRequest] Error when get custodian state with key %v\n: ", cusStateKey)
			continue
		}
		updateCustodianStateAfterExpiredPortingReq(custodianState, matchCusDetail.LockedAmountCollateral, matchCusDetail.Amount, tokenID)
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

// convertPortalTimeOutToBeaconBlocks returns number of beacon blocks corresponding to duration time
func (blockchain *BlockChain) convertPortalTimeOutToBeaconBlocks(duration time.Duration) uint64 {
	return uint64(duration.Seconds() / blockchain.config.ChainParams.MinBeaconBlockInterval.Seconds())
}

func (blockchain *BlockChain) checkAndBuildInstForExpiredWaitingPortingRequest(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
) ([][]string, error) {
	insts := [][]string{}
	sortedWaitingPortingReqKeys := make([]string, 0)
	for key := range currentPortalState.WaitingPortingRequests {
		sortedWaitingPortingReqKeys = append(sortedWaitingPortingReqKeys, key)
	}
	sort.Strings(sortedWaitingPortingReqKeys)
	for _, portingReqKey := range sortedWaitingPortingReqKeys {
		portingReq := currentPortalState.WaitingPortingRequests[portingReqKey]
		if (beaconHeight+1)-portingReq.BeaconHeight() >= blockchain.convertPortalTimeOutToBeaconBlocks(common.PortalTimeOutWaitingPortingRequest) {
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

func checkAndBuildInstForTPExchangeRateRedeemRequest(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	exchangeRate *statedb.FinalExchangeRatesState,
	liquidatedCustodianState *statedb.CustodianState,
	tokenID string,
) ([][]string, error) {

	insts := [][]string{}

	// calculate total amount of matching redeem amount with the liquidated custodian
	totalMatchingRedeemAmountPubToken := uint64(0)
	for _, redeemReq := range currentPortalState.WaitingRedeemRequests {
		if redeemReq.GetTokenID() == tokenID {
			for _, cus := range redeemReq.GetCustodians() {
				if cus.GetIncognitoAddress() == liquidatedCustodianState.GetIncognitoAddress() {
					totalMatchingRedeemAmountPubToken += cus.GetAmount()
				}
			}
		}
	}

	convertExchangeRatesObj := NewConvertExchangeRatesObject(exchangeRate)

	// calculate total minted amount prv for liquidate (maximum 120% amount)
	totalMatchingRedeemAmountPubTokenInPRV, err := convertExchangeRatesObj.ExchangePToken2PRVByTokenId(tokenID, totalMatchingRedeemAmountPubToken)
	if err != nil {
		Logger.log.Errorf("[checkAndBuildInstForTPExchangeRateRedeemRequest] Error when convert total amount public token to prv %v", err)
		return insts, err
	}

	totalMintedTmp := new(big.Int).Mul(new(big.Int).SetUint64(totalMatchingRedeemAmountPubTokenInPRV), new(big.Int).SetUint64(common.PercentReceivedCollateralAmount))
	totalMintedAmountPRV := new(big.Int).Div(totalMintedTmp, new(big.Int).SetUint64(100)).Uint64()

	if totalMintedAmountPRV > liquidatedCustodianState.GetLockedAmountCollateral()[tokenID] {
		totalMintedAmountPRV = liquidatedCustodianState.GetLockedAmountCollateral()[tokenID]
	}

	// calculate minted amount prv for each matching redeem requests
	// rely on percent matching redeem amount and total matching redeem amount
	liquidatedByExchangeRate := true
	sortedWaitingRedeemReqKeys := make([]string, 0)
	for key := range currentPortalState.WaitingRedeemRequests {
		sortedWaitingRedeemReqKeys = append(sortedWaitingRedeemReqKeys, key)
	}
	sort.Strings(sortedWaitingRedeemReqKeys)

	for _, redeemReqKey := range sortedWaitingRedeemReqKeys {
		redeemReq := currentPortalState.WaitingRedeemRequests[redeemReqKey]
		if redeemReq.GetTokenID() == tokenID {
			for _, matchCustodian := range redeemReq.GetCustodians() {
				if matchCustodian.GetIncognitoAddress() == liquidatedCustodianState.GetIncognitoAddress() {
					tmp := new(big.Int).Mul(new(big.Int).SetUint64(matchCustodian.GetAmount()), new(big.Int).SetUint64(totalMintedAmountPRV))
					mintedAmountPRV := new(big.Int).Div(tmp, new(big.Int).SetUint64(totalMatchingRedeemAmountPubToken)).Uint64()

					// get shardId of redeemer
					redeemerKey, err := wallet.Base58CheckDeserialize(redeemReq.GetRedeemerAddress())
					if err != nil {
						Logger.log.Errorf("[checkAndBuildInstForTPExchangeRateRedeemRequest] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
							redeemReq.GetUniqueRedeemID(), err)
						continue
					}
					shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

					// remove matching custodian from matching custodians list in waiting redeem request
					updatedCustodians, _ := removeCustodianFromMatchingRedeemCustodians(
						currentPortalState.WaitingRedeemRequests[redeemReqKey].GetCustodians(), matchCustodian.GetIncognitoAddress())
					currentPortalState.WaitingRedeemRequests[redeemReqKey].SetCustodians(updatedCustodians)

					// build instruction
					//todo: need to update remainUnlockAmount
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.GetUniqueRedeemID(),
						redeemReq.GetTokenID(),
						matchCustodian.GetAmount(),
						mintedAmountPRV,
						0,
						redeemReq.GetRedeemerAddress(),
						matchCustodian.GetIncognitoAddress(),
						liquidatedByExchangeRate,
						metadata.PortalLiquidateCustodianMeta,
						shardID,
						common.PortalLiquidateCustodianSuccessChainStatus,
					)
					insts = append(insts, inst)

				}
			}
			// remove redeem request from waiting redeem requests list
			if len(currentPortalState.WaitingRedeemRequests[redeemReqKey].GetCustodians()) == 0 {
				deleteWaitingRedeemRequest(currentPortalState, redeemReqKey)
			}
		}
	}
	// update custodian state (update locked amount, holding public token amount)
	custodianStateKey := statedb.GenerateCustodianStateObjectKey(beaconHeight, liquidatedCustodianState.GetIncognitoAddress())
	custodianStateKeyStr := custodianStateKey.String()

	lockedAmountTmp := currentPortalState.CustodianPoolState[custodianStateKeyStr].GetLockedAmountCollateral()
	lockedAmountTmp[tokenID] -= totalMintedAmountPRV
	currentPortalState.CustodianPoolState[custodianStateKeyStr].SetLockedAmountCollateral(lockedAmountTmp)

	return insts, nil
}

func checkAndBuildInstRejectRedeemRequestByLiquidationExchangeRate(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
	liquidatedCustodianState *statedb.CustodianState,
	tokenID string,
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
				common.PortalRedeemRequestRejectedByLiquidationChainStatus,
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
func buildInstForLiquidationTopPercentileExchangeRates(beaconHeight uint64, currentPortalState *CurrentPortalState) ([][]string, error) {
	if len(currentPortalState.CustodianPoolState) <= 0 {
		return [][]string{}, nil
	}

	insts := [][]string{}
	keyExchangeRate := statedb.GeneratePortalFinalExchangeRatesStateObjectKey(beaconHeight)
	keyExchangeRateStr := keyExchangeRate.String()
	exchangeRate, ok := currentPortalState.FinalExchangeRatesState[keyExchangeRateStr]
	if !ok {
		Logger.log.Errorf("Exchange key %+v rate not found", keyExchangeRateStr)
		return [][]string{}, nil
	}

	custodianPoolState := currentPortalState.CustodianPoolState
	sortedCustodianStateKeys := make([]string, 0)
	for key := range custodianPoolState {
		sortedCustodianStateKeys = append(sortedCustodianStateKeys, key)
	}
	sort.Strings(sortedCustodianStateKeys)
	//for _, custodianKey := range sortedCustodianStateKeys {
	//	custodianState := custodianPoolState[custodianKey]
	//
	//	calTPRatio, err := calculateTPRatio(custodianState.GetHoldingPublicTokens(), custodianState.GetLockedAmountCollateral(), exchangeRate)
	//	if err != nil {
	//		Logger.log.Errorf("Auto liquidation: cal tp ratio error %v", err)
	//		continue
	//	}
	//
	//	//filter TP by TP 120 or TP130
	//	detectTp, err := detectTopPercentileLiquidation(custodianState, calTPRatio)
	//	if err != nil {
	//		Logger.log.Errorf("Auto liquidation: detect cal tp ratio error %v", err)
	//		continue
	//	}
	//
	//	isUpdateDetectTp := false
	//	if len(detectTp) > 0 {
	//		sortedDetectTPKeys := make([]string, 0)
	//		for key := range detectTp {
	//			sortedDetectTPKeys = append(sortedDetectTPKeys, key)
	//		}
	//		sort.Strings(sortedDetectTPKeys)
	//
	//		for _, pTokenID := range sortedDetectTPKeys {
	//			v := detectTp[pTokenID]
	//			if v.HoldAmountFreeCollateral > 0 {
	//				// check and build instruction for waiting redeem request
	//				instsFromRedeemRequest, err := checkAndBuildInstForTPExchangeRateRedeemRequest(
	//					beaconHeight,
	//					currentPortalState,
	//					exchangeRate,
	//					custodianState,
	//					pTokenID,
	//				)
	//				if err != nil {
	//					Logger.log.Errorf("Error when check and build instruction from redeem request %v\n", err)
	//					continue
	//				}
	//				if len(instsFromRedeemRequest) > 0 {
	//					isUpdateDetectTp = true
	//					Logger.log.Infof("There is %v instructions for tp exchange rate for redeem request", len(instsFromRedeemRequest))
	//					insts = append(insts, instsFromRedeemRequest...)
	//				}
	//
	//				// Note: don't liquidate waiting porting requests in this case
	//			}
	//		}
	//	}
	//
	//	// re-calculate detect tp
	//	if isUpdateDetectTp {
	//		calTPRatio, err = calculateTPRatio(custodianState.GetHoldingPublicTokens(), custodianState.GetLockedAmountCollateral(), exchangeRate)
	//		if err != nil {
	//			Logger.log.Errorf("Auto liquidation: cal tp ratio error %v", err)
	//			continue
	//		}
	//
	//		//filter TP by TP 120 or TP130
	//		detectTp, err = detectTopPercentileLiquidation(custodianState, calTPRatio)
	//		if err != nil {
	//			Logger.log.Errorf("Auto liquidation: detect cal tp ratio error %v", err)
	//			continue
	//		}
	//	}
	//
	//	if len(detectTp) > 0 {
	//		// remove locked amount and holding public token in waiting porting request before pushing into liquidation pool
	//		detectTp = updateDetectTPExcludeWaitingPorting(detectTp, currentPortalState, custodianState)
	//		//Logger.log.Errorf("buildInstForLiquidationTopPercentileExchangeRates custodianState.GetHoldingPublicTokens() %v\n", custodianState.GetHoldingPublicTokens())
	//		inst := buildTopPercentileExchangeRatesLiquidationInst(
	//			custodianState.GetIncognitoAddress(),
	//			metadata.PortalLiquidateTPExchangeRatesMeta,
	//			common.PortalLiquidateTPExchangeRatesSuccessChainStatus,
	//			detectTp,
	//		)
	//
	//		//update current portal state
	//		updateCurrentPortalStateOfLiquidationExchangeRates(beaconHeight, currentPortalState, custodianKey, custodianState, detectTp)
	//
	//		insts = append(insts, inst)
	//	}
	//}

	for _, custodianKey := range sortedCustodianStateKeys {
		custodianState := custodianPoolState[custodianKey]
		tpRatios, err := calAndCheckTPRatio(currentPortalState, custodianState, exchangeRate)
		if err != nil {
			Logger.log.Errorf("Error when calculating and checking tp ratio %v", err)
		}

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
					)
					if err != nil {
						Logger.log.Errorf("Error when check and build instruction from redeem request %tpRatioDetail\n", err)
						continue
					}
					if len(instsFromRedeemRequest) > 0 {
						Logger.log.Infof("There is % tpRatioDetail instructions for tp exchange rate for redeem request", len(instsFromRedeemRequest))
						insts = append(insts, instsFromRedeemRequest...)
					}
				}
			}

			//update current portal state
			updateCurrentPortalStateOfLiquidationExchangeRates(beaconHeight, currentPortalState, custodianKey, custodianState, tpRatios)
			inst := buildTopPercentileExchangeRatesLiquidationInst(
				custodianState.GetIncognitoAddress(),
				metadata.PortalLiquidateTPExchangeRatesMeta,
				common.PortalLiquidateTPExchangeRatesSuccessChainStatus,
				tpRatios,
			)
			insts = append(insts, inst)
		}
	}

	return insts, nil
}

func updateDetectTPExcludeWaitingPorting(
	detectTp map[string]metadata.LiquidateTopPercentileExchangeRatesDetail,
	currentPortalState *CurrentPortalState,
	custodianState *statedb.CustodianState,
) map[string]metadata.LiquidateTopPercentileExchangeRatesDetail {
	for tokenID, value := range detectTp {
		if value.HoldAmountPubToken <= 0 {
			continue
		}
		for _, portingReq := range currentPortalState.WaitingPortingRequests {
			if portingReq.TokenID() != tokenID {
				continue
			}
			for _, cus := range portingReq.Custodians() {
				if cus.IncAddress != custodianState.GetIncognitoAddress() {
					continue
				}
				value.HoldAmountPubToken -= cus.Amount
				value.HoldAmountFreeCollateral -= cus.LockedAmountCollateral
				detectTp[tokenID] = value
				break
			}
		}
	}
	return detectTp
}

func (blockchain *BlockChain) buildInstructionsForLiquidationRedeemPTokenExchangeRates(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
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
	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null.")
		// need to mint ptoken to user
		inst := buildRedeemLiquidateExchangeRatesInst(
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			0,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	//get exchange rates
	exchangeRatesKey := statedb.GeneratePortalFinalExchangeRatesStateObjectKey(beaconHeight)
	exchangeRatesState, ok := currentPortalState.FinalExchangeRatesState[exchangeRatesKey.String()]
	if !ok {
		Logger.log.Errorf("exchange rates not found")
		inst := buildRedeemLiquidateExchangeRatesInst(
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			0,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	minRedeemFee, err := CalMinRedeemFee(meta.RedeemAmount, meta.TokenID, exchangeRatesState)
	if err != nil {
		Logger.log.Errorf("Error when calculating minimum redeem fee %v", err)
		inst := buildRedeemLiquidateExchangeRatesInst(
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			0,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	if meta.RedeemFee < minRedeemFee {
		Logger.log.Errorf("Redeem fee is invalid, minRedeemFee %v, but get %v\n", minRedeemFee, meta.RedeemFee)
		inst := buildRedeemLiquidateExchangeRatesInst(
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			0,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	//check redeem amount
	liquidateExchangeRatesKey := statedb.GeneratePortalLiquidateExchangeRatesPoolObjectKey(beaconHeight)
	liquidateExchangeRates, ok := currentPortalState.LiquidateExchangeRatesPool[liquidateExchangeRatesKey.String()]

	if !ok {
		Logger.log.Errorf("Liquidate exchange rates not found")
		inst := buildRedeemLiquidateExchangeRatesInst(
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			0,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	liquidateByTokenID, ok := liquidateExchangeRates.Rates()[meta.TokenID]

	if !ok {
		Logger.log.Errorf("Liquidate exchange rates not found")
		inst := buildRedeemLiquidateExchangeRatesInst(
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			0,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	totalPrv, err := calTotalLiquidationByExchangeRates(meta.RedeemAmount, liquidateByTokenID)

	if err != nil {
		Logger.log.Errorf("Calculate total liquidation error %v", err)
		inst := buildRedeemLiquidateExchangeRatesInst(
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			0,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	//todo: review
	if totalPrv > liquidateByTokenID.HoldAmountFreeCollateral || liquidateByTokenID.HoldAmountFreeCollateral <= 0 {
		Logger.log.Errorf("amout free collateral not enough, need prv %v != hold amount free collateral %v", totalPrv, liquidateByTokenID.HoldAmountFreeCollateral)
		inst := buildRedeemLiquidateExchangeRatesInst(
			meta.TokenID,
			meta.RedeemAmount,
			meta.RedeemerIncAddressStr,
			meta.RemoteAddress,
			meta.RedeemFee,
			0,
			meta.Type,
			actionData.ShardID,
			actionData.TxReqID,
			common.PortalRedeemLiquidateExchangeRatesRejectedChainStatus,
		)
		return [][]string{inst}, nil
	}

	Logger.log.Infof("Redeem Liquidation: Amount refund to user amount ptoken %v, amount prv %v", meta.RedeemAmount, totalPrv)
	liquidateExchangeRates.Rates()[meta.TokenID] = statedb.LiquidateExchangeRatesDetail{
		HoldAmountFreeCollateral: liquidateByTokenID.HoldAmountFreeCollateral - totalPrv,
		HoldAmountPubToken:       liquidateByTokenID.HoldAmountPubToken - meta.RedeemAmount,
	}

	currentPortalState.LiquidateExchangeRatesPool[liquidateExchangeRatesKey.String()] = liquidateExchangeRates

	inst := buildRedeemLiquidateExchangeRatesInst(
		meta.TokenID,
		meta.RedeemAmount,
		meta.RedeemerIncAddressStr,
		meta.RemoteAddress,
		meta.RedeemFee,
		totalPrv,
		meta.Type,
		actionData.ShardID,
		actionData.TxReqID,
		common.PortalRedeemLiquidateExchangeRatesSuccessChainStatus,
	)
	return [][]string{inst}, nil
}

func (blockchain *BlockChain) buildInstructionsForLiquidationCustodianDeposit(
	contentStr string,
	shardID byte,
	metaType int,
	currentPortalState *CurrentPortalState,
	beaconHeight uint64,
) ([][]string, error) {
	// parse instruction
	actionContentBytes, err := base64.StdEncoding.DecodeString(contentStr)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while decoding content string of portal liquidation custodian deposit action: %+v", err)
		return [][]string{}, nil
	}
	var actionData metadata.PortalLiquidationCustodianDepositAction
	err = json.Unmarshal(actionContentBytes, &actionData)
	if err != nil {
		Logger.log.Errorf("ERROR: an error occurred while unmarshal portal liquidation custodian deposit action: %+v", err)
		return [][]string{}, nil
	}

	if currentPortalState == nil {
		Logger.log.Warn("Current Portal state is null.")
		// need to refund collateral to custodian
		inst := buildLiquidationCustodianDepositInst(
			actionData.Meta.PTokenId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedChainStatus,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	meta := actionData.Meta

	keyCustodianState := statedb.GenerateCustodianStateObjectKey(beaconHeight, meta.IncogAddressStr)
	custodian, ok := currentPortalState.CustodianPoolState[keyCustodianState.String()]

	if !ok {
		Logger.log.Errorf("Custodian not found")
		// need to refund collateral to custodian
		inst := buildLiquidationCustodianDepositInst(
			actionData.Meta.PTokenId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedChainStatus,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	//check exit ptoken
	if _, ok := custodian.GetLockedAmountCollateral()[actionData.Meta.PTokenId]; !ok {
		Logger.log.Errorf("PToken not found")
		// need to refund collateral to custodian
		inst := buildLiquidationCustodianDepositInst(
			actionData.Meta.PTokenId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedChainStatus,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	keyExchangeRate := statedb.GeneratePortalFinalExchangeRatesStateObjectKey(beaconHeight)
	exchangeRate, ok := currentPortalState.FinalExchangeRatesState[keyExchangeRate.String()]
	if !ok {
		Logger.log.Errorf("Exchange rate not found", err)
		inst := buildLiquidationCustodianDepositInst(
			actionData.Meta.PTokenId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedChainStatus,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	calTPRatio, err := calculateTPRatio(custodian.GetHoldingPublicTokens(), custodian.GetLockedAmountCollateral(), exchangeRate)
	if err != nil {
		Logger.log.Errorf("Custodian deposit: cal tp ratio error %v", err)
		inst := buildLiquidationCustodianDepositInst(
			actionData.Meta.PTokenId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedChainStatus,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	//filter TP by TP 120 & TP130
	detectTp, err := detectTopPercentileLiquidation(custodian, calTPRatio)

	if err != nil {
		Logger.log.Errorf("Detect TP liquidation error %v", err)
		inst := buildLiquidationCustodianDepositInst(
			actionData.Meta.PTokenId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedChainStatus,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	tpItem, ok := detectTp[actionData.Meta.PTokenId]

	if tpItem.TPKey != common.TP130 || !ok {
		Logger.log.Errorf("TP value is must TP130")
		inst := buildLiquidationCustodianDepositInst(
			actionData.Meta.PTokenId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedChainStatus,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	amountNeeded, totalFreeCollateralNeeded, remainFreeCollateral, err := CalAmountNeededDepositLiquidate(custodian, exchangeRate, actionData.Meta.PTokenId, actionData.Meta.FreeCollateralSelected)

	if err != nil {
		Logger.log.Errorf("Calculate amount needed deposit err %v", err)
		inst := buildLiquidationCustodianDepositInst(
			actionData.Meta.PTokenId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedChainStatus,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	if actionData.Meta.DepositedAmount < amountNeeded {
		Logger.log.Errorf("Deposited amount is not enough, expect %v, data sent %v", amountNeeded, actionData.Meta.DepositedAmount)
		inst := buildLiquidationCustodianDepositInst(
			actionData.Meta.PTokenId,
			actionData.Meta.IncogAddressStr,
			actionData.Meta.DepositedAmount,
			actionData.Meta.FreeCollateralSelected,
			common.PortalLiquidationCustodianDepositRejectedChainStatus,
			actionData.Meta.Type,
			shardID,
			actionData.TxReqID,
		)

		return [][]string{inst}, nil
	}

	Logger.log.Infof("Deposited amount: expect %v, data sent %v", amountNeeded, actionData.Meta.DepositedAmount)

	remainDepositAmount := actionData.Meta.DepositedAmount - amountNeeded
	custodian.SetTotalCollateral(custodian.GetTotalCollateral() + actionData.Meta.DepositedAmount)

	if actionData.Meta.FreeCollateralSelected == false {
		lockedAmountCollateral := custodian.GetLockedAmountCollateral()
		lockedAmountCollateral[actionData.Meta.PTokenId] = lockedAmountCollateral[actionData.Meta.PTokenId] + amountNeeded

		custodian.SetLockedAmountCollateral(lockedAmountCollateral)
		//update remain
		custodian.SetFreeCollateral(custodian.GetFreeCollateral() + remainDepositAmount)
	} else {
		lockedAmountCollateral := custodian.GetLockedAmountCollateral()
		lockedAmountCollateral[actionData.Meta.PTokenId] = lockedAmountCollateral[actionData.Meta.PTokenId] + amountNeeded + totalFreeCollateralNeeded
		//deposit from free collateral DepositedAmount
		custodian.SetLockedAmountCollateral(lockedAmountCollateral)
		custodian.SetFreeCollateral(remainFreeCollateral + remainDepositAmount)
	}

	currentPortalState.CustodianPoolState[keyCustodianState.String()] = custodian

	inst := buildLiquidationCustodianDepositInst(
		actionData.Meta.PTokenId,
		actionData.Meta.IncogAddressStr,
		actionData.Meta.DepositedAmount,
		actionData.Meta.FreeCollateralSelected,
		common.PortalLiquidationCustodianDepositSuccessChainStatus,
		actionData.Meta.Type,
		shardID,
		actionData.TxReqID,
	)

	return [][]string{inst}, nil
}
