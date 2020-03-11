package blockchain

import (
	"encoding/json"
	"github.com/incognitochain/incognito-chain/common"
	"github.com/incognitochain/incognito-chain/database/lvdb"
	"github.com/incognitochain/incognito-chain/metadata"
	"github.com/incognitochain/incognito-chain/wallet"
	"strconv"
)

// beacon build instruction for portal liquidation when custodians run away - don't send public tokens back to users.
func buildCustodianRunAwayLiquidationInst(
	redeemID string,
	tokenID string,
	redeemPubTokenAmount uint64,
	mintedCollateralAmount uint64,
	redeemerIncAddrStr string,
	custodianIncAddrStr string,
	metaType int,
	shardID byte,
	status string,
) []string {
	liqCustodianContent := metadata.PortalLiquidateCustodianContent{
		UniqueRedeemID:         redeemID,
		TokenID:                tokenID,
		RedeemPubTokenAmount:   redeemPubTokenAmount,
		MintedCollateralAmount: mintedCollateralAmount,
		RedeemerIncAddressStr:  redeemerIncAddrStr,
		CustodianIncAddressStr: custodianIncAddrStr,
		ShardID:                shardID,
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
	tpValue map[string]int,
	custodianAddress string,
	exchangeRates lvdb.FinalExchangeRates,
	metaType int,
	shardID byte,
	status string,
) []string {
	tpContent := metadata.TopPercentileExchangeRatesLiquidationContent{
		TPValue: tpValue,
		CustodianAddress: custodianAddress,
		ExchangeRates: exchangeRates,
		MetaType: metaType,
		Status: status,
		ShardID:                shardID,
	}
	tpContentBytes, _ := json.Marshal(tpContent)
	return []string{
		strconv.Itoa(metaType),
		strconv.Itoa(int(shardID)),
		status,
		string(tpContentBytes),
	}
}


func checkAndBuildInstForCustodianLiquidation(
	beaconHeight uint64,
	currentPortalState *CurrentPortalState,
) ([][]string, error) {

	insts := [][]string{}

	// get exchange rate
	exchangeRateKey := lvdb.NewFinalExchangeRatesKey(beaconHeight)
	exchangeRate := currentPortalState.FinalExchangeRates[exchangeRateKey]
	if exchangeRate == nil {
		Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when get exchange rate")
	}

	for redeemReqKey, redeemReq := range currentPortalState.WaitingRedeemRequests {
		if beaconHeight - redeemReq.BeaconHeight >= common.PortalTimeOutCustodianSendPubTokenBack {
			// get shardId of redeemer
			redeemerKey, err := wallet.Base58CheckDeserialize(redeemReq.RedeemerAddress)
			if err != nil {
				Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when deserializing redeemer address string in redeemID %v - %v\n: ",
					redeemReq.UniqueRedeemID, err)
				continue
			}
			shardID := common.GetShardIDFromLastByte(redeemerKey.KeySet.PaymentAddress.Pk[len(redeemerKey.KeySet.PaymentAddress.Pk)-1])

			// get tokenSymbol from redeemTokenID
			tokenSymbol := ""
			for tokenSym, incTokenID := range metadata.PortalSupportedTokenMap {
				if incTokenID == redeemReq.TokenID {
					tokenSymbol = tokenSym
					break
				}
			}

			for cusIncAddr, matchCusDetail := range redeemReq.Custodians {
				// calculate minted collateral amount
				mintedAmountInPToken := matchCusDetail.Amount * common.PercentReceivedCollateralAmount / 100
				mintedAmountInPRV, err := exchangeRate.ExchangePToken2PRVByTokenId(tokenSymbol, mintedAmountInPToken)
				if err != nil {
					Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when exchanging ptoken to prv amount %v\n: ", err)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.UniqueRedeemID,
						redeemReq.TokenID,
						matchCusDetail.Amount,
						0,
						redeemReq.RedeemerAddress,
						cusIncAddr,
						metadata.PortalLiquidateCustodianMeta,
						shardID,
						common.PortalLiquidateCustodianFailedChainStatus,
					)
					insts = append(insts, inst)
					continue
				}

				// update custodian state (total collateral, holding public tokens, locked amount, free collateral)
				cusStateKey := lvdb.NewCustodianStateKey(beaconHeight, cusIncAddr)
				custodianState := currentPortalState.CustodianPoolState[cusStateKey]
				if custodianState == nil {
					Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when get custodian state with key %v\n: ", cusStateKey)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.UniqueRedeemID,
						redeemReq.TokenID,
						matchCusDetail.Amount,
						0,
						redeemReq.RedeemerAddress,
						cusIncAddr,
						metadata.PortalLiquidateCustodianMeta,
						shardID,
						common.PortalLiquidateCustodianFailedChainStatus,
					)
					insts = append(insts, inst)
					continue
				}

				if custodianState.TotalCollateral < mintedAmountInPRV ||
					custodianState.LockedAmountCollateral[tokenSymbol] < mintedAmountInPRV {
					Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Total collateral %v, locked amount %v " +
						"should be greater than minted amount %v\n: ",
						custodianState.TotalCollateral, custodianState.LockedAmountCollateral[tokenSymbol], mintedAmountInPRV)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.UniqueRedeemID,
						redeemReq.TokenID,
						matchCusDetail.Amount,
						mintedAmountInPRV,
						redeemReq.RedeemerAddress,
						cusIncAddr,
						metadata.PortalLiquidateCustodianMeta,
						shardID,
						common.PortalLiquidateCustodianFailedChainStatus,
					)
					insts = append(insts, inst)
					continue
				}

				err = updateCustodianStateAfterLiquidateCustodian(custodianState, mintedAmountInPRV, tokenSymbol)
				if err != nil {
					Logger.log.Errorf("[checkAndBuildInstForCustodianLiquidation] Error when updating %v\n: ", err)
					inst := buildCustodianRunAwayLiquidationInst(
						redeemReq.UniqueRedeemID,
						redeemReq.TokenID,
						matchCusDetail.Amount,
						mintedAmountInPRV,
						redeemReq.RedeemerAddress,
						cusIncAddr,
						metadata.PortalLiquidateCustodianMeta,
						shardID,
						common.PortalLiquidateCustodianFailedChainStatus,
					)
					insts = append(insts, inst)
					continue
				}

				// remove matching custodian from matching custodians list in waiting redeem request
				delete(currentPortalState.WaitingRedeemRequests[redeemReqKey].Custodians, cusIncAddr)

				// build instruction
				inst := buildCustodianRunAwayLiquidationInst(
					redeemReq.UniqueRedeemID,
					redeemReq.TokenID,
					matchCusDetail.Amount,
					mintedAmountInPRV,
					redeemReq.RedeemerAddress,
					cusIncAddr,
					metadata.PortalLiquidateCustodianMeta,
					shardID,
					common.PortalLiquidateCustodianSuccessChainStatus,
				)
				insts = append(insts, inst)
			}

			// remove redeem request from waiting redeem requests list
			if len(currentPortalState.WaitingRedeemRequests[redeemReqKey].Custodians) == 0 {
				delete(currentPortalState.WaitingRedeemRequests, redeemReqKey)
			}
		}
	}

	return insts, nil
}

/*
Top percentile (TP): 150 (TP150), 130 (TP130), 120 (TP120)
if TP down, we are need liquidation custodian and notify to custodians (or users)
 */
func checkTopPercentileExchangeRatesLiquidationInst(beaconHeight uint64, currentPortalState *CurrentPortalState)  ([][]string, error) {
	if len(currentPortalState.CustodianPoolState) <= 0 {
		return [][]string{}, nil
	}

	insts := [][]string{}

	keyExchangeRate := lvdb.NewFinalExchangeRatesKey(beaconHeight)
	exchangeRate, ok := currentPortalState.FinalExchangeRates[keyExchangeRate]
	if !ok {
		return [][]string{}, nil
	}

	custodianPoolState := currentPortalState.CustodianPoolState

	for custodianKey, v := range custodianPoolState {
		result, err := detectMinAspectRatio(v.HoldingPubTokens, v.LockedAmountCollateral, exchangeRate)
		if err != nil {
			Logger.log.Errorf("Detect TP exchange rates error %v", err)
			inst := buildTopPercentileExchangeRatesLiquidationInst(
				nil,
				v.IncognitoAddress,
				*exchangeRate,
				metadata.PortalLiquidateTPExchangeRatesMeta,
				0, //todo: find shard id
				common.PortalLiquidateTPExchangeRatesFailedChainStatus,
			)
			insts = append(insts, inst)
			continue
		}

		tpList := make(map[string]int)
		for ptoken, tpValue := range result {
			if  tpValue > common.TP120 && tpValue <= common.TP130 {
				tpList[ptoken] = tpValue
				continue
			}

			if  tpValue <= common.TP120 {
				tpList[ptoken] = tpValue

				//exec liquidation
				//update custodian
				lockAmount := v.LockedAmountCollateral[ptoken]
				//holdPubtokens := v.HoldingPubTokens[ptoken]

				newCustodian := v
				newCustodian.LockedAmountCollateral[ptoken] = 0
				newCustodian.HoldingPubTokens[ptoken] = 0
				newCustodian.TotalCollateral = newCustodian.TotalCollateral - lockAmount

				currentPortalState.CustodianPoolState[custodianKey] = newCustodian

				//todo: update liquidationState
				continue
			}
		}

		if len(tpList) > 0 {
			inst := buildTopPercentileExchangeRatesLiquidationInst(
				tpList,
				v.IncognitoAddress,
				*exchangeRate,
				metadata.PortalLiquidateTPExchangeRatesMeta,
				0, //todo: find shard id
				common.PortalLiquidateTPExchangeRatesSuccessChainStatus,
			)
			insts = append(insts, inst)
		}
	}

	return insts, nil
}
